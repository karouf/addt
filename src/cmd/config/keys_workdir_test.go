package config

import (
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestWorkdirAutotrustKeyValidation(t *testing.T) {
	if !IsValidKey("workdir.autotrust") {
		t.Error("IsValidKey(workdir.autotrust) = false, want true")
	}
}

func TestWorkdirAutotrustGetValue(t *testing.T) {
	b := true
	cfg := &cfgtypes.GlobalConfig{
		Workdir: &cfgtypes.WorkdirSettings{
			Autotrust: &b,
		},
	}

	got := GetValue(cfg, "workdir.autotrust")
	if got != "true" {
		t.Errorf("GetValue(workdir.autotrust) = %q, want %q", got, "true")
	}

	// Test with nil Workdir
	nilCfg := &cfgtypes.GlobalConfig{}
	if got := GetValue(nilCfg, "workdir.autotrust"); got != "" {
		t.Errorf("GetValue(workdir.autotrust) with nil Workdir = %q, want empty", got)
	}
}

func TestWorkdirAutotrustSetValue(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}

	SetValue(cfg, "workdir.autotrust", "true")
	if cfg.Workdir == nil || cfg.Workdir.Autotrust == nil || *cfg.Workdir.Autotrust != true {
		t.Error("Autotrust not set correctly to true")
	}

	SetValue(cfg, "workdir.autotrust", "false")
	if *cfg.Workdir.Autotrust != false {
		t.Error("Autotrust not set correctly to false")
	}
}

func TestWorkdirAutotrustUnsetValue(t *testing.T) {
	b := true
	cfg := &cfgtypes.GlobalConfig{
		Workdir: &cfgtypes.WorkdirSettings{
			Autotrust: &b,
		},
	}

	UnsetValue(cfg, "workdir.autotrust")
	if cfg.Workdir.Autotrust != nil {
		t.Error("Autotrust should be nil after unset")
	}
}

func TestWorkdirAutotrustGetDefaultValue(t *testing.T) {
	got := GetDefaultValue("workdir.autotrust")
	if got != "true" {
		t.Errorf("GetDefaultValue(workdir.autotrust) = %q, want %q", got, "true")
	}
}
