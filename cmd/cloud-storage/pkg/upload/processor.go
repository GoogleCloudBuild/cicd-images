// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/yargevad/filepathx"
)

// expandFolderPathWithGlob takes in a folder path and an optional glob string, and will perform
// a glob expansion to return all files that match. A wildcard glob will be used if no glob is
// provided.
func expandFolderPathWithGlob(path, glob string) ([]string, error) {
	if glob == "" {
		glob = "**/*"
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return filepathx.Glob(path + glob)
}

// filterPaths will exclude all folders from the filepaths input.
// Additionally, if useIgnoreList is true, the .gcloudignore list file will be processed,
// and any filepaths that match will also be excluded.
func filterPaths(filepaths []string, useIgnoreList bool) []string {
	var outputList []string

	var rules *ignore.GitIgnore

	if useIgnoreList {
		rules, _ = ignore.CompileIgnoreFile(".gcloudignore")
	}

	for _, f := range filepaths {
		fileInfo, _ := os.Stat(f)
		if fileInfo.IsDir() {
			continue
		}

		if rules != nil && rules.MatchesPath(f) {
			fmt.Printf("Skipping file because it was on ignorelist: %v\n", f)
			continue
		}

		outputList = append(outputList, f)
	}

	return outputList
}

// buildUploadInput will construct the input to be passed to the Uploader.
// Each UploadInput contains both an original filepath string, and an ObjectName
// string that the uploader will use when uploading to GCS. The ObjectName
// will be the filepath unless the includeParent boolean is set to false, in which
// case the original parent directory from the path will be removed from the ObjectName
func buildUploadInput(path, prefix string, filepaths []string, isDir, includeParent bool) ([]UploadInput, error) {
	uploadInputs := make([]UploadInput, len(filepaths))

	base := ""
	if isDir {
		base = path
	} else {
		base = filepath.Dir(path)
	}

	for i, f := range filepaths {
		uploadInputs[i].FilePath = f
		if includeParent {
			uploadInputs[i].ObjectName = filepath.Join(prefix, f)
		} else {
			objectName, err := filepath.Rel(base, f)
			if err != nil {
				return nil, err
			}
			uploadInputs[i].ObjectName = filepath.Join(prefix, objectName)
		}
	}

	return uploadInputs, nil
}

// ProcessPath will convert an input path (file or folder) into a list of UploadInput.
// Each UploadInput object represents a single file that will need to be uploaded to GCS.
// The prefix parameter will be the prefix for every GCS ObjectName
// The glob parameter can be used alongside a folder path to do selective filtering.
// The useIgnoreList parameter can be used to process a .gcloudignore file for additional filtering.
// The includeParent parameter can be used to maintain or ignore the folder structure in the objects uploaded to GCS.
func ProcessPath(path, prefix, glob string, useIgnoreList, includeParent bool) ([]UploadInput, error) {
	fmt.Printf("Path provided: %v with glob %v\n", path, glob)

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error processing provided path: %v, %w", path, err)
	}

	if !fileInfo.IsDir() && glob != "" {
		return nil, fmt.Errorf("can only provide glob when path is a folder")
	}

	if fileInfo.IsDir() {
		fileList, err := expandFolderPathWithGlob(path, glob)
		if err != nil {
			return nil, err
		}
		fileList = filterPaths(fileList, useIgnoreList)
		return buildUploadInput(path, prefix, fileList, true, includeParent)
	} else {
		fileList := []string{path}
		return buildUploadInput(path, prefix, fileList, false, includeParent)
	}
}
