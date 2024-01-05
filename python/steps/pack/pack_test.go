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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudBuild/python-images/steps/internal/auth"
	"github.com/GoogleCloudBuild/python-images/steps/internal/cmd"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var mockLogger = zap.NewNop()

func TestParseArgs(t *testing.T) {
	t.Run("parse valid arguments", func(t *testing.T) {
		f := setupFlags()
		err := f.Parse([]string{"--command=test-command", "--artifactRegistryUrl=test-url",
			"--sourceDistributionResultsPath=test-source.json", "--wheelDistributionResultsPath=test-wheel.json"})
		assert.NoError(t, err)

		args, err := ParseArgs(f)

		assert.NoError(t, err)
		assert.Equal(t, []string{"test-command"}, args.Command)
		assert.Equal(t, "https://test-url", args.ArtifactRegistryUrl)
		assert.Equal(t, "test-source.json", args.SourceDistributionResultsPath)
		assert.Equal(t, "test-wheel.json", args.WheelDistributionResultsPath)
	})

	t.Run("parse empty arguments", func(t *testing.T) {
		f := setupFlags()
		err := f.Parse([]string{})
		assert.NoError(t, err)

		args, err := ParseArgs(f)

		assert.NoError(t, err)
		assert.Equal(t, []string{}, args.Command)
		assert.Equal(t, "", args.ArtifactRegistryUrl)
	})
}

func setupFlags() *pflag.FlagSet {
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("command", "", "")
	f.String("artifactRegistryUrl", "", "")
	f.String("sourceDistributionResultsPath", "", "")
	f.String("wheelDistributionResultsPath", "", "")
	f.Bool("verbose", false, "")
	return f
}

func TestPackagePythonArtifact(t *testing.T) {
	t.Run("successful with commands", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.VirtualEnvPython3, mock.AnythingOfType("[]string")).Return(nil)
		args := Arguments{
			Command: []string{"test-command"},
		}

		err := packagePythonArtifact(mockRunner, args, mockLogger)

		assert.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("failure with commands", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.VirtualEnvPython3, mock.AnythingOfType("[]string")).Return(errors.New("command failed"))
		args := Arguments{
			Command: []string{"test-command"},
		}

		err := packagePythonArtifact(mockRunner, args, mockLogger)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command failed")
		mockRunner.AssertExpectations(t)
	})
}

func TestPushPythonArtifact(t *testing.T) {
	t.Run("successful with artifact registry url", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.VirtualEnvPython3, mock.AnythingOfType("[]string")).Return(nil)
		mockClient := new(auth.MockHTTPClient)
		mockClient.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "foo-token"}`))),
		}, nil)

		args := Arguments{
			ArtifactRegistryUrl: "test-url",
		}

		err := pushPythonArtifact(mockRunner, args, mockClient, mockLogger)

		assert.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("failure with artifact registry url", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.VirtualEnvPython3, mock.AnythingOfType("[]string")).Return(errors.New("command failed"))
		mockClient := new(auth.MockHTTPClient)
		mockClient.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"access_token": "foo-token"}`))),
		}, nil)

		args := Arguments{
			ArtifactRegistryUrl: "test-url",
		}

		err := pushPythonArtifact(mockRunner, args, mockClient, mockLogger)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to push python artifact")
		mockRunner.AssertExpectations(t)
	})

	t.Run("failed to get token", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockClient := new(auth.MockHTTPClient)
		mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("failed to fetch token"))

		args := Arguments{
			ArtifactRegistryUrl: "test-url",
		}

		err := pushPythonArtifact(mockRunner, args, mockClient, mockLogger)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch token")
		mockRunner.AssertExpectations(t)
	})
}

func TestExecute(t *testing.T) {
	t.Run("failed to create virtual environment", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.Python3, mock.AnythingOfType("[]string")).Return(errors.New("failed to create virtual environment"))
		mockClient := new(auth.MockHTTPClient)

		args := Arguments{
			Command:             []string{"test-command"},
			ArtifactRegistryUrl: "test-url",
		}

		err := Execute(mockRunner, args, mockClient)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create virtual environment")
		mockRunner.AssertExpectations(t)
	})

	t.Run("failed to install dependencies for package command", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.Python3, mock.AnythingOfType("[]string")).Return(nil)
		mockRunner.On("Run", cmd.VirtualEnvPip, mock.AnythingOfType("[]string")).Return(errors.New("failed to install dependencies for package command"))
		mockClient := new(auth.MockHTTPClient)

		args := Arguments{
			Command:             []string{"test-command"},
			ArtifactRegistryUrl: "test-url",
		}

		err := Execute(mockRunner, args, mockClient)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to install dependencies for package command")
		mockRunner.AssertExpectations(t)
	})

	t.Run("failed to package python artifact", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.Python3, mock.AnythingOfType("[]string")).Return(nil)
		mockRunner.On("Run", cmd.VirtualEnvPip, mock.AnythingOfType("[]string")).Return(nil)
		mockRunner.On("Run", cmd.VirtualEnvPython3, mock.AnythingOfType("[]string")).Return(errors.New("failed to package python artifact"))
		mockClient := new(auth.MockHTTPClient)

		args := Arguments{
			Command:             []string{"test-command"},
			ArtifactRegistryUrl: "test-url",
		}

		err := Execute(mockRunner, args, mockClient)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to package python artifact")
		mockRunner.AssertExpectations(t)
	})

	t.Run("failed to push python artifact", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.Python3, mock.AnythingOfType("[]string")).Return(nil)
		mockRunner.On("Run", cmd.VirtualEnvPip, mock.AnythingOfType("[]string")).Return(nil)
		mockRunner.On("Run", cmd.VirtualEnvPython3, mock.AnythingOfType("[]string")).Return(nil)
		mockClient := new(auth.MockHTTPClient)
		mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("failed to fetch token"))

		args := Arguments{
			Command:             []string{"test-command"},
			ArtifactRegistryUrl: "test-url",
		}

		err := Execute(mockRunner, args, mockClient)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch token")
		mockRunner.AssertExpectations(t)
	})
}

func TestInstallDependenciesForPackage(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.VirtualEnvPip, mock.AnythingOfType("[]string")).Return(nil)

		err := installDependenciesForPackage(mockRunner, mockLogger)

		assert.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		mockRunner := new(cmd.MockCommandRunner)
		mockRunner.On("Run", cmd.VirtualEnvPip, mock.AnythingOfType("[]string")).Return(errors.New("failed to install dependencies for package command"))

		err := installDependenciesForPackage(mockRunner, mockLogger)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to install dependencies for package command")
		mockRunner.AssertExpectations(t)
	})
}

func TestGenerateProvenance(t *testing.T) {
	testCases := []struct {
		name                 string
		mockFileNames        []string
		expectedSourceOutput ProvenanceOutput
		expectedWheelOutput  ProvenanceOutput
		expectedError        string
	}{
		{
			name:          "valid source and wheel distributions",
			mockFileNames: []string{"foo-1.0.tar.gz", "foo-1.0.whl"},
			expectedSourceOutput: ProvenanceOutput{
				URI:    "pkg:foo.com/ar-repo/foo@1.0",
				Digest: "c4ef48263126ab6d675b3a5fe5b6a075b8e6f9f0e40ea1b31d9ee94cf5cb5d3e",
			},
			expectedWheelOutput: ProvenanceOutput{
				URI:    "pkg:foo.com/ar-repo/foo@1.0",
				Digest: "c4ef48263126ab6d675b3a5fe5b6a075b8e6f9f0e40ea1b31d9ee94cf5cb5d3e",
			},
			expectedError: "",
		},
		{
			name:          "no distributions",
			mockFileNames: []string{},
			expectedSourceOutput: ProvenanceOutput{
				URI:    "",
				Digest: "",
			},
			expectedWheelOutput: ProvenanceOutput{
				URI:    "",
				Digest: "",
			},
			expectedError: "no files found in dist directory",
		},
		{
			name:          "duplicate distributions",
			mockFileNames: []string{"foo-1.0.tar.gz", "foo-2.0.tar.gz"},
			expectedSourceOutput: ProvenanceOutput{
				URI:    "",
				Digest: "",
			},
			expectedWheelOutput: ProvenanceOutput{
				URI:    "",
				Digest: "",
			},
			expectedError: "file already exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDistDir := t.TempDir()
			tempOutputDir := t.TempDir()

			dummyData := []byte("dummy data for testing")
			for _, fileName := range tc.mockFileNames {
				filePath := filepath.Join(tempDistDir, fileName)
				if err := os.WriteFile(filePath, dummyData, 0444); err != nil {
					t.Fatalf("Failed to write to mock file: %v", err)
				}
			}

			args := Arguments{
				ArtifactRegistryUrl:           "http://foo.com/ar-repo",
				SourceDistributionResultsPath: filepath.Join(tempOutputDir, "source"),
				WheelDistributionResultsPath:  filepath.Join(tempOutputDir, "wheel"),
			}

			err := generateProvenance(tempDistDir, args, mockLogger)

			if tc.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					verifyProvenanceOutput(t, args.SourceDistributionResultsPath, tc.expectedSourceOutput)
					verifyProvenanceOutput(t, args.WheelDistributionResultsPath, tc.expectedWheelOutput)
				}
			}
		})
	}
}

func verifyProvenanceOutput(t *testing.T, outputPath string, expectedOutput ProvenanceOutput) {
	t.Helper()
	outputFile, err := os.ReadFile(outputPath)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	var output ProvenanceOutput
	if err := json.Unmarshal(outputFile, &output); err != nil {
		t.Errorf("Failed to unmarshal output file: %v", err)
		return
	}

	if output.URI != expectedOutput.URI || output.Digest != expectedOutput.Digest {
		t.Errorf("Expected output %+v, got %+v", expectedOutput, output)
	}
}

func TestGenerateUri(t *testing.T) {
	t.Run("successful for source distribution", func(t *testing.T) {
		artifactRegistryUrl := "http://foo.com/ar-repo"

		uri, err := generateUri(artifactRegistryUrl, "foo-8.0.0.tar.gz")

		assert.NoError(t, err)
		assert.Equal(t, "pkg:foo.com/ar-repo/foo@8.0.0", uri)
	})

	t.Run("successful for wheel distribution", func(t *testing.T) {
		artifactRegistryUrl := "http://foo.com/ar-repo"

		uri, err := generateUri(artifactRegistryUrl, "foo-8.0.0.whl")

		assert.NoError(t, err)
		assert.Equal(t, "pkg:foo.com/ar-repo/foo@8.0.0", uri)
	})

	t.Run("failed for unknown distribution", func(t *testing.T) {
		artifactRegistryUrl := "http://foo.com/ar-repo"

		uri, err := generateUri(artifactRegistryUrl, "foo-8.0.0.egg")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid filename format")
		assert.Equal(t, "", uri)
	})

	t.Run("failed to parse artifact registry url", func(t *testing.T) {
		artifactRegistryUrl := "http://foo.com/ar-repo%"

		uri, err := generateUri(artifactRegistryUrl, "foo-8.0.0.tar.gz")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing ArtifactRegistryUrl")
		assert.Equal(t, "", uri)
	})
}

func TestComputeDigest(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write([]byte("test"))
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}

		digest, err := computeDigest(tempFile.Name())

		assert.NoError(t, err)
		assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", digest)
	})

	t.Run("failed for empty file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = computeDigest(tempFile.Name())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty file")
	})
}
