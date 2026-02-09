package core

import (
	"strings"
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
		Ports:                   []string{"3000", "8080"},
		PortRangeStart:          30000,
		PortsInjectSystemPrompt: true,
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if env["ADDT_PORT_MAP"] == "" {
		t.Error("ADDT_PORT_MAP not set")
	}
}

func TestBuildEnvironment_NoPortMap(t *testing.T) {
	cfg := &provider.Config{
		Ports:                   []string{},
		PortRangeStart:          30000,
		PortsInjectSystemPrompt: true,
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if _, ok := env["ADDT_PORT_MAP"]; ok {
		t.Error("ADDT_PORT_MAP should not be set when no ports configured")
	}
}

func TestBuildEnvironment_OtelEnabled(t *testing.T) {
	cfg := &provider.Config{}
	cfg.Otel.Enabled = true
	cfg.Otel.Endpoint = "http://host.docker.internal:4318"
	cfg.Otel.Protocol = "http/json"
	cfg.Otel.ServiceName = "addt"
	cfg.Extensions = "claude"
	cfg.Provider = "podman"
	cfg.AddtVersion = "0.0.9"

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if env["OTEL_EXPORTER_OTLP_ENDPOINT"] != "http://host.docker.internal:4318" {
		t.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT = %q, want 'http://host.docker.internal:4318'", env["OTEL_EXPORTER_OTLP_ENDPOINT"])
	}
	if env["OTEL_EXPORTER_OTLP_PROTOCOL"] != "http/json" {
		t.Errorf("OTEL_EXPORTER_OTLP_PROTOCOL = %q, want 'http/json'", env["OTEL_EXPORTER_OTLP_PROTOCOL"])
	}
	if env["OTEL_SERVICE_NAME"] != "addt-claude" {
		t.Errorf("OTEL_SERVICE_NAME = %q, want 'addt-claude'", env["OTEL_SERVICE_NAME"])
	}
	if env["CLAUDE_CODE_ENABLE_TELEMETRY"] != "1" {
		t.Errorf("CLAUDE_CODE_ENABLE_TELEMETRY = %q, want '1'", env["CLAUDE_CODE_ENABLE_TELEMETRY"])
	}
	ra := env["OTEL_RESOURCE_ATTRIBUTES"]
	if ra == "" {
		t.Fatal("OTEL_RESOURCE_ATTRIBUTES not set")
	}
	if !strings.Contains(ra, "addt.extension=claude") {
		t.Errorf("OTEL_RESOURCE_ATTRIBUTES = %q, missing addt.extension=claude", ra)
	}
	if !strings.Contains(ra, "addt.provider=podman") {
		t.Errorf("OTEL_RESOURCE_ATTRIBUTES = %q, missing addt.provider=podman", ra)
	}
}

func TestBuildEnvironment_OtelDisabled(t *testing.T) {
	cfg := &provider.Config{}
	cfg.Otel.Enabled = false

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if _, ok := env["OTEL_EXPORTER_OTLP_ENDPOINT"]; ok {
		t.Error("OTEL_EXPORTER_OTLP_ENDPOINT should not be set when OTEL is disabled")
	}
}

func TestBuildEnvironment_GitDisableHooksEnabled(t *testing.T) {
	cfg := &provider.Config{
		GitDisableHooks: true,
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if env["ADDT_GIT_DISABLE_HOOKS"] != "true" {
		t.Errorf("ADDT_GIT_DISABLE_HOOKS = %q, want 'true'", env["ADDT_GIT_DISABLE_HOOKS"])
	}
}

func TestBuildEnvironment_GitDisableHooksDisabled(t *testing.T) {
	cfg := &provider.Config{
		GitDisableHooks: false,
	}

	env := BuildEnvironment(&mockEnvProvider{}, cfg)

	if _, ok := env["ADDT_GIT_DISABLE_HOOKS"]; ok {
		t.Error("ADDT_GIT_DISABLE_HOOKS should not be set when git hooks neutralization is disabled")
	}
}

func TestAddFlagEnvVars_FlagPresent(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{Extensions: "claude"}
	args := []string{"--yolo", "do something"}

	addFlagEnvVars(env, cfg, args)

	if env["ADDT_EXTENSION_CLAUDE_YOLO"] != "true" {
		t.Errorf("ADDT_EXTENSION_CLAUDE_YOLO = %q, want 'true'", env["ADDT_EXTENSION_CLAUDE_YOLO"])
	}
}

func TestAddFlagEnvVars_FlagAbsent(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{Extensions: "claude"}
	args := []string{"do something"}

	addFlagEnvVars(env, cfg, args)

	if _, ok := env["ADDT_EXTENSION_CLAUDE_YOLO"]; ok {
		t.Error("ADDT_EXTENSION_CLAUDE_YOLO should not be set when --yolo is not passed")
	}
}

func TestAddFlagEnvVars_WrongExtension(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{Extensions: "codex"}
	args := []string{"--yolo"}

	addFlagEnvVars(env, cfg, args)

	if _, ok := env["ADDT_EXTENSION_CLAUDE_YOLO"]; ok {
		t.Error("ADDT_EXTENSION_CLAUDE_YOLO should not be set for non-claude extension")
	}
}

func TestAddFlagEnvVars_ConfigSetting(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{
		Extensions: "claude",
		ExtensionFlagSettings: map[string]map[string]bool{
			"claude": {"yolo": true},
		},
	}
	args := []string{"do something"} // no --yolo flag

	addFlagEnvVars(env, cfg, args)

	if env["ADDT_EXTENSION_CLAUDE_YOLO"] != "true" {
		t.Errorf("ADDT_EXTENSION_CLAUDE_YOLO = %q, want 'true' (from config)", env["ADDT_EXTENSION_CLAUDE_YOLO"])
	}
}

func TestAddFlagEnvVars_ConfigSettingFalse(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{
		Extensions: "claude",
		ExtensionFlagSettings: map[string]map[string]bool{
			"claude": {"yolo": false},
		},
	}
	args := []string{"do something"}

	addFlagEnvVars(env, cfg, args)

	if _, ok := env["ADDT_EXTENSION_CLAUDE_YOLO"]; ok {
		t.Error("ADDT_EXTENSION_CLAUDE_YOLO should not be set when config value is false")
	}
}

func TestAddFlagEnvVars_CLIOverridesConfig(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{
		Extensions: "claude",
		ExtensionFlagSettings: map[string]map[string]bool{
			"claude": {"yolo": false}, // config says false
		},
	}
	args := []string{"--yolo", "do something"} // CLI says true

	addFlagEnvVars(env, cfg, args)

	if env["ADDT_EXTENSION_CLAUDE_YOLO"] != "true" {
		t.Errorf("ADDT_EXTENSION_CLAUDE_YOLO = %q, want 'true' (CLI should override config)", env["ADDT_EXTENSION_CLAUDE_YOLO"])
	}
}

func TestAddFlagEnvVars_NilExtensionFlagSettings(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{
		Extensions:            "claude",
		ExtensionFlagSettings: nil,
	}
	args := []string{"do something"}

	addFlagEnvVars(env, cfg, args)

	if _, ok := env["ADDT_EXTENSION_CLAUDE_YOLO"]; ok {
		t.Error("ADDT_EXTENSION_CLAUDE_YOLO should not be set when no config settings")
	}
}

func TestAddFlagEnvVars_GlobalYoloFallback(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{
		Extensions:            "claude",
		ExtensionFlagSettings: nil,
	}
	cfg.Security.Yolo = true
	args := []string{"do something"} // no --yolo flag

	addFlagEnvVars(env, cfg, args)

	if env["ADDT_EXTENSION_CLAUDE_YOLO"] != "true" {
		t.Errorf("ADDT_EXTENSION_CLAUDE_YOLO = %q, want 'true' (from global security.yolo)", env["ADDT_EXTENSION_CLAUDE_YOLO"])
	}
}

func TestAddFlagEnvVars_PerExtensionOverridesGlobalYolo(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{
		Extensions: "claude",
		ExtensionFlagSettings: map[string]map[string]bool{
			"claude": {"yolo": false}, // per-extension explicitly disables
		},
	}
	cfg.Security.Yolo = true // global enables
	args := []string{"do something"}

	addFlagEnvVars(env, cfg, args)

	if _, ok := env["ADDT_EXTENSION_CLAUDE_YOLO"]; ok {
		t.Error("ADDT_EXTENSION_CLAUDE_YOLO should not be set when per-extension explicitly disables yolo")
	}
}

func TestAddFlagEnvVars_GlobalYoloFalseNoEffect(t *testing.T) {
	env := make(map[string]string)
	cfg := &provider.Config{
		Extensions:            "claude",
		ExtensionFlagSettings: nil,
	}
	cfg.Security.Yolo = false
	args := []string{"do something"}

	addFlagEnvVars(env, cfg, args)

	if _, ok := env["ADDT_EXTENSION_CLAUDE_YOLO"]; ok {
		t.Error("ADDT_EXTENSION_CLAUDE_YOLO should not be set when global security.yolo is false")
	}
}

func TestAddTerminalEnvVars_TerminalIdentification(t *testing.T) {
	// Scenario: host terminal sets identification vars that should be forwarded
	// to the container so tools like Claude Code can detect terminal capabilities
	// (OSC 52 clipboard, rich copy blocks, etc.) when terminal.osc is enabled.
	terminalVars := map[string]string{
		"TERM_PROGRAM":          "ghostty",
		"TERM_PROGRAM_VERSION":  "1.2.0",
		"LC_TERMINAL":           "iTerm2",
		"LC_TERMINAL_VERSION":   "3.5.0",
		"KITTY_WINDOW_ID":       "42",
		"ITERM_SESSION_ID":      "w0t0p0:ABC-123",
		"VTE_VERSION":           "7200",
		"GHOSTTY_RESOURCES_DIR": "/usr/share/ghostty",
	}

	// Set all vars
	for k, v := range terminalVars {
		t.Setenv(k, v)
	}

	cfg := &provider.Config{TerminalOSC: true}
	env := make(map[string]string)
	addTerminalEnvVars(env, cfg)

	for k, v := range terminalVars {
		if env[k] != v {
			t.Errorf("env[%q] = %q, want %q", k, env[k], v)
		}
	}
}

func TestAddTerminalEnvVars_TermAlwaysXterm256color(t *testing.T) {
	// Scenario: host uses a custom TERM value (xterm-kitty, xterm-ghostty, etc.)
	// whose terminfo entry may not exist in the container. The container should
	// always get TERM=xterm-256color so that TUI apps render correctly.
	tests := []struct {
		name    string
		hostVal string
	}{
		{"kitty", "xterm-kitty"},
		{"ghostty", "xterm-ghostty"},
		{"plain xterm", "xterm"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.hostVal)
			cfg := &provider.Config{}
			env := make(map[string]string)
			addTerminalEnvVars(env, cfg)
			if env["TERM"] != "xterm-256color" {
				t.Errorf("env[TERM] = %q, want %q", env["TERM"], "xterm-256color")
			}
		})
	}
}

func TestAddTerminalEnvVars_TerminalIdentificationUnset(t *testing.T) {
	// Scenario: none of the terminal identification vars are set on the host;
	// they should not appear in the container environment even with OSC enabled
	varsToCheck := []string{
		"TERM_PROGRAM",
		"TERM_PROGRAM_VERSION",
		"LC_TERMINAL",
		"LC_TERMINAL_VERSION",
		"KITTY_WINDOW_ID",
		"ITERM_SESSION_ID",
		"VTE_VERSION",
		"GHOSTTY_RESOURCES_DIR",
	}

	// Ensure they are unset
	for _, k := range varsToCheck {
		t.Setenv(k, "")
	}

	cfg := &provider.Config{TerminalOSC: true}
	env := make(map[string]string)
	addTerminalEnvVars(env, cfg)

	for _, k := range varsToCheck {
		if _, ok := env[k]; ok {
			t.Errorf("env[%q] should not be set when host var is empty", k)
		}
	}
}

func TestAddTerminalEnvVars_OSCDisabledSkipsIdentification(t *testing.T) {
	// Scenario: terminal.osc is false (default). Terminal identification vars
	// should NOT be forwarded, preventing apps from detecting OSC capabilities.
	// Basic vars (TERM, COLORTERM, COLUMNS, LINES) must still be set.
	terminalVars := map[string]string{
		"TERM_PROGRAM":          "ghostty",
		"TERM_PROGRAM_VERSION":  "1.2.0",
		"LC_TERMINAL":           "iTerm2",
		"LC_TERMINAL_VERSION":   "3.5.0",
		"KITTY_WINDOW_ID":       "42",
		"ITERM_SESSION_ID":      "w0t0p0:ABC-123",
		"VTE_VERSION":           "7200",
		"GHOSTTY_RESOURCES_DIR": "/usr/share/ghostty",
	}

	for k, v := range terminalVars {
		t.Setenv(k, v)
	}
	t.Setenv("COLORTERM", "truecolor")

	cfg := &provider.Config{TerminalOSC: false}
	env := make(map[string]string)
	addTerminalEnvVars(env, cfg)

	// Terminal identification vars must NOT be forwarded
	for k := range terminalVars {
		if _, ok := env[k]; ok {
			t.Errorf("env[%q] should not be set when TerminalOSC is false", k)
		}
	}

	// Basic terminal vars must still be present
	if env["TERM"] != "xterm-256color" {
		t.Errorf("env[TERM] = %q, want %q", env["TERM"], "xterm-256color")
	}
	if env["COLORTERM"] != "truecolor" {
		t.Errorf("env[COLORTERM] = %q, want %q", env["COLORTERM"], "truecolor")
	}
	if env["COLUMNS"] == "" {
		t.Error("COLUMNS not set")
	}
	if env["LINES"] == "" {
		t.Error("LINES not set")
	}
}

func TestAddTerminalEnvVars_OSCEnabledForwardsIdentification(t *testing.T) {
	// Scenario: terminal.osc is true. Terminal identification vars should be
	// forwarded so apps can detect OSC capabilities (clipboard, links, etc.)
	t.Setenv("TERM_PROGRAM", "kitty")
	t.Setenv("KITTY_WINDOW_ID", "99")

	cfg := &provider.Config{TerminalOSC: true}
	env := make(map[string]string)
	addTerminalEnvVars(env, cfg)

	if env["TERM_PROGRAM"] != "kitty" {
		t.Errorf("env[TERM_PROGRAM] = %q, want %q", env["TERM_PROGRAM"], "kitty")
	}
	if env["KITTY_WINDOW_ID"] != "99" {
		t.Errorf("env[KITTY_WINDOW_ID] = %q, want %q", env["KITTY_WINDOW_ID"], "99")
	}
	if env["TERM"] != "xterm-256color" {
		t.Errorf("env[TERM] = %q, want %q", env["TERM"], "xterm-256color")
	}
	if env["COLUMNS"] == "" {
		t.Error("COLUMNS not set")
	}
}

func TestParseEnvVarSpec(t *testing.T) {
	tests := []struct {
		spec        string
		wantName    string
		wantDefault string
	}{
		{"VAR_NAME", "VAR_NAME", ""},
		{"VAR_NAME=value", "VAR_NAME", "value"},
		{"VAR_NAME=", "VAR_NAME", ""},
		{"VAR_NAME=value=with=equals", "VAR_NAME", "value=with=equals"},
		{"CLAUDE_CODE_ENABLE_TELEMETRY=1", "CLAUDE_CODE_ENABLE_TELEMETRY", "1"},
		{"OTEL_LOG_USER_PROMPTS=true", "OTEL_LOG_USER_PROMPTS", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			name, defaultValue := parseEnvVarSpec(tt.spec)
			if name != tt.wantName {
				t.Errorf("parseEnvVarSpec(%q) name = %q, want %q", tt.spec, name, tt.wantName)
			}
			if defaultValue != tt.wantDefault {
				t.Errorf("parseEnvVarSpec(%q) default = %q, want %q", tt.spec, defaultValue, tt.wantDefault)
			}
		})
	}
}
