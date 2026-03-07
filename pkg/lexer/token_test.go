// pkg/lexer/token_test.go
package lexer

import "testing"

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		token    TokenType
		expected string
	}{
		{TokenInt, "INT"},
		{TokenString, "STRING"},
		{TokenIdent, "IDENT"},
		{TokenEOF, "EOF"},
	}

	for _, tt := range tests {
		if got := string(tt.token); got != tt.expected {
			t.Errorf("TokenType(%q) = %q, want %q", tt.token, got, tt.expected)
		}
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{Type: TokenInt, Literal: "42", Line: 1, Column: 1}
	expected := "INT(42) at 1:1"
	if got := tok.String(); got != expected {
		t.Errorf("Token.String() = %q, want %q", got, expected)
	}
}
