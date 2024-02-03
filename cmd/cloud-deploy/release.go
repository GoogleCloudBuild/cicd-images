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
	"fmt"

	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a Cloud Deploy Release",
	Long:  ``, // TODO: add a longer description
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: add release create business logic
		fmt.Println("release called")
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	// TODO: add cloud deploy release flag
}
