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

# Apple's notary requires the full cert chain (leaf → "Developer ID
# Certification Authority" intermediate → Apple Root CA) to be visible
# inside the signature. A typical Keychain Access .p12 export only
# packages the leaf cert + private key, so we fetch the intermediates
# from Apple's stable URLs and pass them to rcodesign as extra certs to
# embed.
#
# Apple issues Developer ID certs from one of two intermediates:
#   - G1 (legacy):  DeveloperIDCA.cer    — older certs
#   - G2 (current): DeveloperIDG2CA.cer  — most certs issued from ~2018
# We fetch both and embed both; rcodesign accepts --certificate-der-file
# multiple times. Apple's notary picks the matching chain.
intermediate_g1="$tmp/DeveloperIDCA.cer"
intermediate_g2="$tmp/DeveloperIDG2CA.cer"
echo "sign-macos: fetching Developer ID Certification Authority intermediates (G1 + G2)"
curl -fsSL --retry 3 https://www.apple.com/certificateauthority/DeveloperIDCA.cer    -o "$intermediate_g1"
curl -fsSL --retry 3 https://www.apple.com/certificateauthority/DeveloperIDG2CA.cer  -o "$intermediate_g2"

echo "sign-macos: signing $BINARY"
rcodesign sign \
    --p12-file "$p12" \
    --p12-password "$MACOS_SIGN_PASSWORD" \
    --certificate-der-file "$intermediate_g1" \
    --certificate-der-file "$intermediate_g2" \
    --code-signature-flags runtime \
    "$BINARY"

# Best-effort sanity check.
if rcodesign verify "$BINARY" >/dev/null 2>&1; then
    echo "sign-macos: $BINARY verified"
else
    echo "sign-macos: WARNING — $BINARY verify failed (continuing)" >&2
fi
