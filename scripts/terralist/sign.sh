#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_require_cmd gpg

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
terralist_init_paths
[ -f "$SUMS_PATH" ] || terralist_fail "checksum file is missing: $SUMS_PATH"

terralist_read_gpg_config
gpg_cmd=(gpg --batch --yes \
  --homedir "$GPG_HOME" \
  --pinentry-mode "${GPG_PINENTRY_MODE:-loopback}" \
  --local-user "$GPG_KEY_ID" \
  --detach-sign \
  --output "$SUMS_PATH.sig")
if [ -n "${GPG_PASSPHRASE:-}" ]; then
  printf '%s' "$GPG_PASSPHRASE" | "${gpg_cmd[@]}" --passphrase-fd 0 "$SUMS_PATH"
else
  "${gpg_cmd[@]}" "$SUMS_PATH"
fi
printf '[terralist-publish] wrote %s.sig\n' "$SUMS_PATH" >&2
