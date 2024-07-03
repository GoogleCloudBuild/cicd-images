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
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

// ZipDirectory creates a zip archive of files and folders within the specified directory.
// The directory itself is not included.
func ZipDirectory(sourceDir string, out io.Writer) error {
	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		// Propagate any errors encountered during directory traversal.
		if err != nil {
			return err
		}

		// Ignore non-regular files such as symbolic links and UNIX sockets.
		if !info.IsDir() && !info.Mode().IsRegular() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Use forward slashes for zip file paths (cross-platform compatibility).
		header.Name = filepath.ToSlash(relPath)

		// A trailing slash indicates that this file is a directory and should have no data.
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// No content to write for a directory
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// ToGCSZipped uploads a zipped folder to Google Cloud Storage.
func ToGCSZipped(ctx context.Context, client *storage.Client, bucket, object, sourceDir string) error {
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if err := ZipDirectory(sourceDir, wc); err != nil {
		return fmt.Errorf("failed to upload zip: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	log.Printf("Uploaded file to gs://%s/%s\n", bucket, object)
	return nil
}
