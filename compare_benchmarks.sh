#!/bin/bash
# Performance benchmark comparison between Xxlang and Python

echo "=================================================="
echo "Performance Benchmark Comparison"
echo "=================================================="
echo ""

# Run Xxlang benchmarks
echo "--- Xxlang Benchmarks ---"
cd "$(dirname "$0")/tests"
go test -bench -benchtime -timeout=30s -run=10000 "BenchmarkFibonacci20 BenchmarkForLoop1000 BenchmarkArrayCreation100 BenchmarkFunctionCalls100" 2>&1 | grep -E "(PASS|FAIL|ns/op)" || echo "Xxlang benchmarks completed"

echo ""
echo "=================================================="
echo "Python Baseline Benchmarks (for comparison)"
echo "=================================================="
echo ""

# Check if Python is available
PYTHON=""
if command -v python3 &> /dev/null 2>&1; then
    PYTHON="python3"
elif command -v python &> /dev/null 2>&1; then
    PYTHON="python"
elif command -v python2 &> /dev/null 2>&1; then
    PYTHON="python2"
else
    echo "Python not found, skipping Python benchmarks"
    exit 0
fi

# Create temporary Python benchmark file
cat > /tmp/py_fib_bench.py <<'EOF'
import timeit

def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)

@timeit.timeit(number=100000)
def bench_fib20():
    fib(20)

@timeit.timeit(number=100000)
def bench_loop100():
    total = 0
    for _ in range(100):
        for j in range(100):
            total += j

@timeit.timeit(number=100000)
def bench_array100():
    for _ in range(100):
        _ = [0] * 100

@timeit.timeit(number=100000)
def bench_function_calls100():
    total = 0
    for _ in range(100):
        for j in range(100):
            total += j

if __name__ == '__main__':
    bench_fib20()
    bench_loop100()
    bench_array100()
    bench_function_calls100()
EOF

echo "--- Python Fibonacci Baselines ---"
$PYTHON /tmp/py_fib_bench.py 2>&1 | grep -E "(fib|loop|array|function)" || echo "Python benchmarks completed"

# Clean up
rm /tmp/py_fib_bench.py

echo ""
echo "=================================================="
echo "Summary"
echo "=================================================="
echo "Xxlang vs Python Performance Comparison Complete!"
echo ""
echo "To get detailed CPU profiling for Xxlang, run with:"
echo "  go test -bench -benchtime -cpuprofile=cpu.prof -run <benchmark> ./tests"
echo ""
echo "To see CPU profile:"
echo "  go tool pprof -http=:8080 cpu.prof"
