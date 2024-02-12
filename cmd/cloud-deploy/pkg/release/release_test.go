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
	"testing"

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCreateCloudDeployRelease(t *testing.T) {
	// setup
	flags := &config.ReleaseConfiguration{
		ProjectId:        "id",
		Region:           "global",
		DeliveryPipeline: "test-pipeline",
		Source:           "../../test/testdata/testdata.tgz",
		Images:           "tag1=image1,tag2=image2",
	}

	bckName := "test-pipeline-uid_clouddeploy"
	stagedObj := "/source/1701112633.689433-75c0b8929cb04d7a8a008a38368ac250.tgz"

	// Setup the fake server.
	ctx := context.Background()
	cdClient := test.CreateCloudDeployClient(t, ctx)
	gcsClient := test.CreateGCSClient(t, []byte{}, bckName, stagedObj)

	if err := CreateCloudDeployRelease(ctx, cdClient, gcsClient, flags); err != nil {
		t.Fatalf("unexpected error when creating release: %s", err.Error())
	}
}

func TestSetImages_Success(t *testing.T) {
	tcs := []struct {
		name            string
		images          string
		Release         *deploypb.Release
		ExpectedRelease *deploypb.Release
	}{
		{
			name:            "no image",
			Release:         &deploypb.Release{},
			ExpectedRelease: &deploypb.Release{},
		},
		{
			name:    "2 images",
			images:  "image1=tag1,image2=tag2",
			Release: &deploypb.Release{},
			ExpectedRelease: &deploypb.Release{
				BuildArtifacts: []*deploypb.BuildArtifact{
					{
						Tag:   "tag1",
						Image: "image1",
					},
					{
						Tag:   "tag2",
						Image: "image2",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		if err := setImages(tc.images, tc.Release); err != nil {
			t.Fatalf("unexpected error calling setImages(); %s", err)
		}

		if diff := cmp.Diff(tc.ExpectedRelease.BuildArtifacts, tc.Release.BuildArtifacts, cmpopts.IgnoreUnexported(deploypb.BuildArtifact{})); diff != "" {
			t.Errorf("unexpected release config calling setImages(): %s", diff)
		}
	}
}
