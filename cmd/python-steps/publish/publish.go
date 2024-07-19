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
package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/spf13/pflag"
)

var (
	filenameRegex    = regexp.MustCompile(`^(?P<name>[a-zA-Z0-9_\-]+)-(?P<version>[0-9a-zA-Z._\-+]+)\.(tar\.gz|whl)$`) // Matches the following filename format: <name>-<version>.<extension>.
	projectNameRegex = regexp.MustCompile(`(.*)\-[0-9]+\.[0-9]+.*`)
)

type ProvenanceOutput struct {
	URI             string `json:"uri"`
	Digest          string `json:"digest"`
	IsBuildArtifact string `json:"isBuildArtifact"`
}

// Arguments represents the arguments passed to the publish command.
type Arguments struct {
	Repository                    string
	Paths                         []string
	SourceDistributionResultsPath string
	WheelDistributionResultsPath  string
	IsBuildArtifact               string
	Verbose                       bool
}

// ParseArgs parses the arguments passed to the publish command.
func ParseArgs(f *pflag.FlagSet) (Arguments, error) {
	repository, err := f.GetString("repository")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get repository: %w", err)
	}
	repository, err = ensureHTTPSByDefault(repository)
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to ensure repository is https: %w", err)
	}
	paths, err := f.GetStringSlice("paths")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get paths: %w", err)
	}
	if err := validatePaths(paths); err != nil {
		return Arguments{}, fmt.Errorf("paths are invalid: %w", err)
	}
	sourceDistributionResultsPath, err := f.GetString("sourceDistributionResultsPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get sourceDistributionResultsPath: %w", err)
	}
	wheelDistributionResultsPath, err := f.GetString("wheelDistributionResultsPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get wheelDistributionResultsPath: %w", err)
	}
	isBuildArtifact, err := f.GetString("isBuildArtifact")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get isBuildArtifact: %w", err)
	}
	verbose, err := f.GetBool("verbose")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get verbose: %w", err)
	}

	return Arguments{
		Repository:                    repository,
		Paths:                         paths,
		SourceDistributionResultsPath: sourceDistributionResultsPath,
		WheelDistributionResultsPath:  wheelDistributionResultsPath,
		IsBuildArtifact:               isBuildArtifact,
		Verbose:                       verbose,
	}, nil
}

// Execute is the entrypoint for the publish command execution.
func Execute(ctx context.Context, runner helper.CommandRunner, args Arguments) error {
	logger.SetupLogger(args.Verbose)

	gcpToken, err := helper.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	slog.Info("Pushing python artifacts to artifact registry")
	for _, path := range args.Paths {
		if err := pushPythonArtifact(runner, args, path, gcpToken); err != nil {
			return fmt.Errorf("failed to push python artifacts: %w", err)
		}
	}
	slog.Info("Successfully pushed python artifacts to artifact registry")

	slog.Info("Generating build results")
	if err := generateResults(args); err != nil {
		return fmt.Errorf("failed to generate provenance: %w", err)
	}
	slog.Info("Successfully generated build results")

	return nil
}

// Publish the python artifact to Artifact Registry.
func pushPythonArtifact(runner helper.CommandRunner, args Arguments, path, gcpToken string) error {
	if args.Repository == "" {
		return fmt.Errorf("artifactRegistryURL is required")
	}

	commands := []string{
		"-m", "twine", "upload",
		"--repository-url", args.Repository,
		"--username", "oauth2accesstoken",
		"--password", gcpToken,
		path,
	}

	if err := runner.Run("python3", commands...); err != nil {
		return fmt.Errorf("failed to push python artifact: %w", err)
	}

	return nil
}

// Generate results for provenance.
func generateResults(args Arguments) error {
	for _, file := range args.Paths {
		uri := generateURI(args.Repository, file)
		digest, err := helper.ComputeDigest(file)
		if err != nil {
			return fmt.Errorf("error computing digest for %s: %w", file, err)
		}

		outputData := ProvenanceOutput{
			URI:             strings.TrimSpace(uri),
			Digest:          strings.TrimSpace(digest),
			IsBuildArtifact: strings.TrimSpace(args.IsBuildArtifact),
		}
		output, err := json.Marshal(outputData)
		if err != nil {
			return fmt.Errorf("error marshalling output data: %w", err)
		}
		slog.Debug("Generated results", "results", output)

		var outputPath string
		switch {
		case strings.HasSuffix(file, ".tar.gz"):
			outputPath = args.SourceDistributionResultsPath
		case strings.HasSuffix(file, ".whl"):
			outputPath = args.WheelDistributionResultsPath
		default:
			continue
		}

		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("file already exists at %s", outputPath)
		}

		err = os.WriteFile(outputPath, output, 0o600)
		if err != nil {
			return fmt.Errorf("error writing to %s: %w", outputPath, err)
		}
	}

	return nil
}

func generateURI(repository, filePath string) string {
	fileName := filepath.Base(filePath) // extract file name
	projectName := projectNameRegex.FindStringSubmatch(fileName)[1]

	uri := fmt.Sprintf("%s/%s/%s", repository, projectName, fileName)
	return uri
}

// ensureHTTPSByDefault ensures that the given registry URL has a https scheme.
func ensureHTTPSByDefault(registryURL string) (string, error) {
	if registryURL == "" {
		return "", nil
	}

	parsedURL, err := url.Parse(registryURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// If the scheme is missing, default to HTTPS
	if parsedURL.Scheme == "" {
		_, err := url.ParseRequestURI("https://" + registryURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL after adding https scheme: %w", err)
		}
		parsedURL.Scheme = "https"
	}

	return parsedURL.String(), nil
}

// Check the paths are for valid artifacts.
func validatePaths(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("'paths' parameter is empty")
	}

	for _, path := range paths {
		// check if file or dir
		fileInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("error describing path in 'paths' array: %w", err)
		}

		if fileInfo.IsDir() {
			return fmt.Errorf("invalid file path: %s", path)
		}
		fileName := filepath.Base(path)
		// Matches the following filename format: <name>-<version>.<extension>.
		matches := filenameRegex.FindStringSubmatch(fileName)
		if matches == nil || len(matches) < 4 {
			return fmt.Errorf("invalid filename format: %s", fileName)
		}
	}

	return nil
}
