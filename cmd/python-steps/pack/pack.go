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
package pack

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/auth"
	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/command"
	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/package-url/packageurl-go"
	"github.com/spf13/pflag"
)

const (
	PYTHON_DIST_DIR = "dist"
)

var (
	// Matches the following filename format: <name>-<version>.<extension>
	filenameRegex = regexp.MustCompile(`^(?P<name>[a-zA-Z0-9_\-]+)-(?P<version>[0-9a-zA-Z._\-+]+)\.(tar\.gz|whl)$`)
)

type ProvenanceOutput struct {
	URI    string `json:"uri"`
	Digest string `json:"digest"`
}

// Arguments represents the arguments passed to the pack command.
type Arguments struct {
	Command                       []string
	ArtifactRegistryUrl           string
	SourceDistributionResultsPath string
	WheelDistributionResultsPath  string
	Verbose                       bool
}

// ParseArgs parses the arguments passed to the pack command.
func ParseArgs(f *pflag.FlagSet) (Arguments, error) {
	command, err := f.GetString("command")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get command: %w", err)
	}
	artifactRegistryUrl, err := f.GetString("artifactRegistryUrl")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get artifactRegistryUrl: %w", err)
	}
	artifactRegistryUrl, err = auth.EnsureHTTPSByDefault(artifactRegistryUrl)
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to ensure artifactRegistryUrl is https: %w", err)
	}
	sourceDistributionResultsPath, err := f.GetString("sourceDistributionResultsPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get sourceDistributionResultsPath: %w", err)
	}
	wheelDistributionResultsPath, err := f.GetString("wheelDistributionResultsPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get wheelDistributionResultsPath: %w", err)
	}
	verbose, err := f.GetBool("verbose")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get verbose: %w", err)
	}

	return Arguments{
		Command:                       strings.Fields(command),
		ArtifactRegistryUrl:           artifactRegistryUrl,
		SourceDistributionResultsPath: sourceDistributionResultsPath,
		WheelDistributionResultsPath:  wheelDistributionResultsPath,
		Verbose:                       verbose,
	}, nil
}

// Execute is the entrypoint for the package command execution.
func Execute(runner command.CommandRunner, args Arguments, client auth.HTTPClient) error {
	logger.SetupLogger(args.Verbose)
	slog.Info("Executing pack command", "args", args)

	if err := command.CreateVirtualEnv(runner); err != nil {
		return fmt.Errorf("failed to create virtual environment: %w", err)
	}

	if err := installDependenciesForPackage(runner); err != nil {
		return fmt.Errorf("failed to install dependencies for package command: %w", err)
	}

	if err := packagePythonArtifact(runner, args); err != nil {
		return fmt.Errorf("failed to package python artifact: %w", err)
	}

	if err := pushPythonArtifact(runner, args, client); err != nil {
		return fmt.Errorf("failed to push python artifact: %w", err)
	}

	if err := generateResults(PYTHON_DIST_DIR, args); err != nil {
		return fmt.Errorf("failed to generate provenance: %w", err)
	}

	slog.Info("Successfully executed pack command")
	return nil
}

func installDependenciesForPackage(runner command.CommandRunner) error {
	slog.Info("Installing dependencies for package command")
	commands := []string{
		"install", "build", "wheel", "twine",
	}
	if err := runner.Run(command.VirtualEnvPip, commands...); err != nil {
		return err
	}
	slog.Info("Successfully installed dependencies for package command")
	return nil
}

func packagePythonArtifact(runner command.CommandRunner, args Arguments) error {
	slog.Info("Packaging python artifact")
	if len(args.Command) == 0 {
		return fmt.Errorf("command is required")
	}
	if err := runner.Run(command.VirtualEnvPython3, args.Command...); err != nil {
		return fmt.Errorf("failed to package python artifacts: %w", err)
	}
	slog.Info("Successfully packaged python artifact")
	return nil
}

func pushPythonArtifact(runner command.CommandRunner, args Arguments, client auth.HTTPClient) error {
	slog.Info("Pushing python artifact to artifact registry")
	if args.ArtifactRegistryUrl == "" {
		return fmt.Errorf("artifactRegistryUrl is required")
	}

	gcpToken, err := auth.GetGCPToken(client)
	if err != nil {
		return err
	}

	commands := []string{
		"-m", "twine", "upload",
		"--repository-url", args.ArtifactRegistryUrl,
		"--username", "oauth2accesstoken",
		"--password", gcpToken,
		"dist/*",
	}
	if err := runner.Run(command.VirtualEnvPython3, commands...); err != nil {
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

		uri, err := generateUri(args.ArtifactRegistryUrl, file.Name())
		if err != nil {
			return fmt.Errorf("error generating URI for %s: %w", file.Name(), err)
		}
		digest, err := computeDigest(filepath.Join(distDir, file.Name()))
		if err != nil {
			return fmt.Errorf("error computing digest for %s: %w", file.Name(), err)
		}

		outputData := ProvenanceOutput{
			URI:    uri,
			Digest: digest,
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

func generateUri(artifactRegistryUrl string, fileName string) (string, error) {
	// Matches the following filename format: <name>-<version>.<extension>
	matches := filenameRegex.FindStringSubmatch(fileName)
	if matches == nil || len(matches) < 4 {
		return "", fmt.Errorf("invalid filename format: %s", fileName)
	}
	packageName := matches[1]
	packageVersion := matches[2]

	parsedUrl, err := url.Parse(artifactRegistryUrl)
	if err != nil {
		return "", fmt.Errorf("error parsing ArtifactRegistryUrl: %w", err)
	}

	purl := packageurl.NewPackageURL(parsedUrl.Host, parsedUrl.Path, packageName, packageVersion, packageurl.Qualifiers{}, "")
	return purl.ToString(), nil
}

func computeDigest(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %w", filePath, err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("empty file: %s", filePath)
	}

	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:]), nil
}
