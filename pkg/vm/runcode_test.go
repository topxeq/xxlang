// pkg/vm/runcode_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestSetAndGetRunCodeCallback(t *testing.T) {
	// Set a callback
	testCallback := func(code string, args *objects.Map) (objects.Object, error) {
		return &objects.Int{Value: 42}, nil
	}

	SetRunCodeCallback(testCallback)

	// Get it back
	retrieved := GetRunCodeCallback()
	if retrieved == nil {
		t.Fatal("GetRunCodeCallback should return non-nil after SetRunCodeCallback")
	}

	// Call it to verify it's the same
	result, err := retrieved("test", nil)
	if err != nil {
		t.Fatalf("callback returned error: %v", err)
	}

	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}

	if i.Value != 42 {
		t.Errorf("callback: expected 42, got %d", i.Value)
	}

	// Reset to nil
	SetRunCodeCallback(nil)
	if GetRunCodeCallback() != nil {
		t.Error("GetRunCodeCallback should return nil after reset")
	}
}

func TestExecuteRunCodeWithoutCallback(t *testing.T) {
	// Make sure callback is nil
	SetRunCodeCallback(nil)

	// ExecuteRunCode without callback should return error
	result, err := ExecuteRunCode("test", nil)
	if err == nil {
		t.Error("ExecuteRunCode without callback should return error")
	}
	if result != nil {
		t.Error("ExecuteRunCode without callback should return nil result")
	}
}

func TestExecuteRunCodeWithCallback(t *testing.T) {
	// Set a callback
	testCallback := func(code string, args *objects.Map) (objects.Object, error) {
		return &objects.String{Value: code}, nil
	}
	SetRunCodeCallback(testCallback)
	defer SetRunCodeCallback(nil) // Clean up

	result, err := ExecuteRunCode("hello world", nil)
	if err != nil {
		t.Fatalf("ExecuteRunCode returned error: %v", err)
	}

	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}

	if s.Value != "hello world" {
		t.Errorf("ExecuteRunCode: expected 'hello world', got '%s'", s.Value)
	}
}
