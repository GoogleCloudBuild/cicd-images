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
	"os"
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
	testCases := []struct {
		name            string
		path            string
		expectedCommand []string
	}{
		{
			name:            "Successful for source distribution",
			path:            "dist/my-pkg.tar.gz",
			expectedCommand: []string{"python3", "-m", "twine", "upload", "--repository-url", "foo-url", "--username", "oauth2accesstoken", "--password", "foo-token", "dist/my-pkg.tar.gz"},
		},
		{
			name:            "Successful for wheel distribution",
			path:            "dist/my-pkg.whl",
			expectedCommand: []string{"python3", "-m", "twine", "upload", "--repository-url", "foo-url", "--username", "oauth2accesstoken", "--password", "foo-token", "dist/my-pkg.whl"},
		},
	}
	args := Arguments{
		Repository: "foo-url",
	}
	gcpToken := "foo-token"

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			runner := &mockCommandRunner{ExpectedCommands: test.expectedCommand}
			if err := pushPythonArtifact(runner, args, test.path, gcpToken); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGenerateUri(t *testing.T) {
	testCases := []struct {
		name          string
		repository    string
		filePath      string
		expectedURI   string
		expectedError string // Partial error string to check if error occurs
	}{
		{
			name:        "Successful for source distribution",
			repository:  "http://foo.com/ar-repo",
			filePath:    "/workspace/source/dist/foo-8.0.0.tar.gz",
			expectedURI: "http://foo.com/ar-repo/foo/foo-8.0.0.tar.gz",
		},
		{
			name:        "Successful for wheel distribution",
			repository:  "http://foo.com/ar-repo",
			filePath:    "/workspace/source/dist/foo-8.0.0.whl",
			expectedURI: "http://foo.com/ar-repo/foo/foo-8.0.0.whl",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			uri := generateURI(test.repository, test.filePath)

			if uri != test.expectedURI {
				t.Errorf("uri mismatch: got '%s', expected '%s'", uri, test.expectedURI)
			}
		})
	}
}

func TestEnsureHTTPSByDefault(t *testing.T) {
	testCases := []struct {
		name          string
		repository    string
		expectedURI   string
		expectedError string // Partial error string to check if error occurs
	}{
		{
			name:        "Valid repository url",
			repository:  "http://foo.com/ar-repo",
			expectedURI: "http://foo.com/ar-repo",
		},
		{
			name:          "Invalid repository url",
			repository:    "http://foo.com/ar-repo%",
			expectedError: "invalid URL",
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			uri, err := ensureHTTPSByDefault(test.repository)

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
				if uri != test.expectedURI {
					t.Errorf("uri mismatch: got '%s', expected '%s'", uri, test.expectedURI)
				}
			}
		})
	}
}

func TestValidatePaths(t *testing.T) {
	testCases := []struct {
		name          string
		filename      string
		expectedError string // Partial error string to check if error occurs
	}{
		{
			name:     "Valid source distribution",
			filename: "foo-8.0.0.tar.gz",
		},
		{
			name:     "Valid wheel distribution",
			filename: "foo-8.0.0.whl",
		},
		{
			name:          "Invalid distribution format",
			filename:      "foo-8.0.0.egg",
			expectedError: "invalid filename format",
		},
		{
			name:          "Invalid file path",
			filename:      "",
			expectedError: "invalid file path",
		},
	}
	tmpdir := t.TempDir()

	for _, test := range testCases {
		var filePaths []string
		if test.filename != "" {
			// Create temporary files for testing
			tmpfile, err := os.CreateTemp(tmpdir, "*"+test.filename)
			if err != nil {
				t.Errorf("Error creating temporary file: %v", err)
			}
			defer tmpfile.Close()
			filePaths = []string{tmpfile.Name()}
		} else {
			filePaths = []string{tmpdir}
		}

		t.Run(test.name, func(t *testing.T) {
			err := validatePaths(filePaths)

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
			}
		})
	}
}
