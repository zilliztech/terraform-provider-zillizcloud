#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_require_cmd go
terralist_require_cmd zip

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
: "${PUBLISH_PLATFORM:?PUBLISH_PLATFORM is required}"
terralist_parse_platform "$PUBLISH_PLATFORM"

PACKAGE="${PACKAGE:-.}"
CGO_ENABLED="${CGO_ENABLED:-0}"
GO_LDFLAGS="${GO_LDFLAGS:--s -w -X main.version=${PUBLISH_VERSION#v}}"
terralist_init_paths --create-dist

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/terralist-provider-build.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

binary_name="$(terralist_binary_name)"
zip_name="$(terralist_zip_name "$GOOS_VALUE" "$GOARCH_VALUE")"
build_dir="$tmp_dir/${GOOS_VALUE}_${GOARCH_VALUE}"
mkdir -p "$build_dir"

terralist_log "terralist-build" "go build ${GOOS_VALUE}/${GOARCH_VALUE} version=${PUBLISH_VERSION#v}"
(
  cd "$SOURCE_DIR"
  CGO_ENABLED="$CGO_ENABLED" GOOS="$GOOS_VALUE" GOARCH="$GOARCH_VALUE" \
    go build -buildvcs=false -ldflags "$GO_LDFLAGS" -o "$build_dir/$binary_name" "$PACKAGE"
)
chmod 755 "$build_dir/$binary_name"

terralist_log "terralist-build" "zip $zip_name"
rm -f "$DIST_DIR/$zip_name"
(
  cd "$build_dir"
  zip -q -9 -X "$DIST_DIR/$zip_name" "$binary_name"
)
