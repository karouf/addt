//go:build addt

package addt

import (
	"testing"
)

// Scenario: A user enables workdir.autotrust globally. The entrypoint
// should set ADDT_EXT_WORKDIR_AUTOTRUST=true for the extension.
func TestAutotrust_Addt_GlobalEnabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
workdir:
  autotrust: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOTRUST:${ADDT_EXT_WORKDIR_AUTOTRUST:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "AUTOTRUST:")
			if result != "true" {
				t.Errorf("Expected AUTOTRUST:true, got AUTOTRUST:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: A user does not set workdir.autotrust. The default should
// be false (not trusting workspace automatically).
func TestAutotrust_Addt_DefaultDisabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOTRUST:${ADDT_EXT_WORKDIR_AUTOTRUST:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "AUTOTRUST:")
			if result != "false" {
				t.Errorf("Expected AUTOTRUST:false (default), got AUTOTRUST:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: A user sets per-extension autotrust override that differs
// from the global setting. Per-extension should take precedence.
func TestAutotrust_Addt_PerExtensionOverridesGlobal(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
workdir:
  autotrust: false
extensions:
  claude:
    workdir:
      autotrust: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOTRUST:${ADDT_EXT_WORKDIR_AUTOTRUST:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "AUTOTRUST:")
			if result != "true" {
				t.Errorf("Expected AUTOTRUST:true (per-extension override), got AUTOTRUST:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: A user sets autotrust via environment variable.
// The env var should override config.
func TestAutotrust_Addt_EnvVarOverridesConfig(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			// Config says autotrust: false
			dir, cleanup := setupAddtDir(t, prov, `
workdir:
  autotrust: false
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Per-extension env var overrides
			t.Setenv("ADDT_CLAUDE_WORKDIR_AUTOTRUST", "true")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOTRUST:${ADDT_EXT_WORKDIR_AUTOTRUST:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "AUTOTRUST:")
			if result != "true" {
				t.Errorf("Expected AUTOTRUST:true (env var override), got AUTOTRUST:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: When autotrust is enabled for claude, the setup.sh should
// create the workspace trust entry in ~/.claude.json inside the container.
func TestAutotrust_Addt_ClaudeWorkspaceTrusted(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
workdir:
  autotrust: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Check that ~/.claude.json has the workspace trust entry
			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"cat ~/.claude.json | grep -o '/workspace' | head -1 | xargs -I{} echo TRUST_RESULT:{}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "TRUST_RESULT:")
			if result != "/workspace" {
				t.Errorf("Expected /workspace trust entry in ~/.claude.json, got TRUST_RESULT:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}

// Scenario: When autotrust is disabled for claude, the setup.sh should
// NOT create the workspace trust entry in ~/.claude.json.
func TestAutotrust_Addt_ClaudeWorkspaceNotTrusted(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
workdir:
  autotrust: false
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Check that ~/.claude.json does NOT have the workspace trust entry
			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"if grep -q '/workspace' ~/.claude.json 2>/dev/null; then echo TRUST_RESULT:TRUSTED; else echo TRUST_RESULT:NOT_TRUSTED; fi")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			result := extractMarker(output, "TRUST_RESULT:")
			if result != "NOT_TRUSTED" {
				t.Errorf("Expected workspace NOT trusted when autotrust=false, got TRUST_RESULT:%s\nFull output:\n%s",
					result, output)
			}
		})
	}
}
