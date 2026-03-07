// pkg/lexer/lexer_test.go
package lexer

import "testing"

func TestNextToken_BasicTokens(t *testing.T) {
	input := `=+(){},;`

	expected := []Token{
		{Type: TokenAssign, Literal: "=", Line: 1, Column: 1},
		{Type: TokenPlus, Literal: "+", Line: 1, Column: 2},
		{Type: TokenLParen, Literal: "(", Line: 1, Column: 3},
		{Type: TokenRParen, Literal: ")", Line: 1, Column: 4},
		{Type: TokenLBrace, Literal: "{", Line: 1, Column: 5},
		{Type: TokenRBrace, Literal: "}", Line: 1, Column: 6},
		{Type: TokenComma, Literal: ",", Line: 1, Column: 7},
		{Type: TokenSemicolon, Literal: ";", Line: 1, Column: 8},
		{Type: TokenEOF, Literal: "", Line: 1, Column: 9},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_Code(t *testing.T) {
	input := `var x = 5;
var y = 10;
var add = func(a, b) {
	return a + b;
};`

	expected := []Token{
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "x"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenInt, Literal: "5"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "y"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenInt, Literal: "10"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "add"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenFunc, Literal: "func"},
		{Type: TokenLParen, Literal: "("},
		{Type: TokenIdent, Literal: "a"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenIdent, Literal: "b"},
		{Type: TokenRParen, Literal: ")"},
		{Type: TokenLBrace, Literal: "{"},
		{Type: TokenReturn, Literal: "return"},
		{Type: TokenIdent, Literal: "a"},
		{Type: TokenPlus, Literal: "+"},
		{Type: TokenIdent, Literal: "b"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenRBrace, Literal: "}"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_Operators(t *testing.T) {
	input := `== != <= >= && || += -= ++ -- =>`

	expected := []Token{
		{Type: TokenEqual, Literal: "=="},
		{Type: TokenNotEqual, Literal: "!="},
		{Type: TokenLTE, Literal: "<="},
		{Type: TokenGTE, Literal: ">="},
		{Type: TokenAnd, Literal: "&&"},
		{Type: TokenOr, Literal: "||"},
		{Type: TokenPlusAssign, Literal: "+="},
		{Type: TokenMinusAssign, Literal: "-="},
		{Type: TokenIncrement, Literal: "++"},
		{Type: TokenDecrement, Literal: "--"},
		{Type: TokenArrow, Literal: "=>"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_Floats(t *testing.T) {
	input := `3.14 0.5 1e10 2.5e-3 1E+5`

	expected := []Token{
		{Type: TokenFloat, Literal: "3.14"},
		{Type: TokenFloat, Literal: "0.5"},
		{Type: TokenFloat, Literal: "1e10"},
		{Type: TokenFloat, Literal: "2.5e-3"},
		{Type: TokenFloat, Literal: "1E+5"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_Strings(t *testing.T) {
	input := `"hello" "world\n" "tab\t" "quote\"" "backslash\\" "mixed\n\t\\"`

	expected := []Token{
		{Type: TokenString, Literal: "hello"},
		{Type: TokenString, Literal: "world\n"},
		{Type: TokenString, Literal: "tab\t"},
		{Type: TokenString, Literal: "quote\""},
		{Type: TokenString, Literal: "backslash\\"},
		{Type: TokenString, Literal: "mixed\n\t\\"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `// this is a comment
var x = 5; // inline comment
/* multi-line
comment */
var y = 10;
/* another */ var z = 15;`

	expected := []Token{
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "x"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenInt, Literal: "5"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "y"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenInt, Literal: "10"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "z"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenInt, Literal: "15"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	input := `var const func return if else while for in break continue
switch case default try catch finally throw class extends new this
null true false import export`

	expected := []Token{
		{Type: TokenVar, Literal: "var"},
		{Type: TokenConst, Literal: "const"},
		{Type: TokenFunc, Literal: "func"},
		{Type: TokenReturn, Literal: "return"},
		{Type: TokenIf, Literal: "if"},
		{Type: TokenElse, Literal: "else"},
		{Type: TokenWhile, Literal: "while"},
		{Type: TokenFor, Literal: "for"},
		{Type: TokenIn, Literal: "in"},
		{Type: TokenBreak, Literal: "break"},
		{Type: TokenContinue, Literal: "continue"},
		{Type: TokenSwitch, Literal: "switch"},
		{Type: TokenCase, Literal: "case"},
		{Type: TokenDefault, Literal: "default"},
		{Type: TokenTry, Literal: "try"},
		{Type: TokenCatch, Literal: "catch"},
		{Type: TokenFinally, Literal: "finally"},
		{Type: TokenThrow, Literal: "throw"},
		{Type: TokenClass, Literal: "class"},
		{Type: TokenExtends, Literal: "extends"},
		{Type: TokenNew, Literal: "new"},
		{Type: TokenThis, Literal: "this"},
		{Type: TokenNull, Literal: "null"},
		{Type: TokenTrue, Literal: "true"},
		{Type: TokenFalse, Literal: "false"},
		{Type: TokenImport, Literal: "import"},
		{Type: TokenExport, Literal: "export"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_LineColumn(t *testing.T) {
	input := `var x = 5;
var y = 10;`

	expected := []Token{
		{Type: TokenVar, Literal: "var", Line: 1, Column: 1},
		{Type: TokenIdent, Literal: "x", Line: 1, Column: 5},
		{Type: TokenAssign, Literal: "=", Line: 1, Column: 7},
		{Type: TokenInt, Literal: "5", Line: 1, Column: 9},
		{Type: TokenSemicolon, Literal: ";", Line: 1, Column: 10},
		{Type: TokenVar, Literal: "var", Line: 2, Column: 1},
		{Type: TokenIdent, Literal: "y", Line: 2, Column: 5},
		{Type: TokenAssign, Literal: "=", Line: 2, Column: 7},
		{Type: TokenInt, Literal: "10", Line: 2, Column: 9},
		{Type: TokenSemicolon, Literal: ";", Line: 2, Column: 11},
		{Type: TokenEOF, Literal: "", Line: 2, Column: 12},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
		if tok.Line != exp.Line {
			t.Fatalf("tests[%d] - wrong line. expected=%d, got=%d",
				i, exp.Line, tok.Line)
		}
		if tok.Column != exp.Column {
			t.Fatalf("tests[%d] - wrong column. expected=%d, got=%d",
				i, exp.Column, tok.Column)
		}
	}
}

func TestNextToken_ArrayAndObject(t *testing.T) {
	input := `var arr = [1, 2, 3];
var obj = {x: 1, y: 2};`

	expected := []Token{
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "arr"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenLBracket, Literal: "["},
		{Type: TokenInt, Literal: "1"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenInt, Literal: "2"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenInt, Literal: "3"},
		{Type: TokenRBracket, Literal: "]"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenVar, Literal: "var"},
		{Type: TokenIdent, Literal: "obj"},
		{Type: TokenAssign, Literal: "="},
		{Type: TokenLBrace, Literal: "{"},
		{Type: TokenIdent, Literal: "x"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenInt, Literal: "1"},
		{Type: TokenComma, Literal: ","},
		{Type: TokenIdent, Literal: "y"},
		{Type: TokenColon, Literal: ":"},
		{Type: TokenInt, Literal: "2"},
		{Type: TokenRBrace, Literal: "}"},
		{Type: TokenSemicolon, Literal: ";"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNextToken_DotOperator(t *testing.T) {
	input := `obj.method()`

	expected := []Token{
		{Type: TokenIdent, Literal: "obj"},
		{Type: TokenDot, Literal: "."},
		{Type: TokenIdent, Literal: "method"},
		{Type: TokenLParen, Literal: "("},
		{Type: TokenRParen, Literal: ")"},
		{Type: TokenEOF, Literal: ""},
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Fatalf("tests[%d] - wrong token type. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Fatalf("tests[%d] - wrong literal. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}
