#!/usr/bin/env bash
set -euo pipefail

REPO="samuelnp/centinela"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
API_URL="https://api.github.com/repos/$REPO/releases/latest"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1" >&2; exit 1; }
}

need_cmd curl
need_cmd uname
need_cmd awk
need_cmd grep

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac

TAG=$(curl -fsSL "$API_URL" | awk -F '"' '/"tag_name"/{print $4; exit}')
[ -n "$TAG" ] || { echo "could not resolve latest release tag" >&2; exit 1; }

BIN="centinela-${TAG}-${OS}-${ARCH}"
SUMS_URL="https://github.com/$REPO/releases/download/$TAG/SHA256SUMS"
BIN_URL="https://github.com/$REPO/releases/download/$TAG/$BIN"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "$SUMS_URL" -o "$TMP_DIR/SHA256SUMS"
curl -fsSL "$BIN_URL" -o "$TMP_DIR/$BIN"

EXPECTED=$(grep "  $BIN$" "$TMP_DIR/SHA256SUMS" | awk '{print $1}')
[ -n "$EXPECTED" ] || { echo "checksum missing for $BIN" >&2; exit 1; }

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "$TMP_DIR/$BIN" | awk '{print $1}')
else
  need_cmd shasum
  ACTUAL=$(shasum -a 256 "$TMP_DIR/$BIN" | awk '{print $1}')
fi

[ "$EXPECTED" = "$ACTUAL" ] || { echo "checksum verification failed" >&2; exit 1; }

mkdir -p "$INSTALL_DIR"
install -m 0755 "$TMP_DIR/$BIN" "$INSTALL_DIR/centinela"
echo "Installed centinela $TAG to $INSTALL_DIR/centinela"
