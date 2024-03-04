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
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/cmd/gar-upload/pkg/upload"
	"github.com/google/go-containerregistry/pkg/authn"
	"golang.org/x/oauth2/google"

	"github.com/spf13/cobra"
)

var (
	source  string
	target  string
	version string
)

const (
	userAgent      = "google-gitlab-components:artifact-registry-upload/"
	gitlabEndpoint = "https://gitlab.com/api/v4/projects/%s/registry/repositories?job_token=%s"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gar-upload",
	Short: "gar-upload is used to upload docker image to Google Artifact Registry",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		logger := log.Default()
		// init client
		client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return err
		}
		client.Transport = &upload.UserAgentTransport{
			Transport: client.Transport,
			UserAgent: userAgent + version,
		}

		gitlabClient := &http.Client{}
		//TODO(@yongxuanzhang): add user agent for gitlabClient

		// init vars
		projectID := os.Getenv("CI_PROJECT_ID")
		jobToken := os.Getenv("CI_JOB_TOKEN")

		source = strings.TrimSpace(source)
		target = strings.TrimSpace(target)
		projectID = strings.TrimSpace(projectID)
		jobToken = strings.TrimSpace(jobToken)

		auth := authn.FromConfig(authn.AuthConfig{
			Username: os.Getenv("CI_REGISTRY_USER"),
			Password: os.Getenv("CI_REGISTRY_PASSWORD"),
		})
		uploader, err := upload.New(client, gitlabClient, source, target, fmt.Sprintf(gitlabEndpoint, projectID, jobToken), auth)
		if err != nil {
			return fmt.Errorf("creating uploader: %w", err)
		}
		// copy image
		if err := uploader.CopyImage(ctx); err != nil {
			return err
		}
		log.Println("Image pushed to AR")

		// update package annotation
		if err := uploader.UpdateAnnotation(); err != nil {
			logger.Println("Failed to update annotation, ignoring:", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("Failure: ", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&source, "source", "", "", "Source Image Path, e.g. registry.gitlab.com/group/project/image:tag")
	rootCmd.PersistentFlags().StringVarP(&target, "target", "", "", "Target Image Path, e.g. us-central1-docker.pkg.dev/projectID/repo/image:tag")
	rootCmd.PersistentFlags().StringVarP(&version, "version", "", "", "Version of the binary caller, e.g. 0.1.0")
}
