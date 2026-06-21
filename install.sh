#!/usr/bin/env sh
# install.sh — one-line installer for wpci (jaaacki/woodpecker-cli).
#
#   curl -fsSL https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.sh | sh
#   WPCI_VERSION=v0.0.1 sh install.sh   # pin a release
#
# Downloads the release binary matching this OS/arch, verifies its sha256
# checksum (loudly skipping only when it genuinely cannot), installs `wpci` into
# a user-writable bin directory, and prints a PATH hint when needed. Never asks
# for Woodpecker credentials — run `wpci account add` after install.
set -eu

REPO="${WPCI_REPO:-jaaacki/woodpecker-cli}"
VERSION="${WPCI_VERSION:-latest}"
INSTALL_DIR="${WPCI_INSTALL_DIR:-$HOME/.local/bin}"
BIN_NAME="${WPCI_BIN_NAME:-wpci}"

warn() { echo "wpci: $*" >&2; }
die() { echo "wpci: error: $*" >&2; exit 1; }

uname_s="$(uname -s)"
uname_m="$(uname -m)"

case "$uname_s" in
  Darwin) os="darwin" ;;
  Linux)  os="linux" ;;
  MINGW*|MSYS*|CYGWIN*) die "Windows detected; use install.ps1: irm https://raw.githubusercontent.com/$REPO/main/install.ps1 | iex" ;;
  *) die "unsupported OS: $uname_s (expected darwin/linux)" ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) die "unsupported architecture: $uname_m (expected amd64/arm64)" ;;
esac

if command -v curl >/dev/null 2>&1; then
  fetch="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
  fetch="wget -qO-"
else
  die "curl or wget is required"
fi

api_url="https://api.github.com/repos/$REPO/releases/latest"
if [ "$VERSION" = "latest" ]; then
  tag="$($fetch "$api_url" | sed -n 's/.*\"tag_name\": *\"\([^\"]*\)\".*/\1/p' | head -n 1)"
else
  tag="$VERSION"
fi
[ -n "$tag" ] || die "could not resolve a release tag for '$VERSION'"

asset="wpci-$os-$arch"
base="https://github.com/$REPO/releases/download/$tag"
tmp="${TMPDIR:-/tmp}/wpci-install.$$"
mkdir -p "$tmp"
trap 'rm -rf "$tmp"' EXIT INT TERM

echo "Installing $REPO $tag for $os/$arch"
mkdir -p "$INSTALL_DIR"

$fetch "$base/$asset" > "$tmp/wpci" || die "download failed: $base/$asset"

# --- checksum verification (fail closed on mismatch; warn on inability) ------
if ! { command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1; }; then
  warn "no sha256 tool found; skipping checksum verification"
elif $fetch "$base/checksums.txt" > "$tmp/checksums.txt" 2>/dev/null; then
  expected="$(grep "  $asset\$" "$tmp/checksums.txt" | awk '{print $1}')"
  if [ -z "$expected" ]; then
    warn "no checksum entry for $asset; skipping verification"
  else
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "$tmp/wpci" | awk '{print $1}')"
    else
      actual="$(shasum -a 256 "$tmp/wpci" | awk '{print $1}')"
    fi
    [ "$actual" = "$expected" ] || die "checksum mismatch for $asset (expected $expected, got $actual)"
    echo "checksum verified: $actual"
  fi
else
  warn "could not fetch checksums.txt from $base; skipping verification"
fi

chmod +x "$tmp/wpci"
mv "$tmp/wpci" "$INSTALL_DIR/$BIN_NAME"

# --- PATH check --------------------------------------------------------------
on_path=no
oldifs="$IFS"; IFS=":"
for p in $PATH; do [ "$p" = "$INSTALL_DIR" ] && { on_path=yes; break; }; done
IFS="$oldifs"

echo "Installed: $INSTALL_DIR/$BIN_NAME"
if [ "$on_path" = "no" ]; then
  warn "$INSTALL_DIR is not on your PATH. Add it, then run wpci:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
echo "Next: configure a Woodpecker server account:"
echo "  printf '%s' \"\$WPCI_TOKEN\" | $BIN_NAME account add home --server https://ci.example.com --token-stdin"
echo "  $BIN_NAME home doctor"
