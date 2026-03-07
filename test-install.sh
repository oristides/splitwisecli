#!/usr/bin/env sh
# Simulate one-liner install from local build (no GitHub release needed)
# Usage: ./test-install.sh   (run from repo root)

set -e

BINARY="splitwisecli"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS and arch (same as install.sh)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in x86_64|amd64) ARCH="x86_64" ;; aarch64|arm64) ARCH="arm64" ;; i386|i686) ARCH="i386" ;; *) echo "Unsupported: $ARCH"; exit 1 ;; esac
case "$OS" in darwin) OS="Darwin" ;; linux) OS="Linux" ;; *) echo "Unsupported: $OS"; exit 1 ;; esac

echo "Building ${BINARY}..."
go build -o "$BINARY" .

echo "Creating release tarball (simulating goreleaser)..."
TARBALL="${BINARY}_${OS}_${ARCH}.tar.gz"
tar -czf "$TARBALL" "$BINARY"

echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
tar -xzf "$TARBALL" -C "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/$BINARY"

rm -f "$TARBALL" "$BINARY"

echo ""
echo "Installed to $INSTALL_DIR/$BINARY"
echo ""
if ! echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
  echo "Add to PATH: export PATH=\"\$PATH:$INSTALL_DIR\""
  echo ""
fi
echo "Next: run '$BINARY config' to set up your API credentials."
echo "      Credentials are saved to ~/.config/splitwisecli/config.json"
echo ""
# Require interactive terminal for credential setup (no Scenario B fallback)
if [ -t 1 ] && [ -e /dev/tty ]; then
  echo "Running credential setup..."
  "$INSTALL_DIR/$BINARY" config </dev/tty 2>/dev/tty
else
  echo "Installation succeeded. $BINARY is installed at $INSTALL_DIR/$BINARY"
  echo ""
  echo "Credential setup was skipped (no interactive terminal — e.g. output redirected, CI, or background)."
  echo ""
  echo "Next step: run '$BINARY config' from your terminal to set up API credentials."
  echo ""
  echo "To run help: $INSTALL_DIR/$BINARY --help"
  echo ""
  exit 0
fi
echo ""
echo "To run help: $INSTALL_DIR/$BINARY --help"
