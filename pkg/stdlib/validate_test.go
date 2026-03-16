// pkg/stdlib/validate_test.go
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// Helper function to call validate functions
func callValidateFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("validate")
	if mod == nil {
		panic("validate module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

// ============================================
// Tests for validate module
// ============================================

func TestValidateIsEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isEmail", String(tt.input))
		boolResult, ok := result.(*objects.Bool)
		if !ok {
			t.Errorf("isEmail(%q) did not return bool", tt.input)
			continue
		}
		if boolResult.Value != tt.expected {
			t.Errorf("isEmail(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}

	// Test error cases
	result := callValidateFunc("isEmail")
	if !isError(result) {
		t.Error("isEmail() should return error with no arguments")
	}

	result = callValidateFunc("isEmail", Int(42))
	if !isError(result) {
		t.Error("isEmail(int) should return error")
	}
}

func TestValidateIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"http://example.com", true},
		{"https://domain.org/path", true},
		{"ftp://files.server.com", true},
		{"ws://websocket.example.com", true},
		{"wss://secure.websocket.com", true},
		{"invalid-url", false},
		{"htp://wrong-scheme.com", false},
		{"http://nodot", false},
		{"", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isURL", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isURL(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateMatches(t *testing.T) {
	// Test matching pattern
	result := callValidateFunc("matches", String("hello123"), String(`^[a-z]+[0-9]+$`))
	boolResult := result.(*objects.Bool)
	if !boolResult.Value {
		t.Error("matches should return true for matching pattern")
	}

	// Test non-matching pattern
	result = callValidateFunc("matches", String("HELLO"), String(`^[a-z]+$`))
	boolResult = result.(*objects.Bool)
	if boolResult.Value {
		t.Error("matches should return false for non-matching pattern")
	}

	// Test invalid regex
	result = callValidateFunc("matches", String("test"), String(`[invalid`))
	if !isError(result) {
		t.Error("matches should return error for invalid regex")
	}

	// Test error cases
	result = callValidateFunc("matches", String("test"))
	if !isError(result) {
		t.Error("matches() should return error with insufficient arguments")
	}
}

func TestValidateLengthRange(t *testing.T) {
	tests := []struct {
		input    string
		min      int64
		max      int64
		expected bool
	}{
		{"hello", 1, 10, true},
		{"hi", 3, 10, false},
		{"this is a long string", 1, 10, false},
		{"", 1, 10, false},
		{"exact", 5, 5, true},
	}

	for _, tt := range tests {
		result := callValidateFunc("lengthRange", String(tt.input), Int(tt.min), Int(tt.max))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("lengthRange(%q, %d, %d) = %v, expected %v", tt.input, tt.min, tt.max, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"  ", false},
		{"", false},
		{"  value  ", true},
	}

	for _, tt := range tests {
		result := callValidateFunc("required", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("required(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateInArray(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{
		String("a"),
		String("b"),
		String("c"),
	}}

	// Test value in array
	result := callValidateFunc("inArray", String("b"), arr)
	boolResult := result.(*objects.Bool)
	if !boolResult.Value {
		t.Error("inArray should return true for value in array")
	}

	// Test value not in array
	result = callValidateFunc("inArray", String("d"), arr)
	boolResult = result.(*objects.Bool)
	if boolResult.Value {
		t.Error("inArray should return false for value not in array")
	}
}

func TestValidateNotInArray(t *testing.T) {
	arr := &objects.Array{Elements: []objects.Object{
		String("a"),
		String("b"),
	}}

	// Test value not in array
	result := callValidateFunc("notInArray", String("c"), arr)
	boolResult := result.(*objects.Bool)
	if !boolResult.Value {
		t.Error("notInArray should return true for value not in array")
	}

	// Test value in array
	result = callValidateFunc("notInArray", String("a"), arr)
	boolResult = result.(*objects.Bool)
	if boolResult.Value {
		t.Error("notInArray should return false for value in array")
	}
}

func TestValidateInRange(t *testing.T) {
	tests := []struct {
		value    interface{}
		min      interface{}
		max      interface{}
		expected bool
	}{
		{int64(5), int64(1), int64(10), true},
		{int64(0), int64(1), int64(10), false},
		{int64(15), int64(1), int64(10), false},
		{float64(5.5), int64(1), int64(10), true},
		{int64(5), float64(1.0), float64(10.0), true},
	}

	for _, tt := range tests {
		var val, minObj, maxObj objects.Object
		switch v := tt.value.(type) {
		case int64:
			val = Int(v)
		case float64:
			val = &objects.Float{Value: v}
		}
		switch v := tt.min.(type) {
		case int64:
			minObj = Int(v)
		case float64:
			minObj = &objects.Float{Value: v}
		}
		switch v := tt.max.(type) {
		case int64:
			maxObj = Int(v)
		case float64:
			maxObj = &objects.Float{Value: v}
		}

		result := callValidateFunc("inRange", val, minObj, maxObj)
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("inRange(%v, %v, %v) = %v, expected %v", tt.value, tt.min, tt.max, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`{"key": "value"}`, true},
		{`[1, 2, 3]`, true},
		{`not json`, false},
		{`{"incomplete"`, false},
		{``, false},
		{`  {"key": "value"}  `, true},
	}

	for _, tt := range tests {
		result := callValidateFunc("isJSON", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isJSON(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsAlphanumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abc123", true},
		{"ABC", true},
		{"123", true},
		{"abc-123", false},
		{"", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isAlphanumeric", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isAlphanumeric(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsAlpha(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"WORLD", true},
		{"HelloWorld", true},
		{"hello123", false},
		{"", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isAlpha", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isAlpha(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"3.14", true},
		{"-42", true},
		{"abc", false},
		{"", false},
		{"12.34.56", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isNumeric", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isNumeric(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsInteger(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"-42", true},
		{"0", true},
		{"3.14", false},
		{"abc", false},
		{"", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isInteger", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isInteger(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsHexColor(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"#fff", true},
		{"#ffffff", true},
		{"#ABC123", true},
		{"fff", false},
		{"#ff", false},
		{"#fffffff", false},
		{"#ggg", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isHexColor", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isHexColor(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsUUID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"not-a-uuid", false},
		{"550e8400e29b41d4a716446655440000", false}, // Missing dashes
		{"", false},
		{"550e8400-e29b-41d4-a716-44665544000g", false}, // Invalid char
	}

	for _, tt := range tests {
		result := callValidateFunc("isUUID", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isUUID(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsIPv4(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"256.1.1.1", false},
		{"1.2.3", false},
		{"1.2.3.4.5", false},
		{"01.02.03.04", false}, // Leading zeros
		{"abc.def.ghi.jkl", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isIPv4", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isIPv4(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1234567", true},
		{"+1-555-123-4567", true},
		{"(555) 123-4567", true},
		{"123", false},
		{"1234567890123456", false}, // Too long
		{"abc-defg", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isPhone", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isPhone(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsDate(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"2024-01-15", true},
		{"1999-12-31", true},
		{"2024-13-01", false}, // Invalid month
		{"2024-01-32", false}, // Invalid day
		{"01-15-2024", false}, // Wrong format
		{"2024/01/15", false}, // Wrong separator
		{"", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isDate", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isDate(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateIsTime(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"12:30", true},
		{"23:59", true},
		{"00:00", true},
		{"12:30:45", true},
		{"24:00", false}, // Invalid hour
		{"12:60", false}, // Invalid minute
		{"12:30:60", false}, // Invalid second
		{"", false},
	}

	for _, tt := range tests {
		result := callValidateFunc("isTime", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isTime(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

func TestValidateStartsWith(t *testing.T) {
	result := callValidateFunc("startsWith", String("hello world"), String("hello"))
	boolResult := result.(*objects.Bool)
	if !boolResult.Value {
		t.Error("startsWith should return true for matching prefix")
	}

	result = callValidateFunc("startsWith", String("hello world"), String("world"))
	boolResult = result.(*objects.Bool)
	if boolResult.Value {
		t.Error("startsWith should return false for non-matching prefix")
	}
}

func TestValidateEndsWith(t *testing.T) {
	result := callValidateFunc("endsWith", String("hello world"), String("world"))
	boolResult := result.(*objects.Bool)
	if !boolResult.Value {
		t.Error("endsWith should return true for matching suffix")
	}

	result = callValidateFunc("endsWith", String("hello world"), String("hello"))
	boolResult = result.(*objects.Bool)
	if boolResult.Value {
		t.Error("endsWith should return false for non-matching suffix")
	}
}

func TestValidateContains(t *testing.T) {
	result := callValidateFunc("contains", String("hello world"), String("lo wo"))
	boolResult := result.(*objects.Bool)
	if !boolResult.Value {
		t.Error("contains should return true for matching substring")
	}

	result = callValidateFunc("contains", String("hello world"), String("xyz"))
	boolResult = result.(*objects.Bool)
	if boolResult.Value {
		t.Error("contains should return false for non-matching substring")
	}
}

func TestValidateIsCreditCard(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"4532015112830366", true},  // Valid Visa
		{"5425233430109903", true},  // Valid MasterCard
		{"374245455400126", true},   // Valid Amex
		{"1234567812345678", false}, // Invalid (fails Luhn)
		{"1234", false},             // Too short
		{"abcd-efgh-ijkl-mnop", false}, // Non-numeric
	}

	for _, tt := range tests {
		result := callValidateFunc("isCreditCard", String(tt.input))
		boolResult := result.(*objects.Bool)
		if boolResult.Value != tt.expected {
			t.Errorf("isCreditCard(%q) = %v, expected %v", tt.input, boolResult.Value, tt.expected)
		}
	}
}

// Helper function to check if result is an error
func isError(obj objects.Object) bool {
	_, ok := obj.(*objects.Error)
	return ok
}
