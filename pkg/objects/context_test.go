// pkg/objects/context_test.go
package objects

import (
	"testing"
	"time"
)

// TestContextType tests the Context type methods
func TestContextType(t *testing.T) {
	ctx := NewBackgroundContext()

	if got := ctx.Type(); got != ContextType {
		t.Errorf("Context.Type() = %s, want CONTEXT", got)
	}

	if got := ctx.TypeTag(); got != TagContext {
		t.Errorf("Context.TypeTag() = %d, want %d", got, TagContext)
	}
}

func TestContextInspect(t *testing.T) {
	ctx := NewBackgroundContext()
	inspect := ctx.Inspect()

	if inspect == "" {
		t.Error("Context.Inspect() should not be empty")
	}

	// After cancellation
	ctx2 := NewContextWithCancel(nil)
	ctx2.Cancel()
	inspect2 := ctx2.Inspect()
	if inspect2 == "" {
		t.Error("Context.Inspect() should not be empty after cancel")
	}
}

func TestContextToBool(t *testing.T) {
	ctx := NewBackgroundContext()

	// Active context should be true
	if ctx.ToBool() != TRUE {
		t.Error("Active context should be TRUE")
	}

	// Cancelled context should be false
	ctx2 := NewContextWithCancel(nil)
	ctx2.Cancel()
	if ctx2.ToBool() != FALSE {
		t.Error("Cancelled context should be FALSE")
	}
}

func TestContextHashKey(t *testing.T) {
	ctx1 := NewBackgroundContext()
	ctx2 := NewBackgroundContext()

	// Each context should have a unique hash key
	if ctx1.HashKey() == ctx2.HashKey() {
		t.Error("Different contexts should have different hash keys")
	}

	// Same context should return same hash key
	hk := ctx1.HashKey()
	if ctx1.HashKey() != hk {
		t.Error("Same context should return same hash key")
	}
}

// TestContextCancel tests manual cancellation
func TestContextCancel(t *testing.T) {
	ctx := NewContextWithCancel(nil)

	if ctx.IsDone() {
		t.Error("New context should not be done")
	}

	ctx.Cancel()

	if !ctx.IsDone() {
		t.Error("Context should be done after cancel")
	}

	err := ctx.Err()
	if err == nil {
		t.Error("Cancelled context should have an error")
	}

	errStr := ctx.ErrString()
	if errStr == "" {
		t.Error("Error string should not be empty for cancelled context")
	}
}

// TestContextTimeout tests timeout functionality
func TestContextTimeout(t *testing.T) {
	ctx := NewContextWithTimeout(nil, 50*time.Millisecond)

	if ctx.IsDone() {
		t.Error("New timeout context should not be done immediately")
	}

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	if !ctx.IsDone() {
		t.Error("Context should be done after timeout")
	}

	errStr := ctx.ErrString()
	if errStr == "" {
		t.Error("Timed out context should have an error")
	}
}

// TestContextDeadline tests deadline functionality
func TestContextDeadline(t *testing.T) {
	// Context without deadline
	ctx1 := NewBackgroundContext()
	_, hasDeadline := ctx1.Deadline()
	if hasDeadline {
		t.Error("Background context should not have a deadline")
	}

	// Context with deadline
	deadline := time.Now().Add(200 * time.Millisecond)
	ctx2 := NewContextWithDeadline(nil, deadline)

	dl, hasDeadline := ctx2.Deadline()
	if !hasDeadline {
		t.Error("ContextWithDeadline should have a deadline")
	}

	if dl.IsZero() {
		t.Error("Deadline should not be zero")
	}

	// DeadlineString for context without deadline
	dlStr := ctx1.DeadlineString()
	if dlStr != "" {
		t.Error("Background context should have empty deadline string")
	}

	// DeadlineString for context with deadline
	dlStr2 := ctx2.DeadlineString()
	if dlStr2 == "" {
		t.Error("Context with deadline should have non-empty deadline string")
	}
}

// TestContextDone tests the Done tube
func TestContextDone(t *testing.T) {
	ctx := NewContextWithCancel(nil)

	done := ctx.Done()
	if done == nil {
		t.Fatal("Done() should return a tube")
	}

	// Done should return the same tube on subsequent calls
	done2 := ctx.Done()
	if done != done2 {
		t.Error("Done() should return the same tube")
	}

	// Cancel and verify tube is closed
	ctx.Cancel()

	// Give time for the goroutine to close the tube
	time.Sleep(50 * time.Millisecond)

	// Try to receive from done tube - should not block since it's closed
	// TryReceive returns (Object, received, open)
	_, received, open := done.TryReceive()
	if received {
		t.Error("Should not receive from closed tube")
	}
	if open {
		t.Error("Tube should be closed")
	}
}

// TestContextParent tests parent context inheritance
func TestContextParent(t *testing.T) {
	parent := NewContextWithCancel(nil)
	child := NewContextWithCancel(parent)

	// Cancel parent should cancel child
	parent.Cancel()

	// Give time for propagation
	time.Sleep(50 * time.Millisecond)

	if !child.IsDone() {
		t.Error("Child should be done when parent is cancelled")
	}
}

// TestContextParentTimeout tests parent timeout inheritance
func TestContextParentTimeout(t *testing.T) {
	parent := NewContextWithTimeout(nil, 50*time.Millisecond)
	child := NewContextWithCancel(parent)

	// Wait for parent timeout
	time.Sleep(100 * time.Millisecond)

	if !child.IsDone() {
		t.Error("Child should be done when parent times out")
	}
}

// ============================================================
// Builtin Function Tests
// ============================================================

func TestBuiltinNewContext(t *testing.T) {
	fn, ok := Builtins["newContext"]
	if !ok {
		t.Fatal("newContext builtin not found")
	}

	result := fn.Fn()
	ctx, ok := result.(*Context)
	if !ok {
		t.Fatalf("newContext should return *Context, got %T", result)
	}

	if ctx.IsDone() {
		t.Error("New context should not be done")
	}
}

func TestBuiltinContextWithTimeout(t *testing.T) {
	fn, ok := Builtins["contextWithTimeout"]
	if !ok {
		t.Fatal("contextWithTimeout builtin not found")
	}

	// Test with null parent
	result := fn.Fn(NULL, &Int{Value: 50})
	ctx, ok := result.(*Context)
	if !ok {
		t.Fatalf("contextWithTimeout should return *Context, got %T", result)
	}

	if ctx.IsDone() {
		t.Error("New timeout context should not be done immediately")
	}

	// Test error cases
	result = fn.Fn() // No args
	if !isError(result) {
		t.Error("Should return error for no arguments")
	}

	result = fn.Fn(NULL, &String{Value: "not a number"}) // Wrong type
	if !isError(result) {
		t.Error("Should return error for wrong type")
	}
}

func TestBuiltinContextWithCancel(t *testing.T) {
	fn, ok := Builtins["contextWithCancel"]
	if !ok {
		t.Fatal("contextWithCancel builtin not found")
	}

	result := fn.Fn()
	ctx, ok := result.(*Context)
	if !ok {
		t.Fatalf("contextWithCancel should return *Context, got %T", result)
	}

	if ctx.IsDone() {
		t.Error("New cancel context should not be done")
	}

	// Test with parent
	parent := NewBackgroundContext()
	result = fn.Fn(parent)
	ctx, ok = result.(*Context)
	if !ok {
		t.Fatalf("contextWithCancel should return *Context, got %T", result)
	}
}

func TestBuiltinContextWithDeadline(t *testing.T) {
	fn, ok := Builtins["contextWithDeadline"]
	if !ok {
		t.Fatal("contextWithDeadline builtin not found")
	}

	// Create deadline 100ms from now
	deadline := time.Now().Add(100 * time.Millisecond).UnixMilli()

	result := fn.Fn(NULL, &Int{Value: deadline})
	ctx, ok := result.(*Context)
	if !ok {
		t.Fatalf("contextWithDeadline should return *Context, got %T", result)
	}

	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		t.Error("Context should have a deadline")
	}

	// Test error cases
	result = fn.Fn() // No args
	if !isError(result) {
		t.Error("Should return error for no arguments")
	}

	result = fn.Fn(NULL, &String{Value: "not a number"}) // Wrong type
	if !isError(result) {
		t.Error("Should return error for wrong type")
	}
}

func TestBuiltinContextCancel(t *testing.T) {
	fn, ok := Builtins["contextCancel"]
	if !ok {
		t.Fatal("contextCancel builtin not found")
	}

	ctx := NewContextWithCancel(nil)
	result := fn.Fn(ctx)

	if result != NULL {
		t.Errorf("contextCancel should return NULL, got %v", result)
	}

	if !ctx.IsDone() {
		t.Error("Context should be done after cancel")
	}

	// Test error case
	result = fn.Fn() // No args
	if !isError(result) {
		t.Error("Should return error for no arguments")
	}

	result = fn.Fn(&Int{Value: 42}) // Wrong type
	if !isError(result) {
		t.Error("Should return error for wrong type")
	}
}

func TestBuiltinContextDone(t *testing.T) {
	fn, ok := Builtins["contextDone"]
	if !ok {
		t.Fatal("contextDone builtin not found")
	}

	ctx := NewBackgroundContext()
	result := fn.Fn(ctx)

	tube, ok := result.(*Tube)
	if !ok {
		t.Fatalf("contextDone should return *Tube, got %T", result)
	}

	if tube == nil {
		t.Error("Done tube should not be nil")
	}

	// Test error case
	result = fn.Fn() // No args
	if !isError(result) {
		t.Error("Should return error for no arguments")
	}
}

func TestBuiltinContextErr(t *testing.T) {
	fn, ok := Builtins["contextErr"]
	if !ok {
		t.Fatal("contextErr builtin not found")
	}

	// Active context - should return NULL
	ctx := NewBackgroundContext()
	result := fn.Fn(ctx)
	if result != NULL {
		t.Errorf("Active context should have NULL error, got %v", result)
	}

	// Cancelled context - should return error string
	ctx2 := NewContextWithCancel(nil)
	ctx2.Cancel()
	result = fn.Fn(ctx2)
	errStr, ok := result.(*String)
	if !ok {
		t.Fatalf("Cancelled context should return error string, got %T", result)
	}
	if errStr.Value == "" {
		t.Error("Error string should not be empty")
	}

	// Test error case
	result = fn.Fn() // No args
	if !isError(result) {
		t.Error("Should return error for no arguments")
	}
}

func TestBuiltinContextIsDone(t *testing.T) {
	fn, ok := Builtins["contextIsDone"]
	if !ok {
		t.Fatal("contextIsDone builtin not found")
	}

	// Active context
	ctx := NewBackgroundContext()
	result := fn.Fn(ctx)
	if result != FALSE {
		t.Errorf("Active context isDone should be FALSE, got %v", result)
	}

	// Cancelled context
	ctx2 := NewContextWithCancel(nil)
	ctx2.Cancel()
	result = fn.Fn(ctx2)
	if result != TRUE {
		t.Errorf("Cancelled context isDone should be TRUE, got %v", result)
	}

	// Test error case
	result = fn.Fn() // No args
	if !isError(result) {
		t.Error("Should return error for no arguments")
	}
}

func TestBuiltinContextDeadline(t *testing.T) {
	fn, ok := Builtins["contextDeadline"]
	if !ok {
		t.Fatal("contextDeadline builtin not found")
	}

	// Context without deadline
	ctx1 := NewBackgroundContext()
	result := fn.Fn(ctx1)
	if result != NULL {
		t.Errorf("Context without deadline should return NULL, got %v", result)
	}

	// Context with deadline
	deadline := time.Now().Add(500 * time.Millisecond)
	ctx2 := NewContextWithDeadline(nil, deadline)
	result = fn.Fn(ctx2)
	dlInt, ok := result.(*Int)
	if !ok {
		t.Fatalf("Context with deadline should return *Int, got %T", result)
	}
	if dlInt.Value == 0 {
		t.Error("Deadline value should not be zero")
	}

	// Test error case
	result = fn.Fn() // No args
	if !isError(result) {
		t.Error("Should return error for no arguments")
	}
}

// ============================================================
// Context Methods Tests
// ============================================================

func TestContextMethods(t *testing.T) {
	ctx := NewContextWithCancel(nil)

	// Test done method
	method, ok := GetMethod(ContextType, "done")
	if !ok {
		t.Fatal("done method not found")
	}
	result := method.Fn(ctx)
	if _, ok := result.(*Tube); !ok {
		t.Errorf("done method should return Tube, got %T", result)
	}

	// Test cancel method
	method, ok = GetMethod(ContextType, "cancel")
	if !ok {
		t.Fatal("cancel method not found")
	}
	result = method.Fn(ctx)
	if result != NULL {
		t.Errorf("cancel method should return NULL, got %v", result)
	}

	// Test isDone method
	method, ok = GetMethod(ContextType, "isDone")
	if !ok {
		t.Fatal("isDone method not found")
	}
	result = method.Fn(ctx)
	if result != TRUE {
		t.Errorf("isDone should be TRUE for cancelled context, got %v", result)
	}

	// Test err method
	method, ok = GetMethod(ContextType, "err")
	if !ok {
		t.Fatal("err method not found")
	}
	result = method.Fn(ctx)
	if _, ok := result.(*String); !ok {
		t.Errorf("err method should return String, got %T", result)
	}
}

func TestContextMethodDeadlineStr(t *testing.T) {
	ctx := NewBackgroundContext()

	method, ok := GetMethod(ContextType, "deadlineStr")
	if !ok {
		t.Fatal("deadlineStr method not found")
	}

	// No deadline
	result := method.Fn(ctx)
	if result != NULL {
		t.Errorf("deadlineStr should return NULL for context without deadline, got %v", result)
	}

	// With deadline
	deadline := time.Now().Add(1 * time.Hour)
	ctx2 := NewContextWithDeadline(nil, deadline)
	result = method.Fn(ctx2)
	str, ok := result.(*String)
	if !ok {
		t.Fatalf("deadlineStr should return String, got %T", result)
	}
	if str.Value == "" {
		t.Error("deadlineStr should not be empty")
	}
}

// ============================================================
// Integration Tests
// ============================================================

func TestContextIntegrationWithTimeout(t *testing.T) {
	// Create a context with a short timeout
	ctx := NewContextWithTimeout(nil, 50*time.Millisecond)

	// Check initial state
	if ctx.IsDone() {
		t.Error("Context should not be done initially")
	}

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Check final state
	if !ctx.IsDone() {
		t.Error("Context should be done after timeout")
	}

	errStr := ctx.ErrString()
	if errStr == "" {
		t.Error("Should have error after timeout")
	}
}

func TestContextIntegrationCascadingCancel(t *testing.T) {
	// Create a chain of contexts
	parent := NewContextWithCancel(nil)
	child1 := NewContextWithCancel(parent)
	child2 := NewContextWithCancel(child1)

	// Cancel parent
	parent.Cancel()

	// Wait for propagation
	time.Sleep(50 * time.Millisecond)

	// All children should be cancelled
	if !parent.IsDone() {
		t.Error("Parent should be done")
	}
	if !child1.IsDone() {
		t.Error("Child1 should be done")
	}
	if !child2.IsDone() {
		t.Error("Child2 should be done")
	}
}
