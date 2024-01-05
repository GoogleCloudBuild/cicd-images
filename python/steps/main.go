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
	"fmt"
	"net/http"
	"os"

	"github.com/GoogleCloudBuild/python-images/steps/internal/cmd"
	"github.com/GoogleCloudBuild/python-images/steps/pack"
	"github.com/GoogleCloudBuild/python-images/steps/run"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "python-steps"}

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run a Python command.",
		Long: `run is for running a Python command.
It can install Python dependencies from a requirements.txt file or
from a space-separated list of dependencies. It can also pull Python
artifacts from a GCP Artifact Registry.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			runArgs, err := run.ParseArgs(cobraCmd.Flags())
			if err != nil {
				return err
			}
			runner := &cmd.DefaultCommandRunner{}
			client := &http.Client{}
			if err := run.Execute(runner, runArgs, client); err != nil {
				return err
			}
			return nil
		},
	}
	runCmd.Flags().StringP("command", "", "", "The Python command to run, e.g. `main.py` or `-m pytest`.")
	runCmd.Flags().StringP("dependencies", "", "", "The Python dependencies to install split on whitespace, e.g. `flake8 pytest`.")
	runCmd.Flags().StringP("requirementsPath", "", "", "Path to the `requirements.txt` file, e.g. `./requirements.txt`.")
	runCmd.Flags().StringP("artifactRegistryUrl", "", "", "The GCP Artifact Registry URL to pull Python artifacts from, e.g. `https://us-docker.pkg.dev/my-project/my-repo`.")
	runCmd.Flags().BoolP("verbose", "", false, "Whether to print verbose output.")
	runCmd.Flags().StringP("script", "", "", "The Python script to run inline, e.g. `print('hello world')`.")

	var packCmd = &cobra.Command{
		Use:   "package",
		Short: "Package a Python project.",
		Long: `package is for packaging and pushing
a Python project to a GCP Artifact Registry repository.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			packageArgs, err := pack.ParseArgs(cobraCmd.Flags())
			if err != nil {
				return err
			}
			runner := &cmd.DefaultCommandRunner{}
			client := &http.Client{}
			if err := pack.Execute(runner, packageArgs, client); err != nil {
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

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(packCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
