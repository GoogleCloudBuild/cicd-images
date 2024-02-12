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
	"context"
	"fmt"
	"sort"
	"strings"

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
	//TODO: add more release config fields
	releaseConfig := &deploypb.Release{
		Name: fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline, flags.Release),
	}

	if err := gcs.SetSource(ctx, pipelineUUID, flags, gcsClient, releaseConfig); err != nil {
		return fmt.Errorf("failed to set source: %w", err)
	}

	if err := setImages(flags.Images, releaseConfig); err != nil {
		return fmt.Errorf("failed to set images: %w", err)
	}

	req := &deploypb.CreateReleaseRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline),
		ReleaseId: flags.Release,
		Release:   releaseConfig,
		RequestId: uuid.NewString(),
	}

	fmt.Printf("creating Cloud Deploy release: %s... \n", flags.Release)
	op, err := cdClient.CreateRelease(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("created Cloud Deploy release: %s... \n", flags.Release)

	return nil
}

func setImages(images string, releaseConfig *deploypb.Release) error {
	if images == "" {
		return nil
	}

	imgMap, err := parseImgString(images)
	if err != nil {
		return err
	}

	// sort by key for testing
	keys := make([]string, 0, len(imgMap))
	for k := range imgMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buildArtifacts := []*deploypb.BuildArtifact{}
	for _, key := range keys {
		buildArtifacts = append(buildArtifacts, &deploypb.BuildArtifact{
			Image: key,
			Tag:   imgMap[key],
		})
	}

	releaseConfig.BuildArtifacts = buildArtifacts
	return nil
}

// parseImgString converts the image string to a map[string]string
// Ex:
// image1=path/to/image1:v1,image2=path/to/image2:v1 =>
// {"image1": "path/to/image1:v1", "image2": "path/to/image2:v1"}
func parseImgString(imgString string) (map[string]string, error) {
	res := make(map[string]string)

	pairs := strings.Split(imgString, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid key-value image pair: %s", pair)
		}

		img := kv[0]
		tag := kv[1]

		res[img] = tag
	}

	return res, nil
}
