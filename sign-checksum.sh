#!/usr/bin/env bash
set -euo pipefail

SHAFILE="$1"

if [[ -z "$SHAFILE" || ! -f "$SHAFILE" ]]; then
  echo "Error: Checksum file not found: $SHAFILE"
  exit 1
fi

if [[ -z "${GPG_PASSPHRASE:-}" ]]; then
  echo "Error: GPG_PASSPHRASE is not set"
  exit 1
fi

gpg \
  --batch \
  --yes \
  --pinentry-mode loopback \
  --passphrase "${GPG_PASSPHRASE}" \
  --default-key D09266D4C73C32D8 \
  --detach-sign \
  "$SHAFILE"

echo "Signed: ${SHAFILE}.sig"
