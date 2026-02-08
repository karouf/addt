#!/bin/bash
echo "Setup [tessl]: Initializing Tessl environment"

# auth_method: native = tessl init, env = skip (no env-based auth), auto = native
if [ "$ADDT_EXT_AUTH_AUTOLOGIN" = "true" ]; then
    method="${ADDT_EXT_AUTH_METHOD:-auto}"

    if [ "$method" = "native" ] || [ "$method" = "auto" ]; then
        echo "Setup [tessl]: Auto-initializing Tessl"
        tessl init
    fi
fi
