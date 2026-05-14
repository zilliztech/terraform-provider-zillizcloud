#!/usr/bin/env bash
set -euo pipefail

GPG_HOME="$(mktemp -d /tmp/tgpg.XXXXXX)"
rm -rf "$GPG_HOME"
mkdir -p "$GPG_HOME"
chmod 700 "$GPG_HOME"

gpgconf --homedir "$GPG_HOME" --launch gpg-agent || true

gpg --homedir "$GPG_HOME" --batch --pinentry-mode loopback --passphrase '' \
  --import /etc/terralist-gpg/gpg-private-key

gpg --homedir "$GPG_HOME" --batch --pinentry-mode loopback --passphrase '' \
  --import /etc/terralist-gpg/gpg-public-key

FINGERPRINT="$(cat /etc/terralist-gpg/gpg-key-id)"
test -n "$FINGERPRINT"
printf '%s\n' "$FINGERPRINT" > .terralist-gpg-fingerprint
printf '%s\n' "$GPG_HOME" > .terralist-gpg-homedir
