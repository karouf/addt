#!/bin/bash
# OpenAI Codex CLI installation
# https://github.com/openai/codex

set -e

echo "Extension [codex]: Installing OpenAI Codex CLI..."

# Get version from environment (set by main install.sh from config.yaml default or override)
CODEX_VERSION="${CODEX_VERSION:-latest}"

# Install codex globally via npm
if [ "$CODEX_VERSION" = "latest" ]; then
    sudo npm install -g @openai/codex
else
    sudo npm install -g @openai/codex@$CODEX_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(codex --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
echo "Extension [codex]: Done. Installed Codex CLI v${INSTALLED_VERSION}"
