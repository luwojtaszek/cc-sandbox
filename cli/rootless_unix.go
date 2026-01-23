//go:build unix

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// dockerInfo holds cached docker info results from a single call
type dockerInfo struct {
	securityOptions string
	operatingSystem string
	available       bool
}

// Cached detection results
var (
	// Combined docker info cache (single subprocess call)
	dockerInfoOnce  sync.Once
	dockerInfoCache dockerInfo

	// Runtime availability (parallel detection)
	runtimeDetectionOnce  sync.Once
	podmanAvailableResult bool
	dockerAvailableResult bool
)

// getDockerInfo retrieves docker info fields in a single subprocess call.
// Results are cached for the lifetime of the process.
func getDockerInfo() dockerInfo {
	dockerInfoOnce.Do(func() {
		cmd := exec.Command("docker", "info", "--format",
			"{{.SecurityOptions}}|||{{.OperatingSystem}}")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(output)), "|||")
			if len(parts) == 2 {
				dockerInfoCache = dockerInfo{
					securityOptions: parts[0],
					operatingSystem: parts[1],
					available:       true,
				}
			}
		}
	})
	return dockerInfoCache
}

// detectRuntimeAvailability checks docker and podman availability in parallel.
func detectRuntimeAvailability() {
	runtimeDetectionOnce.Do(func() {
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			cmd := exec.Command("docker", "--version")
			dockerAvailableResult = cmd.Run() == nil
		}()

		go func() {
			defer wg.Done()
			cmd := exec.Command("podman", "--version")
			podmanAvailableResult = cmd.Run() == nil
		}()

		wg.Wait()
	})
}

// isRootlessDocker detects if Docker is running in rootless mode.
// Results are cached for the lifetime of the process.
func isRootlessDocker() bool {
	// Method 1: Check docker info output for rootless security option
	info := getDockerInfo()
	if info.available && strings.Contains(info.securityOptions, "rootless") {
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
	detectRuntimeAvailability()
	return podmanAvailableResult
}

// isDockerAvailable checks if docker is available on the system.
// Results are cached for the lifetime of the process.
func isDockerAvailable() bool {
	detectRuntimeAvailability()
	return dockerAvailableResult
}

// isOrbStack detects if Docker is running via OrbStack (macOS).
// Results are cached for the lifetime of the process.
func isOrbStack() bool {
	// Check if socket path contains orbstack
	socket := getDefaultDockerSocket()
	if strings.Contains(socket, "orbstack") {
		return true
	}
	// Use cached docker info
	info := getDockerInfo()
	return info.available && strings.Contains(strings.ToLower(info.operatingSystem), "orbstack")
}
