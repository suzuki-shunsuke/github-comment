#!/usr/bin/env sh

ee() {
  echo "+ $*"
  eval "$@"
}

cd "$(dirname "$0")/.." || exit 1
if [ ! -f .git/hooks/pre-commit ]; then
  ee ln -s ../../githooks/pre-commit.sh .git/hooks/pre-commit || exit 1
fi
ee chmod a+x githooks/*
