package core

import (
	"testing"

	"github.com/jedi4ever/addt/provider"
)

// mockEnvProvider implements the minimal provider interface for env tests
type mockEnvProvider struct{}

func (m *mockEnvProvider) Initialize(cfg *provider.Config) error              { return nil }
func (m *mockEnvProvider) Run(spec *provider.RunSpec) error                   { return nil }
func (m *mockEnvProvider) Shell(spec *provider.RunSpec) error                 { return nil }
func (m *mockEnvProvider) Cleanup() error                                     { return nil }
func (m *mockEnvProvider) Exists(name string) bool                            { return false }
func (m *mockEnvProvider) IsRunning(name string) bool                         { return false }
func (m *mockEnvProvider) Start(name string) error                            { return nil }
func (m *mockEnvProvider) Stop(name string) error                             { return nil }
func (m *mockEnvProvider) Remove(name string) error                           { return nil }
func (m *mockEnvProvider) List() ([]provider.Environment, error)              { return nil, nil }
func (m *mockEnvProvider) GeneratePersistentName() string                     { return "test-persistent" }
func (m *mockEnvProvider) GenerateEphemeralName() string                      { return "test-ephemeral" }
func (m *mockEnvProvider) GetStatus(cfg *provider.Config, name string) string { return "test" }
func (m *mockEnvProvider) GetName() string                                    { return "mock" }
func (m *mockEnvProvider) GetExtensionEnvVars(imageName string) []string      { return nil }
func (m *mockEnvProvider) DetermineImageName() string                         { return "test-image" }
func (m *mockEnvProvider) BuildIfNeeded(rebuild bool, rebuildBase bool) error { return nil }

func TestBuildEnvironment_Basic(t *testing.T) {
	cfg := &provider.Config{}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	// COLUMNS and LINES should always be set
	if env["COLUMNS"] == "" {
		t.Error("COLUMNS not set")
	}

	if env["LINES"] == "" {
		t.Error("LINES not set")
	}
}

func TestBuildEnvironment_Firewall(t *testing.T) {
	cfg := &provider.Config{
		FirewallEnabled: true,
		FirewallMode:    "allowlist",
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if env["ADDT_FIREWALL_ENABLED"] != "true" {
		t.Errorf("ADDT_FIREWALL_ENABLED = %q, want 'true'", env["ADDT_FIREWALL_ENABLED"])
	}

	if env["ADDT_FIREWALL_MODE"] != "allowlist" {
		t.Errorf("ADDT_FIREWALL_MODE = %q, want 'allowlist'", env["ADDT_FIREWALL_MODE"])
	}
}

func TestBuildEnvironment_FirewallDisabled(t *testing.T) {
	cfg := &provider.Config{
		FirewallEnabled: false,
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if _, ok := env["ADDT_FIREWALL_ENABLED"]; ok {
		t.Error("ADDT_FIREWALL_ENABLED should not be set when firewall is disabled")
	}
}

func TestBuildEnvironment_Command(t *testing.T) {
	cfg := &provider.Config{
		Command: "codex",
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if env["ADDT_COMMAND"] != "codex" {
		t.Errorf("ADDT_COMMAND = %q, want 'codex'", env["ADDT_COMMAND"])
	}
}

func TestBuildEnvironment_CommandNotSet(t *testing.T) {
	cfg := &provider.Config{
		Command: "",
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if _, ok := env["ADDT_COMMAND"]; ok {
		t.Error("ADDT_COMMAND should not be set when command is empty")
	}
}

func TestBuildEnvironment_PortMap(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{"3000", "8080"},
		PortRangeStart: 30000,
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if env["ADDT_PORT_MAP"] == "" {
		t.Error("ADDT_PORT_MAP not set")
	}
}

func TestBuildEnvironment_NoPortMap(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{},
		PortRangeStart: 30000,
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if _, ok := env["ADDT_PORT_MAP"]; ok {
		t.Error("ADDT_PORT_MAP should not be set when no ports configured")
	}
}
