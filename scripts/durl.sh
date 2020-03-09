#!/usr/bin/env sh

cd "$(dirname "$0")/.." || exit 1

find . \
  -type d -name .git -prune -o \
  -type f -print |
  durl check || exit 1
