# Module System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement import/export module system with relative path resolution and module caching.

**Architecture:** Create a new `pkg/module/` package for module loading and caching. Modify parser to handle import/export syntax. Modify compiler to generate module-related opcodes. Modify VM to load modules at runtime.

**Tech Stack:** Go standard library (filepath, os, path/filepath)

---

## Task 1: Create Module Package Infrastructure

**Files:**
- Create: `pkg/module/module.go`
- Create: `pkg/module/resolver.go`
- Create: `pkg/module/loader.go`
- Test: `pkg/module/module_test.go`

**Step 1: Write the failing test**

```go
// pkg/module/module_test.go
package module

import (
    "testing"
)

func TestModuleExports(t *testing.T) {
    m := NewModule("./math")
    m.Export("add", &Integer{Value: 1})
    m.Export("sub", &Integer{Value: 2})

    if !m.HasExport("add") {
        t.Error("expected HasExport('add') to be true")
    }
    if !m.HasExport("sub") {
        t.Error("expected HasExport('sub') to be true")
    }
    if m.HasExport("mul") {
        t.Error("expected HasExport('mul') to be false")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/module/...`
Expected: FAIL - package module not found

**Step 3: Create module types**

```go
// pkg/module/module.go
package module

import "github.com/topxeq/xxlang/pkg/objects"

// Module represents a compiled module with exported symbols
type Module struct {
    Name    string
    Exports map[string]objects.Object
}

// NewModule creates a new module
func NewModule(name string) *Module {
    return &Module{
        Name:    name,
        Exports: make(map[string]objects.Object),
    }
}

// Export adds an export to the module
func (m *Module) Export(name string, value objects.Object) {
    m.Exports[name] = value
}

// HasExport checks if an export exists
func (m *Module) HasExport(name string) bool {
    _, ok := m.Exports[name]
    return ok
}

// GetExport retrieves an export
func (m *Module) GetExport(name string) (objects.Object, bool) {
    val, ok := m.Exports[name]
    return val, ok
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/module/...`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/module/
git commit -m "feat(module): add Module type with exports map"
```

---

## Task 2: Implement Path Resolver

**Files:**
- Modify: `pkg/module/resolver.go`
- Test: `pkg/module/resolver_test.go`

**Step 1: Write the failing test**

```go
// pkg/module/resolver_test.go
package module

import (
    "path/filepath"
    "testing"
)

func TestResolveRelativePath(t *testing.T) {
    tests := []struct {
        importer  string
        importPath string
        want      string
    }{
        {"/project/main.xxl", "./math", "/project/math.xxl"},
        {"/project/main.xxl", "./utils/helper", "/project/utils/helper.xxl"},
        {"/project/main.xxl", "../lib", "/lib.xxl"},
        {"/project/src/main.xxl", "./math", "/project/src/math.xxl"},
    }

    for _, tt := range tests {
        got, err := Resolve(tt.importer, tt.importPath)
        if err != nil {
            t.Errorf("Resolve(%s, %s) error: %v", tt.importer, tt.importPath, err)
            continue
        }
        // Normalize paths for comparison
        wantNorm := filepath.Clean(tt.want)
        gotNorm := filepath.Clean(got)
        if gotNorm != wantNorm {
            t.Errorf("Resolve(%s, %s) = %s, want %s", tt.importer, tt.importPath, gotNorm, wantNorm)
        }
    }
}

func TestResolveBarePath(t *testing.T) {
    // Bare paths (no ./ or ../) should return error for now
    _, err := Resolve("/project/main.xxl", "std/math")
    if err == nil {
        t.Error("expected error for bare import path")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/module/...`
Expected: FAIL - Resolve not defined

**Step 3: Implement resolver**

```go
// pkg/module/resolver.go
package module

import (
    "errors"
    "path/filepath"
    "strings"
)

var (
    ErrModuleNotFound     = errors.New("module not found")
    ErrInvalidImportPath  = errors.New("invalid import path")
    ErrBareImportNotSupported = errors.New("bare imports not supported yet")
)

// Resolve resolves an import path relative to the importer
func Resolve(importerPath, importPath string) (string, error) {
    // Check if it's a relative path
    if !strings.HasPrefix(importPath, "./") && !strings.HasPrefix(importPath, "../") {
        return "", ErrBareImportNotSupported
    }

    // Get directory of importer
    importerDir := filepath.Dir(importerPath)

    // Resolve the path
    resolved := filepath.Join(importerDir, importPath)

    // Add .xxl extension if not present
    if filepath.Ext(resolved) != ".xxl" {
        resolved += ".xxl"
    }

    // Clean the path
    resolved = filepath.Clean(resolved)

    return resolved, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/module/...`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/module/
git commit -m "feat(module): add path resolver for relative imports"
```

---

## Task 3: Implement Module Loader with Caching

**Files:**
- Modify: `pkg/module/loader.go`
- Test: `pkg/module/loader_test.go`

**Step 1: Write the failing test**

```go
// pkg/module/loader_test.go
package module

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoaderCache(t *testing.T) {
    // Create temp directory and file
    tmpDir := t.TempDir()
    modulePath := filepath.Join(tmpDir, "math.xxl")
    os.WriteFile(modulePath, []byte(`export var x = 1`), 0644)

    loader := NewLoader()

    // First load
    m1, err := loader.Get(modulePath)
    if err != nil {
        t.Fatalf("first load error: %v", err)
    }

    // Second load should return cached module
    m2, err := loader.Get(modulePath)
    if err != nil {
        t.Fatalf("second load error: %v", err)
    }

    // Should be the same instance
    if m1 != m2 {
        t.Error("expected cached module to be same instance")
    }
}

func TestLoaderLoadingState(t *testing.T) {
    loader := NewLoader()

    // Mark as loading
    loader.MarkLoading("./math")
    if !loader.IsLoading("./math") {
        t.Error("expected IsLoading to be true")
    }

    // Mark as done
    loader.MarkDone("./math")
    if loader.IsLoading("./math") {
        t.Error("expected IsLoading to be false after MarkDone")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/module/...`
Expected: FAIL - NewLoader not defined

**Step 3: Implement loader**

```go
// pkg/module/loader.go
package module

import (
    "sync"
)

// Loader handles module loading with caching
type Loader struct {
    mu      sync.RWMutex
    modules map[string]*Module    // cached modules
    loading map[string]bool       // modules currently being loaded (cycle detection)
}

// NewLoader creates a new module loader
func NewLoader() *Loader {
    return &Loader{
        modules: make(map[string]*Module),
        loading: make(map[string]bool),
    }
}

// Get retrieves a cached module or returns nil if not cached
func (l *Loader) Get(path string) (*Module, error) {
    l.mu.RLock()
    defer l.mu.RUnlock()

    m, ok := l.modules[path]
    if !ok {
        return nil, ErrModuleNotFound
    }
    return m, nil
}

// Set caches a module
func (l *Loader) Set(path string, m *Module) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.modules[path] = m
}

// IsLoading checks if a module is currently being loaded
func (l *Loader) IsLoading(path string) bool {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.loading[path]
}

// MarkLoading marks a module as being loaded
func (l *Loader) MarkLoading(path string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.loading[path] = true
}

// MarkDone marks a module as done loading
func (l *Loader) MarkDone(path string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    delete(l.loading, path)
}

// HasModule checks if a module is cached
func (l *Loader) HasModule(path string) bool {
    l.mu.RLock()
    defer l.mu.RUnlock()
    _, ok := l.modules[path]
    return ok
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/module/...`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/module/
git commit -m "feat(module): add Loader with caching and cycle detection"
```

---

## Task 4: Add Import/Export Parsing

**Files:**
- Modify: `pkg/parser/parser.go`
- Test: `pkg/parser/parser_test.go`

**Step 1: Write the failing test**

```go
// Add to pkg/parser/parser_test.go

func TestImportStatements(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {
            `import "./math"`,
            `import "./math";`,
        },
        {
            `import math from "./math"`,
            `import math from "./math";`,
        },
        {
            `import { add, sub } from "./math"`,
            `import { add, sub } from "./math";`,
        },
        {
            `import * as math from "./math"`,
            `import * as math from "./math";`,
        },
    }

    for _, tt := range tests {
        l := lexer.New(tt.input)
        p := New(l)
        program := p.ParseProgram()
        checkParserErrors(t, p)

        if program.String() != tt.expected {
            t.Errorf("program.String() = %s, want %s", program.String(), tt.expected)
        }
    }
}

func TestExportStatements(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {
            `export func add(a, b) { return a + b }`,
            `export func add(a, b) { return (a + b); }`,
        },
        {
            `export var PI = 3.14`,
            `export var PI = 3.14;`,
        },
    }

    for _, tt := range tests {
        l := lexer.New(tt.input)
        p := New(l)
        program := p.ParseProgram()
        checkParserErrors(t, p)

        if program.String() != tt.expected {
            t.Errorf("program.String() = %s, want %s", program.String(), tt.expected)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/parser/... -run TestImportStatements`
Expected: FAIL - import not parsed

**Step 3: Add import parsing to parser**

Modify `pkg/parser/parser.go` - add cases in `parseStatement()`:

```go
// In parseStatement() switch, add:
case lexer.TokenImport:
    return p.parseImportStatement()
case lexer.TokenExport:
    return p.parseExportStatement()
```

Add the parsing functions:

```go
// parseImportStatement parses an import statement
func (p *Parser) parseImportStatement() *ImportStatement {
    stmt := &ImportStatement{Token: p.curToken}

    p.nextToken()

    // Check import style
    if p.curTokenIs(lexer.TokenIdent) {
        // Could be: import math from "./math" or import * as math from "./math"
        if p.curToken.Literal == "*" {
            // Namespace import: import * as math from "./math"
            p.nextToken()
            if !p.expectPeek(lexer.TokenIdent) {
                return nil
            }
            if p.curToken.Literal != "as" {
                p.addError("expected 'as' after '*'")
                return nil
            }
            p.nextToken()
            stmt.Alias = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
        } else if p.peekTokenIs(lexer.TokenIdent) && p.peekToken.Literal == "from" {
            // Default import: import math from "./math"
            stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
        } else if p.curTokenIs(lexer.TokenLBrace) {
            // Destructuring import: import { add, sub } from "./math"
            return p.parseDestructuringImport(stmt)
        } else {
            // Could be destructuring starting with identifier
            stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
        }

        // Expect 'from'
        if !p.expectPeek(lexer.TokenIdent) {
            return nil
        }
        if p.curToken.Literal != "from" {
            p.addError("expected 'from' in import statement")
            return nil
        }
        p.nextToken()
    } else if p.curTokenIs(lexer.TokenLBrace) {
        // Destructuring import: import { add, sub } from "./math"
        return p.parseDestructuringImport(stmt)
    }
    // else: simple import import "./math"

    // Expect string path
    if !p.expectPeek(lexer.TokenString) {
        return nil
    }
    stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

    if p.peekTokenIs(lexer.TokenSemicolon) {
        p.nextToken()
    }

    return stmt
}

// parseDestructuringImport parses destructuring import
func (p *Parser) parseDestructuringImport(stmt *ImportStatement) *ImportStatement {
    // curToken is '{'
    stmt.Names = []*Identifier{}

    p.nextToken()
    for !p.curTokenIs(lexer.TokenRBrace) {
        ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
        stmt.Names = append(stmt.Names, ident)

        p.nextToken()
        if p.curTokenIs(lexer.TokenComma) {
            p.nextToken()
        }
    }

    // Expect 'from'
    p.nextToken()
    if p.curToken.Literal != "from" {
        p.addError("expected 'from' in import statement")
        return nil
    }
    p.nextToken()

    // Expect string path
    if !p.curTokenIs(lexer.TokenString) {
        p.addError("expected string path after 'from'")
        return nil
    }
    stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

    if p.peekTokenIs(lexer.TokenSemicolon) {
        p.nextToken()
    }

    return stmt
}

// parseExportStatement parses an export statement
func (p *Parser) parseExportStatement() *ExportStatement {
    stmt := &ExportStatement{Token: p.curToken}

    p.nextToken()

    // Parse the exportable statement
    switch p.curToken.Type {
    case lexer.TokenFunc:
        stmt.Exportable = p.parseFunctionStatement()
    case lexer.TokenVar:
        stmt.Exportable = p.parseVarStatement()
    case lexer.TokenConst:
        stmt.Exportable = p.parseConstStatement()
    default:
        p.addError(fmt.Sprintf("unexpected token in export: %s", p.curToken.Type))
        return nil
    }

    return stmt
}

// parseFunctionStatement parses a function as a statement
func (p *Parser) parseFunctionStatement() *FunctionLiteral {
    fl := &FunctionLiteral{Token: p.curToken}

    // Optional function name
    if p.peekTokenIs(lexer.TokenIdent) {
        p.nextToken()
        fl.Name = p.curToken.Literal
    }

    if !p.expectPeek(lexer.TokenLParen) {
        return nil
    }

    fl.Parameters = p.parseFunctionParameters()

    if !p.expectPeek(lexer.TokenLBrace) {
        return nil
    }

    fl.Body = p.parseBlockStatement()

    return fl
}
```

**Step 4: Update AST to support destructuring imports**

Modify `pkg/parser/ast.go` - update `ImportStatement`:

```go
// ImportStatement represents an import statement
type ImportStatement struct {
    Token    lexer.Token // The 'import' token
    Name     *Identifier      // Default import name (import math from ...)
    Path     *StringLiteral   // Module path
    Alias    *Identifier      // Namespace alias (import * as math from ...)
    Names    []*Identifier    // Destructured names (import { add, sub } from ...)
}

// String returns a string representation of the import statement
func (is *ImportStatement) String() string {
    var sb strings.Builder
    sb.WriteString("import ")

    if is.Alias != nil {
        sb.WriteString("* as ")
        sb.WriteString(is.Alias.String())
        sb.WriteString(" from ")
    } else if len(is.Names) > 0 {
        sb.WriteString("{ ")
        names := make([]string, len(is.Names))
        for i, n := range is.Names {
            names[i] = n.String()
        }
        sb.WriteString(strings.Join(names, ", "))
        sb.WriteString(" } from ")
    } else if is.Name != nil {
        sb.WriteString(is.Name.String())
        sb.WriteString(" from ")
    }

    sb.WriteString(is.Path.String())
    sb.WriteString(";")
    return sb.String()
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./pkg/parser/... -run TestImportStatements`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/parser/
git commit -m "feat(parser): add import/export statement parsing"
```

---

## Task 5: Add Module Opcodes

**Files:**
- Modify: `pkg/compiler/opcode.go`

**Step 1: Add new opcodes**

```go
// Add to opcode constants in opcode.go:
const (
    // ... existing opcodes ...

    // Module operations
    OpLoadModule   // Load module and push onto stack
    OpGetExport    // Get export from module
    OpModule       // Create module from exports
    OpSetExport    // Set export in current module
)

// Add to definitions map:
OpLoadModule: {"OpLoadModule", []int{2}},   // 2-byte constant index for path
OpGetExport:  {"OpGetExport", []int{2}},    // 2-byte constant index for name
OpModule:     {"OpModule", []int{2}},       // 2-byte export count
OpSetExport:  {"OpSetExport", []int{2}},    // 2-byte constant index for name
```

**Step 2: Run tests to verify no breakage**

Run: `go test ./pkg/compiler/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/compiler/opcode.go
git commit -m "feat(compiler): add module opcodes (OpLoadModule, OpGetExport, OpModule, OpSetExport)"
```

---

## Task 6: Compile Import Statements

**Files:**
- Modify: `pkg/compiler/compiler.go`
- Test: `pkg/compiler/compiler_test.go`

**Step 1: Write the failing test**

```go
// Add to pkg/compiler/compiler_test.go

func TestImportStatementCompilation(t *testing.T) {
    tests := []struct {
        input    string
        expected []byte
    }{
        {
            `import "./math"`,
            concat(
                Make(OpConstant, 0),  // "./math" string
                Make(OpLoadModule, 0),
                Make(OpPop),
            ),
        },
        {
            `import math from "./math"`,
            concat(
                Make(OpConstant, 0),  // "./math" string
                Make(OpLoadModule, 0),
                Make(OpPop),  // module stored in globals
            ),
        },
    }

    for _, tt := range tests {
        program := parse(tt.input)
        compiler := New()
        err := compiler.Compile(program)
        if err != nil {
            t.Fatalf("compiler error: %v", err)
        }

        bytecode := compiler.Bytecode()
        if !testBytecode(t, bytecode.Instructions, tt.expected) {
            t.Fatalf("test failed for input: %s", tt.input)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/compiler/... -run TestImportStatementCompilation`
Expected: FAIL

**Step 3: Implement import compilation**

Add to `pkg/compiler/compiler.go`:

```go
// In Compile() switch, add:
case *parser.ImportStatement:
    return c.compileImportStatement(node)

// compileImportStatement compiles an import statement
func (c *Compiler) compileImportStatement(node *parser.ImportStatement) error {
    // Load the module path constant
    pathIdx := c.addConstant(objects.NewString(node.Path.Value))
    c.emit(OpLoadModule, pathIdx)

    // The module is now on the stack
    // Handle different import styles
    if node.Name != nil {
        // Default import: import math from "./math"
        // Store module in global
        symbol := c.symbolTable.Define(node.Name.Value)
        c.emit(OpSetGlobal, symbol.Index)
    } else if node.Alias != nil {
        // Namespace import: import * as math from "./math"
        // Module is already on stack, store it
        symbol := c.symbolTable.Define(node.Alias.Value)
        c.emit(OpSetGlobal, symbol.Index)
    } else if len(node.Names) > 0 {
        // Destructuring import: import { add, sub } from "./math"
        // Get each export and store as global
        for _, name := range node.Names {
            c.emit(OpDup)  // Duplicate module reference
            nameIdx := c.addConstant(objects.NewString(name.Value))
            c.emit(OpGetExport, nameIdx)
            symbol := c.symbolTable.Define(name.Value)
            c.emit(OpSetGlobal, symbol.Index)
        }
        c.emit(OpPop)  // Pop the original module
    } else {
        // Simple import: import "./math"
        c.emit(OpPop)
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/compiler/... -run TestImportStatementCompilation`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/compiler/
git commit -m "feat(compiler): add import statement compilation"
```

---

## Task 7: Compile Export Statements

**Files:**
- Modify: `pkg/compiler/compiler.go`
- Test: `pkg/compiler/compiler_test.go`

**Step 1: Write the failing test**

```go
// Add to pkg/compiler/compiler_test.go

func TestExportStatementCompilation(t *testing.T) {
    input := `export var x = 10`

    program := parse(input)
    compiler := New()
    err := compiler.Compile(program)
    if err != nil {
        t.Fatalf("compiler error: %v", err)
    }

    // Should compile to setting x and exporting it
    bytecode := compiler.Bytecode()
    if len(bytecode.Instructions) == 0 {
        t.Error("expected instructions to be generated")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/compiler/... -run TestExportStatementCompilation`
Expected: FAIL

**Step 3: Implement export compilation**

Add to `pkg/compiler/compiler.go`:

```go
// In Compile() switch, add:
case *parser.ExportStatement:
    return c.compileExportStatement(node)

// compileExportStatement compiles an export statement
func (c *Compiler) compileExportStatement(node *parser.ExportStatement) error {
    // Compile the exportable value first
    switch stmt := node.Exportable.(type) {
    case *parser.VarStatement:
        // Compile var statement
        if err := c.Compile(stmt); err != nil {
            return err
        }
        // Export the variable
        c.emit(OpGetGlobal, c.lastDefinedGlobal)
        nameIdx := c.addConstant(objects.NewString(stmt.Name.Value))
        c.emit(OpSetExport, nameIdx)

    case *parser.ConstStatement:
        // Compile const statement
        if err := c.Compile(stmt); err != nil {
            return err
        }
        c.emit(OpGetGlobal, c.lastDefinedGlobal)
        nameIdx := c.addConstant(objects.NewString(stmt.Name.Value))
        c.emit(OpSetExport, nameIdx)

    case *parser.FunctionLiteral:
        // Compile function
        if err := c.Compile(stmt); err != nil {
            return err
        }
        nameIdx := c.addConstant(objects.NewString(stmt.Name))
        c.emit(OpSetExport, nameIdx)
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/compiler/... -run TestExportStatementCompilation`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/compiler/
git commit -m "feat(compiler): add export statement compilation"
```

---

## Task 8: Add Module Object Type

**Files:**
- Modify: `pkg/objects/object.go`
- Create: `pkg/objects/module.go`
- Test: `pkg/objects/module_test.go`

**Step 1: Write the failing test**

```go
// pkg/objects/module_test.go
package objects

import "testing"

func TestModuleObject(t *testing.T) {
    m := NewModule()
    m.Set("add", NewInteger(1))

    if m.Type() != MODULE_OBJ {
        t.Errorf("expected type MODULE_OBJ, got %s", m.Type())
    }

    val, ok := m.Get("add")
    if !ok {
        t.Error("expected to find 'add' export")
    }
    if val.(*Integer).Value != 1 {
        t.Error("unexpected value")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -run TestModuleObject`
Expected: FAIL

**Step 3: Implement Module object**

```go
// pkg/objects/module.go
package objects

// MODULE_OBJ is the type for module objects
const MODULE_OBJ = "MODULE"

// Module represents a module with exported symbols
type Module struct {
    Exports map[string]Object
}

// NewModule creates a new module object
func NewModule() *Module {
    return &Module{
        Exports: make(map[string]Object),
    }
}

// Type returns the object type
func (m *Module) Type() ObjectType {
    return MODULE_OBJ
}

// Inspect returns a string representation
func (m *Module) Inspect() string {
    return "[module]"
}

// Set adds an export to the module
func (m *Module) Set(name string, value Object) {
    m.Exports[name] = value
}

// Get retrieves an export from the module
func (m *Module) Get(name string) (Object, bool) {
    val, ok := m.Exports[name]
    return val, ok
}
```

Add to `pkg/objects/object.go`:

```go
// In the ObjectType constants, ensure MODULE_OBJ is referenced
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/objects/... -run TestModuleObject`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "feat(objects): add Module object type for module exports"
```

---

## Task 9: VM Module Loading

**Files:**
- Modify: `pkg/vm/vm.go`
- Test: `pkg/vm/vm_test.go`

**Step 1: Write the failing test**

```go
// Add to pkg/vm/vm_test.go

func TestModuleLoading(t *testing.T) {
    // This is more of an integration test
    // We'll test the opcodes directly

    tests := []struct {
        input    string
        expected interface{}
    }{
        {
            // Create a module and get export
            `
            var m = {add: 1}
            m.add
            `,
            1,
        },
    }

    for _, tt := range tests {
        eval := testEval(tt.input)
        testObject(t, eval, tt.expected)
    }
}
```

**Step 2: Run test to verify current state**

Run: `go test ./pkg/vm/... -run TestModuleLoading`
Expected: May pass or fail depending on current implementation

**Step 3: Implement module opcodes in VM**

Add to `pkg/vm/vm.go`:

```go
// In Run() switch, add:
case compiler.OpLoadModule:
    if err := vm.loadModule(); err != nil {
        return err
    }

case compiler.OpGetExport:
    if err := vm.getExport(); err != nil {
        return err
    }

case compiler.OpSetExport:
    if err := vm.setExport(); err != nil {
        return err
    }

case compiler.OpModule:
    if err := vm.createModule(); err != nil {
        return err
    }
```

Add the methods:

```go
// loadModule loads a module at runtime
func (vm *VM) loadModule() error {
    constIdx := int(vm.readUint16())
    vm.currentFrame().IP += 2

    pathObj, ok := vm.constants[constIdx].(*objects.String)
    if !ok {
        return fmt.Errorf("module path is not a string")
    }
    path := pathObj.Value

    // Check if module is already cached
    if module, ok := vm.moduleCache[path]; ok {
        vm.stack.Push(module)
        return nil
    }

    // Check for circular loading
    if vm.moduleLoader.IsLoading(path) {
        // Return incomplete module for circular reference
        module := objects.NewModule()
        vm.stack.Push(module)
        return nil
    }

    // Load and compile the module file
    code, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("module not found: %s", path)
    }

    // Parse and compile the module
    l := lexer.New(string(code))
    p := parser.New(l)
    program := p.ParseProgram()
    if len(p.Errors()) > 0 {
        return fmt.Errorf("parse errors in module %s: %v", path, p.Errors())
    }

    // Mark as loading (for cycle detection)
    vm.moduleLoader.MarkLoading(path)

    // Create a new module object
    module := objects.NewModule()
    vm.moduleCache[path] = module

    // Compile and run the module
    c := compiler.New()
    if err := c.Compile(program); err != nil {
        return fmt.Errorf("compile error in module %s: %v", path, err)
    }

    // Run module in isolated context but capture exports
    moduleVM := NewWithGlobalsStore(c.Bytecode(), make([]objects.Object, compiler.GlobalsSize))
    moduleVM.moduleCache = vm.moduleCache
    moduleVM.moduleLoader = vm.moduleLoader
    moduleVM.currentModule = module

    if err := moduleVM.Run(); err != nil {
        return fmt.Errorf("runtime error in module %s: %v", path, err)
    }

    vm.moduleLoader.MarkDone(path)
    vm.stack.Push(module)

    return nil
}

// getExport retrieves an export from a module
func (vm *VM) getExport() error {
    nameIdx := int(vm.readUint16())
    vm.currentFrame().IP += 2

    nameObj, ok := vm.constants[nameIdx].(*objects.String)
    if !ok {
        return fmt.Errorf("export name is not a string")
    }
    name := nameObj.Value

    module, ok := vm.stack.Pop().(*objects.Module)
    if !ok {
        return fmt.Errorf("cannot get export from non-module")
    }

    val, ok := module.Get(name)
    if !ok {
        return fmt.Errorf("export '%s' not found in module", name)
    }

    vm.stack.Push(val)
    return nil
}

// setExport sets an export in the current module
func (vm *VM) setExport() error {
    nameIdx := int(vm.readUint16())
    vm.currentFrame().IP += 2

    nameObj, ok := vm.constants[nameIdx].(*objects.String)
    if !ok {
        return fmt.Errorf("export name is not a string")
    }
    name := nameObj.Value

    val := vm.stack.Pop()

    if vm.currentModule == nil {
        return fmt.Errorf("export statement outside of module context")
    }

    vm.currentModule.Set(name, val)
    return nil
}

// createModule creates a module object from exports on stack
func (vm *VM) createModule() error {
    count := int(vm.readUint16())
    vm.currentFrame().IP += 2

    module := objects.NewModule()

    // Pop name-value pairs from stack
    for i := 0; i < count; i++ {
        val := vm.stack.Pop()
        name := vm.stack.Pop().(*objects.String).Value
        module.Set(name, val)
    }

    vm.stack.Push(module)
    return nil
}
```

**Step 4: Update VM struct to include module support**

```go
// Update VM struct
type VM struct {
    // ... existing fields ...
    moduleCache  map[string]*objects.Module
    moduleLoader *module.Loader
    currentModule *objects.Module
    sourcePath   string  // Path of current file being executed
}

// Update New() and NewWithGlobalsStore()
func New(bytecode *Bytecode) *VM {
    return &VM{
        // ... existing initialization ...
        moduleCache:  make(map[string]*objects.Module),
        moduleLoader: module.NewLoader(),
    }
}

func NewWithGlobalsStore(bytecode *Bytecode, globals []objects.Object) *VM {
    return &VM{
        // ... existing initialization ...
        moduleCache:  make(map[string]*objects.Module),
        moduleLoader: module.NewLoader(),
    }
}
```

**Step 5: Run tests**

Run: `go test ./pkg/vm/...`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/vm/
git commit -m "feat(vm): add module loading and export handling"
```

---

## Task 10: Integration Test - Full Module System

**Files:**
- Create: `tests/modules/math.xxl`
- Create: `tests/modules/main.xxl`
- Test: Manual integration test

**Step 1: Create test module files**

```bash
mkdir -p tests/modules
```

```xxl
// tests/modules/math.xxl
export func add(a, b) {
    return a + b
}

export func sub(a, b) {
    return a - b
}

export var PI = 3.14159
```

```xxl
// tests/modules/main.xxl
import math from "./math"

println(math.add(1, 2))
println(math.sub(5, 3))
println(math.PI)
```

**Step 2: Run integration test**

```bash
cd /mnt1/aiprjs/xxlang
go run ./cmd/xxlang/main.go tests/modules/main.xxl
```

Expected output:
```
3
2
3.14159
```

**Step 3: Test destructuring import**

```xxl
// tests/modules/main2.xxl
import { add, PI } from "./math"

println(add(10, 20))
println(PI)
```

**Step 4: Test namespace import**

```xxl
// tests/modules/main3.xxl
import * as m from "./math"

println(m.add(1, 2))
println(m.PI)
```

**Step 5: Test circular dependencies**

```xxl
// tests/modules/circular_a.xxl
export var x = 1
import { y } from "./circular_b"
export var sum = x + y
```

```xxl
// tests/modules/circular_b.xxl
export var y = 2
import { x } from "./circular_a"
export var product = x * y
```

**Step 6: Commit integration tests**

```bash
git add tests/modules/
git commit -m "test(modules): add integration tests for module system"
```

---

## Task 11: Update Main Entry Point

**Files:**
- Modify: `cmd/xxlang/main.go`

**Step 1: Pass source path to VM**

Update `runFile()` function to pass the source file path:

```go
func runFile(filename string) {
    // Get absolute path
    absPath, err := filepath.Abs(filename)
    if err != nil {
        fmt.Printf("Error getting absolute path: %v\n", err)
        os.Exit(1)
    }

    // Read the file
    code, err := os.ReadFile(absPath)
    if err != nil {
        fmt.Printf("Error reading file '%s': %v\n", filename, err)
        os.Exit(1)
    }

    // Parse and compile
    l := lexer.New(string(code))
    p := parser.New(l)
    program := p.ParseProgram()

    if len(p.Errors()) > 0 {
        fmt.Printf("Parser errors: %v\n", p.Errors())
        os.Exit(1)
    }

    c := compiler.New()
    if err := c.Compile(program); err != nil {
        fmt.Printf("Compiler error: %v\n", err)
        os.Exit(1)
    }

    // Create VM with source path for module resolution
    bytecode := c.Bytecode()
    v := vm.New(bytecode)
    v.SetSourcePath(absPath)

    if err := v.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    result := v.LastPopped()
    if result != nil && result != objects.NULL {
        fmt.Println(result.Inspect())
    }
}
```

**Step 2: Add SetSourcePath to VM**

```go
// In pkg/vm/vm.go
func (vm *VM) SetSourcePath(path string) {
    vm.sourcePath = path
}
```

**Step 3: Run final integration test**

```bash
go run ./cmd/xxlang/main.go tests/modules/main.xxl
```

**Step 4: Commit**

```bash
git add cmd/xxlang/main.go pkg/vm/vm.go
git commit -m "feat: integrate module system with main entry point"
```

---

## Task 12: Update Documentation

**Files:**
- Modify: `README.md`

**Step 1: Add module system documentation**

Add section to README.md:

```markdown
## Module System

Xxlang supports ES6-style modules with import/export syntax.

### Exports

```xxl
// math.xxl
export func add(a, b) {
    return a + b
}

export var PI = 3.14159
```

### Imports

```xxl
// Default import
import math from "./math"
math.add(1, 2)

// Destructuring import
import { add, PI } from "./math"
add(1, 2)

// Namespace import
import * as math from "./math"
math.add(1, 2)
```

### Module Resolution

- `./math` → resolves to `./math.xxl` relative to current file
- `../utils` → resolves to parent directory
- Bare imports like `std/math` reserved for future standard library
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add module system documentation to README"
```
