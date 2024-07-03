// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upload

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/google/go-cmp/cmp"
)

func TestZipDirectory(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		directory     string
		expectedError bool
	}{
		{
			name:          "SuccessWithNestedDir",
			directory:     "../../test/test_dir",
			expectedError: false,
		},
		{
			name:          "SuccessWithEmptyDir",
			directory:     "../../test/empty_dir",
			expectedError: false,
		},
		{
			name:          "NonExistentDirectory",
			directory:     "../../test/non_existent",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture loop variable

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Use an in-memory buffer instead of a temporary file
			var buf bytes.Buffer

			err := ZipDirectory(tc.directory, &buf)
			if tc.expectedError && err == nil {
				t.Fatal("Expected an error, but got nil")
			} else if !tc.expectedError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// If successful, verify the zip file contents
			if !tc.expectedError {
				if err := verifyZipContents(t, buf.Bytes(), tc.directory); err != nil {
					t.Errorf("Zip file contents verification failed: %v", err)
				}
			}
		})
	}
}

func TestToGCSZipped(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name          string
		bucket        string
		object        string
		sourceDir     string
		expectedError bool
	}{
		{
			name:          "SuccessWithNestedDir",
			bucket:        "test-bucket",
			object:        "test.zip",
			sourceDir:     "../../test/test_dir",
			expectedError: false,
		},
		{
			name:          "SuccessWithEmptyDir",
			bucket:        "test-bucket",
			object:        "test-empty.zip",
			sourceDir:     "../../test/empty_dir",
			expectedError: false,
		},
		{
			name:          "NonExistentDirectory",
			bucket:        "test-bucket",
			object:        "non-existent.zip",
			sourceDir:     "../../test/non_existent",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture loop variable

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := createGCSClient(t, []fakestorage.Object{
				{ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName: tc.bucket,
					Name:       tc.object,
				}},
			})

			err := ToGCSZipped(ctx, client, tc.bucket, tc.object, tc.sourceDir)
			if tc.expectedError && err == nil {
				t.Fatal("Expected an error, but got nil")
			} else if !tc.expectedError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// If successful, verify the object exists in the bucket and its content
			if !tc.expectedError {
				reader, err := client.Bucket(tc.bucket).Object(tc.object).NewReader(ctx)
				if err != nil {
					t.Errorf("Object not found in bucket: %v", err)
				}
				defer reader.Close()

				// Read the contents of the object from GCS
				zipData, err := io.ReadAll(reader)
				if err != nil {
					t.Errorf("Failed to read object content: %v", err)
				}

				// Verify the zip file contents
				if err := verifyZipContents(t, zipData, tc.sourceDir); err != nil {
					t.Errorf("Zip file contents verification failed: %v", err)
				}
			}
		})
	}
}

// verifyZipContents compares the contents of a zip file (from a byte slice)
// with a directory on disk.
func verifyZipContents(t *testing.T, zipData []byte, directory string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Create a map to track files AND directories in the zip archive
	zipFileEntries := make(map[string]bool)
	for _, f := range zipReader.File {
		// if a directory, do not include training slash
		name := strings.TrimSuffix(f.Name, "/")
		zipFileEntries[name] = true
	}

	// Walk the directory
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path
		relPath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Check if the entry (file or directory) exists in the zip map
		if _, exists := zipFileEntries[relPath]; !exists {
			return fmt.Errorf("entry missing from zip archive: %s", relPath)
		}

		// If it's a file, compare the contents
		if !info.IsDir() {
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read directory file content: %w", err)
			}

			zipFile, err := zipReader.Open(relPath)
			if err != nil {
				return fmt.Errorf("failed to open file in zip: %w", err)
			}
			defer zipFile.Close()

			zipContent, err := io.ReadAll(zipFile)
			if err != nil {
				return fmt.Errorf("failed to read zip file content: %w", err)
			}

			if !bytes.Equal(fileContent, zipContent) {
				t.Errorf("File contents mismatch:\n%s", cmp.Diff(string(fileContent), string(zipContent)))
			}
		}

		return nil // Continue walking the directory
	})

	return err
}

// createGCSClient creates a new fake GCS client.
func createGCSClient(t *testing.T, objects []fakestorage.Object) *storage.Client {
	t.Helper()
	server := fakestorage.NewServer(objects)
	t.Cleanup(server.Stop)
	return server.Client()
}
