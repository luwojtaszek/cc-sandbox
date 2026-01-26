#!/bin/bash
set -e

# Debug logging - only show messages when CC_SANDBOX_DEBUG=1
debug_log() {
    if [ "$CC_SANDBOX_DEBUG" = "1" ]; then
        echo "$@"
    fi
}

# Permission mode: "skip" or "accept"
# - "skip" uses --dangerously-skip-permissions (for non-root users)
# - "accept" uses --permission-mode acceptEdits (for root users who can't use skip)
PERMISSION_MODE="${CC_SANDBOX_PERMISSION_MODE:-skip}"

# Determine Claude flags based on permission mode
if [ "$PERMISSION_MODE" = "accept" ]; then
    CLAUDE_FLAGS="--permission-mode acceptEdits"
else
    CLAUDE_FLAGS="--dangerously-skip-permissions"
fi

# Determine if we're running as root or non-root
CURRENT_UID=$(id -u)

if [ "$CURRENT_UID" = "0" ]; then
    # Running as root (rootless Docker mode)
    export HOME="/root"
    debug_log "[cc-sandbox] Running as root (rootless Docker mode)"
else
    # Running as non-root user - use fixuid to remap UID/GID
    export HOME="/home/claude"

    # Run fixuid to remap claude user to current UID/GID and fix file ownership
    # fixuid reads from /etc/fixuid/config.yml
    eval "$(fixuid -q)"

    debug_log "[cc-sandbox] Running as user $CURRENT_UID:$(id -g)"
fi

# Handle Claude credentials (same logic, uses $HOME)
if [ -d "/mnt/claude-data" ]; then
    # Ensure .claude directory exists
    mkdir -p /mnt/claude-data/.claude

    if [ -d "$HOME/.claude" ] && [ ! -L "$HOME/.claude" ]; then
        cp -rn "$HOME/.claude/"* /mnt/claude-data/.claude/ 2>/dev/null || true
        rm -rf "$HOME/.claude"
    fi

    if [ ! -L "$HOME/.claude" ]; then
        ln -sf /mnt/claude-data/.claude "$HOME/.claude"
    fi

    debug_log "[cc-sandbox] Using persistent credentials from /mnt/claude-data"
fi

# Handle .claude.json file persistence
if [ -d "/mnt/claude-data" ]; then
    # If .claude.json exists in home and is not a symlink, move it to volume
    if [ -f "$HOME/.claude.json" ] && [ ! -L "$HOME/.claude.json" ]; then
        # Only copy if volume doesn't already have the file (preserve existing)
        if [ ! -f "/mnt/claude-data/.claude.json" ]; then
            cp "$HOME/.claude.json" /mnt/claude-data/.claude.json 2>/dev/null || true
        fi
        rm -f "$HOME/.claude.json"
    fi

    # Create symlink if it doesn't exist
    if [ ! -L "$HOME/.claude.json" ]; then
        ln -sf /mnt/claude-data/.claude.json "$HOME/.claude.json"
    fi
fi

# Handle Claude configuration from host mount or git repo
# Usage: setup_claude_config_from_mount <src_dir> [skip_settings]
setup_claude_config_from_mount() {
    local src_dir="$1"
    local skip_settings="$2"

    # Link skills
    if [ -d "$src_dir/skills" ]; then
        mkdir -p "$HOME/.claude/skills"
        for skill_dir in "$src_dir/skills"/*/; do
            [ -d "$skill_dir" ] || continue
            skill_name=$(basename "$skill_dir")
            [ -e "$HOME/.claude/skills/$skill_name" ] || \
                ln -sf "$skill_dir" "$HOME/.claude/skills/$skill_name"
        done
        debug_log "[cc-sandbox] Linked skills from $src_dir"
    fi

    # Link agents
    if [ -d "$src_dir/agents" ]; then
        mkdir -p "$HOME/.claude/agents"
        for agent in "$src_dir/agents"/*.md; do
            [ -f "$agent" ] || continue
            name=$(basename "$agent")
            [ -e "$HOME/.claude/agents/$name" ] || \
                ln -sf "$agent" "$HOME/.claude/agents/$name"
        done
        debug_log "[cc-sandbox] Linked agents from $src_dir"
    fi

    # Link commands (legacy)
    if [ -d "$src_dir/commands" ]; then
        mkdir -p "$HOME/.claude/commands"
        for cmd in "$src_dir/commands"/*.md; do
            [ -f "$cmd" ] || continue
            name=$(basename "$cmd")
            [ -e "$HOME/.claude/commands/$name" ] || \
                ln -sf "$cmd" "$HOME/.claude/commands/$name"
        done
        debug_log "[cc-sandbox] Linked commands from $src_dir"
    fi

    # Handle settings.json + .credentials coupling (skip for git repo without install.sh)
    if [ "$skip_settings" != "skip_settings" ] && [ -f "$src_dir/settings.json" ]; then
        if [ ! -f "$src_dir/.credentials" ]; then
            echo "[cc-sandbox] ERROR: settings.json found but .credentials missing in $src_dir"
            echo "[cc-sandbox] These files must be mounted together. Exiting."
            exit 1
        fi
        # Copy both files (they're coupled)
        cp "$src_dir/settings.json" "$HOME/.claude/settings.json"
        cp "$src_dir/.credentials" "$HOME/.claude/.credentials"
        debug_log "[cc-sandbox] Applied settings.json + .credentials from $src_dir"
    fi

    # Link CLAUDE.md
    if [ -f "$src_dir/CLAUDE.md" ] && [ ! -f "$HOME/CLAUDE.md" ]; then
        ln -sf "$src_dir/CLAUDE.md" "$HOME/CLAUDE.md"
        debug_log "[cc-sandbox] Linked global CLAUDE.md"
    fi
}

# Process host-mounted config (highest priority)
if [ -d "/mnt/host-claude-config" ]; then
    debug_log "[cc-sandbox] Processing host-mounted Claude config..."
    setup_claude_config_from_mount "/mnt/host-claude-config"
fi

# Process git repo config
if [ -n "$CC_CLAUDE_CONFIG_REPO" ]; then
    CONFIG_DIR="/mnt/claude-config-repo"

    # Clone repo if not present
    if [ ! -d "$CONFIG_DIR/.git" ]; then
        debug_log "[cc-sandbox] Cloning Claude config from: $CC_CLAUDE_CONFIG_REPO"
        git clone --depth 1 "$CC_CLAUDE_CONFIG_REPO" "$CONFIG_DIR" 2>/dev/null || \
            debug_log "[cc-sandbox] Warning: Failed to clone config repo"
    elif [ "$CC_CLAUDE_CONFIG_SYNC" = "1" ]; then
        # Sync requested - pull latest changes
        debug_log "[cc-sandbox] Syncing Claude config from: $CC_CLAUDE_CONFIG_REPO"
        git -C "$CONFIG_DIR" pull --ff-only 2>/dev/null || \
            debug_log "[cc-sandbox] Warning: Failed to pull config repo updates"
    else
        debug_log "[cc-sandbox] Using cached Claude config (use --claude-config-sync to update)"
    fi

    # Run install.sh if present (responsible for settings.json + credentials handling)
    if [ -x "$CONFIG_DIR/install.sh" ]; then
        debug_log "[cc-sandbox] Running install.sh from config repo"
        (cd "$CONFIG_DIR" && ./install.sh)
    else
        # Default: link skills/agents/commands only (skip settings.json for safety)
        setup_claude_config_from_mount "$CONFIG_DIR" "skip_settings"
    fi
fi

# Handle Playwright browsers (shared installation at /opt/ms-playwright)
if [ -d "/opt/ms-playwright" ]; then
    mkdir -p "$HOME/.cache"
    if [ ! -L "$HOME/.cache/ms-playwright" ]; then
        ln -sf /opt/ms-playwright "$HOME/.cache/ms-playwright"
    fi
fi

# Handle Git configuration
# Note: Image has a default .gitconfig with basic settings (init.defaultBranch, etc.)
# We need to merge host config and apply any env var overrides

if [ -f "/mnt/host-config/.gitconfig" ]; then
    # Host gitconfig is mounted - import user.name and user.email from it
    # (keeps image defaults for other settings)
    HOST_GIT_NAME=$(git config -f /mnt/host-config/.gitconfig user.name 2>/dev/null || true)
    HOST_GIT_EMAIL=$(git config -f /mnt/host-config/.gitconfig user.email 2>/dev/null || true)

    if [ -n "$HOST_GIT_NAME" ]; then
        git config --global user.name "$HOST_GIT_NAME"
        debug_log "[cc-sandbox] Git user.name from host: $HOST_GIT_NAME"
    fi
    if [ -n "$HOST_GIT_EMAIL" ]; then
        git config --global user.email "$HOST_GIT_EMAIL"
        debug_log "[cc-sandbox] Git user.email from host: $HOST_GIT_EMAIL"
    fi
fi

# Apply git config from environment variables (highest priority - overrides host)
if [ -n "$CC_GIT_USER_NAME" ]; then
    git config --global user.name "$CC_GIT_USER_NAME"
    debug_log "[cc-sandbox] Git user.name override: $CC_GIT_USER_NAME"
fi

if [ -n "$CC_GIT_USER_EMAIL" ]; then
    git config --global user.email "$CC_GIT_USER_EMAIL"
    debug_log "[cc-sandbox] Git user.email override: $CC_GIT_USER_EMAIL"
fi

# Handle GitHub token
if [ -n "$GH_TOKEN" ] || [ -n "$GITHUB_TOKEN" ]; then
    debug_log "[cc-sandbox] GitHub token detected in environment"
fi

# Handle gh config
if [ -d "/mnt/host-config/gh" ]; then
    mkdir -p "$HOME/.config"
    # Remove empty pre-created directory if it exists (from Dockerfile)
    if [ -d "$HOME/.config/gh" ] && [ -z "$(ls -A "$HOME/.config/gh" 2>/dev/null)" ]; then
        rmdir "$HOME/.config/gh"
    fi
    # Create symlink if directory doesn't exist (or was just removed)
    if [ ! -e "$HOME/.config/gh" ]; then
        ln -sf /mnt/host-config/gh "$HOME/.config/gh"
        debug_log "[cc-sandbox] Using host GitHub CLI config"
    fi
fi

# Configure git to use gh as credential helper for HTTPS auth
if command -v gh &> /dev/null; then
    # Check if gh is authenticated (either via config or token)
    if [ -d "$HOME/.config/gh" ] || [ -n "$GH_TOKEN" ] || [ -n "$GITHUB_TOKEN" ]; then
        # Only set if no existing credential helper is configured for GitHub
        EXISTING_HELPER=$(git config --global --get credential.https://github.com.helper 2>/dev/null || true)
        if [ -z "$EXISTING_HELPER" ]; then
            git config --global credential.https://github.com.helper "!gh auth git-credential"
            debug_log "[cc-sandbox] Git credential helper configured (gh)"
        else
            debug_log "[cc-sandbox] Using existing credential helper: $EXISTING_HELPER"
        fi
    fi
fi

# Handle SSH keys
if [ -d "/mnt/host-config/.ssh" ]; then
    mkdir -p "$HOME/.ssh"
    cp -rn /mnt/host-config/.ssh/* "$HOME/.ssh/" 2>/dev/null || true
    chmod 700 "$HOME/.ssh"
    chmod 600 "$HOME/.ssh/"* 2>/dev/null || true
    debug_log "[cc-sandbox] SSH keys available"
fi

# Docker socket handling (group-add is handled by CLI via --group-add flag)
if [ -S "/var/run/docker.sock" ]; then
    debug_log "[cc-sandbox] Docker socket available"
fi

# Display startup info
debug_log "[cc-sandbox] Starting in /workspace | User: $(whoami) ($(id -u):$(id -g)) | Permission mode: $PERMISSION_MODE"

# Execute command
if [ $# -eq 0 ]; then
    exec claude $CLAUDE_FLAGS
fi

if [ "$1" = "claude" ]; then
    shift
    exec claude $CLAUDE_FLAGS "$@"
fi

exec "$@"
