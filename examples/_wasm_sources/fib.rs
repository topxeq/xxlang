// fib.rs - Fibonacci WASM plugin in Rust
// Build: rustc --target wasm32-unknown-unknown -O --crate-type cdylib -o fib.wasm fib.rs

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
    write_string("fib", result_ptr);
}

/// Plugin version - writes (ptr, size) to result_ptr
#[no_mangle]
pub extern "C" fn plugin_version(result_ptr: u32) {
    write_string("1.0.0-rust", result_ptr);
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

/// Matrix multiplication helper
fn multiply(a: [[i64; 2]; 2], b: [[i64; 2]; 2]) -> [[i64; 2]; 2] {
    [
        [a[0][0] * b[0][0] + a[0][1] * b[1][0], a[0][0] * b[0][1] + a[0][1] * b[1][1]],
        [a[1][0] * b[0][0] + a[1][1] * b[1][0], a[1][0] * b[0][1] + a[1][1] * b[1][1]],
    ]
}

/// Fibonacci - O(log n) matrix algorithm
#[no_mangle]
pub extern "C" fn call_matrix(n: i64) -> i64 {
    if n <= 1 {
        return n;
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

/// Check if number is Fibonacci - returns 0 or 1
#[no_mangle]
pub extern "C" fn call_isFib(n: i64) -> i32 {
    if n < 0 {
        return 0;
    }

    let n2 = n * n;
    let check1 = 5 * n2 + 4;
    let check2 = 5 * n2 - 4;

    fn is_perfect_square(x: i64) -> bool {
        let mut sqrt: i64 = 0;
        while sqrt * sqrt < x {
            sqrt += 1;
        }
        sqrt * sqrt == x
    }

    if is_perfect_square(check1) || is_perfect_square(check2) {
        1
    } else {
        0
    }
}

/// Fibonacci range - returns (ptr, count) for array of i64
#[no_mangle]
pub extern "C" fn call_range_(n: i64, result_ptr: u32) {
    unsafe {
        let result = result_ptr as *mut u32;

        if n < 0 {
            *result = 0;
            *result.add(1) = 0;
            return;
        }

        let count = (n + 1) as u32;
        let size = count * 8;
        let ptr = alloc(size);

        // Fill array with Fibonacci numbers
        let arr = ptr as *mut i64;
        let mut a: i64 = 0;
        let mut b: i64 = 1;
        for i in 0..=n {
            *arr.add(i as usize) = a;
            let tmp = a + b;
            a = b;
            b = tmp;
        }

        *result = ptr;
        *result.add(1) = count;
    }
}
