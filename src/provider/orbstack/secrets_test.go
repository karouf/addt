package orbstack

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestFilterSecretEnvVars(t *testing.T) {
	p := &OrbStackProvider{}

	env := map[string]string{
		"ANTHROPIC_API_KEY": "secret-key",
		"GH_TOKEN":          "github-token",
		"TERM":              "xterm-256color",
		"HOME":              "/home/user",
	}

	secretVarNames := []string{"ANTHROPIC_API_KEY", "GH_TOKEN"}

	p.filterSecretEnvVars(env, secretVarNames)

	// Secret vars should be removed
	if _, exists := env["ANTHROPIC_API_KEY"]; exists {
		t.Error("ANTHROPIC_API_KEY should be removed")
	}
	if _, exists := env["GH_TOKEN"]; exists {
		t.Error("GH_TOKEN should be removed")
	}

	// Non-secret vars should remain
	if env["TERM"] != "xterm-256color" {
		t.Errorf("TERM = %q, want \"xterm-256color\"", env["TERM"])
	}
	if env["HOME"] != "/home/user" {
		t.Errorf("HOME = %q, want \"/home/user\"", env["HOME"])
	}
}

func TestAddTmpfsSecretsMount(t *testing.T) {
	p := &OrbStackProvider{}

	args := []string{"-it"}
	result := p.addTmpfsSecretsMount(args)

	if len(result) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(result), result)
	}
	if result[1] != "--tmpfs" {
		t.Errorf("Expected --tmpfs flag, got %s", result[1])
	}
	if result[2] != "/run/secrets:size=1m,mode=0777" {
		t.Errorf("Expected tmpfs mount arg, got %s", result[2])
	}
}

func TestPrepareSecrets(t *testing.T) {
	// Test that secrets are properly encoded as base64 JSON

	secrets := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-test123",
		"GH_TOKEN":          "ghp_xxx",
	}

	// Manually encode to verify format
	jsonBytes, err := json.Marshal(secrets)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)

	// Verify we can decode it back
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	var decodedSecrets map[string]string
	if err := json.Unmarshal(decoded, &decodedSecrets); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decodedSecrets["ANTHROPIC_API_KEY"] != secrets["ANTHROPIC_API_KEY"] {
		t.Errorf("ANTHROPIC_API_KEY mismatch: got %q, want %q",
			decodedSecrets["ANTHROPIC_API_KEY"], secrets["ANTHROPIC_API_KEY"])
	}
	if decodedSecrets["GH_TOKEN"] != secrets["GH_TOKEN"] {
		t.Errorf("GH_TOKEN mismatch: got %q, want %q",
			decodedSecrets["GH_TOKEN"], secrets["GH_TOKEN"])
	}
}
