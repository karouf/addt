package config

import (
	"os"
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

// HandleGitHubGhAuth auto-detects the GitHub token via `gh auth token`
// when token_source is "gh_auth" and GH_TOKEN is not already set.
func HandleGitHubGhAuth(tokenSource string) {
	if tokenSource == "gh_auth" && os.Getenv("GH_TOKEN") == "" {
		if token := DetectGitHubToken(); token != "" {
			os.Setenv("GH_TOKEN", token)
		}
	}
}
