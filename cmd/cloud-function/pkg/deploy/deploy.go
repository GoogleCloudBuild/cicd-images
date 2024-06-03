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

package deploy

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/utils"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
)

const (
	noLabelsStartingWithDeployMessage = "label keys starting with 'deployment' are reserved for use by deployment tools and cannot be specified manually"
)

// CreateOrUpdateFunction deploys a Cloud Functions. If the function
// doesn't exist, it creates a new one. If the function exists, it updates the
// existing function with the config.DeployOptions.
func CreateOrUpdateFunction(ctx context.Context, client *functions.FunctionClient, fileUtil utils.FileUtil, opts config.DeployOptions) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", opts.ProjectID, opts.Region)
	resourceName := fmt.Sprintf("%s/functions/%s", parent, opts.Name)
	function, err := buildFunctionDefinition(ctx, client, fileUtil, opts, resourceName)
	if err != nil {
		return err
	}
	log.Printf("Deploying a Cloud function [%s] in project [%s] region [%s]\n", opts.Name, opts.ProjectID, opts.Region)
	getRequest := &functionspb.GetFunctionRequest{
		Name: resourceName,
	}
	_, err = client.GetFunction(ctx, getRequest)
	if err != nil {
		req := &functionspb.CreateFunctionRequest{
			Parent:     parent,
			FunctionId: opts.Name,
			Function:   function,
		}
		createOperation, err := client.CreateFunction(ctx, req)
		if err != nil {
			return err
		}

		_, err = createOperation.Wait(ctx)
		if err != nil {
			return err
		}
		log.Printf("Cloud function [%s] created successfully\n", resourceName)
	} else {
		req := &functionspb.UpdateFunctionRequest{
			Function: function,
		}
		updateOperation, err := client.UpdateFunction(ctx, req)
		if err != nil {
			return err
		}

		_, err = updateOperation.Wait(ctx)
		if err != nil {
			return err
		}
		log.Printf("Cloud function [%s] updated successfully\n", resourceName)
	}

	return nil
}

// buildFunctionDefinition creates a new Cloud Function object based on the
// provided DeployOptions.
func buildFunctionDefinition(ctx context.Context, client *functions.FunctionClient, fileUtil utils.FileUtil, opts config.DeployOptions, resourceName string) (*functionspb.Function, error) {
	err := verifyNoDeploymentLabels(opts)
	if err != nil {
		return nil, err
	}
	source, err := utils.BuildSourceDefinition(ctx, client, fileUtil, opts)
	if err != nil {
		return nil, err
	}
	return &functionspb.Function{
		Name:        resourceName,
		Description: opts.Description,
		BuildConfig: &functionspb.BuildConfig{
			Runtime:    opts.Runtime,
			EntryPoint: opts.EntryPoint,
			Source:     source,
		},
		Labels:      opts.Labels,
		Environment: functionspb.Environment_GEN_1,
	}, nil
}

// verifyNoDeploymentLabels checks that there are no reserved labels used in DeployOptions.Labels.
func verifyNoDeploymentLabels(opts config.DeployOptions) error {
	if len(opts.Labels) != 0 {
		for k := range opts.Labels {
			if strings.HasPrefix(k, "deployment") {
				return fmt.Errorf(noLabelsStartingWithDeployMessage)
			}
		}
	}
	return nil
}
