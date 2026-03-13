# Xxlang Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a complete interpreted scripting language in Go with bytecode VM, lightweight OOP, modules, plugins, and comprehensive standard library.

**Architecture:** Layered architecture with lexer → parser → compiler → VM pipeline. All values are objects with a common base class. Stack-based bytecode execution with call frames.

**Tech Stack:** Go 1.21+, Go standard library (database/sql, net/http, crypto, compress, encoding/json, regexp)

---

## Phase 1: Core Language (MVP)

### Task 1: Project Setup

**Files:**
- Create: `go.mod`
- Create: `go.sum` (auto-generated)
- Create: `.gitignore`
- Create: `README.md`

**Step 1: Initialize Go module**

Run: `go mod init github.com/topxeq/xxlang`
Expected: Creates go.mod file

**Step 2: Create .gitignore**

```
# Binaries
xxlang
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo

# Build
/bin/
/dist/

# Plugins
*.so
```

**Step 3: Create README.md**

```markdown
# Xxlang

Xxlang (Chinese: 现象语言) is a line-by-line interpreted scripting language implemented in Go.

## Features

- Bytecode virtual machine for efficient execution
- Lightweight OOP with single inheritance
- Closures and first-class functions
- Module system for code organization
- Plugin system for extending functionality
- Comprehensive standard library

## Installation

```bash
go install github.com/topxeq/xxlang/cmd/xxlang@latest
```

## Quick Start

```bash
# Run a file
xxlang script.xxl

# Start REPL
xxlang -i

# Evaluate code
xxlang -e "println(1 + 2);"
```

## License

MIT
```

**Step 4: Commit**

```bash
git add go.mod .gitignore README.md
git commit -m "chore: initialize project"
```

---

### Task 2: Token Types

**Files:**
- Create: `pkg/lexer/token.go`
- Create: `pkg/lexer/token_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/lexer/... -v`
Expected: FAIL - undefined: TokenType, Token

**Step 3: Write minimal implementation**

```go
// pkg/lexer/token.go
package lexer

import "fmt"

// TokenType represents the type of a token
type TokenType string

// Token types
const (
	// Literals
	TokenInt    TokenType = "INT"
	TokenFloat  TokenType = "FLOAT"
	TokenString TokenType = "STRING"
	TokenIdent  TokenType = "IDENT"

	// Operators
	TokenPlus       TokenType = "+"
	TokenMinus      TokenType = "-"
	TokenAsterisk   TokenType = "*"
	TokenSlash      TokenType = "/"
	TokenPercent    TokenType = "%"
	TokenAssign     TokenType = "="
	TokenPlusAssign TokenType = "+="
	TokenMinusAssign TokenType = "-="
	TokenEqual      TokenType = "=="
	TokenNotEqual   TokenType = "!="
	TokenLT         TokenType = "<"
	TokenGT         TokenType = ">"
	TokenLTE        TokenType = "<="
	TokenGTE        TokenType = ">="
	TokenAnd        TokenType = "&&"
	TokenOr         TokenType = "||"
	TokenNot        TokenType = "!"
	TokenIncrement  TokenType = "++"
	TokenDecrement  TokenType = "--"

	// Delimiters
	TokenComma     TokenType = ","
	TokenColon     TokenType = ":"
	TokenSemicolon TokenType = ";"
	TokenLParen    TokenType = "("
	TokenRParen    TokenType = ")"
	TokenLBrace    TokenType = "{"
	TokenRBrace    TokenType = "}"
	TokenLBracket  TokenType = "["
	TokenRBracket  TokenType = "]"
	TokenDot       TokenType = "."
	TokenArrow     TokenType = "=>"

	// Keywords
	TokenVar      TokenType = "VAR"
	TokenConst    TokenType = "CONST"
	TokenFunc     TokenType = "FUNC"
	TokenReturn   TokenType = "RETURN"
	TokenIf       TokenType = "IF"
	TokenElse     TokenType = "ELSE"
	TokenWhile    TokenType = "WHILE"
	TokenFor      TokenType = "FOR"
	TokenIn       TokenType = "IN"
	TokenBreak    TokenType = "BREAK"
	TokenContinue TokenType = "CONTINUE"
	TokenSwitch   TokenType = "SWITCH"
	TokenCase     TokenType = "CASE"
	TokenDefault  TokenType = "DEFAULT"
	TokenTry      TokenType = "TRY"
	TokenCatch    TokenType = "CATCH"
	TokenFinally  TokenType = "FINALLY"
	TokenThrow    TokenType = "THROW"
	TokenClass    TokenType = "CLASS"
	TokenExtends  TokenType = "EXTENDS"
	TokenNew      TokenType = "NEW"
	TokenThis     TokenType = "THIS"
	TokenNull     TokenType = "NULL"
	TokenTrue     TokenType = "TRUE"
	TokenFalse    TokenType = "FALSE"
	TokenImport   TokenType = "IMPORT"
	TokenExport   TokenType = "EXPORT"

	// Special
	TokenEOF     TokenType = "EOF"
	TokenIllegal TokenType = "ILLEGAL"
)

// Token represents a lexical token
type Token struct {
	Type     TokenType
	Literal  string
	Line     int
	Column   int
}

// String returns a string representation of the token
func (t Token) String() string {
	return fmt.Sprintf("%s(%s) at %d:%d", t.Type, t.Literal, t.Line, t.Column)
}

// Keywords maps keywords to their token types
var Keywords = map[string]TokenType{
	"var":      TokenVar,
	"const":    TokenConst,
	"func":     TokenFunc,
	"return":   TokenReturn,
	"if":       TokenIf,
	"else":     TokenElse,
	"while":    TokenWhile,
	"for":      TokenFor,
	"in":       TokenIn,
	"break":    TokenBreak,
	"continue": TokenContinue,
	"switch":   TokenSwitch,
	"case":     TokenCase,
	"default":  TokenDefault,
	"try":      TokenTry,
	"catch":    TokenCatch,
	"finally":  TokenFinally,
	"throw":    TokenThrow,
	"class":    TokenClass,
	"extends":  TokenExtends,
	"new":      TokenNew,
	"this":     TokenThis,
	"null":     TokenNull,
	"true":     TokenTrue,
	"false":    TokenFalse,
	"import":   TokenImport,
	"export":   TokenExport,
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/lexer/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/lexer/
git commit -m "feat(lexer): add token types"
```

---

### Task 3: Lexer Implementation

**Files:**
- Create: `pkg/lexer/lexer.go`
- Create: `pkg/lexer/lexer_test.go`

**Step 1: Write the failing test for basic tokens**

```go
// pkg/lexer/lexer_test.go
package lexer

import "testing"

func TestNextToken_BasicTokens(t *testing.T) {
	input := `=+(){},;`

	expected := []Token{
		{Type: TokenAssign, Literal: "=", Line: 1, Column: 1},
		{Type: TokenPlus, Literal: "+", Line: 1, Column: 2},
		{Type: TokenLParen, Literal: "(", Line: 1, Column: 3},
		{Type: TokenRParen, Literal: ")", Line: 1, Column: 4},
		{Type: TokenLBrace, Literal: "{", Line: 1, Column: 5},
		{Type: TokenRBrace, Literal: "}", Line: 1, Column: 6},
		{Type: TokenComma, Literal: ",", Line: 1, Column: 7},
		{Type: TokenSemicolon, Literal: ";", Line: 1, Column: 8},
		{Type: TokenEOF, Literal: "", Line: 1, Column: 9},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_Code(t *testing.T) {
	input := `var x = 5;
var y = 10;
var add = func(a, b) {
	return a + b;
};`

	expected := []Token{
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "x"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenInt, Literal: "5"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "y"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenInt, Literal: "10"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "add"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenFunc, Literal: "func"},
		{Type: TokenLParen, Literal: "("},
		{Type: TokenIdent, Literal: "a"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenIdent, Literal: "b"},
		{Type: TokenRParen, Literal: ")"},
		{Type: TokenLBrace, Literal: "{"},
		{Type: TokenReturn, Literal: "return"},
		{Type: TokenIdent, Literal: "a"},
		{Type: TokenPlus, Literal: "+"},
		{Type: TokenIdent, Literal: "b"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenRBrace, Literal: "}"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/lexer/... -v`
Expected: FAIL - undefined: New, Lexer

**Step 3: Write minimal implementation**

```go
// pkg/lexer/lexer.go
package lexer

// Lexer tokenizes input source code
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// New creates a new Lexer for the given input
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenEqual, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenAssign, l.ch, tok.Line, tok.Column)
		}
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenPlusAssign, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenIncrement, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenPlus, l.ch, tok.Line, tok.Column)
		}
	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenMinusAssign, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenDecrement, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenMinus, l.ch, tok.Line, tok.Column)
		}
	case '*':
		tok = newToken(TokenAsterisk, l.ch, tok.Line, tok.Column)
	case '/':
		if l.peekChar() == '/' {
			// Single line comment
			l.skipComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			// Multi-line comment
			l.skipMultiLineComment()
			return l.NextToken()
		}
		tok = newToken(TokenSlash, l.ch, tok.Line, tok.Column)
	case '%':
		tok = newToken(TokenPercent, l.ch, tok.Line, tok.Column)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenNotEqual, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenNot, l.ch, tok.Line, tok.Column)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenLTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenLT, l.ch, tok.Line, tok.Column)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenGTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenGT, l.ch, tok.Line, tok.Column)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenAnd, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenIllegal, l.ch, tok.Line, tok.Column)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TokenOr, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(TokenIllegal, l.ch, tok.Line, tok.Column)
		}
	case ',':
		tok = newToken(TokenComma, l.ch, tok.Line, tok.Column)
	case ':':
		tok = newToken(TokenColon, l.ch, tok.Line, tok.Column)
	case ';':
		tok = newToken(TokenSemicolon, l.ch, tok.Line, tok.Column)
	case '(':
		tok = newToken(TokenLParen, l.ch, tok.Line, tok.Column)
	case ')':
		tok = newToken(TokenRParen, l.ch, tok.Line, tok.Column)
	case '{':
		tok = newToken(TokenLBrace, l.ch, tok.Line, tok.Column)
	case '}':
		tok = newToken(TokenRBrace, l.ch, tok.Line, tok.Column)
	case '[':
		tok = newToken(TokenLBracket, l.ch, tok.Line, tok.Column)
	case ']':
		tok = newToken(TokenRBracket, l.ch, tok.Line, tok.Column)
	case '.':
		tok = newToken(TokenDot, l.ch, tok.Line, tok.Column)
	case '"':
		tok.Type = TokenString
		tok.Literal = l.readString()
	case '\'':
		tok.Type = TokenString
		tok.Literal = l.readStringSingleQuote()
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			tok.Type = LookupIdent(ident)
			tok.Literal = ident
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TokenInt
			tok.Literal = l.readNumber()
			// Check for float
			if l.ch == '.' {
				l.readChar()
				fraction := l.readNumber()
				tok.Type = TokenFloat
				tok.Literal = tok.Literal + "." + fraction
			}
			// Check for exponent
			if l.ch == 'e' || l.ch == 'E' {
				l.readChar()
				exponent := "e"
				if l.ch == '+' || l.ch == '-' {
					exponent += string(l.ch)
					l.readChar()
				}
				exponent += l.readNumber()
				tok.Type = TokenFloat
				tok.Literal = tok.Literal + exponent
			}
			return tok
		} else {
			tok = newToken(TokenIllegal, l.ch, tok.Line, tok.Column)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType TokenType, ch byte, line, column int) Token {
	return Token{Type: tokenType, Literal: string(ch), Line: line, Column: column}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar() // Skip escape character
		}
	}
	str := l.input[position:l.position]
	// Process escape sequences
	return processEscapeSequences(str)
}

func (l *Lexer) readStringSingleQuote() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar() // Skip escape character
		}
	}
	str := l.input[position:l.position]
	return processEscapeSequences(str)
}

func processEscapeSequences(s string) string {
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case '\'':
				result = append(result, '\'')
			case '0':
				result = append(result, 0)
			default:
				result = append(result, s[i+1])
			}
			i += 2
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipMultiLineComment() {
	l.readChar() // Skip *
	l.readChar()
	for {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			break
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		if l.ch == 0 {
			break
		}
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/lexer/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/lexer/
git commit -m "feat(lexer): implement lexer with tokenization"
```

---

### Task 4: Object System - Base Interface

**Files:**
- Create: `pkg/objects/object.go`
- Create: `pkg/objects/object_test.go`

**Step 1: Write the failing test**

```go
// pkg/objects/object_test.go
package objects

import "testing"

func TestObjectType(t *testing.T) {
	tests := []struct {
		obj      Object
		expected ObjectType
	}{
		{NULL, NullType},
		{TRUE, BoolType},
		{FALSE, BoolType},
	}

	for _, tt := range tests {
		if got := tt.obj.Type(); got != tt.expected {
			t.Errorf("object.Type() = %s, want %s", got, tt.expected)
		}
	}
}

func TestNullInspect(t *testing.T) {
	if got := NULL.Inspect(); got != "null" {
		t.Errorf("NULL.Inspect() = %s, want null", got)
	}
}

func TestBoolInspect(t *testing.T) {
	if got := TRUE.Inspect(); got != "true" {
		t.Errorf("TRUE.Inspect() = %s, want true", got)
	}
	if got := FALSE.Inspect(); got != "false" {
		t.Errorf("FALSE.Inspect() = %s, want false", got)
	}
}

func TestBoolToBool(t *testing.T) {
	if TRUE.ToBool() != TRUE {
		t.Error("TRUE.ToBool() should return TRUE")
	}
	if FALSE.ToBool() != FALSE {
		t.Error("FALSE.ToBool() should return FALSE")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -v`
Expected: FAIL - undefined: Object, ObjectType, NULL, etc.

**Step 3: Write minimal implementation**

```go
// pkg/objects/object.go
package objects

import "fmt"

// ObjectType represents the type of an object
type ObjectType string

// Object types
const (
	NullType    ObjectType = "NULL"
	IntType     ObjectType = "INT"
	FloatType   ObjectType = "FLOAT"
	StringType  ObjectType = "STRING"
	BoolType    ObjectType = "BOOL"
	ArrayType   ObjectType = "ARRAY"
	MapType     ObjectType = "MAP"
	FunctionType ObjectType = "FUNCTION"
	BuiltinType ObjectType = "BUILTIN"
	BytesType   ObjectType = "BYTES"
	ClassType   ObjectType = "CLASS"
	InstanceType ObjectType = "INSTANCE"
	ErrorType   ObjectType = "ERROR"
	ReturnType  ObjectType = "RETURN"
)

// Object is the base interface for all values in Xxlang
type Object interface {
	Type() ObjectType
	Inspect() string
	ToBool() *Bool
	HashKey() HashKey
}

// HashKey is used for map keys
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Null represents the null value
type Null struct{}

func (n *Null) Type() ObjectType { return NullType }
func (n *Null) Inspect() string  { return "null" }
func (n *Null) ToBool() *Bool    { return FALSE }
func (n *Null) HashKey() HashKey { return HashKey{Type: NullType, Value: 0} }

// NULL is the singleton null value
var NULL = &Null{}

// Bool represents a boolean value
type Bool struct {
	Value bool
}

func (b *Bool) Type() ObjectType { return BoolType }
func (b *Bool) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}
func (b *Bool) ToBool() *Bool { return b }
func (b *Bool) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	}
	return HashKey{Type: BoolType, Value: value}
}

// TRUE and FALSE are singleton boolean values
var (
	TRUE  = &Bool{Value: true}
	FALSE = &Bool{Value: false}
)

// Error represents a runtime error
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ErrorType }
func (e *Error) Inspect() string  { return fmt.Sprintf("ERROR: %s", e.Message) }
func (e *Error) ToBool() *Bool    { return FALSE }
func (e *Error) HashKey() HashKey { return HashKey{Type: ErrorType, Value: 0} }

// Return represents a return value (used internally)
type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType { return ReturnType }
func (r *Return) Inspect() string  { return r.Value.Inspect() }
func (r *Return) ToBool() *Bool    { return r.Value.ToBool() }
func (r *Return) HashKey() HashKey { return HashKey{Type: ReturnType, Value: 0} }

// IsTruthy checks if an object is truthy
func IsTruthy(obj Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/objects/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "feat(objects): add base object interface, null, bool, error types"
```

---

### Task 5: Object System - Int and Float

**Files:**
- Create: `pkg/objects/int.go`
- Create: `pkg/objects/float.go`
- Create: `pkg/objects/number_test.go`

**Step 1: Write the failing test**

```go
// pkg/objects/number_test.go
package objects

import "testing"

func TestIntInspect(t *testing.T) {
	i := &Int{Value: 42}
	if got := i.Inspect(); got != "42" {
		t.Errorf("Int.Inspect() = %s, want 42", got)
	}
}

func TestIntType(t *testing.T) {
	i := &Int{Value: 42}
	if got := i.Type(); got != IntType {
		t.Errorf("Int.Type() = %s, want INT", got)
	}
}

func TestIntToBool(t *testing.T) {
	zero := &Int{Value: 0}
	if zero.ToBool() != FALSE {
		t.Error("Int(0).ToBool() should be FALSE")
	}

	nonzero := &Int{Value: 42}
	if nonzero.ToBool() != TRUE {
		t.Error("Int(42).ToBool() should be TRUE")
	}
}

func TestIntHashKey(t *testing.T) {
	a := &Int{Value: 42}
	b := &Int{Value: 42}
	c := &Int{Value: 43}

	if a.HashKey() != b.HashKey() {
		t.Error("same int values should have same hash keys")
	}
	if a.HashKey() == c.HashKey() {
		t.Error("different int values should have different hash keys")
	}
}

func TestFloatInspect(t *testing.T) {
	f := &Float{Value: 3.14}
	if got := f.Inspect(); got != "3.14" {
		t.Errorf("Float.Inspect() = %s, want 3.14", got)
	}
}

func TestFloatType(t *testing.T) {
	f := &Float{Value: 3.14}
	if got := f.Type(); got != FloatType {
		t.Errorf("Float.Type() = %s, want FLOAT", got)
	}
}

func TestFloatToBool(t *testing.T) {
	zero := &Float{Value: 0.0}
	if zero.ToBool() != FALSE {
		t.Error("Float(0).ToBool() should be FALSE")
	}

	nonzero := &Float{Value: 3.14}
	if nonzero.ToBool() != TRUE {
		t.Error("Float(3.14).ToBool() should be TRUE")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -v`
Expected: FAIL - undefined: Int, Float

**Step 3: Write minimal implementation**

```go
// pkg/objects/int.go
package objects

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

// Int represents an integer value
type Int struct {
	Value int64
}

func (i *Int) Type() ObjectType { return IntType }
func (i *Int) Inspect() string  { return strconv.FormatInt(i.Value, 10) }
func (i *Int) ToBool() *Bool {
	if i.Value == 0 {
		return FALSE
	}
	return TRUE
}
func (i *Int) HashKey() HashKey {
	return HashKey{Type: IntType, Value: uint64(i.Value)}
}

// Int built-in methods
func (i *Int) Method(name string) Object {
	switch name {
	case "toStr":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &String{Value: i.Inspect()}
			},
		}
	case "toFloat":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &Float{Value: float64(i.Value)}
			},
		}
	case "abs":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if i.Value < 0 {
					return &Int{Value: -i.Value}
				}
				return i
			},
		}
	}
	return nil
}
```

```go
// pkg/objects/float.go
package objects

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
)

// Float represents a floating-point value
type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FloatType }
func (f *Float) Inspect() string  { return strconv.FormatFloat(f.Value, 'f', -1, 64) }
func (f *Float) ToBool() *Bool {
	if f.Value == 0.0 {
		return FALSE
	}
	return TRUE
}
func (f *Float) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%f", f.Value)))
	return HashKey{Type: FloatType, Value: h.Sum64()}
}

// Float built-in methods
func (f *Float) Method(name string) Object {
	switch name {
	case "toStr":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &String{Value: f.Inspect()}
			},
		}
	case "toInt":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &Int{Value: int64(f.Value)}
			},
		}
	case "round":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &Int{Value: int64(math.Round(f.Value))}
			},
		}
	case "floor":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &Int{Value: int64(math.Floor(f.Value))}
			},
		}
	case "ceil":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &Int{Value: int64(math.Ceil(f.Value))}
			},
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/objects/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "feat(objects): add int and float types"
```

---

### Task 6: Object System - String

**Files:**
- Create: `pkg/objects/string.go`
- Create: `pkg/objects/string_test.go`

**Step 1: Write the failing test**

```go
// pkg/objects/string_test.go
package objects

import "testing"

func TestStringInspect(t *testing.T) {
	s := &String{Value: "hello"}
	if got := s.Inspect(); got != "hello" {
		t.Errorf("String.Inspect() = %s, want hello", got)
	}
}

func TestStringType(t *testing.T) {
	s := &String{Value: "hello"}
	if got := s.Type(); got != StringType {
		t.Errorf("String.Type() = %s, want STRING", got)
	}
}

func TestStringToBool(t *testing.T) {
	empty := &String{Value: ""}
	if empty.ToBool() != FALSE {
		t.Error("String(\"\").ToBool() should be FALSE")
	}

	nonempty := &String{Value: "hello"}
	if nonempty.ToBool() != TRUE {
		t.Error("String(\"hello\").ToBool() should be TRUE")
	}
}

func TestStringHashKey(t *testing.T) {
	a := &String{Value: "hello"}
	b := &String{Value: "hello"}
	c := &String{Value: "world"}

	if a.HashKey() != b.HashKey() {
		t.Error("same string values should have same hash keys")
	}
	if a.HashKey() == c.HashKey() {
		t.Error("different string values should have different hash keys")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -v`
Expected: FAIL - undefined: String

**Step 3: Write minimal implementation**

```go
// pkg/objects/string.go
package objects

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

// String represents a string value
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return StringType }
func (s *String) Inspect() string  { return s.Value }
func (s *String) ToBool() *Bool {
	if len(s.Value) == 0 {
		return FALSE
	}
	return TRUE
}
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: StringType, Value: h.Sum64()}
}

// String built-in methods
func (s *String) Method(name string) Object {
	switch name {
	case "len":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &Int{Value: int64(len(s.Value))}
			},
		}
	case "toStr":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return s
			},
		}
	case "upper":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &String{Value: strings.ToUpper(s.Value)}
			},
		}
	case "lower":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &String{Value: strings.ToLower(s.Value)}
			},
		}
	case "trim":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &String{Value: strings.TrimSpace(s.Value)}
			},
		}
	case "trimLeft":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &String{Value: strings.TrimLeft(s.Value, " \t\n\r")}
			},
		}
	case "trimRight":
		return &Builtin{
			Fn: func(args ...Object) Object {
				return &String{Value: strings.TrimRight(s.Value, " \t\n\r")}
			},
		}
	case "contains":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments for contains. got=%d, want=1", len(args))
				}
				sub, ok := args[0].(*String)
				if !ok {
					return newError("argument to contains must be STRING, got %s", args[0].Type())
				}
				if strings.Contains(s.Value, sub.Value) {
					return TRUE
				}
				return FALSE
			},
		}
	case "indexOf":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments for indexOf. got=%d, want=1", len(args))
				}
				sub, ok := args[0].(*String)
				if !ok {
					return newError("argument to indexOf must be STRING, got %s", args[0].Type())
				}
				return &Int{Value: int64(strings.Index(s.Value, sub.Value))}
			},
		}
	case "lastIndexOf":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments for lastIndexOf. got=%d, want=1", len(args))
				}
				sub, ok := args[0].(*String)
				if !ok {
					return newError("argument to lastIndexOf must be STRING, got %s", args[0].Type())
				}
				return &Int{Value: int64(strings.LastIndex(s.Value, sub.Value))}
			},
		}
	case "split":
		return &Builtin{
			Fn: func(args ...Object) Object {
				sep := " "
				if len(args) > 0 {
					sepObj, ok := args[0].(*String)
					if !ok {
						return newError("argument to split must be STRING, got %s", args[0].Type())
					}
					sep = sepObj.Value
				}
				parts := strings.Split(s.Value, sep)
				elements := make([]Object, len(parts))
				for i, p := range parts {
					elements[i] = &String{Value: p}
				}
				return &Array{Elements: elements}
			},
		}
	case "replace":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return newError("wrong number of arguments for replace. got=%d, want=2", len(args))
				}
				old, ok := args[0].(*String)
				if !ok {
					return newError("first argument to replace must be STRING, got %s", args[0].Type())
				}
				newStr, ok := args[1].(*String)
				if !ok {
					return newError("second argument to replace must be STRING, got %s", args[1].Type())
				}
				return &String{Value: strings.ReplaceAll(s.Value, old.Value, newStr.Value)}
			},
		}
	case "substring":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) < 1 || len(args) > 2 {
					return newError("wrong number of arguments for substring. got=%d, want=1 or 2", len(args))
				}
				start, ok := args[0].(*Int)
				if !ok {
					return newError("first argument to substring must be INT, got %s", args[0].Type())
				}
				startIdx := int(start.Value)
				if startIdx < 0 {
					startIdx = 0
				}
				if startIdx > len(s.Value) {
					startIdx = len(s.Value)
				}

				if len(args) == 1 {
					return &String{Value: s.Value[startIdx:]}
				}

				end, ok := args[1].(*Int)
				if !ok {
					return newError("second argument to substring must be INT, got %s", args[1].Type())
				}
				endIdx := int(end.Value)
				if endIdx < startIdx {
					endIdx = startIdx
				}
				if endIdx > len(s.Value) {
					endIdx = len(s.Value)
				}
				return &String{Value: s.Value[startIdx:endIdx]}
			},
		}
	case "startsWith":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments for startsWith. got=%d, want=1", len(args))
				}
				prefix, ok := args[0].(*String)
				if !ok {
					return newError("argument to startsWith must be STRING, got %s", args[0].Type())
				}
				if strings.HasPrefix(s.Value, prefix.Value) {
					return TRUE
				}
				return FALSE
			},
		}
	case "endsWith":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments for endsWith. got=%d, want=1", len(args))
				}
				suffix, ok := args[0].(*String)
				if !ok {
					return newError("argument to endsWith must be STRING, got %s", args[0].Type())
				}
				if strings.HasSuffix(s.Value, suffix.Value) {
					return TRUE
				}
				return FALSE
			},
		}
	case "repeat":
		return &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments for repeat. got=%d, want=1", len(args))
				}
				count, ok := args[0].(*Int)
				if !ok {
					return newError("argument to repeat must be INT, got %s", args[0].Type())
				}
				return &String{Value: strings.Repeat(s.Value, int(count.Value))}
			},
		}
	case "parseInt":
		return &Builtin{
			Fn: func(args ...Object) Object {
				val, err := strconv.ParseInt(s.Value, 0, 64)
				if err != nil {
					return newError("could not parse '%s' as int", s.Value)
				}
				return &Int{Value: val}
			},
		}
	case "parseFloat":
		return &Builtin{
			Fn: func(args ...Object) Object {
				val, err := strconv.ParseFloat(s.Value, 64)
				if err != nil {
					return newError("could not parse '%s' as float", s.Value)
				}
				return &Float{Value: val}
			},
		}
	}
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/objects/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "feat(objects): add string type with methods"
```

---

### Task 7: Object System - Array

**Files:**
- Create: `pkg/objects/array.go`
- Create: `pkg/objects/array_test.go`

**Step 1: Write the failing test**

```go
// pkg/objects/array_test.go
package objects

import "testing"

func TestArrayInspect(t *testing.T) {
	arr := &Array{Elements: []Object{
		&Int{Value: 1},
		&Int{Value: 2},
		&Int{Value: 3},
	}}
	if got := arr.Inspect(); got != "[1, 2, 3]" {
		t.Errorf("Array.Inspect() = %s, want [1, 2, 3]", got)
	}
}

func TestArrayType(t *testing.T) {
	arr := &Array{Elements: []Object{}}
	if got := arr.Type(); got != ArrayType {
		t.Errorf("Array.Type() = %s, want ARRAY", got)
	}
}

func TestArrayToBool(t *testing.T) {
	empty := &Array{Elements: []Object{}}
	if empty.ToBool() != FALSE {
		t.Error("Array([]).ToBool() should be FALSE")
	}

	nonempty := &Array{Elements: []Object{&Int{Value: 1}}}
	if nonempty.ToBool() != TRUE {
		t.Error("Array([1]).ToBool() should be TRUE")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -v`
Expected: FAIL - undefined: Array

**Step 3: Write minimal implementation**

```go
// pkg/objects/array.go
package objects

import (
	"bytes"
	"fmt"
)

// Array represents an array value
type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ArrayType }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i, e := range a.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(e.Inspect())
	}
	out.WriteString("]")
	return out.String()
}
func (a *Array) ToBool() *Bool {
	if len(a.Elements) == 0 {
		return FALSE
	}
	return TRUE
}
func (a *Array) HashKey() HashKey {
	// Arrays are not hashable
	return HashKey{Type: ArrayType, Value: 0}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/objects/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "feat(objects): add array type"
```

---

### Task 8: Object System - Map

**Files:**
- Create: `pkg/objects/map.go`
- Create: `pkg/objects/map_test.go`

**Step 1: Write the failing test**

```go
// pkg/objects/map_test.go
package objects

import "testing"

func TestMapInspect(t *testing.T) {
	m := &Map{Pairs: map[HashKey]MapPair{
		{Type: StringType, Value: 123}: {
			Key:   &String{Value: "a"},
			Value: &Int{Value: 1},
		},
	}}
	// Map order is not guaranteed, just check it contains the right parts
	got := m.Inspect()
	if got[0] != '{' || got[len(got)-1] != '}' {
		t.Errorf("Map.Inspect() = %s, expected to start with { and end with }", got)
	}
}

func TestMapType(t *testing.T) {
	m := &Map{Pairs: map[HashKey]MapPair{}}
	if got := m.Type(); got != MapType {
		t.Errorf("Map.Type() = %s, want MAP", got)
	}
}

func TestMapToBool(t *testing.T) {
	empty := &Map{Pairs: map[HashKey]MapPair{}}
	if empty.ToBool() != FALSE {
		t.Error("Map({}).ToBool() should be FALSE")
	}

	nonempty := &Map{Pairs: map[HashKey]MapPair{
		{Type: StringType, Value: 123}: {
			Key:   &String{Value: "a"},
			Value: &Int{Value: 1},
		},
	}}
	if nonempty.ToBool() != TRUE {
		t.Error("Map({a: 1}).ToBool() should be TRUE")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -v`
Expected: FAIL - undefined: Map, MapPair

**Step 3: Write minimal implementation**

```go
// pkg/objects/map.go
package objects

import (
	"bytes"
	"fmt"
)

// MapPair represents a key-value pair in a map
type MapPair struct {
	Key   Object
	Value Object
}

// Map represents a map value
type Map struct {
	Pairs map[HashKey]MapPair
}

func (m *Map) Type() ObjectType { return MapType }
func (m *Map) Inspect() string {
	var out bytes.Buffer
	out.WriteString("{")
	first := true
	for _, pair := range m.Pairs {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(pair.Key.Inspect())
		out.WriteString(": ")
		out.WriteString(pair.Value.Inspect())
	}
	out.WriteString("}")
	return out.String()
}
func (m *Map) ToBool() *Bool {
	if len(m.Pairs) == 0 {
		return FALSE
	}
	return TRUE
}
func (m *Map) HashKey() HashKey {
	// Maps are not hashable
	return HashKey{Type: MapType, Value: 0}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/objects/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "feat(objects): add map type"
```

---

### Task 9: Object System - Builtin and Function

**Files:**
- Create: `pkg/objects/builtin.go`
- Create: `pkg/objects/function.go`

**Step 1: Write the failing test**

```go
// pkg/objects/builtin_test.go
package objects

import "testing"

func TestBuiltinType(t *testing.T) {
	b := &Builtin{Fn: func(args ...Object) Object { return NULL }}
	if got := b.Type(); got != BuiltinType {
		t.Errorf("Builtin.Type() = %s, want BUILTIN", got)
	}
}

func TestBuiltinInspect(t *testing.T) {
	b := &Builtin{Fn: func(args ...Object) Object { return NULL }}
	if got := b.Inspect(); got != "builtin function" {
		t.Errorf("Builtin.Inspect() = %s, want 'builtin function'", got)
	}
}
```

```go
// pkg/objects/function_test.go
package objects

import "testing"

func TestFunctionType(t *testing.T) {
	f := &Function{Parameters: []*Identifier{}, Body: nil, Env: nil}
	if got := f.Type(); got != FunctionType {
		t.Errorf("Function.Type() = %s, want FUNCTION", got)
	}
}

func TestFunctionInspect(t *testing.T) {
	f := &Function{
		Parameters: []*Identifier{{Value: "x"}},
		Body:       nil,
		Env:        nil,
	}
	got := f.Inspect()
	if got != "func(x) { ... }" {
		t.Errorf("Function.Inspect() = %s, want 'func(x) { ... }'", got)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -v`
Expected: FAIL - undefined: Builtin, Function, Identifier

**Step 3: Write minimal implementation**

```go
// pkg/objects/builtin.go
package objects

// BuiltinFunction is the type for built-in functions
type BuiltinFunction func(args ...Object) Object

// Builtin represents a built-in function
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BuiltinType }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) ToBool() *Bool    { return TRUE }
func (b *Builtin) HashKey() HashKey { return HashKey{Type: BuiltinType, Value: 0} }

// Builtins contains all built-in functions
var Builtins = map[string]*Builtin{
	"len": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for len. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				return &Int{Value: int64(len(arg.Value))}
			case *Array:
				return &Int{Value: int64(len(arg.Elements))}
			case *Map:
				return &Int{Value: int64(len(arg.Pairs))}
			default:
				return newError("argument to 'len' not supported, got %s", args[0].Type())
			}
		},
	},
	"print": {
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			return NULL
		},
	},
	"println": {
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return NULL
		},
	},
	"typeOf": {
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments for typeOf. got=%d, want=1", len(args))
			}
			return &String{Value: string(args[0].Type())}
		},
	},
}
```

```go
// pkg/objects/function.go
package objects

import (
	"bytes"
	"fmt"
)

// Identifier represents an identifier (used in function parameters)
type Identifier struct {
	Value string
}

func (i *Identifier) String() string {
	return i.Value
}

// Environment represents a variable scope
type Environment struct {
	Store map[string]Object
	Outer *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	return &Environment{Store: make(map[string]Object), Outer: nil}
}

// NewEnclosedEnvironment creates a new environment with an outer scope
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{Store: make(map[string]Object), Outer: outer}
}

// Get retrieves a variable from the environment
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.Store[name]
	if !ok && e.Outer != nil {
		obj, ok = e.Outer.Get(name)
	}
	return obj, ok
}

// Set sets a variable in the environment
func (e *Environment) Set(name string, val Object) Object {
	e.Store[name] = val
	return val
}

// Function represents a user-defined function
type Function struct {
	Parameters []*Identifier
	Body       interface{} // Will be *ast.BlockStatement, using interface{} to avoid import cycle
	Env        *Environment
	Name       string // Optional: for named functions
}

func (f *Function) Type() ObjectType { return FunctionType }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	out.WriteString("func(")
	for i, p := range f.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(") { ... }")
	return out.String()
}
func (f *Function) ToBool() *Bool    { return TRUE }
func (f *Function) HashKey() HashKey { return HashKey{Type: FunctionType, Value: 0} }
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/objects/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/objects/
git commit -m "feat(objects): add builtin and function types"
```

---

### Task 10: AST Node Definitions

**Files:**
- Create: `pkg/parser/ast.go`
- Create: `pkg/parser/ast_test.go`

**Step 1: Write the failing test**

```go
// pkg/parser/ast_test.go
package parser

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/lexer"
)

func TestProgramString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Name: &Identifier{Value: "x"},
				Value: &IntegerLiteral{Value: 42},
			},
		},
	}

	expected := "var x = 42;"
	if program.String() != expected {
		t.Errorf("program.String() = %q, want %q", program.String(), expected)
	}
}

func TestIdentifierString(t *testing.T) {
	ident := &Identifier{Value: "foo"}
	if ident.String() != "foo" {
		t.Errorf("Identifier.String() = %q, want 'foo'", ident.String())
	}
}

func TestIntegerLiteralString(t *testing.T) {
	il := &IntegerLiteral{Value: 42}
	if il.String() != "42" {
		t.Errorf("IntegerLiteral.String() = %q, want '42'", il.String())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/parser/... -v`
Expected: FAIL - undefined: Program, Statement, etc.

**Step 3: Write minimal implementation**

```go
// pkg/parser/ast.go
package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/topxeq/xxlang/pkg/lexer"
)

// Node represents a node in the AST
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// VarStatement represents a var declaration
type VarStatement struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Literal }
func (vs *VarStatement) String() string {
	var out bytes.Buffer
	out.WriteString(vs.TokenLiteral() + " ")
	out.WriteString(vs.Name.String())
	out.WriteString(" = ")
	if vs.Value != nil {
		out.WriteString(vs.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ConstStatement represents a const declaration
type ConstStatement struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) String() string {
	var out bytes.Buffer
	out.WriteString(cs.TokenLiteral() + " ")
	out.WriteString(cs.Name.String())
	out.WriteString(" = ")
	if cs.Value != nil {
		out.WriteString(cs.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Token       lexer.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// ExpressionStatement represents an expression as a statement
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement represents a block of statements
type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString(" }")
	return out.String()
}

// Identifier represents an identifier
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()       {}
func (i *Identifier) TokenLiteral() string  { return i.Token.Literal }
func (i *Identifier) String() string        { return i.Value }

// IntegerLiteral represents an integer literal
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()       {}
func (il *IntegerLiteral) TokenLiteral() string  { return il.Token.Literal }
func (il *IntegerLiteral) String() string        { return il.Token.Literal }

// FloatLiteral represents a float literal
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()       {}
func (fl *FloatLiteral) TokenLiteral() string  { return fl.Token.Literal }
func (fl *FloatLiteral) String() string        { return fl.Token.Literal }

// StringLiteral represents a string literal
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()       {}
func (sl *StringLiteral) TokenLiteral() string  { return sl.Token.Literal }
func (sl *StringLiteral) String() string        { return fmt.Sprintf("%q", sl.Value) }

// BooleanLiteral represents a boolean literal
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()       {}
func (bl *BooleanLiteral) TokenLiteral() string  { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string        { return bl.Token.Literal }

// NullLiteral represents a null literal
type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode()       {}
func (nl *NullLiteral) TokenLiteral() string  { return nl.Token.Literal }
func (nl *NullLiteral) String() string        { return "null" }

// ArrayLiteral represents an array literal
type ArrayLiteral struct {
	Token    lexer.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()       {}
func (al *ArrayLiteral) TokenLiteral() string  { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i, e := range al.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(e.String())
	}
	out.WriteString("]")
	return out.String()
}

// MapLiteral represents a map literal
type MapLiteral struct {
	Token lexer.Token
	Pairs map[Expression]Expression
}

func (ml *MapLiteral) expressionNode()       {}
func (ml *MapLiteral) TokenLiteral() string  { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	first := true
	for key, value := range ml.Pairs {
		if !first {
			out.WriteString(", ")
		}
		first = false
		out.WriteString(key.String())
		out.WriteString(": ")
		out.WriteString(value.String())
	}
	out.WriteString("}")
	return out.String()
}

// PrefixExpression represents a prefix expression
type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()       {}
func (pe *PrefixExpression) TokenLiteral() string  { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents an infix expression
type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()       {}
func (ie *InfixExpression) TokenLiteral() string  { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// CallExpression represents a function call
type CallExpression struct {
	Token     lexer.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()       {}
func (ce *CallExpression) TokenLiteral() string  { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	for i, a := range ce.Arguments {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(a.String())
	}
	out.WriteString(")")
	return out.String()
}

// IndexExpression represents an index expression
type IndexExpression struct {
	Token lexer.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()       {}
func (ie *IndexExpression) TokenLiteral() string  { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// DotExpression represents a member access expression
type DotExpression struct {
	Token  lexer.Token
	Left   Expression
	Right  *Identifier
}

func (de *DotExpression) expressionNode()       {}
func (de *DotExpression) TokenLiteral() string  { return de.Token.Literal }
func (de *DotExpression) String() string {
	var out bytes.Buffer
	out.WriteString(de.Left.String())
	out.WriteString(".")
	out.WriteString(de.Right.String())
	return out.String()
}

// AssignmentExpression represents an assignment expression
type AssignmentExpression struct {
	Token lexer.Token
	Left  Expression
	Value Expression
}

func (ae *AssignmentExpression) expressionNode()       {}
func (ae *AssignmentExpression) TokenLiteral() string  { return ae.Token.Literal }
func (ae *AssignmentExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ae.Left.String())
	out.WriteString(" = ")
	out.WriteString(ae.Value.String())
	out.WriteString(")")
	return out.String()
}

// FunctionLiteral represents a function literal
type FunctionLiteral struct {
	Token      lexer.Token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()       {}
func (fl *FunctionLiteral) TokenLiteral() string  { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(fl.TokenLiteral())
	if fl.Name != nil {
		out.WriteString(" ")
		out.WriteString(fl.Name.String())
	}
	out.WriteString("(")
	for i, p := range fl.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// IfStatement represents an if statement
type IfStatement struct {
	Token       lexer.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if ")
	out.WriteString(is.Condition.String())
	out.WriteString(" ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

// WhileStatement represents a while statement
type WhileStatement struct {
	Token      lexer.Token
	Condition  Expression
	Body       *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" ")
	out.WriteString(ws.Body.String())
	return out.String()
}

// ForStatement represents a for-in statement
type ForStatement struct {
	Token    lexer.Token
	Var      *Identifier
	Index    *Identifier // optional index variable
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	if fs.Index != nil {
		out.WriteString(fs.Index.String())
		out.WriteString(", ")
	}
	out.WriteString(fs.Var.String())
	out.WriteString(" in ")
	out.WriteString(fs.Iterable.String())
	out.WriteString(" ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// BreakStatement represents a break statement
type BreakStatement struct {
	Token lexer.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break;" }

// ContinueStatement represents a continue statement
type ContinueStatement struct {
	Token lexer.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue;" }
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/parser/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/parser/
git commit -m "feat(parser): add AST node definitions"
```

---

## Continuing Implementation

The plan continues with:
- Task 11: Parser Implementation
- Task 12-15: Compiler Implementation
- Task 16-20: VM Implementation
- Task 21-30: Standard Library
- Task 31-35: Module/Plugin System
- Task 36-40: CLI and Compilation

Due to length, the remaining tasks follow the same TDD pattern with:
1. Write failing test
2. Run test to verify failure
3. Write minimal implementation
4. Run test to verify pass
5. Commit

Each task focuses on a single component with clear acceptance criteria.
