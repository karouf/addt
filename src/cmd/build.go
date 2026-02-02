package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/dclaude/config"
	"github.com/jedi4ever/dclaude/provider"
)

// HandleBuildCommand handles the "build" command
func HandleBuildCommand(args []string, defaultNodeVersion, defaultGoVersion, defaultUvVersion string, defaultPortRangeStart int) {
	// Load base config
	cfg := config.LoadConfig(defaultNodeVersion, defaultGoVersion, defaultUvVersion, defaultPortRangeStart)

	// Parse --build-arg flags
	buildArgs := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if args[i] == "--build-arg" && i+1 < len(args) {
			parts := strings.SplitN(args[i+1], "=", 2)
			if len(parts) == 2 {
				buildArgs[parts[0]] = parts[1]
			}
			i++ // Skip next arg
		}
	}

	// Apply build args to config
	if ext, ok := buildArgs["DCLAUDE_EXTENSIONS"]; ok {
		cfg.Extensions = ext
	}
	if ver, ok := buildArgs["CLAUDE_VERSION"]; ok {
		cfg.ClaudeVersion = ver
	}
	if ver, ok := buildArgs["NODE_VERSION"]; ok {
		cfg.NodeVersion = ver
	}
	if ver, ok := buildArgs["GO_VERSION"]; ok {
		cfg.GoVersion = ver
	}
	if ver, ok := buildArgs["UV_VERSION"]; ok {
		cfg.UvVersion = ver
	}

	// Convert to provider config
	providerCfg := &provider.Config{
		ClaudeVersion: cfg.ClaudeVersion,
		NodeVersion:   cfg.NodeVersion,
		GoVersion:     cfg.GoVersion,
		UvVersion:     cfg.UvVersion,
		Provider:      cfg.Provider,
		Extensions:    cfg.Extensions,
	}

	// Create provider
	prov, err := NewProvider(cfg.Provider, providerCfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize provider
	if err := prov.Initialize(providerCfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Determine image name
	providerCfg.ImageName = prov.DetermineImageName()

	// Always rebuild when using build command
	if err := prov.BuildIfNeeded(true); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
