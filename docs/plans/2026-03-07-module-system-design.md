# Module System Design

## Overview

A module system for Xxlang that supports code organization through import/export statements with relative path resolution and module caching.

## Syntax

### Import Statements

```xxl
// Default import - module as object
import math from "./math"
math.add(1, 2)

// Destructuring import - import specific exports
import { add, sub } from "./math"
add(1, 2)

// Namespace import - import all exports
import * as math from "./math"
math.add(1, 2)
```

### Export Statements

```xxl
// Declaration-style exports
export func add(a, b) { return a + b }
export func sub(a, b) { return a - b }
export var PI = 3.14159
export const VERSION = "1.0"
```

## Module Resolution

| Import Path | Resolves To |
|-------------|-------------|
| `import "./math"` | `./math.xxl` |
| `import "../utils"` | `../utils.xxl` |
| `import "std/math"` | Standard library (future) |

Resolution rules:
1. Relative paths (`./` or `../`) are resolved relative to the importing file
2. `.xxl` extension is optional
3. Bare imports (no `./` or `../`) reserved for standard library

## Module Object

Each module exports a Module object:

```go
type Module struct {
    Name    string
    Exports map[string]Object
}
```

## Circular Dependencies

Handling strategy: Cache + Partial availability

1. When a module is first imported, it's cached immediately (before execution)
2. If a circular import occurs, the cached (possibly incomplete) module is returned
3. Only already-exported symbols are available in circular scenarios

Example:
```xxl
// a.xxl
export var x = 1
import { y } from "./b"  // b may not have finished exporting

// b.xxl
export var y = 2
import { x } from "./a"  // a.x is available (value: 1)
```

## Implementation Architecture

```
pkg/module/
  ├── module.go      # Module type and exports
  ├── resolver.go    # Path resolution logic
  └── loader.go      # Module loading and caching

Modifications:
  - pkg/lexer/lexer.go    # Tokenize import/export (already done)
  - pkg/parser/parser.go  # Parse import/export statements
  - pkg/compiler/compiler.go # Compile module references
  - pkg/vm/vm.go          # Runtime module loading
```

## Execution Model

1. **Parse Phase**: Parser recognizes import/export statements
2. **Compile Phase**: Compiler resolves imports, creates module symbols
3. **Runtime Phase**: VM loads modules on-demand, caches results

## Module Cache

```go
type ModuleCache struct {
    modules map[string]*Module
    loading map[string]bool  // Track modules being loaded (cycle detection)
}
```

## Error Handling

- Module not found: `Error: module not found: "./math"`
- Circular dependency warning (optional, not error)
- Export not found: `Error: export "foo" not found in module "./math"`
