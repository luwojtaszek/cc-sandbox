//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// replaceBinary replaces the current executable with a new one.
// On Windows, we cannot directly replace a running executable,
// so we rename the old one first, then move the new one in place.
func replaceBinary(currentPath, newPath string) error {
	// Create unique backup path with timestamp to avoid race conditions
	// between concurrent update attempts
	timestamp := time.Now().UnixNano()
	backupPath := fmt.Sprintf("%s.old.%d", currentPath, timestamp)

	// Rename current binary to backup
	if err := os.Rename(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to current path
	if err := os.Rename(newPath, currentPath); err != nil {
		// Try to restore backup on failure
		os.Rename(backupPath, currentPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Clean up old backups (including the one we just created)
	cleanupOldBackups(currentPath)

	return nil
}

// cleanupOldBackups removes old backup files (*.old.*) from previous updates.
// This is best-effort; files still in use won't be deleted.
func cleanupOldBackups(currentPath string) {
	dir := filepath.Dir(currentPath)
	baseName := filepath.Base(currentPath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match pattern: cc-sandbox.exe.old.* or cc-sandbox.old.*
		if strings.HasPrefix(name, baseName+".old.") || name == baseName+".old" {
			oldPath := filepath.Join(dir, name)
			os.Remove(oldPath) // Best effort, ignore errors (file may be in use)
		}
	}
}

// getExecutablePath returns the path of the current executable on Windows.
// This handles symlinks and returns the actual path.
func getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
