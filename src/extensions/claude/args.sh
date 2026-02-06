#!/bin/bash
# Claude Code argument transformer
# Transforms generic addt args to Claude-specific args

ARGS=()

#??
#   "bypassPermissionsModeAccepted": true,
# IS_SANDBOX=1 - for sandbox mode

while [[ $# -gt 0 ]]; do
    case "$1" in
        --yolo)
            # Transform generic --yolo to Claude's flag
            ARGS+=(--dangerously-skip-permissions)
            shift
            ;;
        *)
            ARGS+=("$1")
            shift
            ;;
    esac
done

# If ADDT_EXTENSION_CLAUDE_YOLO is set via config/env and --dangerously-skip-permissions
# wasn't already added by a --yolo CLI flag, inject it now
if [ "${ADDT_EXTENSION_CLAUDE_YOLO}" = "true" ]; then
    already_set=false
    for arg in "${ARGS[@]}"; do
        if [ "$arg" = "--dangerously-skip-permissions" ]; then
            already_set=true
            break
        fi
    done
    if [ "$already_set" = "false" ]; then
        ARGS+=(--dangerously-skip-permissions)
    fi
fi

# Add system prompt if set (for port mappings, etc.)
if [ -n "$ADDT_SYSTEM_PROMPT" ]; then
    ARGS+=(--append-system-prompt "$ADDT_SYSTEM_PROMPT")
fi

# Output transformed args (one per line for proper handling)
printf '%s\n' "${ARGS[@]}"
