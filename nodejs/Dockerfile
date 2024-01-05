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
FROM go-base as builder

USER root
ENV PATH /opt/go-1.21/bin:/root/go/bin:$PATH

WORKDIR /usr/local/go_tools
COPY tools/go.mod tools/go.sum ./
RUN go install github.com/google/go-licenses

# Run binary tests, build and license scan.
WORKDIR /app
COPY steps/ ./

RUN go mod download && \
    go test ./... -timeout 30s && \
    go build -o node-steps

# Scan and save licenses.
RUN go-licenses check . && \
    go-licenses save . --save_path=/THIRD_PARTY_NOTICES/node-steps && \
    go-licenses report . > /THIRD_PARTY_NOTICES/node-steps/licenses.csv

# hadolint ignore=DL3006
FROM base

COPY --from=builder --chown=nonroot:nonroot /app/node-steps /usr/local/bin/node-steps
COPY --from=builder --chown=nonroot:nonroot /THIRD_PARTY_NOTICES/node-steps /THIRD_PARTY_NOTICES/node-steps