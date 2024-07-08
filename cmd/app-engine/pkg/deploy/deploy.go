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
	"cmp"
	"context"
	"fmt"
	"log"
	"os"

	appengine "cloud.google.com/go/appengine/apiv1"
	"cloud.google.com/go/appengine/apiv1/appenginepb"
	"cloud.google.com/go/storage"
	appYAML "github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/appyaml"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/upload"
	"github.com/GoogleCloudBuild/cicd-images/cmd/app-engine/pkg/version"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const DefaultServiceID = "default"

// PromoteVersion promotes the specified version of an App Engine service
// to receive all traffic.
// The function uses a long-running operation to perform the promotion.
func PromoteVersion(ctx context.Context, client *appengine.ServicesClient, serviceName, versionID string) (*appenginepb.Service, error) {
	req := &appenginepb.UpdateServiceRequest{
		Name: serviceName,
		Service: &appenginepb.Service{
			Split: &appenginepb.TrafficSplit{
				// Key: version, value: proportion of traffic routed to it
				Allocations: map[string]float64{versionID: 1.0},
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"split"},
		},
		MigrateTraffic: false,
	}

	op, err := client.UpdateService(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error promoting version: %w", err)
	}
	resp, err := op.Wait(ctx)
	return resp, err
}

// CreateVersion deploys the given version to the specified App Engine service.
// It returns a pointer to the deployed version or an error if the deployment fails.
// The function uses a long-running operation to perform the deployment.
func CreateVersion(ctx context.Context, client *appengine.VersionsClient, serviceName string, version *appenginepb.Version) (*appenginepb.Version, error) {
	req := &appenginepb.CreateVersionRequest{
		Parent:  serviceName,
		Version: version,
	}

	op, err := client.CreateVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	return op.Wait(ctx)
}

// CreateAndDeployVersion parses given runtime configuration and deploys
// the given version to the specified App Engine service.
// It returns an error if the deployment fails.
// The function uses long-running operations to perform the deployment.
func CreateAndDeployVersion(ctx context.Context, versionClient *appengine.VersionsClient, servicesClient *appengine.ServicesClient, storageClient *storage.Client, opts config.AppEngineDeployOptions) error {
	appYAML, err := parseYAMLFromPath(opts)
	if err != nil {
		return err
	}

	serviceID := cmp.Or(appYAML.Service, DefaultServiceID)                       // e.g. "default"
	serviceName := fmt.Sprintf("apps/%s/services/%s", opts.ProjectID, serviceID) // e.g. "apps/my-project/services/default"
	if opts.ImageURL == "" {
		objectName := fmt.Sprintf("%s-%s.zip", serviceID, opts.VersionID)
		// Fall back to default bucket if not specified
		opts.Bucket = cmp.Or(opts.Bucket, fmt.Sprintf("staging.%s.appspot.com", opts.ProjectID))
		if err = upload.ToGCSZipped(ctx, storageClient, opts.Bucket, objectName, "."); err != nil {
			return err
		}
		opts.SourceURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", opts.Bucket, objectName)
	}

	version, err := version.NewVersion(appYAML, opts)
	if err != nil {
		return fmt.Errorf("failed to construct version struct: %w", err)
	}

	return callDeploy(ctx, serviceName, version, versionClient, servicesClient, opts)
}

func parseYAMLFromPath(opts config.AppEngineDeployOptions) (*appYAML.AppYAML, error) {
	appYAMLBytes, err := os.ReadFile(opts.AppYAMLPath)
	if err != nil {
		return nil, err
	}
	appYAML, err := appYAML.ParseAppYAML(appYAMLBytes)
	if err != nil {
		return nil, err
	}
	return appYAML, nil
}

func callDeploy(ctx context.Context, serviceName string, version *appenginepb.Version, versionClient *appengine.VersionsClient, servicesClient *appengine.ServicesClient, opts config.AppEngineDeployOptions) error {
	log.Printf("Deploying service: %s, version:, %s", serviceName, version)
	version, err := CreateVersion(ctx, versionClient, serviceName, version)
	if err != nil {
		return fmt.Errorf("error during version creation: %w", err)
	}
	log.Printf("Version created successfully: %v", version)

	if opts.Promote && servicesClient != nil {
		updateServiceResponse, err := PromoteVersion(ctx, servicesClient, serviceName, version.Id)
		if err != nil {
			return fmt.Errorf("error during promotion: %w", err)
		}
		log.Println("Promoted latest version", updateServiceResponse)
	}
	return nil
}
