// pkg/vm/xxl_version_test.go
// Tests for the getXxlVersion() builtin, which exposes the Xxlang version
// string to scripts.
//
// Since v0.9.7 the version lives in pkg/version as the single source of
// truth. In v0.9.8 the BuildNumber was removed entirely; these tests cover
// the remaining getXxlVersion() builtin.
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

// TestGetXxlBuildNumberRemoved verifies that the getXxlBuildNumber builtin
// (removed in v0.9.8) is no longer in the registry and no longer resolves.
// Scripts that previously called it must now fail to compile.
func TestGetXxlBuildNumberRemoved(t *testing.T) {
	for _, name := range objects.BuiltinRegistry {
		if name == "getXxlBuildNumber" {
			t.Fatal("BuiltinRegistry still contains \"getXxlBuildNumber\" (should have been removed in v0.9.8)")
		}
	}
	st := compiler.NewSymbolTable()
	if _, ok := st.Resolve("getXxlBuildNumber"); ok {
		t.Fatal("compiler still resolves \"getXxlBuildNumber\" (should have been removed in v0.9.8)")
	}
}
