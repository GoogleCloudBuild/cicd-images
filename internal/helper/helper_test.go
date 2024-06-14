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
package helper

import (
	"context"
	"os"
	"strings"
	"testing"
)

type mockMetaDataClient struct{}

var _ metadataClient = (*mockMetaDataClient)(nil)

func (r *mockMetaDataClient) GetWithContext(_ context.Context, _ string) (string, error) {
	return `{"access_token": "fetched_access_token", "expires_in": 3600, "token_type": "Bearer"}`, nil
}

func TestGetAuthenticationToken(t *testing.T) {
	mockClient := &mockMetaDataClient{}
	ctx := context.Background()
	result := "fetched_access_token"

	t.Run("get authentication token", func(t *testing.T) {
		token, err := GetAuthenticationToken(ctx, mockClient)
		if err != nil {
			t.Fatalf("Expected no errors, but got %v", err)
		}

		if token != result {
			t.Errorf("Fetched token does not match expected token. Expected: \n%v\n Got: \n%v\n", result, token)
		}
	})
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

			digest, err := ComputeDigest(tempFile.Name())

			if err != nil && !strings.Contains(err.Error(), test.expectedError) {
				t.Fatalf("error message mismatch: got '%s', expected to contain '%s'", err.Error(), test.expectedError)
			}
			if digest != test.expectedDigest {
				t.Errorf("digest mismatch: got '%s', expected '%s'", digest, test.expectedDigest)
			}
		})
	}
}
