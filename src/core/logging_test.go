package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogCommand_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	LogCommand(logFile, "/home/user", "test-container", []string{"--help"})

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestLogCommand_AppendsEntry(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	LogCommand(logFile, "/home/user", "test-container", []string{"--help"})
	LogCommand(logFile, "/home/user", "test-container", []string{"run", "test"})

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(lines))
	}
}

func TestLogCommand_ContainsExpectedFields(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	LogCommand(logFile, "/home/user/project", "my-container", []string{"arg1", "arg2"})

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logEntry := string(content)

	expectedFields := []string{
		"PWD: /home/user/project",
		"Container: my-container",
		"Command: arg1 arg2",
	}

	for _, field := range expectedFields {
		if !strings.Contains(logEntry, field) {
			t.Errorf("Log entry missing %q\nGot: %s", field, logEntry)
		}
	}
}

func TestLogCommand_HasTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	LogCommand(logFile, "/home/user", "test-container", []string{})

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Check for timestamp format [YYYY-MM-DD HH:MM:SS]
	logEntry := string(content)
	if !strings.HasPrefix(logEntry, "[") || !strings.Contains(logEntry, "]") {
		t.Error("Log entry missing timestamp brackets")
	}
}

func TestLogCommand_EmptyArgs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	LogCommand(logFile, "/home/user", "test-container", []string{})

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "Command: \n") {
		t.Errorf("Expected empty command, got: %s", string(content))
	}
}

func TestLogCommand_InvalidPath(t *testing.T) {
	// Should not panic on invalid path
	LogCommand("/nonexistent/dir/log.txt", "/home/user", "test-container", []string{})
}
