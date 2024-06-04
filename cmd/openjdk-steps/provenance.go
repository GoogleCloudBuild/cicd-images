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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	helper "github.com/GoogleCloudBuild/cicd-images/cmd/openjdk-steps/internal"
	"github.com/spf13/cobra"
)

var (
	repositoryUrl   string
	artifactName    string
	artifactId      string
	groupId         string
	version         string
	isBuildArtifact string
	resultsPath     string
)

type Provenance struct {
	Uri             string `json:"uri"`
	Digest          string `json:"digest"`
	IsBuildArtifact string `json:"isBuildArtifact"`
}

var generateProvenanceCmd = &cobra.Command{
	Use:   "generate-provenance",
	Short: "Get the maven artifact results.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		uri := fmt.Sprintf("%s/%s/%s/%s/%s", repositoryUrl, groupId, artifactId, version, artifactName)

		digest, err := helper.GetCheckSum(uri, ctx)
		if err != nil {
			return fmt.Errorf("error generating checksum: %w", err)
		}

		provenance := &Provenance{
			Uri:             strings.TrimSpace(uri),
			Digest:          strings.TrimSpace(digest),
			IsBuildArtifact: strings.TrimSpace(isBuildArtifact),
		}

		file, err := json.Marshal(provenance)
		if err != nil {
			return fmt.Errorf("error marshaling json %v: %v", provenance, err)
		}
		// Write provenance as json file in path
		if err := os.WriteFile(resultsPath, file, 0444); err != nil {
			return fmt.Errorf("error writing results into %s: %v", resultsPath, err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateProvenanceCmd)

	generateProvenanceCmd.Flags().StringVar(&repositoryUrl, "repositoryUrl", "", "URL of the Artifact Registry repository.")
	generateProvenanceCmd.Flags().StringVar(&artifactName, "artifactName", "", "The name of the artifact.")
	generateProvenanceCmd.Flags().StringVar(&artifactId, "artifactId", "", "The name of the package file created from the build step.")
	generateProvenanceCmd.Flags().StringVar(&groupId, "groupId", "", "ID to uniquely identify the project across all Maven projects.")
	generateProvenanceCmd.Flags().StringVar(&version, "version", "", "The version for the application.")
	generateProvenanceCmd.Flags().StringVar(&isBuildArtifact, "isBuildArtifact", "true", "If the results should be a build artifact.")
	generateProvenanceCmd.Flags().StringVar(&resultsPath, "resultsPath", "", "Path to write the results in.")
}
