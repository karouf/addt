package config

import (
	"strings"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// GetGPGKeys returns all valid GPG config keys
func GetGPGKeys() []KeyInfo {
	return []KeyInfo{
		{Key: "gpg.forward", Description: "GPG forwarding mode: proxy, agent, keys, or off", Type: "string", EnvVar: "ADDT_GPG_FORWARD"},
		{Key: "gpg.allowed_key_ids", Description: "GPG key IDs allowed for signing (comma-separated)", Type: "string", EnvVar: "ADDT_GPG_ALLOWED_KEY_IDS"},
		{Key: "gpg.dir", Description: "GPG directory path (default: ~/.gnupg)", Type: "string", EnvVar: "ADDT_GPG_DIR"},
	}
}

// GetGPGValue retrieves a GPG config value
func GetGPGValue(g *cfgtypes.GPGSettings, key string) string {
	if g == nil {
		return ""
	}
	switch key {
	case "gpg.forward":
		return g.Forward
	case "gpg.allowed_key_ids":
		return strings.Join(g.AllowedKeyIDs, ",")
	case "gpg.dir":
		return g.Dir
	}
	return ""
}

// SetGPGValue sets a GPG config value
func SetGPGValue(g *cfgtypes.GPGSettings, key, value string) {
	switch key {
	case "gpg.forward":
		g.Forward = value
	case "gpg.allowed_key_ids":
		if value == "" {
			g.AllowedKeyIDs = nil
		} else {
			g.AllowedKeyIDs = strings.Split(value, ",")
		}
	case "gpg.dir":
		g.Dir = value
	}
}

// UnsetGPGValue clears a GPG config value
func UnsetGPGValue(g *cfgtypes.GPGSettings, key string) {
	switch key {
	case "gpg.forward":
		g.Forward = ""
	case "gpg.allowed_key_ids":
		g.AllowedKeyIDs = nil
	case "gpg.dir":
		g.Dir = ""
	}
}
