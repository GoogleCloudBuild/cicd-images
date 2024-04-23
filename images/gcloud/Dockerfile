# syntax=docker/dockerfile:1
#
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
#

# hadolint ignore=DL3006
FROM base

USER root

RUN apt-get update -y \
    && clean-install apt-transport-https python3

# Install gcloud CLI and prune unnecessary files
# hadolint ignore=DL4006
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" \
    | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl "https://packages.cloud.google.com/apt/doc/apt-key.gpg" \
    | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    apt-get update -y && \
    clean-install google-cloud-cli && \
    rm -rf /usr/lib/google-cloud-sdk/platform/bundledpythonunix && \
    rm -rf /usr/lib/google-cloud-sdk/.install && \
    find /usr/lib/google-cloud-sdk/ -name "*.pyc" -exec rm -rf '{}' + && \
    find /usr/lib/google-cloud-sdk -type d \( -name 'tests' -o -name 'test' \) -path '*/third_party/*' -exec rm -rf {} +

ENV USER="nonroot"
USER $USER

WORKDIR /home/$USER

ENTRYPOINT ["gcloud"]
