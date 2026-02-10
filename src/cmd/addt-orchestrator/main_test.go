package main

import (
	"testing"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()

	if sm == nil {
		t.Fatal("expected non-nil session manager")
	}

	if sm.sessions == nil {
		t.Fatal("expected non-nil sessions map")
	}

	// Runtime should be detected (either "docker", "rancher", "podman", or "orbstack")
	if sm.runtime != "docker" && sm.runtime != "rancher" && sm.runtime != "podman" && sm.runtime != "orbstack" {
		t.Errorf("expected runtime to be 'docker', 'rancher', 'podman', or 'orbstack', got '%s'", sm.runtime)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"foo", []string{"foo"}},
		{"foo\nbar", []string{"foo", "bar"}},
		{"foo\nbar\n", []string{"foo", "bar"}},
	}

	for _, tc := range tests {
		result := splitLines(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("splitLines(%q): expected %v, got %v", tc.input, tc.expected, result)
			continue
		}
		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("splitLines(%q)[%d]: expected %q, got %q", tc.input, i, tc.expected[i], result[i])
			}
		}
	}
}

func TestSplitTabs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"foo", []string{"foo"}},
		{"foo\tbar", []string{"foo", "bar"}},
		{"foo\tbar\tbaz", []string{"foo", "bar", "baz"}},
	}

	for _, tc := range tests {
		result := splitTabs(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("splitTabs(%q): expected %v, got %v", tc.input, tc.expected, result)
			continue
		}
		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("splitTabs(%q)[%d]: expected %q, got %q", tc.input, i, tc.expected[i], result[i])
			}
		}
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"", "", true},
		{"foo", "foo", true},
		{"foobar", "bar", true},
		{"foo", "bar", false},
		{"Up 2 hours", "Up", true},
		{"Exited (0)", "Up", false},
	}

	for _, tc := range tests {
		result := containsString(tc.s, tc.substr)
		if result != tc.expected {
			t.Errorf("containsString(%q, %q): expected %v, got %v", tc.s, tc.substr, tc.expected, result)
		}
	}
}
