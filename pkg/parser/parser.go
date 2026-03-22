// pkg/parser/parser.go
package parser

import (
	"fmt"
	"strconv"

	"github.com/topxeq/xxlang/pkg/lexer"
)

// Operator precedence constants
const (
	_           int = iota
	LOWEST          // lowest precedence for expression parsing
	TERNARY         // ? : (right-associative)
	ASSIGN          // =, +=, -= (right-associative, binds after LOWEST)
	OR              // ||
	AND             // &&
	EQUALS          // ==, !=
	LESSGREATER     // <, >, <=, >=
	SUM             // +, -
	PRODUCT         // *, /, %
	PREFIX          // !, -
	CALL            // fn()
	INDEX           // arr[i]
	DOT             // obj.field
)

// precedence maps token types to their precedence
var precedence = map[lexer.TokenType]int{
	lexer.TokenOr:             OR,
	lexer.TokenAnd:            AND,
	lexer.TokenEqual:          EQUALS,
	lexer.TokenNotEqual:       EQUALS,
	lexer.TokenLT:             LESSGREATER,
	lexer.TokenGT:             LESSGREATER,
	lexer.TokenLTE:            LESSGREATER,
	lexer.TokenGTE:            LESSGREATER,
	lexer.TokenPlus:           SUM,
	lexer.TokenMinus:          SUM,
	lexer.TokenAsterisk:       PRODUCT,
	lexer.TokenSlash:          PRODUCT,
	lexer.TokenPercent:        PRODUCT,
	lexer.TokenLParen:         CALL,
	lexer.TokenLBracket:       INDEX,
	lexer.TokenDot:            DOT,
	lexer.TokenAssign:         ASSIGN,
	lexer.TokenPlusAssign:     ASSIGN,
	lexer.TokenMinusAssign:    ASSIGN,
	lexer.TokenAsteriskAssign: ASSIGN,
	lexer.TokenSlashAssign:    ASSIGN,
	lexer.TokenPercentAssign:  ASSIGN,
	lexer.TokenIncrement:      CALL,    // Postfix ++ has high precedence
	lexer.TokenDecrement:      CALL,    // Postfix -- has high precedence
	lexer.TokenQuestion:       TERNARY, // Ternary operator ?:
}

// Parser parses tokens into an AST
type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

// New creates a new Parser instance
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[lexer.TokenType]prefixParseFn),
		infixParseFns:  make(map[lexer.TokenType]infixParseFn),
	}

	// Register prefix parse functions
	p.registerPrefix(lexer.TokenIdent, p.parseIdentifier)
	p.registerPrefix(lexer.TokenInt, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TokenFloat, p.parseFloatLiteral)
	p.registerPrefix(lexer.TokenString, p.parseStringLiteral)
	p.registerPrefix(lexer.TokenTrue, p.parseBooleanLiteral)
	p.registerPrefix(lexer.TokenFalse, p.parseBooleanLiteral)
	p.registerPrefix(lexer.TokenNull, p.parseNullLiteral)
	p.registerPrefix(lexer.TokenNot, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenMinus, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenLParen, p.parseGroupedExpression)
	p.registerPrefix(lexer.TokenLBracket, p.parseArrayLiteral)
	p.registerPrefix(lexer.TokenLBrace, p.parseMapLiteral)
	p.registerPrefix(lexer.TokenFunc, p.parseFunctionLiteral)
	p.registerPrefix(lexer.TokenIncrement, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenDecrement, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenNew, p.parseNewExpression)
	p.registerPrefix(lexer.TokenThis, p.parseThisExpression)
	p.registerPrefix(lexer.TokenSuper, p.parseSuperExpression)

	// Register infix parse functions
	p.registerInfix(lexer.TokenPlus, p.parseInfixExpression)
	p.registerInfix(lexer.TokenMinus, p.parseInfixExpression)
	p.registerInfix(lexer.TokenAsterisk, p.parseInfixExpression)
	p.registerInfix(lexer.TokenSlash, p.parseInfixExpression)
	p.registerInfix(lexer.TokenPercent, p.parseInfixExpression)
	p.registerInfix(lexer.TokenEqual, p.parseInfixExpression)
	p.registerInfix(lexer.TokenNotEqual, p.parseInfixExpression)
	p.registerInfix(lexer.TokenLT, p.parseInfixExpression)
	p.registerInfix(lexer.TokenGT, p.parseInfixExpression)
	p.registerInfix(lexer.TokenLTE, p.parseInfixExpression)
	p.registerInfix(lexer.TokenGTE, p.parseInfixExpression)
	p.registerInfix(lexer.TokenAnd, p.parseInfixExpression)
	p.registerInfix(lexer.TokenOr, p.parseInfixExpression)
	p.registerInfix(lexer.TokenLParen, p.parseCallExpression)
	p.registerInfix(lexer.TokenLBracket, p.parseIndexExpression)
	p.registerInfix(lexer.TokenDot, p.parseDotExpression)
	p.registerInfix(lexer.TokenAssign, p.parseAssignmentExpression)
	p.registerInfix(lexer.TokenPlusAssign, p.parseCompoundAssignmentExpression)
	p.registerInfix(lexer.TokenMinusAssign, p.parseCompoundAssignmentExpression)
	p.registerInfix(lexer.TokenAsteriskAssign, p.parseCompoundAssignmentExpression)
	p.registerInfix(lexer.TokenSlashAssign, p.parseCompoundAssignmentExpression)
	p.registerInfix(lexer.TokenPercentAssign, p.parseCompoundAssignmentExpression)
	p.registerInfix(lexer.TokenIncrement, p.parsePostfixExpression)
	p.registerInfix(lexer.TokenDecrement, p.parsePostfixExpression)
	p.registerInfix(lexer.TokenQuestion, p.parseTernaryExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// registerPrefix registers a prefix parse function
func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers an infix parse function
func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// nextToken advances the parser to the next token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs checks if the current token is of the given type
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the peek token is of the given type
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the peek token is of the given type and advances if so
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError adds an error for unexpected peek token
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d:%d: expected next token to be %s, got %s instead",
		p.curToken.Line, p.curToken.Column, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// addError adds an error message
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

// Errors returns all parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

// curPrecedence returns the precedence of the current token
func (p *Parser) curPrecedence() int {
	if p, ok := precedence[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// peekPrecedence returns the precedence of the peek token
func (p *Parser) peekPrecedence() int {
	if p, ok := precedence[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *Program {
	program := &Program{
		Statements: []Statement{},
	}

	for !p.curTokenIs(lexer.TokenEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement parses a statement
func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case lexer.TokenVar:
		return p.parseVarStatement()
	case lexer.TokenConst:
		return p.parseConstStatement()
	case lexer.TokenReturn:
		return p.parseReturnStatement()
	case lexer.TokenIf:
		return p.parseIfStatement()
	case lexer.TokenWhile:
		return p.parseWhileStatement()
	case lexer.TokenFor:
		return p.parseForStatement()
	case lexer.TokenSwitch:
		return p.parseSwitchStatement()
	case lexer.TokenBreak:
		return p.parseBreakStatement()
	case lexer.TokenContinue:
		return p.parseContinueStatement()
	case lexer.TokenImport:
		return p.parseImportStatement()
	case lexer.TokenExport:
		return p.parseExportStatement()
	case lexer.TokenClass:
		return p.parseClassStatement()
	case lexer.TokenTry:
		return p.parseTryStatement()
	case lexer.TokenThrow:
		return p.parseThrowStatement()
	case lexer.TokenLBrace:
		// Check if this is a map literal or block statement
		// Empty {} is treated as map literal
		// { followed by expression-starting tokens that form a key:value pair is a map
		// Otherwise it's a block statement
		if p.peekTokenIs(lexer.TokenRBrace) {
			// Empty map literal
			return p.parseExpressionStatement()
		}
		// Check if it looks like a map literal (has string/int/ident key followed by colon)
		// Common map key tokens: STRING, INT, FLOAT, IDENT, TRUE, FALSE, NULL
		if p.peekTokenIs(lexer.TokenString) ||
			p.peekTokenIs(lexer.TokenInt) ||
			p.peekTokenIs(lexer.TokenFloat) ||
			p.peekTokenIs(lexer.TokenIdent) ||
			p.peekTokenIs(lexer.TokenTrue) ||
			p.peekTokenIs(lexer.TokenFalse) ||
			p.peekTokenIs(lexer.TokenNull) {
			// This is likely a map literal - parse as expression statement
			return p.parseExpressionStatement()
		}
		// Standalone block statement
		return p.parseBlockStatement()
	case lexer.TokenIdent:
		// Check for short variable declaration: ident := value
		if p.peekTokenIs(lexer.TokenColonAssign) {
			return p.parseShortVarStatement()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseVarStatement parses a var statement
func (p *Parser) parseVarStatement() *VarStatement {
	stmt := &VarStatement{Token: p.curToken}

	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.TokenAssign) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseShortVarStatement parses a short variable declaration (:=)
func (p *Parser) parseShortVarStatement() *ShortVarStatement {
	stmt := &ShortVarStatement{Token: p.curToken}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Skip the ':=' token
	p.nextToken()

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseConstStatement parses a const statement
func (p *Parser) parseConstStatement() *ConstStatement {
	stmt := &ConstStatement{Token: p.curToken}

	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.TokenAssign) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement parses a return statement
func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}

	// Check if there's a return value (not followed by semicolon or end of block)
	if p.peekTokenIs(lexer.TokenSemicolon) || p.peekTokenIs(lexer.TokenRBrace) {
		// Return without value
		if p.peekTokenIs(lexer.TokenSemicolon) {
			p.nextToken()
		}
		return stmt
	}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseExpressionStatement parses an expression statement
func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseBlockStatement parses a block statement
func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	p.nextToken()

	for !p.curTokenIs(lexer.TokenRBrace) && !p.curTokenIs(lexer.TokenEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseIfStatement parses an if statement
func (p *Parser) parseIfStatement() *IfStatement {
	stmt := &IfStatement{Token: p.curToken}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(lexer.TokenElse) {
		p.nextToken()

		// Check if this is "else if" - if followed by 'if' keyword
		if p.peekTokenIs(lexer.TokenIf) {
			p.nextToken()
			// Create a block statement containing the if statement
			elseIfStmt := p.parseIfStatement()
			if elseIfStmt == nil {
				return nil
			}
			stmt.Alternative = &BlockStatement{
				Token:      p.curToken,
				Statements: []Statement{elseIfStmt},
			}
		} else {
			// Regular "else { ... }"
			if !p.expectPeek(lexer.TokenLBrace) {
				return nil
			}

			stmt.Alternative = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseWhileStatement parses a while statement
func (p *Parser) parseWhileStatement() *WhileStatement {
	stmt := &WhileStatement{Token: p.curToken}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForStatement parses a for statement (C-style or for-in)
func (p *Parser) parseForStatement() Statement {
	// We need to look ahead to determine if this is a for-in or C-style for loop
	// for (x in arr) or for (k, v in arr) vs for (init; cond; update)

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	p.nextToken()

	// Check if this is a for-in loop by looking at the pattern
	// Pattern 1: ident "in" -> for (x in arr)
	// Pattern 2: ident "," ident "in" -> for (k, v in arr)
	// Otherwise it's a C-style for loop

	// Check if it could be a for-in loop (starts with an identifier, not var/const/etc)
	if p.curTokenIs(lexer.TokenIdent) {
		firstIdent := p.curToken.Literal

		p.nextToken()

		if p.curTokenIs(lexer.TokenIn) {
			// This is for (x in arr)
			return p.parseForInStatementSingle(firstIdent)
		} else if p.curTokenIs(lexer.TokenComma) {
			// This might be for (k, v in arr)
			p.nextToken()
			if p.curTokenIs(lexer.TokenIdent) {
				secondIdent := p.curToken.Literal
				p.nextToken()
				if p.curTokenIs(lexer.TokenIn) {
					return p.parseForInStatementKeyValue(firstIdent, secondIdent)
				}
			}
			// If we get here, it's not a valid for-in
			p.addError(fmt.Sprintf("line %d:%d: expected 'in' after identifiers in for-in loop",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		// It's a C-style for loop with an identifier as init
		// Current token is after the identifier, need to go back
		return p.parseCStyleForStatementWithIdentifier(firstIdent)
	}

	// Not a for-in loop, parse as C-style for loop
	return p.parseCStyleForStatementWithInit()
}

// parseForInStatementSingle parses for (x in arr)
func (p *Parser) parseForInStatementSingle(valueIdent string) *ForInStatement {
	stmt := &ForInStatement{Token: lexer.Token{Type: lexer.TokenFor, Literal: "for"}}

	stmt.Value = &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: valueIdent}, Value: valueIdent}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForInStatementKeyValue parses for (k, v in arr)
func (p *Parser) parseForInStatementKeyValue(keyIdent, valueIdent string) *ForInStatement {
	stmt := &ForInStatement{Token: lexer.Token{Type: lexer.TokenFor, Literal: "for"}}

	stmt.Key = &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: keyIdent}, Value: keyIdent}
	stmt.Value = &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: valueIdent}, Value: valueIdent}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseCStyleForStatement parses for (init; cond; update) { ... }
func (p *Parser) parseCStyleForStatement() *ForStatement {
	stmt := &ForStatement{Token: p.curToken}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	p.nextToken()

	// Parse init (can be var declaration or expression, ends with semicolon)
	if !p.curTokenIs(lexer.TokenSemicolon) {
		stmt.Init = p.parseStatement()
		// After parseStatement, we might have consumed the semicolon already
		// if it was a var statement, or we need to skip it
	}

	// Make sure we're past the first semicolon
	if p.curTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	// Parse condition (expression, ends with semicolon)
	if !p.curTokenIs(lexer.TokenSemicolon) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.TokenSemicolon) {
		return nil
	}

	p.nextToken()

	// Parse update (expression, ends with right paren)
	if !p.curTokenIs(lexer.TokenRParen) {
		stmt.Update = p.parseExpressionStatement()
	}

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseCStyleForStatementWithInit parses C-style for loop when we've already consumed the opening paren
// and the current token is the first token inside the parens (not an identifier that starts for-in).
func (p *Parser) parseCStyleForStatementWithInit() *ForStatement {
	stmt := &ForStatement{Token: lexer.Token{Type: lexer.TokenFor, Literal: "for"}}

	// At this point, current token is the first token after '('
	// This could be 'var', an expression, or ';'

	// Parse init expression/statement
	if !p.curTokenIs(lexer.TokenSemicolon) {
		// Parse as a statement (could be var declaration or expression statement)
		stmt.Init = p.parseStatement()
	}

	// After parsing init, we need to be at the semicolon
	// parseStatement for var leaves curToken at the semicolon it consumed
	// parseExpressionStatement leaves curToken at the semicolon it consumed
	if p.curTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	} else if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken() // move to semicolon
		p.nextToken() // move past semicolon to condition or next part
	}

	// Parse condition (if not empty)
	if !p.curTokenIs(lexer.TokenSemicolon) && !p.curTokenIs(lexer.TokenRParen) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	// After parsing condition, move past the semicolon
	if p.curTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	} else if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken() // move to semicolon
		p.nextToken() // move past semicolon to update or rparen
	}

	// Parse update (if not empty)
	if !p.curTokenIs(lexer.TokenRParen) {
		stmt.Update = p.parseExpressionStatement()
	}

	// After parsing update (or if empty), we should be at ')'
	// parseExpressionStatement leaves curToken at the last token of expression
	// So we need to check if we're at ')' or need to advance to it
	if p.curTokenIs(lexer.TokenRParen) {
		// Already at ')', just advance to '{'
	} else if p.peekTokenIs(lexer.TokenRParen) {
		p.nextToken() // move to ')'
	} else if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseCStyleForStatementWithIdentifier parses C-style for loop when we've seen an identifier
// but it wasn't followed by 'in' (so it's not a for-in loop)
func (p *Parser) parseCStyleForStatementWithIdentifier(ident string) *ForStatement {
	stmt := &ForStatement{Token: lexer.Token{Type: lexer.TokenFor, Literal: "for"}}

	// The current token is after the identifier
	// We need to parse the rest of the init expression
	// Build an expression starting with the identifier

	identExpr := &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: ident}, Value: ident}

	// Check if this is an assignment: ident = value
	if p.curTokenIs(lexer.TokenAssign) || p.curTokenIs(lexer.TokenPlusAssign) || p.curTokenIs(lexer.TokenMinusAssign) {
		// Parse as assignment
		p.nextToken()
		value := p.parseExpression(LOWEST)

		var initExpr Expression
		if p.curTokenIs(lexer.TokenAssign) {
			// We already moved past the operator, need to use previous token
			// This is a bit messy, let's simplify
			initExpr = &AssignmentExpression{
				Token: lexer.Token{Type: lexer.TokenAssign, Literal: "="},
				Left:  identExpr,
				Value: value,
			}
		} else {
			initExpr = &AssignmentExpression{
				Token: lexer.Token{Type: lexer.TokenAssign, Literal: "="},
				Left:  identExpr,
				Value: value,
			}
		}
		stmt.Init = &ExpressionStatement{Token: lexer.Token{Type: lexer.TokenIdent, Literal: ident}, Expression: initExpr}
	} else {
		// Just the identifier as an expression statement
		stmt.Init = &ExpressionStatement{Token: lexer.Token{Type: lexer.TokenIdent, Literal: ident}, Expression: identExpr}
	}

	// Skip semicolon after init
	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}
	p.nextToken()

	// Parse condition
	if !p.curTokenIs(lexer.TokenSemicolon) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.TokenSemicolon) {
		return nil
	}

	p.nextToken()

	// Parse update
	if !p.curTokenIs(lexer.TokenRParen) {
		stmt.Update = p.parseExpressionStatement()
	}

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseBreakStatement parses a break statement
func (p *Parser) parseBreakStatement() *BreakStatement {
	stmt := &BreakStatement{Token: p.curToken}

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseContinueStatement parses a continue statement
func (p *Parser) parseContinueStatement() *ContinueStatement {
	stmt := &ContinueStatement{Token: p.curToken}

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseImportStatement parses an import statement
// Supports:
//   - import "./math" - Simple import (just load module)
//   - import math from "./math" - Default import (module stored as `math`)
//   - import { add, sub } from "./math" - Destructuring import
//   - import * as math from "./math" - Namespace import
func (p *Parser) parseImportStatement() *ImportStatement {
	stmt := &ImportStatement{Token: p.curToken}

	p.nextToken()

	// Check what kind of import this is
	if p.curTokenIs(lexer.TokenString) {
		// Simple import: import "./math"
		stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(lexer.TokenAsterisk) {
		// Namespace import: import * as math from "./math"
		p.nextToken()
		if !p.curTokenIs(lexer.TokenIdent) || p.curToken.Literal != "as" {
			p.addError(fmt.Sprintf("line %d:%d: expected 'as' after '*', got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Literal))
			return nil
		}
		p.nextToken()
		if !p.curTokenIs(lexer.TokenIdent) {
			p.addError(fmt.Sprintf("line %d:%d: expected identifier after 'as', got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		stmt.Alias = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		// Expect 'from'
		if !p.curTokenIs(lexer.TokenIdent) || p.curToken.Literal != "from" {
			p.addError(fmt.Sprintf("line %d:%d: expected 'from', got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Literal))
			return nil
		}
		p.nextToken()
		if !p.curTokenIs(lexer.TokenString) {
			p.addError(fmt.Sprintf("line %d:%d: expected string path, got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(lexer.TokenLBrace) {
		// Destructuring import: import { add, sub } from "./math"
		p.nextToken()
		stmt.Names = []*Identifier{}
		for !p.curTokenIs(lexer.TokenRBrace) {
			if !p.curTokenIs(lexer.TokenIdent) {
				p.addError(fmt.Sprintf("line %d:%d: expected identifier in import list, got %s",
					p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
			stmt.Names = append(stmt.Names, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
			p.nextToken()
			if p.curTokenIs(lexer.TokenComma) {
				p.nextToken()
			} else if !p.curTokenIs(lexer.TokenRBrace) {
				p.addError(fmt.Sprintf("line %d:%d: expected ',' or '}' in import list, got %s",
					p.curToken.Line, p.curToken.Column, p.curToken.Type))
				return nil
			}
		}
		p.nextToken()
		// Expect 'from'
		if !p.curTokenIs(lexer.TokenIdent) || p.curToken.Literal != "from" {
			p.addError(fmt.Sprintf("line %d:%d: expected 'from', got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Literal))
			return nil
		}
		p.nextToken()
		if !p.curTokenIs(lexer.TokenString) {
			p.addError(fmt.Sprintf("line %d:%d: expected string path, got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(lexer.TokenIdent) {
		// Default import: import math from "./math"
		// Could also be the 'from' keyword if this is a side-effect only import with name
		if p.curToken.Literal == "from" {
			// This shouldn't happen - import from "./math" without binding name
			p.addError(fmt.Sprintf("line %d:%d: expected identifier or path after import",
				p.curToken.Line, p.curToken.Column))
			return nil
		}
		stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		// Expect 'from'
		if !p.curTokenIs(lexer.TokenIdent) || p.curToken.Literal != "from" {
			p.addError(fmt.Sprintf("line %d:%d: expected 'from', got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Literal))
			return nil
		}
		p.nextToken()
		if !p.curTokenIs(lexer.TokenString) {
			p.addError(fmt.Sprintf("line %d:%d: expected string path, got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	} else {
		p.addError(fmt.Sprintf("line %d:%d: unexpected token after import: %s",
			p.curToken.Line, p.curToken.Column, p.curToken.Type))
		return nil
	}

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

// parseExportStatement parses an export statement
// Supports:
//   - export func add(a, b) { return a + b } - Export function
//   - export var PI = 3.14 - Export variable
//   - export const VERSION = "1.0" - Export constant
func (p *Parser) parseExportStatement() *ExportStatement {
	stmt := &ExportStatement{Token: p.curToken}

	p.nextToken()

	// Parse the statement to export (var, const, func, etc.)
	var exportable Statement
	switch p.curToken.Type {
	case lexer.TokenVar:
		exportable = p.parseVarStatement()
	case lexer.TokenConst:
		exportable = p.parseConstStatement()
	case lexer.TokenFunc:
		exportable = p.parseExpressionStatement()
	default:
		p.addError(fmt.Sprintf("line %d:%d: cannot export %s, expected var, const, or func",
			p.curToken.Line, p.curToken.Column, p.curToken.Type))
		return nil
	}

	if exportable == nil {
		return nil
	}

	stmt.Exportable = exportable
	return stmt
}

// parseClassStatement parses a class declaration
func (p *Parser) parseClassStatement() *ClassStatement {
	stmt := &ClassStatement{Token: p.curToken}

	// Expect class name
	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for inheritance: supports both "extends" and ":" syntax
	// class Dog extends Animal { ... }
	// class Dog : Animal { ... }
	if p.peekTokenIs(lexer.TokenExtends) || p.peekTokenIs(lexer.TokenColon) {
		p.nextToken() // Move to 'extends' or ':'
		if !p.expectPeek(lexer.TokenIdent) {
			return nil
		}
		stmt.SuperClass = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// Expect {
	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	// Parse class body
	stmt.Methods = []*FunctionLiteral{}
	stmt.Fields = []*VarStatement{}

	p.nextToken() // Advance past { to first token in body

	for !p.curTokenIs(lexer.TokenRBrace) && !p.curTokenIs(lexer.TokenEOF) {
		if p.curTokenIs(lexer.TokenFunc) {
			method := p.parseFunctionLiteral()
			if method != nil {
				if fn, ok := method.(*FunctionLiteral); ok {
					stmt.Methods = append(stmt.Methods, fn)
				}
			}
			// After parseFunctionLiteral, curToken is at the function's closing }
			// Advance past it to continue parsing the class body
			if p.curTokenIs(lexer.TokenRBrace) {
				p.nextToken()
			}
		} else if p.curTokenIs(lexer.TokenVar) {
			field := p.parseVarStatement()
			if field != nil {
				stmt.Fields = append(stmt.Fields, field)
			}
			// After parseVarStatement, curToken is at the last token of the value
			// Skip any semicolon and advance to the next token
			if p.peekTokenIs(lexer.TokenSemicolon) {
				p.nextToken() // Move to semicolon
			}
			p.nextToken() // Move past semicolon (or value) to next token
		} else {
			p.addError(fmt.Sprintf("unexpected token in class body: %s", p.curToken.Type))
			return nil
		}

		// Skip any trailing semicolons
		if p.curTokenIs(lexer.TokenSemicolon) {
			p.nextToken()
		}
	}

	return stmt
}

// parseSwitchStatement parses a switch statement
//
//	switch (expression) {
//	    case value1:
//	        statements
//	    case value2:
//	        statements
//	    default:
//	        statements
//	}
func (p *Parser) parseSwitchStatement() *SwitchStatement {
	stmt := &SwitchStatement{Token: p.curToken}

	// Expect opening parenthesis for switch expression
	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	p.nextToken()

	// Parse cases and default
	stmt.Cases = []*CaseStatement{}

	for !p.curTokenIs(lexer.TokenRBrace) && !p.curTokenIs(lexer.TokenEOF) {
		if p.curTokenIs(lexer.TokenCase) {
			caseStmt := p.parseCaseStatement()
			if caseStmt != nil {
				stmt.Cases = append(stmt.Cases, caseStmt)
			}
			// After parsing, curToken is at next case/default or }
		} else if p.curTokenIs(lexer.TokenDefault) {
			if stmt.Default != nil {
				p.addError(fmt.Sprintf("line %d:%d: multiple default clauses in switch statement",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
			stmt.Default = p.parseDefaultStatement()
			// After parsing, curToken is at next case or }
			// If we see another case after default, that's an error
			if p.curTokenIs(lexer.TokenCase) {
				p.addError(fmt.Sprintf("line %d:%d: 'case' cannot appear after 'default'",
					p.curToken.Line, p.curToken.Column))
				return nil
			}
		} else {
			p.addError(fmt.Sprintf("line %d:%d: expected 'case' or 'default', got %s",
				p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
	}

	return stmt
}

// parseCaseStatement parses a case statement in a switch
func (p *Parser) parseCaseStatement() *CaseStatement {
	stmt := &CaseStatement{Token: p.curToken}

	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	// Parse case body - statements until we hit case, default, or }
	stmt.Consequence = p.parseCaseBody()

	return stmt
}

// parseDefaultStatement parses a default statement in a switch
func (p *Parser) parseDefaultStatement() *DefaultStatement {
	stmt := &DefaultStatement{Token: p.curToken}

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	// Parse default body - statements until we hit case (error) or }
	stmt.Consequence = p.parseCaseBody()

	return stmt
}

// parseCaseBody parses statements inside a case/default until hitting case, default, or }
func (p *Parser) parseCaseBody() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	// Advance to first statement in the case body
	p.nextToken()

	for !p.curTokenIs(lexer.TokenCase) && !p.curTokenIs(lexer.TokenDefault) &&
		!p.curTokenIs(lexer.TokenRBrace) && !p.curTokenIs(lexer.TokenEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseTryStatement parses a try-catch-finally statement
//
//	try {
//	    statements
//	} catch (e) {
//
//	    statements
//	} finally {
//
//	    statements
//	}
func (p *Parser) parseTryStatement() *TryStatement {
	stmt := &TryStatement{Token: p.curToken}

	// Expect '{' for try block
	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Block = p.parseBlockStatement()

	// Check for catch
	if p.peekTokenIs(lexer.TokenCatch) {
		p.nextToken()
		stmt.Catch = p.parseCatchStatement()
		if stmt.Catch == nil {
			return nil
		}
	}

	// Check for finally
	if p.peekTokenIs(lexer.TokenFinally) {
		p.nextToken()
		stmt.Finally = p.parseFinallyStatement()
		if stmt.Finally == nil {
			return nil
		}
	}

	// Must have at least catch or finally
	if stmt.Catch == nil && stmt.Finally == nil {
		p.addError(fmt.Sprintf("line %d:%d: try statement must have catch or finally clause",
			stmt.Token.Line, stmt.Token.Column))
		return nil
	}

	return stmt
}

// parseCatchStatement parses a catch clause
func (p *Parser) parseCatchStatement() *CatchStatement {
	stmt := &CatchStatement{Token: p.curToken}

	// Expect '(' for exception variable
	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	// Expect identifier for exception variable
	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}

	stmt.Exception = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect ')'
	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	// Expect '{' for catch block
	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Block = p.parseBlockStatement()

	return stmt
}

// parseFinallyStatement parses a finally clause
func (p *Parser) parseFinallyStatement() *FinallyStatement {
	stmt := &FinallyStatement{Token: p.curToken}

	// Expect '{' for finally block
	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	stmt.Block = p.parseBlockStatement()

	return stmt
}

// parseThrowStatement parses a throw statement
func (p *Parser) parseThrowStatement() *ThrowStatement {
	stmt := &ThrowStatement{Token: p.curToken}

	// Check if there's an expression after throw
	if p.peekTokenIs(lexer.TokenSemicolon) {
		// Throw without value - will throw null, consume semicolon
		p.nextToken()
		return stmt
	}

	if p.peekTokenIs(lexer.TokenRBrace) {
		// Throw without value at end of block - don't consume the }
		return stmt
	}

	p.nextToken()
	stmt.ErrExpr = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TokenSemicolon) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.TokenSemicolon) && precedence <= p.peekPrecedence() {
		// For postfix ++ and --, only parse as postfix if on the same line
		// This prevents: pln("hello")\n++x  from being parsed as (pln("hello"))++
		if p.peekToken.Type == lexer.TokenIncrement || p.peekToken.Type == lexer.TokenDecrement {
			if p.peekToken.Line != p.curToken.Line {
				// ++ or -- is on a different line, don't parse as postfix
				return leftExp
			}
		}

		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// noPrefixParseFnError adds an error for missing prefix parse function
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d:%d: no prefix parse function for %s found",
		p.curToken.Line, p.curToken.Column, t)
	p.errors = append(p.errors, msg)
}

// parseIdentifier parses an identifier
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral parses an integer literal
func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d:%d: could not parse %q as integer",
			p.curToken.Line, p.curToken.Column, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses a float literal
func (p *Parser) parseFloatLiteral() Expression {
	lit := &FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d:%d: could not parse %q as float",
			p.curToken.Line, p.curToken.Column, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal
func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseBooleanLiteral parses a boolean literal
func (p *Parser) parseBooleanLiteral() Expression {
	return &BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.TokenTrue)}
}

// parseNullLiteral parses a null literal
func (p *Parser) parseNullLiteral() Expression {
	return &NullLiteral{Token: p.curToken}
}

// parseGroupedExpression parses a grouped expression (expr)
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	return exp
}

// parsePrefixExpression parses a prefix expression
func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression parses an infix expression
func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence + 1) // +1 for left associativity

	return expression
}

// parsePostfixExpression parses a postfix expression (++, --)
func (p *Parser) parsePostfixExpression(left Expression) Expression {
	return &PostfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}
}

// parseTernaryExpression parses a ternary expression (condition ? consequent : alternative)
func (p *Parser) parseTernaryExpression(condition Expression) Expression {
	expression := &TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}

	// Skip '?' token
	p.nextToken()

	// Parse consequent (use TERNARY-1 for right associativity)
	expression.Consequent = p.parseExpression(TERNARY - 1)

	// Expect ':'
	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	// Skip ':' token
	p.nextToken()

	// Parse alternative (right associative, so use TERNARY-1)
	expression.Alternative = p.parseExpression(TERNARY - 1)

	return expression
}

// parseCallExpression parses a function call expression
func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(lexer.TokenRParen)
	return exp
}

// parseIndexExpression parses an index or slice expression
func (p *Parser) parseIndexExpression(left Expression) Expression {
	token := p.curToken // The '[' token

	p.nextToken()

	// Check for empty start ([:end] syntax)
	if p.curTokenIs(lexer.TokenColon) {
		// Slice with empty start: [:end]
		sliceExp := &SliceExpression{Token: token, Left: left, Start: nil}

		p.nextToken() // Move past the colon

		// Check if there's an end expression
		if !p.curTokenIs(lexer.TokenRBracket) {
			sliceExp.End = p.parseExpression(LOWEST)
		}

		if !p.expectPeek(lexer.TokenRBracket) {
			return nil
		}

		return sliceExp
	}

	// Parse the first expression (could be index or slice start)
	firstExpr := p.parseExpression(LOWEST)

	// Check if this is a slice expression (has a colon)
	if p.peekTokenIs(lexer.TokenColon) {
		sliceExp := &SliceExpression{Token: token, Left: left, Start: firstExpr}

		p.nextToken() // Move to the colon
		p.nextToken() // Move past the colon

		// Check if there's an end expression
		if !p.curTokenIs(lexer.TokenRBracket) {
			sliceExp.End = p.parseExpression(LOWEST)
		}

		if !p.expectPeek(lexer.TokenRBracket) {
			return nil
		}

		return sliceExp
	}

	// Regular index expression
	exp := &IndexExpression{Token: token, Left: left, Index: firstExpr}

	if !p.expectPeek(lexer.TokenRBracket) {
		return nil
	}

	return exp
}

// parseDotExpression parses a dot expression
func (p *Parser) parseDotExpression(object Expression) Expression {
	exp := &DotExpression{Token: p.curToken, Object: object}

	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}

	exp.Property = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}

// parseAssignmentExpression parses an assignment expression
func (p *Parser) parseAssignmentExpression(left Expression) Expression {
	exp := &AssignmentExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Value = p.parseExpression(LOWEST)

	return exp
}

// parseCompoundAssignmentExpression parses compound assignment (+=, -=)
func (p *Parser) parseCompoundAssignmentExpression(left Expression) Expression {
	exp := &CompoundAssignmentExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	exp.Right = p.parseExpression(LOWEST)

	return exp
}

// parseArrayLiteral parses an array literal
func (p *Parser) parseArrayLiteral() Expression {
	array := &ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(lexer.TokenRBracket)

	return array
}

// parseMapLiteral parses a map literal
func (p *Parser) parseMapLiteral() Expression {
	mapLit := &MapLiteral{Token: p.curToken}
	mapLit.Pairs = make(map[Expression]Expression)

	// Handle empty map: {}
	if p.peekTokenIs(lexer.TokenRBrace) {
		p.nextToken()
		return mapLit
	}

	// Parse first key-value pair
	p.nextToken()
	key := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)

	mapLit.Pairs[key] = value

	// Parse remaining key-value pairs
	for p.peekTokenIs(lexer.TokenComma) {
		p.nextToken() // move to comma
		p.nextToken() // move past comma to next key

		key := p.parseExpression(LOWEST)

		if !p.expectPeek(lexer.TokenColon) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		mapLit.Pairs[key] = value
	}

	if !p.expectPeek(lexer.TokenRBrace) {
		return nil
	}

	return mapLit
}

// parseFunctionLiteral parses a function literal
func (p *Parser) parseFunctionLiteral() Expression {
	lit := &FunctionLiteral{Token: p.curToken}

	// Check for named function: func name() { ... }
	if p.peekTokenIs(lexer.TokenIdent) {
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	lit.Parameters, lit.VariadicParam = p.parseFunctionParameters()

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters parses function parameters, including optional variadic parameter
// Returns regular parameters and optional variadic parameter
func (p *Parser) parseFunctionParameters() ([]*Identifier, *Identifier) {
	identifiers := []*Identifier{}
	var variadic *Identifier

	if p.peekTokenIs(lexer.TokenRParen) {
		p.nextToken()
		return identifiers, nil
	}

	p.nextToken()

	// Check for variadic parameter at the beginning: func (...args)
	if p.curTokenIs(lexer.TokenEllipsis) {
		if !p.expectPeek(lexer.TokenIdent) {
			p.addError("expected identifier after '...' in variadic parameter")
			return nil, nil
		}
		variadic = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectPeek(lexer.TokenRParen) {
			return nil, nil
		}
		return identifiers, variadic
	}

	ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(lexer.TokenComma) {
		p.nextToken() // skip comma
		p.nextToken()

		// Check for variadic parameter: func (a, b, ...rest)
		if p.curTokenIs(lexer.TokenEllipsis) {
			if !p.expectPeek(lexer.TokenIdent) {
				p.addError("expected identifier after '...' in variadic parameter")
				return nil, nil
			}
			variadic = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
			if !p.expectPeek(lexer.TokenRParen) {
				return nil, nil
			}
			return identifiers, variadic
		}

		ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(lexer.TokenRParen) {
		return nil, nil
	}

	return identifiers, nil
}

// parseNewExpression parses a new expression
func (p *Parser) parseNewExpression() Expression {
	expr := &NewExpression{Token: p.curToken}

	// Expect class name
	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}
	expr.Class = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect (
	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	// Parse arguments
	expr.Arguments = p.parseExpressionList(lexer.TokenRParen)

	return expr
}

// parseThisExpression parses this expression
func (p *Parser) parseThisExpression() Expression {
	return &ThisExpression{Token: p.curToken}
}

// parseSuperExpression parses super expression
func (p *Parser) parseSuperExpression() Expression {
	expr := &SuperExpression{Token: p.curToken}

	// Check for super.method() pattern
	if p.peekTokenIs(lexer.TokenDot) {
		p.nextToken() // consume dot
		if !p.expectPeek(lexer.TokenIdent) {
			return nil
		}
		methodName := p.curToken.Literal

		// Check for method call
		if p.peekTokenIs(lexer.TokenLParen) {
			p.nextToken()
			callExpr := &SuperCallExpression{
				Token:  expr.Token,
				Method: methodName,
				Args:   p.parseExpressionList(lexer.TokenRParen),
			}
			return callExpr
		}
	}

	return expr
}

// parseExpressionList parses a comma-separated list of expressions
func (p *Parser) parseExpressionList(end lexer.TokenType) []Expression {
	list := []Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.TokenComma) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}
