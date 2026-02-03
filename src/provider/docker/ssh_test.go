package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHandleSSHForwarding_Disabled(t *testing.T) {
	p := &DockerProvider{}

	testCases := []string{"", "off", "false", "none"}

	for _, mode := range testCases {
		t.Run(mode, func(t *testing.T) {
			args := p.HandleSSHForwarding(mode, "/home/test", "testuser")
			if len(args) != 0 {
				t.Errorf("HandleSSHForwarding(%q) returned %v, want empty", mode, args)
			}
		})
	}
}

func TestHandleSSHForwarding_Keys(t *testing.T) {
	p := &DockerProvider{}

	// Create a temporary home directory with .ssh
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh dir: %v", err)
	}

	// Create some test files
	os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("private"), 0600)
	os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("public"), 0644)
	os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host *"), 0644)

	args := p.HandleSSHForwarding("keys", homeDir, "testuser")

	// Should mount .ssh directory
	expectedMount := sshDir + ":/home/testuser/.ssh:ro"
	if !containsVolume(args, expectedMount) {
		t.Errorf("HandleSSHForwarding(\"keys\") missing mount %q, got %v", expectedMount, args)
	}

	// Should NOT set SSH_AUTH_SOCK
	if containsEnvPrefix(args, "SSH_AUTH_SOCK=") {
		t.Errorf("HandleSSHForwarding(\"keys\") should not set SSH_AUTH_SOCK")
	}
}

func TestHandleSSHForwarding_Keys_NoSSHDir(t *testing.T) {
	p := &DockerProvider{}

	// Create a temporary home directory WITHOUT .ssh
	homeDir := t.TempDir()

	args := p.HandleSSHForwarding("keys", homeDir, "testuser")

	// Should return empty when .ssh doesn't exist
	if len(args) != 0 {
		t.Errorf("HandleSSHForwarding(\"keys\") without .ssh returned %v, want empty", args)
	}
}

func TestHandleSSHForwarding_Agent_NoSocket(t *testing.T) {
	p := &DockerProvider{}

	// Save and clear SSH_AUTH_SOCK
	origSock := os.Getenv("SSH_AUTH_SOCK")
	os.Unsetenv("SSH_AUTH_SOCK")
	defer func() {
		if origSock != "" {
			os.Setenv("SSH_AUTH_SOCK", origSock)
		}
	}()

	args := p.HandleSSHForwarding("agent", "/home/test", "testuser")

	// Should return empty when no SSH agent
	if len(args) != 0 {
		t.Errorf("HandleSSHForwarding(\"agent\") without SSH_AUTH_SOCK returned %v, want empty", args)
	}
}

func TestHandleSSHForwarding_Agent_MacOSSocket(t *testing.T) {
	p := &DockerProvider{}

	// Save original SSH_AUTH_SOCK
	origSock := os.Getenv("SSH_AUTH_SOCK")
	defer func() {
		if origSock != "" {
			os.Setenv("SSH_AUTH_SOCK", origSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	// Set a macOS-style socket path (these don't work in Docker)
	os.Setenv("SSH_AUTH_SOCK", "/var/folders/xx/fake/com.apple.launchd.xxx/Listeners")

	args := p.HandleSSHForwarding("agent", "/home/test", "testuser")

	// Should return empty for macOS sockets (with warning printed)
	if len(args) != 0 {
		t.Errorf("HandleSSHForwarding(\"agent\") with macOS socket returned %v, want empty", args)
	}
}

func TestHandleSSHForwarding_Agent_ValidSocket(t *testing.T) {
	p := &DockerProvider{}

	// Create a socket path that won't trigger macOS detection
	// macOS detection looks for "com.apple.launchd" or "/var/folders/"
	// Use /tmp which is a symlink to /private/tmp on macOS but won't match the pattern
	tmpDir, err := os.MkdirTemp("/tmp", "ssh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "agent.sock")

	// Create the socket file (just a regular file for testing)
	if err := os.WriteFile(socketPath, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create mock socket: %v", err)
	}

	// Create a home directory with .ssh
	homeDir, err := os.MkdirTemp("/tmp", "home-test-*")
	if err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	defer os.RemoveAll(homeDir)

	sshDir := filepath.Join(homeDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host *"), 0644)
	os.WriteFile(filepath.Join(sshDir, "known_hosts"), []byte("github.com ..."), 0644)
	os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("ssh-rsa ..."), 0644)

	// Save and set SSH_AUTH_SOCK
	origSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", socketPath)
	defer func() {
		if origSock != "" {
			os.Setenv("SSH_AUTH_SOCK", origSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	args := p.HandleSSHForwarding("agent", homeDir, "testuser")

	// Should mount the socket
	expectedSocketMount := socketPath + ":/ssh-agent"
	if !containsVolume(args, expectedSocketMount) {
		t.Errorf("HandleSSHForwarding(\"agent\") missing socket mount %q, got %v", expectedSocketMount, args)
	}

	// Should set SSH_AUTH_SOCK env
	if !containsEnv(args, "SSH_AUTH_SOCK=/ssh-agent") {
		t.Errorf("HandleSSHForwarding(\"agent\") missing SSH_AUTH_SOCK env, got %v", args)
	}

	// Should mount safe SSH files (in a temp dir)
	hasSshMount := false
	for i, arg := range args {
		if arg == "-v" && i+1 < len(args) {
			mount := args[i+1]
			if contains(mount, ".ssh:ro") {
				hasSshMount = true
				break
			}
		}
	}
	if !hasSshMount {
		t.Errorf("HandleSSHForwarding(\"agent\") missing .ssh mount, got %v", args)
	}
}

func TestHandleSSHForwarding_True_SameAsAgent(t *testing.T) {
	p := &DockerProvider{}

	// Save and clear SSH_AUTH_SOCK
	origSock := os.Getenv("SSH_AUTH_SOCK")
	os.Unsetenv("SSH_AUTH_SOCK")
	defer func() {
		if origSock != "" {
			os.Setenv("SSH_AUTH_SOCK", origSock)
		}
	}()

	// Both "true" and "agent" should behave the same
	argsTrue := p.HandleSSHForwarding("true", "/home/test", "testuser")
	argsAgent := p.HandleSSHForwarding("agent", "/home/test", "testuser")

	if len(argsTrue) != len(argsAgent) {
		t.Errorf("HandleSSHForwarding(\"true\") = %v, HandleSSHForwarding(\"agent\") = %v, want same", argsTrue, argsAgent)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
