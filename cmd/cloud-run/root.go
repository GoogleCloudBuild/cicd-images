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
	"github.com/spf13/cobra"
)

var (
	projectID string
	region    string
	userAgent string

	rootCmd = &cobra.Command{
		Use:   "cloud-run",
		Short: "A CLI tool to deploy and manage Cloud Run services",
		Long: `cloud-run is a command-line tool for deploying and managing Google Cloud Run services.
It provides a streamlined interface for common Cloud Run operations with support for:
- Deploying from container images or source code
- Managing environment variables
- Handling secrets (as environment variables or mounted volumes)
- Updating existing services`,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&projectID, "project-id", "", "The Google Cloud project ID where the service will be deployed")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "The Google Cloud region where the service will be deployed (e.g., us-central1)")
	rootCmd.PersistentFlags().StringVar(&userAgent, "google-apis-user-agent", "", "Custom user agent string for Google API calls")

	_ = rootCmd.MarkPersistentFlagRequired("project-id")
	_ = rootCmd.MarkPersistentFlagRequired("region")

	// Add commands
	rootCmd.AddCommand(NewDeployCmd())
}

func Execute() error {
	return rootCmd.Execute()
}
