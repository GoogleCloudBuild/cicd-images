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

package main

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	auth "github.com/GoogleCloudBuild/cicd-images/cmd/git-steps/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbuild/v2"
	"google.golang.org/api/option"
)

var (
	url                          string
	sshServerPublicKeys          string
	sshPrivateKeySecretsResource string
	reposResource                string
)
const gcpAccessTokenURL = "https://www.googleapis.com/auth/cloud-platform"

var generateCredentialsCmd = &cobra.Command{
	Use:   "generate-credentials",
	Short: "Fetch secrets and generate credential files for authentication.", //TODO
	Long: `Peform authentication by fetching secrets and generating related credential files.
	Two authentication methods:
	- SSH Keys: provided public keys, fetch private key from Secret Manager, and store the keys in their corresponding .ssh files
	- Repos API: provided url of cloud build repository resource, fetch access token, and store the credentials in the corresponding .git files
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get default access token
		credentials, err := getAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("error finding default credentials: %v", err)
		}

		if reposResource != "" {
			// Connect to Cloud Build API with access token
			cloudbuildService, err := cloudbuild.NewService(ctx, option.WithTokenSource(credentials.TokenSource))
			if err != nil {
				return fmt.Errorf("error creating new cloudbuild service: %v", err)
			}

			arf := &repositoryFetcher{cb: cloudbuildService}

			if err := auth.AuthenticateWithReposAPI(reposResource, arf); err != nil {
				return err
			}
		} else if sshPrivateKeySecretsResource != "" {
			// Authenticate with Secret Manager
			client, err := secretmanager.NewClient(ctx, option.WithTokenSource(credentials.TokenSource))
			if err != nil {
				return fmt.Errorf("error authenticating and connecting to Secret Manager: %v\n", err)
			}
			defer client.Close()

			sf := &secretVersionFetcher{client: client, ctx: ctx}

			if err := auth.AuthenticateWithSSHKeys(sf, sshPrivateKeySecretsResource, url, sshServerPublicKeys); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("parameters for either SSH keys or Cloud Build Repositories must be provided for authentication")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCredentialsCmd)

	generateCredentialsCmd.Flags().StringVar(&url, "url", "", "The URL of the repository.")
	generateCredentialsCmd.Flags().StringVar(&sshServerPublicKeys, "sshServerPublicKeys", `[""]`, "The public keys of the SSH server in an array.")
	generateCredentialsCmd.Flags().StringVar(&sshPrivateKeySecretsResource, "sshPrivateKeySecretsResource", "", "The URL of the SSH private key saved on Secret Manager.")
	generateCredentialsCmd.Flags().StringVar(&reposResource, "reposResource", "", "The URL of the Cloud Build Repository resource.")
}

type repositoryFetcher struct {
	cb *cloudbuild.Service
}

var _ auth.CBRepoClient = (*repositoryFetcher)(nil) // Verify repositoryFetcher implements CBRepoClient

func (r *repositoryFetcher) Get(reposResource string) (string, error) {
	repoDetails, err := cloudbuild.NewProjectsLocationsConnectionsRepositoriesService(r.cb).Get(reposResource).Do()
	if err != nil {
		return "", err
	}
	return repoDetails.RemoteUri, nil
}

func (r *repositoryFetcher) AccessReadWriteToken(reposResource string) (string, error) {
	tokenResponse, err := cloudbuild.NewProjectsLocationsConnectionsRepositoriesService(r.cb).AccessReadWriteToken(reposResource, &cloudbuild.FetchReadWriteTokenRequest{}).Do()
	if err != nil {
		return "", err
	}
	return tokenResponse.Token, nil
}

type secretVersionFetcher struct {
	client *secretmanager.Client
	ctx    context.Context
}

var _ auth.SMClient = (*secretVersionFetcher)(nil) // Verify secretVersionFetcher implements SMClient

func (r *secretVersionFetcher) AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	resp, err := r.client.AccessSecretVersion(r.ctx, req)
	if err != nil {
		return &secretmanagerpb.AccessSecretVersionResponse{}, fmt.Errorf("error accessing secret version: %v\n", err)
	}
	return resp, nil
}

// Get access token for google cloud service.
func getAccessToken(ctx context.Context) (*google.Credentials, error) {
	scopes := []string{
		gcpAccessTokenURL,
	}
	credentials, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return &google.Credentials{}, err
	}

	return credentials, nil
}
