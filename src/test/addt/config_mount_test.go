//go:build addt

package addt

import (
	"testing"
)

// Scenario: A user enables config.automount globally. The env var
// ADDT_CONFIG_AUTOMOUNT should be set to true inside the container.
func TestConfigMount_Addt_GlobalAutomountEnabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
config:
  automount: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOMOUNT:${ADDT_CONFIG_AUTOMOUNT:-UNSET} && echo READONLY:${ADDT_CONFIG_READONLY:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			automount := extractMarker(output, "AUTOMOUNT:")
			if automount != "true" {
				t.Errorf("Expected AUTOMOUNT:true, got AUTOMOUNT:%s\nFull output:\n%s",
					automount, output)
			}
		})
	}
}

// Scenario: Config automount defaults to false when not configured.
func TestConfigMount_Addt_AutomountDefaultDisabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOMOUNT:${ADDT_CONFIG_AUTOMOUNT:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			automount := extractMarker(output, "AUTOMOUNT:")
			if automount != "false" {
				t.Errorf("Expected AUTOMOUNT:false (default), got AUTOMOUNT:%s\nFull output:\n%s",
					automount, output)
			}
		})
	}
}

// Scenario: A user enables config.readonly globally. The env var
// ADDT_CONFIG_READONLY should be set to true inside the container.
func TestConfigMount_Addt_GlobalReadonlyEnabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
config:
  automount: true
  readonly: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo READONLY:${ADDT_CONFIG_READONLY:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			readonly := extractMarker(output, "READONLY:")
			if readonly != "true" {
				t.Errorf("Expected READONLY:true, got READONLY:%s\nFull output:\n%s",
					readonly, output)
			}
		})
	}
}

// Scenario: Config readonly defaults to false when not configured.
func TestConfigMount_Addt_ReadonlyDefaultDisabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo READONLY:${ADDT_CONFIG_READONLY:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			readonly := extractMarker(output, "READONLY:")
			if readonly != "false" {
				t.Errorf("Expected READONLY:false (default), got READONLY:%s\nFull output:\n%s",
					readonly, output)
			}
		})
	}
}

// Scenario: A user sets per-extension config automount override via env var.
func TestConfigMount_Addt_PerExtensionAutomountEnvVar(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			// Global config automount is off
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Per-extension env var turns it on
			t.Setenv("ADDT_CLAUDE_CONFIG_AUTOMOUNT", "true")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOMOUNT_OVERRIDE:${ADDT_CLAUDE_CONFIG_AUTOMOUNT:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "AUTOMOUNT_OVERRIDE:")
			if result != "true" {
				t.Errorf("Expected AUTOMOUNT_OVERRIDE:true (env var), got AUTOMOUNT_OVERRIDE:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: When config.automount is enabled for claude, the ~/.claude
// directory from the host should be mounted into the container.
func TestConfigMount_Addt_ClaudeConfigMounted(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
config:
  automount: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Check if ~/.claude exists as a mount point
			// When automount is on, the entrypoint should detect the mounted config
			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"if [ -d ~/.claude ]; then echo MOUNT_RESULT:EXISTS; else echo MOUNT_RESULT:MISSING; fi")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "MOUNT_RESULT:")
			if result != "EXISTS" {
				t.Errorf("Expected ~/.claude to exist when automount=true, got MOUNT_RESULT:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}
