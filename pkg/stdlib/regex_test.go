// pkg/stdlib/regex_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callRegexFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("std/regex")
	if mod == nil {
		panic("std/regex module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestRegexCompile(t *testing.T) {
	result := callRegexFunc("compile", String(`\d+`))
	cr, ok := result.(*CompiledRegex)
	if !ok {
		t.Fatalf("compile() should return CompiledRegex, got %T", result)
	}
	if cr.Pattern != `\d+` {
		t.Errorf("compile() pattern = %s, want %s", cr.Pattern, `\d+`)
	}

	// Test invalid pattern
	result = callRegexFunc("compile", String(`[invalid`))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("compile() with invalid pattern should return Error, got %T", result)
	}
}

func TestRegexMatch(t *testing.T) {
	// Test with string pattern
	result := callRegexFunc("match", String(`\d+`), String("hello123world"))
	r, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("match() should return Bool, got %T", result)
	}
	if !r.Value {
		t.Errorf("match('\\d+', 'hello123world') should be true")
	}

	result = callRegexFunc("match", String(`^\d+$`), String("hello"))
	r, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("match() should return Bool, got %T", result)
	}
	if r.Value {
		t.Errorf("match('^\\d+$', 'hello') should be false")
	}

	// Test with compiled regex
	compiled := callRegexFunc("compile", String(`[a-z]+`))
	result = callRegexFunc("match", compiled, String("abc"))
	r, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("match() should return Bool, got %T", result)
	}
	if !r.Value {
		t.Errorf("match(compiled, 'abc') should be true")
	}
}

func TestRegexFind(t *testing.T) {
	// Find first match
	result := callRegexFunc("find", String(`\d+`), String("a1b22c333"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("find() should return String, got %T", result)
	}
	if r.Value != "1" {
		t.Errorf("find('\\d+', 'a1b22c333') = %s, want '1'", r.Value)
	}

	// Find with no match
	result = callRegexFunc("find", String(`\d+`), String("no digits"))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("find() should return String, got %T", result)
	}
	if r.Value != "" {
		t.Errorf("find() with no match should return empty string")
	}
}

func TestRegexFindAll(t *testing.T) {
	result := callRegexFunc("findAll", String(`\d+`), String("a1b22c333"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("findAll() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("findAll('\\d+', 'a1b22c333') length = %d, want 3", len(arr.Elements))
	}

	expected := []string{"1", "22", "333"}
	for i, exp := range expected {
		if s, ok := arr.Elements[i].(*objects.String); !ok || s.Value != exp {
			t.Errorf("findAll()[%d] = %v, want %s", i, arr.Elements[i], exp)
		}
	}

	// Test with limit
	result = callRegexFunc("findAll", String(`\d+`), String("a1b22c333"), Int(2))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("findAll() should return Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("findAll() with limit=2 length = %d, want 2", len(arr.Elements))
	}
}

func TestRegexFindGroups(t *testing.T) {
	// Pattern with capture groups
	result := callRegexFunc("findGroups", String(`(\w+)@(\w+)\.(\w+)`), String("email: test@example.com"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("findGroups() should return Array, got %T", result)
	}
	if len(arr.Elements) != 4 {
		t.Errorf("findGroups() length = %d, want 4 (full match + 3 groups)", len(arr.Elements))
	}

	// Check full match
	if s, ok := arr.Elements[0].(*objects.String); !ok || s.Value != "test@example.com" {
		t.Errorf("findGroups()[0] = %v, want 'test@example.com'", arr.Elements[0])
	}

	// Check groups
	expected := []string{"test", "example", "com"}
	for i, exp := range expected {
		if s, ok := arr.Elements[i+1].(*objects.String); !ok || s.Value != exp {
			t.Errorf("findGroups()[%d] = %v, want '%s'", i+1, arr.Elements[i+1], exp)
		}
	}

	// Test with no match
	result = callRegexFunc("findGroups", String(`\d+`), String("no digits"))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("findGroups() with no match should return Null, got %T", result)
	}
}

func TestRegexReplace(t *testing.T) {
	result := callRegexFunc("replace", String(`\d+`), String("a1b22c333"), String("X"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("replace() should return String, got %T", result)
	}
	if r.Value != "aXbXcX" {
		t.Errorf("replace('\\d+', 'a1b22c333', 'X') = %s, want 'aXbXcX'", r.Value)
	}

	// Test with compiled regex
	compiled := callRegexFunc("compile", String(`[aeiou]`))
	result = callRegexFunc("replace", compiled, String("hello world"), String("*"))
	r, ok = result.(*objects.String)
	if !ok {
		t.Fatalf("replace() should return String, got %T", result)
	}
	if r.Value != "h*ll* w*rld" {
		t.Errorf("replace(vowels, 'hello world', '*') = %s, want 'h*ll* w*rld'", r.Value)
	}
}

func TestRegexSplit(t *testing.T) {
	result := callRegexFunc("split", String(`[,.]`), String("a,b,c.d,e"))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("split() should return Array, got %T", result)
	}
	if len(arr.Elements) != 5 {
		t.Errorf("split('[,.]', 'a,b,c.d,e') length = %d, want 5", len(arr.Elements))
	}

	expected := []string{"a", "b", "c", "d", "e"}
	for i, exp := range expected {
		if s, ok := arr.Elements[i].(*objects.String); !ok || s.Value != exp {
			t.Errorf("split()[%d] = %v, want '%s'", i, arr.Elements[i], exp)
		}
	}

	// Test with limit
	result = callRegexFunc("split", String(`,`), String("a,b,c,d,e"), Int(3))
	arr, ok = result.(*objects.Array)
	if !ok {
		t.Fatalf("split() should return Array, got %T", result)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("split() with limit=3 length = %d, want 3", len(arr.Elements))
	}
}

func TestRegexQuote(t *testing.T) {
	result := callRegexFunc("quote", String("a.b*c+d?"))
	r, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("quote() should return String, got %T", result)
	}
	expected := `a\.b\*c\+d\?`
	if r.Value != expected {
		t.Errorf("quote('a.b*c+d?') = %s, want '%s'", r.Value, expected)
	}
}

func TestRegexCount(t *testing.T) {
	result := callRegexFunc("count", String(`\d+`), String("a1b22c333"))
	r, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("count() should return Int, got %T", result)
	}
	if r.Value != 3 {
		t.Errorf("count('\\d+', 'a1b22c333') = %d, want 3", r.Value)
	}

	// Test with no matches
	result = callRegexFunc("count", String(`\d+`), String("no digits"))
	r, ok = result.(*objects.Int)
	if !ok {
		t.Fatalf("count() should return Int, got %T", result)
	}
	if r.Value != 0 {
		t.Errorf("count() with no matches should be 0, got %d", r.Value)
	}
}

func TestRegexErrorCases(t *testing.T) {
	// compile with wrong args
	result := callRegexFunc("compile")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("compile() with no args should return Error, got %T", result)
	}

	// match with wrong args
	result = callRegexFunc("match", String(`\d+`))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("match() with 1 arg should return Error, got %T", result)
	}

	// replace with wrong args
	result = callRegexFunc("replace", String(`\d+`), String("test"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("replace() with 2 args should return Error, got %T", result)
	}

	// find with wrong type
	result = callRegexFunc("find", Int(123), String("test"))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("find() with int pattern should return Error, got %T", result)
	}
}
