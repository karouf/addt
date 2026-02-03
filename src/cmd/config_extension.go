package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/config"
	"github.com/jedi4ever/addt/extensions"
)

func listExtensionConfig(extName string) {
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

	// Get extension defaults from extension's config.yaml
	var extDefaults *extensions.ExtensionConfig
	exts, err := getExtensions()
	if err == nil {
		for _, ext := range exts {
			if ext.Name == extName {
				extDefaults = &ext
				break
			}
		}
	}

	extNameUpper := strings.ToUpper(extName)
	fmt.Printf("Extension: %s\n\n", extName)

	keys := getExtensionConfigKeys()

	// Get extension config from global and project config files
	var globalExtCfg, projectExtCfg *config.ExtensionSettings
	if globalCfg.Extensions != nil {
		globalExtCfg = globalCfg.Extensions[extName]
	}
	if projectCfg.Extensions != nil {
		projectExtCfg = projectCfg.Extensions[extName]
	}

	// Print header
	fmt.Printf("  %-10s   %-15s   %s\n", "Key", "Value", "Source")
	fmt.Printf("  %s   %s   %s\n", strings.Repeat("-", 10), strings.Repeat("-", 15), "--------")

	for _, k := range keys {
		envVar := fmt.Sprintf(k.EnvVar, extNameUpper)
		envValue := os.Getenv(envVar)

		var projectValue, globalValue, defaultValue string

		// Get project config value
		if projectExtCfg != nil {
			switch k.Key {
			case "version":
				projectValue = projectExtCfg.Version
			case "automount":
				if projectExtCfg.Automount != nil {
					projectValue = fmt.Sprintf("%v", *projectExtCfg.Automount)
				}
			}
		}

		// Get global config value
		if globalExtCfg != nil {
			switch k.Key {
			case "version":
				globalValue = globalExtCfg.Version
			case "automount":
				if globalExtCfg.Automount != nil {
					globalValue = fmt.Sprintf("%v", *globalExtCfg.Automount)
				}
			}
		}

		// Get extension default value
		if extDefaults != nil {
			switch k.Key {
			case "version":
				defaultValue = extDefaults.DefaultVersion
			case "automount":
				defaultValue = fmt.Sprintf("%v", extDefaults.AutoMount)
			}
		}

		// Determine effective value and source (env > project > global > default)
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

		if source == "env" || source == "project" || source == "global" {
			fmt.Printf("* %-10s   %-15s   %s\n", k.Key, displayValue, source)
		} else {
			fmt.Printf("  %-10s   %-15s   %s\n", k.Key, displayValue, source)
		}
	}
}

func getExtensionConfig(extName, key string) {
	if !isValidExtensionConfigKey(key) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Println("Available keys: version, automount")
		os.Exit(1)
	}

	cfg, err := config.LoadGlobalConfigFile()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var extCfg *config.ExtensionSettings
	if cfg.Extensions != nil {
		extCfg = cfg.Extensions[extName]
	}

	if extCfg == nil {
		fmt.Printf("%s is not set\n", key)
		return
	}

	var val string
	switch key {
	case "version":
		val = extCfg.Version
	case "automount":
		if extCfg.Automount != nil {
			val = fmt.Sprintf("%v", *extCfg.Automount)
		}
	}

	if val == "" {
		fmt.Printf("%s is not set\n", key)
	} else {
		fmt.Println(val)
	}
}

func setExtensionConfig(extName, key, value string, useProject bool) {
	if !isValidExtensionConfigKey(key) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Println("Available keys: version, automount")
		os.Exit(1)
	}

	// Validate bool values
	if key == "automount" {
		value = strings.ToLower(value)
		if value != "true" && value != "false" {
			fmt.Printf("Invalid value for %s: must be 'true' or 'false'\n", key)
			os.Exit(1)
		}
	}

	var cfg *config.GlobalConfig
	var err error
	if useProject {
		cfg, err = config.LoadProjectConfigFile()
	} else {
		cfg, err = config.LoadGlobalConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize extensions map if needed
	if cfg.Extensions == nil {
		cfg.Extensions = make(map[string]*config.ExtensionSettings)
	}

	// Initialize extension config if needed
	if cfg.Extensions[extName] == nil {
		cfg.Extensions[extName] = &config.ExtensionSettings{}
	}

	extCfg := cfg.Extensions[extName]
	switch key {
	case "version":
		extCfg.Version = value
	case "automount":
		b := value == "true"
		extCfg.Automount = &b
	}

	if useProject {
		if err := config.SaveProjectConfigFile(cfg); err != nil {
			fmt.Printf("Error saving project config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s.%s = %s (project)\n", extName, key, value)
	} else {
		if err := config.SaveGlobalConfigFile(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s.%s = %s\n", extName, key, value)
	}
}

func unsetExtensionConfig(extName, key string, useProject bool) {
	if !isValidExtensionConfigKey(key) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Println("Available keys: version, automount")
		os.Exit(1)
	}

	var cfg *config.GlobalConfig
	var err error
	if useProject {
		cfg, err = config.LoadProjectConfigFile()
	} else {
		cfg, err = config.LoadGlobalConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	configType := "global"
	if useProject {
		configType = "project"
	}

	if cfg.Extensions == nil || cfg.Extensions[extName] == nil {
		fmt.Printf("%s.%s is not set in %s config\n", extName, key, configType)
		return
	}

	extCfg := cfg.Extensions[extName]
	switch key {
	case "version":
		extCfg.Version = ""
	case "automount":
		extCfg.Automount = nil
	}

	// Clean up empty extension config
	if extCfg.Version == "" && extCfg.Automount == nil {
		delete(cfg.Extensions, extName)
	}

	// Clean up empty extensions map
	if len(cfg.Extensions) == 0 {
		cfg.Extensions = nil
	}

	if useProject {
		if err := config.SaveProjectConfigFile(cfg); err != nil {
			fmt.Printf("Error saving project config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Unset %s.%s (project)\n", extName, key)
	} else {
		if err := config.SaveGlobalConfigFile(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Unset %s.%s\n", extName, key)
	}
}
