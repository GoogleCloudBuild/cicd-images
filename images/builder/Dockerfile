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

#hadolint ignore=DL3006
FROM base

USER root
ENV USER=root \
    ENV=/etc/environment \
    HOME=/root

# Convenience script for building on this base image.
COPY installer-scripts/ /opt/installer-scripts/

RUN <<EOF
    clean-install \
        software-properties-common \
        yq
EOF

#hadolint ignore=DL3059
SHELL ["/bin/bash", "-c"]
RUN <<EOF
    /opt/installer-scripts/common-apt.sh
    /opt/installer-scripts/docker.sh
    /opt/installer-scripts/setup-nodejs.sh
EOF

# TODO: Install additional language runtimes when needed
# RUN bash /opt/installer-scripts/setup-go.sh
# RUN bash /opt/installer-scripts/setup-openjdk.sh
# RUN bash /opt/installer-scripts/setup-python.sh

VOLUME /workspace
WORKDIR /workspace
CMD ["/bin/bash"]
