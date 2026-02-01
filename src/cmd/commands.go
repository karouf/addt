package cmd

import (
	"fmt"
	"os"

	"github.com/jedi4ever/dclaude/provider"
)

// HandleContainersCommand handles the containers subcommand using a provider
func HandleContainersCommand(prov provider.Provider, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	action := args[0]
	switch action {
	case "list", "ls":
		envs, err := prov.List()
		if err != nil {
			fmt.Printf("Error listing environments: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Persistent %s environments:\n", prov.GetName())
		fmt.Println("NAME\t\t\t\tSTATUS\t\tCREATED")
		for _, env := range envs {
			fmt.Printf("%s\t%s\t%s\n", env.Name, env.Status, env.CreatedAt)
		}
	case "stop":
		if len(args) < 2 {
			fmt.Println("Usage: dclaude containers stop <name>")
			os.Exit(1)
		}
		if err := prov.Stop(args[1]); err != nil {
			fmt.Printf("Error stopping environment: %v\n", err)
			os.Exit(1)
		}
	case "remove", "rm":
		if len(args) < 2 {
			fmt.Println("Usage: dclaude containers remove <name>")
			os.Exit(1)
		}
		if err := prov.Remove(args[1]); err != nil {
			fmt.Printf("Error removing environment: %v\n", err)
			os.Exit(1)
		}
	case "clean":
		envs, err := prov.List()
		if err != nil {
			fmt.Printf("Error listing environments: %v\n", err)
			os.Exit(1)
		}
		if len(envs) == 0 {
			fmt.Println("No persistent environments found")
			return
		}
		fmt.Println("Removing all persistent environments...")
		for _, env := range envs {
			fmt.Println(env.Name)
			prov.Remove(env.Name)
		}
		fmt.Println("✓ Cleaned")
	default:
		fmt.Println(`Usage: dclaude containers [list|stop|remove|clean]

Commands:
  list, ls    - List all persistent environments
  stop <name> - Stop a persistent environment
  remove <name> - Remove a persistent environment
  clean       - Remove all persistent environments`)
		os.Exit(1)
	}
}
