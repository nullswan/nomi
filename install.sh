#!/bin/bash

set -e

# Define variables
REPO="nullswan/nomi"
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest | jq -r .tag_name)
PLATFORM="$(uname | tr '[:upper:]' '[:lower:]')"
ARCHITECTURE="$(uname -m)"

case "$ARCHITECTURE" in
  x86_64) ARCH="amd64" ;;
  arm64) ARCH="arm64" ;;
  i386) ARCH="386" ;;
  arm) ARCH="arm" ;;
  ppc64le) ARCH="ppc64le" ;;
  *) echo "Unsupported architecture: $ARCHITECTURE"; exit 1 ;;
esac

# Construct the download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/nomi-cli_${LATEST_RELEASE}_${PLATFORM}_${ARCH}.tar.gz"

# Download and extract
echo "Downloading $DOWNLOAD_URL..."
curl -L -o nomi-cli.tar.gz $DOWNLOAD_URL
tar -xzf nomi-cli.tar.gz
chmod +x nomi-cli
mv nomi-cli /usr/local/bin/

# Clean up
rm nomi-cli.tar.gz

echo "nomi-cli installed successfully!"