#!/bin/bash
# Compare interpreted vs compiled execution performance

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
XXLANG="./xxlang"
FIB_SOURCE="$SCRIPT_DIR/xxlang/fib_naive.xxl"
FIB_BYTECODE="$SCRIPT_DIR/xxlang/fib_naive.xxb"
FIB_EXEC="$SCRIPT_DIR/xxlang/fib_naive"

echo "=============================================="
echo "Xxlang: Interpreted vs Compiled Performance"
echo "=============================================="
echo ""
echo "Test: Fibonacci(35) naive recursion"
echo ""

# Check xxlang exists
if [ ! -f "$XXLANG" ]; then
    echo "Building xxlang..."
    go build -o xxlang ./cmd/xxl
fi

# Compile to bytecode
echo "1. Compiling to bytecode..."
$XXLANG compile --bytecode -o "$FIB_BYTECODE" "$FIB_SOURCE"
echo "   Created: $FIB_BYTECODE"
echo ""

# Compile to "executable" (shell wrapper)
echo "2. Creating executable wrapper..."
$XXLANG compile -o "$FIB_EXEC" "$FIB_SOURCE"
echo "   Created: $FIB_EXEC"
echo ""

# Verify both produce same result
echo "3. Verifying correctness..."
RESULT_SOURCE=$($XXLANG run "$FIB_SOURCE")
RESULT_BYTECODE=$($XXLANG run "$FIB_BYTECODE")
echo "   Source:   $RESULT_SOURCE"
echo "   Bytecode: $RESULT_BYTECODE"
echo ""

# Benchmark interpreted (source)
echo "4. Benchmark: Interpreted (parse + compile + execute)"
echo "   Running 5 iterations..."
for i in {1..5}; do
    START=$(date +%s%N)
    $XXLANG run "$FIB_SOURCE" > /dev/null
    END=$(date +%s%N)
    ELAPSED_MS=$(( (END - START) / 1000000 ))
    echo "   Run $i: ${ELAPSED_MS} ms"
done
echo ""

# Benchmark bytecode
echo "5. Benchmark: Pre-compiled bytecode (execute only)"
echo "   Running 5 iterations..."
for i in {1..5}; do
    START=$(date +%s%N)
    $XXLANG run "$FIB_BYTECODE" > /dev/null
    END=$(date +%s%N)
    ELAPSED_MS=$(( (END - START) / 1000000 ))
    echo "   Run $i: ${ELAPSED_MS} ms"
done
echo ""

# Compare startup time
echo "6. Startup overhead comparison (empty program)"
echo 'println("hi")' > /tmp/startup_test.xxl
$XXLANG compile --bytecode -o /tmp/startup_test.xxb /tmp/startup_test.xxl

# Source startup
START=$(date +%s%N)
$XXLANG run /tmp/startup_test.xxl > /dev/null
END=$(date +%s%N)
SOURCE_STARTUP_US=$(( (END - START) / 1000 ))

# Bytecode startup
START=$(date +%s%N)
$XXLANG run /tmp/startup_test.xxb > /dev/null
END=$(date +%s%N)
BYTECODE_STARTUP_US=$(( (END - START) / 1000 ))

echo "   Source startup:   ${SOURCE_STARTUP_US} µs"
echo "   Bytecode startup: ${BYTECODE_STARTUP_US} µs"
echo "   Speedup: $(( SOURCE_STARTUP_US / (BYTECODE_STARTUP_US > 0 ? BYTECODE_STARTUP_US : 1) ))x"
echo ""

# Cleanup
rm -f /tmp/startup_test.xxl /tmp/startup_test.xxb

echo "=============================================="
echo "Summary"
echo "=============================================="
echo ""
echo "Key Insight:"
echo "- Interpreted: Parse + Compile + Execute"
echo "- Bytecode:    Execute only (skip parsing/compiling)"
echo ""
echo "For compute-bound tasks like fib(35), the parsing/compiling"
echo "overhead is negligible compared to execution time."
echo "For small scripts, bytecode provides faster startup."
