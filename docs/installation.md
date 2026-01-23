# Installation

## One-Line Installer (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/luwojtaszek/cc-sandbox/main/install.sh | bash
```

This downloads the latest binary for your platform and installs it to `~/.local/bin`.

### Installer Options

```bash
# Install specific version
CC_SANDBOX_VERSION=v1.3.0 curl -fsSL https://raw.githubusercontent.com/luwojtaszek/cc-sandbox/main/install.sh | bash

# Install to custom directory
CC_SANDBOX_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/luwojtaszek/cc-sandbox/main/install.sh | bash
```

## Go Install

If you have Go installed:

```bash
go install github.com/luwojtaszek/cc-sandbox/cli@latest
```

The binary will be installed to `$GOPATH/bin` (usually `~/go/bin`).

## Manual Download

Download the binary for your platform from the [releases page](https://github.com/luwojtaszek/cc-sandbox/releases):

```bash
# Linux (x86_64)
curl -LO https://github.com/luwojtaszek/cc-sandbox/releases/latest/download/cc-sandbox-linux-amd64

# Linux (ARM64)
curl -LO https://github.com/luwojtaszek/cc-sandbox/releases/latest/download/cc-sandbox-linux-arm64

# macOS (Intel)
curl -LO https://github.com/luwojtaszek/cc-sandbox/releases/latest/download/cc-sandbox-darwin-amd64

# macOS (Apple Silicon)
curl -LO https://github.com/luwojtaszek/cc-sandbox/releases/latest/download/cc-sandbox-darwin-arm64

# Windows
curl -LO https://github.com/luwojtaszek/cc-sandbox/releases/latest/download/cc-sandbox-windows-amd64.exe
```

Then install:

```bash
chmod +x cc-sandbox-*
sudo mv cc-sandbox-* /usr/local/bin/cc-sandbox
```

## Building from Source

### Prerequisites

- Go 1.22+
- Make
- Docker 20.10+ (for building images)

### Build CLI

```bash
git clone https://github.com/luwojtaszek/cc-sandbox.git
cd cc-sandbox

# Build for current platform
make cli

# Build for all platforms
make cli-all

# Install to ~/.local/bin
make install-user
```

### Build Docker Images

```bash
# Build all images
make images

# Or build individually
make base
make docker
make bun-full
```

## Updating

The CLI includes a built-in update command:

```bash
cc-sandbox update
```

This will:
1. Check for and install the latest CLI version
2. Update any locally installed cc-sandbox Docker images

### Update Options

```bash
# Update only CLI
cc-sandbox update --skip-images

# Update only Docker images
cc-sandbox update --skip-cli

# Force update even if already on latest
cc-sandbox update --force
```

## Verifying Installation

```bash
# Check version
cc-sandbox version

# Test run
cc-sandbox claude --help
```

## Uninstalling

```bash
# Remove CLI binary
rm $(which cc-sandbox)

# Remove Docker images (optional)
docker rmi cc-sandbox:base cc-sandbox:docker cc-sandbox:bun-full

# Remove credentials volume (optional - removes saved Claude credentials)
docker volume rm cc-sandbox-credentials-$(id -u)
```
