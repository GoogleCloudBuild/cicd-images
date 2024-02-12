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

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/test"
)

func TestCreateRollout(t *testing.T) {
	// setup
	flags := &config.ReleaseConfiguration{
		ProjectId:        "id",
		Region:           "global",
		DeliveryPipeline: "test-pipeline",
		Images:           "tag1=image1,tag2=image2",
	}

	// Setup the fake server.
	ctx := context.Background()
	cdClient := test.CreateCloudDeployClient(t, ctx)

	if err := CreateRollout(ctx, cdClient, flags); err != nil {
		t.Fatalf("unexpected error when calling CreateRollout: %s", err.Error())
	}
}
