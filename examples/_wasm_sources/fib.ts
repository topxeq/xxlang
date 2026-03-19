// fib.ts - Fibonacci WASM plugin in AssemblyScript
// Build: asc fib.ts -o fib.wasm --optimize --runtime stub --initialMemory 2

// Heap starts at 64KB (plenty of room within 2 pages = 128KB)
var heapPtr: usize = 65536;

// Memory allocator
export function alloc(size: u32): u32 {
    if (heapPtr % 8 != 0) {
        heapPtr += 8 - (heapPtr % 8);
    }
    const result = heapPtr as u32;
    heapPtr += size as usize;
    return result;
}

// Plugin name
export function plugin_name(resultPtr: u32): void {
    const name = "fib";
    const len = 3;
    const ptr = alloc(len);

    for (let i = 0; i < len; i++) {
        store<u8>(ptr + i, name.charCodeAt(i));
    }

    store<u32>(resultPtr, ptr);
    store<u32>(resultPtr + 4, len);
}

// Plugin version
export function plugin_version(resultPtr: u32): void {
    const version = "1.0.0-as";
    const len = 8;
    const ptr = alloc(len);

    for (let i = 0; i < len; i++) {
        store<u8>(ptr + i, version.charCodeAt(i));
    }

    store<u32>(resultPtr, ptr);
    store<u32>(resultPtr + 4, len);
}

// Fibonacci - O(n)
export function call_fast(n: i64): i64 {
    if (n <= 1) return n;
    let a: i64 = 0;
    let b: i64 = 1;
    for (let i: i64 = 2; i <= n; i++) {
        const tmp = a + b;
        a = b;
        b = tmp;
    }
    return b;
}

// Fibonacci - O(log n)
export function call_matrix(n: i64): i64 {
    if (n <= 1) return n;

    var r00: i64 = 1, r01: i64 = 0, r10: i64 = 0, r11: i64 = 1;
    var b00: i64 = 1, b01: i64 = 1, b10: i64 = 1, b11: i64 = 0;
    var m = n;

    while (m > 0) {
        if (m & 1) {
            const t00 = r00 * b00 + r01 * b10;
            const t01 = r00 * b01 + r01 * b11;
            const t10 = r10 * b00 + r11 * b10;
            const t11 = r10 * b01 + r11 * b11;
            r00 = t00; r01 = t01; r10 = t10; r11 = t11;
        }
        const t00 = b00 * b00 + b01 * b10;
        const t01 = b00 * b01 + b01 * b11;
        const t10 = b10 * b00 + b11 * b10;
        const t11 = b10 * b01 + b11 * b11;
        b00 = t00; b01 = t01; b10 = t10; b11 = t11;
        m >>= 1;
    }
    return r01;
}

// Check if Fibonacci
export function call_isFib(n: i64): i32 {
    if (n < 0) return 0;
    const n2 = n * n;
    const c1 = 5 * n2 + 4;
    const c2 = 5 * n2 - 4;

    var s1: i64 = 0;
    while (s1 * s1 < c1) s1++;
    if (s1 * s1 == c1) return 1;

    var s2: i64 = 0;
    while (s2 * s2 < c2) s2++;
    if (s2 * s2 == c2) return 1;

    return 0;
}

// Fibonacci range
export function call_range_(n: i64, resultPtr: u32): void {
    if (n < 0) {
        store<u32>(resultPtr, 0);
        store<u32>(resultPtr + 4, 0);
        return;
    }

    const count: u32 = (n + 1) as u32;
    const ptr = alloc(count * 8);

    var a: i64 = 0;
    var b: i64 = 1;
    for (let i: i64 = 0; i <= n; i++) {
        store<i64>(ptr + (i as usize) * 8, a);
        const tmp = a + b;
        a = b;
        b = tmp;
    }

    store<u32>(resultPtr, ptr);
    store<u32>(resultPtr + 4, count);
}
