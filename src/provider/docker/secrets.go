package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// prepareSecrets collects secret environment variables and returns them as base64-encoded JSON
// Returns the base64 string and the list of secret variable names
func (p *DockerProvider) prepareSecrets(imageName string, env map[string]string) (string, []string, error) {
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

	// Encode as JSON then base64
	jsonBytes, err := json.Marshal(secrets)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal secrets: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return encoded, writtenSecrets, nil
}

// addTmpfsSecretsMount adds a tmpfs mount for secrets at /run/secrets
// The entrypoint will decode secrets from env var and write to this tmpfs
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
