#!/bin/bash
# Cursor CLI Agent installation
# https://cursor.com/docs/cli/installation

set -e

echo "Extension [cursor]: Installing Cursor CLI..."

# Get version from environment (set by main install.sh from config.yaml default or override)
# Note: The official Cursor installer doesn't support version pinning
CURSOR_VERSION="${CURSOR_VERSION:-latest}"

# Install Cursor CLI using official installer
curl https://cursor.com/install -fsSL | bash

# Create 'cursor' symlink for convenience (official installer creates 'agent')
if [ -f "$HOME/.local/bin/agent" ] && [ ! -f "$HOME/.local/bin/cursor" ]; then
    ln -s agent "$HOME/.local/bin/cursor"
fi

# Verify installation
if command -v agent &> /dev/null; then
    INSTALLED_VERSION=$(agent --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
    echo "Extension [cursor]: Done. Installed Cursor CLI v${INSTALLED_VERSION}"
else
    echo "Warning: agent command not found after installation"
    echo "You may need to add ~/.local/bin to your PATH or install manually from https://cursor.com/cli"
fi
