# cc-sandbox:base
# Base image for Claude Code sandbox with UID/GID mapping support

FROM ubuntu:24.04

LABEL org.opencontainers.image.source="https://github.com/luwojtaszek/cc-sandbox"
LABEL org.opencontainers.image.description="Claude Code Sandbox - Base Image"
LABEL org.opencontainers.image.licenses="MIT"

ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies (no sudo - security best practice)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    wget \
    gnupg \
    git \
    jq \
    ripgrep \
    fzf \
    tree \
    unzip \
    zip \
    locales \
    && rm -rf /var/lib/apt/lists/*

# Setup locales
RUN locale-gen en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

# Install Node.js 22 LTS
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/* \
    && npm install -g npm@latest

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# Install gosu for privilege dropping
ARG GOSU_VERSION=1.17
RUN case "$(dpkg --print-architecture)" in \
        amd64) GOSU_ARCH="amd64" ;; \
        arm64) GOSU_ARCH="arm64" ;; \
        *) echo "Unsupported architecture" && exit 1 ;; \
    esac \
    && curl -fsSL "https://github.com/tianon/gosu/releases/download/${GOSU_VERSION}/gosu-${GOSU_ARCH}" -o /usr/local/bin/gosu \
    && chmod +x /usr/local/bin/gosu \
    && gosu --version

# Install fixuid for dynamic UID/GID mapping
ARG FIXUID_VERSION=0.6.0
RUN case "$(dpkg --print-architecture)" in \
        amd64) FIXUID_ARCH="amd64" ;; \
        arm64) FIXUID_ARCH="arm64" ;; \
        *) echo "Unsupported architecture" && exit 1 ;; \
    esac \
    && curl -fsSL "https://github.com/boxboat/fixuid/releases/download/v${FIXUID_VERSION}/fixuid-${FIXUID_VERSION}-linux-${FIXUID_ARCH}.tar.gz" | tar -C /usr/local/bin -xzf - \
    && chmod 4755 /usr/local/bin/fixuid \
    && mkdir -p /etc/fixuid

# Create non-root user 'claude' (no sudo - security best practice)
# Remove existing user/group with UID/GID 1000 if present (ubuntu base image has 'ubuntu' user)
RUN userdel -r ubuntu 2>/dev/null || true \
    && groupdel ubuntu 2>/dev/null || true \
    && groupadd --gid 1000 claude \
    && useradd --uid 1000 --gid claude --shell /bin/bash --create-home claude

# Configure fixuid
RUN echo "user: claude\ngroup: claude\npaths:\n  - /home/claude\n  - /workspace\n  - /mnt/claude-data" > /etc/fixuid/config.yml

# Create workspace and config directories
RUN mkdir -p /workspace \
    && mkdir -p /mnt/claude-data \
    && mkdir -p /home/claude/.claude \
    && mkdir -p /home/claude/.config/gh \
    && chown claude:claude /workspace \
    && chown claude:claude /mnt/claude-data \
    && chown -R claude:claude /home/claude

# Switch to claude user for Claude Code installation
USER claude
WORKDIR /home/claude

# Install Claude Code for claude user and add ~/.local/bin to PATH
RUN curl -fsSL https://claude.ai/install.sh | bash \
    && echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc

# Configure git defaults for claude user
RUN git config --global init.defaultBranch main \
    && git config --global core.editor "nano" \
    && git config --global pull.rebase false

# Switch to root for root user setup
USER root

# Install Claude Code for root user and add ~/.local/bin to PATH
RUN curl -fsSL https://claude.ai/install.sh | bash \
    && echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc

# Set up root home directory
RUN mkdir -p /root/.claude \
    && mkdir -p /root/.config/gh

# Configure git defaults for root user
RUN git config --global init.defaultBranch main \
    && git config --global core.editor "nano" \
    && git config --global pull.rebase false

# Copy entrypoint script (as root, will handle privilege dropping)
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Add ~/.local/bin to PATH for non-interactive shells
ENV PATH="/root/.local/bin:/home/claude/.local/bin:${PATH}"

# NOTE: No final USER directive! Container starts as root.
# Entrypoint handles privilege dropping based on CC_SANDBOX_USER env var.

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
