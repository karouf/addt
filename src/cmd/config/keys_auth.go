package config

import (
	"fmt"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// GetAuthKeys returns all valid auth config keys
func GetAuthKeys() []KeyInfo {
	return []KeyInfo{
		{Key: "auth.autologin", Description: "Automatically handle authentication on first launch (default: true)", Type: "bool", EnvVar: "ADDT_AUTH_AUTOLOGIN"},
		{Key: "auth.method", Description: "Authentication method: native, env, auto (default: auto)", Type: "string", EnvVar: "ADDT_AUTH_METHOD"},
	}
}

// GetAuthValue retrieves an auth config value
func GetAuthValue(a *cfgtypes.AuthSettings, key string) string {
	if a == nil {
		return ""
	}
	switch key {
	case "auth.autologin":
		if a.Autologin != nil {
			return fmt.Sprintf("%v", *a.Autologin)
		}
	case "auth.method":
		return a.Method
	}
	return ""
}

// SetAuthValue sets an auth config value
func SetAuthValue(a *cfgtypes.AuthSettings, key, value string) {
	switch key {
	case "auth.autologin":
		b := value == "true"
		a.Autologin = &b
	case "auth.method":
		a.Method = value
	}
}

// UnsetAuthValue clears an auth config value
func UnsetAuthValue(a *cfgtypes.AuthSettings, key string) {
	switch key {
	case "auth.autologin":
		a.Autologin = nil
	case "auth.method":
		a.Method = ""
	}
}
