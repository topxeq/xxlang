// pkg/stdlib/coverage_extra_test.go
// Additional tests to improve code coverage for the stdlib package
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Tests for csv module
// ============================================

func TestCSVModuleParse(t *testing.T) {
	mod := Get("csv")
	if mod == nil {
		t.Skip("csv module not found")
	}

	t.Run("parse basic", func(t *testing.T) {
		result := callStdlibFunc("csv", "parse", String("a,b,c\n1,2,3"))
		if result == nil {
			t.Fatal("parse returned nil")
		}
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 rows, got %d", len(arr.Elements))
		}
	})

	t.Run("parseWithHeader", func(t *testing.T) {
		result := callStdlibFunc("csv", "parseWithHeader", String("name,value\ntest,42"))
		if result == nil {
			t.Fatal("parseWithHeader returned nil")
		}
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 1 {
			t.Errorf("expected 1 row, got %d", len(arr.Elements))
		}
	})

	t.Run("stringify", func(t *testing.T) {
		arr := &objects.Array{
			Elements: []objects.Object{
				&objects.Array{Elements: []objects.Object{String("a"), String("b")}},
				&objects.Array{Elements: []objects.Object{String("1"), String("2")}},
			},
		}
		result := callStdlibFunc("csv", "stringify", arr)
		if result == nil {
			t.Fatal("stringify returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("stringify should return non-empty string")
		}
	})
}

// TestFindColumnIndex tests the findColumnIndex helper
func TestFindColumnIndex(t *testing.T) {
	header := &objects.Array{
		Elements: []objects.Object{
			String("name"),
			String("value"),
			String("count"),
		},
	}

	if idx := findColumnIndex(header, "value"); idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if idx := findColumnIndex(header, "notfound"); idx != -1 {
		t.Errorf("expected index -1 for not found, got %d", idx)
	}
}

// ============================================
// Tests for collections module
// ============================================

func TestCollectionsModule(t *testing.T) {
	mod := Get("collections")
	if mod == nil {
		t.Skip("collections module not found")
	}

	// Note: "range" is not a function in collections, it's a language keyword
	// Test other functions instead
	t.Run("repeat", func(t *testing.T) {
		result := callStdlibFunc("collections", "repeat", String("x"), Int(3))
		if result == nil {
			t.Fatal("repeat returned nil")
		}
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 3 {
			t.Errorf("expected 3 elements, got %d", len(arr.Elements))
		}
	})
}

// ============================================
// Tests for os module
// ============================================

func TestOSModule(t *testing.T) {
	mod := Get("os")
	if mod == nil {
		t.Skip("os module not found")
	}

	t.Run("arch", func(t *testing.T) {
		result := callStdlibFunc("os", "arch")
		if result == nil {
			t.Fatal("arch returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("os arch should not be empty")
		}
	})

	t.Run("getEnv", func(t *testing.T) {
		// Test getting an environment variable
		result := callStdlibFunc("os", "getEnv", String("PATH"))
		if result == nil {
			// May return nil if PATH doesn't exist, which is fine
			return
		}
		// If it exists, should be a string
		_, ok := result.(*objects.String)
		if !ok && result != objects.NULL {
			t.Logf("getEnv returned %T", result)
		}
	})
}

// ============================================
// Tests for filepath module
// ============================================

func TestFilepathModule(t *testing.T) {
	mod := Get("filepath")
	if mod == nil {
		t.Skip("filepath module not found")
	}

	t.Run("join", func(t *testing.T) {
		result := callStdlibFunc("filepath", "join", String("a"), String("b"), String("c"))
		if result == nil {
			t.Fatal("join returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("join should return non-empty string")
		}
	})

	t.Run("base", func(t *testing.T) {
		result := callStdlibFunc("filepath", "base", String("/path/to/file.txt"))
		if result == nil {
			t.Fatal("base returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "file.txt" {
			t.Errorf("expected 'file.txt', got '%s'", s.Value)
		}
	})

	t.Run("dir", func(t *testing.T) {
		result := callStdlibFunc("filepath", "dir", String("/path/to/file.txt"))
		if result == nil {
			t.Fatal("dir returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("dir should return non-empty string")
		}
	})

	t.Run("ext", func(t *testing.T) {
		result := callStdlibFunc("filepath", "ext", String("/path/to/file.txt"))
		if result == nil {
			t.Fatal("ext returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != ".txt" {
			t.Errorf("expected '.txt', got '%s'", s.Value)
		}
	})
}

// ============================================
// Tests for random module
// ============================================

func TestRandomModule(t *testing.T) {
	mod := Get("random")
	if mod == nil {
		t.Skip("random module not found")
	}

	t.Run("int", func(t *testing.T) {
		result := callStdlibFunc("random", "int", Int(1), Int(100))
		if result == nil {
			t.Fatal("int returned nil")
		}
		n, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if n.Value < 1 || n.Value > 100 {
			t.Errorf("random int %d out of range [1, 100]", n.Value)
		}
	})

	t.Run("float", func(t *testing.T) {
		result := callStdlibFunc("random", "float")
		if result == nil {
			t.Fatal("float returned nil")
		}
		f, ok := result.(*objects.Float)
		if !ok {
			t.Fatalf("expected Float, got %T", result)
		}
		if f.Value < 0 || f.Value >= 1 {
			t.Errorf("random float %f out of range [0, 1)", f.Value)
		}
	})

	t.Run("string", func(t *testing.T) {
		result := callStdlibFunc("random", "string", Int(10))
		if result == nil {
			t.Fatal("string returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) != 10 {
			t.Errorf("expected string length 10, got %d", len(s.Value))
		}
	})
}

// ============================================
// Tests for encoding module
// ============================================

func TestEncodingModule(t *testing.T) {
	mod := Get("encoding")
	if mod == nil {
		t.Skip("encoding module not found")
	}

	t.Run("base64Encode", func(t *testing.T) {
		result := callStdlibFunc("encoding", "base64Encode", String("hello"))
		if result == nil {
			t.Fatal("base64Encode returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("base64Encode should return non-empty string")
		}
	})

	t.Run("base64Decode", func(t *testing.T) {
		// "hello" in base64 is "aGVsbG8="
		result := callStdlibFunc("encoding", "base64Decode", String("aGVsbG8="))
		if result == nil {
			t.Fatal("base64Decode returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "hello" {
			t.Errorf("expected 'hello', got '%s'", s.Value)
		}
	})

	t.Run("hexEncode", func(t *testing.T) {
		result := callStdlibFunc("encoding", "hexEncode", String("hello"))
		if result == nil {
			t.Fatal("hexEncode returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("hexEncode should return non-empty string")
		}
	})

	t.Run("hexDecode", func(t *testing.T) {
		// "hello" in hex is "68656c6c6f"
		result := callStdlibFunc("encoding", "hexDecode", String("68656c6c6f"))
		if result == nil {
			t.Fatal("hexDecode returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "hello" {
			t.Errorf("expected 'hello', got '%s'", s.Value)
		}
	})
}

// ============================================
// Tests for strings module
// ============================================

func TestStringsModuleExtra(t *testing.T) {
	mod := Get("strings")
	if mod == nil {
		t.Skip("strings module not found")
	}

	t.Run("hasPrefix", func(t *testing.T) {
		result := callStdlibFunc("strings", "hasPrefix", String("hello world"), String("hello"))
		if result == nil {
			t.Fatal("hasPrefix returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("hasPrefix should return true")
		}
	})

	t.Run("hasSuffix", func(t *testing.T) {
		result := callStdlibFunc("strings", "hasSuffix", String("hello world"), String("world"))
		if result == nil {
			t.Fatal("hasSuffix returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("hasSuffix should return true")
		}
	})

	t.Run("contains", func(t *testing.T) {
		result := callStdlibFunc("strings", "contains", String("hello world"), String("wor"))
		if result == nil {
			t.Fatal("contains returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("contains should return true")
		}
	})

	t.Run("repeat", func(t *testing.T) {
		result := callStdlibFunc("strings", "repeat", String("ab"), Int(3))
		if result == nil {
			t.Fatal("repeat returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "ababab" {
			t.Errorf("expected 'ababab', got '%s'", s.Value)
		}
	})

	t.Run("title", func(t *testing.T) {
		result := callStdlibFunc("strings", "title", String("hello world"))
		if result == nil {
			t.Fatal("title returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("title should return non-empty string")
		}
	})

	t.Run("toLower", func(t *testing.T) {
		result := callStdlibFunc("strings", "toLower", String("HELLO"))
		if result == nil {
			t.Fatal("toLower returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "hello" {
			t.Errorf("expected 'hello', got '%s'", s.Value)
		}
	})

	t.Run("toUpper", func(t *testing.T) {
		result := callStdlibFunc("strings", "toUpper", String("hello"))
		if result == nil {
			t.Fatal("toUpper returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "HELLO" {
			t.Errorf("expected 'HELLO', got '%s'", s.Value)
		}
	})
}

// ============================================
// Tests for json module
// ============================================

func TestJSONModuleExtra(t *testing.T) {
	mod := Get("json")
	if mod == nil {
		t.Skip("json module not found")
	}

	t.Run("parse object", func(t *testing.T) {
		result := callStdlibFunc("json", "parse", String(`{"name":"test","value":42}`))
		if result == nil {
			t.Fatal("parse returned nil")
		}
		m, ok := result.(*objects.Map)
		if !ok {
			t.Fatalf("expected Map, got %T", result)
		}
		if len(m.Pairs) != 2 {
			t.Errorf("expected 2 pairs, got %d", len(m.Pairs))
		}
	})

	t.Run("parse array", func(t *testing.T) {
		result := callStdlibFunc("json", "parse", String(`[1,2,3]`))
		if result == nil {
			t.Fatal("parse returned nil")
		}
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 3 {
			t.Errorf("expected 3 elements, got %d", len(arr.Elements))
		}
	})

	t.Run("stringify", func(t *testing.T) {
		m := &objects.Map{
			Pairs: map[objects.HashKey]objects.MapPair{
				String("name").HashKey(): {Key: String("name"), Value: String("test")},
			},
		}
		result := callStdlibFunc("json", "stringify", m)
		if result == nil {
			t.Fatal("stringify returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if len(s.Value) == 0 {
			t.Error("stringify should return non-empty string")
		}
	})
}

// ============================================
// Tests for array module
// ============================================

func TestArrayModuleExtra(t *testing.T) {
	mod := Get("array")
	if mod == nil {
		t.Skip("array module not found")
	}

	arr := &objects.Array{
		Elements: []objects.Object{Int(3), Int(1), Int(4), Int(1), Int(5)},
	}

	t.Run("indexOf", func(t *testing.T) {
		result := callStdlibFunc("array", "indexOf", arr, Int(4))
		if result == nil {
			t.Fatal("indexOf returned nil")
		}
		n, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if n.Value != 2 {
			t.Errorf("expected index 2, got %d", n.Value)
		}
	})

	t.Run("contains true", func(t *testing.T) {
		result := callStdlibFunc("array", "contains", arr, Int(4))
		if result == nil {
			t.Fatal("contains returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("contains should return true")
		}
	})

	t.Run("contains false", func(t *testing.T) {
		result := callStdlibFunc("array", "contains", arr, Int(99))
		if result == nil {
			t.Fatal("contains returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value {
			t.Error("contains should return false")
		}
	})

	t.Run("first", func(t *testing.T) {
		result := callStdlibFunc("array", "first", arr)
		if result == nil {
			t.Fatal("first returned nil")
		}
		n, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if n.Value != 3 {
			t.Errorf("expected 3, got %d", n.Value)
		}
	})

	t.Run("last", func(t *testing.T) {
		result := callStdlibFunc("array", "last", arr)
		if result == nil {
			t.Fatal("last returned nil")
		}
		n, ok := result.(*objects.Int)
		if !ok {
			t.Fatalf("expected Int, got %T", result)
		}
		if n.Value != 5 {
			t.Errorf("expected 5, got %d", n.Value)
		}
	})
}

// ============================================
// Tests for map module
// ============================================

func TestMapModuleExtra(t *testing.T) {
	mod := Get("map")
	if mod == nil {
		t.Skip("map module not found")
	}

	pairs := map[objects.HashKey]objects.MapPair{
		String("a").HashKey(): {Key: String("a"), Value: Int(1)},
		String("b").HashKey(): {Key: String("b"), Value: Int(2)},
	}
	m := &objects.Map{Pairs: pairs}

	t.Run("keys", func(t *testing.T) {
		result := callStdlibFunc("map", "keys", m)
		if result == nil {
			t.Fatal("keys returned nil")
		}
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 keys, got %d", len(arr.Elements))
		}
	})

	t.Run("values", func(t *testing.T) {
		result := callStdlibFunc("map", "values", m)
		if result == nil {
			t.Fatal("values returned nil")
		}
		arr, ok := result.(*objects.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 values, got %d", len(arr.Elements))
		}
	})

	t.Run("hasKey true", func(t *testing.T) {
		result := callStdlibFunc("map", "hasKey", m, String("a"))
		if result == nil {
			t.Fatal("hasKey returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("hasKey should return true")
		}
	})

	t.Run("hasKey false", func(t *testing.T) {
		result := callStdlibFunc("map", "hasKey", m, String("z"))
		if result == nil {
			t.Fatal("hasKey returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if b.Value {
			t.Error("hasKey should return false")
		}
	})
}

// ============================================
// Tests for unicode module
// ============================================

func TestUnicodeModule(t *testing.T) {
	mod := Get("unicode")
	if mod == nil {
		t.Skip("unicode module not found")
	}

	t.Run("isLetter", func(t *testing.T) {
		result := callStdlibFunc("unicode", "isLetter", String("a"))
		if result == nil {
			t.Fatal("isLetter returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("'a' should be a letter")
		}
	})

	t.Run("isDigit", func(t *testing.T) {
		result := callStdlibFunc("unicode", "isDigit", String("5"))
		if result == nil {
			t.Fatal("isDigit returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("'5' should be a digit")
		}
	})

	t.Run("isSpace", func(t *testing.T) {
		result := callStdlibFunc("unicode", "isSpace", String(" "))
		if result == nil {
			t.Fatal("isSpace returned nil")
		}
		b, ok := result.(*objects.Bool)
		if !ok {
			t.Fatalf("expected Bool, got %T", result)
		}
		if !b.Value {
			t.Error("' ' should be a space")
		}
	})

	t.Run("toLower", func(t *testing.T) {
		result := callStdlibFunc("unicode", "toLower", String("A"))
		if result == nil {
			t.Fatal("toLower returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "a" {
			t.Errorf("expected 'a', got '%s'", s.Value)
		}
	})

	t.Run("toUpper", func(t *testing.T) {
		result := callStdlibFunc("unicode", "toUpper", String("a"))
		if result == nil {
			t.Fatal("toUpper returned nil")
		}
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("expected String, got %T", result)
		}
		if s.Value != "A" {
			t.Errorf("expected 'A', got '%s'", s.Value)
		}
	})
}
