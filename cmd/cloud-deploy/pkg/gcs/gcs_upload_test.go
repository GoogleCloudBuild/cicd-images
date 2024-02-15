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
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/google/uuid"
)

func CreateGCSClient(t *testing.T, content []byte, bucketName, objName string) *storage.Client {
	t.Helper()
	server := fakestorage.NewServer([]fakestorage.Object{{
		Content: content,
		ObjectAttrs: fakestorage.ObjectAttrs{
			BucketName: bucketName,
			Name:       objName,
		},
	},
	})
	t.Cleanup(server.Stop)
	return server.Client()
}

func TestSetSource_TarFile(t *testing.T) {
	// setup
	ctx := context.Background()
	flags := &config.ReleaseConfiguration{
		ProjectId: "id",
		Source:    "./test_data/testdata.tgz",
	}
	pipelineUUID := "test-pipeline-uid"
	release := &deploypb.Release{}

	bucketName, err := getDefaultbucket(pipelineUUID)
	if err != nil {
		t.Fatalf("unexpected error getting bucket name: %s", err)
	}
	stagedObj := fmt.Sprintf("source/%v-%s.tgz", time.Now().UnixMicro(), uuid.New())
	client := CreateGCSClient(t, []byte{}, bucketName, stagedObj)

	// test
	err = SetSource(ctx, pipelineUUID, flags, client, release)
	if err != nil {
		t.Fatalf("unexpected error calling FetchReleasePipeline: %v", err.Error())
	}
}
