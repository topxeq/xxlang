package jsengine

import (
	"strconv"
)

// parserState holds the state of the parser for backtracking
type parserState struct {
	lexerPos int
	lexerCh  byte
	cur      Token
	peek     Token
}

type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
	pos   int // Current position for backtracking
}

func NewParser(input string) *Parser {
	p := &Parser{lexer: NewLexer(input)}
	p.nextToken()
	p.nextToken()
	return p
}

// save returns the current parser state for later restoration
func (p *Parser) save() parserState {
	return parserState{
		lexerPos: p.lexer.pos,
		lexerCh:  p.lexer.ch,
		cur:      p.cur,
		peek:     p.peek,
	}
}

// restore restores the parser to a previously saved state
func (p *Parser) restore(s parserState) {
	p.lexer.pos = s.lexerPos
	p.lexer.ch = s.lexerCh
	p.cur = s.cur
	p.peek = s.peek
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.lexer.NextToken()
}

func (p *Parser) Parse() *Program {
	prog := &Program{}
	for p.cur.Type != TokEOF {
		if p.cur.Type == TokSemi {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		} else {
			p.nextToken()
		}
	}
	return prog
}

func (p *Parser) parseStatement() Statement {
	switch p.cur.Type {
	case TokKeyword:
		switch p.cur.Literal {
		case "var", "let", "const":
			return p.parseVarDecl()
		case "if":
			return p.parseIfStmt()
		case "for":
			// Check if this is a for-in loop by looking ahead
			return p.parseForStmtOrForIn()
		case "while":
			return p.parseWhileStmt()
		case "return":
			return p.parseReturnStmt()
		case "break":
			p.nextToken()
			return &BreakStmt{}
		case "continue":
			p.nextToken()
			return &ContinueStmt{}
		case "function":
			// Check for generator function: function*
			p.nextToken()
			if p.cur.Type == TokStar {
				// This is a generator function
				p.nextToken()
				return p.parseGeneratorDecl()
			}
			// Not a generator, p.cur is now the function name
			// Parse as regular function from current position
			return p.parseFunctionDeclFromName()
		case "yield":
			// yield statement in generator
			p.nextToken()
			var val Expression
			if p.cur.Type != TokSemi && p.cur.Type != TokRBrace {
				val = p.parseExpression()
			}
			return &YieldStmt{Value: val}
		case "async":
			// async function declaration
			return p.parseAsyncFunctionDecl()
		case "try":
			return p.parseTryStmt()
		case "throw":
			return p.parseThrowStmt()
		case "switch":
			return p.parseSwitchStmt()
		case "class":
			return p.parseClassDecl()
		}
	}

	expr := p.parseExpression()
	if expr != nil {
		return &ExpressionStmt{Expr: expr}
	}
	return nil
}

func (p *Parser) parseVarDecl() *VarDecl {
	keyword := p.cur.Literal
	p.nextToken()

	_ = keyword

	// Check for array destructuring: var [a, b] = ...
	if p.cur.Type == TokLBracket {
		return p.parseArrayDestructDecl()
	}

	// Check for object destructuring: var {a, b} = ...
	if p.cur.Type == TokLBrace {
		return p.parseObjectDestructDecl()
	}

	// Regular variable declaration
	name := p.cur.Literal
	p.nextToken()

	var val Expression
	if p.cur.Type == TokEq {
		p.nextToken()
		val = p.parseExpression()
	}

	return &VarDecl{Name: name, Value: val}
}

// parseArrayDestructDecl parses array destructuring: var [a, b] = arr
func (p *Parser) parseArrayDestructDecl() *VarDecl {
	p.nextToken() // skip '['

	var elements []DestructElement
	for p.cur.Type != TokRBracket {
		name := p.cur.Literal
		p.nextToken()

		var defaultVal Expression
		if p.cur.Type == TokEq {
			p.nextToken()
			defaultVal = p.parseExpression()
		}

		elements = append(elements, DestructElement{Name: name, Default: defaultVal})

		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken() // skip ']'

	var val Expression
	if p.cur.Type == TokEq {
		p.nextToken()
		val = p.parseExpression()
	}

	return &VarDecl{
		IsDestructuring: true,
		DestructPattern: &DestructPattern{IsArray: true, Elements: elements},
		Value:           val,
	}
}

// parseObjectDestructDecl parses object destructuring: var {a, b: c} = obj
func (p *Parser) parseObjectDestructDecl() *VarDecl {
	p.nextToken() // skip '{'

	var elements []DestructElement
	for p.cur.Type != TokRBrace {
		propName := p.cur.Literal
		p.nextToken()

		var varName string
		var defaultVal Expression

		// Check for rename: {a: b} or default: {a = defaultVal}
		if p.cur.Type == TokColon {
			p.nextToken()
			varName = p.cur.Literal
			p.nextToken()
		} else {
			varName = propName
		}

		// Check for default value
		if p.cur.Type == TokEq {
			p.nextToken()
			defaultVal = p.parseExpression()
		}

		elements = append(elements, DestructElement{
			Name:     varName,
			Property: propName,
			Default:  defaultVal,
		})

		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken() // skip '}'

	var val Expression
	if p.cur.Type == TokEq {
		p.nextToken()
		val = p.parseExpression()
	}

	return &VarDecl{
		IsDestructuring: true,
		DestructPattern: &DestructPattern{IsArray: false, Elements: elements},
		Value:           val,
	}
}

func (p *Parser) parseIfStmt() *IfStmt {
	p.nextToken()
	p.nextToken()
	cond := p.parseExpression()
	p.nextToken()
	body := p.parseBlockStmt()

	stmt := &IfStmt{Cond: cond, Body: body.Statements}

	if p.cur.Type == TokKeyword && p.cur.Literal == "else" {
		p.nextToken()
		elseBody := p.parseBlockStmt()
		stmt.Else = elseBody.Statements
	}

	return stmt
}

func (p *Parser) parseForStmt() *ForStmt {
	p.nextToken()
	p.nextToken()

	var init Statement
	var cond Expression
	var post Statement

	if p.cur.Type != TokSemi {
		if p.cur.Type == TokKeyword && (p.cur.Literal == "var" || p.cur.Literal == "let") {
			init = p.parseVarDecl()
		} else {
			init = &ExpressionStmt{Expr: p.parseExpression()}
		}
	}
	if p.cur.Type == TokSemi {
		p.nextToken()
	}

	if p.cur.Type != TokSemi {
		cond = p.parseExpression()
	}
	if p.cur.Type == TokSemi {
		p.nextToken()
	}

	if p.cur.Type != TokRParen {
		post = &ExpressionStmt{Expr: p.parseExpression()}
	}
	if p.cur.Type == TokRParen {
		p.nextToken()
	}

	body := p.parseBlockStmt()

	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body.Statements}
}

// parseForStmtOrForIn determines if this is a regular for loop or for-in loop
func (p *Parser) parseForStmtOrForIn() Statement {
	// Save state before parsing
	state := p.save()
	p.nextToken() // skip 'for'
	p.nextToken() // skip '('

	// Check if this looks like for-in: for (var x in obj) or for-of: for (var x of obj)
	isForIn := false
	isForOf := false
	if p.cur.Type == TokKeyword && (p.cur.Literal == "var" || p.cur.Literal == "let") {
		p.nextToken() // skip var/let
		if p.cur.Type == TokIdent {
			p.nextToken()
			if p.cur.Type == TokKeyword && p.cur.Literal == "in" {
				isForIn = true
			} else if p.cur.Type == TokKeyword && p.cur.Literal == "of" {
				isForOf = true
			}
		}
	} else if p.cur.Type == TokIdent {
		p.nextToken()
		if p.cur.Type == TokKeyword && p.cur.Literal == "in" {
			isForIn = true
		} else if p.cur.Type == TokKeyword && p.cur.Literal == "of" {
			isForOf = true
		}
	}

	// Restore to original state
	p.restore(state)

	if isForIn {
		return p.parseForInStmt()
	}
	if isForOf {
		return p.parseForOfStmt()
	}
	return p.parseForStmt()
}

// parseForInStmt parses a for-in loop: for (var key in obj) { body }
func (p *Parser) parseForInStmt() *ForInStmt {
	p.nextToken() // skip 'for'
	p.nextToken() // skip '('

	// Parse variable declaration
	var varName string
	if p.cur.Type == TokKeyword && (p.cur.Literal == "var" || p.cur.Literal == "let") {
		p.nextToken()
	}
	if p.cur.Type == TokIdent {
		varName = p.cur.Literal
		p.nextToken()
	}

	// Expect 'in' keyword
	if p.cur.Type == TokKeyword && p.cur.Literal == "in" {
		p.nextToken()
	}

	// Parse object expression
	obj := p.parseExpression()

	// Expect ')'
	if p.cur.Type == TokRParen {
		p.nextToken()
	}

	body := p.parseBlockStmt()

	return &ForInStmt{VarName: varName, Object: obj, Body: body.Statements}
}

// parseForOfStmt parses a for-of loop: for (var x of iterable)
func (p *Parser) parseForOfStmt() *ForOfStmt {
	p.nextToken() // skip 'for'
	p.nextToken() // skip '('

	// Parse variable declaration
	var varName string
	if p.cur.Type == TokKeyword && (p.cur.Literal == "var" || p.cur.Literal == "let") {
		p.nextToken()
	}
	if p.cur.Type == TokIdent {
		varName = p.cur.Literal
		p.nextToken()
	}

	// Expect 'of' keyword
	if p.cur.Type == TokKeyword && p.cur.Literal == "of" {
		p.nextToken()
	}

	// Parse iterable expression
	obj := p.parseExpression()

	// Expect ')'
	if p.cur.Type == TokRParen {
		p.nextToken()
	}

	body := p.parseBlockStmt()

	return &ForOfStmt{VarName: varName, Object: obj, Body: body.Statements}
}

func (p *Parser) parseWhileStmt() *WhileStmt {
	p.nextToken()
	p.nextToken()
	cond := p.parseExpression()
	p.nextToken()
	body := p.parseBlockStmt()
	return &WhileStmt{Cond: cond, Body: body.Statements}
}

func (p *Parser) parseReturnStmt() *ReturnStmt {
	p.nextToken()
	if p.cur.Type == TokSemi || p.cur.Type == TokRBrace || p.cur.Type == TokEOF {
		return &ReturnStmt{}
	}
	val := p.parseExpression()
	return &ReturnStmt{Value: val}
}

// parseGeneratorDecl parses a generator function declaration
func (p *Parser) parseGeneratorDecl() *GeneratorDecl {
	name := p.cur.Literal // save function name
	p.nextToken()         // skip function name
	p.nextToken()         // skip '('

	var params []string
	var restParam string
	for p.cur.Type != TokRParen {
		if p.cur.Type == TokSpread {
			p.nextToken()
			if p.cur.Type == TokIdent {
				restParam = p.cur.Literal
				p.nextToken()
			}
		} else if p.cur.Type == TokIdent {
			params = append(params, p.cur.Literal)
			p.nextToken()
		}
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	body := p.parseBlockStmt()

	return &GeneratorDecl{Name: name, Params: params, RestParam: restParam, Body: body.Statements}
}

func (p *Parser) parseFunctionDecl() *FunctionDecl {
	p.nextToken() // skip 'function'
	name := p.cur.Literal
	p.nextToken()
	p.nextToken()

	var params []string
	var restParam string
	defaultVals := make(map[string]Expression)
	for p.cur.Type != TokRParen {
		if p.cur.Type == TokSpread {
			// Rest parameter: ...args
			p.nextToken()
			if p.cur.Type == TokIdent {
				restParam = p.cur.Literal
				p.nextToken()
			}
		} else if p.cur.Type == TokIdent {
			paramName := p.cur.Literal
			params = append(params, paramName)
			p.nextToken()
			// Handle default parameter value: = expr
			if p.cur.Type == TokEq {
				p.nextToken()
				defaultVals[paramName] = p.parseExpression()
			}
		}
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	body := p.parseBlockStmt()


	return &FunctionDecl{Name: name, Params: params, DefaultVals: defaultVals, RestParam: restParam, Body: body.Statements}
}

// parseFunctionDeclFromName parses a function declaration when current token is already the function name
func (p *Parser) parseFunctionDeclFromName() *FunctionDecl {
	name := p.cur.Literal
	p.nextToken()
	p.nextToken()

	var params []string
	var restParam string
	defaultVals := make(map[string]Expression)
	for p.cur.Type != TokRParen {
		if p.cur.Type == TokSpread {
			p.nextToken()
			if p.cur.Type == TokIdent {
				restParam = p.cur.Literal
				p.nextToken()
			}
		} else if p.cur.Type == TokIdent {
			paramName := p.cur.Literal
			params = append(params, paramName)
			p.nextToken()
			if p.cur.Type == TokEq {
				p.nextToken()
				defaultVals[paramName] = p.parseExpression()
			}
		}
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	body := p.parseBlockStmt()

	return &FunctionDecl{Name: name, Params: params, DefaultVals: defaultVals, RestParam: restParam, Body: body.Statements}
}

// parseAsyncFunctionDecl parses an async function declaration
func (p *Parser) parseAsyncFunctionDecl() *FunctionDecl {
	p.nextToken() // skip 'async'
	fd := p.parseFunctionDecl()
	fd.IsAsync = true
	return fd
}

func (p *Parser) parseBlockStmt() *BlockStmt {
	p.nextToken()
	var stmts []Statement
	for p.cur.Type != TokRBrace && p.cur.Type != TokEOF {
		if p.cur.Type == TokSemi {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		} else {
			p.nextToken()
		}
	}
	p.nextToken()
	return &BlockStmt{Statements: stmts}
}

func (p *Parser) parseExpression() Expression {
	return p.parseAssignExpr()
}

func (p *Parser) parseAssignExpr() Expression {
	left := p.parseTernaryExpr()
	if left == nil {
		return nil
	}

	if p.cur.Type == TokEq || p.cur.Type == TokPlusEq || p.cur.Type == TokMinusEq ||
		p.cur.Type == TokStarEq || p.cur.Type == TokSlashEq {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseAssignExpr()
		// Even if right is nil, we should return the assignment expression
		// The right side might be nil if we're at the end of an expression (e.g., before a closing paren)
		return &AssignExpr{Left: left, Op: op, Right: right}
	}

	return left
}

func (p *Parser) parseTernaryExpr() Expression {
	cond := p.parseOrExpr()

	// Handle nullish coalescing ?? (lower precedence than ternary)
	if p.cur.Type == TokNullish {
		p.nextToken()
		right := p.parseTernaryExpr()
		return &NullishCoalescingExpr{Left: cond, Right: right}
	}

	if p.cur.Type == TokQuestion {
		p.nextToken()
		trueExpr := p.parseExpression()
		p.nextToken()
		falseExpr := p.parseExpression()
		return &TernaryExpr{Cond: cond, True: trueExpr, False: falseExpr}
	}

	return cond
}

func (p *Parser) parseOrExpr() Expression {
	left := p.parseAndExpr()
	for p.cur.Type == TokOr {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseAndExpr()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAndExpr() Expression {
	left := p.parseEqualityExpr()
	for p.cur.Type == TokAnd {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseEqualityExpr()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseEqualityExpr() Expression {
	left := p.parseRelationalExpr()
	for p.cur.Type == TokEqEq || p.cur.Type == TokEqEqEq || p.cur.Type == TokNeq || p.cur.Type == TokNeqEq {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseRelationalExpr()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseRelationalExpr() Expression {
	left := p.parseAdditiveExpr()
	for p.cur.Type == TokLt || p.cur.Type == TokGt || p.cur.Type == TokLte || p.cur.Type == TokGte {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseAdditiveExpr()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	// Handle instanceof and in operators
	for p.cur.Type == TokKeyword && (p.cur.Literal == "instanceof" || p.cur.Literal == "in") {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseAdditiveExpr()
		if op == "instanceof" {
			left = &InstanceofExpr{Object: left, Constructor: right}
		} else {
			left = &InExpr{Property: left, Object: right}
		}
	}
	return left
}

func (p *Parser) parseAdditiveExpr() Expression {
	left := p.parseMultiplicativeExpr()
	for p.cur.Type == TokPlus || p.cur.Type == TokMinus {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseMultiplicativeExpr()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplicativeExpr() Expression {
	left := p.parseUnaryExpr()
	for p.cur.Type == TokStar || p.cur.Type == TokSlash || p.cur.Type == TokPercent {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseUnaryExpr()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseUnaryExpr() Expression {
	// Handle prefix increment/decrement: ++i, --i
	if p.cur.Type == TokInc || p.cur.Type == TokDec {
		op := p.cur.Literal
		p.nextToken()
		expr := p.parseUnaryExpr()
		return &UpdateExpr{Operator: op, Prefix: true, Operand: expr}
	}
	if p.cur.Type == TokNot || p.cur.Type == TokMinus || p.cur.Type == TokPlus {
		op := p.cur.Literal
		p.nextToken()
		expr := p.parseUnaryExpr()
		return &UnaryExpr{Op: op, Expr: expr}
	}
	if p.cur.Type == TokKeyword && p.cur.Literal == "typeof" {
		p.nextToken()
		expr := p.parseUnaryExpr()
		return &TypeOfExpr{Expr: expr}
	}
	if p.cur.Type == TokKeyword && p.cur.Literal == "delete" {
		p.nextToken()
		expr := p.parseUnaryExpr()
		return &UnaryExpr{Op: "delete", Expr: expr}
	}
	if p.cur.Type == TokKeyword && p.cur.Literal == "await" {
		p.nextToken()
		expr := p.parseUnaryExpr()
		return &AwaitExpr{Expr: expr}
	}
	return p.parseCallExpr()
}

func (p *Parser) parseCallExpr() Expression {
	left := p.parsePrimaryExpr()
	if left == nil {
		return nil
	}

	for {
		if p.cur.Type == TokLParen {
			p.nextToken()
			var args []Expression
			for p.cur.Type != TokRParen {
				args = append(args, p.parseExpression())
				if p.cur.Type == TokComma {
					p.nextToken()
				}
			}
			p.nextToken()
			left = &CallExpr{Callee: left, Args: args}
		} else if p.cur.Type == TokDot {
			p.nextToken()
			prop := &Ident{Name: p.cur.Literal}
			p.nextToken()
			left = &MemberExpr{Object: left, Property: prop, Computed: false}
		} else if p.cur.Type == TokOptChain {
			// Optional chaining ?.
			p.nextToken()
			if p.cur.Type == TokLParen {
				// obj?.()
				p.nextToken()
				var args []Expression
				for p.cur.Type != TokRParen {
					args = append(args, p.parseExpression())
					if p.cur.Type == TokComma {
						p.nextToken()
					}
				}
				p.nextToken()
				left = &OptionalChainExpr{Object: left, Property: &CallExpr{Callee: left, Args: args}, Computed: false}
			} else if p.cur.Type == TokLBracket {
				// obj?.[key]
				p.nextToken()
				prop := p.parseExpression()
				p.nextToken()
				left = &OptionalChainExpr{Object: left, Property: prop, Computed: true}
			} else {
				// obj?.prop
				prop := &Ident{Name: p.cur.Literal}
				p.nextToken()
				left = &OptionalChainExpr{Object: left, Property: prop, Computed: false}
			}
		} else if p.cur.Type == TokLBracket {
			p.nextToken()
			prop := p.parseExpression()
			p.nextToken()
			left = &MemberExpr{Object: left, Property: prop, Computed: true}
		} else if p.cur.Type == TokInc || p.cur.Type == TokDec {
			// Postfix increment/decrement: i++, i--
			op := p.cur.Literal
			p.nextToken()
			left = &UpdateExpr{Operator: op, Prefix: false, Operand: left}
		} else {
			break
		}
	}

	return left
}

func (p *Parser) parsePrimaryExpr() Expression {
	switch p.cur.Type {
	case TokIdent:
		name := p.cur.Literal
		p.nextToken()
		// Check for single-param arrow function: x => expr
		if p.cur.Type == TokArrow {
			p.nextToken()
			return p.parseArrowFunctionBody([]string{name})
		}
		return &Ident{Name: name}
	case TokNumber:
		val, _ := strconv.ParseFloat(p.cur.Literal, 64)
		p.nextToken()
		return &NumberLit{Value: val}
	case TokBigInt:
		val := p.cur.Literal
		p.nextToken()
		return &BigIntLit{Value: val}
	case TokString:
		val := p.cur.Literal
		p.nextToken()
		return &StringLit{Value: val}
	case TokTemplate:
		return p.parseTemplateLit()
	case TokKeyword:
		switch p.cur.Literal {
		case "true":
			p.nextToken()
			return &BoolLit{Value: true}
		case "false":
			p.nextToken()
			return &BoolLit{Value: false}
		case "null":
			p.nextToken()
			return &NullLit{}
		case "undefined":
			p.nextToken()
			return &UndefinedLit{}
		case "function":
			return p.parseFunctionExpr()
		case "new":
			return p.parseNewExpr()
		case "class":
			return p.parseClassExpr()
		case "this":
			p.nextToken()
			return &ThisExpr{}
		case "super":
			p.nextToken()
			if p.cur.Type == TokLParen {
				p.nextToken()
				var args []Expression
				for p.cur.Type != TokRParen {
					args = append(args, p.parseExpression())
					if p.cur.Type == TokComma {
						p.nextToken()
					}
				}
				p.nextToken()
				return &SuperExpr{Args: args}
			}
			return &Ident{Name: "super"}
		}
	case TokLParen:
		// Could be grouped expression or arrow function
		return p.parseParenExprOrArrowFunc()
	case TokLBracket:
		return p.parseArrayLit()
	case TokLBrace:
		return p.parseObjectLit()
	case TokSemi, TokRBrace, TokRParen, TokRBracket, TokEOF:
		return nil
	}

	p.nextToken()
	return &Ident{Name: "_"}
}

// parseParenExprOrArrowFunc handles (expr) or (params) => body
func (p *Parser) parseParenExprOrArrowFunc() Expression {
	p.nextToken() // skip (

	// Save state for backtracking
	savedPos := p.lexer.pos
	savedCur := p.cur
	savedPeek := p.peek

	// Check if this could be arrow function parameters
	var params []string
	isArrowFunc := false

	if p.cur.Type == TokRParen {
		// () => ...  empty params
		p.nextToken()
		if p.cur.Type == TokArrow {
			isArrowFunc = true
		}
	} else {
		// Try to parse as params
		for p.cur.Type != TokRParen && p.cur.Type != TokEOF {
			if p.cur.Type == TokIdent {
				params = append(params, p.cur.Literal)
				p.nextToken()
				if p.cur.Type == TokComma {
					p.nextToken()
				}
			} else {
				break
			}
		}

		if p.cur.Type == TokRParen {
			p.nextToken()
			if p.cur.Type == TokArrow {
				isArrowFunc = true
			}
		}
	}

	if isArrowFunc {
		p.nextToken() // skip =>
		return p.parseArrowFunctionBody(params)
	}

	// Not an arrow function - backtrack and parse as grouped expression
	p.lexer.pos = savedPos
	p.cur = savedCur
	p.peek = savedPeek
	expr := p.parseExpression()
	// Skip to the closing paren
	for p.cur.Type != TokRParen && p.cur.Type != TokEOF {
		p.nextToken()
	}
	if p.cur.Type == TokRParen {
		p.nextToken() // skip )
	}
	return expr
}

// parseArrowFunctionBody parses the body of an arrow function
func (p *Parser) parseArrowFunctionBody(params []string) *ArrowFunctionExpr {
	// Check if body is a block or expression
	if p.cur.Type == TokLBrace {
		// Block body: { statements }
		body := p.parseBlockStmt()
		return &ArrowFunctionExpr{
			Params: params,
			Body:   body.Statements,
		}
	}

	// Expression body: implicit return
	expr := p.parseExpression()
	return &ArrowFunctionExpr{
		Params:     params,
		Expression: true,
		Expr:       expr,
		Body: []Statement{
			&ReturnStmt{Value: expr},
		},
	}
}

func (p *Parser) parseFunctionExpr() *FunctionExpr {
	p.nextToken()
	var name string
	if p.cur.Type == TokIdent {
		name = p.cur.Literal
		p.nextToken()
	}
	p.nextToken()

	var params []string
	var restParam string
	defaultVals := make(map[string]Expression)
	for p.cur.Type != TokRParen {
		if p.cur.Type == TokSpread {
			// Rest parameter: ...args
			p.nextToken()
			if p.cur.Type == TokIdent {
				restParam = p.cur.Literal
				p.nextToken()
			}
		} else if p.cur.Type == TokIdent {
			paramName := p.cur.Literal
			params = append(params, paramName)
			p.nextToken()
			// Handle default parameter value: = expr
			if p.cur.Type == TokEq {
				p.nextToken()
				defaultVals[paramName] = p.parseExpression()
			}
		}
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	body := p.parseBlockStmt()

	return &FunctionExpr{Name: name, Params: params, DefaultVals: defaultVals, RestParam: restParam, Body: body.Statements}
}

func (p *Parser) parseNewExpr() *NewExpr {
	p.nextToken()
	callee := p.parsePrimaryExpr()

	var args []Expression
	if p.cur.Type == TokLParen {
		p.nextToken()
		for p.cur.Type != TokRParen {
			args = append(args, p.parseExpression())
			if p.cur.Type == TokComma {
				p.nextToken()
			}
		}
		p.nextToken()
	}

	return &NewExpr{Callee: callee, Args: args}
}

func (p *Parser) parseArrayLit() *ArrayLit {
	p.nextToken()
	var elements []Expression
	for p.cur.Type != TokRBracket {
		// Check for spread operator
		if p.cur.Type == TokSpread {
			p.nextToken()
			elements = append(elements, &SpreadExpr{Argument: p.parseExpression()})
		} else {
			elements = append(elements, p.parseExpression())
		}
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	return &ArrayLit{Elements: elements}
}


// parseTemplateLit parses a template literal: `Hello ${name}!`
func (p *Parser) parseTemplateLit() *TemplateLit {
	raw := p.cur.Literal
	p.nextToken()

	var parts []TemplatePart
	i := 0
	for i < len(raw) {
		// Look for ${...}
		if i+1 < len(raw) && raw[i] == '$' && raw[i+1] == '{' {
			// Find matching }
			braceCount := 1
			j := i + 2
			for j < len(raw) && braceCount > 0 {
				if raw[j] == '{' {
					braceCount++
				} else if raw[j] == '}' {
					braceCount--
				}
				j++
			}
			// Parse the expression inside ${...}
			exprStr := raw[i+2 : j-1]
			if exprStr != "" {
				subParser := NewParser(exprStr)
				expr := subParser.parseExpression()
				if expr != nil {
					parts = append(parts, TemplatePart{IsExpr: true, Expr: expr})
				}
			}
			i = j
		} else {
			// Collect text until next ${
			start := i
			for i < len(raw) && !(i+1 < len(raw) && raw[i] == '$' && raw[i+1] == '{') {
				i++
			}
			if start < i {
				parts = append(parts, TemplatePart{IsExpr: false, Text: raw[start:i]})
			}
		}
	}

	return &TemplateLit{Parts: parts}
}

func (p *Parser) parseObjectLit() *ObjectLit {
	p.nextToken()
	var props []ObjectProperty
	for p.cur.Type != TokRBrace {
		// Check for spread operator: ...obj
		if p.cur.Type == TokSpread {
			p.nextToken()
			value := p.parseExpression()
			props = append(props, ObjectProperty{Key: "", Value: value, Spread: true})
		} else if p.cur.Type == TokLBracket {
			// Computed property name: [expr]
			p.nextToken()
			keyExpr := p.parseExpression()
			p.nextToken() // skip ]
			p.nextToken() // skip :
			value := p.parseExpression()
			props = append(props, ObjectProperty{Key: "", Value: value, Computed: true, KeyExpr: keyExpr})
		} else if p.cur.Type == TokIdent {
			key := p.cur.Literal
			p.nextToken()
			// Check for shorthand { x } or method shorthand { f() {} }
			if p.cur.Type == TokComma || p.cur.Type == TokRBrace {
				// Shorthand: { x } means { x: x }
				props = append(props, ObjectProperty{Key: key, Value: &Ident{Name: key}, Shorthand: true})
			} else if p.cur.Type == TokLParen {
				// Method shorthand: { f() {} }
				p.nextToken()
				var params []string
				for p.cur.Type != TokRParen {
					if p.cur.Type == TokIdent {
						params = append(params, p.cur.Literal)
					}
					p.nextToken()
					if p.cur.Type == TokComma {
						p.nextToken()
					}
				}
				p.nextToken()
				body := p.parseBlockStmt()
				fn := &Function{Params: params, Body: body.Statements, Env: nil}
				props = append(props, ObjectProperty{Key: key, Value: &FunctionExpr{Params: params, Body: body.Statements}})
				_ = fn
			} else if p.cur.Type == TokColon {
				p.nextToken()
				value := p.parseExpression()
				props = append(props, ObjectProperty{Key: key, Value: value})
			}
		} else {
			key := p.cur.Literal
			p.nextToken()
			p.nextToken()
			value := p.parseExpression()
			props = append(props, ObjectProperty{Key: key, Value: value})
		}
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	return &ObjectLit{Properties: props}
}

func (p *Parser) parseTryStmt() Statement {
	p.nextToken()
	body := p.parseBlockStmt()

	var catchVar string
	var catch []Statement
	if p.cur.Type == TokKeyword && p.cur.Literal == "catch" {
		p.nextToken()
		// Parse catch variable
		if p.cur.Type == TokLParen {
			p.nextToken()
			if p.cur.Type == TokIdent {
				catchVar = p.cur.Literal
				p.nextToken()
			}
			p.nextToken() // skip )
		}
		catch = p.parseBlockStmt().Statements
	}

	var finally []Statement
	if p.cur.Type == TokKeyword && p.cur.Literal == "finally" {
		p.nextToken()
		finally = p.parseBlockStmt().Statements
	}

	return &TryStmt{
		Body:     body.Statements,
		CatchVar: catchVar,
		Catch:    catch,
		Finally:  finally,
	}
}

func (p *Parser) parseThrowStmt() Statement {
	p.nextToken()
	val := p.parseExpression()
	return &ThrowStmt{Value: val}
}

func (p *Parser) parseSwitchStmt() Statement {
	p.nextToken()
	p.nextToken()
	_ = p.parseExpression()
	p.nextToken()
	p.nextToken()

	for p.cur.Type != TokRBrace && p.cur.Type != TokEOF {
		if p.cur.Type == TokKeyword && (p.cur.Literal == "case" || p.cur.Literal == "default") {
			p.nextToken()
			if p.cur.Literal != "default" {
				_ = p.parseExpression()
			}
			p.nextToken()
		}
		_ = p.parseStatement()
	}
	p.nextToken()

	return &ExpressionStmt{Expr: &Ident{Name: "undefined"}}
}

// parseClassDecl parses a class declaration: class Name [extends SuperClass] { body }
func (p *Parser) parseClassDecl() *ClassDecl {
	p.nextToken() // skip 'class'
	
	name := ""
	if p.cur.Type == TokIdent {
		name = p.cur.Literal
		p.nextToken()
	}
	
	superClass := ""
	if p.cur.Type == TokKeyword && p.cur.Literal == "extends" {
		p.nextToken()
		if p.cur.Type == TokIdent {
			superClass = p.cur.Literal
			p.nextToken()
		}
	}
	
	body := p.parseClassBody()
	
	return &ClassDecl{
		Name:       name,
		SuperClass: superClass,
		Body:       body,
	}
}

// parseClassExpr parses a class expression (used in expressions)
func (p *Parser) parseClassExpr() *ClassExpr {
	p.nextToken() // skip 'class'
	
	name := ""
	if p.cur.Type == TokIdent {
		name = p.cur.Literal
		p.nextToken()
	}
	
	superClass := ""
	if p.cur.Type == TokKeyword && p.cur.Literal == "extends" {
		p.nextToken()
		if p.cur.Type == TokIdent {
			superClass = p.cur.Literal
			p.nextToken()
		}
	}
	
	body := p.parseClassBody()
	
	return &ClassExpr{
		Name:       name,
		SuperClass: superClass,
		Body:       body,
	}
}

// parseClassBody parses the body of a class: { methods }
func (p *Parser) parseClassBody() []ClassElement {
	var elements []ClassElement
	
	p.nextToken() // skip '{'
	
	for p.cur.Type != TokRBrace && p.cur.Type != TokEOF {
		elem := p.parseClassElement()
		if elem != nil {
			elements = append(elements, *elem)
		} else {
			p.nextToken()
		}
	}
	
	p.nextToken() // skip '}'
	
	return elements
}

// parseClassElement parses a single element in a class body
func (p *Parser) parseClassElement() *ClassElement {
	isStatic := false
	isGetter := false
	isSetter := false
	
	// Check for static keyword
	if p.cur.Type == TokKeyword && p.cur.Literal == "static" {
		isStatic = true
		p.nextToken()
	}
	
	// Check for getter/setter
	if p.cur.Type == TokKeyword && p.cur.Literal == "get" {
		isGetter = true
		p.nextToken()
	} else if p.cur.Type == TokKeyword && p.cur.Literal == "set" {
		isSetter = true
		p.nextToken()
	}
	
	// Method name
	if p.cur.Type != TokIdent {
		return nil
	}
	
	name := p.cur.Literal
	p.nextToken()
	
	// Parse parameters
	p.nextToken() // skip '('
	var params []string
	for p.cur.Type != TokRParen && p.cur.Type != TokEOF {
		if p.cur.Type == TokIdent {
			params = append(params, p.cur.Literal)
		}
		p.nextToken()
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken() // skip ')'
	
	// Parse body
	body := p.parseBlockStmt()
	
	return &ClassElement{
		Type:     "method",
		Name:     name,
		Params:   params,
		Body:     body.Statements,
		IsStatic: isStatic,
		IsGetter: isGetter,
		IsSetter: isSetter,
	}
}
