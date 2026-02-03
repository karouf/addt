package core

import (
	"strings"
	"testing"

	"github.com/jedi4ever/addt/provider"
)

func TestAddAIContext_WithPorts(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{"3000", "8080"},
		PortRangeStart: 30000,
	}

	env := make(map[string]string)
	AddAIContext(env, cfg)

	if env["ADDT_PORT_MAP"] == "" {
		t.Error("ADDT_PORT_MAP not set")
	}
}

func TestAddAIContext_NoPorts(t *testing.T) {
	cfg := &provider.Config{
		Ports:          []string{},
		PortRangeStart: 30000,
	}

	env := make(map[string]string)
	AddAIContext(env, cfg)

	if _, ok := env["ADDT_PORT_MAP"]; ok {
		t.Error("ADDT_PORT_MAP should not be set when no ports")
	}
}

func TestBuildSystemPromptPortSection_Empty(t *testing.T) {
	prompt := BuildSystemPromptPortSection("")

	if prompt != "" {
		t.Errorf("Expected empty prompt, got %q", prompt)
	}
}

func TestBuildSystemPromptPortSection_Single(t *testing.T) {
	prompt := BuildSystemPromptPortSection("3000:30000")

	expectedPhrases := []string{
		"Port Mapping Information",
		"Container port 3000",
		"Host port 30000",
		"localhost:30000",
		"IMPORTANT",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(prompt, phrase) {
			t.Errorf("Prompt missing %q\nGot: %s", phrase, prompt)
		}
	}
}

func TestBuildSystemPromptPortSection_Multiple(t *testing.T) {
	prompt := BuildSystemPromptPortSection("3000:30000,8080:30001")

	expectedPhrases := []string{
		"Container port 3000",
		"Host port 30000",
		"Container port 8080",
		"Host port 30001",
		"localhost:30000",
		"localhost:30001",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(prompt, phrase) {
			t.Errorf("Prompt missing %q\nGot: %s", phrase, prompt)
		}
	}
}

func TestSplitPortMap_Empty(t *testing.T) {
	mappings := splitPortMap("")

	if mappings != nil {
		t.Errorf("Expected nil, got %v", mappings)
	}
}

func TestSplitPortMap_Single(t *testing.T) {
	mappings := splitPortMap("3000:30000")

	if len(mappings) != 1 {
		t.Fatalf("Expected 1 mapping, got %d", len(mappings))
	}

	if mappings[0] != "3000:30000" {
		t.Errorf("Mapping = %q, want '3000:30000'", mappings[0])
	}
}

func TestSplitPortMap_Multiple(t *testing.T) {
	mappings := splitPortMap("3000:30000,8080:30001,5432:30002")

	if len(mappings) != 3 {
		t.Fatalf("Expected 3 mappings, got %d", len(mappings))
	}

	expected := []string{"3000:30000", "8080:30001", "5432:30002"}
	for i, exp := range expected {
		if mappings[i] != exp {
			t.Errorf("Mapping %d = %q, want %q", i, mappings[i], exp)
		}
	}
}

func TestParsePortMapping_Valid(t *testing.T) {
	container, host := parsePortMapping("3000:30000")

	if container != "3000" {
		t.Errorf("Container = %q, want '3000'", container)
	}

	if host != "30000" {
		t.Errorf("Host = %q, want '30000'", host)
	}
}

func TestParsePortMapping_NoColon(t *testing.T) {
	container, host := parsePortMapping("3000")

	if container != "" || host != "" {
		t.Errorf("Expected empty strings for invalid mapping, got container=%q, host=%q", container, host)
	}
}

func TestFormatPortMappingsForPrompt(t *testing.T) {
	result := formatPortMappingsForPrompt("3000:30000,8080:30001")

	expectedLines := []string{
		"- Container port 3000 → Host port 30000",
		"- Container port 8080 → Host port 30001",
		"localhost:30000",
		"localhost:30001",
	}

	for _, line := range expectedLines {
		if !strings.Contains(result, line) {
			t.Errorf("Result missing %q\nGot: %s", line, result)
		}
	}
}
