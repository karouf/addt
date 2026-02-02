#!/bin/bash
# Kiro CLI installation (AWS)
# https://kiro.dev/docs/cli/installation/

set -e

echo "Extension [kiro]: Installing Kiro CLI..."

# Get version from environment (set by main install.sh from config.yaml default or override)
# Note: The official Kiro installer may not support version pinning
KIRO_VERSION="${KIRO_VERSION:-latest}"

# Install required dependencies
sudo apt-get update && sudo apt-get install -y unzip

# Install Kiro CLI using official installer
curl -fsSL https://cli.kiro.dev/install | bash

# Verify installation
if command -v kiro-cli &> /dev/null; then
    INSTALLED_VERSION=$(kiro-cli --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
    echo "Extension [kiro]: Done. Installed Kiro CLI v${INSTALLED_VERSION}"
else
    if [ -f "$HOME/.local/bin/kiro-cli" ]; then
        echo "Extension [kiro]: Done. Installed to ~/.local/bin/kiro-cli"
    else
        echo "Warning: kiro-cli command not found after installation"
    fi
fi
