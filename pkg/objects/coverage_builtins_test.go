// pkg/objects/coverage_builtins_test.go
package objects

import (
	"testing"
)

// TestGetMethodExtra tests method lookup edge cases
func TestGetMethodExtra(t *testing.T) {
	_, ok := GetMethod(ObjectType("UNKNOWN"), "abs")
	if ok {
		t.Error("should not find method for unknown type")
	}
}

// TestArrayMethodErrors tests array method error cases
func TestArrayMethodErrors(t *testing.T) {
	t.Run("len wrong args", func(t *testing.T) {
		method, _ := GetMethod(ArrayType, "len")
		result := method.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("len with no args should return error")
		}
	})

	t.Run("push wrong args", func(t *testing.T) {
		method, _ := GetMethod(ArrayType, "push")
		arr := &Array{Elements: []Object{}}
		result := method.Fn(arr)
		if _, ok := result.(*Error); !ok {
			t.Error("push with no element should return error")
		}
	})

	t.Run("indexOf wrong args", func(t *testing.T) {
		method, _ := GetMethod(ArrayType, "indexOf")
		arr := &Array{Elements: []Object{NewInt(1)}}
		result := method.Fn(arr)
		if _, ok := result.(*Error); !ok {
			t.Error("indexOf with no element should return error")
		}
	})

	t.Run("contains wrong args", func(t *testing.T) {
		method, _ := GetMethod(ArrayType, "contains")
		arr := &Array{Elements: []Object{NewInt(1)}}
		result := method.Fn(arr)
		if _, ok := result.(*Error); !ok {
			t.Error("contains with no element should return error")
		}
	})

	t.Run("join wrong args", func(t *testing.T) {
		method, _ := GetMethod(ArrayType, "join")
		arr := &Array{Elements: []Object{NewString("a")}}
		result := method.Fn(arr)
		if _, ok := result.(*Error); !ok {
			t.Error("join with no separator should return error")
		}
	})

	t.Run("join wrong separator type", func(t *testing.T) {
		method, _ := GetMethod(ArrayType, "join")
		arr := &Array{Elements: []Object{NewString("a")}}
		result := method.Fn(arr, NewInt(1))
		if _, ok := result.(*Error); !ok {
			t.Error("join with non-string separator should return error")
		}
	})
}

// TestIntMethodErrors tests int method error cases
func TestIntMethodErrors(t *testing.T) {
	t.Run("abs wrong args", func(t *testing.T) {
		method, _ := GetMethod(IntType, "abs")
		result := method.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("abs with no args should return error")
		}
	})

	t.Run("toFloat wrong args", func(t *testing.T) {
		method, _ := GetMethod(IntType, "toFloat")
		result := method.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("toFloat with no args should return error")
		}
	})
}

// TestFloatMethodErrors tests float method error cases
func TestFloatMethodErrors(t *testing.T) {
	t.Run("floor wrong args", func(t *testing.T) {
		method, _ := GetMethod(FloatType, "floor")
		result := method.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("floor with no args should return error")
		}
	})

	t.Run("ceil wrong args", func(t *testing.T) {
		method, _ := GetMethod(FloatType, "ceil")
		result := method.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("ceil with no args should return error")
		}
	})

	t.Run("round wrong args", func(t *testing.T) {
		method, _ := GetMethod(FloatType, "round")
		result := method.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("round with no args should return error")
		}
	})
}

// TestStringMethodErrors tests string method error cases
func TestStringMethodErrors(t *testing.T) {
	t.Run("split wrong args", func(t *testing.T) {
		method, _ := GetMethod(StringType, "split")
		result := method.Fn(NewString("a,b"))
		if _, ok := result.(*Error); !ok {
			t.Error("split with no separator should return error")
		}
	})

	t.Run("contains wrong args", func(t *testing.T) {
		method, _ := GetMethod(StringType, "contains")
		result := method.Fn(NewString("hello"))
		if _, ok := result.(*Error); !ok {
			t.Error("contains with no substr should return error")
		}
	})

	t.Run("indexOf wrong args", func(t *testing.T) {
		method, _ := GetMethod(StringType, "indexOf")
		result := method.Fn(NewString("hello"))
		if _, ok := result.(*Error); !ok {
			t.Error("indexOf with no substr should return error")
		}
	})

	t.Run("startsWith wrong args", func(t *testing.T) {
		method, _ := GetMethod(StringType, "startsWith")
		result := method.Fn(NewString("hello"))
		if _, ok := result.(*Error); !ok {
			t.Error("startsWith with no prefix should return error")
		}
	})

	t.Run("endsWith wrong args", func(t *testing.T) {
		method, _ := GetMethod(StringType, "endsWith")
		result := method.Fn(NewString("hello"))
		if _, ok := result.(*Error); !ok {
			t.Error("endsWith with no suffix should return error")
		}
	})

	t.Run("subStr wrong args", func(t *testing.T) {
		method, _ := GetMethod(StringType, "subStr")
		result := method.Fn(NewString("hello"))
		if _, ok := result.(*Error); !ok {
			t.Error("subStr with no start should return error")
		}
	})

	t.Run("subStr too many args", func(t *testing.T) {
		method, _ := GetMethod(StringType, "subStr")
		result := method.Fn(NewString("hello"), NewInt(0), NewInt(3), NewInt(4))
		if _, ok := result.(*Error); !ok {
			t.Error("subStr with too many args should return error")
		}
	})
}

// TestMapMethodErrors tests map method error cases
func TestMapMethodErrors(t *testing.T) {
	m := &Map{Pairs: make(map[HashKey]MapPair)}

	t.Run("len wrong args", func(t *testing.T) {
		method, _ := GetMethod(MapType, "len")
		result := method.Fn()
		if _, ok := result.(*Error); !ok {
			t.Error("len with no args should return error")
		}
	})

	t.Run("hasKey wrong args", func(t *testing.T) {
		method, _ := GetMethod(MapType, "hasKey")
		result := method.Fn(m)
		if _, ok := result.(*Error); !ok {
			t.Error("hasKey with no key should return error")
		}
	})

	t.Run("delete wrong args", func(t *testing.T) {
		method, _ := GetMethod(MapType, "delete")
		result := method.Fn(m)
		if _, ok := result.(*Error); !ok {
			t.Error("delete with no key should return error")
		}
	})
}

// TestCharsMethodsExtra tests chars type methods
func TestCharsMethodsExtra(t *testing.T) {
	tests := []struct {
		name   string
		val    string
		method string
		args   []Object
		check  func(Object) bool
	}{
		{"len", "hello", "len", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 5
		}},
		{"upper", "hello", "upper", nil, func(r Object) bool {
			c, ok := r.(*Chars)
			return ok && string(c.Value) == "HELLO"
		}},
		{"lower", "HELLO", "lower", nil, func(r Object) bool {
			c, ok := r.(*Chars)
			return ok && string(c.Value) == "hello"
		}},
		{"trim", "  hi  ", "trim", nil, func(r Object) bool {
			c, ok := r.(*Chars)
			return ok && string(c.Value) == "hi"
		}},
		{"at", "hello", "at", []Object{NewInt(1)}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "e"
		}},
		{"at negative", "hello", "at", []Object{NewInt(-1)}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "o"
		}},
		{"reverse", "hello", "reverse", nil, func(r Object) bool {
			c, ok := r.(*Chars)
			return ok && string(c.Value) == "olleh"
		}},
		{"repeat", "ab", "repeat", []Object{NewInt(3)}, func(r Object) bool {
			c, ok := r.(*Chars)
			return ok && string(c.Value) == "ababab"
		}},
		{"repeat zero", "ab", "repeat", []Object{NewInt(0)}, func(r Object) bool {
			return r == CHARS_EMPTY
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(CharsType, tt.method)
			if !ok {
				t.Fatalf("method %s not found", tt.method)
			}
			args := append([]Object{NewCharsFromString(tt.val)}, tt.args...)
			result := method.Fn(args...)
			if err, ok := result.(*Error); ok {
				t.Fatalf("unexpected error: %s", err.Message)
			}
			if !tt.check(result) {
				t.Errorf("check failed for method %s, got %s", tt.method, result.Inspect())
			}
		})
	}
}

// TestCharsMethodErrors tests chars method error cases
func TestCharsMethodErrors(t *testing.T) {
	c := NewCharsFromString("hello")

	t.Run("at wrong args", func(t *testing.T) {
		method, _ := GetMethod(CharsType, "at")
		result := method.Fn(c)
		if _, ok := result.(*Error); !ok {
			t.Error("at with no index should return error")
		}
	})

	t.Run("at out of bounds", func(t *testing.T) {
		method, _ := GetMethod(CharsType, "at")
		result := method.Fn(c, NewInt(100))
		if _, ok := result.(*Error); !ok {
			t.Error("at out of bounds should return error")
		}
	})

	t.Run("repeat negative", func(t *testing.T) {
		method, _ := GetMethod(CharsType, "repeat")
		result := method.Fn(c, NewInt(-1))
		if _, ok := result.(*Error); !ok {
			t.Error("repeat with negative count should return error")
		}
	})
}

// TestNewCharsFromStringExtra tests chars creation
func TestNewCharsFromStringExtra(t *testing.T) {
	c := NewCharsFromString("hello")
	if c == nil {
		t.Fatal("NewCharsFromString returned nil")
	}
	if len(c.Value) != 5 {
		t.Errorf("expected 5 chars, got %d", len(c.Value))
	}
	if c.Type() != CharsType {
		t.Errorf("expected CharsType, got %s", c.Type())
	}
}

// TestCharsInspectExtra tests chars inspect
func TestCharsInspectExtra(t *testing.T) {
	c := NewCharsFromString("hello")
	if c.Inspect() != "hello" {
		t.Errorf("expected 'hello', got %s", c.Inspect())
	}
}

// TestNewArrayExtra tests array creation
func TestNewArrayExtra(t *testing.T) {
	elements := []Object{NewInt(1), NewInt(2)}
	arr := NewArray(elements)
	if arr == nil {
		t.Fatal("NewArray returned nil")
	}
	if len(arr.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr.Elements))
	}
}

// TestReleaseMapSliceExtra tests releasing map slice
func TestReleaseMapSliceExtra(t *testing.T) {
	m1 := NewMapWithCapacity(2)
	m2 := NewMapWithCapacity(0) // fresh empty map
	objs := []*Map{m1, m2}
	ReleaseMapSlice(objs)
}

// TestNewErrorExtra tests error creation
func TestNewErrorExtra(t *testing.T) {
	err := newError("test error: %s", "value")
	if err.Message != "test error: value" {
		t.Errorf("expected 'test error: value', got %s", err.Message)
	}
}
