package extensions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedi4ever/addt/extensions"
)

// Create creates a new local extension with template files
func Create(name string) {
	// Validate name
	if name == "" {
		fmt.Println("Error: extension name cannot be empty")
		os.Exit(1)
	}

	// Check if name contains only valid characters
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			fmt.Printf("Error: extension name can only contain lowercase letters, numbers, hyphens, and underscores\n")
			os.Exit(1)
		}
	}

	// Get local extensions directory
	localDir := extensions.GetLocalExtensionsDir()
	if localDir == "" {
		fmt.Println("Error: could not determine local extensions directory")
		os.Exit(1)
	}

	extDir := filepath.Join(localDir, name)

	// Check if extension already exists
	if _, err := os.Stat(extDir); err == nil {
		fmt.Printf("Error: extension '%s' already exists at %s\n", name, extDir)
		os.Exit(1)
	}

	// Create extension directory
	if err := os.MkdirAll(extDir, 0755); err != nil {
		fmt.Printf("Error: failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Create config.yaml
	configContent := fmt.Sprintf(`name: %s
description: Description of your extension
entrypoint: %s
default_version: latest
auto_mount: false
dependencies: []
env_vars: []
mounts: []
flags: []
`, name, name)

	if err := os.WriteFile(filepath.Join(extDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		fmt.Printf("Error: failed to create config.yaml: %v\n", err)
		os.Exit(1)
	}

	// Create install.sh
	installContent := fmt.Sprintf(`#!/bin/bash
# %s - Installation script
# This script runs during 'addt build %s'

set -e

echo "Extension [%s]: Installing..."

# Get version from environment or default to latest
VERSION="${%s_VERSION:-latest}"

# TODO: Add your installation commands here
# Examples:
#   sudo npm install -g your-package
#   pip install your-package
#   go install github.com/your/package@latest

echo "Extension [%s]: Done."
`, name, name, name, strings.ToUpper(strings.ReplaceAll(name, "-", "_")), name)

	if err := os.WriteFile(filepath.Join(extDir, "install.sh"), []byte(installContent), 0755); err != nil {
		fmt.Printf("Error: failed to create install.sh: %v\n", err)
		os.Exit(1)
	}

	// Create setup.sh
	setupContent := fmt.Sprintf(`#!/bin/bash
# %s - Setup script
# This script runs at container startup before the entrypoint

echo "Setup [%s]: Initializing environment"

# TODO: Add any runtime setup commands here
# Examples:
#   export MY_VAR="value"
#   source ~/.my-config
`, name, name)

	if err := os.WriteFile(filepath.Join(extDir, "setup.sh"), []byte(setupContent), 0755); err != nil {
		fmt.Printf("Error: failed to create setup.sh: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created extension '%s' at:\n", name)
	fmt.Printf("  %s\n", extDir)
	fmt.Println()
	fmt.Println("Files created:")
	fmt.Println("  config.yaml  - Extension configuration")
	fmt.Println("  install.sh   - Installation script (runs during build)")
	fmt.Println("  setup.sh     - Setup script (runs at container start)")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s/config.yaml to configure your extension\n", extDir)
	fmt.Printf("  2. Edit %s/install.sh to add installation commands\n", extDir)
	fmt.Printf("  3. Build with: addt build %s\n", name)
	fmt.Printf("  4. Run with:   addt run %s\n", name)
}
