#!/bin/sh
set -e

(
  cd "$(dirname "$0")"
  go build -o /tmp/dns-server app/*.go
)

exec /tmp/dns-server "$@"
