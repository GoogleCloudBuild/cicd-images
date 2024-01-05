// Copyright 2023 Google LLC
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
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var mockLogger = zap.NewNop()

func TestCreateVirtualEnv(t *testing.T) {
	mockRunner := new(MockCommandRunner)
	mockRunner.On("Run", "python3", []string{"-m", "venv", "venv"}).Return(nil)

	err := CreateVirtualEnv(mockRunner, mockLogger)

	assert.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

func TestRun(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		runner := &DefaultCommandRunner{}

		err := runner.Run(mockLogger, "echo", "foo")

		assert.NoError(t, err)
	})

	t.Run("failed execution", func(t *testing.T) {
		runner := &DefaultCommandRunner{}

		err := runner.Run(mockLogger, "non-existent-command", "foo")

		assert.Error(t, err)
	})
}
