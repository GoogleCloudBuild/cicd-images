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
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
)

type Provenance struct {
	URI             string `json:"uri"`
	Digest          string `json:"digest"`
	IsBuildArtifact string `json:"isBuildArtifact"`
}

type PackageJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func GenerateProvenance(provenancePath, repository, isBuildArtifact string) error {
	pkg, err := parsePackageJSON()
	if err != nil {
		return err
	}

	fileName := strings.ReplaceAll(pkg.Name, "@", "")
	fileName = strings.ReplaceAll(fileName, "/", "-")

	tarballName := fmt.Sprintf("%s-%s.tgz", fileName, pkg.Version)
	uri := fmt.Sprintf("%s/%s/%s/%s", repository, pkg.Name, pkg.Version, tarballName)

	digest, err := helper.ComputeDigest(tarballName)
	if err != nil {
		return fmt.Errorf("error computing digest for %s: %w", tarballName, err)
	}

	provenance := &Provenance{
		URI:             strings.TrimSpace(uri),
		Digest:          strings.TrimSpace(digest),
		IsBuildArtifact: strings.TrimSpace(isBuildArtifact),
	}

	file, err := json.Marshal(provenance)
	if err != nil {
		return fmt.Errorf("error marshaling json %v: %w", provenance, err)
	}

	// Write provenance as json file in path
	err = os.WriteFile(provenancePath, file, 0444)
	if err != nil {
		return fmt.Errorf("error writing file into %s: %w", provenancePath, err)
	}

	return nil
}

func parsePackageJSON() (PackageJSON, error) {
	// Parse package.json file to extract name & version
	data, err := os.ReadFile("./package.json")
	if err != nil {
		return PackageJSON{}, fmt.Errorf("error reading package.json file: %w", err)
	}

	var pkg PackageJSON
	err = json.Unmarshal(data, &pkg)
	if err != nil {
		return PackageJSON{}, fmt.Errorf("error unmarshalling package.json file: %w", err)
	}

	return pkg, nil
}
