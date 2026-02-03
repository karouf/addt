package cmd

import (
	"fmt"
	"os"

	"github.com/jedi4ever/addt/config"
)

// HandleConfigCommand handles the config subcommand
func HandleConfigCommand(args []string) {
	if len(args) == 0 {
		printConfigHelp()
		return
	}

	switch args[0] {
	case "global":
		handleGlobalConfig(args[1:])
	case "project":
		handleProjectConfig(args[1:])
	case "extension":
		handleExtensionConfig(args[1:])
	case "path":
		fmt.Printf("Global config:  %s\n", config.GetGlobalConfigPath())
		fmt.Printf("Project config: %s\n", config.GetProjectConfigPath())
	default:
		fmt.Printf("Unknown config command: %s\n", args[0])
		printConfigHelp()
		os.Exit(1)
	}
}

// handleGlobalConfig handles global config subcommands
func handleGlobalConfig(args []string) {
	if len(args) == 0 {
		printGlobalConfigHelp()
		return
	}

	switch args[0] {
	case "list":
		listConfig()
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: addt config global get <key>")
			os.Exit(1)
		}
		getConfig(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: addt config global set <key> <value>")
			os.Exit(1)
		}
		setConfig(args[1], args[2])
	case "unset":
		if len(args) < 2 {
			fmt.Println("Usage: addt config global unset <key>")
			os.Exit(1)
		}
		unsetConfig(args[1])
	default:
		fmt.Printf("Unknown global config command: %s\n", args[0])
		printGlobalConfigHelp()
		os.Exit(1)
	}
}

// handleProjectConfig handles project-level config subcommands
func handleProjectConfig(args []string) {
	if len(args) == 0 {
		printProjectConfigHelp()
		return
	}

	switch args[0] {
	case "list":
		listProjectConfig()
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: addt config project get <key>")
			os.Exit(1)
		}
		getProjectConfig(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: addt config project set <key> <value>")
			os.Exit(1)
		}
		setProjectConfig(args[1], args[2])
	case "unset":
		if len(args) < 2 {
			fmt.Println("Usage: addt config project unset <key>")
			os.Exit(1)
		}
		unsetProjectConfig(args[1])
	default:
		fmt.Printf("Unknown project config command: %s\n", args[0])
		printProjectConfigHelp()
		os.Exit(1)
	}
}

// handleExtensionConfig handles extension-specific config subcommands
func handleExtensionConfig(args []string) {
	if len(args) == 0 {
		printExtensionConfigHelp()
		return
	}

	// Check for --project flag anywhere in args
	useProject := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--project" {
			useProject = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	if len(args) == 0 {
		printExtensionConfigHelp()
		return
	}

	extName := args[0]

	// Check if first arg is a subcommand (user forgot extension name)
	if extName == "list" || extName == "get" || extName == "set" || extName == "unset" {
		fmt.Println("Error: extension name required")
		fmt.Println()
		printExtensionConfigHelp()
		os.Exit(1)
	}

	// Validate that the extension exists
	if !extensionExists(extName) {
		fmt.Printf("Error: extension '%s' does not exist\n", extName)
		fmt.Println("Run 'addt extensions list' to see available extensions")
		os.Exit(1)
	}

	if len(args) < 2 {
		// Default to list for extension
		listExtensionConfig(extName)
		return
	}

	switch args[1] {
	case "list":
		listExtensionConfig(extName)
	case "get":
		if len(args) < 3 {
			fmt.Println("Usage: addt config extension <name> get <key>")
			os.Exit(1)
		}
		getExtensionConfig(extName, args[2])
	case "set":
		if len(args) < 4 {
			fmt.Println("Usage: addt config extension <name> set <key> <value> [--project]")
			os.Exit(1)
		}
		setExtensionConfig(extName, args[2], args[3], useProject)
	case "unset":
		if len(args) < 3 {
			fmt.Println("Usage: addt config extension <name> unset <key> [--project]")
			os.Exit(1)
		}
		unsetExtensionConfig(extName, args[2], useProject)
	default:
		fmt.Printf("Unknown extension config command: %s\n", args[1])
		printExtensionConfigHelp()
		os.Exit(1)
	}
}

// extensionExists checks if an extension with the given name exists
func extensionExists(name string) bool {
	exts, err := getExtensions()
	if err != nil {
		return false
	}
	for _, ext := range exts {
		if ext.Name == name {
			return true
		}
	}
	return false
}

func printConfigHelp() {
	fmt.Println("Usage: addt config <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  global <subcommand>              Manage global configuration (~/.addt/config.yaml)")
	fmt.Println("  project <subcommand>             Manage project configuration (.addt.yaml)")
	fmt.Println("  extension <name> <subcommand>    Manage extension-specific configuration")
	fmt.Println("  path                             Show config file paths")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  addt config global list")
	fmt.Println("  addt config global set docker_cpus 2")
	fmt.Println("  addt config project set persistent true")
	fmt.Println("  addt config extension claude set version 1.0.5")
	fmt.Println()
	fmt.Println("Precedence (highest to lowest):")
	fmt.Println("  1. Environment variables (e.g., ADDT_DOCKER_CPUS)")
	fmt.Println("  2. Project config (.addt.yaml in current directory)")
	fmt.Println("  3. Global config (~/.addt/config.yaml)")
	fmt.Println("  4. Default values")
}

func printGlobalConfigHelp() {
	fmt.Println("Usage: addt config global <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List all global configuration values")
	fmt.Println("  get <key>         Get a configuration value")
	fmt.Println("  set <key> <value> Set a configuration value")
	fmt.Println("  unset <key>       Remove a configuration value (use default)")
	fmt.Println()
	fmt.Println("Available keys:")
	keys := getConfigKeys()
	maxKeyLen := 0
	for _, k := range keys {
		if len(k.Key) > maxKeyLen {
			maxKeyLen = len(k.Key)
		}
	}
	for _, k := range keys {
		fmt.Printf("  %-*s  %s\n", maxKeyLen, k.Key, k.Description)
	}
}

func printProjectConfigHelp() {
	fmt.Println("Usage: addt config project <command>")
	fmt.Println()
	fmt.Println("Manage project-level configuration stored in .addt.yaml")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List all project configuration values")
	fmt.Println("  get <key>         Get a configuration value")
	fmt.Println("  set <key> <value> Set a configuration value")
	fmt.Println("  unset <key>       Remove a configuration value")
	fmt.Println()
	fmt.Println("Project config overrides global config but is overridden by env vars.")
	fmt.Println()
	fmt.Println("Available keys:")
	keys := getConfigKeys()
	maxKeyLen := 0
	for _, k := range keys {
		if len(k.Key) > maxKeyLen {
			maxKeyLen = len(k.Key)
		}
	}
	for _, k := range keys {
		fmt.Printf("  %-*s  %s\n", maxKeyLen, k.Key, k.Description)
	}
}

func printExtensionConfigHelp() {
	fmt.Println("Usage: addt config extension <name> <command> [--project]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List extension configuration")
	fmt.Println("  get <key>         Get a configuration value")
	fmt.Println("  set <key> <value> Set a configuration value")
	fmt.Println("  unset <key>       Remove a configuration value")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --project         Save to project config (.addt.yaml) instead of global")
	fmt.Println()
	fmt.Println("Available keys:")
	fmt.Println("  version     Extension version (e.g., \"1.0.5\", \"latest\", \"stable\")")
	fmt.Println("  automount   Auto-mount extension config directories (true/false)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  addt config extension claude list")
	fmt.Println("  addt config extension claude set version 1.0.5")
	fmt.Println("  addt config extension claude set automount false --project")
}
