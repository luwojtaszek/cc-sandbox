//go:build unix

package main

import (
	"fmt"
	"os"
)

// replaceBinary replaces the current executable with a new one.
// On Unix, we can rename over the running binary.
func replaceBinary(currentPath, newPath string) error {
	// Make the new binary executable
	if err := os.Chmod(newPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// On Unix, we can atomically replace the binary by renaming
	// The running process keeps its file descriptor to the old binary
	if err := os.Rename(newPath, currentPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}
