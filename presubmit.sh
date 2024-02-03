#!/usr/bin/env bash
set -euo pipefail

# Test image builds if the source was modified
changed_image_dirs=$(git show --name-only --pretty="" HEAD | xargs dirname | grep -E "^images/[^/]+$" | uniq)

for dir in $changed_image_dirs
do
  if [ -d "${dir}" ]; then
      make -C "${dir}" build test
  fi
done