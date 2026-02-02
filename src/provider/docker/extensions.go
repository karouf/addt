package docker

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtensionMount represents a mount point for an extension
type ExtensionMount struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// ExtensionMetadata represents metadata for an installed extension
type ExtensionMetadata struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Entrypoint  string           `json:"entrypoint"`
	Mounts      []ExtensionMount `json:"mounts"`
}

// ExtensionsConfig represents the extensions.json file structure
type ExtensionsConfig struct {
	Extensions map[string]ExtensionMetadata `json:"extensions"`
}

// GetExtensionMounts reads extension metadata from image and returns all mounts
func (p *DockerProvider) GetExtensionMounts(imageName string) []ExtensionMount {
	var mounts []ExtensionMount

	// Read extensions.json from the image
	cmd := exec.Command("docker", "run", "--rm", "--entrypoint", "cat", imageName,
		"/home/claude/.dclaude/extensions.json")
	output, err := cmd.Output()
	if err != nil {
		// No extensions metadata or file doesn't exist
		return mounts
	}

	// Parse the JSON
	var config ExtensionsConfig
	if err := json.Unmarshal(output, &config); err != nil {
		return mounts
	}

	// Collect all mounts from all extensions
	for _, ext := range config.Extensions {
		mounts = append(mounts, ext.Mounts...)
	}

	return mounts
}

// AddExtensionMounts adds extension mount volumes to docker args
func (p *DockerProvider) AddExtensionMounts(dockerArgs []string, imageName, homeDir string) []string {
	extMounts := p.GetExtensionMounts(imageName)
	for _, extMount := range extMounts {
		// Skip claude config mounts if MountClaudeConfig is disabled
		if !p.config.MountClaudeConfig {
			if strings.Contains(extMount.Target, "/.claude") {
				continue
			}
		}

		// Expand ~ to home directory
		source := extMount.Source
		if strings.HasPrefix(source, "~/") {
			source = filepath.Join(homeDir, source[2:])
		}

		// Check if source exists, create if it's a directory path
		if info, err := os.Stat(source); err == nil {
			// Source exists (file or directory)
			dockerArgs = append(dockerArgs, "-v", source+":"+extMount.Target)
		} else if os.IsNotExist(err) {
			// Source doesn't exist - create directory if path doesn't look like a file
			if !strings.Contains(filepath.Base(source), ".") {
				// Looks like a directory (no extension)
				if err := os.MkdirAll(source, 0755); err == nil {
					dockerArgs = append(dockerArgs, "-v", source+":"+extMount.Target)
				}
			}
			// Skip files that don't exist (e.g., ~/.claude.json on fresh install)
		} else if info != nil {
			dockerArgs = append(dockerArgs, "-v", source+":"+extMount.Target)
		}
	}
	return dockerArgs
}
