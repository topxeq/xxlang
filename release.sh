#!/bin/bash
#
# Release build script for xxlang
# Builds for multiple platforms and creates compressed archives
#
# Usage:
#   ./release.sh [version]
#
# The version is automatically detected from cmd/xxl/main.go if not provided
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Get version from main.go if not provided
get_version() {
    if [ -n "$1" ]; then
        echo "$1"
        return
    fi

    # Extract version from cmd/xxl/main.go
    local version
    version=$(grep -E '^\s*Version\s*=' cmd/xxl/main.go | head -1 | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        error "Could not detect version from cmd/xxl/main.go"
        exit 1
    fi

    echo "$version"
}

# Build for a specific platform
build_platform() {
    local os="$1"
    local arch="$2"
    local output_dir="$3"

    local binary_name
    if [ "$os" = "windows" ]; then
        binary_name="xxl.exe"
    else
        binary_name="xxl"
    fi

    local archive_name
    if [ "$os" = "windows" ]; then
        archive_name="xxlang-${os}-${arch}.zip"
    else
        archive_name="xxlang-${os}-${arch}.tar.gz"
    fi

    info "Building for $os/$arch..."

    # Set environment for cross-compilation
    export GOOS="$os"
    export GOARCH="$arch"
    export CGO_ENABLED=0

    # Build
    local build_dir="${output_dir}/build/${os}-${arch}"
    mkdir -p "$build_dir"

    go build -ldflags="-s -w" -o "${build_dir}/${binary_name}" ./cmd/xxl

    # Create archive
    info "Creating archive: $archive_name"

    # Use absolute path for archive
    local archive_path="$(cd "$output_dir" && pwd)/${archive_name}"

    if [ "$os" = "windows" ]; then
        # Create zip for Windows
        (cd "$build_dir" && zip -q "$archive_path" "$binary_name")
    else
        # Create tar.gz for Linux/macOS
        (cd "$build_dir" && tar -czf "$archive_path" "$binary_name")
    fi

    # Show file size
    local size
    size=$(ls -lh "${output_dir}/${archive_name}" | awk '{print $5}')
    success "Created ${archive_name} (${size})"
}

# Main
main() {
    local version
    version=$(get_version "$1")

    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}    Xxlang Release Builder           ${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""

    info "Version: $version"

    # Create output directory
    local output_dir="release-v${version}"
    mkdir -p "$output_dir"

    # Build for all supported platforms
    # Linux
    build_platform "linux" "amd64" "$output_dir"
    build_platform "linux" "arm64" "$output_dir"

    # macOS
    build_platform "darwin" "amd64" "$output_dir"
    build_platform "darwin" "arm64" "$output_dir"

    # Windows
    build_platform "windows" "amd64" "$output_dir"
    build_platform "windows" "arm64" "$output_dir"

    echo ""
    success "Release build complete!"
    echo ""
    echo "Archives created in: $output_dir/"
    echo ""
    ls -lh "$output_dir"
    echo ""
    echo "Upload these files to GitHub Releases:"
    echo "  https://github.com/topxeq/xxlang/releases/new?tag=v${version}"
    echo ""
}

main "$@"
