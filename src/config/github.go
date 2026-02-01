package config

import (
	"os/exec"
	"strings"
)

// DetectGitHubToken attempts to get the GitHub token from the gh CLI
func DetectGitHubToken() string {
	if _, err := exec.LookPath("gh"); err != nil {
		return ""
	}
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
