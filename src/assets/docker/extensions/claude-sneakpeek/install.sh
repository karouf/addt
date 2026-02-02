#!/bin/bash
# Claude Sneakpeek installation
# https://github.com/mikekelly/claude-sneakpeek

set -e

echo "Extension [claude-sneakpeek]: Installing Claude Sneakpeek..."

# Get version from environment (set by main install.sh from config.yaml default or override)
CLAUDE_SNEAKPEEK_VERSION="${CLAUDE_SNEAKPEEK_VERSION:-latest}"

# Install using npx quick installer with custom name
# Note: npx installs latest by default, version pinning via npx is limited
if [ "$CLAUDE_SNEAKPEEK_VERSION" = "latest" ]; then
    npx @realmikekelly/claude-sneakpeek quick --name claudesp
else
    npx @realmikekelly/claude-sneakpeek@$CLAUDE_SNEAKPEEK_VERSION quick --name claudesp
fi

# Verify installation
if command -v claudesp &> /dev/null; then
    echo "Extension [claude-sneakpeek]: Done. Installed claudesp"
else
    if [ -f "$HOME/.local/bin/claudesp" ]; then
        echo "Extension [claude-sneakpeek]: Done. Installed to ~/.local/bin/claudesp"
    else
        echo "Warning: claudesp command not found after installation"
    fi
fi
