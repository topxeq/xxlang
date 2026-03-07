// pkg/parser/parser.go
package parser

import (
	"fmt"
	"strconv"

	"github.com/topxeq/xxlang/pkg/lexer"
)

// Operator precedence constants
const (
	_ int = iota
	LOWEST      // lowest precedence for expression parsing
	ASSIGN      // =, +=, -= (right-associative, binds after LOWEST)
	OR          // ||
	AND         // &&
	EQUALS      // ==, !=
	LESSGREATER // <, >, <=, >=
	SUM         // +, -
	PRODUCT     // *, /, %
	PREFIX      // !, -
	CALL        // fn()
	INDEX       // arr[i]
	DOT         // obj.field
)

// precedence maps token types to their precedence
var precedence = map[lexer.TokenType]int{
	lexer.TokenOr:          OR,
	lexer.TokenAnd:         AND,
	lexer.TokenEqual:       EQUALS,
	lexer.TokenNotEqual:    EQUALS,
	lexer.TokenLT:          LESSGREATER,
	lexer.TokenGT:          LESSGREATER,
	lexer.TokenLTE:         LESSGREATER,
	lexer.TokenGTE:         LESSGREATER,
	lexer.TokenPlus:        SUM,
	lexer.TokenMinus:       SUM,
	lexer.TokenAsterisk:    PRODUCT,
	lexer.TokenSlash:       PRODUCT,
	lexer.TokenPercent:     PRODUCT,
	lexer.TokenLParen:      CALL,
	lexer.TokenLBracket:    INDEX,
	lexer.TokenDot:         DOT,
	lexer.TokenAssign:      ASSIGN,
	lexer.TokenPlusAssign:  ASSIGN,
	lexer.TokenMinusAssign: ASSIGN,
	lexer.TokenAsteriskAssign: ASSIGN,
	lexer.TokenSlashAssign: ASSIGN,
	lexer.TokenPercentAssign: ASSIGN,
	lexer.TokenIncrement:   CALL, // Postfix ++ has high precedence
	lexer.TokenDecrement:   CALL, // Postfix -- has high precedence
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
	case lexer.TokenBreak:
		return p.parseBreakStatement()
	case lexer.TokenContinue:
		return p.parseContinueStatement()
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

		if !p.expectPeek(lexer.TokenLBrace) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
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

// parseExpression parses an expression with the given precedence
func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.TokenSemicolon) && precedence <= p.peekPrecedence() {
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

// parseCallExpression parses a function call expression
func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(lexer.TokenRParen)
	return exp
}

// parseIndexExpression parses an index expression
func (p *Parser) parseIndexExpression(left Expression) Expression {
	exp := &IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

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

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters parses function parameters
func (p *Parser) parseFunctionParameters() []*Identifier {
	identifiers := []*Identifier{}

	if p.peekTokenIs(lexer.TokenRParen) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(lexer.TokenComma) {
		p.nextToken()
		p.nextToken()
		ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(lexer.TokenRParen) {
		return nil
	}

	return identifiers
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
