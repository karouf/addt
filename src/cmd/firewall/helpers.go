package firewall

import (
	"fmt"

	"github.com/jedi4ever/addt/config"
)

// containsString checks if a slice contains a string
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a string from a slice
func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// saveGlobalConfig saves global config
func saveGlobalConfig(cfg *config.GlobalConfig) {
	if err := config.SaveGlobalConfig(cfg); err != nil {
		fmt.Printf("Error saving global config: %v\n", err)
	}
}

// saveProjectConfig saves project config
func saveProjectConfig(cfg *config.GlobalConfig) {
	if err := config.SaveProjectConfig(cfg); err != nil {
		fmt.Printf("Error saving project config: %v\n", err)
	}
}

// printDomainList prints a list of domains with a label
func printDomainList(label string, domains []string, defaults []string, denied []string) {
	if len(domains) == 0 && len(defaults) == 0 {
		fmt.Printf("%s: (none)\n", label)
		return
	}

	fmt.Printf("%s:\n", label)
	for _, d := range domains {
		if containsString(denied, d) {
			fmt.Printf("    - %s (overridden by deny)\n", d)
		} else {
			fmt.Printf("    - %s\n", d)
		}
	}
	if len(defaults) > 0 {
		if len(domains) == 0 {
			fmt.Printf("    (defaults):\n")
		} else {
			fmt.Printf("    (+ defaults):\n")
		}
		for _, d := range defaults {
			if containsString(denied, d) {
				fmt.Printf("    - %s (overridden by deny)\n", d)
			} else if !containsString(domains, d) {
				fmt.Printf("    - %s\n", d)
			}
		}
	}
}

// ensureFirewall initializes the Firewall settings struct if nil
func ensureFirewall(cfg *config.GlobalConfig) *config.FirewallSettings {
	if cfg.Firewall == nil {
		cfg.Firewall = &config.FirewallSettings{}
	}
	return cfg.Firewall
}

// removeDomainFromConfig removes a domain from both allowed and denied lists
func removeDomainFromConfig(cfg *config.GlobalConfig, domain string) bool {
	fw := ensureFirewall(cfg)
	removed := false

	newAllowed := removeString(fw.Allowed, domain)
	if len(newAllowed) < len(fw.Allowed) {
		fw.Allowed = newAllowed
		removed = true
		fmt.Printf("Removed '%s' from allowed domains\n", domain)
	}

	newDenied := removeString(fw.Denied, domain)
	if len(newDenied) < len(fw.Denied) {
		fw.Denied = newDenied
		removed = true
		fmt.Printf("Removed '%s' from denied domains\n", domain)
	}

	return removed
}
