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
    -o xtrace \
    -o pipefail

 . "$(dirname "$(readlink $0 -f)")"/env.sh

# Only check var declaration after defaults are set
set -o nounset

# Install docker
DOCKER_VERSION=$(yq '."build-tools".docker.version' $PACKAGES)
DOCKER_BUILDX_VERSION=$(yq '."build-tools"."docker-buildx".version' $PACKAGES)
DOCKER_COMPOSE_VERSION=$(yq '."build-tools"."docker-compose".version' $PACKAGES)

install -m 0755 -d /etc/apt/keyrings;
DOCKER_GPG_KEY=/etc/apt/keyrings/docker.gpg
DOCKER_GPG_KEY_URL="https://download.docker.com/linux/ubuntu/gpg"
curl -fsSL  $DOCKER_GPG_KEY_URL | gpg --dearmor -o $DOCKER_GPG_KEY

ARCH=$(dpkg --print-architecture)
DOCKER_APT_URL="https://download.docker.com/linux/ubuntu"
echo -e \
  "# Add the docker repository to Apt sources:\n" \
  "deb [arch=$ARCH signed-by=$DOCKER_GPG_KEY] $DOCKER_APT_URL \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

clean-install \
  docker-ce-cli="5:${DOCKER_VERSION}.*" \
  docker-compose-plugin="${DOCKER_COMPOSE_VERSION}.*" \
  docker-buildx-plugin="${DOCKER_BUILDX_VERSION}.*" \
  ;
ln -fs /usr/libexec/docker/cli-plugins/* /usr/local/bin 

docker --version
docker-compose version
docker-buildx version


