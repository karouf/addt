#!/bin/bash
# Extension installer for DClaude
# Usage: install.sh [extension1,extension2,...]
#
# Extensions are directories containing:
#   - config.yaml  - metadata (name, description, entrypoint, dependencies)
#   - install.sh   - install script

set -e

EXTENSIONS_DIR="${EXTENSIONS_DIR:-/usr/local/share/dclaude/extensions}"
METADATA_FILE="${METADATA_FILE:-/home/claude/.dclaude/extensions.json}"
EXTENSIONS="${1:-$DCLAUDE_EXTENSIONS}"

# Ensure metadata directory exists
mkdir -p "$(dirname "$METADATA_FILE")"

if [ -z "$EXTENSIONS" ] || [ "$EXTENSIONS" = "none" ]; then
    echo "Extensions: No extensions requested"
    echo '{"extensions":{}}' > "$METADATA_FILE"
    exit 0
fi

# Simple YAML parser - extract value for a key
yaml_get() {
    local file="$1"
    local key="$2"
    grep "^${key}:" "$file" 2>/dev/null | sed "s/^${key}:[[:space:]]*//" | tr -d '"' || echo ""
}

# Parse dependencies from YAML (handles both [] and list format)
yaml_get_deps() {
    local file="$1"
    local in_deps=false
    local deps=""

    while IFS= read -r line; do
        if [[ "$line" =~ ^dependencies: ]]; then
            # Check for inline empty array
            if [[ "$line" =~ \[\] ]]; then
                echo ""
                return
            fi
            in_deps=true
            continue
        fi
        if $in_deps; then
            # Stop if we hit another top-level key
            if [[ "$line" =~ ^[a-z] ]] && [[ ! "$line" =~ ^[[:space:]] ]]; then
                break
            fi
            # Extract dependency name (- item format)
            if [[ "$line" =~ ^[[:space:]]*-[[:space:]]*(.+) ]]; then
                dep="${BASH_REMATCH[1]}"
                deps="$deps $dep"
            fi
        fi
    done < "$file"

    echo "$deps" | xargs
}

# Build installation order with dependencies
declare -A installed
declare -a install_order

resolve_extension() {
    local ext="$1"

    # Skip if already processed
    if [ "${installed[$ext]}" = "1" ]; then
        return
    fi

    local ext_dir="$EXTENSIONS_DIR/$ext"
    local config="$ext_dir/config.yaml"
    local script="$ext_dir/install.sh"

    if [ ! -d "$ext_dir" ]; then
        echo "Extensions: Warning - extension '$ext' not found (no $ext/ directory), skipping"
        return
    fi

    if [ ! -f "$config" ]; then
        echo "Extensions: Warning - extension '$ext' has no config.yaml, skipping"
        return
    fi

    if [ ! -f "$script" ]; then
        echo "Extensions: Warning - extension '$ext' has no install.sh, skipping"
        return
    fi

    # Process dependencies first
    local deps=$(yaml_get_deps "$config")
    for dep in $deps; do
        resolve_extension "$dep"
    done

    # Add to install order
    install_order+=("$ext")
    installed[$ext]=1
}

echo "Extensions: Resolving dependencies..."

# Parse comma-separated extension list
IFS=',' read -ra EXT_ARRAY <<< "$EXTENSIONS"
for ext in "${EXT_ARRAY[@]}"; do
    ext=$(echo "$ext" | xargs)  # trim whitespace
    [ -n "$ext" ] && resolve_extension "$ext"
done

if [ ${#install_order[@]} -eq 0 ]; then
    echo "Extensions: No valid extensions to install"
    echo '{"extensions":{}}' > "$METADATA_FILE"
    exit 0
fi

echo "Extensions: Installation order: ${install_order[*]}"

# Install each extension
for ext in "${install_order[@]}"; do
    ext_dir="$EXTENSIONS_DIR/$ext"
    config="$ext_dir/config.yaml"
    script="$ext_dir/install.sh"

    description=$(yaml_get "$config" "description")

    echo "=========================================="
    echo "Extensions: Installing '$ext'"
    [ -n "$description" ] && echo "  $description"
    echo "=========================================="
    bash "$script"
done

# Write metadata JSON
echo "Extensions: Writing metadata to $METADATA_FILE"
{
    echo '{"extensions":{'
    first=true
    for ext in "${install_order[@]}"; do
        config="$EXTENSIONS_DIR/$ext/config.yaml"
        name=$(yaml_get "$config" "name")
        description=$(yaml_get "$config" "description")
        entrypoint=$(yaml_get "$config" "entrypoint")

        [ "$first" = true ] && first=false || echo ","
        printf '"%s":{"name":"%s","description":"%s","entrypoint":"%s"}' \
            "$ext" "$name" "$description" "$entrypoint"
    done
    echo '}}'
} > "$METADATA_FILE"

echo "=========================================="
echo "Extensions: All extensions installed successfully"
echo "Installed: ${install_order[*]}"
echo "=========================================="
