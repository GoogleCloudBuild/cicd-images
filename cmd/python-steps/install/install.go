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
package install

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/auth"
	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/command"
	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/spf13/pflag"
)

const (
	IndexURL = "https://pypi.org/simple/"
)

// Arguments represents the arguments passed to the run command.
type Arguments struct {
	Dependencies        []string
	RequirementsPath    string
	ArtifactRegistryURL string
	Verbose             bool
}

// ParseArgs parses the arguments passed to the run command.
func ParseArgs(f *pflag.FlagSet) (Arguments, error) {
	dependencies, err := f.GetString("dependencies")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get dependencies: %w", err)
	}
	requirementsPath, err := f.GetString("requirementsPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get requirementsPath: %w", err)
	}
	artifactRegistryURL, err := f.GetString("artifactRegistryUrl")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get artifactRegistryURL: %w", err)
	}
	artifactRegistryURL, err = auth.EnsureHTTPSByDefault(artifactRegistryURL)
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to ensure artifactRegistryURL is https: %w", err)
	}
	verbose, err := f.GetBool("verbose")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get verbose: %w", err)
	}

	return Arguments{
		Dependencies:        strings.Fields(dependencies),
		RequirementsPath:    requirementsPath,
		ArtifactRegistryURL: artifactRegistryURL,
		Verbose:             verbose,
	}, nil
}

// Execute is the entrypoint for the run command execution.
func Execute(ctx context.Context, runner command.CommandRunner, args Arguments) error {
	logger.SetupLogger(args.Verbose)
	slog.Info("Executing run command", "args", args)

	if err := command.CreateVirtualEnv(runner); err != nil {
		return fmt.Errorf("failed to create virtual environment: %w", err)
	}

	token, err := helper.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("error getting authentication token: %w", err)
	}

	if err := installDependencies(runner, args, token); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	slog.Info("Successfully executed run command")
	return nil
}

func installDependencies(runner command.CommandRunner, args Arguments, gcpToken string) error {
	indexFlags, err := getFlags(args.ArtifactRegistryURL, gcpToken)
	if err != nil {
		return err
	}

	if args.RequirementsPath != "" {
		if err := installFromRequirementsFile(runner, args.RequirementsPath, indexFlags); err != nil {
			return err
		}
	}

	for _, dep := range args.Dependencies {
		if err := installDependency(runner, dep, indexFlags...); err != nil {
			return err
		}
	}

	return nil
}

func getFlags(artifactRegistryURL, gcpToken string) ([]string, error) {
	if artifactRegistryURL == "" {
		return []string{"--index-url=" + IndexURL}, nil
	}

	authenticatedArtifactRegistryURL, err := auth.ConstructAuthenticatedURL(artifactRegistryURL, gcpToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated Artifact Registry URL: %w", err)
	}

	slog.Info("Successfully constructed artifact registry url")
	return []string{"--index-url=" + IndexURL, "--extra-index-url=" + authenticatedArtifactRegistryURL}, nil
}

func installFromRequirementsFile(runner command.CommandRunner, requirementsPath string, flags []string) error {
	slog.Info("Installing dependencies from requirements file", "requirementsPath", requirementsPath)
	args := append([]string{"install", "-r", requirementsPath}, flags...)
	slog.Info("Successfully installed dependencies from requirements file")
	return runner.Run(command.VirtualEnvPip, args...)
}

func installDependency(runner command.CommandRunner, dep string, flags ...string) error {
	slog.Info("Installing a dependency", "dependency", dep)
	args := append([]string{"install", dep}, flags...)
	slog.Info("Successfully installed the dependency")
	return runner.Run(command.VirtualEnvPip, args...)
}
