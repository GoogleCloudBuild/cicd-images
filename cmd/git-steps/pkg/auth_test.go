//  Copyright 2024 Google LLC
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

package auth

import (
	"bytes"
	"os"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type fakeSecretVersionFetcher struct {
	resp *secretmanagerpb.AccessSecretVersionResponse
}

var _ SMClient = (*fakeSecretVersionFetcher)(nil) // Verify fakeSecretVersionFetcher implements SMClient

func (f *fakeSecretVersionFetcher) AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return f.resp, nil
}

func TestSshAuth(t *testing.T) {
	privateKey := []byte(`-----BEGIN PRIVATE KEY-----
privatekey
-----END PRIVATE KEY-----`)
	sshPrivateKeySecretsResource := ""
	repoUrl := "git@github.com:test/test-repo.git"
	sshServerPublicKeys := `["ssh-ed25519 key1", "ssh-rsa key2"]`
	publicKeyResult := `github.com ssh-ed25519 key1
github.com ssh-rsa key2
`

	sf := &fakeSecretVersionFetcher{
		resp: &secretmanagerpb.AccessSecretVersionResponse{
			Name: ``,
			Payload: &secretmanagerpb.SecretPayload{
				Data: privateKey,
			},
		},
	}

	tmpdir := t.TempDir()
	// Create temporary file for testing
	tmpfile, err := os.CreateTemp(tmpdir, "*")
	if err != nil {
		t.Errorf("Error creating temporary file: %v", err)
	}
	defer tmpfile.Close()

	t.Run("ssh auth", func(t *testing.T) {
		if err := AuthenticateWithSSHKeys(sf, sshPrivateKeySecretsResource, repoUrl, sshServerPublicKeys, tmpfile.Name()); err != nil {
			t.Fatalf("Expected no errors, but got %v", err)
		}
		defer os.RemoveAll(".ssh")

		// check private key
		idRsaFile, err := os.ReadFile(".ssh/id_rsa")
		if err != nil {
			t.Fatalf("Error reading id_rsa file: %v", err)
		}

		if !bytes.Equal(privateKey, idRsaFile) {
			t.Errorf("Generated file does not match expected results. Expected:\n%v\n Got:\n%v\n", string(privateKey), string(idRsaFile))
		}

		// check public keys
		knownHostFile, err := os.ReadFile(".ssh/known_hosts")
		if err != nil {
			t.Fatalf("Error reading generated known_hosts file: %v", err)
		}

		if !bytes.Equal([]byte(publicKeyResult), knownHostFile) {
			t.Errorf("Generated file does not match expected results. Expected:\n%v\n Got:\n%v\n", publicKeyResult, string(knownHostFile))
		}

		url, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			t.Fatalf("Error reading .gitconfig file: %v", err)
		}
		if !bytes.Equal(url, []byte(repoUrl)) {
			t.Errorf("Generated url file does not match expected results. Expected:\n%v\n Got:\n%v\n", repoUrl, string(url))
		}
	})
}

type fakeRepositoryFetcher struct {
	fakeRepoUri string
	fakeToken   string
}

var _ CBRepoClient = (*fakeRepositoryFetcher)(nil) // Verify fakeRepositoryFetcher implements CBRepoClient

func (f *fakeRepositoryFetcher) Get(string) (string, error) {
	return f.fakeRepoUri, nil
}

func (f *fakeRepositoryFetcher) AccessReadWriteToken(string) (string, error) {
	return f.fakeToken, nil
}

func TestReposApiAuth(t *testing.T) {
	rf := &fakeRepositoryFetcher{
		fakeRepoUri: "https://fakeDomain/test/private-repo.git",
		fakeToken:   "fakeToken",
	}
	reposResource := ""
	configResult := "[credential \"https://fakeDomain\"]\n  helper = store\n"
	credResult := "https://x-access-token:fakeToken@fakeDomain\n"

	tmpdir := t.TempDir()
	// Create temporary file for testing
	tmpfile, err := os.CreateTemp(tmpdir, "*")
	if err != nil {
		t.Errorf("Error creating temporary file: %v", err)
	}
	defer tmpfile.Close()

	t.Run("repos api auth", func(t *testing.T) {
		if err := AuthenticateWithReposAPI(reposResource, rf, tmpfile.Name()); err != nil {
			t.Fatalf("Expected no errors, but got %v", err)
		}
		defer os.Remove(".gitconfig")
		defer os.Remove(".git-credentials")

		configFile, err := os.ReadFile(".gitconfig")
		if err != nil {
			t.Fatalf("Error reading .gitconfig file: %v", err)
		}

		if !bytes.Equal([]byte(configResult), configFile) {
			t.Errorf("Generated .gitconfig file does not match expected results. Expected:\n%v\n Got:\n%v\n", configResult, string(configFile))
		}

		credFile, err := os.ReadFile(".git-credentials")
		if err != nil {
			t.Fatalf("Error reading .git-credentials file: %v", err)
		}

		if !bytes.Equal([]byte(credResult), credFile) {
			t.Errorf("Generated .git-credentials file does not match expected results. Expected:\n%v\n Got:\n%v\n", credResult, string(credFile))
		}

		url, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			t.Fatalf("Error reading .gitconfig file: %v", err)
		}
		if !bytes.Equal(url, []byte(rf.fakeRepoUri)) {
			t.Errorf("Generated url file does not match expected results. Expected:\n%v\n Got:\n%v\n", rf.fakeRepoUri, string(url))
		}
	})
}
