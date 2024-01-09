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

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/spf13/cobra"
)

const (
	uploadPattern             = `Uploaded to [^\s]+: (https:\/\/[^ ]+)`
	repositoryPattern         = `https?:\/\/[^\s]+`
	logPathDefault            = "./deploy.txt"
	outputPathDefault         = "/mavenconfigs/artifacts.list"
	artifactsDirectoryDefault = "./target"
)

var (
	logPath             string
	mavenPackagingTypes = []string{"jar", "ejb", "war", "ear", "rar"}
	outputPath          string
	artifactsDirectory  string
)

func init() {
	deployScanCmd.Flags().StringVar(&logPath, "logPath", logPathDefault, "Path of output log of deploy plugin")
	deployScanCmd.Flags().StringVar(&outputPath, "outputPath", outputPathDefault, "Path of retrieved artifacts output")
}

var deployScanCmd = &cobra.Command{
	Use: "deploy-scan",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read the file contents
		content, err := ioutil.ReadFile(logPath)
		if err != nil {
			return err
		}

		// Convert the file content to a string
		logs := string(content)

		artifacts := retrieveArtifacts(logs)
		// Create or open the file for writing
		file, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer file.Close() // Ensure the file is closed when we're done with it

		checksums, err := generateCheckSum(artifacts)
		if err != nil {
			return err
		}
		// Write each string to the file
		for uri, digest := range checksums {
			str := fmt.Sprintf("%s %s\n", uri, digest)
			fmt.Printf("Write: %s\n", str)
			_, err := fmt.Fprintf(file, fmt.Sprintf("%s %s\n", uri, digest))
			if err != nil {
				return err
			}
		}

		fmt.Printf("Strings written to %s\n", outputPath)
		return nil
	},
}

func generateCheckSum(artifacts map[string]string) (map[string]string, error) {
	checksums := map[string]string{}

	c := metadata.NewClient(&http.Client{})
	token, err := getAuthenticationToken(c)
	if err != nil {
		return nil, err
	}
	fmt.Println("Token retrieval success!")

	for uri, _ := range artifacts {
		// TODO: Should be using a passed context
		req, err := http.NewRequest("GET", fmt.Sprintf("%s.sha256", uri), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error: Could not download checksum: %s", err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return nil, fmt.Errorf("error: Could not read checksum: %s", err)
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
