package jsengine

type TokenType int

const (
	TokEOF TokenType = iota
	TokIdent
	TokNumber
	TokString
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
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"var":       TokKeyword,
	"let":       TokKeyword,
	"const":     TokKeyword,
	"if":        TokKeyword,
	"else":      TokKeyword,
	"for":       TokKeyword,
	"while":     TokKeyword,
	"do":        TokKeyword,
	"break":     TokKeyword,
	"continue":  TokKeyword,
	"return":    TokKeyword,
	"function":  TokKeyword,
	"true":      TokKeyword,
	"false":     TokKeyword,
	"null":      TokKeyword,
	"undefined": TokKeyword,
	"new":       TokKeyword,
	"typeof":    TokKeyword,
	"delete":    TokKeyword,
	"in":        TokKeyword,
	"switch":    TokKeyword,
	"case":      TokKeyword,
	"default":   TokKeyword,
	"try":       TokKeyword,
	"catch":     TokKeyword,
	"finally":   TokKeyword,
	"throw":     TokKeyword,
}

type Lexer struct {
	input string
	pos   int
	ch    byte
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
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{TokSlashEq, "/="}
		} else {
			tok = Token{TokSlash, "/"}
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
		} else {
			tok = Token{TokLt, "<"}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{TokGte, ">="}
		} else {
			tok = Token{TokGt, ">"}
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = Token{TokAnd, "&&"}
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = Token{TokOr, "||"}
		}
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
		tok = Token{TokDot, "."}
	case '?':
		tok = Token{TokQuestion, "?"}
	case ':':
		tok = Token{TokColon, ":"}
	case '"', '\'':
		tok = Token{TokString, l.readString(l.ch)}
		return tok
	default:
		if isDigit(l.ch) {
			tok = Token{TokNumber, l.readNumber()}
			return tok
		}
		if isIdentStart(l.ch) {
			ident := l.readIdent()
			if kw, ok := keywords[ident]; ok {
				tok = Token{kw, ident}
			} else {
				tok = Token{TokIdent, ident}
			}
			return tok
		}
		tok = Token{TokIdent, string(l.ch)}
	}

	l.readChar()
	return tok
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

func (l *Lexer) readNumber() string {
	start := l.pos - 1
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
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
