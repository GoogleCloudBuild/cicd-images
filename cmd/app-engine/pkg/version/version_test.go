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

	"cloud.google.com/go/appengine/apiv1/appenginepb"
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
		want        *appenginepb.Version
	}{
		{
			name: "SuccessWithImage",
			appYAML: &appYAML.AppYAML{
				Runtime: "go122",
				Env:     "flex",
			},
			opts: config.AppEngineDeployOptions{
				ImageURL:  "us-docker.pkg.dev/appengine/container/hello",
				VersionID: "test-version",
			},
			expectedErr: "",
			want: &appenginepb.Version{
				Env:     "flex",
				Id:      "test-version",
				Runtime: "go122",
				Deployment: &appenginepb.Deployment{
					Container: &appenginepb.ContainerInfo{
						Image: "us-docker.pkg.dev/appengine/container/hello",
					},
				},
			},
		},
		{
			name: "SuccessWithSource",
			appYAML: &appYAML.AppYAML{
				Runtime: "go122",
				Env:     "flex",
				Service: "testservice",
			},
			opts: config.AppEngineDeployOptions{
				SourceURL:   "gs://test-bucket/test-source.zip",
				VersionID:   "test-version",
				AppYAMLPath: "app.yaml",
			},
			expectedErr: "",
			want: &appenginepb.Version{
				Env:     "flex",
				Id:      "test-version",
				Runtime: "go122",
				Deployment: &appenginepb.Deployment{
					Zip: &appenginepb.ZipInfo{
						SourceUrl: "gs://test-bucket/test-source.zip",
					},
					CloudBuildOptions: &appenginepb.CloudBuildOptions{
						AppYamlPath: "app.yaml",
					},
				},
			},
		},
		{
			name: "MissingRuntime",
			appYAML: &appYAML.AppYAML{
				Env: "flex",
			},
			opts: config.AppEngineDeployOptions{
				ImageURL:  "us-docker.pkg.dev/appengine/container/hello",
				VersionID: "test-version",
			},
			expectedErr: "runtime is required",
		},
		{
			name: "InvalidEnv",
			appYAML: &appYAML.AppYAML{
				Runtime: "go122",
				Env:     "standard", // Invalid env
			},
			opts: config.AppEngineDeployOptions{
				ImageURL:  "us-docker.pkg.dev/appengine/container/hello",
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
				assert.Equal(t, tc.want, version)
			}
		})
	}
}
