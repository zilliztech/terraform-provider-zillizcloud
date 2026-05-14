#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_unset_proxy
terralist_require_cmd curl
terralist_require_cmd jq
terralist_require_cmd kubectl

: "${TERRALIST_NAMESPACE:?TERRALIST_NAMESPACE is required}"
: "${TERRALIST_URL:?TERRALIST_URL is required}"
MASTER_API_KEY_VALUE="$(terralist_api_key)"

authorities="$(terralist_api_request GET '/v1/api/authorities/')"
authority_id="$(jq -r --arg name "$TERRALIST_NAMESPACE" 'first(.[]? | select(.name == $name) | .id) // empty' <<< "$authorities")"
if [ -n "$authority_id" ] && [ "$authority_id" != "null" ]; then
  printf '%s\n' "$authorities" | jq -c --arg id "$authority_id" '.[] | select((.id|tostring) == $id)'
  exit 0
fi

printf '[terralist-publish] creating Terralist authority %s\n' "$TERRALIST_NAMESPACE" >&2
body_file="$(mktemp)"
trap 'rm -f "$body_file"' EXIT
jq -n --arg name "$TERRALIST_NAMESPACE" '{name: $name, public: true}' > "$body_file"
terralist_api_request POST '/v1/api/authorities/' "$body_file" '201'
