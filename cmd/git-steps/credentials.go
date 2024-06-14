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
	"log/slog"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	auth "github.com/GoogleCloudBuild/cicd-images/cmd/git-steps/pkg"
	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"google.golang.org/api/developerconnect/v1"
	"google.golang.org/api/option"
)

var (
	url                          string
	urlPath                      string
	sshServerPublicKeys          []string
	sshPrivateKeySecretsResource string
	gitRepositoryLink            string
)

const gcpAccessTokenURL = "https://www.googleapis.com/auth/cloud-platform"

var generateCredentialsCmd = &cobra.Command{
	Use:   "generate-credentials",
	Short: "Fetch secrets and generate credential files for authentication.",
	Long: `Peform authentication by fetching secrets and generating related credential files.
	Two authentication methods:
	- SSH Keys: provided public keys, fetch private key from Secret Manager, and store the keys in their corresponding .ssh files
	- Developer Connect: provided url of developer connect git repository, fetch access token, and store the credentials in the corresponding .git files
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		logger.SetupLogger(verbose)
		slog.Info("Executing generate-credentials command")

		// Get default access token
		accessToken, err := helper.GetAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("error finding default credentials: %v", err)
		}
		slog.Info("Successfully retrieved default credentials")

		if gitRepositoryLink != "" {
			// Connect to Developer Connect API with access token
			developerconnectService, err := developerconnect.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})))
			if err != nil {
				return fmt.Errorf("error creating new developer connect service: %v", err)
			}

			arf := &repositoryFetcher{dc: developerconnectService}

			if err := auth.AuthenticateWithDeveloperConnect(gitRepositoryLink, arf, urlPath); err != nil {
				return err
			}
			slog.Info("Successfully authenticated with Developer Connect")
		} else if sshPrivateKeySecretsResource != "" {
			// Authenticate with Secret Manager
			client, err := secretmanager.NewClient(ctx, option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})))
			if err != nil {
				return fmt.Errorf("error authenticating and connecting to Secret Manager: %v\n", err)
			}
			defer client.Close()

			sf := &secretVersionFetcher{client: client, ctx: ctx}

			if err := auth.AuthenticateWithSSHKeys(sf, sshPrivateKeySecretsResource, url, urlPath, sshServerPublicKeys); err != nil {
				return err
			}
			slog.Info("Successfully authenticated with SSH Keys")
		} else {
			// for public repositories, just store url without auth
			if err := auth.StoreURL(url, urlPath); err != nil {
				return fmt.Errorf("error storing url without auth: %v", err)
			}
		}

		slog.Info("Successfully executed generate-credentials command")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCredentialsCmd)

	generateCredentialsCmd.Flags().StringVar(&url, "url", "", "The URL of the repository.")
	generateCredentialsCmd.Flags().StringVar(&urlPath, "urlPath", "", "The path to store the extracted URL of the repository.")
	generateCredentialsCmd.Flags().StringSliceVar(&sshServerPublicKeys, "sshServerPublicKeys", []string{}, "The public keys of the SSH server in an array.")
	generateCredentialsCmd.Flags().StringVar(&sshPrivateKeySecretsResource, "sshPrivateKeySecretsResource", "", "The secret version resource name of the SSH private key saved on Secret Manager.")
	generateCredentialsCmd.Flags().StringVar(&gitRepositoryLink, "gitRepositoryLink", "", "The resource name of the repository linked to Developer Connect.")
	generateCredentialsCmd.Flags().BoolVar(&verbose, "verbose", false, "Whether to print verbose output.")
}

type repositoryFetcher struct {
	dc *developerconnect.Service
}

var _ auth.DCRepoClient = (*repositoryFetcher)(nil) // Verify repositoryFetcher implements DCRepoClient

func (r *repositoryFetcher) Get(gitRepositoryLink string) (string, error) {
	repoDetails, err := developerconnect.NewProjectsLocationsConnectionsGitRepositoryLinksService(r.dc).Get(gitRepositoryLink).Do()
	if err != nil {
		return "", err
	}
	return repoDetails.CloneUri, nil
}

func (r *repositoryFetcher) AccessReadWriteToken(gitRepositoryLink string) (string, error) {
	tokenResponse, err := developerconnect.NewProjectsLocationsConnectionsGitRepositoryLinksService(r.dc).FetchReadWriteToken(gitRepositoryLink, &developerconnect.FetchReadWriteTokenRequest{}).Do()
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
