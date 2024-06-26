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
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/GoogleCloudBuild/cicd-images/cmd/nodejs-steps/internal"
	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	RESULTS_ARG = "results-path"
	PACK_ARG    = "pack-args"
	PUBLISH_ARG = "publish-args"
	VERBOSE     = "verbose"
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

		logger.SetupLogger(publishCmdArgs.verbose)
		slog.Info("Executing publish command")

		cmd.SilenceUsage = true
		ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
		defer cf()

		// fetch Artifact Registry Token.
		token, err := helper.GetAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch Artifact Registry token: %w", err)
		}
		slog.Info("Successfully fetched Artifact Registry token")

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
		slog.Info("Successfully packed the Node module")

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
		slog.Info("Successfully published the Node module")
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
			slog.Info("Successfully generated provenance")
		}

		slog.Info("Successfully executed publish command")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringP(RESULTS_ARG, "", "", "Path to write the results in.")
	publishCmd.Flags().StringP(PACK_ARG, "", "", "Extra arguments for npm pack command.")
	publishCmd.Flags().StringP(PUBLISH_ARG, "", "", "Extra arguments for npm publish command.")
	publishCmd.Flags().BoolP(VERBOSE, "", false, "Whether to print verbose output.")
}

type publishArguments struct {
	resultsPath string
	packArgs    []string
	publishArgs []string
	verbose     bool
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

	verbose, err := f.GetBool(VERBOSE)
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get verbose: %w", err)
	}

	return publishArguments{
		resultsPath: resultsPath,
		packArgs:    strings.Fields(packArgs),
		publishArgs: strings.Fields(publishArgs),
		verbose:     verbose,
	}, nil
}
