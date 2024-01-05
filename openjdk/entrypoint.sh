#!/usr/bin/env sh

MAJOR_VERSION=$(echo $GOOGLE_RUNTIME_VERSION | cut -d. -f1)

BIN_DIR="/opt/jdk-${MAJOR_VERSION}/bin"

if [ ! -d $BIN_DIR ]; then
    echo "openjdk version $GOOGLE_RUNTIME_VERSION not installed" 1>&2
    exit 1
fi

export JAVA_HOME="/opt/jdk-${MAJOR_VERSION}"
export PATH="$BIN_DIR:$PATH"
export MAVEN_HOME="/opt/maven-${MAVEN_VERSION}"
export PATH="${MAVEN_HOME}/bin:/opt/gradle-${GRADLE_VERSION}/bin":$PATH

exec "$@"