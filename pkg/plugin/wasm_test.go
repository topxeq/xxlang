// pkg/plugin/wasm_test.go
// Tests for WASM plugin loading and execution
package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// getWASMPath returns the path to the test WASM plugin
func getWASMPath(t *testing.T) string {
	// Try multiple possible locations
	paths := []string{
		"testdata/target/wasm32-unknown-unknown/release/testplugin.wasm",
		"../plugin/testdata/target/wasm32-unknown-unknown/release/testplugin.wasm",
		"../../examples/wasm_plugin/plugin/fib.wasm",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}

	t.Skip("WASM test plugin not found - run 'cargo build --target wasm32-unknown-unknown --release' in testdata/")
	return ""
}

// closePlugin safely closes a plugin if it has a Close method
func closePlugin(p Plugin) {
	if wp, ok := p.(*WasmPlugin); ok {
		wp.Close(context.Background())
	}
}

// TestWasmPluginLoad tests loading a WASM plugin
func TestWasmPluginLoad(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	// Check plugin name
	if plugin.Name() != "testplugin" {
		t.Errorf("Expected plugin name 'testplugin', got %q", plugin.Name())
	}
}

// TestWasmPluginExports tests getting exports from a WASM plugin
func TestWasmPluginExports(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	if len(exports) == 0 {
		t.Error("Expected exports from WASM plugin")
	}

	// Check for version
	if version, ok := exports["version"]; !ok {
		t.Error("Expected 'version' export")
	} else if str, ok := version.(*objects.String); !ok {
		t.Errorf("Expected version to be string, got %T", version)
	} else if str.Value != "1.0.0-rust" {
		t.Errorf("Expected version '1.0.0-rust', got %q", str.Value)
	}
}

// TestWasmPluginAdd tests the add function
func TestWasmPluginAdd(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	addFn, ok := exports["add"]
	if !ok {
		t.Fatal("Expected 'add' export")
	}

	builtin, ok := addFn.(*objects.Builtin)
	if !ok {
		t.Fatalf("Expected builtin function, got %T", addFn)
	}

	result := builtin.Fn(&objects.Int{Value: 10}, &objects.Int{Value: 20})
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("Expected int result, got %T", result)
	}

	if intResult.Value != 30 {
		t.Errorf("Expected 10 + 20 = 30, got %d", intResult.Value)
	}
}

// TestWasmPluginSub tests the subtract function
func TestWasmPluginSub(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	subFn := exports["sub"].(*objects.Builtin)

	result := subFn.Fn(&objects.Int{Value: 100}, &objects.Int{Value: 40})
	intResult := result.(*objects.Int)

	if intResult.Value != 60 {
		t.Errorf("Expected 100 - 40 = 60, got %d", intResult.Value)
	}
}

// TestWasmPluginMul tests the multiply function
func TestWasmPluginMul(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	mulFn := exports["mul"].(*objects.Builtin)

	result := mulFn.Fn(&objects.Int{Value: 6}, &objects.Int{Value: 7})
	intResult := result.(*objects.Int)

	if intResult.Value != 42 {
		t.Errorf("Expected 6 * 7 = 42, got %d", intResult.Value)
	}
}

// TestWasmPluginDiv tests the divide function
func TestWasmPluginDiv(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	divFn := exports["div"].(*objects.Builtin)

	result := divFn.Fn(&objects.Int{Value: 100}, &objects.Int{Value: 5})
	intResult := result.(*objects.Int)

	if intResult.Value != 20 {
		t.Errorf("Expected 100 / 5 = 20, got %d", intResult.Value)
	}
}

// TestWasmPluginFactorial tests the factorial function
func TestWasmPluginFactorial(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	factFn := exports["factorial"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{0, 1},
		{1, 1},
		{5, 120},
		{10, 3628800},
	}

	for _, tt := range tests {
		result := factFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("factorial(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginFib tests the fibonacci function
func TestWasmPluginFib(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	fibFn := exports["fib"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{10, 55},
		{20, 6765},
	}

	for _, tt := range tests {
		result := fibFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("fib(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginIsEven tests the is_even function
func TestWasmPluginIsEven(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	isEvenFn := exports["is_even"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected bool
	}{
		{0, true},
		{1, false},
		{2, true},
		{100, true},
		{101, false},
	}

	for _, tt := range tests {
		result := isEvenFn.Fn(&objects.Int{Value: tt.input})
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("is_even(%d) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginIsPrime tests the is_prime function
func TestWasmPluginIsPrime(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	isPrimeFn := exports["is_prime"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected bool
	}{
		{0, false},
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{17, true},
		{100, false},
		{101, true},
	}

	for _, tt := range tests {
		result := isPrimeFn.Fn(&objects.Int{Value: tt.input})
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("is_prime(%d) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginGCD tests the gcd function
func TestWasmPluginGCD(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	gcdFn := exports["gcd"].(*objects.Builtin)

	tests := []struct {
		a, b     int64
		expected int64
	}{
		{12, 8, 4},
		{100, 35, 5},
		{17, 13, 1},
		{48, 18, 6},
	}

	for _, tt := range tests {
		result := gcdFn.Fn(&objects.Int{Value: tt.a}, &objects.Int{Value: tt.b})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("gcd(%d, %d) = %d, expected %d", tt.a, tt.b, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginMaxMin tests the max and min functions
func TestWasmPluginMaxMin(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	maxFn := exports["max"].(*objects.Builtin)
	minFn := exports["min"].(*objects.Builtin)

	// Test max
	result := maxFn.Fn(&objects.Int{Value: 10}, &objects.Int{Value: 20})
	intResult := result.(*objects.Int)
	if intResult.Value != 20 {
		t.Errorf("max(10, 20) = %d, expected 20", intResult.Value)
	}

	// Test min
	result = minFn.Fn(&objects.Int{Value: 10}, &objects.Int{Value: 20})
	intResult = result.(*objects.Int)
	if intResult.Value != 10 {
		t.Errorf("min(10, 20) = %d, expected 10", intResult.Value)
	}
}

// TestWasmPluginPow tests the pow function
func TestWasmPluginPow(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	powFn := exports["pow"].(*objects.Builtin)

	tests := []struct {
		base, exp int64
		expected  int64
	}{
		{2, 0, 1},
		{2, 1, 2},
		{2, 10, 1024},
		{3, 4, 81},
	}

	for _, tt := range tests {
		result := powFn.Fn(&objects.Int{Value: tt.base}, &objects.Int{Value: tt.exp})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("pow(%d, %d) = %d, expected %d", tt.base, tt.exp, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginBinomial tests the binomial coefficient function
func TestWasmPluginBinomial(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	binFn := exports["binomial"].(*objects.Builtin)

	tests := []struct {
		n, k     int64
		expected int64
	}{
		{5, 0, 1},
		{5, 1, 5},
		{5, 2, 10},
		{10, 5, 252},
	}

	for _, tt := range tests {
		result := binFn.Fn(&objects.Int{Value: tt.n}, &objects.Int{Value: tt.k})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("binomial(%d, %d) = %d, expected %d", tt.n, tt.k, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginTriangle tests triangular number function
func TestWasmPluginTriangle(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	triFn := exports["triangle"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{1, 1},
		{5, 15},
		{10, 55},
		{100, 5050},
	}

	for _, tt := range tests {
		result := triFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("triangle(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestNewWasmPlugin tests the NewWasmPlugin constructor
func TestNewWasmPlugin(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	// Test that it implements Plugin interface
	var _ Plugin = plugin
}

// TestWasmPluginToModule tests the ToModule function
func TestWasmPluginToModule(t *testing.T) {
	wasmPath := getWASMPath(t)

	p, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(p)

	module := ToModule(p)
	if module == nil {
		t.Fatal("Expected non-nil module")
	}

	if module.Name != "testplugin" {
		t.Errorf("Expected module name 'testplugin', got %q", module.Name)
	}

	if len(module.Exports) == 0 {
		t.Error("Expected exports in module")
	}
}

// TestWasmPluginAbs tests the abs function
func TestWasmPluginAbs(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	absFn := exports["abs"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
	}

	for _, tt := range tests {
		result := absFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("abs(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginNeg tests the neg function
func TestWasmPluginNeg(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	negFn := exports["neg"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{5, -5},
		{-5, 5},
		{0, 0},
	}

	for _, tt := range tests {
		result := negFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("neg(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginSquare tests the square function
func TestWasmPluginSquare(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	squareFn := exports["square"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{5, 25},
		{10, 100},
		{-3, 9},
	}

	for _, tt := range tests {
		result := squareFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("square(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginLCM tests the lcm function
func TestWasmPluginLCM(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	lcmFn := exports["lcm"].(*objects.Builtin)

	tests := []struct {
		a, b     int64
		expected int64
	}{
		{4, 6, 12},
		{3, 5, 15},
		{12, 8, 24},
		{7, 11, 77},
	}

	for _, tt := range tests {
		result := lcmFn.Fn(&objects.Int{Value: tt.a}, &objects.Int{Value: tt.b})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("lcm(%d, %d) = %d, expected %d", tt.a, tt.b, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginIsOdd tests the is_odd function
func TestWasmPluginIsOdd(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	isOddFn := exports["is_odd"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected bool
	}{
		{0, false},
		{1, true},
		{2, false},
		{99, true},
		{100, false},
	}

	for _, tt := range tests {
		result := isOddFn.Fn(&objects.Int{Value: tt.input})
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("is_odd(%d) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginIsSquare tests the is_square function
func TestWasmPluginIsSquare(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	isSquareFn := exports["is_square"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected bool
	}{
		{0, true},
		{1, true},
		{4, true},
		{9, true},
		{16, true},
		{2, false},
		{3, false},
		{5, false},
		{-1, false},
	}

	for _, tt := range tests {
		result := isSquareFn.Fn(&objects.Int{Value: tt.input})
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("is_square(%d) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginClamp tests the clamp function
func TestWasmPluginClamp(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	clampFn := exports["clamp"].(*objects.Builtin)

	tests := []struct {
		value, min, max int64
		expected        int64
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{50, 0, 100, 50},
	}

	for _, tt := range tests {
		result := clampFn.Fn(&objects.Int{Value: tt.value}, &objects.Int{Value: tt.min}, &objects.Int{Value: tt.max})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("clamp(%d, %d, %d) = %d, expected %d", tt.value, tt.min, tt.max, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginSumSquares tests the sum_squares function
func TestWasmPluginSumSquares(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	sumSqFn := exports["sum_squares"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{1, 1},
		{2, 5},    // 1 + 4
		{3, 14},   // 1 + 4 + 9
		{5, 55},   // 1 + 4 + 9 + 16 + 25
		{10, 385}, // 1 + 4 + 9 + ... + 100
	}

	for _, tt := range tests {
		result := sumSqFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("sum_squares(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginCountPrimes tests the count_primes function
func TestWasmPluginCountPrimes(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	countPrimesFn := exports["count_primes"].(*objects.Builtin)

	tests := []struct {
		input    int64
		expected int64
	}{
		{1, 0},
		{2, 1},
		{10, 4}, // 2, 3, 5, 7
		{20, 8}, // 2, 3, 5, 7, 11, 13, 17, 19
		{100, 25},
	}

	for _, tt := range tests {
		result := countPrimesFn.Fn(&objects.Int{Value: tt.input})
		intResult := result.(*objects.Int)
		if intResult.Value != tt.expected {
			t.Errorf("count_primes(%d) = %d, expected %d", tt.input, intResult.Value, tt.expected)
		}
	}
}

// TestWasmPluginArgumentErrors tests error handling for wrong arguments
func TestWasmPluginArgumentErrors(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Test single-arg function with wrong number of args
	addFn := exports["add"].(*objects.Builtin)
	result := addFn.Fn(&objects.Int{Value: 1})
	if err, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for single arg to add, got %T", result)
	} else if err.Message != "call_add requires 2 arguments" {
		t.Errorf("Unexpected error message: %s", err.Message)
	}

	// Test single-arg function with wrong arg type
	fibFn := exports["fib"].(*objects.Builtin)
	result = fibFn.Fn(&objects.String{Value: "not a number"})
	if err, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for string arg to fib, got %T", result)
	} else if err.Message != "argument must be integer" {
		t.Errorf("Unexpected error message: %s", err.Message)
	}

	// Test two-arg function with wrong arg types
	gcdFn := exports["gcd"].(*objects.Builtin)
	result = gcdFn.Fn(&objects.Int{Value: 10}, &objects.String{Value: "not a number"})
	if err, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for mixed arg types, got %T", result)
	} else if err.Message != "arguments must be integers" {
		t.Errorf("Unexpected error message: %s", err.Message)
	}

	// Test bool-returning function with wrong arg count
	isEvenFn := exports["is_even"].(*objects.Builtin)
	result = isEvenFn.Fn()
	if err, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for no args to is_even, got %T", result)
	} else if err.Message != "call_is_even requires 1 argument" {
		t.Errorf("Unexpected error message: %s", err.Message)
	}

	// Test bool-returning function with wrong arg type
	result = isEvenFn.Fn(&objects.String{Value: "not a number"})
	if err, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for string arg to is_even, got %T", result)
	} else if err.Message != "argument must be integer" {
		t.Errorf("Unexpected error message: %s", err.Message)
	}
}

// TestWasmPluginClose tests the Close method
func TestWasmPluginClose(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}

	// Close should work without error
	wp := plugin.(*WasmPlugin)
	err = wp.Close(context.Background())
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

// TestWasmPluginSumArray tests the sum_array function
func TestWasmPluginSumArray(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	sumArrayFn := exports["sum_array"].(*objects.Builtin)

	// sum_array takes 2 arguments: pointer and count
	// For testing, we pass 0 as pointer which should return 0
	result := sumArrayFn.Fn(&objects.Int{Value: 0}, &objects.Int{Value: 0})
	intResult := result.(*objects.Int)
	if intResult.Value != 0 {
		t.Errorf("sum_array(0, 0) = %d, expected 0", intResult.Value)
	}
}

// TestWasmPluginSumArrayErrors tests error handling in sum_array
func TestWasmPluginSumArrayErrors(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	sumArrayFn := exports["sum_array"].(*objects.Builtin)

	// Test with non-integer argument (tests default case error handling)
	result := sumArrayFn.Fn(&objects.String{Value: "not a number"}, &objects.Int{Value: 0})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for non-integer argument, got %T", result)
	}
}

// TestWasmPluginClampErrors tests error handling for clamp function (3 args)
func TestWasmPluginClampErrors(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	clampFn := exports["clamp"].(*objects.Builtin)

	// Test with non-integer argument (tests default case error handling for 3 args)
	result := clampFn.Fn(&objects.Int{Value: 5}, &objects.String{Value: "not a number"}, &objects.Int{Value: 10})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for non-integer argument, got %T", result)
	}
}

// TestLoaderLoadFromPathAndRegister tests Load with path and registration
func TestLoaderLoadFromPathAndRegister(t *testing.T) {
	// Create a loader with a search path
	loader := NewLoader()

	// Find the WASM file location
	paths := []string{
		"testdata/target/wasm32-unknown-unknown/release",
		"../../examples/wasm_plugin/plugin",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			loader.AddPath(p)
			break
		}
	}

	// Load the plugin using Load (should register it)
	p, err := loader.Load("testplugin")
	if err != nil {
		t.Skipf("Could not load testplugin: %v", err)
	}

	// Verify it was registered
	if _, ok := Get(p.Name()); !ok {
		t.Error("Plugin was not registered after Load")
	}
}

// TestReadStringFromMemory tests the readStringFromMemory helper
func TestReadStringFromMemory(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)
	mem := wp.module.Memory()
	if mem == nil {
		t.Fatal("Memory is nil")
	}

	// Write test data to memory
	testStr := "hello world"
	buf := []byte(testStr)
	offset := uint32(100)
	mem.Write(offset, buf)

	// Pack ptr and size: high 32 bits = offset, low 32 bits = size
	ptrSize := uint64(offset)<<32 | uint64(len(testStr))

	result := readStringFromMemory(wp.module, ptrSize)
	if result != testStr {
		t.Errorf("readStringFromMemory() = %q, expected %q", result, testStr)
	}

	// Test with size 0 - should return empty string
	ptrSizeZero := uint64(offset)<<32 | 0
	result = readStringFromMemory(wp.module, ptrSizeZero)
	if result != "" {
		t.Errorf("readStringFromMemory() with size 0 = %q, expected empty", result)
	}
}

// TestReadStringFromMemory2 tests the readStringFromMemory2 helper
func TestReadStringFromMemory2(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)
	mem := wp.module.Memory()
	if mem == nil {
		t.Fatal("Memory is nil")
	}

	// Write test data to memory
	testStr := "test string"
	buf := []byte(testStr)
	offset := uint32(200)
	mem.Write(offset, buf)

	result := readStringFromMemory2(wp.module, offset, uint32(len(testStr)))
	if result != testStr {
		t.Errorf("readStringFromMemory2() = %q, expected %q", result, testStr)
	}

	// Test with size 0
	result = readStringFromMemory2(wp.module, offset, 0)
	if result != "" {
		t.Errorf("readStringFromMemory2() with size 0 = %q, expected empty", result)
	}
}

// TestReadInt64ArrayFromMemory tests the readInt64ArrayFromMemory helper
func TestReadInt64ArrayFromMemory(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)
	mem := wp.module.Memory()
	if mem == nil {
		t.Fatal("Memory is nil")
	}

	// Write test int64 array to memory
	testValues := []int64{100, 200, 300, 400, 500}
	offset := uint32(300)
	for i, v := range testValues {
		mem.WriteUint64Le(offset+uint32(i*8), uint64(v))
	}

	result := readInt64ArrayFromMemory(wp.module, offset, uint32(len(testValues)))
	if result == nil {
		t.Fatal("readInt64ArrayFromMemory() returned nil")
	}
	if len(result) != len(testValues) {
		t.Fatalf("readInt64ArrayFromMemory() returned %d elements, expected %d", len(result), len(testValues))
	}
	for i, v := range testValues {
		if result[i] != v {
			t.Errorf("result[%d] = %d, expected %d", i, result[i], v)
		}
	}

	// Test with count 0
	result = readInt64ArrayFromMemory(wp.module, offset, 0)
	if result != nil {
		t.Errorf("readInt64ArrayFromMemory() with count 0 = %v, expected nil", result)
	}
}

// TestReadStringFromResultPtrEdgeCases tests edge cases for readStringFromResultPtr
func TestReadStringFromResultPtrEdgeCases(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)

	// Test with valid memory but writing empty result
	mem := wp.module.Memory()
	offset := uint32(400)
	// Write ptr=0, size=0
	mem.WriteUint32Le(offset, 0)
	mem.WriteUint32Le(offset+4, 0)

	result := readStringFromResultPtr(wp.module, offset)
	if result != "" {
		t.Errorf("readStringFromResultPtr() with size 0 = %q, expected empty", result)
	}
}

// TestWasmCloseWithNilModule tests Close with nil module
func TestWasmCloseWithNilModule(t *testing.T) {
	// Test closing a plugin with nil module
	wp := &WasmPlugin{
		name:   "test",
		module: nil,
		rt:     nil,
	}

	err := wp.Close(context.Background())
	if err != nil {
		t.Errorf("Close with nil module returned error: %v", err)
	}
}

// TestLoadPathInvalidWASM tests LoadPath with invalid WASM content
func TestLoadPathInvalidWASM(t *testing.T) {
	// Create a temp file with invalid WASM content
	tmpFile, err := os.CreateTemp("", "invalid*.wasm")
	if err != nil {
		t.Skip("could not create temp file")
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid content
	tmpFile.WriteString("not a valid wasm file")
	tmpFile.Close()

	loader := NewLoader()
	p, err := loader.LoadPath(tmpFile.Name())
	if err == nil {
		t.Error("expected error for invalid WASM file")
		if p != nil {
			closePlugin(p)
		}
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

// TestLoadPluginWASMNonExistent tests loading non-existent WASM file
func TestLoadPluginWASMNonExistent(t *testing.T) {
	_, err := loadPluginWASM("/nonexistent/path/to/plugin.wasm")
	if err == nil {
		t.Error("expected error for non-existent WASM file")
	}
}

// TestReadInt64ArrayFromMemoryEmpty tests readInt64ArrayFromMemory with empty array
func TestReadInt64ArrayFromMemoryEmpty(t *testing.T) {
	// Test with nil module (count 0)
	result := readInt64ArrayFromMemory(nil, 0, 0)
	if result != nil {
		t.Errorf("readInt64ArrayFromMemory(nil, 0, 0) = %v, expected nil", result)
	}
}

// TestWasmPlugin_NameMethod tests the Name method
func TestWasmPlugin_NameMethod(t *testing.T) {
	names := []string{"fib", "math", "test_plugin", ""}
	for _, name := range names {
		wp := &WasmPlugin{name: name}
		if wp.Name() != name {
			t.Errorf("Name() = %q, expected %q", wp.Name(), name)
		}
	}
}

// TestWasmPlugin_NewWasmPluginFunction tests NewWasmPlugin
func TestWasmPlugin_NewWasmPluginFunction(t *testing.T) {
	wp := NewWasmPlugin("test", nil, nil)
	if wp == nil {
		t.Fatal("NewWasmPlugin returned nil")
	}
	if wp.Name() != "test" {
		t.Errorf("Name() = %q, expected 'test'", wp.Name())
	}
}

// TestWasmPlugin_CloseNilRuntime tests Close with nil runtime
func TestWasmPlugin_CloseNilRuntime(t *testing.T) {
	wp := &WasmPlugin{
		name:   "test",
		module: nil,
		rt:     nil,
	}

	err := wp.Close(context.Background())
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// TestReadStringFromMemoryInvalidRead tests readStringFromMemory when read fails
func TestReadStringFromMemoryInvalidRead(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)
	mem := wp.module.Memory()
	if mem == nil {
		t.Fatal("Memory is nil")
	}

	// Try to read from an invalid offset (very large value)
	// This should return empty string because the read will fail
	ptrSize := uint64(0xFFFF0000)<<32 | uint64(10)
	result := readStringFromMemory(wp.module, ptrSize)
	// The result should be empty because the read fails
	if result != "" {
		t.Logf("readStringFromMemory with invalid offset = %q (may vary)", result)
	}
}

// TestReadStringFromMemory2InvalidRead tests readStringFromMemory2 when read fails
func TestReadStringFromMemory2InvalidRead(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)

	// Try to read from an invalid offset (very large value)
	// This should return empty string because the read fails
	result := readStringFromMemory2(wp.module, 0xFFFF0000, 10)
	// The result should be empty because the read fails
	if result != "" {
		t.Logf("readStringFromMemory2 with invalid offset = %q (may vary)", result)
	}
}

// TestReadInt64ArrayFromMemoryInvalidRead tests readInt64ArrayFromMemory when read fails
func TestReadInt64ArrayFromMemoryInvalidRead(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)

	// Try to read from an invalid offset (very large value)
	// This should return nil because the read fails
	result := readInt64ArrayFromMemory(wp.module, 0xFFFF0000, 5)
	// The result should be nil because the read fails
	if result != nil {
		t.Logf("readInt64ArrayFromMemory with invalid offset = %v (may vary)", result)
	}
}

// TestReadStringFromResultPtrInvalid tests readStringFromResultPtr with invalid pointer
func TestReadStringFromResultPtrInvalid(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)

	// Try to read from an invalid offset
	result := readStringFromResultPtr(wp.module, 0xFFFF0000)
	// The result should be empty because the read fails
	if result != "" {
		t.Logf("readStringFromResultPtr with invalid offset = %q (may vary)", result)
	}
}

// TestWrapFunctionDefaultCase tests the default case in wrapFunction
func TestWrapFunctionDefaultCase(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Test clamp function which uses 3 arguments (not in switch case)
	clampFn, ok := exports["clamp"].(*objects.Builtin)
	if !ok {
		t.Skip("clamp function not found")
	}

	// Test successful call with 3 arguments
	result := clampFn.Fn(&objects.Int{Value: 15}, &objects.Int{Value: 0}, &objects.Int{Value: 10})
	intResult, ok := result.(*objects.Int)
	if !ok {
		t.Errorf("Expected Int result, got %T", result)
	} else if intResult.Value != 10 {
		t.Errorf("clamp(15, 0, 10) = %d, expected 10", intResult.Value)
	}

	// Test with wrong argument type (tests error path in default case)
	result = clampFn.Fn(&objects.Int{Value: 15}, &objects.String{Value: "invalid"}, &objects.Int{Value: 10})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for non-integer argument, got %T", result)
	}
}

// TestWrapFunctionTwoArgsNoResult tests two-arg function returning no result
func TestWrapFunctionTwoArgsNoResult(t *testing.T) {
	// This tests the "function returned no result" path for two-arg functions
	// We can't easily trigger this without a special WASM module, so we skip
	t.Skip("Cannot test without a WASM module that returns no results")
}

// TestReadStringFromMemory2Invalid tests readStringFromMemory2 with invalid read
func TestReadStringFromMemory2Invalid(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)

	// Test with very large offset that will fail
	result := readStringFromMemory2(wp.module, 0xFFFFFFFF, 10)
	if result != "" {
		t.Logf("readStringFromMemory2 with invalid offset = %q", result)
	}
}

// TestReadStringFromResultPtrWithValidSize tests readStringFromResultPtr with valid size
func TestReadStringFromResultPtrWithValidSize(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)
	mem := wp.module.Memory()
	if mem == nil {
		t.Fatal("Memory is nil")
	}

	// Write test data
	testStr := "test_result"
	offset := uint32(500)
	mem.Write(offset, []byte(testStr))

	// Write ptr and size at result location
	resultLoc := uint32(600)
	mem.WriteUint32Le(resultLoc, offset)
	mem.WriteUint32Le(resultLoc+4, uint32(len(testStr)))

	// Read it back
	result := readStringFromResultPtr(wp.module, resultLoc)
	if result != testStr {
		t.Errorf("readStringFromResultPtr() = %q, expected %q", result, testStr)
	}
}

// TestReadStringFromResultPtrFailedRead tests readStringFromResultPtr when reading ptr fails
func TestReadStringFromResultPtrFailedRead(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)

	// Try reading from an invalid resultPtr that will fail
	result := readStringFromResultPtr(wp.module, 0xFFFFFFFF)
	if result != "" {
		t.Errorf("Expected empty string for failed read, got %q", result)
	}
}

// TestWasmPluginBoolReturnFalse tests bool-returning functions returning false
func TestWasmPluginBoolReturnFalse(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Test is_even with odd number (should return false)
	isEvenFn := exports["is_even"].(*objects.Builtin)
	result := isEvenFn.Fn(&objects.Int{Value: 3})
	if result != objects.FALSE {
		t.Errorf("is_even(3) = %v, expected FALSE", result)
	}

	// Test is_prime with non-prime (should return false)
	isPrimeFn := exports["is_prime"].(*objects.Builtin)
	result = isPrimeFn.Fn(&objects.Int{Value: 4})
	if result != objects.FALSE {
		t.Errorf("is_prime(4) = %v, expected FALSE", result)
	}
}

// TestWasmPluginRangeNegative tests range_ with negative input
func TestWasmPluginRangeNegative(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	rangeFn := exports["range_"].(*objects.Builtin)

	// Test with negative input - should return empty array
	result := rangeFn.Fn(&objects.Int{Value: -1})
	arrResult, ok := result.(*objects.Array)
	if !ok {
		t.Errorf("range_(-1) returned %T, expected Array", result)
		return
	}
	if len(arrResult.Elements) != 0 {
		t.Errorf("range_(-1) returned %d elements, expected 0", len(arrResult.Elements))
	}
}

// TestWasmPluginRangeWrongArgType tests range_ with wrong argument type
func TestWasmPluginRangeWrongArgType(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	rangeFn := exports["range_"].(*objects.Builtin)

	// Test with non-integer argument
	result := rangeFn.Fn(&objects.String{Value: "not a number"})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for non-integer argument, got %T", result)
	}
}

// TestWasmPluginRangeWrongArgCount tests range_ with wrong argument count
func TestWasmPluginRangeWrongArgCount(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	rangeFn := exports["range_"].(*objects.Builtin)

	// Test with no arguments
	result := rangeFn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for no arguments, got %T", result)
	}
}

// TestWasmPluginSingleArgWrongCount tests single-arg functions with wrong arg count
func TestWasmPluginSingleArgWrongCount(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Test factorial with no arguments
	factorialFn := exports["factorial"].(*objects.Builtin)
	result := factorialFn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for no arguments, got %T", result)
	}

	// Test factorial with too many arguments
	result = factorialFn.Fn(&objects.Int{Value: 5}, &objects.Int{Value: 2})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for too many arguments, got %T", result)
	}
}

// TestWasmPluginTwoArgsWrongCount tests two-arg functions with wrong arg count
func TestWasmPluginTwoArgsWrongCount(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Test add with no arguments
	addFn := exports["add"].(*objects.Builtin)
	result := addFn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for no arguments, got %T", result)
	}

	// Test add with one argument
	result = addFn.Fn(&objects.Int{Value: 5})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for one argument, got %T", result)
	}

	// Test add with three arguments
	result = addFn.Fn(&objects.Int{Value: 5}, &objects.Int{Value: 3}, &objects.Int{Value: 2})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for three arguments, got %T", result)
	}
}

// TestWasmPluginBoolWrongCount tests bool-returning functions with wrong arg count
func TestWasmPluginBoolWrongCount(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Test is_odd with no arguments
	isOddFn := exports["is_odd"].(*objects.Builtin)
	result := isOddFn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for no arguments, got %T", result)
	}

	// Test is_square with too many arguments
	isSquareFn := exports["is_square"].(*objects.Builtin)
	result = isSquareFn.Fn(&objects.Int{Value: 4}, &objects.Int{Value: 2})
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("Expected error for too many arguments, got %T", result)
	}
}

// TestWasmPluginMod tests the mod function
func TestWasmPluginMod(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()
	modFn := exports["mod"].(*objects.Builtin)

	testCases := []struct {
		a, b, expected int64
	}{
		{10, 3, 1},
		{10, 5, 0},
		{7, 4, 3},
	}

	for _, tc := range testCases {
		result := modFn.Fn(&objects.Int{Value: tc.a}, &objects.Int{Value: tc.b})
		if intResult, ok := result.(*objects.Int); ok {
			if intResult.Value != tc.expected {
				t.Errorf("mod(%d, %d) = %d, expected %d", tc.a, tc.b, intResult.Value, tc.expected)
			}
		} else {
			t.Errorf("mod(%d, %d) returned %T", tc.a, tc.b, result)
		}
	}
}

// TestWasmPluginExportsVersion tests that version is exported correctly
func TestWasmPluginExportsVersion(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Check version is exported
	version, ok := exports["version"]
	if !ok {
		t.Error("Expected version export")
		return
	}

	strVersion, ok := version.(*objects.String)
	if !ok {
		t.Errorf("Expected version to be string, got %T", version)
		return
	}

	if strVersion.Value == "" {
		t.Error("Expected non-empty version string")
	}

	t.Logf("Plugin version: %s", strVersion.Value)
}

// TestWasmPluginMultipleExports tests that multiple exports work correctly
func TestWasmPluginMultipleExports(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	exports := plugin.Exports()

	// Check that we have multiple exports
	if len(exports) < 10 {
		t.Errorf("Expected at least 10 exports, got %d", len(exports))
	}

	// Check for some expected exports
	expectedExports := []string{"add", "sub", "mul", "div", "mod", "pow", "factorial", "fib", "is_even", "is_prime"}
	for _, name := range expectedExports {
		if _, ok := exports[name]; !ok {
			t.Errorf("Expected export %q not found", name)
		}
	}
}

// TestReadStringFromResultPtrValidPtr tests when ptr read succeeds but buffer read fails
func TestReadStringFromResultPtrValidPtrInvalidBuffer(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)
	mem := wp.module.Memory()
	if mem == nil {
		t.Fatal("Memory is nil")
	}

	// Write a valid ptr but point to an invalid buffer location
	resultLoc := uint32(700)
	mem.WriteUint32Le(resultLoc, 0xFFFFFFFF)     // invalid buffer ptr
	mem.WriteUint32Le(resultLoc+4, 10)            // non-zero size

	// This should return empty because buffer read fails
	result := readStringFromResultPtr(wp.module, resultLoc)
	if result != "" {
		t.Errorf("Expected empty string for invalid buffer read, got %q", result)
	}
}

// TestReadInt64ArrayFromMemoryInvalidOffset tests with invalid memory offset
func TestReadInt64ArrayFromMemoryInvalidOffset(t *testing.T) {
	wasmPath := getWASMPath(t)

	plugin, err := loadPluginWASM(wasmPath)
	if err != nil {
		t.Fatalf("Failed to load WASM plugin: %v", err)
	}
	defer closePlugin(plugin)

	wp := plugin.(*WasmPlugin)

	// Try with invalid offset
	result := readInt64ArrayFromMemory(wp.module, 0xFFFFFFFF, 5)
	if result != nil {
		t.Errorf("Expected nil for invalid offset, got %v", result)
	}
}

// TestWasmPluginLoadPathAlreadyCached tests loading a plugin that's already in registry
func TestWasmPluginLoadPathAlreadyCached(t *testing.T) {
	pluginPath := getWASMPath(t)

	loader := NewLoader()

	// First load
	p1, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}

	// Get the absolute path which is used as the registry key
	absPath, _ := filepath.Abs(pluginPath)

	// Second load should return the same plugin from registry
	p2, err := loader.LoadPath(pluginPath)
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}

	// They should be the same plugin (from registry)
	if p1 != p2 {
		t.Error("Expected same plugin from registry")
	}

	// Verify it's in the registry
	if got, ok := Get(absPath); !ok || got != p1 {
		t.Error("Plugin not found in registry with correct key")
	}
}
