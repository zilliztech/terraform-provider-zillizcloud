#!/usr/bin/env bash
set -euo pipefail

if [ -f .terralist-gpg-homedir ]; then
  GPG_HOME="$(cat .terralist-gpg-homedir)"
  if [ -n "$GPG_HOME" ] && [ -d "$GPG_HOME" ]; then
    gpgconf --homedir "$GPG_HOME" --kill gpg-agent || true
    rm -rf "$GPG_HOME"
  fi
fi
