#!/bin/bash
set -e

# Run fixuid to map container user to host user's UID/GID
eval "$(fixuid -q)"

# Handle Claude credentials
if [ -d "/mnt/claude-data" ]; then
    mkdir -p /mnt/claude-data/.claude

    if [ -d "$HOME/.claude" ] && [ ! -L "$HOME/.claude" ]; then
        cp -rn "$HOME/.claude/"* /mnt/claude-data/.claude/ 2>/dev/null || true
        rm -rf "$HOME/.claude"
    fi

    if [ ! -L "$HOME/.claude" ]; then
        ln -sf /mnt/claude-data/.claude "$HOME/.claude"
    fi

    echo "[cc-sandbox] Using persistent credentials from /mnt/claude-data"
fi

# Handle Git configuration
if [ -f "/mnt/host-config/.gitconfig" ] && [ ! -f "$HOME/.gitconfig" ]; then
    ln -sf /mnt/host-config/.gitconfig "$HOME/.gitconfig"
    echo "[cc-sandbox] Using host .gitconfig"
fi

# Handle GitHub token
if [ -n "$GH_TOKEN" ] || [ -n "$GITHUB_TOKEN" ]; then
    echo "[cc-sandbox] GitHub token detected in environment"
fi

# Handle gh config
if [ -d "/mnt/host-config/gh" ] && [ ! -d "$HOME/.config/gh" ]; then
    mkdir -p "$HOME/.config"
    ln -sf /mnt/host-config/gh "$HOME/.config/gh"
    echo "[cc-sandbox] Using host GitHub CLI config"
fi

# Handle SSH keys
if [ -d "/mnt/host-config/.ssh" ]; then
    mkdir -p "$HOME/.ssh"
    cp -rn /mnt/host-config/.ssh/* "$HOME/.ssh/" 2>/dev/null || true
    chmod 700 "$HOME/.ssh"
    chmod 600 "$HOME/.ssh/"* 2>/dev/null || true
    echo "[cc-sandbox] SSH keys available"
fi

# Docker socket handling (group-add is handled by CLI via --group-add flag)
if [ -S "/var/run/docker.sock" ]; then
    echo "[cc-sandbox] Docker socket available"
fi

# Display startup info
echo "[cc-sandbox] Starting in /workspace"
echo "[cc-sandbox] User: $(whoami) ($(id -u):$(id -g))"

# Execute the command
if [ $# -eq 0 ]; then
    exec claude
fi

if [ "$1" = "claude" ]; then
    shift
    exec claude --dangerously-skip-permissions "$@"
fi

exec "$@"
