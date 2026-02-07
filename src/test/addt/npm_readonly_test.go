//go:build addt

package addt

import (
	"os"
	"strings"
	"testing"

	configcmd "github.com/jedi4ever/addt/cmd/config"
)

// --- Config tests (in-process, no container needed) ---

func TestNpmReadonly_Addt_DefaultValue(t *testing.T) {
	// Scenario: User starts with no security config. The default value
	// for security.read_only_rootfs should be false with source=default.
	_, cleanup := setupAddtDir(t, "", ``)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "security.read_only_rootfs") {
			if !strings.Contains(line, "false") {
				t.Errorf("Expected security.read_only_rootfs default=false, got line: %s", line)
			}
			if !strings.Contains(line, "default") {
				t.Errorf("Expected security.read_only_rootfs source=default, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected security.read_only_rootfs key in config list, got:\n%s", output)
}

func TestNpmReadonly_Addt_ConfigLoaded(t *testing.T) {
	// Scenario: User sets security.read_only_rootfs: true in .addt.yaml,
	// then verifies it appears in config list with value=true and source=project.
	_, cleanup := setupAddtDir(t, "", `
security:
  read_only_rootfs: true
`)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "security.read_only_rootfs") {
			if !strings.Contains(line, "true") {
				t.Errorf("Expected security.read_only_rootfs=true, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected security.read_only_rootfs source=project, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected security.read_only_rootfs key in config list, got:\n%s", output)
}

func TestNpmReadonly_Addt_ConfigViaSet(t *testing.T) {
	// Scenario: User enables readonly root via 'config set' command,
	// then verifies it appears in config list with value=true and source=project.
	_, cleanup := setupAddtDir(t, "", ``)
	defer cleanup()

	captureOutput(t, func() {
		configcmd.HandleCommand([]string{"set", "security.read_only_rootfs", "true"})
	})

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "security.read_only_rootfs") {
			if !strings.Contains(line, "true") {
				t.Errorf("Expected security.read_only_rootfs=true after config set, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected security.read_only_rootfs source=project after config set, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected security.read_only_rootfs key in config list, got:\n%s", output)
}

// --- Container tests (subprocess, both providers) ---

func TestNpmReadonly_Addt_NpmInstallWithReadonly(t *testing.T) {
	// Scenario: User enables readonly root filesystem for security hardening.
	// They install a global npm package (cowsay) inside the container.
	// This should succeed because NPM_CONFIG_PREFIX points to /home/addt/.npm-global
	// which lives on the writable tmpfs mount, even though root is readonly.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
security:
  read_only_rootfs: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// Set env var for subprocess robustness
			origReadonly := os.Getenv("ADDT_SECURITY_READ_ONLY_ROOTFS")
			os.Setenv("ADDT_SECURITY_READ_ONLY_ROOTFS", "true")
			defer func() {
				if origReadonly != "" {
					os.Setenv("ADDT_SECURITY_READ_ONLY_ROOTFS", origReadonly)
				} else {
					os.Unsetenv("ADDT_SECURITY_READ_ONLY_ROOTFS")
				}
			}()

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "npm install -g cowsay && cowsay --version && echo NPM_INSTALL:ok")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("npm install with readonly root failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "NPM_INSTALL:")
			if result != "ok" {
				t.Errorf("Expected NPM_INSTALL:ok, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestNpmReadonly_Addt_NpmInstallWithoutReadonly(t *testing.T) {
	// Scenario: Baseline test — user installs a global npm package without
	// readonly root enabled. This confirms npm install works normally and
	// serves as a control for the readonly test.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "npm install -g cowsay && cowsay --version && echo NPM_INSTALL:ok")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("npm install without readonly root failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "NPM_INSTALL:")
			if result != "ok" {
				t.Errorf("Expected NPM_INSTALL:ok, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestNpmReadonly_Addt_ReadonlyRootWriteFails(t *testing.T) {
	// Scenario: User enables readonly root and verifies it is actually
	// effective. Attempting to write to a system path (/usr/local/testfile)
	// should fail, confirming the root filesystem is truly readonly.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
security:
  read_only_rootfs: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			origReadonly := os.Getenv("ADDT_SECURITY_READ_ONLY_ROOTFS")
			os.Setenv("ADDT_SECURITY_READ_ONLY_ROOTFS", "true")
			defer func() {
				if origReadonly != "" {
					os.Setenv("ADDT_SECURITY_READ_ONLY_ROOTFS", origReadonly)
				} else {
					os.Unsetenv("ADDT_SECURITY_READ_ONLY_ROOTFS")
				}
			}()

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "touch /usr/local/testfile 2>/dev/null && echo WRITE:ok || echo WRITE:denied")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("readonly root write test failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "WRITE:")
			if result != "denied" {
				t.Errorf("Expected WRITE:denied (root should be readonly), got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestNpmReadonly_Addt_TmpWritable(t *testing.T) {
	// Scenario: User enables readonly root and verifies that /tmp is still
	// writable. The readonly root feature mounts /tmp as a tmpfs, so writes
	// to /tmp should succeed even though the root filesystem is readonly.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
security:
  read_only_rootfs: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			origReadonly := os.Getenv("ADDT_SECURITY_READ_ONLY_ROOTFS")
			os.Setenv("ADDT_SECURITY_READ_ONLY_ROOTFS", "true")
			defer func() {
				if origReadonly != "" {
					os.Setenv("ADDT_SECURITY_READ_ONLY_ROOTFS", origReadonly)
				} else {
					os.Unsetenv("ADDT_SECURITY_READ_ONLY_ROOTFS")
				}
			}()

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "touch /tmp/test_writable && echo TMP:ok || echo TMP:denied")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("tmp writable test failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "TMP:")
			if result != "ok" {
				t.Errorf("Expected TMP:ok (/tmp should be writable via tmpfs), got %q\nFull output:\n%s", result, output)
			}
		})
	}
}
