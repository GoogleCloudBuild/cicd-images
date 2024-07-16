package build

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	"cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
)

type Options struct {
	ProjectID string
	Region    string
	Service   string
	Source    string
}

const (
	URLVar  = "CI_REPOSITORY_URL"
	HashVar = "CI_COMMIT_SHA"
)

// Run builds a container image using Cloud Build and returns the image URI.
func Run(ctx context.Context, c *cloudbuild.Client, opts Options) (string, error) {
	req, err := CreateBuildRequest(opts)
	if err != nil {
		return "", err
	}
	log.Println("Created Build Request:", req)
	op, err := c.CreateBuild(ctx, req)
	if err != nil {
		return "", err
	}

	b, err := op.Wait(ctx)
	if b != nil {
		log.Println("Build log:", b.LogUrl)
	}
	if err != nil {
		return "", err
	}
	if len(b.Images) != 1 {
		return "", fmt.Errorf("expected 1 image, got %d", len(b.Images))
	}
	log.Println("Build completed:", b)

	image := b.Images[0]
	log.Println("Built image:", image)
	return image, nil
}

// CreateBuildRequest creates a Cloud Build request to build a container image.
// Default values are taken from gcloud.
// Environment variables are GitLab CI/CD specific.
func CreateBuildRequest(opts Options) (*cloudbuildpb.CreateBuildRequest, error) {
	image := fmt.Sprintf("%s-docker.pkg.dev/%s/cloud-run-source-deploy/%s", opts.Region, opts.ProjectID, opts.Service)
	parent := fmt.Sprintf("projects/%s/locations/global", opts.ProjectID)

	var steps []*cloudbuildpb.BuildStep
	if _, err := os.Stat(filepath.Join(opts.Source, "Dockerfile")); err == nil {
		step := &cloudbuildpb.BuildStep{Name: "gcr.io/cloud-builders/docker", Args: []string{"build", "-t", image, "."}}
		steps = append(steps, step)
	} else {
		step := &cloudbuildpb.BuildStep{Name: "gcr.io/k8s-skaffold/pack", Args: []string{"build", image, "--builder", "gcr.io/buildpacks/builder:v1"}}
		steps = append(steps, step)
	}
	url := os.Getenv(URLVar)
	if url == "" {
		return nil, fmt.Errorf(URLVar + " is not defined")
	}
	hash := os.Getenv(HashVar)
	if hash == "" {
		return nil, fmt.Errorf(HashVar + " is not defined")
	}
	remoteSource := &cloudbuildpb.Source{
		Source: &cloudbuildpb.Source_GitSource{
			GitSource: &cloudbuildpb.GitSource{
				Url:      url,
				Dir:      opts.Source,
				Revision: hash,
			},
		},
	}

	req := &cloudbuildpb.CreateBuildRequest{
		Parent:    parent,
		ProjectId: opts.ProjectID,
		Build: &cloudbuildpb.Build{
			Source: remoteSource,
			Steps:  steps,
			Images: []string{image},
		},
	}
	return req, nil
}
