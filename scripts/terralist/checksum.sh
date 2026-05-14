#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_unset_proxy
terralist_require_cmd sha256sum

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
: "${PUBLISH_PLATFORMS:?PUBLISH_PLATFORMS is required}"
terralist_init_paths

zip_paths=()
platforms="$(terralist_platforms_from_csv "$PUBLISH_PLATFORMS")"
while IFS= read -r platform; do
  terralist_parse_platform "$platform"
  zip_paths+=("$(terralist_zip_path "$GOOS_VALUE" "$GOARCH_VALUE")")
done <<< "$platforms"
[ "${#zip_paths[@]}" -gt 0 ] || terralist_fail "PUBLISH_PLATFORMS did not include any platforms"

: > "$SUMS_PATH"
while IFS= read -r zip_path; do
  [ -f "$zip_path" ] || terralist_fail "artifact is missing: $zip_path"
  checksum="$(sha256sum "$zip_path" | awk '{print $1}')"
  printf '%s  %s\n' "$checksum" "$(basename "$zip_path")" >> "$SUMS_PATH"
done < <(printf '%s\n' "${zip_paths[@]}" | sort)
printf '[terralist-publish] wrote %s\n' "$SUMS_PATH" >&2
