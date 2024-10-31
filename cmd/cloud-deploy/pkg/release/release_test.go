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
	"path/filepath"
	"testing"

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/test"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TODO (@zhangquan): Refactor tests to tests public-facing methods.
// TODO (@zhangquan): Improve tests coverage.
func TestCreateCloudDeployRelease(t *testing.T) {
	// setup
	flags := &config.ReleaseConfiguration{
		ProjectId:        "id",
		Region:           "global",
		DeliveryPipeline: "test-pipeline",
		Source:           "../../test/testdata/testdata.tgz",
		Images: map[string]string{
			"tag1": "image1",
			"tag2": "image2",
		},
	}

	bucket := "test-pipeline-uid_clouddeploy"
	object := "/source/1701112633.689433-75c0b8929cb04d7a8a008a38368ac250.tgz"

	mockObjs := []fakestorage.Object{{
		ObjectAttrs: fakestorage.ObjectAttrs{
			BucketName: bucket,
			Name:       object,
		},
	}}

	// Setup the fake server.
	ctx := context.Background()
	cdClient := test.CreateCloudDeployClient(t, ctx)
	gcsClient := test.CreateGCSClient(t, mockObjs)

	if err := CreateCloudDeployRelease(ctx, cdClient, gcsClient, flags); err != nil {
		t.Fatalf("unexpected error when creating release: %s", err.Error())
	}
}

func TestSetImages_Success(t *testing.T) {
	tcs := []struct {
		name            string
		images          map[string]string
		Release         *deploypb.Release
		ExpectedRelease *deploypb.Release
	}{
		{
			name:    "no image",
			Release: &deploypb.Release{},
			ExpectedRelease: &deploypb.Release{
				BuildArtifacts: []*deploypb.BuildArtifact{},
			},
		},
		{
			name: "2 images",
			images: map[string]string{
				"image1": "tag1",
				"image2": "tag2",
			},
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

func TestParseDictString_Success(t *testing.T) {
	tcs := []struct {
		name         string
		input        string
		expectedDict map[string]string
	}{
		{
			name:  "annotations",
			input: "annotation1=val1,annotation2=val2",
			expectedDict: map[string]string{
				"annotation1": "val1",
				"annotation2": "val2",
			},
		},
		{
			name:  "labels",
			input: "label1=val1,label2=val2",
			expectedDict: map[string]string{
				"label1": "val1",
				"label2": "val2",
			},
		},
	}
	for _, tc := range tcs {
		res, err := ParseDictString(tc.input)
		if err != nil {
			t.Fatalf("unexpected error calling ParseDictString(); %s", err)
		}
		if diff := cmp.Diff(tc.expectedDict, res); diff != "" {
			t.Errorf("unexpected dict calling ParseDictString(): %s", diff)
		}
	}
}

func TestValidateSupportedSkaffoldVersion_Success(t *testing.T) {
	ctx := context.Background()
	flags := &config.ReleaseConfiguration{
		ProjectId: "test-project",
		Region:    "us-central1",
	}

	tcs := []struct {
		name            string
		skaffoldVersion string
	}{
		{
			name:            "valid skaffold version",
			skaffoldVersion: "2.13",
		}, {
			name:            "skaffold version not found",
			skaffoldVersion: "skaffold_preview",
		},
	}

	for _, tc := range tcs {
		flags.SkaffoldVersion = tc.skaffoldVersion
		cdClient := test.CreateCloudDeployClient(t, ctx)

		if err := validateSupportedSkaffoldVersion(ctx, flags, cdClient); err != nil {
			t.Errorf("unexpected error calling validateSupportedSkaffoldVersion(): %s", err)
		}
	}
}

func TestValidateSupportedSkaffoldVersion_Failed(t *testing.T) {
	ctx := context.Background()
	tcs := []struct {
		name            string
		skaffoldVersion string
		expectedErr     error
	}{
		{
			name:            "maintenance period",
			skaffoldVersion: "2.0",
			expectedErr: fmt.Errorf("the Skaffold version you've chosen is no longer supported.\n" +
				"https://cloud.google.com/deploy/docs/using-skaffold/select-skaffold#skaffold_version_deprecation_and_maintenance_policy"),
		},
	}

	flags := &config.ReleaseConfiguration{
		ProjectId: "test-project",
		Region:    "us-central1",
	}
	cdClient := test.CreateCloudDeployClient(t, ctx)

	for _, tc := range tcs {
		flags.SkaffoldVersion = tc.skaffoldVersion
		err := validateSupportedSkaffoldVersion(ctx, flags, cdClient)
		if diff := cmp.Diff(tc.expectedErr.Error(), err.Error()); diff != "" {
			t.Errorf("mismatched error: %s", diff)
		}
	}
}

func TestSetSkaffoldFile_RelativePath_Success(t *testing.T) {
	ctx := context.Background()
	tcs := []struct {
		name                       string
		flags                      *config.ReleaseConfiguration
		expectedSkaffoldConfigPath string
	}{
		{
			name: "directory source with relative skaffold file",
			flags: &config.ReleaseConfiguration{
				Source:       "../../test/test_dir",
				SkaffoldFile: "skaffold-custom.yaml",
			},
			expectedSkaffoldConfigPath: "skaffold-custom.yaml",
		}, {
			name: "directory source with default skaffold file",
			flags: &config.ReleaseConfiguration{
				Source: "../../test/test_dir",
			},
			expectedSkaffoldConfigPath: "",
		}, {
			name: "tar source with relative skaffold file",
			flags: &config.ReleaseConfiguration{
				Source:       "../../test/testdata/testdata.tgz",
				SkaffoldFile: "skaffold.yaml",
			},
			expectedSkaffoldConfigPath: "skaffold.yaml",
		}, {
			name: "tar source with default skaffold file",
			flags: &config.ReleaseConfiguration{
				Source: "../../test/testdata/testdata.tgz",
			},
			expectedSkaffoldConfigPath: "",
		}, {
			name: "remote gcs source with relative skaffold file",
			flags: &config.ReleaseConfiguration{
				Source:       "gs://remote-bucket/obj.tgz",
				SkaffoldFile: "skaffold-custom.yaml",
			},
			expectedSkaffoldConfigPath: "skaffold-custom.yaml",
		}, {
			name: "remote gcs source with reldefaultative skaffold file",
			flags: &config.ReleaseConfiguration{
				Source: "gs://remote-bucket/obj.tgz",
			},
			expectedSkaffoldConfigPath: "",
		},
	}

	for _, tc := range tcs {
		release := &deploypb.Release{}
		if err := setSkaffoldFile(ctx, tc.flags, release); err != nil {
			t.Fatalf("unexpected err calling setSkaffoldFile(): %s", err)
		}
		if diff := cmp.Diff(tc.expectedSkaffoldConfigPath, release.SkaffoldConfigPath); diff != "" {
			t.Errorf("release.SkaffoldConfigPath does not match: %s", diff)
		}
	}
}

func TestSetSkaffoldFile_AbsolutePath_Success(t *testing.T) {
	ctx := context.Background()
	absSkaffoldPath, err := filepath.Abs("../../test/test_dir/skaffold-custom.yaml")
	if err != nil {
		t.Fatalf("unexpected err finding abs path of skaffold file")
	}
	absSourcePath, err := filepath.Abs("../../test/test_dir")
	if err != nil {
		t.Fatalf("unexpected err finding abs path of source dir")
	}

	tcs := []struct {
		name                       string
		flags                      *config.ReleaseConfiguration
		expectedSkaffoldConfigPath string
	}{
		{
			name: "abs source and abs skaffold file",
			flags: &config.ReleaseConfiguration{
				Source:       absSourcePath,
				SkaffoldFile: absSkaffoldPath,
			},
			expectedSkaffoldConfigPath: "skaffold-custom.yaml",
		}, {
			name: "abs source and relative skaffold file",
			flags: &config.ReleaseConfiguration{
				Source:       absSourcePath,
				SkaffoldFile: "../../test/test_dir/skaffold-custom.yaml",
			},
			expectedSkaffoldConfigPath: "../../test/test_dir/skaffold-custom.yaml",
		}, {
			name: "relative source and abs skaffold file",
			flags: &config.ReleaseConfiguration{
				Source:       "../../test/test_dir",
				SkaffoldFile: absSkaffoldPath,
			},
			expectedSkaffoldConfigPath: "skaffold-custom.yaml",
		},
	}

	for _, tc := range tcs {
		release := &deploypb.Release{}
		if err := setSkaffoldFile(ctx, tc.flags, release); err != nil {
			t.Fatalf("unexpected err calling setSkaffoldFile(): %s", err)
		}
		if diff := cmp.Diff(tc.expectedSkaffoldConfigPath, release.SkaffoldConfigPath); diff != "" {
			t.Errorf("release.SkaffoldConfigPath does not match: %s", diff)
		}
	}
}
