package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Config struct {
	Image        string
	Registry     string
	DockerSocket string
	Workdir      string
	Mounts       []string
	EnvVars      []string
	MountDocker  bool
	MountGit     bool
	MountGH      bool
	MountSSH     bool
	Interactive  bool
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cfg := &Config{}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSandbox(cfg, args)
		},
	}

	rootCmd.Flags().StringVarP(&cfg.Image, "image", "i", "", "Docker image tag (default: base)")
	rootCmd.Flags().StringArrayVarP(&cfg.Mounts, "mount", "m", nil, "Additional volume mounts (host:container)")
	rootCmd.Flags().StringArrayVarP(&cfg.EnvVars, "env", "e", nil, "Environment variables (KEY=value)")
	rootCmd.Flags().StringVarP(&cfg.Workdir, "workdir", "w", "", "Working directory (default: current directory)")
	rootCmd.Flags().BoolVar(&cfg.MountDocker, "docker", false, "Mount Docker socket")
	rootCmd.Flags().BoolVar(&cfg.MountGit, "git", true, "Mount .gitconfig from host")
	rootCmd.Flags().BoolVar(&cfg.MountGH, "gh", true, "Mount GitHub CLI config from host")
	rootCmd.Flags().BoolVar(&cfg.MountSSH, "ssh", false, "Mount SSH keys from host")
	rootCmd.Flags().BoolVarP(&cfg.Interactive, "interactive", "t", true, "Run in interactive mode with TTY")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cc-sandbox %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
			fmt.Printf("  go:     %s\n", runtime.Version())
		},
	})

	return rootCmd
}

func runSandbox(cfg *Config, args []string) error {
	cfg.Registry = getEnv("CC_SANDBOX_REGISTRY", "")
	cfg.DockerSocket = getEnv("CC_SANDBOX_DOCKER_SOCKET", getDefaultDockerSocket())

	if cfg.Image == "" {
		cfg.Image = getEnv("CC_SANDBOX_DEFAULT_IMAGE", "base")
	}

	if cfg.Workdir == "" {
		var err error
		cfg.Workdir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	imageName := buildImageName(cfg.Registry, cfg.Image)
	dockerArgs := buildDockerArgs(cfg, imageName, args)

	dockerCmd := exec.Command("docker", dockerArgs...)
	dockerCmd.Stdin = os.Stdin
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	return dockerCmd.Run()
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

func buildDockerArgs(cfg *Config, imageName string, containerArgs []string) []string {
	args := []string{"run", "--rm"}

	if cfg.Interactive && isTerminal() {
		args = append(args, "-it")
	}

	args = append(args, "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()))
	args = append(args, "-v", cfg.Workdir+":/workspace")
	args = append(args, "-w", "/workspace")
	args = append(args, "-v", "cc-sandbox-credentials:/mnt/claude-data")

	homeDir, _ := os.UserHomeDir()

	if cfg.MountGit {
		gitconfig := filepath.Join(homeDir, ".gitconfig")
		if fileExists(gitconfig) {
			args = append(args, "-v", gitconfig+":/mnt/host-config/.gitconfig:ro")
		}
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

	if cfg.MountDocker {
		if fileExists(cfg.DockerSocket) {
			args = append(args, "-v", cfg.DockerSocket+":/var/run/docker.sock")
			// Add docker socket's group to allow access without sudo
			if gid := getFileGID(cfg.DockerSocket); gid > 0 {
				args = append(args, "--group-add", fmt.Sprintf("%d", gid))
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

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func getFileGID(path string) int {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(stat.Gid)
	}
	return 0
}
