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

// Package appyaml provides utilities for parsing and working with app.yaml files
// for Google App Engine flexible environment deployments.
package appyaml

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

// AppYAML represents the structure of an app.yaml file.
// See https://cloud.google.com/appengine/docs/flexible/reference/app-yaml
type AppYAML struct {
	Runtime string `yaml:"runtime"`
	Service string `yaml:"service"`
	Env     string `yaml:"env"`
}

// ParseAppYAML parses the provided app.yaml data (in byte format)
// and returns a pointer to an `AppYAML` struct containing the parsed data.
func ParseAppYAML(data []byte) (*AppYAML, error) {
	appYAML := &AppYAML{}
	err := yaml.Unmarshal(data, appYAML)
	if err != nil {
		return nil, fmt.Errorf("error parsing app.yaml: %w", err)
	}
	return appYAML, nil
}
