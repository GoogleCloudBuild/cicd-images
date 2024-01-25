#!/usr/bin/env bash

# Get the changed top-level directories from the current commit.
changed_dirs=$(git show --name-only --pretty="" HEAD | cut -d "/" -f1 | xargs dirname | uniq)

for dir in $changed_dirs
do
    if [ "${dir}" == "." ]; then
      echo "Not running build/test in root directory"
      continue
    fi
    make -C "${dir}" build test
done