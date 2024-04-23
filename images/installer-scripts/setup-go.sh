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

set -o errexit  \
    -o nounset  \
    -o pipefail \
    -o nounset
#    -o xtrace \

 . "$(dirname "$(readlink $0 -f)")"/env.sh

install_dependencies(){
  GIT=$(which git)
  if [[ -z "$GIT" ]] || [[ ! -x "$GIT" ]]; then
    bash git.sh
  fi
}

# Set Default versions
GO_KEY='language-runtimes.go'
GO_VERSION=$(yq ".${GO_KEY}.default-version" <$PACKAGES)

# Override versions based on args
GO_VERSION=${1:-$GO_VERSION}
if [[ "$#" -gt 1 ]]; then
  err_exit "$0: Unexpected args $*"
fi

export GO_VERSION

# Install Go
MINOR_VERSION=$(echo $GO_VERSION | cut -d. -f1-2)

BIN_DIR="/opt/go-$MINOR_VERSION/bin"

INSTALL="false"
if [[ ! -d $BIN_DIR ]]; then
    is_version_supported $GO_VERSION $GO_KEY \
    || err_exit "Go version $GO_VERSION is not supported!"
    INSTALL="true"
fi

RUNTIME_URL="https://go.dev/dl/"

CURL_CMD="curl -fsSL -A GCPBuildpacks"

if [[ "$INSTALL" == "true" ]]; then
  VERS_URL="${RUNTIME_URL}?mode=json&include=all"
  FULL_VERSION=$($CURL_CMD $VERS_URL | jq -r '.[] | .version' | \
                grep "^go${MINOR_VERSION}" | sort -V | tail -1 )

  #Normalize FULL_VERSION
  FULL_VERSION=$(echo $FULL_VERSION | tr '+' '_')
  #Install required version
  DOWNLOAD_FILE="${FULL_VERSION}.linux-amd64.tar.gz"
  DOWNLOAD_URL="${RUNTIME_URL}/${DOWNLOAD_FILE}"
  $CURL_CMD -o /tmp/${DOWNLOAD_FILE} $DOWNLOAD_URL
  mkdir -p "/opt/go-${MINOR_VERSION}"
  tar --strip-components=1 -zxf /tmp/${DOWNLOAD_FILE} \
      -C "/opt/go-${MINOR_VERSION}"
  rm /tmp/${DOWNLOAD_FILE}
fi

#Create links to installed binaries
ln -s ${BIN_DIR}/* -t /usr/local/bin

copy_licenses "$( dirname ${BIN_DIR} )"

mkdir -p "${HOME}/go/src" "${HOME}/go/bin"

update_env PATH "${HOME}/go/bin:${BIN_DIR}:${PATH}"
