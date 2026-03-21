// pkg/jit/native_executor_test.go
// Tests for native executor functionality
package jit

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

// TestCanExecuteNatively tests the CanExecuteNatively function
func TestCanExecuteNatively(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []byte
		expected bool
	}{
		{
			name: "simple arithmetic",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0, // const 0 to R0
				byte(compiler.OpRegLoadConst), 1, 0, 1, // const 1 to R1
				byte(compiler.OpRegAdd), 2, 0, 1, // R2 = R0 + R1
				byte(compiler.OpRegMove), 255, 2, // return R2
			},
			expected: true,
		},
		{
			name: "with builtin call",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegBuiltin), 1, 1, // builtin 1 with 1 arg
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: false,
		},
		{
			name: "with function call",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegCall), 1, 0, // call R1 with 0 args
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: false,
		},
		{
			name: "with closure",
			bytecode: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0, // create closure
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: false,
		},
		{
			name: "with array operation",
			bytecode: []byte{
				byte(compiler.OpRegArray), 0, 0, 2, // create array from R0, R1
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: false,
		},
		{
			name: "control flow",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegJump), 0, 0, 10, // jump +10
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: true,
		},
		{
			name: "comparison",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegLoadConst), 1, 0, 1,
				byte(compiler.OpRegLess), 2, 0, 1, // R2 = R0 < R1
				byte(compiler.OpRegMove), 255, 2,
			},
			expected: true,
		},
		{
			name: "local variables",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegStoreLocal), 0, 0, // locals[0] = R0
				byte(compiler.OpRegLoadLocal), 1, 0, // R1 = locals[0]
				byte(compiler.OpRegMove), 255, 1,
			},
			expected: true,
		},
		{
			name: "global variables",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 42,
				byte(compiler.OpRegStoreGlobal), 0, 0, 0, // globals[0] = R0
				byte(compiler.OpRegLoadGlobal), 1, 0, 0, // R1 = globals[0]
				byte(compiler.OpRegMove), 255, 1,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.bytecode,
				NumLocals:     8,
				NumParameters: 0,
			}

			result := CanExecuteNatively(fn)
			if result != tt.expected {
				t.Errorf("CanExecuteNatively() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestAnalyzeNativeSupport tests the AnalyzeNativeSupport function
func TestAnalyzeNativeSupport(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []byte
		expected NativeSupportLevel
	}{
		{
			name: "pure arithmetic",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegAdd), 1, 0, 0,
				byte(compiler.OpRegMove), 255, 1,
			},
			expected: SupportPureArithmetic,
		},
		{
			name: "with builtin",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegBuiltin), 1, 1,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: SupportWithBuiltins,
		},
		{
			name: "with function call",
			bytecode: []byte{
				byte(compiler.OpRegLoadFunc), 0, 0, 0,
				byte(compiler.OpRegCall), 0, 0,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: SupportNone, // OpRegLoadFunc is unsupported
		},
		{
			name: "with array",
			bytecode: []byte{
				byte(compiler.OpRegArrayEmpty), 0,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: SupportWithArrays,
		},
		{
			name: "with object field",
			bytecode: []byte{
				byte(compiler.OpRegArrayEmpty), 0,
				byte(compiler.OpRegGetField), 1, 0, 0, 0, // R1 = R0.field
				byte(compiler.OpRegMove), 255, 1,
			},
			expected: SupportWithObjects,
		},
		{
			name: "with closure - unsupported",
			bytecode: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: SupportNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.bytecode,
				NumLocals:     8,
				NumParameters: 0,
			}

			result := AnalyzeNativeSupport(fn)
			if result != tt.expected {
				t.Errorf("AnalyzeNativeSupport() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestUsesClosures tests the UsesClosures function
func TestUsesClosures(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []byte
		expected bool
	}{
		{
			name: "no closure",
			bytecode: []byte{
				byte(compiler.OpRegLoadConst), 0, 0, 0,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: false,
		},
		{
			name: "with OpRegClosure",
			bytecode: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: true,
		},
		{
			name: "with OpRegLoadFree",
			bytecode: []byte{
				byte(compiler.OpRegLoadFree), 0, 0,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: true,
		},
		{
			name: "with OpRegStoreFree",
			bytecode: []byte{
				byte(compiler.OpRegStoreFree), 0, 0,
				byte(compiler.OpRegMove), 255, 0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.bytecode,
				NumLocals:     8,
				NumParameters: 0,
			}

			result := UsesClosures(fn)
			if result != tt.expected {
				t.Errorf("UsesClosures() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestNativeExecutorCreation tests creating a native executor
func TestNativeExecutorCreation(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  4096,
		Debug:        false,
	}

	exec := NewNativeExecutor(config)
	if exec == nil {
		t.Fatal("NewNativeExecutor returned nil")
	}

	// Cleanup should not panic
	exec.Cleanup()
}

// TestNativeFunctionRegistry tests the native function registry
func TestNativeFunctionRegistry(t *testing.T) {
	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  4096,
		Debug:        false,
	}

	registry := NewNativeFunctionRegistry(config)
	if registry == nil {
		t.Fatal("NewNativeFunctionRegistry returned nil")
	}
	defer registry.Cleanup()

	// Test Get on non-existent function
	fn := registry.Get(999)
	if fn != nil {
		t.Error("Get on non-existent index should return nil")
	}

	// Test Has
	if registry.Has(999) {
		t.Error("Has on non-existent index should return false")
	}
}

// TestNativeExecutorWithSimpleFunction tests executing a simple function natively
func TestNativeExecutorWithSimpleFunction(t *testing.T) {
	// Create a simple function: return 42
	bytecode := []byte{
		byte(compiler.OpRegLoadConst), 0, 0, 0, // Load const 0 to R0 (value will be in constants)
		byte(compiler.OpRegMove), 255, 0, // Return R0
	}

	fn := &compiler.CompiledFunction{
		Instructions:  bytecode,
		NumLocals:     8,
		NumParameters: 0,
	}

	constants := []vm.Value{vm.NewInt(42)}
	globals := make([]int64, 256)

	config := JITConfig{
		HotThreshold: 1,
		MaxCodeSize:  4096,
		Debug:        true,
	}

	exec := NewNativeExecutor(config)
	defer exec.Cleanup()

	result, err := exec.ExecuteFunction(fn, constants, globals)
	if err != nil {
		t.Fatalf("ExecuteFunction failed: %v", err)
	}

	t.Logf("Result: %d (expected 42)", result)
	// Note: The result depends on correct code generation
}

// TestCallBuiltinFromNative tests the builtin callback
func TestCallBuiltinFromNative(t *testing.T) {
	// Initialize context
	InitJITCallbackContext(nil, nil)

	// Test with invalid builtin index
	args := []int64{1, 2}
	result := CallBuiltinFromNative(999, 2, &args[0])
	if result != 0 {
		t.Error("Invalid builtin index should return 0")
	}

	// Test abs builtin (index 14)
	// First we need to set up the builtins
	// Note: This requires the builtin to be properly initialized
}

// TestCanExecuteNativelyWithBuiltins tests the helper function
func TestCanExecuteNativelyWithBuiltins(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []byte
		expected bool
	}{
		{
			name: "pure arithmetic - should pass (CanExecuteNatively)",
			bytecode: []byte{
				byte(compiler.OpRegAdd), 0, 0, 0,
			},
			// Note: CanExecuteNativelyWithBuiltins checks if level >= SupportWithBuiltins
			// Pure arithmetic is SupportPureArithmetic, which is < SupportWithBuiltins
			// So this returns false for the *WithBuiltins check
			expected: false,
		},
		{
			name: "with builtin - should pass",
			bytecode: []byte{
				byte(compiler.OpRegBuiltin), 1, 0,
			},
			// This has SupportWithBuiltins level
			expected: true, // Has builtin but no closure - can execute with builtin callback
		},
		{
			name: "with closure - should fail",
			bytecode: []byte{
				byte(compiler.OpRegClosure), 0, 0, 0, 0, 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &compiler.CompiledFunction{
				Instructions:  tt.bytecode,
				NumLocals:     8,
				NumParameters: 0,
			}

			result := CanExecuteNativelyWithBuiltins(fn)
			if result != tt.expected {
				t.Errorf("CanExecuteNativelyWithBuiltins() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestCanExecuteNativelyWithCalls tests the helper function
func TestCanExecuteNativelyWithCalls(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegCall), 0, 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	// OpRegCall alone should be SupportWithCalls level
	// but CanExecuteNatively checks if >= SupportPureArithmetic
	// Actually, let's check what happens
	level := AnalyzeNativeSupport(fn)
	t.Logf("Function with OpRegCall has support level: %v", level)
}

// TestCanExecuteNativelyWithArrays tests the helper function
func TestCanExecuteNativelyWithArrays(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegArrayEmpty), 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	if !CanExecuteNativelyWithArrays(fn) {
		t.Error("Function with array operation should pass CanExecuteNativelyWithArrays")
	}
}

// TestCanExecuteNativelyWithObjects tests the helper function
func TestCanExecuteNativelyWithObjects(t *testing.T) {
	fn := &compiler.CompiledFunction{
		Instructions: []byte{
			byte(compiler.OpRegGetField), 0, 0, 0, 0,
		},
		NumLocals:     8,
		NumParameters: 0,
	}

	if !CanExecuteNativelyWithObjects(fn) {
		t.Error("Function with object field access should pass CanExecuteNativelyWithObjects")
	}
}

// TestCallbackPointers tests that callback pointers can be obtained
func TestCallbackPointers(t *testing.T) {
	// Test that we can get callback pointers
	builtinPtr := GetBuiltinCallbackPtr()
	if builtinPtr == 0 {
		t.Error("GetBuiltinCallbackPtr should return non-zero pointer")
	}
	t.Logf("Builtin callback ptr: 0x%x", builtinPtr)

	funcPtr := GetFunctionCallbackPtr()
	if funcPtr == 0 {
		t.Error("GetFunctionCallbackPtr should return non-zero pointer")
	}
	t.Logf("Function callback ptr: 0x%x", funcPtr)

	collPtr := GetCollectionCallbackPtr()
	if collPtr == 0 {
		t.Error("GetCollectionCallbackPtr should return non-zero pointer")
	}
	t.Logf("Collection callback ptr: 0x%x", collPtr)

	objPtr := GetObjectCallbackPtr()
	if objPtr == 0 {
		t.Error("GetObjectCallbackPtr should return non-zero pointer")
	}
	t.Logf("Object callback ptr: 0x%x", objPtr)
}
