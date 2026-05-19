#!/usr/bin/env bash
# notarize-macos.sh — submit darwin binaries from a goreleaser dist/ dir
# to Apple's notary service via rcodesign.
#
# Called as a post-goreleaser step in release.yaml. For each
# refuse_darwin_*.tar.gz in dist/, extract the binary, zip it on its
# own (Apple's notary service accepts .zip, not .tar.gz), submit, and
# wait for approval. We don't staple the ticket — stapling only works
# on bundles/.pkg/.dmg, not plain Mach-O binaries — but the
# notarization is recorded server-side and Gatekeeper online-queries
# Apple at first launch.
#
# Requires (set as workflow env from secrets):
#   MACOS_NOTARY_KEY        — base64 of the App Store Connect .p8 key
#   MACOS_NOTARY_KEY_ID     — the 10-char Key ID
#   MACOS_NOTARY_ISSUER_ID  — the UUID Issuer ID
#
# Exits 0 (silently) when secrets aren't set. Fails loudly on Apple's
# rejection so the release workflow surfaces the problem.

set -euo pipefail

DIST="${1:-dist}"

if [ -z "${MACOS_NOTARY_KEY:-}" ]; then
    echo "notarize-macos: MACOS_NOTARY_KEY unset — skipping" >&2
    exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# Build an rcodesign-compatible API key JSON from the three secrets.
p8="$tmp/AuthKey.p8"
printf '%s' "$MACOS_NOTARY_KEY" | base64 -d > "$p8"

api_key_json="$tmp/api-key.json"
rcodesign encode-app-store-connect-api-key \
    -o "$api_key_json" \
    "$MACOS_NOTARY_ISSUER_ID" \
    "$MACOS_NOTARY_KEY_ID" \
    "$p8"

shopt -s nullglob
archives=( "$DIST"/refuse_darwin_*.tar.gz )
if [ ${#archives[@]} -eq 0 ]; then
    echo "notarize-macos: no darwin archives in $DIST — nothing to do" >&2
    exit 0
fi

for archive in "${archives[@]}"; do
    name=$(basename "$archive" .tar.gz)
    work="$tmp/$name"
    mkdir -p "$work"

    echo "notarize-macos: preparing $name"
    tar -xzf "$archive" -C "$work"

    zip="$tmp/$name.zip"
    ( cd "$work" && zip -q "$zip" refuse )

    echo "notarize-macos: submitting $name.zip to Apple notary"
    rcodesign notary-submit \
        --api-key-file "$api_key_json" \
        --wait \
        "$zip"
done

echo "notarize-macos: all darwin archives notarized"
