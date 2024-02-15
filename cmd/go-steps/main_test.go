//  Copyright 2023 Google LLC
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

package main

import (
	"reflect"
	"testing"
)

func TestExtractLocation(t *testing.T) {
	type test struct {
		name          string
		url           string
		want          string
		expectedError string
	}

	tests := []test{
		{name: "valid url", url: "us-east-go.pkg.dev", want: "us-east"},
		{name: "valid url with https", url: "https://us-go.pkg.dev/foo/baz", want: "us"},
		{name: "valid url with http", url: "http://us-go.pkg.dev/foo/baz", want: "us"},
		{
			name: "valid url with multiple proxies",
			url:  "https://us-go.pkg.dev/foo/baz,https://proxy.golang.org,direct",
			want: "us",
		},
		{
			name: "valid url with multiple proxies",
			url:  "us-go.pkg.dev/foo/baz,https://proxy.golang.org,direct",
			want: "us",
		},
		{
			name: "valid url with multiple proxies in the beginning",
			url:  "https://proxy.golang.org,direct,us-go.pkg.dev/foo/baz",
			want: "us",
		},
		{
			name: "valid url with multiple proxies in the beginning and https",
			url:  "https://proxy.golang.org,direct,https://us-go.pkg.dev/foo/baz",
			want: "us",
		},
		{
			name:          "invalid url",
			url:           "https://proxy.golang.org",
			want:          "us-east",
			expectedError: "Could not get location from proxy. please include a proxy with the format: https://{LOCATION}-go.pkg.dev/foo/baz"},
	}

	for _, tc := range tests {
		got, err := extractLocation(tc.url)

		if err != nil && tc.expectedError != err.Error() {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(tc.want, got) && reflect.DeepEqual(tc.expectedError, "") {
			t.Fatalf("Test %v, expected: %v, got: %v", tc.name, tc.want, got)
		}
	}
}
