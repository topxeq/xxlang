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
	TokenPlus           TokenType = "+"
	TokenMinus          TokenType = "-"
	TokenAsterisk       TokenType = "*"
	TokenSlash          TokenType = "/"
	TokenPercent        TokenType = "%"
	TokenAssign         TokenType = "="
	TokenPlusAssign     TokenType = "+="
	TokenMinusAssign    TokenType = "-="
	TokenAsteriskAssign TokenType = "*="
	TokenSlashAssign    TokenType = "/="
	TokenPercentAssign  TokenType = "%="
	TokenEqual          TokenType = "=="
	TokenNotEqual       TokenType = "!="
	TokenLT             TokenType = "<"
	TokenGT             TokenType = ">"
	TokenLTE            TokenType = "<="
	TokenGTE            TokenType = ">="
	TokenAnd            TokenType = "&&"
	TokenOr             TokenType = "||"
	TokenNot            TokenType = "!"
	TokenIncrement      TokenType = "++"
	TokenDecrement      TokenType = "--"

	// Delimiters
	TokenComma      TokenType = ","
	TokenColon      TokenType = ":"
	TokenColonAssign TokenType = ":="
	TokenSemicolon  TokenType = ";"
	TokenLParen    TokenType = "("
	TokenRParen    TokenType = ")"
	TokenLBrace    TokenType = "{"
	TokenRBrace    TokenType = "}"
	TokenLBracket  TokenType = "["
	TokenRBracket  TokenType = "]"
	TokenDot       TokenType = "."
	TokenArrow     TokenType = "=>"
	TokenQuestion  TokenType = "?"

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
	TokenSuper    TokenType = "SUPER"
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
	Type    TokenType
	Literal string
	Line    int
	Column  int
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
	"super":    TokenSuper,
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
