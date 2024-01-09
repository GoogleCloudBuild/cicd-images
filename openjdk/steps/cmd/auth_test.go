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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/openjdk/steps/internal/xmlmodules"
)

func TestRetrieveRepos(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		arg         string
		expected    map[string]string
		expectedErr error
	}{
		{
			name: "valid pom.xml",
			arg:  "./testdata/pom-test.xml",
			expected: map[string]string{
				"valid-artifact-registry":    "https://us-central1-maven.pkg.dev/random-project/random-repo",
				"valid-artifact-registry2":   "https://us-central1-maven.pkg.dev/random-project2/random-repo2",
				"invalid-artifact-registry3": "",
				"valid-artifact-registry4":   "https://us-central1-maven.pkg.dev/random-project4/random-repo4",
				"valid-artifact-registry5":   "https://rampdom-repository/random-project5/random-repo5",
			},
			expectedErr: nil,
		},
		{
			name:        "pom.xml nonexistent",
			arg:         "./testdata/nonexistent.xml",
			expected:    nil,
			expectedErr: fmt.Errorf("Error: retrieving pom.xml: open ./testdata/nonexistent.xml: no such file or directory"),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call function
			result, err := retrieveRepos(tc.arg)

			if err != nil {
				if tc.expectedErr == nil {
					t.Errorf("Expected no errors, but got %v", err)
				} else if tc.expectedErr != nil && reflect.DeepEqual(err.Error(), tc.expectedErr.Error()) {
					t.Errorf("Expected %v, but got %v", tc.expectedErr, err)
				}
			} else if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

func TestPomParse(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		arg         string
		expected    map[string]string
		expectedErr error
	}{
		{
			name: "valid multi pom.xml",
			arg:  "./testdata/pom-test-multi-module.xml",
			expected: map[string]string{
				"artifact-registry":  "https://us-central1-maven.pkg.dev/gcb-catalog-bugbash/java-bugbash-repo",
				"artifact-registry2": "https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo",
				"central":            "https://repo.maven.apache.org/maven2",
			},
			expectedErr: nil,
		},
		{
			name:        "pom.xml nonexistent",
			arg:         "./testdata/nonexistent.xml",
			expected:    nil,
			expectedErr: fmt.Errorf("Error: retrieving pom.xml: open ./testdata/nonexistent.xml: no such file or directory"),
		},
		{
			name: "single layer pom.xml",
			arg:  "./testdata/pom-test.xml",
			expected: map[string]string{
				"valid-artifact-registry":    "https://us-central1-maven.pkg.dev/random-project/random-repo",
				"valid-artifact-registry2":   "https://us-central1-maven.pkg.dev/random-project2/random-repo2",
				"invalid-artifact-registry3": "",
				"valid-artifact-registry4":   "https://us-central1-maven.pkg.dev/random-project4/random-repo4",
				"valid-artifact-registry5":   "https://rampdom-repository/random-project5/random-repo5",
			},
			expectedErr: nil,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call function
			result, err := retrieveRepos(tc.arg)

			if err != nil {
				if tc.expectedErr == nil {
					t.Errorf("Expected no errors, but got %v", err)
				} else if tc.expectedErr != nil && reflect.DeepEqual(err.Error(), tc.expectedErr.Error()) {
					t.Errorf("Expected %v, but got %v", tc.expectedErr, err)
				}
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
	token := "test_token"

	t.Run("write setting.xml", func(t *testing.T) {
		// Call the function being tested
		writeSettingXML(settingXMLName, repos, token)

		// Read the output file and unmarshal the XML data into a struct
		data, err := ioutil.ReadFile(settingXMLName)
		if err != nil {
			t.Fatalf("Error reading output file: %v", err)
		}
		var settings xmlmodules.Settings
		err = xml.Unmarshal(data, &settings)
		if err != nil {
			t.Fatalf("Error unmarshaling XML data: %v", err)
		}

		// Check that the output struct contains the expected data
		if len(settings.Servers.Server) != 2 {
			t.Fatalf("%v", settings)
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
	tmpfile, err := ioutil.TempFile("", settingXMLName)
	if err != nil {
		t.Fatalf("Error creating temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Define test input data
	repos := map[string]string{
		"repo1": "https://us-central1-maven.pkg.dev/random-project/random-repo",
		"repo2": "https://us-central1-maven.pkg.dev/random-project/random-repo2",
		"repo3": "https://repo3.example.com",
		"repo4": "https://repo4.example.com",
	}
	token := "test_token"

	t.Run("write setting.xml", func(t *testing.T) {
		// Call the function being tested
		writeSettingXML(settingXMLName, repos, token)

		// Read the output file and unmarshal the XML data into a struct
		data, err := ioutil.ReadFile(tmpfile.Name())
		if err != nil {
			t.Fatalf("Error reading output file: %v", err)
		}
		if len(string(data)) != 0 {
			t.Fatalf("No data should be writing to an existing settings.xml")
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
			result := validateMavenInstallFlags(tc.flags)
			if result != tc.expected {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}
