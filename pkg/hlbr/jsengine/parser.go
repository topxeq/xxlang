package jsengine

import (
	"strconv"
)

type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
}

func NewParser(input string) *Parser {
	p := &Parser{lexer: NewLexer(input)}
	p.nextToken()
	p.nextToken()
	return p
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
			return p.parseForStmt()
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
			return p.parseFunctionDecl()
		case "try":
			return p.parseTryStmt()
		case "throw":
			return p.parseThrowStmt()
		case "switch":
			return p.parseSwitchStmt()
		}
	case TokLBrace:
		return p.parseBlockStmt()
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
	name := p.cur.Literal
	p.nextToken()

	_ = keyword

	var val Expression
	if p.cur.Type == TokEq {
		p.nextToken()
		val = p.parseExpression()
	}

	return &VarDecl{Name: name, Value: val}
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

func (p *Parser) parseFunctionDecl() *FunctionDecl {
	p.nextToken()
	name := p.cur.Literal
	p.nextToken()
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

	return &FunctionDecl{Name: name, Params: params, Body: body.Statements}
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

	if p.cur.Type == TokEq || p.cur.Type == TokPlusEq || p.cur.Type == TokMinusEq ||
		p.cur.Type == TokStarEq || p.cur.Type == TokSlashEq {
		op := p.cur.Literal
		p.nextToken()
		right := p.parseAssignExpr()
		return &AssignExpr{Left: left, Op: op, Right: right}
	}

	return left
}

func (p *Parser) parseTernaryExpr() Expression {
	cond := p.parseOrExpr()

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
		} else if p.cur.Type == TokLBracket {
			p.nextToken()
			prop := p.parseExpression()
			p.nextToken()
			left = &MemberExpr{Object: left, Property: prop, Computed: true}
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
		return &Ident{Name: name}
	case TokNumber:
		val, _ := strconv.ParseFloat(p.cur.Literal, 64)
		p.nextToken()
		return &NumberLit{Value: val}
	case TokString:
		val := p.cur.Literal
		p.nextToken()
		return &StringLit{Value: val}
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
		}
	case TokLParen:
		p.nextToken()
		expr := p.parseExpression()
		p.nextToken()
		return expr
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

func (p *Parser) parseFunctionExpr() *FunctionExpr {
	p.nextToken()
	var name string
	if p.cur.Type == TokIdent {
		name = p.cur.Literal
		p.nextToken()
	}
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

	return &FunctionExpr{Name: name, Params: params, Body: body.Statements}
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
		elements = append(elements, p.parseExpression())
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	return &ArrayLit{Elements: elements}
}

func (p *Parser) parseObjectLit() *ObjectLit {
	p.nextToken()
	var props []ObjectProperty
	for p.cur.Type != TokRBrace {
		key := p.cur.Literal
		p.nextToken()
		p.nextToken()
		value := p.parseExpression()
		props = append(props, ObjectProperty{Key: key, Value: value})
		if p.cur.Type == TokComma {
			p.nextToken()
		}
	}
	p.nextToken()
	return &ObjectLit{Properties: props}
}

func (p *Parser) parseTryStmt() Statement {
	p.nextToken()
	_ = p.parseBlockStmt()

	if p.cur.Type == TokKeyword && p.cur.Literal == "catch" {
		p.nextToken()
		p.nextToken()
		p.nextToken()
		_ = p.parseBlockStmt()
	}

	if p.cur.Type == TokKeyword && p.cur.Literal == "finally" {
		p.nextToken()
		_ = p.parseBlockStmt()
	}

	return &ExpressionStmt{Expr: &Ident{Name: "undefined"}}
}

func (p *Parser) parseThrowStmt() Statement {
	p.nextToken()
	_ = p.parseExpression()
	return &ExpressionStmt{Expr: &Ident{Name: "undefined"}}
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
