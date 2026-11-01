#!/bin/sh
# Cross-compiles wavemark for every supported platform and writes the
# release files (archives and checksums.txt) to the output folder.
#
#   scripts/release.sh v0.1.0 [out]
set -eu

version="${1:?usage: scripts/release.sh <version> [out]}"
out="${2:-release}"
targets="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

rm -rf "$out"
mkdir -p "$out"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

for target in $targets; do
  os="${target%/*}"
  arch="${target#*/}"
  bin=wavemark
  [ "$os" = windows ] && bin=wavemark.exe
  dir="$work/${os}_${arch}"
  mkdir -p "$dir"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath \
    -ldflags "-s -w -X main.version=$version" -o "$dir/$bin" ./cmd/wavemark
  cp LICENSE README.md "$dir/" 2>/dev/null || true
  if [ "$os" = windows ]; then
    (cd "$dir" && zip -q -r "../wavemark_${os}_${arch}.zip" .)
    mv "$work/wavemark_${os}_${arch}.zip" "$out/"
  else
    tar -czf "$out/wavemark_${os}_${arch}.tar.gz" -C "$dir" .
  fi
  echo "built ${os}/${arch}"
done

(cd "$out" && sha256sum wavemark_* > checksums.txt)
echo "release files in $out:"
ls -l "$out"
