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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/config"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/option"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
)

var OKWithBody = func(in any, t *testing.T) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		marshal, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("failed to marsharl %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(marshal)
		if err != nil {
			t.Fatalf("failed to write mocked response body %v", err)
		}
	}
}

func TestPackAndUploadSourceCodeToCloudBucket(t *testing.T) {
	type TestCase struct {
		description  string
		opts         config.DeployOptions
		uploadMock   func(w http.ResponseWriter, r *http.Request)
		mockFileUtil MockFileUtil
		wantErr      bool
	}
	tests := []TestCase{
		{
			description: "empty string, success",
			opts: config.DeployOptions{
				Source: "",
			},
			uploadMock: OKWithBody(&functionspb.GenerateUploadUrlResponse{
				UploadUrl: "http://domain.com/upload",
			}, t),
			mockFileUtil: MockFileUtil{ZipFileName: "tmp/function.zip"},
			wantErr:      false,
		},
		{
			description: "valid path, success",
			opts: config.DeployOptions{
				Source: "/tmp/functions",
			},
			uploadMock: OKWithBody(&functionspb.GenerateUploadUrlResponse{
				UploadUrl: "http://domain.com/upload",
			}, t),
			mockFileUtil: MockFileUtil{ZipFileName: "tmp/function.zip"},
			wantErr:      false,
		},
		{
			description: "path to a file, failure",
			opts: config.DeployOptions{
				Source: "/tmp/functions/hello.txt",
			},
			mockFileUtil: MockFileUtil{ZipCreationError: fmt.Errorf(pathIsNotDirectoryErrorMessage)},
			wantErr:      true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost && strings.Contains(r.RequestURI, "functions:generateUploadUrl") {
					tc.uploadMock(w, r)
				}
			}))

			ctx := context.Background()
			client, err := functions.NewFunctionRESTClient(ctx, option.WithEndpoint(ts.URL), option.WithoutAuthentication())
			if err != nil {
				t.Fatalf("failed to create client %v", err)
			}

			_, err = packAndUploadSourceCodeToCloudBucket(ctx, client, tc.mockFileUtil, tc.opts)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr %t, err %v", tc.wantErr, err)
			}
		})
	}
}

func TestBuildSourceCSR(t *testing.T) {
	type TestCase struct {
		description string
		source      string
		repoSource  *functionspb.RepoSource
		wantErr     bool
	}
	tests := []TestCase{
		{
			description: "minimal url provided, success",
			source:      "https://source.developers.google.com/projects/test-project/repos/test-repository",
			repoSource: &functionspb.RepoSource{
				ProjectId: "test-project",
				RepoName:  "test-repository",
				Revision: &functionspb.RepoSource_BranchName{
					BranchName: "master",
				},
			},
			wantErr: false,
		},
		{
			description: "url with branch name provided, success",
			source:      "https://source.developers.google.com/projects/test-project/repos/test-repository/moveable-aliases/branch-name1",
			repoSource: &functionspb.RepoSource{
				ProjectId: "test-project",
				RepoName:  "test-repository",
				Revision: &functionspb.RepoSource_BranchName{
					BranchName: "branch-name1",
				},
			},
			wantErr: false,
		},
		{
			description: "url with tag name provided, success",
			source:      "https://source.developers.google.com/projects/test-project/repos/test-repository/fixed-aliases/tag-name1",
			repoSource: &functionspb.RepoSource{
				ProjectId: "test-project",
				RepoName:  "test-repository",
				Revision: &functionspb.RepoSource_TagName{
					TagName: "tag-name1",
				},
			},
			wantErr: false,
		},
		{
			description: "url with commit sha provided, success",
			source:      "https://source.developers.google.com/projects/test-project/repos/test-repository/revisions/commit-sha1",
			repoSource: &functionspb.RepoSource{
				ProjectId: "test-project",
				RepoName:  "test-repository",
				Revision: &functionspb.RepoSource_CommitSha{
					CommitSha: "commit-sha1",
				},
			},
			wantErr: false,
		},
		{
			description: "url with branch name and path to dir provided, success",
			source:      "https://source.developers.google.com/projects/test-project/repos/test-repository/moveable-aliases/branch-name1/paths/tmp",
			repoSource: &functionspb.RepoSource{
				ProjectId: "test-project",
				RepoName:  "test-repository",
				Revision: &functionspb.RepoSource_BranchName{
					BranchName: "branch-name1",
				},
				Dir: "tmp",
			},
			wantErr: false,
		},
		{
			description: "url with path to dir provided, success",
			source:      "https://source.developers.google.com/projects/test-project/repos/test-repository/paths/tmp",
			repoSource: &functionspb.RepoSource{
				ProjectId: "test-project",
				RepoName:  "test-repository",
				Revision: &functionspb.RepoSource_BranchName{
					BranchName: "master",
				},
				Dir: "tmp",
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			repoSource, err := buildSourceCSR(tc.source)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr %t, err %v", tc.wantErr, err)
			}

			if tc.wantErr && !strings.EqualFold(err.Error(), csrSourceErrorMessage) {
				t.Fatalf("wantErr %t, err %v", tc.wantErr, err)
			}

			if err == nil {
				if diff := cmp.Diff(repoSource.RepoSource, tc.repoSource, cmpopts.IgnoreUnexported(functionspb.RepoSource{})); diff != "" {
					t.Errorf("%T differ (-got, +want): %s", tc.repoSource, diff)
				}
			}
		})
	}
}

func TestBuildSourceGCS(t *testing.T) {
	type TestCase struct {
		description   string
		source        string
		storageSource *functionspb.StorageSource
		wantErr       bool
	}
	tests := []TestCase{
		{
			description: "path with subfolder and a zip archive, success",
			source:      "gs://bucket-name/subfolder/object-name.zip",
			storageSource: &functionspb.StorageSource{
				Bucket: "bucket-name",
				Object: "subfolder/object-name.zip",
			},
			wantErr: false,
		},
		{
			description: "path with zip archive, success",
			source:      "gs://bucket-name/object-name.zip",
			storageSource: &functionspb.StorageSource{
				Bucket: "bucket-name",
				Object: "object-name.zip",
			},
			wantErr: false,
		},
		{
			description: "full path, not a zip archive, fail",
			source:      "gs://bucket-name/object-name.txt",
			wantErr:     true,
		},
		{
			description: "full path, not a file, fail",
			source:      "gs://bucket-name/some-path",
			wantErr:     true,
		},
		{
			description: "bucket only, missing slash, fail",
			source:      "gs://bucket-name",
			wantErr:     true,
		},
		{
			description: "bucket only, fail",
			source:      "gs://bucket-name/",
			wantErr:     true,
		},
		{
			description: "object only, fail",
			source:      "gs:///object-name",
			wantErr:     true,
		},
		{
			description: "empty path, fail",
			source:      "gs://",
			wantErr:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			storageSource, err := buildSourceGCS(tc.source)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr %t, err %v", tc.wantErr, err)
			}

			if tc.wantErr && !strings.EqualFold(err.Error(), ccsSourceErrorMessage) {
				t.Fatalf("wantErr %t, err %v", tc.wantErr, err)
			}

			if err == nil {
				if diff := cmp.Diff(storageSource.StorageSource, tc.storageSource, cmpopts.IgnoreUnexported(functionspb.StorageSource{})); diff != "" {
					t.Errorf("%T differ (-got, +want): %s", tc.storageSource, diff)
				}
			}
		})
	}
}
