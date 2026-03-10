# Comprehensive Test Suite Design

## Overview

Create a comprehensive test suite for xxlang with 80%+ coverage across all core packages using a layer-by-layer approach.

## Current State

| Package | Coverage | Status |
|---------|----------|--------|
| `pkg/lexer` | 84.0% | Good |
| `pkg/module` | 97.4% | Excellent |
| `pkg/parser` | 71.7% | Needs improvement |
| `pkg/stdlib` | 57.6% | Needs improvement |
| `pkg/vm` | 55.0% | Needs improvement |
| `pkg/compiler` | 48.0% | Needs improvement |
| `pkg/objects` | 8.2% | Critical |
| `pkg/interpreter` | 0.0% | No tests |
| `pkg/plugin` | 0.0% | No tests |

## Target Coverage Goals

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| `pkg/lexer` | 84% | 90% | Low |
| `pkg/parser` | 72% | 85% | Medium |
| `pkg/compiler` | 48% | 80% | High |
| `pkg/vm` | 55% | 80% | High |
| `pkg/objects` | 8% | 70% | High |
| `pkg/stdlib` | 58% | 75% | Medium |
| `pkg/interpreter` | 0% | 80% | High |
| `pkg/plugin` | 0% | 70% | Medium |

## Test Structure

### Unit Tests (per package)

```
pkg/
├── lexer/
│   └── lexer_test.go      # +edge cases: unicode, escape sequences, errors
├── parser/
│   └── parser_test.go     # +error recovery, malformed input
├── compiler/
│   └── compiler_test.go   # +opcode verification, constant table
├── vm/
│   └── vm_test.go         # +opcode execution, stack boundaries
├── objects/
│   └── *_test.go          # +all type methods, conversions
├── interpreter/
│   └── interpreter_test.go # NEW - embedding API
└── plugin/
    └── plugin_test.go      # NEW - plugin interface
```

### Test Pattern

Each test follows this structure:
```go
func TestFeature(t *testing.T) {
    t.Run("happy path", func(t *testing.T) { ... })
    t.Run("edge cases", func(t *testing.T) { ... })
    t.Run("errors", func(t *testing.T) { ... })
}
```

## New Test Files

### 1. `pkg/interpreter/interpreter_test.go`

Test the embedding API:
- `TestNew` - Constructor with options (WithStdlib, WithGlobals)
- `TestEval` - Code evaluation, return values
- `TestEvalFile` - File execution
- `TestSetGlobal/GetGlobal` - Variable passing between Go and xxlang
- `TestToGo/FromGo` - Type conversion helpers
- `TestErrorHandling` - Compile errors, runtime errors
- `TestThreadSafety` - Concurrent access patterns

### 2. `pkg/plugin/plugin_test.go`

Test the plugin system:
- `TestPluginInterface` - Plugin struct, Name(), Exports()
- `TestRegister` - Plugin registration
- `TestLoad` - Loading .so files (mocked)
- `TestExportsConversion` - Converting Go functions to builtins

### 3. `pkg/objects/methods_test.go` (expand existing)

Test all type methods:
- Integer: toFloat, toStr, abs
- Float: toInt, toStr, abs, floor, ceil, round
- String: len, upper, lower, trim, split, contains, indexOf, startsWith, endsWith, toInt, toFloat
- Array: len, push, pop, first, last, indexOf, contains, reverse
- Map: len, keys, values, hasKey, delete

## Error Testing Strategy

### Categories

| Layer | Error Types |
|-------|-------------|
| Lexer | Invalid chars, unterminated strings, bad numbers |
| Parser | Syntax errors, unexpected tokens, missing elements |
| Compiler | Undefined variables, redeclaration, type mismatches |
| VM | Division by zero, index out of bounds, wrong arity |
| Interpreter | File not found, invalid input, timeout |

### Error Verification Pattern

```go
func TestParserErrors(t *testing.T) {
    tests := []struct {
        input       string
        expectError string // substring to match
    }{
        {"var x = ", "expected expression"},
        {"func f( { }", "expected ')'"},
        {"class { }", "expected class name"},
    }

    for _, tt := range tests {
        _, err := parse(tt.input)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), tt.expectError)
    }
}
```

### Recovery Testing

- Parser should continue after errors to report multiple issues
- VM should clean up stack after panics
- Interpreter should remain usable after errors

## Enhanced Existing Tests

### `pkg/compiler/compiler_test.go`

Add tests for:
- All opcodes generation
- Constant pooling and deduplication
- Source map generation
- Nested scopes and closures
- Class compilation
- Module compilation

### `pkg/vm/vm_test.go`

Add tests for:
- Stack boundary conditions
- Call frame management
- Error propagation
- Built-in function execution
- Class instantiation
- Method calls and super

### `pkg/stdlib/*_test.go`

Add missing function tests:
- `io`: printf, readLine
- `string`: all functions with edge cases
- `math`: all functions with edge cases
- `array`: all functions with edge cases
- `json`: parse, stringify with complex structures
- `regex`: all functions
- `crypto`: all functions

## Implementation Order

1. **Phase 1: Critical Gaps** (High Priority)
   - `pkg/interpreter/interpreter_test.go` (NEW)
   - `pkg/objects/` expanded tests
   - `pkg/compiler/` expanded tests
   - `pkg/vm/` expanded tests

2. **Phase 2: Medium Priority**
   - `pkg/plugin/plugin_test.go` (NEW)
   - `pkg/parser/` expanded tests
   - `pkg/stdlib/` expanded tests

3. **Phase 3: Polish** (Low Priority)
   - `pkg/lexer/` edge cases
   - Integration test additions

## Verification

After implementation:
```bash
# Run all tests
go test ./... -v

# Check coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Target: All packages at 70%+ coverage, core packages (lexer, parser, compiler, vm) at 80%+.
