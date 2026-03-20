#!/usr/bin/env python3
"""
Python benchmarks for comparison with Xxlang
Pure Python implementation (no C optimizations like NumPy)
"""

import time
import sys

def fib_recursive(n):
    """Recursive Fibonacci (naive)"""
    if n <= 1:
        return n
    return fib_recursive(n - 1) + fib_recursive(n - 2)

def fib_iterative(n):
    """Iterative Fibonacci"""
    a, b = 0, 1
    for _ in range(n):
        a, b = b, a + b
    return a

def factorial(n):
    """Factorial"""
    result = 1
    for i in range(2, n + 1):
        result *= i
    return result

def for_loop_sum(n):
    """For loop sum"""
    total = 0
    for i in range(n):
        total += i
    return total

def while_loop_sum(n):
    """While loop sum"""
    total = 0
    i = 0
    while i < n:
        total += i
        i += 1
    return total

def nested_loop(n):
    """Nested loop"""
    total = 0
    for i in range(n):
        for j in range(n):
            total += i * j
    return total

def prime_check(limit):
    """Prime number check"""
    count = 0
    for n in range(2, limit + 1):
        is_prime = True
        for i in range(2, int(n ** 0.5) + 1):
            if n % i == 0:
                is_prime = False
                break
        if is_prime:
            count += 1
    return count

def bubble_sort(arr):
    """Bubble sort"""
    n = len(arr)
    for i in range(n):
        for j in range(0, n - i - 1):
            if arr[j] > arr[j + 1]:
                arr[j], arr[j + 1] = arr[j + 1], arr[j]
    return arr

def string_concat(n):
    """String concatenation"""
    result = ""
    for i in range(n):
        result += str(i)
    return result

def array_sum(n):
    """Array sum"""
    arr = list(range(n))
    total = 0
    for x in arr:
        total += x
    return total

def arithmetic_ops(n):
    """Arithmetic operations"""
    result = 0
    for i in range(n):
        result = result + i
        result = result * 2
        result = result - i
    return result

def function_calls(n):
    """Function calls"""
    def add(a, b):
        return a + b

    def mul(a, b):
        return a * b

    result = 0
    for i in range(n):
        result = add(result, i)
        result = mul(result, 1)
    return result

def benchmark(func, *args, iterations=10):
    """Run benchmark and return average time in microseconds"""
    times = []
    for _ in range(iterations):
        start = time.perf_counter()
        result = func(*args)
        end = time.perf_counter()
        times.append((end - start) * 1_000_000)  # Convert to microseconds
    return sum(times) / len(times), result

def main():
    print("=" * 70)
    print("Python Performance Benchmarks (Pure Python, no C optimizations)")
    print(f"Python version: {sys.version}")
    print("=" * 70)

    benchmarks = [
        ("Fibonacci10 (recursive)", fib_recursive, 10, 10),
        ("Fibonacci15 (recursive)", fib_recursive, 15, 10),
        ("Fibonacci20 (recursive)", fib_recursive, 20, 5),
        ("Fibonacci35 (recursive)", fib_recursive, 35, 1),
        ("Fibonacci10 (iterative)", fib_iterative, 10, 100000),
        ("Fibonacci15 (iterative)", fib_iterative, 15, 100000),
        ("Fibonacci1000 (iterative)", fib_iterative, 1000, 10000),
        ("Factorial10", factorial, 10, 100000),
        ("Factorial20", factorial, 20, 100000),
        ("ForLoop100", for_loop_sum, 100, 100000),
        ("ForLoop1000", for_loop_sum, 1000, 10000),
        ("WhileLoop100", while_loop_sum, 100, 100000),
        ("WhileLoop1000", while_loop_sum, 1000, 10000),
        ("NestedLoop10", nested_loop, 10, 10000),
        ("PrimeCheck100", prime_check, 100, 10000),
        ("BubbleSort10", bubble_sort, list(range(10, 0, -1)), 100000),
        ("StringConcat100", string_concat, 100, 10000),
        ("ArraySum1000", array_sum, 1000, 10000),
        ("ArithmeticOps100", arithmetic_ops, 100, 100000),
        ("FunctionCalls100", function_calls, 100, 10000),
    ]

    print(f"\n{'Benchmark':<30} {'Time (µs)':<15} {'Result':<20}")
    print("-" * 70)

    for name, func, args, iterations in benchmarks:
        if isinstance(args, list):
            # For bubble sort, need to copy the list each time
            import copy
            def bench_func():
                arr = copy.copy(args)
                return func(arr)
            time_us, result = benchmark(bench_func, iterations=iterations)
        else:
            time_us, result = benchmark(func, args, iterations=iterations)

        result_str = str(result)[:20]
        print(f"{name:<30} {time_us:<15.2f} {result_str:<20}")

    print("=" * 70)

if __name__ == "__main__":
    main()
