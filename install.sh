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
curl --location --progress-bar --output nomi-cli $DOWNLOAD_URL

echo "Installing nomi..."
mkdir -p ~/.local/bin
mv nomi-cli ~/.local/bin/nomi
chmod +x ~/.local/bin/nomi
rm -f nomi-cli

read -p "Do you want to add ~/.local/bin to your PATH? (y/n): " response < /dev/tty
if [[ "$response" =~ ^[Yy]$ ]]; then
    for rc in ~/.bashrc ~/.zshrc; do
        if ! grep -q 'export PATH="$HOME/.local/bin:$PATH"' "$rc"; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$rc"
            echo "Added to $rc"
        fi
    done
    echo "PATH updated. Please restart your terminal or run 'source ~/.bashrc' or 'source ~/.zshrc'."
fi

echo "nomi installed successfully!"