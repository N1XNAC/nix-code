#!/usr/bin/env bash
set -euo pipefail

REPO="N1XNAC/nix-code"
BINARY="n1x"
VERSION="${1:-latest}"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}"
echo "  Installing N1X Code - Terminal AI Coding Agent"
echo -e "${NC}"

detect_arch() {
  local arch
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "unsupported: $arch" >&2; exit 1 ;;
  esac
}

detect_os() {
  local os
  os=$(uname -s)
  case "$os" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) echo "unsupported: $os" >&2; exit 1 ;;
  esac
}

INSTALL_DIR="${INSTALL_DIR:-$HOME/.n1x/bin}"
mkdir -p "$INSTALL_DIR"

OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$VERSION" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/${BINARY}_${OS}_${ARCH}.tar.gz"
else
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/v$VERSION/${BINARY}_${OS}_${ARCH}.tar.gz"
fi

echo "  OS: $OS"
echo "  Arch: $ARCH"
echo "  Downloading from: $DOWNLOAD_URL"
echo ""

TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

if command -v curl &>/dev/null; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/nix.tar.gz"
elif command -v wget &>/dev/null; then
  wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/nix.tar.gz"
else
  echo "Error: need curl or wget" >&2
  exit 1
fi

tar -xzf "$TMP_DIR/nix.tar.gz" -C "$TMP_DIR"
mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

add_to_path() {
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    local shell_config
    case "${SHELL:-}" in
      */zsh) shell_config="$HOME/.zshrc" ;;
      */bash) shell_config="$HOME/.bashrc" ;;
      *) shell_config="$HOME/.profile" ;;
    esac
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$shell_config"
    echo "  Added $INSTALL_DIR to PATH in $shell_config"
  fi
}

add_to_path

echo -e "${GREEN}"
  echo "  ✓ N1X Code installed successfully!"
echo -e "${NC}"
  echo "  Run 'n1x config' to set up your API keys"
  echo "  Run 'n1x run \"your prompt\"' to start coding"
  echo "  Run 'n1x' to launch the interactive TUI"
echo ""
echo "  Restart your terminal or run:"
  echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
