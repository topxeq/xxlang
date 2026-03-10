#!/usr/bin/env python3
# benchmarks/python/fib.py
# Python baseline benchmarks for comparison

import time
import sys

def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

def fib_iter(n):
    if n <= 1:
        return n
    a, b = 0, 1
    for _ in range(2, n + 1):
        a, b = b, a + b
    return b

def loop_sum(n):
    total = 0
    for i in range(n):
        total += i
    return total

def array_sum(n):
    arr = list(range(n))
    return sum(arr)

def run_benchmark(name, func, iterations=3):
    start = time.perf_counter()
    for _ in range(iterations):
        func()
    elapsed = (time.perf_counter() - start) / iterations
    return elapsed

def main():
    print("=== Python Performance Benchmarks ===\n")

    benchmarks = [
        ("fib(10)", lambda: fib(10)),
        ("fib(20)", lambda: fib(20)),
        ("fib(30)", lambda: fib(30)),
        ("fib(35)", lambda: fib(35)),
        ("fibIter(35)", lambda: fib_iter(35)),
        ("loopSum(10000)", lambda: loop_sum(10000)),
        ("arraySum(1000)", lambda: array_sum(1000)),
    ]

    for name, func in benchmarks:
        elapsed = run_benchmark(name, func)
        print(f"{name:20s}: {elapsed*1000:.2f} ms")

if __name__ == "__main__":
    main()
