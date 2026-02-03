package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/config"
)

func listConfig() {
	globalCfg, err := config.LoadGlobalConfigFile()
	if err != nil {
		fmt.Printf("Error loading global config: %v\n", err)
		os.Exit(1)
	}

	projectCfg, err := config.LoadProjectConfigFile()
	if err != nil {
		fmt.Printf("Error loading project config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Global config:  %s\n", config.GetGlobalConfigPath())
	fmt.Printf("Project config: %s\n\n", config.GetProjectConfigPath())

	keys := getConfigKeys()

	// Calculate column widths based on content
	maxKeyLen := 3 // "Key"
	maxValLen := 5 // "Value"
	for _, k := range keys {
		if len(k.Key) > maxKeyLen {
			maxKeyLen = len(k.Key)
		}
		envValue := os.Getenv(k.EnvVar)
		projectValue := getConfigValue(projectCfg, k.Key)
		globalValue := getConfigValue(globalCfg, k.Key)
		defaultValue := getDefaultValue(k.Key)
		val := envValue
		if val == "" {
			val = projectValue
		}
		if val == "" {
			val = globalValue
		}
		if val == "" {
			val = defaultValue
		}
		if val == "" {
			val = "-"
		}
		if len(val) > maxValLen {
			maxValLen = len(val)
		}
	}

	// Print header
	fmt.Printf("  %-*s   %-*s   %s\n", maxKeyLen, "Key", maxValLen, "Value", "Source")
	fmt.Printf("  %s   %s   %s\n", strings.Repeat("-", maxKeyLen), strings.Repeat("-", maxValLen), "--------")

	for _, k := range keys {
		envValue := os.Getenv(k.EnvVar)
		projectValue := getConfigValue(projectCfg, k.Key)
		globalValue := getConfigValue(globalCfg, k.Key)
		defaultValue := getDefaultValue(k.Key)

		var displayValue, source string
		if envValue != "" {
			displayValue = envValue
			source = "env"
		} else if projectValue != "" {
			displayValue = projectValue
			source = "project"
		} else if globalValue != "" {
			displayValue = globalValue
			source = "global"
		} else if defaultValue != "" {
			displayValue = defaultValue
			source = "default"
		} else {
			displayValue = "-"
			source = ""
		}

		// Highlight non-default values
		if source == "env" || source == "project" || source == "global" {
			fmt.Printf("* %-*s   %-*s   %s\n", maxKeyLen, k.Key, maxValLen, displayValue, source)
		} else {
			fmt.Printf("  %-*s   %-*s   %s\n", maxKeyLen, k.Key, maxValLen, displayValue, source)
		}
	}
}

func getConfig(key string) {
	// Validate key
	if !isValidConfigKey(key) {
		fmt.Printf("Unknown config key: %s\n", key)
		fmt.Println("Use 'addt config list' to see available keys.")
		os.Exit(1)
	}

	cfg, err := config.LoadGlobalConfigFile()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	val := getConfigValue(cfg, key)
	if val == "" {
		fmt.Printf("%s is not set\n", key)
	} else {
		fmt.Println(val)
	}
}

func setConfig(key, value string) {
	// Validate key
	keyInfo := getConfigKeyInfo(key)
	if keyInfo == nil {
		fmt.Printf("Unknown config key: %s\n", key)
		fmt.Println("Use 'addt config --help' to see available keys.")
		os.Exit(1)
	}

	// Validate value based on type
	if keyInfo.Type == "bool" {
		value = strings.ToLower(value)
		if value != "true" && value != "false" {
			fmt.Printf("Invalid value for %s: must be 'true' or 'false'\n", key)
			os.Exit(1)
		}
	}

	cfg, err := config.LoadGlobalConfigFile()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	setConfigValue(cfg, key, value)

	if err := config.SaveGlobalConfigFile(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Set %s = %s\n", key, value)
}

func unsetConfig(key string) {
	// Validate key
	if !isValidConfigKey(key) {
		fmt.Printf("Unknown config key: %s\n", key)
		fmt.Println("Use 'addt config list' to see available keys.")
		os.Exit(1)
	}

	cfg, err := config.LoadGlobalConfigFile()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	unsetConfigValue(cfg, key)

	if err := config.SaveGlobalConfigFile(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Unset %s\n", key)
}
