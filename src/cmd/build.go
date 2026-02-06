package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/provider"
)

// HandleBuildCommand handles the build command
func HandleBuildCommand(prov provider.Provider, cfg *provider.Config, args []string, noCache bool, rebuildBase bool) {
	// Parse --build-arg flags
	for i := 0; i < len(args); i++ {
		if args[i] == "--build-arg" && i+1 < len(args) {
			parts := strings.SplitN(args[i+1], "=", 2)
			if len(parts) == 2 {
				key, val := parts[0], parts[1]
				switch {
				case key == "ADDT_EXTENSIONS":
					cfg.Extensions = val
				case key == "NODE_VERSION":
					cfg.NodeVersion = val
				case key == "GO_VERSION":
					cfg.GoVersion = val
				case key == "UV_VERSION":
					cfg.UvVersion = val
				case strings.HasSuffix(key, "_VERSION"):
					// Per-extension versions (e.g., CLAUDE_VERSION, CODEX_VERSION)
					extName := strings.TrimSuffix(key, "_VERSION")
					extName = strings.ToLower(extName)
					if cfg.ExtensionVersions == nil {
						cfg.ExtensionVersions = make(map[string]string)
					}
					cfg.ExtensionVersions[extName] = val
				}
			}
			i++ // Skip next arg
		}
	}

	// Determine image name
	cfg.ImageName = prov.DetermineImageName()

	// Set no-cache flag
	cfg.NoCache = noCache

	// Always rebuild extension image when using build command
	// Base image is rebuilt if --rebuild-base flag is set
	if err := prov.BuildIfNeeded(true, rebuildBase); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printBuildHelp() {
	fmt.Println("Usage: addt build [options]")
	fmt.Println()
	fmt.Println("Build the container image for the configured extension(s).")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --no-cache              Build without using cache")
	fmt.Println("  --rebuild-base          Rebuild the base image before building extension image")
	fmt.Println("  --build-arg KEY=VALUE   Set build-time variables")
	fmt.Println()
	fmt.Println("Build arguments:")
	fmt.Println("  ADDT_EXTENSIONS         Comma-separated list of extensions")
	fmt.Println("  NODE_VERSION            Node.js version")
	fmt.Println("  GO_VERSION              Go version")
	fmt.Println("  UV_VERSION              UV Python version")
	fmt.Println("  <EXT>_VERSION           Version for specific extension")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  addt build")
	fmt.Println("  addt build --no-cache")
	fmt.Println("  addt build --rebuild-base")
	fmt.Println("  addt build --rebuild-base --no-cache")
	fmt.Println("  addt build --build-arg ADDT_EXTENSIONS=claude,codex")
	fmt.Println("  addt build --build-arg CLAUDE_VERSION=1.0.5")
}
