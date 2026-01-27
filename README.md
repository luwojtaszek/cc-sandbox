<p align="center">
  <img src="docs/assets/logo.svg" alt="cc-sandbox logo" width="80" />
</p>

<p align="center">
<strong>cc-sandbox</strong><br>
  Run Claude Code in isolated Docker containers. Works with Docker Desktop, OrbStack, and Podman.
</p>

<p align="center">
  <a href="./LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square" /></a>
  <a href="https://go.dev"><img alt="Built with Go" src="https://img.shields.io/badge/Built%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white" /></a>
  <a href="https://github.com/luwojtaszek/cc-sandbox/releases/latest"><img alt="Latest Release" src="https://img.shields.io/github/v/release/luwojtaszek/cc-sandbox?style=flat-square&color=green" /></a>
</p>

---


## Quick Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/luwojtaszek/cc-sandbox/main/install.sh | bash

# Run Claude Code
cc-sandbox claude

# Run with a prompt
cc-sandbox claude -p "fix the bug in main.go"
```

## Features

- **Git/GitHub CLI integration** - Your git config and `gh` auth work seamlessly
- **Credential persistence** - Authenticate once, credentials persist across sessions
- **Docker-in-Docker support** - Run Docker commands inside the container
- **Playwright/agent-browser** - Web automation with headless Chromium
- **Cross-platform** - Works on Linux, macOS, and Windows

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│  Host Machine                                                   │
│  ┌──────────────┐                                               │
│  │ cc-sandbox   │ CLI detects runtime, configures mounts        │
│  └──────┬───────┘                                               │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  Docker Container (cc-sandbox:base/docker/bun-full)         ││
│  │                                                             ││
│  │  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐       ││
│  │  │ Claude Code │  │ Your Project │  │ Credentials   │       ││
│  │  │             │  │ /workspace   │  │ /mnt/claude   │       ││
│  │  └─────────────┘  └──────────────┘  └───────────────┘       ││
│  │                                                             ││
│  │  Mounted from host:                                         ││
│  │  • Current directory → /workspace                           ││
│  │  • Git config (read-only)                                   ││
│  │  • GitHub CLI auth (read-only)                              ││
│  │  • Docker socket (optional)                                 ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
│  UID/GID mapping via fixuid ensures correct file ownership      │
└─────────────────────────────────────────────────────────────────┘
```

## Supported Platforms

| Platform | Runtime                         |
|----------|---------------------------------|
| macOS    | Docker Desktop, OrbStack        |
| Linux    | Docker, Podman, rootless Docker |

## Basic Commands

```bash
cc-sandbox claude              # Interactive session
cc-sandbox claude -p "prompt"  # One-shot prompt
cc-sandbox claude -c           # Continue previous conversation
cc-sandbox auth                # Authenticate Claude credentials
cc-sandbox update              # Update CLI and images
```

## Image Variants

| Image      | Size   | Description                                     |
|------------|--------|-------------------------------------------------|
| `base`     | ~800MB | Node.js 22, Git, GitHub CLI, Claude Code        |
| `docker`   | ~900MB | Base + Docker CLI, Compose                      |
| `bun-full` | ~2GB   | Docker + Bun, Python, Playwright, agent-browser |

```bash
cc-sandbox -i docker claude      # With Docker access (auto-enabled)
cc-sandbox -i bun-full claude    # Full dev environment with Docker
```

## Documentation

See the [`docs/`](docs/) directory for comprehensive documentation:

- [Installation](docs/installation.md) - Install methods, updating, building from source
- [CLI Reference](docs/cli-reference.md) - Commands, flags, environment variables
- [Docker Images](docs/images.md) - Image variants, contents, building custom images
- [Container Runtimes](docs/container-runtimes.md) - Docker Desktop, OrbStack, Podman
- [Credentials & Configuration](docs/credentials.md) - Claude credentials, Git/SSH config
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

## License

MIT License - see [LICENSE](LICENSE) for details.
