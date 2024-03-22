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
package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

const (
	GKE_METADATA_SERVER_ENDPOINT = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
)

// HTTPClient is an interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TokenResponse represents the response from the GKE metadata server.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

// GetArtifactRegistryURL returns the authenticated URL for the given artifact registry URL.
func GetArtifactRegistryURL(client HTTPClient, registryURL string, logger *zap.Logger) (string, error) {
	logger.Info("Getting the artifact registry URL")
	gcpToken, err := GetGCPToken(client, logger)
	if err != nil {
		return "", err
	}

	authenticatedURL, err := constructAuthenticatedURL(registryURL, gcpToken)
	if err != nil {
		return "", err
	}

	logger.Info("Successfully got the artifact registry URL")
	return authenticatedURL, nil
}

// GetGCPToken retrieves a GCP token from the GKE metadata server.
func GetGCPToken(client HTTPClient, logger *zap.Logger) (string, error) {
	logger.Info("Getting the GCP token")
	// Create a HTTP request for the GKE metadata server endpoint.
	req, err := http.NewRequest("GET", GKE_METADATA_SERVER_ENDPOINT, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decode the JSON response.
	var tokenData TokenResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&tokenData); err != nil {
		return "", err
	}

	if tokenData.AccessToken == "" {
		return "", fmt.Errorf("no access token in the response from metadata server")
	}

	logger.Info("Successfully got the GCP token")
	return tokenData.AccessToken, nil
}

func constructAuthenticatedURL(registryURL, token string) (string, error) {
	u, err := url.Parse(registryURL)
	if err != nil {
		return "", err
	}
	// add "/simple" to the end of the URL to make it a valid PEP 503 URL.
	u = u.JoinPath("simple")
	u.User = url.UserPassword("oauth2accesstoken", token)
	return u.String(), nil
}

// EnsureHTTPSByDefault ensures that the given registry URL has a https scheme.
func EnsureHTTPSByDefault(registryURL string) (string, error) {
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
