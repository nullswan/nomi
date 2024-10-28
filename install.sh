#!/bin/bash

set -e

# Define variables
REPO="nullswan/nomi"
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest | jq -r .tag_name)
PLATFORM="$(uname | tr '[:upper:]' '[:lower:]')"
ARCHITECTURE="$(uname -m)"

case "${PLATFORM}-${ARCHITECTURE}" in
  darwin-arm64)
    ARCH="arm64"
    ;;
  linux-386)
    ARCH="386"
    ;;
  linux-amd64)
    ARCH="amd64"
    ;;
  linux-arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported platform or architecture: ${PLATFORM}-${ARCHITECTURE}"
    exit 1
    ;;
esac

# Construct the download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/nomi-cli-${PLATFORM}-${ARCH}"

# Download
echo "Downloading $DOWNLOAD_URL..."
curl -L -o nomi-cli $DOWNLOAD_URL
chmod +x nomi-cli
mv nomi-cli /usr/local/bin/

echo "nomi-cli installed successfully!"