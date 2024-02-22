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
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"testing"

	xmlmodules "github.com/GoogleCloudBuild/cicd-images/cmd/openjdk-steps/internal/setup/xmlmodules"
	"github.com/vifraa/gopom"
)

func TestRetrieveRepos(t *testing.T) {
	testFile := "./testdata/pom-test.xml"

	// should retrieve repositories that have id AND url from the pom file with format {id: url}
	expectedResult := map[string]string{
		"valid-artifact-registry-snapshot":                "https://us-central1-maven.pkg.dev/dummy-project/dummy-repo-url",
		"valid-artifact-registry-distribution-management": "https://us-central1-maven.pkg.dev/dummy-project2/dummy-repo-url2",
		"valid-artifact-registry":                         "https://us-central1-maven.pkg.dev/dummy-project3/dummy-repo-url3",
	}
	artifactDirectory := "./target"

	tmpdir := t.TempDir()
	fileName := createTempTestingFile(t, tmpdir, testFile)

	t.Run("valid pom.xml", func(t *testing.T) {
		result, err := RetrieveRepos(fileName, artifactDirectory)
		if err != nil {
			t.Errorf("Expected no errors, but got %v", err)
		} else if !reflect.DeepEqual(result, expectedResult) {
			t.Errorf("Expected %v, but got %v", expectedResult, result)
		}
	})

}

func TestRetrieveReposNonexistent(t *testing.T) {
	fileName := "./testdata/nonexistent.xml"
	expectedErr := fmt.Errorf("Error: retrieving pom.xml: open ./testdata/nonexistent.xml: no such file or directory")
	artifactDirectory := "./target"

	t.Run("pom.xml nonexistent", func(t *testing.T) {
		_, err := RetrieveRepos(fileName, artifactDirectory)
		if reflect.DeepEqual(err.Error(), expectedErr.Error()) {
			t.Errorf("Expected %v, but got %v", expectedErr, err)
		}
	})

}

func TestPomParse(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		arg      string
		expected map[string]string
	}{
		{
			name: "multi pom.xml",
			arg:  "./testdata/pom-test-multi-module.xml",
			expected: map[string]string{
				"artifact-registry":  "https://us-central1-maven.pkg.dev/dummy-project/dummy-repo-url",
				"artifact-registry2": "https://us-central1-maven.pkg.dev/dummy-project2/dummy-quickstart-repo-url",
				"central":            "https://repo.maven.apache.org/maven2",
			},
		},
		{
			name: "single layer pom.xml",
			arg:  "./testdata/pom-test.xml",
			expected: map[string]string{
				"valid-artifact-registry-snapshot":                "https://us-central1-maven.pkg.dev/dummy-project/dummy-repo-url",
				"valid-artifact-registry-distribution-management": "https://us-central1-maven.pkg.dev/dummy-project2/dummy-repo-url2",
				"valid-artifact-registry":                         "https://us-central1-maven.pkg.dev/dummy-project3/dummy-repo-url3",
			},
		},
	}
	artifactDirectory := "./target"

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpdir := t.TempDir()
			fileName := createTempTestingFile(t, tmpdir, tc.arg)

			result, err := RetrieveRepos(fileName, artifactDirectory)
			if err != nil {
				t.Errorf("Expected no errors, but got %v", err)
			} else if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

func TestWriteSettingXML(t *testing.T) {
	// Create a temporary file for testing
	settingXMLName := "test_setting.xml"

	// Define test input data
	repos := map[string]string{
		"repo1": "https://us-central1-maven.pkg.dev/random-project/random-repo",
		"repo2": "https://us-central1-maven.pkg.dev/random-project/random-repo2",
		"repo3": "https://repo3.example.com",
		"repo4": "https://repo4.example.com",
	}
	debug := false
	localRepository := "~/.m2/repository"
	token := "test_token"

	t.Run("write setting.xml", func(t *testing.T) {
		// Call the function being tested
		err := WriteSettingXML(debug, settingXMLName, localRepository, repos, token)
		if err != nil {
			t.Errorf("Error in writeSettingXML: %v", err)
		}

		// Read the output file and unmarshal the XML data into a struct
		data, err := os.ReadFile(settingXMLName)
		if err != nil {
			t.Errorf("Error reading output file: %v", err)
		}
		var settings xmlmodules.Settings
		err = xml.Unmarshal(data, &settings)
		if err != nil {
			t.Errorf("Error unmarshaling XML data: %v", err)
		}

		// Check that the output struct contains the expected data
		if len(settings.Servers.Server) != 2 {
			t.Errorf("%v", settings)
			t.Errorf("Expected 2 servers, but got %d", len(settings.Servers.Server))
		}
		for _, server := range settings.Servers.Server {
			if server.Username != "oauth2accesstoken" {
				t.Errorf("Expected username 'oauth2accesstoken', but got '%s'", server.Username)
			}
			if server.Password != token {
				t.Errorf("Expected password '%s', but got '%s'", token, server.Password)
			}
		}
		if len(settings.LocalRepository.Path) <= 0 {
			t.Errorf("Expected localRepository '%s', but got '%s'", "~/.m2/repository", settings.LocalRepository.Path)
		}
		defer os.Remove(settingXMLName)
	})
}

func TestWriteSettingXMLExist(t *testing.T) {
	// Create a temporary file for testing
	settingXMLName := "test_setting.xml"
	tmpfile, err := os.CreateTemp("", settingXMLName)
	if err != nil {
		t.Errorf("Error creating temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Define test input data
	repos := map[string]string{
		"repo1": "https://us-central1-maven.pkg.dev/random-project/random-repo",
		"repo2": "https://us-central1-maven.pkg.dev/random-project/random-repo2",
		"repo3": "https://repo3.example.com",
		"repo4": "https://repo4.example.com",
	}
	debug := false
	localRepository := "~/.m2/repository"
	token := "test_token"

	t.Run("write setting.xml", func(t *testing.T) {
		// Call the function being tested
		err := WriteSettingXML(debug, settingXMLName, localRepository, repos, token)
		if err != nil {
			t.Errorf("Error in writeSettingXML: %v", err)
		}

		// Read the output file and unmarshal the XML data into a struct
		data, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			t.Errorf("Error reading output file: %v", err)
		}
		if len(string(data)) != 0 {
			t.Errorf("No data should be writing to an existing settings.xml")
		}
		defer os.Remove(settingXMLName)

	})
}

func TestValidateMavenInstallFlags(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		flags    string
		expected string
	}{
		{
			name:     "simple flags",
			flags:    "-e",
			expected: "",
		},
		{
			name:     "all allowed flags except one unknown flags",
			flags:    "-DallowIncompleteProjects -DinstallAtEnd -Dmaven.install.skip --offline --quiet --debug --no-transfer-progress --batch-mode --fail-fast --errors -e --fail-never unknownflags",
			expected: "unknownflags",
		},
		{
			name:     "disallow special symbol",
			flags:    "$aaa",
			expected: "$aaa",
		},
		{
			name:     "disallow complex special symbol",
			flags:    "'$aaa'",
			expected: "'$aaa'",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call function
			result := ValidateMavenInstallFlags(tc.flags)
			if result != tc.expected {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

func TestUpdatePomSingle(t *testing.T) {
	// Define test case
	pomFile := "./testdata/pom-test.xml"
	artifactDirectory := "./testDirectory"

	tmpdir := t.TempDir()
	fileName := createTempTestingFile(t, tmpdir, pomFile)

	t.Run("single project pom.xml", func(t *testing.T) {
		_, err := RetrieveRepos(fileName, artifactDirectory)

		if err != nil {
			t.Errorf("Expected no errors, but got %v", err)
		} else {
			// Read the pom file and unmarshal the XML data into a struct
			project, err := gopom.Parse(fileName)
			if err != nil {
				t.Errorf("Error parsing pom file: %v", err)
			}
			if *project.Build.Directory != artifactDirectory {
				t.Errorf("Expected %v, but got %v", artifactDirectory, project.Build.Directory)
			}

		}
	})

}

func TestUpdatePomMultiple(t *testing.T) {
	// Define test case
	pomFile := "./testdata/pom-test-multi-module.xml"
	artifactDirectory := "./testDirectory"

	type Projects struct {
		Projects []gopom.Project `xml:"project"`
	}

	tmpdir := t.TempDir()
	fileName := createTempTestingFile(t, tmpdir, pomFile)

	t.Run("multi-project pom.xml", func(t *testing.T) {
		_, err := RetrieveRepos(fileName, artifactDirectory)
		if err != nil {
			t.Errorf("Expected no errors, but got %v", err)
		} else {
			// Read the pom file and unmarshal the XML data into a struct
			content, err := os.ReadFile(fileName)
			if err != nil {
				t.Errorf("Error reading temp pom file: %v", err)
			}

			var projects Projects
			err = xml.Unmarshal(content, &projects)
			if err != nil {
				t.Errorf("Error parsing xml data: %v", err)
			}
			for _, project := range projects.Projects {
				if *project.Build.Directory != artifactDirectory {
					t.Errorf("Expected %v, but got %v", artifactDirectory, project.Build.Directory)
				}
			}
		}
	})
}

// Helper function to create copy of the testing files
func createTempTestingFile(t *testing.T, dir string, testFile string) string {
	t.Helper()

	// Create temporary file for testing
	tmpfile, err := os.CreateTemp(dir, "*")
	if err != nil {
		t.Errorf("Error creating temporary file: %v", err)
	}
	defer tmpfile.Close()

	// Copy data from pom file to temp file
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Error reading pom file: %v", err)
	}
	if _, err := tmpfile.Write(content); err != nil {
		t.Errorf("Error writing temp pom file: %v", err)
	}

	return tmpfile.Name()
}
