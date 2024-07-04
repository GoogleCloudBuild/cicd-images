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
package publish

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// MockCommandRunner is a mock implementation of CommandRunner.
type mockCommandRunner struct {
	CalledCommands   []string
	ExpectedCommands []string
	Err              error
}

func (m *mockCommandRunner) Run(cmd string, args ...string) error {
	m.CalledCommands = append(m.CalledCommands, cmd)
	m.CalledCommands = append(m.CalledCommands, args...)

	if m.Err != nil {
		return m.Err
	}

	if !reflect.DeepEqual(m.CalledCommands, m.ExpectedCommands) {
		return fmt.Errorf("unexpected commands: got %v, expected %v", m.CalledCommands, m.ExpectedCommands)
	}

	return nil
}

func TestPushPythonArtifact(t *testing.T) {
	runner := &mockCommandRunner{
		ExpectedCommands: []string{"python3", "-m", "twine", "upload", "--repository-url", "foo-url", "--username", "oauth2accesstoken", "--password", "foo-token", "dist/*"},
	}
	args := Arguments{
		ArtifactRegistryURL: "foo-url",
		ArtifactDir:         "dist",
	}
	gcpToken := "foo-token"

	if err := pushPythonArtifact(runner, args, gcpToken); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateUri(t *testing.T) {
	testCases := []struct {
		name                string
		artifactRegistryURL string
		filename            string
		expectedUri         string
		expectedError       string // Partial error string to check if error occurs
	}{
		{
			name:                "Successful for source distribution",
			artifactRegistryURL: "http://foo.com/ar-repo",
			filename:            "foo-8.0.0.tar.gz",
			expectedUri:         "pkg:foo.com/ar-repo/foo@8.0.0",
		},
		{
			name:                "Successful for wheel distribution",
			artifactRegistryURL: "http://foo.com/ar-repo",
			filename:            "foo-8.0.0.whl",
			expectedUri:         "pkg:foo.com/ar-repo/foo@8.0.0",
		},
		{
			name:                "Failed for unknown distribution",
			artifactRegistryURL: "http://foo.com/ar-repo",
			filename:            "foo-8.0.0.egg",
			expectedError:       "invalid filename format",
		},
		{
			name:                "Failed to parse artifact registry url",
			artifactRegistryURL: "http://foo.com/ar-repo%",
			filename:            "foo-8.0.0.tar.gz",
			expectedError:       "error parsing artifactRegistryURL",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			uri, err := generateURI(test.artifactRegistryURL, test.filename)

			if test.expectedError != "" {
				if err == nil {
					t.Fatal("expected an error, but got nil")
				} else if !strings.Contains(err.Error(), test.expectedError) {
					t.Fatalf("error message mismatch: got '%s', expected to contain '%s'", err.Error(), test.expectedError)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if uri != test.expectedUri {
					t.Errorf("uri mismatch: got '%s', expected '%s'", uri, test.expectedUri)
				}
			}
		})
	}
}
