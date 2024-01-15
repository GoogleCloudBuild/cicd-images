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

# Adopted from https://github.com/kubernetes/release/blob/master/images/build/debian-base/bullseye/Dockerfile.build
# hadolint ignore=DL3007
FROM marketplace.gcr.io/google/ubuntu2204:latest

ENV DEBIAN_FRONTEND=noninteractive

# Smaller package install size.
COPY excludes /etc/dpkg/dpkg.cfg.d/excludes

# Convenience script for building on this base image.
COPY clean-install /usr/local/bin/clean-install

# Keep base image up-to-date
RUN apt-get update && \
    apt-get upgrade -y && \
    rm -rf /var/lib/apt/lists/*

# No-op stubs replace some unnecessary binaries that may be depended on in the install process (in
# particular we don't run an init process).
WORKDIR /usr/local/bin
RUN touch noop && \
    chmod 555 noop && \
    ln -s noop runlevel && \
    ln -s noop invoke-rc.d && \
    ln -s noop update-rc.d

WORKDIR /

# Hold required packages to avoid breaking the installation of packages
RUN apt-mark hold apt gpgv adduser passwd libsemanage-common libsemanage2 libcap2 && \
    apt-get purge -y --allow-remove-essential \
    e2fsprogs \
    libmount1 \
    libsmartcols1 \
    libblkid1 \
    libss2 \
    ncurses-base \
    ncurses-bin \
    util-linux

# Install common packages needed by CICD processes.
RUN clean-install \
    bash \
    ca-certificates \
    curl \
    wget \
    gnupg \
    unzip

# Cleanup cached and unnecessary files.
# Keep copyright notices and licenses available under /usr/share/copyrights.tar.gz.
RUN apt-get autoremove -y && \
    apt-get clean -y && \
    tar -czf /usr/share/copyrights.tar.gz /usr/share/common-licenses /usr/share/doc/*/copyright && \
    rm -rf \
    /usr/share/doc \
    /usr/share/man \
    /usr/share/info \
    /usr/share/locale \
    /var/lib/apt/lists/* \
    /var/log/* \
    /var/cache/debconf/* \
    /usr/share/common-licenses* \
    /usr/share/bash-completion \
    ~/.bashrc \
    ~/.profile \
    /etc/systemd \
    /lib/lsb \
    /lib/udev \
    /usr/lib/x86_64-linux-gnu/gconv/IBM* \
    /usr/lib/x86_64-linux-gnu/gconv/EBC* && \
    mkdir -p /usr/share/man/man1 /usr/share/man/man2 \
    /usr/share/man/man3 /usr/share/man/man4 \
    /usr/share/man/man5 /usr/share/man/man6 \
    /usr/share/man/man7 /usr/share/man/man8

# We default to use the same uid as non-root distroless images for consistency.
# See https://github.com/GoogleContainerTools/distroless.
ARG USER
ARG USERID

ENV USER=${USER:-nonroot}
ENV USERID=${USERID:-65532}
ENV GROUPID=${USERID}

RUN addgroup --gid $GROUPID $USER && \
    adduser --disabled-password --gecos "" --uid $USERID --gid $GROUPID --home /home/$USER --shell /bin/sh $USER

USER $USERID:$GROUPID
WORKDIR /home/$USER