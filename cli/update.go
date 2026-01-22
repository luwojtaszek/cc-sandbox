package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// Constants are defined in constants.go

type UpdateConfig struct {
	SkipCLI    bool
	SkipImages bool
	Force      bool
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func newUpdateCmd() *cobra.Command {
	cfg := &UpdateConfig{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update cc-sandbox CLI and Docker images",
		Long: `Update cc-sandbox to the latest version.

This command will:
  1. Check for and install the latest CLI version
  2. Update any locally installed cc-sandbox Docker images

Only images that are already pulled locally will be updated.
Images you haven't used won't be downloaded.

Examples:
  cc-sandbox update              # Update CLI and images
  cc-sandbox update --skip-cli   # Update only Docker images
  cc-sandbox update --skip-images # Update only CLI`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUpdate(cfg)
		},
	}

	cmd.Flags().BoolVar(&cfg.SkipCLI, "skip-cli", false, "Skip CLI update")
	cmd.Flags().BoolVar(&cfg.SkipImages, "skip-images", false, "Skip Docker images update")
	cmd.Flags().BoolVarP(&cfg.Force, "force", "f", false, "Force update even if already on latest version")

	return cmd
}

func runUpdate(cfg *UpdateConfig) error {
	fmt.Println()
	fmt.Println("\033[32m╔══════════════════════════════════════╗\033[0m")
	fmt.Println("\033[32m║     cc-sandbox updater               ║\033[0m")
	fmt.Println("\033[32m╚══════════════════════════════════════╝\033[0m")
	fmt.Println()

	var cliUpdated, imagesUpdated bool
	var err error

	if !cfg.SkipCLI {
		cliUpdated, err = updateCLI(cfg.Force)
		if err != nil {
			fmt.Printf("\033[31m[ERROR]\033[0m Failed to update CLI: %v\n", err)
		}
	} else {
		fmt.Println("\033[33m[SKIP]\033[0m Skipping CLI update")
	}

	fmt.Println()

	if !cfg.SkipImages {
		imagesUpdated, err = updateImages()
		if err != nil {
			fmt.Printf("\033[31m[ERROR]\033[0m Failed to update images: %v\n", err)
		}
	} else {
		fmt.Println("\033[33m[SKIP]\033[0m Skipping Docker images update")
	}

	fmt.Println()
	if cliUpdated || imagesUpdated {
		fmt.Println("\033[32m[OK]\033[0m Update complete!")
	} else {
		fmt.Println("\033[34m[INFO]\033[0m Everything is up to date.")
	}

	return nil
}

func updateCLI(force bool) (bool, error) {
	fmt.Println("\033[34m[INFO]\033[0m Checking for CLI updates...")

	latestVersion, err := getLatestVersion()
	if err != nil {
		return false, fmt.Errorf("failed to fetch latest version: %w", err)
	}

	currentVersion := version
	if currentVersion == "dev" {
		fmt.Println("\033[33m[WARN]\033[0m Running development version, skipping CLI update")
		return false, nil
	}

	// Compare versions using semver
	// Ensure both versions have 'v' prefix for semver.Compare
	latestSemver := latestVersion
	if !strings.HasPrefix(latestSemver, "v") {
		latestSemver = "v" + latestSemver
	}
	currentSemver := currentVersion
	if !strings.HasPrefix(currentSemver, "v") {
		currentSemver = "v" + currentSemver
	}

	// Use semver comparison if both are valid semver, otherwise fall back to string comparison
	var needsUpdate bool
	if semver.IsValid(latestSemver) && semver.IsValid(currentSemver) {
		// semver.Compare returns: -1 if current < latest, 0 if equal, 1 if current > latest
		needsUpdate = semver.Compare(currentSemver, latestSemver) < 0
	} else {
		// Fallback to string comparison for non-semver versions
		latestNorm := strings.TrimPrefix(latestVersion, "v")
		currentNorm := strings.TrimPrefix(currentVersion, "v")
		needsUpdate = latestNorm != currentNorm
	}

	if !force && !needsUpdate {
		fmt.Printf("\033[32m[OK]\033[0m CLI is already at latest version (%s)\n", currentVersion)
		return false, nil
	}

	fmt.Printf("\033[34m[INFO]\033[0m Current version: %s\n", currentVersion)
	fmt.Printf("\033[34m[INFO]\033[0m Latest version:  %s\n", latestVersion)
	fmt.Println("\033[34m[INFO]\033[0m Downloading update...")

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Download new binary
	osName := runtime.GOOS
	arch := runtime.GOARCH
	suffix := ""
	if osName == "windows" {
		suffix = ".exe"
	}

	binaryFilename := fmt.Sprintf("cc-sandbox-%s-%s%s", osName, arch, suffix)
	binaryURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
		GitHubRepo, latestVersion, binaryFilename)
	checksumURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/checksums.txt",
		GitHubRepo, latestVersion)

	tempFile, err := downloadAndVerify(binaryURL, checksumURL, binaryFilename)
	if err != nil {
		return false, fmt.Errorf("failed to download update: %w", err)
	}
	defer func() { _ = os.Remove(tempFile) }()

	// Replace executable
	err = replaceBinary(execPath, tempFile)
	if err != nil {
		return false, fmt.Errorf("failed to install update: %w", err)
	}

	fmt.Printf("\033[32m[OK]\033[0m CLI updated to %s\n", latestVersion)
	return true, nil
}

func updateImages() (bool, error) {
	fmt.Println("\033[34m[INFO]\033[0m Checking for Docker image updates...")

	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Println("\033[33m[WARN]\033[0m Docker not found, skipping image updates")
		return false, nil
	}

	// Get list of locally installed cc-sandbox images
	localImages, err := getLocalImages()
	if err != nil {
		return false, fmt.Errorf("failed to list local images: %w", err)
	}

	if len(localImages) == 0 {
		fmt.Println("\033[34m[INFO]\033[0m No cc-sandbox images found locally")
		return false, nil
	}

	fmt.Printf("\033[34m[INFO]\033[0m Found %d local cc-sandbox image(s)\n", len(localImages))

	var updated bool
	for _, img := range localImages {
		fmt.Printf("\033[34m[INFO]\033[0m Updating %s...\n", img)

		cmd := exec.Command("docker", "pull", img)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("\033[31m[ERROR]\033[0m Failed to update %s: %v\n", img, err)
			continue
		}

		fmt.Printf("\033[32m[OK]\033[0m Updated %s\n", img)
		updated = true
	}

	return updated, nil
}

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func getLocalImages() ([]string, error) {
	// List all local images matching cc-sandbox pattern
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var images []string
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		image := strings.TrimSpace(scanner.Text())
		if image == "" || image == "<none>:<none>" {
			continue
		}

		// Check if this is a cc-sandbox image
		if isRelevantImage(image) {
			// Convert local image name to registry image for pulling
			registryImage := toRegistryImage(image)
			if registryImage != "" && !seen[registryImage] {
				images = append(images, registryImage)
				seen[registryImage] = true
			}
		}
	}

	return images, scanner.Err()
}

func isRelevantImage(image string) bool {
	// Match patterns:
	// - cc-sandbox:base, cc-sandbox:docker, cc-sandbox:bun-full (local)
	// - ghcr.io/luwojtaszek/cc-sandbox:base (registry)

	for _, tag := range KnownImageTags {
		if image == "cc-sandbox:"+tag {
			return true
		}
		if image == DefaultRegistry+"/cc-sandbox:"+tag {
			return true
		}
	}

	return false
}

func toRegistryImage(image string) string {
	// Convert local image name to registry image.
	// Example: cc-sandbox:base -> ghcr.io/luwojtaszek/cc-sandbox:base

	if strings.HasPrefix(image, DefaultRegistry+"/") {
		return image
	}

	for _, tag := range KnownImageTags {
		if image == "cc-sandbox:"+tag {
			return DefaultRegistry + "/cc-sandbox:" + tag
		}
	}

	return ""
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "cc-sandbox-update-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = tempFile.Close() }()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// parseChecksumFile parses a checksums.txt file and returns the checksum for the target filename.
// Format: <sha256hash>  <filename> (two spaces between hash and filename)
func parseChecksumFile(checksumPath, targetFilename string) (string, error) {
	file, err := os.Open(checksumPath)
	if err != nil {
		return "", fmt.Errorf("failed to open checksum file: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: <hash>  <filename> (two spaces) or <hash> <filename> (one space)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			filename := parts[len(parts)-1] // Take last part as filename
			if filename == targetFilename {
				return strings.ToLower(hash), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading checksum file: %w", err)
	}

	return "", fmt.Errorf("checksum not found for %s", targetFilename)
}

// calculateSHA256 computes the SHA256 hash of a file.
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// downloadAndVerify downloads a binary and verifies its SHA256 checksum.
// Returns the path to the verified binary file.
func downloadAndVerify(binaryURL, checksumURL, filename string) (string, error) {
	// Download the checksum file
	checksumFile, err := downloadFile(checksumURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksum file: %w", err)
	}
	defer func() { _ = os.Remove(checksumFile) }()

	// Parse expected checksum
	expectedHash, err := parseChecksumFile(checksumFile, filename)
	if err != nil {
		return "", fmt.Errorf("failed to parse checksum: %w", err)
	}

	// Download the binary
	binaryFile, err := downloadFile(binaryURL)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	// Calculate actual checksum
	actualHash, err := calculateSHA256(binaryFile)
	if err != nil {
		_ = os.Remove(binaryFile)
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Verify checksum
	if actualHash != expectedHash {
		_ = os.Remove(binaryFile)
		return "", fmt.Errorf("checksum verification failed: expected %s, got %s", expectedHash, actualHash)
	}

	fmt.Println("\033[32m[OK]\033[0m Checksum verified")
	return binaryFile, nil
}
