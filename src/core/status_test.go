package core

import (
	"strings"
	"testing"

	"github.com/jedi4ever/addt/config/security"
	"github.com/jedi4ever/addt/provider"
)

func TestSecurityPostureLine_DefaultConfig(t *testing.T) {
	// Default config has relaxed settings: firewall off, audit off, rootfs rw, network bridge
	cfg := &provider.Config{
		WorkdirAutomount: true,
		Security:         security.DefaultConfig(),
	}

	posture, allLocked := SecurityPostureLine(cfg)

	if allLocked {
		t.Error("Default config should NOT be fully locked down")
	}

	// Should contain relaxed indicators
	if !strings.Contains(posture, "firewall:off") {
		t.Errorf("Expected firewall:off, got %q", posture)
	}
	if !strings.Contains(posture, "network:bridge") {
		t.Errorf("Expected network:bridge, got %q", posture)
	}
	if !strings.Contains(posture, "workdir:rw") {
		t.Errorf("Expected workdir:rw (automount default with rw), got %q", posture)
	}
	if !strings.Contains(posture, "rootfs:rw") {
		t.Errorf("Expected rootfs:rw, got %q", posture)
	}
	if !strings.Contains(posture, "audit:off") {
		t.Errorf("Expected audit:off, got %q", posture)
	}

	// Secrets should be isolated by default, so "secrets:exposed" should NOT appear
	if strings.Contains(posture, "secrets:exposed") {
		t.Errorf("Default config has IsolateSecrets=true, should not show secrets:exposed, got %q", posture)
	}

	// Default pids (200) should not be shown
	if strings.Contains(posture, "pids:") {
		t.Errorf("Default pids limit should not be shown, got %q", posture)
	}
}

func TestSecurityPostureLine_StrictConfig(t *testing.T) {
	// Strict-like config: firewall on, but still has network bridge and rootfs rw
	sec := security.DefaultConfig()
	sec.ReadOnlyRootfs = true
	sec.AuditLog = true

	cfg := &provider.Config{
		FirewallEnabled:  true,
		FirewallMode:     "strict",
		WorkdirAutomount: true,
		WorkdirReadonly:  true,
		Security:         sec,
	}

	posture, allLocked := SecurityPostureLine(cfg)

	if allLocked {
		t.Error("Strict config with network:bridge should NOT be fully locked down")
	}

	if !strings.Contains(posture, "firewall:strict") {
		t.Errorf("Expected firewall:strict, got %q", posture)
	}
	if !strings.Contains(posture, "workdir:ro") {
		t.Errorf("Expected workdir:ro, got %q", posture)
	}
	if !strings.Contains(posture, "rootfs:ro") {
		t.Errorf("Expected rootfs:ro, got %q", posture)
	}
	if !strings.Contains(posture, "audit:on") {
		t.Errorf("Expected audit:on, got %q", posture)
	}

	// Still has bridge network so not fully locked
	if !strings.Contains(posture, "network:bridge") {
		t.Errorf("Expected network:bridge, got %q", posture)
	}
}

func TestSecurityPostureLine_ParanoiaConfig(t *testing.T) {
	// Paranoia config: everything locked down
	sec := security.DefaultConfig()
	sec.NetworkMode = "none"
	sec.ReadOnlyRootfs = true
	sec.AuditLog = true
	sec.TimeLimit = 120
	sec.PidsLimit = 100
	sec.IsolateSecrets = true

	cfg := &provider.Config{
		FirewallEnabled:  true,
		FirewallMode:     "strict",
		WorkdirAutomount: true,
		WorkdirReadonly:  true,
		Security:         sec,
	}

	posture, allLocked := SecurityPostureLine(cfg)

	if !allLocked {
		t.Error("Paranoia config should be fully locked down")
	}

	if !strings.Contains(posture, "firewall:strict") {
		t.Errorf("Expected firewall:strict, got %q", posture)
	}
	if !strings.Contains(posture, "network:none") {
		t.Errorf("Expected network:none, got %q", posture)
	}
	if !strings.Contains(posture, "workdir:ro") {
		t.Errorf("Expected workdir:ro, got %q", posture)
	}
	if !strings.Contains(posture, "rootfs:ro") {
		t.Errorf("Expected rootfs:ro, got %q", posture)
	}
	if !strings.Contains(posture, "audit:on") {
		t.Errorf("Expected audit:on, got %q", posture)
	}
	if !strings.Contains(posture, "time:120m") {
		t.Errorf("Expected time:120m, got %q", posture)
	}
	if !strings.Contains(posture, "pids:100") {
		t.Errorf("Expected pids:100, got %q", posture)
	}
}

func TestSecurityPostureLine_WorkdirNotMounted(t *testing.T) {
	cfg := &provider.Config{
		WorkdirAutomount: false,
		Security:         security.DefaultConfig(),
	}

	posture, _ := SecurityPostureLine(cfg)

	if !strings.Contains(posture, "workdir:none") {
		t.Errorf("Expected workdir:none when not automounted, got %q", posture)
	}
}

func TestSecurityPostureLine_SecretsExposed(t *testing.T) {
	sec := security.DefaultConfig()
	sec.IsolateSecrets = false

	cfg := &provider.Config{
		Security: sec,
	}

	posture, allLocked := SecurityPostureLine(cfg)

	if allLocked {
		t.Error("Config with exposed secrets should NOT be fully locked down")
	}

	if !strings.Contains(posture, "secrets:exposed") {
		t.Errorf("Expected secrets:exposed, got %q", posture)
	}
}

func TestSecurityPostureLine_NonDefaultPids(t *testing.T) {
	sec := security.DefaultConfig()
	sec.PidsLimit = 50

	cfg := &provider.Config{
		Security: sec,
	}

	posture, _ := SecurityPostureLine(cfg)

	if !strings.Contains(posture, "pids:50") {
		t.Errorf("Expected pids:50, got %q", posture)
	}
}

func TestSecurityPostureLine_FirewallPermissive(t *testing.T) {
	cfg := &provider.Config{
		FirewallEnabled: true,
		FirewallMode:    "permissive",
		Security:        security.DefaultConfig(),
	}

	posture, allLocked := SecurityPostureLine(cfg)

	if allLocked {
		t.Error("Permissive firewall should NOT be fully locked down")
	}

	if !strings.Contains(posture, "firewall:permissive") {
		t.Errorf("Expected firewall:permissive, got %q", posture)
	}
}
