#!/bin/bash
# Claude Code - AI coding assistant by Anthropic
# https://github.com/anthropics/claude-code

set -e

echo "Extension [claude]: Installing Claude Code..."

# Get version from environment or default to latest
CLAUDE_VERSION="${CLAUDE_VERSION:-latest}"

# Install via npm (globally, requires root)
if [ "$CLAUDE_VERSION" = "latest" ] || [ "$CLAUDE_VERSION" = "stable" ]; then
    sudo npm install -g @anthropic-ai/claude-code
else
    sudo npm install -g @anthropic-ai/claude-code@$CLAUDE_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(claude --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
echo "Extension [claude]: Done. Installed Claude Code v${INSTALLED_VERSION}"
