package cmd

import (
	"fmt"

	"github.com/jedi4ever/addt/config"
)

// configKeyInfo holds metadata about a config key
type configKeyInfo struct {
	Key         string
	Description string
	Type        string // "bool", "string", "int"
	EnvVar      string
}

// getConfigKeys returns all valid config keys with their metadata (sorted alphabetically)
func getConfigKeys() []configKeyInfo {
	keys := []configKeyInfo{
		{Key: "dind", Description: "Enable Docker-in-Docker", Type: "bool", EnvVar: "ADDT_DIND"},
		{Key: "dind_mode", Description: "Docker-in-Docker mode: host or isolated", Type: "string", EnvVar: "ADDT_DIND_MODE"},
		{Key: "docker_cpus", Description: "CPU limit for container (e.g., \"2\", \"0.5\")", Type: "string", EnvVar: "ADDT_DOCKER_CPUS"},
		{Key: "docker_memory", Description: "Memory limit for container (e.g., \"512m\", \"2g\")", Type: "string", EnvVar: "ADDT_DOCKER_MEMORY"},
		{Key: "firewall", Description: "Enable network firewall", Type: "bool", EnvVar: "ADDT_FIREWALL"},
		{Key: "firewall_mode", Description: "Firewall mode: strict, permissive, off", Type: "string", EnvVar: "ADDT_FIREWALL_MODE"},
		{Key: "github_detect", Description: "Auto-detect GitHub token from gh CLI", Type: "bool", EnvVar: "ADDT_GITHUB_DETECT"},
		{Key: "go_version", Description: "Go version", Type: "string", EnvVar: "ADDT_GO_VERSION"},
		{Key: "gpg_forward", Description: "Enable GPG forwarding", Type: "bool", EnvVar: "ADDT_GPG_FORWARD"},
		{Key: "log", Description: "Enable command logging", Type: "bool", EnvVar: "ADDT_LOG"},
		{Key: "log_file", Description: "Log file path", Type: "string", EnvVar: "ADDT_LOG_FILE"},
		{Key: "node_version", Description: "Node.js version", Type: "string", EnvVar: "ADDT_NODE_VERSION"},
		{Key: "persistent", Description: "Enable persistent container mode", Type: "bool", EnvVar: "ADDT_PERSISTENT"},
		{Key: "port_range_start", Description: "Starting port for auto allocation", Type: "int", EnvVar: "ADDT_PORT_RANGE_START"},
		{Key: "ssh_forward", Description: "SSH forwarding mode: agent or keys", Type: "string", EnvVar: "ADDT_SSH_FORWARD"},
		{Key: "uv_version", Description: "UV Python package manager version", Type: "string", EnvVar: "ADDT_UV_VERSION"},
		{Key: "workdir", Description: "Override working directory (default: current directory)", Type: "string", EnvVar: "ADDT_WORKDIR"},
		{Key: "workdir_automount", Description: "Auto-mount working directory to /workspace", Type: "bool", EnvVar: "ADDT_WORKDIR_AUTOMOUNT"},
	}
	return keys
}

// getExtensionConfigKeys returns all valid extension config keys with their metadata
func getExtensionConfigKeys() []configKeyInfo {
	return []configKeyInfo{
		{Key: "version", Description: "Extension version", Type: "string", EnvVar: "ADDT_%s_VERSION"},
		{Key: "automount", Description: "Auto-mount extension config directories", Type: "bool", EnvVar: "ADDT_%s_AUTOMOUNT"},
	}
}

// getDefaultValue returns the default value for a config key
func getDefaultValue(key string) string {
	switch key {
	case "docker_cpus":
		return ""
	case "dind":
		return "false"
	case "dind_mode":
		return "isolated"
	case "firewall":
		return "false"
	case "firewall_mode":
		return "strict"
	case "github_detect":
		return "false"
	case "go_version":
		return "latest"
	case "gpg_forward":
		return "false"
	case "log":
		return "false"
	case "log_file":
		return "addt.log"
	case "docker_memory":
		return ""
	case "node_version":
		return "22"
	case "persistent":
		return "false"
	case "port_range_start":
		return "30000"
	case "ssh_forward":
		return "agent"
	case "uv_version":
		return "latest"
	case "workdir":
		return "."
	case "workdir_automount":
		return "true"
	}
	return ""
}

// isValidConfigKey checks if a key is a valid config key
func isValidConfigKey(key string) bool {
	for _, k := range getConfigKeys() {
		if k.Key == key {
			return true
		}
	}
	return false
}

// getConfigKeyInfo returns the metadata for a config key, or nil if not found
func getConfigKeyInfo(key string) *configKeyInfo {
	for _, k := range getConfigKeys() {
		if k.Key == key {
			return &k
		}
	}
	return nil
}

// isValidExtensionConfigKey checks if a key is a valid extension config key
func isValidExtensionConfigKey(key string) bool {
	for _, k := range getExtensionConfigKeys() {
		if k.Key == key {
			return true
		}
	}
	return false
}

// getConfigValue retrieves a config value from the config struct
func getConfigValue(cfg *config.GlobalConfig, key string) string {
	switch key {
	case "docker_cpus":
		return cfg.DockerCPUs
	case "dind":
		if cfg.Dind != nil {
			return fmt.Sprintf("%v", *cfg.Dind)
		}
	case "dind_mode":
		return cfg.DindMode
	case "firewall":
		if cfg.Firewall != nil {
			return fmt.Sprintf("%v", *cfg.Firewall)
		}
	case "firewall_mode":
		return cfg.FirewallMode
	case "github_detect":
		if cfg.GitHubDetect != nil {
			return fmt.Sprintf("%v", *cfg.GitHubDetect)
		}
	case "go_version":
		return cfg.GoVersion
	case "gpg_forward":
		if cfg.GPGForward != nil {
			return fmt.Sprintf("%v", *cfg.GPGForward)
		}
	case "log":
		if cfg.Log != nil {
			return fmt.Sprintf("%v", *cfg.Log)
		}
	case "log_file":
		return cfg.LogFile
	case "docker_memory":
		return cfg.DockerMemory
	case "node_version":
		return cfg.NodeVersion
	case "persistent":
		if cfg.Persistent != nil {
			return fmt.Sprintf("%v", *cfg.Persistent)
		}
	case "port_range_start":
		if cfg.PortRangeStart != nil {
			return fmt.Sprintf("%d", *cfg.PortRangeStart)
		}
	case "ssh_forward":
		return cfg.SSHForward
	case "uv_version":
		return cfg.UvVersion
	case "workdir":
		return cfg.Workdir
	case "workdir_automount":
		if cfg.WorkdirAutomount != nil {
			return fmt.Sprintf("%v", *cfg.WorkdirAutomount)
		}
	}
	return ""
}

// setConfigValue sets a config value in the config struct
func setConfigValue(cfg *config.GlobalConfig, key, value string) {
	switch key {
	case "docker_cpus":
		cfg.DockerCPUs = value
	case "dind":
		b := value == "true"
		cfg.Dind = &b
	case "dind_mode":
		cfg.DindMode = value
	case "firewall":
		b := value == "true"
		cfg.Firewall = &b
	case "firewall_mode":
		cfg.FirewallMode = value
	case "github_detect":
		b := value == "true"
		cfg.GitHubDetect = &b
	case "go_version":
		cfg.GoVersion = value
	case "gpg_forward":
		b := value == "true"
		cfg.GPGForward = &b
	case "log":
		b := value == "true"
		cfg.Log = &b
	case "log_file":
		cfg.LogFile = value
	case "docker_memory":
		cfg.DockerMemory = value
	case "node_version":
		cfg.NodeVersion = value
	case "persistent":
		b := value == "true"
		cfg.Persistent = &b
	case "port_range_start":
		var i int
		fmt.Sscanf(value, "%d", &i)
		cfg.PortRangeStart = &i
	case "ssh_forward":
		cfg.SSHForward = value
	case "uv_version":
		cfg.UvVersion = value
	case "workdir":
		cfg.Workdir = value
	case "workdir_automount":
		b := value == "true"
		cfg.WorkdirAutomount = &b
	}
}

// unsetConfigValue clears a config value in the config struct
func unsetConfigValue(cfg *config.GlobalConfig, key string) {
	switch key {
	case "docker_cpus":
		cfg.DockerCPUs = ""
	case "dind":
		cfg.Dind = nil
	case "dind_mode":
		cfg.DindMode = ""
	case "firewall":
		cfg.Firewall = nil
	case "firewall_mode":
		cfg.FirewallMode = ""
	case "github_detect":
		cfg.GitHubDetect = nil
	case "go_version":
		cfg.GoVersion = ""
	case "gpg_forward":
		cfg.GPGForward = nil
	case "log":
		cfg.Log = nil
	case "log_file":
		cfg.LogFile = ""
	case "docker_memory":
		cfg.DockerMemory = ""
	case "node_version":
		cfg.NodeVersion = ""
	case "persistent":
		cfg.Persistent = nil
	case "port_range_start":
		cfg.PortRangeStart = nil
	case "ssh_forward":
		cfg.SSHForward = ""
	case "uv_version":
		cfg.UvVersion = ""
	case "workdir":
		cfg.Workdir = ""
	case "workdir_automount":
		cfg.WorkdirAutomount = nil
	}
}
