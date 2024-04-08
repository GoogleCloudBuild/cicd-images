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
	"encoding/json"
	"fmt"
	"os"

	url "github.com/whilp/git-urls"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type SMClient interface {
	AccessSecretVersion(req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

type CBRepoClient interface {
	AccessReadWriteToken(string) (string, error)
	Get(string) (string, error)
}

// Authenticate using ssh public and private keys.
func AuthenticateWithSSHKeys(sf SMClient, sshPrivateKeySecretsResource string, repoUrl string, sshServerPublicKeys string, urlPath string) error {
	if err := getPrivateKey(sf, sshPrivateKeySecretsResource); err != nil {
		return err
	}

	if err := storePublicKeys(repoUrl, sshServerPublicKeys, urlPath); err != nil {
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
	err = os.MkdirAll(".ssh", 0700) // rwx permissions for owner only
	if err != nil {
		return fmt.Errorf("error creating .ssh directory: %v\n", err)
	}

	id_rsa, err := os.OpenFile(".ssh/id_rsa", os.O_CREATE|os.O_WRONLY, 0700) // rwx permissions for owner only
	if err != nil {
		return fmt.Errorf("error opening id_rsa file: %v\n", err)
	}

	if _, err := id_rsa.Write(payload); err != nil {
		return fmt.Errorf("error writing key to id_rsa file: %v\n", err)
	}
	if err := id_rsa.Chmod(0400); err != nil { // change permission to read only for owner
		return fmt.Errorf("error changing permission of id_rsa file: %v", err)
	}

	return nil
}

// Format and store the public keys in .ssh file.
func storePublicKeys(repoUrl string, sshServerPublicKeys string, urlPath string) error {
	// Parse host and port from URL param
	u, err := url.Parse(repoUrl)
	if err != nil {
		return fmt.Errorf("error parsing host and port from URL: %v\n", err)
	}
	host := u.Host

	// store host in urlPath
	if err := StoreURL(repoUrl, urlPath); err != nil {
		return fmt.Errorf("error storing url in path: %v", err)
	}

	// Convert the sshServerPublicKeys string to an array.
	var publicKeys []string
	if err := json.Unmarshal([]byte(sshServerPublicKeys), &publicKeys); err != nil {
		return fmt.Errorf("error converting public keys string to array: %v", err)
	}

	// Add the ssh public keys to .ssh file
	file, err := os.OpenFile(".ssh/known_hosts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening known_hosts file: %v\n", err)
	}
	defer file.Close()

	for _, key := range publicKeys {
		if _, err := file.WriteString(host + " " + key + "\n"); err != nil {
			return fmt.Errorf("error writing public keys to known_hosts file: %v\n", err)
		}
	}

	return nil
}

// Authenticate using Repos API.
func AuthenticateWithReposAPI(reposResource string, rf CBRepoClient, urlPath string) error {
	domain, err := getRemoteGitRepoUrl(rf, reposResource, urlPath)
	if err != nil {
		return err
	}

	if err := getReadWriteAccessToken(rf, reposResource, domain); err != nil {
		return err
	}

	return nil
}

// Call the Cloud Build API service to get repository details and store in .gitconfig file.
func getRemoteGitRepoUrl(rf CBRepoClient, reposResource string, urlPath string) (string, error) {
	// Get details of remote repo and extract url
	remoteUri, err := rf.Get(reposResource)
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
	configFile, err := os.OpenFile(".gitconfig", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// Call the Cloud Build API service to get the read-write access token and store in .git-credentials file.
func getReadWriteAccessToken(rf CBRepoClient, reposResource string, domain string) error {
	// Get the read-write access token for the repository
	reposAPIToken, err := rf.AccessReadWriteToken(reposResource)
	if err != nil {
		return fmt.Errorf("error getting read-write access token: %v", err)
	}

	// Store formatted token in .git-credentials file
	credFile, err := os.OpenFile(".git-credentials", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening .git-credentials file: %v\n", err)
	}
	defer credFile.Close()

	tokenCredentials := fmt.Sprintf("https://x-access-token:%s@%s\n", reposAPIToken, domain)
	if _, err := credFile.WriteString(tokenCredentials); err != nil {
		return fmt.Errorf("error writing to .git-credentials file: %v\n", err)
	}

	return nil
}

func StoreURL(data string, urlPath string) error {
	if urlPath == "" { // if urlPath is empty, then do not need to write URL
		return nil
	}
	// Write provenance as json file in path
	if err := os.WriteFile(urlPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("error writing results into %s: %v", urlPath, err)
	}
	return nil
}