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

package main

import (
	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	"github.com/dijarvrella/cicd-images/cmd/cloud-run/pkg/build"
	"github.com/dijarvrella/cicd-images/cmd/cloud-run/pkg/config"
	"github.com/dijarvrella/cicd-images/cmd/cloud-run/pkg/deploy"

	"fmt"
	"log"
	"unicode"

	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"google.golang.org/api/run/v2"
)

var opts config.DeployOptions

func NewDeployCmd() *cobra.Command {
	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Create or update a Cloud Run service",
		Long: `Deploy creates a new Cloud Run service or updates an existing one.

You can deploy from a container image or source code:
- From container image: Use --image flag
- From source code: Use --source flag (defaults to current directory)

Environment Variables:
- --set-env-vars KEY=VALUE     Set environment variables (removes existing ones)
- --env-vars-file FILE         Load environment variables from a YAML file
- --remove-env-vars KEY        Remove specific environment variables
- --update-env-vars KEY=VALUE  Update or add new environment variables
- --clear-env-vars            Remove all environment variables

Secrets:
- --set-secrets KEY=SECRET:VERSION    Set secrets (removes existing ones)
  For mounted volumes: /path/to/mount=SECRET:VERSION
  For env vars: ENV_VAR=SECRET:VERSION
- --remove-secrets KEY        Remove specific secrets
- --update-secrets KEY=VALUE  Update or add new secrets
- --clear-secrets            Remove all secrets

Access and Traffic Configuration:
- --allow-unauthenticated      Allow unauthenticated access to the service
- --no-allow-unauthenticated   Require authentication for access to the service
- --ingress TYPE               Set the ingress traffic settings (all, internal, internal-and-cloud-load-balancing)
- --default-url                Use the default URL for the service (default)
- --no-default-url             Disable the default URL for the service

VPC Connectivity:
- --vpc-connector              The VPC connector to use for this service
- --vpc-network                The VPC network to connect to
- --vpc-subnetwork             The VPC subnetwork to connect to
- --vpc-egress                 VPC egress setting (private-ranges-only or all-traffic)

Examples:
  # Deploy from container image
  cloud-run deploy --project-id=my-project --region=us-central1 --service=myapp --image=gcr.io/myproject/myapp:v1

  # Deploy from source code
  cloud-run deploy --project-id=my-project --region=us-central1 --service=myapp --source=./src

  # Set environment variables
  cloud-run deploy --service=myapp --set-env-vars=DB_HOST=localhost,DB_PORT=5432

  # Mount a secret as a volume
  cloud-run deploy --service=myapp --set-secrets=/secrets/api/key=mysecret:latest

  # Set a secret as an environment variable
  cloud-run deploy --service=myapp --set-secrets=API_KEY=mysecret:1

  # Deploy internal service requiring authentication
  cloud-run deploy --service=myapp --image=gcr.io/myproject/myapp:v1 --no-allow-unauthenticated --ingress internal

  # Deploy service without a default URL
  cloud-run deploy --service=myapp --image=gcr.io/myproject/myapp:v1 --no-default-url`,
		RunE: deployService,
	}
	deployCmd.Flags().StringVar(&opts.Image, "image", "", "The container image to deploy (e.g., gcr.io/project/image:tag)")
	deployCmd.Flags().StringVar(&opts.Service, "service", "", "The name of the Cloud Run service to create or update")
	deployCmd.Flags().StringVar(&opts.Source, "source", ".", "The source directory to deploy from")

	deployCmd.Flags().StringToStringVar(&opts.EnvVars, "set-env-vars", nil, "List of key-value pairs to set as environment variables (removes existing ones)")
	deployCmd.Flags().StringVar(&opts.EnvVarsFile, "env-vars-file", "", "Path to a local YAML file with environment variable definitions")
	deployCmd.Flags().StringSliceVar(&opts.RemoveEnvVars, "remove-env-vars", nil, "List of environment variables to remove")
	deployCmd.Flags().StringToStringVar(&opts.UpdateEnvVars, "update-env-vars", nil, "List of key-value pairs to update or add as environment variables")
	deployCmd.Flags().BoolVar(&opts.ClearEnvVars, "clear-env-vars", false, "Remove all environment variables")

	deployCmd.Flags().StringToStringVar(&opts.Secrets, "set-secrets", nil, "List of key-value pairs to set as secrets (removes existing ones)")
	deployCmd.Flags().StringSliceVar(&opts.RemoveSecrets, "remove-secrets", nil, "List of secrets to remove")
	deployCmd.Flags().StringToStringVar(&opts.UpdateSecrets, "update-secrets", nil, "List of key-value pairs to update or add as secrets")
	deployCmd.Flags().BoolVar(&opts.ClearSecrets, "clear-secrets", false, "Remove all secrets")

	// Add access and traffic configuration flags
	deployCmd.Flags().BoolVar(&opts.AllowUnauthenticated, "allow-unauthenticated", true, "Allow unauthenticated access to the service")
	deployCmd.Flags().StringVar(&opts.Ingress, "ingress", "all", "Set the ingress traffic settings (all, internal, internal-and-cloud-load-balancing)")

	// Add both --default-url and --no-default-url flags as mutually exclusive
	deployCmd.Flags().BoolVar(&opts.DefaultURL, "default-url", true, "Use the default URL for the service")
	deployCmd.Flags().Bool("no-default-url", false, "Disable the default URL for the service")

	// Add VPC connectivity flags
	deployCmd.Flags().StringVar(&opts.VpcConnector, "vpc-connector", "", "The VPC connector to use for this service")
	deployCmd.Flags().StringVar(&opts.VpcNetwork, "vpc-network", "default", "The VPC network to connect to")
	deployCmd.Flags().StringVar(&opts.VpcSubnetwork, "vpc-subnetwork", "default", "The VPC subnetwork to connect to")
	deployCmd.Flags().StringVar(&opts.VpcEgress, "vpc-egress", "private-ranges-only", "VPC egress setting (private-ranges-only or all-traffic)")

	// Link the no-default-url flag to DefaultURL
	deployCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		noDefaultURL, _ := cmd.Flags().GetBool("no-default-url")
		if noDefaultURL {
			opts.DefaultURL = false
		}

		// Mark flags as mutually exclusive
		deployCmd.MarkFlagsMutuallyExclusive("default-url", "no-default-url")

		return nil
	}

	// Flag validations
	// Only one of these env var flags can be used at a time, but all are optional
	if deployCmd.Flags().Lookup("set-env-vars").Changed ||
		deployCmd.Flags().Lookup("update-env-vars").Changed ||
		deployCmd.Flags().Lookup("clear-env-vars").Changed ||
		deployCmd.Flags().Lookup("env-vars-file").Changed {
		deployCmd.MarkFlagsMutuallyExclusive("set-env-vars", "update-env-vars", "clear-env-vars", "env-vars-file")
	}

	// Only one of these secret flags can be used at a time, but all are optional
	if deployCmd.Flags().Lookup("set-secrets").Changed ||
		deployCmd.Flags().Lookup("update-secrets").Changed ||
		deployCmd.Flags().Lookup("clear-secrets").Changed {
		deployCmd.MarkFlagsMutuallyExclusive("set-secrets", "update-secrets", "clear-secrets")
	}

	// Add no-wait flag
	deployCmd.Flags().Bool("no-wait", false, "Skip waiting for service to be ready")

	_ = deployCmd.MarkFlagRequired("service")
	deployCmd.MarkFlagsOneRequired(
		"image",
		"source",
	)
	deployCmd.MarkFlagsMutuallyExclusive(
		"image",
		"source",
	)
	return deployCmd
}

func deployService(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Add more robust validation for service name to ensure it meets Cloud Run requirements
	if !isValidServiceName(opts.Service) {
		return fmt.Errorf("invalid service name: %s. Service names must contain only lowercase letters, numbers, and hyphens, must begin with a letter, cannot end with a hyphen, and must be less than 50 characters", opts.Service)
	}

	log.Printf("Service name validation passed for: %s", opts.Service)

	cloudbuildClient, err := cloudbuild.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cloudbuildClient.Close()
	if opts.Image == "" {
		// Source deploy is implemented as a combination of image build
		// and image deploy, same as gcloud.
		opts.Image, err = build.Run(ctx, cloudbuildClient, build.Options{
			ProjectID: projectID,
			Region:    region,
			Service:   opts.Service,
			Source:    opts.Source,
		})
		if err != nil {
			return err
		}
	}

	// Create a Cloud Run v2 service client
	runService, err := run.NewService(ctx, option.WithUserAgent(userAgent))
	if err != nil {
		return err
	}

	servicesClient := run.NewProjectsLocationsServicesService(runService)

	log.Printf("Using service name: %s", opts.Service)
	err = deploy.CreateOrUpdateServiceV2(servicesClient, projectID, region, opts)
	if err != nil {
		return err
	}

	// Check for no-wait flag
	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		// Wait for service to be ready
		if err := deploy.WaitForServiceReadyV2(ctx, servicesClient, projectID, region, opts.Service); err != nil {
			return err
		}
	}
	return nil
}

// isValidServiceName validates that a service name meets Cloud Run requirements:
// - Only lowercase letters, numbers, and hyphens
// - Must begin with a letter
// - Cannot end with a hyphen
// - Must be less than 50 characters
func isValidServiceName(name string) bool {
	if len(name) == 0 || len(name) >= 50 {
		log.Printf("Service name validation failed: length check. Name: %s, Length: %d", name, len(name))
		return false
	}

	// Must start with a letter
	if !unicode.IsLetter(rune(name[0])) {
		log.Printf("Service name validation failed: must start with a letter. Name: %s", name)
		return false
	}

	// Cannot end with a hyphen
	if name[len(name)-1] == '-' {
		log.Printf("Service name validation failed: cannot end with hyphen. Name: %s", name)
		return false
	}

	// Only lowercase letters, numbers, and hyphens allowed
	for i, r := range name {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '-' {
			log.Printf("Service name validation failed: invalid character at position %d: '%c'. Name: %s", i, r, name)
			return false
		}
	}

	log.Printf("Service name validation successful for: %s", name)
	return true
}
