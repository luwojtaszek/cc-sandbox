//go:build unix

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// isRootlessDocker detects if Docker is running in rootless mode.
// It uses multiple detection methods for reliability.
func isRootlessDocker() bool {
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
func isPodmanAvailable() bool {
	cmd := exec.Command("podman", "--version")
	return cmd.Run() == nil
}

// isDockerAvailable checks if docker is available on the system.
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

// isOrbStack detects if Docker is running via OrbStack (macOS).
func isOrbStack() bool {
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
