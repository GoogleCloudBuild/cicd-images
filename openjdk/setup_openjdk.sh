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

MAJOR_VERSION=$(echo $GOOGLE_RUNTIME_VERSION | cut -d. -f1)

BIN_DIR="/opt/jdk-${MAJOR_VERSION}/bin"

if [ ! -d $BIN_DIR ]; then
    echo "openjdk version $GOOGLE_RUNTIME_VERSION not installed" 1>&2
    exit 1
fi

export JAVA_HOME="/opt/jdk-${MAJOR_VERSION}"
export PATH="$BIN_DIR:$PATH"