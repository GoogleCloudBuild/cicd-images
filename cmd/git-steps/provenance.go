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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var resultsPath string

type Provenance struct {
	Uri    string `json:"uri"`
	Digest string `json:"digest"`
	Ref    string `json:"ref"`
}

var generateProvenanceCmd = &cobra.Command{
	Use:   "generate-provenance",
	Short: "Get the git artifact results.",
	RunE: func(cmd *cobra.Command, args []string) error {
		uriCmd := exec.Command("git", strings.Fields("config --get remote.origin.url")...)
		uri, err := uriCmd.Output()
		if err != nil {
			return fmt.Errorf("error running 'git config --get remote.origin.url': %v", err)
		}

		digestCmd := exec.Command("git", strings.Fields("rev-parse HEAD")...)
		digest, err := digestCmd.Output()
		if err != nil {
			return fmt.Errorf("error running 'git rev-parse HEAD': %v", err)
		}

		refCmd := exec.Command("git", strings.Fields("symbolic-ref HEAD")...)
		ref, err := refCmd.Output()
		if err != nil {
			return fmt.Errorf("error running 'git symbolic-ref HEAD': %v", err)
		}

		provenance := &Provenance{
			Uri:    string(uri),
			Digest: string(digest),
			Ref:    string(ref),
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

	generateProvenanceCmd.Flags().StringVar(&resultsPath, "resultsPath", "", "Path to write the results in.")
}
