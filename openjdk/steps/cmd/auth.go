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
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/GoogleCloudBuild/cicd-images/openjdk/steps/internal/xmlmodules"
	"github.com/spf13/cobra"
	"github.com/vifraa/gopom"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
)

const (
	customizedSettings = "/mavenconfigs/settings.xml"
)

var (
	debug                    string
	pomPath                  string
	settingPath              string
	localRepository          string
	installFlags             string
	settingSecretName        string
	arUrlPrefixRegex         = regexp.MustCompile(`^https:\/\/[a-z0-9\-]+\.pkg\.dev\/.+$`)
	allowedMavenInstallFlags = []string{"-DallowIncompleteProjects", "-DinstallAtEnd", "-Dmaven.install.skip", "--offline", "--quiet", "--debug", "--no-transfer-progress", "--batch-mode", "--fail-fast", "--errors", "-e", "--fail-never"}
)

func init() {
	authCmd.Flags().StringVar(&debug, "debug", "no", "Whether to turn on debug mode")
	authCmd.Flags().StringVar(&pomPath, "pomPath", "./pom.xml", "Location of pom.xml")
	authCmd.Flags().StringVar(&settingPath, "settingPath", customizedSettings, "Location of setting.xml")
	authCmd.Flags().StringVar(&localRepository, "localRepository", "~/.m2/repository", "Location of local repository")
	authCmd.Flags().StringVar(&installFlags, "installFlags", "", "Flags being passed to maven-install Task")
	authCmd.Flags().StringVar(&settingSecretName, "settingSecretName", "", "Secret name in Secret Manager to store customized settings.xml. Format: projects/*/secrets/*")
}

var authCmd = &cobra.Command{
	Use: "auth",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Local repository is: %s\n", localRepository)

		c := metadata.NewClient(&http.Client{})

		token, err := getAuthenticationToken(c)
		if err != nil {
			return err
		}
		fmt.Println("Token retrieval success!")

		sc, err := secretmanager.NewClient(cmd.Context())
		if err != nil {
			return err
		}
		defer sc.Close()

		if settingSecretName != "" {
			fmt.Printf("Downloading custom settings from Secret Manager...")
			data, err := retrieveSettingSecret(settingSecretName, sc, cmd.Context())
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(customizedSettings, data, 0644)
			if err != nil {
				return err
			}

			fmt.Printf("Secret payload saved to %s\n", customizedSettings)
		} else {
			repos, err := retrieveRepos(pomPath)
			if err != nil {
				return err
			}
			fmt.Printf("Main repos:\n%v\n", repos)
			err = writeSettingXML(settingPath, repos, token)
			if err != nil {
				return err
			}
		}
		if installFlags != "" {
			unallowedFlag := validateMavenInstallFlags(installFlags)
			if unallowedFlag != "" {
				return fmt.Errorf("Unallowed flags for maven-install Task: %v", unallowedFlag)
			}
		}
		return nil
	},
}

type Projects struct {
	Projects []gopom.Project `xml:"project"`
}

func PomParse(path string) (*Projects, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var projects Projects

	err = xml.Unmarshal(b, &projects)
	if err != nil {
		return nil, err
	}
	return &projects, nil
}

// Retrieves repositories at multi project level.
func retrieveRepos(pomPath string) (map[string]string, error) {
	fmt.Printf("POM path: %s\n", pomPath)
	pom, err := gopom.Parse(pomPath)
	repos := map[string]string{}

	if err != nil {
		projects, err := PomParse(pomPath)
		if err != nil {
			return nil, fmt.Errorf("Error: retrieving pom.xml: %v\n", err)
		}
		for _, project := range projects.Projects {
			retrieveProjectRepos(&project, repos)
		}
	} else {
		retrieveProjectRepos(pom, repos)
	}

	fmt.Printf("POM.XML scan done.")
	return repos, nil
}

// Retrieves repositories at single project level.
func retrieveProjectRepos(project *gopom.Project, repos map[string]string) {
	// Get the distribution management repository
	(repos)[project.DistributionManagement.SnapshotRepository.ID] = project.DistributionManagement.SnapshotRepository.URL
	(repos)[project.DistributionManagement.Repository.ID] = project.DistributionManagement.Repository.URL

	// Get the repositories in "repositories"
	for _, repo := range project.Repositories {
		(repos)[repo.ID] = repo.URL
	}

	// Get the repositories in "pluginRepositories"
	for _, repo := range project.PluginRepositories {
		(repos)[repo.ID] = repo.URL
	}
}

func writeSettingXML(settingXMLPath string, repos map[string]string, token string) error {
	settings := xmlmodules.Settings{
		Servers:         xmlmodules.Servers{},
		LocalRepository: xmlmodules.LocalRepository{Path: localRepository},
	}

	for repoId, url := range repos {
		if arUrlPrefixRegex.MatchString(url) {
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
	}

	file, err := os.Open(settingXMLPath) // For read access.
	if err == nil {
		// Users have provided a settings.xml, we skip writting auth tokens.
		file.Close()
		return err
	}
	file.Close()
	file, err = os.OpenFile(settingXMLPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	// Write the XML data to a file
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	err = encoder.Encode(settings)
	if err != nil {
		return err
	}

	if debug == "yes" {
		data, err := ioutil.ReadFile(settingXMLPath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return err
		}

		fmt.Println(string(data))
		fmt.Println("Debug mode: printing settings.xml")
		fmt.Printf("%v", settings)
	}

	fmt.Println("SETTING.XML written to file successfully!")
	return nil
}

func validateMavenInstallFlags(flags string) string {
	installFlags := strings.Fields(flags)
	for _, flag := range installFlags {
		found := false
		for _, allowedFlag := range allowedMavenInstallFlags {
			if flag == allowedFlag {
				found = true
			}

		}
		if found == false {
			return flag
		}

	}
	return ""
}

func retrieveSettingSecret(secretName string, c *secretmanager.Client, ctx context.Context) ([]byte, error) {

	// Build the secret version request.
	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	// Access the secret version.
	result, err := c.AccessSecretVersion(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("Failed to access secret version: %v\n", err)
	}

	return result.Payload.Data, nil
}
