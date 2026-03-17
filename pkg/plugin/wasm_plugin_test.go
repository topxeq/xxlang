// pkg/plugin/wasm_plugin_test.go
// Comprehensive tests for WASM plugin loading and execution
package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Test helpers
// ============================================

func getWasmPluginPath(t *testing.T, name string) string {
	t.Helper()

	// Try multiple possible locations
	paths := []string{
		filepath.Join("examples", "wasm_plugin", "plugin", name+".wasm"),
		filepath.Join("..", "..", "examples", "wasm_plugin", "plugin", name+".wasm"),
		filepath.Join("testdata", "target", "wasm32-unknown-unknown", "release", name+".wasm"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			absPath, err := filepath.Abs(p)
			if err == nil {
				return absPath
			}
		}
	}

	t.Skipf("WASM plugin %s.wasm not found, skipping test", name)
	return ""
}

// ============================================
// WASM Loading Tests
// ============================================

func TestWasmLoader_LoadFibPlugin(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load fib.wasm: %v", err)
	}

	if p == nil {
		t.Fatal("expected plugin, got nil")
	}

	if p.Name() == "" {
		t.Error("expected non-empty plugin name")
	}

	t.Logf("Loaded plugin: %s", p.Name())
}

func TestWasmLoader_LoadTestPlugin(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "testplugin")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load testplugin.wasm: %v", err)
	}

	if p == nil {
		t.Fatal("expected plugin, got nil")
	}

	t.Logf("Loaded plugin: %s", p.Name())
}

// ============================================
// WASM Function Execution Tests
// ============================================

func TestWasmPlugin_FibFast(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()
	if len(exports) == 0 {
		t.Fatal("expected at least one export")
	}

	// Check for fast function
	fastFn, ok := exports["fast"]
	if !ok {
		t.Skip("fast function not exported")
	}

	builtin, ok := fastFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", fastFn)
	}

	// Test Fibonacci numbers
	testCases := []struct {
		input    int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{5, 5},
		{10, 55},
		{20, 6765},
	}

	for _, tc := range testCases {
		result := builtin.Fn(&objects.Int{Value: tc.input})
		intResult, ok := result.(*objects.Int)
		if !ok {
			t.Errorf("fast(%d) returned %T, expected Int", tc.input, result)
			continue
		}
		if intResult.Value != tc.expected {
			t.Errorf("fast(%d) = %d, expected %d", tc.input, intResult.Value, tc.expected)
		}
	}
}

func TestWasmPlugin_FibMatrix(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	matrixFn, ok := exports["matrix"]
	if !ok {
		t.Skip("matrix function not exported")
	}

	builtin, ok := matrixFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", matrixFn)
	}

	// Test matrix Fibonacci algorithm
	testCases := []struct {
		input    int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{10, 55},
		{15, 610},
	}

	for _, tc := range testCases {
		result := builtin.Fn(&objects.Int{Value: tc.input})
		intResult, ok := result.(*objects.Int)
		if !ok {
			t.Errorf("matrix(%d) returned %T, expected Int", tc.input, result)
			continue
		}
		if intResult.Value != tc.expected {
			t.Errorf("matrix(%d) = %d, expected %d", tc.input, intResult.Value, tc.expected)
		}
	}
}

func TestWasmPlugin_IsFib(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	isFibFn, ok := exports["isFib"]
	if !ok {
		t.Skip("isFib function not exported")
	}

	builtin, ok := isFibFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", isFibFn)
	}

	// Test Fibonacci number detection
	testCases := []struct {
		input    int64
		expected bool
	}{
		{0, true},
		{1, true},
		{2, true},
		{3, true},
		{5, true},
		{8, true},
		{13, true},
		{4, false},
		{6, false},
		{7, false},
	}

	for _, tc := range testCases {
		result := builtin.Fn(&objects.Int{Value: tc.input})
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Errorf("isFib(%d) returned %T, expected Bool", tc.input, result)
			continue
		}
		if boolResult.Value != tc.expected {
			t.Errorf("isFib(%d) = %v, expected %v", tc.input, boolResult.Value, tc.expected)
		}
	}
}

func TestWasmPlugin_Version(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	version, ok := exports["version"]
	if !ok {
		t.Skip("version not exported")
	}

	strVersion, ok := version.(*objects.String)
	if !ok {
		t.Fatalf("expected String, got %T", version)
	}

	if strVersion.Value == "" {
		t.Error("expected non-empty version string")
	}

	t.Logf("Plugin version: %s", strVersion.Value)
}

// ============================================
// WASM Error Handling Tests
// ============================================

func TestWasmPlugin_InvalidArgument(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	fastFn, ok := exports["fast"]
	if !ok {
		t.Skip("fast function not exported")
	}

	builtin, ok := fastFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", fastFn)
	}

	// Test with wrong argument type
	result := builtin.Fn(&objects.String{Value: "not a number"})
	if result == nil {
		t.Error("expected error for wrong argument type")
	}

	if errObj, ok := result.(*objects.Error); ok {
		if errObj.Message == "" {
			t.Error("expected error message")
		}
	}
}

func TestWasmPlugin_WrongArgumentCount(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	fastFn, ok := exports["fast"]
	if !ok {
		t.Skip("fast function not exported")
	}

	builtin, ok := fastFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", fastFn)
	}

	// Test with wrong number of arguments
	result := builtin.Fn(&objects.Int{Value: 1}, &objects.Int{Value: 2})
	if result == nil {
		t.Error("expected error for wrong argument count")
	}

	if errObj, ok := result.(*objects.Error); ok {
		if errObj.Message == "" {
			t.Error("expected error message")
		}
	}
}

// ============================================
// LoadPath Coverage Tests
// ============================================
// LoadPath Coverage Tests
// ============================================

func TestLoadPath_NonExistentFile(t *testing.T) {
	loader := NewLoader()

	p, err := loader.LoadPath("/nonexistent/path/to/plugin.wasm")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if p != nil {
		t.Error("expected nil plugin for non-existent file")
	}
}

func TestLoadPath_EmptyPath(t *testing.T) {
	loader := NewLoader()

	p, err := loader.LoadPath("")
	if err == nil {
		t.Error("expected error for empty path")
	}
	if p != nil {
		t.Error("expected nil plugin for empty path")
	}
}

func TestLoadPath_InvalidWasm(t *testing.T) {
	// Create a temp file with invalid WASM content
	tmpFile, err := os.CreateTemp("", "invalid*.wasm")
	if err != nil {
		t.Skip("could not create temp file")
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid content
	_, _ = tmpFile.WriteString("not a valid wasm file")
	_ = tmpFile.Close()

	loader := NewLoader()
	p, err := loader.LoadPath(tmpFile.Name())
	if err == nil {
		t.Error("expected error for invalid WASM file")
	}
	if p != nil {
		t.Error("expected nil plugin for invalid WASM")
	}
}

// ============================================
// Plugin Registry with WASM Tests
// ============================================

func TestWasmPlugin_RegisterAndRetrieve(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	// Register the plugin
	Register(p)

	// Retrieve it
	retrieved, ok := Get(p.Name())
	if !ok {
		t.Error("expected to retrieve registered plugin")
	}
	if retrieved == nil {
		t.Error("expected non-nil plugin")
	}
	if retrieved.Name() != p.Name() {
		t.Errorf("retrieved plugin name = %s, expected %s", retrieved.Name(), p.Name())
	}
}

// ============================================
// ToModule Tests with WASM
// ============================================

func TestToModule_WasmPlugin(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	m := ToModule(p)
	if m == nil {
		t.Fatal("expected module, got nil")
	}

	if m.Name != p.Name() {
		t.Errorf("module name = %s, expected %s", m.Name, p.Name())
	}

	if len(m.Exports) == 0 {
		t.Error("expected at least one export in module")
	}
}

// ============================================
// Large Fibonacci Test
// ============================================

func TestWasmPlugin_LargeFibonacci(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	matrixFn, ok := exports["matrix"]
	if !ok {
		t.Skip("matrix function not exported")
	}

	builtin, ok := matrixFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", matrixFn)
	}

	// Test large Fibonacci number (matrix algorithm handles this efficiently)
	result := builtin.Fn(&objects.Int{Value: 50})
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Errorf("matrix(50) returned %T, expected Int", result)
		return
	}

	// Fibonacci(50) = 12586269025
	if intResult.Value <= 0 {
		t.Errorf("matrix(50) = %d, expected positive number", intResult.Value)
	}

	t.Logf("Fibonacci(50) = %d", intResult.Value)
}

// ============================================
// Range Function Test
// ============================================

func TestWasmPlugin_Range(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	rangeFn, ok := exports["range_"]
	if !ok {
		t.Skip("range_ function not exported")
	}

	builtin, ok := rangeFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", rangeFn)
	}

	// Test range(10) - should return first 11 Fibonacci numbers
	result := builtin.Fn(&objects.Int{Value: 10})
	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Errorf("range_(10) returned %T, expected Array", result)
		return
	}

	if len(arrResult.Elements) != 11 {
		t.Errorf("range_(10) returned %d elements, expected 11", len(arrResult.Elements))
		return
	}

	// Verify first few values
	expectedFirstFive := []int64{0, 1, 1, 2, 3}
	for i, expected := range expectedFirstFive {
		val, ok := arrResult.Elements[i].(*objects.Int)
		if !ok {
			t.Errorf("element %d is %T, expected Int", i, arrResult.Elements[i])
			continue
		}
		if val.Value != expected {
			t.Errorf("range_(10)[%d] = %d, expected %d", i, val.Value, expected)
		}
	}
}

// ============================================
// Concurrent Access Tests
// ============================================

func TestWasmPlugin_ConcurrentAccess(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	exports := p.Exports()

	fastFn, ok := exports["fast"]
	if !ok {
		t.Skip("fast function not exported")
	}

	builtin, ok := fastFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", fastFn)
	}

	// Run concurrent calls
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			result := builtin.Fn(&objects.Int{Value: int64(n)})
			if result == nil {
				t.Errorf("concurrent call %d returned nil", n)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================
// VM Integration Test
// ============================================

func TestWasmPlugin_VMIntegration(t *testing.T) {
	// This tests that WASM plugins work correctly when loaded through VM
	// The actual VM tests are in pkg/vm, but we test the plugin side here

	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	// Convert to module for VM use
	m := ToModule(p)

	// Verify the module has callable functions
	for name, export := range m.Exports {
		if builtin, ok := export.(*objects.Builtin); ok {
			// Test that the function is callable
			if name == "fast" {
				result := builtin.Fn(&objects.Int{Value: 10})
				if result == nil {
					t.Errorf("function %s returned nil", name)
				}
			}
		}
	}
}

// ============================================
// WasmPlugin Close Tests
// ============================================
// NOTE: This test must run LAST because closing a WASM plugin
// affects other plugins loaded from the same file due to shared
// module state in the gowasm runtime.

func TestWasmPlugin_Close(t *testing.T) {
	pluginPath := getWasmPluginPath(t, "fib")

	loader := NewLoader()
	p, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("failed to load plugin: %v", err)
	}

	// Get exports before close
	exports := p.Exports()
	if len(exports) == 0 {
		t.Error("expected exports before close")
	}

	// Close should not panic
	ctx := context.Background()
	if wp, ok := p.(*WasmPlugin); ok {
		err := wp.Close(ctx)
		if err != nil {
			t.Errorf("Close returned error: %v", err)
		}
	}
}
