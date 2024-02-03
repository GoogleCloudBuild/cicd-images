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

WORKDIR /tools
COPY tools/go.mod tools/go.sum ./
RUN go install github.com/google/go-licenses && \
    go install github.com/anchore/syft/cmd/syft && \
    go install github.com/GoogleCloudPlatform/docker-credential-gcr/v2

RUN go-licenses check --ignore github.com/xi2/xz --ignore modernc.org/mathutil github.com/anchore/syft/cmd/syft && \
    go-licenses save --ignore github.com/xi2/xz --ignore modernc.org/mathutil github.com/anchore/syft/cmd/syft --save_path=/THIRD_PARTY_NOTICES/syft && \
    go-licenses report --ignore github.com/xi2/xz --ignore modernc.org/mathutil github.com/anchore/syft/cmd/syft > /THIRD_PARTY_NOTICES/syft/licenses.csv

RUN go-licenses check github.com/GoogleCloudPlatform/docker-credential-gcr/v2 && \
    go-licenses save github.com/GoogleCloudPlatform/docker-credential-gcr/v2 --save_path=/THIRD_PARTY_NOTICES/docker-credential-gcr && \
    go-licenses report github.com/GoogleCloudPlatform/docker-credential-gcr/v2 > /THIRD_PARTY_NOTICES/docker-credential-gcr/licenses.csv

# hadolint ignore=DL3006
FROM base

ARG USER=nonroot

COPY --from=builder /go/bin/syft /usr/local/bin/syft
COPY --from=builder /THIRD_PARTY_NOTICES/syft /THIRD_PARTY_NOTICES/syft

COPY --from=builder /go/bin/docker-credential-gcr /usr/local/bin/docker-credential-gcr
COPY --from=builder /THIRD_PARTY_NOTICES/docker-credential-gcr /THIRD_PARTY_NOTICES/docker-credential-gcr

USER $USER:$USER

RUN docker-credential-gcr configure-docker -include-artifact-registry

ENTRYPOINT ["syft"]