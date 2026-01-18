# cc-sandbox:bun-full
# Full-featured development environment

FROM ghcr.io/luwojtaszek/cc-sandbox:docker

LABEL org.opencontainers.image.source="https://github.com/luwojtaszek/cc-sandbox"
LABEL org.opencontainers.image.description="Claude Code Sandbox - Full Development Environment"
LABEL org.opencontainers.image.licenses="MIT"

# Install Python 3, pip, Playwright dependencies, and additional development tools (already root from docker image)
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-pip \
    dnsutils \
    iputils-ping \
    netcat-openbsd \
    htop \
    vim-tiny \
    nano \
    p7zip-full \
    # Playwright/Chromium dependencies for Ubuntu 24.04
    libglib2.0-0t64 \
    libnspr4 \
    libnss3 \
    libdbus-1-3 \
    libatk1.0-0t64 \
    libatk-bridge2.0-0t64 \
    libcups2t64 \
    libxcb1 \
    libxkbcommon0 \
    libatspi2.0-0t64 \
    libx11-6 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libcairo2 \
    libpango-1.0-0 \
    libasound2t64 \
    && rm -rf /var/lib/apt/lists/* \
    && ln -sf /usr/bin/python3 /usr/bin/python

# Install Bun for claude user
USER claude
RUN curl -fsSL https://bun.sh/install | bash
USER root

# Install Bun for root user
RUN curl -fsSL https://bun.sh/install | bash

# Set Bun paths for both users
ENV BUN_INSTALL_CLAUDE="/home/claude/.bun"
ENV BUN_INSTALL_ROOT="/root/.bun"
ENV PATH="${BUN_INSTALL_ROOT}/bin:${BUN_INSTALL_CLAUDE}/bin:${PATH}"

# Set shared Playwright browsers path for all users
ENV PLAYWRIGHT_BROWSERS_PATH=/opt/ms-playwright

# Install pnpm, agent-browser, and Chromium with dependencies
RUN corepack enable && corepack prepare pnpm@latest --activate \
    && npm install -g agent-browser \
    && npx playwright install-deps chromium \
    && agent-browser install --with-deps \
    && chmod -R 755 /opt/ms-playwright \
    && npm cache clean --force \
    && echo 'export PLAYWRIGHT_BROWSERS_PATH=/opt/ms-playwright' >> /etc/profile.d/playwright.sh \
    && echo 'export PLAYWRIGHT_BROWSERS_PATH=/opt/ms-playwright' >> /etc/bash.bashrc

# Install useful Python packages for claude user
USER claude
RUN pip3 install --user --break-system-packages --no-cache-dir \
    httpie \
    rich \
    pyyaml
USER root

# Install useful Python packages for root user
RUN pip3 install --user --break-system-packages --no-cache-dir \
    httpie \
    rich \
    pyyaml

ENV PATH="/root/.local/bin:/home/claude/.local/bin:${PATH}"

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

# NOTE: No final USER directive! Container starts as root.
# Entrypoint handles privilege dropping based on CC_SANDBOX_USER env var.

WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
