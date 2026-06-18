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

# Resolve the latest version. Primary path follows the github.com
# /releases/latest redirect (served from the web/CDN, NOT the 60-req/hr
# unauthenticated api.github.com), which is far less likely to rate-limit or
# 5xx during a traffic spike. Falls back to the JSON API if the redirect
# can't be read, and retries transient failures before giving up.
resolve_latest() {
  v=$(curl -fsSL -I -o /dev/null -w '%{url_effective}' \
        "https://github.com/$REPO/releases/latest" 2>/dev/null \
        | sed -E 's#.*/tag/v?##' | tr -d '[:space:]')
  [ -n "$v" ] && { echo "$v"; return 0; }
  curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
    | grep -E '"tag_name":' | head -n1 | sed -E 's/.*"v?([^"]+)".*/\1/'
}

if [ "$VERSION" = latest ]; then
  i=1
  while [ "$i" -le 3 ]; do
    VERSION=$(resolve_latest)
    [ -n "$VERSION" ] && break
    echo "refuse: version lookup failed (attempt $i/3), retrying..." >&2
    sleep 2
    i=$((i + 1))
  done
  if [ -z "$VERSION" ]; then
    echo "could not resolve latest version" >&2
    echo "GitHub may be having a transient issue — retry shortly, or pin a version:" >&2
    echo "  curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/install.sh | REFUSE_VERSION=1.4.0 sh" >&2
    exit 1
  fi
fi

URL="https://github.com/$REPO/releases/download/v$VERSION/refuse_${OS}_${ARCH}.${EXT}"
SUMS_URL="https://github.com/$REPO/releases/download/v$VERSION/checksums.txt"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "refuse: downloading $URL"
curl -sSfL --retry 3 --retry-delay 1 "$URL" -o "$TMP/refuse.$EXT"

echo "refuse: verifying checksum"
curl -sSfL --retry 3 --retry-delay 1 "$SUMS_URL" -o "$TMP/checksums.txt"
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
  *":$INSTALL_DIR:"*)
    # Already on PATH (reinstall, or a shell that already sourced the rc edits).
    echo
    echo "refuse: ready — run  refuse --version   (then  refuse init  to get started)"
    ;;
  *)
    for rc in "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile"; do
      patch_rc "$rc"
    done

    # NOTE: a piped `curl | sh` runs in a child shell and cannot change the PATH
    # of the terminal you're sitting in. New terminals pick it up from the rc
    # edits above; for the current one, the export below is required. Make it
    # the most prominent, last thing the user sees so it isn't skipped.
    echo
    echo "refuse: installed. New terminals will find it automatically."
    echo "        To use it in THIS terminal right now, run:"
    echo
    echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
    echo
    echo "        Then:  refuse --version   (and  refuse init  to get started)"
    ;;
esac
