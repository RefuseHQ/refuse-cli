#!/usr/bin/env bash
# sign-macos.sh — code-sign a single darwin Go binary with rcodesign.
#
# Called by goreleaser's `builds.hooks.post` once per built binary. We
# can't use goreleaser's native `notarize.macos` block because its Go
# x509 verifier chokes on Apple's critical extension OIDs (e.g.
# 1.2.840.113635.100.6.1.13 — Developer ID Application). rcodesign
# (Rust) handles them correctly.
#
# Usage:  sign-macos.sh BINARY_PATH GOOS
#
# Requires (set as workflow env from secrets):
#   MACOS_SIGN_P12       — base64 of the Developer ID .p12
#   MACOS_SIGN_PASSWORD  — password the .p12 was exported with
#
# Exits 0 on success, 0 (silently) when GOOS != darwin or when the
# signing secret isn't set (forks / local snapshot builds).

set -euo pipefail

BINARY="${1:?missing BINARY arg}"
OS="${2:?missing GOOS arg}"

if [ "$OS" != "darwin" ]; then
    exit 0
fi
if [ -z "${MACOS_SIGN_P12:-}" ]; then
    echo "sign-macos: MACOS_SIGN_P12 unset — leaving $BINARY unsigned" >&2
    exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

p12="$tmp/cert.p12"
printf '%s' "$MACOS_SIGN_P12" | base64 -d > "$p12"

echo "sign-macos: signing $BINARY"
rcodesign sign \
    --p12-file "$p12" \
    --p12-password "$MACOS_SIGN_PASSWORD" \
    --code-signature-flags runtime \
    "$BINARY"

# Best-effort sanity check.
if rcodesign verify "$BINARY" >/dev/null 2>&1; then
    echo "sign-macos: $BINARY verified"
else
    echo "sign-macos: WARNING — $BINARY verify failed (continuing)" >&2
fi
