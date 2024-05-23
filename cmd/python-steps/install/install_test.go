// Copyright 2023 Google LLC
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
package install

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/auth"
	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/command"
)

func TestInstallDependency(t *testing.T) {
	dep := "foo"
	flags := []string{"--index-url=" + IndexURL}
	runner := &command.MockCommandRunner{
		ExpectedCommands: []string{command.VirtualEnvPip, "install", dep, "--index-url=" + IndexURL},
	}
	if err := installDependency(runner, dep, flags...); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuthenticateRegistryAndGetFlags(t *testing.T) {
	registryURL := "https://foo-registry.com"
	expectedFlags := []string{"--index-url=https://pypi.org/simple/", "--extra-index-url=https://oauth2accesstoken:foo-token@foo-registry.com/simple"}
	client := &auth.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "foo-token"}`))),
			}, nil

		},
	}

	flags, err := authenticateRegistryAndGetFlags(registryURL, client)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(flags, expectedFlags) {
		t.Errorf("unexpected flags: got %v, expected %v", flags, expectedFlags)
	}
}

func TestInstallFromRequirementsFile(t *testing.T) {
	requirementsPath := "requirements.txt"
	flags := []string{"--index-url=" + IndexURL}
	runner := &command.MockCommandRunner{
		ExpectedCommands: []string{command.VirtualEnvPip, "install", "-r", requirementsPath, "--index-url=" + IndexURL},
	}
	if err := installFromRequirementsFile(runner, requirementsPath, flags); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
