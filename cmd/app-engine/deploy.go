// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/deploy"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/version"

	appengine "cloud.google.com/go/appengine/apiv1"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var opts config.AppEngineDeployOptions

func newDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new version to App Engine",
		RunE:  runDeploy,
	}

	cmd.Flags().StringVarP(&opts.AppYAMLPath, "app-yaml", "a", "app.yaml", "Path to the app.yaml file")
	cmd.Flags().StringVarP(&opts.ImageURL, "image", "i", "", "URL of the Artifact Registry container image to deploy (optional)")
	cmd.Flags().BoolVar(&opts.Promote, "promote", true, "Promote the deployed version to receive 100% traffic")
	cmd.Flags().StringVarP(&opts.VersionID, "version", "v", version.ID(), "Version ID (leave empty to auto-generate)")
	cmd.Flags().StringVarP(&opts.Bucket, "bucket", "b", "", "Cloud Storage Bucket to upload source zip file (optional, ignored if deploying a container image)")

	return cmd
}

func runDeploy(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	versionClient, err := appengine.NewVersionsClient(ctx, option.WithUserAgent(userAgent))
	if err != nil {
		return fmt.Errorf("failed to create versions client: %w", err)
	}
	defer versionClient.Close()

	var servicesClient *appengine.ServicesClient
	if opts.Promote {
		servicesClient, err = appengine.NewServicesClient(ctx, option.WithUserAgent(userAgent))
		if err != nil {
			return fmt.Errorf("error creating services client: %w", err)
		}
		defer servicesClient.Close()
	}

	var storageClient *storage.Client
	// If not deploying an image, then upload source
	if opts.ImageURL == "" {
		storageClient, err = storage.NewClient(ctx, option.WithUserAgent(userAgent))
		if err != nil {
			return fmt.Errorf("error creating storage client: %w", err)
		}
		defer storageClient.Close()
	}

	return deploy.CreateAndDeployVersion(ctx, versionClient, servicesClient, storageClient, opts)
}
