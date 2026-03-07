#!/usr/bin/env sh
# Install splitwisecli from GitHub releases
# Usage: curl -fsSL https://raw.githubusercontent.com/oristides/splitwisecli/main/install.sh | sh

set -e

REPO="oristides/splitwisecli"
BINARY="splitwisecli"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  i386|i686) ARCH="i386" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin) OS="Darwin" ;;
  linux) OS="Linux" ;;
  mingw*|msys*|cygwin*) OS="Windows" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Windows uses zip, others use tar.gz
if [ "$OS" = "Windows" ]; then
  SUFFIX="_${OS}_${ARCH}.zip"
else
  SUFFIX="_${OS}_${ARCH}.tar.gz"
fi

# Get latest release
LATEST_URL=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep "browser_download_url.*${SUFFIX}" | cut -d '"' -f 4)

if [ -z "$LATEST_URL" ]; then
  echo "No release found for ${OS}/${ARCH}"
  echo ""
  echo "This usually means no GitHub release has been published yet."
  echo "Install with Go instead:"
  echo "  go install github.com/${REPO}@latest"
  echo ""
  echo "Or build from source:"
  echo "  git clone https://github.com/${REPO}.git"
  echo "  cd splitwisecli && go build -o splitwisecli"
  echo ""
  echo "To publish releases, tag a version and run: goreleaser release"
  echo "See: https://github.com/${REPO}/releases"
  exit 0
fi

echo "Downloading ${BINARY}..."
mkdir -p "$INSTALL_DIR"
TMP_FILE=$(mktemp ${TMPDIR:-/tmp}/splitwisecli.XXXXXX)

curl -fsSL -o "$TMP_FILE" "$LATEST_URL"

if [ "$OS" = "Windows" ]; then
  unzip -q -o "$TMP_FILE" -d "$INSTALL_DIR"
  BINARY_PATH="$INSTALL_DIR/${BINARY}.exe"
else
  tar -xzf "$TMP_FILE" -C "$INSTALL_DIR"
  BINARY_PATH="$INSTALL_DIR/${BINARY}"
fi

rm -f "$TMP_FILE"
chmod +x "$BINARY_PATH" 2>/dev/null || true

echo ""
echo "Installed to $BINARY_PATH"
echo ""
if ! echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
  echo "Add to PATH: export PATH=\"\$PATH:$INSTALL_DIR\""
  echo ""
fi
echo "Next: run '${BINARY} config' to set up your API credentials."
echo "      Credentials are saved to ~/.config/splitwisecli/config.json"
echo ""
# Require interactive terminal for credential setup (no Scenario B fallback)
if [ "$OS" = "Windows" ]; then
  echo "Run '${BINARY} config' to set up credentials."
elif [ -t 1 ] && [ -e /dev/tty ]; then
  echo "Running credential setup..."
  "$BINARY_PATH" config </dev/tty 2>/dev/tty
else
  echo "Installation succeeded. ${BINARY} is installed at $BINARY_PATH"
  echo ""
  echo "Credential setup was skipped (no interactive terminal — e.g. output redirected, CI, or background)."
  echo ""
  echo "Next step: run '${BINARY} config' from your terminal to set up API credentials."
  echo ""
  echo "To run help: $BINARY_PATH --help"
  echo ""
  echo "Install and read the SKILL on https://github.com/${REPO}/blob/main/skill/SKILL.md to learn how to use splitwisecli"
  echo ""
  exit 0
fi
echo ""
echo "To run help: $BINARY_PATH --help"
echo ""
echo "Install and read the SKILL on https://github.com/${REPO}/blob/main/skill/SKILL.md to learn how to use splitwisecli"
