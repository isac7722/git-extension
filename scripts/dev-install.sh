#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
INSTALL_DIR="${GE_INSTALL_DIR:-/usr/local/bin}"

echo "==> Building ge..."
cd "$PROJECT_DIR"
go build -o ge ./cmd/ge/

echo "==> Installing to $INSTALL_DIR/ge..."
if [[ -w "$INSTALL_DIR" ]]; then
  cp ge "$INSTALL_DIR/ge"
else
  sudo cp ge "$INSTALL_DIR/ge"
fi
chmod +x "$INSTALL_DIR/ge"
rm ge

echo "==> Setting up shell integration..."
"$INSTALL_DIR/ge" setup 2>&1 || true

echo ""
echo "Done! Restart your shell or run:"
echo "  source ~/.zshrc"
