package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestEnv creates temporary directories for testing and sets ADDT_CONFIG_DIR
// Returns cleanup function to restore original state
func setupTestEnv(t *testing.T) (globalDir, projectDir string, cleanup func()) {
	t.Helper()

	// Save original env vars
	origConfigDir := os.Getenv("ADDT_CONFIG_DIR")
	origCwd, _ := os.Getwd()

	// Create temp directories
	globalDir = t.TempDir()
	projectDir = t.TempDir()

	// Set ADDT_CONFIG_DIR to isolate from real home directory
	os.Setenv("ADDT_CONFIG_DIR", globalDir)

	// Change to project directory
	os.Chdir(projectDir)

	cleanup = func() {
		os.Setenv("ADDT_CONFIG_DIR", origConfigDir)
		os.Chdir(origCwd)
	}

	return globalDir, projectDir, cleanup
}

func TestGetConfigFilePath(t *testing.T) {
	globalDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	path := GetConfigFilePath()

	// Check that path ends with config.yaml and contains our temp dir name
	// (avoid macOS /var vs /private/var symlink issues)
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("GetConfigFilePath() should end with config.yaml, got %q", path)
	}
	if filepath.Base(filepath.Dir(path)) != filepath.Base(globalDir) {
		t.Errorf("GetConfigFilePath() dir = %q, want dir containing %q", filepath.Dir(path), filepath.Base(globalDir))
	}
}

func TestGetProjectConfigFilePath(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	path := GetProjectConfigFilePath()

	// Check that path ends with .addt.yaml (avoid macOS /var vs /private/var issues)
	if filepath.Base(path) != ".addt.yaml" {
		t.Errorf("GetProjectConfigFilePath() = %q, want path ending in .addt.yaml", path)
	}
}

func TestLoadGlobalConfig_NonExistent(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadGlobalConfig() returned nil")
	}

	// Should return empty config when file doesn't exist
	if cfg.NodeVersion != "" {
		t.Errorf("Expected empty NodeVersion, got %q", cfg.NodeVersion)
	}
}

func TestSaveAndLoadGlobalConfig(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create and save config
	cfg := &GlobalConfig{
		NodeVersion:  "20",
		GoVersion:    "1.21",
		DockerCPUs:   "2",
		DockerMemory: "4g",
	}

	err := SaveGlobalConfig(cfg)
	if err != nil {
		t.Fatalf("SaveGlobalConfig() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() error = %v", err)
	}

	if loaded.NodeVersion != "20" {
		t.Errorf("NodeVersion = %q, want %q", loaded.NodeVersion, "20")
	}
	if loaded.GoVersion != "1.21" {
		t.Errorf("GoVersion = %q, want %q", loaded.GoVersion, "1.21")
	}
	if loaded.DockerCPUs != "2" {
		t.Errorf("DockerCPUs = %q, want %q", loaded.DockerCPUs, "2")
	}
	if loaded.DockerMemory != "4g" {
		t.Errorf("DockerMemory = %q, want %q", loaded.DockerMemory, "4g")
	}
}

func TestSaveAndLoadProjectConfig(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create and save project config
	persistent := true
	cfg := &GlobalConfig{
		Persistent:   &persistent,
		FirewallMode: "permissive",
	}

	err := SaveProjectConfig(cfg)
	if err != nil {
		t.Fatalf("SaveProjectConfig() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadProjectConfig()
	if err != nil {
		t.Fatalf("LoadProjectConfig() error = %v", err)
	}

	if loaded.Persistent == nil || *loaded.Persistent != true {
		t.Errorf("Persistent = %v, want true", loaded.Persistent)
	}
	if loaded.FirewallMode != "permissive" {
		t.Errorf("FirewallMode = %q, want %q", loaded.FirewallMode, "permissive")
	}
}

func TestExtensionSettings(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config with extension settings
	automount := true
	cfg := &GlobalConfig{
		Extensions: map[string]*ExtensionSettings{
			"claude": {
				Version:   "1.0.5",
				Automount: &automount,
			},
		},
	}

	err := SaveGlobalConfig(cfg)
	if err != nil {
		t.Fatalf("SaveGlobalConfig() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("LoadGlobalConfig() error = %v", err)
	}

	if loaded.Extensions == nil {
		t.Fatal("Extensions is nil")
	}

	claudeCfg := loaded.Extensions["claude"]
	if claudeCfg == nil {
		t.Fatal("claude extension config is nil")
	}

	if claudeCfg.Version != "1.0.5" {
		t.Errorf("claude.Version = %q, want %q", claudeCfg.Version, "1.0.5")
	}
	if claudeCfg.Automount == nil || *claudeCfg.Automount != true {
		t.Errorf("claude.Automount = %v, want true", claudeCfg.Automount)
	}
}

func TestExtensionSettingsInProjectConfig(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create project config with extension settings
	automount := false
	cfg := &GlobalConfig{
		Extensions: map[string]*ExtensionSettings{
			"claude": {
				Automount: &automount,
			},
		},
	}

	err := SaveProjectConfig(cfg)
	if err != nil {
		t.Fatalf("SaveProjectConfig() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadProjectConfig()
	if err != nil {
		t.Fatalf("LoadProjectConfig() error = %v", err)
	}

	if loaded.Extensions == nil {
		t.Fatal("Extensions is nil")
	}

	claudeCfg := loaded.Extensions["claude"]
	if claudeCfg == nil {
		t.Fatal("claude extension config is nil")
	}

	if claudeCfg.Automount == nil || *claudeCfg.Automount != false {
		t.Errorf("claude.Automount = %v, want false", claudeCfg.Automount)
	}
}

func TestConfigKeyValidation(t *testing.T) {
	validKeys := []string{
		"docker_cpus", "docker_memory", "dind", "dind_mode",
		"firewall", "firewall_mode", "node_version", "go_version",
		"persistent", "workdir", "workdir_automount",
	}

	for _, key := range validKeys {
		if !isValidConfigKey(key) {
			t.Errorf("isValidConfigKey(%q) = false, want true", key)
		}
	}

	invalidKeys := []string{"invalid", "foo", "bar", "version"}
	for _, key := range invalidKeys {
		if isValidConfigKey(key) {
			t.Errorf("isValidConfigKey(%q) = true, want false", key)
		}
	}
}

func TestExtensionConfigKeyValidation(t *testing.T) {
	validKeys := []string{"version", "automount"}
	for _, key := range validKeys {
		if !isValidExtensionConfigKey(key) {
			t.Errorf("isValidExtensionConfigKey(%q) = false, want true", key)
		}
	}

	invalidKeys := []string{"invalid", "foo", "node_version"}
	for _, key := range invalidKeys {
		if isValidExtensionConfigKey(key) {
			t.Errorf("isValidExtensionConfigKey(%q) = true, want false", key)
		}
	}
}

func TestGetConfigValue(t *testing.T) {
	persistent := true
	portStart := 35000
	cfg := &GlobalConfig{
		NodeVersion:    "20",
		DockerCPUs:     "4",
		Persistent:     &persistent,
		PortRangeStart: &portStart,
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"node_version", "20"},
		{"docker_cpus", "4"},
		{"persistent", "true"},
		{"port_range_start", "35000"},
		{"go_version", ""}, // not set
	}

	for _, tt := range tests {
		got := getConfigValue(cfg, tt.key)
		if got != tt.expected {
			t.Errorf("getConfigValue(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestSetConfigValue(t *testing.T) {
	cfg := &GlobalConfig{}

	setConfigValue(cfg, "node_version", "18")
	if cfg.NodeVersion != "18" {
		t.Errorf("NodeVersion = %q, want %q", cfg.NodeVersion, "18")
	}

	setConfigValue(cfg, "persistent", "true")
	if cfg.Persistent == nil || *cfg.Persistent != true {
		t.Errorf("Persistent = %v, want true", cfg.Persistent)
	}

	setConfigValue(cfg, "port_range_start", "40000")
	if cfg.PortRangeStart == nil || *cfg.PortRangeStart != 40000 {
		t.Errorf("PortRangeStart = %v, want 40000", cfg.PortRangeStart)
	}
}

func TestUnsetConfigValue(t *testing.T) {
	persistent := true
	cfg := &GlobalConfig{
		NodeVersion: "20",
		Persistent:  &persistent,
	}

	unsetConfigValue(cfg, "node_version")
	if cfg.NodeVersion != "" {
		t.Errorf("NodeVersion = %q, want empty", cfg.NodeVersion)
	}

	unsetConfigValue(cfg, "persistent")
	if cfg.Persistent != nil {
		t.Errorf("Persistent = %v, want nil", cfg.Persistent)
	}
}

func TestGetDefaultValue(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"node_version", "22"},
		{"firewall", "false"},
		{"firewall_mode", "strict"},
		{"persistent", "false"},
		{"workdir_automount", "true"},
		{"port_range_start", "30000"},
		{"ssh_forward", "agent"},
	}

	for _, tt := range tests {
		got := getDefaultValue(tt.key)
		if got != tt.expected {
			t.Errorf("getDefaultValue(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}
