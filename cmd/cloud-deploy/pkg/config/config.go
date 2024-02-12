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

// ReleaseConfiguration is used to configure the cloud deploy release.
type ReleaseConfiguration struct {
	DeliveryPipeline string // ex: web-app
	Region           string // ex: us-central1
	ProjectId        string // ex: my-project
	Source           string // ex: .
	Images           string // ex: image1=path/to/image1:v1,image2=path/to/image2:v1
	Release          string // ex: test-release
}
