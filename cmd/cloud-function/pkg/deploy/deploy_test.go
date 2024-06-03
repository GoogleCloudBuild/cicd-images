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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/config"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/utils"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
	pb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/genproto/googleapis/rpc/status"
)

var NotFound = func(w http.ResponseWriter, _ *http.Request) {
	err := googleapi.CheckResponse(&http.Response{
		StatusCode: http.StatusNotFound,
		Status:     http.StatusText(http.StatusNotFound),
		Body:       io.NopCloser(bytes.NewReader(nil)),
	})
	http.Error(w, err.Error(), http.StatusNotFound)
}

var Unauthorized = func(w http.ResponseWriter, _ *http.Request) {
	err := googleapi.CheckResponse(&http.Response{
		StatusCode: http.StatusUnauthorized,
		Status:     http.StatusText(http.StatusUnauthorized),
		Body:       io.NopCloser(bytes.NewReader(nil)),
	})
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

var OKWithBody = func(in protoreflect.ProtoMessage, t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		mo := protojson.MarshalOptions{UseProtoNames: true}
		marshal, err := mo.Marshal(in)
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

var OngoingFunctionOperation = func(name string) *pb.Operation {
	return &pb.Operation{
		Name: name,
		Done: false,
	}
}

var SuccessfulFunctionOperation = func(name string, function *functionspb.Function, t *testing.T) *pb.Operation {
	responseAny, err := anypb.New(function)
	if err != nil {
		t.Fatal(err)
	}
	return &pb.Operation{
		Name: name,
		Done: true,
		Result: &pb.Operation_Response{
			Response: responseAny,
		},
	}
}

var FailedFunctionOperation = func(name string, errCode codes.Code, errMsg string, t *testing.T) *pb.Operation {
	details := &errdetails.ErrorInfo{Reason: errMsg}
	a, err := anypb.New(details)
	if err != nil {
		t.Fatal(err)
	}
	return &pb.Operation{
		Name: name,
		Done: true,
		Result: &pb.Operation_Error{
			Error: &status.Status{
				Code:    int32(errCode),
				Message: errMsg,
				Details: []*anypb.Any{a},
			},
		},
	}
}

func TestCreateOrUpdateFunction(t *testing.T) {
	type TestCase struct {
		description           string
		opts                  config.DeployOptions
		getFunctionMock       func(w http.ResponseWriter, r *http.Request)
		getOperationMock      func(w http.ResponseWriter, r *http.Request)
		createMock            func(w http.ResponseWriter, r *http.Request)
		replaceMock           func(w http.ResponseWriter, r *http.Request)
		generateUploadURLMock func(w http.ResponseWriter, r *http.Request)
		httpCalls             []string
		wantErr               bool
	}
	tests := []TestCase{
		{
			description: "test deploy from GSR, not supported, fail",
			opts: config.DeployOptions{
				ProjectID:  "test-project",
				Region:     "us-central1",
				Name:       "test-function",
				Source:     "https://source.developers.google.com/projects/test-project/repos/test-repository/moveable-aliases/branch1",
				EntryPoint: "hello-world-function",
				Runtime:    "go121",
			},
			wantErr: true,
		},
		{
			description: "test deploy from GCS, not supported, fail",
			opts: config.DeployOptions{
				ProjectID:  "test-project",
				Region:     "us-central1",
				Name:       "test-function",
				Source:     "gs://bucket/object.zip",
				EntryPoint: "hello-world-function",
				Runtime:    "go121",
			},
			wantErr: true,
		},
		{
			description: "test deploy from loacl file system with labels, not found then create, success",
			opts: config.DeployOptions{
				ProjectID:  "test-project",
				Region:     "us-central1",
				Name:       "test-function",
				Source:     "tmp/source-code",
				EntryPoint: "hello-world-function",
				Runtime:    "go121",
				Labels: map[string]string{
					"label1": "value",
				},
			},
			getFunctionMock: NotFound,
			createMock:      OKWithBody(OngoingFunctionOperation("operation-id"), t),
			getOperationMock: OKWithBody(SuccessfulFunctionOperation("operation-id", &functionspb.Function{
				Name: "test-function",
				Labels: map[string]string{
					"label1": "value",
				},
				BuildConfig: &functionspb.BuildConfig{
					Runtime:    "go121",
					EntryPoint: "hello-world-function",
					Source: &functionspb.Source{
						Source: &functionspb.Source_StorageSource{
							StorageSource: &functionspb.StorageSource{
								Bucket: "",
								Object: "",
							},
						},
					},
				},
				Environment: functionspb.Environment_GEN_1,
			}, t), t),
			generateUploadURLMock: OKWithBody(&functionspb.GenerateUploadUrlResponse{
				UploadUrl: "http://domain.com/upload",
			}, t),
			httpCalls: []string{http.MethodPost, http.MethodGet, http.MethodPost, http.MethodGet},
		},
		{
			description: "test deploy from local file system,, function exists, update, success",
			opts: config.DeployOptions{
				ProjectID:  "test-project",
				Region:     "us-central1",
				Name:       "test-function",
				Source:     "tmp/function",
				EntryPoint: "bye-world-function",
				Runtime:    "go121",
			},
			getFunctionMock: OKWithBody(&functionspb.Function{
				Name:        "test-function",
				Environment: functionspb.Environment_GEN_1,
			}, t),
			replaceMock: OKWithBody(OngoingFunctionOperation("operation-id"), t),
			getOperationMock: OKWithBody(SuccessfulFunctionOperation("operation-id", &functionspb.Function{
				Name: "test-function",
				BuildConfig: &functionspb.BuildConfig{
					Runtime:    "go121",
					EntryPoint: "bye-world-function",
					Source: &functionspb.Source{
						Source: &functionspb.Source_StorageSource{
							StorageSource: &functionspb.StorageSource{
								Bucket: "",
								Object: "",
							},
						},
					},
				},
				Environment: functionspb.Environment_GEN_1,
			}, t), t),
			generateUploadURLMock: OKWithBody(&functionspb.GenerateUploadUrlResponse{
				UploadUrl: "http://domain.com/upload",
			}, t),
			httpCalls: []string{http.MethodPost, http.MethodGet, http.MethodPatch, http.MethodGet},
		},
		{
			description: "test deploy from local file system, not found then create with permission denied, fail",
			opts: config.DeployOptions{
				ProjectID:  "test-project",
				Region:     "us-central1",
				Name:       "test-function",
				Source:     "tmp/source-code",
				EntryPoint: "hello-world-function",
				Runtime:    "go121",
			},
			getFunctionMock:  NotFound,
			createMock:       OKWithBody(OngoingFunctionOperation("operation-id"), t),
			getOperationMock: OKWithBody(FailedFunctionOperation("operation-id", codes.PermissionDenied, "permission denied", t), t),
			generateUploadURLMock: OKWithBody(&functionspb.GenerateUploadUrlResponse{
				UploadUrl: "http://domain.com/upload",
			}, t),
			httpCalls: []string{http.MethodPost, http.MethodGet, http.MethodPost, http.MethodGet},
			wantErr:   true,
		},
		{
			description: "test deploy from local file system, not found then create, success",
			opts: config.DeployOptions{
				ProjectID:  "test-project",
				Region:     "us-central1",
				Name:       "test-function",
				Source:     ".",
				EntryPoint: "hello-world-function",
				Runtime:    "go121",
			},
			getFunctionMock: NotFound,
			createMock:      OKWithBody(OngoingFunctionOperation("operation-id"), t),
			getOperationMock: OKWithBody(SuccessfulFunctionOperation("operation-id", &functionspb.Function{
				Name: "test-function",
				BuildConfig: &functionspb.BuildConfig{
					Runtime:    "go121",
					EntryPoint: "bye-world-function",
					Source: &functionspb.Source{
						Source: &functionspb.Source_StorageSource{
							StorageSource: &functionspb.StorageSource{
								Bucket: "",
								Object: "",
							},
						},
					},
				},
				Environment: functionspb.Environment_GEN_1,
			}, t), t),
			generateUploadURLMock: OKWithBody(&functionspb.GenerateUploadUrlResponse{
				UploadUrl: "http://domain.com/upload",
			}, t),
			httpCalls: []string{http.MethodPost, http.MethodGet, http.MethodPost, http.MethodGet},
			wantErr:   false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var calls []string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet:
					calls = append(calls, http.MethodGet)
					switch {
					case strings.Contains(r.RequestURI, "projects/test-project/locations/us-central1/functions/test-function"):
						if tc.getFunctionMock == nil {
							t.Fatalf("getFunctionMock not set")
						}
						tc.getFunctionMock(w, r)
					case strings.Contains(r.RequestURI, "/v2/operation-id"):
						if tc.getOperationMock == nil {
							t.Fatalf("getOperationMock not set")
						}
						tc.getOperationMock(w, r)
					default:
						w.WriteHeader(http.StatusOK)
					}
				case r.Method == http.MethodPost:
					calls = append(calls, http.MethodPost)
					switch {
					case strings.Contains(r.RequestURI, "functions:generateUploadUrl"):
						if tc.generateUploadURLMock == nil {
							t.Fatalf("generateUploadURLMock not set")
						}
						tc.generateUploadURLMock(w, r)
					default:
						if tc.createMock == nil {
							t.Fatalf("createMock not set")
						}
						tc.createMock(w, r)
					}
				case r.Method == http.MethodPatch:
					calls = append(calls, http.MethodPatch)
					if tc.replaceMock == nil {
						t.Fatalf("replaceMock not set")
					}
					tc.replaceMock(w, r)
				default:
					t.Fatalf("unexpected http method")
				}
			}))

			ctx := context.Background()
			client, err := functions.NewFunctionRESTClient(ctx, option.WithEndpoint(ts.URL), option.WithoutAuthentication())
			if err != nil {
				t.Fatalf("failed to create client %v", err)
			}

			err = CreateOrUpdateFunction(ctx, client, utils.MockFileUtil{}, tc.opts)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr %t, err %v", tc.wantErr, err)
			}

			if diff := cmp.Diff(calls, tc.httpCalls); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", tc.httpCalls, diff)
			}
		})
	}
}
