// Copyright 2023 Google LLC All Rights Reserved.
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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoogleCloudBuild/node-images/steps/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const COMMAND_ARG = "command"

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "runner",
	Short: "Run a Node command.",
	Long: `node is for running a Node command.
It can perform commands from npm, npx and yarn ClIs. It can also
pull Node artifacts from a GCP Artifact Registry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeCmdArgs, err := parseNodeArgs(cmd.Flags())
		if err != nil {
			return err
		}

		command := nodeCmdArgs.Command
		var runtime string

		runtime = command[0]

		if runtime != "npm" && runtime != "npx" && runtime != "yarn" && runtime != "node" {
			return fmt.Errorf("not a valid executing command. Expecting: (npm|npx|yarn), got: %s", runtime)
		}

		// fetch Artifact Registry Token.
		token, err := internal.GetToken(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to fetch Artifact Registry token: %v", err)
		}

		// check if npmrc exists to try to authenticate with AR
		if _, err := os.Stat(".npmrc"); err == nil {
			if err := internal.AuthenticateNpmrcFile(token); err != nil {
				return fmt.Errorf("failed to authenticate npmrc file: %v", err)
			}
		} else {
			fmt.Println("Warning: No .npmrc file detected, creating a new one with Artifact Registry authentication.")
			f, err := os.OpenFile(".npmrc", os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// Handle error opening the file
				return fmt.Errorf("Error creating .npmrc file: %v", err)
			}

			// Write configuration lines to the file
			defer f.Close() // Close the file after writing
			_, err = f.WriteString(`@artifact-registry:always-auth=true
			//artifact-registry.googleapis.com:_authToken=` + token + `
			`)
			if err != nil {
				// Handle error writing to the file
				return fmt.Errorf("Error writing to .npmrc file: %v", err)
			}
		}

		// extract arguements to execute with defined runtime.
		if subCommand := command[1:]; subCommand != nil {
			out, err := internal.RunCmd(runtime, subCommand...)
			if err != nil {
				return fmt.Errorf("error executing '%s %s': %s", runtime, strings.Join(subCommand, " "), err)
			}
			fmt.Println(out)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	nodeCmd.Flags().StringP(COMMAND_ARG, "", "", "The Node command to run, e.g. `npm run start` | `yarn ...`")
	nodeCmd.MarkFlagRequired(COMMAND_ARG)
}

type nodeArguments struct {
	Command []string
}

func parseNodeArgs(f *pflag.FlagSet) (nodeArguments, error) {
	command, err := f.GetString(COMMAND_ARG)
	if err != nil {
		return nodeArguments{}, fmt.Errorf("failed to get command: %v", err)
	}
	if len(command) < 1 {
		return nodeArguments{}, fmt.Errorf("Invalid command. Command can not be empty.")
	}
	return nodeArguments{
		Command: strings.Fields(command),
	}, nil
}
