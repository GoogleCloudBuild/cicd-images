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
	"fmt"
	"net/url"
)

const (
	GKE_METADATA_SERVER_ENDPOINT = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
)

func ConstructAuthenticatedURL(registryURL, token string) (string, error) {
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
