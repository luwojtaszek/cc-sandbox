# Custom Images Guide

This guide explains how to build custom cc-sandbox images and configure them for use.

## Overview

You can extend cc-sandbox images to include your own tools, runtimes, or configurations. Custom images can:

- Add project-specific tooling (Go, Rust, Java, etc.)
- Pre-install dependencies
- Include Docker-in-Docker support
- Set custom defaults

## Step 1: Build Your Custom Image

### Basic Custom Image

Start from any cc-sandbox base image:

```dockerfile
# Dockerfile
FROM ghcr.io/luwojtaszek/cc-sandbox:base

# Add your tools
RUN apt-get update && apt-get install -y golang-go && rm -rf /var/lib/apt/lists/*

# Keep the entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
```

### Custom Image with Docker Support

To enable Docker-in-Docker, start from `cc-sandbox:docker` or `cc-sandbox:bun-full`:

```dockerfile
# Dockerfile
FROM ghcr.io/luwojtaszek/cc-sandbox:docker

# Add Go toolchain
RUN apt-get update && apt-get install -y golang-go && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
```

Build your image:

```bash
docker build -t cc-sandbox:golang-full .
```

## Step 2: Configure cc-sandbox to Use Your Image

### Option A: CLI Flag (Local Images)

If you have a local image named `cc-sandbox:<tag>`, just use the tag:

```bash
# Build your custom image
docker build -t cc-sandbox:golang-full .

# Use it directly - local images are automatically detected
cc-sandbox -i golang-full claude
```

The CLI automatically checks for local images first. If `cc-sandbox:golang-full` exists locally, it's used. Otherwise, the CLI falls back to the registry.

### Option B: Environment Variable (Recommended for Defaults)

Set `CC_SANDBOX_DEFAULT_IMAGE` to use your image by default:

```bash
export CC_SANDBOX_DEFAULT_IMAGE=golang-full
cc-sandbox claude  # Uses golang-full automatically
```

### Option C: Custom Registry

If you publish images to your own registry:

```bash
export CC_SANDBOX_REGISTRY=ghcr.io/myorg
cc-sandbox -i golang-full claude
# Pulls: ghcr.io/myorg/cc-sandbox:golang-full (if not found locally)
```

Note: Local images are always checked first. Setting `CC_SANDBOX_REGISTRY=""` is only needed if you want to prevent any registry fallback.

## Step 3: Enable Docker Socket Auto-Mount

By default, only `docker` and `bun-full` images auto-mount the Docker socket. To enable this for your custom images:

### Option A: Use --docker Flag

```bash
cc-sandbox -i golang-full --docker claude
```

### Option B: Environment Variable (Recommended)

Add your image to `CC_SANDBOX_DOCKER_IMAGES`:

```bash
export CC_SANDBOX_DOCKER_IMAGES="golang-full,rust-docker"
cc-sandbox -i golang-full claude  # Docker socket auto-mounted
```

Multiple images can be specified (comma-separated).

## Complete Setup Example

Add to your shell profile (`~/.bashrc`, `~/.zshrc`):

```bash
# Default to your custom image
export CC_SANDBOX_DEFAULT_IMAGE=golang-full

# Auto-mount Docker for custom images
export CC_SANDBOX_DOCKER_IMAGES="golang-full,rust-docker"

# Git identity (optional)
export CC_SANDBOX_GIT_USER_NAME="Your Name"
export CC_SANDBOX_GIT_USER_EMAIL="your@email.com"

# Optional: Override registry fallback (local images are checked first automatically)
# export CC_SANDBOX_REGISTRY=""
```

Now simply run:

```bash
cc-sandbox claude  # Uses golang-full with Docker support (local image preferred)
```

## Reference: All Configuration Options

| Variable                    | Description                       | Default               |
|-----------------------------|-----------------------------------|-----------------------|
| `CC_SANDBOX_DEFAULT_IMAGE`  | Default image tag                 | `base`                |
| `CC_SANDBOX_REGISTRY`       | Registry prefix (empty for local) | `ghcr.io/luwojtaszek` |
| `CC_SANDBOX_DOCKER_IMAGES`  | Images that auto-mount Docker     | none                  |
| `CC_SANDBOX_GIT_USER_NAME`  | Git user.name override            | auto-detected         |
| `CC_SANDBOX_GIT_USER_EMAIL` | Git user.email override           | auto-detected         |

See [CLI Reference](cli-reference.md) for all options.
