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

package appyaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAppYAML_ValidYaml(t *testing.T) {
	validYaml := []byte(`
env: flexible
service: my-service
runtime: go122
`)
	expectedAppYAML := &AppYAML{
		Env:     "flexible",
		Service: "my-service",
		Runtime: "go122",
	}

	actualAppYAML, err := ParseAppYAML(validYaml)

	assert.Nil(t, err, "unexpected error while parsing valid yaml")
	assert.Equal(t, expectedAppYAML, actualAppYAML, "parsed app.yaml does not match expected")
}
