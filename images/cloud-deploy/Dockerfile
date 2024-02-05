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

FROM google-go.pkg.dev/golang:1.21.5 as builder

WORKDIR /src
COPY --from=src . .

WORKDIR /src/tools
RUN go install github.com/google/go-licenses

# Run binary tests, build and license scan.
WORKDIR /src
RUN go mod download && \
    go test ./cmd/cloud-deploy/... -timeout 30s && \
    go test ./cmd/google-cloud-auth/... -timeout 30s && \
    go build ./cmd/cloud-deploy && \
    go build ./cmd/google-cloud-auth

# Scan and save licenses.
RUN go-licenses check ./cmd/cloud-deploy && \
    go-licenses save ./cmd/cloud-deploy --save_path=/THIRD_PARTY_NOTICES/cloud-deploy && \
    go-licenses report ./cmd/cloud-deploy > /THIRD_PARTY_NOTICES/cloud-deploy/licenses.csv

# Scan and save licenses.
RUN go-licenses check ./cmd/google-cloud-auth && \
    go-licenses save ./cmd/google-cloud-auth --save_path=/THIRD_PARTY_NOTICES/google-cloud-auth && \
    go-licenses report ./cmd/google-cloud-auth > /THIRD_PARTY_NOTICES/google-cloud-auth/licenses.csv

# hadolint ignore=DL3006
FROM base

COPY --from=builder /src/cloud-deploy /usr/local/bin/cloud-deploy
COPY --from=builder /THIRD_PARTY_NOTICES/cloud-deploy /THIRD_PARTY_NOTICES/cloud-deploy

COPY --from=builder /src/google-cloud-auth /usr/local/bin/google-cloud-auth
COPY --from=builder /THIRD_PARTY_NOTICES/google-cloud-auth /THIRD_PARTY_NOTICES/google-cloud-auth