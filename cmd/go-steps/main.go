//  Copyright 2023 Google LLC
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
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/auth"
	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/netrc"
	"github.com/pkg/errors"
)

const (
	defaultProxy      = "https://proxy.golang.org,direct"
	arToolsCommand    = "run github.com/GoogleCloudPlatform/artifact-registry-go-tools/cmd/auth@latest"
	repositoryPattern = `([^/,]+)-go\.pkg\.dev`
)

func main() {
	// Authenticate with Artifact Registry.
	goproxy := os.Getenv("GOPROXY")
	if strings.Contains(goproxy, "go.pkg.dev") {
		err := authenticateArtifactRegistry(goproxy)
		if err != nil {
			log.Println(err)
		}
	}
}

// authenticateArtifactRegistry authenticates to Artifact Registry by running an auth tool binary
func authenticateArtifactRegistry(goproxy string) error {
	location, err := extractLocation(goproxy)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	// Add location to .netrcfile
	err = addLocations(location)
	if err != nil {
		return fmt.Errorf("error: Could not authenticate to Artifact Registry: %w", err)
	}

	// Refresh token
	err = refreshToken()
	if err != nil {
		return fmt.Errorf("error: Could not refresh Artifact Registry token: %w", err)
	}

	return nil
}

// extractLocation returns the repository location from a provided string.
func extractLocation(str string) (string, error) {
	pattern := regexp.MustCompile(repositoryPattern)
	match := pattern.FindStringSubmatch(str)
	if match != nil {
		return match[1], nil
	} else {
		return "", errors.New("Could not get location from proxy. please include a proxy with the format: https://{LOCATION}-go.pkg.dev/foo/baz")
	}
}

func refreshToken() error {
	ctx := context.Background()
	ctx, cf := context.WithTimeout(ctx, 30*time.Second)
	defer cf()

	p, config, err := netrc.Load()
	if err != nil {
		return err
	}

	token, err := auth.Token(ctx)
	if err != nil {
		return err
	}

	config = netrc.Refresh(config, token)
	if err := netrc.Save(config, p); err != nil {
		return err
	}
	return nil
}

func addLocations(locations string) error {
	if locations == "" {
		return fmt.Errorf("-locations is required")
	}
	ll := strings.Split(locations, ",")
	p, config, err := netrc.Load()
	if err != nil {
		return err
	}
	newCfg, err := netrc.AddConfigs(ll, config, "%s-go.pkg.dev", "")
	if err != nil {
		return err
	}
	if err := netrc.Save(newCfg, p); err != nil {
		return err
	}
	return nil
}
