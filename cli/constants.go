package main

// Shared constants for the cc-sandbox CLI

const (
	// DefaultRegistry is the container registry for cc-sandbox images
	DefaultRegistry = "ghcr.io/luwojtaszek"

	// GitHubRepo is the GitHub repository for cc-sandbox
	GitHubRepo = "luwojtaszek/cc-sandbox"

	// RuntimePodman is the podman container runtime
	RuntimePodman = "podman"

	// RuntimeDocker is the docker container runtime
	RuntimeDocker = "docker"
)

// KnownImageTags lists all known cc-sandbox image variants
var KnownImageTags = []string{"base", "docker", "bun-full"}
