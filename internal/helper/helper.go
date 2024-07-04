// Copyright 2024 Google LLC
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
package helper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/auth"
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

// Get oauth2 access token from Application Default Credentials (ADC) for Google Cloud Service.
func GetAccessToken(ctx context.Context) (string, error) {
	return auth.Token(ctx)
}

// Compute SHA256 digest of packaged file.
func ComputeDigest(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %w", filePath, err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("empty file: %s", filePath)
	}

	hash := sha256.Sum256(data)

	return "sha256:" + hex.EncodeToString(hash[:]), nil
}
