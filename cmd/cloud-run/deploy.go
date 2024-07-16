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
	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-run/pkg/build"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-run/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-run/pkg/deploy"

	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"google.golang.org/api/run/v1"
)

var opts config.DeployOptions

func NewDeployCmd() *cobra.Command {
	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Create or update a Cloud Run service",
		RunE:  deployService,
	}
	deployCmd.Flags().StringVar(&opts.Image, "image", "", "The image to deployService")
	deployCmd.Flags().StringVar(&opts.Service, "service", "", "The service name to deployService")
	deployCmd.Flags().StringVar(&opts.Source, "source", ".", "The source directory to deployService")

	_ = deployCmd.MarkFlagRequired("service")
	deployCmd.MarkFlagsOneRequired(
		"image",
		"source",
	)
	deployCmd.MarkFlagsMutuallyExclusive(
		"image",
		"source",
	)
	return deployCmd
}

func deployService(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cloudbuildClient, err := cloudbuild.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cloudbuildClient.Close()
	if opts.Image == "" {
		// Source deploy is implemented as a combination of image build
		// and image deploy, same as gcloud.
		opts.Image, err = build.Run(ctx, cloudbuildClient, build.Options{
			ProjectID: projectID,
			Region:    region,
			Service:   opts.Service,
			Source:    opts.Source,
		})
		if err != nil {
			return err
		}
	}

	runService, err := run.NewService(ctx, option.WithUserAgent(userAgent))
	if err != nil {
		return err
	}
	err = deploy.CreateOrUpdateService(runService, projectID, region, opts)
	if err != nil {
		return err
	}
	return deploy.WaitForServiceReady(ctx, runService, projectID, region, opts.Service)
}
