# cc-sandbox Documentation

Welcome to the cc-sandbox documentation. This guide covers everything you need to run Claude Code in isolated Docker
containers.

## Quick Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/luwojtaszek/cc-sandbox/main/install.sh | bash

# Run
cc-sandbox claude
```

## Documentation

| Guide                                         | Description                                        |
|-----------------------------------------------|----------------------------------------------------|
| [Installation](installation.md)               | Install methods, updating, building from source    |
| [CLI Reference](cli-reference.md)             | Commands, flags, environment variables             |
| [Docker Images](images.md)                    | Image variants, contents, building custom images   |
| [Container Runtimes](container-runtimes.md)   | Docker Desktop, OrbStack, Podman, rootless Docker  |
| [Credentials & Configuration](credentials.md) | Claude credentials, Git/GitHub/SSH config mounting |
| [Troubleshooting](troubleshooting.md)         | Common issues and solutions                        |

## Getting Started Checklist

1. **Install the CLI** - Use the one-line installer or `go install`
2. **Pull an image** - Happens automatically on first run
3. **Run Claude Code** - `cc-sandbox claude` in any project directory
4. **Authenticate** - Follow the Claude Code login prompts (one-time)

## Key Concepts

### Images

cc-sandbox provides three Docker image variants:

- **base** - Minimal: Node.js, Git, GitHub CLI, Claude Code
- **docker** - Base + Docker CLI for container management
- **bun-full** - Full dev environment with Bun, Python, Playwright

### Credential Persistence

Your Claude credentials are stored in a Docker volume and persist across container restarts. You only need to
authenticate once.

### Host Configuration

By default, cc-sandbox mounts your Git and GitHub CLI configuration from the host, so commits and `gh` commands work
seamlessly.

## Getting Help

- [GitHub Issues](https://github.com/luwojtaszek/cc-sandbox/issues) - Bug reports and feature requests
- [CLI Help](cli-reference.md) - `cc-sandbox --help`
