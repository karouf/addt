package core

import (
	"fmt"
	"strings"

	"github.com/jedi4ever/addt/provider"
)

// SecurityPostureLine builds a compact security posture summary from the config.
// Returns the posture string (without prefix icon) and whether all settings are locked down.
func SecurityPostureLine(cfg *provider.Config) (string, bool) {
	var parts []string
	allLocked := true

	// Firewall
	if cfg.FirewallEnabled {
		parts = append(parts, fmt.Sprintf("firewall:%s", cfg.FirewallMode))
		// "strict" is fully locked down; anything else is permissive
		if cfg.FirewallMode != "strict" {
			allLocked = false
		}
	} else {
		parts = append(parts, "firewall:off")
		allLocked = false
	}

	// Network mode
	networkMode := cfg.Security.NetworkMode
	if networkMode == "" {
		networkMode = "bridge"
	}
	parts = append(parts, fmt.Sprintf("network:%s", networkMode))
	if networkMode != "none" {
		allLocked = false
	}

	// Workdir
	if !cfg.WorkdirAutomount {
		parts = append(parts, "workdir:none")
	} else if cfg.WorkdirReadonly {
		parts = append(parts, "workdir:ro")
	} else {
		parts = append(parts, "workdir:rw")
		allLocked = false
	}

	// Root filesystem
	if cfg.Security.ReadOnlyRootfs {
		parts = append(parts, "rootfs:ro")
	} else {
		parts = append(parts, "rootfs:rw")
		allLocked = false
	}

	// Audit logging
	if cfg.Security.AuditLog {
		parts = append(parts, "audit:on")
	} else {
		parts = append(parts, "audit:off")
		allLocked = false
	}

	// Time limit (only show when > 0)
	if cfg.Security.TimeLimit > 0 {
		parts = append(parts, fmt.Sprintf("time:%dm", cfg.Security.TimeLimit))
	}

	// Pids limit (only show when non-default, default is 200)
	if cfg.Security.PidsLimit != 200 && cfg.Security.PidsLimit != 0 {
		parts = append(parts, fmt.Sprintf("pids:%d", cfg.Security.PidsLimit))
	}

	// Secrets isolation
	if !cfg.Security.IsolateSecrets {
		parts = append(parts, "secrets:exposed")
		allLocked = false
	}

	return strings.Join(parts, " "), allLocked
}

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

	// Line 1: experimental warning
	fmt.Printf("⚠ addt:%s is experimental - things are not perfect yet\n", extension)

	// Line 2: provider + features status
	fmt.Printf("✓ %s\n", status)

	// Line 3: security posture
	posture, allLocked := SecurityPostureLine(cfg)
	if allLocked {
		fmt.Printf("✓ %s\n", posture)
	} else {
		fmt.Printf("⚠ %s\n", posture)
	}

	// Line 4: yolo warning (conditional, at end)
	if cfg.Security.Yolo {
		fmt.Printf("⚠ security.yolo is enabled - extensions will run with --yolo flag\n")
	}
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
