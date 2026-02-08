package config

import (
	"fmt"
	"os"
	"strings"

	cfgtypes "github.com/jedi4ever/addt/config"
	"github.com/jedi4ever/addt/extensions"
)

func listExtension(extName string, useGlobal, verbose bool) {
	// Get extension defaults from extension's config.yaml
	var extDefaults *extensions.ExtensionConfig
	exts, err := extensions.GetExtensions()
	if err == nil {
		for _, ext := range exts {
			if ext.Name == extName {
				extDefaults = &ext
				break
			}
		}
	}

	extNameUpper := strings.ToUpper(extName)
	scope := "project"
	if useGlobal {
		scope = "global"
	}
	fmt.Printf("Extension: %s (%s)\n\n", extName, scope)

	keys := GetAllExtensionKeys(extName)

	// Load the appropriate config
	var cfg *cfgtypes.GlobalConfig
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var extCfg *cfgtypes.ExtensionSettings
	if cfg.Extensions != nil {
		extCfg = cfg.Extensions[extName]
	}

	rows := make([]configRow, 0, len(keys))
	for _, k := range keys {
		var envVar string
		if strings.Contains(k.EnvVar, "%s") {
			envVar = fmt.Sprintf(k.EnvVar, extNameUpper)
		} else {
			envVar = k.EnvVar
		}
		envValue := os.Getenv(envVar)

		var configValue, defaultValue string

		// Get config value
		if extCfg != nil {
			switch k.Key {
			case "version":
				configValue = extCfg.Version
			case "config.automount":
				if extCfg.Config != nil && extCfg.Config.Automount != nil {
					configValue = fmt.Sprintf("%v", *extCfg.Config.Automount)
				}
			case "config.readonly":
				if extCfg.Config != nil && extCfg.Config.Readonly != nil {
					configValue = fmt.Sprintf("%v", *extCfg.Config.Readonly)
				}
			case "workdir.autotrust":
				if extCfg.Workdir != nil && extCfg.Workdir.Autotrust != nil {
					configValue = fmt.Sprintf("%v", *extCfg.Workdir.Autotrust)
				}
			case "auth.autologin":
				if extCfg.Auth != nil && extCfg.Auth.Autologin != nil {
					configValue = fmt.Sprintf("%v", *extCfg.Auth.Autologin)
				}
			case "auth.method":
				if extCfg.Auth != nil {
					configValue = extCfg.Auth.Method
				}
			default:
				// Check flag keys
				if IsFlagKey(k.Key, extName) && extCfg.Flags != nil {
					if v, ok := extCfg.Flags[k.Key]; ok && v != nil {
						configValue = fmt.Sprintf("%v", *v)
					}
				}
			}
		}

		// Get extension default value
		if extDefaults != nil {
			switch k.Key {
			case "version":
				defaultValue = extDefaults.DefaultVersion
			case "config.automount":
				defaultValue = fmt.Sprintf("%v", extDefaults.Config.Automount)
			case "config.readonly":
				defaultValue = fmt.Sprintf("%v", extDefaults.Config.Readonly)
			case "workdir.autotrust":
				defaultValue = "true"
			case "auth.autologin":
				defaultValue = fmt.Sprintf("%v", extDefaults.Auth.Autologin)
			case "auth.method":
				if extDefaults.Auth.Method != "" {
					defaultValue = extDefaults.Auth.Method
				} else {
					defaultValue = "auto"
				}
			default:
				// Flag keys default to "false"
				if IsFlagKey(k.Key, extName) {
					defaultValue = "false"
				}
			}
		}

		// Determine effective value and source (env > config > default)
		var displayValue, source string
		if envValue != "" {
			displayValue = envValue
			source = "env"
		} else if configValue != "" {
			displayValue = configValue
			source = scope
		} else if defaultValue != "" {
			displayValue = defaultValue
			source = "default"
		} else {
			displayValue = "-"
			source = ""
		}

		def := defaultValue
		if def == "" {
			def = "-"
		}

		rows = append(rows, configRow{
			Key:          k.Key,
			Value:        displayValue,
			Default:      def,
			Source:       source,
			IsOverridden: source == "env" || source == scope,
			Description:  k.Description,
		})
	}

	printRows(rows, verbose)
}

func getExtension(extName, key string, useGlobal bool) {
	if !IsValidExtensionKey(key, extName) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Printf("Available keys: %s\n", AvailableExtensionKeyNames(extName))
		os.Exit(1)
	}

	var cfg *cfgtypes.GlobalConfig
	var err error
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var extCfg *cfgtypes.ExtensionSettings
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
	case "config.automount":
		if extCfg.Config != nil && extCfg.Config.Automount != nil {
			val = fmt.Sprintf("%v", *extCfg.Config.Automount)
		}
	case "config.readonly":
		if extCfg.Config != nil && extCfg.Config.Readonly != nil {
			val = fmt.Sprintf("%v", *extCfg.Config.Readonly)
		}
	case "workdir.autotrust":
		if extCfg.Workdir != nil && extCfg.Workdir.Autotrust != nil {
			val = fmt.Sprintf("%v", *extCfg.Workdir.Autotrust)
		}
	case "auth.autologin":
		if extCfg.Auth != nil && extCfg.Auth.Autologin != nil {
			val = fmt.Sprintf("%v", *extCfg.Auth.Autologin)
		}
	case "auth.method":
		if extCfg.Auth != nil {
			val = extCfg.Auth.Method
		}
	default:
		// Check flag keys
		if IsFlagKey(key, extName) && extCfg.Flags != nil {
			if v, ok := extCfg.Flags[key]; ok && v != nil {
				val = fmt.Sprintf("%v", *v)
			}
		}
	}

	if val == "" {
		fmt.Printf("%s is not set\n", key)
	} else {
		fmt.Println(val)
	}
}

func setExtension(extName, key, value string, useGlobal bool) {
	if !IsValidExtensionKey(key, extName) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Printf("Available keys: %s\n", AvailableExtensionKeyNames(extName))
		os.Exit(1)
	}

	// Validate bool values for automount, workdir.autotrust, auth.autologin, and flag keys
	if key == "config.automount" || key == "config.readonly" || key == "workdir.autotrust" || key == "auth.autologin" || IsFlagKey(key, extName) {
		value = strings.ToLower(value)
		if value != "true" && value != "false" {
			fmt.Printf("Invalid value for %s: must be 'true' or 'false'\n", key)
			os.Exit(1)
		}
	}

	var cfg *cfgtypes.GlobalConfig
	var err error
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize extensions map if needed
	if cfg.Extensions == nil {
		cfg.Extensions = make(map[string]*cfgtypes.ExtensionSettings)
	}

	// Initialize extension config if needed
	if cfg.Extensions[extName] == nil {
		cfg.Extensions[extName] = &cfgtypes.ExtensionSettings{}
	}

	extCfg := cfg.Extensions[extName]
	switch key {
	case "version":
		extCfg.Version = value
	case "config.automount":
		if extCfg.Config == nil {
			extCfg.Config = &cfgtypes.ConfigSettings{}
		}
		b := value == "true"
		extCfg.Config.Automount = &b
	case "config.readonly":
		if extCfg.Config == nil {
			extCfg.Config = &cfgtypes.ConfigSettings{}
		}
		b := value == "true"
		extCfg.Config.Readonly = &b
	case "workdir.autotrust":
		if extCfg.Workdir == nil {
			extCfg.Workdir = &cfgtypes.ExtensionWorkdirSettings{}
		}
		b := value == "true"
		extCfg.Workdir.Autotrust = &b
	case "auth.autologin":
		if extCfg.Auth == nil {
			extCfg.Auth = &cfgtypes.AuthSettings{}
		}
		b := value == "true"
		extCfg.Auth.Autologin = &b
	case "auth.method":
		if extCfg.Auth == nil {
			extCfg.Auth = &cfgtypes.AuthSettings{}
		}
		extCfg.Auth.Method = value
	default:
		// Handle flag keys
		if IsFlagKey(key, extName) {
			if extCfg.Flags == nil {
				extCfg.Flags = make(map[string]*bool)
			}
			b := value == "true"
			extCfg.Flags[key] = &b
		}
	}

	scope := "project"
	if useGlobal {
		if err := cfgtypes.SaveGlobalConfigFile(cfg); err != nil {
			fmt.Printf("Error saving global config: %v\n", err)
			os.Exit(1)
		}
		scope = "global"
	} else {
		if err := cfgtypes.SaveProjectConfigFile(cfg); err != nil {
			fmt.Printf("Error saving project config: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Set %s.%s = %s (%s)\n", extName, key, value, scope)
}

func unsetExtension(extName, key string, useGlobal bool) {
	if !IsValidExtensionKey(key, extName) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Printf("Available keys: %s\n", AvailableExtensionKeyNames(extName))
		os.Exit(1)
	}

	var cfg *cfgtypes.GlobalConfig
	var err error
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	scope := "project"
	if useGlobal {
		scope = "global"
	}

	if cfg.Extensions == nil || cfg.Extensions[extName] == nil {
		fmt.Printf("%s.%s is not set in %s config\n", extName, key, scope)
		return
	}

	extCfg := cfg.Extensions[extName]
	switch key {
	case "version":
		extCfg.Version = ""
	case "config.automount":
		if extCfg.Config != nil {
			extCfg.Config.Automount = nil
		}
	case "config.readonly":
		if extCfg.Config != nil {
			extCfg.Config.Readonly = nil
		}
	case "workdir.autotrust":
		if extCfg.Workdir != nil {
			extCfg.Workdir.Autotrust = nil
		}
	case "auth.autologin":
		if extCfg.Auth != nil {
			extCfg.Auth.Autologin = nil
		}
	case "auth.method":
		if extCfg.Auth != nil {
			extCfg.Auth.Method = ""
		}
	default:
		// Handle flag keys
		if IsFlagKey(key, extName) && extCfg.Flags != nil {
			delete(extCfg.Flags, key)
			if len(extCfg.Flags) == 0 {
				extCfg.Flags = nil
			}
		}
	}

	// Clean up empty extension config
	if isExtensionSettingsEmpty(extCfg) {
		delete(cfg.Extensions, extName)
	}

	// Clean up empty extensions map
	if len(cfg.Extensions) == 0 {
		cfg.Extensions = nil
	}

	if useGlobal {
		if err := cfgtypes.SaveGlobalConfigFile(cfg); err != nil {
			fmt.Printf("Error saving global config: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := cfgtypes.SaveProjectConfigFile(cfg); err != nil {
			fmt.Printf("Error saving project config: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Unset %s.%s (%s)\n", extName, key, scope)
}

// isExtensionSettingsEmpty returns true if all fields are zero/nil
func isExtensionSettingsEmpty(e *cfgtypes.ExtensionSettings) bool {
	if e.Version != "" || len(e.Flags) > 0 || len(e.FirewallAllowed) > 0 || len(e.FirewallDenied) > 0 {
		return false
	}
	if e.Config != nil && (e.Config.Automount != nil || e.Config.Readonly != nil) {
		return false
	}
	if e.Workdir != nil && e.Workdir.Autotrust != nil {
		return false
	}
	if e.Auth != nil && (e.Auth.Autologin != nil || e.Auth.Method != "") {
		return false
	}
	return true
}
