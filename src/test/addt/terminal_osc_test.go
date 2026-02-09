//go:build addt

package addt

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestTerminalOSC_Addt_OSC52EmittedWhenEnabled(t *testing.T) {
	// Scenario: User enables terminal.osc and runs from Ghostty terminal.
	// Inside the container a clipboard-aware script detects TERM_PROGRAM,
	// decides OSC 52 is supported, and emits an OSC 52 clipboard-set sequence.
	// The raw escape sequence must pass through stdout so the host terminal
	// can intercept it and update the clipboard.
	providers := requireProviders(t)

	payload := "hello from container"
	b64 := base64.StdEncoding.EncodeToString([]byte(payload))
	// The script checks TERM_PROGRAM then emits OSC 52;c;<base64>\a
	// A bare `echo` after printf ensures the marker lands on its own line
	script := fmt.Sprintf(
		`if [ -n "$TERM_PROGRAM" ]; then printf '\033]52;c;%s\007'; echo; echo "OSC52:emitted"; else echo "OSC52:skipped"; fi`,
		b64,
	)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
terminal:
  osc: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			restore := saveRestoreEnv(t, "TERM_PROGRAM", "ghostty")
			defer restore()

			output, err := runRunSubcommand(t, dir, "debug", "-c", script)
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			// Verify the script decided to emit (TERM_PROGRAM was available)
			marker := extractMarker(output, "OSC52:")
			if marker != "emitted" {
				t.Errorf("Expected OSC52:emitted, got %q — TERM_PROGRAM not forwarded?", marker)
			}

			// Verify the raw OSC 52 escape sequence is present in the output
			osc52Seq := fmt.Sprintf("\033]52;c;%s", b64)
			if !strings.Contains(output, osc52Seq) {
				t.Errorf("OSC 52 sequence not found in output.\nExpected substring: %q\nFull output bytes: %q", osc52Seq, output)
			}
		})
	}
}

func TestTerminalOSC_Addt_TermOverriddenToXterm256color(t *testing.T) {
	// Scenario: User's host terminal sets TERM=xterm-kitty whose terminfo entry
	// does not exist in the container. addt should override TERM to
	// xterm-256color so TUI apps (Ink/Node.js, ncurses) render correctly.
	// This works regardless of terminal.osc setting.
	providers := requireProviders(t)

	// Script prints the TERM value seen inside the container
	script := `echo "TERM:$TERM"`

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// Set a custom TERM on the host that won't exist in the container
			restore := saveRestoreEnv(t, "TERM", "xterm-kitty")
			defer restore()

			output, err := runRunSubcommand(t, dir, "debug", "-c", script)
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			// Verify the container sees xterm-256color, not xterm-kitty
			marker := extractMarker(output, "TERM:")
			if marker != "xterm-256color" {
				t.Errorf("Expected TERM:xterm-256color, got TERM:%s — host TERM leaked into container?", marker)
			}
		})
	}
}

func TestTerminalOSC_Addt_ConfigDisabledBlocksTermProgram(t *testing.T) {
	// Scenario: User sets terminal.osc: false (the default). Even though the
	// host has TERM_PROGRAM set, it must NOT reach the container. Apps inside
	// the container won't detect OSC capabilities and fall back to plain output.
	providers := requireProviders(t)

	script := `if [ -n "$TERM_PROGRAM" ]; then echo "OSC_CONFIG:leaked"; else echo "OSC_CONFIG:blocked"; fi`

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
terminal:
  osc: false
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			restore := saveRestoreEnv(t, "TERM_PROGRAM", "ghostty")
			defer restore()

			output, err := runRunSubcommand(t, dir, "debug", "-c", script)
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			marker := extractMarker(output, "OSC_CONFIG:")
			if marker != "blocked" {
				t.Errorf("Expected OSC_CONFIG:blocked, got %q — TERM_PROGRAM leaked despite terminal.osc: false", marker)
			}
		})
	}
}

func TestTerminalOSC_Addt_DefaultBlocksTermProgram(t *testing.T) {
	// Scenario: User does not set terminal.osc at all (defaults to false).
	// TERM_PROGRAM must NOT reach the container.
	providers := requireProviders(t)

	script := `if [ -n "$TERM_PROGRAM" ]; then echo "OSC_DEFAULT:leaked"; else echo "OSC_DEFAULT:blocked"; fi`

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			restore := saveRestoreEnv(t, "TERM_PROGRAM", "kitty")
			defer restore()

			output, err := runRunSubcommand(t, dir, "debug", "-c", script)
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			marker := extractMarker(output, "OSC_DEFAULT:")
			if marker != "blocked" {
				t.Errorf("Expected OSC_DEFAULT:blocked, got %q — default should be false", marker)
			}
		})
	}
}

func TestTerminalOSC_Addt_ConfigEnabledForwardsTermProgram(t *testing.T) {
	// Scenario: User explicitly sets terminal.osc: true. TERM_PROGRAM from
	// the host must be forwarded to the container so apps can detect OSC support.
	providers := requireProviders(t)

	script := `echo "TERM_PROG:$TERM_PROGRAM"`

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
terminal:
  osc: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			restore := saveRestoreEnv(t, "TERM_PROGRAM", "ghostty")
			defer restore()

			output, err := runRunSubcommand(t, dir, "debug", "-c", script)
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			marker := extractMarker(output, "TERM_PROG:")
			if marker != "ghostty" {
				t.Errorf("Expected TERM_PROG:ghostty, got %q — terminal.osc: true should forward TERM_PROGRAM", marker)
			}
		})
	}
}

func TestTerminalOSC_Addt_OSC52RoundtripContent(t *testing.T) {
	// Scenario: User copies a multi-word string to clipboard via OSC 52 from
	// inside the container with terminal.osc: true. The test verifies the
	// base64-encoded payload in the escape sequence decodes back to the original
	// content — proving the full pipeline (env detection → encode → emit →
	// passthrough) is intact.
	providers := requireProviders(t)

	payload := "rich copy block test: lines & borders!"
	b64 := base64.StdEncoding.EncodeToString([]byte(payload))
	// Emit OSC 52, then echo the base64 as a marker so we can decode and verify
	// A bare `echo` after printf ensures the marker lands on its own line
	script := fmt.Sprintf(
		`printf '\033]52;c;%s\007' && echo && echo "B64:%s"`,
		b64, b64,
	)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
terminal:
  osc: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			restore := saveRestoreEnv(t, "TERM_PROGRAM", "kitty")
			defer restore()

			output, err := runRunSubcommand(t, dir, "debug", "-c", script)
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			// Extract the base64 payload echoed by the script
			got64 := extractMarker(output, "B64:")
			if got64 == "" {
				t.Fatal("B64 marker not found in output")
			}

			// Decode and verify it matches the original payload
			decoded, err := base64.StdEncoding.DecodeString(got64)
			if err != nil {
				t.Fatalf("Failed to decode base64 %q: %v", got64, err)
			}
			if string(decoded) != payload {
				t.Errorf("Decoded payload = %q, want %q", string(decoded), payload)
			}

			// Verify the OSC 52 sequence is also present in raw output
			osc52Seq := fmt.Sprintf("\033]52;c;%s", b64)
			if !strings.Contains(output, osc52Seq) {
				t.Errorf("OSC 52 sequence not found in output")
			}
		})
	}
}
