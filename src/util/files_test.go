package util

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestExpandTilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"no tilde", "/some/path", "/some/path"},
		{"tilde prefix", "~/Documents", filepath.Join(homeDir, "Documents")},
		{"tilde only slash", "~/", homeDir},
		{"tilde without slash", "~foo", "~foo"},
		{"relative path", "some/path", "some/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandTilde(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandTilde(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAddtHome_Default(t *testing.T) {
	// Ensure ADDT_HOME is not set
	orig := os.Getenv("ADDT_HOME")
	os.Unsetenv("ADDT_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("ADDT_HOME", orig)
		}
	}()

	currentUser, err := user.Current()
	if err != nil {
		t.Skip("cannot determine current user")
	}

	expected := filepath.Join(currentUser.HomeDir, ".addt")
	result := GetAddtHome()
	if result != expected {
		t.Errorf("GetAddtHome() = %q, want %q", result, expected)
	}
}

func TestGetAddtHome_EnvOverride(t *testing.T) {
	orig := os.Getenv("ADDT_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("ADDT_HOME", orig)
		} else {
			os.Unsetenv("ADDT_HOME")
		}
	}()

	os.Setenv("ADDT_HOME", "/custom/addt/home")
	result := GetAddtHome()
	if result != "/custom/addt/home" {
		t.Errorf("GetAddtHome() = %q, want %q", result, "/custom/addt/home")
	}
}

func TestGetAddtHome_TildeExpansion(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	orig := os.Getenv("ADDT_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("ADDT_HOME", orig)
		} else {
			os.Unsetenv("ADDT_HOME")
		}
	}()

	os.Setenv("ADDT_HOME", "~/my-addt")
	expected := filepath.Join(homeDir, "my-addt")
	result := GetAddtHome()
	if result != expected {
		t.Errorf("GetAddtHome() = %q, want %q", result, expected)
	}
}
