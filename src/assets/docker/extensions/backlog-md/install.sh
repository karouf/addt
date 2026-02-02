#!/bin/bash
# Backlog.md installation
# https://github.com/MrLesk/Backlog.md

set -e

echo "Extension [backlog-md]: Installing Backlog.md..."

# Get version from environment (set by main install.sh from config.yaml default or override)
# Note: The env var name uses underscore: BACKLOG_MD_VERSION
BACKLOG_MD_VERSION="${BACKLOG_MD_VERSION:-latest}"

# Install Backlog.md globally via npm
if [ "$BACKLOG_MD_VERSION" = "latest" ]; then
    sudo npm install -g backlog.md
else
    sudo npm install -g backlog.md@$BACKLOG_MD_VERSION
fi

# Verify installation
INSTALLED_VERSION=$(backlog --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
echo "Extension [backlog-md]: Done. Installed Backlog.md v${INSTALLED_VERSION}"
