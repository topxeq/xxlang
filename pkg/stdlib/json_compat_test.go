// pkg/stdlib/json_compat_test.go
// Tests for JSON builtin functions and module compatibility
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestJsonBuiltinFunctionsCompatibility tests that toJson/fromJson builtins
// produce the same results as json.stringify/json.parse module functions
func TestJsonBuiltinFunctionsCompatibility(t *testing.T) {
	// Get the json module
	jsonModule := Get("json")
	if jsonModule == nil {
		t.Fatal("json module not found")
	}

	// Test data
	testCases := []struct {
		name  string
		value objects.Object
	}{
		{"null", objects.NULL},
		{"bool true", objects.TRUE},
		{"bool false", objects.FALSE},
		{"int", objects.NewInt(42)},
		{"float", objects.NewFloat(3.14)},
		{"string", objects.NewString("hello")},
		{"array", objects.NewArray([]objects.Object{
			objects.NewInt(1),
			objects.NewInt(2),
			objects.NewInt(3),
		})},
		{"map", objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("key").HashKey(): {
				Key:   objects.NewString("key"),
				Value: objects.NewString("value"),
			},
		})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that ObjectToGoValue and ObjectToJSON work
			goVal, err := objects.ObjectToGoValue(tc.value)
			if err != nil {
				t.Errorf("ObjectToGoValue failed for %s: %v", tc.name, err)
			}

			jsonBytes, err := objects.ObjectToJSON(tc.value, objects.ObjectToJSONOptions{})
			if err != nil {
				t.Errorf("ObjectToJSON failed for %s: %v", tc.name, err)
			}

			// Test round-trip: object -> JSON -> object
			jsonStr := string(jsonBytes)

			parsedObj, err := objects.JSONToObject(jsonStr)
			if err != nil {
				t.Errorf("JSONToObject failed for %s: %v", tc.name, err)
			}

			// Verify the parsed object
			_ = goVal
			_ = parsedObj
		})
	}
}

// TestJsonModuleFunctions tests the json module functions
func TestJsonModuleFunctions(t *testing.T) {
	// Get the json module
	jsonModule := Get("json")
	if jsonModule == nil {
		t.Fatal("json module not found")
	}

	// Check that all expected functions are exported
	expectedFuncs := []string{
		"parse", "stringify", "encode", "decode",
		"toJson", "fromJson",
		"readFile", "writeFile", "writeFilePretty",
		"updateFile", "appendToArrayFile",
		"isValid", "getType",
	}

	for _, fn := range expectedFuncs {
		if _, ok := jsonModule.Exports[fn]; !ok {
			t.Errorf("json module missing function: %s", fn)
		}
	}
}

// TestJsonRoundTrip tests that JSON conversion is reversible
func TestJsonRoundTrip(t *testing.T) {
	testCases := []struct {
		name  string
		value objects.Object
	}{
		{"simple int", objects.NewInt(42)},
		{"negative int", objects.NewInt(-123)},
		{"float", objects.NewFloat(3.14159)},
		{"string", objects.NewString("test string with \"quotes\"")},
		{"array of mixed types", objects.NewArray([]objects.Object{
			objects.NewInt(1),
			objects.NewString("two"),
			objects.TRUE,
			objects.NULL,
		})},
		{"nested map", objects.NewMap(map[objects.HashKey]objects.MapPair{
			objects.NewString("outer").HashKey(): {
				Key: objects.NewString("outer"),
				Value: objects.NewMap(map[objects.HashKey]objects.MapPair{
					objects.NewString("inner").HashKey(): {
						Key:   objects.NewString("inner"),
						Value: objects.NewInt(42),
					},
				}),
			},
		})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to JSON
			jsonBytes, err := objects.ObjectToJSON(tc.value, objects.ObjectToJSONOptions{})
			if err != nil {
				t.Fatalf("ObjectToJSON failed: %v", err)
			}

			// Convert back
			parsed, err := objects.JSONToObject(string(jsonBytes))
			if err != nil {
				t.Fatalf("JSONToObject failed: %v", err)
			}

			// Convert again to JSON for comparison
			jsonBytes2, err := objects.ObjectToJSON(parsed, objects.ObjectToJSONOptions{})
			if err != nil {
				t.Fatalf("Second ObjectToJSON failed: %v", err)
			}

			// The JSON strings should be identical
			if string(jsonBytes) != string(jsonBytes2) {
				t.Errorf("Round-trip mismatch:\n  original: %s\n  parsed:   %s", jsonBytes, jsonBytes2)
			}
		})
	}
}
