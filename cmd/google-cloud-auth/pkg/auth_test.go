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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSetupApplicationDefaultCredential_WIF(t *testing.T) {
	defaltOutputPath := "/tmp/gcp-credentials.json"
	tcs := []struct {
		name                    string
		ouptutPath              string
		jwtEnvVar               string
		audience                string
		serviceAccount          string
		expectedCredentialsPath string
		expectedCredentials     ExternalAccountConfig
	}{
		{
			name:                    "direct WIF",
			ouptutPath:              "/tmp/custom.json",
			jwtEnvVar:               "JWT_ENV_VAR",
			audience:                "test-audience",
			expectedCredentialsPath: "/tmp/custom.json",
			expectedCredentials: ExternalAccountConfig{
				Type:             "external_account",
				Audience:         "test-audience",
				SubjectTokenType: "urn:ietf:params:oauth:token-type:jwt",
				TokenURL:         "https://sts.googleapis.com/v1/token",
				CredentialSource: CredentialSource{
					File: "/tmp/oidc-jwt.txt",
					Format: Format{
						Type: "text",
					},
				},
			},
		}, {
			name:                    "WIF with SA impersonation",
			ouptutPath:              defaltOutputPath,
			jwtEnvVar:               "JWT_ENV_VAR",
			audience:                "test-audience",
			serviceAccount:          "test-sa",
			expectedCredentialsPath: defaltOutputPath,
			expectedCredentials: ExternalAccountConfig{
				Type:                           "external_account",
				Audience:                       "test-audience",
				SubjectTokenType:               "urn:ietf:params:oauth:token-type:jwt",
				TokenURL:                       "https://sts.googleapis.com/v1/token",
				ServiceAccountImpersonationURL: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/test-sa:generateAccessToken",
				CredentialSource: CredentialSource{
					File: "/tmp/oidc-jwt.txt",
					Format: Format{
						Type: "text",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		if err := os.Setenv("JWT_ENV_VAR", "jwt-content"); err != nil {
			t.Fatalf("unexpected err setting JWT_ENV_VAR: %v", err.Error())
		}

		if err := SetupApplicationDefaultCredential("", tc.ouptutPath, tc.jwtEnvVar, tc.serviceAccount, tc.audience); err != nil {
			t.Fatalf("unexpected err calling SetupApplicationDefaultCredential: %v", err.Error())
		}

		cont, err := os.ReadFile(tc.expectedCredentialsPath)
		if err != nil {
			t.Fatalf("Error reading file: %v", err)
		}
		var credentials ExternalAccountConfig
		if err := json.Unmarshal(cont, &credentials); err != nil {
			t.Fatalf("Error unmarshalling JSON: %v", err)
		}
		if d := cmp.Diff(tc.expectedCredentials, credentials); d != "" {
			t.Errorf("credentials does not match: %s", d)
		}

		// clean up
		err = os.Remove(tc.expectedCredentialsPath)
		if err != nil {
			t.Fatalf("Error deleting file: %v", err)
		}
	}
}

func TestSetupApplicationDefaultCredential_UserProvidedJson(t *testing.T) {
	content := "my credentials"
	outputPath := "/tmp/gcp-credentials.json"

	if err := os.Setenv("CREDENTIALS_ENV_VAR", content); err != nil {
		t.Fatalf("unexpected err setting JWT_ENV_VAR: %v", err.Error())
	}

	if err := SetupApplicationDefaultCredential("CREDENTIALS_ENV_VAR", outputPath, "", "", ""); err != nil {
		t.Fatalf("unexpected err calling SetupApplicationDefaultCredential: %v", err.Error())
	}

	cont, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	credentials := string(cont)
	if d := cmp.Diff(content, credentials); d != "" {
		t.Errorf("credentials does not match: %s", d)
	}

	// clean up
	err = os.Remove(outputPath)
	if err != nil {
		t.Fatalf("Error deleting file: %v", err)
	}

}
