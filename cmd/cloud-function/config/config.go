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

package config

// DeployOptions is used to configure the cloud function.
type DeployOptions struct {
	Name        string            // The name of the function
	Description string            // The description of the function
	Region      string            // ex: us-central1
	ProjectID   string            // ex: my-project
	Labels      map[string]string // List of key-value pairs to set as function labels in the form label1=VALUE1,label2=VALUE2
	Runtime     string            // The runtime in which to run the function
	EntryPoint  string            // The name of a Google Cloud Function (as defined in source code) that will be executed
	Source      string            // Location of source code to deploy. Can be one Google Cloud Storage, Google Source Repository or a Local filesystem path
}
