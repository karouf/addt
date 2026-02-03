package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/extensions"
)

// PrintVersion prints addt version and loaded extension version
func PrintVersion(version, defaultNodeVersion, defaultGoVersion, defaultUvVersion string) {
	fmt.Printf("addt %s\n", version)
	fmt.Println()

	// Default tool versions
	fmt.Println("Tools:")
	fmt.Printf("  Node.js:  %s\n", defaultNodeVersion)
	fmt.Printf("  Go:       %s\n", defaultGoVersion)
	fmt.Printf("  UV:       %s\n", defaultUvVersion)
	fmt.Println()

	// Get loaded extension (from env or binary name)
	extName := os.Getenv("ADDT_EXTENSIONS")
	if extName == "" {
		extName = os.Getenv("ADDT_COMMAND")
	}
	// No default - extension must be explicitly set via symlink or env
	// Take first extension if comma-separated
	if idx := strings.Index(extName, ","); idx != -1 {
		extName = extName[:idx]
	}

	// Get version for this extension
	extVersion := os.Getenv("ADDT_" + strings.ToUpper(extName) + "_VERSION")
	if extVersion == "" {
		// Look up default version from config
		exts, err := extensions.GetExtensions()
		if err == nil {
			for _, ext := range exts {
				if ext.Name == extName {
					extVersion = ext.DefaultVersion
					break
				}
			}
		}
	}
	if extVersion == "" {
		extVersion = "latest"
	}

	fmt.Println("Extension:")
	fmt.Printf("  %-16s %s\n", extName, extVersion)
}
