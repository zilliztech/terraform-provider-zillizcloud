#!/bin/sh
set -eu

case "$(uname -s)" in
  Linux) verify_os="linux" ;;
  Darwin) verify_os="darwin" ;;
  *) verify_os="" ;;
esac
case "$(uname -m)" in
  x86_64|amd64) verify_arch="amd64" ;;
  aarch64|arm64) verify_arch="arm64" ;;
  *) verify_arch="" ;;
esac
verify_platform="${verify_os}/${verify_arch}"

if [ -z "$verify_os" ] || [ -z "$verify_arch" ]; then
  echo "Skipping terraform install verify for unsupported agent platform: $(uname -s)/$(uname -m)"
  exit 0
fi

case ",${PUBLISH_PLATFORMS}," in
  *",${verify_platform},"*) ;;
  *)
    echo "Skipping terraform install verify: published platforms are ${PUBLISH_PLATFORMS}, agent platform is ${verify_platform}"
    exit 0
    ;;
esac

verify_dir="$(mktemp -d)"
trap 'rm -rf "$verify_dir"' EXIT

cat > "${verify_dir}/main.tf" <<EOF
terraform {
  required_providers {
    ${PROVIDER_NAME} = {
      source  = "terralist.zilliz.cc/${TERRALIST_NAMESPACE}/${PROVIDER_NAME}"
      version = "${PUBLISH_VERSION}"
    }
  }
}
EOF

(
  cd "$verify_dir"
  env -u http_proxy -u https_proxy -u HTTP_PROXY -u HTTPS_PROXY \
    TF_CLI_CONFIG_FILE=/dev/null \
    terraform init -input=false
)
