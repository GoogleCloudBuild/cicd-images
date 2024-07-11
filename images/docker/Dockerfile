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
ENV USER=root \
    HOME=/root

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG DOCKER_VERSION='26' \
    CONTAINERD_VERSION='1'
# hadolint ignore=DL3008,SC1091
RUN <<EOF
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
        tee /etc/apt/sources.list.d/docker.list > /dev/null
    clean-install \
        docker-ce="5:${DOCKER_VERSION}.*" containerd.io="${CONTAINERD_VERSION}.*" \
        btrfs-progs e2fsprogs xfsprogs iptables uidmap pigz mount util-linux
    dockerd --version
    containerd --version
EOF

ARG DOCKER_CREDENTIAL_GCR_VERSION=2.1.22
ARG DOCKER_CREDENTIAL_GCR_CHECKSUM=443e897dc383d69e55e6dbcb13802f4ec88444848612e83f0381df2ddd721694
RUN <<EOF
    curl -L -o /tmp/docker-credential-gcr_linux_amd64-${DOCKER_CREDENTIAL_GCR_VERSION}.tar.gz \
        "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v${DOCKER_CREDENTIAL_GCR_VERSION}/docker-credential-gcr_linux_amd64-${DOCKER_CREDENTIAL_GCR_VERSION}.tar.gz"
    echo ${DOCKER_CREDENTIAL_GCR_CHECKSUM} /tmp/docker-credential-gcr_linux_amd64-${DOCKER_CREDENTIAL_GCR_VERSION}.tar.gz | sha256sum -c - && \
    tar -xvzf /tmp/docker-credential-gcr_linux_amd64-${DOCKER_CREDENTIAL_GCR_VERSION}.tar.gz && \
    mv docker-credential-gcr /usr/local/bin/docker-credential-gcr && \
    chmod +x /usr/local/bin/docker-credential-gcr && \
    rm /tmp/docker-credential-gcr_linux_amd64-${DOCKER_CREDENTIAL_GCR_VERSION}.tar.gz
EOF

COPY dockerd-entrypoint.sh /usr/local/bin/
COPY docker-entrypoint.sh /usr/local/bin/
COPY dind /usr/local/bin/dind

RUN set -eux; \
    chmod +x /usr/local/bin/dind /usr/local/bin/dockerd-entrypoint.sh

VOLUME /var/lib/docker
EXPOSE 2376/TCP
EXPOSE 2375/TCP

ENTRYPOINT ["dockerd-entrypoint.sh"]
CMD ["dockerd"]
