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
	"os"
	"time"

	"github.com/GoogleCloudBuild/cicd-images/cmd/maven-steps/publish"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "maven-steps",
	}

	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a maven artifact",
		Long:  "Publish a maven artifact to Artifact Registry & generate maven artifact results.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			publishArgs, err := publish.ParseArgs(cmd.Flags())
			if err != nil {
				return err
			}

			ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
			defer cf()
			if err := publish.Execute(ctx, publishArgs); err != nil {
				return err
			}

			return nil
		},
	}
	publishCmd.Flags().String("repository", "", "URL of the Artifact Registry repository.")
	publishCmd.Flags().String("artifactPath", "", "The path to the packaged file.")
	publishCmd.Flags().String("artifactId", "", "The name of the package file created from the build step.")
	publishCmd.Flags().String("groupId", "", "ID to uniquely identify the project across all Maven projects.")
	publishCmd.Flags().String("version", "", "The version for the application.")
	publishCmd.Flags().Bool("verbose", false, "Whether to print verbose output.")
	publishCmd.Flags().String("isBuildArtifact", "true", "If the results should be a build artifact.")
	publishCmd.Flags().String("resultsPath", "", "Path to write the results in.")

	rootCmd.AddCommand(publishCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
