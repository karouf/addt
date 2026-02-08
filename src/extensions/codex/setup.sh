#!/bin/bash
echo "Setup [codex]: Initializing OpenAI Codex environment"

CODEX_DIR="$HOME/.codex"
CODEX_CONFIG="$CODEX_DIR/config.toml"

# Only create config if .codex doesn't exist yet (respect mounted config from automount)
if [ ! -d "$CODEX_DIR" ]; then
    mkdir -p "$CODEX_DIR"

    # Auto-trust workspace directory if configured
    # https://developers.openai.com/codex/config-reference/
    if [ "$ADDT_EXT_WORKDIR_AUTOTRUST" = "true" ]; then
        echo "Setup [codex]: Auto-trusting /workspace directory"
        cat > "$CODEX_CONFIG" << 'EOF'
[projects."/workspace"]
trust_level = "trusted"
EOF
    fi
else
    echo "Setup [codex]: Found existing .codex config (likely from automount), not modifying"
fi

# auth_method: env = API key, native = interactive login, auto = try env first
if [ "$ADDT_EXT_AUTH_AUTOLOGIN" = "true" ]; then
    method="${ADDT_EXT_AUTH_METHOD:-auto}"

    if [ "$method" = "env" ] || [ "$method" = "auto" ]; then
        if [ -n "$OPENAI_API_KEY" ]; then
            echo "Setup [codex]: Auto-logging in with API key"
            printenv OPENAI_API_KEY | codex login --with-api-key
        fi
    fi
fi
