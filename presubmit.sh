#!/usr/bin/env bash

# Get the changed top-level directories from the current commit.
changed_dirs=$(git show --name-only --pretty="" HEAD | xargs dirname | cut -d "/" -f1 | uniq)

for dir in $changed_dirs
do
    if [ "${dir}" == "." ]; then
      echo "Not running build/test in root directory"
      continue
    fi
    make -C "${dir}" build test
done