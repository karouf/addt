package core

import (
	"github.com/jedi4ever/addt/provider"
)

// AddAIContext adds environment variables that provide context to the AI
// This includes port mappings that get converted to system prompts by the entrypoint
func AddAIContext(env map[string]string, cfg *provider.Config) {
	// Add port map for system prompt generation
	// The entrypoint script reads ADDT_PORT_MAP and generates ADDT_SYSTEM_PROMPT
	// which tells the AI how to communicate port information to users
	portMap := BuildPortMapString(cfg)
	if portMap != "" {
		env["ADDT_PORT_MAP"] = portMap
	}
}

// BuildSystemPromptPortSection generates the port mapping section of the system prompt
// This is what the entrypoint script generates from ADDT_PORT_MAP
// Included here for documentation and testing purposes
func BuildSystemPromptPortSection(portMap string) string {
	if portMap == "" {
		return ""
	}

	return `# Port Mapping Information

When you start a service inside this container on certain ports, tell the user the correct HOST port to access it from their browser.

Port mappings (container→host):
` + formatPortMappingsForPrompt(portMap) + `

IMPORTANT:
- When testing/starting services inside the container, use the container ports (e.g., http://localhost:3000)
- When telling the USER where to access services in their browser, use the HOST ports (e.g., http://localhost:30000)
- Always remind the user to use the host port in their browser`
}

// formatPortMappingsForPrompt formats port mappings for the system prompt
func formatPortMappingsForPrompt(portMap string) string {
	// This mirrors what the entrypoint script does
	// Port map format: "3000:30000,8080:30001"
	// Output format:
	// - Container port 3000 → Host port 30000 (user accesses: http://localhost:30000)
	// - Container port 8080 → Host port 30001 (user accesses: http://localhost:30001)

	result := ""
	mappings := splitPortMap(portMap)
	for _, mapping := range mappings {
		containerPort, hostPort := parsePortMapping(mapping)
		if containerPort != "" && hostPort != "" {
			result += "- Container port " + containerPort + " → Host port " + hostPort +
				" (user accesses: http://localhost:" + hostPort + ")\n"
		}
	}
	return result
}

// splitPortMap splits a comma-separated port map string
func splitPortMap(portMap string) []string {
	if portMap == "" {
		return nil
	}

	var mappings []string
	current := ""
	for _, c := range portMap {
		if c == ',' {
			if current != "" {
				mappings = append(mappings, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		mappings = append(mappings, current)
	}
	return mappings
}

// parsePortMapping parses a single port mapping "container:host"
func parsePortMapping(mapping string) (containerPort, hostPort string) {
	colonIdx := -1
	for i, c := range mapping {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return "", ""
	}
	return mapping[:colonIdx], mapping[colonIdx+1:]
}
