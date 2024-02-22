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

package installscanlog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
)

const (
	buildPattern     = `Building [^\s]+: [^\s]+`
	directoryPattern = `(?:\/workspace\/source\/)(.+[/])` // relative path of artifact directories
)

// Go through the install log and scan to retrieve artifact output paths.
func ScanInstallLog(logPath string, outputPath string) error {
	// Read the file contents
	content, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("Error reading log file: %v\n", err)
	}

	// Convert file content to a string
	logs := string(content)

	artifactPaths, err := retrieveArtifactPaths(logs)
	if err != nil {
		return fmt.Errorf("Failed to retrieve artifact paths: %v\n", err)
	}

	// Create list string of all the paths
	var paths []string
	for path := range artifactPaths {
		paths = append(paths, path)
	}

	data, err := json.Marshal(paths) // to preserve quotes
	if err != nil {
		return fmt.Errorf("Failed to write marshal data: %v\n", err)
	}
	log.Printf("Write: %s\n", string(data))

	// Write string to output file
	err = os.WriteFile(outputPath, data, 0444)
	if err != nil {
		return fmt.Errorf("Failed to write payload to file: %v\n", err)
	}

	return nil
}

// Get all artifact paths from log
func retrieveArtifactPaths(logs string) (map[string]bool, error) {
	results := map[string]bool{} // using a map instead of an array to store unique values
	re := regexp.MustCompile(buildPattern)
	directoryRE := regexp.MustCompile(directoryPattern)

	// Get all lines with "Build __: /____/____"
	matches := re.FindAllStringSubmatch(logs, -1)

	for _, match := range matches {
		// Extract the directory paths
		directoryMatch := directoryRE.FindStringSubmatch(match[0])
		if len(directoryMatch) == 0 {
			return results, fmt.Errorf("Failed to retrieve a directory match")
		}
		log.Printf("Retrieved artifact path: %v\n", directoryMatch[1])
		results[directoryMatch[1]] = true
	}

	return results, nil
}
