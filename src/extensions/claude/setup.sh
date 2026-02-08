#!/bin/bash
set -e
echo "Setup [claude]: Initializing Claude Code environment"

CLAUDE_JSON="$HOME/.claude.json"
CLAUDE_DIR="$HOME/.claude"
CLAUDE_INTERNAL_JSON="$CLAUDE_DIR/claude.json"
CLAUDE_CREDENTIALS_FILE="$CLAUDE_DIR/.credentials.json"

# Check if user already has authentication configured (from mounted config via automount)
if [ -d "$CLAUDE_DIR" ] "$CLAUDE_JSON" 2>/dev/null; then
    echo "Setup [claude]: Found existing Claude config (likely from automount), not modifying"
    exit 0
fi

# Continuining with setup as no $ClAUDE dir exists
echo "Setup [claude]: No $CLAUDE_DIR found, creating it"
mkdir -p "$CLAUDE_DIR"

# if no config file, create it
echo "Setup [claude]: Creating $CLAUDE_JSON"

# bypassPermissionsModeAccepted controlled by ADDT_EXTENSION_CLAUDE_YOLO (default: false)
BYPASS="${ADDT_EXTENSION_CLAUDE_YOLO:-false}"
echo "Setup [claude]: bypassPermissionsModeAccepted=$BYPASS"

# Build the base config JSON
if [ "$ADDT_EXT_AUTOTRUST" = "true" ]; then
    echo "Setup [claude]: Auto-trusting /workspace directory"
    cat > "$CLAUDE_JSON" << EOF
{
  "hasCompletedOnboarding": true,
  "hasTrustDialogAccepted": true,
  "bypassPermissionsModeAccepted": $BYPASS,
  "projects": {
    "/workspace": {
      "allowedTools": [],
      "hasTrustDialogAccepted": true,
      "hasCompletedProjectOnboarding": true
    }
  }
}
EOF
else
    cat > "$CLAUDE_JSON" << EOF
{
  "hasCompletedOnboarding": true,
  "hasTrustDialogAccepted": true,
  "bypassPermissionsModeAccepted": $BYPASS
}
EOF
fi

# Create internal config (~/.claude/claude.json) - hooks trust dialog
echo "Setup [claude]: Creating $CLAUDE_INTERNAL_JSON (trusting hooks)"
cat > "$CLAUDE_INTERNAL_JSON" << 'EOF'
{
  "hasTrustDialogHooksAccepted": true,
  "hasCompletedOnboarding": true
}
EOF

# If auto_login is enabled, configure authentication based on login_method
# login_method: env = API key, native = OAuth, auto = try env first then native
if [ "$ADDT_EXT_AUTO_LOGIN" = "true" ]; then
    method="${ADDT_EXT_LOGIN_METHOD:-auto}"

    # env or auto: configure API key authentication if ANTHROPIC_API_KEY is available
    if [ "$method" = "env" ] || [ "$method" = "auto" ]; then
        if [ -n "$ANTHROPIC_API_KEY" ]; then
            API_KEY_LAST_20="${ANTHROPIC_API_KEY: -20}"
            echo "Setup [claude]: Found ANTHROPIC_API_KEY, configuring for API key authentication"
            if command -v jq >/dev/null 2>&1; then
                tmpfile="$(mktemp)"
                jq --arg ak "$API_KEY_LAST_20" '
                    .customApiKeyResponses = {
                        approved: [$ak],
                        rejected: []
                    }
                ' "$CLAUDE_JSON" > "$tmpfile" && mv "$tmpfile" "$CLAUDE_JSON"
            else
                echo "Warning: jq not found, cannot insert API key into $CLAUDE_JSON"
                exit 1
            fi
            echo "Setup [claude]: Configured for API key authentication"
        fi
    fi

    # native or auto: configure OAuth credentials if available
    if [ "$method" = "native" ] || [ "$method" = "auto" ]; then
        if [ -n "$CLAUDE_OAUTH_CREDENTIALS" ]; then
            echo "Setup [claude]: Found OAuth credentials, configuring for OAuth authentication"
            echo "$CLAUDE_OAUTH_CREDENTIALS" | base64 -d > "$CLAUDE_CREDENTIALS_FILE"
            echo "Setup [claude]: Decoded $CLAUDE_CREDENTIALS_FILE"
            echo "Setup [claude]: Completed OAuth authentication setup"
        fi
    fi
fi

echo "Setup [claude]: Completed Claude Code environment setup"