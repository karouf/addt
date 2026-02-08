package config

import (
	"fmt"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// GetWorkdirKeys returns all valid workdir config keys
func GetWorkdirKeys() []KeyInfo {
	return []KeyInfo{
		{Key: "workdir.path", Description: "Override working directory (default: current directory)", Type: "string", EnvVar: "ADDT_WORKDIR"},
		{Key: "workdir.automount", Description: "Auto-mount working directory to /workspace", Type: "bool", EnvVar: "ADDT_WORKDIR_AUTOMOUNT"},
		{Key: "workdir.readonly", Description: "Mount working directory as read-only", Type: "bool", EnvVar: "ADDT_WORKDIR_READONLY"},
		{Key: "workdir.autotrust", Description: "Trust /workspace directory on first launch (default: true)", Type: "bool", EnvVar: "ADDT_WORKDIR_AUTOTRUST"},
	}
}

// GetWorkdirValue retrieves a workdir config value
func GetWorkdirValue(w *cfgtypes.WorkdirSettings, key string) string {
	if w == nil {
		return ""
	}
	switch key {
	case "workdir.path":
		return w.Path
	case "workdir.automount":
		if w.Automount != nil {
			return fmt.Sprintf("%v", *w.Automount)
		}
	case "workdir.readonly":
		if w.Readonly != nil {
			return fmt.Sprintf("%v", *w.Readonly)
		}
	case "workdir.autotrust":
		if w.Autotrust != nil {
			return fmt.Sprintf("%v", *w.Autotrust)
		}
	}
	return ""
}

// SetWorkdirValue sets a workdir config value
func SetWorkdirValue(w *cfgtypes.WorkdirSettings, key, value string) {
	switch key {
	case "workdir.path":
		w.Path = value
	case "workdir.automount":
		b := value == "true"
		w.Automount = &b
	case "workdir.readonly":
		b := value == "true"
		w.Readonly = &b
	case "workdir.autotrust":
		b := value == "true"
		w.Autotrust = &b
	}
}

// UnsetWorkdirValue clears a workdir config value
func UnsetWorkdirValue(w *cfgtypes.WorkdirSettings, key string) {
	switch key {
	case "workdir.path":
		w.Path = ""
	case "workdir.automount":
		w.Automount = nil
	case "workdir.readonly":
		w.Readonly = nil
	case "workdir.autotrust":
		w.Autotrust = nil
	}
}
