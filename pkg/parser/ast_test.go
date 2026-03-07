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
				Token: lexer.Token{Type: lexer.TokenVar, Literal: "var"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"},
					Value: "x",
				},
				Value: &IntegerLiteral{
					Token: lexer.Token{Type: lexer.TokenInt, Literal: "42"},
					Value: 42,
				},
			},
		},
	}

	expected := "var x = 42;"
	if program.String() != expected {
		t.Errorf("program.String() = %q, want %q", program.String(), expected)
	}
}

func TestIdentifierString(t *testing.T) {
	ident := &Identifier{
		Token: lexer.Token{Type: lexer.TokenIdent, Literal: "foo"},
		Value: "foo",
	}
	if ident.String() != "foo" {
		t.Errorf("Identifier.String() = %q, want 'foo'", ident.String())
	}
}

func TestIntegerLiteralString(t *testing.T) {
	il := &IntegerLiteral{
		Token: lexer.Token{Type: lexer.TokenInt, Literal: "42"},
		Value: 42,
	}
	if il.String() != "42" {
		t.Errorf("IntegerLiteral.String() = %q, want '42'", il.String())
	}
}

func TestFloatLiteralString(t *testing.T) {
	fl := &FloatLiteral{
		Token: lexer.Token{Type: lexer.TokenFloat, Literal: "3.14"},
		Value: 3.14,
	}
	if fl.String() != "3.14" {
		t.Errorf("FloatLiteral.String() = %q, want '3.14'", fl.String())
	}
}

func TestStringLiteralString(t *testing.T) {
	sl := &StringLiteral{
		Token: lexer.Token{Type: lexer.TokenString, Literal: `"hello"`},
		Value: "hello",
	}
	if sl.String() != `"hello"` {
		t.Errorf(`StringLiteral.String() = %q, want '"hello"'`, sl.String())
	}
}

func TestBooleanLiteralString(t *testing.T) {
	tests := []struct {
		literal *BooleanLiteral
		want    string
	}{
		{
			literal: &BooleanLiteral{
				Token: lexer.Token{Type: lexer.TokenTrue, Literal: "true"},
				Value: true,
			},
			want: "true",
		},
		{
			literal: &BooleanLiteral{
				Token: lexer.Token{Type: lexer.TokenFalse, Literal: "false"},
				Value: false,
			},
			want: "false",
		},
	}

	for _, tt := range tests {
		if got := tt.literal.String(); got != tt.want {
			t.Errorf("BooleanLiteral.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestNullLiteralString(t *testing.T) {
	nl := &NullLiteral{
		Token: lexer.Token{Type: lexer.TokenNull, Literal: "null"},
	}
	if nl.String() != "null" {
		t.Errorf("NullLiteral.String() = %q, want 'null'", nl.String())
	}
}

func TestArrayLiteralString(t *testing.T) {
	al := &ArrayLiteral{
		Token: lexer.Token{Type: lexer.TokenLBracket, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "2"}, Value: 2},
			&IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "3"}, Value: 3},
		},
	}
	expected := "[1, 2, 3]"
	if al.String() != expected {
		t.Errorf("ArrayLiteral.String() = %q, want %q", al.String(), expected)
	}
}

func TestMapLiteralString(t *testing.T) {
	ml := &MapLiteral{
		Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
		Pairs: map[Expression]Expression{
			&StringLiteral{Token: lexer.Token{Type: lexer.TokenString, Literal: `"key"`}, Value: "key"}: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "42"}, Value: 42},
		},
	}
	// Map order is not guaranteed, so we just check it contains the expected parts
	got := ml.String()
	if got[0] != '{' || got[len(got)-1] != '}' {
		t.Errorf("MapLiteral.String() = %q, should be enclosed in braces", got)
	}
}

func TestPrefixExpressionString(t *testing.T) {
	tests := []struct {
		expr     *PrefixExpression
		expected string
	}{
		{
			expr: &PrefixExpression{
				Token:    lexer.Token{Type: lexer.TokenMinus, Literal: "-"},
				Operator: "-",
				Right:    &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "5"}, Value: 5},
			},
			expected: "(-5)",
		},
		{
			expr: &PrefixExpression{
				Token:    lexer.Token{Type: lexer.TokenNot, Literal: "!"},
				Operator: "!",
				Right:    &BooleanLiteral{Token: lexer.Token{Type: lexer.TokenTrue, Literal: "true"}, Value: true},
			},
			expected: "(!true)",
		},
	}

	for _, tt := range tests {
		if got := tt.expr.String(); got != tt.expected {
			t.Errorf("PrefixExpression.String() = %q, want %q", got, tt.expected)
		}
	}
}

func TestInfixExpressionString(t *testing.T) {
	expr := &InfixExpression{
		Token:    lexer.Token{Type: lexer.TokenPlus, Literal: "+"},
		Left:     &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "5"}, Value: 5},
		Operator: "+",
		Right:    &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "10"}, Value: 10},
	}
	expected := "(5 + 10)"
	if got := expr.String(); got != expected {
		t.Errorf("InfixExpression.String() = %q, want %q", got, expected)
	}
}

func TestCallExpressionString(t *testing.T) {
	ce := &CallExpression{
		Token:    lexer.Token{Type: lexer.TokenLParen, Literal: "("},
		Function: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "add"}, Value: "add"},
		Arguments: []Expression{
			&IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "2"}, Value: 2},
		},
	}
	expected := "add(1, 2)"
	if ce.String() != expected {
		t.Errorf("CallExpression.String() = %q, want %q", ce.String(), expected)
	}
}

func TestIndexExpressionString(t *testing.T) {
	ie := &IndexExpression{
		Token: lexer.Token{Type: lexer.TokenLBracket, Literal: "["},
		Left:  &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "arr"}, Value: "arr"},
		Index: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "0"}, Value: 0},
	}
	expected := "(arr[0])"
	if ie.String() != expected {
		t.Errorf("IndexExpression.String() = %q, want %q", ie.String(), expected)
	}
}

func TestDotExpressionString(t *testing.T) {
	de := &DotExpression{
		Token:    lexer.Token{Type: lexer.TokenDot, Literal: "."},
		Object:   &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "obj"}, Value: "obj"},
		Property: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "prop"}, Value: "prop"},
	}
	expected := "(obj.prop)"
	if de.String() != expected {
		t.Errorf("DotExpression.String() = %q, want %q", de.String(), expected)
	}
}

func TestAssignmentExpressionString(t *testing.T) {
	ae := &AssignmentExpression{
		Token: lexer.Token{Type: lexer.TokenAssign, Literal: "="},
		Left:  &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
		Value: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "10"}, Value: 10},
	}
	expected := "(x = 10)"
	if ae.String() != expected {
		t.Errorf("AssignmentExpression.String() = %q, want %q", ae.String(), expected)
	}
}

func TestFunctionLiteralString(t *testing.T) {
	fn := &FunctionLiteral{
		Token: lexer.Token{Type: lexer.TokenFunc, Literal: "func"},
		Parameters: []*Identifier{
			{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "a"}, Value: "a"},
			{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "b"}, Value: "b"},
		},
		Body: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ReturnStatement{
					Token: lexer.Token{Type: lexer.TokenReturn, Literal: "return"},
					ReturnValue: &InfixExpression{
						Token:    lexer.Token{Type: lexer.TokenPlus, Literal: "+"},
						Left:     &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "a"}, Value: "a"},
						Operator: "+",
						Right:    &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "b"}, Value: "b"},
					},
				},
			},
		},
	}
	expected := "func(a, b) { return (a + b); }"
	if fn.String() != expected {
		t.Errorf("FunctionLiteral.String() = %q, want %q", fn.String(), expected)
	}
}

func TestVarStatementString(t *testing.T) {
	vs := &VarStatement{
		Token: lexer.Token{Type: lexer.TokenVar, Literal: "var"},
		Name:  &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
		Value: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "42"}, Value: 42},
	}
	expected := "var x = 42;"
	if vs.String() != expected {
		t.Errorf("VarStatement.String() = %q, want %q", vs.String(), expected)
	}
}

func TestConstStatementString(t *testing.T) {
	cs := &ConstStatement{
		Token: lexer.Token{Type: lexer.TokenConst, Literal: "const"},
		Name:  &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "PI"}, Value: "PI"},
		Value: &FloatLiteral{Token: lexer.Token{Type: lexer.TokenFloat, Literal: "3.14"}, Value: 3.14},
	}
	expected := "const PI = 3.14;"
	if cs.String() != expected {
		t.Errorf("ConstStatement.String() = %q, want %q", cs.String(), expected)
	}
}

func TestReturnStatementString(t *testing.T) {
	rs := &ReturnStatement{
		Token:       lexer.Token{Type: lexer.TokenReturn, Literal: "return"},
		ReturnValue: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "42"}, Value: 42},
	}
	expected := "return 42;"
	if rs.String() != expected {
		t.Errorf("ReturnStatement.String() = %q, want %q", rs.String(), expected)
	}
}

func TestIfStatementString(t *testing.T) {
	is := &IfStatement{
		Token:     lexer.Token{Type: lexer.TokenIf, Literal: "if"},
		Condition: &BooleanLiteral{Token: lexer.Token{Type: lexer.TokenTrue, Literal: "true"}, Value: true},
		Consequence: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      lexer.Token{Type: lexer.TokenInt, Literal: "1"},
					Expression: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "1"}, Value: 1},
				},
			},
		},
	}
	expected := "if (true) { 1 }"
	if is.String() != expected {
		t.Errorf("IfStatement.String() = %q, want %q", is.String(), expected)
	}
}

func TestIfElseStatementString(t *testing.T) {
	is := &IfStatement{
		Token:     lexer.Token{Type: lexer.TokenIf, Literal: "if"},
		Condition: &BooleanLiteral{Token: lexer.Token{Type: lexer.TokenTrue, Literal: "true"}, Value: true},
		Consequence: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      lexer.Token{Type: lexer.TokenInt, Literal: "1"},
					Expression: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "1"}, Value: 1},
				},
			},
		},
		Alternative: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      lexer.Token{Type: lexer.TokenInt, Literal: "2"},
					Expression: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "2"}, Value: 2},
				},
			},
		},
	}
	expected := "if (true) { 1 } else { 2 }"
	if is.String() != expected {
		t.Errorf("IfStatement.String() = %q, want %q", is.String(), expected)
	}
}

func TestWhileStatementString(t *testing.T) {
	ws := &WhileStatement{
		Token:     lexer.Token{Type: lexer.TokenWhile, Literal: "while"},
		Condition: &BooleanLiteral{Token: lexer.Token{Type: lexer.TokenTrue, Literal: "true"}, Value: true},
		Body: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&BreakStatement{Token: lexer.Token{Type: lexer.TokenBreak, Literal: "break"}},
			},
		},
	}
	expected := "while (true) { break; }"
	if ws.String() != expected {
		t.Errorf("WhileStatement.String() = %q, want %q", ws.String(), expected)
	}
}

func TestForStatementString(t *testing.T) {
	fs := &ForStatement{
		Token: lexer.Token{Type: lexer.TokenFor, Literal: "for"},
		Init: &VarStatement{
			Token: lexer.Token{Type: lexer.TokenVar, Literal: "var"},
			Name:  &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "i"}, Value: "i"},
			Value: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "0"}, Value: 0},
		},
		Condition: &InfixExpression{
			Token:    lexer.Token{Type: lexer.TokenLT, Literal: "<"},
			Left:     &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "i"}, Value: "i"},
			Operator: "<",
			Right:    &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "10"}, Value: 10},
		},
		Update: &ExpressionStatement{
			Token: lexer.Token{Type: lexer.TokenIncrement, Literal: "++"},
			Expression: &PostfixExpression{
				Token:    lexer.Token{Type: lexer.TokenIncrement, Literal: "++"},
				Left:     &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "i"}, Value: "i"},
				Operator: "++",
			},
		},
		Body: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      lexer.Token{Type: lexer.TokenIdent, Literal: "i"},
					Expression: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "i"}, Value: "i"},
				},
			},
		},
	}
	expected := "for (var i = 0; (i < 10); (i++)) { i }"
	if fs.String() != expected {
		t.Errorf("ForStatement.String() = %q, want %q", fs.String(), expected)
	}
}

func TestBreakStatementString(t *testing.T) {
	bs := &BreakStatement{Token: lexer.Token{Type: lexer.TokenBreak, Literal: "break"}}
	expected := "break;"
	if bs.String() != expected {
		t.Errorf("BreakStatement.String() = %q, want %q", bs.String(), expected)
	}
}

func TestContinueStatementString(t *testing.T) {
	cs := &ContinueStatement{Token: lexer.Token{Type: lexer.TokenContinue, Literal: "continue"}}
	expected := "continue;"
	if cs.String() != expected {
		t.Errorf("ContinueStatement.String() = %q, want %q", cs.String(), expected)
	}
}

func TestExpressionStatementString(t *testing.T) {
	es := &ExpressionStatement{
		Token:      lexer.Token{Type: lexer.TokenInt, Literal: "42"},
		Expression: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "42"}, Value: 42},
	}
	expected := "42"
	if es.String() != expected {
		t.Errorf("ExpressionStatement.String() = %q, want %q", es.String(), expected)
	}
}

func TestBlockStatementString(t *testing.T) {
	bs := &BlockStatement{
		Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
		Statements: []Statement{
			&VarStatement{
				Token: lexer.Token{Type: lexer.TokenVar, Literal: "var"},
				Name:  &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
				Value: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "1"}, Value: 1},
			},
			&VarStatement{
				Token: lexer.Token{Type: lexer.TokenVar, Literal: "var"},
				Name:  &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "y"}, Value: "y"},
				Value: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "2"}, Value: 2},
			},
		},
	}
	expected := "{ var x = 1; var y = 2; }"
	if bs.String() != expected {
		t.Errorf("BlockStatement.String() = %q, want %q", bs.String(), expected)
	}
}

func TestNodeTokenLiteral(t *testing.T) {
	tests := []struct {
		name     string
		node     Node
		expected string
	}{
		{
			name: "Identifier",
			node: &Identifier{
				Token: lexer.Token{Type: lexer.TokenIdent, Literal: "foo"},
				Value: "foo",
			},
			expected: "foo",
		},
		{
			name: "IntegerLiteral",
			node: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.TokenInt, Literal: "42"},
				Value: 42,
			},
			expected: "42",
		},
		{
			name: "StringLiteral",
			node: &StringLiteral{
				Token: lexer.Token{Type: lexer.TokenString, Literal: `"hello"`},
				Value: "hello",
			},
			expected: `"hello"`,
		},
		{
			name: "BooleanLiteral",
			node: &BooleanLiteral{
				Token: lexer.Token{Type: lexer.TokenTrue, Literal: "true"},
				Value: true,
			},
			expected: "true",
		},
		{
			name: "VarStatement",
			node: &VarStatement{
				Token: lexer.Token{Type: lexer.TokenVar, Literal: "var"},
			},
			expected: "var",
		},
		{
			name: "ReturnStatement",
			node: &ReturnStatement{
				Token: lexer.Token{Type: lexer.TokenReturn, Literal: "return"},
			},
			expected: "return",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.TokenLiteral(); got != tt.expected {
				t.Errorf("TokenLiteral() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPostfixExpressionString(t *testing.T) {
	pe := &PostfixExpression{
		Token:    lexer.Token{Type: lexer.TokenIncrement, Literal: "++"},
		Left:     &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
		Operator: "++",
	}
	expected := "(x++)"
	if pe.String() != expected {
		t.Errorf("PostfixExpression.String() = %q, want %q", pe.String(), expected)
	}
}

func TestTernaryExpressionString(t *testing.T) {
	te := &TernaryExpression{
		Token:       lexer.Token{Type: lexer.TokenIdent, Literal: "?"},
		Condition:   &BooleanLiteral{Token: lexer.Token{Type: lexer.TokenTrue, Literal: "true"}, Value: true},
		Consequent:  &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "1"}, Value: 1},
		Alternative: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenInt, Literal: "2"}, Value: 2},
	}
	expected := "(true ? 1 : 2)"
	if te.String() != expected {
		t.Errorf("TernaryExpression.String() = %q, want %q", te.String(), expected)
	}
}

func TestArrowFunctionLiteralString(t *testing.T) {
	afl := &ArrowFunctionLiteral{
		Token: lexer.Token{Type: lexer.TokenArrow, Literal: "=>"},
		Parameters: []*Identifier{
			{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
		},
		Expression: &IntegerLiteral{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: 0},
	}
	expected := "(x) => x"
	if afl.String() != expected {
		t.Errorf("ArrowFunctionLiteral.String() = %q, want %q", afl.String(), expected)
	}
}

func TestNamedFunctionLiteralString(t *testing.T) {
	fn := &FunctionLiteral{
		Token: lexer.Token{Type: lexer.TokenFunc, Literal: "func"},
		Name:  "add",
		Parameters: []*Identifier{
			{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "a"}, Value: "a"},
			{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "b"}, Value: "b"},
		},
		Body: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ReturnStatement{
					Token: lexer.Token{Type: lexer.TokenReturn, Literal: "return"},
					ReturnValue: &InfixExpression{
						Token:    lexer.Token{Type: lexer.TokenPlus, Literal: "+"},
						Left:     &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "a"}, Value: "a"},
						Operator: "+",
						Right:    &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "b"}, Value: "b"},
					},
				},
			},
		},
	}
	expected := "func add(a, b) { return (a + b); }"
	if fn.String() != expected {
		t.Errorf("FunctionLiteral.String() = %q, want %q", fn.String(), expected)
	}
}

func TestForInStatementString(t *testing.T) {
	fis := &ForInStatement{
		Token: lexer.Token{Type: lexer.TokenFor, Literal: "for"},
		Value: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
		Iterable: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "arr"}, Value: "arr"},
		Body: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      lexer.Token{Type: lexer.TokenIdent, Literal: "print"},
					Expression: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
				},
			},
		},
	}
	expected := "for (x in arr) { x }"
	if fis.String() != expected {
		t.Errorf("ForInStatement.String() = %q, want %q", fis.String(), expected)
	}
}

func TestForInStatementWithKeyString(t *testing.T) {
	fis := &ForInStatement{
		Token: lexer.Token{Type: lexer.TokenFor, Literal: "for"},
		Key:   &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "i"}, Value: "i"},
		Value: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
		Iterable: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "arr"}, Value: "arr"},
		Body: &BlockStatement{
			Token: lexer.Token{Type: lexer.TokenLBrace, Literal: "{"},
			Statements: []Statement{
				&ExpressionStatement{
					Token:      lexer.Token{Type: lexer.TokenIdent, Literal: "print"},
					Expression: &Identifier{Token: lexer.Token{Type: lexer.TokenIdent, Literal: "x"}, Value: "x"},
				},
			},
		},
	}
	expected := "for (i, x in arr) { x }"
	if fis.String() != expected {
		t.Errorf("ForInStatement.String() = %q, want %q", fis.String(), expected)
	}
}
