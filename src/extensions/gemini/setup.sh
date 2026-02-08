#!/bin/bash
echo "Setup [gemini]: Initializing Gemini CLI environment"

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
                echo '{"security":{"auth":{"selectedType":"gemini-api-key"}}}' > "$HOME/.gemini/settings.json"
            fi
        fi
    fi
else
    echo "Setup [gemini]: Found existing .gemini config (likely from automount), not modifying"
fi
