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

package client

import (
	"context"
	"fmt"

	deploy "cloud.google.com/go/deploy/apiv1"
	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

type ICloudDeployClient interface {
	GetDeliveryPipeline(ctx context.Context, req *deploypb.GetDeliveryPipelineRequest, options ...gax.CallOption) (*deploypb.DeliveryPipeline, error)
}

type CloudDeployClient struct {
	*deploy.CloudDeployClient
}

// NewCloudDeployClient returns Google Cloud DeployClient with specified user agent or an error if failed
func NewCloudDeployClient(ctx context.Context, userAgent string) (ICloudDeployClient, error) {
	c, err := deploy.NewCloudDeployClient(ctx, option.WithUserAgent(userAgent))
	if err != nil {
		return nil, fmt.Errorf("CloudDeployClient init error: %w", err)
	}
	return &CloudDeployClient{CloudDeployClient: c}, err
}
