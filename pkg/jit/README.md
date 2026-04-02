# JIT Compilation Support for Xxlang

This document describes the JIT (Just-In-Time) compilation support in Xxlang, including supported platforms, operations, limitations, and and best practices.

## Supported Platforms

| Platform | Support Level | Notes |
|----------|---------------|-------|
| Windows AMD64 | Full | Native x86-64 code generation |
| Linux AMD64 | Full | Native x86-64 code generation |
| macOS AMD64 | Full | Native x86-64 code generation |
| Linux ARM64 | Partial | Native ARM64 code generation (experimental) |
| macOS ARM64 | Partial | Native ARM64 code generation (experimental) |
| Other | None | Falls back to interpreter |

## Supported Operations

### Fully Supported (Native Execution)

- **Arithmetic**: `OpRegAdd`, `OpRegSub`, `OpRegMul`, `OpRegDiv`, `OpRegMod`, `OpRegNeg`
- **Comparison**: `OpRegLess`, `OpRegGreater`, `OpRegEqual`, `OpRegNotEqual`, `OpRegLessEqual`, `OpRegGreaterEqual`
- **Logical**: `OpRegAnd`, `OpRegOr`, `OpRegNot`
- **Data Movement**: `OpRegLoadConst`, `OpRegMove`, `OpRegLoadLocal`, `OpRegStoreLocal`, `OpRegLoadGlobal`, `OpRegStoreGlobal`
- **Control Flow**: `OpRegJump`, `OpRegJumpIfTrue`, `OpRegJumpIfFalse`, `OpRegReturn`
- **Literals**: `OpRegNull`, `OpRegTrue`, `OpRegFalse`
- **Increment/Decrement**: `OpRegIncLocal`, `OpRegDecLocal`
- **Loop Optimizations**: `OpRegLoopCountAdd`, `OpRegLoopBodyAdd`, `OpRegLoopIncCheck`

### Partially Supported (Fallback to Interpreter)

- **Function Calls**: `OpRegCall`, `OpRegTailCall` - Supported for self-recursive functions
- **Arrays**: `OpRegArray`, `OpRegArrayEmpty`, `OpRegArrayAppend`, `OpRegIndex`, `OpRegSetIndex` - Returns placeholder values
- **Maps**: `OpRegMap`, `OpRegMapEmpty`, `OpRegMapSet` - Returns placeholder values

### Not Supported (Always Falls Back to Interpreter)

- **Closures**: `OpRegClosure`, `OpRegLoadFree`, `OpRegStoreFree`
- **Builtin Calls**: `OpRegBuiltin` - Requires Go runtime callback
- **Methods**: `OpRegGetMethod`, `OpRegCallMethod`
- **Fields**: `OpRegGetField`, `OpRegSetField`
- **Classes**: `OpRegClass`, `OpRegNew`
- **Exceptions**: `OpRegThrow`, `OpRegPushHandler`, `OpRegPopHandler`
- **Modules**: `OpRegLoadModule`, `OpRegGetExport`, `OpRegSetExport`
- **Iterators**: `OpRegIterKey`, `OpRegIterValue`
- **Concurrency**: All concurrency operations (`OpRegRunStart`, `OpRegMakeTube`, etc.)

## Limitations

