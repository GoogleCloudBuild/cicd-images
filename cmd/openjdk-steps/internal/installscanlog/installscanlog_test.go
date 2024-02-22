// Copyright 2024 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package installscanlog

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetrieveArtifactPaths(t *testing.T) {
	// Get test data and convert to a string
	content, err := os.ReadFile("testdata/test-install-log.txt")
	assert.NoError(t, err)

	logs := string(content)
	expectedResult := map[string]bool{
		"./target/": true,
	}

	t.Run("Full log of mvn install", func(t *testing.T) {
		result, err := retrieveArtifactPaths(logs)
		assert.NoError(t, err)

		if !reflect.DeepEqual(result, expectedResult) {
			t.Errorf("Expected %v\n, but got %v", expectedResult, result)
		}
	})

}
