// pkg/stdlib/db_extra2_test.go
// Additional tests for db module to increase coverage, focusing on dbValueToObj.
package stdlib

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestDBValueToObj_AllTypes tests dbValueToObj conversion across various Go types.
func TestDBValueToObj_AllTypes(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedType  string      // expected object type name (e.g., "Int", "Float", "String", "Bool", "Null")
		expectedValue interface{} // optional expected value (if nil, just check type)
	}{
		{"nil", nil, "Null", nil},
		{"int", 42, "Int", int64(42)},
		{"int64", int64(9223372036854775807), "Int", int64(9223372036854775807)},
		{"int32", int32(-1000), "Int", int64(-1000)},
		{"int16", int16(-123), "Int", int64(-123)},
		{"int8", int8(-128), "Int", int64(-128)},
		{"uint", uint(4294967295), "Int", int64(4294967295)},
		{"uint64", uint64(18446744073709551615), "Int", nil}, // value may overflow int64; check type only
		{"uint32", uint32(4294967295), "Int", int64(4294967295)},
		{"uint16", uint16(65535), "Int", int64(65535)},
		{"uint8", uint8(255), "Int", int64(255)},
		{"float32", float32(3.14159), "Float", float64(3.14159)},
		{"float64", float64(2.71828), "Float", float64(2.71828)},
		{"float64 NaN", math.NaN(), "Float", nil},   // NaN cannot be compared directly
		{"float64 +Inf", math.Inf(1), "Float", nil}, // Inf cannot be compared directly
		{"float64 -Inf", math.Inf(-1), "Float", nil},
		{"string", "hello world", "String", "hello world"},
		{"bytes", []byte("ABC"), "String", "ABC"}, // []byte converts to String
		{"bool true", true, "Bool", true},
		{"bool false", false, "Bool", false},
		// default case: custom type that doesn't match any; fmt.Sprintf("%v", v)
		{"time.Time", time.Now(), "String", nil}, // time.Time becomes string via fmt.Sprintf
		{"struct", struct{ X int }{X: 5}, "String", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := dbValueToObj(tt.input)
			if obj == nil {
				t.Fatal("dbValueToObj returned nil")
			}
			typ := reflect.TypeOf(obj)
			if typ == nil {
				t.Fatal("reflect.TypeOf returned nil")
			}
			typeName := typ.String()
			// The type string is like "*stdlib.Int", "*stdlib.String", etc.
			// We'll check that it contains the expected type name.
			if !strings.Contains(typeName, tt.expectedType) {
				t.Fatalf("expected type containing %q, got %s", tt.expectedType, typeName)
			}
			// If expectedValue is provided and not nil, check equality.
			if tt.expectedValue != nil {
				switch tt.expectedType {
				case "Int":
					if intObj, ok := obj.(*objects.Int); ok {
						if intObj.Value != tt.expectedValue.(int64) {
							t.Errorf("expected value %v, got %v", tt.expectedValue, intObj.Value)
						}
					} else {
						t.Fatalf("expected *objects.Int, got %T", obj)
					}
				case "Float":
					if floatObj, ok := obj.(*objects.Float); ok {
						if tt.expectedValue != nil {
							// For floats, compare with tolerance
							gotVal := floatObj.Value
							wantVal := tt.expectedValue.(float64)
							if diff := math.Abs(gotVal - wantVal); diff > 1e-6 {
								t.Errorf("expected value %v, got %v", wantVal, gotVal)
							}
						}
					} else {
						t.Fatalf("expected *objects.Float, got %T", obj)
					}
				case "String":
					if strObj, ok := obj.(*objects.String); ok {
						if strObj.Value != tt.expectedValue.(string) {
							t.Errorf("expected value %q, got %q", tt.expectedValue, strObj.Value)
						}
					} else {
						t.Fatalf("expected *objects.String, got %T", obj)
					}
				case "Bool":
					if boolObj, ok := obj.(*objects.Bool); ok {
						if boolObj.Value != tt.expectedValue.(bool) {
							t.Errorf("expected value %v, got %v", tt.expectedValue, boolObj.Value)
						}
					} else {
						t.Fatalf("expected *objects.Bool, got %T", obj)
					}
				}
			}
		})
	}
}

// TestDBValueToObj_DefaultFallback tests that unknown types are converted to string via fmt.Sprintf.
func TestDBValueToObj_DefaultFallback(t *testing.T) {
	// Custom struct without String() method
	type customStruct struct{ ID int }
	cs := customStruct{ID: 123}
	obj := dbValueToObj(cs)
	strObj, ok := obj.(*objects.String)
	if !ok {
		t.Fatalf("expected *objects.String for custom struct, got %T", obj)
	}
	// The string should be the default fmt.Sprintf output: "{123}" or similar.
	// We'll just check that it's non-empty and contains "123".
	if !strings.Contains(strObj.Value, "123") {
		t.Errorf("expected string representation to contain 123, got %q", strObj.Value)
	}
}
