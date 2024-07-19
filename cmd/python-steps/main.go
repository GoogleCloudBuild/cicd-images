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
		RunE: func(cobraCmd *cobra.Command, _ []string) error {
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
	publishCmd.Flags().String("repository", "", "The GCP Artifact Registry repository URL to push Python artifacts to, e.g. `https://us-docker.pkg.dev/my-project/my-repo`.")
	publishCmd.Flags().StringSlice("paths", []string{}, "The paths of the python package distributions to publish to Artifact Registry. Can include a source distribution and a wheel distribution. Ex. ['dist/my-pkg.tar.gz','dist/my-pkg.whl']")
	publishCmd.Flags().String("sourceDistributionResultsPath", "", "The path to the source distribution results.")
	publishCmd.Flags().String("wheelDistributionResultsPath", "", "The path to the wheel distribution results.")
	publishCmd.Flags().String("isBuildArtifact", "true", "The uploaded Python artifacts are reported as build artifact in provenance.")
	publishCmd.Flags().Bool("verbose", false, "Whether to print verbose output.")

	rootCmd.AddCommand(publishCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error executing 'python-steps': %w", err)
		os.Exit(1)
	}
}
