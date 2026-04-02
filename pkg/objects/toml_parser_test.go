// pkg/objects/toml_parser_test.go
// Tests for TOML parser
package objects

import (
	"testing"
)

func TestParseToml(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		validate  func(t *testing.T, result Object)
	}{
		{
			name:      "empty input",
			input:     "",
			wantError: false,
			validate: func(t *testing.T, result Object) {
				_, ok := result.(*TomlDocument)
				if !ok {
					t.Errorf("expected TomlDocument for empty input, got %T", result)
				}
			},
		},
		{
			name:      "simple key-value",
			input:     "name = \"John\"",
			wantError: false,
			validate: func(t *testing.T, result Object) {
				doc, ok := result.(*TomlDocument)
				if !ok {
					t.Errorf("expected TomlDocument, got %T", result)
					return
				}
				// Document should have some content
				if doc == nil {
					t.Errorf("expected non-nil document")
				}
			},
		},
		{
			name:      "simple table",
			input:     "[user]\nname = \"Jane\"",
			wantError: false,
			validate: func(t *testing.T, result Object) {
				_, ok := result.(*TomlDocument)
				if !ok {
					t.Errorf("expected TomlDocument, got %T", result)
				}
			},
		},
		{
			name:      "invalid syntax",
			input:     "name = \"unclosed string",
			wantError: true,
		},
		{
			name:      "array of integers",
			input:     "numbers = [1, 2, 3]",
			wantError: false,
		},
		{
			name:      "nested table",
			input:     "[database]\n  [database.server]\n    host = \"localhost\"",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseToml(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected non-nil result")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
