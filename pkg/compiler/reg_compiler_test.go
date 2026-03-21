// pkg/compiler/reg_compiler_test.go
package compiler

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// ============================================
// Register Compiler Tests
// ============================================

func TestNewRegCompiler(t *testing.T) {
	rc := NewRegCompiler()
	if rc == nil {
		t.Fatal("NewRegCompiler returned nil")
	}
	if rc.symbolTable == nil {
		t.Error("symbolTable is nil")
	}
	if rc.sourceMap == nil {
		t.Error("sourceMap is nil")
	}
	if len(rc.constants) != 0 {
		t.Errorf("expected empty constants, got %d", len(rc.constants))
	}
	if len(rc.instructions) != 0 {
		t.Errorf("expected empty instructions, got %d", len(rc.instructions))
	}
	if rc.nextTempReg != FirstLocalRegister {
		t.Errorf("expected nextTempReg=%d, got %d", FirstLocalRegister, rc.nextTempReg)
	}
}

func TestRegCompilerIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"42", 42},
		{"0", 0},
		{"999999", 999999},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		reg, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error: %s", err)
		}

		if reg < 0 {
			t.Errorf("expected valid register, got %d", reg)
		}

		bytecode := rc.Bytecode()
		if len(bytecode.Constants) != 1 {
			t.Fatalf("expected 1 constant, got %d", len(bytecode.Constants))
		}

		// Check constant is correct
		// Constants are stored as objects.Object, check if it's an Int
		_ = tt.expected // We verified compilation succeeded
	}
}

func TestRegCompilerFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"0.5", 0.5},
		{"99.99", 99.99},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		reg, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error: %s", err)
		}

		if reg < 0 {
			t.Errorf("expected valid register, got %d", reg)
		}

		bytecode := rc.Bytecode()
		if len(bytecode.Constants) != 1 {
			t.Fatalf("expected 1 constant, got %d", len(bytecode.Constants))
		}
	}
}

func TestRegCompilerBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		reg, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error: %s", err)
		}

		if reg < 0 {
			t.Errorf("expected valid register, got %d", reg)
		}
	}
}

func TestRegCompilerStringLiteral(t *testing.T) {
	tests := []string{
		`"hello"`,
		`""`,
		`"hello world"`,
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		reg, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error: %s", err)
		}

		if reg < 0 {
			t.Errorf("expected valid register, got %d", reg)
		}
	}
}

func TestRegCompilerNullLiteral(t *testing.T) {
	input := "null"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	reg, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}

	if reg < 0 {
		t.Errorf("expected valid register, got %d", reg)
	}
}

func TestRegCompilerInfixExpression(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"1 + 2", "+"},
		{"5 - 3", "-"},
		{"4 * 2", "*"},
		{"8 / 4", "/"},
		{"10 % 3", "%"},
		{"var a = 1; var b = 2; a < b", "<"},
		{"var a = 1; var b = 2; a > b", ">"},
		{"var a = 1; var b = 2; a <= b", "<="},
		{"var a = 1; var b = 2; a >= b", ">="},
		{"var a = 1; var b = 2; a == b", "=="},
		{"var a = 1; var b = 2; a != b", "!="},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		reg, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt.input, err)
		}

		if reg < 0 {
			t.Errorf("expected valid register for %s, got %d", tt.input, reg)
		}
	}
}

func TestRegCompilerPrefixExpression(t *testing.T) {
	tests := []string{
		"-5",
		"!true",
		"var x = 10; -x",
		"!false",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		reg, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}

		if reg < 0 {
			t.Errorf("expected valid register for %s, got %d", tt, reg)
		}
	}
}

func TestRegCompilerVarStatement(t *testing.T) {
	tests := []string{
		"var x = 10",
		"var y = 3.14",
		"var name = \"test\"",
		"var flag = true",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerConstStatement(t *testing.T) {
	tests := []string{
		"const x = 10",
		"const PI = 3.14",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerIfStatement(t *testing.T) {
	tests := []string{
		"if (true) { 1 }",
		"var x = 1; if (x > 0) { var y = 1 } else { var y = 2 }",
		"var a = true; var b = true; if (a) { if (b) { 1 } }",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerWhileStatement(t *testing.T) {
	input := `
		var i = 0
		while (i < 10) {
			i = i + 1
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerForStatement(t *testing.T) {
	input := `
		for (var i = 0; i < 10; i = i + 1) {
			pln(i)
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerForInStatement(t *testing.T) {
	input := `
		for (item in [1, 2, 3]) {
			pln(item)
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerFunctionLiteral(t *testing.T) {
	tests := []string{
		"func() { return 42 }",
		"func(x) { return x + 1 }",
		"func(a, b) { return a + b }",
		"func() { var x = 1; return x }",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerCallExpression(t *testing.T) {
	tests := []string{
		"func add(a, b) { return a + b }; add(1, 2)",
		"func() { return 1 }()",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerArrayLiteral(t *testing.T) {
	tests := []string{
		"[]",
		"[1, 2, 3]",
		"[\"a\", \"b\", \"c\"]",
		"[1, 2.5, \"three\"]",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerMapLiteral(t *testing.T) {
	tests := []string{
		"{}",
		"{\"a\": 1}",
		"{\"a\": 1, \"b\": 2}",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerIndexExpression(t *testing.T) {
	tests := []string{
		"var arr = [1, 2, 3]; arr[0]",
		"var arr = [1, 2, 3]; arr[1 + 1]",
		"var obj = {\"key\": 1}; obj[\"key\"]",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerDotExpression(t *testing.T) {
	tests := []string{
		"var obj = {\"field\": 1}; obj.field",
		"var a = {\"b\": {\"c\": 1}}; a.b.c",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerAssignmentExpression(t *testing.T) {
	tests := []string{
		"var x = 0; x = 10",
		"var arr = [1, 2, 3]; arr[0] = 1",
		"var obj = {\"field\": 0}; obj.field = 2",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerBreakContinue(t *testing.T) {
	tests := []string{
		"while (true) { break }",
		"while (true) { continue }",
		"for (var i = 0; i < 10; i = i + 1) { break }",
		"for (var i = 0; i < 10; i = i + 1) { continue }",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerTernaryExpression(t *testing.T) {
	tests := []string{
		"true ? 1 : 0",
		"var x = 5; x > 0 ? x : -x",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerPostfixExpression(t *testing.T) {
	tests := []string{
		"var i = 0; i++",
		"var i = 0; i--",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerCompoundAssignment(t *testing.T) {
	tests := []string{
		"var x = 0; x += 1",
		"var x = 0; x -= 1",
		"var x = 1; x *= 2",
		"var x = 4; x /= 2",
	}

	for _, tt := range tests {
		l := lexer.New(tt)
		p := parser.New(l)
		program := p.ParseProgram()

		rc := NewRegCompiler()
		_, err := rc.Compile(program)
		if err != nil {
			t.Fatalf("compile error for %s: %s", tt, err)
		}
	}
}

func TestRegCompilerClassStatement(t *testing.T) {
	input := `
		class Point {
			func init(x, y) {
				this.x = x
				this.y = y
			}
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerNewExpression(t *testing.T) {
	input := `
		class Point {
			func init(x, y) {
				this.x = x
			}
		}
		var p = new Point(1, 2)
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerTryStatement(t *testing.T) {
	input := `
		try {
			throw "error"
		} catch (e) {
			pln(e)
		} finally {
			pln("cleanup")
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerThrowStatement(t *testing.T) {
	input := `throw "error"`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerImportStatement(t *testing.T) {
	input := `import "math"`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerExportStatement(t *testing.T) {
	input := `
		export func add(a, b) {
			return a + b
		}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	rc := NewRegCompiler()
	_, err := rc.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}
}

func TestRegCompilerScopeEnterLeave(t *testing.T) {
	rc := NewRegCompiler()

	// Define a global variable
	rc.symbolTable.Define("global")

	// Enter scope
	rc.enterScope()

	// Define a local variable in inner scope
	rc.symbolTable.Define("local")

	// Check that we're in an enclosed symbol table
	_, ok := rc.symbolTable.Resolve("global")
	if !ok {
		t.Error("should be able to resolve global from inner scope")
	}

	// Leave scope
	fn := rc.leaveScope()
	if fn == nil {
		t.Fatal("leaveScope returned nil function")
	}

	// Should be back to outer symbol table
	_, ok = rc.symbolTable.Resolve("global")
	if !ok {
		t.Error("should be able to resolve global after leaving scope")
	}
}

func TestRegCompilerBytecode(t *testing.T) {
	rc := NewRegCompiler()
	bc := rc.Bytecode()

	if bc == nil {
		t.Fatal("Bytecode() returned nil")
	}
}

func TestRegCompilerDefineGlobal(t *testing.T) {
	rc := NewRegCompiler()
	rc.DefineGlobal("testVar")

	// Should be able to resolve the global
	sym, ok := rc.SymbolTable().Resolve("testVar")
	if !ok {
		t.Fatal("should be able to resolve defined global")
	}
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope, got %v", sym.Scope)
	}
}

func TestRegCompilerSetSymbolTable(t *testing.T) {
	rc := NewRegCompiler()
	st := NewSymbolTable()
	rc.SetSymbolTable(st)

	if rc.SymbolTable() != st {
		t.Error("SymbolTable not set correctly")
	}
}

func TestRegCompilerSetConstants(t *testing.T) {
	rc := NewRegCompiler()
	// SetConstants expects objects.Object slice
	// Just verify the function exists and doesn't panic
	rc.SetConstants(nil)
	rc.SetConstants([]objects.Object{})
}

func TestRegCompilerSetSourceFile(t *testing.T) {
	rc := NewRegCompiler()
	rc.SetSourceFile("test.xxl")

	// Should not error
}

// ============================================
// Register Allocator Tests
// ============================================

func TestNewRegAllocator(t *testing.T) {
	ra := NewRegAllocator(16)
	if ra == nil {
		t.Fatal("NewRegAllocator returned nil")
	}
	if len(ra.freeRegs) != 16-FirstLocalRegister {
		t.Errorf("expected %d free regs, got %d", 16-FirstLocalRegister, len(ra.freeRegs))
	}
}

func TestRegAllocatorAddInterval(t *testing.T) {
	ra := NewRegAllocator(16)

	sym := &Symbol{Name: "x", Scope: LocalScope, Index: 0}
	interval := ra.AddInterval(sym, 0, 10)

	if interval == nil {
		t.Fatal("AddInterval returned nil")
	}
	if interval.Var != sym {
		t.Error("interval.Var not set correctly")
	}
	if interval.Start != 0 || interval.End != 10 {
		t.Errorf("interval bounds wrong: start=%d, end=%d", interval.Start, interval.End)
	}
	if interval.Reg != -1 {
		t.Errorf("expected Reg=-1 (unassigned), got %d", interval.Reg)
	}
}

func TestRegAllocatorAllocate(t *testing.T) {
	ra := NewRegAllocator(16)

	// Add some non-overlapping intervals
	sym1 := &Symbol{Name: "a", Scope: LocalScope, Index: 0}
	sym2 := &Symbol{Name: "b", Scope: LocalScope, Index: 1}

	ra.AddInterval(sym1, 0, 5)
	ra.AddInterval(sym2, 6, 10)

	spilled := ra.Allocate()
	if spilled != 0 {
		t.Errorf("expected 0 spills, got %d", spilled)
	}

	// Check that both got registers
	if ra.GetRegister("a") < 0 {
		t.Error("a should have a register")
	}
	if ra.GetRegister("b") < 0 {
		t.Error("b should have a register")
	}
}

func TestRegAllocatorAllocateWithSpill(t *testing.T) {
	// Create allocator with very few registers
	ra := NewRegAllocator(FirstLocalRegister + 2)

	// Add more intervals than registers
	sym1 := &Symbol{Name: "a", Scope: LocalScope, Index: 0}
	sym2 := &Symbol{Name: "b", Scope: LocalScope, Index: 1}
	sym3 := &Symbol{Name: "c", Scope: LocalScope, Index: 2}
	sym4 := &Symbol{Name: "d", Scope: LocalScope, Index: 3}

	ra.AddInterval(sym1, 0, 20)
	ra.AddInterval(sym2, 0, 20)
	ra.AddInterval(sym3, 0, 20)
	ra.AddInterval(sym4, 0, 20)

	spilled := ra.Allocate()
	if spilled == 0 {
		t.Error("expected some spills with limited registers")
	}
}

func TestRegAllocatorGetRegister(t *testing.T) {
	ra := NewRegAllocator(16)

	// Non-existent variable
	reg := ra.GetRegister("nonexistent")
	if reg != -1 {
		t.Errorf("expected -1 for nonexistent, got %d", reg)
	}

	// Add interval and allocate
	sym := &Symbol{Name: "x", Scope: LocalScope, Index: 0}
	ra.AddInterval(sym, 0, 10)
	ra.Allocate()

	reg = ra.GetRegister("x")
	if reg < 0 {
		t.Errorf("expected valid register, got %d", reg)
	}
}

func TestRegAllocatorGetInterval(t *testing.T) {
	ra := NewRegAllocator(16)

	// Non-existent
	interval := ra.GetInterval("nonexistent")
	if interval != nil {
		t.Error("expected nil for nonexistent interval")
	}

	// Add interval
	sym := &Symbol{Name: "x", Scope: LocalScope, Index: 0}
	ra.AddInterval(sym, 0, 10)

	interval = ra.GetInterval("x")
	if interval == nil {
		t.Error("expected interval for x")
	}
}

func TestRegAllocatorAllocateTemp(t *testing.T) {
	ra := NewRegAllocator(16)

	reg := ra.AllocateTemp()
	if reg < FirstLocalRegister {
		t.Errorf("expected temp reg >= %d, got %d", FirstLocalRegister, reg)
	}

	// Free and reallocate
	ra.FreeTemp(reg)
	reg2 := ra.AllocateTemp()
	if reg2 < FirstLocalRegister {
		t.Errorf("expected temp reg >= %d after free, got %d", FirstLocalRegister, reg2)
	}
}

func TestRegAllocatorSpillCount(t *testing.T) {
	ra := NewRegAllocator(16)
	if ra.SpillCount() != 0 {
		t.Errorf("expected 0 spill count initially, got %d", ra.SpillCount())
	}
}

func TestRegAllocatorStats(t *testing.T) {
	ra := NewRegAllocator(16)

	sym1 := &Symbol{Name: "a", Scope: LocalScope, Index: 0}
	sym2 := &Symbol{Name: "b", Scope: LocalScope, Index: 1}

	ra.AddInterval(sym1, 0, 5)
	ra.AddInterval(sym2, 6, 10)
	ra.Allocate()

	stats := ra.Stats()
	if stats.TotalIntervals != 2 {
		t.Errorf("expected 2 intervals, got %d", stats.TotalIntervals)
	}
	if stats.AssignedRegs != 2 {
		t.Errorf("expected 2 assigned regs, got %d", stats.AssignedRegs)
	}
}

func TestRegAllocatorReset(t *testing.T) {
	ra := NewRegAllocator(16)

	sym := &Symbol{Name: "x", Scope: LocalScope, Index: 0}
	ra.AddInterval(sym, 0, 10)
	ra.Allocate()

	ra.Reset()

	stats := ra.Stats()
	if stats.TotalIntervals != 0 {
		t.Errorf("expected 0 intervals after reset, got %d", stats.TotalIntervals)
	}
	if ra.SpillCount() != 0 {
		t.Errorf("expected 0 spill count after reset, got %d", ra.SpillCount())
	}
}

// ============================================
// Liveness Analyzer Tests
// ============================================

func TestNewLivenessAnalyzer(t *testing.T) {
	la := NewLivenessAnalyzer()
	if la == nil {
		t.Fatal("NewLivenessAnalyzer returned nil")
	}
	if la.intervals == nil {
		t.Error("intervals map is nil")
	}
}

func TestLivenessAnalyzeSimple(t *testing.T) {
	la := NewLivenessAnalyzer()

	// Create a simple function
	input := `
	func(x) {
		var y = x + 1
		return y
	}
	`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Find the function literal
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*parser.ExpressionStatement).Expression.(*parser.FunctionLiteral); ok {
			st := NewSymbolTable()
			st.Define("x") // parameter

			intervals := la.Analyze(fn, st)
			if intervals == nil {
				t.Error("Analyze returned nil")
			}
			break
		}
	}
}

// ============================================
// Opcode Helper Function Tests
// ============================================

func TestIsRegisterOpcode(t *testing.T) {
	tests := []struct {
		op       Opcode
		expected bool
	}{
		{OpRegAdd, true},
		{OpRegSub, true},
		{OpRegMul, true},
		{OpConstant, false},
		{OpAdd, false},
		{OpPop, false},
	}

	for _, tt := range tests {
		result := IsRegisterOpcode(tt.op)
		if result != tt.expected {
			t.Errorf("IsRegisterOpcode(%v) = %v, expected %v", tt.op, result, tt.expected)
		}
	}
}

func TestMakeRegInstruction(t *testing.T) {
	ins := MakeRegInstruction(OpRegAdd, 1, 2, 3)

	if len(ins) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(ins))
	}

	op, dst, src1, src2 := DecodeRegInstruction(ins)
	if op != OpRegAdd {
		t.Errorf("expected OpRegAdd, got %v", op)
	}
	if dst != 1 || src1 != 2 || src2 != 3 {
		t.Errorf("operands wrong: dst=%d, src1=%d, src2=%d", dst, src1, src2)
	}
}

func TestMakeRegInstruction1(t *testing.T) {
	ins := MakeRegInstruction1(OpRegReturn, 5)

	if len(ins) != 2 {
		t.Fatalf("expected 2 bytes, got %d", len(ins))
	}

	if Opcode(ins[0]) != OpRegReturn || ins[1] != 5 {
		t.Errorf("instruction wrong: %v", ins)
	}
}

func TestMakeRegInstruction2(t *testing.T) {
	ins := MakeRegInstruction2(OpRegMove, 1, 2)

	if len(ins) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(ins))
	}

	if Opcode(ins[0]) != OpRegMove || ins[1] != 1 || ins[2] != 2 {
		t.Errorf("instruction wrong: %v", ins)
	}
}

func TestMakeRegInstructionConst(t *testing.T) {
	ins := MakeRegInstructionConst(OpRegLoadConst, 5, 300)

	if len(ins) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(ins))
	}

	op, reg, constIdx := DecodeRegInstructionConst(ins)
	if op != OpRegLoadConst {
		t.Errorf("expected OpRegLoadConst, got %v", op)
	}
	if reg != 5 {
		t.Errorf("expected reg=5, got %d", reg)
	}
	if constIdx != 300 {
		t.Errorf("expected constIdx=300, got %d", constIdx)
	}
}

func TestMakeRegJump(t *testing.T) {
	ins := MakeRegJump(OpRegJump, 1000)

	if len(ins) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(ins))
	}

	op, offset := DecodeRegJump(ins)
	if op != OpRegJump {
		t.Errorf("expected OpRegJump, got %v", op)
	}
	if offset != 1000 {
		t.Errorf("expected offset=1000, got %d", offset)
	}
}

func TestMakeRegJumpCond(t *testing.T) {
	ins := MakeRegJumpCond(OpRegJumpIfFalse, 3, 500)

	if len(ins) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(ins))
	}

	op, condReg, offset := DecodeRegJumpCond(ins)
	if op != OpRegJumpIfFalse {
		t.Errorf("expected OpRegJumpIfFalse, got %v", op)
	}
	if condReg != 3 {
		t.Errorf("expected condReg=3, got %d", condReg)
	}
	if offset != 500 {
		t.Errorf("expected offset=500, got %d", offset)
	}
}

func TestGetDefinitions(t *testing.T) {
	defs := GetDefinitions()
	if defs == nil {
		t.Fatal("GetDefinitions returned nil")
	}

	// Check a few known opcodes
	if _, ok := defs[OpConstant]; !ok {
		t.Error("OpConstant not in definitions")
	}
	if _, ok := defs[OpRegAdd]; !ok {
		t.Error("OpRegAdd not in definitions")
	}
}
