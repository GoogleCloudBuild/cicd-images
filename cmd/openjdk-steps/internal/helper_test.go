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
	"encoding/xml"
	"os"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/cmd/openjdk-steps/internal/xmlmodules"
)

func TestWriteSettingXML(t *testing.T) {
	// Define test input data
	repos := []string{
		"https://us-central1-maven.pkg.dev/random-project/random-repo",
		"https://us-central1-maven.pkg.dev/random-project/random-repo2",
	}
	localRepository := "~/.m2/repository"
	token := "test_token"
	settingXMLName := "test_setting.xml"

	t.Run("write setting.xml", func(t *testing.T) {
		err := WriteSettingsXML(token, localRepository, repos, settingXMLName)
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
		if len(settings.Servers.Server) != len(repos) {
			t.Errorf("%v", settings)
			t.Errorf("Expected %d servers, but got %d", len(repos), len(settings.Servers.Server))
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
	repos := []string{
		"https://us-central1-maven.pkg.dev/random-project/random-repo",
		"https://us-central1-maven.pkg.dev/random-project/random-repo2",
	}
	localRepository := "~/.m2/repository"
	token := "test_token"

	t.Run("write setting.xml", func(t *testing.T) {
		// Call the function being tested
		err := WriteSettingsXML(token, localRepository, repos, settingXMLName)
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
