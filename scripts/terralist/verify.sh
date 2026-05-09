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
terralist_init_source_dir

MASTER_API_KEY_VALUE="$(terralist_api_key)"

versions_path="/v1/providers/$(terralist_urlencode "$TERRALIST_NAMESPACE")/$(terralist_urlencode "$PROVIDER_NAME")/versions"
versions="$(terralist_api_request GET "$versions_path")"
printf '[terralist-publish] versions endpoint: %s\n' "$(jq -c . <<< "$versions" 2>/dev/null || printf '%s' "$versions")" >&2

first_platform="$(terralist_first_platform "$PUBLISH_PLATFORMS")"
terralist_parse_platform "$first_platform"

download_path="/v1/providers/$(terralist_urlencode "$TERRALIST_NAMESPACE")/$(terralist_urlencode "$PROVIDER_NAME")/$(terralist_urlencode "$PUBLISH_VERSION")/download/$(terralist_urlencode "$GOOS_VALUE")/$(terralist_urlencode "$GOARCH_VALUE")"
download="$(terralist_api_request GET "$download_path")"
jq -e 'type == "object" and has("download_url")' <<< "$download" >/dev/null || terralist_fail "download endpoint did not return provider download metadata"
printf '[terralist-publish] download endpoint verified for %s/%s\n' "$GOOS_VALUE" "$GOARCH_VALUE" >&2
