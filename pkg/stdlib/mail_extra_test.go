// pkg/stdlib/mail_extra_test.go
// Additional tests for mail module internal helpers
package stdlib

import (
	"testing"
)

// TestSplitMailEmails_Extra tests splitMailEmails function with additional cases.
func TestSplitMailEmails_Extra(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a@example.com", []string{"a@example.com"}},
		{"a@example.com, b@example.com", []string{"a@example.com", "b@example.com"}},
		{"a@example.com,b@example.com", []string{"a@example.com", "b@example.com"}},
		{"  a@example.com , b@example.com  ", []string{"a@example.com", "b@example.com"}},
		{`"Alice <alice@example.com>"`, []string{`"Alice <alice@example.com>"`}}, // as is, since parsing may not strip quotes
		{"", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitMailEmails(tt.input)
			// The function may trim spaces but not resolve quoted names
			if len(result) != len(tt.expected) {
				t.Fatalf("splitMailEmails(%q) = %v, want %v", tt.input, result, tt.expected)
			}
			for i, r := range result {
				if r != tt.expected[i] {
					t.Errorf("result[%d] = %q, want %q", i, r, tt.expected[i])
				}
			}
		})
	}
}

// TestFormatMailAddress_Extra tests formatMailAddress function with additional cases.
func TestFormatMailAddress_Extra(t *testing.T) {
	tests := []struct {
		email, name string
		expected    string
	}{
		{"user@example.com", "", "user@example.com"},
		{"user@example.com", "Alice", "Alice <user@example.com>"},
		{"user@example.com", "Alice Smith", "Alice Smith <user@example.com>"},
		{"user@example.com", "Bob & Alice", "Bob & Alice <user@example.com>"},
	}
	for _, tt := range tests {
		t.Run(tt.email+"/"+tt.name, func(t *testing.T) {
			result := formatMailAddress(tt.email, tt.name)
			if result != tt.expected {
				t.Errorf("formatMailAddress(%q, %q) = %q, want %q", tt.email, tt.name, result, tt.expected)
			}
		})
	}
}

// TestParseMailPortFromString_Extra tests parseMailPortFromString function with additional cases.
func TestParseMailPortFromString_Extra(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0}, // no digits -> 0
		{"25", 25},
		{"587", 587},
		{"465", 465},
		{"smtp", 0},       // no digits -> 0
		{"SMTP", 0},       // no digits -> 0
		{"submission", 0}, // no digits -> 0
		{"25tcp", 25},     // extracts leading digits
		{"tcp25", 25},     // extracts trailing digits
		{"587smtp", 587},  // digits followed by text
		{"smtp587", 587},  // text followed by digits
		{"12a34", 1234},   // already in original, but keep
		{"8080", 8080},
		{"0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseMailPortFromString(tt.input)
			if result != tt.expected {
				t.Errorf("parseMailPortFromString(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
