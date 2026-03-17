// testplugin.rs - Test WASM plugin for unit tests
// Build: rustc --target wasm32-unknown-unknown -O --crate-type cdylib -o testplugin.wasm testplugin.rs

#![no_std]
#![no_main]

use core::panic::PanicInfo;

#[panic_handler]
fn panic(_info: &PanicInfo) -> ! {
    loop {}
}

static mut HEAP_PTR: usize = 65536;

/// Memory allocator for WASM
#[no_mangle]
pub extern "C" fn alloc(size: u32) -> u32 {
    unsafe {
        // Ensure 8-byte alignment for i64
        if HEAP_PTR % 8 != 0 {
            HEAP_PTR += 8 - (HEAP_PTR % 8);
        }
        let result = HEAP_PTR as u32;
        HEAP_PTR += size as usize;
        result
    }
}

/// Write string to memory and return (ptr, size) via result_ptr
fn write_string(s: &str, result_ptr: u32) {
    unsafe {
        let bytes = s.as_bytes();
        let offset = alloc(bytes.len() as u32);

        // Write string to memory
        let mem = offset as *mut u8;
        for (i, &byte) in bytes.iter().enumerate() {
            *mem.add(i) = byte;
        }

        // Write (ptr, size) to result_ptr
        let result = result_ptr as *mut u32;
        *result = offset;
        *result.add(1) = bytes.len() as u32;
    }
}

/// Plugin name - writes (ptr, size) to result_ptr
#[no_mangle]
pub extern "C" fn plugin_name(result_ptr: u32) {
    write_string("testplugin", result_ptr);
}

/// Plugin version - writes (ptr, size) to result_ptr
#[no_mangle]
pub extern "C" fn plugin_version(result_ptr: u32) {
    write_string("1.0.0-rust", result_ptr);
}

/// Add two numbers
#[no_mangle]
pub extern "C" fn call_add(a: i64, b: i64) -> i64 {
    a + b
}

/// Subtract two numbers
#[no_mangle]
pub extern "C" fn call_sub(a: i64, b: i64) -> i64 {
    a - b
}

/// Multiply two numbers
#[no_mangle]
pub extern "C" fn call_mul(a: i64, b: i64) -> i64 {
    a * b
}

/// Divide two numbers
#[no_mangle]
pub extern "C" fn call_div(a: i64, b: i64) -> i64 {
    a / b
}

/// Modulo operation
#[no_mangle]
pub extern "C" fn call_mod(a: i64, b: i64) -> i64 {
    a % b
}

/// Power function
#[no_mangle]
pub extern "C" fn call_pow(base: i64, exp: i64) -> i64 {
    let mut result = 1i64;
    let mut b = base;
    let mut e = exp;
    while e > 0 {
        if e & 1 == 1 {
            result *= b;
        }
        b *= b;
        e >>= 1;
    }
    result
}

/// Absolute value
#[no_mangle]
pub extern "C" fn call_abs(n: i64) -> i64 {
    if n < 0 { -n } else { n }
}

/// Negate
#[no_mangle]
pub extern "C" fn call_neg(n: i64) -> i64 {
    -n
}

/// Square
#[no_mangle]
pub extern "C" fn call_square(n: i64) -> i64 {
    n * n
}

/// Factorial
#[no_mangle]
pub extern "C" fn call_factorial(n: i64) -> i64 {
    if n <= 1 {
        return 1;
    }
    let mut result = 1i64;
    let mut i = 2;
    while i <= n {
        result *= i;
        i += 1;
    }
    result
}

/// Fibonacci - O(n) algorithm
#[no_mangle]
pub extern "C" fn call_fast(n: i64) -> i64 {
    if n <= 1 {
        return n;
    }
    let mut a: i64 = 0;
    let mut b: i64 = 1;
    for _ in 2..=n {
        let tmp = a + b;
        a = b;
        b = tmp;
    }
    b
}

/// Fibonacci - O(log n) matrix algorithm
#[no_mangle]
pub extern "C" fn call_fib(n: i64) -> i64 {
    if n <= 1 {
        return n;
    }

    fn multiply(a: [[i64; 2]; 2], b: [[i64; 2]; 2]) -> [[i64; 2]; 2] {
        [
            [a[0][0] * b[0][0] + a[0][1] * b[1][0], a[0][0] * b[0][1] + a[0][1] * b[1][1]],
            [a[1][0] * b[0][0] + a[1][1] * b[1][0], a[1][0] * b[0][1] + a[1][1] * b[1][1]],
        ]
    }

    let mut result = [[1i64, 0], [0, 1]];
    let mut base = [[1i64, 1], [1, 0]];
    let mut m = n;

    while m > 0 {
        if m & 1 == 1 {
            result = multiply(result, base);
        }
        base = multiply(base, base);
        m >>= 1;
    }

    result[0][1]
}

/// GCD
#[no_mangle]
pub extern "C" fn call_gcd(a: i64, b: i64) -> i64 {
    let mut x = a.abs();
    let mut y = b.abs();
    while y != 0 {
        let t = y;
        y = x % y;
        x = t;
    }
    x
}

/// LCM
#[no_mangle]
pub extern "C" fn call_lcm(a: i64, b: i64) -> i64 {
    if a == 0 || b == 0 {
        return 0;
    }
    (a.abs() / call_gcd(a, b)) * b.abs()
}

/// Max of two numbers
#[no_mangle]
pub extern "C" fn call_max(a: i64, b: i64) -> i64 {
    if a > b { a } else { b }
}

/// Min of two numbers
#[no_mangle]
pub extern "C" fn call_min(a: i64, b: i64) -> i64 {
    if a < b { a } else { b }
}

/// Clamp value to range [min, max]
#[no_mangle]
pub extern "C" fn call_clamp(value: i64, min: i64, max: i64) -> i64 {
    if value < min {
        min
    } else if value > max {
        max
    } else {
        value
    }
}

/// Check if number is even
#[no_mangle]
pub extern "C" fn call_is_even(n: i64) -> i32 {
    if n % 2 == 0 { 1 } else { 0 }
}

/// Check if number is odd
#[no_mangle]
pub extern "C" fn call_is_odd(n: i64) -> i32 {
    if n % 2 != 0 { 1 } else { 0 }
}

/// Check if number is prime
#[no_mangle]
pub extern "C" fn call_is_prime(n: i64) -> i32 {
    if n < 2 {
        return 0;
    }
    if n == 2 {
        return 1;
    }
    if n % 2 == 0 {
        return 0;
    }
    let mut i = 3;
    while i * i <= n {
        if n % i == 0 {
            return 0;
        }
        i += 2;
    }
    1
}

/// Check if number is perfect square
#[no_mangle]
pub extern "C" fn call_is_square(n: i64) -> i32 {
    if n < 0 {
        return 0;
    }
    let mut sqrt: i64 = 0;
    while sqrt * sqrt < n {
        sqrt += 1;
    }
    if sqrt * sqrt == n { 1 } else { 0 }
}

/// Check if number is Fibonacci
#[no_mangle]
pub extern "C" fn call_isFib(n: i64) -> i32 {
    if n < 0 {
        return 0;
    }

    fn is_perfect_square(x: i64) -> bool {
        let mut sqrt: i64 = 0;
        while sqrt * sqrt < x {
            sqrt += 1;
        }
        sqrt * sqrt == x
    }

    let n2 = n * n;
    let check1 = 5 * n2 + 4;
    let check2 = 5 * n2 - 4;

    if is_perfect_square(check1) || is_perfect_square(check2) {
        1
    } else {
        0
    }
}

/// Triangle number (sum of 1 to n)
#[no_mangle]
pub extern "C" fn call_triangle(n: i64) -> i64 {
    n * (n + 1) / 2
}

/// Triangular number (same as triangle)
#[no_mangle]
pub extern "C" fn call_triangular(n: i64) -> i64 {
    n * (n + 1) / 2
}

/// Sum of squares from 1 to n
#[no_mangle]
pub extern "C" fn call_sum_squares(n: i64) -> i64 {
    n * (n + 1) * (2 * n + 1) / 6
}

/// Count primes up to n
#[no_mangle]
pub extern "C" fn call_count_primes(n: i64) -> i64 {
    if n < 2 {
        return 0;
    }
    let mut count = 0i64;
    let mut i = 2;
    while i <= n {
        if call_is_prime(i) == 1 {
            count += 1;
        }
        i += 1;
    }
    count
}

/// Binomial coefficient C(n, k)
#[no_mangle]
pub extern "C" fn call_binomial(n: i64, k: i64) -> i64 {
    if k < 0 || k > n {
        return 0;
    }
    if k == 0 || k == n {
        return 1;
    }
    let k = if k > n - k { n - k } else { k };
    let mut result = 1i64;
    let mut i = 0;
    while i < k {
        result = result * (n - i) / (i + 1);
        i += 1;
    }
    result
}

/// Sum array - takes ptr and count, returns sum
#[no_mangle]
pub extern "C" fn call_sum_array(ptr: u32, count: u32) -> i64 {
    unsafe {
        let arr = ptr as *const i64;
        let mut sum = 0i64;
        for i in 0..count {
            sum += *arr.add(i as usize);
        }
        sum
    }
}

/// Range - returns (ptr, count) for array of i64 from 0 to n-1
#[no_mangle]
pub extern "C" fn call_range_(n: i64, result_ptr: u32) {
    unsafe {
        let result = result_ptr as *mut u32;

        if n < 0 {
            *result = 0;
            *result.add(1) = 0;
            return;
        }

        let count = n as u32;
        let size = count * 8;
        let ptr = alloc(size);

        // Fill array with 0..n
        let arr = ptr as *mut i64;
        for i in 0..n {
            *arr.add(i as usize) = i;
        }

        *result = ptr;
        *result.add(1) = count;
    }
}
