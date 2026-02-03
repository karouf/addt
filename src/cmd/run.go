package cmd

import (
	"fmt"
	"os"

	extcmd "github.com/jedi4ever/addt/cmd/extensions"
)

// HandleRunCommand handles the "addt run <extension>" command.
// It validates the extension and sets up environment variables.
// Returns the remaining args for execution, or nil if the command was fully handled (help/error).
func HandleRunCommand(args []string) []string {
	if len(args) < 1 {
		printRunHelp()
		return nil
	}

	extName := args[0]

	// Check for help flag
	if extName == "--help" || extName == "-h" {
		printRunHelp()
		return nil
	}

	// Validate extension exists
	if !extcmd.Exists(extName) {
		fmt.Printf("Error: extension '%s' does not exist\n", extName)
		fmt.Println("Run 'addt extensions list' to see available extensions")
		os.Exit(1)
	}

	// Set the extension environment variables
	os.Setenv("ADDT_EXTENSIONS", extName)
	os.Setenv("ADDT_COMMAND", extcmd.GetEntrypoint(extName))

	// Return remaining args for execution
	if len(args) > 1 {
		return args[1:]
	}
	return []string{}
}

func printRunHelp() {
	fmt.Println("Usage: addt run <extension> [args...]")
	fmt.Println()
	fmt.Println("Run a specific extension in a container.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <extension>    Name of the extension to run")
	fmt.Println("  [args...]      Arguments to pass to the extension")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  addt run claude \"Fix the bug\"")
	fmt.Println("  addt run codex --help")
	fmt.Println("  addt run gemini")
	fmt.Println()
	fmt.Println("To see available extensions:")
	fmt.Println("  addt extensions list")
}
