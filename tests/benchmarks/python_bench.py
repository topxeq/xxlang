#!/usr/bin/env python3
"""
Python performance benchmarks for comparison with xxlang
Run with: python3 python_bench.py
"""

import time
import statistics

def benchmark(func, name, iterations=1000):
    """Run a benchmark function multiple times and return stats"""
    times = []
    for _ in range(iterations):
        start = time.perf_counter_ns()
        func()
        end = time.perf_counter_ns()
        times.append(end - start)

    avg_ns = statistics.mean(times)
    min_ns = min(times)
    max_ns = max(times)
    print(f"{name:35} {avg_ns/1000:10.1f} µs/op ({min_ns/1000:.1f} - {max_ns/1000:.1f})")
    return avg_ns

# Fibonacci recursive
def fib(n):
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)

def bench_fib15():
    return fib(15)

def bench_fib20():
    return fib(20)

# Fibonacci iterative
def fib_iter(n):
    a, b = 0, 1
    for _ in range(n):
        a, b = b, a + b
    return a

def bench_fib_iterative():
    return fib_iter(30)

# Factorial
def fact(n):
    if n <= 1:
        return 1
    return n * fact(n - 1)

def bench_factorial():
    return fact(12)

# For loop
def bench_for_loop_100():
    s = 0
    for i in range(100):
        s += i

def bench_for_loop_1000():
    s = 0
    for i in range(1000):
        s += i

# Function calls
def add(a, b):
    return a + b

def mul(a, b):
    return a * b

def bench_function_calls():
    s = 0
    for i in range(100):
        s = add(s, i)
        s = mul(s, 1)

# Arithmetic
def bench_arithmetic():
    a = 10
    b = 20
    c = a + b
    d = c * 2
    e = d - 5
    f = e / 3
    return f

# Intensive arithmetic
def bench_intensive_arithmetic():
    x = 0
    for i in range(1000):
        x = x + i * 2 - i // 2
        x = x % 1000
    return x

# Comparisons
def bench_comparisons():
    a = 10
    b = 20
    c = 0
    for _ in range(100):
        if a < b:
            c += 1
        if a > b:
            c += 1
        if a == b:
            c += 1
        if a != b:
            c += 1

# While loop
def bench_while_loop():
    i = 0
    s = 0
    while i < 100:
        s += i
        i += 1

# Nested expressions
def bench_nested_expressions():
    a = 1 + 2 * 3 - 4 // 2
    b = (1 + 2) * (3 - 4) // 2
    c = a + b * 2 - a // b
    d = (a + b) * (c - a) // (b + 1)
    return d

# Prime check
def is_prime(n):
    if n < 2:
        return False
    i = 2
    while i * i <= n:
        if n % i == 0:
            return False
        i += 1
    return True

def bench_prime_check():
    for j in range(100):
        is_prime(j)

# Bubble sort
def bubble_sort(arr):
    n = len(arr)
    for i in range(n - 1):
        for j in range(n - i - 1):
            if arr[j] > arr[j + 1]:
                arr[j], arr[j + 1] = arr[j + 1], arr[j]

def bench_bubble_sort():
    arr = [5, 3, 8, 4, 2, 7, 1, 10, 6, 9]
    bubble_sort(arr)

if __name__ == "__main__":
    print("=" * 60)
    print("Python Performance Benchmarks")
    print("=" * 60)

    # Run benchmarks with appropriate iterations
    benchmarks = [
        (bench_fib15, "Fibonacci15", 100),
        (bench_fib20, "Fibonacci20", 20),
        (bench_fib_iterative, "FibonacciIterative", 1000),
        (bench_factorial, "Factorial", 1000),
        (bench_for_loop_100, "ForLoop100", 10000),
        (bench_for_loop_1000, "ForLoop1000", 5000),
        (bench_function_calls, "FunctionCalls", 5000),
        (bench_arithmetic, "Arithmetic", 50000),
        (bench_intensive_arithmetic, "IntensiveArithmetic", 1000),
        (bench_comparisons, "Comparisons", 5000),
        (bench_while_loop, "WhileLoop", 10000),
        (bench_nested_expressions, "NestedExpressions", 50000),
        (bench_prime_check, "PrimeCheck100", 1000),
        (bench_bubble_sort, "BubbleSort10", 10000),
    ]

    results = {}
    for func, name, iters in benchmarks:
        results[name] = benchmark(func, name, iters)

    print()
    print("=" * 60)
    print("Summary (µs/op)")
    print("=" * 60)
    for name, ns in sorted(results.items(), key=lambda x: -x[1]):
        print(f"  {name:30} {ns/1000:10.1f}")
