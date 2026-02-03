package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/config"
)

func listProjectConfig() {
	cfg, err := config.LoadProjectConfigFile()
	if err != nil {
		fmt.Printf("Error loading project config: %v\n", err)
		os.Exit(1)
	}

	configPath := config.GetProjectConfigPath()
	fmt.Printf("Project config: %s\n\n", configPath)

	keys := getConfigKeys()

	// Calculate column widths
	maxKeyLen := 3
	maxValLen := 5
	for _, k := range keys {
		if len(k.Key) > maxKeyLen {
			maxKeyLen = len(k.Key)
		}
		val := getConfigValue(cfg, k.Key)
		if val == "" {
			val = "-"
		}
		if len(val) > maxValLen {
			maxValLen = len(val)
		}
	}

	// Print header
	fmt.Printf("  %-*s   %-*s\n", maxKeyLen, "Key", maxValLen, "Value")
	fmt.Printf("  %s   %s\n", strings.Repeat("-", maxKeyLen), strings.Repeat("-", maxValLen))

	hasValues := false
	for _, k := range keys {
		val := getConfigValue(cfg, k.Key)
		if val != "" {
			hasValues = true
			fmt.Printf("* %-*s   %-*s\n", maxKeyLen, k.Key, maxValLen, val)
		}
	}

	if !hasValues {
		fmt.Println("  (no project config set)")
	}
}

func getProjectConfig(key string) {
	if !isValidConfigKey(key) {
		fmt.Printf("Unknown config key: %s\n", key)
		fmt.Println("Use 'addt config project list' to see available keys.")
		os.Exit(1)
	}

	cfg, err := config.LoadProjectConfigFile()
	if err != nil {
		fmt.Printf("Error loading project config: %v\n", err)
		os.Exit(1)
	}

	val := getConfigValue(cfg, key)
	if val == "" {
		fmt.Printf("%s is not set in project config\n", key)
	} else {
		fmt.Println(val)
	}
}

func setProjectConfig(key, value string) {
	keyInfo := getConfigKeyInfo(key)
	if keyInfo == nil {
		fmt.Printf("Unknown config key: %s\n", key)
		fmt.Println("Use 'addt config project --help' to see available keys.")
		os.Exit(1)
	}

	if keyInfo.Type == "bool" {
		value = strings.ToLower(value)
		if value != "true" && value != "false" {
			fmt.Printf("Invalid value for %s: must be 'true' or 'false'\n", key)
			os.Exit(1)
		}
	}

	cfg, err := config.LoadProjectConfigFile()
	if err != nil {
		fmt.Printf("Error loading project config: %v\n", err)
		os.Exit(1)
	}

	setConfigValue(cfg, key, value)

	if err := config.SaveProjectConfigFile(cfg); err != nil {
		fmt.Printf("Error saving project config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Set %s = %s (project)\n", key, value)
}

func unsetProjectConfig(key string) {
	if !isValidConfigKey(key) {
		fmt.Printf("Unknown config key: %s\n", key)
		fmt.Println("Use 'addt config project list' to see available keys.")
		os.Exit(1)
	}

	cfg, err := config.LoadProjectConfigFile()
	if err != nil {
		fmt.Printf("Error loading project config: %v\n", err)
		os.Exit(1)
	}

	unsetConfigValue(cfg, key)

	if err := config.SaveProjectConfigFile(cfg); err != nil {
		fmt.Printf("Error saving project config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Unset %s (project)\n", key)
}
