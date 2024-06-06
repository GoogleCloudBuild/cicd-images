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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudBuild/cicd-images/cmd/maven-steps/internal/xmlmodules"
)

const tokenPathSuffix = "instance/service-accounts/default/token"

type response struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// Get access token to call GCP API's with, can pass scopes as an query param
func GetAuthenticationToken(c *metadata.Client, ctx context.Context) (string, error) {
	accessTokenJSON, err := c.GetWithContext(ctx, tokenPathSuffix)
	if err != nil {
		return "", fmt.Errorf("requesting token: %w", err)
	}
	var jsonRes response
	err = json.NewDecoder(strings.NewReader(accessTokenJSON)).Decode(&jsonRes)
	if err != nil {
		return "", fmt.Errorf("retrieving auth token: %w", err)
	}
	return string(jsonRes.AccessToken), nil
}

// Create settings.xml with authentication token
func WriteSettingsXML(token string, localRepository string, repositoryIds []string, settingsPath string) error {
	settings := xmlmodules.Settings{
		Servers:         xmlmodules.Servers{},
		LocalRepository: xmlmodules.LocalRepository{Path: localRepository},
	}

	for _, repoId := range repositoryIds {
		server := xmlmodules.Server{
			ID: repoId,
			Configuration: xmlmodules.Configuration{
				HttpConfiguration: xmlmodules.HttpConfiguration{
					Get:  true,
					Head: true,
					Put: xmlmodules.PutParams{
						Property: []xmlmodules.Property{
							{
								Name:  "http.protocol.expect-continue",
								Value: "false",
							},
						},
					},
				},
			},
			// Authentication token reference: https://cloud.google.com/artifact-registry/docs/helm/authentication#token
			Username: "oauth2accesstoken",
			Password: token,
		}
		settings.Servers.Server = append(settings.Servers.Server, server)
	}

	file, err := os.OpenFile(settingsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening settings.xml: %w", err)
	}
	defer file.Close()
	// Write the XML data to a file
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	err = encoder.Encode(settings)
	if err != nil {
		return fmt.Errorf("error writing to settings.xml file: %w", err)
	}

	return nil
}

// Get SHA1 checksum of remote artifact in Artifact Registry
func GetCheckSum(artifactRegistryUrl string, ctx context.Context) (string, error) {
	c := metadata.NewClient(&http.Client{})
	token, err := GetAuthenticationToken(c, ctx)
	if err != nil {
		return "", fmt.Errorf("error getting auth token: %w", err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s.sha1", artifactRegistryUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error downloading checksum: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	checksum := fmt.Sprintf("sha1:%s", string(body))
	return checksum, nil
}
