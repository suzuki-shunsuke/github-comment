#!/usr/bin/env bash

set -eu

ee() {
  echo "+ $*"
  eval "$@"
}

cd "$(dirname "$0")/.."
pwd

if [ $# -eq 0 ]; then
  target="$(find pkg -type d | fzf)"
  if [ "$target" = "" ]; then
    exit 0
  fi
elif [ $# -eq 1 ]; then
  target=$1
else
  echo "too many arguments are given: $*" >&2
  exit 1
fi

if [ ! -d "$target" ]; then
  echo "$target is not found" >&2
  exit 1
fi

ee mkdir -p .coverage/"$target"
ee go test "./$target" -coverprofile=".coverage/$target/coverage.txt" -covermode=atomic
ee go tool cover -html=".coverage/$target/coverage.txt"
