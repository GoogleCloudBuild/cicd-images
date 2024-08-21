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

FROM google-go.pkg.dev/golang:1.22.4 as builder

WORKDIR /src
COPY --from=src . .

RUN go mod download && \
    go test ./cmd/cloud-run/... -timeout 30s && \
    go build ./cmd/cloud-run

# hadolint ignore=DL3006
FROM base

COPY --from=builder /src/cloud-run /usr/local/bin/cloud-run
COPY --from=src --chmod=755 ./THIRD_PARTY_NOTICES.tar.gz /home