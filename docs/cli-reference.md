# CLI Reference

## Commands

### `cc-sandbox [flags] [command] [args...]`

Run Claude Code in a Docker container. If no command is specified, runs `claude` by default.

```bash
cc-sandbox claude                    # Interactive Claude session
cc-sandbox claude -p "fix the bug"   # One-shot prompt
cc-sandbox claude -c                 # Continue previous conversation
```

### `cc-sandbox update`

Update the CLI and Docker images.

```bash
cc-sandbox update              # Update CLI and images
cc-sandbox update --skip-cli   # Update only Docker images
cc-sandbox update --skip-images # Update only CLI
cc-sandbox update --force      # Force update even if on latest
```

### `cc-sandbox auth`

Authenticate Claude Code credentials. Runs `claude setup-token` interactively and stores the OAuth token in a persistent volume for reuse across sessions.

```bash
cc-sandbox auth              # Authenticate credentials
cc-sandbox auth --uid 10000  # Authenticate for specific UID
```

| Flag          | Description                       | Default      |
|---------------|-----------------------------------|--------------|
| `--uid <uid>` | Target UID for credentials volume | current user |

Credentials are stored in a Docker volume named `cc-sandbox-credentials-{UID}` at `/mnt/claude-data/.oauth-token`. The OAuth token is automatically injected into containers via the `CLAUDE_CODE_OAUTH_TOKEN` environment variable in subsequent `cc-sandbox claude` sessions.

### `cc-sandbox version`

Print version information.

```bash
cc-sandbox version
```

## Flags

### Image Selection

| Flag                | Description             | Default |
|---------------------|-------------------------|---------|
| `-i, --image <tag>` | Docker image tag to use | `base`  |

```bash
cc-sandbox -i docker claude      # Use docker image
cc-sandbox -i bun-full claude    # Use full development image
```

### Volume Mounts

| Flag                           | Description              | Default |
|--------------------------------|--------------------------|---------|
| `-m, --mount <host:container>` | Additional volume mounts | none    |

```bash
cc-sandbox -m ~/data:/data claude              # Mount data directory
cc-sandbox -m ~/data:/data:ro claude           # Mount as read-only
cc-sandbox -m /tmp/cache:/cache -m ~/lib:/lib claude  # Multiple mounts
```

### Environment Variables

| Flag                    | Description               | Default |
|-------------------------|---------------------------|---------|
| `-e, --env <KEY=value>` | Set environment variables | none    |

```bash
cc-sandbox -e DEBUG=1 claude
cc-sandbox -e API_KEY=xxx -e DEBUG=1 claude
```

### Working Directory

| Flag                   | Description               | Default           |
|------------------------|---------------------------|-------------------|
| `-w, --workdir <path>` | Working directory on host | current directory |

```bash
cc-sandbox -w ~/projects/myapp claude
```

### Docker Socket

| Flag       | Description                        | Default      |
|------------|------------------------------------|--------------|
| `--docker` | Mount Docker socket into container | auto-enabled |

Enables Docker commands inside the container. **Automatically enabled** when using `-i docker` or `-i bun-full` images.

```bash
cc-sandbox -i docker claude        # Docker socket auto-mounted
cc-sandbox -i base --docker claude # Explicit flag for base image
```

### Host Configuration Mounting

| Flag    | Description                  | Default |
|---------|------------------------------|---------|
| `--git` | Mount ~/.gitconfig from host | `true`  |
| `--gh`  | Mount ~/.config/gh from host | `true`  |
| `--ssh` | Mount ~/.ssh from host       | `false` |

```bash
cc-sandbox --ssh claude              # Enable SSH key access
cc-sandbox --git=false claude        # Disable git config mounting
cc-sandbox --gh=false --git=false claude  # Disable all host config
```

### Claude Code Configuration

| Flag                           | Description                              | Default |
|--------------------------------|------------------------------------------|---------|
| `-C, --claude-config <path>`   | Mount Claude config directory from host  | none    |
| `--claude-config-repo <url>`   | Git repository URL for Claude config     | none    |
| `--claude-config-sync`         | Pull latest changes from config repo     | `false` |

```bash
# Mount local Claude config
cc-sandbox --claude-config ~/.claude claude

# Use team config from Git repo
cc-sandbox --claude-config-repo https://github.com/org/claude-config.git claude

# Sync latest from repo
cc-sandbox --claude-config-repo https://github.com/org/claude-config.git --claude-config-sync claude
```

### Git Identity

| Flag                       | Description             | Default       |
|----------------------------|-------------------------|---------------|
| `--git-user-name <name>`   | Override git user.name  | auto-detected |
| `--git-user-email <email>` | Override git user.email | auto-detected |

```bash
cc-sandbox --git-user-name "John Doe" --git-user-email "john@example.com" claude
```

### Container Runtime

| Flag                  | Description                                   | Default |
|-----------------------|-----------------------------------------------|---------|
| `--runtime <runtime>` | Container runtime: `auto`, `docker`, `podman` | `auto`  |
| `--root <mode>`       | Run as root: `auto`, `true`, `false`          | `auto`  |
| `--host-network`      | Use host network mode                         | `false` |

```bash
cc-sandbox --runtime podman claude   # Use Podman
cc-sandbox --root=true claude        # Force root mode
cc-sandbox --host-network claude     # Use host network (for DinD localhost access)
```

### Interactive Mode

| Flag                | Description           | Default |
|---------------------|-----------------------|---------|
| `-t, --interactive` | Run with TTY attached | `true`  |

```bash
cc-sandbox -t=false claude -p "run tests"  # Non-interactive mode
```

## Environment Variables

These environment variables configure cc-sandbox behavior:

| Variable                        | Description                                     | Default               |
|---------------------------------|-------------------------------------------------|-----------------------|
| `CC_SANDBOX_DEFAULT_IMAGE`      | Default image tag                               | `base`                |
| `CC_SANDBOX_REGISTRY`           | Registry prefix for images                      | `ghcr.io/luwojtaszek` |
| `CC_SANDBOX_DOCKER_IMAGES`      | Additional images that auto-mount Docker socket | none                  |
| `CC_SANDBOX_DOCKER_SOCKET`      | Docker socket path                              | auto-detected         |
| `CC_SANDBOX_ROOT`               | Run as root: `auto`, `true`, `false`            | `auto`                |
| `CC_SANDBOX_RUNTIME`            | Container runtime: `auto`, `docker`, `podman`   | `auto`                |
| `CC_SANDBOX_GIT_USER_NAME`      | Override git user.name                          | none                  |
| `CC_SANDBOX_GIT_USER_EMAIL`     | Override git user.email                         | none                  |
| `CC_SANDBOX_CLAUDE_CONFIG`      | Path to host Claude config directory            | none                  |
| `CC_SANDBOX_CLAUDE_CONFIG_REPO` | Git repository URL for Claude config            | none                  |
| `CC_SANDBOX_DEBUG`              | Enable debug output (`1` to enable)             | none                  |

```bash
# Use docker image by default
export CC_SANDBOX_DEFAULT_IMAGE=docker

# Use local images instead of registry
export CC_SANDBOX_REGISTRY=""

# Enable debug output
CC_SANDBOX_DEBUG=1 cc-sandbox claude
```

## Examples

### Basic Usage

```bash
# Interactive session
cc-sandbox claude

# One-shot prompt
cc-sandbox claude -p "explain this codebase"

# Continue previous conversation
cc-sandbox claude -c
```

### Docker-in-Docker

```bash
# Run with Docker access (auto-enabled for docker image)
cc-sandbox -i docker claude

# With host network for localhost port access
cc-sandbox -i docker --host-network claude
```

### Full Development Environment

```bash
# Use bun-full with all tools (Docker auto-enabled)
cc-sandbox -i bun-full claude

# With SSH keys
cc-sandbox -i bun-full --ssh claude
```

### CI/Non-Interactive Mode

```bash
# Run a specific command non-interactively
cc-sandbox -t=false claude -p "run the tests and report results"
```

### Custom Environment

```bash
# Pass environment variables and mounts
cc-sandbox \
  -e DEBUG=1 \
  -e API_KEY=xxx \
  -m ~/data:/data:ro \
  -m ~/cache:/cache \
  claude
```

### Using Podman

```bash
# Explicitly use Podman
cc-sandbox --runtime podman claude
```

### Claude Code Configuration

```bash
# Mount local Claude config (skills, settings, etc.)
cc-sandbox --claude-config ~/.claude claude

# Use team config from Git repo
cc-sandbox --claude-config-repo https://github.com/org/claude-config.git claude

# Combine team config with local overrides
cc-sandbox \
  --claude-config-repo https://github.com/org/team-config.git \
  --claude-config ~/.claude \
  claude
```
