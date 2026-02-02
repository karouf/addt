#!/bin/bash
# Kiro CLI installation (AWS)
# https://kiro.dev/docs/cli/installation/

set -e

echo "Installing Kiro CLI..."

# Install required dependencies
sudo apt-get update && sudo apt-get install -y unzip

# Install Kiro CLI using official installer
curl -fsSL https://cli.kiro.dev/install | bash

# Verify installation
if command -v kiro-cli &> /dev/null; then
    echo "Kiro CLI installed successfully: $(kiro-cli --version 2>/dev/null || echo 'version unknown')"
else
    # Check common install locations
    if [ -f "$HOME/.local/bin/kiro-cli" ]; then
        echo "Kiro CLI installed to ~/.local/bin/kiro-cli"
    else
        echo "Warning: kiro-cli command not found after installation"
    fi
fi
