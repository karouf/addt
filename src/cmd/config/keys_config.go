package config

import (
	"fmt"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// GetConfigKeys returns all valid config section keys
func GetConfigKeys() []KeyInfo {
	return []KeyInfo{
		{Key: "config.automount", Description: "Auto-mount extension config directories (default: false)", Type: "bool", EnvVar: "ADDT_CONFIG_AUTOMOUNT"},
		{Key: "config.readonly", Description: "Mount extension config directories as read-only (default: false)", Type: "bool", EnvVar: "ADDT_CONFIG_READONLY"},
	}
}

// GetConfigValue retrieves a config section value
func GetConfigValue(c *cfgtypes.ConfigSettings, key string) string {
	if c == nil {
		return ""
	}
	switch key {
	case "config.automount":
		if c.Automount != nil {
			return fmt.Sprintf("%v", *c.Automount)
		}
	case "config.readonly":
		if c.Readonly != nil {
			return fmt.Sprintf("%v", *c.Readonly)
		}
	}
	return ""
}

// SetConfigValue sets a config section value
func SetConfigValue(c *cfgtypes.ConfigSettings, key, value string) {
	switch key {
	case "config.automount":
		b := value == "true"
		c.Automount = &b
	case "config.readonly":
		b := value == "true"
		c.Readonly = &b
	}
}

// UnsetConfigValue clears a config section value
func UnsetConfigValue(c *cfgtypes.ConfigSettings, key string) {
	switch key {
	case "config.automount":
		c.Automount = nil
	case "config.readonly":
		c.Readonly = nil
	}
}
