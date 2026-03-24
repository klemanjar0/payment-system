#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
KEYS_DIR="$SCRIPT_DIR/../keys"

mkdir -p "$KEYS_DIR"

openssl ecparam -name prime256v1 -genkey -noout -out "$KEYS_DIR/private.pem"
openssl ec -in "$KEYS_DIR/private.pem" -pubout -out "$KEYS_DIR/public.pem"

echo "Keys generated in $KEYS_DIR"
