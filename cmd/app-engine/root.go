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
	"github.com/spf13/cobra"
)

var userAgent string

func newAppEngineCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "app-engine deploy",
		Short: "a CLI tool to use App Engine",
	}

	rootCmd.PersistentFlags().StringVarP(&opts.ProjectID, "project-id", "p", "", "Google Cloud Project ID")
	rootCmd.PersistentFlags().StringVar(&userAgent, "google-apis-user-agent", "", "The user-agent to be applied when calling Google APIs")

	_ = rootCmd.MarkPersistentFlagRequired("project-id")

	return rootCmd
}
