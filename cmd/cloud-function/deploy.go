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

package main

import (
	functions "cloud.google.com/go/functions/apiv2"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/pkg/deploy"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/utils"
	"google.golang.org/api/option"

	"github.com/spf13/cobra"
)

var opts config.DeployOptions

func NewDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Create or update a Cloud Function",
		RunE:  deployFunction,
	}
	deployCmd.Flags().StringVar(&opts.Name, "name", "", "The name of the function")
	deployCmd.Flags().StringVar(&opts.Description, "description", "", "The description of the function")
	deployCmd.Flags().StringToStringVar(&opts.Labels, "labels", nil, "List of key-value pairs to set as function labels in the form label1=VALUE1,label2=VALUE2")
	deployCmd.Flags().StringVar(&opts.Runtime, "runtime", "", "The runtime in which to run the function")
	deployCmd.Flags().StringVar(&opts.EntryPoint, "entry-point", "", "The name of a Google Cloud Function (as defined in source code) that will be executed")
	deployCmd.Flags().StringVar(&opts.Source, "source", "", "Location of source code to deploy. Can be one Google Cloud Storage, Google Source Repository or a Local filesystem path")

	_ = deployCmd.MarkFlagRequired("name")
	_ = deployCmd.MarkFlagRequired("entry-point")
	_ = deployCmd.MarkFlagRequired("runtime")
	_ = deployCmd.MarkFlagRequired("source")

	return deployCmd
}

func deployFunction(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	client, err := functions.NewFunctionClient(ctx, option.WithUserAgent(userAgent))
	if err != nil {
		return err
	}
	defer client.Close()
	opts.ProjectID = projectID
	opts.Region = region
	return deploy.CreateOrUpdateFunction(ctx, client, utils.OsFileUtil{}, opts)
}
