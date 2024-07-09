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
	"time"

	"github.com/GoogleCloudBuild/cicd-images/cmd/nodejs-steps/internal"
	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish a Node package",
	Long:  "publish is for pushing a Node project to a GCP Artifact Registry repository.",
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

		if err := authenticateAR(ctx); err != nil {
			return err
		}

		if err := publishPackage(); err != nil {
			return err
		}

		// generate provenance and write it into provenance json path
		if publishCmdArgs.resultsPath != "" {
			if err := internal.GenerateProvenance(publishCmdArgs.resultsPath, publishCmdArgs.repository, publishCmdArgs.isBuildArtifact); err != nil {
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

	publishCmd.Flags().StringP("resultsPath", "", "", "Path to write the results in.")
	publishCmd.Flags().StringP("repository", "", "", "Artifact Registry repository to publish the package to.")
	publishCmd.Flags().StringP("isBuildArtifact", "", "true", "A boolean flag specifying if the results should be a build artifact.")
	publishCmd.Flags().BoolP("verbose", "", false, "Whether to print verbose output.")
}

type publishArguments struct {
	resultsPath     string
	repository      string
	isBuildArtifact string
	verbose         bool
}

func parsePublishArgs(f *pflag.FlagSet) (publishArguments, error) {
	resultsPath, err := f.GetString("resultsPath")
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get results path: %w", err)
	}

	repository, err := f.GetString("repository")
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get repository: %w", err)
	}

	isBuildArtifact, err := f.GetString("isBuildArtifact")
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get isBuildArtifact flag: %w", err)
	}

	verbose, err := f.GetBool("verbose")
	if err != nil {
		return publishArguments{}, fmt.Errorf("failed to get verbose: %w", err)
	}

	return publishArguments{
		resultsPath:     resultsPath,
		repository:      repository,
		isBuildArtifact: isBuildArtifact,
		verbose:         verbose,
	}, nil
}

// Authenticate to Artifact Registry with .npmrc file.
func authenticateAR(ctx context.Context) error {
	// fetch Artifact Registry Token.
	token, err := helper.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch Artifact Registry token: %w", err)
	}
	slog.Info("Successfully fetched Artifact Registry token")

	// check if npmrc exists to try to authenticate with AR
	if _, err := os.Stat(".npmrc"); err == nil {
		if err := internal.AuthenticateNpmrcFile(token); err != nil {
			return fmt.Errorf("failed to authenticate npmrc file: %w", err)
		}
	} else {
		slog.Info("Warning: No .npmrc file detected, creating a new one with Artifact Registry authentication.")
		f, err := os.OpenFile(".npmrc", os.O_CREATE|os.O_WRONLY, 0o644)
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

	slog.Info("Successfully authenticated to Artifact Registry")
	return nil
}

// Publish node package to Artifact Registry.
func publishPackage() error {
	c := exec.Command("npm", "publish")
	var stderr bytes.Buffer
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		return fmt.Errorf("error executing 'npm publish': %s\n%w", stderr.String(), err)
	}

	slog.Info("Successfully published the Node package")
	return nil
}
