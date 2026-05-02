package jsengine

type TokenType int

const (
	TokEOF TokenType = iota
	TokIdent
	TokNumber
	TokBigInt  // BigInt literal
	TokString
	TokTemplate // Template literal with backticks
	TokPlus
	TokMinus
	TokStar
	TokSlash
	TokPercent
	TokEq
	TokEqEq
	TokEqEqEq
	TokNeq
	TokNeqEq
	TokLt
	TokGt
	TokLte
	TokGte
	TokAnd
	TokOr
	TokNot
	TokBitAnd   // &
	TokBitOr    // |
	TokBitXor   // ^
	TokBitNot   // ~
	TokShiftL   // <<
	TokShiftR   // >>
	TokUShiftR  // >>>
	TokLParen
	TokRParen
	TokLBrace
	TokRBrace
	TokLBracket
	TokRBracket
	TokSemi
	TokComma
	TokDot
	TokQuestion
	TokColon
	TokPlusEq
	TokMinusEq
	TokStarEq
	TokSlashEq
	TokInc
	TokDec
	TokKeyword
	TokArrow   // => for arrow functions
	TokSpread  // ... for spread operator
	TokOptChain // ?. for optional chaining
	TokNullish  // ?? for nullish coalescing
	TokRegex    // /pattern/flags regex literal
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"var":        TokKeyword,
	"let":        TokKeyword,
	"const":      TokKeyword,
	"if":         TokKeyword,
	"else":       TokKeyword,
	"for":        TokKeyword,
	"while":      TokKeyword,
	"do":         TokKeyword,
	"break":      TokKeyword,
	"continue":   TokKeyword,
	"return":     TokKeyword,
	"function":   TokKeyword,
	"true":       TokKeyword,
	"false":      TokKeyword,
	"null":       TokKeyword,
	"undefined":  TokKeyword,
	"new":        TokKeyword,
	"typeof":     TokKeyword,
	"delete":     TokKeyword,
	"void":       TokKeyword,
	"with":       TokKeyword,
	"in":         TokKeyword,
	"instanceof": TokKeyword,
	"switch":     TokKeyword,
	"case":       TokKeyword,
	"default":    TokKeyword,
	"try":        TokKeyword,
	"catch":      TokKeyword,
	"finally":    TokKeyword,
	"throw":      TokKeyword,
	"class":      TokKeyword,
	"extends":    TokKeyword,
	"super":      TokKeyword,
	"static":     TokKeyword,
	"get":        TokKeyword,
	"set":        TokKeyword,
	"this":       TokKeyword,
	"async":      TokKeyword,
	"await":      TokKeyword,
	"of":         TokKeyword,
	"yield":      TokKeyword,
}

type Lexer struct {
	input   string
	pos     int
	ch      byte
	prevTok TokenType
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.pos]
	}
	l.pos++
}

func (l *Lexer) peekChar() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekPeekChar() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	switch l.ch {
	case 0:
		tok = Token{TokEOF, ""}
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			tok = Token{TokInc, "++"}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = Token{TokPlusEq, "+="}
		} else {
			tok = Token{TokPlus, "+"}
		}
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			tok = Token{TokDec, "--"}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = Token{TokMinusEq, "-="}
		} else {
			tok = Token{TokMinus, "-"}
		}
	case '*':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{TokStarEq, "*="}
		} else {
			tok = Token{TokStar, "*"}
		}
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		}
		if l.peekChar() == '*' {
			l.skipBlockComment()
			return l.NextToken()
		}
		// Determine if / starts a regex literal or is a division operator.
		// After a value-producing token (ident, number, string, ), ], ++, --),
		// / is division. Otherwise, / starts a regex.
		if l.canBeDivision() {
			if l.peekChar() == '=' {
				l.readChar()
				tok = Token{TokSlashEq, "/="}
			} else {
				tok = Token{TokSlash, "/"}
			}
		} else {
			// Read regex literal: /pattern/flags
			pattern, flags := l.readRegex()
			tok = Token{TokRegex, pattern + "/" + flags}
			l.prevTok = TokRegex
			return tok
		}
	case '%':
		tok = Token{TokPercent, "%"}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = Token{TokEqEqEq, "==="}
			} else {
				tok = Token{TokEqEq, "=="}
			}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = Token{TokArrow, "=>"}
		} else {
			tok = Token{TokEq, "="}
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = Token{TokNeqEq, "!=="}
			} else {
				tok = Token{TokNeq, "!="}
			}
		} else {
			tok = Token{TokNot, "!"}
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{TokLte, "<="}
		} else if l.peekChar() == '<' {
			l.readChar()
			tok = Token{TokShiftL, "<<"}
		} else {
			tok = Token{TokLt, "<"}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{TokGte, ">="}
		} else if l.peekChar() == '>' {
			l.readChar()
			if l.peekChar() == '>' {
				l.readChar()
				tok = Token{TokUShiftR, ">>>"}
			} else {
				tok = Token{TokShiftR, ">>"}
			}
		} else {
			tok = Token{TokGt, ">"}
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = Token{TokAnd, "&&"}
		} else {
			tok = Token{TokBitAnd, "&"}
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = Token{TokOr, "||"}
		} else {
			tok = Token{TokBitOr, "|"}
		}
	case '^':
		tok = Token{TokBitXor, "^"}
	case '~':
		tok = Token{TokBitNot, "~"}
	case '(':
		tok = Token{TokLParen, "("}
	case ')':
		tok = Token{TokRParen, ")"}
	case '{':
		tok = Token{TokLBrace, "{"}
	case '}':
		tok = Token{TokRBrace, "}"}
	case '[':
		tok = Token{TokLBracket, "["}
	case ']':
		tok = Token{TokRBracket, "]"}
	case ';':
		tok = Token{TokSemi, ";"}
	case ',':
		tok = Token{TokComma, ","}
	case '.':
		// Check for spread operator ...
		if l.peekChar() == '.' && l.peekPeekChar() == '.' {
			l.readChar()
			l.readChar()
			tok = Token{TokSpread, "..."}
		} else {
			tok = Token{TokDot, "."}
		}
	case '?':
		// Check for optional chaining ?. or nullish coalescing ??
		if l.peekChar() == '.' {
			l.readChar()
			tok = Token{TokOptChain, "?."}
		} else if l.peekChar() == '?' {
			l.readChar()
			tok = Token{TokNullish, "??"}
		} else {
			tok = Token{TokQuestion, "?"}
		}
	case ':':
		tok = Token{TokColon, ":"}
	case '"', '\'':
		tok = Token{TokString, l.readString(l.ch)}
		l.prevTok = TokString
		return tok
	case '`':
		tok = Token{TokTemplate, l.readTemplate()}
		l.prevTok = TokTemplate
		return tok
	default:
		if isDigit(l.ch) {
			num := l.readNumber()
			// Check if it's a BigInt (ends with 'n')
			if len(num) > 0 && num[len(num)-1] == 'n' {
				tok = Token{TokBigInt, num[:len(num)-1]}
				l.prevTok = TokBigInt
			} else {
				tok = Token{TokNumber, num}
				l.prevTok = TokNumber
			}
			return tok
		}
		if isIdentStart(l.ch) {
			ident := l.readIdent()
			if kw, ok := keywords[ident]; ok {
				tok = Token{kw, ident}
				l.prevTok = kw
			} else {
				tok = Token{TokIdent, ident}
				l.prevTok = TokIdent
			}
			return tok
		}
		tok = Token{TokIdent, string(l.ch)}
	}

	l.readChar()
	l.prevTok = tok.Type
	return tok
}

// canBeDivision returns true if the previous token context indicates that
// a / should be treated as a division operator rather than a regex literal start.
// After value-producing tokens (identifiers, numbers, strings, ), ], ++, --,
// template literals, etc.), / is division. After operators, (, [, {, ;, comma,
// and keywords like return/typeof/in/case, / starts a regex.
// This follows the ECMAScript specification for automatic semicolon insertion
// and regex/division disambiguation.
func (l *Lexer) canBeDivision() bool {
	switch l.prevTok {
	case TokIdent, TokNumber, TokString, TokBigInt, TokRegex, TokTemplate,
		TokRParen, TokRBracket, TokRBrace,
		TokInc, TokDec,
		TokGt, TokGte, TokLte:
		return true
	default:
		return false
	}
}

// readRegex reads a regex literal starting after the opening /.
// Returns (pattern, flags) where pattern does not include the delimiters.
// If no closing / is found before EOF, returns what was consumed and an
// empty flags string (graceful degradation rather than consuming the
// entire remaining input).
func (l *Lexer) readRegex() (string, string) {
	// Skip the opening / - l.ch is currently '/'
	l.readChar()
	// Now l.ch is the first char of the pattern, l.pos is past it
	start := l.pos - 1
	inCharClass := false
	closed := false
	for l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			if l.ch != 0 {
				l.readChar()
			}
			continue
		}
		if l.ch == '[' && !inCharClass {
			inCharClass = true
			l.readChar()
			continue
		}
		if l.ch == ']' && inCharClass {
			inCharClass = false
			l.readChar()
			continue
		}
		if l.ch == '/' && !inCharClass {
			closed = true
			break
		}
		l.readChar()
	}
	pattern := l.input[start : l.pos-1]

	// Read flags after closing /
	var flags string
	if closed {
		l.readChar() // skip closing /
		for isIdentStart(l.ch) || isDigit(l.ch) {
			flags += string(l.ch)
			l.readChar()
		}
	}

	return pattern, flags
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}
}

func (l *Lexer) skipBlockComment() {
	l.readChar() // skip '*'
	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			return
		}
		l.readChar()
	}
}

func (l *Lexer) readString(quote byte) string {
	l.readChar()
	start := l.pos - 1
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
		}
		l.readChar()
	}
	result := l.input[start : l.pos-1]
	l.readChar()
	return result
}

// readTemplate reads a template literal: `Hello ${name}!`
// Returns the raw template string with ${...} placeholders preserved
func (l *Lexer) readTemplate() string {
	l.readChar() // skip backtick
	start := l.pos - 1
	for l.ch != '`' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
		}
		l.readChar()
	}
	result := l.input[start : l.pos-1]
	l.readChar() // skip closing backtick
	return result
}

func (l *Lexer) readNumber() string {
	start := l.pos - 1
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	// Check for BigInt suffix 'n'
	if l.ch == 'n' {
		l.readChar()
		return l.input[start : l.pos-1]
	}
	return l.input[start : l.pos-1]
}

func (l *Lexer) readIdent() string {
	start := l.pos - 1
	for isIdentStart(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start : l.pos-1]
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$'
}
