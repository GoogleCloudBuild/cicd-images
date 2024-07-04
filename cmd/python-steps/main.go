// Copyright 2023 Google LLC
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
	"os"
	"time"

	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/publish"
	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "python-steps"}

	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a Python project.",
		Long:  "Publish is for pushing a Python artifact to a GCP Artifact Registry repository.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			publishArgs, err := publish.ParseArgs(cobraCmd.Flags())
			if err != nil {
				return err
			}
			runner := &helper.DefaultCommandRunner{}
			ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
			defer cf()
			if err := publish.Execute(ctx, runner, publishArgs); err != nil {
				return err
			}
			return nil
		},
	}
	publishCmd.Flags().StringP("artifactRegistryUrl", "", "", "The GCP Artifact Registry URL to push Python artifacts to, e.g. `https://us-docker.pkg.dev/my-project/my-repo`.")
	publishCmd.Flags().StringP("artifactDir", "", "dist", "The local directory in which the packaged artifact is located.")
	publishCmd.Flags().StringP("sourceDistributionResultsPath", "", "", "The path to the source distribution results.")
	publishCmd.Flags().StringP("wheelDistributionResultsPath", "", "", "The path to the wheel distribution results.")
	publishCmd.Flags().StringP("isBuildArtifact", "", "true", "A boolean flag specifying if the results should be a build artifact.")
	publishCmd.Flags().BoolP("verbose", "", false, "Whether to print verbose output.")

	rootCmd.AddCommand(publishCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
