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
package cmd

import (
	"os"
	"os/exec"

	"go.uber.org/zap"
)

const (
	Python3           = "python3"
	VirtualEnvName    = "venv"
	VirtualEnvPip     = VirtualEnvName + "/bin/pip"
	VirtualEnvPython3 = VirtualEnvName + "/bin/python3"
)

// CommandRunner is an interface for running commands.
type CommandRunner interface {
	Run(logger *zap.Logger, cmd string, args ...string) error
}

// DefaultCommandRunner is the default implementation of CommandRunner.
type DefaultCommandRunner struct{}

// Run runs the given command with the given arguments.
func (r *DefaultCommandRunner) Run(logger *zap.Logger, cmd string, args ...string) error {
	logger.Debug("Running command", zap.String("cmd", cmd), zap.Strings("args", args))
	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

// CreateVirtualEnv creates a python virtual environment.
func CreateVirtualEnv(runner CommandRunner, logger *zap.Logger) error {
	logger.Info("Creating virtual environment", zap.String("virtualEnvName", VirtualEnvName))
	return runner.Run(logger, Python3, "-m", "venv", VirtualEnvName)
}
