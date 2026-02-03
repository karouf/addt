package docker

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

// ImageExists checks if a Docker image exists
func (p *DockerProvider) ImageExists(imageName string) bool {
	cmd := exec.Command("docker", "image", "inspect", imageName)
	return cmd.Run() == nil
}

// FindImageByLabel finds an image by a specific label value
func (p *DockerProvider) FindImageByLabel(label, value string) string {
	cmd := exec.Command("docker", "images",
		"--filter", fmt.Sprintf("label=%s=%s", label, value),
		"--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" && !strings.Contains(line, "<none>") {
			return line
		}
	}
	return ""
}

// GetImageLabel retrieves a specific label value from an image
func (p *DockerProvider) GetImageLabel(imageName, label string) string {
	cmd := exec.Command("docker", "inspect",
		"--format", fmt.Sprintf("{{index .Config.Labels %q}}", label),
		imageName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetBaseImageName returns the base image name for the current config
func (p *DockerProvider) GetBaseImageName() string {
	// Get current user info for UID/GID in tag
	currentUser, err := user.Current()
	if err != nil {
		return "addt-base:latest"
	}
	// Base image is tagged with node version and UID to ensure compatibility
	return fmt.Sprintf("addt-base:node%s-uid%s", p.config.NodeVersion, currentUser.Uid)
}
