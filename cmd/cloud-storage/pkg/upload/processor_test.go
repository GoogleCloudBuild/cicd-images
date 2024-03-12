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
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// setupTestDirectory will create and return tmp test directory with the following structure:
// tmpdir/
//
//	a/foo.txt
//	b/bar.txt
//	c/foobar.txt
//	file1.txt
//	file2.txt
//	file3.go
//
// Additionally, the caller should call os.RemoveAll to clean up the temp test directory.
func setupTestDirectory(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "testDir")
	if err != nil {
		t.Fatalf("Error creating temporary directory %v", err)
	}
	for _, dir := range []string{"a", "b", "c"} {
		path := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			t.Fatalf("Error creating test file %v", err)
		}
	}

	testFiles := []string{"a/foo.txt", "b/bar.txt", "c/foobar.txt", "file1.txt", "file2.txt", "file3.go"}
	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		if _, err := os.Create(path); err != nil {
			t.Fatalf("Error creating test file %v", err)
		}
	}

	return tempDir
}

// getAllFilesAsUploadInput returns an array of UploadInputs.
// This is used across a handful of tests so broke it into it's own function
func getAllFilesAsUploadInput(tempDir, prefix string) []UploadInput {
	return []UploadInput{
		{
			FilePath:   filepath.Join(tempDir, "a", "foo.txt"),
			ObjectName: prefix + "a/foo.txt",
		},
		{
			FilePath:   filepath.Join(tempDir, "b", "bar.txt"),
			ObjectName: prefix + "b/bar.txt",
		},
		{
			FilePath:   filepath.Join(tempDir, "c", "foobar.txt"),
			ObjectName: prefix + "c/foobar.txt",
		},
		{
			FilePath:   filepath.Join(tempDir, "file1.txt"),
			ObjectName: prefix + "file1.txt",
		},
		{
			FilePath:   filepath.Join(tempDir, "file2.txt"),
			ObjectName: prefix + "file2.txt",
		},
		{
			FilePath:   filepath.Join(tempDir, "file3.go"),
			ObjectName: prefix + "file3.go",
		},
	}
}

func TestProcessPath_SimpleFolder_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)

	want := getAllFilesAsUploadInput(tempDir, "")
	got, err := ProcessPath(tempDir, "", "", false, false)
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProcessPath(tempDir, '', '', false, false) diff (-want +got):\n%s", diff)
	}
}

func TestProcessPath_SingleFile_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)

	want := []UploadInput{
		{
			FilePath:   filepath.Join(tempDir, "file2.txt"),
			ObjectName: "file2.txt",
		},
	}

	got, _ := ProcessPath(want[0].FilePath, "", "", false, false)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProcessPath(want[0].FilePath, '', '', false, false) diff (-want +got):\n%s", diff)
	}
}

func TestProcessPath_SingleFileWithPrefix_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)

	want := []UploadInput{
		{
			FilePath:   filepath.Join(tempDir, "file2.txt"),
			ObjectName: "myPrefix/file2.txt",
		},
	}

	got, _ := ProcessPath(want[0].FilePath, "myPrefix", "", false, false)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProcessPath(want[0].FilePath, 'myPrefix', '', false, false) diff (-want +got):\n%s", diff)
	}
}

func TestProcessPath_FolderWithPrefix_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)

	myPrefix := "myPrefix/"
	want := getAllFilesAsUploadInput(tempDir, myPrefix)

	got, _ := ProcessPath(tempDir, myPrefix, "", false, false)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProcessPath(tempDir, 'myPrefix', '', false, false) diff (-want +got):\n%s", diff)
	}
}

func TestProcessPath_FolderWithIncludeParent_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)

	want := []UploadInput{
		{
			FilePath:   filepath.Join(tempDir, "a", "foo.txt"),
			ObjectName: filepath.Join(tempDir, "a", "foo.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "b", "bar.txt"),
			ObjectName: filepath.Join(tempDir, "b", "bar.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "c", "foobar.txt"),
			ObjectName: filepath.Join(tempDir, "c", "foobar.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "file1.txt"),
			ObjectName: filepath.Join(tempDir, "file1.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "file2.txt"),
			ObjectName: filepath.Join(tempDir, "file2.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "file3.go"),
			ObjectName: filepath.Join(tempDir, "file3.go"),
		},
	}

	got, _ := ProcessPath(tempDir, "", "", false, true)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProcessPath(tempDir, '', '', false, true) diff (-want +got):\n%s", diff)
	}
}

func TestProcessPath_FolderWithPrefixAndIncludeParent_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)
	myPrefix := "myPrefix/"

	want := []UploadInput{
		{
			FilePath:   filepath.Join(tempDir, "a", "foo.txt"),
			ObjectName: filepath.Join(myPrefix, tempDir, "a", "foo.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "b", "bar.txt"),
			ObjectName: filepath.Join(myPrefix, tempDir, "b", "bar.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "c", "foobar.txt"),
			ObjectName: filepath.Join(myPrefix, tempDir, "c", "foobar.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "file1.txt"),
			ObjectName: filepath.Join(myPrefix, tempDir, "file1.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "file2.txt"),
			ObjectName: filepath.Join(myPrefix, tempDir, "file2.txt"),
		},
		{
			FilePath:   filepath.Join(tempDir, "file3.go"),
			ObjectName: filepath.Join(myPrefix, tempDir, "file3.go"),
		},
	}

	got, _ := ProcessPath(tempDir, myPrefix, "", false, true)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProcessPath(tempDir, myPrefix, '', false, true) diff (-want +got):\n%s", diff)
	}
}

func TestProcessPath_FolderWithIgnoreList_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)

	// create an gcloudignore file with the following data
	// *.go
	// foo.txt
	// c/*
	// We expect this to do the following:
	// 1) Skip all go files
	// 2) Skip any foo.txt files, one exists in a/foo.txt
	// 3) Skip all files in the c/ directory.
	gcloudfile := ".gcloudignore"
	data := []byte("*.go\nfoo.txt\nc/*")
	e := os.WriteFile(gcloudfile, data, 0600)
	if e != nil {
		t.Errorf("os.WriteFile(path, data, 0600) = %v", e)
	}
	defer os.Remove(gcloudfile)

	want := []UploadInput{
		{
			FilePath:   filepath.Join(tempDir, "b", "bar.txt"),
			ObjectName: "b/bar.txt",
		},
		{
			FilePath:   filepath.Join(tempDir, "file1.txt"),
			ObjectName: "file1.txt",
		},
		{
			FilePath:   filepath.Join(tempDir, "file2.txt"),
			ObjectName: "file2.txt",
		},
	}

	got, _ := ProcessPath(tempDir, "", "", true, false)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProcessPath(tempDir, '', '', true, false) diff (-want +got):\n%s", diff)
	}
}

func TestProcessPath_FolderWithGlobs_Success(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer os.RemoveAll(tempDir)
	tests := []struct {
		name           string
		glob           string
		expectedResult []UploadInput
	}{
		{
			name: "TextFileGlob",
			glob: "file*.txt",
			expectedResult: []UploadInput{
				{
					FilePath:   filepath.Join(tempDir, "file1.txt"),
					ObjectName: "file1.txt",
				},
				{
					FilePath:   filepath.Join(tempDir, "file2.txt"),
					ObjectName: "file2.txt",
				},
			},
		},
		{
			name: "GoFileGlob",
			glob: "*.go",
			expectedResult: []UploadInput{
				{
					FilePath:   filepath.Join(tempDir, "file3.go"),
					ObjectName: "file3.go",
				},
			},
		},
		{
			name: "Matching1Or3InFileName",
			glob: "*[13]*",
			expectedResult: []UploadInput{
				{
					FilePath:   filepath.Join(tempDir, "file1.txt"),
					ObjectName: "file1.txt",
				},
				{
					FilePath:   filepath.Join(tempDir, "file3.go"),
					ObjectName: "file3.go",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := ProcessPath(tempDir, "", test.glob, false, false)
			if diff := cmp.Diff(test.expectedResult, got); diff != "" {
				t.Errorf("ProcessPath(tempDir, '', test.glob, false, false) diff (-want +got):\n%s", diff)
			}
		})
	}
}
