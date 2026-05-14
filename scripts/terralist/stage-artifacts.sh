#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/lib.sh"

terralist_require_cmd kubectl

: "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
: "${PUBLISH_PLATFORMS:?PUBLISH_PLATFORMS is required}"
: "${KUBE_NAMESPACE:?KUBE_NAMESPACE is required}"

TERRALIST_DEPLOYMENT="${TERRALIST_DEPLOYMENT:-terralist}"
PVC_SOURCE_ROOT="${PVC_SOURCE_ROOT:-/data/source}"
terralist_init_paths

artifacts=()
platforms="$(terralist_platforms_from_csv "$PUBLISH_PLATFORMS")"
while IFS= read -r platform; do
  terralist_parse_platform "$platform"
  artifacts+=("$(terralist_zip_path "$GOOS_VALUE" "$GOARCH_VALUE")")
done <<< "$platforms"
artifacts+=("$SUMS_PATH" "$SUMS_PATH.sig")

for artifact in "${artifacts[@]}"; do
  [ -f "$artifact" ] || terralist_fail "artifact is missing: $artifact"
done

pod="$(kubectl -n "$KUBE_NAMESPACE" get pod -l "app.kubernetes.io/name=$TERRALIST_DEPLOYMENT" -o 'jsonpath={.items[0].metadata.name}')"
[ -n "$pod" ] || terralist_fail "could not find Terralist pod for deployment $TERRALIST_DEPLOYMENT"

remote_dir="${PVC_SOURCE_ROOT%/}/${PROVIDER_NAME}/${PUBLISH_VERSION}"
kubectl -n "$KUBE_NAMESPACE" exec "deploy/$TERRALIST_DEPLOYMENT" -- mkdir -p "$remote_dir" >&2
for artifact in "${artifacts[@]}"; do
  kubectl -n "$KUBE_NAMESPACE" cp "$artifact" "$pod:$remote_dir/$(basename "$artifact")" >&2
done
printf '%s\n' "$remote_dir"
