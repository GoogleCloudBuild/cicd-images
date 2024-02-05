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
#

set -eu

# "modprobe" without modprobe
# https://twitter.com/lucabruno/status/902934379835662336

# this isn't 100% fool-proof, but it'll have a much higher success rate than simply using the "real" modprobe

# Docker often uses "modprobe -va foo bar baz"
# so we ignore modules that start with "-"
for module; do
	if [ "${module#-}" = "$module" ]; then
		ip link show "$module" || true
		lsmod | grep "$module" || true
	fi
done

# remove /usr/local/... from PATH so we can exec the real modprobe as a last resort
export PATH='/usr/sbin:/usr/bin:/sbin:/bin'
exec modprobe "$@"
