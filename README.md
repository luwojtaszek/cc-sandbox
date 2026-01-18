# cc-sandbox

A customizable Docker sandbox for running Claude Code in isolated containers with proper UID/GID mapping, credential management, and optional Docker socket access.

## Features

- **Secure by default** - No sudo in containers, non-root user with minimal privileges
- **UID/GID mapping** - Files created in containers have correct host ownership via fixuid
- **Multiple image variants** - From minimal base to full development environment
- **Credential persistence** - Claude credentials persist across container restarts
- **Host config mounting** - Optionally mount Git, GitHub CLI, and SSH configs
- **Docker-in-Docker** - Optional Docker socket access for container management
- **Cross-platform CLI** - Go-based CLI works on Linux, macOS, and Windows

## Quick Start

### Using Docker directly

```bash
# Pull the base image
docker pull ghcr.io/luwojtaszek/cc-sandbox:base

# Run Claude Code in current directory
docker run -it --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  --user $(id -u):$(id -g) \
  ghcr.io/luwojtaszek/cc-sandbox:base
```

### Using the CLI (recommended)

```bash
# Install CLI
go install github.com/luwojtaszek/cc-sandbox/cli@latest

# Or download from releases
curl -LO https://github.com/luwojtaszek/cc-sandbox/releases/latest/download/cc-sandbox-linux-amd64
chmod +x cc-sandbox-linux-amd64
sudo mv cc-sandbox-linux-amd64 /usr/local/bin/cc-sandbox

# Run Claude Code
cc-sandbox claude

# Run with a prompt
cc-sandbox claude -p "fix the bug in main.go"

# Continue previous conversation
cc-sandbox claude -c
```

## Image Variants

| Image | Size | Description |
|-------|------|-------------|
| `cc-sandbox:base` | ~800MB | Ubuntu 24.04, Node.js 22, Git, GitHub CLI, Claude Code |
| `cc-sandbox:docker` | ~900MB | Base + Docker CLI and Compose plugin |
| `cc-sandbox:bun-full` | ~2GB | Docker + Bun, Python 3, pnpm, Playwright, agent-browser |

### Base Image (`cc-sandbox:base`)

Minimal image suitable for most tasks:

- Ubuntu 24.04 LTS
- Node.js 22 LTS
- Git, GitHub CLI (gh)
- Claude Code (latest)
- Common utilities: ripgrep, fzf, jq, tree

### Docker Image (`cc-sandbox:docker`)

For tasks requiring Docker access:

- Everything in base
- Docker CLI
- Docker Compose plugin

### Full Image (`cc-sandbox:bun-full`)

Full development environment:

- Everything in docker
- Bun runtime
- Python 3 with pip
- pnpm package manager
- Playwright with Chromium
- agent-browser for web automation
- Additional dev tools: htop, vim, nano

## CLI Usage

```bash
cc-sandbox [flags] [command] [args...]

Flags:
  -i, --image string      Docker image tag (default: base)
  -m, --mount strings     Additional volume mounts (host:container)
  -e, --env strings       Environment variables (KEY=value)
  -w, --workdir string    Working directory (default: current directory)
      --docker            Mount Docker socket
      --git               Mount .gitconfig from host (default: true)
      --gh                Mount GitHub CLI config from host (default: true)
      --ssh               Mount SSH keys from host
  -t, --interactive       Run in interactive mode with TTY (default: true)
  -h, --help              Help for cc-sandbox
  -v, --version           Version for cc-sandbox
```

### Examples

```bash
# Basic usage - interactive Claude session
cc-sandbox claude

# One-shot prompt
cc-sandbox claude -p "explain this codebase"

# Continue previous conversation
cc-sandbox claude -c

# Use docker variant with socket access
cc-sandbox -i docker --docker claude

# Use full environment
cc-sandbox -i bun-full claude

# Pass environment variables
cc-sandbox -e DEBUG=1 -e API_KEY=xxx claude

# Mount additional directories
cc-sandbox -m ~/data:/data:ro claude

# Use registry image
CC_SANDBOX_REGISTRY=ghcr.io/luwojtaszek cc-sandbox claude
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CC_SANDBOX_DEFAULT_IMAGE` | Default image tag to use | `base` |
| `CC_SANDBOX_REGISTRY` | Registry prefix for images | (empty = local) |
| `CC_SANDBOX_DOCKER_SOCKET` | Docker socket path | Auto-detected |

### Credential Persistence

Claude credentials are automatically persisted in a Docker volume (`cc-sandbox-credentials`). This means:

- You only need to authenticate once
- Credentials persist across container restarts
- Different projects share the same credentials

### Host Configuration Mounting

By default, the CLI mounts:

- `~/.gitconfig` (read-only) - Git configuration
- `~/.config/gh` (read-only) - GitHub CLI authentication

Optional mounts:

- `~/.ssh` (read-only) - SSH keys for Git operations (use `--ssh`)

## Building from Source

### Prerequisites

- Docker 20.10+
- Go 1.22+ (for CLI)
- Make

### Build Images

```bash
# Clone repository
git clone https://github.com/luwojtaszek/cc-sandbox.git
cd cc-sandbox

# Build all images
make images

# Or build individually
make base
make docker
make bun-full
```

### Build CLI

```bash
# Build for current platform
make cli

# Build for all platforms
make cli-all

# Install locally
make install-user  # Installs to ~/.local/bin
```

### Push to Registry

```bash
make push REGISTRY=ghcr.io/yourusername
```

## Security

This project follows security best practices:

- **No sudo** - The `claude` user has no elevated privileges
- **Non-root execution** - All processes run as non-root user
- **Read-only mounts** - Host configs are mounted read-only by default
- **UID/GID mapping** - fixuid ensures files have correct host ownership
- **Docker socket opt-in** - Docker socket access requires explicit `--docker` flag

### Docker Socket Access

When using `--docker`, the CLI:

1. Mounts the Docker socket into the container
2. Adds the socket's group ID via `--group-add`
3. The container can then run Docker commands without sudo

This is more secure than adding sudo to the container or running as root.

## Troubleshooting

### Permission denied on files

Ensure you're running with correct UID/GID mapping:

```bash
# CLI handles this automatically
cc-sandbox claude

# Manual Docker run
docker run --user $(id -u):$(id -g) ...
```

### Claude authentication issues

Clear the credentials volume and re-authenticate:

```bash
docker volume rm cc-sandbox-credentials
cc-sandbox claude
```

### Docker socket permission denied

Make sure you're using the `--docker` flag:

```bash
cc-sandbox -i docker --docker claude
```

### Image not found

Pull from registry or build locally:

```bash
# From registry
docker pull ghcr.io/luwojtaszek/cc-sandbox:base

# Or build locally
make images
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
