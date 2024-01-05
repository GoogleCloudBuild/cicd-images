#!/usr/bin/env sh

MINOR_VERSION=$(echo $GOOGLE_GO_VERSION | cut -d. -f1-2)

BIN_DIR="/usr/lib/go-$MINOR_VERSION/bin"

if [ ! -d $BIN_DIR ]; then
    echo "go version $GOOGLE_GO_VERSION not installed" 1>&2
    exit 1
fi

export PATH="$BIN_DIR:$PATH"

exec "$@"