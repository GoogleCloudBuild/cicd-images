# syntax=docker/dockerfile:1
# 
# Copyright 2024 Google LLC
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

FROM google-go.pkg.dev/golang:1.21.5 as builder

WORKDIR /shared-tools
COPY --from=src tools/ ./
RUN go install github.com/google/go-licenses

WORKDIR /tools
COPY tools/ ./
RUN go install github.com/GoogleCloudPlatform/cloud-builders/gke-deploy
RUN go-licenses check github.com/GoogleCloudPlatform/cloud-builders/gke-deploy && \
    go-licenses save github.com/GoogleCloudPlatform/cloud-builders/gke-deploy --save_path=/THIRD_PARTY_NOTICES/gke-deploy && \
    go-licenses report github.com/GoogleCloudPlatform/cloud-builders/gke-deploy > /THIRD_PARTY_NOTICES/gke-deploy/licenses.csv

WORKDIR /src
COPY --from=src . .
RUN go mod download && \
    go test ./cmd/google-cloud-auth/... -timeout 30s && \
    go build ./cmd/google-cloud-auth
RUN go-licenses check ./cmd/google-cloud-auth && \
    go-licenses save ./cmd/google-cloud-auth --save_path=/THIRD_PARTY_NOTICES/google-cloud-auth && \
    go-licenses report ./cmd/google-cloud-auth > /THIRD_PARTY_NOTICES/google-cloud-auth/licenses.csv

# hadolint ignore=DL3006
FROM base
## Install kubectl and gke-gcloud-auth-plugin
USER root:root
SHELL [ "/bin/bash", "-exo", "pipefail", "-c" ]
# hadolint ignore=DL3008
RUN apt-get install --no-install-recommends -y apt-transport-https
RUN curl -fsSL "https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key" | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
RUN echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list
RUN curl "https://packages.cloud.google.com/apt/doc/apt-key.gpg" | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" |  tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
RUN apt-get update && \
clean-install kubectl google-cloud-cli-gke-gcloud-auth-plugin

ARG USER=nonroot
USER $USER:$USER

COPY --from=builder /go/bin/gke-deploy /usr/local/bin/gke-deploy
COPY --from=builder /THIRD_PARTY_NOTICES/gke-deploy /THIRD_PARTY_NOTICES/gke-deploy
COPY --from=builder /src/google-cloud-auth  /usr/local/bin/google-cloud-auth
COPY --from=builder /THIRD_PARTY_NOTICES/google-cloud-auth /THIRD_PARTY_NOTICES/google-cloud-auth
