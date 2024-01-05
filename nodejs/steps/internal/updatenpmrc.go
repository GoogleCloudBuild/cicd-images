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
package internal

import (
	"context"
	"os"
	"path"

	"github.com/GoogleCloudBuild/node-images/steps/pkg/npmrc"
	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/auth"
)

// GetToken receives a context and returns the Token from Artifact Registry.
func GetToken(ctx context.Context) (string, error) {
	token, err := auth.Token(ctx)
	if err != nil {
		return "", err
	}
	return token, nil
}

// AuthenticateNpmrcFile receives a token to write into the user's .npmrc file
func AuthenticateNpmrcFile(token string) error {
	// Load per-project npmrc file
	projectConfig, err := npmrc.Load(".npmrc")
	if err != nil {
		return err
	}

	// Add token to npmrc file
	config := npmrc.AddTokenToConfigFile(projectConfig, token)

	// Getting home directory.
	h, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Save updated npmrc file to per-user config file.
	if err := npmrc.Save(config, path.Join(h, ".npmrc")); err != nil {
		return err
	}

	return nil
}
