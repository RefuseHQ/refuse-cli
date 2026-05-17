#!/usr/bin/env sh
# Install the refuse CLI from the latest GitHub release.
#
# Usage:  curl -sSL https://raw.githubusercontent.com/RefuseHQ/refuse-cli/main/scripts/install.sh | sh
#
# Honours $REFUSE_INSTALL_DIR (default ~/.refuse/bin) and $REFUSE_VERSION
# (default latest). Verifies a sha256 checksum from the release.

set -eu

REPO="RefuseHQ/refuse-cli"
INSTALL_DIR="${REFUSE_INSTALL_DIR:-$HOME/.refuse/bin}"
VERSION="${REFUSE_VERSION:-latest}"

uname_os() {
  case "$(uname -s)" in
    Darwin) echo darwin;;
    Linux)  echo linux;;
    MINGW*|MSYS*|CYGWIN*) echo windows;;
    *) echo "unsupported OS: $(uname -s)" >&2; exit 1;;
  esac
}

uname_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo x86_64;;
    arm64|aarch64) echo arm64;;
    *) echo "unsupported arch: $(uname -m)" >&2; exit 1;;
  esac
}

OS=$(uname_os)
ARCH=$(uname_arch)
EXT=tar.gz
[ "$OS" = windows ] && EXT=zip

if [ "$VERSION" = latest ]; then
  VERSION=$(curl -sSfL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep -E '"tag_name":' | head -n1 | sed -E 's/.*"v?([^"]+)".*/\1/')
  [ -z "$VERSION" ] && { echo "could not resolve latest version" >&2; exit 1; }
fi

URL="https://github.com/$REPO/releases/download/v$VERSION/refuse_${OS}_${ARCH}.${EXT}"
SUMS_URL="https://github.com/$REPO/releases/download/v$VERSION/checksums.txt"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "refuse: downloading $URL"
curl -sSfL "$URL" -o "$TMP/refuse.$EXT"

echo "refuse: verifying checksum"
curl -sSfL "$SUMS_URL" -o "$TMP/checksums.txt"
EXPECTED=$(awk -v f="refuse_${OS}_${ARCH}.${EXT}" '$2==f{print $1}' "$TMP/checksums.txt")
[ -z "$EXPECTED" ] && { echo "checksum line for refuse_${OS}_${ARCH}.${EXT} not found" >&2; exit 1; }
ACTUAL=$(shasum -a 256 "$TMP/refuse.$EXT" | cut -d ' ' -f 1)
if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "checksum mismatch: expected $EXPECTED got $ACTUAL" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
if [ "$EXT" = tar.gz ]; then
  tar -xzf "$TMP/refuse.$EXT" -C "$TMP"
else
  unzip -q "$TMP/refuse.$EXT" -d "$TMP"
fi
mv "$TMP/refuse" "$INSTALL_DIR/refuse"
chmod +x "$INSTALL_DIR/refuse"

echo "refuse: installed $INSTALL_DIR/refuse"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo
    echo "refuse: add $INSTALL_DIR to your PATH, e.g."
    echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.zshrc"
    echo
    ;;
esac

echo "refuse: try \`refuse --version\` and then \`refuse init\`"
