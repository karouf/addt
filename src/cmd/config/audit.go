package config

import (
	"fmt"
	"os"
	"strings"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// ResolvedKey holds a config key with its resolved value and source.
type ResolvedKey struct {
	Key    string
	Value  string
	Source string // "env", "project", "global", "default", or ""
}

// GroupPosture represents the security assessment of an audit group.
type GroupPosture struct {
	Secure bool
	Tags   []string
}

// AuditGroup defines a security audit category with its keys and evaluator.
type AuditGroup struct {
	Name     string
	Keys     []string
	Evaluate func(resolved map[string]ResolvedKey) GroupPosture
}

// AuditGroupResult holds the evaluation result for one audit group.
type AuditGroupResult struct {
	Group   AuditGroup
	Posture GroupPosture
	Keys    []ResolvedKey
}

// AuditResult holds the complete audit output.
type AuditResult struct {
	Groups       []AuditGroupResult
	TotalGroups  int
	RelaxedCount int
}

// GetAuditGroups returns the 6 security audit groups.
func GetAuditGroups() []AuditGroup {
	return []AuditGroup{
		{
			Name: "Network",
			Keys: []string{
				"firewall.enabled",
				"firewall.mode",
				"security.network_mode",
				"docker.dind.enable",
			},
			Evaluate: evaluateNetwork,
		},
		{
			Name: "Filesystem",
			Keys: []string{
				"workdir.automount",
				"workdir.readonly",
				"security.read_only_rootfs",
				"config.automount",
				"config.readonly",
			},
			Evaluate: evaluateFilesystem,
		},
		{
			Name: "Credentials",
			Keys: []string{
				"ssh.forward_keys",
				"ssh.forward_mode",
				"github.forward_token",
				"github.scope_token",
				"security.isolate_secrets",
			},
			Evaluate: evaluateCredentials,
		},
		{
			Name: "Limits",
			Keys: []string{
				"container.cpus",
				"container.memory",
				"security.pids_limit",
				"security.time_limit",
			},
			Evaluate: evaluateLimits,
		},
		{
			Name: "Isolation",
			Keys: []string{
				"security.no_new_privileges",
				"security.cap_drop",
				"security.cap_add",
				"security.yolo",
				"git.disable_hooks",
				"security.seccomp_profile",
				"security.user_namespace",
				"security.disable_devices",
				"security.disable_ipc",
			},
			Evaluate: evaluateIsolation,
		},
		{
			Name: "Audit",
			Keys: []string{
				"security.audit_log",
				"security.audit_log_file",
			},
			Evaluate: evaluateAuditGroup,
		},
	}
}

func val(resolved map[string]ResolvedKey, key string) string {
	if r, ok := resolved[key]; ok {
		return r.Value
	}
	return "-"
}

func evaluateNetwork(resolved map[string]ResolvedKey) GroupPosture {
	fw := val(resolved, "firewall.enabled")
	fwMode := val(resolved, "firewall.mode")
	netMode := val(resolved, "security.network_mode")
	dind := val(resolved, "docker.dind.enable")

	var tags []string

	fwOn := strings.EqualFold(fw, "true")
	if fwOn {
		tags = append(tags, "firewall:on")
	} else {
		tags = append(tags, "firewall:off")
	}

	if netMode == "none" {
		tags = append(tags, "network:none")
	} else if netMode == "" || netMode == "-" {
		tags = append(tags, "network:bridge")
	} else {
		tags = append(tags, "network:"+netMode)
	}

	if strings.EqualFold(dind, "true") {
		tags = append(tags, "dind:on")
	}

	secure := fwOn && strings.EqualFold(fwMode, "strict") && netMode == "none" && !strings.EqualFold(dind, "true")
	return GroupPosture{Secure: secure, Tags: tags}
}

func evaluateFilesystem(resolved map[string]ResolvedKey) GroupPosture {
	workdirRo := val(resolved, "workdir.readonly")
	rootfsRo := val(resolved, "security.read_only_rootfs")

	var tags []string
	if strings.EqualFold(workdirRo, "true") {
		tags = append(tags, "workdir:ro")
	} else {
		tags = append(tags, "workdir:rw")
	}
	if strings.EqualFold(rootfsRo, "true") {
		tags = append(tags, "rootfs:ro")
	} else {
		tags = append(tags, "rootfs:rw")
	}

	secure := strings.EqualFold(workdirRo, "true") && strings.EqualFold(rootfsRo, "true")
	return GroupPosture{Secure: secure, Tags: tags}
}

func evaluateCredentials(resolved map[string]ResolvedKey) GroupPosture {
	sshFwd := val(resolved, "ssh.forward_keys")
	sshMode := val(resolved, "ssh.forward_mode")
	ghFwd := val(resolved, "github.forward_token")
	ghScope := val(resolved, "github.scope_token")
	isolate := val(resolved, "security.isolate_secrets")

	var tags []string

	// SSH tag
	if !strings.EqualFold(sshFwd, "true") {
		tags = append(tags, "ssh:off")
	} else if strings.EqualFold(sshMode, "proxy") {
		tags = append(tags, "ssh:proxy")
	} else {
		tags = append(tags, "ssh:"+sshMode)
	}

	// GitHub tag
	if !strings.EqualFold(ghFwd, "true") {
		tags = append(tags, "github:off")
	} else if strings.EqualFold(ghScope, "true") {
		tags = append(tags, "github:scoped")
	} else {
		tags = append(tags, "github:unscoped")
	}

	// Secrets tag
	if strings.EqualFold(isolate, "true") {
		tags = append(tags, "secrets:isolated")
	} else {
		tags = append(tags, "secrets:shared")
	}

	sshOk := !strings.EqualFold(sshFwd, "true") || strings.EqualFold(sshMode, "proxy")
	ghOk := !strings.EqualFold(ghFwd, "true") || strings.EqualFold(ghScope, "true")
	secretsOk := strings.EqualFold(isolate, "true")
	secure := sshOk && ghOk && secretsOk

	return GroupPosture{Secure: secure, Tags: tags}
}

func evaluateLimits(resolved map[string]ResolvedKey) GroupPosture {
	cpus := val(resolved, "container.cpus")
	mem := val(resolved, "container.memory")
	pids := val(resolved, "security.pids_limit")
	timeLimit := val(resolved, "security.time_limit")

	var tags []string

	if cpus == "" || cpus == "-" {
		tags = append(tags, "cpu:-")
	} else {
		tags = append(tags, "cpu:"+cpus)
	}

	if mem == "" || mem == "-" {
		tags = append(tags, "mem:-")
	} else {
		tags = append(tags, "mem:"+mem)
	}

	tags = append(tags, "pids:"+pids)

	if timeLimit == "0" || timeLimit == "" || timeLimit == "-" {
		tags = append(tags, "time:unlimited")
	} else {
		tags = append(tags, "time:"+timeLimit+"s")
	}

	secure := timeLimit != "0" && timeLimit != "" && timeLimit != "-"
	return GroupPosture{Secure: secure, Tags: tags}
}

func evaluateIsolation(resolved map[string]ResolvedKey) GroupPosture {
	noNewPrivs := val(resolved, "security.no_new_privileges")
	capDrop := val(resolved, "security.cap_drop")
	yolo := val(resolved, "security.yolo")
	hooks := val(resolved, "git.disable_hooks")

	var tags []string

	if strings.EqualFold(noNewPrivs, "true") {
		tags = append(tags, "no-new-privs")
	} else {
		tags = append(tags, "new-privs-allowed")
	}

	if strings.EqualFold(capDrop, "ALL") {
		tags = append(tags, "caps:restricted")
	} else {
		tags = append(tags, "caps:permissive")
	}

	if strings.EqualFold(hooks, "true") {
		tags = append(tags, "hooks:disabled")
	} else {
		tags = append(tags, "hooks:enabled")
	}

	if strings.EqualFold(yolo, "true") {
		tags = append(tags, "yolo:on")
	}

	secure := strings.EqualFold(noNewPrivs, "true") &&
		strings.EqualFold(capDrop, "ALL") &&
		!strings.EqualFold(yolo, "true") &&
		strings.EqualFold(hooks, "true")

	return GroupPosture{Secure: secure, Tags: tags}
}

func evaluateAuditGroup(resolved map[string]ResolvedKey) GroupPosture {
	auditLog := val(resolved, "security.audit_log")

	var tags []string
	if strings.EqualFold(auditLog, "true") {
		tags = append(tags, "logging:on")
	} else {
		tags = append(tags, "logging:off")
	}

	secure := strings.EqualFold(auditLog, "true")
	return GroupPosture{Secure: secure, Tags: tags}
}

// RunAudit resolves all security-relevant keys and evaluates each group.
func RunAudit(projectCfg, globalCfg *cfgtypes.GlobalConfig) AuditResult {
	groups := GetAuditGroups()
	allKeys := GetKeys()

	// Build KeyInfo lookup
	keyInfoMap := make(map[string]KeyInfo, len(allKeys))
	for _, k := range allKeys {
		keyInfoMap[k.Key] = k
	}

	var results []AuditGroupResult
	relaxedCount := 0

	for _, g := range groups {
		resolved := make(map[string]ResolvedKey, len(g.Keys))
		var orderedKeys []ResolvedKey

		for _, keyName := range g.Keys {
			ki, exists := keyInfoMap[keyName]
			var value, source string
			if exists {
				value, source = resolveValueAndSource(ki, projectCfg, globalCfg)
			} else {
				value, source = "-", ""
			}
			rk := ResolvedKey{Key: keyName, Value: value, Source: source}
			resolved[keyName] = rk
			orderedKeys = append(orderedKeys, rk)
		}

		posture := g.Evaluate(resolved)
		if !posture.Secure {
			relaxedCount++
		}

		results = append(results, AuditGroupResult{
			Group:   g,
			Posture: posture,
			Keys:    orderedKeys,
		})
	}

	return AuditResult{
		Groups:       results,
		TotalGroups:  len(groups),
		RelaxedCount: relaxedCount,
	}
}

// auditCommand loads configs and runs the security audit.
func auditCommand() {
	projectCfg, err := cfgtypes.LoadProjectConfigFile()
	if err != nil {
		fmt.Printf("Error loading project config: %v\n", err)
		os.Exit(1)
	}

	globalCfg, err := cfgtypes.LoadGlobalConfigFile()
	if err != nil {
		fmt.Printf("Error loading global config: %v\n", err)
		os.Exit(1)
	}

	result := RunAudit(projectCfg, globalCfg)
	printAudit(result)
}
