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

package gcs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/client"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/google/uuid"
)

// SetSource the source for the release config and creates a default Cloud Storage bucket for staging
func SetSource(ctx context.Context, pipelineUUID string, flags *config.ReleaseConfiguration, client client.IGCSClient, release *deploypb.Release) error {
	source := flags.Source
	if err := validateSource(source); err != nil {
		return err
	}

	bucket, err := getDefaultbucket(pipelineUUID)
	if err != nil {
		return err
	}
	if err := createBucketIfNotExist(ctx, bucket, flags.ProjectId, client); err != nil {
		return err
	}

	// write to GCS
	// TODO: support other source type upload
	object := fmt.Sprintf("source/%v-%s%s", time.Now().UnixMicro(), uuid.New(), filepath.Ext(source))
	if err := setLocalArchiveSource(ctx, source, bucket, object, client, release); err != nil {
		return err
	}

	// TODO: add skaffold_version
	return nil
}

func setLocalArchiveSource(ctx context.Context, source, bucket, object string, client client.IGCSClient, release *deploypb.Release) error {
	fmt.Printf("Uploading local archive file %s to gs://%s/%s \n", source, bucket, object)
	if err := uploadTarball(ctx, client, object, bucket, source); err != nil {
		return err
	}
	release.SkaffoldConfigPath = fmt.Sprintf("gs://%s/%s", bucket, object)

	return nil
}

func getDefaultbucket(pipelineUUID string) (string, error) {
	bucketName := pipelineUUID + "_clouddeploy"

	// Bucket name length constraint
	if len(bucketName) > 63 {
		return "", fmt.Errorf("The length of the bucket id: %s must not exceed 63 characters", bucketName)
	}

	return bucketName, nil
}

func createBucketIfNotExist(ctx context.Context, bucketName, projectId string, client client.IGCSClient) error {
	bktHandler := client.Bucket(bucketName)
	// TODO: optimize to "Get-then-Insert" mechanism
	if err := bktHandler.Create(ctx, projectId, nil); err != nil {
		fmt.Printf("GCS bucket already exist, skipping creation... \n")
	}

	// enable uniform level access
	enableUniformBucketLevelAccess := storage.BucketAttrsToUpdate{
		UniformBucketLevelAccess: &storage.UniformBucketLevelAccess{
			Enabled: true,
		},
	}
	if _, err := bktHandler.Update(ctx, enableUniformBucketLevelAccess); err != nil {
		return fmt.Errorf("Bucket(%q).Update: %w", bucketName, err)
	}

	return nil
}

func uploadTarball(ctx context.Context, gcsClient client.IGCSClient, stagedObj, bucketName, fileToUpload string) error {
	data, err := os.ReadFile(fileToUpload)
	if err != nil {
		return fmt.Errorf("unable to read file to upload: %w", err)
	}

	o := gcsClient.Bucket(bucketName).Object(stagedObj)
	wc := o.NewWriter(ctx)
	wc.ContentType = "application/x-tar"

	if _, err := wc.Write(data); err != nil {
		return fmt.Errorf("unable to write data to bucket %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("unable to close bucket writer %w", err)
	}

	return err
}

// TODO: allow cloud storage and directory source input
func validateSource(source string) error {
	if strings.HasPrefix(source, "gs://") {
		return fmt.Errorf("google cloud storage source is not supported")
	}

	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return fmt.Errorf("source: %s does not exist", source)
	}

	// TODO: check if the souce is an Archive in addition to checking the suffix
	if !info.Mode().IsRegular() || (!strings.HasSuffix(source, ".zip") && !strings.HasSuffix(source, ".tgz") && strings.HasSuffix(source, ".gz")) {
		return fmt.Errorf("source: %s is none of .zip, .tgz, .gz (directory source is not implemented yet)", source)
	}

	return nil
}
