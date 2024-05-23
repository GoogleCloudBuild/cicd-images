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
package command

import (
	"fmt"
	"reflect"
)

// MockCommandRunner is a mock implementation of CommandRunner.
type MockCommandRunner struct {
	CalledCommands   []string
	ExpectedCommands []string
	Err              error
}

func (m *MockCommandRunner) Run(cmd string, args ...string) error {
	m.CalledCommands = append(m.CalledCommands, cmd)
	m.CalledCommands = append(m.CalledCommands, args...)

	if m.Err != nil {
		return m.Err
	}

	if !reflect.DeepEqual(m.CalledCommands, m.ExpectedCommands) {
		return fmt.Errorf("unexpected commands: got %v, expected %v", m.CalledCommands, m.ExpectedCommands)
	}

	return nil
}
