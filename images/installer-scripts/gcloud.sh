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

CLOUD_SDK_KEY=/etc/apt/keyrings/cloud.google.gpg
CLOUD_SDK_URL="https://packages.cloud.google.com/apt"
echo "deb [signed-by=$CLOUD_SDK_KEY] $CLOUD_SDK_URL cloud-sdk main" \
    | tee /etc/apt/sources.list.d/google-cloud-sdk.list
curl  -fsSL "https://packages.cloud.google.com/apt/doc/apt-key.gpg" \
    | gpg --yes --dearmor -o $CLOUD_SDK_KEY

clean-install apt-transport-https python3 google-cloud-cli

COMMON_COMPONENTS="google-cloud-cli-package-go-module"
ALL_COMPONENTS="
    $COMMON_COMPONENTS \
    google-cloud-cli-app-engine-go \
    google-cloud-cli-app-engine-java \
    google-cloud-cli-app-engine-python \
    google-cloud-cli-app-engine-python-extras \
    google-cloud-cli-gke-gcloud-auth-plugin \
    google-cloud-cli-cbt \
    google-cloud-cli-bigtable-emulator \
    google-cloud-cli-datastore-emulator \
    google-cloud-cli-firestore-emulator \
    google-cloud-cli-kpt \
    google-cloud-cli-local-extract \
    google-cloud-cli-pubsub-emulator \
    google-cloud-cli-skaffold \
    google-cloud-cli-anthos-auth \
    google-cloud-cli-app-engine-grpc \
    google-cloud-cli-config-connector \
    google-cloud-cli-istioctl \
    google-cloud-cli-kubectl-oidc \
    google-cloud-cli-nomos \
    google-cloud-cli-spanner-migration-tool \
    google-cloud-cli-terraform-tools \
    google-cloud-sdk \
    google-cloud-sdk-anthos-auth \
    google-cloud-sdk-app-engine-go \
    google-cloud-sdk-app-engine-grpc \
    google-cloud-sdk-app-engine-java \
    google-cloud-sdk-app-engine-python \
    google-cloud-sdk-app-engine-python-extras \
    google-cloud-sdk-cbt \
    google-cloud-sdk-config-connector \
    google-cloud-sdk-datalab \
    google-cloud-sdk-gke-gcloud-auth-plugin \
    google-cloud-sdk-istioctl \
    google-cloud-sdk-kpt \
    google-cloud-sdk-kubectl-oidc \
    google-cloud-sdk-nomos \
    google-cloud-sdk-package-go-module \
    google-cloud-sdk-skaffold \
    google-cloud-sdk-spanner-migration-tool \
    google-cloud-sdk-terraform-tools \
    google-cloud-sdk-terraform-validator \
    kubectl
"

INSTALL_COMPONENTS="$COMMON_COMPONENTS"
INSTALL_OPTS=""
if [[ $# -gt 0 ]]; then
    INSTALL_OPTS="$1"
fi

case $INSTALL_OPTS in
  -full)
    INSTALL_COMPONENTS="$ALL_COMPONENTS"
    ;;
  -slim)
    INSTALL_COMPONENTS="$COMMON_COMPONENTS"
    ;;
esac

if [[ -n "$INSTALL_COMPONENTS" ]]; then
  clean-install $INSTALL_COMPONENTS
fi

# anthoscli is for legacy usage, we don't support it
[[ ! -e /usr/lib/google-cloud-sdk/bin/anthoscli ]] || rm /usr/lib/google-cloud-sdk/bin/anthoscli
#Bundeled python has some patching issues, we have installed python3
rm -rf /usr/lib/google-cloud-sdk/platform/bundledpythonunix

# Cleaning reduces image size
rm -rf /usr/lib/google-cloud-sdk/.install && \
find /usr/lib/google-cloud-sdk/ -name "*.pyc" -exec rm -rf '{}' + && \
find /usr/lib/google-cloud-sdk \
  -type d \( -name 'tests' -o -name 'test' \) \
  -path '*/third_party/*' -exec rm -rf {} +
