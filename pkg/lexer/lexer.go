// pkg/lexer/lexer.go
package lexer

// Lexer tokenizes source code into tokens
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number (1-based)
	column       int  // current column number (1-based)
}

// New creates a new Lexer instance for the given input
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

	// Update line and column tracking
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekNextChar returns the character two positions ahead
func (l *Lexer) peekNextChar() byte {
	if l.readPosition+1 >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition+1]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()
	l.skipComment()

	// Save position for token
	line, column := l.line, l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenEqual, Literal: literal, Line: line, Column: column}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenArrow, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenAssign, l.ch, line, column)
		}
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenPlusAssign, Literal: literal, Line: line, Column: column}
		} else if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenIncrement, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenPlus, l.ch, line, column)
		}
	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenMinusAssign, Literal: literal, Line: line, Column: column}
		} else if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenDecrement, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenMinus, l.ch, line, column)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenNotEqual, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenNot, l.ch, line, column)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenLTE, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenLT, l.ch, line, column)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenGTE, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenGT, l.ch, line, column)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenAnd, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenIllegal, l.ch, line, column)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenOr, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenIllegal, l.ch, line, column)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenAsteriskAssign, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenAsterisk, l.ch, line, column)
		}
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenSlashAssign, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenSlash, l.ch, line, column)
		}
	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenPercentAssign, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenPercent, l.ch, line, column)
		}
	case ',':
		tok = newToken(TokenComma, l.ch, line, column)
	case ':':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenColonAssign, Literal: literal, Line: line, Column: column}
		} else {
			tok = newToken(TokenColon, l.ch, line, column)
		}
	case ';':
		tok = newToken(TokenSemicolon, l.ch, line, column)
	case '(':
		tok = newToken(TokenLParen, l.ch, line, column)
	case ')':
		tok = newToken(TokenRParen, l.ch, line, column)
	case '{':
		tok = newToken(TokenLBrace, l.ch, line, column)
	case '}':
		tok = newToken(TokenRBrace, l.ch, line, column)
	case '[':
		tok = newToken(TokenLBracket, l.ch, line, column)
	case ']':
		tok = newToken(TokenRBracket, l.ch, line, column)
	case '.':
		// Check for ellipsis (...)
		if l.peekChar() == '.' && l.peekNextChar() == '.' {
			l.readChar() // advance to second .
			l.readChar() // advance to third .
			tok = Token{Type: TokenEllipsis, Literal: "...", Line: line, Column: column}
		} else {
			tok = newToken(TokenDot, l.ch, line, column)
		}
	case '?':
		tok = newToken(TokenQuestion, l.ch, line, column)
	case '"':
		tok.Type = TokenString
		tok.Literal = l.readString()
		tok.Line = line
		tok.Column = column
		return tok // readString already advanced past closing quote, don't readChar again
	case '`':
		tok.Type = TokenString
		tok.Literal = l.readRawString()
		tok.Line = line
		tok.Column = column
		return tok // readRawString already advanced past closing backtick, don't readChar again
	case 0:
		tok.Literal = ""
		tok.Type = TokenEOF
		tok.Line = line
		tok.Column = column
	default:
		if isLetter(l.ch) || l.ch == '_' {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			tok.Line = line
			tok.Column = column
			return tok // readIdentifier already advanced, don't readChar again
		} else if isDigit(l.ch) {
			tok.Literal, tok.Type = l.readNumber()
			tok.Line = line
			tok.Column = column
			return tok // readNumber already advanced, don't readChar again
		} else {
			tok = newToken(TokenIllegal, l.ch, line, column)
		}
	}

	l.readChar()
	return tok
}

// newToken creates a new token with the given type and character
func newToken(tokenType TokenType, ch byte, line, column int) Token {
	return Token{
		Type:    tokenType,
		Literal: string(ch),
		Line:    line,
		Column:  column,
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number literal (integer or float)
func (l *Lexer) readNumber() (string, TokenType) {
	position := l.position
	tokenType := TokenInt

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = TokenFloat
		l.readChar() // read '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		tokenType = TokenFloat
		l.readChar() // read 'e' or 'E'
		if l.ch == '+' || l.ch == '-' {
			l.readChar() // read sign
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position], tokenType
}

// readString reads a string literal
func (l *Lexer) readString() string {
	var result []byte

	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			if l.ch == 0 {
				break
			}
			processed := processEscapeSequence(l.ch)
			result = append(result, processed)
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}

	l.readChar() // skip closing quote (or handle EOF)

	return string(result)
}

// readRawString reads a raw string literal (backtick string)
// No escape sequences are processed, newlines are preserved
func (l *Lexer) readRawString() string {
	var result []byte

	l.readChar() // skip opening backtick

	for l.ch != '`' && l.ch != 0 {
		result = append(result, l.ch)
		l.readChar()
	}

	l.readChar() // skip closing backtick (or handle EOF)

	return string(result)
}

// processEscapeSequence converts an escape sequence character to its actual value
func processEscapeSequence(ch byte) byte {
	switch ch {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	case '\\':
		return '\\'
	case '"':
		return '"'
	case '0':
		return 0
	default:
		return ch // return as-is if not recognized
	}
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment skips single-line and multi-line comments
func (l *Lexer) skipComment() {
	for {
		// Check for single-line comment
		if l.ch == '/' && l.peekChar() == '/' {
			l.skipSingleLineComment()
			l.skipWhitespace()
			continue
		}

		// Check for multi-line comment
		if l.ch == '/' && l.peekChar() == '*' {
			l.skipMultiLineComment()
			l.skipWhitespace()
			continue
		}

		// No more comments
		break
	}
}

// skipSingleLineComment skips a single-line comment
func (l *Lexer) skipSingleLineComment() {
	// Skip '//'
	l.readChar()
	l.readChar()

	// Read until end of line or EOF
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipMultiLineComment skips a multi-line comment
func (l *Lexer) skipMultiLineComment() {
	// Skip '/*'
	l.readChar()
	l.readChar()

	// Read until '*/' or EOF
	for {
		if l.ch == 0 {
			break // Unterminated comment
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip '*'
			l.readChar() // skip '/'
			break
		}
		l.readChar()
	}
}

// isLetter checks if a character is a letter
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
