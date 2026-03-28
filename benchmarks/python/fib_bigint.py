#!/usr/bin/env python3
"""
Fibonacci Benchmark for Python
Python's int is arbitrary precision by default
"""
import time

def fib_int(n):
    """Recursive Fibonacci"""
    if n <= 1:
        return n
    return fib_int(n - 1) + fib_int(n - 2)

def fib_int_iter(n):
    """Iterative Fibonacci"""
    if n <= 1:
        return n
    a, b = 0, 1
    for _ in range(2, n + 1):
        a, b = b, a + b
    return b

def fib_bigint_iter(n):
    """Iterative Fibonacci (Python int is already arbitrary precision)"""
    return fib_int_iter(n)

def bit_length(n):
    """Get bit length of integer"""
    return n.bit_length()

def main():
    print("=== Python Fibonacci Benchmark ===\n")

    # Int recursive
    print("--- Int (recursive) ---")
    start = time.time() * 1000
    result = fib_int(35)
    elapsed = int(time.time() * 1000 - start)
    print(f"fib(35) = {result} | time: {elapsed} ms")

    start = time.time() * 1000
    result = fib_int(40)
    elapsed = int(time.time() * 1000 - start)
    print(f"fib(40) = {result} | time: {elapsed} ms")

    # Int iterative
    print("\n--- Int (iterative) ---")
    start = time.time() * 1000
    result = fib_int_iter(90)
    elapsed = int(time.time() * 1000 - start)
    print(f"fib(90) = {result} | time: {elapsed} ms")

    # BigInt iterative (Python int is arbitrary precision)
    print("\n--- BigInt (iterative) - Python int is arbitrary precision ---")
    start = time.time() * 1000
    result = fib_bigint_iter(100)
    elapsed = int(time.time() * 1000 - start)
    print(f"fib(100) bits: {bit_length(result)} | time: {elapsed} ms")

    start = time.time() * 1000
    result = fib_bigint_iter(1000)
    elapsed = int(time.time() * 1000 - start)
    print(f"fib(1000) bits: {bit_length(result)} | time: {elapsed} ms")

    start = time.time() * 1000
    result = fib_bigint_iter(10000)
    elapsed = int(time.time() * 1000 - start)
    print(f"fib(10000) bits: {bit_length(result)} | time: {elapsed} ms")

    start = time.time() * 1000
    result = fib_bigint_iter(100000)
    elapsed = int(time.time() * 1000 - start)
    print(f"fib(100000) bits: {bit_length(result)} | time: {elapsed} ms")

    str_result = str(result)
    print(f"fib(100000) length: {len(str_result)} digits")
    print(f"fib(100000) first 50: {str_result[:50]}")
    print(f"fib(100000) last 50: {str_result[-50:]}")

if __name__ == "__main__":
    main()
