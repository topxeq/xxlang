#!/bin/bash
# Performance benchmark comparison between Xxlang, Go, Java, and Python

set -e

echo "=================================================="
echo "Performance Benchmark Comparison"
echo "=================================================="
echo ""

cd "$(dirname "$0")/tests"

# Run Xxlang benchmarks with 1000000 iterations
echo "--- Xxlang Benchmarks ---"
go test -bench -benchtime -timeout=30s -run=1000000 "BenchmarkFibonacci25 BenchmarkFibonacci20 BenchmarkForLoop1000 BenchmarkArrayCreation100 BenchmarkFunctionCalls100" 2>&1 | grep -E "(PASS|FAIL|ns/op)"

echo ""
echo "=================================================="
echo "Go Baseline Benchmarks (for comparison)"
echo "=================================================="
echo ""

# Create and run Go benchmarks
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

func BenchmarkGoFib25(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(25)
	}
}

func BenchmarkForLoop1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 1000; j++ {
			sum += j
		}
	}
}

func BenchmarkArrayCreation100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]int, 100)
	}
}

func BenchmarkFunctionCalls100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 100; j++ {
			sum += j
		}
	}
}
EOF

echo "--- Go Fibonacci Benchmarks ---"
go test -bench -benchtime -timeout=30s /tmp/go_fib_bench.go 2>&1 | grep -E "(PASS|FAIL|ns/op)" || echo "Go benchmarks failed"

rm /tmp/go_fib_bench.go

echo ""
echo "=================================================="
echo "Java Baseline Benchmarks (for comparison)"
echo "=================================================="
echo ""

# Create and run Java benchmarks
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
            fib(25);
        }
        long end = System.nanoTime();
        System.out.printf("fib(25): %.2f ns/op%n", (end - start) / (double)iterations);
    }
}

echo "--- Java Fibonacci Baseline ---"
if java -cp .:/tmp/Fibonacci.java 2>&1; then
    java -cp .:/tmp/Fibonacci.class 2>&1 || echo "Java compilation failed"
    java -cp .:/tmp/Fibonacci 2>&1
    java /tmp/Fibonacci 2>&1 | grep -A 1 "fib(25)" || echo "Java execution failed"
    java -cp .:/tmp/Fibonacci.java .:/tmp/Fibonacci.class 2>&1
    rm /tmp/Fibonacci.java /tmp/Fibonacci.class
else
    echo "Java not found or javac not configured"
fi

echo ""
echo "=================================================="
echo "Python Baseline Benchmarks (for comparison)"
echo "=================================================="
echo ""

# Check which Python is available
PYTHON_CMD=""
if command -v python3 &> /dev/null 2>&1; then
    PYTHON_CMD="python3"
elif command -v python2 &> /dev/null 2>&1; then
    PYTHON_CMD="python2"
elif command -v python &> /dev/null 2>&1; then
    PYTHON_CMD="python"
else
    echo "Python not found, skipping Python benchmarks"
    PYTHON_CMD=""
fi

if [ -n "$PYTHON_CMD" ]; then
    # Create and run Python benchmarks
    cat > /tmp/py_fib_bench.py <<'EOF'
import timeit

def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)

@timeit.timeit(number=100000)
def bench_fib25():
    fib(25)

@timeit.timeit(number=100000)
def bench_fib20():
    fib(20)

@timeit.timeit(number=100000)
def bench_loop1000():
    total = 0
    for _ in range(1000):
        for j in range(1000):
            total += j

@timeit.timeit(number=100000)
def bench_array100():
    for _ in range(1000):
        _ = [0] * 100

@timeit.timeit(number=100000)
def bench_function_calls100():
    total = 0
    for _ in range(1000):
        for j in range(1000):
            total += j

if __name__ == '__main__':
    bench_fib25()
    bench_fib20()
    bench_loop1000()
    bench_array100()
    bench_function_calls100()
EOF

    echo "--- Python Fibonacci Benchmarks ---"
    if [ -n "$PYTHON_CMD" ]; then
        $PYTHON_CMD /tmp/py_fib_bench.py 2>&1 | grep -E "(fib|loop|array|function)" || echo "Python benchmarks failed"
    fi

    rm /tmp/py_fib_bench.py
fi

echo ""
echo "=================================================="
echo "Summary"
echo "=================================================="
echo ""
echo "Benchmark comparison complete!"
echo ""
echo ""
echo "To get detailed CPU profiling for Xxlang, run with:"
echo "  go test -bench -benchtime -cpuprofile=cpu.prof -run <benchmark> ./tests"
echo ""
echo "To see CPU profile:"
echo "  go tool pprof -http=:8080 cpu.prof"
echo ""
echo "To get memory profiling (B/op/s), add -benchmem flag"
echo ""
