package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Constants are defined in constants.go

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const envVarHelpText = `
Environment Variables:
  CC_SANDBOX_DEFAULT_IMAGE      Default image tag (default: base)
  CC_SANDBOX_REGISTRY           Registry prefix for images (default: ghcr.io/luwojtaszek)
  CC_SANDBOX_GIT_USER_NAME      Override git user.name in container
  CC_SANDBOX_GIT_USER_EMAIL     Override git user.email in container
  CC_SANDBOX_DOCKER_IMAGES      Additional images that auto-mount Docker socket (comma-separated)
  CC_SANDBOX_DOCKER_SOCKET      Docker socket path (default: auto-detected)
  CC_SANDBOX_ROOT               Run as root: auto, true, or false (default: auto)
  CC_SANDBOX_RUNTIME            Container runtime: auto, docker, or podman (default: auto)
  CC_SANDBOX_DEBUG              Enable debug output (set to 1 to enable)
  CC_SANDBOX_CLAUDE_CONFIG      Claude config directory path (e.g., ~/.claude)
  CC_SANDBOX_CLAUDE_CONFIG_REPO Git repository URL for Claude config
`

type Config struct {
	Image            string
	Registry         string
	DockerSocket     string
	Workdir          string
	Mounts           []string
	EnvVars          []string
	MountDocker      bool
	MountGit         bool
	MountGH          bool
	MountSSH         bool
	Interactive      bool
	Root             *bool  // nil = auto-detect, true = run as root, false = run as claude
	Runtime          string // "auto", "docker", "podman"
	GitUserName      string // Override git user.name
	GitUserEmail     string // Override git user.email
	HostNetwork      bool   // Use host network mode for DinD localhost access
	ClaudeConfigPath string // Host path to mount (e.g., ~/.claude)
	ClaudeConfigRepo string // Git repo URL for config
	ClaudeConfigSync bool   // Pull latest changes from repo
}

// flagsWithValues contains flags that require a separate value argument.
// Used by arg parsing to skip values when finding the first positional argument.
var flagsWithValues = map[string]bool{
	"-i": true, "--image": true,
	"-m": true, "--mount": true,
	"-e": true, "--env": true,
	"-w": true, "--workdir": true,
	"--root": true, "--runtime": true,
	"--git-user-name": true, "--git-user-email": true,
	"-C": true, "--claude-config": true,
	"--claude-config-repo": true,
}

func main() {
	rootCmd := newRootCmd()

	// Workaround for Cobra treating first positional arg as subcommand.
	// Find the first positional argument and insert "--" before it if needed.
	args := os.Args[1:]
	knownCommands := map[string]bool{"version": true, "help": true, "update": true, "completion": true}

	// Find the index of the first positional argument (not a flag)
	firstPosIdx := -1
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Already has separator, no need to modify
			firstPosIdx = -1
			break
		}
		if strings.HasPrefix(arg, "-") {
			// Skip flags: --flag, --flag=value, -f, -f=value
			// If flag uses = syntax, value is included, no skip needed
			if !strings.Contains(arg, "=") && flagsWithValues[arg] {
				// This flag takes a separate value, skip the next arg
				i++
			}
			continue
		}
		// Found a positional argument
		firstPosIdx = i
		break
	}

	if firstPosIdx >= 0 && !knownCommands[args[firstPosIdx]] {
		// Insert "--" before the first positional argument
		newArgs := make([]string, 0, len(args)+1)
		newArgs = append(newArgs, args[:firstPosIdx]...)
		newArgs = append(newArgs, "--")
		newArgs = append(newArgs, args[firstPosIdx:]...)
		rootCmd.SetArgs(newArgs)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cfg := &Config{}
	var rootFlag string

	rootCmd := &cobra.Command{
		Use:   "cc-sandbox [flags] [command] [args...]",
		Short: "Claude Code Sandbox - Run Claude Code in isolated Docker containers",
		Long: `cc-sandbox runs Claude Code in isolated Docker containers with proper
UID/GID mapping, optional credential mounting, and Docker socket access.

Examples:
  cc-sandbox claude                      # Interactive Claude session
  cc-sandbox claude -p "fix the bug"     # One-shot prompt
  cc-sandbox -i docker claude            # Use docker image variant
  cc-sandbox --docker claude             # With Docker socket access
  cc-sandbox -e DEBUG=1 claude -c        # Continue with env var`,
		Version:               fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		RunE: func(_ *cobra.Command, args []string) error {
			cfg.Root = parseRootFlag(rootFlag)
			return runSandbox(cfg, args)
		},
	}

	rootCmd.SetHelpTemplate(rootCmd.HelpTemplate() + envVarHelpText)

	rootCmd.Flags().StringVarP(&cfg.Image, "image", "i", "", "Docker image tag (default: base)")
	rootCmd.Flags().StringArrayVarP(&cfg.Mounts, "mount", "m", nil, "Additional volume mounts (host:container)")
	rootCmd.Flags().StringArrayVarP(&cfg.EnvVars, "env", "e", nil, "Environment variables (KEY=value)")
	rootCmd.Flags().StringVarP(&cfg.Workdir, "workdir", "w", "", "Working directory (default: current directory)")
	rootCmd.Flags().BoolVar(&cfg.MountDocker, "docker", false, "Mount Docker socket")
	rootCmd.Flags().BoolVar(&cfg.MountGit, "git", true, "Mount .gitconfig from host")
	rootCmd.Flags().BoolVar(&cfg.MountGH, "gh", true, "Mount GitHub CLI config from host")
	rootCmd.Flags().BoolVar(&cfg.MountSSH, "ssh", false, "Mount SSH keys from host")
	rootCmd.Flags().BoolVarP(&cfg.Interactive, "interactive", "t", true, "Run in interactive mode with TTY")
	rootCmd.Flags().StringVar(&rootFlag, "root", "auto", "Run as root user: auto, true, or false")
	rootCmd.Flags().StringVar(&cfg.Runtime, "runtime", "auto", "Container runtime: auto, docker, or podman")
	rootCmd.Flags().StringVar(&cfg.GitUserName, "git-user-name", "", "Override git user.name in container")
	rootCmd.Flags().StringVar(&cfg.GitUserEmail, "git-user-email", "", "Override git user.email in container")
	rootCmd.Flags().BoolVar(&cfg.HostNetwork, "host-network", false, "Use host network mode (enables localhost access for DinD port mappings)")
	rootCmd.Flags().StringVarP(&cfg.ClaudeConfigPath, "claude-config", "C", "", "Mount Claude config directory from host")
	rootCmd.Flags().StringVar(&cfg.ClaudeConfigRepo, "claude-config-repo", "", "Git repository URL for Claude config")
	rootCmd.Flags().BoolVar(&cfg.ClaudeConfigSync, "claude-config-sync", false, "Pull latest changes from config repo")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("cc-sandbox %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
			fmt.Printf("  go:     %s\n", runtime.Version())
		},
	})

	rootCmd.AddCommand(newUpdateCmd())

	return rootCmd
}

func runSandbox(cfg *Config, args []string) error {
	cfg.Registry = getEnv("CC_SANDBOX_REGISTRY", DefaultRegistry)
	cfg.DockerSocket = getEnv("CC_SANDBOX_DOCKER_SOCKET", getDefaultDockerSocket())

	// Environment variable overrides CLI flag for root mode
	if envRoot := os.Getenv("CC_SANDBOX_ROOT"); envRoot != "" {
		cfg.Root = parseRootFlag(envRoot)
	}

	// Environment variable overrides CLI flag for runtime
	if envRuntime := os.Getenv("CC_SANDBOX_RUNTIME"); envRuntime != "" {
		cfg.Runtime = envRuntime
	}

	if cfg.Image == "" {
		cfg.Image = getEnv("CC_SANDBOX_DEFAULT_IMAGE", "base")
	}

	// Environment variable overrides for Claude config
	if cfg.ClaudeConfigPath == "" {
		cfg.ClaudeConfigPath = os.Getenv("CC_SANDBOX_CLAUDE_CONFIG")
	}
	if cfg.ClaudeConfigRepo == "" {
		cfg.ClaudeConfigRepo = os.Getenv("CC_SANDBOX_CLAUDE_CONFIG_REPO")
	}

	// Auto-enable Docker socket for docker and bun-full images
	applyDockerAutoMount(cfg)

	if cfg.Workdir == "" {
		var err error
		cfg.Workdir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Detect runtime
	runtime := detectRuntime(cfg)

	imageName := resolveImageName(cfg.Registry, cfg.Image, runtime)

	// Pull image if it's from a registry and not available locally
	if isRegistryImage(imageName) && !imageExistsLocally(imageName, runtime) {
		fmt.Fprintf(os.Stderr, "Pulling image %s...\n", imageName)
		if err := pullImage(imageName, runtime); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
	}

	containerArgs := buildContainerArgs(cfg, runtime, imageName, args)

	containerCmd := exec.Command(runtime, containerArgs...)
	containerCmd.Stdin = os.Stdin
	containerCmd.Stdout = os.Stdout
	containerCmd.Stderr = os.Stderr

	return containerCmd.Run()
}

func buildImageName(registry, image string) string {
	if strings.Contains(image, "/") || strings.Contains(image, ":") && strings.Contains(strings.Split(image, ":")[0], ".") {
		return image
	}

	baseName := "cc-sandbox"
	if !strings.HasPrefix(image, baseName) {
		image = baseName + ":" + image
	}

	if registry != "" {
		return registry + "/" + image
	}

	return image
}

func buildContainerArgs(cfg *Config, containerRuntime, imageName string, containerArgs []string) []string {
	args := []string{"run", "--rm"}

	if cfg.Interactive && isTerminal() {
		args = append(args, "-it")
	}

	// Use host network mode if requested (enables localhost access for DinD port mappings)
	if cfg.HostNetwork {
		args = append(args, "--network=host")
	}

	// Determine if we should run as root based on runtime and config
	runAsRoot := shouldUseRootMode(cfg, containerRuntime)

	// Set container user and permission mode based on runtime and root mode
	if containerRuntime == RuntimePodman {
		// Podman: use --userns=keep-id for UID mapping
		args = append(args, "--userns=keep-id")
		args = append(args, "-u", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()))
		args = append(args, "-e", "CC_SANDBOX_PERMISSION_MODE=skip")
	} else if runAsRoot {
		// Docker rootless: add --userns=host to run as root
		args = append(args, "--userns=host")
		args = append(args, "-u", "0:0")
		// Root in container can't use --dangerously-skip-permissions
		args = append(args, "-e", "CC_SANDBOX_PERMISSION_MODE=accept")
	} else {
		// Regular Docker or OrbStack: use -u flag for proper fixuid support
		args = append(args, "-u", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()))
		args = append(args, "-e", "CC_SANDBOX_PERMISSION_MODE=skip")
	}

	args = append(args, "-v", cfg.Workdir+":/workspace")
	args = append(args, "-w", "/workspace")

	// Mount bare repository if running in a git worktree
	if bareRepoPath := resolveGitWorktreePaths(cfg.Workdir); bareRepoPath != "" {
		args = append(args, "-v", bareRepoPath+":"+bareRepoPath)
	}

	// User-specific credentials volume
	args = append(args, "-v", getCredentialsVolumeName()+":/mnt/claude-data")

	homeDir, _ := os.UserHomeDir()

	if cfg.MountGit {
		gitconfig := filepath.Join(homeDir, ".gitconfig")
		if fileExists(gitconfig) {
			args = append(args, "-v", gitconfig+":/mnt/host-config/.gitconfig:ro")
		}
	}

	// Pass resolved git user config as env vars
	userName, userEmail := resolveGitUserConfig(cfg)
	if userName != "" {
		args = append(args, "-e", "CC_GIT_USER_NAME="+userName)
	}
	if userEmail != "" {
		args = append(args, "-e", "CC_GIT_USER_EMAIL="+userEmail)
	}

	// Pass debug flag to container
	if os.Getenv("CC_SANDBOX_DEBUG") == "1" {
		args = append(args, "-e", "CC_SANDBOX_DEBUG=1")
	}

	if cfg.MountGH {
		ghConfig := filepath.Join(homeDir, ".config", "gh")
		if dirExists(ghConfig) {
			args = append(args, "-v", ghConfig+":/mnt/host-config/gh:ro")
		}
		if token := os.Getenv("GH_TOKEN"); token != "" {
			args = append(args, "-e", "GH_TOKEN")
		}
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			args = append(args, "-e", "GITHUB_TOKEN")
		}
	}

	if cfg.MountSSH {
		sshDir := filepath.Join(homeDir, ".ssh")
		if dirExists(sshDir) {
			args = append(args, "-v", sshDir+":/mnt/host-config/.ssh:ro")
		}
	}

	// Claude config from host path
	if cfg.ClaudeConfigPath != "" {
		configPath := expandPath(cfg.ClaudeConfigPath)
		if dirExists(configPath) {
			args = append(args, "-v", configPath+":/mnt/host-claude-config:ro")
		}
	}

	// Claude config from git repo
	if cfg.ClaudeConfigRepo != "" {
		args = append(args, "-e", "CC_CLAUDE_CONFIG_REPO="+cfg.ClaudeConfigRepo)
		args = append(args, "-v", getClaudeConfigVolumeName()+":/mnt/claude-config-repo")
		if cfg.ClaudeConfigSync {
			args = append(args, "-e", "CC_CLAUDE_CONFIG_SYNC=1")
		}
	}

	if cfg.MountDocker {
		if fileExists(cfg.DockerSocket) {
			args = append(args, "-v", cfg.DockerSocket+":/var/run/docker.sock")
			// Add docker socket's group to allow access without sudo
			// On Linux, the socket typically has a 'docker' group (e.g., GID 999)
			// On macOS with Docker Desktop/Orbstack, the socket appears as root:root (GID 0)
			// inside the container, regardless of host permissions
			if gid := getFileGID(cfg.DockerSocket); gid > 0 {
				args = append(args, "--group-add", strconv.Itoa(gid))
			}
			// Add root group (0) only on macOS for Docker Desktop/Orbstack compatibility
			// where the socket is mounted as root:root inside the Linux VM
			if runtime.GOOS == "darwin" {
				args = append(args, "--group-add", "0")
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Docker socket not found at %s\n", cfg.DockerSocket)
		}
	}

	for _, mount := range cfg.Mounts {
		args = append(args, "-v", mount)
	}

	for _, env := range cfg.EnvVars {
		args = append(args, "-e", env)
	}

	args = append(args, imageName)

	if len(containerArgs) > 0 {
		args = append(args, containerArgs...)
	}

	return args
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// debugLog prints a debug message if CC_SANDBOX_DEBUG=1 is set.
func debugLog(format string, args ...interface{}) {
	if os.Getenv("CC_SANDBOX_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// applyDockerAutoMount auto-enables Docker socket mounting for Docker-enabled images.
// Built-in images: "docker" and "bun-full"
// Additional images can be specified via CC_SANDBOX_DOCKER_IMAGES (comma-separated)
func applyDockerAutoMount(cfg *Config) {
	// Built-in images that auto-mount Docker
	if cfg.Image == "docker" || cfg.Image == "bun-full" {
		cfg.MountDocker = true
		return
	}

	// Check user-specified images from environment variable
	if envImages := os.Getenv("CC_SANDBOX_DOCKER_IMAGES"); envImages != "" {
		for _, img := range strings.Split(envImages, ",") {
			if strings.TrimSpace(img) == cfg.Image {
				cfg.MountDocker = true
				return
			}
		}
	}
}

// parseRootFlag parses a root flag string value.
// Returns nil for "auto", true for "true/yes/1", false for "false/no/0".
func parseRootFlag(flag string) *bool {
	switch strings.ToLower(flag) {
	case "true", "yes", "1":
		v := true
		return &v
	case "false", "no", "0":
		v := false
		return &v
	default:
		return nil
	}
}

// detectRuntime determines which container runtime to use.
func detectRuntime(cfg *Config) string {
	if cfg.Runtime != "" && cfg.Runtime != "auto" {
		return cfg.Runtime
	}
	// Check if podman is available and docker is not
	if isPodmanAvailable() && !isDockerAvailable() {
		return RuntimePodman
	}
	return RuntimeDocker
}

// shouldUseRootMode determines if container should run as root (no privilege drop).
// Returns true if root mode is explicitly enabled or auto-detected for rootless Docker.
func shouldUseRootMode(cfg *Config, runtime string) bool {
	if cfg.Root != nil {
		return *cfg.Root
	}
	// Podman with --userns=keep-id doesn't need root mode
	if runtime == RuntimePodman {
		return false
	}
	// OrbStack doesn't need root mode (acts like regular Docker)
	if isOrbStack() {
		return false
	}
	// Auto-detect: use root mode for rootless Docker only
	return isRootlessDocker()
}

// getCredentialsVolumeName returns a user-specific volume name.
// It also handles migration from the old shared volume name.
func getCredentialsVolumeName() string {
	var newVolume string

	// Unix (Linux/macOS): use UID for uniqueness
	if uid := os.Getuid(); uid >= 0 {
		newVolume = "cc-sandbox-credentials-" + strconv.Itoa(uid)
	} else if u, err := user.Current(); err == nil {
		// Windows: use username (os.Getuid() returns -1)
		// Sanitize username for Docker volume name (alphanumeric, dash, underscore)
		name := sanitizeVolumeName(u.Username)
		newVolume = "cc-sandbox-credentials-" + name
	} else {
		// Fallback (shouldn't happen)
		return "cc-sandbox-credentials"
	}

	// Optimization: Check new volume FIRST (most common case - already migrated)
	// This avoids the old volume check in the common case
	if volumeExists(newVolume) {
		return newVolume
	}

	// Only check old volume if new doesn't exist (migration needed or first run)
	oldVolume := "cc-sandbox-credentials"
	if volumeExists(oldVolume) {
		if err := migrateCredentialsVolume(oldVolume, newVolume); err != nil {
			debugLog("Failed to migrate credentials volume: %v", err)
			// Fall back to old volume if migration fails
			return oldVolume
		}
	}

	return newVolume
}

// volumeExists checks if a Docker volume exists.
func volumeExists(name string) bool {
	cmd := exec.Command("docker", "volume", "inspect", name)
	return cmd.Run() == nil
}

// migrateCredentialsVolume copies data from old volume to new volume.
func migrateCredentialsVolume(oldVolume, newVolume string) error {
	debugLog("[cc-sandbox] Migrating credentials from %s to %s...", oldVolume, newVolume)

	// Create new volume
	createCmd := exec.Command("docker", "volume", "create", newVolume)
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create new volume: %w", err)
	}

	// Copy data using a temporary container
	// Use alpine:latest for minimal overhead
	copyCmd := exec.Command("docker", "run", "--rm",
		"-v", oldVolume+":/src:ro",
		"-v", newVolume+":/dst",
		"alpine:latest",
		"sh", "-c", "cp -a /src/. /dst/")
	if err := copyCmd.Run(); err != nil {
		// Clean up new volume on failure
		_ = exec.Command("docker", "volume", "rm", newVolume).Run()
		return fmt.Errorf("failed to copy data: %w", err)
	}

	debugLog("[cc-sandbox] Credentials migration complete. Old volume (%s) preserved.", oldVolume)
	return nil
}

// sanitizeVolumeName sanitizes a string for use in Docker volume names.
// Docker volume names: [a-zA-Z0-9][a-zA-Z0-9_.-]*.
func sanitizeVolumeName(s string) string {
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}
	if result.Len() == 0 {
		return "default"
	}
	return result.String()
}

// getClaudeConfigVolumeName returns a user-specific volume name for Claude config repo cache.
func getClaudeConfigVolumeName() string {
	if uid := os.Getuid(); uid >= 0 {
		return "cc-sandbox-claude-config-" + strconv.Itoa(uid)
	}
	if u, err := user.Current(); err == nil {
		return "cc-sandbox-claude-config-" + sanitizeVolumeName(u.Username)
	}
	return "cc-sandbox-claude-config"
}

// expandPath expands ~ in paths to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func getDefaultDockerSocket() string {
	sockets := []string{
		"/var/run/docker.sock",
		filepath.Join(os.Getenv("HOME"), ".orbstack/run/docker.sock"),
		filepath.Join(os.Getenv("HOME"), ".docker/run/docker.sock"),
	}

	for _, sock := range sockets {
		if fileExists(sock) {
			return sock
		}
	}

	return "/var/run/docker.sock"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// resolveGitWorktreePaths returns paths that need to be mounted for git worktrees.
// Returns the path to the bare repository if workdir is a worktree, empty string otherwise.
func resolveGitWorktreePaths(workdir string) string {
	gitPath := filepath.Join(workdir, ".git")
	debugLog("Checking git worktree at: %s", gitPath)

	// Check if .git is a file (worktree) rather than a directory (regular repo)
	info, err := os.Stat(gitPath)
	if err != nil {
		debugLog("Git worktree check: .git stat failed: %v", err)
		return ""
	}
	if info.IsDir() {
		debugLog("Git worktree check: .git is a directory (regular repo)")
		return ""
	}

	// Read the gitdir reference from the .git file
	content, err := os.ReadFile(gitPath)
	if err != nil {
		debugLog("Git worktree check: failed to read .git file: %v", err)
		return ""
	}

	// Parse "gitdir: /path/to/repo.git/worktrees/name"
	line := strings.TrimSpace(string(content))
	if !strings.HasPrefix(line, "gitdir: ") {
		debugLog("Git worktree check: .git file does not contain gitdir reference: %s", line)
		return ""
	}

	gitdir := strings.TrimPrefix(line, "gitdir: ")
	debugLog("Git worktree check: gitdir = %s", gitdir)

	// The gitdir points to repo.git/worktrees/name
	// We need to mount the parent bare repo (repo.git)
	// Go up two levels: worktrees/name -> worktrees -> repo.git
	if strings.Contains(gitdir, "/worktrees/") {
		idx := strings.Index(gitdir, "/worktrees/")
		bareRepo := gitdir[:idx]
		debugLog("Git worktree check: found bare repository at %s", bareRepo)
		return bareRepo
	}

	debugLog("Git worktree check: gitdir does not contain /worktrees/ path")
	return ""
}

// getGitUserConfigBatched retrieves user.name and user.email in a single subprocess call.
func getGitUserConfigBatched() (name, email string) {
	cmd := exec.Command("git", "config", "--global", "--get-regexp", "^user\\.(name|email)$")
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	// Parse output: "user.name John Doe\nuser.email john@example.com\n"
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "user.name ") {
			name = strings.TrimPrefix(line, "user.name ")
		} else if strings.HasPrefix(line, "user.email ") {
			email = strings.TrimPrefix(line, "user.email ")
		}
	}
	return name, email
}

// resolveGitUserConfig determines git user.name and user.email
// Priority: CLI flags > Env vars > Auto-detected from host
func resolveGitUserConfig(cfg *Config) (string, string) {
	userName, userEmail := getGitUserConfigBatched()

	if envName := os.Getenv("CC_SANDBOX_GIT_USER_NAME"); envName != "" {
		userName = envName
	}
	if envEmail := os.Getenv("CC_SANDBOX_GIT_USER_EMAIL"); envEmail != "" {
		userEmail = envEmail
	}

	if cfg.GitUserName != "" {
		userName = cfg.GitUserName
	}
	if cfg.GitUserEmail != "" {
		userEmail = cfg.GitUserEmail
	}

	return userName, userEmail
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func isRegistryImage(imageName string) bool {
	// Image is from registry if it contains a domain (has dots before first slash)
	if idx := strings.Index(imageName, "/"); idx > 0 {
		domain := imageName[:idx]
		return strings.Contains(domain, ".")
	}
	return false
}

// imageExistsLocally checks if a Docker image exists locally.
// This is a variable to allow mocking in tests.
var imageExistsLocally = func(imageName, runtime string) bool {
	cmd := exec.Command(runtime, "image", "inspect", imageName)
	return cmd.Run() == nil
}

// resolveImageName determines the final image name, preferring local images over registry.
// For simple tags (e.g., "golang-full"), checks if cc-sandbox:<tag> exists locally first.
func resolveImageName(registry, image, runtime string) string {
	// If image already has a path or domain, use as-is
	if strings.Contains(image, "/") || strings.Contains(image, ":") && strings.Contains(strings.Split(image, ":")[0], ".") {
		return image
	}

	// Build local image name (cc-sandbox:<tag>)
	localImage := "cc-sandbox:" + image
	if strings.HasPrefix(image, "cc-sandbox") {
		localImage = image
	}

	// Check if local image exists - if so, prefer it over registry
	if imageExistsLocally(localImage, runtime) {
		return localImage
	}

	// Fall back to registry image
	if registry != "" {
		return registry + "/" + localImage
	}

	return localImage
}

func pullImage(imageName, runtime string) error {
	cmd := exec.Command(runtime, "pull", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
