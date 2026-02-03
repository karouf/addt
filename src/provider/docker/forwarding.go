package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/jedi4ever/addt/internal/util"
)

// HandleSSHForwarding configures SSH forwarding based on config
func (p *DockerProvider) HandleSSHForwarding(sshForward, homeDir, username string) []string {
	var args []string

	if sshForward == "agent" || sshForward == "true" {
		sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
		if sshAuthSock != "" {
			// Check if socket exists and is accessible
			if _, err := os.Stat(sshAuthSock); err == nil {
				// Check for macOS launchd sockets (won't work)
				if strings.Contains(sshAuthSock, "com.apple.launchd") || strings.Contains(sshAuthSock, "/var/folders/") {
					fmt.Println("Warning: SSH agent forwarding not supported on macOS (use ADDT_SSH_FORWARD=keys)")
				} else {
					args = append(args, "-v", fmt.Sprintf("%s:/ssh-agent", sshAuthSock))
					args = append(args, "-e", "SSH_AUTH_SOCK=/ssh-agent")

					// Mount safe SSH files only
					sshDir := filepath.Join(homeDir, ".ssh")
					if _, err := os.Stat(sshDir); err == nil {
						tmpDir, err := os.MkdirTemp("", "ssh-safe-*")
						if err == nil {
							p.tempDirs = append(p.tempDirs, tmpDir)

							// Copy safe files
							util.SafeCopyFile(filepath.Join(sshDir, "config"), filepath.Join(tmpDir, "config"))
							util.SafeCopyFile(filepath.Join(sshDir, "known_hosts"), filepath.Join(tmpDir, "known_hosts"))

							// Copy public keys
							files, _ := filepath.Glob(filepath.Join(sshDir, "*.pub"))
							for _, f := range files {
								util.SafeCopyFile(f, filepath.Join(tmpDir, filepath.Base(f)))
							}

							args = append(args, "-v", fmt.Sprintf("%s:/home/%s/.ssh:ro", tmpDir, username))
						}
					}
				}
			}
		}
	} else if sshForward == "keys" {
		sshDir := filepath.Join(homeDir, ".ssh")
		if _, err := os.Stat(sshDir); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:/home/%s/.ssh:ro", sshDir, username))
		}
	}

	return args
}

// HandleDockerForwarding configures Docker-in-Docker or host Docker socket forwarding
func (p *DockerProvider) HandleDockerForwarding(dindMode, containerName string) []string {
	var args []string

	if dindMode == "host" {
		socketPath := "/var/run/docker.sock"
		if _, err := os.Stat(socketPath); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:%s", socketPath, socketPath))

			// Get socket group ID using stat command for cross-platform compatibility
			gid := getDockerSocketGID(socketPath)
			if gid > 0 {
				args = append(args, "--group-add", fmt.Sprintf("%d", gid))
				if gid != 102 {
					args = append(args, "--group-add", "102")
				}
				if gid != 999 {
					args = append(args, "--group-add", "999")
				}
			} else {
				fmt.Println("Warning: Could not detect Docker socket group, using common defaults")
				args = append(args, "--group-add", "102", "--group-add", "999")
			}
		} else {
			fmt.Println("Warning: ADDT_DIND_MODE=host but /var/run/docker.sock not found")
		}
	} else if dindMode == "isolated" || dindMode == "true" {
		args = append(args, "--privileged")
		args = append(args, "-v", fmt.Sprintf("addt-docker-%s:/var/lib/docker", containerName))
		args = append(args, "-e", "ADDT_DIND=true")
	}

	return args
}

// getDockerSocketGID returns the group ID of the Docker socket
func getDockerSocketGID(socketPath string) int {
	// Try using syscall.Stat_t first (works on Linux)
	if info, err := os.Stat(socketPath); err == nil {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			return int(stat.Gid)
		}
	}

	// Fallback: use stat command (works on macOS and Linux)
	// Try GNU stat format first (Linux)
	cmd := exec.Command("stat", "-c", "%g", socketPath)
	if output, err := cmd.Output(); err == nil {
		if gid, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			return gid
		}
	}

	// Try BSD stat format (macOS)
	cmd = exec.Command("stat", "-f", "%g", socketPath)
	if output, err := cmd.Output(); err == nil {
		if gid, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			return gid
		}
	}

	return 0
}
