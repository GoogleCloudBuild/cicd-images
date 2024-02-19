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
	"testing"
	"time"

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/test"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/google/uuid"
)

func TestSetSource_LocalTarFile(t *testing.T) {
	// setup
	ctx := context.Background()
	flags := &config.ReleaseConfiguration{
		ProjectId: "id",
		Source:    "../../test/testdata/testdata.tgz",
	}
	pipelineUUID := "test-pipeline-uid"
	release := &deploypb.Release{}

	bucket, err := getDefaultbucket(pipelineUUID)
	if err != nil {
		t.Fatalf("unexpected error getting bucket name: %s", err)
	}
	object := fmt.Sprintf("source/%v-%s.tgz", time.Now().UnixMicro(), uuid.New())
	mockObjs := []fakestorage.Object{{
		ObjectAttrs: fakestorage.ObjectAttrs{
			BucketName: bucket,
			Name:       object,
		},
	}}
	client := test.CreateGCSClient(t, mockObjs)

	// test
	err = SetSource(ctx, pipelineUUID, flags, client, release)
	if err != nil {
		t.Fatalf("unexpected error calling FetchReleasePipeline: %v", err.Error())
	}
}

func TestSetSource_LocalDirectory(t *testing.T) {
	// setup
	ctx := context.Background()
	flags := &config.ReleaseConfiguration{
		ProjectId: "id",
		Source:    "../../test/test_dir/",
	}
	pipelineUUID := "test-pipeline-uid"
	release := &deploypb.Release{}

	bucket, err := getDefaultbucket(pipelineUUID)
	if err != nil {
		t.Fatalf("unexpected error getting bucket name: %s", err)
	}
	object := fmt.Sprintf("source/%v-%s.tgz", time.Now().UnixMicro(), uuid.New())
	mockObjs := []fakestorage.Object{{
		ObjectAttrs: fakestorage.ObjectAttrs{
			BucketName: bucket,
			Name:       object,
		},
	}}
	client := test.CreateGCSClient(t, mockObjs)

	// test
	err = SetSource(ctx, pipelineUUID, flags, client, release)
	if err != nil {
		t.Fatalf("unexpected error calling FetchReleasePipeline: %v", err.Error())
	}
}

func TestSetSource_RemoteGCS(t *testing.T) {
	// setup
	ctx := context.Background()
	flags := &config.ReleaseConfiguration{
		ProjectId: "id",
		Source:    "gs://src_bucket/src_obj.zip",
	}
	pipelineUUID := "test-pipeline-uid"
	release := &deploypb.Release{}
	dstBucket, err := getDefaultbucket(pipelineUUID)
	if err != nil {
		t.Fatalf("unexpected error getting bucket name: %s", err)
	}
	destObject := fmt.Sprintf("source/%v-%s.tgz", time.Now().UnixMicro(), uuid.New())

	mockObjs := []fakestorage.Object{{
		ObjectAttrs: fakestorage.ObjectAttrs{
			BucketName: dstBucket,
			Name:       destObject,
		},
	}, {
		ObjectAttrs: fakestorage.ObjectAttrs{
			BucketName: "src_bucket",
			Name:       "src_obj.zip",
		},
	}}
	client := test.CreateGCSClient(t, mockObjs)

	// test
	err = SetSource(ctx, pipelineUUID, flags, client, release)
	if err != nil {
		t.Fatalf("unexpected error calling FetchReleasePipeline: %v", err.Error())
	}
}
