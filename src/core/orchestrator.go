package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedi4ever/dclaude/internal/ports"
	"github.com/jedi4ever/dclaude/internal/terminal"
	"github.com/jedi4ever/dclaude/internal/util"
	"github.com/jedi4ever/dclaude/provider"
)

// Orchestrator coordinates provider operations with business logic
type Orchestrator struct {
	provider provider.Provider
	config   *provider.Config
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(p provider.Provider, cfg *provider.Config) *Orchestrator {
	return &Orchestrator{
		provider: p,
		config:   cfg,
	}
}

// RunClaude orchestrates the Claude Code execution
func (o *Orchestrator) RunClaude(args []string, openShell bool) error {
	// Determine environment name
	var envName string
	if o.config.Persistent {
		envName = o.generatePersistentName()
	} else {
		envName = o.generateEphemeralName()
	}

	// Build RunSpec from config
	spec := o.buildRunSpec(envName, args, openShell)

	// Display status
	o.displayStatus(envName)

	// Execute via provider
	if openShell {
		return o.provider.Shell(spec)
	}
	return o.provider.Run(spec)
}

// buildRunSpec builds a RunSpec from the current configuration
func (o *Orchestrator) buildRunSpec(name string, args []string, openShell bool) *provider.RunSpec {
	cwd, _ := os.Getwd()

	spec := &provider.RunSpec{
		Name:          name,
		ImageName:     o.config.ImageName,
		Args:          args,
		WorkDir:       cwd,
		Interactive:   terminal.IsTerminal(),
		Persistent:    o.config.Persistent,
		Volumes:       o.buildVolumes(cwd),
		Ports:         o.buildPorts(),
		Env:           o.buildEnvironment(),
		SSHForward:    o.config.SSHForward,
		GPGForward:    o.config.GPGForward,
		DockerForward: o.config.DockerForward,
	}

	// Special handling for shell mode with args
	if openShell && len(args) > 0 {
		spec.Args = args
	} else if openShell {
		spec.Args = []string{}
	} else {
		// Normal mode - add "claude" command
		spec.Args = append([]string{"claude"}, args...)
	}

	// Log command if enabled
	if o.config.LogEnabled {
		util.LogCommand(o.config.LogFile, cwd, name, args)
	}

	// Store env file path if exists
	envFilePath := o.config.EnvFile
	if envFilePath == "" {
		envFilePath = ".env"
	}
	if !filepath.IsAbs(envFilePath) {
		envFilePath = filepath.Join(cwd, envFilePath)
	}
	if info, err := os.Stat(envFilePath); err == nil && !info.IsDir() {
		spec.Env["DCLAUDE_ENV_FILE"] = envFilePath
	}

	return spec
}

// buildVolumes builds volume mounts
func (o *Orchestrator) buildVolumes(cwd string) []provider.VolumeMount {
	volumes := []provider.VolumeMount{
		{
			Source:   cwd,
			Target:   "/workspace",
			ReadOnly: false,
		},
	}
	return volumes
}

// buildPorts builds port mappings
func (o *Orchestrator) buildPorts() []provider.PortMapping {
	var portsList []provider.PortMapping
	hostPort := o.config.PortRangeStart

	for _, containerPort := range o.config.Ports {
		containerPort = strings.TrimSpace(containerPort)
		hostPort = ports.FindAvailablePort(hostPort)

		// Parse container port as int
		var containerPortInt int
		fmt.Sscanf(containerPort, "%d", &containerPortInt)

		portsList = append(portsList, provider.PortMapping{
			Container: containerPortInt,
			Host:      hostPort,
		})
		hostPort++
	}

	return portsList
}

// buildEnvironment builds environment variables map
func (o *Orchestrator) buildEnvironment() map[string]string {
	env := make(map[string]string)

	// Pass configured environment variables
	for _, varName := range o.config.EnvVars {
		if value := os.Getenv(varName); value != "" {
			env[varName] = value
		}
	}

	// Pass terminal environment variables for proper paste handling
	if term := os.Getenv("TERM"); term != "" {
		env["TERM"] = term
	}
	if colorterm := os.Getenv("COLORTERM"); colorterm != "" {
		env["COLORTERM"] = colorterm
	}

	// Pass terminal size variables (critical for proper line handling in containers)
	cols, lines := terminal.GetTerminalSize()
	env["COLUMNS"] = fmt.Sprintf("%d", cols)
	env["LINES"] = fmt.Sprintf("%d", lines)

	// Build port map string for display
	if len(o.config.Ports) > 0 {
		var portMappings []string
		hostPort := o.config.PortRangeStart
		for _, containerPort := range o.config.Ports {
			containerPort = strings.TrimSpace(containerPort)
			hostPort = ports.FindAvailablePort(hostPort)
			portMappings = append(portMappings, fmt.Sprintf("%s:%d", containerPort, hostPort))
			hostPort++
		}
		portMapString := strings.Join(portMappings, ",")
		env["DCLAUDE_PORT_MAP"] = portMapString
	}

	return env
}

// displayStatus displays the status line
func (o *Orchestrator) displayStatus(envName string) {
	status := o.provider.GetStatus(o.config, envName)

	// Add port mappings to status
	if len(o.config.Ports) > 0 {
		hostPort := o.config.PortRangeStart
		var portMappings []string
		for _, containerPort := range o.config.Ports {
			containerPort = strings.TrimSpace(containerPort)
			hostPort = ports.FindAvailablePort(hostPort)
			portMappings = append(portMappings, fmt.Sprintf("%s→%d", containerPort, hostPort))
			hostPort++
		}
		portMapDisplay := strings.Join(portMappings, ",")
		status += fmt.Sprintf(" | Ports:%s", portMapDisplay)
	}

	fmt.Printf("✓ %s\n", status)
}

// generatePersistentName generates a persistent container name
func (o *Orchestrator) generatePersistentName() string {
	return o.provider.GeneratePersistentName()
}

// generateEphemeralName generates an ephemeral container name
func (o *Orchestrator) generateEphemeralName() string {
	return o.provider.GenerateEphemeralName()
}

// BuildIfNeeded ensures the image/environment is ready
func (o *Orchestrator) BuildIfNeeded(embeddedDockerfile, embeddedEntrypoint []byte, rebuild bool) error {
	// Only Docker provider needs to build images
	dockerProvider, ok := o.provider.(*docker.DockerProvider)
	if !ok {
		// Other providers don't need to build
		return nil
	}

	// Handle --rebuild flag
	if rebuild {
		fmt.Printf("Rebuilding %s...\n", o.config.ImageName)
		if dockerProvider.ImageExists(o.config.ImageName) {
			fmt.Println("Removing existing image...")
			cmd := exec.Command("docker", "rmi", o.config.ImageName)
			cmd.Run()
		}
	}

	// Build image if needed
	if !dockerProvider.ImageExists(o.config.ImageName) {
		return dockerProvider.BuildImage(embeddedDockerfile, embeddedEntrypoint)
	}

	return nil
}

// DetermineImageName determines the appropriate image name based on config
func (o *Orchestrator) DetermineImageName() string {
	// Only Docker provider uses images
	dockerProvider, ok := o.provider.(*docker.DockerProvider)
	if !ok {
		return ""
	}

	if o.config.ClaudeVersion == "latest" {
		// Query npm registry for latest version
		npmLatest := GetNpmLatestVersion()
		if npmLatest != "" {
			// Check if we already have an image with this version
			existingImage := dockerProvider.FindImageByLabel("tools.claude.version", npmLatest)
			if existingImage != "" {
				return existingImage
			}
			o.config.ClaudeVersion = npmLatest
			return fmt.Sprintf("dclaude:claude-%s", npmLatest)
		}
		return "dclaude:latest"
	}

	// Specific version requested - validate it exists
	if !ValidateNpmVersion(o.config.ClaudeVersion) {
		fmt.Printf("Error: Claude Code version %s does not exist in npm\n", o.config.ClaudeVersion)
		fmt.Println("Available versions: https://www.npmjs.com/package/@anthropic-ai/claude-code?activeTab=versions")
		os.Exit(1)
	}

	// Check if image exists
	existingImage := dockerProvider.FindImageByLabel("tools.claude.version", o.config.ClaudeVersion)
	if existingImage != "" {
		return existingImage
	}
	return fmt.Sprintf("dclaude:claude-%s", o.config.ClaudeVersion)
}
