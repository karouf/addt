package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jedi4ever/addt/util"
	"gopkg.in/yaml.v3"
)

// GetGlobalConfigPath returns the path to the global config file
// Can be overridden with ADDT_CONFIG_DIR (for config only) or ADDT_HOME (for all addt data)
func GetGlobalConfigPath() string {
	configDir := os.Getenv("ADDT_CONFIG_DIR")
	if configDir == "" {
		configDir = util.GetAddtHome()
	}
	if configDir == "" {
		return ""
	}
	return filepath.Join(configDir, "config.yaml")
}

// GetProjectConfigPath returns the path to the project config file
func GetProjectConfigPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, ".addt.yaml")
}

// loadGlobalConfig loads the global config from ~/.addt/config.yaml
// Can be overridden with ADDT_CONFIG_DIR environment variable
func loadGlobalConfig() *GlobalConfig {
	configPath := GetGlobalConfigPath()
	if configPath == "" {
		return &GlobalConfig{}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &GlobalConfig{}
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return &GlobalConfig{}
	}

	return &cfg
}

// LoadGlobalConfigFile loads the global config from ~/.addt/config.yaml with error handling
func LoadGlobalConfigFile() (*GlobalConfig, error) {
	configPath := GetGlobalConfigPath()
	if configPath == "" {
		return &GlobalConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{}, nil
		}
		return nil, err
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveGlobalConfigFile saves the global config to ~/.addt/config.yaml
func SaveGlobalConfigFile(cfg *GlobalConfig) error {
	configPath := GetGlobalConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine config file path")
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// loadProjectConfig loads the project config from .addt.yaml in current directory
func loadProjectConfig() *GlobalConfig {
	configPath := GetProjectConfigPath()
	if configPath == "" {
		return &GlobalConfig{}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &GlobalConfig{}
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return &GlobalConfig{}
	}

	return &cfg
}

// LoadProjectConfigFile loads the project config from .addt.yaml in current directory with error handling
func LoadProjectConfigFile() (*GlobalConfig, error) {
	configPath := GetProjectConfigPath()
	if configPath == "" {
		return &GlobalConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{}, nil
		}
		return nil, err
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config file: %w", err)
	}

	return &cfg, nil
}

// SaveProjectConfigFile saves the project config to .addt.yaml in current directory
func SaveProjectConfigFile(cfg *GlobalConfig) error {
	configPath := GetProjectConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine project config file path")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	return nil
}

// LoadGlobalConfig loads the global config (exported wrapper)
func LoadGlobalConfig() *GlobalConfig {
	return loadGlobalConfig()
}

// LoadProjectConfig loads the project config (exported wrapper)
func LoadProjectConfig() *GlobalConfig {
	return loadProjectConfig()
}

// SaveGlobalConfig saves the global config (exported wrapper)
func SaveGlobalConfig(cfg *GlobalConfig) error {
	return SaveGlobalConfigFile(cfg)
}

// SaveProjectConfig saves the project config (exported wrapper)
func SaveProjectConfig(cfg *GlobalConfig) error {
	return SaveProjectConfigFile(cfg)
}
