# cc-sandbox:bun-full
# Full-featured development environment

FROM cc-sandbox:docker

LABEL org.opencontainers.image.source="https://github.com/luwojtaszek/cc-sandbox"
LABEL org.opencontainers.image.description="Claude Code Sandbox - Full Development Environment"
LABEL org.opencontainers.image.licenses="MIT"

USER root

# Install Python 3 and pip
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-pip \
    python3-venv \
    python3-dev \
    && rm -rf /var/lib/apt/lists/* \
    && ln -sf /usr/bin/python3 /usr/bin/python

# Install additional development tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    dnsutils \
    iputils-ping \
    netcat-openbsd \
    htop \
    vim-tiny \
    nano \
    p7zip-full \
    && rm -rf /var/lib/apt/lists/*

USER claude

# Install Bun
RUN curl -fsSL https://bun.sh/install | bash
ENV BUN_INSTALL="/home/claude/.bun"
ENV PATH="${BUN_INSTALL}/bin:${PATH}"

# Install pnpm and agent-browser globally (as root)
USER root
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN npm install -g agent-browser

# Install Chromium and system dependencies for agent-browser
RUN agent-browser install --with-deps

USER claude

# Install useful Python packages
RUN pip3 install --user --break-system-packages \
    httpie \
    rich \
    pyyaml

ENV PATH="/home/claude/.local/bin:${PATH}"

# Verify installations
RUN echo "=== Installed versions ===" \
    && node --version \
    && npm --version \
    && bun --version \
    && pnpm --version \
    && python --version \
    && docker --version \
    && gh --version \
    && echo "agent-browser: $(agent-browser --version 2>/dev/null || echo 'installed')"

WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
