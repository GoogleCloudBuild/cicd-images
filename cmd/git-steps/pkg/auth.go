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
	"fmt"
	"os"

	url "github.com/whilp/git-urls"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type SMClient interface {
	AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

type DCRepoClient interface {
	AccessReadWriteToken(string) (string, error)
	Get(string) (string, error)
}

// Authenticate using ssh public and private keys.
func AuthenticateWithSSHKeys(sf SMClient, sshPrivateKeySecretsResource, repoURL, urlPath string, sshServerPublicKeys []string) error {
	if err := getPrivateKey(sf, sshPrivateKeySecretsResource); err != nil {
		return err
	}

	if err := storePublicKeys(repoURL, urlPath, sshServerPublicKeys); err != nil {
		return err
	}

	return nil
}

// Get private key from secret manager and store in id_rsa file.
func getPrivateKey(sf SMClient, sshPrivateKeySecretsResource string) error {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: sshPrivateKeySecretsResource,
	}

	resp, err := sf.AccessSecretVersion(req)
	if err != nil {
		return fmt.Errorf("error accessing secret version: %v\n", err)
	}
	payload := resp.GetPayload().GetData()

	// Place private key in id_rsa file, and set file permissions
	err = os.MkdirAll(".ssh", 0o700) // rwx permissions for owner only
	if err != nil {
		return fmt.Errorf("error creating .ssh directory: %v\n", err)
	}

	idRsa, err := os.OpenFile(".ssh/id_rsa", os.O_CREATE|os.O_WRONLY, 0o700) // rwx permissions for owner only
	if err != nil {
		return fmt.Errorf("error opening id_rsa file: %v\n", err)
	}

	if _, err := idRsa.Write(payload); err != nil {
		return fmt.Errorf("error writing key to id_rsa file: %v\n", err)
	}
	if err := idRsa.Chmod(0o400); err != nil { // change permission to read only for owner
		return fmt.Errorf("error changing permission of id_rsa file: %v", err)
	}

	return nil
}

// Format and store the public keys in .ssh file.
func storePublicKeys(repoURL, urlPath string, sshServerPublicKeys []string) error {
	// Parse host and port from URL param
	u, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("error parsing host and port from URL: %v\n", err)
	}
	host := u.Host

	// store host in urlPath
	if err := StoreURL(repoURL, urlPath); err != nil {
		return fmt.Errorf("error storing url in path: %v", err)
	}

	// Add the ssh public keys to .ssh file
	file, err := os.OpenFile(".ssh/known_hosts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening known_hosts file: %v\n", err)
	}
	defer file.Close()

	for _, key := range sshServerPublicKeys {
		if _, err := file.WriteString(host + " " + key + "\n"); err != nil {
			return fmt.Errorf("error writing public keys to known_hosts file: %v\n", err)
		}
	}

	return nil
}

// Authenticate using Developer Connect.
func AuthenticateWithDeveloperConnect(gitRepositoryLink string, rf DCRepoClient, urlPath string) error {
	domain, err := getRemoteGitRepoURL(rf, gitRepositoryLink, urlPath)
	if err != nil {
		return err
	}

	if err := getReadWriteAccessToken(rf, gitRepositoryLink, domain); err != nil {
		return err
	}

	return nil
}

// Call the Developer Connect API service to get repository details and store in .gitconfig file.
func getRemoteGitRepoURL(rf DCRepoClient, gitRepositoryLink, urlPath string) (string, error) {
	// Get details of remote repo and extract url
	remoteUri, err := rf.Get(gitRepositoryLink)
	if err != nil {
		return "", fmt.Errorf("error getting repository details: %v", err)
	}

	u, err := url.Parse(remoteUri)
	if err != nil {
		return "", fmt.Errorf("error parsing host and port from URL: %v\n", err)
	}
	domain := u.Host

	// store domain in urlPath
	if err := StoreURL(remoteUri, urlPath); err != nil {
		return "", fmt.Errorf("error storing url in path: %v", err)
	}

	// Store the domain in the .gitconfig file
	configFile, err := os.OpenFile(".gitconfig", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("error opening .gitconfig file: %v\n", err)
	}
	defer configFile.Close()

	credentialConfig := fmt.Sprintf("[credential \"https://%s\"]\n  helper = store\n", domain)
	if _, err := configFile.WriteString(credentialConfig); err != nil {
		return "", fmt.Errorf("error writing to .gitconfig file: %v\n", err)
	}

	return domain, nil
}

// Call the Developer Connect API service to get the read-write access token and store in .git-credentials file.
func getReadWriteAccessToken(rf DCRepoClient, gitRepositoryLink, domain string) error {
	// Get the read-write access token for the repository
	repoToken, err := rf.AccessReadWriteToken(gitRepositoryLink)
	if err != nil {
		return fmt.Errorf("error getting read-write access token: %v", err)
	}

	// Store formatted token in .git-credentials file
	credFile, err := os.OpenFile(".git-credentials", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening .git-credentials file: %v\n", err)
	}
	defer credFile.Close()

	tokenCredentials := fmt.Sprintf("https://x-access-token:%s@%s\n", repoToken, domain)
	if _, err := credFile.WriteString(tokenCredentials); err != nil {
		return fmt.Errorf("error writing to .git-credentials file: %v\n", err)
	}

	return nil
}

func StoreURL(data, urlPath string) error {
	if urlPath == "" { // if urlPath is empty, then do not need to write URL
		return nil
	}
	// Write provenance as json file in path
	if err := os.WriteFile(urlPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("error writing results into %s: %v", urlPath, err)
	}
	return nil
}
