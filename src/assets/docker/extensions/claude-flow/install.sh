#!/bin/bash
# Claude Flow installation
# https://github.com/ruvnet/claude-flow

set -e

echo "Extension [claude-flow]: Installing Claude Flow..."

# Get version from environment (set by main install.sh from config.yaml default or override)
# Default is "alpha" for v3 features
CLAUDE_FLOW_VERSION="${CLAUDE_FLOW_VERSION:-alpha}"

# Install Claude Flow globally via npm
if [ "$CLAUDE_FLOW_VERSION" = "latest" ]; then
    sudo npm install -g claude-flow
else
    sudo npm install -g claude-flow@$CLAUDE_FLOW_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(claude-flow --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
echo "Extension [claude-flow]: Done. Installed Claude Flow v${INSTALLED_VERSION}"
