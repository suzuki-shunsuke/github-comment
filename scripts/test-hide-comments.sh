#!/usr/bin/env bash

set -eux

cd "$(dirname "$0")/.."

go run ./cmd/github-comment --log-level debug post -k hello
