#!/bin/bash
set -e

# cc-sandbox installer
# Usage: curl -fsSL https://raw.githubusercontent.com/luwojtaszek/cc-sandbox/main/install.sh | bash

REPO="luwojtaszek/cc-sandbox"
INSTALL_DIR="${CC_SANDBOX_INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="cc-sandbox"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub API
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "Failed to fetch latest version. Check your internet connection or try again later."
    fi
    echo "$version"
}

# Main installation
main() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     cc-sandbox installer             ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""

    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected platform: ${OS}/${ARCH}"

    # Get version (use provided or fetch latest)
    VERSION="${CC_SANDBOX_VERSION:-$(get_latest_version)}"
    info "Installing version: ${VERSION}"

    # Construct download URL
    BINARY_SUFFIX=""
    [ "$OS" = "windows" ] && BINARY_SUFFIX=".exe"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}${BINARY_SUFFIX}"

    info "Downloading from: ${DOWNLOAD_URL}"

    # Create install directory
    mkdir -p "$INSTALL_DIR"

    # Download binary
    TEMP_FILE=$(mktemp)
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
        rm -f "$TEMP_FILE"
        error "Failed to download binary. Check if version ${VERSION} exists."
    fi

    # Install binary
    INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}${BINARY_SUFFIX}"
    mv "$TEMP_FILE" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"

    success "Installed to: ${INSTALL_PATH}"

    # Verify installation
    if "$INSTALL_PATH" version &>/dev/null; then
        success "Installation verified!"
        echo ""
        "$INSTALL_PATH" version
    else
        warn "Binary installed but verification failed"
    fi

    echo ""

    # Check if install dir is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "Add it to your shell configuration:"
        echo ""
        case "$SHELL" in
            */zsh)
                echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
                echo "  source ~/.zshrc"
                ;;
            */bash)
                echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
                echo "  source ~/.bashrc"
                ;;
            *)
                echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
                ;;
        esac
        echo ""
    fi

    success "Installation complete!"
    echo ""
    echo "Usage:"
    echo "  cc-sandbox claude              # Start Claude Code"
    echo "  cc-sandbox claude -p \"prompt\"  # One-shot prompt"
    echo "  cc-sandbox --help              # Show all options"
    echo ""
}

main "$@"
