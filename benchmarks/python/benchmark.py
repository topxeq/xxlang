# benchmarks/python/benchmark.py
# Python performance benchmarks for comparison

import time
import sys

# Fibonacci recursive
def fib(n):
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)

def benchmark_fibonacci15():
    fib(15)

def benchmark_fibonacci20():
    fib(20)

# Fibonacci iterative
def fib_iter(n):
    a, b = 0, 1
    for _ in range(n):
        a, b = b, a + b
    return a

def benchmark_fibonacci_iterative():
    fib_iter(30)

# Factorial
def fact(n):
    if n <= 1:
        return 1
    return n * fact(n - 1)

def benchmark_factorial():
    fact(12)

# For loops
def benchmark_for_loop_100():
    sum_val = 0
    for i in range(100):
        sum_val += i

def benchmark_for_loop_1000():
    sum_val = 0
    for i in range(1000):
        sum_val += i

# While loop
def benchmark_while_loop():
    i = 0
    sum_val = 0
    while i < 100:
        sum_val += i
        i += 1

# Function calls
def add(a, b):
    return a + b

def mul(a, b):
    return a * b

def benchmark_function_calls():
    sum_val = 0
    for i in range(100):
        sum_val = add(sum_val, i)
        sum_val = mul(sum_val, 1)

# Arithmetic
def benchmark_arithmetic():
    a = 10
    b = 20
    c = a + b
    d = c * 2
    e = d - 5
    f = e / 3

# Intensive arithmetic
def benchmark_intensive_arithmetic():
    x = 0
    for i in range(1000):
        x = x + i * 2 - i // 2
        x = x % 1000

# Nested expressions
def benchmark_nested_expressions():
    a = 1 + 2 * 3 - 4 // 2
    c = 10
    d = (a + 5) * (c - a) // (5 + 1)

# Comparisons
def benchmark_comparisons():
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

def benchmark_prime_check_100():
    for j in range(100):
        is_prime(j)

# Bubble sort
def bubble_sort(arr):
    n = len(arr)
    for i in range(n - 1):
        for j in range(n - i - 1):
            if arr[j] > arr[j + 1]:
                arr[j], arr[j + 1] = arr[j + 1], arr[j]

def benchmark_bubble_sort_10():
    arr = [5, 3, 8, 4, 2, 7, 1, 10, 6, 9]
    bubble_sort(arr)

# Array sum
def benchmark_array_sum_1000():
    arr = list(range(1000))
    sum_val = sum(arr)

# String concatenation
def benchmark_string_concat_100():
    s = ""
    for _ in range(100):
        s += "x"

def run_benchmark(name, func, iterations=100):
    # Warmup
    for _ in range(5):
        func()

    # Measure
    start = time.perf_counter()
    for _ in range(iterations):
        func()
    elapsed = time.perf_counter() - start

    avg_ns = (elapsed / iterations) * 1e9
    print(f"{name}: {avg_ns:.2f} ns/op")

def main():
    benchmarks = [
        ("Fibonacci15", benchmark_fibonacci15),
        ("Fibonacci20", benchmark_fibonacci20),
        ("FibonacciIterative", benchmark_fibonacci_iterative),
        ("Factorial", benchmark_factorial),
        ("ForLoop100", benchmark_for_loop_100),
        ("ForLoop1000", benchmark_for_loop_1000),
        ("WhileLoop", benchmark_while_loop),
        ("FunctionCalls", benchmark_function_calls),
        ("Arithmetic", benchmark_arithmetic),
        ("IntensiveArithmetic", benchmark_intensive_arithmetic),
        ("NestedExpressions", benchmark_nested_expressions),
        ("Comparisons", benchmark_comparisons),
        ("PrimeCheck100", benchmark_prime_check_100),
        ("BubbleSort10", benchmark_bubble_sort_10),
        ("ArraySum1000", benchmark_array_sum_1000),
        ("StringConcat100", benchmark_string_concat_100),
    ]

    print("Python Performance Benchmarks")
    print("=" * 40)

    for name, func in benchmarks:
        # Adjust iterations for slower benchmarks
        if name in ["Fibonacci20", "FibonacciIterative"]:
            run_benchmark(name, func, iterations=10)
        elif name in ["Fibonacci15", "Factorial"]:
            run_benchmark(name, func, iterations=50)
        else:
            run_benchmark(name, func, iterations=100)

if __name__ == "__main__":
    main()
