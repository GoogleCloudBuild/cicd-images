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
package command

import (
	"log/slog"
	"os"
	"os/exec"
)

const (
	Python3           = "python3"
	VirtualEnvName    = "venv"
	VirtualEnvPip     = VirtualEnvName + "/bin/pip"
	VirtualEnvPython3 = VirtualEnvName + "/bin/python3"
)

// CommandRunner is an interface for running commands.
type CommandRunner interface {
	Run(cmd string, args ...string) error
}

// DefaultCommandRunner is the default implementation of CommandRunner.
type DefaultCommandRunner struct{}

// Run runs the given command with the given arguments.
func (r *DefaultCommandRunner) Run(cmd string, args ...string) error {
	slog.Debug("Running command", "cmd", cmd, "args", args)
	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

// CreateVirtualEnv creates a python virtual environment.
func CreateVirtualEnv(runner CommandRunner) error {
	slog.Info("Creating virtual environment")
	return runner.Run(Python3, "-m", "venv", VirtualEnvName)
}
