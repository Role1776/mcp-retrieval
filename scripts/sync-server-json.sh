#!/usr/bin/env bash
# Makes the git tag the single source of truth for the registry entry.
#
# server.json is committed with the OCI package only. This script pins the
# version into `.version` and into the image identifier, then appends one mcpb
# package per bundle in dist/mcpb — reusing the OCI package's environment
# variable list so it is declared in exactly one place.
#
# Running it twice is safe: every mcpb package is rebuilt from dist/mcpb rather
# than appended to whatever was already there.
#
# Usage: scripts/sync-server-json.sh <version> [--no-bundles]
#
#   --no-bundles  only pin the version and the image tag. Without it a missing
#                 or empty dist/mcpb is an error, so a release can never
#                 silently ship with the Docker package alone.
set -euo pipefail

version="${1:?usage: sync-server-json.sh <version> [--no-bundles]}"
want_bundles=1
[ "${2:-}" = "--no-bundles" ] && want_bundles=0

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
file="$root/server.json"
bundles="$root/dist/mcpb"
release_base="https://github.com/Role1776/mcp-retrieval/releases/download/v${version}"

# The image path must be lowercase (ghcr.io rejects uppercase repository
# names); the server *name* and the Dockerfile label keep the account's real
# casing. The registry never compares the two, so this is not a mismatch.
image="ghcr.io/role1776/mcp-retrieval:${version}"

# Collect the inputs before touching server.json, so a missing bundle aborts
# with the file untouched rather than half-rewritten.
found=()
if [ "$want_bundles" -eq 1 ]; then
  shopt -s nullglob
  found=("$bundles"/*.mcpb)
  shopt -u nullglob

  if [ "${#found[@]}" -eq 0 ]; then
    echo "no .mcpb bundles in $bundles" >&2
    echo "run scripts/build-mcpb.sh <version> first, or pass --no-bundles on purpose" >&2
    exit 1
  fi
fi

tmp="$(mktemp)"
jq --arg v "$version" --arg image "$image" '
  .version = $v
  | .packages = [.packages[] | select(.registryType == "oci") | .identifier = $image]
' "$file" > "$tmp"
mv "$tmp" "$file"

if [ "$want_bundles" -eq 0 ]; then
  echo "server.json synced to $version (version and image only, no mcpb packages)"
  exit 0
fi

for bundle in "${found[@]}"; do
    name="$(basename "$bundle")"
    # `shasum -a 256` on macOS, `sha256sum` on Linux runners.
    sha="$(if command -v sha256sum >/dev/null; then sha256sum "$bundle"; else shasum -a 256 "$bundle"; fi | cut -d' ' -f1)"

    tmp="$(mktemp)"
    jq --arg v "$version" --arg url "$release_base/$name" --arg sha "$sha" '
      .packages += [{
        registryType: "mcpb",
        identifier: $url,
        version: $v,
        fileSha256: $sha,
        transport: { type: "stdio" },
        # The env list lives on the OCI package; mirror it so both install
        # paths advertise the same configuration.
        environmentVariables: (.packages[] | select(.registryType == "oci") | .environmentVariables)
      }]
    ' "$file" > "$tmp"
    mv "$tmp" "$file"
    echo "added mcpb package $name ($sha)"
done

jq -e '.version' "$file" > /dev/null
echo "server.json synced to $version: 1 oci + $(jq '[.packages[] | select(.registryType == "mcpb")] | length' "$file") mcpb packages"
