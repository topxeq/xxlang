// pkg/objects/toml_lexer.go
// TOML lexer - pure Go implementation
package objects

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType represents a token type
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenComment
	TokenKey
	TokenEquals
	TokenString
	TokenMultiLineString
	TokenLiteralString
	TokenMultiLineLiteralString
	TokenInteger
	TokenFloat
	TokenBoolean
	TokenDatetime
	TokenDate
	TokenTime
	TokenLBracket       // [
	TokenRBracket       // ]
	TokenDoubleLBracket // [[
	TokenDoubleRBracket // ]]
	TokenLBrace         // {
	TokenRBrace         // }
	TokenLParen         // (
	TokenRParen         // )
	TokenComma
	TokenDot
	TokenNewline
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "ERROR"
	case TokenComment:
		return "COMMENT"
	case TokenKey:
		return "KEY"
	case TokenEquals:
		return "EQUALS"
	case TokenString:
		return "STRING"
	case TokenMultiLineString:
		return "MULTILINE_STRING"
	case TokenLiteralString:
		return "LITERAL_STRING"
	case TokenMultiLineLiteralString:
		return "MULTILINE_LITERAL_STRING"
	case TokenInteger:
		return "INTEGER"
	case TokenFloat:
		return "FLOAT"
	case TokenBoolean:
		return "BOOLEAN"
	case TokenDatetime:
		return "DATETIME"
	case TokenDate:
		return "DATE"
	case TokenTime:
		return "TIME"
	case TokenLBracket:
		return "LBRACKET"
	case TokenRBracket:
		return "RBRACKET"
	case TokenDoubleLBracket:
		return "DOUBLE_LBRACKET"
	case TokenDoubleRBracket:
		return "DOUBLE_RBRACKET"
	case TokenLBrace:
		return "LBRACE"
	case TokenRBrace:
		return "RBRACE"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenComma:
		return "COMMA"
	case TokenDot:
		return "DOT"
	case TokenNewline:
		return "NEWLINE"
	default:
		return "UNKNOWN"
	}
}

// Token represents a scanned token
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// Lexer holds the lexer state
type Lexer struct {
	input     string
	pos       int
	line      int
	col       int
	lastToken Token
}

// NewLexer creates a new lexer
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		line:  1,
		col:   1,
	}
}

// NextToken returns the next token
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		token := Token{Type: TokenEOF, Line: l.line, Col: l.col}
		l.lastToken = token
		return token
	}

	ch := l.peek()

	// Check for comments
	if ch == '#' {
		token := l.scanComment()
		l.lastToken = token
		return token
	}

	// Check for newlines
	if ch == '\n' {
		l.advance()
		token := Token{Type: TokenNewline, Line: l.line - 1, Col: 0}
		l.lastToken = token
		return token
	}

	// Check for strings
	if ch == '"' {
		token := l.scanString()
		l.lastToken = token
		return token
	}
	if ch == '\'' {
		token := l.scanLiteralString()
		l.lastToken = token
		return token
	}

	// Check for brackets
	// We always return single brackets, parser will combine [[ and ]] when needed
	if ch == '[' {
		l.advance()
		token := Token{Type: TokenLBracket, Value: "[", Line: l.line, Col: l.col - 1}
		l.lastToken = token
		return token
	}

	if ch == ']' {
		l.advance()
		token := Token{Type: TokenRBracket, Value: "]", Line: l.line, Col: l.col - 1}
		l.lastToken = token
		return token
	}

	// Check for braces
	if ch == '{' {
		l.advance()
		token := Token{Type: TokenLBrace, Value: "{", Line: l.line, Col: l.col - 1}
		l.lastToken = token
		return token
	}

	if ch == '}' {
		l.advance()
		token := Token{Type: TokenRBrace, Value: "}", Line: l.line, Col: l.col - 1}
		l.lastToken = token
		return token
	}

	// Check for punctuation
	if ch == '=' {
		l.advance()
		token := Token{Type: TokenEquals, Value: "=", Line: l.line, Col: l.col - 1}
		l.lastToken = token
		return token
	}

	if ch == '.' {
		l.advance()
		token := Token{Type: TokenDot, Value: ".", Line: l.line, Col: l.col - 1}
		l.lastToken = token
		return token
	}

	if ch == ',' {
		l.advance()
		token := Token{Type: TokenComma, Value: ",", Line: l.line, Col: l.col - 1}
		l.lastToken = token
		return token
	}

	// Check for boolean, numbers, datetime
	if unicode.IsDigit(rune(ch)) || ch == '-' || ch == '+' {
		token := l.scanNumberOrDatetime()
		l.lastToken = token
		return token
	}

	// Check for boolean or special float values (inf, nan)
	if ch == 't' || ch == 'f' || ch == 'i' || ch == 'n' {
		token := l.scanBooleanOrSpecial()
		l.lastToken = token
		return token
	}

	// Otherwise, scan as key
	token := l.scanKey()
	l.lastToken = token
	return token
}

// AllTokens returns all tokens
func (l *Lexer) AllTokens() []Token {
	var tokens []Token
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.Type == TokenEOF || token.Type == TokenError {
			break
		}
	}
	return tokens
}

// Helper methods

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) scanComment() Token {
	start := l.pos
	line := l.line
	col := l.col

	for l.pos < len(l.input) && l.peek() != '\n' {
		l.advance()
	}

	return Token{
		Type:  TokenComment,
		Value: l.input[start:l.pos],
		Line:  line,
		Col:   col,
	}
}

func (l *Lexer) scanString() Token {
	line := l.line
	col := l.col
	l.advance() // Skip opening "

	// Check for multi-line string
	if l.peek() == '"' && l.peekNext() == '"' {
		l.advance()
		l.advance()
		return l.scanMultiLineString(line, col)
	}

	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == '"' {
			l.advance()
			return Token{Type: TokenString, Value: sb.String(), Line: line, Col: col}
		}
		if ch == '\\' {
			l.advance()
			escaped := l.scanEscape()
			sb.WriteRune(escaped)
		} else {
			sb.WriteByte(ch)
			l.advance()
		}
	}

	return Token{Type: TokenError, Value: "unterminated string", Line: line, Col: col}
}

func (l *Lexer) scanMultiLineString(line, col int) Token {
	var sb strings.Builder

	// Skip initial newline if present
	if l.peek() == '\n' {
		l.advance()
	}

	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == '"' && l.peekNext() == '"' {
			l.advance()
			l.advance()
			if l.peek() == '"' {
				l.advance()
				return Token{Type: TokenMultiLineString, Value: sb.String(), Line: line, Col: col}
			}
			sb.WriteString(`""`)
			continue
		}
		if ch == '\\' {
			l.advance()
			escaped := l.scanEscape()
			sb.WriteRune(escaped)
		} else {
			sb.WriteByte(ch)
			l.advance()
		}
	}

	return Token{Type: TokenError, Value: "unterminated multi-line string", Line: line, Col: col}
}

func (l *Lexer) scanLiteralString() Token {
	line := l.line
	col := l.col
	l.advance() // Skip opening '

	// Check for multi-line literal string
	if l.peek() == '\'' && l.peekNext() == '\'' {
		l.advance()
		l.advance()
		return l.scanMultiLineLiteralString(line, col)
	}

	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == '\'' {
			l.advance()
			return Token{Type: TokenLiteralString, Value: sb.String(), Line: line, Col: col}
		}
		sb.WriteByte(ch)
		l.advance()
	}

	return Token{Type: TokenError, Value: "unterminated literal string", Line: line, Col: col}
}

func (l *Lexer) scanMultiLineLiteralString(line, col int) Token {
	var sb strings.Builder

	// Skip initial newline if present
	if l.peek() == '\n' {
		l.advance()
	}

	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == '\'' && l.peekNext() == '\'' {
			l.advance()
			l.advance()
			if l.peek() == '\'' {
				l.advance()
				return Token{Type: TokenMultiLineLiteralString, Value: sb.String(), Line: line, Col: col}
			}
			sb.WriteString(`''`)
			continue
		}
		sb.WriteByte(ch)
		l.advance()
	}

	return Token{Type: TokenError, Value: "unterminated multi-line literal string", Line: line, Col: col}
}

func (l *Lexer) scanEscape() rune {
	ch := l.peek()
	l.advance()

	switch ch {
	case 'b':
		return '\b'
	case 't':
		return '\t'
	case 'n':
		return '\n'
	case 'f':
		return '\f'
	case 'r':
		return '\r'
	case '"':
		return '"'
	case '\\':
		return '\\'
	case 'u':
		// Unicode escape \uXXXX
		return l.scanUnicodeEscape(4)
	case 'U':
		// Unicode escape \UXXXXXXXX
		return l.scanUnicodeEscape(8)
	default:
		return rune(ch)
	}
}

func (l *Lexer) scanUnicodeEscape(length int) rune {
	var r rune
	for i := 0; i < length && l.pos < len(l.input); i++ {
		ch := l.peek()
		l.advance()
		r = r << 4
		if ch >= '0' && ch <= '9' {
			r |= rune(ch - '0')
		} else if ch >= 'a' && ch <= 'f' {
			r |= rune(ch - 'a' + 10)
		} else if ch >= 'A' && ch <= 'F' {
			r |= rune(ch - 'A' + 10)
		}
	}
	return r
}

func (l *Lexer) scanNumberOrDatetime() Token {
	line := l.line
	col := l.col
	start := l.pos

	// Handle sign
	hasSign := false
	if l.peek() == '-' || l.peek() == '+' {
		hasSign = true
		l.advance()
	}

	// Check for special values (inf, nan) - must check before scanning digits
	if l.peek() == 'i' {
		// Check for "inf" or "infinity"
		if l.matchWord("inf") || l.matchWord("infinity") {
			return Token{Type: TokenFloat, Value: l.input[start:l.pos], Line: line, Col: col}
		}
	}
	if l.peek() == 'n' {
		// Check for "nan"
		if l.matchWord("nan") {
			return Token{Type: TokenFloat, Value: l.input[start:l.pos], Line: line, Col: col}
		}
	}

	// If had sign but not inf/nan, continue with number
	_ = hasSign

	// Scan digits before decimal/hour (including underscores)
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsDigit(rune(ch)) || ch == '_' {
			l.advance()
		} else {
			break
		}
	}

	// Check for datetime format: YYYY-MM-DD
	if l.peek() == '-' && l.pos-start >= 4 {
		return l.scanDatetime(start, line, col)
	}

	// Check for time format: HH:MM:SS
	if l.peek() == ':' {
		return l.scanTime(start, line, col)
	}

	// Check for hex, octal, binary
	if l.input[start] == '0' && l.pos-start == 1 {
		ch := l.peek()
		if ch == 'x' || ch == 'o' || ch == 'b' {
			return l.scanSpecialNumber(start, line, col)
		}
	}

	// Check for float
	if l.peek() == '.' || l.peek() == 'e' || l.peek() == 'E' {
		return l.scanFloat(start, line, col)
	}

	return Token{Type: TokenInteger, Value: l.input[start:l.pos], Line: line, Col: col}
}

func (l *Lexer) scanDatetime(start, line, col int) Token {
	// Already have YYYY, scan -MM-DD
	l.advance() // skip -
	for i := 0; i < 2 && l.pos < len(l.input); i++ {
		if !unicode.IsDigit(rune(l.peek())) {
			return Token{Type: TokenError, Value: "invalid date", Line: line, Col: col}
		}
		l.advance()
	}
	if l.peek() != '-' {
		return Token{Type: TokenDate, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	l.advance()
	for i := 0; i < 2 && l.pos < len(l.input); i++ {
		if !unicode.IsDigit(rune(l.peek())) {
			return Token{Type: TokenError, Value: "invalid date", Line: line, Col: col}
		}
		l.advance()
	}

	// Check for time component
	if l.peek() != 'T' && l.peek() != ' ' && l.peek() != 't' {
		return Token{Type: TokenDate, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	l.advance() // skip T/space

	// Scan time HH:MM:SS
	for i := 0; i < 2 && l.pos < len(l.input); i++ {
		if !unicode.IsDigit(rune(l.peek())) {
			return Token{Type: TokenError, Value: "invalid datetime", Line: line, Col: col}
		}
		l.advance()
	}
	if l.peek() != ':' {
		return Token{Type: TokenDatetime, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	l.advance()
	for i := 0; i < 2 && l.pos < len(l.input); i++ {
		if !unicode.IsDigit(rune(l.peek())) {
			return Token{Type: TokenError, Value: "invalid datetime", Line: line, Col: col}
		}
		l.advance()
	}
	if l.peek() != ':' {
		return Token{Type: TokenDatetime, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	l.advance()
	for i := 0; i < 2 && l.pos < len(l.input); i++ {
		if !unicode.IsDigit(rune(l.peek())) {
			return Token{Type: TokenError, Value: "invalid datetime", Line: line, Col: col}
		}
		l.advance()
	}

	// Scan fractional seconds
	if l.peek() == '.' {
		l.advance()
		for l.pos < len(l.input) && unicode.IsDigit(rune(l.peek())) {
			l.advance()
		}
	}

	// Scan timezone
	if l.peek() == 'Z' || l.peek() == 'z' {
		l.advance()
	} else if l.peek() == '+' || l.peek() == '-' {
		l.advance()
		for i := 0; i < 2 && l.pos < len(l.input); i++ {
			if !unicode.IsDigit(rune(l.peek())) {
				break
			}
			l.advance()
		}
		if l.peek() == ':' {
			l.advance()
			for i := 0; i < 2 && l.pos < len(l.input); i++ {
				if !unicode.IsDigit(rune(l.peek())) {
					break
				}
				l.advance()
			}
		}
	}

	return Token{Type: TokenDatetime, Value: l.input[start:l.pos], Line: line, Col: col}
}

func (l *Lexer) scanTime(start, line, col int) Token {
	// Already have HH, scan :MM:SS
	l.advance() // skip :
	for i := 0; i < 2 && l.pos < len(l.input); i++ {
		if !unicode.IsDigit(rune(l.peek())) {
			return Token{Type: TokenError, Value: "invalid time", Line: line, Col: col}
		}
		l.advance()
	}
	if l.peek() != ':' {
		return Token{Type: TokenTime, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	l.advance()
	for i := 0; i < 2 && l.pos < len(l.input); i++ {
		if !unicode.IsDigit(rune(l.peek())) {
			return Token{Type: TokenError, Value: "invalid time", Line: line, Col: col}
		}
		l.advance()
	}

	// Scan fractional seconds
	if l.peek() == '.' {
		l.advance()
		for l.pos < len(l.input) && unicode.IsDigit(rune(l.peek())) {
			l.advance()
		}
	}

	return Token{Type: TokenTime, Value: l.input[start:l.pos], Line: line, Col: col}
}

func (l *Lexer) scanFloat(start, line, col int) Token {
	// Scan decimal part (including underscores)
	if l.peek() == '.' {
		l.advance()
		for l.pos < len(l.input) {
			ch := l.peek()
			if unicode.IsDigit(rune(ch)) || ch == '_' {
				l.advance()
			} else {
				break
			}
		}
	}

	// Scan exponent (including underscores in exponent)
	if l.peek() == 'e' || l.peek() == 'E' {
		l.advance()
		if l.peek() == '+' || l.peek() == '-' {
			l.advance()
		}
		for l.pos < len(l.input) {
			ch := l.peek()
			if unicode.IsDigit(rune(ch)) || ch == '_' {
				l.advance()
			} else {
				break
			}
		}
	}

	return Token{Type: TokenFloat, Value: l.input[start:l.pos], Line: line, Col: col}
}

func (l *Lexer) scanSpecialNumber(start, line, col int) Token {
	l.advance() // skip x/o/b
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsDigit(rune(ch)) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') || ch == '_' {
			l.advance()
		} else {
			break
		}
	}
	return Token{Type: TokenInteger, Value: l.input[start:l.pos], Line: line, Col: col}
}

func (l *Lexer) scanBoolean() Token {
	line := l.line
	col := l.col
	start := l.pos

	if l.matchWord("true") {
		return Token{Type: TokenBoolean, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	if l.matchWord("false") {
		return Token{Type: TokenBoolean, Value: l.input[start:l.pos], Line: line, Col: col}
	}

	// Not a boolean, scan as key
	return l.scanKey()
}

func (l *Lexer) scanBooleanOrSpecial() Token {
	line := l.line
	col := l.col
	start := l.pos

	// Check for boolean
	if l.matchWord("true") {
		return Token{Type: TokenBoolean, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	if l.matchWord("false") {
		return Token{Type: TokenBoolean, Value: l.input[start:l.pos], Line: line, Col: col}
	}

	// Check for special float values
	if l.matchWord("inf") || l.matchWord("infinity") {
		return Token{Type: TokenFloat, Value: l.input[start:l.pos], Line: line, Col: col}
	}
	if l.matchWord("nan") {
		return Token{Type: TokenFloat, Value: l.input[start:l.pos], Line: line, Col: col}
	}

	// Not a boolean or special value, scan as key
	return l.scanKey()
}

func (l *Lexer) matchWord(word string) bool {
	if len(l.input)-l.pos < len(word) {
		return false
	}
	if strings.ToLower(l.input[l.pos:l.pos+len(word)]) != word {
		return false
	}
	// Check that the word is not part of a longer identifier
	if l.pos+len(word) < len(l.input) {
		next := rune(l.input[l.pos+len(word)])
		if unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_' || next == '-' {
			return false
		}
	}
	l.pos += len(word)
	l.col += len(word)
	return true
}

func (l *Lexer) scanKey() Token {
	line := l.line
	col := l.col
	start := l.pos

	// Bare key: letters, digits, underscore, dash
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' || ch == '-' {
			l.advance()
		} else {
			break
		}
	}

	if l.pos == start {
		return Token{Type: TokenError, Value: fmt.Sprintf("unexpected character: %c", l.peek()), Line: line, Col: col}
	}

	return Token{Type: TokenKey, Value: l.input[start:l.pos], Line: line, Col: col}
}

// IsValidKey checks if a string is a valid bare key
func IsValidKey(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' && ch != '-' {
			return false
		}
	}
	return true
}

// utf8RuneCount counts runes in a string
func utf8RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}
