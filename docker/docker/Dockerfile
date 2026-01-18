# cc-sandbox:docker
# Extends base image with Docker CLI

FROM ghcr.io/luwojtaszek/cc-sandbox:base

LABEL org.opencontainers.image.source="https://github.com/luwojtaszek/cc-sandbox"
LABEL org.opencontainers.image.description="Claude Code Sandbox - Docker Image"
LABEL org.opencontainers.image.licenses="MIT"

# Install Docker CLI and Compose plugin (already root from base)
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        docker-ce-cli \
        docker-compose-plugin \
    && rm -rf /var/lib/apt/lists/*

# Create docker group and add both users
RUN groupadd -g 999 docker || true \
    && usermod -aG docker claude \
    && usermod -aG docker root

# NOTE: No final USER directive! Container starts as root.
# Entrypoint handles privilege dropping based on CC_SANDBOX_USER env var.

WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["claude"]
