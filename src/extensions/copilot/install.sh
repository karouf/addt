#!/bin/bash
# GitHub Copilot CLI installation
# https://github.com/github/copilot-cli

set -e

echo "Extension [copilot]: Installing GitHub Copilot CLI..."

# Get version from environment (set by main install.sh from config.yaml default or override)
COPILOT_VERSION="${COPILOT_VERSION:-latest}"

# Install Copilot CLI globally via npm
if [ "$COPILOT_VERSION" = "latest" ]; then
    npm install -g @github/copilot
else
    npm install -g @github/copilot@$COPILOT_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(copilot --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
# Clean up the .copilot directory, at this first run creates it
rm -rf "$HOME/.copilot"

echo "Extension [copilot]: Done. Installed Copilot CLI v${INSTALLED_VERSION}"
