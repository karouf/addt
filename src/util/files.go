package util

import (
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// GetAddtHome returns the base directory for addt data files.
// Checks ADDT_HOME env var first, then falls back to ~/.addt.
func GetAddtHome() string {
	if v := os.Getenv("ADDT_HOME"); v != "" {
		return ExpandTilde(v)
	}
	currentUser, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(currentUser.HomeDir, ".addt")
}

// ExpandTilde expands a leading "~/" in a path to the user's home directory.
// Returns the path unchanged if it doesn't start with "~/" or if the home
// directory cannot be determined.
func ExpandTilde(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return path
	}
	return filepath.Join(homeDir, path[2:])
}

// SafeCopyFile copies a file if it exists
func SafeCopyFile(src, dst string) {
	if _, err := os.Stat(src); err != nil {
		return
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}
	os.WriteFile(dst, data, 0600)
}

// SafeCopyDir copies a directory recursively if it exists
func SafeCopyDir(src, dst string) {
	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		return
	}

	if err := os.MkdirAll(dst, 0700); err != nil {
		return
	}

	filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return nil
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			os.MkdirAll(dstPath, 0700)
		} else {
			SafeCopyFile(path, dstPath)
		}

		return nil
	})
}
