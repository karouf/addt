//go:build addt

package addt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Helper to extract run test markers from subprocess output ---

func extractRunResult(output, marker string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, marker) {
			return strings.TrimPrefix(line, marker)
		}
	}
	return ""
}

// --- Container tests (subprocess, both providers) ---

func TestRun_Addt_BasicExecution(t *testing.T) {
	// Scenario: User runs `addt run debug -c "echo RUN_TEST:hello"`.
	// The debug extension has entrypoint /bin/bash, so this executes
	// a bash command inside the container. Verifies the run subcommand
	// path works end-to-end.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo RUN_TEST:hello")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractRunResult(output, "RUN_TEST:")
			if result != "hello" {
				t.Errorf("Expected RUN_TEST:hello, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestRun_Addt_WorkdirMounted(t *testing.T) {
	// Scenario: User creates a marker file in the project directory, then
	// runs `addt run debug` to check the file exists at /workspace/.
	// This confirms workdir mounting works through the run subcommand path.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// Create a marker file in the project directory
			markerFile := filepath.Join(dir, "run_test_marker.txt")
			if err := os.WriteFile(markerFile, []byte("MARKER_FOUND"), 0o644); err != nil {
				t.Fatalf("Failed to write marker file: %v", err)
			}

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "cat /workspace/run_test_marker.txt && echo WORKDIR_OK:yes || echo WORKDIR_OK:no")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractRunResult(output, "WORKDIR_OK:")
			if result != "yes" {
				t.Errorf("Expected WORKDIR_OK:yes, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestRun_Addt_EntrypointIsExtension(t *testing.T) {
	// Scenario: User runs `addt run debug` and checks that ADDT_COMMAND
	// is set to /bin/bash (the debug extension's entrypoint), confirming
	// the run path uses the extension's entrypoint rather than a hardcoded value.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo RUN_CMD:$ADDT_COMMAND")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractRunResult(output, "RUN_CMD:")
			if result != "/bin/bash" {
				t.Errorf("Expected ADDT_COMMAND=/bin/bash, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestRun_Addt_EnvVarsForwarded(t *testing.T) {
	// Scenario: User configures a custom env var in project config, then runs
	// `addt run debug`. The env var should be available inside the container,
	// confirming env forwarding works through the run subcommand path.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
env:
  - "RUN_TEST_VAR=myvalue"
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo ENVVAR:${RUN_TEST_VAR:-NOTSET}")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractRunResult(output, "ENVVAR:")
			if result != "myvalue" {
				t.Errorf("Expected RUN_TEST_VAR=myvalue, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestRun_Addt_UserIsAddt(t *testing.T) {
	// Scenario: User runs `whoami` inside the container via `addt run debug`
	// and expects the output to be "addt" (not root), confirming user mapping
	// works through the run subcommand path.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo WHOAMI:$(whoami)")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractRunResult(output, "WHOAMI:")
			if result != "addt" {
				t.Errorf("Expected whoami=addt, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}
