// Copyright 2024 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"

	"cloud.google.com/go/compute/metadata"
	helper "github.com/GoogleCloudBuild/cicd-images/cmd/maven-steps/internal"
	"github.com/spf13/cobra"
)

var (
	repositoryIds []string
	settingsPath  string
)

const (
	localRepository = "~/.m2/repository"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Authenticate and setup settings file.",
	Long:  "Retrieve authentication token from GKE metadata server, and configure settings.xml",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cmd.SilenceUsage = true

		c := metadata.NewClient(&http.Client{})

		token, err := helper.GetAuthenticationToken(c, ctx)
		if err != nil {
			return fmt.Errorf("error getting authentication token: %w", err)
		}

		if err = helper.WriteSettingsXML(token, localRepository, repositoryIds, settingsPath); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().StringSliceVar(&repositoryIds, "repositoryIds", []string{}, "Repository IDs for which to setup authentication to Artifact Registry.")
	setupCmd.Flags().StringVar(&settingsPath, "settingsPath", "", "Path to store generated settings.xml file.")
	setupCmd.MarkFlagRequired("settingsPath")

}
