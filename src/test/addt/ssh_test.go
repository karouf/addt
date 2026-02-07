//go:build addt

package addt

import (
	"runtime"
	"strings"
	"testing"

	configcmd "github.com/jedi4ever/addt/cmd/config"
)

func TestSSH_Addt_ConfigLoaded(t *testing.T) {
	// Config list runs in-process (no os.Exit on success path).
	_, cleanup := setupAddtDir(t, "", `
ssh:
  forward_keys: true
  forward_mode: "proxy"
`)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	if !strings.Contains(output, "ssh.forward_mode") {
		t.Errorf("Expected output to contain ssh.forward_mode, got:\n%s", output)
	}
	if !strings.Contains(output, "proxy") {
		t.Errorf("Expected output to contain 'proxy', got:\n%s", output)
	}
	if !strings.Contains(output, "project") {
		t.Errorf("Expected output to contain 'project' source, got:\n%s", output)
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ssh.forward_keys") {
			if !strings.Contains(line, "true") {
				t.Errorf("Expected ssh.forward_keys to be true, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected ssh.forward_keys source to be project, got line: %s", line)
			}
		}
	}
}

func TestSSH_Addt_ProxyMode(t *testing.T) {
	providers := requireProviders(t)
	requireSSHAgent(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
ssh:
  forward_keys: true
  forward_mode: "proxy"
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// Check safe SSH files are present
			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "ls -la /home/addt/.ssh/")
			if err != nil {
				t.Fatalf("run ls .ssh failed: %v\nOutput: %s", err, output)
			}

			if !strings.Contains(output, ".pub") &&
				!strings.Contains(output, "config") &&
				!strings.Contains(output, "known_hosts") {
				t.Logf("SSH directory contents:\n%s", output)
				t.Log("Warning: no .pub, config, or known_hosts found")
			}

			// Check SSH_AUTH_SOCK is set
			output2, err := runRunSubcommand(t, dir, "debug",
				"-c", "printenv SSH_AUTH_SOCK")
			if err != nil {
				t.Fatalf("run printenv SSH_AUTH_SOCK failed: %v\nOutput: %s", err, output2)
			}

			// Find the socket path in the output (filter test framework noise)
			found := false
			for _, line := range strings.Split(output2, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "/") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected SSH_AUTH_SOCK to be set in proxy mode, output:\n%s", output2)
			}
		})
	}
}

func TestSSH_Addt_KeysMode(t *testing.T) {
	providers := requireProviders(t)
	requireSSHAgent(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
ssh:
  forward_keys: true
  forward_mode: "keys"
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "ls -la /home/addt/.ssh/")
			if err != nil {
				t.Fatalf("run ls .ssh failed: %v\nOutput: %s", err, output)
			}

			if strings.Contains(output, "No such file") || strings.Contains(output, "cannot access") {
				t.Errorf("Expected SSH directory to exist in keys mode, got:\n%s", output)
			}

			t.Logf("Keys mode SSH directory:\n%s", output)
		})
	}
}

func TestSSH_Addt_AgentMode(t *testing.T) {
	providers := requireProviders(t)
	requireSSHAgent(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			// On macOS, podman can't forward Unix sockets into the VM
			if prov == "podman" && runtime.GOOS == "darwin" {
				t.Skip("podman on macOS cannot forward Unix sockets (use proxy mode)")
			}

			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
ssh:
  forward_keys: true
  forward_mode: "agent"
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "printenv SSH_AUTH_SOCK")
			if err != nil {
				t.Skipf("agent mode may not be supported on this platform: %v\nOutput: %s", err, output)
			}

			for _, line := range strings.Split(output, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "/") {
					t.Logf("Agent mode SSH_AUTH_SOCK: %s", line)
					return
				}
			}
			t.Errorf("SSH_AUTH_SOCK path not found in output:\n%s", output)
		})
	}
}

func TestSSH_Addt_Disabled(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
ssh:
  forward_keys: false
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "test -d /home/addt/.ssh && echo exists || echo missing")
			if err != nil {
				t.Fatalf("run test .ssh failed: %v\nOutput: %s", err, output)
			}

			if !strings.Contains(output, "missing") {
				t.Errorf("Expected SSH directory to be missing when forward_keys=false, got:\n%s", output)
			}
		})
	}
}

func TestSSH_Addt_CustomDir(t *testing.T) {
	_, cleanup := setupAddtDir(t, "", `
ssh:
  forward_keys: true
  dir: "/opt/custom/.ssh"
`)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	if !strings.Contains(output, "ssh.dir") {
		t.Errorf("Expected output to contain ssh.dir, got:\n%s", output)
	}
	if !strings.Contains(output, "/opt/custom/.ssh") {
		t.Errorf("Expected output to contain '/opt/custom/.ssh', got:\n%s", output)
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ssh.dir") {
			if !strings.Contains(line, "project") {
				t.Errorf("Expected ssh.dir source to be project, got line: %s", line)
			}
		}
	}
}

func TestSSH_Addt_GithubConnect(t *testing.T) {
	providers := requireProviders(t)
	requireSSHAgent(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
ssh:
  forward_keys: true
  forward_mode: "proxy"
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// ssh -T git@github.com returns exit code 1 with "successfully authenticated"
			output, _ := runRunSubcommand(t, dir, "debug",
				"-c", "ssh -T -o StrictHostKeyChecking=no -o ConnectTimeout=10 git@github.com")

			outputLower := strings.ToLower(output)
			if !strings.Contains(outputLower, "successfully authenticated") {
				t.Errorf("Expected 'successfully authenticated' in output, got:\n%s", output)
			}
		})
	}
}
