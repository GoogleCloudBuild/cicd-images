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

package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
)

const (
	arEndpoint     = "https://artifactregistry.googleapis.com/v1beta2"
	gitlabEndpoint = "https://gitlab.com/api/v4/projects/%s/registry/repositories?job_token=%s"
	annotationKey  = "console.cloud.google.com/external_link"
	dockerSuffix   = "-docker.pkg.dev"
)

// CopyImage pulls the image from source and push to target
func (u *Uploader) CopyImage(ctx context.Context) error {
	img, err := crane.Pull(u.source, crane.WithAuth(u.auth))
	if err != nil {
		return fmt.Errorf("pulling image %s: %v", u.source, err)
	}
	if err := crane.Push(img, u.target, crane.WithTransport(u.client.Transport)); err != nil {
		return fmt.Errorf("pushing image %s: %v", u.target, err)
	}
	return nil
}

// UpdateAnnotation gets the GitLab link via rest api, fetch the annotation from the AR package,
// and update the annotation's console.cloud.google.com/external_link key with value of the GitLab link.
func (u *Uploader) UpdateAnnotation() error {
	// this is a call to GitLab to fetch the url
	// TODO(@yongxuanzhang): investigate why cannot use Uploader.Client to make the request to GitLab
	link, err := u.getGitLabLink()
	if err != nil {
		return fmt.Errorf("getting GitLab link for image %s: %v", u.source, err)
	}

	// fetch annotations
	path, err := imgPath(u.target)
	if err != nil {
		return fmt.Errorf("getting path for image %s: %v", u.target, err)
	}

	getReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", arEndpoint, path), nil)
	if err != nil {
		return fmt.Errorf("creating get request: %v", err)
	}
	metadata, err := u.packageMetaData(getReq)
	if err != nil {
		return fmt.Errorf("getting metadata for image %s: %v", u.target, err)
	}

	// only update annotations's AnnotationKey
	updateReq, err := updateRequest(link, path, fmt.Sprintf("%s/%s?update_mask=annotations", arEndpoint, path), metadata.Annotations)
	if err != nil {
		return fmt.Errorf("creating update request: %v", err)
	}
	_, err = u.packageMetaData(updateReq)
	if err != nil {
		return fmt.Errorf("updating metadata for image %s: %v", u.target, err)
	}
	log.Println("Annotations updated with", link)
	return nil
}

func imgPath(target string) (string, error) {
	split := strings.SplitN(target, "/", 4)
	if len(split) != 4 {
		return "", fmt.Errorf("got %d parts but expect %s is split to 4 parts", len(split), target)
	}
	hostname, project, repo, imageWithTag := split[0], split[1], split[2], split[3]
	if !strings.HasSuffix(hostname, dockerSuffix) {
		return "", fmt.Errorf("expect %s to have suffix %s", hostname, dockerSuffix)
	}
	location := strings.TrimSuffix(hostname, dockerSuffix)
	split = strings.Split(imageWithTag, ":")
	image := split[0]
	// image name may contain "/", so need to escape to avoid errors
	escapedImage := url.QueryEscape(image)
	path := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s", project, location, repo, escapedImage)
	return path, nil
}

// packageMetaData take the request to call Artifact Registry's API and return PackageMetadata
func (u *Uploader) packageMetaData(req *http.Request) (*packageMetadata, error) {
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	// Process the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("calling api, unexpected status code got=%d want=%d, response body %s", resp.StatusCode, http.StatusOK, string(body))
	}
	metadata := packageMetadata{}
	err = json.Unmarshal(body, &metadata)
	return &metadata, err
}

type packageMetadata struct {
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
}

func updateRequest(link, path, url string, annotations map[string]string) (*http.Request, error) {
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[annotationKey] = link
	payload := packageMetadata{
		Name:        path,
		Annotations: annotations,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payloadBytes))
}

type repository struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Location  string `json:"location"`
	CreatedAt string `json:"created_at"`
}

// / getGitLabLink retrieves the GitLab Registry Container link for a given image source.
func (u *Uploader) getGitLabLink() (string, error) {
	// Create the HTTP request
	req, err := http.NewRequest("GET", u.gitLabURL, nil)
	if err != nil {
		return "", err
	}

	// Execute the request
	resp, err := u.gitlabClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var repositories []repository // Define a struct to match the JSON structure
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return "", err
	}
	link := ""

	// remove the tag of the GitLab image if it exists, e.g. registry.gitlab.com/group/myapp/img:latest
	// is splitted to registry.gitlab.com/group/myapp/img and latest
	splitted := strings.Split(u.source, ":")

	for _, repo := range repositories {
		if repo.Location == splitted[0] {
			proj := strings.TrimSuffix(repo.Path, repo.Name)
			proj = strings.TrimSuffix(proj, "/")
			link = fmt.Sprintf("https://gitlab.com/%s/container_registry/%d", proj, repo.ID)
			return link, nil
		}
	}
	return "", fmt.Errorf("cannot find image link with image %s", u.source)
}
