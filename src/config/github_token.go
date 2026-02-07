package config

// HandleGitHubToken filters the GH_TOKEN env var from the list
// when token forwarding is disabled.
func HandleGitHubToken(forwardToken bool, envVars []string) []string {
	if !forwardToken {
		filtered := make([]string, 0, len(envVars))
		for _, v := range envVars {
			if v != "GH_TOKEN" {
				filtered = append(filtered, v)
			}
		}
		return filtered
	}
	return envVars
}
