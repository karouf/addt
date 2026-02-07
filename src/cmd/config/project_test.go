package config

import (
	"path/filepath"
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestGetProjectConfigPath(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	path := cfgtypes.GetProjectConfigPath()

	// Check that path ends with .addt.yaml (avoid macOS /var vs /private/var issues)
	if filepath.Base(path) != ".addt.yaml" {
		t.Errorf("GetProjectConfigPath() = %q, want path ending in .addt.yaml", path)
	}
}

func TestSaveAndLoadProjectConfig(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create and save project config
	persistent := true
	cfg := &cfgtypes.GlobalConfig{
		Persistent: &persistent,
		Firewall:   &cfgtypes.FirewallSettings{Mode: "permissive"},
	}

	err := cfgtypes.SaveProjectConfigFile(cfg)
	if err != nil {
		t.Fatalf("SaveProjectConfigFile() error = %v", err)
	}

	// Load and verify
	loaded, err := cfgtypes.LoadProjectConfigFile()
	if err != nil {
		t.Fatalf("LoadProjectConfigFile() error = %v", err)
	}

	if loaded.Persistent == nil || *loaded.Persistent != true {
		t.Errorf("Persistent = %v, want true", loaded.Persistent)
	}
	if loaded.Firewall == nil || loaded.Firewall.Mode != "permissive" {
		t.Errorf("Firewall.Mode = %v, want %q", loaded.Firewall, "permissive")
	}
}
