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

var (
	initialRolloutAnnotationStr string
	initialRolloutLabelStr      string
	imagesStr                   string
	userAgent                   string
)

const (
	defaultSource = "."
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a Cloud Deploy Release",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if strings.Contains(flags.DeliveryPipeline, "/") {
			return fmt.Errorf("invalid --delivery-pipeline value: %s, only lower-case letters, numbers, and hyphens are allowed", flags.DeliveryPipeline)
		}

		flags.InitialRolloutAnotations, err = release.ParseDictString(initialRolloutAnnotationStr)
		if err != nil {
			return fmt.Errorf("invalid --initial-rollout-annotations value: %s", err)
		}
		flags.InitialRolloutLabels, err = release.ParseDictString(initialRolloutLabelStr)
		if err != nil {
			return fmt.Errorf("invalid --initial-rollout-labels value: %s", err)
		}
		flags.Images, err = release.ParseDictString(imagesStr)
		if err != nil {
			return fmt.Errorf("invalid --images value: %s", err)
		}
		if flags.Source == "" {
			flags.Source = defaultSource
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
	releaseCmd.PersistentFlags().StringVar(&flags.Description, "description", "", "The description of the release")
	releaseCmd.PersistentFlags().StringVar(&flags.InitialRolloutPhaseId, "initial-rollout-phase-id", "", "The phase to start the initial rollout at when creating the release")
	releaseCmd.PersistentFlags().StringVar(&flags.ToTarget, "to-target", "", "The target to deliver into upon release creation")
	releaseCmd.PersistentFlags().StringVar(&flags.SkaffoldVersion, "skaffold-version", "", "The version of the Skaffold binary")
	releaseCmd.PersistentFlags().StringVar(&flags.SkaffoldFile, "skaffold-file", "", "The path of the skaffold file absolute or relative to the source directory.")

	releaseCmd.PersistentFlags().StringVar(&imagesStr, "images", "", "The images associated with the release")
	releaseCmd.PersistentFlags().StringVar(&initialRolloutAnnotationStr, "initial-rollout-annotations", "", "Annotations to apply to the initial rollout when creating the release")
	releaseCmd.PersistentFlags().StringVar(&initialRolloutLabelStr, "initial-rollout-labels", "", "Labels to apply to the initial rollout when creating the release")

	releaseCmd.PersistentFlags().StringVar(&userAgent, "google-apis-user-agent", "", "The user-agent to be applied when calling Google APIs")

	releaseCmd.MarkPersistentFlagRequired("delivery-pipeline")
	releaseCmd.MarkPersistentFlagRequired("region")
	releaseCmd.MarkPersistentFlagRequired("project-id")
	releaseCmd.MarkPersistentFlagRequired("name")
}
