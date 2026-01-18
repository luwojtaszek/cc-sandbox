//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// replaceBinary replaces the current executable with a new one.
// On Windows, we cannot directly replace a running executable,
// so we rename the old one first, then move the new one in place.
func replaceBinary(currentPath, newPath string) error {
	// Create backup path for the old binary
	backupPath := currentPath + ".old"

	// Remove any existing backup
	os.Remove(backupPath)

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

	// Schedule old binary for deletion on next reboot (best effort)
	// Windows will clean it up, or user can delete manually
	// We leave it for now as .old file

	// Try to delete old binary (might fail if still in use)
	os.Remove(backupPath)

	return nil
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
