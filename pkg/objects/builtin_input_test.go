// pkg/objects/builtin_input_test.go
package objects

import (
	"testing"
)

func TestBuiltinGetInput(t *testing.T) {
	fn, ok := Builtins["getInput"]
	if !ok {
		t.Fatal("getInput builtin not found")
	}

	// Test error case - too many args
	result := fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetInputf(t *testing.T) {
	fn, ok := Builtins["getInputf"]
	if !ok {
		t.Fatal("getInputf builtin not found")
	}

	// Test error case - no args
	result := fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	// Test error case - non-string format
	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string format")
	}
}

func TestBuiltinGetChar(t *testing.T) {
	fn, ok := Builtins["getChar"]
	if !ok {
		t.Fatal("getChar builtin not found")
	}

	// Test error case - extra args
	result := fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetMultiLineInput(t *testing.T) {
	fn, ok := Builtins["getMultiLineInput"]
	if !ok {
		t.Fatal("getMultiLineInput builtin not found")
	}

	// Test too many arguments
	result := fn.Fn(NewString("arg1"), NewString("arg2"))
	if !isError(result) {
		t.Error("expected error for too many args")
	}

	// Test invalid endMarker type
	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string endMarker")
	}
}

func TestBuiltinGetPassword(t *testing.T) {
	fn, ok := Builtins["getPassword"]
	if !ok {
		t.Fatal("getPassword builtin not found")
	}

	// Test too many arguments
	result := fn.Fn(NewString("prompt"), NewString("extra"))
	if !isError(result) {
		t.Error("expected error for too many args")
	}

	// Test non-string prompt
	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string prompt")
	}
}

func TestBuiltinConfirm(t *testing.T) {
	fn, ok := Builtins["confirm"]
	if !ok {
		t.Fatal("confirm builtin not found")
	}

	// Test no args
	result := fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	// Test too many args
	result = fn.Fn(NewString("prompt"), NewInt(1), NewString("extra"))
	if !isError(result) {
		t.Error("expected error for too many args")
	}

	// Test non-string first arg
	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string first arg")
	}

	// Test non-bool second arg (should not error, just ignore)
	// The function should accept any type as second arg and treat as default false
	_ = fn.Fn(NewString("prompt"), NewInt(456)) // should not error
}

func TestBuiltinReadLine(t *testing.T) {
	fn, ok := Builtins["readLine"]
	if !ok {
		t.Fatal("readLine builtin not found")
	}

	// Test error case - extra args
	result := fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinGetClipText(t *testing.T) {
	fn, ok := Builtins["getClipText"]
	if !ok {
		t.Fatal("getClipText builtin not found")
	}

	// Test error case - extra args
	result := fn.Fn(NewString("extra"))
	if !isError(result) {
		t.Error("expected error for extra args")
	}
}

func TestBuiltinSetClipText(t *testing.T) {
	fn, ok := Builtins["setClipText"]
	if !ok {
		t.Fatal("setClipText builtin not found")
	}

	// Test error case - no args
	result := fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	// Test error case - non-string arg
	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string arg")
	}
}
