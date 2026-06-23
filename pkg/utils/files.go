package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", path, err)
	}
	return nil
}

// EnsureDirAll creates all required project directories.
func EnsureDirAll(baseDir string, dirs []string) error {
	for _, d := range dirs {
		path := filepath.Join(baseDir, d)
		if err := EnsureDir(path); err != nil {
			return err
		}
	}
	return nil
}

// PathExists checks if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsWritable checks if a directory is writable.
func IsWritable(path string) bool {
	testFile := filepath.Join(path, ".datadock_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	return true
}
