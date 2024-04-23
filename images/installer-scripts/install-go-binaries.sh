#!/bin/bash

# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit \
    -o nounset \
    -o pipefail
#    -o xtrace \

 . "$(dirname "$(readlink $0 -f)")"/env.sh

get_mod_dir() {
  while getopts ":m:" opt; do
    case $opt in
      m)
        MODULE_DIR="$OPTARG"
        ;;
      \?)
        err_exit "$0: Invalid option: -$OPTARG"
        ;;
    esac
  done
}

# We assume current dir is a go module,
# unless user has provided a diffrent dir
MODULE_DIR=$(pwd)
get_mod_dir "$@"
shift $((OPTIND-1)) # Shift arguments past the options
if [[ "$#" != 0 ]]; then
  err_exit "$0: Unexpected args $*"
fi

# go-tools installation depends on `yq` and `go-license`
# install these first
(
  cd $SRC/tools;
  go install github.com/mikefarah/yq/v4
  go install github.com/google/go-licenses
)

cd $MODULE_DIR

# We expect $MODULE_DIR to be a valid module dir
if [[ "$(go env GOMOD)" == "/dev/null" ]]; then
  err_exit "$MODULE_DIR is not a valid go module"
fi

tools=$(yq '.go-tools|keys|.[]' < $PACKAGES)
IGNORE_LICENSE=( "--ignore" "github.com/xi2/xz"
                  "--ignore" "modernc.org/mathutil"
                   "--ignore" "github.com/deitch/magic/pkg/magic"
                )

set -o xtrace
for tool in $tools; do

  source=$(yq ".go-tools | filter( key == \"$tool\").[].source" < $PACKAGES)
  #go get "${source}"
  go install "${source}"

  go-licenses check "${IGNORE_LICENSE[@]}" "${source}"
  go-licenses save  "${IGNORE_LICENSE[@]}" "${source}"  \
    --save_path="/THIRD_PARTY_NOTICES/${tool}"
  go-licenses report  "${IGNORE_LICENSE[@]}" \
    "${source}" >> /THIRD_PARTY_NOTICES/${tool}/licenses.csv

done
