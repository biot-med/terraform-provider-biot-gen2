#!/bin/bash
# Script to sign the checksum file with GPG passphrase from environment variable

SHAFILE="$1"
if [ -z "$SHAFILE" ] || [ ! -f "$SHAFILE" ]; then
    echo "Error: Checksum file not found: $SHAFILE"
    exit 1
fi

echo "${GPG_PASSPHRASE}" | gpg --batch --yes --pinentry-mode loopback --passphrase-fd 0 --detach-sign --armor --default-key 7C56BDFFED7D41BE "$SHAFILE"
if [ $? -eq 0 ]; then
    mv "${SHAFILE}.asc" "${SHAFILE}.sig"
    echo "Signed: ${SHAFILE}.sig"
else
    echo "Error: Failed to sign $SHAFILE"
    exit 1
fi

