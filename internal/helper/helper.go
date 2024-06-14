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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/auth"
)

const tokenPathSuffix = "instance/service-accounts/default/token"

type response struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type metadataClient interface {
	GetWithContext(ctx context.Context, suffix string) (string, error)
}

// Get GCP authentication token from the metadata server.
// Parameter 'client' is of type *metadata.Client.
// Refer to https://pkg.go.dev/cloud.google.com/go/compute/metadata.
func GetAuthenticationToken(ctx context.Context, client metadataClient) (string, error) {
	accessTokenJSON, err := client.GetWithContext(ctx, tokenPathSuffix)
	if err != nil {
		return "", fmt.Errorf("requesting token: %w", err)
	}

	var jsonRes response
	err = json.NewDecoder(strings.NewReader(accessTokenJSON)).Decode(&jsonRes)
	if err != nil {
		return "", fmt.Errorf("retrieving auth token: %w", err)
	}

	if jsonRes.AccessToken == "" {
		return "", fmt.Errorf("no access token in the response from metadata server")
	}

	return string(jsonRes.AccessToken), nil
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

	return hex.EncodeToString(hash[:]), nil
}
