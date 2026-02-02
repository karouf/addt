#!/bin/bash
# OpenClaw installation (formerly Clawdbot/Moltbot)
# https://github.com/openclaw/openclaw

set -e

echo "Extension [openclaw]: Installing OpenClaw..."

# Get version from environment (set by main install.sh from config.yaml default or override)
OPENCLAW_VERSION="${OPENCLAW_VERSION:-latest}"

# Install OpenClaw globally via npm
if [ "$OPENCLAW_VERSION" = "latest" ]; then
    sudo npm install -g openclaw
else
    sudo npm install -g openclaw@$OPENCLAW_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(openclaw --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
echo "Extension [openclaw]: Done. Installed OpenClaw v${INSTALLED_VERSION}"
