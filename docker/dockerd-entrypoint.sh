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

set -eu

# Validate TLS key cert pair exists and cert is signed by the given ca.pem
valid_certs_exist() {
    [ -n "${DOCKER_CERT_PATH:-}" ] \
    && [ -s "$DOCKER_CERT_PATH/ca.pem" ] \
    && [ -s "$DOCKER_CERT_PATH/cert.pem" ] \
    && [ -s "$DOCKER_CERT_PATH/key.pem" ] \
    && openssl verify -CAfile "$DOCKER_CERT_PATH/ca.pem" "$DOCKER_CERT_PATH/cert.pem" >/dev/null
}

# Set default options when user runs with no command or `dockerd`
# We will incorporate additional options set by user to the default ones
if [ "$#" -eq 0 ] || [ "${1#-}" != "$1" ]; then
    # Set default command and add user provided options at the end
    set -- dockerd "$@"
fi

if [ "$1" = 'dockerd' ]; then
    # set "dockerSocket" to the default *unix socket*
    dockerSocket='unix:///var/run/docker.sock'
    HOSTS="--host=${dockerSocket}"
    # When user specifies a socket, we use user provided value
    # For other type of hosts we append to default socket
    case "${DOCKER_HOST:-}" in
        "")
            echo "DOCKER_HOST is not defined, using default ${HOSTS}"
            ;;
        unix://*)
            HOSTS="--host=${DOCKER_HOST}"
            ;;
        *)
            HOSTS="$HOSTS --host=${DOCKER_HOST}"
            ;;
    esac


    # Enable TCP only when valid certs exists in user provided DOCKER_CERT_PATH
    # We will not generate self-signed certs
    #
    TLS_OPTIONS=''
    if [ -n "${DOCKER_TLS_CERTDIR:-}" ] && [ valid_certs_exist ]; then
        HOSTS="$HOSTS --host=tcp://0.0.0.0:2376"
        TLS_OPTIONS=" --tlsverify"
        TLS_OPTIONS="${TLS_OPTIONS} --tlscacert ${DOCKER_CERT_PATH}/ca.pem"
        TLS_OPTIONS="${TLS_OPTIONS} --tlscert   ${DOCKER_CERT_PATH}/cert.pem"
        TLS_OPTIONS="${TLS_OPTIONS} --tlskey    ${DOCKER_CERT_PATH}/key.pem"
    fi
    # Add host options
    set "$@" $HOSTS $TLS_OPTIONS

    # explicitly remove Docker's default PID file to ensure that it can start properly if it was stopped uncleanly (and thus didn't clean up the PID file)
    find /run /var/run -iname 'docker*.pid' -delete || :

    if ! iptables -nL > /dev/null 2>&1; then
        # if iptables fails to run, chances are high the necessary kernel modules aren't loaded (perhaps the host is using nftables with the translating "iptables" wrappers, for example)
        # https://github.com/docker-library/docker/issues/350
        # https://github.com/moby/moby/issues/26824
        modprobe ip_tables || :
    fi
    # use dind wrapper
    set -- '/usr/local/bin/dind' "$@"
else
    # User provided a command that is not dockerd, we assume intention is to run docker-cli,
    # so we delegate to cli entrypoint

    set -- docker-entrypoint.sh "$@"
fi

exec "$@"
