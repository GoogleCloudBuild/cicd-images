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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

type Provenance struct {
	Uri    string
	Digest string
}

func GenerateProvenance(provenancePath string, packageFileName string, uri string) error {
	digest, err := computeDigest(packageFileName)
	if err != nil {
		return fmt.Errorf("error computing digest for %s: %v", packageFileName, err)
	}
	fmt.Printf("digest: %s uri: %s", digest, uri)

	provenance := &Provenance{
		Uri:    uri,
		Digest: digest,
	}

	file, err := json.Marshal(provenance)
	fmt.Printf("provenance: %s", string(file[:]))
	if err != nil {
		return fmt.Errorf("error marshaling json %v: %v", provenance, err)
	}

	// Write provenance as json file in path
	err = os.WriteFile(provenancePath, file, 0444)
	if err != nil {
		return fmt.Errorf("error writing file into %s: %v", provenancePath, err)
	}

	return nil
}

func computeDigest(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %v", filePath, err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("empty file: %s", filePath)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
