// Copyright 2024 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/GoogleCloudBuild/cicd-images/cmd/openjdk-steps/internal/setup"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const DEBUG_ARG = "debug"
const INPUTPOMFILE_ARG = "inputPomFile"
const EFFECTIVEPOMPATH_ARG = "effectivePomPath"
const SETTINGPATH_ARG = "settingPath"
const LOCALREPOSITORY_ARG = "localRepository"
const INSTALLFLAGS_ARG = "installFlags"
const SETTINGSECRETNAME_ARG = "settingSecretName"
const ARTIFACTDIRECTORY_ARG = "artifactDirectory"
const customizedSettings = "/mavenconfigs/settings.xml"

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate, and setup pom file and settings file.",
	Long:  "Retrieve authentication token from GKE metadata server, scan pom.xml from pomPath, update artifactDirectory, and configure settings.xml",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		authCmdArgs, err := parseAuthArgs(cmd.Flags())
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true

		fmt.Printf("Local repository is: %s\n", authCmdArgs.localRepository)

		c := metadata.NewClient(&http.Client{})

		err = setup.GenerateEffectivePom(authCmdArgs.inputPomFile, authCmdArgs.effectivePomPath)
		if err != nil {
			return fmt.Errorf("error generating effective pom file: %w", err)
		}

		token, err := setup.GetAuthenticationToken(c)
		if err != nil {
			return fmt.Errorf("getAuthenticationToken errors: %w", err)
		}
		fmt.Println("Token retrieval success!")

		sc, err := secretmanager.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("error getting secret manager client %w", err)
		}
		defer sc.Close()

		if authCmdArgs.settingSecretName != "" {
			fmt.Printf("Downloading custom settings from Secret Manager...")
			data, err := setup.RetrieveSettingSecret(authCmdArgs.settingSecretName, sc, ctx)
			if err != nil {
				return fmt.Errorf("retrieveSettingSecret error: %w", err)
			}
			err = os.WriteFile(customizedSettings, data, 0644)
			if err != nil {
				return fmt.Errorf("Failed to write payload to file: %w", err)
			}

			fmt.Printf("Secret payload saved to %s\n", customizedSettings)
		} else {
			repos, err := setup.RetrieveRepos(authCmdArgs.effectivePomPath, authCmdArgs.artifactDirectory)
			if err != nil {
				return fmt.Errorf("retrieveRepos errors: %w", err)
			}
			fmt.Printf("Main repos:\n%v\n", repos)
			err = setup.WriteSettingXML(authCmdArgs.debug, authCmdArgs.settingPath, authCmdArgs.localRepository, repos, token)
			if err != nil {
				return fmt.Errorf("writeSettingXML errors: %w", err)
			}
		}
		if authCmdArgs.installFlags != "" {
			unallowedFlag := setup.ValidateMavenInstallFlags(authCmdArgs.installFlags)
			if unallowedFlag != "" {
				return fmt.Errorf("Unallowed flags for maven-install Task %v", unallowedFlag)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.Flags().BoolP(DEBUG_ARG, "", false, "Whether to turn on debug mode")
	authCmd.Flags().StringP(INPUTPOMFILE_ARG, "", "./pom.xml", "Location of pom.xml")
	authCmd.Flags().StringP(EFFECTIVEPOMPATH_ARG, "", "./pom.xml", "Location of generated effective pom file.")
	authCmd.Flags().StringP(SETTINGPATH_ARG, "", customizedSettings, "Location of settings.xml")
	authCmd.Flags().StringP(LOCALREPOSITORY_ARG, "", "~/.m2/repository", "Location of local repository")
	authCmd.Flags().StringP(INSTALLFLAGS_ARG, "", "", "Flags being passed to maven-install Task")
	authCmd.Flags().StringP(SETTINGSECRETNAME_ARG, "", "", "Secret name in Secret Manager to store customized settings.xml. Format: projects/*/secrets/*")
	authCmd.Flags().StringP(ARTIFACTDIRECTORY_ARG, "", "./target", "Location where artifacts are installed")
}

type authArguments struct {
	debug             bool
	inputPomFile      string
	effectivePomPath  string
	settingPath       string
	localRepository   string
	installFlags      string
	settingSecretName string
	artifactDirectory string
}

func parseAuthArgs(f *pflag.FlagSet) (authArguments, error) {
	debug, err := f.GetBool(DEBUG_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get debug argument: %w", err)
	}

	inputPomFile, err := f.GetString(INPUTPOMFILE_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get input pom file: %w", err)
	}

	effectivePomPath, err := f.GetString(EFFECTIVEPOMPATH_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get pom path: %w", err)
	}

	settingPath, err := f.GetString(SETTINGPATH_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get setting path: %w", err)
	}

	localRepository, err := f.GetString(LOCALREPOSITORY_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get local repository: %w", err)
	}

	installFlags, err := f.GetString(INSTALLFLAGS_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get install flags: %w", err)
	}

	settingSecretName, err := f.GetString(SETTINGSECRETNAME_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get setting secret name: %w", err)
	}

	artifactDirectory, err := f.GetString(ARTIFACTDIRECTORY_ARG)
	if err != nil {
		return authArguments{}, fmt.Errorf("failed to get artifact directory: %w", err)
	}

	return authArguments{
		debug:             debug,
		inputPomFile:      inputPomFile,
		effectivePomPath:  effectivePomPath,
		settingPath:       settingPath,
		localRepository:   localRepository,
		installFlags:      installFlags,
		settingSecretName: settingSecretName,
		artifactDirectory: artifactDirectory,
	}, nil
}
