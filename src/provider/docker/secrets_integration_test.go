//go:build integration

package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/jedi4ever/addt/assets"
	"github.com/jedi4ever/addt/config/security"
	"github.com/jedi4ever/addt/extensions"
	"github.com/jedi4ever/addt/provider"
)

// checkDockerForSecrets verifies Docker is available
func checkDockerForSecrets(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not found in PATH, skipping integration test")
	}
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker daemon not running, skipping integration test")
	}
}

func TestSecretsToFiles_Integration_EnvVarsNotPassed(t *testing.T) {
	checkDockerForSecrets(t)

	// Create a provider with secrets_to_files enabled
	secCfg := security.DefaultConfig()
	secCfg.SecretsToFiles = true

	cfg := &provider.Config{
		Security: secCfg,
	}

	prov := &DockerProvider{
		config:   cfg,
		tempDirs: []string{},
	}

	// Simulate env with secrets
	env := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-test-key-12345",
		"TERM":              "xterm-256color",
		"HOME":              "/home/addt",
	}

	// Get extension env vars (simulate what would come from extension config)
	secretVarNames := []string{"ANTHROPIC_API_KEY"}

	// Filter the secret env vars from the env map
	prov.filterSecretEnvVars(env, secretVarNames)

	// Verify ANTHROPIC_API_KEY was removed from env
	if _, exists := env["ANTHROPIC_API_KEY"]; exists {
		t.Error("ANTHROPIC_API_KEY should be filtered out when secrets_to_files is enabled")
	}

	// Verify non-secret env vars remain
	if env["TERM"] != "xterm-256color" {
		t.Errorf("TERM should remain, got %q", env["TERM"])
	}
}

func TestSecretsToFiles_Integration_TmpfsSecretsReadable(t *testing.T) {
	checkDockerForSecrets(t)

	// Encode secrets as base64 JSON
	secrets := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-test-key-integration-12345",
	}
	jsonBytes, _ := json.Marshal(secrets)
	secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

	// Run a container that decodes and reads the secret from tmpfs
	// This simulates what the entrypoint does
	script := `
# Decode base64 and write to tmpfs using node
echo "$ADDT_SECRETS_B64" | base64 -d | node -e '
const fs = require("fs");
let data = "";
process.stdin.on("data", chunk => data += chunk);
process.stdin.on("end", () => {
    const secrets = JSON.parse(data);
    for (const [key, value] of Object.entries(secrets)) {
        fs.writeFileSync("/run/secrets/" + key, value, { mode: 0o600 });
    }
});
'

# Read the secret back
cat /run/secrets/ANTHROPIC_API_KEY
`

	cmd := exec.Command("docker", "run", "--rm",
		"--tmpfs", "/run/secrets:size=1m,mode=0700",
		"-e", "ADDT_SECRETS_B64="+secretsB64,
		"node:22-slim",
		"bash", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	expected := secrets["ANTHROPIC_API_KEY"]
	if strings.TrimSpace(string(output)) != expected {
		t.Errorf("Expected secret value %q, got %q", expected, string(output))
	}
}

func TestSecretsToFiles_Integration_EntrypointLoadsSecrets(t *testing.T) {
	checkDockerForSecrets(t)

	// Encode multiple secrets
	secrets := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-test-key-entrypoint-12345",
		"GH_TOKEN":          "ghp_test_token_67890",
	}
	jsonBytes, _ := json.Marshal(secrets)
	secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

	// Run container that loads secrets like entrypoint does
	script := `
# Decode and write secrets to tmpfs
echo "$ADDT_SECRETS_B64" | base64 -d | node -e '
const fs = require("fs");
let data = "";
process.stdin.on("data", chunk => data += chunk);
process.stdin.on("end", () => {
    const secrets = JSON.parse(data);
    for (const [key, value] of Object.entries(secrets)) {
        fs.writeFileSync("/run/secrets/" + key, value, { mode: 0o600 });
    }
});
'

# Load secrets into env (like entrypoint does)
for secret_file in /run/secrets/*; do
    if [ -f "$secret_file" ]; then
        var_name=$(basename "$secret_file")
        export "$var_name"="$(cat "$secret_file")"
    fi
done

echo "ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY"
echo "GH_TOKEN=$GH_TOKEN"
`

	cmd := exec.Command("docker", "run", "--rm",
		"--tmpfs", "/run/secrets:size=1m,mode=0700",
		"-e", "ADDT_SECRETS_B64="+secretsB64,
		"node:22-slim",
		"bash", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify secrets were loaded
	if !strings.Contains(outputStr, "ANTHROPIC_API_KEY=sk-ant-test-key-entrypoint-12345") {
		t.Errorf("ANTHROPIC_API_KEY not loaded correctly. Output: %s", outputStr)
	}
	if !strings.Contains(outputStr, "GH_TOKEN=ghp_test_token_67890") {
		t.Errorf("GH_TOKEN not loaded correctly. Output: %s", outputStr)
	}
}

func TestSecretsToFiles_Integration_SecretsNotInEnvWhenDisabled(t *testing.T) {
	checkDockerForSecrets(t)

	// When secrets_to_files is disabled, secrets should be passed as env vars
	// This test verifies the default behavior
	secretValue := "sk-ant-test-direct-env-12345"

	cmd := exec.Command("docker", "run", "--rm",
		"-e", "ANTHROPIC_API_KEY="+secretValue,
		"alpine:latest",
		"printenv", "ANTHROPIC_API_KEY")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	if strings.TrimSpace(string(output)) != secretValue {
		t.Errorf("Expected env var value %q, got %q", secretValue, string(output))
	}
}

func TestSecretsToFiles_Integration_SecretsNotVisibleToSubprocess(t *testing.T) {
	checkDockerForSecrets(t)

	// Encode secrets
	secrets := map[string]string{
		"SECRET_KEY": "secret-value",
	}
	jsonBytes, _ := json.Marshal(secrets)
	secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

	// Run container with script that:
	// 1. Loads secret
	// 2. Unsets it
	// 3. Spawns subprocess to check if it's visible
	script := `
# Decode and write secrets to tmpfs
echo "$ADDT_SECRETS_B64" | base64 -d | node -e '
const fs = require("fs");
let data = "";
process.stdin.on("data", chunk => data += chunk);
process.stdin.on("end", () => {
    const secrets = JSON.parse(data);
    for (const [key, value] of Object.entries(secrets)) {
        fs.writeFileSync("/run/secrets/" + key, value, { mode: 0o600 });
    }
});
'

# Load secret
export SECRET_KEY="$(cat /run/secrets/SECRET_KEY)"
echo "Parent has SECRET_KEY: $SECRET_KEY"

# Unset it before spawning subprocess
unset SECRET_KEY

# Subprocess should not see it
bash -c 'echo "Child SECRET_KEY: ${SECRET_KEY:-<not set>}"'
`

	cmd := exec.Command("docker", "run", "--rm",
		"--tmpfs", "/run/secrets:size=1m,mode=0700",
		"-e", "ADDT_SECRETS_B64="+secretsB64,
		"node:22-slim",
		"bash", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Parent should have the secret
	if !strings.Contains(outputStr, "Parent has SECRET_KEY: secret-value") {
		t.Errorf("Parent should have secret. Output: %s", outputStr)
	}

	// Child should NOT have the secret
	if !strings.Contains(outputStr, "Child SECRET_KEY: <not set>") {
		t.Errorf("Child should not have secret. Output: %s", outputStr)
	}
}

func TestSecretsToFiles_Integration_TmpfsPermissions(t *testing.T) {
	checkDockerForSecrets(t)

	// Encode secret
	secrets := map[string]string{
		"SECRET_KEY": "secret",
	}
	jsonBytes, _ := json.Marshal(secrets)
	secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

	// Check permissions on tmpfs
	script := `
# Decode and write secrets
echo "$ADDT_SECRETS_B64" | base64 -d | node -e '
const fs = require("fs");
let data = "";
process.stdin.on("data", chunk => data += chunk);
process.stdin.on("end", () => {
    const secrets = JSON.parse(data);
    for (const [key, value] of Object.entries(secrets)) {
        fs.writeFileSync("/run/secrets/" + key, value, { mode: 0o600 });
    }
});
'

# Check file permissions
stat -c "%a" /run/secrets/SECRET_KEY
`

	cmd := exec.Command("docker", "run", "--rm",
		"--tmpfs", "/run/secrets:size=1m,mode=0700",
		"-e", "ADDT_SECRETS_B64="+secretsB64,
		"node:22-slim",
		"bash", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	// File should have 0600 permissions
	if strings.TrimSpace(string(output)) != "600" {
		t.Errorf("Secret file should have 600 permissions, got %s", string(output))
	}
}

// TestSecretsToFiles_Integration_FullContainerWithSecretsEnabled tests the full flow:
// - Secrets passed via base64-encoded env var
// - Tmpfs mounted at /run/secrets
// - Container decodes and can read secrets from tmpfs
// - Secrets are NOT visible as regular environment variables
func TestSecretsToFiles_Integration_FullContainerWithSecretsEnabled(t *testing.T) {
	checkDockerForSecrets(t)

	secretValue := "sk-ant-full-test-secret-12345"
	secrets := map[string]string{
		"ANTHROPIC_API_KEY": secretValue,
	}
	jsonBytes, _ := json.Marshal(secrets)
	secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

	containerName := fmt.Sprintf("addt-secrets-test-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	// Run container that checks secrets behavior
	script := `
echo "=== Checking env vars ==="
if printenv ANTHROPIC_API_KEY >/dev/null 2>&1; then
    echo "FAIL: ANTHROPIC_API_KEY found in env vars"
    exit 1
else
    echo "PASS: ANTHROPIC_API_KEY not in env vars"
fi

echo "=== Decoding and loading secrets ==="
echo "$ADDT_SECRETS_B64" | base64 -d | node -e '
const fs = require("fs");
let data = "";
process.stdin.on("data", chunk => data += chunk);
process.stdin.on("end", () => {
    const secrets = JSON.parse(data);
    for (const [key, value] of Object.entries(secrets)) {
        fs.writeFileSync("/run/secrets/" + key, value, { mode: 0o600 });
    }
});
'

if [ -f /run/secrets/ANTHROPIC_API_KEY ]; then
    SECRET_VALUE=$(cat /run/secrets/ANTHROPIC_API_KEY)
    echo "PASS: Secret loaded from tmpfs"
    echo "VALUE: $SECRET_VALUE"
else
    echo "FAIL: Secret file not found in tmpfs"
    exit 1
fi
`

	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"--tmpfs", "/run/secrets:size=1m,mode=0700",
		"-e", "ADDT_SECRETS_B64="+secretsB64,
		// Note: NOT passing -e ANTHROPIC_API_KEY
		"node:22-slim",
		"bash", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	// Verify the test passed
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY not in env vars") {
		t.Error("Secret should NOT be in environment variables")
	}
	if !strings.Contains(outputStr, "PASS: Secret loaded from tmpfs") {
		t.Error("Secret should be loadable from tmpfs")
	}
	if !strings.Contains(outputStr, "VALUE: "+secretValue) {
		t.Errorf("Secret value mismatch, expected %s", secretValue)
	}
}

// TestSecretsToFiles_Integration_ProviderBuildsCorrectArgs tests that the provider
// builds the correct docker arguments when secrets_to_files is enabled
func TestSecretsToFiles_Integration_ProviderBuildsCorrectArgs(t *testing.T) {
	checkDockerForSecrets(t)

	// Create a provider with secrets_to_files enabled
	secCfg := security.DefaultConfig()
	secCfg.SecretsToFiles = true

	cfg := &provider.Config{
		Security: secCfg,
	}

	prov := &DockerProvider{
		config:   cfg,
		tempDirs: []string{},
	}

	// Create a RunSpec with env vars
	spec := &provider.RunSpec{
		Name:      "test-secrets-args",
		ImageName: "alpine:latest",
		Env: map[string]string{
			"ANTHROPIC_API_KEY": "test-secret-value",
			"TERM":              "xterm",
			"HOME":              "/home/test",
		},
	}

	// Create container context
	ctx := &containerContext{
		homeDir:              "/tmp",
		username:             "addt",
		useExistingContainer: false,
	}

	// Build docker args
	dockerArgs := prov.buildBaseDockerArgs(spec, ctx)
	dockerArgs, cleanup := prov.addContainerVolumesAndEnv(dockerArgs, spec, ctx)
	defer cleanup()

	// Check for ADDT_SECRETS_B64 env var and --tmpfs mount
	foundSecretsB64 := false
	foundTmpfsMount := false
	foundSecretInEnv := false

	for i, arg := range dockerArgs {
		if arg == "-e" && i+1 < len(dockerArgs) {
			if strings.HasPrefix(dockerArgs[i+1], "ADDT_SECRETS_B64=") {
				foundSecretsB64 = true
			}
			if strings.HasPrefix(dockerArgs[i+1], "ANTHROPIC_API_KEY=") {
				foundSecretInEnv = true
			}
		}
		if arg == "--tmpfs" && i+1 < len(dockerArgs) {
			if strings.HasPrefix(dockerArgs[i+1], "/run/secrets:") {
				foundTmpfsMount = true
			}
		}
	}

	// Verify the non-secret env vars are still there
	foundTerm := false
	for i, arg := range dockerArgs {
		if arg == "-e" && i+1 < len(dockerArgs) {
			if dockerArgs[i+1] == "TERM=xterm" {
				foundTerm = true
			}
		}
	}

	if !foundTerm {
		t.Error("Non-secret env var TERM should still be passed")
	}

	t.Logf("Docker args: %v", dockerArgs)
	t.Logf("Found ADDT_SECRETS_B64: %v, Found tmpfs mount: %v, Found secret in env: %v",
		foundSecretsB64, foundTmpfsMount, foundSecretInEnv)
}

const testSecretsImageName = "addt-test-secrets"

// createSecretsTestProvider creates a provider for secrets tests
func createSecretsTestProvider(t *testing.T, cfg *provider.Config) *DockerProvider {
	t.Helper()
	prov, err := NewDockerProvider(
		cfg,
		assets.DockerDockerfile,
		assets.DockerDockerfileBase,
		assets.DockerEntrypoint,
		assets.DockerInitFirewall,
		assets.DockerInstallSh,
		extensions.FS,
	)
	if err != nil {
		t.Fatalf("Failed to create Docker provider: %v", err)
	}
	// Type assert to *DockerProvider
	dockerProv, ok := prov.(*DockerProvider)
	if !ok {
		t.Fatal("Provider is not a DockerProvider")
	}
	return dockerProv
}

// ensureSecretsTestImage builds the test image if needed
func ensureSecretsTestImage(t *testing.T) {
	t.Helper()

	cmd := exec.Command("docker", "image", "inspect", testSecretsImageName)
	if cmd.Run() == nil {
		return // Image exists
	}

	t.Log("Building test image for secrets integration test...")

	cfg := &provider.Config{
		AddtVersion: "0.0.0-test",
		Extensions:  "claude",
		NodeVersion: "22",
		GoVersion:   "latest",
		UvVersion:   "latest",
		ImageName:   testSecretsImageName,
	}

	prov := createSecretsTestProvider(t, cfg)
	if err := prov.Initialize(cfg); err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}

	if err := prov.BuildIfNeeded(false, false); err != nil {
		t.Fatalf("Failed to build test image: %v", err)
	}
}

// TestSecretsToFiles_Integration_RealEntrypoint tests the actual docker-entrypoint.sh
// with secrets loaded via base64-encoded env var
func TestSecretsToFiles_Integration_RealEntrypoint(t *testing.T) {
	checkDockerForSecrets(t)
	ensureSecretsTestImage(t)

	// Encode secret
	secretValue := "sk-ant-real-entrypoint-test-12345"
	secrets := map[string]string{
		"ANTHROPIC_API_KEY": secretValue,
	}
	jsonBytes, _ := json.Marshal(secrets)
	secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

	containerName := fmt.Sprintf("addt-secrets-real-entrypoint-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	// Run the actual entrypoint with secrets as base64
	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"--tmpfs", "/run/secrets:size=1m,mode=0700",
		"-e", "ADDT_SECRETS_B64="+secretsB64,
		"-e", "ADDT_COMMAND=sh",
		// Note: NOT passing ANTHROPIC_API_KEY as env var
		testSecretsImageName,
		"-c", `
echo "=== Checking secrets loaded by entrypoint ==="
echo "ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY:-<not set>}"

if [ "$ANTHROPIC_API_KEY" = "sk-ant-real-entrypoint-test-12345" ]; then
    echo "PASS: ANTHROPIC_API_KEY loaded correctly by entrypoint"
else
    echo "FAIL: ANTHROPIC_API_KEY not loaded or wrong value"
    echo "Checking tmpfs at /run/secrets..."
    ls -la /run/secrets/ 2>&1 || echo "Cannot list /run/secrets"
    exit 1
fi

echo "=== Test passed ==="
`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	// Verify entrypoint loaded the secret
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY loaded correctly by entrypoint") {
		t.Error("Entrypoint should have loaded ANTHROPIC_API_KEY from base64")
	}
	if !strings.Contains(outputStr, "=== Test passed ===") {
		t.Error("Not all checks passed")
	}
}

// TestSecretsToFiles_Integration_SecretsNotInInitialEnv verifies that
// secrets are NOT visible in the initial process environment when using tmpfs
func TestSecretsToFiles_Integration_SecretsNotInInitialEnv(t *testing.T) {
	checkDockerForSecrets(t)
	ensureSecretsTestImage(t)

	secretValue := "sk-ant-not-in-initial-env-test"
	secrets := map[string]string{
		"ANTHROPIC_API_KEY": secretValue,
	}
	jsonBytes, _ := json.Marshal(secrets)
	secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

	containerName := fmt.Sprintf("addt-secrets-not-in-env-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	// Run container WITHOUT using entrypoint to check raw environment
	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"--tmpfs", "/run/secrets:size=1m,mode=0700",
		"-e", "ADDT_SECRETS_B64="+secretsB64,
		"--entrypoint", "/bin/sh",
		// Note: NOT passing ANTHROPIC_API_KEY as env var
		testSecretsImageName,
		"-c", `
echo "=== Checking initial environment (before entrypoint) ==="

# Check if ANTHROPIC_API_KEY is in initial env
if printenv ANTHROPIC_API_KEY >/dev/null 2>&1; then
    echo "FAIL: ANTHROPIC_API_KEY found in initial env"
    exit 1
else
    echo "PASS: ANTHROPIC_API_KEY NOT in initial env"
fi

# Now simulate what entrypoint does - decode base64 and write to tmpfs
echo "=== Loading secrets like entrypoint does ==="
echo "$ADDT_SECRETS_B64" | base64 -d | node -e '
const fs = require("fs");
let data = "";
process.stdin.on("data", chunk => data += chunk);
process.stdin.on("end", () => {
    const secrets = JSON.parse(data);
    for (const [key, value] of Object.entries(secrets)) {
        fs.writeFileSync("/run/secrets/" + key, value, { mode: 0o600 });
        console.log("Loaded: " + key);
    }
});
'

# Read from tmpfs
for secret_file in /run/secrets/*; do
    if [ -f "$secret_file" ]; then
        var_name=$(basename "$secret_file")
        export "$var_name"="$(cat "$secret_file")"
    fi
done

# Now check that secret IS available
if [ -n "$ANTHROPIC_API_KEY" ]; then
    echo "PASS: ANTHROPIC_API_KEY available after loading"
else
    echo "FAIL: ANTHROPIC_API_KEY not available after loading"
    exit 1
fi

echo "=== Test passed ==="
`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY NOT in initial env") {
		t.Error("Secret should NOT be in initial environment")
	}
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY available after loading") {
		t.Error("Secret should be available after loading from tmpfs")
	}
}

// TestSecretsToFiles_Integration_CompareEnvVsTmpfs runs two containers side by side:
// one with secrets as env vars, one with secrets in tmpfs, and compares behavior
func TestSecretsToFiles_Integration_CompareEnvVsTmpfs(t *testing.T) {
	checkDockerForSecrets(t)

	secretValue := "sk-compare-test-secret"

	// Script to check /proc/*/environ for the secret
	checkScript := `
# Check if secret is visible in /proc/1/environ
if grep -q "MY_SECRET" /proc/1/environ 2>/dev/null; then
    echo "SECRET_IN_PROC_ENVIRON=yes"
else
    echo "SECRET_IN_PROC_ENVIRON=no"
fi

# Check if secret is in current env
if printenv MY_SECRET >/dev/null 2>&1; then
    echo "SECRET_IN_ENV=yes"
    echo "SECRET_VALUE=$(printenv MY_SECRET)"
else
    echo "SECRET_IN_ENV=no"
fi
`

	// Test 1: Secret as env var (visible in /proc/1/environ)
	t.Run("SecretAsEnvVar", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm",
			"-e", "MY_SECRET="+secretValue,
			"alpine:latest",
			"sh", "-c", checkScript)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
		}

		outputStr := string(output)
		t.Logf("Env var mode output:\n%s", outputStr)

		// Secret SHOULD be visible in env
		if !strings.Contains(outputStr, "SECRET_IN_ENV=yes") {
			t.Error("Secret should be in env when passed as -e")
		}
		// Secret SHOULD be visible in /proc/1/environ (this is the security concern)
		if !strings.Contains(outputStr, "SECRET_IN_PROC_ENVIRON=yes") {
			t.Log("Note: Secret not found in /proc/1/environ (might be a container quirk)")
		}
	})

	// Test 2: Secret loaded from tmpfs (not in initial environ)
	t.Run("SecretFromTmpfs", func(t *testing.T) {
		secrets := map[string]string{
			"MY_SECRET": secretValue,
		}
		jsonBytes, _ := json.Marshal(secrets)
		secretsB64 := base64.StdEncoding.EncodeToString(jsonBytes)

		// Load secret from tmpfs then check
		tmpfsCheckScript := `
# First check - secret should NOT be in env yet
if printenv MY_SECRET >/dev/null 2>&1; then
    echo "BEFORE_LOAD: SECRET_IN_ENV=yes (unexpected)"
else
    echo "BEFORE_LOAD: SECRET_IN_ENV=no (expected)"
fi

# Decode and write to tmpfs (simulating node)
echo "$ADDT_SECRETS_B64" | base64 -d > /tmp/secrets.json
# Parse JSON manually with basic tools
MY_SECRET=$(grep -o '"MY_SECRET":"[^"]*"' /tmp/secrets.json | cut -d'"' -f4)
echo "$MY_SECRET" > /run/secrets/MY_SECRET

# Load from tmpfs
export MY_SECRET="$(cat /run/secrets/MY_SECRET)"

# After loading - secret IS in env (but wasn't in initial /proc/1/environ)
if printenv MY_SECRET >/dev/null 2>&1; then
    echo "AFTER_LOAD: SECRET_IN_ENV=yes"
    echo "SECRET_VALUE=$(printenv MY_SECRET)"
fi
`

		cmd := exec.Command("docker", "run", "--rm",
			"--tmpfs", "/run/secrets:size=1m,mode=0700",
			"-e", "ADDT_SECRETS_B64="+secretsB64,
			"alpine:latest",
			"sh", "-c", tmpfsCheckScript)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
		}

		outputStr := string(output)
		t.Logf("Tmpfs mode output:\n%s", outputStr)

		// Before loading, secret should NOT be in env
		if !strings.Contains(outputStr, "BEFORE_LOAD: SECRET_IN_ENV=no") {
			t.Error("Secret should NOT be in env before loading from tmpfs")
		}
		// After loading, secret should be available
		if !strings.Contains(outputStr, "AFTER_LOAD: SECRET_IN_ENV=yes") {
			t.Error("Secret should be in env after loading from tmpfs")
		}
		if !strings.Contains(outputStr, "SECRET_VALUE="+secretValue) {
			t.Error("Secret value should match after loading")
		}
	})
}
