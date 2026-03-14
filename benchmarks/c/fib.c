// benchmarks/c/fib.c
// C baseline benchmarks for comparison
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <sys/time.h>

// Recursive Fibonacci (naive implementation for fair comparison)
long fib(int n) {
    if (n <= 1) return n;
    return fib(n - 1) + fib(n - 2);
}

// Iterative Fibonacci (optimized)
long fib_iter(int n) {
    if (n <= 1) return n;
    long a = 0, b = 1;
    for (int i = 2; i <= n; i++) {
        long temp = a + b;
        a = b;
        b = temp;
    }
    return b;
}

// Loop sum
long loop_sum(int n) {
    long sum = 0;
    for (int i = 0; i < n; i++) {
        sum += i;
    }
    return sum;
}

// Array sum
long array_sum(int n) {
    int *arr = malloc(n * sizeof(int));
    for (int i = 0; i < n; i++) {
        arr[i] = i;
    }
    long sum = 0;
    for (int i = 0; i < n; i++) {
        sum += arr[i];
    }
    free(arr);
    return sum;
}

// Function call overhead
long add(long a, long b) {
    return a + b;
}

long func_call(int n) {
    long sum = 0;
    for (int i = 0; i < n; i++) {
        sum = add(sum, i);
    }
    return sum;
}

// Get current time in nanoseconds
long long get_ns() {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (long long)ts.tv_sec * 1000000000LL + ts.tv_nsec;
}

// Benchmark runner
void benchmark(const char *name, long (*fn)(), int iterations) {
    // Warmup
    fn();

    long long start = get_ns();
    for (int i = 0; i < iterations; i++) {
        fn();
    }
    long long end = get_ns();

    double ns_per_op = (double)(end - start) / iterations;
    printf("%-20s: %12.2f ns/op\n", name, ns_per_op);
}

int main() {
    printf("=== C Performance Benchmarks ===\n\n");

    // Warmup
    fib(10);

    // Run benchmarks
    benchmark("fib(10)", (long (*)())fib, 100000);
    benchmark("fib(20)", (long (*)())fib, 1000);
    benchmark("fib(30)", (long (*)())fib, 10);
    benchmark("fib(35)", (long (*)())fib, 3);
    benchmark("fibIter(35)", (long (*)())fib_iter, 1000000);
    benchmark("loopSum(10000)", (long (*)())loop_sum, 10000);
    benchmark("arraySum(1000)", (long (*)())array_sum, 100000);
    benchmark("funcCall(1000)", (long (*)())func_call, 100000);

    printf("\n");
    return 0;
}
