// pkg/stdlib/stdlib_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestRegistry(t *testing.T) {
	modules := []string{"math", "strings", "array", "io"}

	for _, name := range modules {
		if !Has(name) {
			t.Errorf("module %s not registered", name)
		}

		mod := Get(name)
		if mod == nil {
			t.Errorf("module %s is nil", name)
			continue
		}

		if mod.Name != name {
			t.Errorf("module name mismatch: got %s, want %s", mod.Name, name)
		}

		if len(mod.Exports) == 0 {
			t.Errorf("module %s has no exports", name)
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	f := Float(3.14)
	if f.Value != 3.14 {
		t.Errorf("Float() = %v, want 3.14", f.Value)
	}

	s := String("hello")
	if s.Value != "hello" {
		t.Errorf("String() = %v, want hello", s.Value)
	}

	i := Int(42)
	if i.Value != 42 {
		t.Errorf("Int() = %v, want 42", i.Value)
	}

	if !Bool(true).Value {
		t.Error("Bool(true) should be true")
	}
	if Bool(false).Value {
		t.Error("Bool(false) should be false")
	}

	arr := Array(Int(1), Int(2), Int(3))
	if len(arr.Elements) != 3 {
		t.Errorf("Array() length = %d, want 3", len(arr.Elements))
	}

	if Null() != objects.NULL {
		t.Error("Null() should return objects.NULL")
	}

	err := Error("test error")
	if err.Message != "test error" {
		t.Errorf("Error() = %v, want 'test error'", err.Message)
	}
}

func TestBuiltinFunc(t *testing.T) {
	called := false
	fn := BuiltinFunc(func(args ...objects.Object) objects.Object {
		called = true
		return Int(int64(len(args)))
	})

	if fn.Fn == nil {
		t.Fatal("BuiltinFunc created nil function")
	}

	result := fn.Fn(Int(1), Int(2), Int(3))
	if !called {
		t.Error("function was not called")
	}

	if result.(*objects.Int).Value != 3 {
		t.Errorf("function returned %v, want 3", result)
	}
}

func TestAllModules(t *testing.T) {
	// Test that all expected modules are registered
	expectedModules := []string{
		"math", "strings", "array", "io", "json", "time",
		"crypto", "regex", "fmt", "os", "env", "log",
		"encoding", "strconv", "text", "uuid", "debug",
		"collections", "bytes", "csv", "sort", "net", "fp",
	}

	for _, name := range expectedModules {
		if !Has(name) {
			t.Errorf("module %s not registered", name)
		}
	}
}

func TestEnvModule(t *testing.T) {
	mod := Get("env")
	if mod == nil {
		t.Fatal("env module not found")
	}

	// Test args function
	argsFn, ok := mod.Exports["args"]
	if !ok {
		t.Fatal("args function not found in env module")
	}

	builtin, ok := argsFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected *Builtin, got %T", argsFn)
	}

	result := builtin.Fn()
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) < 1 {
		t.Error("args() should return at least one element")
	}
}

func TestSetScriptArgs(t *testing.T) {
	// Test SetScriptArgs function
	testArgs := []string{"test1", "test2"}
	SetScriptArgs(testArgs)

	mod := Get("env")
	if mod == nil {
		t.Fatal("env module not found")
	}

	// Test scriptArgs function (not args)
	scriptArgsFn, ok := mod.Exports["scriptArgs"]
	if !ok {
		t.Fatal("scriptArgs function not found")
	}

	builtin, ok := scriptArgsFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected *Builtin, got %T", scriptArgsFn)
	}

	result := builtin.Fn()
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 args, got %d", len(arr.Elements))
	}
}

func callEnvFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("env")
	if mod == nil {
		panic("env module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestEnvGetSet(t *testing.T) {
	// Test set and get
	result := callEnvFunc("set", String("XXLANG_TEST_VAR"), String("test_value"))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("set() should return Null, got %T", result)
	}

	result = callEnvFunc("get", String("XXLANG_TEST_VAR"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("get() should return String, got %T", result)
	}
	if s.Value != "test_value" {
		t.Errorf("get() = %s, want 'test_value'", s.Value)
	}

	// Test getOr with existing key
	result = callEnvFunc("getOr", String("XXLANG_TEST_VAR"), String("default"))
	s, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("getOr() should return String, got %T", result)
	}
	if s.Value != "test_value" {
		t.Errorf("getOr() = %s, want 'test_value'", s.Value)
	}

	// Test getOr with non-existing key
	result = callEnvFunc("getOr", String("XXLANG_NONEXISTENT_VAR"), String("default_value"))
	s, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("getOr() should return String, got %T", result)
	}
	if s.Value != "default_value" {
		t.Errorf("getOr() = %s, want 'default_value'", s.Value)
	}

	// Clean up
	callEnvFunc("unset", String("XXLANG_TEST_VAR"))
}

func TestEnvHas(t *testing.T) {
	// Set a test variable
	callEnvFunc("set", String("XXLANG_HAS_TEST"), String("value"))

	// Test has with existing key
	result := callEnvFunc("has", String("XXLANG_HAS_TEST"))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("has() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("has() should return true for existing key")
	}

	// Test has with non-existing key
	result = callEnvFunc("has", String("XXLANG_NONEXISTENT"))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("has() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("has() should return false for non-existing key")
	}

	// Clean up
	callEnvFunc("unset", String("XXLANG_HAS_TEST"))
}

func TestEnvAll(t *testing.T) {
	result := callEnvFunc("all")
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("all() should return Array, got %T", result)
	}
	if len(arr.Elements) == 0 {
		t.Error("all() should return at least one environment variable")
	}
}

func TestEnvMap(t *testing.T) {
	result := callEnvFunc("map")
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("map() should return Map, got %T", result)
	}
	if len(m.Pairs) == 0 {
		t.Error("map() should return at least one environment variable")
	}
}

func TestEnvExpand(t *testing.T) {
	// Set a test variable
	callEnvFunc("set", String("XXLANG_EXPAND_TEST"), String("expanded_value"))

	result := callEnvFunc("expand", String("value is ${XXLANG_EXPAND_TEST}"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expand() should return String, got %T", result)
	}
	if s.Value != "value is expanded_value" {
		t.Errorf("expand() = %s, want 'value is expanded_value'", s.Value)
	}

	// Clean up
	callEnvFunc("unset", String("XXLANG_EXPAND_TEST"))
}

func TestEnvCwd(t *testing.T) {
	result := callEnvFunc("cwd")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("cwd() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("cwd() should return non-empty string")
	}
}

func TestEnvPid(t *testing.T) {
	result := callEnvFunc("pid")
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("pid() should return Int, got %T", result)
	}
	if i.Value <= 0 {
		t.Errorf("pid() = %d, should be positive", i.Value)
	}
}

func TestEnvPpid(t *testing.T) {
	result := callEnvFunc("ppid")
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("ppid() should return Int, got %T", result)
	}
	if i.Value < 0 {
		t.Errorf("ppid() = %d, should be non-negative", i.Value)
	}
}

func TestEnvExe(t *testing.T) {
	result := callEnvFunc("exe")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("exe() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("exe() should return non-empty string")
	}
}

func TestEnvStreams(t *testing.T) {
	result := callEnvFunc("streams")
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("streams() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("streams() should return 3 elements, got %d", len(arr.Elements))
	}
}

func TestEnvLookup(t *testing.T) {
	// Set a test variable
	callEnvFunc("set", String("XXLANG_LOOKUP_TEST"), String("lookup_value"))

	result := callEnvFunc("lookup", String("XXLANG_LOOKUP_TEST"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("lookup() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("lookup() should return 2 elements, got %d", len(arr.Elements))
	}
	// First element should be true (exists)
	if b, ok := arr.Elements[0].(*objects.Bool); !ok || !b.Value {
		t.Error("lookup()[0] should be true for existing key")
	}

	// Clean up
	callEnvFunc("unset", String("XXLANG_LOOKUP_TEST"))
}

func TestEnvGetInt(t *testing.T) {
	// Set a test variable with integer value
	callEnvFunc("set", String("XXLANG_INT_TEST"), String("42"))

	result := callEnvFunc("getInt", String("XXLANG_INT_TEST"))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("getInt() should return Int, got %T", result)
	}
	if i.Value != 42 {
		t.Errorf("getInt() = %d, want 42", i.Value)
	}

	// Test with default value for non-existing key
	result = callEnvFunc("getInt", String("XXLANG_NONEXISTENT_INT"), Int(99))
	i, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("getInt() should return Int, got %T", result)
	}
	if i.Value != 99 {
		t.Errorf("getInt() with default = %d, want 99", i.Value)
	}

	// Clean up
	callEnvFunc("unset", String("XXLANG_INT_TEST"))
}

func TestEnvGetBool(t *testing.T) {
	// Set test variables with boolean values
	callEnvFunc("set", String("XXLANG_BOOL_TRUE"), String("true"))
	callEnvFunc("set", String("XXLANG_BOOL_FALSE"), String("false"))

	result := callEnvFunc("getBool", String("XXLANG_BOOL_TRUE"))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("getBool() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("getBool() should return true for 'true' value")
	}

	result = callEnvFunc("getBool", String("XXLANG_BOOL_FALSE"))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("getBool() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("getBool() should return false for 'false' value")
	}

	// Clean up
	callEnvFunc("unset", String("XXLANG_BOOL_TRUE"))
	callEnvFunc("unset", String("XXLANG_BOOL_FALSE"))
}

func TestEnvMixArgs(t *testing.T) {
	// Test mixArgs when scriptArgs is set
	SetScriptArgs([]string{"arg1", "arg2"})
	result := callEnvFunc("mixArgs")
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("mixArgs() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("mixArgs() with scriptArgs should return 2 elements, got %d", len(arr.Elements))
	}
}

func callLogFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("log")
	if mod == nil {
		panic("log module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestLogModule(t *testing.T) {
	// Test setLevel and getLevel
	callLogFunc("setLevel", String("debug"))
	result := callLogFunc("getLevel")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("getLevel() should return String, got %T", result)
	}
	if s.Value != "DEBUG" {
		t.Errorf("getLevel() = %s, want 'DEBUG'", s.Value)
	}

	// Test debug logging
	result = callLogFunc("debug", String("test debug message"))
	s, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("debug() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("debug() should return non-empty message")
	}

	// Test info logging
	result = callLogFunc("info", String("test info message"))
	s, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("info() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("info() should return non-empty message")
	}

	// Test warn logging
	result = callLogFunc("warn", String("test warn message"))
	s, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("warn() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("warn() should return non-empty message")
	}

	// Test error logging
	result = callLogFunc("error", String("test error message"))
	s, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("error() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("error() should return non-empty message")
	}

	// Reset to info level
	callLogFunc("setLevel", String("info"))
}

func TestLogFormat(t *testing.T) {
	result := callLogFunc("format", String("info"), String("formatted"), String("message"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("format() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("format() should return non-empty message")
	}
}

func TestLogIsLevel(t *testing.T) {
	callLogFunc("setLevel", String("warn"))

	// debug should be disabled
	result := callLogFunc("isLevel", String("debug"))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isLevel() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isLevel('debug') should be false when level is warn")
	}

	// error should be enabled
	result = callLogFunc("isLevel", String("error"))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isLevel() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isLevel('error') should be true when level is warn")
	}

	// Reset to info
	callLogFunc("setLevel", String("info"))
}

func TestLogPrint(t *testing.T) {
	result := callLogFunc("print", String("test print"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("print() should return String, got %T", result)
	}
	if s.Value != "test print" {
		t.Errorf("print() = %s, want 'test print'", s.Value)
	}
}

func TestLogPrintNoNL(t *testing.T) {
	result := callLogFunc("printNoNL", String("test"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("printNoNL() should return String, got %T", result)
	}
	if s.Value != "test" {
		t.Errorf("printNoNL() = %s, want 'test'", s.Value)
	}
}

func TestLogPrintf(t *testing.T) {
	result := callLogFunc("printf", String("value: %d"), Int(42))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("printf() should return String, got %T", result)
	}
	if s.Value != "value: 42" {
		t.Errorf("printf() = %s, want 'value: 42'", s.Value)
	}
}

func TestLogWithPrefix(t *testing.T) {
	result := callLogFunc("withPrefix", String("PREFIX"), String("test message"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("withPrefix() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("withPrefix() should return non-empty message")
	}
}

func TestLogJson(t *testing.T) {
	result := callLogFunc("json", String("info"), String("test json log"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("json() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("json() should return non-empty message")
	}
}

func TestLogStack(t *testing.T) {
	result := callLogFunc("stack")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("stack() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("stack() should return non-empty stack trace")
	}
}
