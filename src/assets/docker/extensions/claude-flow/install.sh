#!/bin/bash
# Claude Flow installation
# https://github.com/ruvnet/claude-flow

set -e

echo "Installing Claude Flow..."

# Install Claude Flow globally via npm (alpha version for v3 features)
sudo npm install -g claude-flow@alpha

# Verify installation
if command -v claude-flow &> /dev/null; then
    echo "Claude Flow installed successfully: $(claude-flow --version 2>/dev/null || echo 'version unknown')"
else
    echo "Warning: claude-flow command not found after installation"
fi
