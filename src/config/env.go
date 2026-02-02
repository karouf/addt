package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(envFile string) error {
	specifiedByUser := envFile != ""
	if envFile == "" {
		envFile = ".env"
	}

	file, err := os.Open(envFile)
	if err != nil {
		// Only warn if user explicitly specified the env file
		if specifiedByUser {
			fmt.Printf("Warning: Specified env file not found: %s\n", envFile)
		}
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			os.Setenv(key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}
	return nil
}
