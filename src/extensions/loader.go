package extensions

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/jedi4ever/addt/util"
	"gopkg.in/yaml.v3"
)

// GetLocalExtensionsDir returns the path to local extensions directory (~/.addt/extensions)
func GetLocalExtensionsDir() string {
	addtHome := util.GetAddtHome()
	if addtHome == "" {
		return ""
	}
	return filepath.Join(addtHome, "extensions")
}

// GetExtensions reads all extension configs from embedded filesystem and local ~/.addt/extensions/
func GetExtensions() ([]ExtensionConfig, error) {
	configMap := make(map[string]ExtensionConfig)

	// First, read embedded extensions
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		configPath := entry.Name() + "/config.yaml"
		data, err := FS.ReadFile(configPath)
		if err != nil {
			continue // Skip directories without config.yaml
		}

		var cfg ExtensionConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue // Skip invalid configs
		}

		cfg.IsLocal = false
		configMap[cfg.Name] = cfg
	}

	// Then, read local extensions (override embedded ones with same name)
	localExtsDir := GetLocalExtensionsDir()
	if localExtsDir != "" {
		if entries, err := os.ReadDir(localExtsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				configPath := filepath.Join(localExtsDir, entry.Name(), "config.yaml")
				data, err := os.ReadFile(configPath)
				if err != nil {
					continue // Skip directories without config.yaml
				}

				var cfg ExtensionConfig
				if err := yaml.Unmarshal(data, &cfg); err != nil {
					continue // Skip invalid configs
				}

				cfg.IsLocal = true
				configMap[cfg.Name] = cfg // Override embedded extension if exists
			}
		}
	}

	// Convert map to slice
	var configs []ExtensionConfig
	for _, cfg := range configMap {
		configs = append(configs, cfg)
	}

	// Sort by name
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})

	return configs, nil
}
