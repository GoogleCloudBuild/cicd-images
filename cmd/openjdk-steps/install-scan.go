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

	"github.com/GoogleCloudBuild/cicd-images/cmd/openjdk-steps/internal/installscanlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const INSTALLLOGPATH_ARG = "logPath"
const INSTALLOUTPUTPATH_ARG = "outputPath"
const defaultLogPath = "/mavenconfigs/install-output.txt"
const defaultOutputPath = "/mavenconfigs/artifact-directory-paths.txt"

var installScanCmd = &cobra.Command{
	Use:   "install-scan",
	Short: "Scan the install logs to retrieve relative artifact output paths.",
	RunE: func(cmd *cobra.Command, args []string) error {
		installScanCmdArgs, err := parseInstallScanArgs(cmd.Flags())
		if err != nil {
			return err
		}
		cmd.SilenceUsage = true

		return installscanlog.ScanInstallLog(installScanCmdArgs.logPath, installScanCmdArgs.outputPath)
	},
}

func init() {
	rootCmd.AddCommand(installScanCmd)

	installScanCmd.Flags().StringP(INSTALLLOGPATH_ARG, "", defaultLogPath, "Path of output log of build.")
	installScanCmd.Flags().StringP(INSTALLOUTPUTPATH_ARG, "", defaultOutputPath, "Path of retreived list of artifact directories.")
}

type installScanArguments struct {
	logPath    string
	outputPath string
}

func parseInstallScanArgs(f *pflag.FlagSet) (installScanArguments, error) {
	logPath, err := f.GetString(INSTALLLOGPATH_ARG)
	if err != nil {
		return installScanArguments{}, fmt.Errorf("failed to get logPath argument: %w", err)
	}

	outputPath, err := f.GetString(INSTALLOUTPUTPATH_ARG)
	if err != nil {
		return installScanArguments{}, fmt.Errorf("failed to get outputPath argument: %w", err)
	}

	return installScanArguments{
		logPath:    logPath,
		outputPath: outputPath,
	}, nil
}
