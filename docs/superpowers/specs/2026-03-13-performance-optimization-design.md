# Performance Optimization Design for Xxlang

**Date:** 2026-03-13
**Status:** Pending Implementation
**Type:** Optimization Design

## Executive Summary

This document describes a hybrid approach to optimize Xxlang's performance, combining bytecode-level optimizations with VM-level improvements. The goal is to achieve 5-10x overall performance improvement while maintaining code clarity and testability.

**Target Improvements:**
- Function call overhead: 2,197x → < 1,000x (vs Go)
- Fibonacci(10): 806x → < 400x (vs Go)
- Loop overhead: 103x → < 80x (vs Go)
- Maintain 67%+ test coverage
- Keep memory overhead under 2x

---

## Current Performance Baseline

From benchmarks run on 2026-03-13:

| Benchmark | Go | Xxlang | Slowdown |
|-----------|-----|--------|----------|
| fib(10) | 1,225 ns | 987,366 ns | 806x |
| fib(20) | 48,140 ns | 8,848,071 ns | 184x |
| LoopSum(10000) | 3,669 ns | 377,872 ns | 103x |
| ArraySum(1000) | 581 ns | 486,202 ns | 837x |
| FunctionCalls(1000) | 287 ns | 630,330 ns | 2,197x |

**Identified Bottlenecks:**
1. Function call overhead (closure allocation, stack operations)
2. Method/function lookup overhead (hash map access)
3. Array bounds checking on every access
4. Type assertions in arithmetic operations
5. Generic opcode dispatch

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Compilation Phase                          │
│  ┌────────────┐    ┌────────────┐    ┌──────────┐ │
│  │   Source    │ →  │   Parser    │ →  │   AST    │ │
│  └────────────┘    └────────────┘    └──────────┘ │
│                         v                                 │
│  ┌─────────────────────────────────────┐            │
│  │  Bytecode Compiler                │            │
│  └─────────────────────────────────────┘            │
│                         v                                 │
│  ┌────────────┐    ┌────────────┐                 │
│  │  Bytecode   │ →  │ Optimizer  │                 │
│  └────────────┘    └────────────┘                 │
│                         v                                 │
│  ┌─────────────────────────────────────┐            │
│  │  Optimized Bytecode            │            │
│  └─────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────┐
│                    Execution Phase                        │
│  ┌─────────────────────────────────────┐             │
│  │  VM with Optimized Components    │             │
│  │  - Inline Cache               │             │
│  │  - Closure Pool              │             │
│  │  - Type Specialized Opcodes   │             │
│  └─────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Design

### 1. Bytecode Optimizer

**File:** `pkg/compiler/optimizer.go`

**Purpose:** Transform bytecode before execution to reduce instruction count and improve patterns.

**Passes:**

#### Pass 1: Constant Folding

Evaluates constant expressions at compile time:

```
Before:                          After:
-------                          ------
push const 1                     push const 3
push const 2
add
```

**Supported operations:**
- Binary: `+`, `-`, `*`, `/` on integer/float constants
- Unary: `-`, `!` on boolean/integer constants
- Comparison: `<`, `>`, `<=`, `>=`, `==`, `!=` on constants

#### Pass 2: Dead Code Elimination

Removes unreachable instructions after unconditional jumps.

#### Pass 3: Peephole Optimization

Simplifies local patterns:
- `push X; pop` → remove both (common after if statements)
- `getLocal N; setLocal N` → optimize to keep operation if result used

**API:**

```go
type Optimizer struct {
    bytecode *Bytecode
}

// Run all optimization passes
func (o *Optimizer) Optimize() *Bytecode

// Run specific pass
func (o *Optimizer) FoldConstants() *Bytecode
func (o *Optimizer) EliminateDeadCode() *Bytecode
func (o *Optimizer) OptimizePeephole() *Bytecode

// Check if optimization preserved semantics
func (o *Optimizer) Verify(original, optimized *Bytecode) bool
```

---

### 2. Inline Cache

**File:** `pkg/compiler/inline_cache.go`

**Purpose:** Cache method/function lookups at specific bytecode positions to avoid hash map lookups during execution.

**Data Structure:**

```go
type InlineCacheEntry struct {
    Position int       // Bytecode IP
    Object   Object  // Cached function/method object
    Type     ObjectType // For type checking
    Depth     int    // Call stack depth for invalidation
}

type InlineCache struct {
    entries []InlineCacheEntry
    size    int
}
```

**Usage:**

1. **Compiler:** Fills cache during compilation
   - `OpGetMethod` → add cache entry for next `OpCallMethod`
   - `OpCall` → add cache entry for function

2. **VM:** Checks cache before hash map lookup
   - `GetMethod()` → check cache first, fallback to hash map
   - Cache hit avoids `map[string]Object` lookup (~50-100ns saved per call)

**Integration:**

```go
type Bytecode struct {
    Instructions []byte
    Constants   []Object
    SourceMap   *SourceMap
    InlineCache *InlineCache  // New field
}

// In compiler:
func (c *Compiler) emitMethodCall(nameIdx int) {
    // ... compile method lookup and arguments ...
    // Add cache entry
    c.bytecode.InlineCache.Set(c.lastIP, nameIdx)
}
```

**Invalidation:**
- Cache entries are per-bytecode (immutable after compilation)
- No runtime invalidation needed
- Module bytecode gets its own cache

---

### 3. Closure Pool

**File:** `pkg/vm/closure_pool.go`

**Purpose:** Reuse closure objects instead of allocating new ones for every function call.

**Problem:** Recursive functions like Fibonacci allocate thousands of closures.

**Solution:** Object pool with tiered sizing.

**Data Structure:**

```go
type ClosurePool struct {
    mu     sync.RWMutex
    tiers  map[int][]*objects.Closure  // Keyed by numFreeVars
    stats   PoolStats
}

type PoolStats struct {
    Acquired  int64
    Released  int64
    Created   int64  // New allocations (pool miss)
    Reused    int64  // Pool hits
}
```

**Pool Tiers:**
- `tier0`: 0-2 free variables (small, most common)
- `tier1`: 3-5 free variables (medium)
- `tier2`: 6+ free variables (large, rare)

**Initial Size:** 16 per tier, grows exponentially.

**API:**

```go
// Acquire a closure from pool or allocate new
func (p *ClosurePool) Acquire(numFree int) *objects.Closure

// Return closure to pool for reuse
func (p *ClosurePool) Release(closure *objects.Closure)

// Get statistics for profiling
func (p *ClosurePool) Stats() PoolStats

// Reset pool (for testing)
func (p *ClosurePool) Reset()
```

**VM Integration:**

```go
type VM struct {
    // ... existing fields ...
    closurePool *ClosurePool  // New field
}

// In executeClosure():
func (vm *VM) executeClosure() error {
    fnIndex := vm.readUint16()
    freeVars := vm.readUint8()

    // Try to acquire from pool first
    closure := vm.closurePool.Acquire(freeVars)
    closure.Fn = vm.constants[fnIndex].(*CompiledFunction)
    // ... set free variables ...

    vm.stack.Push(closure)
    return nil
}

// When frame is popped, return closure to pool
func (vm *VM) popFrame() *Frame {
    frame := vm.frames[vm.frameIndex-1]
    vm.frameIndex--

    // Return closure to pool if applicable
    if frame.Closure != nil {
        vm.closurePool.Release(frame.Closure)
    }

    return frame
}
```

---

### 4. Type Specialization

**Purpose:** Add specialized opcodes for common type combinations to avoid runtime type assertions.

**New Opcodes:**

| Opcode | Description | When Used |
|--------|-------------|-------------|
| `OpAddInt` | Integer-only addition | Both operands are int constants |
| `OpAddFloat` | Float-only addition | Both operands are float constants |
| `OpMulIntBy2` | Multiply by 2 | Right operand is constant 2 |
| `OpMulIntBy10` | Multiply by 10 | Formatting common pattern |
| `OpSetLocalByConst` | Set local to constant | Eliminates push/pop |

**Compiler Changes:**

```go
// In compileInfixExpression():
case "*":
    // Check if right operand is constant
    if constLit, ok := node.Right.(*IntegerLiteral); ok {
        if constLit.Value == 2 {
            c.emit(OpMulIntBy2)  // Specialized opcode
            return nil
        }
        if constLit.Value == 10 {
            c.emit(OpMulIntBy10)
            return nil
        }
    }
    // Fall back to generic OpMul
    c.emit(OpMul)
```

**VM Changes:**

```go
case compiler.OpAddInt:
    left := vm.stack.Pop().(*Int)
    right := vm.stack.Pop().(*Int)
    vm.stack.Push(&Int{Value: left.Value + right.Value})
    // No type assertions needed!
```

**Fallback Strategy:**
- Specialized opcodes have direct type access (known at compile time)
- Generic opcodes remain for dynamic types
- No changes to language semantics

---

### 5. Performance Profiling

**File:** `pkg/vm/profiler.go`

**Purpose:** Count opcode executions to identify hot paths.

**API:**

```go
type Profiler struct {
    opcodeCounts map[Opcode]int64
    enabled     bool
}

func (p *Profiler) Record(op Opcode)
func (p *Profiler) Report() string
func (p *Profiler) Reset()
```

**Usage:**

```go
type VM struct {
    profiler *Profiler
}

// In Run():
op := Opcode(ins[ip])
vm.profiler.Record(op)
// ... execute opcode ...
```

**Output Example:**

```
Opcode Execution Counts:
OpCall:        10,543
OpGetLocal:    52,189
OpSetLocal:    31,205
OpAdd:         25,672
OpReturn:       10,542
```

---

## Configuration

All optimizations are configurable at runtime:

```go
type OptimizationFlags struct {
    BytecodeOptimizer  bool  // Enable bytecode passes
    InlineCache      bool  // Enable inline caching
    ClosurePool      bool  // Enable closure pooling
    TypeSpecialization bool // Enable specialized opcodes
}

// Default: all enabled
func DefaultOptimizations() OptimizationFlags {
    return OptimizationFlags{
        BytecodeOptimizer:  true,
        InlineCache:       true,
        ClosurePool:       true,
        TypeSpecialization: true,
    }
}
```

---

## Implementation Phases

### Phase 1: Foundation (Low Risk)
- [ ] Create `pkg/compiler/optimizer.go` structure
- [ ] Implement constant folding pass
- [ ] Add optimization enable/disable flags
- [ ] Create baseline benchmarks
- [ ] Ensure all existing tests pass

### Phase 2: Inline Cache (Medium Risk)
- [ ] Create `pkg/compiler/inline_cache.go`
- [ ] Integrate inline cache into compiler
- [ ] Update VM to use cache
- [ ] Add cache hit/miss tests
- [ ] Benchmark cache effectiveness

### Phase 3: Closure Pool (Medium Risk)
- [ ] Create `pkg/vm/closure_pool.go`
- [ ] Integrate pool into VM
- [ ] Add stress tests for pool
- [ ] Benchmark pool hit rate
- [ ] Ensure no memory leaks

### Phase 4: Type Specialization (Higher Risk)
- [ ] Add specialized opcodes to compiler
- [ ] Implement specialized VM handlers
- [ ] Add fallback tests for dynamic types
- [ ] Benchmark specialized vs generic paths

### Phase 5: Integration & Validation
- [ ] Run full test suite
- [ ] Run benchmarks
- [ ] Measure performance improvements
- [ ] Update RESULTS.md
- [ ] Document any known limitations

---

## Error Handling Strategy

### Optimizer Failures
- Optimization errors should log warning but not fail compilation
- Fallback to unoptimized bytecode
- Verification step catches semantic changes

### Pool Exhaustion
- Gracefully fallback to new allocation
- Log warning for capacity monitoring
- Consider auto-growing pool in production

### Cache Behavior
- Cache misses are normal (first execution)
- Cache hits are the optimization win
- No invalidation needed (bytecode is immutable)

---

## Testing Strategy

### Unit Tests

```go
// Test each optimization pass independently
func TestConstantFolding(t *testing.T) {
    cases := []struct {
        input    string
        expected  []byte
    }{
        {"push 1; push 2; add", []byte{push_const, 3}},
        // ... more cases
    }
    for _, c := range cases {
        result := Optimize(c.input)
        assert.Equal(t, c.expected, result)
    }
}

func TestInlineCache(t *testing.T) {
    cache := NewInlineCache()
    cache.Set(10, &Function{Name: "foo"})

    fn, found := cache.Get(10)
    assert.True(t, found)
    assert.Equal(t, "foo", fn.Name)
}

func TestClosurePool(t *testing.T) {
    pool := NewClosurePool()

    c1 := pool.Acquire(2)
    assert.NotNil(t, c1)

    pool.Release(c1)
    c2 := pool.Acquire(2)

    assert.Equal(t, c1, c2)  // Reused!
}
```

### Regression Tests

Ensure optimizations preserve semantics:

```go
func TestOptimizationPreservesSemantics(t *testing.T) {
    testCases := []string{
        "1 + 2",
        "func fib(n) { if (n <= 1) { return n } return fib(n-1) + fib(n-2) } fib(10)",
        // ... more comprehensive tests
    }

    for _, code := range testCases {
        result1 := ExecuteOriginal(code)
        result2 := ExecuteOptimized(code)

        assert.Equal(t, result1, result2,
            "Optimization changed semantics for: %s", code)
    }
}
```

### Benchmark Suite

```go
func BenchmarkFib10_Original(b *testing.B) {
    DisableAllOptimizations()
    runFibBenchmark(10, b)
}

func BenchmarkFib10_Optimized(b *testing.B) {
    EnableAllOptimizations()
    runFibBenchmark(10, b)
}

func BenchmarkLoopSum_Original(b *testing.B) {
    DisableAllOptimizations()
    runLoopBenchmark(10000, b)
}

func BenchmarkLoopSum_Optimized(b *testing.B) {
    EnableAllOptimizations()
    runLoopBenchmark(10000, b)
}
```

---

## Success Criteria

| Metric | Current | Target | How to Measure |
|---------|---------|---------|----------------|
| fib(10) vs Go | 806x | < 400x | `go test -bench=BenchmarkFib10` |
| fib(20) vs Python | 4x faster | < 3x faster | `go test -bench=BenchmarkFib20` |
| Loop overhead vs Go | 103x | < 80x | `go test -bench=BenchmarkLoopSum` |
| Function call overhead vs Go | 2,197x | < 1,000x | `go test -bench=BenchmarkFunctionCalls` |
| Test coverage | 67% | ≥ 67% | `go test -cover` |
| Memory overhead vs baseline | 1x | < 2x | `go test -benchmem` |
| Compilation time | ~130µs | < 200µs | `go test -bench=BenchmarkCompile` |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|-------|------------|--------|-------------|
| Optimization changes semantics | Low | High | Comprehensive regression tests |
| Pool causes memory leaks | Low | Medium | Reference counting + GC tests |
| Type specialization breaks dynamic code | Medium | High | Fallback to generic opcodes |
| Inline cache grows unbounded | Low | Low | Fixed size per bytecode |
| Performance regressions | Low | Medium | Before/after benchmarks, CI checks |

---

## Future Work (Out of Scope)

### Long-term Optimizations
1. **JIT Compilation**: Compile hot functions to native Go code
2. **Parallel Execution**: Support concurrent execution of independent code
3. **Native Function Bindings**: Allow critical paths to use Go directly
4. **Escape Analysis**: Reduce heap allocations through escape analysis

### Measurement Improvements
1. **Hot Path Profiling**: Identify actual bottlenecks at runtime
2. **Memory Profiling**: Track allocations by operation type
3. **Flame Graphs**: Visualize call stack and hot paths

---

## Appendix: Performance References

- **LuaJIT**: Achieves 2-5x slower than C via JIT
- **CPython**: ~50-100x slower than C, heavily optimized
- **V8 (Node.js)**: ~5-10x slower than C++ via JIT
- **Python 3.11+**: Improved interpreter with specialized bytecodes

Xxlang's naive implementation is at the lower end of typical interpreter performance. These optimizations aim to move toward LuaJIT/PyPy levels while maintaining simplicity.
