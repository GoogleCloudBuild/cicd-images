// Copyright 2023 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/cmd/nodejs-steps/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	RESULTS_ARG = "results-path"
	PACK_ARG    = "pack-args"
	PUBLISH_ARG = "publish-args"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish a Node package",
	Long: `package is for packaging and pushing
	a Node project to a GCP Artifact Registry repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		publishCmdArgs, err := parsePublishArgs(cmd.Flags())
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true

		// fetch Artifact Registry Token.
		token, err := internal.GetToken(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to fetch Artifact Registry token: %w", err)
		}

		// check if npmrc exists to try to authenticate with AR
		if _, err := os.Stat(".npmrc"); err == nil {
			if err := internal.AuthenticateNpmrcFile(token); err != nil {
				return fmt.Errorf("failed to install dependencies: %w", err)
			}
		} else {
			fmt.Println("Warning: No .netrc file detected, skipping Artifact Registry Authentication.")
		}

		// Pack the node module into a tar file.
		packCommand := append([]string{"pack"}, publishCmdArgs.packArgs...)
		c := exec.Command("npm", packCommand...)
		var stdout, stderr bytes.Buffer
		c.Stdout = &stdout
		c.Stderr = &stderr
		err = c.Run()
		packageName := strings.TrimSpace(stdout.String())
		if err != nil {
			return fmt.Errorf("error executing 'npm %s': %s\n%w", strings.Join(packCommand, " "), stderr.String(), err)
		}

		// publish the tar file.
		publishArgs := append([]string{"publish", packageName}, publishCmdArgs.publishArgs...)
		c = exec.Command("npm", publishArgs...)
		c.Stdout = &stdout
		c.Stderr = &stderr
		err = c.Run()
		uri := strings.TrimSpace(stdout.String())
		if err != nil {
			return fmt.Errorf("error executing 'npm %s': %s\n%w", strings.Join(publishArgs, " "), stderr.String(), err)
		}
		// remove first two characters of npm publish (e.g. "+  @SCOPE/package@0.0.0" ->  "@SCOPE/package@0.0.0")
		// Define regular expression to match leading whitespace characters
		removeWhitespace := regexp.MustCompile(`^\s+`)
		// Replace leading whitespace with an empty string
		uri = removeWhitespace.ReplaceAllString(uri, "")

		// generate provenance and write ir into provenance json path
		if publishCmdArgs.resultsPath != "" {
			if err := internal.GenerateProvenance(publishCmdArgs.resultsPath, packageName, uri); err != nil {
				return fmt.Errorf("failed to generate provenance: %w", err)
			}
		}

		fmt.Printf("Package %s published successfully!", packageName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringP(RESULTS_ARG, "", "", "Path to write the results in.")
	publishCmd.Flags().StringP(PACK_ARG, "", "", "Extra arguments for npm pack command.")
	publishCmd.Flags().StringP(PUBLISH_ARG, "", "", "Extra arguments for npm publish command.")
}

type publishArguments struct {
	resultsPath string
	packArgs    []string
	publishArgs []string
}

func parsePublishArgs(f *pflag.FlagSet) (publishArguments, error) {
	resultsPath, err := f.GetString(RESULTS_ARG)
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get results path: %w", err)
	}

	packArgs, err := f.GetString(PACK_ARG)
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get `npm pack` extra arguments: %w", err)
	}

	publishArgs, err := f.GetString(PUBLISH_ARG)
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get `npm publish` extra arguments: %w", err)
	}

	return publishArguments{
		resultsPath: resultsPath,
		packArgs:    strings.Fields(packArgs),
		publishArgs: strings.Fields(publishArgs),
	}, nil
}
