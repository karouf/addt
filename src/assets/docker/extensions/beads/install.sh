#!/bin/bash
# Beads - Git-backed issue tracker for AI agents
# https://github.com/steveyegge/beads

set -e

echo "Extension [beads]: Installing Beads (bd)..."

# Get version from environment (set by main install.sh from config.yaml default or override)
# For Go, "latest" uses @latest, otherwise use @vX.Y.Z format
BEADS_VERSION="${BEADS_VERSION:-latest}"

if [ "$BEADS_VERSION" = "latest" ]; then
    /usr/local/go/bin/go install github.com/steveyegge/beads/cmd/bd@latest
else
    /usr/local/go/bin/go install github.com/steveyegge/beads/cmd/bd@v$BEADS_VERSION
fi

echo "Extension [beads]: Done. Installed bd at ~/go/bin/bd"
