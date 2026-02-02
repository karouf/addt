package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration options
type Config struct {
	NodeVersion          string
	GoVersion            string
	UvVersion            string
	EnvVars              []string
	GitHubDetect         bool
	Ports                []string
	PortRangeStart       int
	SSHForward           string
	GPGForward           bool
	DindMode             string
	EnvFile              string
	LogEnabled           bool
	LogFile              string
	ImageName            string
	Persistent           bool              // Enable persistent container mode
	MountWorkdir         bool              // Mount present working directory
	FirewallEnabled      bool              // Enable network firewall
	FirewallMode         string            // Firewall mode: strict, permissive, off
	Mode                 string            // container or shell
	Provider             string            // Provider type: docker or daytona
	Extensions           string            // Comma-separated list of extensions to install (e.g., "gastown,beads")
	Command              string            // Command to run instead of claude (e.g., "gt" for gastown)
	ExtensionVersions    map[string]string // Per-extension versions (e.g., {"claude": "1.0.5", "codex": "latest"})
	MountExtensionConfig map[string]bool   // Per-extension mount control (e.g., {"claude": true, "codex": false})
}

// LoadConfig loads configuration from environment variables
func LoadConfig(defaultNodeVersion string, defaultGoVersion string, defaultUvVersion string, defaultPortRangeStart int) *Config {
	cfg := &Config{
		NodeVersion:          getEnvOrDefault("DCLAUDE_NODE_VERSION", defaultNodeVersion),
		GoVersion:            getEnvOrDefault("DCLAUDE_GO_VERSION", defaultGoVersion),
		UvVersion:            getEnvOrDefault("DCLAUDE_UV_VERSION", defaultUvVersion),
		EnvVars:              strings.Split(getEnvOrDefault("DCLAUDE_ENV_VARS", "ANTHROPIC_API_KEY,GH_TOKEN"), ","),
		GitHubDetect:         getEnvOrDefault("DCLAUDE_GITHUB_DETECT", "false") == "true",
		PortRangeStart:       getEnvInt("DCLAUDE_PORT_RANGE_START", defaultPortRangeStart),
		SSHForward:           os.Getenv("DCLAUDE_SSH_FORWARD"),
		GPGForward:           os.Getenv("DCLAUDE_GPG_FORWARD") == "true",
		DindMode:             os.Getenv("DCLAUDE_DIND_MODE"),
		EnvFile:              os.Getenv("DCLAUDE_ENV_FILE"), // Empty means use default .env
		LogEnabled:           os.Getenv("DCLAUDE_LOG") == "true",
		LogFile:              getEnvOrDefault("DCLAUDE_LOG_FILE", "dclaude.log"),
		Persistent:           os.Getenv("DCLAUDE_PERSISTENT") == "true",
		MountWorkdir:         getEnvOrDefault("DCLAUDE_MOUNT_WORKDIR", "true") != "false",
		FirewallEnabled:      os.Getenv("DCLAUDE_FIREWALL") == "true",
		FirewallMode:         getEnvOrDefault("DCLAUDE_FIREWALL_MODE", "strict"),
		Mode:                 getEnvOrDefault("DCLAUDE_MODE", "container"),
		Provider:             getEnvOrDefault("DCLAUDE_PROVIDER", "docker"),
		Extensions:           getEnvOrDefault("DCLAUDE_EXTENSIONS", "claude"),
		Command:              os.Getenv("DCLAUDE_COMMAND"), // Empty means use default "claude"
		ExtensionVersions:    make(map[string]string),
		MountExtensionConfig: make(map[string]bool),
	}

	// Load per-extension versions and mount configs from environment
	// Pattern: DCLAUDE_<EXT>_VERSION and DCLAUDE_MOUNT_<EXT>_CONFIG
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		// Check for DCLAUDE_<EXT>_VERSION pattern
		if strings.HasPrefix(key, "DCLAUDE_") && strings.HasSuffix(key, "_VERSION") {
			// Extract extension name (e.g., "DCLAUDE_CLAUDE_VERSION" -> "claude")
			extName := strings.TrimPrefix(key, "DCLAUDE_")
			extName = strings.TrimSuffix(extName, "_VERSION")
			extName = strings.ToLower(extName)
			// Skip non-extension versions (node, go, uv)
			if extName != "node" && extName != "go" && extName != "uv" {
				cfg.ExtensionVersions[extName] = value
			}
		}

		// Check for DCLAUDE_<EXT>_MOUNT_CONFIG pattern
		if strings.HasPrefix(key, "DCLAUDE_") && strings.HasSuffix(key, "_MOUNT_CONFIG") {
			// Extract extension name (e.g., "DCLAUDE_CLAUDE_MOUNT_CONFIG" -> "claude")
			extName := strings.TrimPrefix(key, "DCLAUDE_")
			extName = strings.TrimSuffix(extName, "_MOUNT_CONFIG")
			extName = strings.ToLower(extName)
			cfg.MountExtensionConfig[extName] = value != "false"
		}
	}

	// Set default version for claude if not specified
	if _, exists := cfg.ExtensionVersions["claude"]; !exists {
		cfg.ExtensionVersions["claude"] = "stable"
	}

	// Parse ports
	if ports := os.Getenv("DCLAUDE_PORTS"); ports != "" {
		cfg.Ports = strings.Split(ports, ",")
		for i := range cfg.Ports {
			cfg.Ports[i] = strings.TrimSpace(cfg.Ports[i])
		}
	}

	// Trim env vars
	for i := range cfg.EnvVars {
		cfg.EnvVars[i] = strings.TrimSpace(cfg.EnvVars[i])
	}

	return cfg
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
