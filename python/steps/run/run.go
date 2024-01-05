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
package run

import (
	"fmt"
	"strings"

	"github.com/GoogleCloudBuild/python-images/steps/internal/auth"
	"github.com/GoogleCloudBuild/python-images/steps/internal/cmd"
	"github.com/GoogleCloudBuild/python-images/steps/internal/logger"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const (
	IndexURL = "https://pypi.org/simple/"
)

// Arguments represents the arguments passed to the run command.
type Arguments struct {
	Command             string
	Dependencies        []string
	RequirementsPath    string
	ArtifactRegistryUrl string
	Verbose             bool
	Script              string
}

// ParseArgs parses the arguments passed to the run command.
func ParseArgs(f *pflag.FlagSet) (Arguments, error) {
	command, err := f.GetString("command")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get command: %v", err)
	}
	dependencies, err := f.GetString("dependencies")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get dependencies: %v", err)
	}
	requirementsPath, err := f.GetString("requirementsPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get requirementsPath: %v", err)
	}
	artifactRegistryUrl, err := f.GetString("artifactRegistryUrl")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get artifactRegistryUrl: %v", err)
	}
	artifactRegistryUrl, err = auth.EnsureHTTPSByDefault(artifactRegistryUrl)
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to ensure artifactRegistryUrl is https: %v", err)
	}
	verbose, err := f.GetBool("verbose")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get verbose: %v", err)
	}
	script, err := f.GetString("script")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get script: %v", err)
	}

	return Arguments{
		Command:             command,
		Dependencies:        strings.Fields(dependencies),
		RequirementsPath:    requirementsPath,
		ArtifactRegistryUrl: artifactRegistryUrl,
		Verbose:             verbose,
		Script:              script,
	}, nil
}

// Execute is the entrypoint for the run command execution.
func Execute(runner cmd.CommandRunner, args Arguments, client auth.HTTPClient) error {
	logger, err := logger.SetupLogger(args.Verbose)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %v", err)
	}
	defer logger.Sync()
	logger.Info("Executing run command", zap.Any("args", args))

	if err := cmd.CreateVirtualEnv(runner, logger); err != nil {
		return fmt.Errorf("failed to create virtual environment: %v", err)
	}

	if err := installDependencies(runner, args, client, logger); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	if script := args.Script; script != "" {
		if err := runPythonScript(runner, script, logger); err != nil {
			return fmt.Errorf("failed to run inline script: %v", err)
		}
	} else if command := args.Command; command != "" {
		if err := runPythonCommand(runner, command, logger); err != nil {
			return fmt.Errorf("failed to run command: %v", err)
		}
	}

	logger.Info("Successfully executed run command")
	return nil
}

func installDependencies(runner cmd.CommandRunner, args Arguments, client auth.HTTPClient, logger *zap.Logger) error {
	logger.Info("Installing dependencies", zap.Strings("dependencies", args.Dependencies))

	indexFlags, err := authenticateRegistryAndGetFlags(args.ArtifactRegistryUrl, client, logger)
	if err != nil {
		return err
	}

	if args.RequirementsPath != "" {
		if err := installFromRequirementsFile(runner, args.RequirementsPath, indexFlags, logger); err != nil {
			return err
		}
	}

	for _, dep := range args.Dependencies {
		if err := installDependency(logger, runner, dep, indexFlags...); err != nil {
			return err
		}
	}

	logger.Info("Successfully installed dependencies")
	return nil
}

func authenticateRegistryAndGetFlags(artifactRegistryUrl string, client auth.HTTPClient, logger *zap.Logger) ([]string, error) {
	logger.Info("Authenticating registry", zap.String("artifactRegistryUrl", artifactRegistryUrl))

	if artifactRegistryUrl == "" {
		return []string{"--index-url=" + IndexURL}, nil
	}

	authenticatedArtifactRegistryURL, err := auth.GetArtifactRegistryURL(client, artifactRegistryUrl, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated Artifact Registry URL: %v", err)
	}

	logger.Info("Successfully authenticated registry")
	return []string{"--index-url=" + IndexURL, "--extra-index-url=" + authenticatedArtifactRegistryURL}, nil
}

func installFromRequirementsFile(runner cmd.CommandRunner, requirementsPath string, flags []string, logger *zap.Logger) error {
	logger.Info("Installing dependencies from requirements file", zap.String("requirementsPath", requirementsPath))
	args := append([]string{"install", "-r", requirementsPath}, flags...)
	logger.Info("Successfully installed dependencies from requirements file")
	return runner.Run(logger, cmd.VirtualEnvPip, args...)
}

func installDependency(logger *zap.Logger, runner cmd.CommandRunner, dep string, flags ...string) error {
	logger.Info("Installing dependency", zap.String("dependency", dep))
	args := append([]string{"install", dep}, flags...)
	logger.Info("Successfully installed dependency")
	return runner.Run(logger, cmd.VirtualEnvPip, args...)
}

func runPythonCommand(runner cmd.CommandRunner, command string, logger *zap.Logger) error {
	logger.Info("Running python command", zap.String("command", command))
	if command == "" {
		return fmt.Errorf("command is required")
	}
	commands := strings.Fields(command)
	if err := runner.Run(logger, cmd.VirtualEnvPython3, commands...); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}
	logger.Info("Successfully ran python command")
	return nil
}

func runPythonScript(runner cmd.CommandRunner, script string, logger *zap.Logger) error {
	logger.Info("Running supplied python script")
	if script == "" {
		return fmt.Errorf("script is required")
	}
	commands := []string{"-c", script}
	if err := runner.Run(logger, cmd.VirtualEnvPython3, commands...); err != nil {
		return fmt.Errorf("failed to run the supplied python script: %v", err)
	}
	logger.Info("Successfully ran the python script")
	return nil
}
