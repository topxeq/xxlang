#!/bin/bash
# Build script for xxlang
# Produces 'xx' executable on Linux/macOS, 'xx.exe' on Windows

set -e

# Determine output name based on OS
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    OUTPUT="xx.exe"
else
    OUTPUT="xx"
fi

# Allow overriding output name
if [[ -n "$1" ]]; then
    OUTPUT="$1"
fi

echo "Building $OUTPUT..."

# Build with optimizations
go build -ldflags="-s -w" -o "$OUTPUT" ./cmd/xxlang

echo "Build complete: $OUTPUT"
echo "Version: $(./$OUTPUT version)"
