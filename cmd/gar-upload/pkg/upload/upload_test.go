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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-containerregistry/pkg/authn"
)

func TestImgPath_Success(t *testing.T) {
	imgPathTests := []struct {
		name     string
		target   string
		expected string
	}{
		{
			name:     "Standard Input",
			target:   "us-central1-docker.pkg.dev/project/repo/image:imagetag",
			expected: "projects/project/locations/us-central1/repositories/repo/packages/image"},
		{
			name:     "No image tag",
			target:   "us-central1-docker.pkg.dev/project/repo/image",
			expected: "projects/project/locations/us-central1/repositories/repo/packages/image"}, // Should probably return an error, but assuming current logic
		{
			name:     "image name contains `/`",
			target:   "us-central1-docker.pkg.dev/project/repo/image/test",
			expected: "projects/project/locations/us-central1/repositories/repo/packages/image%2Ftest"},
	}
	for _, tt := range imgPathTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := imgPath(tt.target)
			if err != nil {
				t.Error(err)
			}
			if result != tt.expected {
				t.Errorf("ImgPath(%q) failed. Expected: %q, Got: %q", tt.target, tt.expected, result)
			}
		})
	}
}

func TestPackageMetaData_Update(t *testing.T) {
	path := "imgPath"
	endpoint := "http://artifactregistry.googleapis.com/v1beta2"
	gitlabUrl := "https://gitlab.com/user/app/container_registry/111]"
	annotations := map[string]string{}

	req, err := updateRequest(gitlabUrl, path, fmt.Sprintf("%s/%s?update_mask=annotations", endpoint, path), annotations)
	if err != nil {
		t.Error(err)
	}

	metadata := packageMetadata{
		Name: path,
		Annotations: map[string]string{
			annotationKey: gitlabUrl,
		},
	}

	resp, err := json.Marshal(metadata)
	if err != nil {
		t.Error(err)
	}

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer ts.Close()

	// Create a custom HTTP client that points to the test server
	mockClient := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(ts.URL)
			},
		},
	}
	auth := authn.FromConfig(authn.AuthConfig{})
	uploader, err := New(mockClient, &http.Client{}, "source", "target", "url", auth)
	if err != nil {
		t.Fatal(err)
	}
	got, err := uploader.packageMetaData(req)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(got, &metadata); diff != "" {
		t.Error(diff)
	}
}

func TestGetGitLabLink(t *testing.T) {
	repositories := []repository{{
		ID:       1,
		Name:     "",
		Path:     "group/project/",
		Location: "registry.gitlab.com/my-group/test-repo",
	}}

	resp, err := json.Marshal(repositories)
	if err != nil {
		t.Error(err)
	}

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer ts.Close()

	// Create a custom HTTP client that points to the test server
	mockClient := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(ts.URL)
			},
		},
	}
	auth := authn.FromConfig(authn.AuthConfig{})
	uploader, err := New(&http.Client{}, mockClient, "registry.gitlab.com/my-group/test-repo:latest", "target", ts.URL, auth)
	if err != nil {
		t.Fatal(err)
	}
	link, err := uploader.getGitLabLink()
	if err != nil {
		t.Fatal(err)
	}

	expectedLink := fmt.Sprintf("https://gitlab.com/%scontainer_registry/%d", repositories[0].Path, repositories[0].ID)
	if link != expectedLink {
		t.Errorf("Expected link %s, got %s", expectedLink, link)
	}
}
