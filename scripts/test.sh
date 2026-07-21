#!/usr/bin/env bash
# test.sh — run Pr!nt's checks like CI does.
set -euo pipefail

echo "== build =="
go build ./...

echo "== vet =="
go vet ./...

echo "== test =="
go test ./...

echo "== fmt =="
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "unformatted files:"
  echo "$unformatted"
  exit 1
fi

echo "all checks passed"
