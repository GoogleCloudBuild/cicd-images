// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"
	"time"

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-run/pkg/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-run/pkg/utils"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/run/v1"
)

// CreateOrUpdateService deploys a service to Cloud run. If the service
// doesn't exist, it creates a new one. If the service exists, it updates the
// existing service with the config.DeployOptions.
func CreateOrUpdateService(runAPIClient *run.APIService, projectID, region string, opts config.DeployOptions) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)
	resourceName := fmt.Sprintf("%s/services/%s", parent, opts.Service)
	log.Printf("Deploying container to Cloud Run service [%s] in project [%s] region [%s]\n", opts.Service, projectID, region)

	existingService, err := runAPIClient.Projects.Locations.Services.Get(resourceName).Do()
	if err != nil {
		gErr, ok := err.(*googleapi.Error)
		if !ok || gErr.Code != http.StatusNotFound {
			return err
		}
		serviceDefinition := buildServiceDefinition(projectID, opts)
		log.Printf("Creating a new serivce %s\n", opts.Service)
		createCall := runAPIClient.Projects.Locations.Services.Create(parent, &serviceDefinition)
		_, err = createCall.Do()
		if err != nil {
			return err
		}
	} else {
		log.Printf("Replacing the existing serivce %s\n", opts.Service)
		updateWithOptions(existingService, opts)
		replaceCall := runAPIClient.Projects.Locations.Services.ReplaceService(resourceName, existingService)
		_, err = replaceCall.Do()
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForServiceReady waits for a Cloud Run service to reach a ready state
// by polling its status.
func WaitForServiceReady(ctx context.Context, runAPIClient *run.APIService, projectID, region, service string) error {
	resourceName := fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, region, service)
	if err := utils.PollWithInterval(ctx, time.Minute*2, time.Second, func() (bool, error) {
		runService, err := runAPIClient.Projects.Locations.Services.Get(resourceName).Do()
		if err != nil {
			return false, err
		}

		// Clients polling for completed reconciliation should poll until
		// observedGeneration = metadata.generation and the Ready condition's
		// status is True or False
		// see details in: https://github.com/googleapis/google-api-go-client/blob/v0.169.0/run/v1/run-gen.go#L961-L967
		if runService.Status.ObservedGeneration != runService.Metadata.Generation {
			return false, nil
		}
		for _, c := range runService.Status.Conditions {
			if c.Type == "Ready" {
				log.Println(c.Message)
				if c.Status == "True" {
					return true, nil
				}
				if c.Status == "False" {
					return false, fmt.Errorf("failed to deploy the latest revision of the service %s", service)
				}
			}
		}
		return false, nil
	}); err != nil {
		return err
	}
	runService, err := runAPIClient.Projects.Locations.Services.Get(resourceName).Do()
	log.Printf("Service [%s] with revision [%s] is deployed successfully, serving %d percent of traffic.\n",
		service, runService.Status.LatestReadyRevisionName, runService.Status.Traffic[0].Percent)
	log.Printf("Service URL: %s \n ", runService.Status.Url)
	if err != nil {
		return err
	}
	return nil
}

// updateWithOptions updates the image of an existing Cloud Run Service object
// based on the provided config.DeployOptions.
func updateWithOptions(service *run.Service, opts config.DeployOptions) {
	service.Spec.Template.Spec.Containers[0].Image = opts.Image
}

// buildServiceDefinition creates a new Cloud Run Service object based on the
// provided projectID and DeployOptions.
func buildServiceDefinition(projectID string, opts config.DeployOptions) run.Service {
	rService := run.Service{
		ApiVersion: "serving.knative.dev/v1",
		Kind:       "Service",
		Metadata:   &run.ObjectMeta{Namespace: projectID, Name: opts.Service},
		Spec: &run.ServiceSpec{
			Template: &run.RevisionTemplate{
				Spec: &run.RevisionSpec{
					Containers: []*run.Container{{Image: opts.Image}},
				},
			},
		},
	}

	return rService
}
