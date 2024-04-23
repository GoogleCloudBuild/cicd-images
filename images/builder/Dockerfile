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

FROM golang:1.21 as go-build

COPY --from=src --chmod=755 / /src
COPY ./ /opt/go-binaries

RUN bash /src/images/installer-scripts/install-go-binaries.sh -m /opt/go-binaries/tools

# Builder image
#hadolint ignore=DL3006
FROM base

USER root
ENV HOME=/builder/home \
    USER=root \
    ENV=/etc/environment \
    BASH_ENV=/etc/environment

COPY --chmod=755 --from=go-build /go/bin/* /usr/local/bin/
COPY --chmod=644 --from=go-build /THIRD_PARTY_NOTICES /THIRD_PARTY_NOTICES

COPY --from=src  --chmod=755 /images/installer-scripts /opt/installer-scripts

RUN bash /opt/installer-scripts/common-apt.sh

# Install build tools
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /opt/installer-scripts/docker.sh
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /opt/installer-scripts/git.sh -git-lfs
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /opt/installer-scripts/setup-nodejs.sh
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /opt/installer-scripts/setup-openjdk.sh
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /opt/installer-scripts/setup-go.sh
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /opt/installer-scripts/setup-python.sh
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /opt/installer-scripts/gcloud.sh -full

# Clean $HOME dir
#hadolint ignore=DL3059
RUN rm -rf /builder/home && \
    mkdir -p /builder/home


VOLUME /builder/home
VOLUME /workspace
WORKDIR $HOME
CMD ["/bin/bash"]
