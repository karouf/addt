package config

import (
	"fmt"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// GetFirewallKeys returns all valid firewall config keys
func GetFirewallKeys() []KeyInfo {
	return []KeyInfo{
		{Key: "firewall.enabled", Description: "Enable network firewall (default: false)", Type: "bool", EnvVar: "ADDT_FIREWALL"},
		{Key: "firewall.mode", Description: "Firewall mode: strict, permissive, off (default: strict)", Type: "string", EnvVar: "ADDT_FIREWALL_MODE"},
	}
}

// GetFirewallValue retrieves a firewall config value
func GetFirewallValue(f *cfgtypes.FirewallSettings, key string) string {
	if f == nil {
		return ""
	}
	switch key {
	case "firewall.enabled":
		if f.Enabled != nil {
			return fmt.Sprintf("%v", *f.Enabled)
		}
	case "firewall.mode":
		return f.Mode
	}
	return ""
}

// SetFirewallValue sets a firewall config value
func SetFirewallValue(f *cfgtypes.FirewallSettings, key, value string) {
	switch key {
	case "firewall.enabled":
		b := value == "true"
		f.Enabled = &b
	case "firewall.mode":
		f.Mode = value
	}
}

// UnsetFirewallValue clears a firewall config value
func UnsetFirewallValue(f *cfgtypes.FirewallSettings, key string) {
	switch key {
	case "firewall.enabled":
		f.Enabled = nil
	case "firewall.mode":
		f.Mode = ""
	}
}
