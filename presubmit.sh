#!/usr/bin/env bash
# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# TODO(kmonty): Move this to the Makefile so it's easier to know how to run locally too.
set -eu

make install-tools

# Run pre-commit hooks
make pre-commit

# Test image builds if the source was modified
changed_image_dirs=$(git show --name-only --pretty="" HEAD | xargs dirname | grep -E "^images/" | sed 's#\(images/[^/]*/\).*#\1#' | uniq)

for dir in $changed_image_dirs; do
  if [ -d "${dir}" ] ; then
      make -C "${dir}" build test
  fi
done
