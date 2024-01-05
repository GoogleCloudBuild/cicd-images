#!/usr/bin/env sh

MINOR_VERSION=$(echo $GOOGLE_PYTHON_VERSION | cut -d. -f1-2)

BIN_DIR="/opt/python${MINOR_VERSION}/bin"

if [ $MINOR_VERSION != "3.10" ] && [ ! -d $BIN_DIR ]; then
    echo "python version $GOOGLE_PYTHON_VERSION not installed" 1>&2
    exit 1
fi

if [ $MINOR_VERSION != "3.10" ]; then
    export PATH="/opt/python${MINOR_VERSION}/bin":$PATH
fi

exec "$@"