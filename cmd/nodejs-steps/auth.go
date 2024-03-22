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

package main

import (
	"fmt"
	"os"

	"github.com/GoogleCloudBuild/cicd-images/cmd/nodejs-steps/internal"
	"github.com/spf13/cobra"
)

const COMMAND_ARG = "command"

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate npmrc file for use with the Artifact Registry.",
	Long: `auth is authenticating with Artifact Registry.
It fetches an access token from application default credentials
or gcloud CLI and writes it into the users npmrc file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// fetch Artifact Registry Token.
		token, err := internal.GetToken(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to fetch Artifact Registry token: %w", err)
		}

		// check if npmrc exists to try to authenticate with AR
		if _, err := os.Stat(".npmrc"); err == nil {
			if err := internal.AuthenticateNpmrcFile(token); err != nil {
				return fmt.Errorf("failed to authenticate npmrc file: %w", err)
			}
		} else {
			fmt.Println("Warning: No .npmrc file detected, creating a new one with Artifact Registry authentication.")
			f, err := os.OpenFile(".npmrc", os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// Handle error opening the file
				return fmt.Errorf("error creating .npmrc file: %w", err)
			}

			// Write configuration lines to the file
			defer f.Close() // Close the file after writing
			_, err = f.WriteString(`@artifact-registry:always-auth=true
//artifact-registry.googleapis.com:_authToken=` + token + `
			`)
			if err != nil {
				// Handle error writing to the file
				return fmt.Errorf("error writing to .npmrc file: %w", err)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
