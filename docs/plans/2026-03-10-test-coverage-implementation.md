# Comprehensive Test Suite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Achieve 80%+ test coverage across all core xxlang packages with layer-by-layer unit tests.

**Architecture:** Each package gets focused unit tests organized by happy path, edge cases, and errors. New test files for interpreter and plugin packages. Follow TDD: write failing test first, then implement.

**Tech Stack:** Go testing package, table-driven tests, t.Run() for subtests

---

## Phase 1: Critical Gaps (High Priority)

### Task 1: Create pkg/interpreter Tests

**Files:**
- Create: `pkg/interpreter/interpreter_test.go`

**Step 1: Write the failing test for New() and Eval()**

```go
// pkg/interpreter/interpreter_test.go
package interpreter

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestInterpreter_New(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		interp := New()
		if interp == nil {
			t.Fatal("expected interpreter, got nil")
		}
	})

	t.Run("with stdlib", func(t *testing.T) {
		interp := New(WithStdlib())
		result, err := interp.Eval(`len("hello")`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
	})
}

func TestInterpreter_Eval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"integer", "42", int64(42)},
		{"float", "3.14", 3.14},
		{"string", `"hello"`, "hello"},
		{"boolean", "true", true},
		{"arithmetic", "2 + 3 * 4", int64(14)},
		{"comparison", "1 < 2", true},
		{"array", "[1, 2, 3][0]", int64(1)},
		{"map", `{"a": 1}["a"]`, int64(1)},
	}

	interp := New(WithStdlib())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := interp.Eval(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			verifyResult(t, result, tt.expected)
		})
	}
}

func TestInterpreter_SetGetGlobal(t *testing.T) {
	interp := New(WithStdlib())

	t.Run("set and get integer", func(t *testing.T) {
		interp.SetGlobal("x", 42)
		val, ok := interp.GetGlobal("x")
		if !ok {
			t.Fatal("expected to find global x")
		}
		if intVal, ok := val.(*objects.Int); !ok {
			t.Fatalf("expected Int, got %T", val)
		} else if intVal.Value != 42 {
			t.Fatalf("expected 42, got %d", intVal.Value)
		}
	})

	t.Run("use in code", func(t *testing.T) {
		interp.SetGlobal("y", 10)
		result, err := interp.Eval("y * 2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		verifyResult(t, result, int64(20))
	})
}

func TestInterpreter_Errors(t *testing.T) {
	interp := New(WithStdlib())

	t.Run("compile error", func(t *testing.T) {
		_, err := interp.Eval("var x = ")
		if err == nil {
			t.Fatal("expected error for incomplete expression")
		}
	})

	t.Run("runtime error", func(t *testing.T) {
		_, err := interp.Eval("1 / 0")
		if err == nil {
			t.Fatal("expected error for division by zero")
		}
	})
}

func verifyResult(t *testing.T, obj objects.Object, expected interface{}) {
	t.Helper()
	switch exp := expected.(type) {
	case int64:
		if intVal, ok := obj.(*objects.Int); !ok {
			t.Fatalf("expected Int, got %T", obj)
		} else if intVal.Value != exp {
			t.Fatalf("expected %d, got %d", exp, intVal.Value)
		}
	case float64:
		if floatVal, ok := obj.(*objects.Float); !ok {
			t.Fatalf("expected Float, got %T", obj)
		} else if floatVal.Value != exp {
			t.Fatalf("expected %f, got %f", exp, floatVal.Value)
		}
	case string:
		if strVal, ok := obj.(*objects.String); !ok {
			t.Fatalf("expected String, got %T", obj)
		} else if strVal.Value != exp {
			t.Fatalf("expected %q, got %q", exp, strVal.Value)
		}
	case bool:
		if boolVal, ok := obj.(*objects.Bool); !ok {
			t.Fatalf("expected Bool, got %T", obj)
		} else if boolVal.Value != exp {
			t.Fatalf("expected %t, got %t", exp, boolVal.Value)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/interpreter/... -v`
Expected: PASS (interpreter package already exists)

**Step 3: Verify coverage**

Run: `go test ./pkg/interpreter/... -cover`
Expected: 80%+ coverage

**Step 4: Commit**

```bash
git add pkg/interpreter/interpreter_test.go
git commit -m "test(interpreter): add comprehensive interpreter tests"
```

---

### Task 2: Expand pkg/objects Tests - Type Methods

**Files:**
- Modify: `pkg/objects/string_test.go`
- Modify: `pkg/objects/array_test.go`
- Modify: `pkg/objects/map_test.go`

**Step 1: Write failing tests for string methods**

```go
// Add to pkg/objects/string_test.go

func TestString_Methods(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		method   string
		args     []Object
		expected Object
	}{
		{"len", "hello", "len", nil, &Int{Value: 5}},
		{"upper", "hello", "upper", nil, &String{Value: "HELLO"}},
		{"lower", "HELLO", "lower", nil, &String{Value: "hello"}},
		{"trim", "  hi  ", "trim", nil, &String{Value: "hi"}},
		{"contains true", "hello", "contains", []Object{&String{Value: "ell"}}, TRUE},
		{"contains false", "hello", "contains", []Object{&String{Value: "xyz"}}, FALSE},
		{"indexOf found", "hello", "indexOf", []Object{&String{Value: "l"}}, &Int{Value: 2}},
		{"indexOf not found", "hello", "indexOf", []Object{&String{Value: "z"}}, &Int{Value: -1}},
		{"startsWith true", "hello", "startsWith", []Object{&String{Value: "he"}}, TRUE},
		{"startsWith false", "hello", "startsWith", []Object{&String{Value: "lo"}}, FALSE},
		{"endsWith true", "hello", "endsWith", []Object{&String{Value: "lo"}}, TRUE},
		{"endsWith false", "hello", "endsWith", []Object{&String{Value: "he"}}, FALSE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &String{Value: tt.input}
			result, err := s.InvokeMethod(tt.method, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			compareObjects(t, result, tt.expected)
		})
	}
}
```

**Step 2: Write failing tests for array methods**

```go
// Add to pkg/objects/array_test.go

func TestArray_Methods(t *testing.T) {
	tests := []struct {
		name     string
		input    *Array
		method   string
		args     []Object
		expected Object
	}{
		{"len", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "len", nil, &Int{Value: 2}},
		{"first", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "first", nil, &Int{Value: 1}},
		{"last", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "last", nil, &Int{Value: 2}},
		{"indexOf found", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "indexOf", []Object{&Int{Value: 2}}, &Int{Value: 1}},
		{"indexOf not found", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "indexOf", []Object{&Int{Value: 3}}, &Int{Value: -1}},
		{"contains true", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "contains", []Object{&Int{Value: 1}}, TRUE},
		{"contains false", &Array{Elements: []Object{&Int{Value: 1}, &Int{Value: 2}}}, "contains", []Object{&Int{Value: 3}}, FALSE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.InvokeMethod(tt.method, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			compareObjects(t, result, tt.expected)
		})
	}
}
```

**Step 3: Write failing tests for map methods**

```go
// Add to pkg/objects/map_test.go

func TestMap_Methods(t *testing.T) {
	pairs := map[string]Object{"a": &Int{Value: 1}, "b": &Int{Value: 2}}
	m := &Map{Pairs: pairs}

	tests := []struct {
		name     string
		method   string
		args     []Object
		expected Object
	}{
		{"len", "len", nil, &Int{Value: 2}},
		{"hasKey true", "hasKey", []Object{&String{Value: "a"}}, TRUE},
		{"hasKey false", "hasKey", []Object{&String{Value: "c"}}, FALSE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.InvokeMethod(tt.method, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			compareObjects(t, result, tt.expected)
		})
	}
}
```

**Step 4: Run tests to verify coverage**

Run: `go test ./pkg/objects/... -v -cover`
Expected: Coverage increases from 8% to 50%+

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "test(objects): add type method tests for string, array, map"
```

---

### Task 3: Expand pkg/compiler Tests

**Files:**
- Modify: `pkg/compiler/compiler_test.go`

**Step 1: Add opcode verification tests**

```go
// Add to pkg/compiler/compiler_test.go

func TestCompiler_OpcodeVerification(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedOps   []byte
		unexpectedOps []byte
	}{
		{
			name:          "constant load",
			input:         "42",
			expectedOps:   []byte{byte(OpConstant)},
			unexpectedOps: nil,
		},
		{
			name:          "addition",
			input:         "1 + 2",
			expectedOps:   []byte{byte(OpAdd)},
			unexpectedOps: nil,
		},
		{
			name:          "comparison",
			input:         "1 < 2",
			expectedOps:   []byte{byte(OpLessThan)},
			unexpectedOps: nil,
		},
		{
			name:          "function call",
			input:         "func f() { return 1 }; f()",
			expectedOps:   []byte{byte(OpCall)},
			unexpectedOps: nil,
		},
		{
			name:          "array creation",
			input:         "[1, 2, 3]",
			expectedOps:   []byte{byte(OpArray)},
			unexpectedOps: nil,
		},
		{
			name:          "map creation",
			input:         `{"a": 1}`,
			expectedOps:   []byte{byte(OpMap)},
			unexpectedOps: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(t, tt.input)
			c := New()
			err := c.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %v", err)
			}

			bytecode := c.Bytecode()
			for _, expectedOp := range tt.expectedOps {
				if !containsOpcode(bytecode.Instructions, expectedOp) {
					t.Errorf("expected opcode %d in bytecode", expectedOp)
				}
			}
			for _, unexpectedOp := range tt.unexpectedOps {
				if containsOpcode(bytecode.Instructions, unexpectedOp) {
					t.Errorf("unexpected opcode %d in bytecode", unexpectedOp)
				}
			}
		})
	}
}

func TestCompiler_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError string
	}{
		{"undefined variable", "x", "undefined"},
		{"redeclare variable", "var x = 1; var x = 2", "redeclare"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(t, tt.input)
			c := New()
			err := c.Compile(program)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func containsOpcode(instructions []byte, opcode byte) bool {
	for i := 0; i < len(instructions); i++ {
		if instructions[i] == opcode {
			return true
		}
	}
	return false
}
```

**Step 2: Run tests**

Run: `go test ./pkg/compiler/... -v -cover`
Expected: Coverage increases from 48% to 65%+

**Step 3: Commit**

```bash
git add pkg/compiler/compiler_test.go
git commit -m "test(compiler): add opcode verification and error tests"
```

---

### Task 4: Expand pkg/vm Tests - Error Handling

**Files:**
- Modify: `pkg/vm/vm_test.go`

**Step 1: Add error condition tests**

```go
// Add to pkg/vm/vm_test.go

func TestVM_ErrorConditions(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError string
	}{
		{
			name:        "division by zero",
			input:       "1 / 0",
			expectError: "division by zero",
		},
		{
			name:        "array index out of bounds - negative",
			input:       "var a = [1, 2]; a[-1]",
			expectError: "index out of bounds",
		},
		{
			name:        "array index out of bounds - too high",
			input:       "var a = [1, 2]; a[10]",
			expectError: "index out of bounds",
		},
		{
			name:        "wrong number of arguments",
			input:       "func f(x) { return x }; f()",
			expectError: "wrong number of arguments",
		},
		{
			name:        "call non-function",
			input:       "var x = 1; x()",
			expectError: "not a function",
		},
		{
			name:        "index non-indexable",
			input:       "var x = 1; x[0]",
			expectError: "not indexable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(t, tt.input)
			c := compiler.New()
			err := c.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %v", err)
			}

			v := New(c.Bytecode())
			err = v.Run()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func TestVM_StackOperations(t *testing.T) {
	t.Run("deeply nested expressions", func(t *testing.T) {
		input := "1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10"
		result := testEval(t, input)
		testIntegerObject(t, result, 55)
	})

	t.Run("nested function calls", func(t *testing.T) {
		input := `
			func a() { return 1 }
			func b() { return a() + 1 }
			func c() { return b() + 1 }
			func d() { return c() + 1 }
			d()
		`
		result := testEval(t, input)
		testIntegerObject(t, result, 4)
	})
}
```

**Step 2: Run tests**

Run: `go test ./pkg/vm/... -v -cover`
Expected: Coverage increases from 55% to 70%+

**Step 3: Commit**

```bash
git add pkg/vm/vm_test.go
git commit -m "test(vm): add error condition and stack operation tests"
```

---

## Phase 2: Medium Priority

### Task 5: Create pkg/plugin Tests

**Files:**
- Create: `pkg/plugin/plugin_test.go`

**Step 1: Write plugin tests**

```go
// pkg/plugin/plugin_test.go
package plugin

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestPlugin_Interface(t *testing.T) {
	t.Run("plugin struct", func(t *testing.T) {
		p := &Plugin{
			Name: "test",
			Exports: map[string]objects.Object{
				"hello": &objects.String{Value: "world"},
			},
		}

		if p.Name != "test" {
			t.Errorf("expected name 'test', got %q", p.Name)
		}
		if len(p.Exports) != 1 {
			t.Errorf("expected 1 export, got %d", len(p.Exports))
		}
	})
}

func TestPlugin_Registry(t *testing.T) {
	t.Run("register and get", func(t *testing.T) {
		// Clear registry for test
		registryMu.Lock()
		registry = make(map[string]*Plugin)
		registryMu.Unlock()

		p := &Plugin{Name: "testplugin"}
		Register(p)

		got := Get("testplugin")
		if got == nil {
			t.Fatal("expected plugin, got nil")
		}
		if got.Name != "testplugin" {
			t.Errorf("expected name 'testplugin', got %q", got.Name)
		}
	})

	t.Run("get non-existent", func(t *testing.T) {
		got := Get("nonexistent")
		if got != nil {
			t.Errorf("expected nil for non-existent plugin, got %v", got)
		}
	})
}

func TestPlugin_List(t *testing.T) {
	registryMu.Lock()
	registry = make(map[string]*Plugin)
	registry["a"] = &Plugin{Name: "a"}
	registry["b"] = &Plugin{Name: "b"}
	registryMu.Unlock()

	list := List()
	if len(list) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(list))
	}
}
```

**Step 2: Run tests**

Run: `go test ./pkg/plugin/... -v -cover`
Expected: Tests pass, coverage 70%+

**Step 3: Commit**

```bash
git add pkg/plugin/plugin_test.go
git commit -m "test(plugin): add plugin registry and interface tests"
```

---

### Task 6: Expand pkg/parser Error Tests

**Files:**
- Modify: `pkg/parser/parser_test.go`

**Step 1: Add error message tests**

```go
// Add to pkg/parser/parser_test.go

func TestParser_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError string
	}{
		{
			name:        "missing expression",
			input:       "var x = ",
			expectError: "expected expression",
		},
		{
			name:        "unclosed parenthesis",
			input:       "(1 + 2",
			expectError: "expected ')'",
		},
		{
			name:        "unclosed bracket",
			input:       "[1, 2",
			expectError: "expected ']'",
		},
		{
			name:        "unclosed brace",
			input:       "{x: 1",
			expectError: "expected '}'",
		},
		{
			name:        "invalid assignment target",
			input:       "1 = 2",
			expectError: "expected identifier",
		},
		{
			name:        "missing function name",
			input:       "func () { }",
			expectError: "expected identifier",
		},
		{
			name:        "missing class name",
			input:       "class { }",
			expectError: "expected identifier",
		},
		{
			name:        "unclosed string",
			input:       `"hello`,
			expectError: "unterminated string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Fatal("expected parser error, got none")
			}

			found := false
			for _, err := range errors {
				if strings.Contains(err, tt.expectError) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got %v", tt.expectError, errors)
			}
		})
	}
}

func TestParser_ErrorRecovery(t *testing.T) {
	t.Run("multiple errors", func(t *testing.T) {
		input := `
			var x =
			func f( { }
			class { }
		`
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		errors := p.Errors()
		if len(errors) < 2 {
			t.Errorf("expected multiple errors, got %d: %v", len(errors), errors)
		}
	})
}
```

**Step 2: Run tests**

Run: `go test ./pkg/parser/... -v -cover`
Expected: Coverage increases from 72% to 80%+

**Step 3: Commit**

```bash
git add pkg/parser/parser_test.go
git commit -m "test(parser): add error message and recovery tests"
```

---

### Task 7: Expand pkg/stdlib Tests

**Files:**
- Modify: `pkg/stdlib/io_test.go`
- Modify: `pkg/stdlib/string_test.go`
- Modify: `pkg/stdlib/array_test.go`

**Step 1: Add io tests**

```go
// Add to pkg/stdlib/io_test.go

func TestIO_Printf(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	interp := interpreter.New(WithStdlib())
	interp.Eval(`printf("Hello %s, you are %d years old", "Alice", 30)`)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Hello Alice") {
		t.Errorf("expected output to contain 'Hello Alice', got %q", output)
	}
}
```

**Step 2: Add string edge case tests**

```go
// Add to pkg/stdlib/string_test.go

func TestString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"empty string len", `len("")`, 0},
		{"unicode len", `len("你好")`, 2},
		{"unicode upper", `upper("hello")`, "HELLO"},
		{"trim all whitespace", `trim("   ")`, ""},
		{"substr empty", `substr("", 0, 0)`, ""},
		{"indexOf empty", `indexOf("hello", "")`, 0},
		{"split empty", `len(split("", ","))`, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.input)
			verifyResult(t, result, tt.expected)
		})
	}
}
```

**Step 3: Run tests**

Run: `go test ./pkg/stdlib/... -v -cover`
Expected: Coverage increases from 58% to 70%+

**Step 4: Commit**

```bash
git add pkg/stdlib/
git commit -m "test(stdlib): add io printf and string edge case tests"
```

---

## Phase 3: Polish

### Task 8: Add pkg/lexer Edge Cases

**Files:**
- Modify: `pkg/lexer/lexer_test.go`

**Step 1: Add edge case tests**

```go
// Add to pkg/lexer/lexer_test.go

func TestLexer_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokenExpect
	}{
		{
			name:  "unicode string",
			input: `"你好世界"`,
			expected: []tokenExpect{
				{token.STRING, "你好世界"},
			},
		},
		{
			name:  "escape sequences",
			input: `"hello\nworld\t"`,
			expected: []tokenExpect{
				{token.STRING, "hello\nworld\t"},
			},
		},
		{
			name:  "large integer",
			input: "9223372036854775807", // max int64
			expected: []tokenExpect{
				{token.INT, "9223372036854775807"},
			},
		},
		{
			name:  "float with exponent",
			input: "1.5e10",
			expected: []tokenExpect{
				{token.FLOAT, "1.5e10"},
			},
		},
		{
			name:  "negative float",
			input: "-3.14",
			expected: []tokenExpect{
				{token.MINUS, "-"},
				{token.FLOAT, "3.14"},
			},
		},
		{
			name:  "consecutive operators",
			input: "==!=",
			expected: []tokenExpect{
				{token.EQ, "=="},
				{token.NOT_EQ, "!="},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for _, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.typ {
					t.Errorf("expected token type %v, got %v", exp.typ, tok.Type)
				}
				if tok.Literal != exp.literal {
					t.Errorf("expected literal %q, got %q", exp.literal, tok.Literal)
				}
			}
		})
	}
}

func TestLexer_Errors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"unterminated string", `"hello`, true},
		{"invalid escape", `"hello\g"`, false}, // Should be handled at runtime
		{"invalid char", "@", false},            // Should produce ILLEGAL token
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			hasError := false
			for {
				tok := l.NextToken()
				if tok.Type == token.ILLEGAL {
					hasError = true
				}
				if tok.Type == token.EOF {
					break
				}
			}
			if tt.expectError && !hasError {
				t.Error("expected error, got none")
			}
		})
	}
}
```

**Step 2: Run tests**

Run: `go test ./pkg/lexer/... -v -cover`
Expected: Coverage increases from 84% to 90%+

**Step 3: Commit**

```bash
git add pkg/lexer/lexer_test.go
git commit -m "test(lexer): add unicode, escape, and edge case tests"
```

---

### Task 9: Final Verification and Coverage Report

**Step 1: Run all tests**

Run: `go test ./... -v 2>&1 | tail -20`
Expected: All tests pass

**Step 2: Generate coverage report**

Run: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
Expected: All packages 70%+, core packages 80%+

**Step 3: Commit coverage report**

```bash
git add coverage.out
git commit -m "test: add coverage report baseline"
```

---

## Summary

| Phase | Tasks | Target Coverage |
|-------|-------|-----------------|
| 1: Critical | Tasks 1-4 | interpreter 80%+, objects 50%+, compiler 65%+, vm 70%+ |
| 2: Medium | Tasks 5-7 | plugin 70%+, parser 80%+, stdlib 70%+ |
| 3: Polish | Tasks 8-9 | lexer 90%+, final verification |

**Total estimated time:** 2-3 hours
