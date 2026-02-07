package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jedi4ever/addt/util"
)

// ApplySettings applies Settings overrides to a Config
func ApplySettings(cfg *Config, settings *Settings) {
	if settings == nil {
		return
	}
	if settings.PidsLimit != nil {
		cfg.PidsLimit = *settings.PidsLimit
	}
	if settings.UlimitNofile != "" {
		cfg.UlimitNofile = settings.UlimitNofile
	}
	if settings.UlimitNproc != "" {
		cfg.UlimitNproc = settings.UlimitNproc
	}
	if settings.NoNewPrivileges != nil {
		cfg.NoNewPrivileges = *settings.NoNewPrivileges
	}
	if len(settings.CapDrop) > 0 {
		cfg.CapDrop = settings.CapDrop
	}
	if len(settings.CapAdd) > 0 {
		cfg.CapAdd = settings.CapAdd
	}
	if settings.ReadOnlyRootfs != nil {
		cfg.ReadOnlyRootfs = *settings.ReadOnlyRootfs
	}
	if settings.TmpfsTmpSize != "" {
		cfg.TmpfsTmpSize = settings.TmpfsTmpSize
	}
	if settings.TmpfsHomeSize != "" {
		cfg.TmpfsHomeSize = settings.TmpfsHomeSize
	}
	if settings.SeccompProfile != "" {
		cfg.SeccompProfile = settings.SeccompProfile
	}
	if settings.NetworkMode != "" {
		cfg.NetworkMode = settings.NetworkMode
	}
	if settings.DisableIPC != nil {
		cfg.DisableIPC = *settings.DisableIPC
	}
	if settings.TimeLimit != nil {
		cfg.TimeLimit = *settings.TimeLimit
	}
	if settings.UserNamespace != "" {
		cfg.UserNamespace = settings.UserNamespace
	}
	if settings.DisableDevices != nil {
		cfg.DisableDevices = *settings.DisableDevices
	}
	if settings.MemorySwap != "" {
		cfg.MemorySwap = settings.MemorySwap
	}
	if settings.IsolateSecrets != nil {
		cfg.IsolateSecrets = *settings.IsolateSecrets
	}
	if settings.AuditLog != nil {
		cfg.AuditLog = *settings.AuditLog
	}
	if settings.AuditLogFile != "" {
		cfg.AuditLogFile = settings.AuditLogFile
	}
}

// ApplyEnvOverrides applies environment variable overrides to a Config
func ApplyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ADDT_SECURITY_PIDS_LIMIT"); v != "" {
		if pids, err := strconv.Atoi(v); err == nil {
			cfg.PidsLimit = pids
		}
	}
	if v := os.Getenv("ADDT_SECURITY_ULIMIT_NOFILE"); v != "" {
		cfg.UlimitNofile = v
	}
	if v := os.Getenv("ADDT_SECURITY_ULIMIT_NPROC"); v != "" {
		cfg.UlimitNproc = v
	}
	if v := os.Getenv("ADDT_SECURITY_NO_NEW_PRIVILEGES"); v != "" {
		cfg.NoNewPrivileges = v != "false"
	}
	if v := os.Getenv("ADDT_SECURITY_CAP_DROP"); v != "" {
		cfg.CapDrop = strings.Split(v, ",")
	}
	if v := os.Getenv("ADDT_SECURITY_CAP_ADD"); v != "" {
		cfg.CapAdd = strings.Split(v, ",")
	}
	if v := os.Getenv("ADDT_SECURITY_READ_ONLY_ROOTFS"); v != "" {
		cfg.ReadOnlyRootfs = v == "true"
	}
	if v := os.Getenv("ADDT_SECURITY_TMPFS_TMP_SIZE"); v != "" {
		cfg.TmpfsTmpSize = v
	}
	if v := os.Getenv("ADDT_SECURITY_TMPFS_HOME_SIZE"); v != "" {
		cfg.TmpfsHomeSize = v
	}
	if v := os.Getenv("ADDT_SECURITY_SECCOMP_PROFILE"); v != "" {
		cfg.SeccompProfile = v
	}
	if v := os.Getenv("ADDT_SECURITY_NETWORK_MODE"); v != "" {
		cfg.NetworkMode = v
	}
	if v := os.Getenv("ADDT_SECURITY_DISABLE_IPC"); v != "" {
		cfg.DisableIPC = v == "true"
	}
	if v := os.Getenv("ADDT_SECURITY_TIME_LIMIT"); v != "" {
		if minutes, err := strconv.Atoi(v); err == nil && minutes > 0 {
			cfg.TimeLimit = minutes
		}
	}
	if v := os.Getenv("ADDT_SECURITY_USER_NAMESPACE"); v != "" {
		cfg.UserNamespace = v
	}
	if v := os.Getenv("ADDT_SECURITY_DISABLE_DEVICES"); v != "" {
		cfg.DisableDevices = v == "true"
	}
	if v := os.Getenv("ADDT_SECURITY_MEMORY_SWAP"); v != "" {
		cfg.MemorySwap = v
	}
	if v := os.Getenv("ADDT_SECURITY_ISOLATE_SECRETS"); v != "" {
		cfg.IsolateSecrets = v == "true"
	}
	if v := os.Getenv("ADDT_SECURITY_AUDIT_LOG"); v != "" {
		cfg.AuditLog = v == "true"
	}
	if v := os.Getenv("ADDT_SECURITY_AUDIT_LOG_FILE"); v != "" {
		cfg.AuditLogFile = v
	}
}

// LoadConfig loads security configuration with full precedence chain
func LoadConfig(globalSettings, projectSettings *Settings) Config {
	cfg := DefaultConfig()
	ApplySettings(&cfg, globalSettings)
	ApplySettings(&cfg, projectSettings)
	ApplyEnvOverrides(&cfg)
	return cfg
}

// InitAuditLog initializes audit logging if enabled in config
func InitAuditLog(cfg *Config) error {
	if !cfg.AuditLog {
		return nil
	}

	logPath := cfg.AuditLogFile
	if logPath == "" {
		// Default to <addt_home>/audit.log
		addtHome := util.GetAddtHome()
		if addtHome == "" {
			return fmt.Errorf("failed to determine addt home directory")
		}
		logPath = filepath.Join(addtHome, "audit.log")
	}

	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return EnableAuditLog(logPath)
}
