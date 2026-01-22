//go:build unix

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Cached detection results
var (
	rootlessDockerOnce   sync.Once
	rootlessDockerResult bool

	orbStackOnce   sync.Once
	orbStackResult bool

	podmanAvailableOnce   sync.Once
	podmanAvailableResult bool

	dockerAvailableOnce   sync.Once
	dockerAvailableResult bool
)

// isRootlessDocker detects if Docker is running in rootless mode.
// Results are cached for the lifetime of the process.
func isRootlessDocker() bool {
	rootlessDockerOnce.Do(func() {
		rootlessDockerResult = detectRootlessDocker()
	})
	return rootlessDockerResult
}

// detectRootlessDocker performs the actual rootless Docker detection.
// It uses multiple detection methods for reliability.
func detectRootlessDocker() bool {
	// Method 1: Check docker info output for rootless security option
	cmd := exec.Command("docker", "info", "--format", "{{.SecurityOptions}}")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "rootless") {
		return true
	}

	// Method 2: Check if Docker socket is in user's home directory
	// (typical for rootless Docker installations)
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return false
	}

	sockets := []string{
		filepath.Join(homeDir, ".docker/run/docker.sock"),
		filepath.Join(homeDir, ".local/share/docker/run/docker.sock"),
	}
	for _, sock := range sockets {
		if fileExists(sock) {
			return true
		}
	}

	return false
}

// isPodmanAvailable checks if podman is available on the system.
// Results are cached for the lifetime of the process.
func isPodmanAvailable() bool {
	podmanAvailableOnce.Do(func() {
		cmd := exec.Command("podman", "--version")
		podmanAvailableResult = cmd.Run() == nil
	})
	return podmanAvailableResult
}

// isDockerAvailable checks if docker is available on the system.
// Results are cached for the lifetime of the process.
func isDockerAvailable() bool {
	dockerAvailableOnce.Do(func() {
		cmd := exec.Command("docker", "--version")
		dockerAvailableResult = cmd.Run() == nil
	})
	return dockerAvailableResult
}

// isOrbStack detects if Docker is running via OrbStack (macOS).
// Results are cached for the lifetime of the process.
func isOrbStack() bool {
	orbStackOnce.Do(func() {
		orbStackResult = detectOrbStack()
	})
	return orbStackResult
}

// detectOrbStack performs the actual OrbStack detection.
func detectOrbStack() bool {
	// Check if socket path contains orbstack
	socket := getDefaultDockerSocket()
	if strings.Contains(socket, "orbstack") {
		return true
	}
	// Also check docker info for OrbStack
	cmd := exec.Command("docker", "info", "--format", "{{.OperatingSystem}}")
	output, err := cmd.Output()
	return err == nil && strings.Contains(strings.ToLower(string(output)), "orbstack")
}
