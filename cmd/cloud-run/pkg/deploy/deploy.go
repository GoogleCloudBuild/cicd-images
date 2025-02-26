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
	"path/filepath"
	"strings"
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

// updateWithOptions updates the image and configuration of an existing Cloud Run Service object
// based on the provided config.DeployOptions.
func updateWithOptions(service *run.Service, opts config.DeployOptions) {
	service.Spec.Template.Spec.Containers[0].Image = opts.Image

	// Handle environment variables
	container := service.Spec.Template.Spec.Containers[0]
	if opts.ClearEnvVars {
		container.Env = nil
	} else if opts.EnvVars != nil && len(opts.EnvVars) > 0 {
		container.Env = make([]*run.EnvVar, 0, len(opts.EnvVars))
		for k, v := range opts.EnvVars {
			container.Env = append(container.Env, &run.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	} else {
		if opts.RemoveEnvVars != nil && len(opts.RemoveEnvVars) > 0 && container.Env != nil {
			// Remove specified environment variables
			newEnv := make([]*run.EnvVar, 0, len(container.Env))
			for _, env := range container.Env {
				shouldKeep := true
				for _, remove := range opts.RemoveEnvVars {
					if env.Name == remove {
						shouldKeep = false
						break
					}
				}
				if shouldKeep {
					newEnv = append(newEnv, env)
				}
			}
			container.Env = newEnv
		}
		if opts.UpdateEnvVars != nil && len(opts.UpdateEnvVars) > 0 {
			// Update or add new environment variables
			if container.Env == nil {
				container.Env = make([]*run.EnvVar, 0, len(opts.UpdateEnvVars))
			}
			for k, v := range opts.UpdateEnvVars {
				found := false
				for _, env := range container.Env {
					if env.Name == k {
						env.Value = v
						found = true
						break
					}
				}
				if !found {
					container.Env = append(container.Env, &run.EnvVar{
						Name:  k,
						Value: v,
					})
				}
			}
		}
	}

	// Handle secrets
	if opts.ClearSecrets {
		container.EnvFrom = nil
		container.VolumeMounts = nil
		service.Spec.Template.Spec.Volumes = nil
	} else if opts.Secrets != nil && len(opts.Secrets) > 0 {
		container.EnvFrom = make([]*run.EnvFromSource, 0)
		container.VolumeMounts = make([]*run.VolumeMount, 0)
		service.Spec.Template.Spec.Volumes = make([]*run.Volume, 0)

		for k, v := range opts.Secrets {
			if k[0] == '/' {
				// Mount secret as volume
				mountPath := k
				secretName := v[:strings.Index(v, ":")]
				version := v[strings.Index(v, ":")+1:]

				container.VolumeMounts = append(container.VolumeMounts, &run.VolumeMount{
					Name:      secretName,
					MountPath: mountPath,
				})

				service.Spec.Template.Spec.Volumes = append(service.Spec.Template.Spec.Volumes, &run.Volume{
					Name: secretName,
					Secret: &run.SecretVolumeSource{
						SecretName: secretName,
						Items: []*run.KeyToPath{{
							Key:  version,
							Path: filepath.Base(mountPath),
						}},
					},
				})
			} else {
				// Set as environment variable
				container.EnvFrom = append(container.EnvFrom, &run.EnvFromSource{
					SecretRef: &run.SecretEnvSource{
						LocalObjectReference: &run.LocalObjectReference{
							Name: v[:strings.Index(v, ":")],
						},
					},
				})
			}
		}
	} else {
		if opts.RemoveSecrets != nil && len(opts.RemoveSecrets) > 0 {
			// Remove specified secrets
			if container.EnvFrom != nil {
				newEnvFrom := make([]*run.EnvFromSource, 0, len(container.EnvFrom))
				for _, envFrom := range container.EnvFrom {
					if envFrom.SecretRef != nil {
						shouldKeep := true
						for _, remove := range opts.RemoveSecrets {
							if envFrom.SecretRef.LocalObjectReference.Name == remove {
								shouldKeep = false
								break
							}
						}
						if shouldKeep {
							newEnvFrom = append(newEnvFrom, envFrom)
						}
					} else {
						newEnvFrom = append(newEnvFrom, envFrom)
					}
				}
				container.EnvFrom = newEnvFrom
			}
		}
		if opts.UpdateSecrets != nil && len(opts.UpdateSecrets) > 0 {
			// Update or add new secrets
			for k, v := range opts.UpdateSecrets {
				if k[0] == '/' {
					// Mount secret as volume
					mountPath := k
					secretName := v[:strings.Index(v, ":")]
					version := v[strings.Index(v, ":")+1:]

					// Update existing volume mount or add new one
					found := false
					if container.VolumeMounts != nil {
						for _, mount := range container.VolumeMounts {
							if mount.MountPath == mountPath {
								mount.Name = secretName
								found = true
								break
							}
						}
					}
					if !found {
						if container.VolumeMounts == nil {
							container.VolumeMounts = make([]*run.VolumeMount, 0)
						}
						container.VolumeMounts = append(container.VolumeMounts, &run.VolumeMount{
							Name:      secretName,
							MountPath: mountPath,
						})
					}

					// Update existing volume or add new one
					found = false
					if service.Spec.Template.Spec.Volumes != nil {
						for _, vol := range service.Spec.Template.Spec.Volumes {
							if vol.Name == secretName {
								vol.Secret.SecretName = secretName
								vol.Secret.Items[0].Key = version
								found = true
								break
							}
						}
					}
					if !found {
						if service.Spec.Template.Spec.Volumes == nil {
							service.Spec.Template.Spec.Volumes = make([]*run.Volume, 0)
						}
						service.Spec.Template.Spec.Volumes = append(service.Spec.Template.Spec.Volumes, &run.Volume{
							Name: secretName,
							Secret: &run.SecretVolumeSource{
								SecretName: secretName,
								Items: []*run.KeyToPath{{
									Key:  version,
									Path: filepath.Base(mountPath),
								}},
							},
						})
					}
				} else {
					// Set as environment variable
					if container.EnvFrom == nil {
						container.EnvFrom = make([]*run.EnvFromSource, 0)
					}
					found := false
					for _, envFrom := range container.EnvFrom {
						if envFrom.SecretRef != nil && envFrom.SecretRef.LocalObjectReference.Name == k {
							envFrom.SecretRef.LocalObjectReference.Name = v[:strings.Index(v, ":")]
							found = true
							break
						}
					}
					if !found {
						container.EnvFrom = append(container.EnvFrom, &run.EnvFromSource{
							SecretRef: &run.SecretEnvSource{
								LocalObjectReference: &run.LocalObjectReference{
									Name: v[:strings.Index(v, ":")],
								},
							},
						})
					}
				}
			}
		}
	}
}

// buildServiceDefinition creates a new Cloud Run Service object based on the
// provided projectID and DeployOptions.
func buildServiceDefinition(projectID string, opts config.DeployOptions) run.Service {
	container := &run.Container{Image: opts.Image}

	// Set environment variables if provided and not empty
	if opts.EnvVars != nil && len(opts.EnvVars) > 0 {
		container.Env = make([]*run.EnvVar, 0, len(opts.EnvVars))
		for k, v := range opts.EnvVars {
			container.Env = append(container.Env, &run.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	// Set secrets if provided and not empty
	var volumes []*run.Volume
	if opts.Secrets != nil && len(opts.Secrets) > 0 {
		container.EnvFrom = make([]*run.EnvFromSource, 0)
		container.VolumeMounts = make([]*run.VolumeMount, 0)
		volumes = make([]*run.Volume, 0)

		for k, v := range opts.Secrets {
			if k[0] == '/' {
				// Mount secret as volume
				mountPath := k
				secretName := v[:strings.Index(v, ":")]
				version := v[strings.Index(v, ":")+1:]

				container.VolumeMounts = append(container.VolumeMounts, &run.VolumeMount{
					Name:      secretName,
					MountPath: mountPath,
				})

				volumes = append(volumes, &run.Volume{
					Name: secretName,
					Secret: &run.SecretVolumeSource{
						SecretName: secretName,
						Items: []*run.KeyToPath{{
							Key:  version,
							Path: filepath.Base(mountPath),
						}},
					},
				})
			} else {
				// Set as environment variable
				container.EnvFrom = append(container.EnvFrom, &run.EnvFromSource{
					SecretRef: &run.SecretEnvSource{
						LocalObjectReference: &run.LocalObjectReference{
							Name: v[:strings.Index(v, ":")],
						},
					},
				})
			}
		}
	}

	rService := run.Service{
		ApiVersion: "serving.knative.dev/v1",
		Kind:       "Service",
		Metadata:   &run.ObjectMeta{Namespace: projectID, Name: opts.Service},
		Spec: &run.ServiceSpec{
			Template: &run.RevisionTemplate{
				Spec: &run.RevisionSpec{
					Containers: []*run.Container{container},
					// Only set volumes if we have any
					Volumes: volumes,
				},
			},
		},
	}

	return rService
}
