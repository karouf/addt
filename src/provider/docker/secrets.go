package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// prepareSecretsJSON collects secret environment variables and returns them as JSON
// Returns the JSON string and the list of secret variable names
func (p *DockerProvider) prepareSecretsJSON(imageName string, env map[string]string) (string, []string, error) {
	// Get extension env vars (these are the "secrets")
	secretVarNames := p.GetExtensionEnvVars(imageName)
	if len(secretVarNames) == 0 {
		return "", nil, nil
	}

	// Collect secrets that have values
	secrets := make(map[string]string)
	writtenSecrets := []string{}
	for _, varName := range secretVarNames {
		value, exists := env[varName]
		if !exists || value == "" {
			continue
		}
		secrets[varName] = value
		writtenSecrets = append(writtenSecrets, varName)
	}

	if len(writtenSecrets) == 0 {
		return "", nil, nil
	}

	// Encode as JSON
	jsonBytes, err := json.Marshal(secrets)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal secrets: %w", err)
	}

	return string(jsonBytes), writtenSecrets, nil
}

// copySecretsToContainer copies secrets JSON to the container's tmpfs via docker cp
func (p *DockerProvider) copySecretsToContainer(containerName, secretsJSON string) error {
	// Write secrets to a temp file
	tmpFile, err := os.CreateTemp("", "addt-secrets-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(secretsJSON); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write secrets: %w", err)
	}
	tmpFile.Close()

	// Set restrictive permissions
	if err := os.Chmod(tmpPath, 0600); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Copy to container's /run/secrets/.secrets
	cmd := exec.Command("docker", "cp", tmpPath, containerName+":/run/secrets/.secrets")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker cp failed: %w\n%s", err, string(output))
	}

	return nil
}

// copySecretsToContainerPodman copies secrets JSON to the container's tmpfs via podman cp
func copySecretsToContainerPodman(containerName, secretsJSON string) error {
	// Write secrets to a temp file
	tmpFile, err := os.CreateTemp("", "addt-secrets-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(secretsJSON); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write secrets: %w", err)
	}
	tmpFile.Close()

	// Set restrictive permissions
	if err := os.Chmod(tmpPath, 0600); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Copy to container's /run/secrets/.secrets
	cmd := exec.Command("podman", "cp", tmpPath, containerName+":/run/secrets/.secrets")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("podman cp failed: %w\n%s", err, string(output))
	}

	return nil
}

// addTmpfsSecretsMount adds a tmpfs mount for secrets at /run/secrets
func (p *DockerProvider) addTmpfsSecretsMount(dockerArgs []string) []string {
	return append(dockerArgs, "--tmpfs", "/run/secrets:size=1m,mode=0700")
}

// filterSecretEnvVars removes secret env vars from the env map
// This prevents secrets from being passed as -e flags
func (p *DockerProvider) filterSecretEnvVars(env map[string]string, secretVarNames []string) {
	for _, varName := range secretVarNames {
		delete(env, varName)
	}
}

// writeSecretsFile writes secrets JSON to a file for later docker cp
func writeSecretsFile(secretsJSON string) (string, error) {
	// Create secrets directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}

	secretsDir := filepath.Join(homeDir, ".addt", "secrets")
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create secrets dir: %w", err)
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp(secretsDir, "secrets-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := tmpFile.WriteString(secretsJSON); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write secrets: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to set permissions: %w", err)
	}

	return tmpFile.Name(), nil
}
