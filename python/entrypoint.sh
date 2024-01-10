#!/usr/bin/env sh

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