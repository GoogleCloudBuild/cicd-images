// Copyright 2023 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deployscanlog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/compute/metadata"
)

const (
	uploadPattern             = `Uploaded to [^\s]+: (https:\/\/[^ ]+)`
	repositoryPattern         = `https?:\/\/[^\s]+`
	artifactsDirectoryDefault = "./target"
	tokenPathSuffix           = "instance/service-accounts/default/token"
)

var (
	mavenPackagingTypes = []string{"jar", "ejb", "war", "ear", "rar"}
	artifactsDirectory  string
)

// Go through the maven deploy log and scan to retrieve artifacts.
func ScanDeployLog(logPath string, outputPath string) error {
	// Read the file contents
	content, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("reading log file %w", err)
	}

	// Convert the file content to a string
	logs := string(content)

	artifacts := retrieveArtifacts(logs)
	// Create or open the file for writing
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close() // Ensure the file is closed when we're done with it

	checksums, err := generateCheckSum(artifacts)
	if err != nil {
		return fmt.Errorf("generateCheckSum: %w", err)
	}
	// Write each string to the file

	for uri, digest := range checksums {
		str := fmt.Sprintf("%s %s\n", uri, digest)
		fmt.Printf("Write: %s\n", str)
		_, err := fmt.Fprintf(file, fmt.Sprintf("%s %s\n", uri, digest))
		if err != nil {
			return fmt.Errorf("writing to file: %w", err)
		}
	}

	fmt.Printf("Strings written to %s\n", outputPath)
	return nil
}

func generateCheckSum(artifacts map[string]string) (map[string]string, error) {
	checksums := map[string]string{}

	c := metadata.NewClient(&http.Client{})
	token, err := getAuthenticationToken(c)
	if err != nil {
		return checksums, fmt.Errorf("getAuthenticationToken errors: %w", err)
	}
	fmt.Println("Token retrieval success!")

	for uri, _ := range artifacts {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s.sha256", uri), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("downloading checksum: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return nil, fmt.Errorf("reading checksum: %w", err)
		}

		fmt.Printf("SHA-256 Checksum: %s\n", string(body))

		checksums[uri] = fmt.Sprintf("sha256:%s", string(body))
	}

	return checksums, nil
}

func retrieveArtifacts(logs string) map[string]string {
	results := map[string]string{}
	re := regexp.MustCompile(uploadPattern)
	repositoryRE := regexp.MustCompile(repositoryPattern)

	matches := re.FindAllStringSubmatch(logs, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(match[1], " ")
			for _, line := range lines {
				repoMatch := repositoryRE.FindStringSubmatch(line)
				if len(repoMatch) > 0 {
					if hasSuffixInList(repoMatch[0], mavenPackagingTypes) {
						fmt.Printf("Retrieved artifacts: %v\n", repoMatch[0])
						results[repoMatch[0]] = repoMatch[0]
					}
				}
			}

		}
	}

	return results
}

func hasSuffixInList(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// TODO(yawenluo): combine this function with that in maven-auth module.
type response struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func getAuthenticationToken(c *metadata.Client) (string, error) {
	// get access token to call gcp api's with, can pass scopes as an query param
	accessTokenJSON, err := c.Get(tokenPathSuffix)
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
