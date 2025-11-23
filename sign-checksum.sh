#!/bin/bash
# Script to sign the checksum file with GPG passphrase from environment variable

SHAFILE="$1"
if [ -z "$SHAFILE" ] || [ ! -f "$SHAFILE" ]; then
    echo "Error: Checksum file not found: $SHAFILE"
    exit 1
fi

echo "${GPG_PASSPHRASE}" | gpg --batch --yes --pinentry-mode loopback --passphrase-fd 0 --detach-sign --default-key 7C56BDFFED7D41BE "$SHAFILE"
if [ $? -eq 0 ]; then
    # GPG creates .sig file directly (binary signature, not armored)
    if [ -f "${SHAFILE}.sig" ]; then
        echo "Signed: ${SHAFILE}.sig"
    else
        echo "Error: Signature file not created"
        exit 1
    fi
else
    echo "Error: Failed to sign $SHAFILE"
    exit 1
fi

