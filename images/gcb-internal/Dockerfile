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

# Install go-steps binary
# hadolint ignore=DL3007
FROM gcr.io/cloud-marketplace/google/debian12:latest as builder

ENV GOPATH=$HOME/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# hadolint ignore=DL3008
RUN apt-get update -y && \
    apt-get -y upgrade && \
    apt-get install -y --no-install-recommends ca-certificates golang git

# Get version Go 1.22.4
RUN go install golang.org/dl/go1.22.4@latest && \
    go1.22.4 download

WORKDIR /src
COPY --from=src . .

RUN go1.22.4 mod download && \
    go1.22.4 test ./cmd/go-steps/... -timeout 30s && \
    go1.22.4 build ./cmd/go-steps

# hadolint ignore=DL3007
FROM gcr.io/cloud-marketplace/google/debian12:latest

# hadolint ignore=DL3008
RUN <<EOF
    apt-get update -y
    apt-get -y upgrade
    apt-get install -y --no-install-recommends \
        xxd \
        nodejs npm \
        maven \
        twine python3-pip
    pip install keyrings.google-artifactregistry-auth==1.1.2 --break-system-packages --no-cache-dir
EOF

COPY --from=builder /src/go-steps /usr/local/bin/go-steps

