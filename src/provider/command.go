package provider

import (
	"os"
	"os/exec"
	"strings"
	"sync"
)

// DockerCmd creates an exec.Cmd for docker targeting a specific context.
// This ensures each provider (docker, orbstack) hits the correct daemon
// regardless of which Docker context is currently active.
func DockerCmd(context string, args ...string) *exec.Cmd {
	cmd := exec.Command("docker", args...)
	cmd.Env = append(os.Environ(), "DOCKER_CONTEXT="+context)
	return cmd
}

var (
	dockerContexts     []string
	dockerContextsOnce sync.Once
)

// loadDockerContexts parses `docker context ls` once and caches the result.
func loadDockerContexts() {
	dockerContextsOnce.Do(func() {
		cmd := exec.Command("docker", "context", "ls", "--format", "{{.Name}}")
		out, err := cmd.Output()
		if err != nil {
			return
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				dockerContexts = append(dockerContexts, line)
			}
		}
	})
}

// HasDockerContext checks if a named Docker context exists.
func HasDockerContext(name string) bool {
	loadDockerContexts()
	for _, ctx := range dockerContexts {
		if ctx == name {
			return true
		}
	}
	return false
}

// DockerContextNames returns all available Docker context names.
func DockerContextNames() []string {
	loadDockerContexts()
	return dockerContexts
}
