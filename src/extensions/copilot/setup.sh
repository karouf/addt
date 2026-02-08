#!/bin/bash
echo "Setup [copilot]: Initializing GitHub Copilot CLI environment"

# Only create config if .copilot doesn't exist yet (respect mounted config from automount)
if [ ! -d "$HOME/.copilot" ]; then
    # Pre-trust the /workspace directory so copilot skips the interactive trust prompt.
    # Trust state is stored in ~/.copilot/config.
    if [ "$ADDT_EXT_WORKDIR_AUTOTRUST" = "true" ]; then
        echo "Setup [copilot]: Auto-trusting /workspace directory"
        mkdir -p "$HOME/.copilot"
        cat > "$HOME/.copilot/config" <<'EOF'
{
  "trusted_folders": ["/workspace"],
  "banner": "never"
}
EOF
    fi
else
    echo "Setup [copilot]: Found existing .copilot config (likely from automount), not modifying"
fi
