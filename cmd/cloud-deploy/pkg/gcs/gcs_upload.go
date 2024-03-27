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
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/google/uuid"
)

const (
	tmpTarPath    = "/tmp/cloud-deploy-tmp-tar.tgz"
	gitIgnoreFile = ".gitignore"
	gitFile       = ".git"
)

// SetSource sets the source for the release config and creates a default Cloud Storage bucket for staging
func SetSource(ctx context.Context, pipelineUUID string, flags *config.ReleaseConfiguration, client *storage.Client, release *deploypb.Release) error {
	source := flags.Source
	object := fmt.Sprintf("source/%v-%s", time.Now().UnixMicro(), uuid.New())
	skaffoldConfigUri := ""
	bucket, err := getDefaultbucket(pipelineUUID)
	if err != nil {
		return err
	}

	if strings.HasPrefix(source, "gs://") {
		// remote gcs source
		object += filepath.Ext(source)
		skaffoldConfigUri, err = copyRemoteGCS(ctx, source, bucket, object, flags.ProjectId, client)
		if err != nil {
			return err
		}
	} else {
		// local source
		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("local source: %s does not exist", source)
		}
		if !info.Mode().IsDir() && !strings.HasSuffix(source, ".zip") && !strings.HasSuffix(source, ".tgz") && !strings.HasSuffix(source, ".gz") {
			return fmt.Errorf("local source: %s is none of local .zip, .tgz, .gz, or directory", source)
		}

		// upload
		if info.Mode().IsDir() {
			// create a tarball if source is a local directory
			if err := createTarball(source, tmpTarPath); err != nil {
				return fmt.Errorf("failed to create local tar: %w", err)
			}
			object += ".tgz"
			source = tmpTarPath
		} else {
			object += filepath.Ext(source)
		}

		skaffoldConfigUri, err = uploadLocalArchiveToGCS(ctx, source, bucket, object, flags.ProjectId, client)
		if err != nil {
			return err
		}
	}

	release.SkaffoldConfigUri = skaffoldConfigUri
	return nil
}

func uploadLocalArchiveToGCS(ctx context.Context, source, bucket, object, projectId string, client *storage.Client) (string, error) {
	if err := createBucketIfNotExist(ctx, bucket, projectId, client); err != nil {
		return "", err
	}

	fmt.Printf("Uploading local archive file %s to gs://%s/%s \n", source, bucket, object)
	if err := uploadTarball(ctx, client, object, bucket, source); err != nil {
		return "", err
	}

	return fmt.Sprintf("gs://%s/%s", bucket, object), nil
}

func uploadTarball(ctx context.Context, gcsClient *storage.Client, object, bucket, fileToUpload string) error {
	data, err := os.ReadFile(fileToUpload)
	if err != nil {
		return fmt.Errorf("unable to read file to upload: %w", err)
	}

	o := gcsClient.Bucket(bucket).Object(object)
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

func createTarball(folderToTarball, tarballPath string) error {
	// Create a temp tarball file
	tarballFile, err := os.Create(tarballPath)
	if err != nil {
		return err
	}
	defer tarballFile.Close()

	// create a tar writer
	gzipWriter := gzip.NewWriter(tarballFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(folderToTarball, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(folderToTarball, path)
		if err != nil {
			return err
		}

		// skip uploading .git and .gitignore files by default
		// this is the default behavior of `gcloud deploy release create ...`
		if relPath == gitIgnoreFile || strings.HasPrefix(relPath, gitFile) {
			return nil
		}
		header.Name = relPath

		// write the header to the tarball
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func copyRemoteGCS(ctx context.Context, source, destBucket, destObj, projectId string, client *storage.Client) (string, error) {
	// parse source
	srcBucket, srcObj, err := parseRemoteGCSSource(source)
	if err != nil {
		return "", err
	}

	if err := createBucketIfNotExist(ctx, destBucket, projectId, client); err != nil {
		return "", err
	}

	// Copy content
	dstUri := fmt.Sprintf("gs://%s/%s", destBucket, destObj)
	fmt.Printf("Copying remote storage %s to %s \n", source, dstUri)

	src := client.Bucket(srcBucket).Object(srcObj)
	dst := client.Bucket(destBucket).Object(destObj)
	_, err = dst.CopierFrom(src).Run(ctx)
	if err != nil {
		return "", fmt.Errorf("error running gcs copier: %v", err)
	}

	return dstUri, nil
}

// parseRemoteGCSSource parse an input remote source to bucket name and object name, or error.
// e.g. input: gs://bucket_name/object_name.xyz
// output: bucket_name, object_name.xyz, nil
func parseRemoteGCSSource(source string) (string, string, error) {
	source = strings.TrimPrefix(source, "gs://")
	parts := strings.SplitN(source, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid remote GCS source format: %s", source)
	}

	return parts[0], parts[1], nil
}

func createBucketIfNotExist(ctx context.Context, bucketName, projectId string, client *storage.Client) error {
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
		return fmt.Errorf("bucket(%q).Update: %w", bucketName, err)
	}

	return nil
}

func getDefaultbucket(pipelineUUID string) (string, error) {
	bucketName := pipelineUUID + "_clouddeploy"

	// Bucket name length constraint
	if len(bucketName) > 63 {
		return "", fmt.Errorf("the length of the bucket id: %s must not exceed 63 characters", bucketName)
	}

	return bucketName, nil
}
