package core

import (
	"fmt"

	"github.com/jedi4ever/addt/provider"
)

// Runner coordinates container execution
type Runner struct {
	provider provider.Provider
	config   *provider.Config
}

// NewRunner creates a new runner
func NewRunner(p provider.Provider, cfg *provider.Config) *Runner {
	return &Runner{
		provider: p,
		config:   cfg,
	}
}

// Run executes the container with the configured extension
func (r *Runner) Run(args []string) error {
	return r.execute(args, false)
}

// Shell opens an interactive shell in the container
func (r *Runner) Shell(args []string) error {
	return r.execute(args, true)
}

// execute is the common execution logic for Run and Shell
func (r *Runner) execute(args []string, openShell bool) error {
	// Determine container name
	name := r.generateName()

	// Build run options
	opts := BuildRunOptions(r.provider, r.config, name, args, openShell)

	// Display status
	DisplayStatus(r.provider, r.config, name)

	// Execute via provider
	if openShell {
		return r.provider.Shell(opts)
	}
	return r.provider.Run(opts)
}

// generateName generates the container name based on persistence mode
func (r *Runner) generateName() string {
	if r.config.Persistent {
		return r.provider.GeneratePersistentName()
	}
	return r.provider.GenerateEphemeralName()
}

// GetExtensionName returns the current extension name
func (r *Runner) GetExtensionName() string {
	if r.config.Command != "" {
		return r.config.Command
	}
	return "claude"
}

// DisplayWarning shows the experimental warning
func (r *Runner) DisplayWarning() {
	fmt.Printf("⚠ addt:%s is experimental - things are not perfect yet\n", r.GetExtensionName())
}
