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

package auth

import (
	"encoding/json"
	"fmt"
	"os"
)

// SetupApplicationDefaultCredential builds the credential json and stores it at jwtJsonOutputPath.
// Tthe file content is built on full oidcJwt and workloadIdentityProvider and audience.
// If serviceAccount is provided, the service Account impersonation is applied during authentication.
func SetupApplicationDefaultCredential(credentialsOutputPath, oidcJwtEnvVar, serviceAccount, workloadIdentityProvider string) error {
	// write oidcJwt to jwtFilePath
	jwtFilePath := "/tmp/oidc-jwt.txt"
	jwtContent := os.Getenv(oidcJwtEnvVar)
	err := createJWTFile(jwtFilePath, jwtContent)
	if err != nil {
		return err
	}

	// compose the credential file using the JWT file
	err = createCredentialFile(credentialsOutputPath, jwtFilePath, serviceAccount, workloadIdentityProvider)
	if err != nil {
		return err
	}

	fmt.Printf("auth completed, file: %s\n", credentialsOutputPath)
	return nil
}

func createJWTFile(outputPath, oidcJwt string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	data := []byte(oidcJwt)
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func createCredentialFile(credentialJsonOutputPath, jwtFilePath, serviceAccount, workloadIdentityProvider string) error {
	config := ExternalAccountConfig{
		Type:             "external_account",
		Audience:         workloadIdentityProvider,
		SubjectTokenType: "urn:ietf:params:oauth:token-type:jwt",
		TokenURL:         "https://sts.googleapis.com/v1/token",
		CredentialSource: CredentialSource{
			File: jwtFilePath,
			Format: Format{
				Type: "text",
			},
		},
	}
	if serviceAccount != "" {
		fmt.Println("service account provided, authenticating use Workload Identity Federation with Service Account impersonation...")
		config.ServiceAccountImpersonationURL = fmt.Sprintf("https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken", serviceAccount)
	}

	// Convert the struct to JSON and write to file
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(credentialJsonOutputPath, jsonBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
