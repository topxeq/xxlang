#!/bin/bash
# benchmarks/run.sh
# Run all benchmarks and compare results

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
RESULTS_DIR="$SCRIPT_DIR/results"

echo "========================================"
echo "  xxlang Performance Benchmark Suite"
echo "========================================"
echo ""

# Create results directory
mkdir -p "$RESULTS_DIR"

# Build xxlang first
echo "Building xxlang..."
cd "$PROJECT_ROOT"
go build -o "$SCRIPT_DIR/xxlang" ./cmd/xxlang
echo ""

# Function to run command with timeout
run_with_timeout() {
    local timeout_sec=$1
    shift
    timeout "$timeout_sec" "$@"
}

echo "=== 1. Go Baseline Benchmarks ==="
echo ""

cd "$SCRIPT_DIR/go"
go run fib_test.go
echo ""

echo "=== 2. Python Benchmarks ==="
echo ""
if command -v python3 &> /dev/null; then
    python3 "$SCRIPT_DIR/python/fib.py"
else
    echo "Python3 not found, skipping"
fi
echo ""

echo "=== 3. xxlang Benchmarks ==="
echo ""

# Run Fibonacci benchmarks
echo "--- Fibonacci ---"
for n in 10 20 30 35; do
    echo -n "fib($n): "
    start=$(date +%s%N)
    if [ $n -eq 35 ]; then
        timeout=120
    elif [ $n -eq 30 ]; then
        timeout=30
    else
        timeout=10
    fi

    # Create temp script for specific n
    cat > /tmp/fib_bench.xxl << EOF
func fib(n) {
    if (n <= 1) {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}
print(fib($n))
EOF

    if run_with_timeout $timeout "$SCRIPT_DIR/xxlang" run /tmp/fib_bench.xxl 2>/dev/null; then
        end=$(date +%s%N)
        elapsed_ms=$(( (end - start) / 1000000 ))
        echo "  (${elapsed_ms} ms)"
    else
        echo "  TIMEOUT or ERROR"
    fi
done
echo ""

# Run Loop benchmark
echo "--- Loop Performance ---"
echo -n "loopSum(10000): "
start=$(date +%s%N)
"$SCRIPT_DIR/xxlang" run "$SCRIPT_DIR/xxlang/loop.xxl" 2>/dev/null
end=$(date +%s%N)
elapsed_ms=$(( (end - start) / 1000000 ))
echo "  (${elapsed_ms} ms)"
echo ""

# Run Array benchmark
echo "--- Array Operations ---"
echo -n "arraySum(1000): "
start=$(date +%s%N)
"$SCRIPT_DIR/xxlang" run "$SCRIPT_DIR/xxlang/array.xxl" 2>/dev/null
end=$(date +%s%N)
elapsed_ms=$(( (end - start) / 1000000 ))
echo "  (${elapsed_ms} ms)"
echo ""

# Run Function Calls benchmark
echo "--- Function Call Overhead ---"
echo -n "functionCalls(10000): "
start=$(date +%s%N)
"$SCRIPT_DIR/xxlang" run "$SCRIPT_DIR/xxlang/function_calls.xxl" 2>/dev/null
end=$(date +%s%N)
elapsed_ms=$(( (end - start) / 1000000 ))
echo "  (${elapsed_ms} ms)"
echo ""

echo "========================================"
echo "  Benchmark Complete"
echo "========================================"

# Cleanup
rm -f /tmp/fib_bench.xxl
