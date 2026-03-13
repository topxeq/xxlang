#!/bin/bash
# Performance benchmark comparison between Xxlang, Go, and Python

echo "=================================================="
echo "Performance Benchmark Comparison"
echo "=================================================="
echo ""

# Run Xxlang benchmarks with 1000 iterations
echo "--- Xxlang Benchmarks ---"
go test -bench -benchtime -timeout=30s -run=1000 ./tests 2>&1 | grep -E "(PASS|FAIL|ns/op)" || echo "Xxlang benchmarks completed"

echo ""
echo "=================================================="
echo "Go Baseline Benchmarks (for comparison)"
echo "=================================================="
echo ""

# Create temporary Go benchmark file
cat > /tmp/go_fib_bench.go <<'EOF'
package main

import "testing"

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func BenchmarkGoFib20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(20)
	}
}

func BenchmarkGoLoop1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 1000; j++ {
			sum += j
		}
	}
}
EOF

echo "--- Go Fibonacci Baselines ---"
go test -bench -benchtime -timeout=30s /tmp/go_fib_bench.go 2>&1 | grep -E "(PASS|FAIL|ns/op)" || echo "Go benchmarks completed"

rm /tmp/go_fib_bench.go

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
else
	echo "Python not found, skipping Python benchmarks"
	PYTHON=""
fi

if [ -n "$PYTHON" ]; then
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
def bench_loop1000():
    total = 0
    for _ in range(100000):
        for j in range(1000):
            total += j

@timeit.timeit(number=100000)
def bench_array100():
    for _ in range(100000):
        _ = [0] * 100

@timeit.timeit(number=100000)
def bench_function_calls100():
    total = 0
    for _ in range(100000):
        for j in range(1000):
            total += j

if __name__ == '__main__':
    bench_fib20()
    bench_loop1000()
    bench_array100()
    bench_function_calls100()
EOF

	echo "--- Python Fibonacci Baselines ---"
	$PYTHON /tmp/py_fib_bench.py 2>&1 | grep -E "(fib|loop|array|function)" || echo "Python benchmarks completed"

	rm /tmp/py_fib_bench.py
fi

echo ""
echo "=================================================="
echo "Java Baseline Benchmarks (for comparison)"
echo "=================================================="
echo ""

# Check if Java is available
if command -v javac &> /dev/null 2>&1; then
	# Create temporary Java benchmark file
	cat > /tmp/Fibonacci.java <<'EOF'
public class Fibonacci {
    public static long fib(int n) {
        if (n <= 1) {
            return n;
        }
        return fib(n-1) + fib(n-2);
    }

    public static void main(String[] args) {
        int iterations = 1000000;
        long start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            fib(20);
        }
        long end = System.nanoTime();
        System.out.printf("fib(20): %.2f ns/op%n", (end - start) / (double)iterations);
    }
EOF

	echo "--- Java Fibonacci Baselines ---"
	javac /tmp/Fibonacci.java 2>&1 | grep -A 1 "fib(20)" || echo "Java benchmarks completed"

	rm /tmp/Fibonacci.java /tmp/Fibonacci.class
else
	echo "Java not found, skipping Java benchmarks"

echo ""
echo "=================================================="
echo "Summary"
echo "=================================================="
echo "For detailed CPU profiling, use:"
echo "  go test -bench -benchtime -cpuprofile=cpu.prof -run <benchmark> ./tests"
echo ""
echo "To view CPU profile:"
echo "  go tool pprof -http=:8080 cpu.prof"
echo ""
echo "To get memory profiling (B/op/s), add -benchmem flag"
