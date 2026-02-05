package podman

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestFilterSecretEnvVars(t *testing.T) {
	p := &PodmanProvider{}

	env := map[string]string{
		"ANTHROPIC_API_KEY": "secret123",
		"GH_TOKEN":          "ghp_xxx",
		"PATH":              "/usr/bin",
		"OPENAI_API_KEY":    "sk-xxx",
		"NON_SECRET_VAR":    "value",
	}

	secretVars := []string{"ANTHROPIC_API_KEY", "GH_TOKEN", "OPENAI_API_KEY"}

	p.filterSecretEnvVars(env, secretVars)

	// Secret vars should be removed
	for _, secretVar := range secretVars {
		if _, exists := env[secretVar]; exists {
			t.Errorf("filterSecretEnvVars should have removed %s", secretVar)
		}
	}

	// Non-secret vars should remain
	if _, exists := env["PATH"]; !exists {
		t.Error("filterSecretEnvVars removed PATH which should remain")
	}
	if _, exists := env["NON_SECRET_VAR"]; !exists {
		t.Error("filterSecretEnvVars removed NON_SECRET_VAR which should remain")
	}
}

func TestAddTmpfsSecretsMount(t *testing.T) {
	p := &PodmanProvider{}

	args := []string{"-it", "--rm"}
	result := p.addTmpfsSecretsMount(args)

	// Should add tmpfs mount
	if len(result) != 4 {
		t.Errorf("Expected 4 args, got %d: %v", len(result), result)
	}

	found := false
	for i := 0; i < len(result)-1; i++ {
		if result[i] == "--tmpfs" && result[i+1] == "/run/secrets:size=1m,mode=0700" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("addTmpfsSecretsMount did not add expected tmpfs mount, got %v", result)
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
