# cc-sandbox

Run Claude Code in isolated Docker containers.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/luwojtaszek/cc-sandbox/main/install.sh | bash
```

## Quick Start

```bash
cc-sandbox claude
```

## Configuration

Configure via environment variables in your shell profile (`~/.bashrc`, `~/.zshrc`):

### Image Selection

| Variable | Description | Default |
|----------|-------------|---------|
| `CC_SANDBOX_DEFAULT_IMAGE` | Image variant | `base` |
| `CC_SANDBOX_REGISTRY` | Registry prefix | `ghcr.io/luwojtaszek` |

Available images:

- **base** - Node.js, Git, GitHub CLI, Claude Code (~800MB)
- **docker** - Base + Docker CLI (~900MB)
- **bun-full** - Full dev: Bun, Python, Playwright (~2GB)

Example:
```bash
export CC_SANDBOX_DEFAULT_IMAGE=docker
```

### Claude Code Configuration

Share your Claude skills, agents, and settings.

| Variable | Description |
|----------|-------------|
| `CC_SANDBOX_CLAUDE_CONFIG` | Path to host `~/.claude` directory |
| `CC_SANDBOX_CLAUDE_CONFIG_REPO` | Git URL for shared config |

Example:
```bash
export CC_SANDBOX_CLAUDE_CONFIG=~/.claude
# or for team shared config:
export CC_SANDBOX_CLAUDE_CONFIG_REPO=https://github.com/org/claude-config.git
```

#### Config Repo Setup

When using `CC_SANDBOX_CLAUDE_CONFIG_REPO`, the entrypoint automatically sets up your configuration:

**Default behavior** (without `install.sh`):
- Auto-links `skills/*/` directories to `~/.claude/skills/`
- Auto-links `agents/*.md` files to `~/.claude/agents/`
- Auto-links `commands/*.md` files to `~/.claude/commands/` (legacy)
- Auto-links `CLAUDE.md` if present
- Skips `settings.json` for safety (requires custom install.sh)

**Custom behavior** (with `install.sh`):

If your config repo contains an executable `install.sh`, it runs instead of auto-linking. This gives you full control over setup and is required for handling `settings.json` and `.credentials`.

Example `install.sh`:
```bash
#!/bin/bash
# install.sh - Custom setup for Claude config repo
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Link skills
if [ -d "$SCRIPT_DIR/skills" ]; then
    mkdir -p "$HOME/.claude/skills"
    for skill in "$SCRIPT_DIR/skills"/*/; do
        [ -d "$skill" ] || continue
        name=$(basename "$skill")
        ln -sf "$skill" "$HOME/.claude/skills/$name"
    done
fi

# Link agents
if [ -d "$SCRIPT_DIR/agents" ]; then
    mkdir -p "$HOME/.claude/agents"
    for agent in "$SCRIPT_DIR/agents"/*.md; do
        [ -f "$agent" ] || continue
        ln -sf "$agent" "$HOME/.claude/agents/$(basename "$agent")"
    done
fi

# Apply settings (only if .credentials also exists)
if [ -f "$SCRIPT_DIR/settings.json" ] && [ -f "$SCRIPT_DIR/.credentials" ]; then
    cp "$SCRIPT_DIR/settings.json" "$HOME/.claude/settings.json"
    cp "$SCRIPT_DIR/.credentials" "$HOME/.claude/.credentials"
fi
```

### Git Identity

| Variable | Description |
|----------|-------------|
| `CC_SANDBOX_GIT_USER_NAME` | Git commit author name |
| `CC_SANDBOX_GIT_USER_EMAIL` | Git commit author email |

Auto-detected from host git config if not set.

### Custom Images

Build your own image extending cc-sandbox:

```dockerfile
FROM ghcr.io/luwojtaszek/cc-sandbox:base
RUN apt-get update && apt-get install -y golang-go
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
```

Then:
```bash
docker build -t cc-sandbox:my-image .
export CC_SANDBOX_DEFAULT_IMAGE=my-image
```

For Docker socket auto-mount on custom images:
```bash
export CC_SANDBOX_DOCKER_IMAGES=my-image
```

## Credentials

- Claude credentials persist in Docker volume (one-time login)
- Git/GitHub CLI config auto-mounted from host
- Use `--ssh` flag for SSH key access
- `GH_TOKEN`/`GITHUB_TOKEN` passed through automatically

## Updating

```bash
cc-sandbox update
```

## Reference

- [CLI Reference](cli-reference.md) - All flags and commands
- [Troubleshooting](troubleshooting.md) - Common issues
