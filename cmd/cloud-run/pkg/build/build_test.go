package build

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	"cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// fakeCloudBuildServer is a fake implementation of the Cloud Build server.
type fakeCloudBuildServer struct {
	cloudbuildpb.UnimplementedCloudBuildServer
	createBuildFunc func(*cloudbuildpb.CreateBuildRequest) (*longrunningpb.Operation, error)
}

func (f *fakeCloudBuildServer) CreateBuild(_ context.Context, req *cloudbuildpb.CreateBuildRequest) (*longrunningpb.Operation, error) {
	if f.createBuildFunc != nil {
		return f.createBuildFunc(req)
	}
	// Simulate a successful build creation with a long-running operation.
	build := &cloudbuildpb.Build{
		Id: "1234567890",
		Images: []string{
			"us-central1-docker.pkg.dev/test-project/cloud-run-source-deploy/test-service",
		},
	}
	anyBuild, err := anypb.New(build)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version to Any: %w", err)
	}
	return &longrunningpb.Operation{
		Name: "cloud-build-operation",
		Done: true,
		Result: &longrunningpb.Operation_Response{
			Response: anyBuild,
		},
	}, nil
}

func setupFakeCloudBuildServer(t *testing.T) (*cloudbuild.Client, *fakeCloudBuildServer) {
	ctx := context.Background()
	// Create a fake Cloud Build server.
	fakeServer := &fakeCloudBuildServer{}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	gsrv := grpc.NewServer()
	cloudbuildpb.RegisterCloudBuildServer(gsrv, fakeServer)
	fakeServerAddr := l.Addr().String()
	go func() {
		if err := gsrv.Serve(l); err != nil {
			panic(err)
		}
	}()

	c, err := cloudbuild.NewClient(ctx,
		option.WithEndpoint(fakeServerAddr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		t.Fatal(err)
	}
	return c, fakeServer
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	c, fakeServer := setupFakeCloudBuildServer(t)
	defer c.Close()

	opts := Options{
		ProjectID: "test-project",
		Region:    "us-central1",
		Service:   "test-service",
		Source:    "test-source",
	}
	t.Setenv(URLVar, "https://github.com/test/repo.git")
	t.Setenv(HashVar, "test-hash")

	t.Run("Success", func(t *testing.T) {
		build, err := Run(ctx, c, opts)
		if err != nil {
			t.Fatalf("Run() unexpected error: %v", err)
		}

		want := "us-central1-docker.pkg.dev/test-project/cloud-run-source-deploy/test-service"

		if build != want {
			t.Errorf("Run() = %v, want %v", build, want)
		}
	})

	t.Run("CreateBuildError", func(t *testing.T) {
		fakeServer.createBuildFunc = func(_ *cloudbuildpb.CreateBuildRequest) (*longrunningpb.Operation, error) {
			return nil, fmt.Errorf("injected error")
		}
		_, err := Run(ctx, c, opts)
		if err == nil {
			t.Fatal("Run() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "injected error") {
			t.Errorf("Run() error = %v, want error containing 'injected error'", err)
		}
	})
}

func TestCreateBuildRequest(t *testing.T) {
	tests := []struct {
		name       string
		opts       Options
		wantImage  string
		wantParent string
		wantSteps  []*cloudbuildpb.BuildStep
		wantSource *cloudbuildpb.Source
		wantErr    bool
		wantErrMsg string
		dockerfile bool
		envURL     string
		envHash    string
	}{
		{
			name: "Dockerfile",
			opts: Options{
				ProjectID: "test-project",
				Region:    "us-central1",
				Service:   "test-service",
				Source:    "test-source",
			},
			wantImage:  "us-central1-docker.pkg.dev/test-project/cloud-run-source-deploy/test-service",
			wantParent: "projects/test-project/locations/global",
			wantSteps: []*cloudbuildpb.BuildStep{
				{Name: "gcr.io/cloud-builders/docker", Args: []string{"build", "-t", "us-central1-docker.pkg.dev/test-project/cloud-run-source-deploy/test-service", "."}},
			},
			wantSource: &cloudbuildpb.Source{
				Source: &cloudbuildpb.Source_GitSource{
					GitSource: &cloudbuildpb.GitSource{
						Url:      "https://github.com/test/repo.git",
						Dir:      "test-source",
						Revision: "test-hash",
					},
				},
			},
			dockerfile: true,
			envURL:     "https://github.com/test/repo.git",
			envHash:    "test-hash",
		},
		{
			name: "No Dockerfile",
			opts: Options{
				ProjectID: "test-project",
				Region:    "us-central1",
				Service:   "test-service",
				Source:    "test-source",
			},
			wantImage:  "us-central1-docker.pkg.dev/test-project/cloud-run-source-deploy/test-service",
			wantParent: "projects/test-project/locations/global",
			wantSteps: []*cloudbuildpb.BuildStep{
				{Name: "gcr.io/k8s-skaffold/pack", Args: []string{"build", "us-central1-docker.pkg.dev/test-project/cloud-run-source-deploy/test-service", "--builder", "gcr.io/buildpacks/builder:v1"}},
			},
			wantSource: &cloudbuildpb.Source{
				Source: &cloudbuildpb.Source_GitSource{
					GitSource: &cloudbuildpb.GitSource{
						Url:      "https://github.com/test/repo.git",
						Dir:      "test-source",
						Revision: "test-hash",
					},
				},
			},
			dockerfile: false,
			envURL:     "https://github.com/test/repo.git",
			envHash:    "test-hash",
		},
		{
			name:       "Missing URL",
			opts:       Options{ProjectID: "test-project", Region: "us-central1", Service: "test-service", Source: "test-source"},
			wantErr:    true,
			wantErrMsg: URLVar + " is not defined",
		},
		{
			name:       "Missing Hash",
			opts:       Options{ProjectID: "test-project", Region: "us-central1", Service: "test-service", Source: "test-source"},
			wantErr:    true,
			wantErrMsg: HashVar + " is not defined",
			envURL:     "https://github.com/test/repo.git",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dockerfile {
				if err := os.MkdirAll(filepath.Join(tt.opts.Source, "Dockerfile"), 0o755); err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(filepath.Join(tt.opts.Source, "Dockerfile"))
			}
			if tt.envURL != "" {
				os.Setenv(URLVar, tt.envURL)
			} else {
				os.Unsetenv(URLVar)
			}
			if tt.envHash != "" {
				os.Setenv(HashVar, tt.envHash)
			} else {
				os.Unsetenv(HashVar)
			}
			got, err := CreateBuildRequest(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBuildRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.wantErrMsg {
					t.Errorf("CreateBuildRequest() error = %v, wantErrMsg %v", err, tt.wantErrMsg)
				}
				return
			}
			if got.Parent != tt.wantParent {
				t.Errorf("CreateBuildRequest() got.Parent = %v, want %v", got.Parent, tt.wantParent)
			}
			if got.ProjectId != tt.opts.ProjectID {
				t.Errorf("CreateBuildRequest() got.ProjectId = %v, want %v", got.ProjectId, tt.opts.ProjectID)
			}
			if got.Build.Images[0] != tt.wantImage {
				t.Errorf("CreateBuildRequest() got.Build.Images[0] = %v, want %v", got.Build.Images[0], tt.wantImage)
			}
			if !proto.Equal(got.Build.Source, tt.wantSource) {
				t.Errorf("CreateBuildRequest() got.Build.Source = %v, want %v", got.Build.Source, tt.wantSource)
			}
			if len(got.Build.Steps) != len(tt.wantSteps) {
				t.Errorf("CreateBuildRequest() got.Build.Steps = %v, want %v", got.Build.Steps, tt.wantSteps)
			}
			for i := range got.Build.Steps {
				if !proto.Equal(got.Build.Steps[i], tt.wantSteps[i]) {
					t.Errorf("CreateBuildRequest() got.Build.Steps[%d] = %v, want %v", i, got.Build.Steps[i], tt.wantSteps[i])
				}
			}
		})
	}
}
