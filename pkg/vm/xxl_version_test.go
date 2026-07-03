// pkg/vm/xxl_version_test.go
// Tests for the getXxlVersion() and getXxlBuildNumber() builtins, which
// expose the Xxlang version string and build number to scripts.
//
// Since v0.9.7 both values live in pkg/version as the single source of
// truth; these tests verify the builtins read from that package.
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/version"
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

// TestGetXxlVersionDefault verifies that the builtin returns the value
// defined in pkg/version (the single source of truth).
func TestGetXxlVersionDefault(t *testing.T) {
	src := `return getXxlVersion()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != version.Version {
		t.Errorf("default version = %q, want %q (version.Version)", out, version.Version)
	}
}

// TestGetXxlVersionOverride verifies that a value set on version.Version
// flows through to the builtin. (Production code must never reassign this;
// the override here is only for test isolation.)
func TestGetXxlVersionOverride(t *testing.T) {
	saved := version.Version
	defer func() { version.Version = saved }()

	version.Version = "9.9.9-test"
	src := `return getXxlVersion()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "9.9.9-test" {
		t.Errorf("overridden version = %q, want %q", out, "9.9.9-test")
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

// TestGetXxlBuildNumberDefault verifies that the builtin returns the value
// defined in pkg/version (the single source of truth).
func TestGetXxlBuildNumberDefault(t *testing.T) {
	src := `return getXxlBuildNumber()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != version.BuildNumber {
		t.Errorf("default build number = %q, want %q (version.BuildNumber)", out, version.BuildNumber)
	}
}

// TestGetXxlBuildNumberOverride verifies that a value set on
// version.BuildNumber flows through to the builtin.
func TestGetXxlBuildNumberOverride(t *testing.T) {
	saved := version.BuildNumber
	defer func() { version.BuildNumber = saved }()

	version.BuildNumber = "2026070301"
	src := `return getXxlBuildNumber()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "2026070301" {
		t.Errorf("overridden build number = %q, want %q", out, "2026070301")
	}
}

// TestGetXxlVersionAndBuildNumberCombined verifies the common health.xxl
// usage: concatenating version and build number to match CLI output.
func TestGetXxlVersionAndBuildNumberCombined(t *testing.T) {
	savedV := version.Version
	savedB := version.BuildNumber
	defer func() {
		version.Version = savedV
		version.BuildNumber = savedB
	}()

	version.Version = "0.9.7"
	version.BuildNumber = "2026070302"
	src := `return "v" + getXxlVersion() + "." + getXxlBuildNumber()`
	out, err := runRegScriptReturn(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "v0.9.7.2026070302"
	if out != want {
		t.Errorf("combined = %q, want %q", out, want)
	}
}
