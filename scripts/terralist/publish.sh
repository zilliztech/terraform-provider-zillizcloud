#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_unset_proxy

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
: "${PUBLISH_PLATFORMS:?PUBLISH_PLATFORMS is required}"

bash "$TERRALIST_SCRIPT_DIR/check-artifacts.sh"
bash "$TERRALIST_SCRIPT_DIR/checksum.sh"
bash "$TERRALIST_SCRIPT_DIR/sign.sh"

authority_file="$(mktemp)"
trap 'rm -f "$authority_file"' EXIT
bash "$TERRALIST_SCRIPT_DIR/ensure-authority.sh" > "$authority_file"

if [ "${SKIP_KEY_UPLOAD:-0}" != "1" ]; then
  AUTHORITY_FILE="$authority_file" bash "$TERRALIST_SCRIPT_DIR/ensure-key.sh"
fi

remote_dir="$(bash "$TERRALIST_SCRIPT_DIR/stage-artifacts.sh")"
REMOTE_DIR="$remote_dir" bash "$TERRALIST_SCRIPT_DIR/upload.sh"

if [ "${SKIP_VERIFY:-0}" != "1" ]; then
  bash "$TERRALIST_SCRIPT_DIR/verify.sh"
fi

printf '[terralist-publish] done\n'
