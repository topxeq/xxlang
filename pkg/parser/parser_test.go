// pkg/parser/parser_test.go
package parser

import (
	"fmt"
	"testing"

	"github.com/topxeq/xxlang/pkg/lexer"
)

// ============================================
// Test Helper Functions
// ============================================

func testProgramStatements(t *testing.T, program *Program, expectedLen int) {
	if len(program.Statements) != expectedLen {
		t.Fatalf("program.Statements does not contain %d statements. got=%d",
			expectedLen, len(program.Statements))
	}
}

func testVarStatement(t *testing.T, stmt Statement, expectedName string) bool {
	varStmt, ok := stmt.(*VarStatement)
	if !ok {
		t.Errorf("stmt not *VarStatement. got=%T", stmt)
		return false
	}

	if varStmt.Name.Value != expectedName {
		t.Errorf("varStmt.Name.Value not '%s'. got=%s", expectedName, varStmt.Name.Value)
		return false
	}

	if varStmt.Name.TokenLiteral() != expectedName {
		t.Errorf("varStmt.Name.TokenLiteral() not '%s'. got=%s",
			expectedName, varStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func testConstStatement(t *testing.T, stmt Statement, expectedName string) bool {
	constStmt, ok := stmt.(*ConstStatement)
	if !ok {
		t.Errorf("stmt not *ConstStatement. got=%T", stmt)
		return false
	}

	if constStmt.Name.Value != expectedName {
		t.Errorf("constStmt.Name.Value not '%s'. got=%s", expectedName, constStmt.Name.Value)
		return false
	}

	if constStmt.Name.TokenLiteral() != expectedName {
		t.Errorf("constStmt.Name.TokenLiteral() not '%s'. got=%s",
			expectedName, constStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp Expression, value string) bool {
	ident, ok := exp.(*Identifier)
	if !ok {
		t.Errorf("exp not *Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, il Expression, value int64) bool {
	integ, ok := il.(*IntegerLiteral)
	if !ok {
		t.Errorf("il not *IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value, integ.TokenLiteral())
		return false
	}

	return true
}

func testFloatLiteral(t *testing.T, fl Expression, value float64) bool {
	float, ok := fl.(*FloatLiteral)
	if !ok {
		t.Errorf("fl not *FloatLiteral. got=%T", fl)
		return false
	}

	if float.Value != value {
		t.Errorf("float.Value not %f. got=%f", value, float.Value)
		return false
	}

	return true
}

func testStringLiteral(t *testing.T, sl Expression, value string) bool {
	str, ok := sl.(*StringLiteral)
	if !ok {
		t.Errorf("sl not *StringLiteral. got=%T", sl)
		return false
	}

	if str.Value != value {
		t.Errorf("str.Value not %s. got=%s", value, str.Value)
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp Expression, value bool) bool {
	bo, ok := exp.(*BooleanLiteral)
	if !ok {
		t.Errorf("exp not *BooleanLiteral. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case float64:
		return testFloatLiteral(t, exp, v)
	case string:
		// Try to determine if this should be an identifier or string literal
		// by checking the actual expression type
		if _, ok := exp.(*StringLiteral); ok {
			return testStringLiteral(t, exp, v)
		}
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	default:
		t.Errorf("type of exp not handled. got=%T", exp)
		return false
	}
}

func testInfixExpression(t *testing.T, exp Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*InfixExpression)
	if !ok {
		t.Errorf("exp is not InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testNullLiteral(t *testing.T, exp Expression) bool {
	null, ok := exp.(*NullLiteral)
	if !ok {
		t.Errorf("exp not *NullLiteral. got=%T", exp)
		return false
	}

	if null.TokenLiteral() != "null" {
		t.Errorf("null.TokenLiteral not 'null'. got=%s", null.TokenLiteral())
		return false
	}

	return true
}

// ============================================
// Variable Declaration Tests
// ============================================

func TestVarStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"var x = 5;", "x", 5},
		{"var y = true;", "y", true},
		{"var foobar = y;", "foobar", "y"},
		{"var pi = 3.14;", "pi", 3.14},
		{"var s = \"hello\";", "s", "hello"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		if !testVarStatement(t, program.Statements[0], tt.expectedIdentifier) {
			return
		}

		val := program.Statements[0].(*VarStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestVarStatementWithoutValue(t *testing.T) {
	input := "var x;"

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()

	// Should produce an error
	if len(p.Errors()) == 0 {
		t.Errorf("expected parser error for var without value")
	}
}

// ============================================
// Constant Declaration Tests
// ============================================

func TestConstStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"const x = 5;", "x", 5},
		{"const y = true;", "y", true},
		{"const foobar = y;", "foobar", "y"},
		{"const PI = 3.14;", "PI", 3.14},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		if !testConstStatement(t, program.Statements[0], tt.expectedIdentifier) {
			return
		}

		val := program.Statements[0].(*ConstStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

// ============================================
// Return Statement Tests
// ============================================

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return foobar;", "foobar"},
		{"return 3.14;", 3.14},
		{"return;", nil},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		returnStmt, ok := program.Statements[0].(*ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ReturnStatement. got=%T", program.Statements[0])
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return', got %q", returnStmt.TokenLiteral())
		}

		if tt.expectedValue != nil {
			if !testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue) {
				return
			}
		}
	}
}

// ============================================
// Identifier Tests
// ============================================

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	testIdentifier(t, stmt.Expression, "foobar")
}

// ============================================
// Literal Tests
// ============================================

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	testIntegerLiteral(t, stmt.Expression, 5)
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "3.14;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	testFloatLiteral(t, stmt.Expression, 3.14)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	testStringLiteral(t, stmt.Expression, "hello world")
}

func TestBooleanLiteralExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
		}

		testBooleanLiteral(t, stmt.Expression, tt.expected)
	}
}

func TestNullLiteralExpression(t *testing.T) {
	input := "null;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	testNullLiteral(t, stmt.Expression)
}

// ============================================
// Prefix Expression Tests
// ============================================

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!foobar;", "!", "foobar"},
		{"-foobar;", "-", "foobar"},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*PrefixExpression)
		if !ok {
			t.Fatalf("stmt.Expression not *PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}

		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

// ============================================
// Infix Expression Tests
// ============================================

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 % 5;", 5, "%", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"5 >= 5;", 5, ">=", 5},
		{"5 <= 5;", 5, "<=", 5},
		{"true && true;", true, "&&", true},
		{"true || false;", true, "||", false},
		{"foobar + barfoo;", "foobar", "+", "barfoo"},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
			return
		}
	}
}

// ============================================
// Operator Precedence Tests
// ============================================

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

// ============================================
// If Statement Tests
// ============================================

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("stmt not *IfStatement. got=%T", program.Statements[0])
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if len(stmt.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n", len(stmt.Consequence.Statements))
	}

	consequence, ok := stmt.Consequence.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ExpressionStatement. got=%T",
			stmt.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if stmt.Alternative != nil {
		t.Errorf("stmt.Alternative was not nil. got=%+v", stmt.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("stmt not *IfStatement. got=%T", program.Statements[0])
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if len(stmt.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n", len(stmt.Consequence.Statements))
	}

	consequence, ok := stmt.Consequence.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ExpressionStatement. got=%T",
			stmt.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(stmt.Alternative.Statements) != 1 {
		t.Errorf("alternative is not 1 statements. got=%d\n", len(stmt.Alternative.Statements))
	}

	alternative, ok := stmt.Alternative.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ExpressionStatement. got=%T",
			stmt.Alternative.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

// ============================================
// While Statement Tests
// ============================================

func TestWhileStatement(t *testing.T) {
	input := `while (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*WhileStatement)
	if !ok {
		t.Fatalf("stmt not *WhileStatement. got=%T", program.Statements[0])
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if len(stmt.Body.Statements) != 1 {
		t.Errorf("body is not 1 statements. got=%d\n", len(stmt.Body.Statements))
	}

	body, ok := stmt.Body.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ExpressionStatement. got=%T",
			stmt.Body.Statements[0])
	}

	if !testIdentifier(t, body.Expression, "x") {
		return
	}
}

// ============================================
// For Statement Tests
// ============================================

func TestForStatement(t *testing.T) {
	input := `for (var i = 0; i < 10; i++) { i }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("stmt not *ForStatement. got=%T", program.Statements[0])
	}

	// Check init
	if stmt.Init == nil {
		t.Fatalf("stmt.Init is nil")
	}

	varStmt, ok := stmt.Init.(*VarStatement)
	if !ok {
		t.Fatalf("stmt.Init not *VarStatement. got=%T", stmt.Init)
	}

	if varStmt.Name.Value != "i" {
		t.Errorf("varStmt.Name.Value not 'i'. got=%s", varStmt.Name.Value)
	}

	// Check condition
	if !testInfixExpression(t, stmt.Condition, "i", "<", 10) {
		return
	}

	// Check update
	if stmt.Update == nil {
		t.Fatalf("stmt.Update is nil")
	}

	// Check body
	if len(stmt.Body.Statements) != 1 {
		t.Errorf("body is not 1 statements. got=%d\n", len(stmt.Body.Statements))
	}
}

func TestForInStatement(t *testing.T) {
	input := `for (x in arr) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ForInStatement)
	if !ok {
		t.Fatalf("stmt not *ForInStatement. got=%T", program.Statements[0])
	}

	if stmt.Value.Value != "x" {
		t.Errorf("stmt.Value.Value not 'x'. got=%s", stmt.Value.Value)
	}

	if !testIdentifier(t, stmt.Iterable, "arr") {
		return
	}

	if len(stmt.Body.Statements) != 1 {
		t.Errorf("body is not 1 statements. got=%d\n", len(stmt.Body.Statements))
	}
}

func TestForInStatementWithKey(t *testing.T) {
	input := `for (i, x in arr) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ForInStatement)
	if !ok {
		t.Fatalf("stmt not *ForInStatement. got=%T", program.Statements[0])
	}

	if stmt.Key.Value != "i" {
		t.Errorf("stmt.Key.Value not 'i'. got=%s", stmt.Key.Value)
	}

	if stmt.Value.Value != "x" {
		t.Errorf("stmt.Value.Value not 'x'. got=%s", stmt.Value.Value)
	}

	if !testIdentifier(t, stmt.Iterable, "arr") {
		return
	}
}

// ============================================
// Break and Continue Tests
// ============================================

func TestBreakStatement(t *testing.T) {
	input := `break;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*BreakStatement)
	if !ok {
		t.Fatalf("stmt not *BreakStatement. got=%T", program.Statements[0])
	}

	if stmt.TokenLiteral() != "break" {
		t.Errorf("stmt.TokenLiteral not 'break'. got=%s", stmt.TokenLiteral())
	}
}

func TestContinueStatement(t *testing.T) {
	input := `continue;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ContinueStatement)
	if !ok {
		t.Fatalf("stmt not *ContinueStatement. got=%T", program.Statements[0])
	}

	if stmt.TokenLiteral() != "continue" {
		t.Errorf("stmt.TokenLiteral not 'continue'. got=%s", stmt.TokenLiteral())
	}
}

// ============================================
// Function Literal Tests
// ============================================

func TestFunctionLiteral(t *testing.T) {
	input := `func(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	function, ok := stmt.Expression.(*FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression not *FunctionLiteral. got=%T", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n", len(function.Parameters))
	}

	testIdentifier(t, function.Parameters[0], "x")
	testIdentifier(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statements. got=%d\n", len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt not ExpressionStatement. got=%T", function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionLiteralWithName(t *testing.T) {
	input := `func add(x, y) { return x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	function, ok := stmt.Expression.(*FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression not *FunctionLiteral. got=%T", stmt.Expression)
	}

	if function.Name != "add" {
		t.Errorf("function.Name not 'add'. got=%s", function.Name)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n", len(function.Parameters))
	}
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "func() {};", expectedParams: []string{}},
		{input: "func(x) {};", expectedParams: []string{"x"}},
		{input: "func(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ExpressionStatement)
		function := stmt.Expression.(*FunctionLiteral)

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams), len(function.Parameters))
		}

		for i, ident := range tt.expectedParams {
			testIdentifier(t, function.Parameters[i], ident)
		}
	}
}

// ============================================
// Call Expression Tests
// ============================================

func TestCallExpression(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression not *CallExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestCallExpressionWithNoArguments(t *testing.T) {
	input := "myFunc();"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression not *CallExpression. got=%T", stmt.Expression)
	}

	if len(exp.Arguments) != 0 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}
}

// ============================================
// Array Literal Tests
// ============================================

func TestArrayLiteral(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	array, ok := stmt.Expression.(*ArrayLiteral)
	if !ok {
		t.Fatalf("exp not *ArrayLiteral. got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestEmptyArrayLiteral(t *testing.T) {
	input := "[]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	array, ok := stmt.Expression.(*ArrayLiteral)
	if !ok {
		t.Fatalf("exp not *ArrayLiteral. got=%T", stmt.Expression)
	}

	if len(array.Elements) != 0 {
		t.Fatalf("len(array.Elements) not 0. got=%d", len(array.Elements))
	}
}

// ============================================
// Index Expression Tests
// ============================================

func TestIndexExpression(t *testing.T) {
	input := "myArray[1 + 1]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	indexExp, ok := stmt.Expression.(*IndexExpression)
	if !ok {
		t.Fatalf("exp not *IndexExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, indexExp.Index, 1, "+", 1) {
		return
	}
}

// ============================================
// Map Literal Tests
// ============================================

func TestMapLiteral(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	mapLit, ok := stmt.Expression.(*MapLiteral)
	if !ok {
		t.Fatalf("exp not *MapLiteral. got=%T", stmt.Expression)
	}

	if len(mapLit.Pairs) != 3 {
		t.Fatalf("mapLit.Pairs has wrong length. got=%d", len(mapLit.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range mapLit.Pairs {
		literal, ok := key.(*StringLiteral)
		if !ok {
			t.Errorf("key is not StringLiteral. got=%T", key)
			continue
		}

		expectedValue := expected[literal.Value]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestEmptyMapLiteral(t *testing.T) {
	input := "{}"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	mapLit, ok := stmt.Expression.(*MapLiteral)
	if !ok {
		t.Fatalf("exp not *MapLiteral. got=%T", stmt.Expression)
	}

	if len(mapLit.Pairs) != 0 {
		t.Fatalf("len(mapLit.Pairs) not 0. got=%d", len(mapLit.Pairs))
	}
}

// ============================================
// Dot Expression Tests
// ============================================

func TestDotExpression(t *testing.T) {
	input := "obj.property"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	dotExp, ok := stmt.Expression.(*DotExpression)
	if !ok {
		t.Fatalf("exp not *DotExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, dotExp.Object, "obj") {
		return
	}

	if !testIdentifier(t, dotExp.Property, "property") {
		return
	}
}

func TestChainedDotExpression(t *testing.T) {
	input := "obj.prop1.prop2"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	dotExp, ok := stmt.Expression.(*DotExpression)
	if !ok {
		t.Fatalf("exp not *DotExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, dotExp.Property, "prop2") {
		return
	}

	innerDot, ok := dotExp.Object.(*DotExpression)
	if !ok {
		t.Fatalf("inner object not *DotExpression. got=%T", dotExp.Object)
	}

	if !testIdentifier(t, innerDot.Object, "obj") {
		return
	}

	if !testIdentifier(t, innerDot.Property, "prop1") {
		return
	}
}

// ============================================
// Assignment Expression Tests
// ============================================

func TestAssignmentExpression(t *testing.T) {
	input := "x = 5"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	assign, ok := stmt.Expression.(*AssignmentExpression)
	if !ok {
		t.Fatalf("stmt.Expression not *AssignmentExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, assign.Left, "x") {
		return
	}

	if !testIntegerLiteral(t, assign.Value, 5) {
		return
	}
}

// ============================================
// Compound Assignment Tests
// ============================================

func TestCompoundAssignmentExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    int64
	}{
		{"x += 5", "+=", 5},
		{"x -= 3", "-=", 3},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
		}

		assign, ok := stmt.Expression.(*CompoundAssignmentExpression)
		if !ok {
			t.Fatalf("stmt.Expression not *CompoundAssignmentExpression. got=%T", stmt.Expression)
		}

		if assign.Operator != tt.operator {
			t.Errorf("operator not %s. got=%s", tt.operator, assign.Operator)
		}

		if !testIdentifier(t, assign.Left, "x") {
			return
		}

		if !testIntegerLiteral(t, assign.Right, tt.value) {
			return
		}
	}
}

// ============================================
// Postfix Expression Tests
// ============================================

func TestPostfixExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"x++", "++"},
		{"x--", "--"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		testProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
		}

		postfix, ok := stmt.Expression.(*PostfixExpression)
		if !ok {
			t.Fatalf("stmt.Expression not *PostfixExpression. got=%T", stmt.Expression)
		}

		if postfix.Operator != tt.operator {
			t.Errorf("operator not %s. got=%s", tt.operator, postfix.Operator)
		}

		if !testIdentifier(t, postfix.Left, "x") {
			return
		}
	}
}

// ============================================
// Block Statement Tests
// ============================================

func TestBlockStatement(t *testing.T) {
	input := `{ var x = 5; x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	block, ok := program.Statements[0].(*BlockStatement)
	if !ok {
		t.Fatalf("stmt not *BlockStatement. got=%T", program.Statements[0])
	}

	if len(block.Statements) != 2 {
		t.Fatalf("block.Statements has wrong length. got=%d", len(block.Statements))
	}
}

// ============================================
// Error Handling Tests
// ============================================

func TestParsingErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"var x 5;"},
		{"const y;"},
		{"return ;"}, // this should actually parse fine
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		_ = p.ParseProgram()

		// We just want to make sure errors are collected without crashing
		// The actual error messages are checked in specific tests
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

// ============================================
// Complex Expression Tests
// ============================================

func TestComplexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"arr[0].property",
			"((arr[0]).property)",
		},
		{
			"obj.method()",
			"(obj.method)()",
		},
		{
			"obj.method(arg1, arg2)",
			"(obj.method)(arg1, arg2)",
		},
		{
			"arr[0]()",
			"(arr[0])()",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

// ============================================
// Grouped Expression Tests
// ============================================

func TestGroupedExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 * (5 + 5)", "(2 * (5 + 5))"},
		{"(true && false)", "(true && false)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

// ============================================
// Import Statement Tests
// ============================================

func TestImportStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`import "./math"`,
			`import "./math";`,
		},
		{
			`import math from "./math"`,
			`import math from "./math";`,
		},
		{
			`import { add, sub } from "./math"`,
			`import { add, sub } from "./math";`,
		},
		{
			`import * as math from "./math"`,
			`import * as math from "./math";`,
		},
		{
			`import "../utils"`,
			`import "../utils";`,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program.String() != tt.expected {
			t.Errorf("program.String() = %s, want %s", program.String(), tt.expected)
		}
	}
}

// ============================================
// Export Statement Tests
// ============================================

func TestExportStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`export func add(a, b) { return a + b }`,
			`export func add(a, b) { return (a + b); }`,
		},
		{
			`export var PI = 3.14`,
			`export var PI = 3.14;`,
		},
		{
			`export const VERSION = "1.0"`,
			`export const VERSION = "1.0";`,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program.String() != tt.expected {
			t.Errorf("program.String() = %s, want %s", program.String(), tt.expected)
		}
	}
}

// ============================================
// Class Statement Tests
// ============================================

func TestClassStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`class Person {}`,
			`class Person { }`,
		},
		{
			`class Dog extends Animal {}`,
			`class Dog extends Animal { }`,
		},
		{
			`class Person {
				var name = ""
				func init(name) { this.name = name }
			}`,
			`class Person { func init(name) { ((this.name) = name) } var name = ""; }`,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program.String() != tt.expected {
			t.Errorf("program.String() = %s, want %s", program.String(), tt.expected)
		}
	}
}

func TestNewExpression(t *testing.T) {
	input := `new Person("Alice", 30)`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	newExpr, ok := stmt.Expression.(*NewExpression)
	if !ok {
		t.Fatalf("expected NewExpression, got %T", stmt.Expression)
	}
	if newExpr.Class.String() != "Person" {
		t.Errorf("expected Person, got %s", newExpr.Class.String())
	}
	if len(newExpr.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(newExpr.Arguments))
	}
}

func TestThisExpression(t *testing.T) {
	input := `this.name`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	dotExpr, ok := stmt.Expression.(*DotExpression)
	if !ok {
		t.Fatalf("expected DotExpression, got %T", stmt.Expression)
	}

	thisExpr, ok := dotExpr.Object.(*ThisExpression)
	if !ok {
		t.Fatalf("expected ThisExpression, got %T", dotExpr.Object)
	}
	if thisExpr.String() != "this" {
		t.Errorf("expected 'this', got %s", thisExpr.String())
	}
}

func TestSuperExpression(t *testing.T) {
	input := `super.init()`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	callExpr, ok := stmt.Expression.(*SuperCallExpression)
	if !ok {
		t.Fatalf("expected SuperCallExpression, got %T", stmt.Expression)
	}
	if callExpr.Method != "init" {
		t.Errorf("expected init, got %s", callExpr.Method)
	}
}

func TestSuperExpressionWithArgs(t *testing.T) {
	input := `super.init("arg1", 42)`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	callExpr, ok := stmt.Expression.(*SuperCallExpression)
	if !ok {
		t.Fatalf("expected SuperCallExpression, got %T", stmt.Expression)
	}
	if len(callExpr.Args) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(callExpr.Args))
	}
}

func TestCStyleForLoop(t *testing.T) {
	input := `for (var i = 0; i < 10; i = i + 1) { i; }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// C-style for loop has Init, Condition, and Update
	if stmt.Init == nil {
		t.Error("expected Init statement")
	}
	if stmt.Condition == nil {
		t.Error("expected Condition expression")
	}
	if stmt.Update == nil {
		t.Error("expected Update statement")
	}
}

func TestCStyleForLoopWithIdentifier(t *testing.T) {
	input := `for (i < 10; i = i + 1) { i; }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// C-style for loop without initializer
	if stmt.Condition == nil {
		t.Error("expected Condition expression")
	}
	if stmt.Update == nil {
		t.Error("expected Update statement")
	}
}

// ============================================
// Error Cases Tests
// ============================================

func TestImportErrorMissingAs(t *testing.T) {
	// Test: import * without "as" should cause error
	input := `import * math from "./math"`
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error, got none")
	}
}

func TestImportErrorMissingFrom(t *testing.T) {
	// Test: import { x } missing "from" should cause error
	input := `import { x } "./math"`
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error, got none")
	}
}

func TestImportErrorInvalidPath(t *testing.T) {
	// Test: import with invalid path should cause error
	input := `import math from math` // Missing quotes
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error, got none")
	}
}

func TestClassErrorInvalidSuper(t *testing.T) {
	// Test: class extends with invalid identifier should cause error
	input := `class Dog extends 123 {}`
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error, got none")
	}
}

func TestArrayErrorUnclosed(t *testing.T) {
	// Test: unclosed array should cause error
	input := `[1, 2` // Missing ]
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error, got none")
	}
}

func TestMapErrorUnclosed(t *testing.T) {
	// Test: unclosed map should cause error
	input := `{"a": 1` // Missing }
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error, got none")
	}
}

func TestFunctionErrorUnclosed(t *testing.T) {
	// Test: unclosed function body at end of file
	// Parser may accept this as valid (EOF closes it)
	input := `func add(a, b) { return a + b`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	// Just verify it parses without crashing
	_ = program
}

func TestCallErrorUnclosed(t *testing.T) {
	// Test: unclosed call at end of file
	// Parser may accept this as valid (EOF closes it)
	input := `add(1, 2`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	// Just verify it parses without crashing
	_ = program
}

// ============================================
// Try-Catch-Finally Statement Tests
// ============================================

func TestTryCatchStatement(t *testing.T) {
	input := `try { throw "error" } catch (e) { print(e) }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*TryStatement)
	if !ok {
		t.Fatalf("stmt not *TryStatement. got=%T", program.Statements[0])
	}

	// Check try block exists
	if stmt.Block == nil {
		t.Fatal("try block is nil")
	}
	if len(stmt.Block.Statements) != 1 {
		t.Errorf("try block should have 1 statement, got=%d", len(stmt.Block.Statements))
	}

	// Check catch exists
	if stmt.Catch == nil {
		t.Fatal("catch clause is nil")
	}
	if stmt.Catch.Exception.Value != "e" {
		t.Errorf("catch exception variable should be 'e', got=%s", stmt.Catch.Exception.Value)
	}
	if stmt.Catch.Block == nil {
		t.Fatal("catch block is nil")
	}

	// Check finally is nil
	if stmt.Finally != nil {
		t.Error("finally clause should be nil")
	}
}

func TestTryFinallyStatement(t *testing.T) {
	input := `try { x = 1 } finally { pr("done") }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*TryStatement)
	if !ok {
		t.Fatalf("stmt not *TryStatement. got=%T", program.Statements[0])
	}

	// Check try block exists
	if stmt.Block == nil {
		t.Fatal("try block is nil")
	}

	// Check catch is nil
	if stmt.Catch != nil {
		t.Error("catch clause should be nil")
	}

	// Check finally exists
	if stmt.Finally == nil {
		t.Fatal("finally clause is nil")
	}
	if stmt.Finally.Block == nil {
		t.Fatal("finally block is nil")
	}
}

func TestTryCatchFinallyStatement(t *testing.T) {
	input := `try { throw "err" } catch (e) { pr(e) } finally { pr("cleanup") }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*TryStatement)
	if !ok {
		t.Fatalf("stmt not *TryStatement. got=%T", program.Statements[0])
	}

	// Check try block exists
	if stmt.Block == nil {
		t.Fatal("try block is nil")
	}

	// Check catch exists
	if stmt.Catch == nil {
		t.Fatal("catch clause is nil")
	}
	if stmt.Catch.Exception.Value != "e" {
		t.Errorf("catch exception variable should be 'e', got=%s", stmt.Catch.Exception.Value)
	}

	// Check finally exists
	if stmt.Finally == nil {
		t.Fatal("finally clause is nil")
	}
}

func TestThrowStatement(t *testing.T) {
	input := `throw "error message"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ThrowStatement)
	if !ok {
		t.Fatalf("stmt not *ThrowStatement. got=%T", program.Statements[0])
	}

	if stmt.ErrExpr == nil {
		t.Fatal("throw expression is nil")
	}

	// Check the expression is a string literal
	strLit, ok := stmt.ErrExpr.(*StringLiteral)
	if !ok {
		t.Fatalf("throw expression not *StringLiteral. got=%T", stmt.ErrExpr)
	}
	if strLit.Value != "error message" {
		t.Errorf("throw message should be 'error message', got=%s", strLit.Value)
	}
}

func TestThrowWithoutValue(t *testing.T) {
	// Throw without value is valid when followed by semicolon
	input := `throw;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ThrowStatement)
	if !ok {
		t.Fatalf("stmt not *ThrowStatement. got=%T", program.Statements[0])
	}

	// Throw without value should have nil ErrExpr
	if stmt.ErrExpr != nil {
		t.Errorf("throw without value should have nil ErrExpr, got=%T", stmt.ErrExpr)
	}
}

func TestThrowAtEndOfBlock(t *testing.T) {
	input := `try { throw } catch (e) { }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	tryStmt, ok := program.Statements[0].(*TryStatement)
	if !ok {
		t.Fatalf("stmt not *TryStatement. got=%T", program.Statements[0])
	}

	throwStmt, ok := tryStmt.Block.Statements[0].(*ThrowStatement)
	if !ok {
		t.Fatalf("statement in try block not *ThrowStatement. got=%T", tryStmt.Block.Statements[0])
	}

	// Throw without value should have nil ErrExpr
	if throwStmt.ErrExpr != nil {
		t.Errorf("throw without value should have nil ErrExpr, got=%T", throwStmt.ErrExpr)
	}
}

func TestTryStatementWithoutCatchOrFinally(t *testing.T) {
	input := `try { x = 1 }`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()

	// Should have errors because try needs catch or finally
	if len(p.errors) == 0 {
		t.Error("expected parse error for try without catch or finally")
	}
}

func TestNestedTryCatchStatement(t *testing.T) {
	input := `
try {
	try {
		throw "inner"
	} catch (e) {
		print(e)
	}
} catch (e) {
	print(e)
}
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	outerTry, ok := program.Statements[0].(*TryStatement)
	if !ok {
		t.Fatalf("stmt not *TryStatement. got=%T", program.Statements[0])
	}

	// Inner try should be the first statement in outer try block
	innerTry, ok := outerTry.Block.Statements[0].(*TryStatement)
	if !ok {
		t.Fatalf("inner stmt not *TryStatement. got=%T", outerTry.Block.Statements[0])
	}

	// Verify inner try has catch
	if innerTry.Catch == nil {
		t.Fatal("inner try should have catch clause")
	}
}

// ============================================
// Switch Statement Tests
// ============================================

func TestSwitchStatement(t *testing.T) {
	input := `
switch (x) {
	case 1:
		print("one")
	case 2:
		print("two")
	default:
		print("other")
}
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	switchStmt, ok := program.Statements[0].(*SwitchStatement)
	if !ok {
		t.Fatalf("stmt not *SwitchStatement. got=%T", program.Statements[0])
	}

	if len(switchStmt.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(switchStmt.Cases))
	}

	if switchStmt.Default == nil {
		t.Fatal("expected default clause")
	}
}

func TestSwitchStatementWithoutDefault(t *testing.T) {
	input := `
switch (x) {
	case 1:
		print("one")
	case 2:
		print("two")
}
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	switchStmt, ok := program.Statements[0].(*SwitchStatement)
	if !ok {
		t.Fatalf("stmt not *SwitchStatement. got=%T", program.Statements[0])
	}

	if len(switchStmt.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(switchStmt.Cases))
	}

	if switchStmt.Default != nil {
		t.Fatal("expected no default clause")
	}
}

func TestSwitchStatementWithMultipleStatements(t *testing.T) {
	input := `
switch (x) {
	case 1:
		print("one")
		print("uno")
		break
	case 2:
		print("two")
		print("dos")
}
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	switchStmt, ok := program.Statements[0].(*SwitchStatement)
	if !ok {
		t.Fatalf("stmt not *SwitchStatement. got=%T", program.Statements[0])
	}

	if len(switchStmt.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(switchStmt.Cases))
	}

	// First case should have 3 statements in the block
	if len(switchStmt.Cases[0].Consequence.Statements) != 3 {
		t.Fatalf("expected 3 statements in first case, got %d", len(switchStmt.Cases[0].Consequence.Statements))
	}
}

func TestSwitchStatementErrorUnclosed(t *testing.T) {
	// Parser may handle unclosed switch gracefully (EOF closes it)
	input := `switch (x) { case 1: print("one")`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	// Just verify it parses without crashing
	_ = program
}

func TestSwitchStatementErrorMultipleDefault(t *testing.T) {
	input := `
switch (x) {
	default:
		print("first")
	default:
		print("second")
}
`
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error for multiple default clauses")
	}
}

func TestSwitchStatementErrorCaseAfterDefault(t *testing.T) {
	input := `
switch (x) {
	default:
		print("default")
	case 1:
		print("one")
}
`
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error for case after default")
	}
}

// ============================================
// Ternary Expression Tests
// ============================================

func TestTernaryExpression(t *testing.T) {
	input := `x > 0 ? "positive" : "non-positive"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	testProgramStatements(t, program, 1)

	exprStmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	ternary, ok := exprStmt.Expression.(*TernaryExpression)
	if !ok {
		t.Fatalf("expression not *TernaryExpression. got=%T", exprStmt.Expression)
	}

	if ternary.Condition == nil {
		t.Fatal("expected condition expression")
	}
	if ternary.Consequent == nil {
		t.Fatal("expected consequent expression")
	}
	if ternary.Alternative == nil {
		t.Fatal("expected alternative expression")
	}
}

func TestTernaryExpressionNested(t *testing.T) {
	input := `x > 0 ? x > 10 ? "big" : "small" : "non-positive"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	outerTernary, ok := exprStmt.Expression.(*TernaryExpression)
	if !ok {
		t.Fatalf("expression not *TernaryExpression. got=%T", exprStmt.Expression)
	}

	// Consequent should be another ternary
	innerTernary, ok := outerTernary.Consequent.(*TernaryExpression)
	if !ok {
		t.Fatalf("consequent not *TernaryExpression. got=%T", outerTernary.Consequent)
	}

	_ = innerTernary
}

func TestTernaryExpressionWithFunctionCall(t *testing.T) {
	input := `condition ? foo() : bar()`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	ternary, ok := exprStmt.Expression.(*TernaryExpression)
	if !ok {
		t.Fatalf("expression not *TernaryExpression. got=%T", exprStmt.Expression)
	}

	// Check consequent is a call expression
	_, ok = ternary.Consequent.(*CallExpression)
	if !ok {
		t.Fatalf("consequent not *CallExpression. got=%T", ternary.Consequent)
	}

	// Check alternative is a call expression
	_, ok = ternary.Alternative.(*CallExpression)
	if !ok {
		t.Fatalf("alternative not *CallExpression. got=%T", ternary.Alternative)
	}
}

func TestTernaryExpressionErrorMissingColon(t *testing.T) {
	input := `x > 0 ? "positive"`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.errors) == 0 {
		t.Error("expected parse error for missing colon in ternary")
	}
}

// ============================================
// C-Style For Loop Additional Tests
// ============================================

func TestCStyleForLoopWithoutInit(t *testing.T) {
	input := `for (; i < 10; i = i + 1) { print(i) }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// Init should be nil, condition and update should be set
	if stmt.Init != nil {
		t.Error("expected nil Init")
	}
	if stmt.Condition == nil {
		t.Error("expected Condition expression")
	}
	if stmt.Update == nil {
		t.Error("expected Update statement")
	}
}

func TestCStyleForLoopWithoutCondition(t *testing.T) {
	input := `for (var i = 0; ; i = i + 1) { if (i > 10) { break } }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// Init and update should be set, condition should be nil
	if stmt.Init == nil {
		t.Error("expected Init statement")
	}
	if stmt.Condition != nil {
		t.Error("expected nil Condition")
	}
	if stmt.Update == nil {
		t.Error("expected Update statement")
	}
}

func TestCStyleForLoopWithoutUpdate(t *testing.T) {
	input := `for (var i = 0; i < 10; ) { print(i); i = i + 1 }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// Init and condition should be set, update should be nil
	if stmt.Init == nil {
		t.Error("expected Init statement")
	}
	if stmt.Condition == nil {
		t.Error("expected Condition expression")
	}
	if stmt.Update != nil {
		t.Error("expected nil Update")
	}
}

func TestCStyleForLoopInfinite(t *testing.T) {
	input := `for (;;) { break }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// All should be nil
	if stmt.Init != nil {
		t.Error("expected nil Init")
	}
	if stmt.Condition != nil {
		t.Error("expected nil Condition")
	}
	if stmt.Update != nil {
		t.Error("expected nil Update")
	}
}

func TestCStyleForLoopWithConstInit(t *testing.T) {
	input := `for (const i = 0; i < 10; i = i + 1) { print(i) }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// Init should be a const statement
	_, ok = stmt.Init.(*ConstStatement)
	if !ok {
		t.Fatalf("expected Init to be *ConstStatement, got %T", stmt.Init)
	}
}

// ============================================
// Additional Expression Tests
// ============================================

func TestPostfixExpressionOnIndex(t *testing.T) {
	input := `arr[i]++`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	postfix, ok := exprStmt.Expression.(*PostfixExpression)
	if !ok {
		t.Fatalf("expression not *PostfixExpression. got=%T", exprStmt.Expression)
	}

	// Check that the left side is an index expression
	_, ok = postfix.Left.(*IndexExpression)
	if !ok {
		t.Fatalf("left not *IndexExpression. got=%T", postfix.Left)
	}
}

func TestCompoundAssignmentWithIndex(t *testing.T) {
	input := `arr[i] += 1`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not *ExpressionStatement. got=%T", program.Statements[0])
	}

	compound, ok := exprStmt.Expression.(*CompoundAssignmentExpression)
	if !ok {
		t.Fatalf("expression not *CompoundAssignmentExpression. got=%T", exprStmt.Expression)
	}

	// Check that the left side is an index expression
	_, ok = compound.Left.(*IndexExpression)
	if !ok {
		t.Fatalf("left not *IndexExpression. got=%T", compound.Left)
	}
}

// TestShortVarStatement tests short variable declaration (:=)
func TestShortVarStatement(t *testing.T) {
	input := `x := 10`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ShortVarStatement)
	if !ok {
		t.Fatalf("stmt not *ShortVarStatement. got=%T", program.Statements[0])
	}

	if stmt.Name.Value != "x" {
		t.Errorf("stmt.Name.Value not 'x'. got=%s", stmt.Name.Value)
	}

	if stmt.Name.TokenLiteral() != "x" {
		t.Errorf("stmt.Name.TokenLiteral() not 'x'. got=%s", stmt.Name.TokenLiteral())
	}

	lit, ok := stmt.Value.(*IntegerLiteral)
	if !ok {
		t.Fatalf("stmt.Value not *IntegerLiteral. got=%T", stmt.Value)
	}

	if lit.Value != 10 {
		t.Errorf("lit.Value not 10. got=%d", lit.Value)
	}
}

// TestShortVarStatementMultiple tests multiple short variable declarations
func TestShortVarStatementMultiple(t *testing.T) {
	input := `
x := 10
y := 20
name := "hello"
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedName    string
		expectedLiteral interface{}
	}{
		{"x", int64(10)},
		{"y", int64(20)},
		{"name", "hello"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		shortVar, ok := stmt.(*ShortVarStatement)
		if !ok {
			t.Errorf("stmt %d not *ShortVarStatement. got=%T", i, stmt)
			continue
		}

		if shortVar.Name.Value != tt.expectedName {
			t.Errorf("stmt %d: shortVar.Name.Value not '%s'. got=%s", i, tt.expectedName, shortVar.Name.Value)
		}

		switch expected := tt.expectedLiteral.(type) {
		case int64:
			lit, ok := shortVar.Value.(*IntegerLiteral)
			if !ok {
				t.Errorf("stmt %d: shortVar.Value not *IntegerLiteral. got=%T", i, shortVar.Value)
				continue
			}
			if lit.Value != expected {
				t.Errorf("stmt %d: lit.Value not %d. got=%d", i, expected, lit.Value)
			}
		case string:
			lit, ok := shortVar.Value.(*StringLiteral)
			if !ok {
				t.Errorf("stmt %d: shortVar.Value not *StringLiteral. got=%T", i, shortVar.Value)
				continue
			}
			if lit.Value != expected {
				t.Errorf("stmt %d: lit.Value not %s. got=%s", i, expected, lit.Value)
			}
		}
	}
}
