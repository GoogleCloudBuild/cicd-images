#!/bin/bash

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

set -o errexit  \
    -o nounset  \
    -o pipefail \
    -o nounset
#    -o xtrace \

 . "$(dirname "$(readlink $0 -f)")"/env.sh

show_help() {
  echo "Usage: $0 [-j OpenJDK-Version] [-m Maven-Version] [-g Gradle-Version]"
}

requested_versions() {
  while getopts ":j:m:g:" opt; do
    case $opt in
      j)
        OPENJDK_VERSION="$OPTARG"
        ;;
      m)
        MAVEN_VERSION="$OPTARG"
        ;;
      g)
        GRADLE_VERSION="$OPTARG"
        ;;
      \?)
        show_help
        err_exit "Invalid option: -$OPTARG"
        ;;
    esac
  done
}

# Set Default versions
OPENJDK_VERSION=$(yq '."language-runtimes".openjdk.default-version' $PACKAGES)
MAVEN_VERSION=$(yq '."build-tools".maven.default-version' $PACKAGES)
GRADLE_VERSION=$(yq '."build-tools".gradle.default-version' $PACKAGES)

# Override versions based on args
requested_versions "$@"
shift $((OPTIND-1)) # Shift arguments past the options
if [[ "$#" != 0 ]]; then
  err_exit "Unexpected args $*"
fi

export OPENJDK_VERSION MAVEN_VERSION GRADLE_VERSION

# Install OpenJDK

MAJOR_VERSION=$(echo $OPENJDK_VERSION | cut -d. -f1)

INSTALL_DIR="/opt/jdk-${MAJOR_VERSION}"
BIN_DIR="${INSTALL_DIR}/bin"

INSTALL="false"
if [[ ! -d $BIN_DIR ]]; then
    is_version_supported $MAJOR_VERSION 'language-runtimes' 'openjdk' \
    || err_exit "OpenJDK version $OPENJDK_VERSION is not supported!"
    INSTALL="true"
fi

RUNTIME_URL="https://dl.google.com/runtimes/ubuntu2204"

CURL_CMD="curl -fsSL -A GCPBuildpacks"
if [[ "$INSTALL" == "true" ]]; then
  VERS_URL="${RUNTIME_URL}/openjdk/version.json"
  FQ_VERSION=$($CURL_CMD $VERS_URL | jq '.[]' | sed -e 's/"//g' | \
                grep "^$MAJOR_VERSION" | sort -V | tail -1 )

  #Normalize FQ_VERSION
  FQ_VERSION=$(echo $FQ_VERSION | tr '+' '_')
  #Install required version
  CURL_CMD="curl -fsSL -A GCPBuildpacks"
  DOWNLOAD_FILE="openjdk-${FQ_VERSION}.tar.gz"
  DOWNLOAD_URL="${RUNTIME_URL}/openjdk/${DOWNLOAD_FILE}"
  $CURL_CMD -o /tmp/${DOWNLOAD_FILE} $DOWNLOAD_URL
  mkdir -p "${INSTALL_DIR}"
  tar --strip-components=1 -zxf /tmp/${DOWNLOAD_FILE} \
      -C "${INSTALL_DIR}"
  rm /tmp/${DOWNLOAD_FILE}
fi

copy_licenses "${INSTALL_DIR}"

#Create links to installed binaries
ln -s ${BIN_DIR}/* -t /usr/local/bin

update_env PATH "${BIN_DIR}:${PATH}"
update_env JAVA_HOME "${INSTALL_DIR}"
update_env JDK_HOME "${INSTALL_DIR}"
update_env JRE_HOME "${INSTALL_DIR}"


# Install maven

INSTALL_DIR="/opt/maven-${MAVEN_VERSION}"
BIN_DIR="${INSTALL_DIR}/bin"
INSTALL="false"
MAVEN_KEY="maven"
if [[ ! -d $BIN_DIR ]]; then
    is_version_supported "$MAVEN_VERSION" "build-tools" "$MAVEN_KEY" \
    || err_exit "Maven version $MAVEN_VERSION is not supported!"
    INSTALL="true"
fi

if [[ "$INSTALL" == "true" ]]; then
  MAVEN_SHA=$( yq --arg key "${MAVEN_KEY}" --arg version "${MAVEN_VERSION}" '."build-tools".[$key]."supported-versions"[] | \
                select ( .version == $version ) | .digest' $PACKAGES )
  MAVEN_BASE_URL=https://downloads.apache.org/maven/maven-3/${MAVEN_VERSION}/binaries
  curl -fsSLO --output-dir /tmp --compressed ${MAVEN_BASE_URL}/apache-maven-${MAVEN_VERSION}-bin.tar.gz
  echo "${MAVEN_SHA} /tmp/apache-maven-${MAVEN_VERSION}-bin.tar.gz" | sha512sum -c -
  curl -fsSLO --output-dir /tmp --compressed ${MAVEN_BASE_URL}/apache-maven-${MAVEN_VERSION}-bin.tar.gz.asc
  GPG_DIGEST=$( yq --arg key "${MAVEN_KEY}" --arg version "${MAVEN_VERSION}" '."build-tools".[$key]."supported-versions"[] | \
                select ( .version == $version ) | .gpg' $PACKAGES )
  gpg --batch --keyserver hkps://keyserver.ubuntu.com --recv-keys $GPG_DIGEST
  gpg --batch --verify  /tmp/apache-maven-${MAVEN_VERSION}-bin.tar.gz.asc \
                        /tmp/apache-maven-${MAVEN_VERSION}-bin.tar.gz
  mkdir "$INSTALL_DIR"
  tar -xzf /tmp/apache-maven-${MAVEN_VERSION}-bin.tar.gz \
      -C ${INSTALL_DIR} --strip-components=1
  rm /tmp/apache-maven-${MAVEN_VERSION}-bin.tar.gz.asc \
    /tmp/apache-maven-${MAVEN_VERSION}-bin.tar.gz
fi

copy_licenses "$INSTALL_DIR"

#Create links to installed binaries
ln -s ${BIN_DIR}/* -t /usr/local/bin

update_env M2_HOME "${INSTALL_DIR}"
update_env M2 "$HOME/.m2"
update_env PATH "${BIN_DIR}:${PATH}"

# Install gradle

INSTALL_DIR="/opt/gradle-${GRADLE_VERSION}"
BIN_DIR="${INSTALL_DIR}/bin"
INSTALL="false"
GRADLE_KEY="gradle"
if [[ ! -d $BIN_DIR ]]; then
    INSTALL="true"
    is_version_supported "$GRADLE_VERSION" "build-tools" "$GRADLE_KEY" \
    || err_exit "Gradle version $GRADLE_VERSION is not supported!"
fi

if [[ "$INSTALL" == "true" ]]; then

  GRADLE_SHA=$( yq --arg key "${GRADLE_KEY}" --arg version "${GRADLE_VERSION}" '."build-tools".[$key]."supported-versions"[] | \
                select ( .version == $version ) | .digest' $PACKAGES )
  GRADLE_BASE_URL=https://services.gradle.org/distributions

  curl -fsSLO --output-dir /tmp --compressed ${GRADLE_BASE_URL}/gradle-${GRADLE_VERSION}-bin.zip
  echo "${GRADLE_SHA} /tmp/gradle-${GRADLE_VERSION}-bin.zip" | sha256sum -c -
  unzip -qd /opt /tmp/gradle-${GRADLE_VERSION}-bin.zip
  rm /tmp/gradle-${GRADLE_VERSION}-bin.zip

fi

copy_licenses "$INSTALL_DIR"

#Create links to installed binaries
ln -s ${BIN_DIR}/* -t /usr/local/bin

update_env GRADLE_HOME "$INSTALL_DIR"
update_env GRADLE_USER_HOME "$HOME/.gradle"
update_env PATH "${BIN_DIR}:${PATH}"
