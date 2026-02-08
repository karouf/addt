package config

import (
	"strings"

	cfgtypes "github.com/jedi4ever/addt/config"
	"github.com/jedi4ever/addt/extensions"
)

// KeyInfo holds metadata about a config key
type KeyInfo struct {
	Key         string
	Description string
	Type        string // "bool", "string", "int"
	EnvVar      string
}

// GetKeys returns all valid config keys with their metadata (sorted alphabetically)
func GetKeys() []KeyInfo {
	return registryGetKeys()
}

// GetDefaultValue returns the default value for a config key
func GetDefaultValue(key string) string {
	return registryGetDefaultValue(key)
}

// IsValidKey checks if a key is a valid config key
func IsValidKey(key string) bool {
	return registryIsValidKey(key)
}

// GetKeyInfo returns the metadata for a config key, or nil if not found
func GetKeyInfo(key string) *KeyInfo {
	return registryGetKeyInfo(key)
}

// GetValue retrieves a config value from the config struct
func GetValue(cfg *cfgtypes.GlobalConfig, key string) string {
	return reflectGetValue(cfg, key)
}

// SetValue sets a config value in the config struct
func SetValue(cfg *cfgtypes.GlobalConfig, key, value string) {
	reflectSetValue(cfg, key, value)
}

// UnsetValue clears a config value in the config struct
func UnsetValue(cfg *cfgtypes.GlobalConfig, key string) {
	reflectUnsetValue(cfg, key)
}

// GetExtensionKeys returns all valid extension config keys with their metadata
func GetExtensionKeys() []KeyInfo {
	return registryGetExtensionKeys()
}

// GetExtensionFlagKeys returns dynamic extension keys derived from an extension's config.yaml flags
func GetExtensionFlagKeys(extName string) []KeyInfo {
	exts, err := extensions.GetExtensions()
	if err != nil {
		return nil
	}
	var keys []KeyInfo
	for _, ext := range exts {
		if ext.Name != extName {
			continue
		}
		for _, flag := range ext.Flags {
			if flag.EnvVar == "" {
				continue
			}
			key := strings.TrimPrefix(flag.Flag, "--")
			keys = append(keys, KeyInfo{
				Key:         key,
				Description: flag.Description,
				Type:        "bool",
				EnvVar:      flag.EnvVar,
			})
		}
		break
	}
	return keys
}

// GetAllExtensionKeys returns both static and dynamic (flag) keys for an extension
func GetAllExtensionKeys(extName string) []KeyInfo {
	keys := GetExtensionKeys()
	keys = append(keys, GetExtensionFlagKeys(extName)...)
	return keys
}

// AvailableExtensionKeyNames returns a comma-separated list of all valid extension key names
func AvailableExtensionKeyNames(extName string) string {
	keys := GetAllExtensionKeys(extName)
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = k.Key
	}
	return strings.Join(names, ", ")
}

// IsValidExtensionKey checks if a key is a valid extension config key (static or dynamic flag)
func IsValidExtensionKey(key string, extName string) bool {
	for _, k := range GetAllExtensionKeys(extName) {
		if k.Key == key {
			return true
		}
	}
	return false
}

// IsFlagKey checks if a key corresponds to a dynamic flag key for the given extension
func IsFlagKey(key string, extName string) bool {
	for _, k := range GetExtensionFlagKeys(extName) {
		if k.Key == key {
			return true
		}
	}
	return false
}
