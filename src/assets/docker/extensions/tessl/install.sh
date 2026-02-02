#!/bin/bash
# Tessl - Agent enablement platform
# https://tessl.io/
# https://docs.tessl.io/

set -e

echo "Extension [tessl]: Installing Tessl CLI..."

# Get version from environment (set by main install.sh from config.yaml default or override)
TESSL_VERSION="${TESSL_VERSION:-latest}"

# Install via npm (globally, requires root)
if [ "$TESSL_VERSION" = "latest" ]; then
    sudo npm install -g @tessl/cli
else
    sudo npm install -g @tessl/cli@$TESSL_VERSION
fi

echo "Extension [tessl]: Done. Installed tessl CLI"
echo "  Run 'tessl init' to authenticate and configure"
echo "  Run 'tessl skill search' to find skills"
echo "  Run 'tessl mcp' to start MCP server"
