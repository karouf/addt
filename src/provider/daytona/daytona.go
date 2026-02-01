package daytona

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jedi4ever/dclaude/provider"
)

// DaytonaProvider implements the Provider interface for Daytona
type DaytonaProvider struct {
	config *provider.Config
}

// NewDaytonaProvider creates a new Daytona provider
func NewDaytonaProvider(cfg *provider.Config) (provider.Provider, error) {
	return &DaytonaProvider{
		config: cfg,
	}, nil
}

// Initialize initializes the Daytona provider
func (p *DaytonaProvider) Initialize(cfg *provider.Config) error {
	p.config = cfg
	return p.CheckPrerequisites()
}

// GetName returns the provider name
func (p *DaytonaProvider) GetName() string {
	return "daytona"
}

// CheckPrerequisites verifies Daytona is installed and running
func (p *DaytonaProvider) CheckPrerequisites() error {
	// Check Daytona is installed
	if _, err := exec.LookPath("daytona"); err != nil {
		return fmt.Errorf("Daytona is not installed. Please install Daytona from: https://github.com/daytonaio/daytona")
	}

	// Check Daytona server is running
	cmd := exec.Command("daytona", "server", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Daytona server is not running. Please start Daytona with: daytona server start")
	}

	return nil
}

// Exists checks if a workspace exists
func (p *DaytonaProvider) Exists(name string) bool {
	cmd := exec.Command("daytona", "list", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	// Simple check - just see if the name appears in the output
	// In production, parse JSON properly
	return strings.Contains(string(output), name)
}

// IsRunning checks if a workspace is currently running
func (p *DaytonaProvider) IsRunning(name string) bool {
	// Daytona workspaces are always "running" once created
	return p.Exists(name)
}

// Start starts a stopped workspace (no-op for Daytona)
func (p *DaytonaProvider) Start(name string) error {
	// Daytona workspaces don't need explicit start
	return nil
}

// Stop stops a running workspace
func (p *DaytonaProvider) Stop(name string) error {
	cmd := exec.Command("daytona", "stop", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Remove removes a workspace
func (p *DaytonaProvider) Remove(name string) error {
	cmd := exec.Command("daytona", "delete", name, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// List lists all dclaude workspaces
func (p *DaytonaProvider) List() ([]provider.Environment, error) {
	cmd := exec.Command("daytona", "list", "--format", "table")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var envs []provider.Environment
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			// Skip header
			continue
		}
		// Parse table output - this is a simple implementation
		// In production, use JSON format and proper parsing
		parts := strings.Fields(line)
		if len(parts) > 0 && strings.HasPrefix(parts[0], "dclaude-") {
			envs = append(envs, provider.Environment{
				Name:      parts[0],
				Status:    "running",
				CreatedAt: "",
			})
		}
	}
	return envs, nil
}

// Run runs a command in a workspace
func (p *DaytonaProvider) Run(spec *provider.RunSpec) error {
	workspaceName := spec.Name

	// Check if workspace exists
	if !p.Exists(workspaceName) {
		// Create workspace
		fmt.Printf("Creating Daytona workspace: %s\n", workspaceName)
		createArgs := []string{"create", workspaceName}

		// Add image if specified
		if spec.ImageName != "" {
			createArgs = append(createArgs, "--image", spec.ImageName)
		}

		cmd := exec.Command("daytona", createArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create Daytona workspace: %w", err)
		}
	} else {
		fmt.Printf("Using existing Daytona workspace: %s\n", workspaceName)
	}

	// Connect to workspace and run command
	connectArgs := []string{"connect", workspaceName}
	if len(spec.Args) > 0 {
		connectArgs = append(connectArgs, "--")
		connectArgs = append(connectArgs, spec.Args...)
	}

	cmd := exec.Command("daytona", connectArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Shell opens a shell in a workspace
func (p *DaytonaProvider) Shell(spec *provider.RunSpec) error {
	workspaceName := spec.Name

	// Check if workspace exists
	if !p.Exists(workspaceName) {
		// Create workspace
		fmt.Printf("Creating Daytona workspace: %s\n", workspaceName)
		createArgs := []string{"create", workspaceName}

		// Add image if specified
		if spec.ImageName != "" {
			createArgs = append(createArgs, "--image", spec.ImageName)
		}

		cmd := exec.Command("daytona", createArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create Daytona workspace: %w", err)
		}
	} else {
		fmt.Printf("Using existing Daytona workspace: %s\n", workspaceName)
	}

	// Connect to workspace with shell
	fmt.Println("Opening shell in Daytona workspace...")
	cmd := exec.Command("daytona", "connect", workspaceName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Cleanup cleans up resources
func (p *DaytonaProvider) Cleanup() error {
	// Nothing to clean up for Daytona
	return nil
}

// GetStatus returns a status string for display
func (p *DaytonaProvider) GetStatus(cfg *provider.Config, envName string) string {
	status := fmt.Sprintf("Provider:%s Mode:workspace", p.GetName())

	// Workspace name
	if cfg.Persistent {
		status += fmt.Sprintf(" | Workspace:%s", envName)
	}

	// GitHub token status
	if os.Getenv("GH_TOKEN") != "" {
		status += " | GH:✓"
	} else {
		status += " | GH:-"
	}

	// SSH is built-in to Daytona
	status += " | SSH:builtin"

	// Docker support (if workspace has it)
	if cfg.DockerForward != "" {
		status += " | Docker:workspace"
	}

	return status
}

// GeneratePersistentName generates a workspace name for persistent mode
func (p *DaytonaProvider) GeneratePersistentName() string {
	return "dclaude-workspace"
}

// GenerateEphemeralName generates a unique workspace name for ephemeral mode
func (p *DaytonaProvider) GenerateEphemeralName() string {
	return fmt.Sprintf("dclaude-%d", os.Getpid())
}

// BuildIfNeeded is a no-op for Daytona (no image building needed)
func (p *DaytonaProvider) BuildIfNeeded(rebuild bool) error {
	return nil
}

// DetermineImageName returns empty string for Daytona (no image concept)
func (p *DaytonaProvider) DetermineImageName() string {
	return ""
}
