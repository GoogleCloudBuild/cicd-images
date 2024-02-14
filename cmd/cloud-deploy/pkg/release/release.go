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

	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/client"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
)

// CreateCloudDeployRelease is the main entry to create a Release
func CreateCloudDeployRelease(ctx context.Context, cdClient client.ICloudDeployClient, flags *config.Config) error {
	// TODO: Add implementation
	_, err := FetchReleasePipeline(ctx, cdClient, flags)
	return err
}

// FetchReleasePipeline calls Cloud Deploy API to get the target Delivery Pipeline.
// It returns the ID of the Delivery Pipeline is found, return error otherwise.
func FetchReleasePipeline(ctx context.Context, cdClient client.ICloudDeployClient, flags *config.Config) (string, error) {
	req := &deploypb.GetDeliveryPipelineRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline),
	}
	dp, err := cdClient.GetDeliveryPipeline(ctx, req)
	if err != nil {
		return "", err
	}

	return dp.Uid, nil
}
