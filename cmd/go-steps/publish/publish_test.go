//  Copyright 2024 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package publish

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/api/artifactregistry/v1"
)

func TestPackageModule(t *testing.T) {
	tmpdir := t.TempDir()
	modPath := filepath.Dir(createTestGoModule(t, tmpdir))
	args := Arguments{
		ModulePath: "github.com/test/module/path",
		Version:    "v0.0.0",
	}

	t.Run("package module", func(t *testing.T) {
		filePath, err := packageModule(args, modPath)
		if err != nil {
			t.Fatalf("Expected no errors, but got %v", err)
		}

		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) || filepath.Base(filePath) != "go.zip" {
			t.Fatalf("Expected file 'go.zip' to exist at path %v, but it does not exist.", filePath)
		}
	})
}

type mockGoModService struct{}

var _ ArtifactRegistryGoModulesService = (*mockGoModService)(nil) // Verify mockGoModService implements ArtifactRegistryGoModulesService

func (m *mockGoModService) Upload(_ context.Context, _ string, _ *os.File) (*artifactregistry.UploadGoModuleMediaResponse, error) {
	return &artifactregistry.UploadGoModuleMediaResponse{
		Operation: &artifactregistry.Operation{
			Name: "mock-operation",
		},
	}, nil
}

type mockOperationService struct{}

var _ ArtifactRegistryOperationService = (*mockOperationService)(nil) // Verify mockOperationService implements ArtifactRegistryOperationService

func (m *mockOperationService) Get(_ string) (*artifactregistry.Operation, error) {
	return &artifactregistry.Operation{
		Done: true,
	}, nil
}

func TestExecuteUpload(t *testing.T) {
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	defer cf()

	resourceName := "projects/test-project/locations/test-location/repositories/test-repository"
	zipFile := &os.File{}
	gmService := &mockGoModService{}
	opService := &mockOperationService{}

	t.Run("execute upload", func(t *testing.T) {
		if err := executeUpload(ctx, resourceName, zipFile, gmService, opService); err != nil {
			t.Fatalf("Expected no errors, but got %v", err)
		}
	})
}

func createTestGoModule(t *testing.T, dir string) string {
	t.Helper()
	modString := `
module github.com/test/module/path

go 1.20

require github.com/test/test v1.0.1
`

	tmpFile, err := os.CreateTemp(dir, "go.mod")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(modString); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}
