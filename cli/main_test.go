package main

import (
	"os"
	"testing"
)

func TestBuildImageName(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		image    string
		want     string
	}{
		{"simple tag", "ghcr.io/user", "base", "ghcr.io/user/cc-sandbox:base"},
		{"empty registry", "", "base", "cc-sandbox:base"},
		{"full image path", "ghcr.io/user", "other/image:tag", "other/image:tag"},
		{"already prefixed", "ghcr.io/user", "cc-sandbox:docker", "ghcr.io/user/cc-sandbox:docker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildImageName(tt.registry, tt.image)
			if got != tt.want {
				t.Errorf("buildImageName(%q, %q) = %q, want %q", tt.registry, tt.image, got, tt.want)
			}
		})
	}
}

func TestParseRootFlag(t *testing.T) {
	tests := []struct {
		input string
		want  *bool
	}{
		{"true", boolPtr(true)},
		{"True", boolPtr(true)},
		{"yes", boolPtr(true)},
		{"1", boolPtr(true)},
		{"false", boolPtr(false)},
		{"False", boolPtr(false)},
		{"no", boolPtr(false)},
		{"0", boolPtr(false)},
		{"auto", nil},
		{"", nil},
		{"invalid", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseRootFlag(tt.input)
			if !boolPtrEqual(got, tt.want) {
				t.Errorf("parseRootFlag(%q) = %v, want %v", tt.input, ptrStr(got), ptrStr(tt.want))
			}
		})
	}
}

func TestSanitizeVolumeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"user-name", "user-name"},
		{"user_name", "user_name"},
		{"user@domain", "userdomain"},
		{"DOMAIN\\user", "DOMAINuser"},
		{"user name", "username"},
		{"", "default"},
		{"@#$%", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeVolumeName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeVolumeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsRegistryImage(t *testing.T) {
	tests := []struct {
		image string
		want  bool
	}{
		{"ghcr.io/user/image:tag", true},
		{"docker.io/library/alpine", true},
		{"cc-sandbox:base", false},
		{"myimage:latest", false},
		{"localhost:5000/image", false}, // localhost has no dots
	}

	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			got := isRegistryImage(tt.image)
			if got != tt.want {
				t.Errorf("isRegistryImage(%q) = %v, want %v", tt.image, got, tt.want)
			}
		})
	}
}

func TestResolveImageName(t *testing.T) {
	tests := []struct {
		name        string
		registry    string
		image       string
		localExists bool
		want        string
	}{
		// Local image exists - should prefer local
		{"local golang-full exists", "ghcr.io/user", "golang-full", true, "cc-sandbox:golang-full"},
		{"local rust-dev exists", "ghcr.io/user", "rust-dev", true, "cc-sandbox:rust-dev"},

		// Local image doesn't exist - should use registry
		{"no local, use registry", "ghcr.io/user", "golang-full", false, "ghcr.io/user/cc-sandbox:golang-full"},
		{"no local, empty registry", "", "golang-full", false, "cc-sandbox:golang-full"},

		// Full paths should pass through unchanged
		{"full path unchanged", "ghcr.io/user", "myregistry.io/custom:tag", false, "myregistry.io/custom:tag"},
		{"path with slash unchanged", "ghcr.io/user", "user/image:tag", false, "user/image:tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock imageExistsLocally
			originalFn := imageExistsLocally
			imageExistsLocally = func(_, _ string) bool { return tt.localExists }
			defer func() { imageExistsLocally = originalFn }()

			got := resolveImageName(tt.registry, tt.image, "docker")
			if got != tt.want {
				t.Errorf("resolveImageName(%q, %q) = %q, want %q", tt.registry, tt.image, got, tt.want)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	// Set test env var
	_ = os.Setenv("TEST_VAR_EXISTS", "value")
	defer func() { _ = os.Unsetenv("TEST_VAR_EXISTS") }()

	tests := []struct {
		key    string
		defVal string
		want   string
	}{
		{"TEST_VAR_EXISTS", "default", "value"},
		{"TEST_VAR_NOT_EXISTS", "default", "default"},
		{"TEST_VAR_NOT_EXISTS", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := getEnv(tt.key, tt.defVal)
			if got != tt.want {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defVal, got, tt.want)
			}
		})
	}
}

func TestApplyDockerAutoMount(t *testing.T) {
	tests := []struct {
		name        string
		image       string
		envValue    string
		mountDocker bool
		wantDocker  bool
	}{
		// Built-in behavior
		{"docker image auto-enables", "docker", "", false, true},
		{"bun-full image auto-enables", "bun-full", "", false, true},
		{"base image does not auto-enable", "base", "", false, false},
		{"explicit --docker on base works", "base", "", true, true},
		// Env var behavior
		{"custom image via env var", "golang-full", "golang-full", false, true},
		{"multiple custom images", "rust-docker", "golang-full,rust-docker", false, true},
		{"image not in env var list", "python", "golang-full,rust-docker", false, false},
		{"built-in still works with env var set", "docker", "golang-full", false, true},
		{"whitespace handling", "golang-full", " golang-full , rust-docker ", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				_ = os.Setenv("CC_SANDBOX_DOCKER_IMAGES", tt.envValue)
				defer func() { _ = os.Unsetenv("CC_SANDBOX_DOCKER_IMAGES") }()
			} else {
				_ = os.Unsetenv("CC_SANDBOX_DOCKER_IMAGES")
			}

			cfg := &Config{
				Image:       tt.image,
				MountDocker: tt.mountDocker,
			}
			applyDockerAutoMount(cfg)
			if cfg.MountDocker != tt.wantDocker {
				t.Errorf("applyDockerAutoMount() MountDocker = %v, want %v", cfg.MountDocker, tt.wantDocker)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tilde expansion", "~/.claude", home + "/.claude"},
		{"tilde with subpath", "~/foo/bar", home + "/foo/bar"},
		{"absolute path unchanged", "/absolute/path", "/absolute/path"},
		{"relative path unchanged", "relative/path", "relative/path"},
		{"tilde only", "~", "~"}, // Only expands ~/ not ~
		{"no tilde", "no-tilde", "no-tilde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetClaudeConfigVolumeName(t *testing.T) {
	// Test that it returns a valid volume name
	name := getClaudeConfigVolumeName()

	// Should start with the expected prefix
	if !contains(name, "cc-sandbox-claude-config") {
		t.Errorf("getClaudeConfigVolumeName() = %q, want prefix 'cc-sandbox-claude-config'", name)
	}

	// Should be a valid Docker volume name (alphanumeric, dash, underscore)
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			t.Errorf("getClaudeConfigVolumeName() = %q contains invalid character %q", name, r)
		}
	}
}

func TestBuildContainerArgsWithClaudeConfig(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name             string
		claudeConfigPath string
		claudeConfigRepo string
		claudeConfigSync bool
		wantContains     []string
		wantNotContains  []string
	}{
		{
			name:             "host config mount",
			claudeConfigPath: tmpDir,
			wantContains:     []string{"-v", tmpDir + ":/mnt/host-claude-config:ro"},
		},
		{
			name:             "git repo config",
			claudeConfigRepo: "https://github.com/user/config",
			wantContains:     []string{"-e", "CC_CLAUDE_CONFIG_REPO=https://github.com/user/config", "-v", "cc-sandbox-claude-config"},
			wantNotContains:  []string{"CC_CLAUDE_CONFIG_SYNC"},
		},
		{
			name:             "git repo config with sync",
			claudeConfigRepo: "https://github.com/user/config",
			claudeConfigSync: true,
			wantContains:     []string{"CC_CLAUDE_CONFIG_REPO=https://github.com/user/config", "CC_CLAUDE_CONFIG_SYNC=1"},
		},
		{
			name:             "both host and git repo",
			claudeConfigPath: tmpDir,
			claudeConfigRepo: "https://github.com/user/config",
			wantContains:     []string{tmpDir + ":/mnt/host-claude-config:ro", "CC_CLAUDE_CONFIG_REPO"},
		},
		{
			name:             "no config",
			claudeConfigPath: "",
			claudeConfigRepo: "",
			wantNotContains:  []string{"/mnt/host-claude-config", "CC_CLAUDE_CONFIG_REPO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Workdir:          tmpDir,
				ClaudeConfigPath: tt.claudeConfigPath,
				ClaudeConfigRepo: tt.claudeConfigRepo,
				ClaudeConfigSync: tt.claudeConfigSync,
			}

			args := buildContainerArgs(cfg, "docker", "test-image", nil)
			argsStr := joinArgs(args)

			for _, want := range tt.wantContains {
				if !contains(argsStr, want) {
					t.Errorf("buildContainerArgs() missing expected %q in args: %v", want, args)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if contains(argsStr, notWant) {
					t.Errorf("buildContainerArgs() should not contain %q in args: %v", notWant, args)
				}
			}
		})
	}
}

func TestClaudeConfigEnvVarOverrides(t *testing.T) {
	tests := []struct {
		name           string
		envConfig      string
		envConfigRepo  string
		flagConfig     string
		flagConfigRepo string
		wantConfig     string
		wantConfigRepo string
	}{
		{
			name:       "flag overrides env",
			envConfig:  "/env/path",
			flagConfig: "/flag/path",
			wantConfig: "/flag/path",
		},
		{
			name:       "env used when no flag",
			envConfig:  "/env/path",
			flagConfig: "",
			wantConfig: "/env/path",
		},
		{
			name:           "repo flag overrides env",
			envConfigRepo:  "https://github.com/env/repo",
			flagConfigRepo: "https://github.com/flag/repo",
			wantConfigRepo: "https://github.com/flag/repo",
		},
		{
			name:           "repo env used when no flag",
			envConfigRepo:  "https://github.com/env/repo",
			flagConfigRepo: "",
			wantConfigRepo: "https://github.com/env/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			if tt.envConfig != "" {
				_ = os.Setenv("CC_SANDBOX_CLAUDE_CONFIG", tt.envConfig)
				defer func() { _ = os.Unsetenv("CC_SANDBOX_CLAUDE_CONFIG") }()
			} else {
				_ = os.Unsetenv("CC_SANDBOX_CLAUDE_CONFIG")
			}

			if tt.envConfigRepo != "" {
				_ = os.Setenv("CC_SANDBOX_CLAUDE_CONFIG_REPO", tt.envConfigRepo)
				defer func() { _ = os.Unsetenv("CC_SANDBOX_CLAUDE_CONFIG_REPO") }()
			} else {
				_ = os.Unsetenv("CC_SANDBOX_CLAUDE_CONFIG_REPO")
			}

			cfg := &Config{
				ClaudeConfigPath: tt.flagConfig,
				ClaudeConfigRepo: tt.flagConfigRepo,
			}

			// Apply env var overrides (simulating what runSandbox does)
			if cfg.ClaudeConfigPath == "" {
				cfg.ClaudeConfigPath = os.Getenv("CC_SANDBOX_CLAUDE_CONFIG")
			}
			if cfg.ClaudeConfigRepo == "" {
				cfg.ClaudeConfigRepo = os.Getenv("CC_SANDBOX_CLAUDE_CONFIG_REPO")
			}

			if tt.wantConfig != "" && cfg.ClaudeConfigPath != tt.wantConfig {
				t.Errorf("ClaudeConfigPath = %q, want %q", cfg.ClaudeConfigPath, tt.wantConfig)
			}
			if tt.wantConfigRepo != "" && cfg.ClaudeConfigRepo != tt.wantConfigRepo {
				t.Errorf("ClaudeConfigRepo = %q, want %q", cfg.ClaudeConfigRepo, tt.wantConfigRepo)
			}
		})
	}
}

// Helper functions
func boolPtr(b bool) *bool { return &b }

func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrStr(p *bool) string {
	if p == nil {
		return "nil"
	}
	if *p {
		return "true"
	}
	return "false"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func joinArgs(args []string) string {
	result := ""
	for _, arg := range args {
		result += arg + " "
	}
	return result
}
