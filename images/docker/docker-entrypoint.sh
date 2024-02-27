#!/bin/sh

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

set -eu

# first arg is `-f` or `--some-option`
if [ "${1#-}" != "$1" ]; then
    set -- docker "$@"
fi

# if our command is a valid Docker subcommand, let's invoke it through Docker instead
# (this allows for "docker run docker ps", etc)
if docker help "$1" > /dev/null 2>&1; then
    set -- docker "$@"
fi

# we only support docker or sh command "default run mode"
if [ "$1" != 'docker' -a "$1" != 'sh' ]; then
cat >&2 <<-'EOM'
    This image only supports "docker" command or a sub-command, \"$1\" is not supported.

    Documentation and example of "docker" CLI commands is available at

    https://docs.docker.com/engine/reference/commandline/cli/
EOM
sleep 5
fi


valid_certs_exist() {
    [ -n "${DOCKER_CERT_PATH:-}" ] \
    && [ -s "$DOCKER_CERT_PATH/ca.pem" ] \
    && [ -s "$DOCKER_CERT_PATH/cert.pem" ] \
    && [ -s "$DOCKER_CERT_PATH/key.pem" ]
}

# if DOCKER_HOST is not defined but default Unix socket exists,
# define DOCKER_HOST explicitly
if [ -z "${DOCKER_HOST:-}" ] && [ -S /var/run/docker.sock ]; then
    export DOCKER_HOST=unix:///var/run/docker.sock
fi

# Configure Artifact Registry auth via GCP's ADC
docker-credential-gcr  configure-docker -include-artifact-registry;

exec "$@"
