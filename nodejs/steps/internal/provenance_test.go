// Copyright 2023 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package internal

import (
	"testing"
)

func TestComputeDigest(t *testing.T) {
	cases := []struct {
		name       string
		filePath   string
		wantDigest string
		wantErr    string
	}{
		{
			name:       "valid tar file",
			filePath:   "testdata/test.tar",
			wantDigest: "fb4c942ae8d452441f0f1cf5f6b0cc507e30d2cd3b251532b685f1e22d4329a3",
		},
		{
			name:     "empty file",
			filePath: "testdata/test.txt",
			wantErr:  "empty file: testdata/test.txt",
		},
		{
			name:     "no existent file",
			filePath: "non-existent.file",
			wantErr:  "error reading non-existent.file: open non-existent.file: no such file or directory",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			digest, err := computeDigest(tc.filePath)
			if err != nil && err.Error() != tc.wantErr {
				t.Errorf("got error %v, want error %v", err, tc.wantErr)
			}
			if digest != tc.wantDigest {
				t.Errorf("unexpected digest: got %q, want %q", digest, tc.wantDigest)
			}
		})
	}

}
