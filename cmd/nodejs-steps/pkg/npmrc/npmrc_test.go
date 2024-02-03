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

func TestAddTokenToConfigFile(t *testing.T) {
	cases := []struct {
		name          string
		existingNpmrc string
		wantNpmrc     string
		token         string
	}{
		{
			name: "npmrc scoped with ar registry",
			existingNpmrc: `@catalog:registry=https://us-npm.pkg.dev/some-project/some-repo/
//us-npm.pkg.dev/some-project/some-repo/:always-auth=true`,
			token: "myToken",
			wantNpmrc: `@catalog:registry=https://us-npm.pkg.dev/some-project/some-repo/
//us-npm.pkg.dev/some-project/some-repo/:_authToken=myToken
//us-npm.pkg.dev/some-project/some-repo/:always-auth=true`,
		},
		{
			name: "npmrc no scope with ar registry",
			existingNpmrc: `registry=https://us-npm.pkg.dev/some-project/some-repo/
//us-npm.pkg.dev/some-project/some-repo/:always-auth=true`,
			token: "myToken",
			wantNpmrc: `registry=https://us-npm.pkg.dev/some-project/some-repo/
//us-npm.pkg.dev/some-project/some-repo/:_authToken=myToken
//us-npm.pkg.dev/some-project/some-repo/:always-auth=true`,
		},
		{
			name:          "empty registry",
			existingNpmrc: ``,
			wantNpmrc:     ``,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			npmrc := AddTokenToConfigFile(tc.existingNpmrc, tc.token)

			if npmrc != tc.wantNpmrc {
				t.Errorf("unexpected npmrc: got %q, want %q", npmrc, tc.wantNpmrc)
			}

		})
	}
}
