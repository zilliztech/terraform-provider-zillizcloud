#!/usr/bin/env bash

terralist_script_dir() {
  cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P
}

TERRALIST_SCRIPT_DIR="$(terralist_script_dir)"

terralist_fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

terralist_log() {
  local prefix="$1"
  shift
  printf '[%s] %s\n' "$prefix" "$*"
}

terralist_require_cmd() {
  command -v "$1" >/dev/null 2>&1 || terralist_fail "$1 is required"
}

terralist_unset_proxy() {
  unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY
}

terralist_init_source_dir() {
  PROVIDER_NAME="${PROVIDER_NAME:-zillizcloud}"
  SOURCE_DIR="${WORKSPACE:-$PWD}"
  SOURCE_DIR="$(cd "$SOURCE_DIR" && pwd -P)"
}

terralist_init_paths() {
  local create_dist="${1:-}"

  : "${PUBLISH_VERSION:?PUBLISH_VERSION is required}"
  terralist_init_source_dir
  DIST_DIR="${DIST_DIR:-$SOURCE_DIR/dist/terralist/$PUBLISH_VERSION}"
  if [ "$create_dist" = "--create-dist" ]; then
    mkdir -p "$DIST_DIR"
  fi
  DIST_DIR="$(cd "$DIST_DIR" && pwd -P)"
  SUMS_PATH="$DIST_DIR/terraform-provider-${PROVIDER_NAME}_${PUBLISH_VERSION}_SHA256SUMS"
}

terralist_validate_platform() {
  local value="$1"
  if [[ ! "$value" =~ ^[a-z0-9]+/[a-z0-9_]+$ ]]; then
    terralist_fail "invalid platform '$value', expected GOOS/GOARCH"
  fi
}

terralist_parse_platform() {
  local value="$1"
  terralist_validate_platform "$value"
  GOOS_VALUE="${value%%/*}"
  GOARCH_VALUE="${value##*/}"
}

terralist_infer_platforms() {
  local os arch

  case "$(uname -s)" in
    Darwin) os="darwin" ;;
    Linux) os="linux" ;;
    *) terralist_fail "cannot infer GOOS from $(uname -s); set PUBLISH_PLATFORMS" ;;
  esac

  case "$(uname -m)" in
    arm64|aarch64) arch="arm64" ;;
    x86_64|amd64) arch="amd64" ;;
    *) terralist_fail "cannot infer GOARCH from $(uname -m); set PUBLISH_PLATFORMS" ;;
  esac

  printf '%s/%s\n' "$os" "$arch"
}

terralist_platforms_from_csv() {
  local raw_platforms="$1"
  local platform found_count
  found_count=0

  IFS=',' read -r -a TERRALIST_PLATFORMS <<< "$raw_platforms"
  for platform in "${TERRALIST_PLATFORMS[@]}"; do
    platform="${platform//[[:space:]]/}"
    [ -n "$platform" ] || continue
    terralist_validate_platform "$platform"
    printf '%s\n' "$platform"
    found_count=$((found_count + 1))
  done

  [ "$found_count" -gt 0 ] || terralist_fail "PUBLISH_PLATFORMS did not include any platforms"
}

terralist_first_platform() {
  local raw_platforms="$1"
  local platform

  IFS=',' read -r -a TERRALIST_PLATFORMS <<< "$raw_platforms"
  for platform in "${TERRALIST_PLATFORMS[@]}"; do
    platform="${platform//[[:space:]]/}"
    if [ -n "$platform" ]; then
      terralist_validate_platform "$platform"
      printf '%s\n' "$platform"
      return 0
    fi
  done

  terralist_fail "PUBLISH_PLATFORMS did not include any platforms"
}

terralist_binary_name() {
  printf 'terraform-provider-%s_v%s\n' "$PROVIDER_NAME" "$PUBLISH_VERSION"
}

terralist_zip_name() {
  local goos="$1"
  local goarch="$2"
  printf 'terraform-provider-%s_%s_%s_%s.zip\n' "$PROVIDER_NAME" "$PUBLISH_VERSION" "$goos" "$goarch"
}

terralist_zip_path() {
  local goos="$1"
  local goarch="$2"
  printf '%s/%s\n' "$DIST_DIR" "$(terralist_zip_name "$goos" "$goarch")"
}

terralist_read_gpg_config() {
  terralist_init_source_dir
  GPG_HOME="$(cat "$SOURCE_DIR/.terralist-gpg-homedir")"
  GPG_KEY_ID="$(cat "$SOURCE_DIR/.terralist-gpg-fingerprint")"
  [ -n "$GPG_HOME" ] || terralist_fail ".terralist-gpg-homedir is empty"
  [ -n "$GPG_KEY_ID" ] || terralist_fail ".terralist-gpg-fingerprint is empty"
}

terralist_api_key() {
  if [ -n "${MASTER_API_KEY:-}" ]; then
    printf '%s\n' "$MASTER_API_KEY"
    return 0
  fi

  : "${KUBE_NAMESPACE:?KUBE_NAMESPACE is required}"
  kubectl -n "$KUBE_NAMESPACE" get secret terralist-secret \
    -o jsonpath='{.data.MASTER_API_KEY}' | { base64 -d 2>/dev/null || base64 -D; }
}

terralist_urlencode() {
  jq -rn --arg value "$1" '$value|@uri'
}

terralist_api_request() {
  local method="$1"
  local path="$2"
  local body_file="${3:-}"
  local expected_status="${4:-200}"
  local response_file status

  response_file="$(mktemp)"
  local curl_args=(-sS -o "$response_file" -w '%{http_code}' -X "$method" -H "X-API-Key: $MASTER_API_KEY_VALUE")
  if [ -n "$body_file" ]; then
    curl_args+=(-H 'Content-Type: application/json' --data-binary "@$body_file")
  fi

  status="$(curl "${curl_args[@]}" "${TERRALIST_URL%/}$path")" || {
    cat "$response_file" >&2 || true
    rm -f "$response_file"
    terralist_fail "$method ${TERRALIST_URL%/}$path failed"
  }

  if [ "$status" != "$expected_status" ]; then
    printf '%s %s returned HTTP %s: ' "$method" "${TERRALIST_URL%/}$path" "$status" >&2
    cat "$response_file" >&2 || true
    printf '\n' >&2
    rm -f "$response_file"
    exit 1
  fi

  cat "$response_file"
  rm -f "$response_file"
}
