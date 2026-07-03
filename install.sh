#!/bin/bash
#
# Xxlang Installation / Update Script
# Downloads and installs the latest version of Xxlang from GitHub Releases.
# If Xxlang is already installed, compares versions and skips the download
# when the installed version matches the latest release (use --force to
# reinstall anyway).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
#   or
#   wget -qO- https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash
#   or, to force reinstall:
#   curl -fsSL https://raw.githubusercontent.com/topxeq/xxlang/master/install.sh | bash -s -- --force
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
FORCE=0

# Parse flags
for arg in "$@"; do
    case "$arg" in
        --force|-f)
            FORCE=1
            ;;
        --help|-h)
            echo "Usage: $0 [--force]"
            echo "  --force  Reinstall even if the installed version is up to date."
            exit 0
            ;;
    esac
done

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

# Compare two semantic version strings of the form X.Y.Z.
# Returns 0 if equal, 1 if $1 < $2, 2 if $1 > $2.
version_compare() {
    local a="$1" b="$2"
    if [ "$a" = "$b" ]; then
        return 0
    fi
    # Split on '.' and read into three positional parameters. Using `read`
    # with IFS=. avoids the word-splitting pitfall of `set --` under a
    # non-default IFS, and works on bash 3+ (macOS) and bash 4+ alike.
    local a1 a2 a3 b1 b2 b3
    IFS=. read -r a1 a2 a3 <<< "$a"
    IFS=. read -r b1 b2 b3 <<< "$b"
    # Empty segments default to 0
    a1="${a1:-0}"; a2="${a2:-0}"; a3="${a3:-0}"
    b1="${b1:-0}"; b2="${b2:-0}"; b3="${b3:-0}"
    if [ "$a1" -lt "$b1" ]; then return 1; fi
    if [ "$a1" -gt "$b1" ]; then return 2; fi
    if [ "$a2" -lt "$b2" ]; then return 1; fi
    if [ "$a2" -gt "$b2" ]; then return 2; fi
    if [ "$a3" -lt "$b3" ]; then return 1; fi
    if [ "$a3" -gt "$b3" ]; then return 2; fi
    return 0
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

# Get the version of the currently installed xxl binary, if any.
# Prints the version string (e.g. "0.9.10") on success, empty on failure.
get_installed_version() {
    local bin="$1"
    if [ -z "$bin" ] || [ ! -x "$bin" ]; then
        return 0
    fi
    # `xxl version` prints "Xxlang v0.9.10". Extract the part after "v".
    "$bin" version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 | sed -E 's/^v//'
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

# Extract tar.gz archive
extract_targz() {
    local archive="$1"
    local dest="$2"

    if command -v tar &> /dev/null; then
        tar -xzf "$archive" -C "$dest"
    else
        error "tar is required for extraction. Please install it."
        exit 1
    fi
}

# Extract zip archive
extract_zip() {
    local archive="$1"
    local dest="$2"

    if command -v unzip &> /dev/null; then
        unzip -o -q "$archive" -d "$dest"
    elif command -v 7z &> /dev/null; then
        7z x -y -o"$dest" "$archive" > /dev/null
    else
        error "unzip or 7z is required for extraction. Please install it."
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

    # Determine binary name and install directory first, so we can check
    # the currently installed version before hitting the network.
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="xxl.exe"
    else
        BINARY_NAME="xxl"
    fi

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

    INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"

    # Get latest version
    info "Fetching latest version..."
    VERSION=$(get_latest_version)
    info "Latest version: $VERSION"

    # Check the currently installed version (if any) and skip the download
    # when it already matches the latest release.
    INSTALLED_VERSION=$(get_installed_version "$INSTALL_PATH")
    # Also try `xxl` from PATH in case INSTALL_PATH doesn't point at it.
    if [ -z "$INSTALLED_VERSION" ] && command -v xxl &> /dev/null; then
        INSTALLED_VERSION=$(get_installed_version "$(command -v xxl)")
    fi

    if [ -n "$INSTALLED_VERSION" ]; then
        info "Installed version: $INSTALLED_VERSION"
        if [ "$FORCE" -eq 0 ]; then
            version_compare "$INSTALLED_VERSION" "$VERSION"
            cmp=$?
            if [ "$cmp" -eq 0 ]; then
                success "Already up to date (v${INSTALLED_VERSION}). Nothing to do."
                success "Use --force to reinstall."
                exit 0
            elif [ "$cmp" -eq 2 ]; then
                # Installed is newer than the latest release tag — happens for
                # local builds ahead of a release. Don't downgrade.
                warn "Installed version (${INSTALLED_VERSION}) is newer than the latest release (${VERSION}). Not downgrading."
                exit 0
            fi
            info "Update available: ${INSTALLED_VERSION} -> ${VERSION}"
        else
            warn "Force reinstall requested — ignoring installed version ${INSTALLED_VERSION}."
        fi
    else
        info "No previous installation detected — performing fresh install."
    fi

    # Build download URL - using compressed archives
    # Format: xxlang-{os}-{arch}.tar.gz (Linux/macOS) or xxlang-{os}-{arch}.zip (Windows)
    if [ "$OS" = "windows" ]; then
        ARCHIVE_NAME="xxlang-${OS}-${ARCH}.zip"
    else
        ARCHIVE_NAME="xxlang-${OS}-${ARCH}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE_NAME}"

    info "Download URL: $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    ARCHIVE_FILE="${TMP_DIR}/${ARCHIVE_NAME}"

    # Download archive
    info "Downloading Xxlang v${VERSION}..."
    if ! download_file "$DOWNLOAD_URL" "$ARCHIVE_FILE"; then
        error "Failed to download Xxlang"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Extract archive
    info "Extracting..."
    if [ "$OS" = "windows" ]; then
        extract_zip "$ARCHIVE_FILE" "$TMP_DIR"
    else
        extract_targz "$ARCHIVE_FILE" "$TMP_DIR"
    fi

    # Find the extracted binary
    EXTRACTED_BINARY="${TMP_DIR}/${BINARY_NAME}"
    if [ ! -f "$EXTRACTED_BINARY" ]; then
        error "Binary not found in archive: $BINARY_NAME"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Make executable
    chmod +x "$EXTRACTED_BINARY"

    # Replace the existing binary. On Unix we can overwrite directly via
    # rename. On Windows (Git Bash / MSYS) the running exe cannot be removed
    # or overwritten while it is mapped; renaming it aside first matches the
    # strategy used by `xxl update` (v0.9.9+) and lets the new file take its
    # place.
    if [ -f "$INSTALL_PATH" ]; then
        info "Replacing existing installation..."
        if [ "$OS" = "windows" ]; then
            OLD_PATH="${INSTALL_PATH}.old"
            rm -f "$OLD_PATH"
            mv "$INSTALL_PATH" "$OLD_PATH" 2>/dev/null || true
            mv "$EXTRACTED_BINARY" "$INSTALL_PATH"
            rm -f "$OLD_PATH" 2>/dev/null || true
        else
            # On Unix, rename atomically replaces the file (even if running).
            mv "$EXTRACTED_BINARY" "$INSTALL_PATH"
        fi
    else
        info "Installing to $INSTALL_PATH..."
        mv "$EXTRACTED_BINARY" "$INSTALL_PATH"
    fi

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
