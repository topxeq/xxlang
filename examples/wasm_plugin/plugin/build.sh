#!/bin/bash
# build.sh - Build WASM plugin from C, Go (TinyGo), Zig, or Rust source
#
# Usage: ./build.sh [source]
#
# Supported sources:
#   fib.c   - C source (requires clang with wasm32 target)
#   fib.go  - Go source (requires TinyGo)
#   fib.zig - Zig source (requires Zig)
#   fib.rs  - Rust source (requires Rust with wasm32-unknown-unknown target)
#
# Requirements:
#   clang with wasm32 target: apt install clang lld
#   TinyGo: https://tinygo.org/getting-started/install/
#   Zig: https://ziglang.org/learn/getting-started/
#   Rust: rustup target add wasm32-unknown-unknown

cd "$(dirname "$0")"

SOURCE="${1:-fib.c}"
BASENAME="${SOURCE%.*}"
EXT="${SOURCE##*.}"

case "$EXT" in
    c)
        OUTPUT="${BASENAME}.wasm"
        echo "Building C source: $SOURCE -> $OUTPUT"
        clang -o "$OUTPUT" --target=wasm32 -O2 "$SOURCE" \
            -nostdlib -nostartfiles \
            -Wl,--no-entry -Wl,--export-all
        ;;
    go)
        OUTPUT="${BASENAME}.wasm"
        echo "Building Go source (TinyGo): $SOURCE -> $OUTPUT"
        if ! command -v tinygo &> /dev/null; then
            echo "Error: TinyGo not installed"
            echo "Install from: https://tinygo.org/getting-started/install/"
            exit 1
        fi
        tinygo build -o "$OUTPUT" -target=wasi "$SOURCE"
        ;;
    zig)
        OUTPUT="${BASENAME}.wasm"
        echo "Building Zig source: $SOURCE -> $OUTPUT"
        if ! command -v zig &> /dev/null; then
            echo "Error: Zig not installed"
            echo "Install from: https://ziglang.org/learn/getting-started/"
            exit 1
        fi
        zig build-exe "$SOURCE" -target wasm32-freestanding -O ReleaseSmall -fno-entry -rdynamic
        ;;
    rs)
        OUTPUT="${BASENAME}.wasm"
        echo "Building Rust source: $SOURCE -> $OUTPUT"
        if ! command -v rustc &> /dev/null; then
            echo "Error: Rust not installed"
            echo "Install from: https://rustup.rs/"
            exit 1
        fi
        rustc --target wasm32-unknown-unknown -O --crate-type cdylib -o "$OUTPUT" "$SOURCE"
        ;;
    ts)
        OUTPUT="${BASENAME}.wasm"
        echo "Building AssemblyScript source: $SOURCE -> $OUTPUT"
        if ! command -v asc &> /dev/null; then
            echo "Error: AssemblyScript not installed"
            echo "Install: npm install -g assemblyscript"
            exit 1
        fi
        asc "$SOURCE" -o "$OUTPUT" --optimize --runtime stub --initialMemory 2
        ;;
    *)
        echo "Error: Unknown source type: $EXT"
        echo "Supported: .c, .go, .zig, .rs, .ts"
        exit 1
        ;;
esac

if [ $? -eq 0 ]; then
    echo "✓ Build successful: $OUTPUT"
    echo ""
    echo "Exports:"
    wasm-objdump -x "$OUTPUT" 2>/dev/null | grep "func.*<" | head -10 || true
else
    echo "✗ Build failed"
    exit 1
fi
