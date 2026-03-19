// fib.c - A WebAssembly plugin for Fibonacci calculation
// Build: clang -o fib.wasm --target=wasm32 -O2 fib.c -nostdlib -nostartfiles -Wl,--no-entry -Wl,--export-all
//
// Plugin conventions:
// - alloc(size) -> ptr: allocate memory from WASM heap
// - plugin_name(ptr): write name string info to ptr (ptr, size)
// - plugin_version(ptr): write version string info to ptr (ptr, size)
// - call_xxx(args...): exported functions accessible as "xxx" in Xxlang

#include <stdint.h>
#include <stddef.h>

// External symbols provided by the linker
extern unsigned char __heap_base;
extern unsigned char __data_end;

// Simple bump allocator using WASM linear memory
static uintptr_t heap_ptr = 0;

// Memory allocation from WASM heap
uint32_t alloc(uint32_t size) {
    // Initialize heap pointer if needed
    if (heap_ptr == 0) {
        heap_ptr = (uintptr_t)&__heap_base;
    }

    // Ensure 8-byte alignment for int64
    if (heap_ptr % 8 != 0) {
        heap_ptr += 8 - (heap_ptr % 8);
    }

    uintptr_t result = heap_ptr;
    heap_ptr += size;
    return (uint32_t)result;
}

// Helper: write string to WASM memory
static uint32_t write_string(const char* s, uint32_t len, uint32_t result_ptr) {
    uint32_t offset = alloc(len);

    // Write string to memory
    unsigned char* mem = (unsigned char*)offset;
    for (uint32_t i = 0; i < len; i++) {
        mem[i] = (unsigned char)s[i];
    }

    // Write ptr and size to result_ptr
    uint32_t* result = (uint32_t*)result_ptr;
    result[0] = offset;
    result[1] = len;

    return offset;
}

// Plugin name - writes (ptr, size) to result_ptr
void plugin_name(uint32_t result_ptr) {
    const char* name = "fib";
    write_string(name, 3, result_ptr);
}

// Plugin version - writes (ptr, size) to result_ptr
void plugin_version(uint32_t result_ptr) {
    const char* version = "1.0.0-c";
    write_string(version, 7, result_ptr);
}

// Fibonacci - O(n) algorithm
int64_t call_fast(int64_t n) {
    if (n <= 1) return n;
    int64_t a = 0, b = 1;
    for (int64_t i = 2; i <= n; i++) {
        int64_t tmp = a + b;
        a = b;
        b = tmp;
    }
    return b;
}

// Matrix multiplication helper for matrix algorithm
static void multiply(int64_t a[2][2], int64_t b[2][2], int64_t result[2][2]) {
    result[0][0] = a[0][0] * b[0][0] + a[0][1] * b[1][0];
    result[0][1] = a[0][0] * b[0][1] + a[0][1] * b[1][1];
    result[1][0] = a[1][0] * b[0][0] + a[1][1] * b[1][0];
    result[1][1] = a[1][0] * b[0][1] + a[1][1] * b[1][1];
}

// Fibonacci - O(log n) matrix algorithm
int64_t call_matrix(int64_t n) {
    if (n <= 1) return n;

    int64_t result[2][2] = {{1, 0}, {0, 1}};
    int64_t base[2][2] = {{1, 1}, {1, 0}};
    int64_t temp[2][2];

    while (n > 0) {
        if (n & 1) {
            multiply(result, base, temp);
            result[0][0] = temp[0][0]; result[0][1] = temp[0][1];
            result[1][0] = temp[1][0]; result[1][1] = temp[1][1];
        }
        multiply(base, base, temp);
        base[0][0] = temp[0][0]; base[0][1] = temp[0][1];
        base[1][0] = temp[1][0]; base[1][1] = temp[1][1];
        n >>= 1;
    }

    return result[0][1];
}

// Check if number is Fibonacci - returns 0 or 1
int32_t call_isFib(int64_t n) {
    if (n < 0) return 0;

    // Check if 5*n^2+4 or 5*n^2-4 is a perfect square
    int64_t n2 = n * n;
    int64_t check1 = 5 * n2 + 4;
    int64_t check2 = 5 * n2 - 4;

    // Check perfect square for check1
    int64_t sqrt1 = 0;
    while (sqrt1 * sqrt1 < check1) sqrt1++;
    if (sqrt1 * sqrt1 == check1) return 1;

    // Check perfect square for check2
    int64_t sqrt2 = 0;
    while (sqrt2 * sqrt2 < check2) sqrt2++;
    if (sqrt2 * sqrt2 == check2) return 1;

    return 0;
}

// Fibonacci range - returns (ptr, count) for array of int64
// Writes result to result_ptr: [ptr:u32, count:u32]
void call_range_(int64_t n, uint32_t result_ptr) {
    uint32_t* result = (uint32_t*)result_ptr;

    if (n < 0) {
        result[0] = 0;
        result[1] = 0;
        return;
    }

    uint32_t count = (uint32_t)(n + 1);
    uint32_t size = count * 8;
    uint32_t ptr = alloc(size);

    // Fill array with Fibonacci numbers
    int64_t* arr = (int64_t*)ptr;
    int64_t a = 0, b = 1;
    for (int64_t i = 0; i <= n; i++) {
        arr[i] = a;
        int64_t tmp = a + b;
        a = b;
        b = tmp;
    }

    // Write ptr and count to result
    result[0] = ptr;
    result[1] = count;
}

// _start is required for WASI modules
void _start(void) {
    // Initialize heap pointer
    heap_ptr = (uintptr_t)&__heap_base;
}
