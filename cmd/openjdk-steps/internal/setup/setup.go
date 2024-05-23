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
package setup

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	xmlmodules "github.com/GoogleCloudBuild/cicd-images/cmd/openjdk-steps/internal/setup/xmlmodules"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/vifraa/gopom"
)

const tokenPathSuffix = "instance/service-accounts/default/token"

var (
	arUrlPrefixRegex         = regexp.MustCompile(`^https:\/\/[a-z0-9\-]+\.pkg\.dev\/.+$`)
	allowedMavenInstallFlags = []string{"-DallowIncompleteProjects", "-DinstallAtEnd", "-Dmaven.install.skip", "--offline", "--quiet", "--debug", "--no-transfer-progress", "--batch-mode", "--fail-fast", "--errors", "-e", "--fail-never"}
)

type response struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// generate an effective-pom.xml file
func GenerateEffectivePom(inputPomFile string, effectivePomPath string) error {
	generateArgs := []string{"help:effective-pom", "--file=" + inputPomFile, "-Doutput=" + effectivePomPath}
	pomCmd := exec.Command("mvn", generateArgs...)
	pomCmd.Stdout = os.Stdout
	pomCmd.Stderr = os.Stderr
	err := pomCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// get access token to call gcp api's with, can pass scopes as an query param
func GetAuthenticationToken(c *metadata.Client) (string, error) {
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

type Projects struct {
	Projects []gopom.Project `xml:"project"`
}

// Manually parsing multi-project pom files (github.com/vifraa/gopom does not handle these)
func pomParse(path string) (*Projects, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
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
func RetrieveRepos(pomPath string) (map[string]string, error) {
	fmt.Printf("POM path: %s\n", pomPath)
	pom, err := gopom.Parse(pomPath)
	repos := map[string]string{}

	if err != nil {
		projects, err := pomParse(pomPath)
		if err != nil {
			return nil, fmt.Errorf("retrieving pom.xml: %w", err)
		}
		for _, project := range projects.Projects {
			retrieveProjectRepos(&project, repos)
		}
	} else {
		retrieveProjectRepos(pom, repos)
	}

	fmt.Printf("POM.XML scan done.\n")
	return repos, nil
}

// Retrieves repositories at single project level.
func retrieveProjectRepos(project *gopom.Project, repos map[string]string) {
	// Get the distribution management repository
	if project.DistributionManagement != nil {
		if project.DistributionManagement.SnapshotRepository != nil && project.DistributionManagement.SnapshotRepository.ID != nil && project.DistributionManagement.SnapshotRepository.URL != nil {
			(repos)[*project.DistributionManagement.SnapshotRepository.ID] = *project.DistributionManagement.SnapshotRepository.URL
		} else {
			log.Printf("Snapshot repository in distribution management does not exist or is missing ID/URL.")
		}

		if project.DistributionManagement.Repository != nil && project.DistributionManagement.Repository.ID != nil && project.DistributionManagement.Repository.URL != nil {
			(repos)[*project.DistributionManagement.Repository.ID] = *project.DistributionManagement.Repository.URL
		} else {
			log.Printf("Repository in distribution management does not exist or is missing ID/URL.")
		}
	} else {
		log.Printf("Distribution management field does not exist.")
	}

	// Get the repositories in "repositories"
	if project.Repositories != nil {
		for _, repo := range *project.Repositories {
			if repo.ID != nil && repo.URL != nil {
				(repos)[*repo.ID] = *repo.URL
			} else {
				log.Printf("Repository in repositories is missing ID or URL.")
			}
		}
	} else {
		log.Println("No repositories exist.")
	}

	// Get the repositories in "pluginRepositories"
	if project.PluginRepositories != nil {
		for _, repo := range *project.PluginRepositories {
			if repo.ID != nil && repo.URL != nil {
				(repos)[*repo.ID] = *repo.URL
			} else {
				log.Printf("Repository in repositories is missing ID or URL.")
			}
		}
	} else {
		log.Println("No pluginRepositories exist.")
	}
}

// create/update setting.xml with auth tokens
func WriteSettingXML(debug bool, settingXMLPath string, localRepository string, repos map[string]string, token string) error {
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
		return nil
	}
	file.Close()
	file, err = os.OpenFile(settingXMLPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("openning settings.xml: %w", err)
	}
	defer file.Close()
	// Write the XML data to a file
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	err = encoder.Encode(settings)
	if err != nil {
		return fmt.Errorf("writing SETTINGS.XML to file: %w", err)
	}

	if debug {
		data, err := os.ReadFile(settingXMLPath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return nil
		}

		fmt.Println(string(data))
		fmt.Println("Debug mode: printing settings.xml")
		fmt.Printf("%v", settings)
	}

	fmt.Println("SETTINGS.XML written to file successfully!")
	return nil

}

// Validate maven install flags provided
func ValidateMavenInstallFlags(flags string) string {
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

// Retrieve setting secret from secretmanager
func RetrieveSettingSecret(secretName string, c *secretmanager.Client, ctx context.Context) ([]byte, error) {

	// Build the secret version request.
	req := secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	// Access the secret version.
	result, err := c.AccessSecretVersion(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("access secret version: %w", err)
	}

	return result.Payload.Data, nil
}

// Update the pom.xml file with the artifact directory
func UpdatePomBuildDirectory(pomPath string, artifactDirectory string) error {
	pom, err := gopom.Parse(pomPath)
	if err != nil {
		return fmt.Errorf("parsing pom.xml file: %w", err)
	}

	// Update the build directory
	pom.Build.Directory = &artifactDirectory

	updatedPom, err := xml.MarshalIndent(pom, "", "  ")
	if err != nil {
		return fmt.Errorf("encode in XML: %w", err)
	}

	if err = os.WriteFile(pomPath, updatedPom, 0444); err != nil {
		return fmt.Errorf("write payload to pom.xml file: %w", err)
	}

	return nil
}
