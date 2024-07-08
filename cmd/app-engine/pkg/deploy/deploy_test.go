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

package deploy

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"

	appengine "cloud.google.com/go/appengine/apiv1"
	"cloud.google.com/go/appengine/apiv1/appenginepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/config"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// MockAppengineServer is a mock implementation of the App Engine Admin API.
type MockAppengineServer struct {
	appenginepb.UnimplementedVersionsServer
	appenginepb.UnimplementedServicesServer

	mu sync.Mutex
	// versions is a map of service names to versions, where each version is a map from version ID to version object.
	versions map[string]map[string]*appenginepb.Version
	// services is a map of service names to service objects.
	services map[string]*appenginepb.Service

	// CreateVersionFunc allows injecting an error into CreateVersion
	CreateVersionFunc func(ctx context.Context, req *appenginepb.CreateVersionRequest) (*longrunningpb.Operation, error)
	// UpdateServiceFunc allows injecting an error into UpdateService
	UpdateServiceFunc func(ctx context.Context, req *appenginepb.UpdateServiceRequest) (*longrunningpb.Operation, error)
}

func NewMockAppengineServer() *MockAppengineServer {
	return &MockAppengineServer{
		versions: make(map[string]map[string]*appenginepb.Version),
		services: make(map[string]*appenginepb.Service),
	}
}

// Start starts the mock server on a random port.
// It returns the address the server is listening on and a cleanup function.
func (s *MockAppengineServer) Start() (string, func(), error) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	appenginepb.RegisterVersionsServer(grpcServer, s)
	appenginepb.RegisterServicesServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), grpcServer.GracefulStop, nil
}

func (s *MockAppengineServer) CreateVersion(ctx context.Context, req *appenginepb.CreateVersionRequest) (*longrunningpb.Operation, error) {
	if s.CreateVersionFunc != nil {
		return s.CreateVersionFunc(ctx, req)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	serviceName := req.GetParent()
	version := req.GetVersion()

	if _, ok := s.versions[serviceName]; !ok {
		s.versions[serviceName] = make(map[string]*appenginepb.Version)
	}
	s.versions[serviceName][version.Id] = version

	anyVersion, err := anypb.New(version)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version to Any: %w", err)
	}

	return &longrunningpb.Operation{
		Name: "create-version-operation",
		Done: true,
		Result: &longrunningpb.Operation_Response{
			Response: anyVersion,
		},
	}, nil
}

func (s *MockAppengineServer) UpdateService(ctx context.Context, req *appenginepb.UpdateServiceRequest) (*longrunningpb.Operation, error) {
	if s.UpdateServiceFunc != nil {
		return s.UpdateServiceFunc(ctx, req)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	serviceName := req.GetName()
	service := req.GetService()

	// Basic validation for testing purposes.
	if service.GetSplit() == nil {
		return nil, status.Error(codes.InvalidArgument, "missing traffic split")
	}

	s.services[serviceName] = service

	anyService, err := anypb.New(service)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service to Any: %w", err)
	}

	return &longrunningpb.Operation{
		Name: "update-service-operation",
		Done: true,
		Result: &longrunningpb.Operation_Response{
			Response: anyService,
		},
	}, nil
}

// setupMockServerConnection creates a connection to the mock server and returns the connection.
func setupMockServerConnection(t *testing.T, mockServer *MockAppengineServer) *grpc.ClientConn {
	t.Helper()

	addr, cleanup, err := mockServer.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to mock server: %v", err)
	}

	t.Cleanup(func() {
		conn.Close()
		cleanup()
	})

	return conn
}

func setupFakeStorage(t *testing.T, objects []fakestorage.Object) *storage.Client {
	t.Helper()
	server := fakestorage.NewServer(objects)
	t.Cleanup(server.Stop)
	return server.Client()
}

func TestDeployService(t *testing.T) {
	mockServer := NewMockAppengineServer()
	conn := setupMockServerConnection(t, mockServer)

	client, err := appengine.NewVersionsClient(context.Background(), option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("Failed to create VersionsClient: %v", err)
	}

	serviceName := "test-service"
	version := &appenginepb.Version{
		Id: "v1",
	}

	deployedVersion, err := CreateVersion(context.Background(), client, serviceName, version)
	if err != nil {
		t.Fatalf("DeployService returned an error: %v", err)
	}

	if !proto.Equal(deployedVersion, version) {
		t.Fatalf("Deployed version does not match expected version: got %v, want %v", deployedVersion, version)
	}
}

func TestUpdateService(t *testing.T) {
	mockServer := NewMockAppengineServer()
	conn := setupMockServerConnection(t, mockServer)

	client, err := appengine.NewServicesClient(context.Background(), option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("Failed to create ServicesClient: %v", err)
	}

	serviceName := "test-service"
	version := &appenginepb.Version{
		Id: "v1",
	}

	updatedService, err := PromoteVersion(context.Background(), client, serviceName, version.Id)
	if err != nil {
		t.Fatalf("UpdateService returned an error: %v", err)
	}

	expectedService := &appenginepb.Service{
		Split: &appenginepb.TrafficSplit{
			Allocations: map[string]float64{version.Id: 1.0},
		},
	}

	if !proto.Equal(updatedService.GetSplit(), expectedService.GetSplit()) {
		t.Fatalf("Updated service split does not match expected split: got %v, want %v", updatedService.GetSplit(), expectedService.GetSplit())
	}
}

func TestCreateVersionErrors(t *testing.T) {
	tests := []struct {
		name         string
		grpcError    error
		wantErr      bool
		mockVersions map[string]map[string]*appenginepb.Version
	}{
		{
			name:      "gRPC Error",
			grpcError: status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			wantErr:   true,
		},
		{
			name:         "Marshaling Error",
			grpcError:    fmt.Errorf("failed to marshal version to Any"),
			wantErr:      true,
			mockVersions: make(map[string]map[string]*appenginepb.Version),
		},
	}

	//nolint: dupl // this test is superficially similar to another but has different semantics
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := NewMockAppengineServer()
			if tc.mockVersions != nil {
				mockServer.versions = tc.mockVersions
			}
			mockServer.CreateVersionFunc = func(context.Context, *appenginepb.CreateVersionRequest) (*longrunningpb.Operation, error) {
				return nil, tc.grpcError
			}
			conn := setupMockServerConnection(t, mockServer)

			client, err := appengine.NewVersionsClient(context.Background(), option.WithGRPCConn(conn))
			if err != nil {
				t.Fatalf("Failed to create VersionsClient: %v", err)
			}

			serviceName := "test-service"
			version := &appenginepb.Version{
				Id: "v1",
			}

			_, err = CreateVersion(context.Background(), client, serviceName, version)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CreateVersion() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestPromoteVersionErrors(t *testing.T) {
	tests := []struct {
		name         string
		grpcError    error
		wantErr      bool
		mockServices map[string]*appenginepb.Service
	}{
		{
			name:      "gRPC Error",
			grpcError: status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			wantErr:   true,
		},
		{
			name:         "Marshaling Error",
			grpcError:    fmt.Errorf("failed to marshal service to Any"),
			wantErr:      true,
			mockServices: make(map[string]*appenginepb.Service),
		},
		{
			name:      "Missing Traffic Split",
			grpcError: status.Error(codes.InvalidArgument, "missing traffic split"),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := NewMockAppengineServer()
			if tc.mockServices != nil {
				mockServer.services = tc.mockServices
			}
			mockServer.UpdateServiceFunc = func(context.Context, *appenginepb.UpdateServiceRequest) (*longrunningpb.Operation, error) {
				return nil, tc.grpcError
			}

			conn := setupMockServerConnection(t, mockServer)

			client, err := appengine.NewServicesClient(context.Background(), option.WithGRPCConn(conn))
			if err != nil {
				t.Fatalf("Failed to create ServicesClient: %v", err)
			}

			serviceName := "test-service"
			version := &appenginepb.Version{
				Id: "v1",
			}

			_, err = PromoteVersion(context.Background(), client, serviceName, version.Id)
			if (err != nil) != tc.wantErr {
				t.Fatalf("PromoteVersion() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestCreateAndDeployVersion(t *testing.T) {
	// Setup a valid app.yaml for testing
	validYAMLContent := `
runtime: go122
env: flex`
	path := createTempYAML(t, validYAMLContent, "app.yaml")

	testCases := []struct {
		name           string
		opts           config.AppEngineDeployOptions
		createErr      error
		promoteErr     error
		mockVersions   map[string]map[string]*appenginepb.Version
		mockServices   map[string]*appenginepb.Service
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "Success",
			opts: config.AppEngineDeployOptions{
				AppYAMLPath: path,
				ImageURL:    "test-image-url",
				ProjectID:   "test-project-id",
				Promote:     true,
				VersionID:   "test-version-id",
			},
			mockVersions: map[string]map[string]*appenginepb.Version{
				"apps/test-project-id/services/default": {
					"test-version-id": {Id: "test-version-id"},
				},
			},
			mockServices: map[string]*appenginepb.Service{
				"apps/test-project-id/services/default": {}, // Empty service for promotion
			},
			expectedOutput: "Deploying service: apps/test-project-id/services/default, version:, env:\"flex\" id:\"test-version-id\" runtime:\"go122\"\nVersion created successfully: env:\"flex\" id:\"test-version-id\" runtime:\"go122\"\nPromoted latest version ",
		},
		{
			name: "SuccessWithoutPromote",
			opts: config.AppEngineDeployOptions{
				AppYAMLPath: path,
				ImageURL:    "test-image-url",
				ProjectID:   "test-project-id",
				Promote:     false,
				VersionID:   "test-version-id",
			},
			mockVersions: map[string]map[string]*appenginepb.Version{
				"apps/test-project-id/services/default": {
					"test-version-id": {Id: "test-version-id"},
				},
			},
			expectedOutput: "Deploying service: apps/test-project-id/services/default, version:, env:\"flex\" id:\"test-version-id\" runtime:\"go122\"\nVersion created successfully: env:\"flex\" id:\"test-version-id\" runtime:\"go122\"\n",
		},
		{
			name: "ErrorParsingYAML",
			opts: config.AppEngineDeployOptions{
				AppYAMLPath: "", // Invalid path
				ImageURL:    "test-image-url",
				ProjectID:   "test-project-id",
				Promote:     true,
				VersionID:   "test-version-id",
			},
			expectedErr: "open : no such file or directory",
		},
		{
			name: "ErrorCreateVersion",
			opts: config.AppEngineDeployOptions{
				AppYAMLPath: path,
				ImageURL:    "test-image-url",
				ProjectID:   "test-project-id",
				Promote:     true,
				VersionID:   "test-version-id",
			},
			createErr:   status.Error(codes.Internal, "mock create error"),
			expectedErr: "error during version creation: failed to create version: rpc error: code = Internal desc = mock create error",
		},
		{
			name: "ErrorPromoteVersion",
			opts: config.AppEngineDeployOptions{
				AppYAMLPath: path,
				ImageURL:    "test-image-url",
				ProjectID:   "test-project-id",
				Promote:     true,
				VersionID:   "test-version-id",
			},
			mockVersions: map[string]map[string]*appenginepb.Version{
				"apps/test-project-id/services/default": {
					"test-version-id": {Id: "test-version-id"},
				},
			},
			promoteErr:  status.Error(codes.Internal, "mock promote error"),
			expectedErr: "error during promotion: error promoting version: rpc error: code = Internal desc = mock promote error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := NewMockAppengineServer()
			if tc.mockVersions != nil {
				mockServer.versions = tc.mockVersions
			}
			if tc.mockServices != nil {
				mockServer.services = tc.mockServices
			}
			if tc.createErr != nil {
				mockServer.CreateVersionFunc = func(_ context.Context, _ *appenginepb.CreateVersionRequest) (*longrunningpb.Operation, error) {
					return nil, tc.createErr
				}
			}
			if tc.promoteErr != nil {
				mockServer.UpdateServiceFunc = func(_ context.Context, _ *appenginepb.UpdateServiceRequest) (*longrunningpb.Operation, error) {
					return nil, tc.promoteErr
				}
			}
			conn := setupMockServerConnection(t, mockServer)

			versionClient, err := appengine.NewVersionsClient(context.Background(), option.WithGRPCConn(conn))
			assert.NoError(t, err, "failed to create VersionsClient")

			servicesClient, err := appengine.NewServicesClient(context.Background(), option.WithGRPCConn(conn))
			assert.NoError(t, err, "failed to create ServicesClient")

			storageClient := setupFakeStorage(t, []fakestorage.Object{})

			// Capture log output
			log.SetOutput(os.Stdout)

			err = CreateAndDeployVersion(context.Background(), versionClient, servicesClient, storageClient, tc.opts)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createTempYAML(t *testing.T, content, filename string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", filename)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}
