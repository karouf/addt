package cmd

import (
	"fmt"
	"os"
)

// handleCliCommand handles the "cli" subcommand
func handleCliCommand(args []string, version string) {
	if len(args) == 0 {
		fmt.Println("Usage: addt cli <command>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  update    Install addt updates")
		return
	}
	switch args[0] {
	case "update":
		UpdateAddt(version)
	default:
		fmt.Printf("Unknown cli command: %s\n", args[0])
		os.Exit(1)
	}
}

// printAddtSubcommandUsage prints the "addt" namespace usage (when invoked via agent)
func printAddtSubcommandUsage() {
	fmt.Println("Usage: <agent> addt <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build [--build-arg ...]   Build the container image")
	fmt.Println("  shell                     Open bash shell in container")
	fmt.Println("  containers <subcommand>   Manage containers (list, stop, rm, clean)")
	fmt.Println("  firewall <subcommand>     Manage firewall (list, add, remove, reset)")
	fmt.Println("  extensions <subcommand>   Manage extensions (list, info, new)")
	fmt.Println("  config <subcommand>       Manage config (global, project, extension)")
	fmt.Println("  cli <subcommand>          Manage addt CLI (update)")
	fmt.Println("  version                   Show version info")
}
