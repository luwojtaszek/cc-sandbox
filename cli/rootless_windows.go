//go:build windows

package main

// isRootlessDocker returns false on Windows as rootless Docker
// is a Linux-specific concept. Docker Desktop on Windows uses
// a different architecture (WSL2 or Hyper-V backend).
func isRootlessDocker() bool {
	return false
}

// isPodmanAvailable returns false on Windows.
// Podman on Windows uses WSL, so treat as Docker.
func isPodmanAvailable() bool {
	return false
}

// isDockerAvailable returns true on Windows assuming Docker Desktop is installed.
func isDockerAvailable() bool {
	return true
}

// isOrbStack returns false on Windows as OrbStack is macOS only.
func isOrbStack() bool {
	return false
}
