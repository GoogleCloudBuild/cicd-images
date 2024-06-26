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
	"fmt"
	"time"

	"cloud.google.com/go/appengine/apiv1/appenginepb"

	appYAML "github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/appyaml"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/config"
)

// NewVersion creates a new `appenginepb.Version` struct to be used
// as an argument for CreateVersion.
func NewVersion(appYAML *appYAML.AppYAML, opts config.AppEngineDeployOptions) (*appenginepb.Version, error) {
	if appYAML.Runtime == "" {
		return nil, fmt.Errorf("runtime is required")
	}
	if appYAML.Env != "flex" && appYAML.Env != "flexible" {
		return nil, fmt.Errorf("expected flexible environment, got %s", appYAML.Env)
	}

	version := &appenginepb.Version{
		Env: appYAML.Env,
		Id:  opts.VersionID,
		Deployment: &appenginepb.Deployment{
			Container: &appenginepb.ContainerInfo{
				Image: opts.ImageURL,
			},
		},
		Runtime: appYAML.Runtime,
	}
	return version, nil
}

// ID returns a timestamp string suitable for use as an App Engine version ID.
// The format is YYYYMMDDtHHMMSS, consistent with the default of "gcloud app deploy".
func ID() string {
	return time.Now().UTC().Format("20060102t150405")
}
