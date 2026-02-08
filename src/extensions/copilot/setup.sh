#!/bin/bash
echo "Setup [copilot]: Initializing GitHub Copilot CLI environment"

# Pre-trust the /workspace directory so copilot skips the interactive trust prompt.
# Trust state is stored in ~/.copilot/config.
if [ "$ADDT_EXT_AUTOTRUST" = "true" ]; then
    echo "Setup [copilot]: Auto-trusting /workspace directory"
    mkdir -p "$HOME/.copilot"
    cat > "$HOME/.copilot/config" <<'EOF'
{
  "trusted_folders": ["/workspace"],
  "banner": "never"
}
EOF
fi
