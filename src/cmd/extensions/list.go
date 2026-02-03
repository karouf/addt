package extensions

import (
	"fmt"
	"strings"

	"github.com/jedi4ever/addt/extensions"
)

// List prints all available extensions
func List() {
	exts, err := extensions.GetExtensions()
	if err != nil {
		fmt.Printf("Error reading extensions: %v\n", err)
		return
	}

	// Find max lengths for alignment
	maxName := 4   // "Name"
	maxEntry := 10 // "Entrypoint"
	maxVer := 7    // "Version"
	for _, ext := range exts {
		if len(ext.Name) > maxName {
			maxName = len(ext.Name)
		}
		if len(ext.Entrypoint) > maxEntry {
			maxEntry = len(ext.Entrypoint)
		}
		ver := ext.DefaultVersion
		if ver == "" {
			ver = "latest"
		}
		if len(ver) > maxVer {
			maxVer = len(ver)
		}
	}

	// Print header
	fmt.Printf("  #  %-*s  %-*s  %-*s  %-6s  %s\n", maxName, "Name", maxEntry, "Entrypoint", maxVer, "Version", "Source", "Description")
	fmt.Printf("  -  %-*s  %-*s  %-*s  %-6s  %s\n", maxName, strings.Repeat("-", maxName), maxEntry, strings.Repeat("-", maxEntry), maxVer, strings.Repeat("-", maxVer), "------", "-----------")

	// Print rows
	for i, ext := range exts {
		version := ext.DefaultVersion
		if version == "" {
			version = "latest"
		}
		source := "built-in"
		if ext.IsLocal {
			source = "local"
		}
		fmt.Printf("%3d  %-*s  %-*s  %-*s  %-6s  %s\n", i+1, maxName, ext.Name, maxEntry, ext.Entrypoint, maxVer, version, source, ext.Description)
	}

	// Show local extensions directory info
	localDir := extensions.GetLocalExtensionsDir()
	if localDir != "" {
		fmt.Printf("\nLocal extensions directory: %s\n", localDir)
	}
}
