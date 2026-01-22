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
