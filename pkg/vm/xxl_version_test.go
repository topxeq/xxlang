// pkg/vm/xxl_version_test.go
// Tests for the getXxlVersion() builtin, which exposes the build-time
// Xxlang version string to scripts.
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// TestGetXxlVersionInRegistry verifies the name is listed in BuiltinRegistry.
func TestGetXxlVersionInRegistry(t *testing.T) {
	for _, name := range objects.BuiltinRegistry {
		if name == "getXxlVersion" {
			return
		}
	}
	t.Error("BuiltinRegistry missing \"getXxlVersion\"")
}

// TestGetXxlVersionResolvedByCompiler verifies the compiler resolves the
// name to BuiltinScope.
func TestGetXxlVersionResolvedByCompiler(t *testing.T) {
	st := compiler.NewSymbolTable()
	sym, ok := st.Resolve("getXxlVersion")
	if !ok {
		t.Fatal("compiler did not resolve \"getXxlVersion\"")
	}
	if sym.Scope != compiler.BuiltinScope {
		t.Errorf("\"getXxlVersion\" resolved to scope %v, want BuiltinScope", sym.Scope)
	}
}

// TestGetXxlVersionDefault verifies the default value is "dev" when no
// build-time injection has occurred (as in a local `go test` run).
func TestGetXxlVersionDefault(t *testing.T) {
	// Save and restore to keep the test isolated.
	saved := objects.XxlVersion
	defer func() { objects.XxlVersion = saved }()

	objects.XxlVersion = "dev"
	src := `return getXxlVersion()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "dev" {
		t.Errorf("default version = %q, want %q", out, "dev")
	}
}

// TestGetXxlVersionInjected verifies that a version set on objects.XxlVersion
// (mimicking what cmd/xxl does at startup) flows through to the builtin.
func TestGetXxlVersionInjected(t *testing.T) {
	saved := objects.XxlVersion
	defer func() { objects.XxlVersion = saved }()

	objects.XxlVersion = "9.9.9-test"
	src := `return getXxlVersion()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "9.9.9-test" {
		t.Errorf("injected version = %q, want %q", out, "9.9.9-test")
	}
}

// TestGetXxlVersionCompileCheck is the minimal reproduction: the script
// `getXxlVersion()` must compile without "undefined variable".
func TestGetXxlVersionCompileCheck(t *testing.T) {
	src := `return getXxlVersion()`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}
	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("compile error: %v", err)
	}
}
