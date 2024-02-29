// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package release

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	deploy "cloud.google.com/go/deploy/apiv1"
	"cloud.google.com/go/deploy/apiv1/deploypb"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/gcs"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/rollout"
	"github.com/google/uuid"
)

// CreateCloudDeployRelease is the main entry to create a Release
func CreateCloudDeployRelease(ctx context.Context, cdClient *deploy.CloudDeployClient, gcsClient *storage.Client, flags *config.ReleaseConfiguration) error {
	// TODO: Add implementation
	uuid, err := fetchReleasePipeline(ctx, cdClient, flags)
	if err != nil {
		return err
	}

	// create a release
	if err := createRelease(ctx, cdClient, gcsClient, flags, uuid); err != nil {
		return err
	}

	// rollout to target
	if err := rollout.CreateRollout(ctx, cdClient, flags); err != nil {
		return err
	}

	return nil
}

// fetchReleasePipeline calls Cloud Deploy API to get the target Delivery Pipeline.
// It returns the ID of the Delivery Pipeline is found, return error otherwise.
func fetchReleasePipeline(ctx context.Context, cdClient *deploy.CloudDeployClient, flags *config.ReleaseConfiguration) (string, error) {
	req := &deploypb.GetDeliveryPipelineRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline),
	}
	dp, err := cdClient.GetDeliveryPipeline(ctx, req)
	if err != nil {
		return "", err
	}

	return dp.Uid, nil
}

func createRelease(ctx context.Context, cdClient *deploy.CloudDeployClient, gcsClient *storage.Client, flags *config.ReleaseConfiguration, pipelineUUID string) error {
	release := &deploypb.Release{
		Name:        fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline, flags.Release),
		Description: flags.Description,
	}

	if flags.SkaffoldVersion != "" {
		if err := validateSupportedSkaffoldVersion(ctx, flags, cdClient); err != nil {
			return err
		}

		release.SkaffoldVersion = flags.SkaffoldVersion
	}

	if err := gcs.SetSource(ctx, pipelineUUID, flags, gcsClient, release); err != nil {
		return fmt.Errorf("failed to set source: %w", err)
	}
	if err := setImages(flags.Images, release); err != nil {
		return fmt.Errorf("failed to set images: %w", err)
	}
	if err := setSkaffoldFile(ctx, flags, release); err != nil {
		return fmt.Errorf("failed to set skaffold file: %w", err)
	}

	req := &deploypb.CreateReleaseRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline),
		ReleaseId: flags.Release,
		Release:   release,
		RequestId: uuid.NewString(),
	}

	fmt.Printf("Creating Cloud Deploy release: %s... \n", flags.Release)
	op, err := cdClient.CreateRelease(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Created Cloud Deploy release: %s... \n", flags.Release)

	return nil
}

// setImages sets the release.BuildArtifacts field or it returns error when failed
func setImages(images map[string]string, releaseConfig *deploypb.Release) error {
	// sort by key for testing
	keys := make([]string, 0, len(images))
	for k := range images {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buildArtifacts := []*deploypb.BuildArtifact{}
	for _, key := range keys {
		buildArtifacts = append(buildArtifacts, &deploypb.BuildArtifact{
			Image: key,
			Tag:   images[key],
		})
	}

	releaseConfig.BuildArtifacts = buildArtifacts
	return nil
}

// setSkaffoldFile sets the release.SkaffoldConfigPath field or it returns error when failed
func setSkaffoldFile(ctx context.Context, flags *config.ReleaseConfiguration, releaseConfig *deploypb.Release) error {
	parsedSkaffoldFile := flags.SkaffoldFile
	var err error

	// only when skaffold is absolute path need to be handled here
	if flags.SkaffoldFile != "" && filepath.IsAbs(flags.SkaffoldFile) {
		parsedSkaffoldFile, err = skaffoldFileAbsolutePath(ctx, flags)
		if err != nil {
			return err
		}
	}

	if err := validateSkaffoldFileExist(flags.Source, parsedSkaffoldFile); err != nil {
		return err
	}

	releaseConfig.SkaffoldConfigPath = parsedSkaffoldFile
	return nil
}

// skaffoldFileAbsolutePath returns the absolute skaffold file path relative to source, or returns error when it fails
func skaffoldFileAbsolutePath(ctx context.Context, flags *config.ReleaseConfiguration) (string, error) {
	var sourcePath, desc string
	if flags.Source == "." {
		sourcePath, _ = os.Getwd()
		desc = "current working directory"
	} else {
		sourcePath = flags.Source
		desc = "source"
	}

	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", fmt.Errorf("unexpected err when finding source abs path: %s", err)
	}
	parent := filepath.Clean(absSource)
	child := filepath.Clean(flags.SkaffoldFile)

	info, err := os.Stat(parent)
	if err != nil {
		return "", fmt.Errorf("cannot open local source: %s, err: %s", parent, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("local source: %s is not a directory", parent)
	}

	relPath, err := filepath.Rel(parent, child)
	if err != nil {
		return "", fmt.Errorf("unexpected err when finding skaffold file relative path: %s", err)
	}
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("the skaffold file %s could not be found in the %s. Please enter a valid Skaffold file path", flags.SkaffoldFile, desc)
	}

	return relPath, nil
}

func validateSkaffoldFileExist(source, skaffoldFile string) error {
	if strings.HasPrefix(source, "gs://") {
		fmt.Println("Skipping skaffold file check: source is not a local archive or directory")
		return nil
	}

	if skaffoldFile == "" {
		skaffoldFile = "skaffold.yaml"
	}
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("local source: %s does not exist", source)
	}
	if info.Mode().IsDir() {
		pathToSkaffold := filepath.Join(source, skaffoldFile)
		if _, err := os.Stat(pathToSkaffold); err != nil {
			return fmt.Errorf("could not find skaffold file: %s does not exist", pathToSkaffold)
		}
	} else {
		return validateSkaffoldIsInArchive(source, skaffoldFile)
	}

	return nil
}

func validateSkaffoldIsInArchive(source, skaffoldFile string) error {
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("specified source file: %s does not exist", source)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("specified source file: %s is not a readable compressed file archive", source)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if header.Name == skaffoldFile {
			return nil // File found
		}
	}

	return fmt.Errorf("specified skaffold file: %s is not found in source archive: %s", skaffoldFile, source)
}

// validateSupportedSkaffoldVersion validates the --skaffold-version input against Skaffold maintainance and support window
func validateSupportedSkaffoldVersion(ctx context.Context, flags *config.ReleaseConfiguration, client *deploy.CloudDeployClient) error {
	config, err := getCloudDeployConfig(ctx, flags.ProjectId, flags.Region, client)
	if err != nil {
		return err
	}

	var versionObj *deploypb.SkaffoldVersion
	for _, v := range config.SupportedVersions {
		if v.Version == flags.SkaffoldVersion {
			versionObj = v
		}
	}
	if versionObj == nil {
		return nil
	}

	// validate skaffold version support window
	maintenanceDt := versionObj.MaintenanceModeTime.AsTime()
	if err != nil {
		return err
	}
	supportExpirationDt := versionObj.SupportExpirationTime.AsTime()
	if err != nil {
		return err
	}

	if maintenanceDt.Sub(time.Now()) <= 28*24*time.Hour { // 28 days
		fmt.Printf(
			"WARNING: this release's Skaffold version will be in maintenance mode beginning on %s. "+
				"After that you won't be able to create releases using this version of Skaffold.\n"+
				"https://cloud.google.com/deploy/docs/using-skaffold/select-skaffold#skaffold_version_deprecation_and_maintenance_policy",
			maintenanceDt.Format("2006-01-02"),
		)
	}
	if time.Now().After(supportExpirationDt) {
		return fmt.Errorf("the Skaffold version you've chosen is no longer supported.\n" +
			"https://cloud.google.com/deploy/docs/using-skaffold/select-skaffold#skaffold_version_deprecation_and_maintenance_policy")
	}
	if time.Now().After(maintenanceDt) {
		return fmt.Errorf("you can't create a new release using a Skaffold version that is in maintenance mode.\n" +
			"https://cloud.google.com/deploy/docs/using-skaffold/select-skaffold#skaffold_version_deprecation_and_maintenance_policy")
	}

	return nil
}

// getCloudDeployConfig returns deploypb.Config based on projectId and region
func getCloudDeployConfig(ctx context.Context, projectId, region string, client *deploy.CloudDeployClient) (*deploypb.Config, error) {
	req := &deploypb.GetConfigRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/config", projectId, region),
	}
	config, err := client.GetConfig(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get Cloud Deploy Config: %v", err)
	}

	return config, nil
}

// parseDictString converts the image string to a map[string]string
// Ex:
// image1=path/to/image1:v1,image2=path/to/image2:v1 =>
// {"image1": "path/to/image1:v1", "image2": "path/to/image2:v1"}
func ParseDictString(input string) (map[string]string, error) {
	if input == "" {
		return nil, nil
	}

	if strings.Contains(input, " ") {
		return nil, fmt.Errorf("invalid dict value: %s, the dict string should not contain space", input)
	}
	res := make(map[string]string)
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid key-value pair: %s", pair)
		}

		key := kv[0]
		val := kv[1]

		res[key] = val
	}

	return res, nil
}
