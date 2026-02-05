package docker

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// BuildIfNeeded ensures the Docker image is ready
func (p *DockerProvider) BuildIfNeeded(rebuild bool, rebuildBase bool) error {
	// Handle --addt-rebuild-base flag - rebuild base image first
	if rebuildBase {
		baseImageName := p.GetBaseImageName()
		fmt.Printf("Rebuilding base image %s...\n", baseImageName)
		if p.ImageExists(baseImageName) {
			cmd := exec.Command("docker", "rmi", baseImageName)
			cmd.Run()
		}
		if err := p.BuildBaseImage(); err != nil {
			return err
		}
	}

	imageExists := p.ImageExists(p.config.ImageName)

	// Handle --addt-rebuild flag
	if rebuild {
		if imageExists {
			fmt.Printf("Rebuilding %s...\n", p.config.ImageName)
			fmt.Println("Removing existing image...")
			cmd := exec.Command("docker", "rmi", p.config.ImageName)
			cmd.Run()
		}
		return p.BuildImage(p.embeddedDockerfile, p.embeddedEntrypoint)
	}

	// If image doesn't exist, build it
	if !imageExists {
		return p.BuildImage(p.embeddedDockerfile, p.embeddedEntrypoint)
	}

	// Image exists with matching tag - versions are encoded in tag, no rebuild needed
	return nil
}

// DetermineImageName determines the appropriate Docker image name based on installed extensions
func (p *DockerProvider) DetermineImageName() string {
	// Parse extensions list (comma-separated)
	extensions := strings.Split(p.config.Extensions, ",")
	for i := range extensions {
		extensions[i] = strings.TrimSpace(extensions[i])
	}

	// Filter empty entries and sort alphabetically for consistent naming
	var validExts []string
	for _, ext := range extensions {
		if ext != "" {
			validExts = append(validExts, ext)
		}
	}
	sort.Strings(validExts)

	// First, check if we already have a matching image (avoids npm lookups)
	// This prevents rebuilds when npm version changes or network is flaky
	if existingImage := p.findExistingImage(validExts); existingImage != "" {
		return existingImage
	}

	// No existing image found, resolve versions and build image name
	var tagParts []string
	for _, ext := range validExts {
		version := p.resolveExtensionVersion(ext)
		tagParts = append(tagParts, fmt.Sprintf("%s-%s", ext, version))
	}

	// Join with underscore
	tag := strings.Join(tagParts, "_")
	if tag == "" {
		tag = "base"
	}

	// Prefix with addt version so images are rebuilt when addt is updated
	imageName := fmt.Sprintf("addt:v%s_%s", p.config.AddtVersion, tag)
	return imageName
}

// findExistingImage looks for an existing image matching the extensions
// This avoids npm lookups when we already have a usable image
func (p *DockerProvider) findExistingImage(extensions []string) string {
	if len(extensions) == 0 {
		// Check for base image
		baseImage := fmt.Sprintf("addt:v%s_base", p.config.AddtVersion)
		if p.ImageExists(baseImage) {
			return baseImage
		}
		return ""
	}

	// Build a pattern to find images: addt:v{version}_{ext1}-*_{ext2}-*
	prefix := fmt.Sprintf("addt:v%s_", p.config.AddtVersion)

	// List all addt images
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}", "addt")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse output and find matching images
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" || !strings.HasPrefix(line, prefix) {
			continue
		}

		// Check if this image has all required extensions
		tag := strings.TrimPrefix(line, prefix)
		if p.imageTagMatchesExtensions(tag, extensions) {
			return line
		}
	}

	return ""
}

// imageTagMatchesExtensions checks if an image tag contains all required extensions
func (p *DockerProvider) imageTagMatchesExtensions(tag string, extensions []string) bool {
	// Tag format: ext1-version1_ext2-version2
	parts := strings.Split(tag, "_")

	// Build a set of extensions in the tag
	tagExts := make(map[string]bool)
	for _, part := range parts {
		// Extract extension name (everything before the last dash)
		if idx := strings.LastIndex(part, "-"); idx > 0 {
			extName := part[:idx]
			tagExts[extName] = true
		}
	}

	// Check all required extensions are present
	for _, ext := range extensions {
		if !tagExts[ext] {
			return false
		}
	}

	// Also ensure we don't have extra extensions
	return len(tagExts) == len(extensions)
}

// resolveExtensionVersion resolves the version for an extension, handling dist-tags
func (p *DockerProvider) resolveExtensionVersion(extName string) string {
	version := p.getExtensionVersion(extName)

	// For claude extension, handle npm dist-tags (latest, stable, next)
	if extName == "claude" && (version == "latest" || version == "stable" || version == "next") {
		npmVersion := p.getNpmVersionByTag(version)
		if npmVersion != "" {
			p.setExtensionVersion(extName, npmVersion)
			return npmVersion
		}
	}

	// For claude with specific version, validate it exists
	if extName == "claude" && version != "latest" && version != "stable" && version != "next" {
		if !p.validateNpmVersion(version) {
			fmt.Printf("Error: Claude Code version %s does not exist in npm\n", version)
			fmt.Println("Available versions: https://www.npmjs.com/package/@anthropic-ai/claude-code?activeTab=versions")
			os.Exit(1)
		}
	}

	return version
}

// getExtensionVersion returns the version for an extension, defaulting to "stable" for claude
func (p *DockerProvider) getExtensionVersion(extName string) string {
	if p.config.ExtensionVersions == nil {
		if extName == "claude" {
			return "stable"
		}
		return "latest"
	}
	if ver, ok := p.config.ExtensionVersions[extName]; ok {
		return ver
	}
	if extName == "claude" {
		return "stable"
	}
	return "latest"
}

// setExtensionVersion sets the version for an extension
func (p *DockerProvider) setExtensionVersion(extName, version string) {
	if p.config.ExtensionVersions == nil {
		p.config.ExtensionVersions = make(map[string]string)
	}
	p.config.ExtensionVersions[extName] = version
}
