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

# Set Default versions
NODEJS_KEY='language-runtimes.nodejs'
NODEJS_VERSION=$(yq ".${NODEJS_KEY}.default-version" <$PACKAGES)

# Override versions based on args
NODEJS_VERSION=${1:-$NODEJS_VERSION}
if [[ "$#" -gt 1 ]]; then
  err_exit "$0: Unexpected args $*"
fi
export NODEJS_VERSION

MAJOR_VERSION=$(echo $NODEJS_VERSION | cut -d. -f1)

BIN_DIR="/opt/node-${MAJOR_VERSION}/bin"

INSTALL="false"
if [[ ! -d $BIN_DIR ]]; then
    is_version_supported $NODEJS_VERSION $NODEJS_KEY \
    || err_exit "Nodejs version $NODEJS_VERSION is not supported!"
    INSTALL="true"
fi

RUNTIME_URL="https://dl.google.com/runtimes/ubuntu2204"

CURL_CMD="curl -fsSL -A GCPBuildpacks"
if [[ "$INSTALL" == "true" ]]; then
  VERS_URL="${RUNTIME_URL}/nodejs/version.json"
  FULL_VERSION=$($CURL_CMD $VERS_URL | jq '.[]' | sed -e 's/"//g' | \
                grep "^$MAJOR_VERSION" | sort -V | tail -1 )
  #Install required version
  CURL_CMD="curl -fsSL -A GCPBuildpacks"
  DOWNLOAD_FILE="nodejs-${FULL_VERSION}.tar.gz"
  DOWNLOAD_URL="${RUNTIME_URL}/nodejs/${DOWNLOAD_FILE}"
  $CURL_CMD -o /tmp/${DOWNLOAD_FILE} $DOWNLOAD_URL
  mkdir -p "/opt/node-${MAJOR_VERSION}"
  tar -zxf /tmp/${DOWNLOAD_FILE} \
    -C "/opt/node-${MAJOR_VERSION}"
fi

copy_licenses "$( dirname ${BIN_DIR} )"

#Create links to installed binaries
ln -s ${BIN_DIR}/* -t /usr/local/bin

update_env PATH "${BIN_DIR}:${PATH}"

corepack enable
