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

	"github.com/GoogleCloudBuild/cicd-images/cmd/openjdk-steps/internal/deployscanlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const DEPLOYLOGPATH_ARG 	= "logPath"
const DEPLOYOUTPUTPATH_ARG 	= "outputPath"
const logPathDefault        = "./deploy.txt"
const outputPathDefault     = "/mavenconfigs/artifacts.list"

var deployScanCmd = &cobra.Command {
	Use: "deploy-scan",
	Short: "Scan the deploy logs to retrieve artifacts.",
	RunE: func(cmd *cobra.Command, args []string) error {
		deployScanCmdArgs, err := parseDeployScanArgs(cmd.Flags())
		if err != nil{
			return err
		}
		cmd.SilenceUsage = true

		err = deployscanlog.ScanDeployLog(deployScanCmdArgs.logPath, deployScanCmdArgs.outputPath)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployScanCmd)

	deployScanCmd.Flags().StringP(DEPLOYLOGPATH_ARG, "", defaultLogPath, "Path of output log of deploy plugin")
	deployScanCmd.Flags().StringP(DEPLOYOUTPUTPATH_ARG, "", defaultOutputPath, "Path of retrieved artifacts output")
}

type deployScanArguments struct {
	logPath 	string
	outputPath 	string
}

func parseDeployScanArgs(f *pflag.FlagSet) (deployScanArguments, error) {
	logPath, err := f.GetString(DEPLOYLOGPATH_ARG)
	if err != nil {
		return deployScanArguments{}, fmt.Errorf("failed to get logPath argument: %v", err)
	}

	outputPath, err := f.GetString(DEPLOYOUTPUTPATH_ARG)
	if err != nil {
		return deployScanArguments{}, fmt.Errorf("failed to get outputPath argument: %v", err)
	}

	return deployScanArguments{
		logPath: logPath,
		outputPath: outputPath,
	}, nil
}