package extensions

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/extensions"
)

// ShowInfo displays detailed info about a specific extension
func ShowInfo(name string) {
	exts, err := extensions.GetExtensions()
	if err != nil {
		fmt.Printf("Error reading extensions: %v\n", err)
		os.Exit(1)
	}

	for _, ext := range exts {
		if ext.Name == name {
			version := ext.DefaultVersion
			if version == "" {
				version = "latest"
			}

			fmt.Printf("%s\n", ext.Name)
			fmt.Printf("%s\n\n", strings.Repeat("=", len(ext.Name)))

			fmt.Printf("  %s\n\n", ext.Description)

			source := "built-in"
			if ext.IsLocal {
				source = "local (~/.addt/extensions/" + ext.Name + ")"
			}

			fmt.Println("Configuration:")
			fmt.Printf("  Entrypoint:  %s\n", ext.Entrypoint)
			fmt.Printf("  Version:     %s\n", version)
			fmt.Printf("  Auto-mount:  %v\n", ext.AutoMount)
			fmt.Printf("  Source:      %s\n", source)

			if len(ext.Dependencies) > 0 {
				fmt.Printf("  Depends on:  %s\n", strings.Join(ext.Dependencies, ", "))
			}

			if len(ext.EnvVars) > 0 {
				fmt.Println("\nEnvironment Variables:")
				for _, env := range ext.EnvVars {
					fmt.Printf("  - %s\n", env)
				}
			}

			if len(ext.Mounts) > 0 {
				fmt.Println("\nMounts:")
				for _, m := range ext.Mounts {
					fmt.Printf("  - %s -> %s\n", m.Source, m.Target)
				}
			}

			if len(ext.Flags) > 0 {
				fmt.Println("\nFlags:")
				for _, f := range ext.Flags {
					fmt.Printf("  %-12s %s\n", f.Flag, f.Description)
				}
			}

			fmt.Println("\nUsage:")
			fmt.Printf("  addt run %s [args...]\n", ext.Name)
			return
		}
	}

	fmt.Printf("Extension not found: %s\n", name)
	fmt.Println("Run 'addt extensions list' to see available extensions")
	os.Exit(1)
}
