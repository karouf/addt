package docker

import (
	"embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/jedi4ever/addt/provider"
)

// DockerProvider implements the Provider interface for Docker
type DockerProvider struct {
	config                 *provider.Config
	tempDirs               []string
	embeddedDockerfile     []byte
	embeddedDockerfileBase []byte
	embeddedEntrypoint     []byte
	embeddedInitFirewall   []byte
	embeddedInstallSh      []byte
	embeddedExtensions     embed.FS
}

// NewDockerProvider creates a new Docker provider
func NewDockerProvider(cfg *provider.Config, dockerfile, dockerfileBase, entrypoint, initFirewall, installSh []byte, extensions embed.FS) (provider.Provider, error) {
	return &DockerProvider{
		config:                 cfg,
		tempDirs:               []string{},
		embeddedDockerfile:     dockerfile,
		embeddedDockerfileBase: dockerfileBase,
		embeddedEntrypoint:     entrypoint,
		embeddedInitFirewall:   initFirewall,
		embeddedInstallSh:      installSh,
		embeddedExtensions:     extensions,
	}, nil
}

// Initialize initializes the Docker provider
func (p *DockerProvider) Initialize(cfg *provider.Config) error {
	p.config = cfg
	return p.CheckPrerequisites()
}

// GetName returns the provider name
func (p *DockerProvider) GetName() string {
	return "docker"
}

// CheckPrerequisites verifies Docker is installed and running
func (p *DockerProvider) CheckPrerequisites() error {
	// Check Docker is installed
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("Docker is not installed. Please install Docker from: https://docs.docker.com/get-docker/")
	}

	// Check Docker daemon is running
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker daemon is not running. Please start Docker and try again")
	}

	return nil
}

// Container lifecycle methods (Exists, IsRunning, Start, Stop, Remove, List)
// and name generation (GenerateContainerName, GenerateEphemeralName, GeneratePersistentName)
// are defined in persistent.go

// Cleanup removes temporary directories
func (p *DockerProvider) Cleanup() error {
	for _, dir := range p.tempDirs {
		os.RemoveAll(dir)
	}
	p.tempDirs = []string{}
	return nil
}
