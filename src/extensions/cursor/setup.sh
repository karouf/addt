#!/bin/bash
echo "Setup [cursor]: Initializing Cursor CLI environment"

# CURSOR_API_KEY is forwarded via env_vars in config.yaml.
# The CLI picks it up automatically for headless authentication.

# Only create config if .cursor doesn't exist yet (respect mounted config from automount)
if [ ! -d "$HOME/.cursor" ]; then
    # Pre-trust the /workspace directory so cursor skips the interactive trust prompt.
    # Cursor stores trust markers per directory in ~/.cursor/projects/<dir-key>/
    if [ "$ADDT_EXT_WORKDIR_AUTOTRUST" = "true" ]; then
        echo "Setup [cursor]: Auto-trusting /workspace directory"
        dir_key=$(echo /workspace | tr '/' '-' | sed 's/^-//')
        mkdir -p "$HOME/.cursor/projects/${dir_key}"
        touch "$HOME/.cursor/projects/${dir_key}/.workspace-trusted"
    fi
else
    echo "Setup [cursor]: Found existing .cursor config (likely from automount), not modifying"
fi
