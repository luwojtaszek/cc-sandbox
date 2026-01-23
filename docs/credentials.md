# Credentials & Configuration

cc-sandbox manages Claude credentials and host configuration mounting to provide a seamless experience.

## Claude Credential Persistence

Your Claude Code credentials are automatically persisted in a Docker volume, so you only need to authenticate once.

### Volume Location

Credentials are stored in a user-specific volume:

- Volume name: `cc-sandbox-credentials-<uid>` (e.g., `cc-sandbox-credentials-501`)
- Container mount point: `/mnt/claude-data`

### How It Works

1. On first run, you'll be prompted to authenticate with Claude
2. Credentials are saved to the Docker volume
3. Subsequent runs automatically use the saved credentials
4. Different host users have separate credential volumes

### Migration from Shared Volume

If you previously used cc-sandbox with the old shared volume (`cc-sandbox-credentials`), your credentials are
automatically migrated to the new user-specific volume on first run.

### Clearing Credentials

To re-authenticate, remove the credentials volume:

```bash
# Remove credentials for current user
docker volume rm cc-sandbox-credentials-$(id -u)

# On Windows
docker volume rm cc-sandbox-credentials-<your-username>
```

### Checking Credential Status

```bash
# List volumes
docker volume ls | grep cc-sandbox

# Inspect volume
docker volume inspect cc-sandbox-credentials-$(id -u)
```

## Host Configuration Mounting

cc-sandbox can mount configuration files from your host machine.

### Git Configuration (`--git`)

**Default:** Enabled

Mounts `~/.gitconfig` as read-only at `/mnt/host-config/.gitconfig`.

This enables:

- Your git aliases
- Credential helpers
- User identity (name, email)
- Custom git settings

```bash
# Enabled by default
cc-sandbox claude

# Disable
cc-sandbox --git=false claude
```

### GitHub CLI Configuration (`--gh`)

**Default:** Enabled

Mounts `~/.config/gh` as read-only at `/mnt/host-config/gh`.

This enables:

- Authenticated `gh` commands
- GitHub token access
- Your `gh` configuration

```bash
# Enabled by default
cc-sandbox claude

# Disable
cc-sandbox --gh=false claude
```

### Environment Token Passthrough

If `GH_TOKEN` or `GITHUB_TOKEN` is set in your environment, it's passed through to the container:

```bash
GH_TOKEN=xxx cc-sandbox claude
```

### SSH Keys (`--ssh`)

**Default:** Disabled (opt-in for security)

Mounts `~/.ssh` as read-only at `/mnt/host-config/.ssh`.

This enables:

- SSH-based git operations (`git@github.com:...`)
- SSH connections from within the container

```bash
# Enable SSH key access
cc-sandbox --ssh claude
```

**Security Note:** SSH keys are mounted read-only, but enable access to any systems your keys can authenticate to.

## Git Identity Configuration

cc-sandbox resolves your git identity in this priority order:

1. CLI flags (`--git-user-name`, `--git-user-email`)
2. Environment variables (`CC_SANDBOX_GIT_USER_NAME`, `CC_SANDBOX_GIT_USER_EMAIL`)
3. Host git configuration (auto-detected from `git config --global`)

### Override Git Identity

```bash
# Via CLI flags
cc-sandbox --git-user-name "John Doe" --git-user-email "john@example.com" claude

# Via environment variables
export CC_SANDBOX_GIT_USER_NAME="John Doe"
export CC_SANDBOX_GIT_USER_EMAIL="john@example.com"
cc-sandbox claude
```

### Auto-Detection

If no override is specified, cc-sandbox reads your host's git configuration:

```bash
# What cc-sandbox detects
git config --global user.name
git config --global user.email
```

These values are passed to the container via environment variables (`CC_GIT_USER_NAME`, `CC_GIT_USER_EMAIL`) and
configured by the entrypoint script.

## Git Worktree Support

cc-sandbox automatically detects and supports git worktrees.

### How It Works

When your working directory is a git worktree:

1. cc-sandbox reads the `.git` file to find the bare repository
2. The bare repository path is mounted into the container
3. Git operations work correctly across the worktree

### Example

```bash
# Create a worktree
git worktree add ../feature-branch feature

# cd into worktree
cd ../feature-branch

# cc-sandbox automatically mounts the bare repo
cc-sandbox claude
```

## Configuration Summary

| Configuration      | Mount Point                   | Default | Flag    |
|--------------------|-------------------------------|---------|---------|
| Claude credentials | `/mnt/claude-data`            | Always  | -       |
| Git config         | `/mnt/host-config/.gitconfig` | On      | `--git` |
| GitHub CLI         | `/mnt/host-config/gh`         | On      | `--gh`  |
| SSH keys           | `/mnt/host-config/.ssh`       | Off     | `--ssh` |

## Best Practices

1. **Keep defaults for development** - Git and GitHub CLI mounting simplifies the workflow
2. **Use SSH flag sparingly** - Only enable when needed for git operations
3. **Override identity for work projects** - Use flags when committing to different repositories
4. **Don't share credentials volume** - Each user should have their own volume
