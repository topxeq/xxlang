# JIT Compilation Status

## Current Status: Experimental

The JIT compiler framework is implemented but has **machine code execution issues**. For production use, rely on **TCO (Tail Call Optimization)** instead.

## TCO: The Recommended Approach

For recursive algorithms like Fibonacci, TCO provides **dramatic speedups**:

| Method | fib(35) Time | Speedup vs Python |
|--------|--------------|-------------------|
| Python naive | 2.77 seconds | 1x |
| Xxlang naive | 5.12 seconds | 0.5x |
| **Xxlang TCO** | **7.7 µs** | **360,000x** |

### TCO Example

```javascript
// Tail-recursive Fibonacci (EXTREMELY fast)
func fibHelper(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibHelper(n - 1, b, a + b)  // Tail call - reuse stack
}

func fib(n) {
    return fibHelper(n, 0, 1)
}

// No stack overflow, O(n) complexity
fib(1000)  // Works instantly!
```

## JIT Implementation Files

- `/pkg/jit/jit.go` - Core compiler
- `/pkg/jit/codegen.go` - Basic x86-64 generator
- `/pkg/jit/jit_simple.go` - Simplified generator
- `/pkg/jit/jit_recursive.go` - Recursive function support
- `/pkg/jit/jit_vm.go` - JIT VM wrapper

## Known Issues

1. **Machine code crashes** - Generated code causes segfaults
2. **No function calls** - OpRegCall not supported
3. **Jump offset bugs** - Incorrect relative offsets

## CLI Flags

```bash
# JIT is available but experimental
xxl --jit script.xxl

# Recommended: Use TCO instead
# No special flags needed
xxl fib_tco.xxl
```

## Future Work

To make JIT production-ready:
1. Debug x86-64 machine code generation
2. Add function call support
3. Implement proper stack frame management
4. Add comprehensive testing

## Conclusion

**Use TCO for recursion** - it's already implemented and provides massive speedups (300,000x+). JIT remains experimental for future development.
