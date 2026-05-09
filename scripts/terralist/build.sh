#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_unset_proxy

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
raw_platforms="${PUBLISH_PLATFORMS:-}"
if [ -z "$raw_platforms" ]; then
  raw_platforms="$(terralist_infer_platforms)"
fi

built_count=0
platforms="$(terralist_platforms_from_csv "$raw_platforms")"
while IFS= read -r platform; do
  PUBLISH_PLATFORM="$platform" bash "$TERRALIST_SCRIPT_DIR/build-platform.sh"
  built_count=$((built_count + 1))
done <<< "$platforms"

[ "$built_count" -gt 0 ] || terralist_fail "PUBLISH_PLATFORMS did not include any platforms"
terralist_log "terralist-build" "built $built_count provider artifact(s) in ${DIST_DIR:-${WORKSPACE:-$PWD}/dist/terralist/$PUBLISH_VERSION}"
