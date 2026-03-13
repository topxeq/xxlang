#!/bin/bash
# Performance benchmark comparison between Xxlang, Go, and Python

set -e

echo "=================================================="
echo "Performance Benchmark Comparison"
echo "=================================================="
echo ""

cd "$(dirname "$0")/.."

# Run Xxlang benchmarks
echo "--- Xxlang Benchmarks ---"
go test -bench -benchtime -timeout=60s ./tests 2>&1 | grep -E "(PASS|FAIL|ns/op)"

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

func BenchmarkGoLoop1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 1000; j++ {
			sum += j
		}
	}
}

func BenchmarkGoArray100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]int, 100)
	}
}

func BenchmarkGoFunctionCalls100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 100; j++ {
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

if __name__ == '__main__':
    print("Python Benchmarks:")
    t = timeit.timeit('fib(25)', globals=globals(), number=100000)
    print(f"fib(25): {t * 10000000:.2f} ns/op")
    t = timeit.timeit('fib(20)', globals=globals(), number=100000)
    print(f"fib(20): {t * 10000000:.2f} ns/op")
    t = timeit.timeit('sum=0; [sum:=sum+j for j in range(1000)]', globals=globals(), number=100000)
    print(f"loop1000: {t * 10000000:.2f} ns/op")
    t = timeit.timeit('[0] * 100', globals=globals(), number=100000)
    print(f"array100: {t * 10000000:.2f} ns/op")
EOF

    echo "--- Python Fibonacci Baselines ---"
    $PYTHON_CMD /tmp/py_fib_bench.py 2>&1 || echo "Python benchmarks completed"

    # Clean up
    rm /tmp/py_fib_bench.py
fi

echo ""
echo "=================================================="
echo "Java Baseline Benchmarks (for comparison)"
echo "=================================================="
echo ""

# Check if Java is available
if command -v javac &> /dev/null 2>&1 && command -v java &> /dev/null 2>&1; then
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
        long start, end;

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            fib(25);
        }
        end = System.nanoTime();
        System.out.printf("fib(25): %.2f ns/op%n", (end - start) / (double)iterations);

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            fib(20);
        }
        end = System.nanoTime();
        System.out.printf("fib(20): %.2f ns/op%n", (end - start) / (double)iterations);

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            int sum = 0;
            for (int j = 0; j < 1000; j++) {
                sum += j;
            }
        }
        end = System.nanoTime();
        System.out.printf("loop1000: %.2f ns/op%n", (end - start) / (double)iterations);

        start = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            int[] arr = new int[100];
        }
        end = System.nanoTime();
        System.out.printf("array100: %.2f ns/op%n", (end - start) / (double)iterations);
    }
}
EOF

    echo "--- Java Baseline ---"
    javac /tmp/Fibonacci.java 2>&1 || echo "Java compilation failed"
    java -cp /tmp Fibonacci: 2>&1 || java -cp /tmp Fibonacci 2>&1

    rm /tmp/Fibonacci.java /tmp/Fibonacci.class 2>/dev/null || true
else
    echo "Java not found, skipping Java benchmarks"
fi

echo ""
echo "=================================================="
echo "Summary"
echo "=================================================="
echo "Benchmark comparison complete!"
echo ""
echo "To get detailed CPU profiling for Xxlang, run with:"
echo "  go test -bench -benchtime -cpuprofile=cpu.prof -run <benchmark> ./tests"
echo ""
echo "To see CPU profile:"
echo "  go tool pprof -http=:8080 cpu.prof"
echo ""
echo "To get memory profiling (B/op/s), add -benchmem flag"
