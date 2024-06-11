// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"

	appYAML "github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/appyaml"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/config"
)

func TestNewVersion(t *testing.T) {
	testCases := []struct {
		name        string
		appYAML     *appYAML.AppYAML
		opts        config.AppEngineDeployOptions
		expectedErr string
	}{
		{
			name: "Success",
			appYAML: &appYAML.AppYAML{
				Runtime: "go122",
				Env:     "flex",
			},
			opts: config.AppEngineDeployOptions{
				ImageURL:  "us-docker.pkg.dev/cloudrun/container/hello",
				VersionID: "test-version",
			},
			expectedErr: "",
		},
		{
			name: "MissingRuntime",
			appYAML: &appYAML.AppYAML{
				Env: "flex",
			},
			opts: config.AppEngineDeployOptions{
				ImageURL:  "us-docker.pkg.dev/cloudrun/container/hello",
				VersionID: "test-version",
			},
			expectedErr: "runtime is required in app.yaml",
		},
		{
			name: "InvalidEnv",
			appYAML: &appYAML.AppYAML{
				Runtime: "go122",
				Env:     "standard", // Invalid env
			},
			opts: config.AppEngineDeployOptions{
				ImageURL:  "us-docker.pkg.dev/cloudrun/container/hello",
				VersionID: "test-version",
			},
			expectedErr: "expected flexible environment, got standard",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version, err := NewVersion(tc.appYAML, tc.opts)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, version)
				assert.Equal(t, "flex", version.Env)
				assert.Equal(t, tc.opts.VersionID, version.Id)
				assert.Equal(t, tc.opts.ImageURL, version.Deployment.Container.Image)
				assert.Equal(t, tc.appYAML.Runtime, version.Runtime)
			}
		})
	}
}
