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
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
)

// Uploader holds the client to interact with AR, and source to pull the image, target to push the image.
type Uploader struct {
	//TODO(@yongxuanzhang): this will be replaced with client from client library
	client       *http.Client
	gitlabClient *http.Client
	auth         authn.Authenticator
	source       string
	target       string
	gitLabURL    string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// New returns a new Uploader with given client, source, and target
func New(client, gitlabClient *http.Client, source, target, gitLabURL string, auth authn.Authenticator) (*Uploader, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if gitlabClient == nil {
		return nil, fmt.Errorf("gitlabClient is nil")
	}
	if source == "" {
		return nil, fmt.Errorf("source is empty")
	}
	if target == "" {
		return nil, fmt.Errorf("target is empty")
	}
	if gitLabURL == "" {
		return nil, fmt.Errorf("gitLabURL is empty")
	}
	if auth == nil {
		return nil, fmt.Errorf("auth is nil")
	}
	return &Uploader{
		client:       client,
		gitlabClient: gitlabClient,
		source:       source,
		target:       target,
		gitLabURL:    gitLabURL,
		auth:         auth,
	}, nil
}

// UserAgentTransport is a custom Transport that attaches a User-Agent header to requests
type UserAgentTransport struct {
	Transport http.RoundTripper
	UserAgent string
}

// RoundTrip adds the User-Agent header before making the actual request
func (t *UserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.UserAgent)
	return t.Transport.RoundTrip(req)
}
