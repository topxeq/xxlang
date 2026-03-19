// fib.zig - A WebAssembly plugin for Fibonacci calculation
// Build: zig build-exe fib.zig -target wasm32-freestanding -O ReleaseSmall -fno-entry
//
// Plugin conventions:
// - alloc(size) -> ptr: allocate memory from WASM heap
// - plugin_name(ptr): write name string info to ptr (ptr, size)
// - plugin_version(ptr): write version string info to ptr (ptr, size)
// - call_xxx(args...): exported functions accessible as "xxx" in Xxlang

// Simple bump allocator using WASM linear memory
// Start from a reasonable offset to avoid conflicts
var heap_ptr: usize = 65536; // Start at 64KB offset

// Memory allocation from WASM heap
export fn alloc(size: u32) u32 {
    // Ensure 8-byte alignment for int64
    if (heap_ptr % 8 != 0) {
        heap_ptr += 8 - (heap_ptr % 8);
    }

    const result: u32 = @intCast(heap_ptr);
    heap_ptr += size;
    return result;
}

// Helper: write string to WASM memory
fn writeString(s: []const u8, result_ptr: u32) void {
    const offset = alloc(@intCast(s.len));

    // Write string to memory
    const mem = @as([*]u8, @ptrFromInt(offset));
    @memcpy(mem, s);

    // Write ptr and size to result_ptr
    const result = @as(*[2]u32, @ptrFromInt(result_ptr));
    result[0] = offset;
    result[1] = @intCast(s.len);
}

// Plugin name - writes (ptr, size) to result_ptr
export fn plugin_name(result_ptr: u32) void {
    writeString("fib", result_ptr);
}

// Plugin version - writes (ptr, size) to result_ptr
export fn plugin_version(result_ptr: u32) void {
    writeString("1.0.0-zig", result_ptr);
}

// Fibonacci - O(n) algorithm
export fn call_fast(n: i64) i64 {
    if (n <= 1) return n;
    var a: i64 = 0;
    var b: i64 = 1;
    var i: i64 = 2;
    while (i <= n) : (i += 1) {
        const tmp = a + b;
        a = b;
        b = tmp;
    }
    return b;
}

// Matrix multiplication helper for matrix algorithm
fn multiply(a: [2][2]i64, b: [2][2]i64) [2][2]i64 {
    return [2][2]i64{
        .{ a[0][0] * b[0][0] + a[0][1] * b[1][0], a[0][0] * b[0][1] + a[0][1] * b[1][1] },
        .{ a[1][0] * b[0][0] + a[1][1] * b[1][0], a[1][0] * b[0][1] + a[1][1] * b[1][1] },
    };
}

// Fibonacci - O(log n) matrix algorithm
export fn call_matrix(n: i64) i64 {
    if (n <= 1) return n;

    var result = [2][2]i64{ .{ 1, 0 }, .{ 0, 1 } };
    var base = [2][2]i64{ .{ 1, 1 }, .{ 1, 0 } };
    var m = n;

    while (m > 0) {
        if (m & 1 == 1) {
            result = multiply(result, base);
        }
        base = multiply(base, base);
        m >>= 1;
    }

    return result[0][1];
}

// Check if number is Fibonacci - returns 0 or 1
export fn call_isFib(n: i64) i32 {
    if (n < 0) return 0;

    // Check if 5*n^2+4 or 5*n^2-4 is a perfect square
    const n2 = n * n;
    const check1 = 5 * n2 + 4;
    const check2 = 5 * n2 - 4;

    // Check perfect square
    const isPerfectSquare = struct {
        fn f(x: i64) bool {
            var sqrt: i64 = 0;
            while (sqrt * sqrt < x) : (sqrt += 1) {}
            return sqrt * sqrt == x;
        }
    }.f;

    if (isPerfectSquare(check1) or isPerfectSquare(check2)) {
        return 1;
    }
    return 0;
}

// Fibonacci range - returns (ptr, count) for array of int64
// Writes result to result_ptr: [ptr:u32, count:u32]
export fn call_range_(n: i64, result_ptr: u32) void {
    const result = @as(*[2]u32, @ptrFromInt(result_ptr));

    if (n < 0) {
        result[0] = 0;
        result[1] = 0;
        return;
    }

    const count: u32 = @intCast(n + 1);
    const size = count * 8;
    const ptr = alloc(size);

    // Fill array with Fibonacci numbers
    const arr = @as([*]i64, @ptrFromInt(ptr));
    var a: i64 = 0;
    var b: i64 = 1;
    for (0..@intCast(n + 1)) |i| {
        arr[i] = a;
        const tmp = a + b;
        a = b;
        b = tmp;
    }

    // Write ptr and count to result
    result[0] = ptr;
    result[1] = count;
}
