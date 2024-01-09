// Copyright 2023 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/spf13/cobra"
)

const (
	tokenPathSuffix = "instance/service-accounts/default/token"
)

type response struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func getAuthenticationToken(c *metadata.Client) (string, error) {
	// get access token to call gcp api's with, can pass scopes as an query param
	accessTokenJSON, err := c.Get(tokenPathSuffix)
	if err != nil {
		return "", fmt.Errorf("Error requesting token: %v", err)
	}
	var jsonRes response
	err = json.NewDecoder(strings.NewReader(accessTokenJSON)).Decode(&jsonRes)
	if err != nil {
		return "", fmt.Errorf("Error: retrieving auth token: %v", err)
	}
	return string(jsonRes.AccessToken), nil
}

var rootCmd = &cobra.Command{
	Use: "openjdk-steps",
}

func init() {
	rootCmd.AddCommand(deployScanCmd)
	rootCmd.AddCommand(authCmd)
}

func Execute(ctx context.Context) error {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
