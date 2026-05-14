#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_unset_proxy
terralist_require_cmd curl
terralist_require_cmd gpg
terralist_require_cmd jq
terralist_require_cmd kubectl

: "${AUTHORITY_FILE:?AUTHORITY_FILE is required}"
: "${TERRALIST_NAMESPACE:?TERRALIST_NAMESPACE is required}"
: "${TERRALIST_URL:?TERRALIST_URL is required}"

terralist_read_gpg_config
authority_json="$(cat "$AUTHORITY_FILE")"
authority_id="$(jq -r '.id // empty' <<< "$authority_json")"
[ -n "$authority_id" ] || terralist_fail "authority response did not include id"

key_id="$(printf '%s' "$GPG_KEY_ID" | tail -c 16 | tr '[:lower:]' '[:upper:]')"
existing_key="$(jq -r --arg key_id "$key_id" 'first(.keys[]? | select((.key_id // "" | ascii_upcase) == $key_id) | .key_id) // empty' <<< "$authority_json")"
if [ -n "$existing_key" ]; then
  printf '[terralist-publish] GPG key %s already exists on authority %s\n' "$key_id" "$TERRALIST_NAMESPACE" >&2
  exit 0
fi

public_key="$(gpg --homedir "$GPG_HOME" --armor --export "$GPG_KEY_ID")"
grep -q 'BEGIN PGP PUBLIC KEY BLOCK' <<< "$public_key" || terralist_fail "gpg did not export an armored public key for $GPG_KEY_ID"

body_file="$(mktemp)"
response_file="$(mktemp)"
trap 'rm -f "$body_file" "$response_file"' EXIT
jq -n \
  --arg key_id "$key_id" \
  --arg ascii_armor "$public_key" \
  '{key_id: $key_id, ascii_armor: $ascii_armor, trust_signature: ""}' > "$body_file"

MASTER_API_KEY_VALUE="$(terralist_api_key)"
path="/v1/api/authorities/$(terralist_urlencode "$authority_id")/keys"
status="$(curl -sS \
  -o "$response_file" \
  -w '%{http_code}' \
  -X POST \
  -H "X-API-Key: $MASTER_API_KEY_VALUE" \
  -H 'Content-Type: application/json' \
  --data-binary "@$body_file" \
  "${TERRALIST_URL%/}$path")"

case "$status" in
  200)
    printf '[terralist-publish] uploaded GPG key %s to authority %s\n' "$key_id" "$TERRALIST_NAMESPACE" >&2
    ;;
  409)
    printf '[terralist-publish] GPG key %s already exists or was rejected as duplicate; continuing\n' "$key_id" >&2
    ;;
  *)
    printf 'POST %s returned HTTP %s: ' "${TERRALIST_URL%/}$path" "$status" >&2
    cat "$response_file" >&2 || true
    printf '\n' >&2
    exit 1
    ;;
esac
