package extensions

import (
	"fmt"
	"os"

	configcmd "github.com/jedi4ever/addt/cmd/config"
)

// HandleCommand handles the "extensions" subcommand
func HandleCommand(args []string) {
	if len(args) == 0 {
		printUsage("")
		return
	}
	switch args[0] {
	case "list":
		List()
	case "info":
		if len(args) < 2 {
			fmt.Println("Usage: addt extensions info <name>")
			os.Exit(1)
		}
		ShowInfo(args[1])
	case "new":
		if len(args) < 2 {
			fmt.Println("Usage: addt extensions new <name>")
			os.Exit(1)
		}
		Create(args[1])
	case "clone":
		if len(args) < 2 {
			fmt.Println("Usage: addt extensions clone <source> [target]")
			os.Exit(1)
		}
		targetName := ""
		if len(args) > 2 {
			targetName = args[2]
		}
		Clone(args[1], targetName)
	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: addt extensions remove <name> [--force]")
			os.Exit(1)
		}
		force := len(args) > 2 && args[2] == "--force"
		Remove(args[1], force)
	case "config":
		handleConfigCommand(args[1:], "addt")
	default:
		fmt.Printf("Unknown extensions command: %s\n", args[0])
		os.Exit(1)
	}
}

// HandleCommandAgent handles the "extensions" subcommand when invoked via agent (e.g., "claude addt extensions")
func HandleCommandAgent(args []string) {
	if len(args) == 0 {
		printUsage("<agent>")
		return
	}
	switch args[0] {
	case "list":
		List()
	case "info":
		if len(args) < 2 {
			fmt.Println("Usage: <agent> addt extensions info <name>")
			os.Exit(1)
		}
		ShowInfo(args[1])
	case "new":
		if len(args) < 2 {
			fmt.Println("Usage: <agent> addt extensions new <name>")
			os.Exit(1)
		}
		Create(args[1])
	case "clone":
		if len(args) < 2 {
			fmt.Println("Usage: <agent> addt extensions clone <source> [target]")
			os.Exit(1)
		}
		targetName := ""
		if len(args) > 2 {
			targetName = args[2]
		}
		Clone(args[1], targetName)
	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: <agent> addt extensions remove <name> [--force]")
			os.Exit(1)
		}
		force := len(args) > 2 && args[2] == "--force"
		Remove(args[1], force)
	case "config":
		handleConfigCommand(args[1:], "<agent>")
	default:
		fmt.Printf("Unknown extensions command: %s\n", args[0])
		os.Exit(1)
	}
}

// handleConfigCommand handles the "extensions config" subcommand
func handleConfigCommand(args []string, prefix string) {
	if len(args) < 1 || args[0] == "--help" || args[0] == "-h" {
		fmt.Printf("Usage: %s addt extensions config <name> <command>\n", prefix)
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  list              List extension configuration")
		fmt.Println("  get <key>         Get a configuration value")
		fmt.Println("  set <key> <value> Set a configuration value")
		fmt.Println("  unset <key>       Remove a configuration value")
		fmt.Println()
		fmt.Println("Available keys:")
		fmt.Println("  version     Extension version (e.g., \"1.0.5\", \"latest\", \"stable\")")
		fmt.Println("  config.automount   Auto-mount extension config directories (true/false)")
		fmt.Println()
		fmt.Println("Examples:")
		if prefix == "<agent>" {
			fmt.Println("  claude addt extensions config claude list")
			fmt.Println("  claude addt extensions config claude set version 1.0.5")
		} else {
			fmt.Println("  addt extensions config claude list")
			fmt.Println("  addt extensions config claude set version 1.0.5")
		}
		return
	}
	// Delegate to HandleConfigCommand with "extension" prefix
	configcmd.HandleCommand(append([]string{"extension"}, args...))
}

// printUsage prints the extensions subcommand usage
func printUsage(prefix string) {
	if prefix == "" {
		prefix = "addt"
	}
	fmt.Printf("Usage: %s extensions <command>\n", prefix)
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                       List available extensions")
	fmt.Println("  info <name>                Show extension details")
	fmt.Println("  new <name>                 Create a new local extension")
	fmt.Println("  clone <source> [target]    Copy built-in extension for customization")
	fmt.Println("  remove <name> [--force]    Remove a local extension")
	fmt.Println("  config <name> <subcommand> Configure extension settings")
}
