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
    i386|i686) echo i386;;
    armv7l|armv7) echo armv7;;
    armv6l|armv6) echo armv6;;
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
# Prefer `sha256sum` (coreutils — preinstalled on Debian/Ubuntu/Alpine/RHEL
# minimal images) over `shasum` (Perl, requires apt-get install perl on slim
# images). Fall back to `shasum -a 256` for macOS, which only ships shasum.
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "$TMP/refuse.$EXT" | cut -d ' ' -f 1)
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "$TMP/refuse.$EXT" | cut -d ' ' -f 1)
else
  echo "neither sha256sum nor shasum found; cannot verify the release" >&2
  exit 1
fi
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

# Write a managed PATH-prepend block to every shell rc that already exists.
# Same delimiters as `refuse install` later uses for its shim block, so the
# two layers don't fight. Only touches existing files — won't create one
# from scratch (matches Homebrew's principle: don't make new dotfiles for
# the user).
patch_rc() {
  rc=$1
  [ -e "$rc" ] || return 0
  marker_begin="# >>> refuse cli (managed) >>>"
  marker_end="# <<< refuse cli (managed) <<<"
  # Skip if already patched (idempotent re-installs).
  if grep -qF "$marker_begin" "$rc" 2>/dev/null; then
    return 0
  fi
  {
    echo ""
    echo "$marker_begin"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
    echo "$marker_end"
  } >> "$rc"
  echo "refuse: added PATH export to $rc"
}

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    for rc in "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile"; do
      patch_rc "$rc"
    done

    echo
    echo "refuse: for THIS shell, run —"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo "(future shells pick it up automatically from the rc edits above)"
    echo
    ;;
esac

echo "refuse: then try \`refuse --version\` and \`refuse init\`"
