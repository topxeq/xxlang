// pkg/stdlib/mail_test.go
// Tests for mail module and its internal helper functions.
package stdlib

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// Module existence test
func TestMailModule_Exists(t *testing.T) {
	mod := Get("mail")
	if mod == nil {
		t.Fatal("mail module not found")
	}
	if mod.Name != "mail" {
		t.Errorf("expected module name 'mail', got %s", mod.Name)
	}
}

// NewClient basic tests
func TestMailModule_NewClient(t *testing.T) {
	mod := Get("mail")
	if mod == nil {
		t.Skip("mail module not found")
	}

	fn := mod.Exports["newClient"].(*objects.Builtin)

	// Wrong arg count
	result := fn.Fn(objects.NewString("host"), objects.NewInt(25))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong arg count")
	}
}

func TestMailModule_NewClientArgTypes(t *testing.T) {
	mod := Get("mail")
	if mod == nil {
		t.Skip("mail module not found")
	}

	fn := mod.Exports["newClient"].(*objects.Builtin)

	// Wrong types for host
	result := fn.Fn(objects.NewInt(0), objects.NewInt(587), objects.NewString("user"), objects.NewString("pass"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong host type")
	}
}

func TestMailModule_NewClientValid(t *testing.T) {
	mod := Get("mail")
	if mod == nil {
		t.Skip("mail module not found")
	}

	fn := mod.Exports["newClient"].(*objects.Builtin)

	result := fn.Fn(
		objects.NewString("smtp.example.com"),
		objects.NewInt(587),
		objects.NewString("user@example.com"),
		objects.NewString("password"),
	)

	client, ok := result.(*objects.MailClient)
	if !ok {
		t.Fatalf("expected MailClient, got %T", result)
	}
	if client.Host != "smtp.example.com" {
		t.Errorf("expected host smtp.example.com, got %s", client.Host)
	}
	if client.Port != 587 {
		t.Errorf("expected port 587, got %d", client.Port)
	}
	if client.User != "user@example.com" {
		t.Errorf("expected user user@example.com, got %s", client.User)
	}
}

func TestMailModule_NewClientWithOptions(t *testing.T) {
	mod := Get("mail")
	if mod == nil {
		t.Skip("mail module not found")
	}

	fn := mod.Exports["newClient"].(*objects.Builtin)

	result := fn.Fn(
		objects.NewString("smtp.example.com"),
		objects.NewInt(587),
		objects.NewString("user@example.com"),
		objects.NewString("password"),
		objects.NewString("-from=custom@example.com"),
		objects.NewString("-fromName=Custom Sender"),
	)

	client, ok := result.(*objects.MailClient)
	if !ok {
		t.Fatalf("expected MailClient, got %T", result)
	}
	if client.From != "custom@example.com" {
		t.Errorf("expected from custom@example.com, got %s", client.From)
	}
	if client.FromName != "Custom Sender" {
		t.Errorf("expected fromName 'Custom Sender', got %s", client.FromName)
	}
}

func TestMailModule_SendArgCount(t *testing.T) {
	mod := Get("mail")
	if mod == nil {
		t.Skip("mail module not found")
	}

	fn := mod.Exports["send"].(*objects.Builtin)

	// No args
	result := fn.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for no args")
	}
}

// Helper function tests (testing internal functions)
// These functions are defined in mail.go and are private, but we can test them
// within the same package.

func TestSplitMailEmails(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"alice@example.com", []string{"alice@example.com"}},
		{"alice@example.com, bob@example.org", []string{"alice@example.com", "bob@example.org"}},
		{"alice@example.com, , bob@example.org", []string{"alice@example.com", "bob@example.org"}},
		{"", []string{}},
		{" , ", []string{}},
		{"alice@example.com; bob@example.org", []string{"alice@example.com", "bob@example.org"}}, // semicolon separator
		{"  alice@example.com  ,  bob@example.org  ", []string{"alice@example.com", "bob@example.org"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitMailEmails(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitMailEmails(%q) returned %d items, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, r := range result {
				if r != tt.expected[i] {
					t.Errorf("splitMailEmails(%q)[%d] = %q, want %q", tt.input, i, r, tt.expected[i])
				}
			}
		})
	}
}

func TestFormatMailAddress(t *testing.T) {
	tests := []struct {
		email    string
		name     string
		expected string
	}{
		{"alice@example.com", "Alice", "Alice <alice@example.com>"},
		{"bob@example.org", "", "bob@example.org"},
		{"", "Anonymous", "Anonymous <>"}, // edge case: empty email
	}

	for _, tt := range tests {
		t.Run(tt.email+":"+tt.name, func(t *testing.T) {
			result := formatMailAddress(tt.email, tt.name)
			if result != tt.expected {
				t.Errorf("formatMailAddress(%q, %q) = %q, want %q", tt.email, tt.name, result, tt.expected)
			}
		})
	}
}

func TestParseMailPortFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"587", 587},
		{"25", 25},
		{"465", 465},
		{"", 0},         // empty yields 0
		{"abc", 0},      // non-digits yield 0
		{"12a34", 1234}, // accumulates all digits
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
