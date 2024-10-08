# syntax=docker/dockerfile:1
#
# Copyright 2024 Google LLC
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

# hadolint ignore=DL3007
FROM gcr.io/cloud-marketplace/google/debian12:latest

# hadolint ignore=DL3008
RUN apt-get update -y && \
    apt-get -y upgrade && \
    apt-get install -y --no-install-recommends xxd
