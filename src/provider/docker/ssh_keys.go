package docker

import (
	"fmt"
	"os"
)

// handleSSHKeysForwarding mounts the entire SSH directory read-only
func (p *DockerProvider) handleSSHKeysForwarding(sshDir, username string) []string {
	var args []string

	if _, err := os.Stat(sshDir); err == nil {
		args = append(args, "-v", fmt.Sprintf("%s:/home/%s/.ssh:ro", sshDir, username))
	}

	return args
}
