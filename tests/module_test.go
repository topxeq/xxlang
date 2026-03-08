// tests/module_test.go
// Integration tests for the module system (import/export)
package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// ============================================
// Test Module Setup
// ============================================

// createTestModule creates a temporary module file for testing
func createTestModule(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test module: %v", err)
	}
	return path
}

// setupTestModulesDir creates a temporary directory with test modules
func setupTestModulesDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create math.xxl module
	createTestModule(t, dir, "math.xxl", `
export func add(a, b) {
    return a + b
}

export func sub(a, b) {
    return a - b
}

export func mul(a, b) {
    return a * b
}

export func div(a, b) {
    return a / b
}

export var PI = 3.14159
export var E = 2.71828
`)

	// Create utils.xxl module
	createTestModule(t, dir, "utils.xxl", `
export func identity(x) {
    return x
}

export func double(x) {
    return x * 2
}

export func square(x) {
    return x * x
}
`)

	// Create strings.xxl module
	createTestModule(t, dir, "strings.xxl", `
export func concat(a, b) {
    return a + b
}

export func repeat(s, n) {
    var result = ""
    for (var i = 0; i < n; i = i + 1) {
        result = result + s
    }
    return result
}

export var EMPTY = ""
`)

	// Create a module that imports another module
	createTestModule(t, dir, "calculator.xxl", `
import "./math.xxl"

export func calculate(a, b, op) {
    if (op == "add") { return add(a, b) }
    if (op == "sub") { return sub(a, b) }
    if (op == "mul") { return mul(a, b) }
    if (op == "div") { return div(a, b) }
    return null
}
`)

	return dir
}

// ============================================
// Module System Tests
// ============================================

func TestModuleExportFunction(t *testing.T) {
	// Test that exported functions are accessible
	// Note: This tests the module structure, actual import requires VM integration

	loader := module.NewLoader()

	// Create a module with exported functions
	m := module.NewModule("./test")
	m.Exports["add"] = &objects.CompiledFunction{
		NumParameters: 2,
		Instructions:  []byte{}, // Empty instructions for test
	}
	m.Exports["PI"] = &objects.Float{Value: 3.14159}

	loader.Set("./test", m)

	// Verify module is cached
	if !loader.HasModule("./test") {
		t.Error("expected module to be cached")
	}

	// Retrieve module
	retrieved, err := loader.Get("./test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.Name != "./test" {
		t.Errorf("expected name ./test, got %s", retrieved.Name)
	}
}

func TestModuleCycleDetection(t *testing.T) {
	loader := module.NewLoader()

	// Mark module as loading
	loader.MarkLoading("./circular")

	if !loader.IsLoading("./circular") {
		t.Error("expected module to be marked as loading")
	}

	// Mark as done
	loader.MarkDone("./circular")

	if loader.IsLoading("./circular") {
		t.Error("expected module to no longer be loading")
	}
}

func TestModuleLoaderCache(t *testing.T) {
	loader := module.NewLoader()

	m := module.NewModule("./cached")
	m.Exports["value"] = &objects.Int{Value: 42}

	loader.Set("./cached", m)

	// Get should return same instance
	m1, err := loader.Get("./cached")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m2, err := loader.Get("./cached")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m1 != m2 {
		t.Error("expected same module instance from cache")
	}
}

func TestModuleNotFound(t *testing.T) {
	loader := module.NewLoader()

	_, err := loader.Get("./nonexistent")
	if err != module.ErrModuleNotFound {
		t.Errorf("expected ErrModuleNotFound, got %v", err)
	}
}

// ============================================
// Import Statement Parsing
// ============================================

func TestImportStatementParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`import "./math.xxl"`, "./math.xxl"},
		{`import "../utils.xxl"`, "../utils.xxl"},
		{`import "/abs/path/module.xxl"`, "/abs/path/module.xxl"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if len(program.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(program.Statements))
		}
	}
}

func TestExportStatementParsing(t *testing.T) {
	tests := []string{
		`export func add(a, b) { return a + b }`,
		`export var PI = 3.14159`,
		`export const E = 2.71828`,
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors for %q: %v", tt, p.Errors())
		}

		if len(program.Statements) != 1 {
			t.Fatalf("expected 1 statement for %q, got %d", tt, len(program.Statements))
		}
	}
}

// ============================================
// Module Integration with VM
// ============================================

func TestModuleExportsInObject(t *testing.T) {
	// Test that module exports can be represented as objects
	exports := make(map[string]objects.Object)
	exports["add"] = &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NULL
			}
			a, ok1 := args[0].(*objects.Int)
			b, ok2 := args[1].(*objects.Int)
			if !ok1 || !ok2 {
				return objects.NULL
			}
			return &objects.Int{Value: a.Value + b.Value}
		},
	}
	exports["PI"] = &objects.Float{Value: 3.14159}

	// Create module
	m := module.NewModule("./math")
	m.Exports = exports

	// Verify exports are accessible
	add, ok := m.Exports["add"]
	if !ok {
		t.Fatal("expected add export")
	}

	// Test calling the builtin
	builtin, ok := add.(*objects.Builtin)
	if !ok {
		t.Fatal("expected builtin function")
	}

	result := builtin.Fn(&objects.Int{Value: 3}, &objects.Int{Value: 4})
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatal("expected int result")
	}
	if intResult.Value != 7 {
		t.Errorf("expected 7, got %d", intResult.Value)
	}
}

// ============================================
// Full Module Loading Test
// ============================================

func TestModuleFullPipeline(t *testing.T) {
	// This test verifies that module definitions can be compiled and executed
	// It simulates what happens when a module is loaded (without export keywords)

	moduleCode := `
func add(a, b) {
    return a + b
}

var PI = 3
add(2, 3)
`

	// Parse the module
	l := lexer.New(moduleCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Compile the module
	c := compiler.New()
	if err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	// Run the module to populate exports
	bytecode := c.Bytecode()
	globals := make([]objects.Object, compiler.GlobalsSize)
	v := vm.NewWithGlobalsStore(bytecode, globals)

	if err := v.Run(); err != nil {
		t.Fatalf("vm error: %v", err)
	}

	// Verify the result
	result := v.LastPopped()
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("expected 5, got %d", intResult.Value)
	}
}

// ============================================
// Module Resolution Tests
// ============================================

func TestModuleResolutionRelative(t *testing.T) {
	tests := []struct {
		importerPath string
		importPath   string
	}{
		{"/project/main.xxl", "./math.xxl"},
		{"/project/main.xxl", "./utils/math.xxl"},
		{"/project/sub/main.xxl", "../parent.xxl"},
	}

	for _, tt := range tests {
		resolved, err := module.Resolve(tt.importerPath, tt.importPath)
		if err != nil {
			t.Errorf("failed to resolve %s: %v", tt.importPath, err)
		}
		if resolved == "" {
			t.Errorf("empty resolved path for %s", tt.importPath)
		}
	}
}

func TestModuleResolutionWithExtension(t *testing.T) {
	// Test that .xxl extension is added if not present
	resolved, err := module.Resolve("/project/main.xxl", "./math")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != "/project/math.xxl" {
		t.Errorf("expected /project/math.xxl, got %s", resolved)
	}
}

func TestModuleResolutionBareImport(t *testing.T) {
	// Bare imports (without ./ or ../) should return an error
	_, err := module.Resolve("/project/main.xxl", "stdlib/math")
	if err != module.ErrBareImportNotSupported {
		t.Errorf("expected ErrBareImportNotSupported, got %v", err)
	}
}

// ============================================
// Test with Actual Module Files
// ============================================

func TestModuleFileLoading(t *testing.T) {
	// Create temporary module directory
	dir := t.TempDir()

	// Create a simple module
	modulePath := filepath.Join(dir, "simple.xxl")
	moduleContent := `
export func getValue() {
    return 42
}
`
	if err := os.WriteFile(modulePath, []byte(moduleContent), 0644); err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	// Read and verify module content
	content, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatalf("failed to read module: %v", err)
	}

	// Parse the module to verify syntax
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Errorf("expected 1 statement, got %d", len(program.Statements))
	}
}
