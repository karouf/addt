package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jedi4ever/addt/provider"
	"github.com/jedi4ever/addt/util"
	"github.com/jedi4ever/addt/util/terminal"
)

var optionsLogger = util.Log("options")

// BuildRunOptions creates a RunSpec from the configuration
func BuildRunOptions(p provider.Provider, cfg *provider.Config, name string, args []string, openShell bool) *provider.RunSpec {
	optionsLogger.Debugf("BuildRunOptions called: name=%s, openShell=%v, args=%v", name, openShell, args)

	// Use configured workdir or fall back to current directory
	cwd := cfg.Workdir
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	optionsLogger.Debugf("Working directory: %s", cwd)

	// Interactive mode: if running in a terminal, allow interactive input
	// Both shell mode and run mode can be interactive when in a terminal
	isTerminal := terminal.IsTerminal()
	isInteractive := isTerminal // Both shell and run can be interactive in a terminal
	optionsLogger.Debugf("Terminal check: isTerminal=%v, openShell=%v, isInteractive=%v", isTerminal, openShell, isInteractive)

	// Build the run spec
	spec := &provider.RunSpec{
		Name:             name,
		ImageName:        cfg.ImageName,
		Args:             args,
		WorkDir:          cwd,
		Interactive:      isInteractive,
		Persistent:       cfg.Persistent,
		Volumes:          BuildVolumes(cfg, cwd),
		Ports:            BuildPorts(cfg),
		Env:              BuildEnvironment(p, cfg),
		SSHForwardKeys:   cfg.SSHForwardKeys,
		SSHForwardMode:   cfg.SSHForwardMode,
		SSHAllowedKeys:   cfg.SSHAllowedKeys,
		TmuxForward:      cfg.TmuxForward,
		HistoryPersist:   cfg.HistoryPersist,
		GPGForward:       cfg.GPGForward,
		GPGAllowedKeyIDs: cfg.GPGAllowedKeyIDs,
		DockerDindMode:   cfg.DockerDindMode,
		ContainerCPUs:    cfg.ContainerCPUs,
		ContainerMemory:  cfg.ContainerMemory,
	}
	// Resolve flag → env var mappings (e.g., --yolo → ADDT_EXTENSION_CLAUDE_YOLO=true)
	addFlagEnvVars(spec.Env, cfg, args)

	optionsLogger.Debugf("RunSpec created: Name=%s, ImageName=%s, Interactive=%v, Persistent=%v, DockerDindMode=%s",
		spec.Name, spec.ImageName, spec.Interactive, spec.Persistent, spec.DockerDindMode)

	// Handle args based on mode
	if openShell && len(args) > 0 {
		spec.Args = args
		optionsLogger.Debugf("Shell mode with args: %v", args)
	} else if openShell {
		spec.Args = []string{}
		optionsLogger.Debug("Shell mode without args")
	} else {
		spec.Args = args
		optionsLogger.Debugf("Run mode with args: %v", args)
	}

	// Log command if enabled
	if cfg.LogEnabled {
		optionsLogger.Debug("Command logging enabled, logging command")
		LogCommand(cfg.LogFile, cwd, name, args)
	} else {
		optionsLogger.Debug("Command logging disabled")
	}

	// Load env file vars directly into spec.Env if enabled
	if cfg.EnvFileLoad {
		loadEnvFileVars(spec, cfg, cwd)
	}

	optionsLogger.Debugf("BuildRunOptions completed: spec.Args=%v, spec.Env count=%d", spec.Args, len(spec.Env))
	return spec
}

// loadEnvFileVars reads the env file and adds its variables directly to spec.Env.
// This ensures env file vars work regardless of IsolateSecrets mode.
func loadEnvFileVars(spec *provider.RunSpec, cfg *provider.Config, cwd string) {
	envFilePath := cfg.EnvFile
	if envFilePath == "" {
		envFilePath = ".env"
	}
	if !filepath.IsAbs(envFilePath) {
		envFilePath = filepath.Join(cwd, envFilePath)
	}
	info, err := os.Stat(envFilePath)
	if err != nil || info.IsDir() {
		return
	}

	vars, err := parseEnvFile(envFilePath)
	if err != nil {
		optionsLogger.Debugf("Failed to parse env file %s: %v", envFilePath, err)
		return
	}

	for k, v := range vars {
		spec.Env[k] = v
	}
	spec.Env["ADDT_ENV_FILE"] = envFilePath
	optionsLogger.Debugf("Loaded %d vars from env file: %s", len(vars), envFilePath)
}

// parseEnvFile reads a .env file and returns key=value pairs.
// Supports comments (#), empty lines, and simple KEY=VALUE format.
func parseEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Strip surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if key != "" {
			vars[key] = value
		}
	}
	return vars, nil
}
