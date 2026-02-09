#!/bin/bash
set -e
echo "Setup [gemini]: Initializing Gemini CLI environment"

# Makes gemini skip IDE integration nudge
unset TERM_PROGRAM
unset GEMINI_CLI_IDE_SERVER_PORT

# Only create config if .gemini doesn't exist yet (respect mounted config from automount)
if [ ! -d "$HOME/.gemini" ]; then
    # Pre-configure auth type so gemini-cli skips the interactive first-run wizard.
    # auth_method: env = API key, native = Google OAuth, auto = try env first
    if [ "$ADDT_EXT_AUTH_AUTOLOGIN" = "true" ]; then
        method="${ADDT_EXT_AUTH_METHOD:-auto}"

        if [ "$method" = "env" ] || [ "$method" = "auto" ]; then
            if [ -n "$GEMINI_API_KEY" ]; then
                echo "Setup [gemini]: Auto-configuring API key authentication"
                mkdir -p "$HOME/.gemini"
                cat > "$HOME/.gemini/settings.json" <<EOF
{
  "security": {
    "auth": {
      "selectedType": "gemini-api-key"
    }
  },
  "hasSeenIdeIntegrationNudge": true
}
EOF
            fi
        fi
    fi
else
    echo "Setup [gemini]: Found existing .gemini config (likely from automount), not modifying"
fi
