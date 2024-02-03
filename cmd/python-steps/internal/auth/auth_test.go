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
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var mockLogger = zap.NewNop()

func matchGCPTokenRequest(req *http.Request) bool {
	return req.Method == "GET" &&
		req.URL.String() == GKE_METADATA_SERVER_ENDPOINT &&
		req.Header.Get("Metadata-Flavor") == "Google"
}

func TestGetGCPToken(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		mockClient.On("Do", mock.MatchedBy(matchGCPTokenRequest)).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "foo-token"}`))),
		}, nil)

		_, err := GetGCPToken(mockClient, mockLogger)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("no token in response", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		mockClient.On("Do", mock.MatchedBy(matchGCPTokenRequest)).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": ""}`))),
		}, nil)

		_, err := GetGCPToken(mockClient, mockLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no access token in the response")
	})

}

func TestConstructAuthenticatedURL(t *testing.T) {
	t.Run("valid url", func(t *testing.T) {
		registryURL := "https://foo-registry.com/python-repo-name"
		token := "foo-token"
		expectedURL := "https://oauth2accesstoken:foo-token@foo-registry.com/python-repo-name/simple"

		constructedURL, err := constructAuthenticatedURL(registryURL, token)

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, constructedURL)
	})

	t.Run("invalid url", func(t *testing.T) {
		invalidURL := "http://%"
		token := "foo-token"

		_, err := constructAuthenticatedURL(invalidURL, token)

		assert.Error(t, err)
	})
}

func TestGetArtifactRegistryURL(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		mockClient.On("Do", mock.MatchedBy(matchGCPTokenRequest)).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "foo-token"}`))),
		}, nil)
		registryURL := "https://example-registry.com"
		expectedAuthenticatedURL := "https://oauth2accesstoken:foo-token@example-registry.com/simple"

		authenticatedURL, err := GetArtifactRegistryURL(mockClient, registryURL, mockLogger)

		assert.NoError(t, err)
		assert.Equal(t, expectedAuthenticatedURL, authenticatedURL)
	})

	t.Run("failed to fetch token", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		mockClient.On("Do", mock.MatchedBy(matchGCPTokenRequest)).Return(&http.Response{}, errors.New("failed to fetch token"))

		_, err := GetArtifactRegistryURL(mockClient, "https://example-registry.com", mockLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch token")
	})

	t.Run("failed to construct authenticated url", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		mockClient.On("Do", mock.MatchedBy(matchGCPTokenRequest)).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "foo-token"}`))),
		}, nil)

		_, err := GetArtifactRegistryURL(mockClient, "http://%/example-registry.com", mockLogger)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid URL")
	})
}

func TestEnsureHTTPSByDefault(t *testing.T) {
	t.Run("valid url", func(t *testing.T) {
		registryURL := "https://example-registry.com"
		expectedURL := "https://example-registry.com"

		constructedURL, err := EnsureHTTPSByDefault(registryURL)

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, constructedURL)
	})

	t.Run("invalid url", func(t *testing.T) {
		invalidURL := "http://%"
		_, err := EnsureHTTPSByDefault(invalidURL)

		assert.Error(t, err)
	})

	t.Run("add https schema if none specified", func(t *testing.T) {
		registryURL := "example-registry.com"
		expectedURL := "https://example-registry.com"

		constructedURL, err := EnsureHTTPSByDefault(registryURL)

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, constructedURL)
	})
}
