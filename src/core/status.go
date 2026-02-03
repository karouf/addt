package core

import (
	"fmt"

	"github.com/jedi4ever/addt/provider"
)

// DisplayStatus displays the status line for the current session
func DisplayStatus(p provider.Provider, cfg *provider.Config, envName string) {
	status := p.GetStatus(cfg, envName)

	// Add port mappings to status
	portDisplay := BuildPortDisplayString(cfg)
	if portDisplay != "" {
		status += fmt.Sprintf(" | Ports:%s", portDisplay)
	}

	// Get extension name
	extension := cfg.Command
	if extension == "" {
		extension = "claude"
	}

	fmt.Printf("⚠ addt:%s is experimental - things are not perfect yet\n", extension)
	fmt.Printf("✓ %s\n", status)
}

// FormatStatus formats a status string (without printing)
func FormatStatus(p provider.Provider, cfg *provider.Config, envName string) string {
	status := p.GetStatus(cfg, envName)

	// Add port mappings to status
	portDisplay := BuildPortDisplayString(cfg)
	if portDisplay != "" {
		status += fmt.Sprintf(" | Ports:%s", portDisplay)
	}

	return status
}
