//go:build addt

package addt

import (
	"testing"
)

// Scenario: A user sets global auth.autologin to true in project config.
// The entrypoint should pass ADDT_EXT_AUTH_AUTOLOGIN=true to the extension.
func TestAuth_Addt_GlobalAutologinDefault(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			// No auth config — defaults should apply (autologin=true, method=auto)
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOLOGIN:${ADDT_EXT_AUTH_AUTOLOGIN:-UNSET} && echo METHOD:${ADDT_EXT_AUTH_METHOD:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			// Claude extension default: autologin=true, method=env
			autologin := extractMarker(output, "AUTOLOGIN:")
			if autologin != "true" {
				t.Errorf("Expected AUTOLOGIN:true (extension default), got AUTOLOGIN:%s\nFull output:\n%s",
					autologin, output)
			}

			method := extractMarker(output, "METHOD:")
			if method != "env" {
				t.Errorf("Expected METHOD:env (claude extension default), got METHOD:%s\nFull output:\n%s",
					method, output)
			}
		})
	}
}

// Scenario: A user overrides auth.autologin to false at the global level.
// All extensions should receive autologin=false.
func TestAuth_Addt_GlobalAutologinDisabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
auth:
  autologin: false
  method: native
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOLOGIN:${ADDT_EXT_AUTH_AUTOLOGIN:-UNSET} && echo METHOD:${ADDT_EXT_AUTH_METHOD:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			autologin := extractMarker(output, "AUTOLOGIN:")
			if autologin != "false" {
				t.Errorf("Expected AUTOLOGIN:false (global override), got AUTOLOGIN:%s\nFull output:\n%s",
					autologin, output)
			}

			method := extractMarker(output, "METHOD:")
			if method != "native" {
				t.Errorf("Expected METHOD:native (global override), got METHOD:%s\nFull output:\n%s",
					method, output)
			}
		})
	}
}

// Scenario: A user sets per-extension auth override that differs from global.
// The per-extension setting should take precedence.
func TestAuth_Addt_PerExtensionOverridesGlobal(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			dir, cleanup := setupAddtDir(t, prov, `
auth:
  autologin: true
  method: auto
extensions:
  claude:
    auth:
      autologin: false
      method: native
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOLOGIN:${ADDT_EXT_AUTH_AUTOLOGIN:-UNSET} && echo METHOD:${ADDT_EXT_AUTH_METHOD:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			// Per-extension should override global
			autologin := extractMarker(output, "AUTOLOGIN:")
			if autologin != "false" {
				t.Errorf("Expected AUTOLOGIN:false (per-extension override), got AUTOLOGIN:%s\nFull output:\n%s",
					autologin, output)
			}

			method := extractMarker(output, "METHOD:")
			if method != "native" {
				t.Errorf("Expected METHOD:native (per-extension override), got METHOD:%s\nFull output:\n%s",
					method, output)
			}
		})
	}
}

// Scenario: A user sets auth via environment variables.
// Env vars should override config file settings.
func TestAuth_Addt_EnvVarOverridesConfig(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			defer setDummyAnthropicKey(t)()

			// Config says autologin=true, method=env
			dir, cleanup := setupAddtDir(t, prov, `
auth:
  autologin: true
  method: env
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Per-extension env var overrides config
			t.Setenv("ADDT_CLAUDE_AUTH_AUTOLOGIN", "false")
			t.Setenv("ADDT_CLAUDE_AUTH_METHOD", "native")

			output, err := runShellCommand(t, dir,
				"claude", "-c",
				"echo AUTOLOGIN:${ADDT_EXT_AUTH_AUTOLOGIN:-UNSET} && echo METHOD:${ADDT_EXT_AUTH_METHOD:-UNSET}")
			if err != nil {
				t.Fatalf("shell command failed: %v\nOutput: %s", err, output)
			}

			autologin := extractMarker(output, "AUTOLOGIN:")
			if autologin != "false" {
				t.Errorf("Expected AUTOLOGIN:false (env var override), got AUTOLOGIN:%s\nFull output:\n%s",
					autologin, output)
			}

			method := extractMarker(output, "METHOD:")
			if method != "native" {
				t.Errorf("Expected METHOD:native (env var override), got METHOD:%s\nFull output:\n%s",
					method, output)
			}
		})
	}
}
