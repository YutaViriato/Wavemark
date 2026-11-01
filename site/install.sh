#!/bin/sh
# Installs the wavemark binary.
#
#   curl -fsSL https://wavemark.viriato.site/install.sh | sh
#
# Environment:
#   WAVEMARK_VERSION       version to install, e.g. v0.1.0 (default: latest)
#   WAVEMARK_INSTALL_DIR   where to put the binary (default: $HOME/.local/bin)
#   WAVEMARK_DOWNLOAD_URL  base URL of the release files (default: GitHub Releases)
set -eu

REPO="YutaViriato/Wavemark"
VERSION="${WAVEMARK_VERSION:-latest}"
INSTALL_DIR="${WAVEMARK_INSTALL_DIR:-$HOME/.local/bin}"

say() { printf 'wavemark: %s\n' "$*" >&2; }
fail() { say "error: $*"; exit 1; }

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  MINGW* | MSYS* | CYGWIN*) fail "on Windows, run in PowerShell: irm https://wavemark.viriato.site/install.ps1 | iex" ;;
  *) fail "unsupported system: $(uname -s)" ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) fail "unsupported processor: $(uname -m)" ;;
esac

if [ -n "${WAVEMARK_DOWNLOAD_URL:-}" ]; then
  base="${WAVEMARK_DOWNLOAD_URL%/}"
elif [ "$VERSION" = latest ]; then
  base="https://github.com/$REPO/releases/latest/download"
else
  base="https://github.com/$REPO/releases/download/$VERSION"
fi
asset="wavemark_${os}_${arch}.tar.gz"

download() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$1" -O "$2"
  else
    fail "curl or wget is required"
  fi
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

say "downloading $asset ($VERSION)"
download "$base/$asset" "$tmp/$asset" || fail "could not download $base/$asset"
download "$base/checksums.txt" "$tmp/checksums.txt" || fail "could not download $base/checksums.txt"

expected="$(grep " $asset\$" "$tmp/checksums.txt" | cut -d' ' -f1)"
[ -n "$expected" ] || fail "$asset is not listed in checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp/$asset" | cut -d' ' -f1)"
elif command -v shasum >/dev/null 2>&1; then
  actual="$(shasum -a 256 "$tmp/$asset" | cut -d' ' -f1)"
else
  fail "sha256sum or shasum is required to verify the download"
fi
[ "$expected" = "$actual" ] || fail "checksum mismatch for $asset"

tar -xzf "$tmp/$asset" -C "$tmp"
mkdir -p "$INSTALL_DIR"
mv "$tmp/wavemark" "$INSTALL_DIR/wavemark"
chmod 755 "$INSTALL_DIR/wavemark"

say "installed $("$INSTALL_DIR/wavemark" version) to $INSTALL_DIR/wavemark"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) say "add it to your PATH:  export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac
say "next: wavemark init my-site && wavemark dev my-site"
