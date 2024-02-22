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

package rollout

import (
	"context"
	"testing"

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/test"
	"github.com/google/go-cmp/cmp"
)

// TODO (@zhangquan): Refactor tests to tests public-facing methods
// TODO (@zhangquan): Improve tests coverage
func TestCreateRollout(t *testing.T) {
	// setup
	flags := &config.ReleaseConfiguration{
		ProjectId:        "id",
		Region:           "global",
		DeliveryPipeline: "test-pipeline",
		ToTarget:         "test-target",
		Images: map[string]string{
			"image1": "tags1",
			"image2": "tags2",
		},
		InitialRolloutPhaseId: "stable",
		InitialRolloutAnotations: map[string]string{
			"annotation1": "val1",
			"annotation2": "val2",
		},
		InitialRolloutLabels: map[string]string{
			"label1": "val1",
			"label2": "val2",
		},
	}

	// Setup the fake server.
	ctx := context.Background()
	cdClient := test.CreateCloudDeployClient(t, ctx)

	if err := CreateRollout(ctx, cdClient, flags); err != nil {
		t.Fatalf("unexpected error when calling CreateRollout: %s", err.Error())
	}
}

func TestGetToTargetId(t *testing.T) {
	release := &deploypb.Release{
		DeliveryPipelineSnapshot: &deploypb.DeliveryPipeline{
			Pipeline: &deploypb.DeliveryPipeline_SerialPipeline{
				SerialPipeline: &deploypb.SerialPipeline{
					Stages: []*deploypb.Stage{
						{
							TargetId: "test-id",
						},
						{
							TargetId: "staging-id",
						},
					},
				},
			},
		},
	}
	tcs := []struct {
		name             string
		flags            *config.ReleaseConfiguration
		expectedTargetId string
	}{
		{
			name: "user-provided target",
			flags: &config.ReleaseConfiguration{
				ToTarget: "staging-id",
				Release:  "test-release",
			},
			expectedTargetId: "staging-id",
		}, {
			name: "user-provided target",
			flags: &config.ReleaseConfiguration{
				Release: "test-release",
			},
			expectedTargetId: "test-id",
		},
	}

	for _, tc := range tcs {
		targetId, err := getToTargetId(context.Background(), release, tc.flags)
		if err != nil {
			t.Fatalf("unexpected error calling getToTargetId(): %s", err)
		}
		if diff := cmp.Diff(tc.expectedTargetId, targetId); diff != "" {
			t.Errorf("mismatched target id when calling getToTargetId(): %s", diff)
		}
	}
}
