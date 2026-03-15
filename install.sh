#!/bin/bash
#
# Xxlang Installation Script
# Downloads and installs the latest version of Xxlang from GitHub Releases
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
#   or
#   wget -qO- https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
REPO="topxeq/xxlang"
BINARY_NAME="xxl"
INSTALL_DIR="/usr/local/bin"

# Print functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        CYGWIN*)    echo "windows" ;;
        MINGW*)     echo "windows" ;;
        MSYS*)      echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        arm*)           echo "arm" ;;
        *)              echo "unknown" ;;
    esac
}

# Get latest release version from GitHub API
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        # Fallback: try to get version from the release page HTML
        version=$(curl -fsSL "https://github.com/${REPO}/releases/latest" 2>/dev/null | grep -oP 'tag/v\K[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    fi

    if [ -z "$version" ]; then
        error "Failed to get latest version from GitHub"
        exit 1
    fi

    echo "$version"
}

# Download file with progress
download_file() {
    local url="$1"
    local output="$2"

    if command -v curl &> /dev/null; then
        curl -fSL --progress-bar -o "$output" "$url"
    elif command -v wget &> /dev/null; then
        wget -q --show-progress -O "$output" "$url"
    else
        error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
}

# Main installation
main() {
    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}    Xxlang Installation Script       ${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""

    # Check for required tools
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        error "Either curl or wget is required. Please install one of them."
        exit 1
    fi

    # Detect OS and architecture
    OS=$(detect_os)
    ARCH=$(detect_arch)

    if [ "$OS" = "unknown" ]; then
        error "Unsupported operating system: $(uname -s)"
        exit 1
    fi

    if [ "$ARCH" = "unknown" ]; then
        error "Unsupported architecture: $(uname -m)"
        exit 1
    fi

    info "Detected OS: $OS"
    info "Detected Architecture: $ARCH"

    # Get latest version
    info "Fetching latest version..."
    VERSION=$(get_latest_version)
    info "Latest version: $VERSION"

    # Build download URL
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="xxl.exe"
        ASSET_NAME="xxlang-${VERSION}-${OS}-${ARCH}.exe"
    else
        BINARY_NAME="xxl"
        ASSET_NAME="xxlang-${VERSION}-${OS}-${ARCH}"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ASSET_NAME}"

    info "Download URL: $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${BINARY_NAME}"

    # Download binary
    info "Downloading Xxlang v${VERSION}..."
    if ! download_file "$DOWNLOAD_URL" "$TMP_FILE"; then
        error "Failed to download Xxlang"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Make executable
    chmod +x "$TMP_FILE"

    # Determine install directory
    if [ "$OS" = "windows" ]; then
        # For Windows (Git Bash, etc.)
        if [ -d "$HOME/bin" ]; then
            INSTALL_DIR="$HOME/bin"
        else
            INSTALL_DIR="$HOME"
        fi
    else
        # For Linux/macOS
        # Check if we can write to /usr/local/bin
        if [ -w "/usr/local/bin" ] || [ "$(id -u)" = "0" ]; then
            INSTALL_DIR="/usr/local/bin"
        elif [ -d "$HOME/.local/bin" ]; then
            INSTALL_DIR="$HOME/.local/bin"
        elif [ -d "$HOME/bin" ]; then
            INSTALL_DIR="$HOME/bin"
        else
            # Create ~/.local/bin
            mkdir -p "$HOME/.local/bin"
            INSTALL_DIR="$HOME/.local/bin"

            # Add to PATH if not already there
            if ! echo "$PATH" | grep -q "$HOME/.local/bin"; then
                warn "Adding $HOME/.local/bin to PATH..."
                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
                export PATH="$HOME/.local/bin:$PATH"
            fi
        fi
    fi

    # Check if already installed
    INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"
    if [ -f "$INSTALL_PATH" ]; then
        info "Removing existing installation..."
        rm -f "$INSTALL_PATH"
    fi

    # Install
    info "Installing to $INSTALL_PATH..."
    mv "$TMP_FILE" "$INSTALL_PATH"

    # Cleanup
    rm -rf "$TMP_DIR"

    # Verify installation
    if [ -f "$INSTALL_PATH" ]; then
        success "Xxlang v${VERSION} installed successfully!"
        echo ""
        "$INSTALL_PATH" version
        echo ""
        success "Installation path: $INSTALL_PATH"

        # Check if in PATH
        if ! command -v xxl &> /dev/null; then
            echo ""
            warn "xxl is not in your PATH. Add the following to your shell profile:"
            echo ""
            echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
            echo ""
            if [ -n "$BASH_VERSION" ]; then
                echo "Then run: source ~/.bashrc"
            elif [ -n "$ZSH_VERSION" ]; then
                echo "Then run: source ~/.zshrc"
            fi
        fi

        echo ""
        echo -e "${GREEN}Quick Start:${NC}"
        echo "    xxl                    # Start REPL"
        echo "    xxl run script.xxl     # Run a script"
        echo "    xxl update             # Update to latest version"
        echo "    xxl help               # Show help"
        echo ""
    else
        error "Installation failed"
        exit 1
    fi
}

# Run main function
main "$@"
