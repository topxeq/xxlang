//go:build !js

// pkg/objects/builtin_system_test.go
package objects

import (
	"runtime"
	"strings"
	"testing"
)

func TestBuiltinSystemCmd(t *testing.T) {
	fn, ok := Builtins["systemCmd"]
	if !ok {
		t.Fatal("systemCmd builtin not found")
	}

	result := fn.Fn(NewString("echo hello"))
	mapResult, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("expected OrderedMap, got %T", result)
	}

	successPair := mapResult.Get(NewString("success"))
	boolResult, ok := successPair.(*Bool)
	if !ok {
		t.Fatalf("expected Bool for success, got %T", successPair)
	}
	if boolResult.Value != true {
		t.Errorf("expected success to be true, got %v", boolResult.Value)
	}

	outputPair := mapResult.Get(NewString("output"))
	_, ok = outputPair.(*String)
	if !ok {
		t.Fatalf("expected String for output, got %T", outputPair)
	}

	exitCodePair := mapResult.Get(NewString("exitCode"))
	intResult, ok := exitCodePair.(*Int)
	if !ok {
		t.Fatalf("expected Int for exitCode, got %T", exitCodePair)
	}
	if intResult.Value != 0 {
		t.Errorf("expected exitCode 0, got %d", intResult.Value)
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string first argument")
	}

	result = fn.Fn(NewString("echo hello"), NewInt(456))
	if !isError(result) {
		t.Error("expected error for non-string subsequent arguments")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSystemCmdDetached(t *testing.T) {
	fn, ok := Builtins["systemCmdDetached"]
	if !ok {
		t.Fatal("systemCmdDetached builtin not found")
	}

	result := fn.Fn(NewString("sleep 1"))
	mapResult, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("expected OrderedMap, got %T", result)
	}

	successPair := mapResult.Get(NewString("success"))
	boolResult, ok := successPair.(*Bool)
	if !ok {
		t.Fatalf("expected Bool for success, got %T", successPair)
	}
	if boolResult.Value != true {
		t.Errorf("expected success to be true, got %v", boolResult.Value)
	}

	pidPair := mapResult.Get(NewString("pid"))
	_, ok = pidPair.(*Int)
	if !ok {
		t.Fatalf("expected Int for pid, got %T", pidPair)
	}

	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string first argument")
	}

	result = fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}
}

func TestBuiltinSystemStart(t *testing.T) {
	fn, ok := Builtins["systemStart"]
	if !ok {
		t.Fatal("systemStart builtin not found")
	}

	// Test with no args
	result := fn.Fn()
	if !isError(result) {
		t.Error("expected error for no args")
	}

	// Test with non-string argument
	result = fn.Fn(NewInt(123))
	if !isError(result) {
		t.Error("expected error for non-string argument")
	}

	// Test with valid path (might succeed or fail depending on OS)
	// We just check that it returns a map with success and error fields
	result = fn.Fn(NewString("echo test"))
	mapResult, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("expected OrderedMap, got %T", result)
	}

	successPair := mapResult.Get(NewString("success"))
	_, ok = successPair.(*Bool)
	if !ok {
		t.Fatalf("expected Bool for success, got %T", successPair)
	}
	// success can be true or false depending on whether the command succeeded

	errorPair := mapResult.Get(NewString("error"))
	_, ok = errorPair.(*String)
	if !ok {
		t.Fatalf("expected String for error, got %T", errorPair)
	}

	// Test with working directory (if it's a valid directory)
	// This is platform-dependent, so we just check that it accepts the argument
	result = fn.Fn(NewString("echo test"), NewString("."))
	mapResult2, ok2 := result.(*OrderedMap)
	if !ok2 {
		t.Fatalf("expected OrderedMap for workingDir test, got %T", result)
	}
	_ = mapResult2 // Just check it doesn't panic
}

func TestBuiltinSystemCmd_WithArgs(t *testing.T) {
	fn, _ := Builtins["systemCmd"]

	// Test with command and arguments
	result := fn.Fn(NewString("echo"), NewString("hello"), NewString("world"))
	mapResult, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("expected OrderedMap, got %T", result)
	}

	successPair := mapResult.Get(NewString("success"))
	boolResult, ok := successPair.(*Bool)
	if !ok {
		t.Fatalf("expected Bool for success, got %T", successPair)
	}
	if !boolResult.Value {
		t.Errorf("expected success to be true for echo command, got %v", boolResult.Value)
	}

	outputPair := mapResult.Get(NewString("output"))
	outputStr, ok := outputPair.(*String)
	if !ok {
		t.Fatalf("expected String for output, got %T", outputPair)
	}
	// Output should contain "hello world" (might have newline)
	if !strings.Contains(outputStr.Value, "hello") && !strings.Contains(outputStr.Value, "Hello") {
		t.Errorf("expected output to contain 'hello', got %q", outputStr.Value)
	}
}

func TestBuiltinSystemCmd_FailingCommand(t *testing.T) {
	fn, _ := Builtins["systemCmd"]

	// Test with a command that fails (non-zero exit code)
	// On Unix: "false" command exits with 1
	// On Windows: "cmd /c exit 1" also exits with 1
	cmd := "false"
	if runtime.GOOS == "windows" {
		cmd = "cmd /c \"exit 1\""
	}

	result := fn.Fn(NewString(cmd))
	mapResult, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("expected OrderedMap, got %T", result)
	}

	successPair := mapResult.Get(NewString("success"))
	boolResult, ok := successPair.(*Bool)
	if !ok {
		t.Fatalf("expected Bool for success, got %T", successPair)
	}
	if boolResult.Value {
		t.Errorf("expected success to be false for failing command, got %v", boolResult.Value)
	}

	exitCodePair := mapResult.Get(NewString("exitCode"))
	intResult, ok := exitCodePair.(*Int)
	if !ok {
		t.Fatalf("expected Int for exitCode, got %T", exitCodePair)
	}
	if intResult.Value == 0 {
		t.Errorf("expected non-zero exitCode for failing command, got %d", intResult.Value)
	}
}

func TestBuiltinSystemCmdDetached_WithArgs(t *testing.T) {
	fn, _ := Builtins["systemCmdDetached"]

	// Test with command and arguments
	result := fn.Fn(NewString("echo"), NewString("hello"))
	mapResult, ok := result.(*OrderedMap)
	if !ok {
		t.Fatalf("expected OrderedMap, got %T", result)
	}

	successPair := mapResult.Get(NewString("success"))
	boolResult, ok := successPair.(*Bool)
	if !ok {
		t.Fatalf("expected Bool for success, got %T", successPair)
	}
	if !boolResult.Value {
		t.Errorf("expected success to be true for echo command, got %v", boolResult.Value)
	}

	pidPair := mapResult.Get(NewString("pid"))
	_, ok = pidPair.(*Int)
	if !ok {
		t.Fatalf("expected Int for pid, got %T", pidPair)
	}

	errorPair := mapResult.Get(NewString("error"))
	_, ok = errorPair.(*String)
	if !ok {
		t.Fatalf("expected String for error, got %T", errorPair)
	}
}
