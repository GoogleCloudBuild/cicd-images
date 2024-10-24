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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/spf13/pflag"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"golang.org/x/oauth2"
	"google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	tmpZip = "go.zip"
	source = "./"
)

type Arguments struct {
	Project    string
	Repository string
	Location   string
	ModulePath string
	Version    string
}

type Provenance struct {
	URI             string `json:"uri"`
	Digest          string `json:"digest"`
	IsBuildArtifact string `json:"isBuildArtifact"`
}

// ParseArgs parses the arguments passed to the publish command.
func ParseArgs(f *pflag.FlagSet) (Arguments, error) {
	project, err := f.GetString("project")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get project: %w", err)
	}
	repository, err := f.GetString("repository")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get repository: %w", err)
	}
	location, err := f.GetString("location")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get location: %w", err)
	}
	modulePath, err := f.GetString("modulePath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get modulePath: %w", err)
	}
	version, err := f.GetString("version")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get version: %w", err)
	}

	return Arguments{
		Project:    project,
		Repository: repository,
		Location:   location,
		ModulePath: modulePath,
		Version:    version,
	}, nil
}

// Execute is the entrypoint for the publish command execution.
func Execute(args Arguments) error {
	filePath, err := packageModule(args, source)
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(filePath))

	if err := uploadModule(args, filePath); err != nil {
		return err
	}

	return nil
}

// Package the Go module into a zip file.
func packageModule(args Arguments, source string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", fmt.Errorf("error creating dir for packaged module: %w", err)
	}

	zipPath := filepath.Join(tmpDir, tmpZip)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("error creating file for packaged module: %w", err)
	}
	defer zipFile.Close()

	if err := zip.CreateFromDir(zipFile, module.Version{
		Path:    args.ModulePath,
		Version: args.Version,
	}, source); err != nil {
		return "", fmt.Errorf("error packaging go module: %w", err)
	}

	return zipPath, nil
}

// Upload the packaged Go module to Artifact Registry.
func uploadModule(args Arguments, filePath string) error {
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	defer cf()

	// Get default access token
	accessToken, err := helper.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("error finding default credentials: %w", err)
	}

	// Create Artifact Registry Services
	arService, err := artifactregistry.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})))
	if err != nil {
		return fmt.Errorf("error creating artifact registry service: %w", err)
	}
	gmService := &goModService{gms: artifactregistry.NewProjectsLocationsRepositoriesGoModulesService(arService)}
	opService := &operationService{ops: artifactregistry.NewProjectsLocationsOperationsService(arService)}

	resourceName := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", args.Project, args.Location, args.Repository)
	zipFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening packaged module for upload: %w", err)
	}
	defer zipFile.Close()

	return executeUpload(ctx, resourceName, zipFile, gmService, opService)
}

type ArtifactRegistryGoModulesService interface {
	Upload(context.Context, string, *os.File) (*artifactregistry.UploadGoModuleMediaResponse, error)
}

type goModService struct {
	gms *artifactregistry.ProjectsLocationsRepositoriesGoModulesService
}

var _ ArtifactRegistryGoModulesService = (*goModService)(nil) // Verify goModService implements ArtifactRegistryGoModulesService

func (g *goModService) Upload(ctx context.Context, resourceName string, zipFile *os.File) (*artifactregistry.UploadGoModuleMediaResponse, error) {
	uploadCall := g.gms.Upload(resourceName, &artifactregistry.UploadGoModuleRequest{}).Context(ctx)
	response, err := uploadCall.Media(zipFile, googleapi.ContentType("application/zip")).Do()
	if err != nil {
		return &artifactregistry.UploadGoModuleMediaResponse{}, fmt.Errorf("error executing upload call: %w", err)
	}
	return response, nil
}

type ArtifactRegistryOperationService interface {
	Get(string) (*artifactregistry.Operation, error)
}

type operationService struct {
	ops *artifactregistry.ProjectsLocationsOperationsService
}

var _ ArtifactRegistryOperationService = (*operationService)(nil) // Verify operationService implements ArtifactRegistryOperationService

func (o *operationService) Get(name string) (*artifactregistry.Operation, error) {
	return o.ops.Get(name).Do()
}

// Execute upload call on packaged Go module.
func executeUpload(ctx context.Context, resourceName string, zipFile *os.File, gmService ArtifactRegistryGoModulesService, opService ArtifactRegistryOperationService) error {
	response, err := gmService.Upload(ctx, resourceName, zipFile)
	if err != nil {
		return fmt.Errorf("error executing upload call: %w", err)
	}

	// Poll Operation to check if upload call was completed
	for {
		op, err := opService.Get(response.Operation.Name)
		if err != nil {
			return fmt.Errorf("error refreshing operation request: %w", err)
		}

		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("error running operation to upload go module with status code: %d", op.Error.Code)
			}
			break
		}

		select {
		case <-ctx.Done():
			// Timeout or cancellation occurred
			return ctx.Err()
		case <-time.After(3 * time.Second):
			// Continue polling
		}
	}

	return nil
}
