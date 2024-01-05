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
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	defaultProxy      = "https://proxy.golang.org,direct"
	arToolsCommand    = "run github.com/GoogleCloudPlatform/artifact-registry-go-tools/cmd/auth@latest"
	repositoryPattern = `([^/,]+)-go\.pkg\.dev`
)

var (
	args string
)

func init() {
	flag.StringVar(&args, "args", "", "Arguments to execute with Go command")
}

func main() {
	flag.Parse()
	argsArray := strings.Fields(args)

	// Authenticate with Artifact Registry.
	goproxy := os.Getenv("GOPROXY")
	if strings.Contains(goproxy, "go.pkg.dev") {
		err := authenticateArtifactRegistry(goproxy)
		if err != nil {
			log.Println(err)
		}
	}

	// Execute input command.
	result, err := run(argsArray)
	if err != nil {
		log.Fatalf("Error: %s\n%s", err, result)
	} else {
		fmt.Print(result)
	}
}

// authenticateArtifactRegistry authenticates to Artifact Registry by running an auth tool binary
func authenticateArtifactRegistry(goproxy string) error {
	// Set default proxy to authenticate to AR.
	// GOPROXY needs to be set to default to retrieve first artifact-registry-go-tools
	// directly from source first and dependencies from https://proxy.golang.org.
	os.Setenv("GOPROXY", defaultProxy)

	// Set repository location in .netrc file
	location, err := extractLocation(goproxy)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	command := fmt.Sprintf("%v add-locations --locations=%s", arToolsCommand, location)
	result, err := run(strings.Fields(command))
	if err != nil {
		return fmt.Errorf("error: Could not authenticate to Artifact Registry: %s", result)
	} else {
		log.Println(result)
	}

	// Refresh token
	command = fmt.Sprintf("%v refresh", arToolsCommand)
	result, err = run(strings.Fields(command))
	if err != nil {
		return fmt.Errorf("error: Could not refresh Artifact Registry token: %s", result)
	} else {
		log.Println(result)
	}

	os.Setenv("GOPROXY", goproxy)
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

// run executes a go command followed by the given suffix and returns the output and error.
func run(args []string) (string, error) {
	c := exec.Command("go", args...)
	var output bytes.Buffer
	c.Stderr = &output
	c.Stdout = &output

	err := c.Run()
	return output.String(), err
}
