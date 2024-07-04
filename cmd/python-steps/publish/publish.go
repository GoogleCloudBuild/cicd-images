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
	"github.com/package-url/packageurl-go"
	"github.com/spf13/pflag"
)

// Matches the following filename format: <name>-<version>.<extension>.
var filenameRegex = regexp.MustCompile(`^(?P<name>[a-zA-Z0-9_\-]+)-(?P<version>[0-9a-zA-Z._\-+]+)\.(tar\.gz|whl)$`)

type ProvenanceOutput struct {
	URI             string `json:"uri"`
	Digest          string `json:"digest"`
	IsBuildArtifact string `json:"isBuildArtifact"`
}

// Arguments represents the arguments passed to the publish command.
type Arguments struct {
	ArtifactRegistryURL           string
	ArtifactDir                   string
	SourceDistributionResultsPath string
	WheelDistributionResultsPath  string
	IsBuildArtifact               string
	Verbose                       bool
}

// ParseArgs parses the arguments passed to the publish command.
func ParseArgs(f *pflag.FlagSet) (Arguments, error) {
	artifactRegistryURL, err := f.GetString("artifactRegistryUrl")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get artifactRegistryURL: %w", err)
	}
	artifactRegistryURL, err = ensureHTTPSByDefault(artifactRegistryURL)
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to ensure artifactRegistryURL is https: %w", err)
	}
	artifactDir, err := f.GetString("artifactDir")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get artifacPath: %w", err)
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
		ArtifactRegistryURL:           artifactRegistryURL,
		ArtifactDir:                   artifactDir,
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

	slog.Info("Publishing python artifact")
	if err := pushPythonArtifact(runner, args, gcpToken); err != nil {
		return fmt.Errorf("failed to push python artifact: %w", err)
	}

	if err := generateResults(args.ArtifactDir, args); err != nil {
		return fmt.Errorf("failed to generate provenance: %w", err)
	}

	slog.Info("Successfully published artifact")
	return nil
}

func pushPythonArtifact(runner helper.CommandRunner, args Arguments, gcpToken string) error {
	slog.Info("Pushing python artifact to artifact registry")
	if args.ArtifactRegistryURL == "" {
		return fmt.Errorf("artifactRegistryURL is required")
	}

	commands := []string{
		"-m", "twine", "upload",
		"--repository-url", args.ArtifactRegistryURL,
		"--username", "oauth2accesstoken",
		"--password", gcpToken,
		args.ArtifactDir + "/*",
	}
	if err := runner.Run("python3", commands...); err != nil {
		return fmt.Errorf("failed to push python artifacts: %w", err)
	}

	slog.Info("Successfully pushed python artifact to artifact registry")
	return nil
}

func generateResults(distDir string, args Arguments) error {
	slog.Info("Generating build results")
	files, err := os.ReadDir(distDir)
	if err != nil {
		return fmt.Errorf("error reading dist directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in dist directory")
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		uri, err := generateURI(args.ArtifactRegistryURL, file.Name())
		if err != nil {
			return fmt.Errorf("error generating URI for %s: %w", file.Name(), err)
		}
		digest, err := helper.ComputeDigest(filepath.Join(distDir, file.Name()))
		if err != nil {
			return fmt.Errorf("error computing digest for %s: %w", file.Name(), err)
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
		case strings.HasSuffix(file.Name(), ".tar.gz"):
			outputPath = args.SourceDistributionResultsPath
		case strings.HasSuffix(file.Name(), ".whl"):
			outputPath = args.WheelDistributionResultsPath
		default:
			continue
		}

		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("file already exists at %s", outputPath)
		}

		err = os.WriteFile(outputPath, output, 0444)
		if err != nil {
			return fmt.Errorf("error writing to %s: %w", outputPath, err)
		}
	}

	slog.Info("Successfully generated build results")
	return nil
}

func generateURI(artifactRegistryURL, fileName string) (string, error) {
	// Matches the following filename format: <name>-<version>.<extension>.
	matches := filenameRegex.FindStringSubmatch(fileName)
	if matches == nil || len(matches) < 4 {
		return "", fmt.Errorf("invalid filename format: %s", fileName)
	}
	packageName := matches[1]
	packageVersion := matches[2]

	parsedURL, err := url.Parse(artifactRegistryURL)
	if err != nil {
		return "", fmt.Errorf("error parsing artifactRegistryURL: %w", err)
	}

	purl := packageurl.NewPackageURL(parsedURL.Host, parsedURL.Path, packageName, packageVersion, packageurl.Qualifiers{}, "")
	return purl.ToString(), nil
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
