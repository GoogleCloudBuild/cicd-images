// Copyright 2023 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package npmrc

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		name         string
		config       string
		wantType     string
		wantScope    string
		wantRegistry string
		wantPassword string
		wantToken    string
	}{
		{
			name:     "registry default",
			config:   "myregistry.someProperty=someValue",
			wantType: configType.Default,
		},
		{
			name:         "auth token",
			config:       "//us-west1-npm.pkg.dev/myproj/myrepo/:_authToken=myToken",
			wantType:     configType.AuthToken,
			wantRegistry: "//us-west1-npm.pkg.dev/myproj/myrepo/",
			wantToken:    "myToken",
		},
		{
			name:         "password",
			config:       "//us-west1-npm.pkg.dev/myproj/myrepo/:_password=myPassword",
			wantType:     configType.Password,
			wantRegistry: "//us-west1-npm.pkg.dev/myproj/myrepo/",
			wantPassword: "myPassword",
		},
		{
			name:         "registry assigned",
			config:       "registry=https://us-west1-npm.pkg.dev/myproj/myrepo/",
			wantType:     configType.Registry,
			wantRegistry: "//us-west1-npm.pkg.dev/myproj/myrepo/",
		},
		{
			name:         "scoped registry",
			config:       "@myscope:registry=https://us-west1-npm.pkg.dev/myproj/myrepo/",
			wantType:     configType.Registry,
			wantRegistry: "//us-west1-npm.pkg.dev/myproj/myrepo/",
			wantScope:    "@myscope",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, config := parseConfig(tc.config)
			if got != tc.config {
				t.Errorf("unexpected netrc: got %v, want %v", got, tc.config)
			}
			if config.Password != tc.wantPassword {
				t.Errorf("unexpected Password: got %v, want %v", config.Password, tc.wantPassword)
			}
			if config.Registry != tc.wantRegistry {
				t.Errorf("unexpected Registry: got %v, want %v", config.Registry, tc.wantRegistry)
			}
			if config.Scope != tc.wantScope {
				t.Errorf("unexpected Scope: got %v, want %v", config.Scope, tc.wantScope)
			}
			if config.Token != tc.wantToken {
				t.Errorf("unexpected Token: got %v, want %v", config.Token, tc.wantToken)
			}
			if config.Type != tc.wantType {
				t.Errorf("unexpected Type: got %v, want %v", config.Type, tc.wantType)
			}
		})
	}
}
