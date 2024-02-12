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
	"fmt"
	"strings"

	deploy "cloud.google.com/go/deploy/apiv1"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/release"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var flags config.ReleaseConfiguration

const userAgent = "google-gitlab-components:create-cloud-deploy-release"

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a Cloud Deploy Release",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if strings.Contains(flags.DeliveryPipeline, "/") {
			return fmt.Errorf("invalid delivery-pipeline value: %s, only lower-case letters, numbers, and hyphens are allowed", flags.DeliveryPipeline)
		}
		if flags.Images != "" && strings.Contains(flags.Images, " ") {
			return fmt.Errorf("invalid images value: %s, the images string should not contain space", flags.Images)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := context.Background()
		cdClient, err := deploy.NewCloudDeployClient(ctx, option.WithUserAgent(userAgent))
		if err != nil {
			return err
		}
		gcsClient, err := storage.NewClient(ctx, option.WithUserAgent(userAgent))
		if err != nil {
			return err
		}

		if err = release.CreateCloudDeployRelease(ctx, cdClient, gcsClient, &flags); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	// TODO: consider extracing fields from delivery-pipeline
	releaseCmd.PersistentFlags().StringVar(&flags.DeliveryPipeline, "delivery-pipeline", "", "The delivery pipeline associated with the release")
	releaseCmd.PersistentFlags().StringVar(&flags.Region, "region", "", "The cloud region for the release")
	releaseCmd.PersistentFlags().StringVar(&flags.ProjectId, "project-id", "", "The GCP project id")
	releaseCmd.PersistentFlags().StringVar(&flags.Release, "name", "", "The name of the release to create")
	releaseCmd.PersistentFlags().StringVar(&flags.Source, "source", ".", "The source location containing skaffold.yaml")
	releaseCmd.PersistentFlags().StringVar(&flags.Images, "images", "", "The images associated with the release")

	releaseCmd.MarkPersistentFlagRequired("delivery-pipeline")
	releaseCmd.MarkPersistentFlagRequired("region")
	releaseCmd.MarkPersistentFlagRequired("project-id")
	releaseCmd.MarkPersistentFlagRequired("name")
}
