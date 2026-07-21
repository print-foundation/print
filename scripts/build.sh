#!/usr/bin/env bash
# build.sh — cross-compile Pr!nt release artifacts into dist/.
set -euo pipefail

VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo dev)}"
OUT="${OUT:-dist}"
LDFLAGS="-X main.version=${VERSION} -trimpath -buildvcs=false"

mkdir -p "$OUT"

targets=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "freebsd/amd64"
  "windows/amd64"
)

for t in "${targets[@]}"; do
  os="${t%/*}"
  arch="${t#*/}"
  name="print-${os}-${arch}"
  [ "$os" = "windows" ] && name="${name}.exe"
  echo "building ${os}/${arch} -> ${OUT}/${name}"
  GOOS="$os" GOARCH="$arch" go build -ldflags "$LDFLAGS" -o "${OUT}/${name}" ./cmd/print
done

echo "done. artifacts in ${OUT}/"
