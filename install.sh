#!/bin/bash

set -e

# Define variables
REPO="nullswan/golem"
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
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/golem-cli_${LATEST_RELEASE}_${PLATFORM}_${ARCH}.tar.gz"

# Download and extract
echo "Downloading $DOWNLOAD_URL..."
curl -L -o golem-cli.tar.gz $DOWNLOAD_URL
tar -xzf golem-cli.tar.gz
chmod +x golem-cli
mv golem-cli /usr/local/bin/

# Clean up
rm golem-cli.tar.gz

echo "golem-cli installed successfully!"