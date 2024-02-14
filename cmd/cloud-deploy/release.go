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
	"context"

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/client"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/release"
	"github.com/spf13/cobra"
)

var flags config.Config

const userAgent = "google-gitlab-components:create-cloud-deploy-release"

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a Cloud Deploy Release",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := context.Background()
		cdClient, err := client.NewCloudDeployClient(ctx, userAgent)
		if err != nil {
			return err
		}
		if err = release.CreateCloudDeployRelease(ctx, cdClient, &flags); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.PersistentFlags().StringVar(&flags.DeliveryPipeline, "delivery-pipeline", "", "The delivery pipeline associated with the release")
	releaseCmd.PersistentFlags().StringVar(&flags.Region, "region", "", "The cloud region for the release")
	releaseCmd.PersistentFlags().StringVar(&flags.ProjectId, "project-id", "", "The GCP project id")
}
