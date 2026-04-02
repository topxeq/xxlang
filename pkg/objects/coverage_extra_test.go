// pkg/objects/coverage_extra_test.go
package objects

import (
	"strings"
	"testing"
)

// TestArrayMethods tests array method coverage
func TestArrayMethods(t *testing.T) {
	tests := []struct {
		name   string
		arr    *Array
		method string
		args   []Object
		check  func(Object) bool
	}{
		{
			name:   "push",
			arr:    &Array{Elements: []Object{NewInt(1)}},
			method: "push",
			args:   []Object{NewInt(2)},
			check: func(r Object) bool {
				arr, ok := r.(*Array)
				return ok && len(arr.Elements) == 2
			},
		},
		{
			name:   "pop",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2)}},
			method: "pop",
			args:   nil,
			check: func(r Object) bool {
				arr, ok := r.(*Array)
				return ok && len(arr.Elements) == 1 && arr.LastPopped != nil
			},
		},
		{
			name:   "first",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2)}},
			method: "first",
			args:   nil,
			check: func(r Object) bool {
				n, ok := r.(*Int)
				return ok && n.Value == 1
			},
		},
		{
			name:   "last",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2)}},
			method: "last",
			args:   nil,
			check: func(r Object) bool {
				n, ok := r.(*Int)
				return ok && n.Value == 2
			},
		},
		{
			name:   "indexOf found",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}},
			method: "indexOf",
			args:   []Object{NewInt(2)},
			check: func(r Object) bool {
				n, ok := r.(*Int)
				return ok && n.Value == 1
			},
		},
		{
			name:   "indexOf not found",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2)}},
			method: "indexOf",
			args:   []Object{NewInt(99)},
			check: func(r Object) bool {
				n, ok := r.(*Int)
				return ok && n.Value == -1
			},
		},
		{
			name:   "contains true",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2)}},
			method: "contains",
			args:   []Object{NewInt(1)},
			check: func(r Object) bool {
				return r == TRUE
			},
		},
		{
			name:   "contains false",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2)}},
			method: "contains",
			args:   []Object{NewInt(99)},
			check: func(r Object) bool {
				return r == FALSE
			},
		},
		{
			name:   "reverse",
			arr:    &Array{Elements: []Object{NewInt(1), NewInt(2), NewInt(3)}},
			method: "reverse",
			args:   nil,
			check: func(r Object) bool {
				arr, ok := r.(*Array)
				if !ok || len(arr.Elements) != 3 {
					return false
				}
				n0 := arr.Elements[0].(*Int)
				n2 := arr.Elements[2].(*Int)
				return n0.Value == 3 && n2.Value == 1
			},
		},
		{
			name:   "join",
			arr:    &Array{Elements: []Object{NewString("a"), NewString("b"), NewString("c")}},
			method: "join",
			args:   []Object{NewString(",")},
			check: func(r Object) bool {
				s, ok := r.(*String)
				return ok && s.Value == "a,b,c"
			},
		},
		{
			name:   "first empty",
			arr:    &Array{Elements: []Object{}},
			method: "first",
			args:   nil,
			check: func(r Object) bool {
				return r == NULL
			},
		},
		{
			name:   "last empty",
			arr:    &Array{Elements: []Object{}},
			method: "last",
			args:   nil,
			check: func(r Object) bool {
				return r == NULL
			},
		},
		{
			name:   "reverse empty",
			arr:    &Array{Elements: []Object{}},
			method: "reverse",
			args:   nil,
			check: func(r Object) bool {
				arr, ok := r.(*Array)
				return ok && len(arr.Elements) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(ArrayType, tt.method)
			if !ok {
				t.Fatalf("method %s not found", tt.method)
			}
			args := append([]Object{tt.arr}, tt.args...)
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

// TestArrayPopError tests pop on empty array
func TestArrayPopError(t *testing.T) {
	arr := &Array{Elements: []Object{}}
	method, _ := GetMethod(ArrayType, "pop")
	result := method.Fn(arr)
	if _, ok := result.(*Error); !ok {
		t.Error("pop on empty array should return error")
	}
}

// TestIntMethodsExtra tests additional int methods
func TestIntMethodsExtra(t *testing.T) {
	tests := []struct {
		name     string
		val      int64
		method   string
		args     []Object
		expected Object
	}{
		{"abs positive", 5, "abs", nil, NewInt(5)},
		{"abs negative", -5, "abs", nil, NewInt(5)},
		{"toFloat", 42, "toFloat", nil, NewFloat(42.0)},
		{"typeOf", 42, "typeOf", nil, NewString("INT")},
		{"toStr", 42, "toStr", nil, NewString("42")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(IntType, tt.method)
			if !ok {
				t.Fatalf("method %s not found", tt.method)
			}
			args := append([]Object{NewInt(tt.val)}, tt.args...)
			result := method.Fn(args...)
			if err, ok := result.(*Error); ok {
				t.Fatalf("unexpected error: %s", err.Message)
			}
			compareObjectsForTest(t, result, tt.expected)
		})
	}
}

// TestFloatMethodsExtra tests additional float methods
func TestFloatMethodsExtra(t *testing.T) {
	tests := []struct {
		name   string
		val    float64
		method string
		args   []Object
		check  func(Object) bool
	}{
		{"abs positive", 5.5, "abs", nil, func(r Object) bool {
			f, ok := r.(*Float)
			return ok && f.Value == 5.5
		}},
		{"abs negative", -5.5, "abs", nil, func(r Object) bool {
			f, ok := r.(*Float)
			return ok && f.Value == 5.5
		}},
		{"floor", 3.7, "floor", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 3
		}},
		{"ceil", 3.2, "ceil", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 4
		}},
		{"round", 3.5, "round", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 4
		}},
		{"toInt", 42.9, "toInt", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 42
		}},
		{"typeOf", 3.14, "typeOf", nil, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "FLOAT"
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(FloatType, tt.method)
			if !ok {
				t.Fatalf("method %s not found", tt.method)
			}
			args := append([]Object{NewFloat(tt.val)}, tt.args...)
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

// TestStringMethodsExtra tests additional string methods
func TestStringMethodsExtra(t *testing.T) {
	tests := []struct {
		name   string
		val    string
		method string
		args   []Object
		check  func(Object) bool
	}{
		{"trimLeft default", "  hello  ", "trimLeft", nil, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "hello  "
		}},
		{"trimRight default", "  hello  ", "trimRight", nil, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "  hello"
		}},
		{"trimLeft custom", "xxhelloxx", "trimLeft", []Object{NewString("x")}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "helloxx"
		}},
		{"trimRight custom", "xxhelloxx", "trimRight", []Object{NewString("x")}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "xxhello"
		}},
		{"subStr with end", "hello", "subStr", []Object{NewInt(1), NewInt(4)}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "ell"
		}},
		{"subStr without end", "hello", "subStr", []Object{NewInt(2)}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "llo"
		}},
		{"subStr negative start", "hello", "subStr", []Object{NewInt(-1)}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == "hello"
		}},
		{"subStr overflow start", "hello", "subStr", []Object{NewInt(100)}, func(r Object) bool {
			s, ok := r.(*String)
			return ok && s.Value == ""
		}},
		{"charLen ascii", "hello", "charLen", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 5
		}},
		{"charLen unicode", "你好", "charLen", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 2
		}},
		{"toInt valid", "42", "toInt", nil, func(r Object) bool {
			n, ok := r.(*Int)
			return ok && n.Value == 42
		}},
		{"toFloat valid", "3.14", "toFloat", nil, func(r Object) bool {
			f, ok := r.(*Float)
			return ok && f.Value == 3.14
		}},
		{"toChars", "hello", "toChars", nil, func(r Object) bool {
			c, ok := r.(*Chars)
			return ok && len(c.Value) == 5
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(StringType, tt.method)
			if !ok {
				t.Fatalf("method %s not found", tt.method)
			}
			args := append([]Object{NewString(tt.val)}, tt.args...)
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

// TestStringToIntError tests invalid int conversion
func TestStringToIntError(t *testing.T) {
	method, _ := GetMethod(StringType, "toInt")
	result := method.Fn(NewString("notanumber"))
	if _, ok := result.(*Error); !ok {
		t.Error("toInt of invalid string should return error")
	}
}

// TestStringToFloatError tests invalid float conversion
func TestStringToFloatError(t *testing.T) {
	method, _ := GetMethod(StringType, "toFloat")
	result := method.Fn(NewString("notanumber"))
	if _, ok := result.(*Error); !ok {
		t.Error("toFloat of invalid string should return error")
	}
}

// TestMapMethodsExtra tests additional map methods
func TestMapMethodsExtra(t *testing.T) {
	pairs := make(map[HashKey]MapPair)
	keyA := NewString("a")
	keyB := NewString("b")
	pairs[keyA.HashKey()] = MapPair{Key: keyA, Value: NewInt(1)}
	pairs[keyB.HashKey()] = MapPair{Key: keyB, Value: NewInt(2)}
	m := &Map{Pairs: pairs}

	t.Run("keys", func(t *testing.T) {
		method, _ := GetMethod(MapType, "keys")
		result := method.Fn(m)
		arr, ok := result.(*Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 keys, got %d", len(arr.Elements))
		}
	})

	t.Run("values", func(t *testing.T) {
		method, _ := GetMethod(MapType, "values")
		result := method.Fn(m)
		arr, ok := result.(*Array)
		if !ok {
			t.Fatalf("expected Array, got %T", result)
		}
		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 values, got %d", len(arr.Elements))
		}
	})

	t.Run("delete", func(t *testing.T) {
		method, _ := GetMethod(MapType, "delete")
		result := method.Fn(m, keyA)
		newMap, ok := result.(*Map)
		if !ok {
			t.Fatalf("expected Map, got %T", result)
		}
		if len(newMap.Pairs) != 1 {
			t.Errorf("expected 1 pair after delete, got %d", len(newMap.Pairs))
		}
	})
}

// TestEnvironment tests environment operations
func TestEnvironment(t *testing.T) {
	env := NewEnvironment()

	env.Set("x", NewInt(42))
	val, ok := env.Get("x")
	if !ok {
		t.Error("expected to find x in environment")
	}
	if val.(*Int).Value != 42 {
		t.Errorf("expected x=42, got %d", val.(*Int).Value)
	}

	_, ok = env.Get("y")
	if ok {
		t.Error("expected not to find y in environment")
	}
}

// TestEnclosedEnvironment tests nested environments
func TestEnclosedEnvironment(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", NewInt(10))

	inner := NewEnclosedEnvironment(outer)
	inner.Set("y", NewInt(20))

	val, ok := inner.Get("x")
	if !ok {
		t.Error("expected to find x in outer environment")
	}
	if val.(*Int).Value != 10 {
		t.Errorf("expected x=10, got %d", val.(*Int).Value)
	}

	val, ok = inner.Get("y")
	if !ok {
		t.Error("expected to find y in inner environment")
	}
	if val.(*Int).Value != 20 {
		t.Errorf("expected y=20, got %d", val.(*Int).Value)
	}
}

// TestFunction tests function object
func TestFunction(t *testing.T) {
	fn := &Function{
		Parameters: []*Identifier{{Value: "x"}, {Value: "y"}},
		Env:        NewEnvironment(),
		Name:       "testFunc",
	}

	if fn.Type() != FunctionType {
		t.Errorf("expected FunctionType, got %s", fn.Type())
	}
	if fn.TypeTag() != TagFunction {
		t.Errorf("expected TagFunction, got %v", fn.TypeTag())
	}
	if fn.ToBool() != TRUE {
		t.Error("function should be truthy")
	}
	if !strings.Contains(fn.Inspect(), "func") {
		t.Errorf("function inspect should contain 'func', got %s", fn.Inspect())
	}
}

// TestCompiledFunction tests compiled function object
func TestCompiledFunction(t *testing.T) {
	cf := &CompiledFunction{
		Instructions:  []byte{1, 2, 3},
		NumLocals:     5,
		NumParameters: 2,
		Name:          "test",
	}

	if cf.Type() != CompiledFunctionType {
		t.Errorf("expected CompiledFunctionType, got %s", cf.Type())
	}
	if cf.TypeTag() != TagCompiledFunction {
		t.Errorf("expected TagCompiledFunction, got %v", cf.TypeTag())
	}
	if cf.ToBool() != TRUE {
		t.Error("compiled function should be truthy")
	}
	if !strings.Contains(cf.Inspect(), "test") {
		t.Errorf("compiled function inspect should contain name, got %s", cf.Inspect())
	}
}

// TestNewMapWithCapacity tests map creation with capacity
func TestNewMapWithCapacity(t *testing.T) {
	m := NewMapWithCapacity(10)
	if m == nil {
		t.Fatal("NewMapWithCapacity returned nil")
	}
	if m.Pairs == nil {
		t.Error("expected Pairs to be initialized")
	}

	empty := NewMapWithCapacity(0)
	if len(empty.Pairs) != 0 {
		t.Error("NewMapWithCapacity(0) should return an empty map")
	}
}

// TestMapPoolStats tests map pool statistics
func TestMapPoolStats(t *testing.T) {
	ResetMapPoolStats()

	m := NewMapWithCapacity(5)
	m.Pairs[NewString("a").HashKey()] = MapPair{Key: NewString("a"), Value: NewInt(1)}

	ReleaseMap(m)

	stats := GetMapPoolStats()
	if stats.Released < 1 {
		t.Errorf("expected at least 1 release, got %d", stats.Released)
	}
}

// TestWarmMapPool tests map pool warming
func TestWarmMapPool(t *testing.T) {
	WarmMapPool(10)
}

// TestMapGetSortedKeys tests sorted keys caching
func TestMapGetSortedKeys(t *testing.T) {
	pairs := make(map[HashKey]MapPair)
	keys := []string{"c", "a", "b"}
	for _, k := range keys {
		key := NewString(k)
		pairs[key.HashKey()] = MapPair{Key: key, Value: NewInt(1)}
	}
	m := &Map{Pairs: pairs}

	sortedKeys := m.GetSortedKeys()
	if len(sortedKeys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(sortedKeys))
	}

	m.InvalidateKeysCache()
	sortedKeys2 := m.GetSortedKeys()
	if len(sortedKeys2) != 3 {
		t.Errorf("expected 3 keys after invalidation, got %d", len(sortedKeys2))
	}
}

// TestBuiltin tests builtin object
func TestBuiltin(t *testing.T) {
	fn := func(args ...Object) Object {
		return NewInt(int64(len(args)))
	}
	b := &Builtin{Fn: fn}

	if b.Type() != BuiltinType {
		t.Errorf("expected BuiltinType, got %s", b.Type())
	}
	if b.TypeTag() != TagBuiltin {
		t.Errorf("expected TagBuiltin, got %v", b.TypeTag())
	}
	if b.ToBool() != TRUE {
		t.Error("builtin should be truthy")
	}
}

// TestModule tests module object
func TestModuleObject(t *testing.T) {
	mod := &Module{
		Name:    "testModule",
		Exports: make(map[string]Object),
	}

	if mod.Type() != ModuleType {
		t.Errorf("expected ModuleType, got %s", mod.Type())
	}
	if mod.TypeTag() != TagModule {
		t.Errorf("expected TagModule, got %v", mod.TypeTag())
	}
	if mod.ToBool() != TRUE {
		t.Error("module should be truthy")
	}
	if !strings.Contains(mod.Inspect(), "testModule") {
		t.Errorf("module inspect should contain name, got %s", mod.Inspect())
	}
}

// TestClearInternCache tests intern cache clearing
func TestClearInternCache(t *testing.T) {
	InternString("unique_test_string")
	ClearInternCache()
	s := InternString("unique_test_string")
	if s == nil {
		t.Error("InternString should work after ClearInternCache")
	}
}

// TestUniversalMethods tests universal type methods
func TestUniversalMethods(t *testing.T) {
	t.Run("typeOf", func(t *testing.T) {
		result := universalTypeOf(NewInt(42))
		s, ok := result.(*String)
		if !ok || s.Value != "INT" {
			t.Errorf("expected 'INT', got %s", result.Inspect())
		}
	})

	t.Run("toStr", func(t *testing.T) {
		result := universalToStr(NewInt(42))
		s, ok := result.(*String)
		if !ok || s.Value != "42" {
			t.Errorf("expected '42', got %s", result.Inspect())
		}
	})

	t.Run("typeOf wrong args", func(t *testing.T) {
		result := universalTypeOf()
		if _, ok := result.(*Error); !ok {
			t.Error("typeOf with no args should return error")
		}
	})

	t.Run("toStr wrong args", func(t *testing.T) {
		result := universalToStr()
		if _, ok := result.(*Error); !ok {
			t.Error("toStr with no args should return error")
		}
	})
}
