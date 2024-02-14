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
	"testing"

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/client"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/google/go-cmp/cmp"
)

const getDeliveryPipelineRequest = "projects/%s/locations/%s/deliveryPipelines/%s"

func TestFetchReleasePipeline_Success(t *testing.T) {
	// setup
	flags := &config.Config{
		ProjectId:        "id",
		Region:           "global",
		DeliveryPipeline: "test-pipeline",
	}
	expectedUID := "test-uid"

	ctx := context.Background()
	req := &deploypb.GetDeliveryPipelineRequest{
		Name: fmt.Sprintf(getDeliveryPipelineRequest, flags.ProjectId, flags.Region, flags.DeliveryPipeline),
	}
	resp := &deploypb.DeliveryPipeline{
		Uid: expectedUID,
	}

	mockCDClient := new(client.MockCloudDeployClient)
	mockCDClient.On("GetDeliveryPipeline", ctx, req).Return(resp, nil)

	// test
	uid, err := FetchReleasePipeline(ctx, mockCDClient, flags)

	// validate
	if err != nil {
		t.Fatalf("unexpected error calling FetchReleasePipeline: %v", err.Error())
	}
	if diff := cmp.Diff(expectedUID, uid); diff != "" {
		t.Errorf("FetchReleasePipeline UID does not match: %v", diff)
	}
}

func TestFetchReleasePipeline_Failure(t *testing.T) {
	// setup
	flags := &config.Config{
		ProjectId:        "id",
		Region:           "global",
		DeliveryPipeline: "test-pipeline",
	}
	expectedErr := fmt.Errorf("failed to get pipeline")

	ctx := context.Background()
	req := &deploypb.GetDeliveryPipelineRequest{
		Name: fmt.Sprintf(getDeliveryPipelineRequest, flags.ProjectId, flags.Region, flags.DeliveryPipeline),
	}

	mockCDClient := new(client.MockCloudDeployClient)
	mockCDClient.On("GetDeliveryPipeline", ctx, req).Return(nil, fmt.Errorf("failed to get pipeline"))

	// test
	_, err := FetchReleasePipeline(ctx, mockCDClient, flags)

	// validate
	if diff := cmp.Diff(expectedErr.Error(), err.Error()); diff != "" {
		t.Errorf("FetchReleasePipeline UID does not match: %v", diff)
	}
}
