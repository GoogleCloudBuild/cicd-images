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
MIIB1QIBADANBgkqhkiG9w0BAQEFAASCAb8wggG7AgEAAl0DH3YqFv4mzt67RAAm
KqZSY32GtoUqkLXzSJOIew2ofiKx3ojdJvL69pXZLKNoKkKb8RQKyWdhAIkbTEFX
3k8mroXea5NMfB9NAH0AASQ6uoK5XYs7mMubQgu1dhcCAwEAAQJdAjrb+LAUaQe8
+cFTze0UeK48Ow5nxn4wvniriIA9v3vaMGJ0Hl6qkFO1qq76O+uvSehxPHnzBrfs
SXkQ8nScyeGpoTpn0DCnMnFRiY1hAMy6SqVdC4t7UP9u6oCBAi8B+POU6nCyUOnL
FlPVGFoBxSoxC7q7tJytq+xaPfGBN63AT3sdnXm06YAH1uE/1wIvAZVPf+1sDjIP
c4hFNPzIPh/x1M3qDN9eBr6tdPwymuPmpQ1lik/b9ZpMfXGns8ECLwDTVfcci+BF
tyP1i06jq4AUKg1u8E+BTxXs37YBOOOxDvpvCYMiln6eP6SITavvAi8A6n71d8rl
p6by4+uOjZXZA6hpw7zfN7hx1I4MugEZRjPiWI7f5/ZN8bjBdylcwQIvAQp1f9vQ
S+P5ktRlO7vEm10LtKotJ85Rp+le7PX56re+nntKVZFsliKW0yPmWJE=
-----END PRIVATE KEY-----`)
	sshPrivateKeySecretsResource := ""
	repoUrl := "git@github.com:gcb-catalog-testing-bot/private-repo.git"
	sshServerPublicKeys := `["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk="]`
	publicKeyResult := `github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
`

	sf := &fakeSecretVersionFetcher{
		resp: &secretmanagerpb.AccessSecretVersionResponse{
			Name: ``,
			Payload: &secretmanagerpb.SecretPayload{
				Data: privateKey,
			},
		},
	}

	t.Run("ssh auth", func(t *testing.T) {
		if err := AuthenticateWithSSHKeys(sf, sshPrivateKeySecretsResource, repoUrl, sshServerPublicKeys); err != nil {
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

	t.Run("repos api auth", func(t *testing.T) {
		if err := AuthenticateWithReposAPI(reposResource, rf); err != nil {
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
	})
}
