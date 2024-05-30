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

GIT_VERSION=$(yq '.build-tools.git.version' <$PACKAGES)

add-apt-repository -ny ppa:git-core/ppa
clean-install git="1:${GIT_VERSION}.*"

if [[ $# -lt 1 ]]; then
  GIT_LFS=""
else
  GIT_LFS=${1#-}
fi
if [[ "$GIT_LFS" == "git-lfs" ]]; then
    install -m 0755 -d /etc/apt/keyrings;
    GPG_KEY_FILE=/etc/apt/keyrings/github_git-lfs.gpg
    gpg_key_url="https://packagecloud.io/github/git-lfs/gpgkey"
    curl -fsSL $gpg_key_url | gpg --dearmor -o $GPG_KEY_FILE ;

    GIT_LFS_URL="https://packagecloud.io/github/git-lfs/ubuntu/"
    ARCH="$(dpkg --print-architecture)"
    echo -e \
    "# Add the git-lfs repository to Apt sources:\n" \
    "deb [arch=$ARCH signed-by=$GPG_KEY_FILE] $GIT_LFS_URL \
    $(. /etc/os-release && echo "$VERSION_CODENAME") main" | \
    tee /etc/apt/sources.list.d/github_git-lfs.list > /dev/null ;
    GIT_LFS_VERSION=$(yq '.build-tools.git-lfs.version' <$PACKAGES)
    clean-install git-lfs="${GIT_LFS_VERSION}.*"
fi

git config --system credential.helper gcloud.sh
git config --global --add safe.directory "*"
