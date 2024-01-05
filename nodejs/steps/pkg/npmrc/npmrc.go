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

// Package netrc provides functions to modify an netrc file.
package npmrc

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Load loads the path and contents of the .npmrc file into memory.
func Load(npmrcPath string) (string, error) {
	if npmrcPath == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot load .npmrc file: %v", err)
		}
		npmrcPath = h
	}

	data, err := ioutil.ReadFile(npmrcPath)
	if err != nil {
		return "", fmt.Errorf("cannot load .npmrc file: %v", err)
	}
	return string(data), nil
}

// Save saves the .npmrc file.
func Save(npmrc, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("Save: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(npmrc); err != nil {
		return fmt.Errorf("Save: %v", err)
	}
	return nil
}

// AddTokenToConfigFile appends the token with a registry to the npmrc file
func AddTokenToConfigFile(npmrc, creds string) string {
	toConfigs := []string{}

	for _, line := range strings.Split(string(npmrc), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matchedLine, cfg := parseConfig(line)

		toConfigs = append(toConfigs, matchedLine)
		if cfg.Type == configType.Registry {
			toConfigs = append(toConfigs, fmt.Sprintf("%s:_authToken=%s", cfg.Registry, creds))
		}
	}

	return strings.Join(toConfigs, "\n")
}
