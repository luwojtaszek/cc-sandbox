//go:build windows

package main

// getFileGID returns 0 on Windows as Unix-style GIDs don't apply.
// Docker socket mounting with --group-add is not applicable on Windows.
func getFileGID(path string) int {
	return 0
}
