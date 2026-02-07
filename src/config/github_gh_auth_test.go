package config

import (
	"os"
	"testing"
)

func TestHandleGitHubGhAuth_EnvSource(t *testing.T) {
	// Save and restore GH_TOKEN
	origToken := os.Getenv("GH_TOKEN")
	defer func() {
		if origToken != "" {
			os.Setenv("GH_TOKEN", origToken)
		} else {
			os.Unsetenv("GH_TOKEN")
		}
	}()

	// When token_source is "env", no detection should happen
	os.Unsetenv("GH_TOKEN")

	HandleGitHubGhAuth("env")

	if os.Getenv("GH_TOKEN") != "" {
		t.Error("GH_TOKEN should not be set when token_source is 'env'")
	}
}

func TestHandleGitHubGhAuth_SkipsWhenTokenExists(t *testing.T) {
	// Save and restore GH_TOKEN
	origToken := os.Getenv("GH_TOKEN")
	defer func() {
		if origToken != "" {
			os.Setenv("GH_TOKEN", origToken)
		} else {
			os.Unsetenv("GH_TOKEN")
		}
	}()

	// If GH_TOKEN is already set, detection should be skipped
	os.Setenv("GH_TOKEN", "existing-token")

	HandleGitHubGhAuth("gh_auth")

	if os.Getenv("GH_TOKEN") != "existing-token" {
		t.Errorf("GH_TOKEN should remain 'existing-token', got %q", os.Getenv("GH_TOKEN"))
	}
}
