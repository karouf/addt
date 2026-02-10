package config

import (
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
	"github.com/jedi4ever/addt/config/otel"
	"github.com/jedi4ever/addt/config/security"
)

func TestRegistryLoadsAllKeys(t *testing.T) {
	if len(allKeyDefs) == 0 {
		t.Fatal("allKeyDefs is empty, YAML not loaded")
	}
	// We expect 78 keys total
	if len(allKeyDefs) != 78 {
		t.Errorf("expected 78 key defs, got %d", len(allKeyDefs))
	}
}

func TestAllKeysResolveAgainstStruct(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	for _, kd := range allKeyDefs {
		if _, ok := resolveField(cfg, kd.Key, false); !ok {
			t.Errorf("key %q does not resolve against GlobalConfig", kd.Key)
		}
	}
}

func TestRegistryGetDefaultValue(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"env_file_load", "true"},
		{"env_file", ".env"},
		{"node_version", "22"},
		{"persistent", "false"},
		{"firewall.enabled", "false"},
		{"firewall.mode", "strict"},
		{"docker.dind.enable", "false"},
		{"docker.dind.mode", "isolated"},
		{"ssh.forward_keys", "false"},
		{"ssh.forward_mode", "proxy"},
		{"security.pids_limit", "200"},
		{"security.no_new_privileges", "true"},
		{"security.cap_add", "CHOWN,SETUID,SETGID"},
		{"otel.enabled", "false"},
		{"otel.endpoint", "http://host.docker.internal:4318"},
		{"workdir.automount", "true"},
		{"log.level", "INFO"},
		{"ports.range_start", "30000"},
	}
	for _, tt := range tests {
		got := registryGetDefaultValue(tt.key)
		if got != tt.expected {
			t.Errorf("registryGetDefaultValue(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestRegistryIsValidKey(t *testing.T) {
	validKeys := []string{
		"env_file_load", "persistent", "firewall.enabled",
		"docker.dind.enable", "ssh.allowed_keys", "security.pids_limit",
		"otel.endpoint",
	}
	for _, key := range validKeys {
		if !registryIsValidKey(key) {
			t.Errorf("registryIsValidKey(%q) = false, want true", key)
		}
	}

	invalidKeys := []string{"invalid", "foo", "bar", "version", "docker.dind.nonexistent"}
	for _, key := range invalidKeys {
		if registryIsValidKey(key) {
			t.Errorf("registryIsValidKey(%q) = true, want false", key)
		}
	}
}

func TestRegistryGetKeys(t *testing.T) {
	keys := registryGetKeys()
	if len(keys) != 78 {
		t.Errorf("registryGetKeys() returned %d keys, want 78", len(keys))
	}
	// Verify sorted
	for i := 1; i < len(keys); i++ {
		if keys[i].Key < keys[i-1].Key {
			t.Errorf("keys not sorted: %q before %q", keys[i-1].Key, keys[i].Key)
		}
	}
	// string_list should show as "string" in KeyInfo
	for _, k := range keys {
		if k.Type == "string_list" {
			t.Errorf("key %q has type 'string_list' in KeyInfo, should be 'string'", k.Key)
		}
	}
}

func TestRegistryGetKeyInfo(t *testing.T) {
	ki := registryGetKeyInfo("firewall.enabled")
	if ki == nil {
		t.Fatal("registryGetKeyInfo(firewall.enabled) returned nil")
	}
	if ki.Type != "bool" {
		t.Errorf("type = %q, want bool", ki.Type)
	}
	if ki.EnvVar != "ADDT_FIREWALL" {
		t.Errorf("env_var = %q, want ADDT_FIREWALL", ki.EnvVar)
	}

	if registryGetKeyInfo("nonexistent") != nil {
		t.Error("registryGetKeyInfo(nonexistent) should return nil")
	}
}

// --- Reflection Get/Set/Unset tests ---

func TestReflectGetValueBoolPointer(t *testing.T) {
	b := true
	cfg := &cfgtypes.GlobalConfig{
		Firewall: &cfgtypes.FirewallSettings{Enabled: &b},
	}
	got := reflectGetValue(cfg, "firewall.enabled")
	if got != "true" {
		t.Errorf("reflectGetValue(firewall.enabled) = %q, want %q", got, "true")
	}

	// nil parent
	nilCfg := &cfgtypes.GlobalConfig{}
	if got := reflectGetValue(nilCfg, "firewall.enabled"); got != "" {
		t.Errorf("reflectGetValue with nil Firewall = %q, want empty", got)
	}
}

func TestReflectGetValueString(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{NodeVersion: "20"}
	got := reflectGetValue(cfg, "node_version")
	if got != "20" {
		t.Errorf("reflectGetValue(node_version) = %q, want %q", got, "20")
	}
}

func TestReflectGetValueIntPointer(t *testing.T) {
	i := 100
	cfg := &cfgtypes.GlobalConfig{
		Security: &security.Settings{PidsLimit: &i},
	}
	got := reflectGetValue(cfg, "security.pids_limit")
	if got != "100" {
		t.Errorf("reflectGetValue(security.pids_limit) = %q, want %q", got, "100")
	}
}

func TestReflectGetValueStringPointer(t *testing.T) {
	ep := "http://localhost:4318"
	cfg := &cfgtypes.GlobalConfig{
		Otel: &otel.Settings{Endpoint: &ep},
	}
	got := reflectGetValue(cfg, "otel.endpoint")
	if got != "http://localhost:4318" {
		t.Errorf("reflectGetValue(otel.endpoint) = %q, want %q", got, "http://localhost:4318")
	}

	// nil pointer
	cfg2 := &cfgtypes.GlobalConfig{Otel: &otel.Settings{}}
	if got := reflectGetValue(cfg2, "otel.endpoint"); got != "" {
		t.Errorf("reflectGetValue with nil endpoint = %q, want empty", got)
	}
}

func TestReflectGetValueStringSlice(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{
		SSH: &cfgtypes.SSHSettings{AllowedKeys: []string{"id_rsa", "id_ed25519"}},
	}
	got := reflectGetValue(cfg, "ssh.allowed_keys")
	if got != "id_rsa,id_ed25519" {
		t.Errorf("reflectGetValue(ssh.allowed_keys) = %q, want %q", got, "id_rsa,id_ed25519")
	}
}

func TestReflectSetValueBoolPointer(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "firewall.enabled", "true")
	if cfg.Firewall == nil || cfg.Firewall.Enabled == nil || *cfg.Firewall.Enabled != true {
		t.Error("firewall.enabled not set correctly")
	}
}

func TestReflectSetValueString(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "node_version", "18")
	if cfg.NodeVersion != "18" {
		t.Errorf("node_version = %q, want %q", cfg.NodeVersion, "18")
	}
}

func TestReflectSetValueIntPointer(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "security.pids_limit", "500")
	if cfg.Security == nil || cfg.Security.PidsLimit == nil || *cfg.Security.PidsLimit != 500 {
		t.Error("security.pids_limit not set correctly")
	}
}

func TestReflectSetValueStringPointer(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "otel.endpoint", "http://example.com:4318")
	if cfg.Otel == nil || cfg.Otel.Endpoint == nil || *cfg.Otel.Endpoint != "http://example.com:4318" {
		t.Error("otel.endpoint not set correctly")
	}
}

func TestReflectSetValueStringSlice(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "ssh.allowed_keys", "id_rsa,id_ed25519")
	if cfg.SSH == nil || len(cfg.SSH.AllowedKeys) != 2 {
		t.Fatal("ssh.allowed_keys not set correctly")
	}
	if cfg.SSH.AllowedKeys[0] != "id_rsa" || cfg.SSH.AllowedKeys[1] != "id_ed25519" {
		t.Errorf("ssh.allowed_keys = %v", cfg.SSH.AllowedKeys)
	}

	// Empty value should nil the slice
	reflectSetValue(cfg, "ssh.allowed_keys", "")
	if cfg.SSH.AllowedKeys != nil {
		t.Error("ssh.allowed_keys should be nil after setting empty")
	}
}

func TestReflectUnsetValueBoolPointer(t *testing.T) {
	b := true
	cfg := &cfgtypes.GlobalConfig{
		Firewall: &cfgtypes.FirewallSettings{Enabled: &b},
	}
	reflectUnsetValue(cfg, "firewall.enabled")
	if cfg.Firewall.Enabled != nil {
		t.Error("firewall.enabled should be nil after unset")
	}
}

func TestReflectUnsetValueString(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{NodeVersion: "20"}
	reflectUnsetValue(cfg, "node_version")
	if cfg.NodeVersion != "" {
		t.Errorf("node_version = %q, want empty", cfg.NodeVersion)
	}
}

func TestReflectUnsetValueIntPointer(t *testing.T) {
	i := 200
	cfg := &cfgtypes.GlobalConfig{
		Security: &security.Settings{PidsLimit: &i},
	}
	reflectUnsetValue(cfg, "security.pids_limit")
	if cfg.Security.PidsLimit != nil {
		t.Error("security.pids_limit should be nil after unset")
	}
}

func TestReflectUnsetValueStringSlice(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{
		Security: &security.Settings{CapAdd: []string{"CHOWN", "SETUID"}},
	}
	reflectUnsetValue(cfg, "security.cap_add")
	if cfg.Security.CapAdd != nil {
		t.Error("security.cap_add should be nil after unset")
	}
}

func TestReflectDeepNesting(t *testing.T) {
	// docker.dind.enable is 3 levels deep
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "docker.dind.enable", "true")
	if cfg.Docker == nil || cfg.Docker.Dind == nil || cfg.Docker.Dind.Enable == nil || *cfg.Docker.Dind.Enable != true {
		t.Error("docker.dind.enable not set correctly through 3 levels")
	}

	got := reflectGetValue(cfg, "docker.dind.enable")
	if got != "true" {
		t.Errorf("reflectGetValue(docker.dind.enable) = %q, want %q", got, "true")
	}

	reflectUnsetValue(cfg, "docker.dind.enable")
	if cfg.Docker.Dind.Enable != nil {
		t.Error("docker.dind.enable should be nil after unset")
	}
}

func TestReflectNilParentAllocation(t *testing.T) {
	// Setting a value on a nil sub-struct should allocate it
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "auth.autologin", "true")
	if cfg.Auth == nil {
		t.Fatal("Auth should be allocated after set")
	}
	if cfg.Auth.Autologin == nil || *cfg.Auth.Autologin != true {
		t.Error("auth.autologin not set correctly")
	}
}

func TestReflectRootLevelKeys(t *testing.T) {
	// Test root-level keys (1-segment path)
	cfg := &cfgtypes.GlobalConfig{}

	reflectSetValue(cfg, "env_file_load", "false")
	if cfg.EnvFileLoad == nil || *cfg.EnvFileLoad != false {
		t.Error("env_file_load not set correctly")
	}

	reflectSetValue(cfg, "persistent", "true")
	if cfg.Persistent == nil || *cfg.Persistent != true {
		t.Error("persistent not set correctly")
	}

	reflectSetValue(cfg, "node_version", "20")
	if cfg.NodeVersion != "20" {
		t.Errorf("node_version = %q, want %q", cfg.NodeVersion, "20")
	}
}

func TestReflectStringListCommaHandling(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "security.cap_add", "CHOWN,SETUID,SETGID")
	if cfg.Security == nil || len(cfg.Security.CapAdd) != 3 {
		t.Fatal("security.cap_add not set correctly")
	}
	got := reflectGetValue(cfg, "security.cap_add")
	if got != "CHOWN,SETUID,SETGID" {
		t.Errorf("reflectGetValue(security.cap_add) = %q, want %q", got, "CHOWN,SETUID,SETGID")
	}
}

func TestReflectGitHubScopeRepos(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "github.scope_repos", "owner/repo1,owner/repo2")
	if cfg.GitHub == nil || len(cfg.GitHub.ScopeRepos) != 2 {
		t.Fatal("github.scope_repos not set correctly")
	}
	got := reflectGetValue(cfg, "github.scope_repos")
	if got != "owner/repo1,owner/repo2" {
		t.Errorf("got %q", got)
	}
}

func TestReflectPortsExpose(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "ports.expose", "3000, 8080")
	if cfg.Ports == nil || len(cfg.Ports.Expose) != 2 {
		t.Fatal("ports.expose not set correctly")
	}
	// string_list type trims spaces
	if cfg.Ports.Expose[0] != "3000" || cfg.Ports.Expose[1] != "8080" {
		t.Errorf("ports.expose = %v, expected trimmed values", cfg.Ports.Expose)
	}
}

func TestReflectPortsRangeStart(t *testing.T) {
	i := 35000
	cfg := &cfgtypes.GlobalConfig{
		Ports: &cfgtypes.PortsSettings{RangeStart: &i},
	}
	got := reflectGetValue(cfg, "ports.range_start")
	if got != "35000" {
		t.Errorf("reflectGetValue(ports.range_start) = %q, want %q", got, "35000")
	}
}

func TestReflectLogMaxFiles(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "log.max_files", "10")
	if cfg.Log == nil || cfg.Log.MaxFiles == nil || *cfg.Log.MaxFiles != 10 {
		t.Error("log.max_files not set correctly")
	}
}

func TestReflectWorkdirSettings(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "workdir.path", "/custom/path")
	if cfg.Workdir == nil || cfg.Workdir.Path != "/custom/path" {
		t.Error("workdir.path not set correctly")
	}

	reflectSetValue(cfg, "workdir.automount", "true")
	if cfg.Workdir.Automount == nil || *cfg.Workdir.Automount != true {
		t.Error("workdir.automount not set correctly")
	}
}

func TestReflectSecurityYolo(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}
	reflectSetValue(cfg, "security.yolo", "true")
	if cfg.Security == nil || cfg.Security.Yolo == nil || *cfg.Security.Yolo != true {
		t.Error("security.yolo not set correctly")
	}
	got := reflectGetValue(cfg, "security.yolo")
	if got != "true" {
		t.Errorf("reflectGetValue(security.yolo) = %q, want %q", got, "true")
	}
}
