#!/usr/bin/env sh

MAJOR_VERSION=$(echo $GOOGLE_NODEJS_VERSION | cut -d. -f1)

BIN_DIR="/opt/node${MAJOR_VERSION}/bin"

if [ ! -d $BIN_DIR ]; then
    echo "nodejs version $GOOGLE_NODEJS_VERSION not installed" 1>&2
    exit 1
fi

export PATH="$BIN_DIR:$PATH"

exec "$@"