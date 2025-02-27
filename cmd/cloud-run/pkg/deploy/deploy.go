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

	"github.com/dijarvrella/cicd-images/cmd/cloud-run/pkg/config"
	"github.com/dijarvrella/cicd-images/cmd/cloud-run/pkg/utils"

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

// parseSecretReference parses a secret reference in the format projects/PROJECT_ID/secrets/SECRET_NAME/versions/VERSION
// or in the format SECRET_NAME:VERSION
// and returns the secret name and version
func parseSecretReference(secretRef string) (secretName, version string) {
	log.Printf("Parsing secret reference: %s", secretRef)

	// Check if it's in the format projects/PROJECT_ID/secrets/SECRET_NAME/versions/VERSION
	if strings.Contains(secretRef, "projects/") && strings.Contains(secretRef, "/secrets/") {
		parts := strings.Split(secretRef, "/secrets/")
		if len(parts) == 2 {
			// Get everything up to the /versions/ part if it exists
			secretVersionParts := strings.Split(parts[1], "/versions/")
			secretName = secretVersionParts[0]

			// Handle version
			if len(secretVersionParts) >= 2 {
				version = secretVersionParts[1]
			} else {
				version = "latest" // Default to latest if not specified
			}

			log.Printf("Parsed from project format - Secret name: %s, Version: %s", secretName, version)
			return
		}
	}

	// Fall back to simpler parsing (SECRET_NAME:VERSION)
	parts := strings.Split(secretRef, ":")
	if len(parts) >= 1 {
		secretName = parts[0]
		if len(parts) >= 2 {
			version = parts[1]
		} else {
			version = "latest" // Default to latest if not specified
		}
	}

	log.Printf("Parsed from simple format - Secret name: %s, Version: %s", secretName, version)
	return
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
		// Clear secret environment variables
		if container.Env != nil {
			// Filter out any secret-referenced env vars
			newEnv := make([]*run.EnvVar, 0, len(container.Env))
			for _, env := range container.Env {
				if env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil {
					newEnv = append(newEnv, env)
				}
			}
			container.Env = newEnv
		}

		// Clear volume mounts and volumes
		container.VolumeMounts = nil
		service.Spec.Template.Spec.Volumes = nil
	} else if opts.Secrets != nil && len(opts.Secrets) > 0 {
		// Clear existing secret-referenced env vars
		if container.Env != nil {
			newEnv := make([]*run.EnvVar, 0, len(container.Env))
			for _, env := range container.Env {
				if env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil {
					newEnv = append(newEnv, env)
				}
			}
			container.Env = newEnv
		} else {
			container.Env = make([]*run.EnvVar, 0)
		}

		// Clear existing volume mounts and volumes
		container.VolumeMounts = make([]*run.VolumeMount, 0)
		service.Spec.Template.Spec.Volumes = make([]*run.Volume, 0)

		// Add new secrets
		for k, v := range opts.Secrets {
			if k[0] == '/' {
				// Mount secret as volume
				mountPath := k
				secretName, version := parseSecretReference(v)

				// Only proceed if we have a valid secretName
				if secretName != "" {
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
				}
			} else {
				// Set as environment variable using ValueFrom
				secretName, version := parseSecretReference(v)

				// Only proceed if we have a valid secretName
				if secretName != "" {
					log.Printf("Adding secret env var %s with secret %s and version %s", k, secretName, version)

					// Create a properly initialized SecretKeySelector with all required fields
					secretKeyRef := &run.SecretKeySelector{
						Key:  version,
						Name: secretName,
						LocalObjectReference: &run.LocalObjectReference{
							Name: secretName,
						},
					}

					// Use direct struct initialization to avoid null JSON fields
					envVar := &run.EnvVar{
						Name:  k,
						Value: "",
						ValueFrom: &run.EnvVarSource{
							SecretKeyRef: secretKeyRef,
						},
					}

					// Add to environment variables
					container.Env = append(container.Env, envVar)

					// Log the resulting structure to verify it's set correctly
					log.Printf("Secret reference set: Name=%s, Key=%s",
						secretName, version)
				}
			}
		}
	} else {
		if opts.RemoveSecrets != nil && len(opts.RemoveSecrets) > 0 {
			// Remove specified secrets from environment variables
			if container.Env != nil {
				newEnv := make([]*run.EnvVar, 0, len(container.Env))
				for _, env := range container.Env {
					shouldKeep := true
					if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
						secretName := env.ValueFrom.SecretKeyRef.LocalObjectReference.Name
						for _, remove := range opts.RemoveSecrets {
							if secretName == remove {
								shouldKeep = false
								break
							}
						}
					}
					if shouldKeep {
						newEnv = append(newEnv, env)
					}
				}
				container.Env = newEnv
			}

			// Remove specified secrets from volume mounts and volumes
			if container.VolumeMounts != nil {
				newVolumeMounts := make([]*run.VolumeMount, 0, len(container.VolumeMounts))
				for _, mount := range container.VolumeMounts {
					shouldKeep := true
					for _, remove := range opts.RemoveSecrets {
						if mount.Name == remove {
							shouldKeep = false
							break
						}
					}
					if shouldKeep {
						newVolumeMounts = append(newVolumeMounts, mount)
					}
				}
				container.VolumeMounts = newVolumeMounts
			}

			if service.Spec.Template.Spec.Volumes != nil {
				newVolumes := make([]*run.Volume, 0, len(service.Spec.Template.Spec.Volumes))
				for _, vol := range service.Spec.Template.Spec.Volumes {
					shouldKeep := true
					for _, remove := range opts.RemoveSecrets {
						if vol.Name == remove {
							shouldKeep = false
							break
						}
					}
					if shouldKeep {
						newVolumes = append(newVolumes, vol)
					}
				}
				service.Spec.Template.Spec.Volumes = newVolumes
			}
		}

		if opts.UpdateSecrets != nil && len(opts.UpdateSecrets) > 0 {
			// Update or add new secrets
			for k, v := range opts.UpdateSecrets {
				if k[0] == '/' {
					// Mount secret as volume
					mountPath := k
					secretName, version := parseSecretReference(v)

					// Only proceed if we have a valid secretName
					if secretName != "" {
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
					}
				} else {
					// Set as environment variable using ValueFrom
					secretName, version := parseSecretReference(v)

					// Only proceed if we have a valid secretName
					if secretName != "" {
						// Check if the environment variable already exists
						found := false
						if container.Env != nil {
							for _, env := range container.Env {
								if env.Name == k {
									if env.ValueFrom == nil {
										env.Value = ""
										env.ValueFrom = &run.EnvVarSource{
											SecretKeyRef: &run.SecretKeySelector{
												Key:  version,
												Name: secretName,
												LocalObjectReference: &run.LocalObjectReference{
													Name: secretName,
												},
											},
										}
									} else if env.ValueFrom.SecretKeyRef != nil {
										env.ValueFrom.SecretKeyRef.Key = version
										env.ValueFrom.SecretKeyRef.Name = secretName
										env.ValueFrom.SecretKeyRef.LocalObjectReference.Name = secretName
									}
									found = true
									break
								}
							}
						}

						if !found {
							if container.Env == nil {
								container.Env = make([]*run.EnvVar, 0)
							}
							container.Env = append(container.Env, &run.EnvVar{
								Name: k,
								ValueFrom: &run.EnvVarSource{
									SecretKeyRef: &run.SecretKeySelector{
										Key:  version,
										Name: secretName,
										LocalObjectReference: &run.LocalObjectReference{
											Name: secretName,
										},
									},
								},
							})
						}
					}
				}
			}
		}
	}

	// Set ingress traffic policy if specified
	if opts.Ingress != "" {
		// Make sure metadata annotations exist
		if service.Metadata.Annotations == nil {
			service.Metadata.Annotations = make(map[string]string)
		}

		// Set the ingress value
		switch opts.Ingress {
		case "internal":
			service.Metadata.Annotations["run.googleapis.com/ingress"] = "internal"
		case "internal-and-cloud-load-balancing":
			service.Metadata.Annotations["run.googleapis.com/ingress"] = "internal-and-cloud-load-balancing"
		default: // "all" is the default
			service.Metadata.Annotations["run.googleapis.com/ingress"] = "all"
		}
		log.Printf("Setting ingress to: %s", opts.Ingress)
	}

	// Set authentication policy
	if service.Metadata.Annotations == nil {
		service.Metadata.Annotations = make(map[string]string)
	}

	if opts.AllowUnauthenticated {
		service.Metadata.Annotations["run.googleapis.com/ingress-status"] = "all"
		log.Println("Allowing unauthenticated access")
	} else {
		service.Metadata.Annotations["run.googleapis.com/ingress-status"] = "internal-and-cloud-load-balancing"
		log.Println("Requiring authentication for access")
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
	} else {
		// Initialize an empty env array to avoid null in JSON
		container.Env = make([]*run.EnvVar, 0)
	}

	// Set secrets if provided and not empty
	var volumes []*run.Volume
	if opts.Secrets != nil && len(opts.Secrets) > 0 {
		// Make sure Env is initialized
		if container.Env == nil {
			container.Env = make([]*run.EnvVar, 0)
		}
		container.VolumeMounts = make([]*run.VolumeMount, 0)
		volumes = make([]*run.Volume, 0)

		for k, v := range opts.Secrets {
			if k[0] == '/' {
				// Mount secret as volume
				mountPath := k
				secretName, version := parseSecretReference(v)

				// Only proceed if we have a valid secretName
				if secretName != "" {
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
				}
			} else {
				// Set as environment variable using ValueFrom
				secretName, version := parseSecretReference(v)

				// Only proceed if we have a valid secretName
				if secretName != "" {
					log.Printf("Building service: Adding secret env var %s with secret %s and version %s", k, secretName, version)

					// Create a properly initialized SecretKeySelector with all required fields
					secretKeyRef := &run.SecretKeySelector{
						Key:  version,
						Name: secretName,
						LocalObjectReference: &run.LocalObjectReference{
							Name: secretName,
						},
					}

					// Use explicit field initialization to avoid null values in JSON
					container.Env = append(container.Env, &run.EnvVar{
						Name:  k,
						Value: "", // Ensure value is empty string, not null
						ValueFrom: &run.EnvVarSource{
							SecretKeyRef: secretKeyRef,
						},
					})

					// Log the resulting structure to verify it's set correctly
					log.Printf("Secret reference set for new service: Name=%s, Key=%s",
						secretName, version)
				}
			}
		}
	}

	// Create the service
	rService := run.Service{
		ApiVersion: "serving.knative.dev/v1",
		Kind:       "Service",
		Metadata: &run.ObjectMeta{
			Namespace:   projectID,
			Name:        opts.Service,
			Annotations: make(map[string]string),
		},
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

	// Set ingress traffic policy
	switch opts.Ingress {
	case "internal":
		rService.Metadata.Annotations["run.googleapis.com/ingress"] = "internal"
		log.Printf("Setting ingress to: internal")
	case "internal-and-cloud-load-balancing":
		rService.Metadata.Annotations["run.googleapis.com/ingress"] = "internal-and-cloud-load-balancing"
		log.Printf("Setting ingress to: internal-and-cloud-load-balancing")
	default: // "all" is the default
		rService.Metadata.Annotations["run.googleapis.com/ingress"] = "all"
		log.Printf("Setting ingress to: all")
	}

	// Set authentication policy
	if opts.AllowUnauthenticated {
		rService.Metadata.Annotations["run.googleapis.com/ingress-status"] = "all"
		log.Println("Allowing unauthenticated access")
	} else {
		rService.Metadata.Annotations["run.googleapis.com/ingress-status"] = "internal-and-cloud-load-balancing"
		log.Println("Requiring authentication for access")
	}

	return rService
}
