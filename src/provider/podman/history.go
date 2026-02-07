package podman

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jedi4ever/addt/util"
)

// HandleHistoryPersist configures shell history persistence.
// When enabled, mounts per-project history files from ~/.addt/history/<project-hash>/
func (p *PodmanProvider) HandleHistoryPersist(enabled bool, projectDir, username string) []string {
	if !enabled {
		return nil
	}

	historyDir, err := getProjectHistoryDir(projectDir)
	if err != nil {
		fmt.Printf("Warning: failed to create history directory: %v\n", err)
		return nil
	}

	var args []string
	homeInContainer := fmt.Sprintf("/home/%s", username)

	// Create and mount bash history
	bashHistory := filepath.Join(historyDir, "bash_history")
	if err := touchFile(bashHistory); err == nil {
		args = append(args, "-v", fmt.Sprintf("%s:%s/.bash_history", bashHistory, homeInContainer))
	}

	// Create and mount zsh history
	zshHistory := filepath.Join(historyDir, "zsh_history")
	if err := touchFile(zshHistory); err == nil {
		args = append(args, "-v", fmt.Sprintf("%s:%s/.zsh_history", zshHistory, homeInContainer))
	}

	return args
}

// getProjectHistoryDir returns the history directory for a project
// Creates <addt_home>/history/<project-hash>/ if it doesn't exist
func getProjectHistoryDir(projectDir string) (string, error) {
	addtHome := util.GetAddtHome()
	if addtHome == "" {
		return "", fmt.Errorf("failed to determine addt home directory")
	}

	// Create hash of project directory for unique but consistent naming
	hash := sha256.Sum256([]byte(projectDir))
	projectHash := hex.EncodeToString(hash[:8]) // Use first 8 bytes (16 hex chars)

	historyDir := filepath.Join(addtHome, "history", projectHash)
	if err := os.MkdirAll(historyDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create history dir: %w", err)
	}

	return historyDir, nil
}

// touchFile creates an empty file if it doesn't exist
func touchFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}
