#!/bin/bash
# Amp CLI installation (Sourcegraph)
# https://ampcode.com/

set -e

echo "Extension [amp]: Installing Amp CLI..."

# Get version from environment (set by main install.sh from config.yaml default or override)
AMP_VERSION="${AMP_VERSION:-latest}"

# Install Amp globally via npm
if [ "$AMP_VERSION" = "latest" ]; then
    sudo npm install -g @sourcegraph/amp
else
    sudo npm install -g @sourcegraph/amp@$AMP_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(amp --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
echo "Extension [amp]: Done. Installed Amp CLI v${INSTALLED_VERSION}"
