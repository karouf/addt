package config

import (
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestAuthAutologinKeyValidation(t *testing.T) {
	if !IsValidKey("auth.autologin") {
		t.Error("IsValidKey(auth.autologin) = false, want true")
	}
}

func TestAuthMethodKeyValidation(t *testing.T) {
	if !IsValidKey("auth.method") {
		t.Error("IsValidKey(auth.method) = false, want true")
	}
}

func TestAuthAutologinGetValue(t *testing.T) {
	b := true
	cfg := &cfgtypes.GlobalConfig{
		Auth: &cfgtypes.AuthSettings{
			Autologin: &b,
		},
	}

	got := GetValue(cfg, "auth.autologin")
	if got != "true" {
		t.Errorf("GetValue(auth.autologin) = %q, want %q", got, "true")
	}

	// Test with nil Auth
	nilCfg := &cfgtypes.GlobalConfig{}
	if got := GetValue(nilCfg, "auth.autologin"); got != "" {
		t.Errorf("GetValue(auth.autologin) with nil Auth = %q, want empty", got)
	}
}

func TestAuthMethodGetValue(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{
		Auth: &cfgtypes.AuthSettings{
			Method: "env",
		},
	}

	got := GetValue(cfg, "auth.method")
	if got != "env" {
		t.Errorf("GetValue(auth.method) = %q, want %q", got, "env")
	}
}

func TestAuthAutologinSetValue(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}

	SetValue(cfg, "auth.autologin", "true")
	if cfg.Auth == nil || cfg.Auth.Autologin == nil || *cfg.Auth.Autologin != true {
		t.Error("Autologin not set correctly to true")
	}

	SetValue(cfg, "auth.autologin", "false")
	if *cfg.Auth.Autologin != false {
		t.Error("Autologin not set correctly to false")
	}
}

func TestAuthMethodSetValue(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}

	SetValue(cfg, "auth.method", "native")
	if cfg.Auth == nil || cfg.Auth.Method != "native" {
		t.Errorf("Method not set correctly, got %q", cfg.Auth.Method)
	}
}

func TestAuthAutologinUnsetValue(t *testing.T) {
	b := true
	cfg := &cfgtypes.GlobalConfig{
		Auth: &cfgtypes.AuthSettings{
			Autologin: &b,
		},
	}

	UnsetValue(cfg, "auth.autologin")
	if cfg.Auth.Autologin != nil {
		t.Error("Autologin should be nil after unset")
	}
}

func TestAuthMethodUnsetValue(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{
		Auth: &cfgtypes.AuthSettings{
			Method: "env",
		},
	}

	UnsetValue(cfg, "auth.method")
	if cfg.Auth.Method != "" {
		t.Error("Method should be empty after unset")
	}
}

func TestAuthAutologinGetDefaultValue(t *testing.T) {
	got := GetDefaultValue("auth.autologin")
	if got != "true" {
		t.Errorf("GetDefaultValue(auth.autologin) = %q, want %q", got, "true")
	}
}

func TestAuthMethodGetDefaultValue(t *testing.T) {
	got := GetDefaultValue("auth.method")
	if got != "auto" {
		t.Errorf("GetDefaultValue(auth.method) = %q, want %q", got, "auto")
	}
}
