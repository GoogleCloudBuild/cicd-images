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
package pack

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/auth"
	"github.com/GoogleCloudBuild/cicd-images/cmd/python-steps/internal/command"
)

func TestInstallDependenciesForPackage(t *testing.T) {
	runner := &command.MockCommandRunner{
		ExpectedCommands: []string{command.VirtualEnvPip, "install", "build", "wheel", "twine"},
	}
	if err := installDependenciesForPackage(runner); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPackagePythonArtifact(t *testing.T) {
	runner := &command.MockCommandRunner{
		ExpectedCommands: []string{command.VirtualEnvPython3, "setup.py", "sdist", "bdist_wheel"},
	}
	args := Arguments{
		Command: []string{"setup.py", "sdist", "bdist_wheel"},
	}
	if err := packagePythonArtifact(runner, args); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPushPythonArtifact(t *testing.T) {
	runner := &command.MockCommandRunner{
		ExpectedCommands: []string{command.VirtualEnvPython3, "-m", "twine", "upload", "--repository-url", "foo-url", "--username", "oauth2accesstoken", "--password", "foo-token", "dist/*"},
	}
	args := Arguments{
		ArtifactRegistryUrl: "foo-url",
	}
	client := &auth.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "foo-token"}`))),
			}, nil
		},
	}
	if err := pushPythonArtifact(runner, args, client); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateUri(t *testing.T) {
	testCases := []struct {
		name                string
		artifactRegistryUrl string
		filename            string
		expectedUri         string
		expectedError       string // Partial error string to check if error occurs
	}{
		{
			name:                "Successful for source distribution",
			artifactRegistryUrl: "http://foo.com/ar-repo",
			filename:            "foo-8.0.0.tar.gz",
			expectedUri:         "pkg:foo.com/ar-repo/foo@8.0.0",
		},
		{
			name:                "Successful for wheel distribution",
			artifactRegistryUrl: "http://foo.com/ar-repo",
			filename:            "foo-8.0.0.whl",
			expectedUri:         "pkg:foo.com/ar-repo/foo@8.0.0",
		},
		{
			name:                "Failed for unknown distribution",
			artifactRegistryUrl: "http://foo.com/ar-repo",
			filename:            "foo-8.0.0.egg",
			expectedError:       "invalid filename format",
		},
		{
			name:                "Failed to parse artifact registry url",
			artifactRegistryUrl: "http://foo.com/ar-repo%",
			filename:            "foo-8.0.0.tar.gz",
			expectedError:       "error parsing ArtifactRegistryUrl",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			uri, err := generateUri(test.artifactRegistryUrl, test.filename)

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

func TestComputeDigest(t *testing.T) {
	testCases := []struct {
		name           string
		fileContent    string
		expectedDigest string
		expectedError  string
	}{
		{
			name:           "Successful",
			fileContent:    "test",
			expectedDigest: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			name:          "Failed for empty file",
			fileContent:   "",
			expectedError: "empty file",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			tempFile, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			if _, err = tempFile.Write([]byte(test.fileContent)); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}

			digest, err := computeDigest(tempFile.Name())

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
				if digest != test.expectedDigest {
					t.Errorf("digest mismatch: got '%s', expected '%s'", digest, test.expectedDigest)
				}
			}
		})
	}
}
