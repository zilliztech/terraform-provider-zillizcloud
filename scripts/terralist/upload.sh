#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_unset_proxy
terralist_require_cmd curl
terralist_require_cmd jq
terralist_require_cmd kubectl

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
: "${PUBLISH_PLATFORMS:?PUBLISH_PLATFORMS is required}"
: "${TERRALIST_NAMESPACE:?TERRALIST_NAMESPACE is required}"
: "${TERRALIST_URL:?TERRALIST_URL is required}"
: "${REMOTE_DIR:?REMOTE_DIR is required}"
terralist_init_paths
[ -f "$SUMS_PATH" ] || terralist_fail "checksum file is missing: $SUMS_PATH"

protocols_json="$(jq -Rn 'input | split(",") | map(gsub("^\\s+|\\s+$"; "")) | map(select(length > 0))' <<< "${TERRAFORM_PROVIDER_PROTOCOLS:-6.0}")"
platforms_file="$(mktemp)"
body_file="$(mktemp)"
trap 'rm -f "$platforms_file" "$body_file"' EXIT
: > "$platforms_file"

platforms="$(terralist_platforms_from_csv "$PUBLISH_PLATFORMS")"
while IFS= read -r platform; do
  terralist_parse_platform "$platform"
  zip_name="$(terralist_zip_name "$GOOS_VALUE" "$GOARCH_VALUE")"
  checksum="$(awk -v name="$zip_name" '$2 == name {print $1}' "$SUMS_PATH")"
  [ -n "$checksum" ] || terralist_fail "missing checksum for expected artifact $zip_name"
  jq -n \
    --arg os "$GOOS_VALUE" \
    --arg arch "$GOARCH_VALUE" \
    --arg download_url "file://$REMOTE_DIR/$zip_name" \
    --arg shasum "$checksum" \
    '{os: $os, arch: $arch, download_url: $download_url, shasum: $shasum}' >> "$platforms_file"
done <<< "$platforms"

sums_name="$(basename "$SUMS_PATH")"
jq -s \
  --argjson protocols "$protocols_json" \
  --arg shasums_url "file://$REMOTE_DIR/$sums_name" \
  --arg signature_url "file://$REMOTE_DIR/$sums_name.sig" \
  '{protocols: $protocols, shasums: {url: $shasums_url, signature_url: $signature_url}, platforms: .}' \
  "$platforms_file" > "$body_file"

MASTER_API_KEY_VALUE="$(terralist_api_key)"
path="/v1/api/providers/$(terralist_urlencode "$TERRALIST_NAMESPACE")/$(terralist_urlencode "$PROVIDER_NAME")/$(terralist_urlencode "$PUBLISH_VERSION")/upload"
response="$(terralist_api_request POST "$path" "$body_file")"
printf '[terralist-publish] provider upload response: %s\n' "$(jq -c . <<< "$response" 2>/dev/null || printf '%s' "$response")" >&2
