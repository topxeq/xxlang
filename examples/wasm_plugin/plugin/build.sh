#!/bin/bash
# build.sh - Build WASM plugin from C source
#
# Usage: ./build.sh [source.c]
#
# Requirements:
#   clang with wasm32 target support
#
# On Ubuntu: apt install clang lld

cd "$(dirname "$0")"

SOURCE="${1:-fib.c}"
OUTPUT="${SOURCE%.c}.wasm"

echo "Building $SOURCE -> $OUTPUT"

clang -o "$OUTPUT" --target=wasm32 -O2 "$SOURCE" \
    -nostdlib -nostartfiles \
    -Wl,--no-entry -Wl,--export-all

if [ $? -eq 0 ]; then
    echo "✓ Build successful: $OUTPUT"
    echo ""
    echo "Exports:"
    wasm-objdump -x "$OUTPUT" 2>/dev/null | grep "func.*<" | head -10 || true
else
    echo "✗ Build failed"
    exit 1
fi
