#!/bin/sh
set -eu
REPO="ujjwalvivek/loom"
BINARY="${1:-loom-mario-term}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64"  ;;
  aarch64|arm64) ARCH="arm64"  ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac
case "$OS" in
  linux|darwin) ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac
echo "Fetching latest release..."
TAG=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4) || {
  echo "ERROR: Could not find any releases."
  exit 1
}
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$TAG/$ARCHIVE"
echo "Downloading $URL ..."
curl -fsSL "$URL" -o "/tmp/$ARCHIVE" || {
  echo "ERROR: Failed to download $URL"
  echo "The binary '$BINARY' may not exist for this platform in release $TAG."
  exit 1
}
echo "Extracting ..."
tar -xzf "/tmp/$ARCHIVE" -C /tmp
mkdir -p "$INSTALL_DIR"
mv "/tmp/$BINARY" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY"
rm -f "/tmp/$ARCHIVE"
echo ""
echo "$BINARY $TAG installed to:"
echo "  $INSTALL_DIR/$BINARY"
echo ""
echo "Run it now:  $INSTALL_DIR/$BINARY"
echo "After restart:  $BINARY"
