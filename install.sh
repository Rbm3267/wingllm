#!/usr/bin/env bash
# WingLLM Universal Installer for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/Rbm3267/wingllm/main/install.sh | bash

set -e

echo "================================================="
echo " WingLLM Installer (Local AI Speculative Engine) "
echo "================================================="

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

INSTALL_DIR="$HOME/.wingllm/bin"
mkdir -p "$INSTALL_DIR"

BINARY="wingllm-${OS}-${ARCH}"
RELEASE_URL="https://github.com/Rbm3267/wingllm/releases/latest/download/${BINARY}"
DEST="$INSTALL_DIR/wingllm"

echo "Downloading WingLLM for ${OS}/${ARCH}..."
if curl -fsSL "$RELEASE_URL" -o "$DEST"; then
  chmod +x "$DEST"
  echo "Downloaded successfully to $DEST"
else
  echo "Pre-built binary not found. Compiling from source..."
  if command -v go >/dev/null 2>&1; then
    go install github.com/Rbm3267/wingllm/cmd/wingllm@latest
    echo "Installed via go install!"
    exit 0
  fi
fi

# Add to PATH in shell profile
PROFILE=""
if [ -f "$HOME/.zshrc" ]; then
  PROFILE="$HOME/.zshrc"
elif [ -f "$HOME/.bashrc" ]; then
  PROFILE="$HOME/.bashrc"
fi

if [ -n "$PROFILE" ] && ! grep -q "$INSTALL_DIR" "$PROFILE"; then
  echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$PROFILE"
  echo "Added $INSTALL_DIR to $PROFILE"
fi

echo ""
echo " WingLLM installed successfully!"
echo "Run 'wingllm list' to inspect your local models."
echo "Run 'wingllm serve <model>' to start your 2x local AI server."
