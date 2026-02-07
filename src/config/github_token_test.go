package config

import (
	"testing"
)

func TestHandleGitHubToken_ForwardDisabled(t *testing.T) {
	envVars := []string{"PATH", "GH_TOKEN", "HOME"}
	result := HandleGitHubToken(false, envVars)

	for _, v := range result {
		if v == "GH_TOKEN" {
			t.Error("GH_TOKEN should be removed when forward is disabled")
		}
	}
	if len(result) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(result))
	}
}

func TestHandleGitHubToken_ForwardEnabled(t *testing.T) {
	envVars := []string{"PATH", "GH_TOKEN", "HOME"}
	result := HandleGitHubToken(true, envVars)

	if len(result) != 3 {
		t.Errorf("expected 3 env vars unchanged, got %d", len(result))
	}
	found := false
	for _, v := range result {
		if v == "GH_TOKEN" {
			found = true
		}
	}
	if !found {
		t.Error("GH_TOKEN should be present when forward is enabled")
	}
}

func TestHandleGitHubToken_NoGHToken(t *testing.T) {
	envVars := []string{"PATH", "HOME"}
	result := HandleGitHubToken(false, envVars)

	if len(result) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(result))
	}
}
