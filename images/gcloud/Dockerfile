# syntax=docker/dockerfile:1
#
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

# hadolint ignore=DL3006
FROM base

USER root

COPY --from=src --chmod=755 / /src

RUN apt-get update -y \
    && clean-install apt-transport-https python3

# Install gcloud CLI and prune unnecessary files
#hadolint ignore=DL3059
RUN --mount=type=bind,target=/src,from=src \
    bash /src/gcloud.sh -full

ENV USER="nonroot"
USER $USER

WORKDIR /home/$USER

ENTRYPOINT ["gcloud"]
