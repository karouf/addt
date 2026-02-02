#!/bin/bash
# Gastown - Multi-agent orchestration for Claude Code
# https://github.com/steveyegge/gastown

set -e

echo "Extension [gastown]: Installing dependencies..."

# Install apt dependencies (requires root)
if [ "$(id -u)" = "0" ]; then
    apt-get update && apt-get install -y --no-install-recommends tmux sqlite3
    apt-get clean && rm -rf /var/lib/apt/lists/*
else
    sudo apt-get update && sudo apt-get install -y --no-install-recommends tmux sqlite3
    sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*
fi

echo "Extension [gastown]: Installing Gastown (gt)..."

# Get version from environment (set by main install.sh from config.yaml default or override)
# For Go, "latest" uses @latest, otherwise use @vX.Y.Z format
GASTOWN_VERSION="${GASTOWN_VERSION:-latest}"

if [ "$GASTOWN_VERSION" = "latest" ]; then
    /usr/local/go/bin/go install github.com/steveyegge/gastown/cmd/gt@latest
else
    /usr/local/go/bin/go install github.com/steveyegge/gastown/cmd/gt@v$GASTOWN_VERSION
fi

echo "Extension [gastown]: Done. Installed gt at ~/go/bin/gt"
