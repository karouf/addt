package config

import (
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestGetAuditGroups_Count(t *testing.T) {
	groups := GetAuditGroups()
	if len(groups) != 6 {
		t.Errorf("expected 6 audit groups, got %d", len(groups))
	}

	expected := []string{"Network", "Filesystem", "Credentials", "Limits", "Isolation", "Audit"}
	for i, g := range groups {
		if g.Name != expected[i] {
			t.Errorf("group %d: expected name %q, got %q", i, expected[i], g.Name)
		}
	}
}

func TestAuditKeysAreValid(t *testing.T) {
	groups := GetAuditGroups()
	for _, g := range groups {
		for _, key := range g.Keys {
			if !IsValidKey(key) {
				t.Errorf("group %q: key %q is not a valid config key", g.Name, key)
			}
		}
	}
}

// Helper to build a resolved map from key-value pairs.
func makeResolved(pairs map[string]string) map[string]ResolvedKey {
	m := make(map[string]ResolvedKey, len(pairs))
	for k, v := range pairs {
		m[k] = ResolvedKey{Key: k, Value: v, Source: "default"}
	}
	return m
}

func TestNetworkPosture_Secure(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"firewall.enabled":      "true",
		"firewall.mode":         "strict",
		"security.network_mode": "none",
		"docker.dind.enable":    "false",
	})
	posture := evaluateNetwork(resolved)
	if !posture.Secure {
		t.Errorf("expected secure posture, got relaxed; tags: %v", posture.Tags)
	}
}

func TestNetworkPosture_Relaxed(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"firewall.enabled":      "false",
		"firewall.mode":         "strict",
		"security.network_mode": "",
		"docker.dind.enable":    "false",
	})
	posture := evaluateNetwork(resolved)
	if posture.Secure {
		t.Errorf("expected relaxed posture, got secure; tags: %v", posture.Tags)
	}
}

func TestFilesystemPosture_Secure(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"workdir.automount":         "true",
		"workdir.readonly":          "true",
		"security.read_only_rootfs": "true",
		"config.automount":          "false",
		"config.readonly":           "false",
	})
	posture := evaluateFilesystem(resolved)
	if !posture.Secure {
		t.Errorf("expected secure posture, got relaxed; tags: %v", posture.Tags)
	}
}

func TestFilesystemPosture_Relaxed(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"workdir.automount":         "true",
		"workdir.readonly":          "false",
		"security.read_only_rootfs": "false",
		"config.automount":          "false",
		"config.readonly":           "false",
	})
	posture := evaluateFilesystem(resolved)
	if posture.Secure {
		t.Errorf("expected relaxed posture, got secure; tags: %v", posture.Tags)
	}
}

func TestCredentialsPosture_Secure(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"ssh.forward_keys":         "true",
		"ssh.forward_mode":         "proxy",
		"github.forward_token":     "true",
		"github.scope_token":       "true",
		"security.isolate_secrets": "true",
	})
	posture := evaluateCredentials(resolved)
	if !posture.Secure {
		t.Errorf("expected secure posture, got relaxed; tags: %v", posture.Tags)
	}
}

func TestCredentialsPosture_Relaxed(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"ssh.forward_keys":         "true",
		"ssh.forward_mode":         "agent",
		"github.forward_token":     "true",
		"github.scope_token":       "false",
		"security.isolate_secrets": "false",
	})
	posture := evaluateCredentials(resolved)
	if posture.Secure {
		t.Errorf("expected relaxed posture, got secure; tags: %v", posture.Tags)
	}
}

func TestLimitsPosture_Secure(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"container.cpus":      "2",
		"container.memory":    "4g",
		"security.pids_limit": "200",
		"security.time_limit": "3600",
	})
	posture := evaluateLimits(resolved)
	if !posture.Secure {
		t.Errorf("expected secure posture, got relaxed; tags: %v", posture.Tags)
	}
}

func TestLimitsPosture_Relaxed(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"container.cpus":      "2",
		"container.memory":    "4g",
		"security.pids_limit": "200",
		"security.time_limit": "0",
	})
	posture := evaluateLimits(resolved)
	if posture.Secure {
		t.Errorf("expected relaxed posture, got secure; tags: %v", posture.Tags)
	}
}

func TestIsolationPosture_Secure(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"security.no_new_privileges": "true",
		"security.cap_drop":          "ALL",
		"security.cap_add":           "CHOWN,SETUID,SETGID",
		"security.yolo":              "false",
		"git.disable_hooks":          "true",
		"security.seccomp_profile":   "",
		"security.user_namespace":    "",
		"security.disable_devices":   "false",
		"security.disable_ipc":       "false",
	})
	posture := evaluateIsolation(resolved)
	if !posture.Secure {
		t.Errorf("expected secure posture, got relaxed; tags: %v", posture.Tags)
	}
}

func TestIsolationPosture_Relaxed(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"security.no_new_privileges": "false",
		"security.cap_drop":          "",
		"security.cap_add":           "ALL",
		"security.yolo":              "true",
		"git.disable_hooks":          "false",
		"security.seccomp_profile":   "",
		"security.user_namespace":    "",
		"security.disable_devices":   "false",
		"security.disable_ipc":       "false",
	})
	posture := evaluateIsolation(resolved)
	if posture.Secure {
		t.Errorf("expected relaxed posture, got secure; tags: %v", posture.Tags)
	}
}

func TestAuditPosture_Secure(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"security.audit_log":      "true",
		"security.audit_log_file": "/var/log/addt.log",
	})
	posture := evaluateAuditGroup(resolved)
	if !posture.Secure {
		t.Errorf("expected secure posture, got relaxed; tags: %v", posture.Tags)
	}
}

func TestAuditPosture_Relaxed(t *testing.T) {
	resolved := makeResolved(map[string]string{
		"security.audit_log":      "false",
		"security.audit_log_file": "",
	})
	posture := evaluateAuditGroup(resolved)
	if posture.Secure {
		t.Errorf("expected relaxed posture, got secure; tags: %v", posture.Tags)
	}
}

func TestRunAudit_AllDefaults(t *testing.T) {
	// Empty configs = all defaults
	projectCfg := &cfgtypes.GlobalConfig{}
	globalCfg := &cfgtypes.GlobalConfig{}

	result := RunAudit(projectCfg, globalCfg)

	if result.TotalGroups != 6 {
		t.Errorf("expected 6 total groups, got %d", result.TotalGroups)
	}

	// With all defaults: Network (relaxed), Filesystem (relaxed), Credentials (secure),
	// Limits (relaxed - time_limit=0), Isolation (secure), Audit (relaxed)
	expectedRelaxed := 4
	if result.RelaxedCount != expectedRelaxed {
		t.Errorf("expected %d relaxed groups with defaults, got %d", expectedRelaxed, result.RelaxedCount)
		for _, gr := range result.Groups {
			t.Logf("  %s: secure=%v tags=%v", gr.Group.Name, gr.Posture.Secure, gr.Posture.Tags)
		}
	}

	// Verify specific groups
	for _, gr := range result.Groups {
		switch gr.Group.Name {
		case "Credentials":
			if !gr.Posture.Secure {
				t.Errorf("Credentials should be secure with defaults, got relaxed; tags: %v", gr.Posture.Tags)
			}
		case "Isolation":
			if !gr.Posture.Secure {
				t.Errorf("Isolation should be secure with defaults, got relaxed; tags: %v", gr.Posture.Tags)
			}
		case "Network":
			if gr.Posture.Secure {
				t.Errorf("Network should be relaxed with defaults (firewall off), got secure")
			}
		case "Audit":
			if gr.Posture.Secure {
				t.Errorf("Audit should be relaxed with defaults (audit_log false), got secure")
			}
		}
	}
}
