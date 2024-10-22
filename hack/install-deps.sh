#!/bin/bash

set -e

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sudo apt-get update
    # gcc
    sudo apt install -y gcc libc6-dev
    # x11
    sudo apt install -y libx11-dev xorg-dev libxtst-dev
    # Clipboard
    sudo apt install -y xsel xclip
    # Bitmap
    sudo apt install -y libpng++-dev
    # GoHook
    sudo apt install -y xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev libxkbcommon-dev
    # portaudio
    sudo apt-get install -y portaudio19-dev
    go mod download
elif [[ "$OSTYPE" == "darwin"* ]]; then
    brew install portaudio
    go mod download
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    # TODO(nullswan): Windows dev setup...
    # you will require to install portaudio manually.
    # Check the release.yml for more details.
    # You can hack the portaudio installation by using the following command:
    # pip install pyaudio
    go mod download
else
    echo "Unsupported OS"
    exit 1
fi
