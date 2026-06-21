#!/usr/bin/env sh
set -eu

REPO="${WPCI_REPO:-jaaacki/woodpecker-cli}"
VERSION="${WPCI_VERSION:-latest}"
INSTALL_DIR="${WPCI_INSTALL_DIR:-$HOME/.local/bin}"
BIN_NAME="${WPCI_BIN_NAME:-wpci}"

uname_s="$(uname -s)"
uname_m="$(uname -m)"

case "$uname_s" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *) echo "unsupported OS: $uname_s" >&2; exit 1 ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported architecture: $uname_m" >&2; exit 1 ;;
esac

if command -v curl >/dev/null 2>&1; then
  fetch="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
  fetch="wget -qO-"
else
  echo "curl or wget is required" >&2
  exit 1
fi

api_url="https://api.github.com/repos/$REPO/releases/latest"
if [ "$VERSION" = "latest" ]; then
  tag="$($fetch "$api_url" | sed -n 's/.*\"tag_name\": *\"\([^\"]*\)\".*/\1/p' | head -n 1)"
else
  tag="$VERSION"
fi

if [ -z "${tag:-}" ]; then
  echo "could not determine release version for $REPO" >&2
  exit 1
fi

asset="wpci-$os-$arch"
base="https://github.com/$REPO/releases/download/$tag"
tmp="${TMPDIR:-/tmp}/wpci-install.$$"
mkdir -p "$tmp"
trap 'rm -rf "$tmp"' EXIT INT TERM

echo "Installing $REPO $tag for $os/$arch"
mkdir -p "$INSTALL_DIR"

$fetch "$base/$asset" > "$tmp/wpci"

if $fetch "$base/checksums.txt" > "$tmp/checksums.txt" 2>/dev/null; then
  expected="$(grep "  $asset\$" "$tmp/checksums.txt" | awk '{print $1}')"
  if [ -n "$expected" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "$tmp/wpci" | awk '{print $1}')"
    else
      actual="$(shasum -a 256 "$tmp/wpci" | awk '{print $1}')"
    fi
    [ "$actual" = "$expected" ] || { echo "checksum mismatch" >&2; exit 1; }
  fi
fi

chmod +x "$tmp/wpci"
mv "$tmp/wpci" "$INSTALL_DIR/$BIN_NAME"

echo "Installed: $INSTALL_DIR/$BIN_NAME"
echo "Next:"
echo "  $BIN_NAME account add home --server https://ci.example.com"
echo "  $BIN_NAME account token set home"
echo "  $BIN_NAME home doctor"

