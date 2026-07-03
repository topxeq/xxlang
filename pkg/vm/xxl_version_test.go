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

// TestGetXxlBuildNumberInRegistry verifies the name is listed in BuiltinRegistry.
func TestGetXxlBuildNumberInRegistry(t *testing.T) {
	for _, name := range objects.BuiltinRegistry {
		if name == "getXxlBuildNumber" {
			return
		}
	}
	t.Error("BuiltinRegistry missing \"getXxlBuildNumber\"")
}

// TestGetXxlBuildNumberResolvedByCompiler verifies the compiler resolves the
// name to BuiltinScope.
func TestGetXxlBuildNumberResolvedByCompiler(t *testing.T) {
	st := compiler.NewSymbolTable()
	sym, ok := st.Resolve("getXxlBuildNumber")
	if !ok {
		t.Fatal("compiler did not resolve \"getXxlBuildNumber\"")
	}
	if sym.Scope != compiler.BuiltinScope {
		t.Errorf("\"getXxlBuildNumber\" resolved to scope %v, want BuiltinScope", sym.Scope)
	}
}

// TestGetXxlBuildNumberDefault verifies the default value is "0" when no
// value has been propagated from cmd/xxl.
func TestGetXxlBuildNumberDefault(t *testing.T) {
	saved := objects.XxlBuildNumber
	defer func() { objects.XxlBuildNumber = saved }()

	objects.XxlBuildNumber = "0"
	src := `return getXxlBuildNumber()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "0" {
		t.Errorf("default build number = %q, want %q", out, "0")
	}
}

// TestGetXxlBuildNumberInjected verifies that a value set on
// objects.XxlBuildNumber (mimicking what cmd/xxl does at startup) flows
// through to the builtin.
func TestGetXxlBuildNumberInjected(t *testing.T) {
	saved := objects.XxlBuildNumber
	defer func() { objects.XxlBuildNumber = saved }()

	objects.XxlBuildNumber = "2026070301"
	src := `return getXxlBuildNumber()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "2026070301" {
		t.Errorf("injected build number = %q, want %q", out, "2026070301")
	}
}

// TestGetXxlVersionAndBuildNumberCombined verifies the common health.xxl
// usage: concatenating version and build number to match CLI output.
func TestGetXxlVersionAndBuildNumberCombined(t *testing.T) {
	savedV := objects.XxlVersion
	savedB := objects.XxlBuildNumber
	defer func() {
		objects.XxlVersion = savedV
		objects.XxlBuildNumber = savedB
	}()

	objects.XxlVersion = "0.9.5"
	objects.XxlBuildNumber = "2026070301"
	src := `return "v" + getXxlVersion() + "." + getXxlBuildNumber()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "v0.9.5.2026070301"
	if out != want {
		t.Errorf("combined = %q, want %q", out, want)
	}
}
