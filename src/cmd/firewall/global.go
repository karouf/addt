package firewall

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/config"
)

func handleGlobal(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: addt firewall global <command>")
		fmt.Println("Commands: allow, deny, remove, list, reset")
		return
	}

	cmd := args[0]
	cfg := config.LoadGlobalConfig()

	switch cmd {
	case "allow":
		globalAllow(cfg, args)
	case "deny":
		globalDeny(cfg, args)
	case "remove", "rm":
		globalRemove(cfg, args)
	case "list", "ls":
		globalList(cfg)
	case "reset":
		globalReset(cfg)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Commands: allow, deny, remove, list, reset")
	}
}

func globalAllow(cfg *config.GlobalConfig, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: addt firewall global allow <domain>")
		os.Exit(1)
	}
	fw := ensureFirewall(cfg)
	domain := strings.TrimSpace(args[1])
	if containsString(fw.Allowed, domain) {
		fmt.Printf("Domain '%s' already in global allowed list\n", domain)
		return
	}
	fw.Allowed = append(fw.Allowed, domain)
	saveGlobalConfig(cfg)
	fmt.Printf("Added '%s' to global allowed domains\n", domain)
}

func globalDeny(cfg *config.GlobalConfig, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: addt firewall global deny <domain>")
		os.Exit(1)
	}
	fw := ensureFirewall(cfg)
	domain := strings.TrimSpace(args[1])
	if containsString(fw.Denied, domain) {
		fmt.Printf("Domain '%s' already in global denied list\n", domain)
		return
	}
	fw.Denied = append(fw.Denied, domain)
	saveGlobalConfig(cfg)
	fmt.Printf("Added '%s' to global denied domains\n", domain)
}

func globalRemove(cfg *config.GlobalConfig, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: addt firewall global remove <domain>")
		os.Exit(1)
	}
	domain := strings.TrimSpace(args[1])
	removed := removeDomainFromConfig(cfg, domain)
	if removed {
		saveGlobalConfig(cfg)
	} else {
		fmt.Printf("Domain '%s' not found in global config\n", domain)
	}
}

func globalList(cfg *config.GlobalConfig) {
	fw := ensureFirewall(cfg)
	fmt.Println("Global firewall rules:")
	printDomainList("  Allowed", fw.Allowed, DefaultAllowedDomains(), fw.Denied)
	fmt.Printf("  Denied:\n")
	if len(fw.Denied) == 0 {
		fmt.Printf("    (none)\n")
	} else {
		for _, d := range fw.Denied {
			fmt.Printf("    - %s\n", d)
		}
	}
}

func globalReset(cfg *config.GlobalConfig) {
	fw := ensureFirewall(cfg)
	fw.Allowed = DefaultAllowedDomains()
	fw.Denied = nil
	saveGlobalConfig(cfg)
	fmt.Println("Reset global firewall rules to defaults")
}
