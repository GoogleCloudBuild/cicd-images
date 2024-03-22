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
	"time"
)

// PollWithInterval repeatedly executes a condition function until the condition is met, a timeout occurs, or the context is canceled.
func PollWithInterval(ctx context.Context, timeout, interval time.Duration, conditionFunc func() (bool, error)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := conditionFunc()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}
