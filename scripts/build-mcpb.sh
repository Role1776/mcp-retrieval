#!/usr/bin/env bash
# Packs each binary GoReleaser produced into an MCP Bundle (.mcpb).
#
# An .mcpb is a zip holding a manifest.json plus the server payload; the
# "binary" server type means the bundle carries a pre-compiled executable and
# needs no runtime on the user's machine. One bundle per OS/arch, because a
# native binary is not portable across them.
#
# Usage: scripts/build-mcpb.sh <version>
# Reads:  dist/artifacts.json (written by `goreleaser release`)
# Writes: dist/mcpb/*.mcpb
set -euo pipefail

version="${1:?usage: build-mcpb.sh <version>}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="$root/dist"
out="$dist/mcpb"

[ -f "$dist/artifacts.json" ] || { echo "no $dist/artifacts.json — run goreleaser first" >&2; exit 1; }

rm -rf "$out"
mkdir -p "$out"

# goos -> the platform identifier the MCPB manifest schema expects
mcpb_platform() {
  case "$1" in
    linux)   echo linux ;;
    darwin)  echo darwin ;;
    windows) echo win32 ;;
    *)       echo "unsupported goos: $1" >&2; return 1 ;;
  esac
}

count=0
while IFS=$'\t' read -r path goos goarch; do
  binary="$(basename "$path")"
  platform="$(mcpb_platform "$goos")"
  stage="$(mktemp -d)"

  mkdir -p "$stage/server"
  cp "$root/$path" "$stage/server/$binary"
  chmod +x "$stage/server/$binary"

  # ${__dirname} is expanded by the host app to the unpacked bundle directory.
  jq -n \
    --arg version "$version" \
    --arg binary "$binary" \
    --arg platform "$platform" \
    '{
      manifest_version: "0.3",
      name: "mcp-retrieval",
      display_name: "MCP Retrieval",
      version: $version,
      description: "Web and image search plus page scraping to Markdown, with no API keys.",
      author: { name: "Role1776", url: "https://github.com/Role1776" },
      homepage: "https://github.com/Role1776/mcp-retrieval",
      repository: { type: "git", url: "https://github.com/Role1776/mcp-retrieval" },
      license: "MIT",
      server: {
        type: "binary",
        entry_point: ("server/" + $binary),
        mcp_config: {
          command: ("${__dirname}/server/" + $binary),
          args: [],
          env: {}
        }
      },
      compatibility: { platforms: [$platform] }
    }' > "$stage/manifest.json"

  bundle="$out/mcp-retrieval_${version}_${goos}_${goarch}.mcpb"
  (cd "$stage" && zip -q -r -X "$bundle" manifest.json server)
  rm -rf "$stage"

  echo "packed $(basename "$bundle")"
  count=$((count + 1))
done < <(jq -r '.[] | select(.type == "Binary") | [.path, .goos, .goarch] | @tsv' "$dist/artifacts.json")

[ "$count" -gt 0 ] || { echo "no binaries found in artifacts.json" >&2; exit 1; }
