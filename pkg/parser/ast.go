// pkg/parser/ast.go
package parser

import (
	"fmt"
	"strings"

	"github.com/topxeq/xxlang/pkg/lexer"
)

// Node represents a node in the Abstract Syntax Tree
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

// Program is the root node of every AST
type Program struct {
	Statements []Statement
}

// TokenLiteral returns the token literal of the first statement
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String returns a string representation of the program
func (p *Program) String() string {
	var sb strings.Builder
	for _, s := range p.Statements {
		sb.WriteString(s.String())
	}
	return sb.String()
}

// ============================================
// Statements
// ============================================

// VarStatement represents a variable declaration statement
type VarStatement struct {
	Token lexer.Token // the 'var' token
	Name  *Identifier
	Value Expression
}

func (vs *VarStatement) statementNode() {}

// TokenLiteral returns the token literal
func (vs *VarStatement) TokenLiteral() string {
	return vs.Token.Literal
}

// String returns a string representation of the var statement
func (vs *VarStatement) String() string {
	var sb strings.Builder
	sb.WriteString(vs.TokenLiteral() + " ")
	sb.WriteString(vs.Name.String())
	sb.WriteString(" = ")
	if vs.Value != nil {
		sb.WriteString(vs.Value.String())
	}
	sb.WriteString(";")
	return sb.String()
}

// ShortVarStatement represents a short variable declaration (:=)
type ShortVarStatement struct {
	Token lexer.Token // the identifier token (first identifier in the declaration)
	Name  *Identifier
	Value Expression
}

func (svs *ShortVarStatement) statementNode() {}

// TokenLiteral returns the token literal
func (svs *ShortVarStatement) TokenLiteral() string {
	return svs.Token.Literal
}

// String returns a string representation of the short var statement
func (svs *ShortVarStatement) String() string {
	var sb strings.Builder
	sb.WriteString(svs.Name.String())
	sb.WriteString(" := ")
	if svs.Value != nil {
		sb.WriteString(svs.Value.String())
	}
	sb.WriteString(";")
	return sb.String()
}

// ConstStatement represents a constant declaration statement
type ConstStatement struct {
	Token lexer.Token // the 'const' token
	Name  *Identifier
	Value Expression
}

func (cs *ConstStatement) statementNode() {}

// TokenLiteral returns the token literal
func (cs *ConstStatement) TokenLiteral() string {
	return cs.Token.Literal
}

// String returns a string representation of the const statement
func (cs *ConstStatement) String() string {
	var sb strings.Builder
	sb.WriteString(cs.TokenLiteral() + " ")
	sb.WriteString(cs.Name.String())
	sb.WriteString(" = ")
	if cs.Value != nil {
		sb.WriteString(cs.Value.String())
	}
	sb.WriteString(";")
	return sb.String()
}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Token       lexer.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns the token literal
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

// String returns a string representation of the return statement
func (rs *ReturnStatement) String() string {
	var sb strings.Builder
	sb.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		sb.WriteString(rs.ReturnValue.String())
	}
	sb.WriteString(";")
	return sb.String()
}

// ExpressionStatement represents an expression statement
type ExpressionStatement struct {
	Token      lexer.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the token literal
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

// String returns a string representation of the expression statement
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement represents a block of statements
type BlockStatement struct {
	Token      lexer.Token // the '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the token literal
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

// String returns a string representation of the block statement
func (bs *BlockStatement) String() string {
	var sb strings.Builder
	sb.WriteString("{ ")
	for _, s := range bs.Statements {
		sb.WriteString(s.String())
		sb.WriteString(" ")
	}
	sb.WriteString("}")
	return sb.String()
}

// IfStatement represents an if statement
type IfStatement struct {
	Token       lexer.Token // the 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode() {}

// TokenLiteral returns the token literal
func (is *IfStatement) TokenLiteral() string {
	return is.Token.Literal
}

// String returns a string representation of the if statement
func (is *IfStatement) String() string {
	var sb strings.Builder
	sb.WriteString("if (")
	sb.WriteString(is.Condition.String())
	sb.WriteString(") ")
	sb.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		sb.WriteString(" else ")
		sb.WriteString(is.Alternative.String())
	}
	return sb.String()
}

// WhileStatement represents a while statement
type WhileStatement struct {
	Token     lexer.Token // the 'while' token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode() {}

// TokenLiteral returns the token literal
func (ws *WhileStatement) TokenLiteral() string {
	return ws.Token.Literal
}

// String returns a string representation of the while statement
func (ws *WhileStatement) String() string {
	var sb strings.Builder
	sb.WriteString("while (")
	sb.WriteString(ws.Condition.String())
	sb.WriteString(") ")
	sb.WriteString(ws.Body.String())
	return sb.String()
}

// ForStatement represents a for statement
type ForStatement struct {
	Token     lexer.Token // the 'for' token
	Init      Statement
	Condition Expression
	Update    Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode() {}

// TokenLiteral returns the token literal
func (fs *ForStatement) TokenLiteral() string {
	return fs.Token.Literal
}

// String returns a string representation of the for statement
func (fs *ForStatement) String() string {
	var sb strings.Builder
	sb.WriteString("for (")
	if fs.Init != nil {
		initStr := fs.Init.String()
		// Remove trailing semicolon from init statement if present
		if strings.HasSuffix(initStr, ";") {
			initStr = initStr[:len(initStr)-1]
		}
		sb.WriteString(initStr)
	}
	sb.WriteString("; ")
	if fs.Condition != nil {
		sb.WriteString(fs.Condition.String())
	}
	sb.WriteString("; ")
	if fs.Update != nil {
		sb.WriteString(fs.Update.String())
	}
	sb.WriteString(") ")
	sb.WriteString(fs.Body.String())
	return sb.String()
}

// BreakStatement represents a break statement
type BreakStatement struct {
	Token lexer.Token // the 'break' token
}

func (bs *BreakStatement) statementNode() {}

// TokenLiteral returns the token literal
func (bs *BreakStatement) TokenLiteral() string {
	return bs.Token.Literal
}

// String returns a string representation of the break statement
func (bs *BreakStatement) String() string {
	return "break;"
}

// ContinueStatement represents a continue statement
type ContinueStatement struct {
	Token lexer.Token // the 'continue' token
}

func (cs *ContinueStatement) statementNode() {}

// TokenLiteral returns the token literal
func (cs *ContinueStatement) TokenLiteral() string {
	return cs.Token.Literal
}

// String returns a string representation of the continue statement
func (cs *ContinueStatement) String() string {
	return "continue;"
}

// ============================================
// Expressions
// ============================================

// Identifier represents an identifier expression
type Identifier struct {
	Token lexer.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns the token literal
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

// String returns a string representation of the identifier
func (i *Identifier) String() string {
	return i.Value
}

// IntegerLiteral represents an integer literal expression
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

// String returns a string representation of the integer literal
func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

// FloatLiteral represents a float literal expression
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (fl *FloatLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

// String returns a string representation of the float literal
func (fl *FloatLiteral) String() string {
	return fl.Token.Literal
}

// StringLiteral represents a string literal expression
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Literal
}

// String returns a string representation of the string literal
func (sl *StringLiteral) String() string {
	return "\"" + sl.Token.Literal + "\""
}

// BooleanLiteral represents a boolean literal expression
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (bl *BooleanLiteral) TokenLiteral() string {
	return bl.Token.Literal
}

// String returns a string representation of the boolean literal
func (bl *BooleanLiteral) String() string {
	return bl.Token.Literal
}

// NullLiteral represents a null literal expression
type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (nl *NullLiteral) TokenLiteral() string {
	return nl.Token.Literal
}

// String returns a string representation of the null literal
func (nl *NullLiteral) String() string {
	return "null"
}

// ArrayLiteral represents an array literal expression
type ArrayLiteral struct {
	Token    lexer.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (al *ArrayLiteral) TokenLiteral() string {
	return al.Token.Literal
}

// String returns a string representation of the array literal
func (al *ArrayLiteral) String() string {
	var sb strings.Builder
	elements := make([]string, len(al.Elements))
	for i, el := range al.Elements {
		elements[i] = el.String()
	}
	sb.WriteString("[")
	sb.WriteString(strings.Join(elements, ", "))
	sb.WriteString("]")
	return sb.String()
}

// MapLiteral represents a map literal expression
type MapLiteral struct {
	Token lexer.Token // the '{' token
	Pairs map[Expression]Expression
}

func (ml *MapLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (ml *MapLiteral) TokenLiteral() string {
	return ml.Token.Literal
}

// String returns a string representation of the map literal
func (ml *MapLiteral) String() string {
	var sb strings.Builder
	pairs := make([]string, 0, len(ml.Pairs))
	for key, value := range ml.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}
	sb.WriteString("{")
	sb.WriteString(strings.Join(pairs, ", "))
	sb.WriteString("}")
	return sb.String()
}

// PrefixExpression represents a prefix expression
type PrefixExpression struct {
	Token    lexer.Token // The prefix token, e.g. !, -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

// String returns a string representation of the prefix expression
func (pe *PrefixExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(pe.Operator)
	sb.WriteString(pe.Right.String())
	sb.WriteString(")")
	return sb.String()
}

// InfixExpression represents an infix expression
type InfixExpression struct {
	Token    lexer.Token // The operator token, e.g. +, -
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

// String returns a string representation of the infix expression
func (ie *InfixExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(ie.Left.String())
	sb.WriteString(" ")
	sb.WriteString(ie.Operator)
	sb.WriteString(" ")
	sb.WriteString(ie.Right.String())
	sb.WriteString(")")
	return sb.String()
}

// CallExpression represents a function call expression
type CallExpression struct {
	Token     lexer.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

// String returns a string representation of the call expression
func (ce *CallExpression) String() string {
	var sb strings.Builder
	args := make([]string, len(ce.Arguments))
	for i, arg := range ce.Arguments {
		args[i] = arg.String()
	}
	sb.WriteString(ce.Function.String())
	sb.WriteString("(")
	sb.WriteString(strings.Join(args, ", "))
	sb.WriteString(")")
	return sb.String()
}

// IndexExpression represents an index expression
type IndexExpression struct {
	Token lexer.Token // The '[' token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (ie *IndexExpression) TokenLiteral() string {
	return ie.Token.Literal
}

// String returns a string representation of the index expression
func (ie *IndexExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(ie.Left.String())
	sb.WriteString("[")
	sb.WriteString(ie.Index.String())
	sb.WriteString("])")
	return sb.String()
}

// DotExpression represents a dot expression for property access
type DotExpression struct {
	Token    lexer.Token // The '.' token
	Object   Expression
	Property *Identifier
}

func (de *DotExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (de *DotExpression) TokenLiteral() string {
	return de.Token.Literal
}

// String returns a string representation of the dot expression
func (de *DotExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(de.Object.String())
	sb.WriteString(".")
	sb.WriteString(de.Property.String())
	sb.WriteString(")")
	return sb.String()
}

// AssignmentExpression represents an assignment expression
type AssignmentExpression struct {
	Token lexer.Token // The '=' token
	Left  Expression
	Value Expression
}

func (ae *AssignmentExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (ae *AssignmentExpression) TokenLiteral() string {
	return ae.Token.Literal
}

// String returns a string representation of the assignment expression
func (ae *AssignmentExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(ae.Left.String())
	sb.WriteString(" = ")
	sb.WriteString(ae.Value.String())
	sb.WriteString(")")
	return sb.String()
}

// FunctionLiteral represents a function literal expression
type FunctionLiteral struct {
	Token      lexer.Token // The 'func' token
	Name       string      // Optional: named function
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

// String returns a string representation of the function literal
func (fl *FunctionLiteral) String() string {
	var sb strings.Builder
	params := make([]string, len(fl.Parameters))
	for i, p := range fl.Parameters {
		params[i] = p.String()
	}
	sb.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		sb.WriteString(" ")
		sb.WriteString(fl.Name)
	}
	sb.WriteString("(")
	sb.WriteString(strings.Join(params, ", "))
	sb.WriteString(") ")
	sb.WriteString(fl.Body.String())
	return sb.String()
}

// ============================================
// Additional AST nodes for future expansion
// ============================================

// SwitchStatement represents a switch statement
type SwitchStatement struct {
	Token      lexer.Token // The 'switch' token
	Expression Expression
	Cases      []*CaseStatement
	Default    *DefaultStatement
}

func (ss *SwitchStatement) statementNode() {}

// TokenLiteral returns the token literal
func (ss *SwitchStatement) TokenLiteral() string {
	return ss.Token.Literal
}

// String returns a string representation of the switch statement
func (ss *SwitchStatement) String() string {
	var sb strings.Builder
	sb.WriteString("switch (")
	sb.WriteString(ss.Expression.String())
	sb.WriteString(") {")
	for _, c := range ss.Cases {
		sb.WriteString(" ")
		sb.WriteString(c.String())
	}
	if ss.Default != nil {
		sb.WriteString(" ")
		sb.WriteString(ss.Default.String())
	}
	sb.WriteString(" }")
	return sb.String()
}

// CaseStatement represents a case statement in switch
type CaseStatement struct {
	Token       lexer.Token // The 'case' token
	Expression  Expression
	Consequence *BlockStatement
}

func (cs *CaseStatement) statementNode() {}

// TokenLiteral returns the token literal
func (cs *CaseStatement) TokenLiteral() string {
	return cs.Token.Literal
}

// String returns a string representation of the case statement
func (cs *CaseStatement) String() string {
	var sb strings.Builder
	sb.WriteString("case ")
	sb.WriteString(cs.Expression.String())
	sb.WriteString(": ")
	sb.WriteString(cs.Consequence.String())
	return sb.String()
}

// DefaultStatement represents a default statement in switch
type DefaultStatement struct {
	Token       lexer.Token // The 'default' token
	Consequence *BlockStatement
}

func (ds *DefaultStatement) statementNode() {}

// TokenLiteral returns the token literal
func (ds *DefaultStatement) TokenLiteral() string {
	return ds.Token.Literal
}

// String returns a string representation of the default statement
func (ds *DefaultStatement) String() string {
	var sb strings.Builder
	sb.WriteString("default: ")
	sb.WriteString(ds.Consequence.String())
	return sb.String()
}

// TryStatement represents a try-catch-finally statement
type TryStatement struct {
	Token   lexer.Token // The 'try' token
	Block   *BlockStatement
	Catch   *CatchStatement
	Finally *FinallyStatement
}

func (ts *TryStatement) statementNode() {}

// TokenLiteral returns the token literal
func (ts *TryStatement) TokenLiteral() string {
	return ts.Token.Literal
}

// String returns a string representation of the try statement
func (ts *TryStatement) String() string {
	var sb strings.Builder
	sb.WriteString("try ")
	sb.WriteString(ts.Block.String())
	if ts.Catch != nil {
		sb.WriteString(" ")
		sb.WriteString(ts.Catch.String())
	}
	if ts.Finally != nil {
		sb.WriteString(" ")
		sb.WriteString(ts.Finally.String())
	}
	return sb.String()
}

// CatchStatement represents a catch statement
type CatchStatement struct {
	Token     lexer.Token // The 'catch' token
	Exception *Identifier
	Block     *BlockStatement
}

func (cs *CatchStatement) statementNode() {}

// TokenLiteral returns the token literal
func (cs *CatchStatement) TokenLiteral() string {
	return cs.Token.Literal
}

// String returns a string representation of the catch statement
func (cs *CatchStatement) String() string {
	var sb strings.Builder
	sb.WriteString("catch (")
	if cs.Exception != nil {
		sb.WriteString(cs.Exception.String())
	}
	sb.WriteString(") ")
	sb.WriteString(cs.Block.String())
	return sb.String()
}

// FinallyStatement represents a finally statement
type FinallyStatement struct {
	Token lexer.Token // The 'finally' token
	Block *BlockStatement
}

func (fs *FinallyStatement) statementNode() {}

// TokenLiteral returns the token literal
func (fs *FinallyStatement) TokenLiteral() string {
	return fs.Token.Literal
}

// String returns a string representation of the finally statement
func (fs *FinallyStatement) String() string {
	var sb strings.Builder
	sb.WriteString("finally ")
	sb.WriteString(fs.Block.String())
	return sb.String()
}

// ThrowStatement represents a throw statement
type ThrowStatement struct {
	Token   lexer.Token // The 'throw' token
	ErrExpr Expression
}

func (ts *ThrowStatement) statementNode() {}

// TokenLiteral returns the token literal
func (ts *ThrowStatement) TokenLiteral() string {
	return ts.Token.Literal
}

// String returns a string representation of the throw statement
func (ts *ThrowStatement) String() string {
	var sb strings.Builder
	sb.WriteString("throw ")
	if ts.ErrExpr != nil {
		sb.WriteString(ts.ErrExpr.String())
	}
	sb.WriteString(";")
	return sb.String()
}

// ImportStatement represents an import statement
type ImportStatement struct {
	Token lexer.Token    // The 'import' token
	Name  *Identifier    // Default import name (import math from ...)
	Path  *StringLiteral // Module path
	Alias *Identifier    // Namespace alias (import * as math from ...)
	Names []*Identifier  // Destructured names (import { add, sub } from ...)
}

func (is *ImportStatement) statementNode() {}

// TokenLiteral returns the token literal
func (is *ImportStatement) TokenLiteral() string {
	return is.Token.Literal
}

// String returns a string representation of the import statement
func (is *ImportStatement) String() string {
	var sb strings.Builder
	sb.WriteString("import ")

	if is.Alias != nil {
		sb.WriteString("* as ")
		sb.WriteString(is.Alias.String())
		sb.WriteString(" from ")
	} else if len(is.Names) > 0 {
		sb.WriteString("{ ")
		names := make([]string, len(is.Names))
		for i, n := range is.Names {
			names[i] = n.String()
		}
		sb.WriteString(strings.Join(names, ", "))
		sb.WriteString(" } from ")
	} else if is.Name != nil {
		sb.WriteString(is.Name.String())
		sb.WriteString(" from ")
	}

	sb.WriteString(is.Path.String())
	sb.WriteString(";")
	return sb.String()
}

// ExportStatement represents an export statement
type ExportStatement struct {
	Token      lexer.Token // The 'export' token
	Exportable Statement
}

func (es *ExportStatement) statementNode() {}

// TokenLiteral returns the token literal
func (es *ExportStatement) TokenLiteral() string {
	return es.Token.Literal
}

// String returns a string representation of the export statement
func (es *ExportStatement) String() string {
	var sb strings.Builder
	sb.WriteString("export ")
	sb.WriteString(es.Exportable.String())
	return sb.String()
}

// ClassStatement represents a class declaration
type ClassStatement struct {
	Token      lexer.Token // The 'class' token
	Name       *Identifier
	SuperClass *Identifier
	Methods    []*FunctionLiteral
	Fields     []*VarStatement
}

func (cs *ClassStatement) statementNode() {}

// TokenLiteral returns the token literal
func (cs *ClassStatement) TokenLiteral() string {
	return cs.Token.Literal
}

// String returns a string representation of the class statement
func (cs *ClassStatement) String() string {
	var sb strings.Builder
	sb.WriteString("class ")
	sb.WriteString(cs.Name.String())
	if cs.SuperClass != nil {
		sb.WriteString(" extends ")
		sb.WriteString(cs.SuperClass.String())
	}
	sb.WriteString(" { ")
	for _, m := range cs.Methods {
		sb.WriteString(m.String())
		sb.WriteString(" ")
	}
	for _, f := range cs.Fields {
		sb.WriteString(f.String())
		sb.WriteString(" ")
	}
	sb.WriteString("}")
	return sb.String()
}

// NewExpression represents a new expression for object instantiation
type NewExpression struct {
	Token     lexer.Token // The 'new' token
	Class     Expression
	Arguments []Expression
}

func (ne *NewExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (ne *NewExpression) TokenLiteral() string {
	return ne.Token.Literal
}

// String returns a string representation of the new expression
func (ne *NewExpression) String() string {
	var sb strings.Builder
	args := make([]string, len(ne.Arguments))
	for i, arg := range ne.Arguments {
		args[i] = arg.String()
	}
	sb.WriteString("new ")
	sb.WriteString(ne.Class.String())
	sb.WriteString("(")
	sb.WriteString(strings.Join(args, ", "))
	sb.WriteString(")")
	return sb.String()
}

// ThisExpression represents the 'this' keyword
type ThisExpression struct {
	Token lexer.Token // The 'this' token
}

func (te *ThisExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (te *ThisExpression) TokenLiteral() string {
	return te.Token.Literal
}

// String returns a string representation of the this expression
func (te *ThisExpression) String() string {
	return "this"
}

// SuperExpression represents the 'super' keyword
type SuperExpression struct {
	Token lexer.Token // The 'super' token
}

func (se *SuperExpression) expressionNode() {}

func (se *SuperExpression) TokenLiteral() string {
	return se.Token.Literal
}

func (se *SuperExpression) String() string {
	return "super"
}

// SuperCallExpression represents a super.method() call
type SuperCallExpression struct {
	Token  lexer.Token // The 'super' token
	Method string
	Args   []Expression
}

func (sc *SuperCallExpression) expressionNode() {}

func (sc *SuperCallExpression) TokenLiteral() string {
	return sc.Token.Literal
}

func (sc *SuperCallExpression) String() string {
	var sb strings.Builder
	args := make([]string, len(sc.Args))
	for i, arg := range sc.Args {
		args[i] = arg.String()
	}
	sb.WriteString("super.")
	sb.WriteString(sc.Method)
	sb.WriteString("(")
	sb.WriteString(strings.Join(args, ", "))
	sb.WriteString(")")
	return sb.String()
}

// ForInStatement represents a for-in loop statement
type ForInStatement struct {
	Token    lexer.Token // The 'for' token
	Key      *Identifier
	Value    *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fis *ForInStatement) statementNode() {}

// TokenLiteral returns the token literal
func (fis *ForInStatement) TokenLiteral() string {
	return fis.Token.Literal
}

// String returns a string representation of the for-in statement
func (fis *ForInStatement) String() string {
	var sb strings.Builder
	sb.WriteString("for (")
	if fis.Key != nil {
		sb.WriteString(fis.Key.String())
		sb.WriteString(", ")
	}
	sb.WriteString(fis.Value.String())
	sb.WriteString(" in ")
	sb.WriteString(fis.Iterable.String())
	sb.WriteString(") ")
	sb.WriteString(fis.Body.String())
	return sb.String()
}

// CompoundAssignmentExpression represents compound assignment operators (+=, -=, etc.)
type CompoundAssignmentExpression struct {
	Token    lexer.Token // The operator token, e.g. +=, -=
	Left     Expression
	Operator string
	Right    Expression
}

func (cae *CompoundAssignmentExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (cae *CompoundAssignmentExpression) TokenLiteral() string {
	return cae.Token.Literal
}

// String returns a string representation of the compound assignment expression
func (cae *CompoundAssignmentExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(cae.Left.String())
	sb.WriteString(" ")
	sb.WriteString(cae.Operator)
	sb.WriteString(" ")
	sb.WriteString(cae.Right.String())
	sb.WriteString(")")
	return sb.String()
}

// PostfixExpression represents a postfix expression (++, --)
type PostfixExpression struct {
	Token    lexer.Token // The operator token, e.g. ++, --
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (pe *PostfixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

// String returns a string representation of the postfix expression
func (pe *PostfixExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(pe.Left.String())
	sb.WriteString(pe.Operator)
	sb.WriteString(")")
	return sb.String()
}

// ArrowFunctionLiteral represents an arrow function expression
type ArrowFunctionLiteral struct {
	Token      lexer.Token // The '=>' token
	Parameters []*Identifier
	Body       *BlockStatement
	Expression Expression // For single expression arrow functions
}

func (afl *ArrowFunctionLiteral) expressionNode() {}

// TokenLiteral returns the token literal
func (afl *ArrowFunctionLiteral) TokenLiteral() string {
	return afl.Token.Literal
}

// String returns a string representation of the arrow function literal
func (afl *ArrowFunctionLiteral) String() string {
	var sb strings.Builder
	params := make([]string, len(afl.Parameters))
	for i, p := range afl.Parameters {
		params[i] = p.String()
	}
	sb.WriteString("(")
	sb.WriteString(strings.Join(params, ", "))
	sb.WriteString(") => ")
	if afl.Body != nil {
		sb.WriteString(afl.Body.String())
	} else if afl.Expression != nil {
		sb.WriteString(afl.Expression.String())
	}
	return sb.String()
}

// SpreadExpression represents spread operator (...)
type SpreadExpression struct {
	Token      lexer.Token // The '...' token
	Expression Expression
}

func (se *SpreadExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (se *SpreadExpression) TokenLiteral() string {
	return se.Token.Literal
}

// String returns a string representation of the spread expression
func (se *SpreadExpression) String() string {
	return "..." + se.Expression.String()
}

// TernaryExpression represents a ternary conditional expression
type TernaryExpression struct {
	Token       lexer.Token // The '?' token
	Condition   Expression
	Consequent  Expression
	Alternative Expression
}

func (te *TernaryExpression) expressionNode() {}

// TokenLiteral returns the token literal
func (te *TernaryExpression) TokenLiteral() string {
	return te.Token.Literal
}

// String returns a string representation of the ternary expression
func (te *TernaryExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(te.Condition.String())
	sb.WriteString(" ? ")
	sb.WriteString(te.Consequent.String())
	sb.WriteString(" : ")
	sb.WriteString(te.Alternative.String())
	sb.WriteString(")")
	return sb.String()
}

// Error helper function
func formatError(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
