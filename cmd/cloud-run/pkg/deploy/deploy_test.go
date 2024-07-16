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

package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-run/pkg/config"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/run/v1"
)

var NotFound = func(w http.ResponseWriter, r *http.Request) {
	err := googleapi.CheckResponse(&http.Response{
		StatusCode: http.StatusNotFound,
		Status:     http.StatusText(http.StatusNotFound),
		Body:       io.NopCloser(bytes.NewReader(nil)),
	})
	http.Error(w, err.Error(), http.StatusNotFound)
}

var Unauthorized = func(w http.ResponseWriter, r *http.Request) {
	err := googleapi.CheckResponse(&http.Response{
		StatusCode: http.StatusUnauthorized,
		Status:     http.StatusText(http.StatusUnauthorized),
		Body:       io.NopCloser(bytes.NewReader(nil)),
	})
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

var OKWithBody = func(in any, t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

func TestBuildServiceDefinition(t *testing.T) {
	testProjectID := "test-project"
	testServiceName := "test-service-name"
	testImage := "docker.io/library/test"
	testCases := []struct {
		name      string
		projectID string
		opts      config.DeployOptions
		want      run.Service
	}{
		{
			name: "basic test",
			opts: config.DeployOptions{},
			want: run.Service{
				ApiVersion: "serving.knative.dev/v1",
				Kind:       "Service",
				Metadata:   &run.ObjectMeta{},
				Spec: &run.ServiceSpec{
					Template: &run.RevisionTemplate{
						Spec: &run.RevisionSpec{
							Containers: []*run.Container{{}},
						},
					},
				},
			},
		},
		{
			name:      "set namespace with projectID",
			projectID: testProjectID,
			opts:      config.DeployOptions{},
			want: run.Service{
				ApiVersion: "serving.knative.dev/v1",
				Kind:       "Service",
				Metadata:   &run.ObjectMeta{Namespace: testProjectID},
				Spec: &run.ServiceSpec{
					Template: &run.RevisionTemplate{
						Spec: &run.RevisionSpec{
							Containers: []*run.Container{{}},
						},
					},
				},
			},
		},
		{
			name: "set service Name with opts.service",
			opts: config.DeployOptions{Service: testServiceName},
			want: run.Service{
				ApiVersion: "serving.knative.dev/v1",
				Kind:       "Service",
				Metadata:   &run.ObjectMeta{Name: testServiceName},
				Spec: &run.ServiceSpec{
					Template: &run.RevisionTemplate{
						Spec: &run.RevisionSpec{
							Containers: []*run.Container{{}},
						},
					},
				},
			},
		},
		{
			name: "set the image of the first container with opts.Image",
			opts: config.DeployOptions{Image: testImage},
			want: run.Service{
				ApiVersion: "serving.knative.dev/v1",
				Kind:       "Service",
				Metadata:   &run.ObjectMeta{},
				Spec: &run.ServiceSpec{
					Template: &run.RevisionTemplate{
						Spec: &run.RevisionSpec{
							Containers: []*run.Container{{Image: testImage}},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildServiceDefinition(tc.projectID, tc.opts)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", tc.want, diff)
			}
		})
	}
}

func TestCreateOrUpdateService(t *testing.T) {
	type TestCase struct {
		description    string
		defaultProject string
		region         string
		opts           config.DeployOptions
		getMock        func(w http.ResponseWriter, r *http.Request)
		createMock     func(w http.ResponseWriter, r *http.Request)
		replaceMock    func(w http.ResponseWriter, r *http.Request)
		httpCalls      []string
		wantErr        bool
	}
	tests := []TestCase{
		{
			description:    "test deploy not found then create, success",
			defaultProject: "testProject",
			region:         "us-central1",
			opts:           config.DeployOptions{Service: "hello-world", Image: "gcr.io/cloudrun/hello"},
			getMock:        NotFound,
			createMock: OKWithBody(&run.Service{
				ApiVersion: "serving.knative.dev/v1",
				Kind:       "Service",
				Metadata:   &run.ObjectMeta{Name: "hello-world", Namespace: "testProject"},
				Spec: &run.ServiceSpec{
					Template: &run.RevisionTemplate{
						Spec: &run.RevisionSpec{
							Containers: []*run.Container{{Image: "gcr.io/cloudrun/default"}},
						},
					},
				}}, t),
			httpCalls: []string{http.MethodGet, http.MethodPost},
		},
		{
			description:    "test deploy not found then create, unauthorized",
			defaultProject: "testProject",
			region:         "us-central1",
			opts:           config.DeployOptions{Service: "hello-world", Image: "gcr.io/cloudrun/hello"},
			getMock:        NotFound,
			createMock:     Unauthorized,
			httpCalls:      []string{http.MethodGet, http.MethodPost},
			wantErr:        true,
		},
		{
			description:    "test deploy GET unauthorized",
			defaultProject: "testProject",
			region:         "us-central1",
			opts:           config.DeployOptions{Service: "hello-world", Image: "gcr.io/cloudrun/hello"},
			getMock:        Unauthorized,
			httpCalls:      []string{http.MethodGet},
			wantErr:        true,
		},
		{
			description:    "test deploy GET found, replace Unauthorized",
			defaultProject: "testProject",
			region:         "us-central1",
			opts:           config.DeployOptions{Service: "hello-world", Image: "gcr.io/cloudrun/hello"},
			getMock: OKWithBody(&run.Service{
				ApiVersion: "serving.knative.dev/v1",
				Kind:       "Service",
				Metadata:   &run.ObjectMeta{},
				Spec: &run.ServiceSpec{
					Template: &run.RevisionTemplate{
						Spec: &run.RevisionSpec{
							Containers: []*run.Container{{Image: "gcr.io/cloudrun/default"}},
						},
					},
				}}, t),
			replaceMock: Unauthorized,
			httpCalls:   []string{http.MethodGet, http.MethodPut},
			wantErr:     true,
		},
		{
			description:    "test deploy GET found, replace success",
			defaultProject: "testProject",
			region:         "us-central1",
			opts:           config.DeployOptions{Service: "hello-world", Image: "gcr.io/cloudrun/updated"},
			getMock: OKWithBody(&run.Service{
				ApiVersion: "serving.knative.dev/v1",
				Kind:       "Service",
				Metadata:   &run.ObjectMeta{},
				Spec: &run.ServiceSpec{
					Template: &run.RevisionTemplate{
						Spec: &run.RevisionSpec{
							Containers: []*run.Container{{Image: "gcr.io/cloudrun/default"}},
						},
					},
				}}, t),
			replaceMock: func(w http.ResponseWriter, r *http.Request) {
				var s run.Service
				err := json.NewDecoder(r.Body).Decode(&s)
				if err != nil {
					t.Fatalf("failed to decode request body to service %v", err)
				}
				got := s.Spec.Template.Spec.Containers[0].Image
				if got != "gcr.io/cloudrun/updated" {
					t.Fatalf("expect image to be updated, but it didn't happen")
				}
				w.WriteHeader(http.StatusOK)
				b, err := json.Marshal(&s)
				if err != nil {
					t.Fatalf("failed to marshal %v", err)
				}
				_, err = w.Write(b)
				if err != nil {
					t.Fatalf("fail to write %v", err)
				}
			},
			httpCalls: []string{http.MethodGet, http.MethodPut},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var calls []string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					calls = append(calls, http.MethodGet)
					if tc.getMock == nil {
						t.Fatalf("getMock not set")
					}
					tc.getMock(w, r)
				} else if r.Method == http.MethodPost {
					calls = append(calls, http.MethodPost)
					if tc.createMock == nil {
						t.Fatalf("createMock not set")
					}
					tc.createMock(w, r)
				} else if r.Method == http.MethodPut {
					calls = append(calls, http.MethodPut)
					if tc.replaceMock == nil {
						t.Fatalf("replaceMock not set")
					}
					tc.replaceMock(w, r)
				}
			}))

			runClient, err := run.NewService(context.Background(), option.WithEndpoint(ts.URL), option.WithoutAuthentication())
			if err != nil {
				t.Fatalf("failed to create runClient %v", err)
			}
			err = CreateOrUpdateService(runClient, tc.defaultProject, tc.region, tc.opts)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr %t, err %v", tc.wantErr, err)
			}

			if diff := cmp.Diff(calls, tc.httpCalls); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", tc.httpCalls, diff)
			}
		})
	}
}
