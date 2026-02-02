#!/bin/bash
# Backlog.md installation
# https://github.com/MrLesk/Backlog.md

set -e

echo "Installing Backlog.md..."

# Install Backlog.md globally via npm
sudo npm install -g backlog.md

# Verify installation
if command -v backlog &> /dev/null; then
    echo "Backlog.md installed successfully: $(backlog --version 2>/dev/null || echo 'version unknown')"
else
    echo "Warning: backlog command not found after installation"
fi
