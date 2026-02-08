package config

import (
	_ "embed"
	"fmt"
	"sort"

	cfgtypes "github.com/jedi4ever/addt/config"
	"gopkg.in/yaml.v3"
)

//go:embed config_keys.yaml
var keysYAML []byte

//go:embed config_extension_keys.yaml
var extensionKeysYAML []byte

// KeyDef holds metadata for a single config key, loaded from config_keys.yaml
type KeyDef struct {
	Key         string `yaml:"key"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`    // "bool", "string", "int", "string_list"
	EnvVar      string `yaml:"env_var"` // e.g. "ADDT_FIREWALL"
	Default     string `yaml:"default"`
	Namespace   string `yaml:"namespace"`
}

type keysFile struct {
	Keys []KeyDef `yaml:"keys"`
}

var (
	allKeyDefs          []KeyDef
	keyDefMap           map[string]*KeyDef
	allExtensionKeyDefs []KeyDef
	extensionKeyDefMap  map[string]*KeyDef
)

func init() {
	var kf keysFile
	if err := yaml.Unmarshal(keysYAML, &kf); err != nil {
		panic(fmt.Sprintf("config: failed to parse config_keys.yaml: %v", err))
	}
	allKeyDefs = kf.Keys
	keyDefMap = make(map[string]*KeyDef, len(allKeyDefs))
	for i := range allKeyDefs {
		keyDefMap[allKeyDefs[i].Key] = &allKeyDefs[i]
	}

	// Validate every key resolves against a zero-value GlobalConfig
	cfg := &cfgtypes.GlobalConfig{}
	for _, kd := range allKeyDefs {
		if _, ok := resolveField(cfg, kd.Key, false); !ok {
			panic(fmt.Sprintf("config: key %q does not resolve against GlobalConfig struct", kd.Key))
		}
	}

	// Load extension keys
	var ekf keysFile
	if err := yaml.Unmarshal(extensionKeysYAML, &ekf); err != nil {
		panic(fmt.Sprintf("config: failed to parse config_extension_keys.yaml: %v", err))
	}
	allExtensionKeyDefs = ekf.Keys
	extensionKeyDefMap = make(map[string]*KeyDef, len(allExtensionKeyDefs))
	for i := range allExtensionKeyDefs {
		extensionKeyDefMap[allExtensionKeyDefs[i].Key] = &allExtensionKeyDefs[i]
	}
}

// registryGetKeys returns all config keys as KeyInfo, sorted alphabetically
func registryGetKeys() []KeyInfo {
	keys := make([]KeyInfo, len(allKeyDefs))
	for i, kd := range allKeyDefs {
		t := kd.Type
		if t == "string_list" {
			t = "string" // external API shows "string" for comma-separated lists
		}
		keys[i] = KeyInfo{
			Key:         kd.Key,
			Description: kd.Description,
			Type:        t,
			EnvVar:      kd.EnvVar,
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Key < keys[j].Key
	})
	return keys
}

// registryGetDefaultValue returns the default value for a config key
func registryGetDefaultValue(key string) string {
	if kd, ok := keyDefMap[key]; ok {
		return kd.Default
	}
	return ""
}

// registryIsValidKey checks if a key is a valid config key
func registryIsValidKey(key string) bool {
	_, ok := keyDefMap[key]
	return ok
}

// registryGetKeyInfo returns the metadata for a config key, or nil if not found
func registryGetKeyInfo(key string) *KeyInfo {
	kd, ok := keyDefMap[key]
	if !ok {
		return nil
	}
	t := kd.Type
	if t == "string_list" {
		t = "string"
	}
	return &KeyInfo{
		Key:         kd.Key,
		Description: kd.Description,
		Type:        t,
		EnvVar:      kd.EnvVar,
	}
}

// GetKeyDef returns the raw KeyDef for a config key, or nil if not found.
// Unlike GetKeyInfo, this includes the default value and namespace.
func GetKeyDef(key string) *KeyDef {
	kd, ok := keyDefMap[key]
	if !ok {
		return nil
	}
	return kd
}

// GetAllKeyDefs returns all key definitions (for audit, display, etc.)
func GetAllKeyDefs() []KeyDef {
	return allKeyDefs
}

// registryGetExtensionKeys returns all extension keys as KeyInfo, sorted alphabetically
func registryGetExtensionKeys() []KeyInfo {
	keys := make([]KeyInfo, len(allExtensionKeyDefs))
	for i, kd := range allExtensionKeyDefs {
		keys[i] = KeyInfo{
			Key:         kd.Key,
			Description: kd.Description,
			Type:        kd.Type,
			EnvVar:      kd.EnvVar,
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Key < keys[j].Key
	})
	return keys
}

// GetExtensionKeyDef returns the raw KeyDef for an extension key, or nil if not found.
func GetExtensionKeyDef(key string) *KeyDef {
	kd, ok := extensionKeyDefMap[key]
	if !ok {
		return nil
	}
	return kd
}

// GetAllExtensionKeyDefs returns all extension key definitions.
func GetAllExtensionKeyDefs() []KeyDef {
	return allExtensionKeyDefs
}
