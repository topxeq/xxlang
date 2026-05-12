// pkg/stdlib/env_parseflags_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// envCall invokes a builtin from the env module.
func envCall(name string, args ...objects.Object) objects.Object {
	mod := Get("env")
	if mod == nil {
		return &objects.Error{Message: "env module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// helper to create a flag spec map
func flagSpec(name, short, ftype, def, desc string) *objects.Map {
	pairs := make(map[objects.HashKey]objects.MapPair)
	pairs[objects.NewString("name").HashKey()] = objects.MapPair{Key: objects.NewString("name"), Value: objects.NewString(name)}
	if short != "" {
		pairs[objects.NewString("short").HashKey()] = objects.MapPair{Key: objects.NewString("short"), Value: objects.NewString(short)}
	}
	if ftype != "" {
		pairs[objects.NewString("type").HashKey()] = objects.MapPair{Key: objects.NewString("type"), Value: objects.NewString(ftype)}
	}
	if def != "" {
		pairs[objects.NewString("default").HashKey()] = objects.MapPair{Key: objects.NewString("default"), Value: objects.NewString(def)}
	}
	if desc != "" {
		pairs[objects.NewString("desc").HashKey()] = objects.MapPair{Key: objects.NewString("desc"), Value: objects.NewString(desc)}
	}
	return &objects.Map{Pairs: pairs}
}

// flagSpecWithDefault creates a flag spec with a non-string default value
func flagSpecWithDefault(name, short, ftype string, def objects.Object, desc string) *objects.Map {
	pairs := make(map[objects.HashKey]objects.MapPair)
	pairs[objects.NewString("name").HashKey()] = objects.MapPair{Key: objects.NewString("name"), Value: objects.NewString(name)}
	if short != "" {
		pairs[objects.NewString("short").HashKey()] = objects.MapPair{Key: objects.NewString("short"), Value: objects.NewString(short)}
	}
	if ftype != "" {
		pairs[objects.NewString("type").HashKey()] = objects.MapPair{Key: objects.NewString("type"), Value: objects.NewString(ftype)}
	}
	if def != nil {
		pairs[objects.NewString("default").HashKey()] = objects.MapPair{Key: objects.NewString("default"), Value: def}
	}
	if desc != "" {
		pairs[objects.NewString("desc").HashKey()] = objects.MapPair{Key: objects.NewString("desc"), Value: objects.NewString(desc)}
	}
	return &objects.Map{Pairs: pairs}
}

// strArray creates an Xxlang string array from Go strings
func strArray(ss ...string) *objects.Array {
	elems := make([]objects.Object, len(ss))
	for i, s := range ss {
		elems[i] = objects.NewString(s)
	}
	return &objects.Array{Elements: elems}
}

// pfGetStr gets a string value from a Map by key name
func pfGetStr(m *objects.Map, key string) (string, bool) {
	k := objects.NewString(key)
	if pair, found := m.Pairs[k.HashKey()]; found {
		if s, ok := pair.Value.(*objects.String); ok {
			return s.Value, true
		}
	}
	return "", false
}

// pfGetInt gets an int value from a Map by key name
func pfGetInt(m *objects.Map, key string) (int64, bool) {
	k := objects.NewString(key)
	if pair, found := m.Pairs[k.HashKey()]; found {
		if n, ok := pair.Value.(*objects.Int); ok {
			return n.Value, true
		}
	}
	return 0, false
}

// pfGetBool gets a bool value from a Map by key name
func pfGetBool(m *objects.Map, key string) (bool, bool) {
	k := objects.NewString(key)
	if pair, found := m.Pairs[k.HashKey()]; found {
		if b, ok := pair.Value.(*objects.Bool); ok {
			return b.Value, true
		}
	}
	return false, false
}

// pfGetArr gets an array value from a Map by key name
func pfGetArr(m *objects.Map, key string) (*objects.Array, bool) {
	k := objects.NewString(key)
	if pair, found := m.Pairs[k.HashKey()]; found {
		if a, ok := pair.Value.(*objects.Array); ok {
			return a, true
		}
	}
	return nil, false
}

func TestParseFlags_BasicLongFlags(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpec("url", "u", "", "http://localhost", "Target URL"),
		flagSpec("user", "U", "", "admin", "Username"),
		flagSpec("pass", "p", "", "", "Password"),
	}}

	args := strArray("--url", "http://example.com", "--user", "bob")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map, got %s", result.Type())
	}

	// Check url
	if v, found := pfGetStr(m, "url"); !found || v != "http://example.com" {
		t.Errorf("url: expected 'http://example.com', got %v", v)
	}
	// Check user
	if v, found := pfGetStr(m, "user"); !found || v != "bob" {
		t.Errorf("user: expected 'bob', got %v", v)
	}
	// Check pass (default)
	if v, found := pfGetStr(m, "pass"); !found || v != "" {
		t.Errorf("pass: expected '', got %v", v)
	}
}

func TestParseFlags_EqualsSyntax(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpec("url", "u", "", "", ""),
		flagSpec("timeout", "t", "int", "", ""),
	}}

	args := strArray("--url=http://test.com", "--timeout=5000")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map, got %s", result.Type())
	}

	if v, found := pfGetStr(m, "url"); !found || v != "http://test.com" {
		t.Errorf("url: expected 'http://test.com', got %v", v)
	}
	if v, found := pfGetInt(m, "timeout"); !found || v != 5000 {
		t.Errorf("timeout: expected 5000, got %v", v)
	}
}

func TestParseFlags_BoolFlags(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpecWithDefault("verbose", "v", "bool", nil, "Verbose"),
		flagSpecWithDefault("debug", "d", "bool", nil, "Debug"),
		flagSpecWithDefault("help", "h", "bool", nil, "Help"),
	}}

	args := strArray("--verbose", "-d")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map")
	}

	if v, found := pfGetBool(m, "verbose"); !found || !v {
		t.Errorf("verbose: expected true, got %v", v)
	}
	if v, found := pfGetBool(m, "debug"); !found || !v {
		t.Errorf("debug: expected true, got %v", v)
	}
	if v, found := pfGetBool(m, "help"); !found || v {
		t.Errorf("help: expected false, got %v", v)
	}
}

func TestParseFlags_ShortFlags(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpec("url", "u", "", "", ""),
		flagSpecWithDefault("verbose", "v", "bool", nil, ""),
		flagSpec("count", "c", "int", "", ""),
	}}

	args := strArray("-u", "http://test.com", "-v", "-c", "42")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map")
	}

	if v, found := pfGetStr(m, "url"); !found || v != "http://test.com" {
		t.Errorf("url: expected 'http://test.com', got %v", v)
	}
	if v, found := pfGetBool(m, "verbose"); !found || !v {
		t.Errorf("verbose: expected true, got %v", v)
	}
	if v, found := pfGetInt(m, "count"); !found || v != 42 {
		t.Errorf("count: expected 42, got %v", v)
	}
}

func TestParseFlags_PositionalArgs(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpec("output", "o", "", "", ""),
	}}

	args := strArray("file1.txt", "--output", "out.txt", "file2.txt")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map")
	}

	// Check _args
	arr, found := pfGetArr(m, "_args")
	if !found {
		t.Fatal("_args not found")
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("_args: expected 2 elements, got %d", len(arr.Elements))
	}
	if arr.Elements[0].(*objects.String).Value != "file1.txt" {
		t.Errorf("_args[0]: expected 'file1.txt', got %v", arr.Elements[0])
	}
	if arr.Elements[1].(*objects.String).Value != "file2.txt" {
		t.Errorf("_args[1]: expected 'file2.txt', got %v", arr.Elements[1])
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpecWithDefault("url", "u", "", objects.NewString("http://localhost"), ""),
		flagSpecWithDefault("timeout", "t", "int", objects.NewInt(10000), ""),
		flagSpecWithDefault("verbose", "v", "bool", nil, ""),
	}}

	// No args provided — should use defaults
	args := strArray()
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map")
	}

	if v, found := pfGetStr(m, "url"); !found || v != "http://localhost" {
		t.Errorf("url default: expected 'http://localhost', got %v", v)
	}
	if v, found := pfGetInt(m, "timeout"); !found || v != 10000 {
		t.Errorf("timeout default: expected 10000, got %v", v)
	}
	if v, found := pfGetBool(m, "verbose"); !found || v {
		t.Errorf("verbose default: expected false, got %v", v)
	}
}

func TestParseFlags_DoubleDashSeparator(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpec("name", "n", "", "", ""),
	}}

	args := strArray("--name", "test", "--", "--not-a-flag", "positional")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map")
	}

	if v, found := pfGetStr(m, "name"); !found || v != "test" {
		t.Errorf("name: expected 'test', got %v", v)
	}
	arr, found := pfGetArr(m, "_args")
	if !found {
		t.Fatal("_args not found")
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("_args: expected 2 elements, got %d", len(arr.Elements))
	}
	if arr.Elements[0].(*objects.String).Value != "--not-a-flag" {
		t.Errorf("_args[0]: expected '--not-a-flag', got %v", arr.Elements[0])
	}
}

func TestParseFlags_HelpText(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpec("url", "u", "", "http://localhost", "Target URL"),
		flagSpecWithDefault("verbose", "v", "bool", nil, "Verbose output"),
	}}

	args := strArray("--help")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map")
	}

	help, found := pfGetStr(m, "_help")
	if !found {
		t.Fatal("_help not found")
	}
	if help == "" {
		t.Error("_help should not be empty when help is requested")
	}
	// Help should contain option descriptions
	if !containsStr(help, "--url") || !containsStr(help, "--verbose") {
		t.Errorf("_help should contain flag descriptions, got: %s", help)
	}
}

func TestParseFlags_FloatType(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpec("rate", "r", "float", "", ""),
	}}

	args := strArray("--rate=3.14")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("parseFlags should return Map")
	}

	k := objects.NewString("rate")
	if pair, found := m.Pairs[k.HashKey()]; found {
		if f, ok := pair.Value.(*objects.Float); !ok || f.Value != 3.14 {
			t.Errorf("rate: expected 3.14, got %v", pair.Value)
		}
	} else {
		t.Error("rate not found in result")
	}
}

func TestParseFlags_NoArgs(t *testing.T) {
	result := envCall("parseFlags")
	if result.Type() != objects.ErrorType {
		t.Error("parseFlags() with no args should return error")
	}
}

func TestParseFlags_BoolWithEquals(t *testing.T) {
	specs := &objects.Array{Elements: []objects.Object{
		flagSpecWithDefault("verbose", "v", "bool", nil, ""),
	}}

	// --verbose=true and --verbose=false
	args := strArray("--verbose=true")
	result := envCall("parseFlags", specs, args)
	if result.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result.Inspect())
	}
	m := result.(*objects.Map)
	if v, found := pfGetBool(m, "verbose"); !found || !v {
		t.Errorf("verbose=true: expected true, got %v", v)
	}

	args2 := strArray("--verbose=false")
	result2 := envCall("parseFlags", specs, args2)
	if result2.Type() == objects.ErrorType {
		t.Fatalf("parseFlags returned error: %s", result2.Inspect())
	}
	m2 := result2.(*objects.Map)
	if v, found := pfGetBool(m2, "verbose"); !found || v {
		t.Errorf("verbose=false: expected false, got %v", v)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
