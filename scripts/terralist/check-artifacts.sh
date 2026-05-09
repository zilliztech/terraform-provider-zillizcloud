#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
: "${PUBLISH_PLATFORMS:?PUBLISH_PLATFORMS is required}"
terralist_init_paths

found_count=0
platforms="$(terralist_platforms_from_csv "$PUBLISH_PLATFORMS")"
while IFS= read -r platform; do
  terralist_parse_platform "$platform"
  zip_path="$(terralist_zip_path "$GOOS_VALUE" "$GOARCH_VALUE")"
  [ -f "$zip_path" ] || terralist_fail "artifact is missing: $zip_path"
  found_count=$((found_count + 1))
done <<< "$platforms"

[ "$found_count" -gt 0 ] || terralist_fail "PUBLISH_PLATFORMS did not include any platforms"
printf '[terralist-publish] checked %s provider artifact(s)\n' "$found_count" >&2
