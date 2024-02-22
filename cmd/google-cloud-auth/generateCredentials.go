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
	auth "github.com/GoogleCloudBuild/cicd-images/cmd/google-cloud-auth/pkg"
	"github.com/spf13/cobra"
)

var (
	oidcJwtEnvVar            string
	serviceAccount           string
	workloadIdentityProvider string
	credentialsOutputPath    string
	credentialsJsonEnvVar    string
)

// authCmd represents the auth command
var generateCredentialsCmd = &cobra.Command{
	Use:   "generate-credentials",
	Short: "Setup Workload Identity Federation credential json",
	Long: `Setup Workload Identity Federation credential json file that can be used to authenticate to GCP services via gcloud
	or cloud client libraries. 

	Set 'GCLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE' env variable to the generated credential file when using gcloud.

	Set 'GOOGLE_APPLICATION_CREDENTIALS' env variable to the generated credential file when using Client Library.

	See more about how Application Default Credentials (ADC) works: https://cloud.google.com/docs/authentication/application-default-credentials

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.SetupApplicationDefaultCredential(credentialsJsonEnvVar, credentialsOutputPath, oidcJwtEnvVar, serviceAccount, workloadIdentityProvider); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCredentialsCmd)

	generateCredentialsCmd.PersistentFlags().StringVar(&oidcJwtEnvVar, "oidc-jwt-env-var", "", "The env var containing full OIDC JWT")
	generateCredentialsCmd.PersistentFlags().StringVar(&workloadIdentityProvider, "workload-identity-provider", "", "The value of the audience(aud) param in the generated credentials file")
	generateCredentialsCmd.PersistentFlags().StringVar(&serviceAccount, "service-account", "", "The Service Account to be impersonated")
	generateCredentialsCmd.PersistentFlags().StringVar(&credentialsOutputPath, "credentials-json-output-path", "/tmp/gcp-credentials.json", "The full file path of the output credentials json")
	generateCredentialsCmd.PersistentFlags().StringVar(&credentialsJsonEnvVar, "credentials-json-env-var", "", "The env var containing user-provided credentials")
}
