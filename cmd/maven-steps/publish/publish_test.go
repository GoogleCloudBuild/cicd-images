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
package publish

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/cmd/maven-steps/publish/xmlmodules"
)

func TestWriteSettingsXML(t *testing.T) {
	// Define test input data
	repoID := "artifact-registry"
	token := "test_token"
	settingXMLName := "test_setting.xml"

	t.Run("write setting.xml", func(t *testing.T) {
		err := writeSettingsXML(token, repoID, settingXMLName)
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
		if len(settings.Servers.Server) != 1 {
			t.Errorf("%v", settings)
			t.Errorf("Expected %d server, but got %d", 1, len(settings.Servers.Server))
		}
		if settings.Servers.Server[0].ID != repoID {
			t.Errorf("Expected repository ID '%s', but got '%s'", repoID, settings.Servers.Server[0].ID)
		}
		if settings.Servers.Server[0].Username != "oauth2accesstoken" {
			t.Errorf("Expected username 'oauth2accesstoken', but got '%s'", settings.Servers.Server[0].Username)
		}
		if settings.Servers.Server[0].Password != token {
			t.Errorf("Expected password '%s', but got '%s'", token, settings.Servers.Server[0].Password)
		}
		defer os.Remove(settingXMLName)
	})
}
