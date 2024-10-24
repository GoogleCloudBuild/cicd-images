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
	"fmt"
	"os"

	"github.com/GoogleCloudBuild/cicd-images/cmd/go-steps/publish"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "go-steps"}

	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a go module to Artifact Registry.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			publishArgs, err := publish.ParseArgs(cmd.Flags())
			if err != nil {
				return err
			}

			if err := publish.Execute(publishArgs); err != nil {
				return err
			}
			return nil
		},
	}
	publishCmd.Flags().String("project", "", "The Google Cloud project ID.")
	publishCmd.Flags().String("repository", "", "Name of Artifact Registry repository to publish the module to.")
	publishCmd.Flags().String("location", "", "The location of the repository.")
	publishCmd.Flags().String("modulePath", "", "The module path of the Go modules. See https://go.dev/ref/mod#module-path.")
	publishCmd.Flags().String("version", "", "The semantic version of the module in the form vX.Y.Z where X is the major version, Y is the minor version, and Z is the patch version. See https://pkg.go.dev/golang.org/x/mod/semver.")

	rootCmd.AddCommand(publishCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error executing 'go-steps': %w", err)
		os.Exit(1)
	}
}
