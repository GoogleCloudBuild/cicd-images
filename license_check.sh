#!/usr/bin/env bash
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
set -e
tar_filename="THIRD_PARTY_NOTICES.tar.gz"

# Run on main module directory. directory
if [[ "$(git diff --name-only HEAD | grep -q "go.mod")" ]] || [[ ! -f "$tar_filename" ]]; then
    if [[ -f "$tar_filename" ]]; then
        rm $tar_filename
    fi
    echo "Creating $tar_filename..."
    go-licenses check ./...
    go-licenses save ./... --save_path=./THIRD_PARTY_NOTICES
    go-licenses report ./... > ./THIRD_PARTY_NOTICES/licenses.csv
    tar -czvf "$tar_filename" ./THIRD_PARTY_NOTICES
    rm -rf ./THIRD_PARTY_NOTICES
fi