# Docker Images

cc-sandbox provides three Docker image variants, each building on the previous one.

## Image Hierarchy

```
base → docker → bun-full
```

| Image                 | Size   | Description                     |
|-----------------------|--------|---------------------------------|
| `cc-sandbox:base`     | ~800MB | Minimal Claude Code environment |
| `cc-sandbox:docker`   | ~900MB | Base + Docker CLI               |
| `cc-sandbox:bun-full` | ~2GB   | Full development environment    |

## Base Image (`cc-sandbox:base`)

Minimal image suitable for most tasks.

### Contents

**Operating System:**

- Ubuntu 24.04 LTS

**Runtime:**

- Node.js 22 LTS
- npm (latest)

**Version Control:**

- Git
- GitHub CLI (`gh`)

**Claude Code:**

- Claude Code (latest)

**Utilities:**

- ripgrep (`rg`) - fast file search
- fzf - fuzzy finder
- jq - JSON processor
- tree - directory listing
- curl, wget - HTTP clients
- zip, unzip - archive tools

### Usage

```bash
cc-sandbox claude                    # Uses base by default
cc-sandbox -i base claude            # Explicit
```

## Docker Image (`cc-sandbox:docker`)

Extends base with Docker CLI for container management.

### Contents

Everything in base, plus:

- Docker CLI
- Docker Compose plugin

### Usage

```bash
cc-sandbox -i docker --docker claude
```

The `--docker` flag mounts the host's Docker socket, enabling Docker commands inside the container.

### Use Cases

- Building and running Docker containers
- Docker Compose workflows
- Container orchestration tasks
- CI/CD pipeline development

## Full Image (`cc-sandbox:bun-full`)

Full development environment with additional runtimes and tools.

### Contents

Everything in docker, plus:

**Additional Runtimes:**

- Bun
- Python 3 with pip
- pnpm

**Browser Automation:**

- Playwright with Chromium
- agent-browser for web automation

**Python Packages:**

- httpie - HTTP client
- rich - terminal formatting
- pyyaml - YAML processing

**Additional Utilities:**

- htop - process viewer
- vim-tiny, nano - text editors
- dnsutils, iputils-ping - network tools
- netcat - network utility
- p7zip-full - archive tool

### Usage

```bash
cc-sandbox -i bun-full claude
```

### Use Cases

- Full-stack development
- Web scraping and automation
- Multi-runtime projects (Node.js, Bun, Python)
- Browser testing

## Using Registry Images

By default, cc-sandbox pulls images from `ghcr.io/luwojtaszek`:

```bash
# These are equivalent
cc-sandbox claude
CC_SANDBOX_REGISTRY=ghcr.io/luwojtaszek cc-sandbox claude
```

Images are automatically pulled on first use.

## Using Local Images

To use locally built images instead of registry images:

```bash
# Build images locally
make images

# Use local images
CC_SANDBOX_REGISTRY="" cc-sandbox claude
```

## Custom Images

See [Custom Images Guide](custom-images.md) for building and configuring your own images.

## Image Architecture

All images support:

- `linux/amd64` (x86_64)
- `linux/arm64` (Apple Silicon, ARM servers)

## Updating Images

```bash
# Update all local images
cc-sandbox update

# Update only images (skip CLI)
cc-sandbox update --skip-cli

# Manual pull
docker pull ghcr.io/luwojtaszek/cc-sandbox:base
```
