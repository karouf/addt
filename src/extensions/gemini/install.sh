#!/bin/bash
# Gemini CLI installation (Google)
# https://github.com/google-gemini/gemini-cli

set -e

echo "Extension [gemini]: Installing Gemini CLI..."

# Get version from environment (set by main install.sh from config.yaml default or override)
GEMINI_VERSION="${GEMINI_VERSION:-latest}"

# Install Gemini CLI globally via npm
if [ "$GEMINI_VERSION" = "latest" ]; then
    npm install -g @google/gemini-cli
else
    npm install -g @google/gemini-cli@$GEMINI_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(gemini --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
# Cleaning up the .gemini directory, at this first run creates it
rm -rf "$HOME/.gemini"
echo "Extension [gemini]: Done. Installed Gemini CLI v${INSTALLED_VERSION}"
