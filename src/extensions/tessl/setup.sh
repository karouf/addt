#!/bin/bash
echo "Setup [tessl]: Initializing Tessl environment"

# Only create config if .tessl doesn't exist yet (respect mounted config from automount)
if [ ! -d "$HOME/.tessl" ]; then
    # auth_method: native = tessl init, env = skip (no env-based auth), auto = native
    if [ "$ADDT_EXT_AUTH_AUTOLOGIN" = "true" ]; then
        method="${ADDT_EXT_AUTH_METHOD:-auto}"

        if [ "$method" = "native" ] || [ "$method" = "auto" ]; then
            echo "Setup [tessl]: Auto-initializing Tessl"
            tessl init
        fi
    fi
else
    echo "Setup [tessl]: Found existing .tessl config (likely from automount), not modifying"
fi
