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
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/install"
	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/command"
	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/pack"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "python-steps"}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install Python dependencies.",
		Long: `install is for installing Python dependencies.
It can install Python dependencies from a requirements.txt file or
from a space-separated list of dependencies. It can also pull Python
artifacts from a GCP Artifact Registry.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			runArgs, err := install.ParseArgs(cobraCmd.Flags())
			if err != nil {
				return err
			}
			runner := &command.DefaultCommandRunner{}
			client := metadata.NewClient(&http.Client{})
			ctx := context.Background()
			if err := install.Execute(ctx, runner, runArgs, client); err != nil {
				return err
			}
			return nil
		},
	}
	installCmd.Flags().StringP("dependencies", "", "", "The Python dependencies to install split on whitespace, e.g. `flake8 pytest`.")
	installCmd.Flags().StringP("requirementsPath", "", "", "Path to the `requirements.txt` file, e.g. `./requirements.txt`.")
	installCmd.Flags().StringP("artifactRegistryUrl", "", "", "The GCP Artifact Registry URL to pull Python artifacts from, e.g. `https://us-docker.pkg.dev/my-project/my-repo`.")
	installCmd.Flags().BoolP("verbose", "", false, "Whether to print verbose output.")
	installCmd.Flags().StringP("script", "", "", "The Python script to run inline, e.g. `print('hello world')`.")

	packCmd := &cobra.Command{
		Use:   "package",
		Short: "Package a Python project.",
		Long: `package is for packaging and pushing
a Python project to a GCP Artifact Registry repository.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			packageArgs, err := pack.ParseArgs(cobraCmd.Flags())
			if err != nil {
				return err
			}
			runner := &command.DefaultCommandRunner{}
			client := metadata.NewClient(&http.Client{})
			ctx := context.Background()
			if err := pack.Execute(ctx, runner, packageArgs, client); err != nil {
				return err
			}
			return nil
		},
	}
	packCmd.Flags().StringP("command", "", "setup.py sdist bdist_wheel", "The Python command to package the Python project, e.g. `setup.py sdist bdist_wheel`.")
	packCmd.Flags().StringP("artifactRegistryUrl", "", "", "The GCP Artifact Registry URL to push Python artifacts to, e.g. `https://us-docker.pkg.dev/my-project/my-repo`.")
	packCmd.MarkFlagRequired("artifactRegistryUrl")
	packCmd.Flags().StringP("sourceDistributionResultsPath", "", "", "The path to the source distribution results.")
	packCmd.Flags().StringP("wheelDistributionResultsPath", "", "", "The path to the wheel distribution results.")
	packCmd.Flags().BoolP("verbose", "", false, "Whether to print verbose output.")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(packCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
