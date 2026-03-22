// pkg/compiler/reg_compiler.go
// Register-based bytecode compiler
package compiler

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// RegCompiler compiles AST to register-based bytecode
type RegCompiler struct {
	constants   []objects.Object
	symbolTable *SymbolTable

	// Code generation
	instructions []byte

	// Scopes for nested function compilation
	// Each entry stores the outer scope's state when entering a new scope
	scopeStack []regScopeState

	// Register allocation
	nextTempReg  int
	maxReg       int
	freeRegs     []int // List of freed registers available for reuse

	// Source mapping
	sourceMap  *SourceMap
	sourceFile string

	// Loop context for break/continue
	loopContexts []regLoopContext

	// Counter for unique loop variable names
	forInLoopCounter int

	// Track depth of try blocks (to disable tail call inside try)
	tryBlockDepth int
}

type regScopeState struct {
	instructions []byte
	nextTempReg  int
	maxReg       int
	freeRegs     []int
}

type regLoopContext struct {
	startPos    int
	breakPos    []int
	continuePos []int
}

// NewRegCompiler creates a new register-based compiler
func NewRegCompiler() *RegCompiler {
	return &RegCompiler{
		constants:    []objects.Object{},
		symbolTable:  NewSymbolTable(),
		instructions: []byte{},
		nextTempReg:  FirstLocalRegister,
		maxReg:       FirstLocalRegister,
		freeRegs:     []int{},
		sourceMap:    NewSourceMap(),
	}
}

// enterScope enters a new compilation scope (for function compilation)
func (c *RegCompiler) enterScope() {
	// Save current state to stack
	c.scopeStack = append(c.scopeStack, regScopeState{
		instructions: c.instructions,
		nextTempReg:  c.nextTempReg,
		maxReg:       c.maxReg,
		freeRegs:     c.freeRegs,
	})

	// Start with fresh instructions for the new scope
	c.instructions = []byte{}

	// Reset register allocation for new scope (parameters will use R0-R7)
	c.nextTempReg = FirstLocalRegister
	c.maxReg = FirstLocalRegister
	c.freeRegs = []int{}

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// leaveScope leaves the current scope and returns the compiled function
func (c *RegCompiler) leaveScope() *CompiledFunction {
	// Capture the function's instructions
	fnInstructions := c.instructions
	numLocals := c.symbolTable.NumDefinitions
	freeVars := make([]Symbol, len(c.symbolTable.FreeSymbols))
	copy(freeVars, c.symbolTable.FreeSymbols)

	// Capture max register used before restoring
	maxRegUsed := c.maxReg

	// Restore outer scope's state
	if len(c.scopeStack) > 0 {
		outer := c.scopeStack[len(c.scopeStack)-1]
		c.scopeStack = c.scopeStack[:len(c.scopeStack)-1]
		c.instructions = outer.instructions
		c.nextTempReg = outer.nextTempReg
		c.maxReg = outer.maxReg
		c.freeRegs = outer.freeRegs
	} else {
		c.instructions = []byte{}
		c.nextTempReg = FirstLocalRegister
		c.maxReg = FirstLocalRegister
		c.freeRegs = []int{}
	}

	c.symbolTable = c.symbolTable.Outer

	return &CompiledFunction{
		Instructions:  fnInstructions,
		NumLocals:     numLocals,
		NumRegs:       maxRegUsed + 1, // +1 because maxReg is 0-indexed
		FreeVariables: freeVars,
	}
}

// Compile compiles an AST node and returns the register containing the result
func (c *RegCompiler) Compile(node parser.Node) (int, error) {
	switch n := node.(type) {
	case *parser.Program:
		return c.compileProgram(n)
	case *parser.ExpressionStatement:
		return c.Compile(n.Expression)
	case *parser.InfixExpression:
		return c.compileInfixExpression(n)
	case *parser.PrefixExpression:
		return c.compilePrefixExpression(n)
	case *parser.IntegerLiteral:
		return c.compileIntegerLiteral(n)
	case *parser.FloatLiteral:
		return c.compileFloatLiteral(n)
	case *parser.BigIntLiteral:
		return c.compileBigIntLiteral(n)
	case *parser.BigFloatLiteral:
		return c.compileBigFloatLiteral(n)
	case *parser.BooleanLiteral:
		return c.compileBooleanLiteral(n)
	case *parser.NullLiteral:
		return c.compileNullLiteral()
	case *parser.StringLiteral:
		return c.compileStringLiteral(n)
	case *parser.Identifier:
		return c.compileIdentifier(n)
	case *parser.VarStatement:
		return c.compileVarStatement(n)
	case *parser.ShortVarStatement:
		return c.compileShortVarStatement(n)
	case *parser.ConstStatement:
		return c.compileConstStatement(n)
	case *parser.BlockStatement:
		return c.compileBlockStatement(n)
	case *parser.IfStatement:
		return c.compileIfStatement(n)
	case *parser.WhileStatement:
		return c.compileWhileStatement(n)
	case *parser.ForStatement:
		return c.compileForStatement(n)
	case *parser.ForInStatement:
		return c.compileForInStatement(n)
	case *parser.ReturnStatement:
		return c.compileReturnStatement(n)
	case *parser.FunctionLiteral:
		return c.compileFunctionLiteral(n)
	case *parser.CallExpression:
		return c.compileCallExpression(n)
	case *parser.ArrayLiteral:
		return c.compileArrayLiteral(n)
	case *parser.MapLiteral:
		return c.compileMapLiteral(n)
	case *parser.IndexExpression:
		return c.compileIndexExpression(n)
	case *parser.SliceExpression:
		return c.compileSliceExpression(n)
	case *parser.DotExpression:
		return c.compileDotExpression(n)
	case *parser.AssignmentExpression:
		return c.compileAssignmentExpression(n)
	case *parser.BreakStatement:
		return c.compileBreakStatement(n)
	case *parser.ContinueStatement:
		return c.compileContinueStatement(n)
	case *parser.TernaryExpression:
		return c.compileTernaryExpression(n)
	case *parser.PostfixExpression:
		return c.compilePostfixExpression(n)
	case *parser.CompoundAssignmentExpression:
		return c.compileCompoundAssignmentExpression(n)
	case *parser.ImportStatement:
		return c.compileImportStatement(n)
	case *parser.ExportStatement:
		return c.compileExportStatement(n)
	case *parser.ClassStatement:
		return c.compileClassStatement(n)
	case *parser.NewExpression:
		return c.compileNewExpression(n)
	case *parser.ThisExpression:
		return c.compileThisExpression(n)
	case *parser.SuperCallExpression:
		return c.compileSuperCallExpression(n)
	case *parser.TryStatement:
		return c.compileTryStatement(n)
	case *parser.ThrowStatement:
		return c.compileThrowStatement(n)
	case *parser.SwitchStatement:
		return c.compileSwitchStatement(n)
	default:
		return 0, fmt.Errorf("unknown node type: %T", node)
	}
}

// compileProgram compiles a program
func (c *RegCompiler) compileProgram(p *parser.Program) (int, error) {
	var lastReg int
	for _, stmt := range p.Statements {
		var err error
		lastReg, err = c.Compile(stmt)
		if err != nil {
			return 0, err
		}
	}
	// Move the last result to the ReturnRegister so it can be retrieved by the VM
	if lastReg != ReturnRegister {
		c.emitRegMove(ReturnRegister, lastReg)
	}
	return ReturnRegister, nil
}

// compileIntegerLiteral compiles an integer literal
func (c *RegCompiler) compileIntegerLiteral(n *parser.IntegerLiteral) (int, error) {
	// Add to constants
	constIdx := c.addConstant(objects.NewInt(n.Value))

	// Allocate a register and load constant
	dst := c.allocTempReg()
	c.emitRegLoadConst(dst, constIdx)
	return dst, nil
}

// compileFloatLiteral compiles a float literal
func (c *RegCompiler) compileFloatLiteral(n *parser.FloatLiteral) (int, error) {
	constIdx := c.addConstant(&objects.Float{Value: n.Value})
	dst := c.allocTempReg()
	c.emitRegLoadConst(dst, constIdx)
	return dst, nil
}

// compileBigIntLiteral compiles a big integer literal
func (c *RegCompiler) compileBigIntLiteral(n *parser.BigIntLiteral) (int, error) {
	bigInt, err := objects.NewBigIntFromString(n.Value)
	if err != nil {
		return -1, fmt.Errorf("could not parse %q as big int: %v", n.Value, err)
	}
	constIdx := c.addConstant(bigInt)
	dst := c.allocTempReg()
	c.emitRegLoadConst(dst, constIdx)
	return dst, nil
}

// compileBigFloatLiteral compiles a big float literal
func (c *RegCompiler) compileBigFloatLiteral(n *parser.BigFloatLiteral) (int, error) {
	bigFloat, err := objects.NewBigFloatFromString(n.Value)
	if err != nil {
		return -1, fmt.Errorf("could not parse %q as big float: %v", n.Value, err)
	}
	constIdx := c.addConstant(bigFloat)
	dst := c.allocTempReg()
	c.emitRegLoadConst(dst, constIdx)
	return dst, nil
}

// compileBooleanLiteral compiles a boolean literal
func (c *RegCompiler) compileBooleanLiteral(n *parser.BooleanLiteral) (int, error) {
	dst := c.allocTempReg()
	if n.Value {
		c.emitRegTrue(dst)
	} else {
		c.emitRegFalse(dst)
	}
	return dst, nil
}

// compileNullLiteral compiles a null literal
func (c *RegCompiler) compileNullLiteral() (int, error) {
	dst := c.allocTempReg()
	c.emitRegNull(dst)
	return dst, nil
}

// compileStringLiteral compiles a string literal
func (c *RegCompiler) compileStringLiteral(n *parser.StringLiteral) (int, error) {
	constIdx := c.addConstant(objects.NewString(n.Value))
	dst := c.allocTempReg()
	c.emitRegLoadConst(dst, constIdx)
	return dst, nil
}

// compileIdentifier compiles an identifier (variable access)
func (c *RegCompiler) compileIdentifier(n *parser.Identifier) (int, error) {
	symbol, ok := c.symbolTable.Resolve(n.Value)
	if !ok {
		return 0, fmt.Errorf("undefined variable: %s", n.Value)
	}

	dst := c.allocTempReg()
	switch symbol.Scope {
	case GlobalScope:
		c.emitRegLoadGlobal(dst, symbol.Index)
	case LocalScope:
		// Load from local slot into register
		c.emitRegLoadLocal(dst, symbol.Index)
	case BuiltinScope:
		// Load builtin function object into register
		c.emitRegLoadBuiltin(dst, symbol.Index)
	case FreeScope:
		c.emitRegLoadFree(dst, symbol.Index)
	}
	return dst, nil
}

// compileInfixExpression compiles an infix expression
func (c *RegCompiler) compileInfixExpression(n *parser.InfixExpression) (int, error) {
	// Compile left operand
	leftReg, err := c.Compile(n.Left)
	if err != nil {
		return 0, err
	}

	// If left operand is in ReturnRegister, move it to a temp register
	// because compiling the right operand might overwrite ReturnRegister
	if leftReg == ReturnRegister {
		newLeftReg := c.allocTempReg()
		c.emitRegMove(newLeftReg, leftReg)
		leftReg = newLeftReg
	}

	// Compile right operand
	rightReg, err := c.Compile(n.Right)
	if err != nil {
		return 0, err
	}

	// Allocate result register
	dst := c.allocTempReg()

	// Emit appropriate instruction
	switch n.Operator {
	case "+":
		c.emitRegAdd(dst, leftReg, rightReg)
	case "-":
		c.emitRegSub(dst, leftReg, rightReg)
	case "*":
		c.emitRegMul(dst, leftReg, rightReg)
	case "/":
		c.emitRegDiv(dst, leftReg, rightReg)
	case "%":
		c.emitRegMod(dst, leftReg, rightReg)
	case "<":
		c.emitRegLess(dst, leftReg, rightReg)
	case ">":
		c.emitRegGreater(dst, leftReg, rightReg)
	case "<=":
		c.emitRegLessEqual(dst, leftReg, rightReg)
	case ">=":
		c.emitRegGreaterEqual(dst, leftReg, rightReg)
	case "==":
		c.emitRegEqual(dst, leftReg, rightReg)
	case "!=":
		c.emitRegNotEqual(dst, leftReg, rightReg)
	case "&&":
		c.emitRegAnd(dst, leftReg, rightReg)
	case "||":
		c.emitRegOr(dst, leftReg, rightReg)
	default:
		return 0, fmt.Errorf("unknown operator: %s", n.Operator)
	}

	// Free temporary registers
	c.freeTempReg(leftReg)
	c.freeTempReg(rightReg)

	return dst, nil
}

// compilePrefixExpression compiles a prefix expression
func (c *RegCompiler) compilePrefixExpression(n *parser.PrefixExpression) (int, error) {
	switch n.Operator {
	case "++", "--":
		// Prefix increment/decrement: ++x or --x
		// Returns the new value (unlike postfix which returns old value)
		switch right := n.Right.(type) {
		case *parser.Identifier:
			symbol, ok := c.symbolTable.Resolve(right.Value)
			if !ok {
				return 0, fmt.Errorf("undefined variable: %s", right.Value)
			}

			// Allocate a register for the current value
			valReg := c.allocTempReg()

			// Load current value
			switch symbol.Scope {
			case GlobalScope:
				c.emitRegLoadGlobal(valReg, symbol.Index)
			case LocalScope:
				c.emitRegLoadLocal(valReg, symbol.Index)
			case FreeScope:
				c.emitRegLoadFree(valReg, symbol.Index)
			}

			// Load constant 1
			oneIdx := c.addConstant(objects.NewInt(1))
			oneReg := c.allocTempReg()
			c.emitRegLoadConst(oneReg, oneIdx)

			// Perform increment or decrement
			switch n.Operator {
			case "++":
				c.emitRegAdd(valReg, valReg, oneReg)
			case "--":
				c.emitRegSub(valReg, valReg, oneReg)
			}

			// Store back to variable
			switch symbol.Scope {
			case GlobalScope:
				c.emitRegStoreGlobal(valReg, symbol.Index)
			case LocalScope:
				c.emitRegStoreLocal(valReg, symbol.Index)
			case FreeScope:
				c.emitRegStoreFree(valReg, symbol.Index)
			}

			// Free the one register
			c.freeTempReg(oneReg)

			// Return the register holding the new value (prefix returns new value)
			return valReg, nil

		default:
			return 0, fmt.Errorf("prefix %s operator not supported for type: %T", n.Operator, right)
		}

	default:
		// Handle other prefix operators: - and !
		rightReg, err := c.Compile(n.Right)
		if err != nil {
			return 0, err
		}

		dst := c.allocTempReg()

		switch n.Operator {
		case "-":
			c.emitRegNeg(dst, rightReg)
		case "!":
			c.emitRegNot(dst, rightReg)
		default:
			return 0, fmt.Errorf("unknown operator: %s", n.Operator)
		}

		c.freeTempReg(rightReg)
		return dst, nil
	}
}

// compileVarStatement compiles a variable declaration
func (c *RegCompiler) compileVarStatement(n *parser.VarStatement) (int, error) {
	// Pre-define the variable if the value is a function literal
	// This allows recursive functions like: var f = func() { f() }
	if fn, ok := n.Value.(*parser.FunctionLiteral); ok {
		// Define the variable first so the function body can reference it
		symbol := c.symbolTable.Define(n.Name.Value)
		// Set the function name so it can reference itself
		fn.Name = n.Name.Value
		// Compile the function literal
		valReg, err := c.Compile(fn)
		if err != nil {
			return 0, err
		}
		// Store to local or global
		if symbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(valReg, symbol.Index)
		} else {
			c.emitRegStoreLocal(valReg, symbol.Index)
		}
		c.freeTempReg(valReg)
		return valReg, nil
	}

	// Regular variable declaration
	symbol := c.symbolTable.Define(n.Name.Value)

	// Compile the initial value
	valReg, err := c.Compile(n.Value)
	if err != nil {
		return 0, err
	}

	// Store to local or global
	if symbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(valReg, symbol.Index)
	} else {
		c.emitRegStoreLocal(valReg, symbol.Index)
	}

	c.freeTempReg(valReg)
	return valReg, nil
}

// compileShortVarStatement compiles a short variable declaration (:=)
func (c *RegCompiler) compileShortVarStatement(n *parser.ShortVarStatement) (int, error) {
	// Short variable declaration - same semantics as var but simpler syntax
	// Pre-define the variable if the value is a function literal
	if fn, ok := n.Value.(*parser.FunctionLiteral); ok {
		symbol := c.symbolTable.Define(n.Name.Value)
		fn.Name = n.Name.Value
		valReg, err := c.Compile(fn)
		if err != nil {
			return 0, err
		}
		if symbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(valReg, symbol.Index)
		} else {
			c.emitRegStoreLocal(valReg, symbol.Index)
		}
		c.freeTempReg(valReg)
		return valReg, nil
	}

	// Regular short variable declaration
	symbol := c.symbolTable.Define(n.Name.Value)

	valReg, err := c.Compile(n.Value)
	if err != nil {
		return 0, err
	}

	if symbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(valReg, symbol.Index)
	} else {
		c.emitRegStoreLocal(valReg, symbol.Index)
	}

	c.freeTempReg(valReg)
	return valReg, nil
}

// compileConstStatement compiles a constant declaration
// Constants are compiled the same as variables at the bytecode level
// The semantic difference (immutability) is enforced at parse time
func (c *RegCompiler) compileConstStatement(n *parser.ConstStatement) (int, error) {
	// Define the constant in the symbol table
	symbol := c.symbolTable.Define(n.Name.Value)

	// Compile the initial value
	valReg, err := c.Compile(n.Value)
	if err != nil {
		return 0, err
	}

	// Store to local or global
	if symbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(valReg, symbol.Index)
	} else {
		c.emitRegStoreLocal(valReg, symbol.Index)
	}

	c.freeTempReg(valReg)
	return valReg, nil
}

// compileBlockStatement compiles a block statement
func (c *RegCompiler) compileBlockStatement(n *parser.BlockStatement) (int, error) {
	var lastReg int
	for _, stmt := range n.Statements {
		var err error
		lastReg, err = c.Compile(stmt)
		if err != nil {
			return 0, err
		}
	}
	return lastReg, nil
}

// compileIfStatement compiles an if statement
func (c *RegCompiler) compileIfStatement(n *parser.IfStatement) (int, error) {
	// Compile condition
	condReg, err := c.Compile(n.Condition)
	if err != nil {
		return 0, err
	}

	// Jump to else/end if condition is false
	jumpIfFalsePos := c.emitRegJumpIfFalse(condReg, 0)
	c.freeTempReg(condReg)

	// Compile consequence
	consequentReg, err := c.Compile(n.Consequence)
	if err != nil {
		return 0, err
	}

	// Allocate result register and move consequent there
	resultReg := c.allocTempReg()
	if consequentReg != resultReg && consequentReg != 0 {
		c.emitRegMove(resultReg, consequentReg)
	}

	// Jump over alternative
	jumpPos := c.emitRegJump(0)

	// Patch jump to here (start of alternative/end)
	c.patchJump(jumpIfFalsePos)

	// Compile alternative if present
	if n.Alternative != nil {
		alternativeReg, err := c.Compile(n.Alternative)
		if err != nil {
			return 0, err
		}
		// Move alternative to result register
		if alternativeReg != resultReg && alternativeReg != 0 {
			c.emitRegMove(resultReg, alternativeReg)
		}
	} else {
		// No alternative - set result to null
		c.emitRegNull(resultReg)
	}

	// Patch jump to here (end)
	c.patchJump(jumpPos)

	return resultReg, nil
}

// compileWhileStatement compiles a while statement
func (c *RegCompiler) compileWhileStatement(n *parser.WhileStatement) (int, error) {
	// Try to optimize prime check while loop pattern
	if optimized := c.tryOptimizePrimeCheckWhile(n); optimized {
		return 0, nil
	}

	// Record loop start
	startPos := len(c.instructions)

	// Compile condition
	condReg, err := c.Compile(n.Condition)
	if err != nil {
		return 0, err
	}

	// Jump to end if condition is false
	jumpIfFalsePos := c.emitRegJumpIfFalse(condReg, 0)
	c.freeTempReg(condReg)

	// Enter loop context
	c.loopContexts = append(c.loopContexts, regLoopContext{
		startPos: startPos,
	})

	// Compile body
	_, err = c.Compile(n.Body)
	if err != nil {
		return 0, err
	}

	// Jump back to start
	c.emitRegJump(startPos - len(c.instructions))

	// Patch jump to end
	c.patchJump(jumpIfFalsePos)

	// Patch continues - for while loops, continue jumps to the start (condition check)
	ctx := c.loopContexts[len(c.loopContexts)-1]
	for _, pos := range ctx.continuePos {
		c.patchJumpTo(pos, startPos)
	}

	// Patch breaks
	for _, pos := range ctx.breakPos {
		c.patchJump(pos)
	}
	c.loopContexts = c.loopContexts[:len(c.loopContexts)-1]

	return 0, nil
}

// compileForStatement compiles a for statement
func (c *RegCompiler) compileForStatement(n *parser.ForStatement) (int, error) {
	// Try to detect simple counting loop pattern for optimization
	if optimized := c.tryOptimizeSimpleCountingLoop(n); optimized {
		return 0, nil
	}

	// Try to detect prime check inner loop pattern
	if optimized := c.tryOptimizePrimeCheckLoop(n); optimized {
		return 0, nil
	}

	// Try to detect nested loop pattern for optimization
	if optimized := c.tryOptimizeNestedLoop(n); optimized {
		return 0, nil
	}

	// Try loop unrolling for small fixed-iteration loops
	if optimized := c.tryUnrollLoop(n); optimized {
		return 0, nil
	}

	// Compile initializer
	if n.Init != nil {
		_, err := c.Compile(n.Init)
		if err != nil {
			return 0, err
		}
	}

	// Record loop start (before condition)
	startPos := len(c.instructions)

	// Compile condition
	var condReg int
	var err error
	if n.Condition != nil {
		condReg, err = c.Compile(n.Condition)
		if err != nil {
			return 0, err
		}
	}

	// Jump to end if condition is false
	var jumpIfFalsePos int
	if n.Condition != nil {
		jumpIfFalsePos = c.emitRegJumpIfFalse(condReg, 0)
		c.freeTempReg(condReg)
	}

	// Enter loop context
	c.loopContexts = append(c.loopContexts, regLoopContext{
		startPos: startPos,
	})

	// Compile body
	_, err = c.Compile(n.Body)
	if err != nil {
		return 0, err
	}

	// Update position for continue
	ctx := &c.loopContexts[len(c.loopContexts)-1]
	for _, pos := range ctx.continuePos {
		c.patchJumpTo(pos, len(c.instructions))
	}

	// Compile update
	if n.Update != nil {
		_, err = c.Compile(n.Update)
		if err != nil {
			return 0, err
		}
	}

	// Jump back to condition check
	c.emitRegJump(startPos - len(c.instructions))

	// Patch jump to end
	if n.Condition != nil {
		c.patchJump(jumpIfFalsePos)
	}

	// Patch breaks
	ctx = &c.loopContexts[len(c.loopContexts)-1]
	for _, pos := range ctx.breakPos {
		c.patchJump(pos)
	}
	c.loopContexts = c.loopContexts[:len(c.loopContexts)-1]

	return 0, nil
}

// tryOptimizeSimpleCountingLoop attempts to detect and optimize simple counting loops
// Pattern: for (var i = start; i < limit; i++) { acc += i }
// Returns true if optimization was applied, false otherwise
func (c *RegCompiler) tryOptimizeSimpleCountingLoop(n *parser.ForStatement) bool {
	// Check for the pattern:
	// 1. Init: var i = <int>
	// 2. Condition: i < <int>
	// 3. Update: i++ or i += 1
	// 4. Body: acc += i (where acc is a variable)

	// For now, check if we have the basic structure
	if n.Init == nil || n.Condition == nil || n.Update == nil || n.Body == nil {
		return false
	}

	// Check init: var i = <int>
	varInit, ok := n.Init.(*parser.VarStatement)
	if !ok {
		return false
	}
	initInt, ok := varInit.Value.(*parser.IntegerLiteral)
	if !ok {
		return false
	}

	// Check condition: i < <int>
	condInfix, ok := n.Condition.(*parser.InfixExpression)
	if !ok {
		return false
	}
	if condInfix.Operator != "<" {
		return false
	}
	condLeft, ok := condInfix.Left.(*parser.Identifier)
	if !ok || condLeft.Value != varInit.Name.Value {
		return false
	}
	condRightInt, ok := condInfix.Right.(*parser.IntegerLiteral)
	if !ok {
		return false
	}

	// Check update: i++ (as expression statement) or i += 1 or i = i + 1
	var isIncrement bool
	switch update := n.Update.(type) {
	case *parser.ExpressionStatement:
		// Check for postfix expression: i++
		if postExpr, ok := update.Expression.(*parser.PostfixExpression); ok {
			if postExpr.Operator != "++" {
				return false
			}
			ident, ok := postExpr.Left.(*parser.Identifier)
			if !ok || ident.Value != varInit.Name.Value {
				return false
			}
			isIncrement = true
		}
		// Check for compound assignment: i += 1
		if compoundExpr, ok := update.Expression.(*parser.CompoundAssignmentExpression); ok {
			if compoundExpr.Operator != "+=" {
				return false
			}
			ident, ok := compoundExpr.Left.(*parser.Identifier)
			if !ok || ident.Value != varInit.Name.Value {
				return false
			}
			updateRight, ok := compoundExpr.Right.(*parser.IntegerLiteral)
			if !ok || updateRight.Value != 1 {
				return false
			}
			isIncrement = true
		}
		// Check for assignment: i = i + 1
		if assignExpr, ok := update.Expression.(*parser.AssignmentExpression); ok {
			assignLeft, ok := assignExpr.Left.(*parser.Identifier)
			if !ok || assignLeft.Value != varInit.Name.Value {
				return false
			}
			// Check: i + 1
			infixRight, ok := assignExpr.Value.(*parser.InfixExpression)
			if !ok || infixRight.Operator != "+" {
				return false
			}
			infixLeft, ok := infixRight.Left.(*parser.Identifier)
			if !ok || infixLeft.Value != varInit.Name.Value {
				return false
			}
			infixRightInt, ok := infixRight.Right.(*parser.IntegerLiteral)
			if !ok || infixRightInt.Value != 1 {
				return false
			}
			isIncrement = true
		}
	default:
		return false
	}

	if !isIncrement {
		return false
	}

	// Check body: simple accumulator pattern
	// Body should be: acc += i or total = total + i
	block := n.Body // n.Body is already *parser.BlockStatement
	if len(block.Statements) != 1 {
		return false
	}

	// Check if body is: acc += i
	compoundAssign, ok := block.Statements[0].(*parser.ExpressionStatement)
	if !ok {
		return false
	}
	compoundExpr, ok := compoundAssign.Expression.(*parser.CompoundAssignmentExpression)
	if !ok {
		// Try infix expression: total = total + i
		assignExpr, ok := compoundAssign.Expression.(*parser.AssignmentExpression)
		if !ok {
			return false
		}
		// Check: total = total + i
		assignLeft, ok := assignExpr.Left.(*parser.Identifier)
		if !ok {
			return false
		}
		infixRight, ok := assignExpr.Value.(*parser.InfixExpression)
		if !ok || infixRight.Operator != "+" {
			return false
		}
		// Check: total + i
		infixLeft, ok := infixRight.Left.(*parser.Identifier)
		if !ok || infixLeft.Value != assignLeft.Value {
			return false
		}
		infixRightIdent, ok := infixRight.Right.(*parser.Identifier)
		if !ok || infixRightIdent.Value != varInit.Name.Value {
			return false
		}

		// Pattern matched! Generate optimized code
		return c.emitOptimizedCountingLoop(assignLeft.Value, varInit.Name.Value, int(initInt.Value), int(condRightInt.Value))
	}

	if compoundExpr.Operator != "+=" {
		return false
	}

	accIdent, ok := compoundExpr.Left.(*parser.Identifier)
	if !ok {
		return false
	}

	counterIdent, ok := compoundExpr.Right.(*parser.Identifier)
	if !ok || counterIdent.Value != varInit.Name.Value {
		return false
	}

	// Pattern matched! Generate optimized code
	return c.emitOptimizedCountingLoop(accIdent.Value, varInit.Name.Value, int(initInt.Value), int(condRightInt.Value))
}

// emitOptimizedCountingLoop emits an optimized counting loop instruction
func (c *RegCompiler) emitOptimizedCountingLoop(accName, counterName string, start, limit int) bool {
	// Get or create symbols for accumulator and counter
	accSymbol, accOk := c.symbolTable.Resolve(accName)
	if !accOk {
		// Accumulator doesn't exist yet, can't optimize
		return false
	}

	counterSymbol := c.symbolTable.Define(counterName)

	// Allocate registers
	accReg := c.allocTempReg()
	counterReg := c.allocTempReg()

	// Add constants for start, limit, step
	startIdx := c.addConstant(objects.NewInt(int64(start)))
	limitIdx := c.addConstant(objects.NewInt(int64(limit)))
	stepIdx := c.addConstant(objects.NewInt(1)) // step = 1

	// Load current accumulator value
	if accSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(accReg, accSymbol.Index)
	} else {
		c.emitRegLoadLocal(accReg, accSymbol.Index)
	}

	// Emit the optimized loop instruction
	// Format: opcode(1) | acc_reg(1) | counter_reg(1) | start_idx(2) | limit_idx(2) | step_idx(2)
	c.instructions = append(c.instructions,
		byte(OpRegLoopCountAdd),
		byte(accReg),
		byte(counterReg),
		byte(startIdx>>8), byte(startIdx),
		byte(limitIdx>>8), byte(limitIdx),
		byte(stepIdx>>8), byte(stepIdx),
	)

	// Store results back
	if accSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(accReg, accSymbol.Index)
	} else {
		c.emitRegStoreLocal(accReg, accSymbol.Index)
	}

	if counterSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(counterReg, counterSymbol.Index)
	} else {
		c.emitRegStoreLocal(counterReg, counterSymbol.Index)
	}

	c.freeTempReg(accReg)
	c.freeTempReg(counterReg)

	return true
}

// tryOptimizePrimeCheckLoop attempts to detect and optimize prime checking inner loops
// Pattern: for (var i = start; i * i <= n; i++) { if (n % i == 0) { result = false; break } }
// Returns true if optimization was applied, false otherwise
func (c *RegCompiler) tryOptimizePrimeCheckLoop(n *parser.ForStatement) bool {
	// Check for the pattern:
	// 1. Init: var i = <int>
	// 2. Condition: i * i <= n
	// 3. Update: i++ or i += 1 or i = i + 1
	// 4. Body: if (n % i == 0) { result = false; break }

	// For now, check if we have the basic structure
	if n.Init == nil || n.Condition == nil || n.Update == nil || n.Body == nil {
		return false
	}

	// Check init: var i = <int>
	varInit, ok := n.Init.(*parser.VarStatement)
	if !ok {
		return false
	}
	initInt, ok := varInit.Value.(*parser.IntegerLiteral)
	if !ok {
		return false
	}

	// Check condition: i * i <= n
	condInfix, ok := n.Condition.(*parser.InfixExpression)
	if !ok {
		return false
	}
	if condInfix.Operator != "<=" {
		return false
	}

	// Check left side: i * i
	mulLeft, ok := condInfix.Left.(*parser.InfixExpression)
	if !ok || mulLeft.Operator != "*" {
		return false
	}
	mulLeftLeft, ok := mulLeft.Left.(*parser.Identifier)
	if !ok || mulLeftLeft.Value != varInit.Name.Value {
		return false
	}
	mulLeftRight, ok := mulLeft.Right.(*parser.Identifier)
	if !ok || mulLeftRight.Value != varInit.Name.Value {
		return false
	}

	// Check right side: n (identifier)
	nIdent, ok := condInfix.Right.(*parser.Identifier)
	if !ok {
		return false
	}

	// Check update: i++ or i += 1 or i = i + 1
	var isIncrement bool
	switch update := n.Update.(type) {
	case *parser.ExpressionStatement:
		if postExpr, ok := update.Expression.(*parser.PostfixExpression); ok {
			if postExpr.Operator == "++" {
				ident, ok := postExpr.Left.(*parser.Identifier)
				if ok && ident.Value == varInit.Name.Value {
					isIncrement = true
				}
			}
		}
		if compoundExpr, ok := update.Expression.(*parser.CompoundAssignmentExpression); ok {
			if compoundExpr.Operator == "+=" {
				ident, ok := compoundExpr.Left.(*parser.Identifier)
				if ok && ident.Value == varInit.Name.Value {
					if rightInt, ok := compoundExpr.Right.(*parser.IntegerLiteral); ok && rightInt.Value == 1 {
						isIncrement = true
					}
				}
			}
		}
		if assignExpr, ok := update.Expression.(*parser.AssignmentExpression); ok {
			assignLeft, ok := assignExpr.Left.(*parser.Identifier)
			if ok && assignLeft.Value == varInit.Name.Value {
				if infixRight, ok := assignExpr.Value.(*parser.InfixExpression); ok && infixRight.Operator == "+" {
					if infixLeft, ok := infixRight.Left.(*parser.Identifier); ok && infixLeft.Value == varInit.Name.Value {
						if rightInt, ok := infixRight.Right.(*parser.IntegerLiteral); ok && rightInt.Value == 1 {
							isIncrement = true
						}
					}
				}
			}
		}
	}

	if !isIncrement {
		return false
	}

	// Check body: if (n % i == 0) { result = false; break }
	block := n.Body
	if len(block.Statements) != 1 {
		return false
	}

	ifStmt, ok := block.Statements[0].(*parser.IfStatement)
	if !ok {
		return false
	}

	// Check condition: n % i == 0
	modCond, ok := ifStmt.Condition.(*parser.InfixExpression)
	if !ok || modCond.Operator != "==" {
		return false
	}
	modLeft, ok := modCond.Left.(*parser.InfixExpression)
	if !ok || modLeft.Operator != "%" {
		return false
	}
	modLeftLeft, ok := modLeft.Left.(*parser.Identifier)
	if !ok || modLeftLeft.Value != nIdent.Value {
		return false
	}
	modLeftRight, ok := modLeft.Right.(*parser.Identifier)
	if !ok || modLeftRight.Value != varInit.Name.Value {
		return false
	}
	modRight, ok := modCond.Right.(*parser.IntegerLiteral)
	if !ok || modRight.Value != 0 {
		return false
	}

	// Check consequence: result = false; break
	consBlock := ifStmt.Consequence
	if len(consBlock.Statements) != 2 {
		return false
	}

	// Check: result = false
	assignStmt, ok := consBlock.Statements[0].(*parser.ExpressionStatement)
	if !ok {
		return false
	}
	assignExpr, ok := assignStmt.Expression.(*parser.AssignmentExpression)
	if !ok {
		return false
	}
	resultIdent, ok := assignExpr.Left.(*parser.Identifier)
	if !ok {
		return false
	}
	falseLit, ok := assignExpr.Value.(*parser.BooleanLiteral)
	if !ok || falseLit.Value {
		return false
	}

	// Check: break
	_, ok = consBlock.Statements[1].(*parser.BreakStatement)
	if !ok {
		return false
	}

	// Pattern matched! Generate optimized code
	return c.emitOptimizedPrimeCheckLoop(nIdent.Value, varInit.Name.Value, resultIdent.Value, int(initInt.Value))
}

// emitOptimizedPrimeCheckLoop emits optimized prime checking loop
func (c *RegCompiler) emitOptimizedPrimeCheckLoop(nVar, iVar, resultVar string, startVal int) bool {
	// Get or create symbols
	nSymbol, nOk := c.symbolTable.Resolve(nVar)
	if !nOk {
		return false
	}

	iSymbol := c.symbolTable.Define(iVar)
	resultSymbol := c.symbolTable.Define(resultVar)

	// Allocate registers
	nReg := c.allocTempReg()
	iReg := c.allocTempReg()
	resultReg := c.allocTempReg()

	// Initialize i = start
	startIdx := c.addConstant(objects.NewInt(int64(startVal)))
	c.emitRegLoadConst(iReg, startIdx)

	// Load n
	if nSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(nReg, nSymbol.Index)
	} else {
		c.emitRegLoadLocal(nReg, nSymbol.Index)
	}

	// Initialize result = true
	c.emitRegTrue(resultReg)

	// Emit OpRegInnerLoopPrime - this handles:
	// 1. Check i*i > n -> done (is prime)
	// 2. Check n % i == 0 -> result=false, done
	// 3. i++, continue
	// We'll patch the jump offsets later
	c.emitRegInnerLoopPrime(nReg, iReg, resultReg, 0, 0) // placeholder offsets
	instrPos := len(c.instructions) - 8

	// Loop end position
	loopEnd := len(c.instructions)

	// Patch the jump offsets
	// jump_is_prime: jump to loopEnd (set result=true already done)
	// jump_done: jump to loopEnd
	offsetIsPrime := loopEnd - instrPos
	offsetDone := loopEnd - instrPos

	c.instructions[instrPos+4] = byte(offsetIsPrime >> 8)
	c.instructions[instrPos+5] = byte(offsetIsPrime)
	c.instructions[instrPos+6] = byte(offsetDone >> 8)
	c.instructions[instrPos+7] = byte(offsetDone)

	// Store results
	if resultSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(resultReg, resultSymbol.Index)
	} else {
		c.emitRegStoreLocal(resultReg, resultSymbol.Index)
	}

	if iSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(iReg, iSymbol.Index)
	} else {
		c.emitRegStoreLocal(iReg, iSymbol.Index)
	}

	c.freeTempReg(nReg)
	c.freeTempReg(iReg)
	c.freeTempReg(resultReg)

	return true
}

// tryOptimizePrimeCheckWhile attempts to optimize while loops used for prime checking
// Pattern: while (i * i <= n) { if (n % i == 0) { return false } i = i + 1 }
func (c *RegCompiler) tryOptimizePrimeCheckWhile(n *parser.WhileStatement) bool {
	// Check condition: i * i <= n
	condInfix, ok := n.Condition.(*parser.InfixExpression)
	if !ok || condInfix.Operator != "<=" {
		return false
	}

	// Check left side: i * i
	mulLeft, ok := condInfix.Left.(*parser.InfixExpression)
	if !ok || mulLeft.Operator != "*" {
		return false
	}
	mulLeftLeft, ok := mulLeft.Left.(*parser.Identifier)
	if !ok {
		return false
	}
	mulLeftRight, ok := mulLeft.Right.(*parser.Identifier)
	if !ok || mulLeftRight.Value != mulLeftLeft.Value {
		return false
	}
	iVar := mulLeftLeft.Value

	// Check right side: n (identifier)
	nIdent, ok := condInfix.Right.(*parser.Identifier)
	if !ok {
		return false
	}
	nVar := nIdent.Value

	// Check body: single if statement with optional increment
	// n.Body is already *parser.BlockStatement
	block := n.Body

	// The body should be: if (n % i == 0) { return false } i = i + 1
	// Or just: if (n % i == 0) { return false / break }
	if len(block.Statements) < 1 {
		return false
	}

	// Check if statement
	ifStmt, ok := block.Statements[0].(*parser.IfStatement)
	if !ok {
		return false
	}

	// Check condition: n % i == 0
	modCond, ok := ifStmt.Condition.(*parser.InfixExpression)
	if !ok || modCond.Operator != "==" {
		return false
	}
	modLeft, ok := modCond.Left.(*parser.InfixExpression)
	if !ok || modLeft.Operator != "%" {
		return false
	}
	modLeftLeft, ok := modLeft.Left.(*parser.Identifier)
	if !ok || modLeftLeft.Value != nVar {
		return false
	}
	modLeftRight, ok := modLeft.Right.(*parser.Identifier)
	if !ok || modLeftRight.Value != iVar {
		return false
	}
	modRight, ok := modCond.Right.(*parser.IntegerLiteral)
	if !ok || modRight.Value != 0 {
		return false
	}

	// Check consequence: return false or isPrime = false; break
	consBlock := ifStmt.Consequence
	if len(consBlock.Statements) < 1 {
		return false
	}

	// Accept: return false OR isPrime = false; break
	var hasReturnOrBreak bool
	if len(consBlock.Statements) == 1 {
		if _, ok := consBlock.Statements[0].(*parser.ReturnStatement); ok {
			hasReturnOrBreak = true
		}
		if _, ok := consBlock.Statements[0].(*parser.BreakStatement); ok {
			hasReturnOrBreak = true
		}
	} else if len(consBlock.Statements) == 2 {
		// Check: result = false; break
		if assignStmt, ok := consBlock.Statements[0].(*parser.ExpressionStatement); ok {
			if assignExpr, ok := assignStmt.Expression.(*parser.AssignmentExpression); ok {
				if _, ok := assignExpr.Left.(*parser.Identifier); ok {
					if falseLit, ok := assignExpr.Value.(*parser.BooleanLiteral); ok && !falseLit.Value {
						if _, ok := consBlock.Statements[1].(*parser.BreakStatement); ok {
							hasReturnOrBreak = true
						}
					}
				}
			}
		}
	}

	if !hasReturnOrBreak {
		return false
	}

	// Check for increment at end of body: i = i + 1
	if len(block.Statements) >= 2 {
		assignStmt, ok := block.Statements[len(block.Statements)-1].(*parser.ExpressionStatement)
		if ok {
			assignExpr, ok := assignStmt.Expression.(*parser.AssignmentExpression)
			if ok {
				assignLeft, ok := assignExpr.Left.(*parser.Identifier)
				if ok && assignLeft.Value == iVar {
					infixRight, ok := assignExpr.Value.(*parser.InfixExpression)
					if ok && infixRight.Operator == "+" {
						infixLeft, ok := infixRight.Left.(*parser.Identifier)
						if ok && infixLeft.Value == iVar {
							if rightInt, ok := infixRight.Right.(*parser.IntegerLiteral); ok && rightInt.Value == 1 {
								// Pattern matched!
								return c.emitOptimizedPrimeCheckWhileLoop(nVar, iVar)
							}
						}
					}
				}
			}
		}
	}

	return false
}

// emitOptimizedPrimeCheckWhileLoop emits optimized prime checking while loop
func (c *RegCompiler) emitOptimizedPrimeCheckWhileLoop(nVar, iVar string) bool {
	// Get or create symbols
	nSymbol, nOk := c.symbolTable.Resolve(nVar)
	if !nOk {
		return false
	}

	iSymbol, iOk := c.symbolTable.Resolve(iVar)
	if !iOk {
		return false
	}

	// Allocate registers
	nReg := c.allocTempReg()
	iReg := c.allocTempReg()

	// Load n
	if nSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(nReg, nSymbol.Index)
	} else {
		c.emitRegLoadLocal(nReg, nSymbol.Index)
	}

	// Load i
	if iSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(iReg, iSymbol.Index)
	} else {
		c.emitRegLoadLocal(iReg, iSymbol.Index)
	}

	// Loop: check i*i > n, if so we're done (prime)
	// else continue checking
	loopStart := len(c.instructions)

	// Emit OpRegLoopMulCheck
	// If i*i > n, jump out of loop
	c.emitRegLoopMulCheck(iReg, nReg, 0) // placeholder jump offset
	mulCheckPos := len(c.instructions) - 5

	// Emit: check n % i == 0
	// If n % i == 0, not prime - return false
	c.emitRegModCheckZero(ReturnRegister, nReg, iReg)
	// If result is false (n % i == 0), return false
	// We need a conditional jump here
	c.emitRegJumpIfFalse(ReturnRegister, 0) // placeholder: jump to "not prime"
	notPrimeJumpPos := len(c.instructions) - 4

	// Increment i
	c.emitRegAddConst(iReg, iReg, c.addConstant(objects.NewInt(1)))

	// Jump back to loop start
	c.emitRegJump(loopStart - len(c.instructions))

	// End of loop (is prime)
	endPos := len(c.instructions)

	// Patch the jump out offset
	offset := endPos - mulCheckPos
	c.instructions[mulCheckPos+3] = byte(offset >> 8)
	c.instructions[mulCheckPos+4] = byte(offset)

	// Return true (is prime)
	c.emitRegTrue(ReturnRegister)
	c.emitRegReturn(ReturnRegister)

	// Patch not prime jump (not used in this case, we return from within)
	notPrimeOffset := len(c.instructions) - notPrimeJumpPos
	c.instructions[notPrimeJumpPos+2] = byte(notPrimeOffset >> 8)
	c.instructions[notPrimeJumpPos+3] = byte(notPrimeOffset)

	// Store i back
	if iSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(iReg, iSymbol.Index)
	} else {
		c.emitRegStoreLocal(iReg, iSymbol.Index)
	}

	c.freeTempReg(nReg)
	c.freeTempReg(iReg)

	return true
}

// tryOptimizeNestedLoop attempts to detect and optimize nested loop patterns
// Pattern: for (i = 0; i < n; i++) { for (j = 0; j < m; j++) { acc += a[i][j] } }
// Returns true if optimization was applied, false otherwise
func (c *RegCompiler) tryOptimizeNestedLoop(n *parser.ForStatement) bool {
	// Check for outer loop structure
	if n.Init == nil || n.Condition == nil || n.Update == nil || n.Body == nil {
		return false
	}

	// Check outer init: var i = 0
	outerInit, ok := n.Init.(*parser.VarStatement)
	if !ok {
		return false
	}
	outerInitInt, ok := outerInit.Value.(*parser.IntegerLiteral)
	if !ok || outerInitInt.Value != 0 {
		return false
	}

	// Check outer condition: i < limit
	outerCond, ok := n.Condition.(*parser.InfixExpression)
	if !ok || outerCond.Operator != "<" {
		return false
	}
	outerCondLeft, ok := outerCond.Left.(*parser.Identifier)
	if !ok || outerCondLeft.Value != outerInit.Name.Value {
		return false
	}
	outerLimit, ok := outerCond.Right.(*parser.IntegerLiteral)
	if !ok {
		return false
	}

	// Check body contains an inner for loop
	block := n.Body
	if len(block.Statements) != 1 {
		return false
	}

	innerFor, ok := block.Statements[0].(*parser.ForStatement)
	if !ok {
		return false
	}

	// Check inner loop structure
	if innerFor.Init == nil || innerFor.Condition == nil || innerFor.Update == nil || innerFor.Body == nil {
		return false
	}

	// Check inner init: var j = 0
	innerInit, ok := innerFor.Init.(*parser.VarStatement)
	if !ok {
		return false
	}
	innerInitInt, ok := innerInit.Value.(*parser.IntegerLiteral)
	if !ok || innerInitInt.Value != 0 {
		return false
	}

	// Check inner condition: j < limit
	innerCond, ok := innerFor.Condition.(*parser.InfixExpression)
	if !ok || innerCond.Operator != "<" {
		return false
	}
	innerCondLeft, ok := innerCond.Left.(*parser.Identifier)
	if !ok || innerCondLeft.Value != innerInit.Name.Value {
		return false
	}
	innerLimit, ok := innerCond.Right.(*parser.IntegerLiteral)
	if !ok {
		return false
	}

	// Check inner body: simple accumulator pattern
	innerBlock := innerFor.Body
	if len(innerBlock.Statements) != 1 {
		return false
	}

	// Check for accumulator pattern: acc += something
	compoundAssign, ok := innerBlock.Statements[0].(*parser.ExpressionStatement)
	if !ok {
		return false
	}
	compoundExpr, ok := compoundAssign.Expression.(*parser.CompoundAssignmentExpression)
	if !ok || compoundExpr.Operator != "+=" {
		return false
	}

	// Pattern matched! We can emit optimized nested loop
	// For now, we'll emit optimized code for the common pattern
	// where we're summing over two dimensions

	// Define the loop variables
	_ = c.symbolTable.Define(outerInit.Name.Value)
	_ = c.symbolTable.Define(innerInit.Name.Value)

	// Get accumulator
	accIdent, ok := compoundExpr.Left.(*parser.Identifier)
	if !ok {
		return false
	}
	accSymbol, accOk := c.symbolTable.Resolve(accIdent.Value)
	if !accOk {
		return false
	}

	// Allocate registers
	accReg := c.allocTempReg()
	iReg := c.allocTempReg()
	jReg := c.allocTempReg()

	// Load accumulator
	if accSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(accReg, accSymbol.Index)
	} else {
		c.emitRegLoadLocal(accReg, accSymbol.Index)
	}

	// Initialize i = 0
	zeroIdx := c.addConstant(objects.NewInt(0))
	c.emitRegLoadConst(iReg, zeroIdx)

	// Outer loop start
	outerStart := len(c.instructions)

	// Check i < outerLimit
	outerLimitIdx := c.addConstant(objects.NewInt(outerLimit.Value))
	outerLimitReg := c.allocTempReg()
	c.emitRegLoadConst(outerLimitReg, outerLimitIdx)
	c.emitRegLess(0, iReg, outerLimitReg) // Use R0 for condition
	outerJumpFalse := c.emitRegJumpIfFalse(0, 0)
	c.freeTempReg(outerLimitReg)

	// Initialize j = 0
	c.emitRegLoadConst(jReg, zeroIdx)

	// Inner loop start
	innerStart := len(c.instructions)

	// Check j < innerLimit
	innerLimitIdx := c.addConstant(objects.NewInt(innerLimit.Value))
	innerLimitReg := c.allocTempReg()
	c.emitRegLoadConst(innerLimitReg, innerLimitIdx)
	c.emitRegLess(0, jReg, innerLimitReg) // Use R0 for condition
	innerJumpFalse := c.emitRegJumpIfFalse(0, 0)
	c.freeTempReg(innerLimitReg)

	// Compile the accumulator update
	// For simplicity, we'll compile the right side normally
	rightReg, err := c.Compile(compoundExpr.Right)
	if err != nil {
		return false
	}

	// acc += right
	c.emitRegAdd(accReg, accReg, rightReg)
	c.freeTempReg(rightReg)

	// j++
	c.emitRegAddConst(jReg, jReg, c.addConstant(objects.NewInt(1)))

	// Jump back to inner loop start
	c.emitRegJump(innerStart - len(c.instructions))

	// Patch inner loop exit
	c.patchJump(innerJumpFalse)

	// i++
	c.emitRegAddConst(iReg, iReg, c.addConstant(objects.NewInt(1)))

	// Jump back to outer loop start
	c.emitRegJump(outerStart - len(c.instructions))

	// Patch outer loop exit
	c.patchJump(outerJumpFalse)

	// Store accumulator back
	if accSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(accReg, accSymbol.Index)
	} else {
		c.emitRegStoreLocal(accReg, accSymbol.Index)
	}

	c.freeTempReg(accReg)
	c.freeTempReg(iReg)
	c.freeTempReg(jReg)

	return true
}

// compileForInStatement compiles a for-in statement
// for (value in iterable) { body }
// for (key, value in iterable) { body }
func (c *RegCompiler) compileForInStatement(n *parser.ForInStatement) (int, error) {
	// Generate unique names for this loop's internal variables
	loopID := c.forInLoopCounter
	c.forInLoopCounter++
	iterVarName := fmt.Sprintf("__for_in_iter_%d__", loopID)
	indexVarName := fmt.Sprintf("__for_in_index_%d__", loopID)

	// Compile iterable expression
	iterReg, err := c.Compile(n.Iterable)
	if err != nil {
		return 0, err
	}

	// If iterReg is ReturnRegister, we need to save it to a safe temp register
	// because subsequent builtin calls (like len) will overwrite ReturnRegister
	if iterReg == ReturnRegister {
		savedReg := c.allocTempReg()
		c.emitRegMove(savedReg, iterReg)
		iterReg = savedReg
	}

	// Save iterable to a local slot to preserve it across iterations
	// This prevents the iterable from being corrupted when the loop body
	// allocates temporary registers
	iterSymbol := c.symbolTable.Define(iterVarName)
	if iterSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(iterReg, iterSymbol.Index)
	} else {
		c.emitRegStoreLocal(iterReg, iterSymbol.Index)
	}
	c.freeTempReg(iterReg)

	// Define index variable (hidden, used for iteration)
	indexSymbol := c.symbolTable.Define(indexVarName)

	// Initialize index to 0
	zeroIdx := c.addConstant(objects.NewInt(0))
	zeroReg := c.allocTempReg()
	c.emitRegLoadConst(zeroReg, zeroIdx)

	// Store index
	if indexSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(zeroReg, indexSymbol.Index)
	} else {
		c.emitRegStoreLocal(zeroReg, indexSymbol.Index)
	}
	c.freeTempReg(zeroReg)

	// Define key variable if present
	var keySymbol Symbol
	if n.Key != nil {
		keySymbol = c.symbolTable.Define(n.Key.Value)
	}

	// Define value variable
	var valueSymbol Symbol
	if n.Value != nil {
		valueSymbol = c.symbolTable.Define(n.Value.Value)
	}

	// Record loop start
	startPos := len(c.instructions)

	// Reload iterable from local slot for this iteration
	iterReg = c.allocTempReg()
	if iterSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(iterReg, iterSymbol.Index)
	} else {
		c.emitRegLoadLocal(iterReg, iterSymbol.Index)
	}

	// Load iterable length using len builtin (index 0)
	// Put iterable in R0 for builtin call
	c.emitRegMove(0, iterReg)
	c.emitRegBuiltin(0, 1) // 0 = len builtin index
	lenReg := c.allocTempReg()
	c.emitRegMove(lenReg, ReturnRegister)

	// Load current index
	indexReg := c.allocTempReg()
	if indexSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(indexReg, indexSymbol.Index)
	} else {
		c.emitRegLoadLocal(indexReg, indexSymbol.Index)
	}

	// Compare index < length
	// Use a dedicated fixed register for condition to avoid conflicts with indexReg
	// This ensures the comparison result doesn't overwrite the index
	// We use register 250 which is below ReturnRegister (255) but above normal temp registers
	const forInCondReg = 250
	c.emitRegLess(forInCondReg, indexReg, lenReg)

	// Jump to end if condition is false
	jumpIfFalsePos := c.emitRegJumpIfFalse(forInCondReg, 0)
	c.freeTempReg(lenReg)

	// Enter loop context
	c.loopContexts = append(c.loopContexts, regLoopContext{
		startPos: startPos,
	})

	// Set key variable (current key at index)
	// Use OpRegIterKey which works correctly for both arrays and maps
	if n.Key != nil {
		keyReg := c.allocTempReg()
		c.emitRegIterKey(keyReg, iterReg, indexReg)
		if keySymbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(keyReg, keySymbol.Index)
		} else {
			c.emitRegStoreLocal(keyReg, keySymbol.Index)
		}
		c.freeTempReg(keyReg)
	}

	// Set value variable (current value at index)
	// Use OpRegIterValue which works correctly for both arrays and maps
	if n.Value != nil {
		elemReg := c.allocTempReg()
		c.emitRegIterValue(elemReg, iterReg, indexReg)
		if valueSymbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(elemReg, valueSymbol.Index)
		} else {
			c.emitRegStoreLocal(elemReg, valueSymbol.Index)
		}
		c.freeTempReg(elemReg)
	}

	c.freeTempReg(indexReg)
	c.freeTempReg(iterReg)

	// Compile body
	_, err = c.Compile(n.Body)
	if err != nil {
		return 0, err
	}

	// Update position for continue
	ctx := &c.loopContexts[len(c.loopContexts)-1]
	for _, pos := range ctx.continuePos {
		c.patchJumpTo(pos, len(c.instructions))
	}

	// Increment index
	incReg := c.allocTempReg()
	if indexSymbol.Scope == GlobalScope {
		c.emitRegLoadGlobal(incReg, indexSymbol.Index)
	} else {
		c.emitRegLoadLocal(incReg, indexSymbol.Index)
	}

	oneIdx := c.addConstant(objects.NewInt(1))
	oneReg := c.allocTempReg()
	c.emitRegLoadConst(oneReg, oneIdx)
	c.emitRegAdd(incReg, incReg, oneReg)
	c.freeTempReg(oneReg)

	if indexSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(incReg, indexSymbol.Index)
	} else {
		c.emitRegStoreLocal(incReg, indexSymbol.Index)
	}
	c.freeTempReg(incReg)

	// Jump back to start
	c.emitRegJump(startPos - len(c.instructions))

	// Patch jump to end
	c.patchJump(jumpIfFalsePos)

	// Patch breaks
	ctx = &c.loopContexts[len(c.loopContexts)-1]
	for _, pos := range ctx.breakPos {
		c.patchJump(pos)
	}
	c.loopContexts = c.loopContexts[:len(c.loopContexts)-1]

	return 0, nil
}

// compileReturnStatement compiles a return statement
// Supports tail call optimization: if return value is a function call,
// emit OpRegTailCall instead of OpRegCall + OpRegReturn
func (c *RegCompiler) compileReturnStatement(n *parser.ReturnStatement) (int, error) {
	if n.ReturnValue == nil {
		c.emitRegReturn(0)
		return 0, nil
	}

	// Check for tail call opportunity: return func(args)
	if call, ok := n.ReturnValue.(*parser.CallExpression); ok {
		return c.compileTailCall(call)
	}

	valReg, err := c.Compile(n.ReturnValue)
	if err != nil {
		return 0, err
	}

	c.emitRegReturn(valReg)
	c.freeTempReg(valReg)
	return 0, nil
}

// compileTailCall compiles a tail call (return func(args))
// This emits OpRegTailCall which reuses the current frame
func (c *RegCompiler) compileTailCall(n *parser.CallExpression) (int, error) {
	// Disable tail call optimization inside try blocks
	// because the exception handler needs to remain active
	if c.tryBlockDepth > 0 {
		valReg, err := c.compileCallExpression(n)
		if err != nil {
			return 0, err
		}
		c.emitRegReturn(valReg)
		c.freeTempReg(valReg)
		return 0, nil
	}

	// Check if this is a direct builtin call - builtins don't benefit from TCO
	// and OpRegTailCall doesn't work with builtins
	if ident, ok := n.Function.(*parser.Identifier); ok {
		symbol, ok := c.symbolTable.Resolve(ident.Value)
		if ok && symbol.Scope == BuiltinScope {
			// Fall back to normal call + return for builtins
			valReg, err := c.compileCallExpression(n)
			if err != nil {
				return 0, err
			}
			c.emitRegReturn(valReg)
			c.freeTempReg(valReg)
			return 0, nil
		}
	}

	// Check if this is a method call
	if dot, ok := n.Function.(*parser.DotExpression); ok {
		// Method tail call optimization
		// Compile the object
		objReg, err := c.Compile(dot.Object)
		if err != nil {
			return 0, err
		}
		if objReg == ReturnRegister {
			tempReg := c.allocTempReg()
			c.emitRegMove(tempReg, objReg)
			objReg = tempReg
		}

		// Compile arguments to temporary registers
		argRegs := make([]int, len(n.Arguments))
		for i, arg := range n.Arguments {
			argReg, err := c.Compile(arg)
			if err != nil {
				return 0, err
			}
			if argReg == ReturnRegister {
				tempReg := c.allocTempReg()
				c.emitRegMove(tempReg, argReg)
				argReg = tempReg
			}
			argRegs[i] = argReg
		}

		// Move arguments to R0-R7
		for i, argReg := range argRegs {
			if argReg != i {
				c.emitRegMove(i, argReg)
			}
		}

		// Free temporary registers
		for i := len(argRegs) - 1; i >= 0; i-- {
			if argRegs[i] >= FirstLocalRegister {
				c.freeTempReg(argRegs[i])
			}
		}

		// Get method name constant
		nameIdx := c.addConstant(objects.InternString(dot.Property.Value))

		// Emit tail call method instruction
		c.emitRegTailCallMethod(objReg, nameIdx, len(n.Arguments))
		if objReg >= FirstLocalRegister {
			c.freeTempReg(objReg)
		}

		return 0, nil
	}

	// Regular function tail call
	// Compile function expression
	funcReg, err := c.Compile(n.Function)
	if err != nil {
		return 0, err
	}

	// Compile arguments to temporary registers
	argRegs := make([]int, len(n.Arguments))
	for i, arg := range n.Arguments {
		argReg, err := c.Compile(arg)
		if err != nil {
			return 0, err
		}
		if argReg == ReturnRegister {
			tempReg := c.allocTempReg()
			c.emitRegMove(tempReg, argReg)
			argReg = tempReg
		}
		argRegs[i] = argReg
	}

	// Move arguments to R0-R7
	for i, argReg := range argRegs {
		if argReg != i {
			c.emitRegMove(i, argReg)
		}
	}

	// Free temporary registers
	for i := len(argRegs) - 1; i >= 0; i-- {
		if argRegs[i] >= FirstLocalRegister {
			c.freeTempReg(argRegs[i])
		}
	}

	// Emit tail call instruction
	c.emitRegTailCall(funcReg, len(n.Arguments))
	c.freeTempReg(funcReg)

	return 0, nil
}

// compileFunctionLiteral compiles a function literal
func (c *RegCompiler) compileFunctionLiteral(n *parser.FunctionLiteral) (int, error) {
	// If named function, define the name first for recursion support
	var funcSymbol Symbol
	if n.Name != "" {
		funcSymbol = c.symbolTable.Define(n.Name)
	}

	// Enter function scope
	c.enterScope()

	// Define parameters as local variables
	// Parameters are passed in R0-R7, copy them to local slots
	for i, p := range n.Parameters {
		symbol := c.symbolTable.Define(p.Value)
		// Copy from argument register to local slot
		// R0-R7 -> Locals[0..n]
		c.emitRegStoreLocal(i, symbol.Index)
	}

	// Define variadic parameter if present
	// The VM will store the variadic args array at this symbol's index
	if n.VariadicParam != nil {
		c.symbolTable.Define(n.VariadicParam.Value)
		// No instruction needed - VM will set the local directly
	}

	// Compile body
	lastReg, err := c.Compile(n.Body)
	if err != nil {
		return 0, err
	}

	// Ensure function ends with return
	if len(c.instructions) == 0 || Opcode(c.instructions[len(c.instructions)-1]) != OpRegReturn {
		// Emit implicit return with the last expression's result
		// If lastReg is 0 (empty block or void), return null
		// Otherwise return the value in the appropriate register
		if lastReg == 0 {
			// Empty block or statement with no value - return null
			c.emitRegReturn(0)
		} else {
			// Return the last expression's result
			c.emitRegReturn(lastReg)
		}
	}

	// Leave scope and get compiled function
	fn := c.leaveScope()
	fn.NumParameters = len(n.Parameters)
	fn.Variadic = n.VariadicParam != nil

	// Add function to constants
	fnIndex := c.addConstant(fn)

	// Allocate register for closure
	dst := c.allocTempReg()

	numFree := len(fn.FreeVariables)
	if numFree == 0 {
		// No free variables - just load the function
		c.emitRegLoadConst(dst, fnIndex)
	} else {
		// Load free variables into temporary registers
		freeStartReg := c.nextTempReg
		for i, freeVar := range fn.FreeVariables {
			freeReg := c.allocTempReg()
			switch freeVar.Scope {
			case GlobalScope:
				c.emitRegLoadGlobal(freeReg, freeVar.Index)
			case LocalScope:
				c.emitRegLoadLocal(freeReg, freeVar.Index)
			case FreeScope:
				c.emitRegLoadFree(freeReg, freeVar.Index)
			}
			_ = i // freeReg already equals freeStartReg + i due to sequential allocation
		}

		// Emit closure instruction
		// Format: OpRegClosure dst func_idx num_free start_reg
		c.instructions = append(c.instructions,
			byte(OpRegClosure),
			byte(dst),
			byte(fnIndex>>8),
			byte(fnIndex),
			byte(numFree),
			byte(freeStartReg),
		)

		// Free the temporary registers used for free variables
		for i := 0; i < numFree; i++ {
			c.freeTempReg(freeStartReg + i)
		}
	}

	// If named function, bind it to its name
	if n.Name != "" {
		switch funcSymbol.Scope {
		case GlobalScope:
			c.emitRegStoreGlobal(dst, funcSymbol.Index)
		case LocalScope:
			c.emitRegStoreLocal(dst, funcSymbol.Index)
		}
		// The function is now stored, result is the function itself
	}

	return dst, nil
}

// compileCallExpression compiles a function call
func (c *RegCompiler) compileCallExpression(n *parser.CallExpression) (int, error) {
	// Check if this is a direct builtin call (e.g., len("hello"))
	if ident, ok := n.Function.(*parser.Identifier); ok {
		symbol, ok := c.symbolTable.Resolve(ident.Value)
		if ok && symbol.Scope == BuiltinScope {
			// This is a builtin - use OpRegBuiltin directly
			// First, compile all arguments to temporary registers
			// This prevents nested calls from overwriting argument registers
			argRegs := make([]int, len(n.Arguments))
			for i, arg := range n.Arguments {
				argReg, err := c.Compile(arg)
				if err != nil {
					return 0, err
				}
				// If the result is in ReturnRegister (from a nested call),
				// move it to a temp register to preserve it
				if argReg == ReturnRegister {
					tempReg := c.allocTempReg()
					c.emitRegMove(tempReg, argReg)
					argReg = tempReg
				}
				argRegs[i] = argReg
			}

			// Now move all arguments to their final positions (R0-R7)
			for i, argReg := range argRegs {
				if argReg != i {
					c.emitRegMove(i, argReg)
				}
			}

			// Free temporary registers (in reverse order for proper stack-like freeing)
			for i := len(argRegs) - 1; i >= 0; i-- {
				if argRegs[i] >= FirstLocalRegister {
					c.freeTempReg(argRegs[i])
				}
			}

			// Emit builtin call
			c.emitRegBuiltin(symbol.Index, len(n.Arguments))
			return ReturnRegister, nil
		}
	}

	// Check if this is a method call (obj.method())
	if dot, ok := n.Function.(*parser.DotExpression); ok {
		// Compile the object
		objReg, err := c.Compile(dot.Object)
		if err != nil {
			return 0, err
		}

		// First, compile all arguments to temporary registers
		argRegs := make([]int, len(n.Arguments))
		for i, arg := range n.Arguments {
			argReg, err := c.Compile(arg)
			if err != nil {
				return 0, err
			}
			// If the result is in ReturnRegister, move it to a temp register
			if argReg == ReturnRegister {
				tempReg := c.allocTempReg()
				c.emitRegMove(tempReg, argReg)
				argReg = tempReg
			}
			argRegs[i] = argReg
		}

		// Now move all arguments to their final positions (R0-R7)
		for i, argReg := range argRegs {
			if argReg != i {
				c.emitRegMove(i, argReg)
			}
		}

		// Free temporary registers
		for i := len(argRegs) - 1; i >= 0; i-- {
			if argRegs[i] >= FirstLocalRegister {
				c.freeTempReg(argRegs[i])
			}
		}

		// Get method name constant
		nameIdx := c.addConstant(objects.InternString(dot.Property.Value))

		// Emit method call
		c.emitRegCallMethod(objReg, nameIdx, len(n.Arguments))
		c.freeTempReg(objReg)

		return ReturnRegister, nil
	}

	// Regular function call
	// Compile function
	funcReg, err := c.Compile(n.Function)
	if err != nil {
		return 0, err
	}

	// First, compile all arguments to temporary registers
	argRegs := make([]int, len(n.Arguments))
	for i, arg := range n.Arguments {
		argReg, err := c.Compile(arg)
		if err != nil {
			return 0, err
		}
		// If the result is in ReturnRegister, move it to a temp register
		if argReg == ReturnRegister {
			tempReg := c.allocTempReg()
			c.emitRegMove(tempReg, argReg)
			argReg = tempReg
		}
		argRegs[i] = argReg
	}

	// Now move all arguments to their final positions (R0-R7)
	for i, argReg := range argRegs {
		if argReg != i {
			c.emitRegMove(i, argReg)
		}
	}

	// Free temporary registers
	for i := len(argRegs) - 1; i >= 0; i-- {
		if argRegs[i] >= FirstLocalRegister {
			c.freeTempReg(argRegs[i])
		}
	}

	// Emit call
	c.emitRegCall(funcReg, len(n.Arguments))
	c.freeTempReg(funcReg)

	// Result is in ReturnRegister
	return ReturnRegister, nil
}

// compileArrayLiteral compiles an array literal
func (c *RegCompiler) compileArrayLiteral(n *parser.ArrayLiteral) (int, error) {
	// Compile elements and collect their result registers
	// Elements may not be in contiguous registers (e.g., nested arrays)
	elementRegs := make([]int, 0, len(n.Elements))

	for _, elem := range n.Elements {
		reg, err := c.Compile(elem)
		if err != nil {
			return 0, err
		}
		elementRegs = append(elementRegs, reg)
	}

	numElements := len(n.Elements)

	// Check if we have enough contiguous register space
	// Need numElements contiguous registers for elements + 1 for dst
	available := NumRegisters - 1 - c.nextTempReg
	if available >= numElements {
		// Enough space - use existing OpRegArray (more efficient)
		// IMPORTANT: To avoid overwriting source registers during moves,
		// we first push all elements to the temp stack, then pop them into
		// contiguous target registers.

		// Push all elements to temp stack
		for _, elemReg := range elementRegs {
			c.emitRegPush(elemReg)
			c.freeTempReg(elemReg)
		}

		// Allocate contiguous target registers and pop from stack
		dst := c.allocTempReg()
		startReg := c.nextTempReg

		for i := 0; i < numElements; i++ {
			c.allocTempReg() // allocate the register slot
		}

		// Pop elements from stack in reverse order (stack is LIFO)
		for i := numElements - 1; i >= 0; i-- {
			targetReg := startReg + i
			c.emitRegPop(targetReg)
		}

		c.emitRegArray(dst, startReg, numElements)

		// Free element registers
		for i := 0; i < numElements; i++ {
			c.freeTempReg(startReg + i)
		}

		return dst, nil
	}

	// Not enough contiguous space - use incremental array building with temp stack
	// Push all elements to the temp stack first to preserve them
	for _, elemReg := range elementRegs {
		c.emitRegPush(elemReg)
	}

	// Free the original element registers
	for _, elemReg := range elementRegs {
		c.freeTempReg(elemReg)
	}

	// Use fixed registers for overflow case
	const overflowArrReg = 253
	const overflowElemReg = 254

	// Create empty array
	dst := overflowArrReg
	c.emitRegArrayEmpty(dst)

	// Pop elements from stack and append to array
	for i := len(elementRegs) - 1; i >= 0; i-- {
		// Pop element to fixed register
		c.emitRegPop(overflowElemReg)
		// Append element to array
		c.emitRegArrayAppend(dst, dst, overflowElemReg)
	}

	return dst, nil
}

// compileMapLiteral compiles a map literal
func (c *RegCompiler) compileMapLiteral(n *parser.MapLiteral) (int, error) {
	// Collect all key-value pairs
	type kvPair struct {
		keyReg, valReg int
	}
	pairs := make([]kvPair, 0, len(n.Pairs))

	// First, collect all pairs to compile (to avoid map iteration order issues)
	type keyVal struct {
		key parser.Expression
		val parser.Expression
	}
	items := make([]keyVal, 0, len(n.Pairs))
	for key, val := range n.Pairs {
		items = append(items, keyVal{key: key, val: val})
	}

	// Reserve registers for all keys and values upfront to prevent corruption
	numPairs := len(items)
	startReg := c.nextTempReg
	numRegs := numPairs * 2

	// Allocate all registers at once
	for i := 0; i < numRegs; i++ {
		c.nextTempReg++
		if c.nextTempReg > c.maxReg {
			c.maxReg = c.nextTempReg
		}
	}

	// Now compile each key-value pair into the reserved registers
	for i, item := range items {
		keyTargetReg := startReg + i*2
		valTargetReg := startReg + i*2 + 1

		// Compile key and move to target register
		keyReg, err := c.Compile(item.key)
		if err != nil {
			return 0, err
		}
		if keyReg != keyTargetReg {
			c.emitRegMove(keyTargetReg, keyReg)
			c.freeTempReg(keyReg)
		}

		// Compile value and move to target register
		valReg, err := c.Compile(item.val)
		if err != nil {
			return 0, err
		}
		if valReg != valTargetReg {
			c.emitRegMove(valTargetReg, valReg)
			c.freeTempReg(valReg)
		}

		pairs = append(pairs, kvPair{keyReg: keyTargetReg, valReg: valTargetReg})
	}

	count := len(pairs)

	// Check if we have enough contiguous register space
	// Need count*2 contiguous registers for key-value pairs
	available := NumRegisters - 1 - c.nextTempReg
	if available >= count*2 {
		// Enough space - use existing OpRegMap (more efficient)
		dst := c.allocTempReg()
		startReg := c.nextTempReg

		for i, pair := range pairs {
			targetKeyReg := startReg + i*2
			targetValReg := startReg + i*2 + 1

			// Allocate these registers
			for c.nextTempReg <= targetValReg {
				c.allocTempReg()
			}

			// Move key and value to target positions
			if pair.keyReg != targetKeyReg {
				c.emitRegMove(targetKeyReg, pair.keyReg)
			}
			if pair.valReg != targetValReg {
				c.emitRegMove(targetValReg, pair.valReg)
			}
		}

		c.emitRegMap(dst, startReg, count)

		// Free all temporary registers
		for i := 0; i < count*2; i++ {
			c.freeTempReg(startReg + i)
		}

		return dst, nil
	}

	// Not enough contiguous space - use incremental map building with temp stack
	// Push all key-value pairs to the temp stack first to preserve them
	for _, pair := range pairs {
		c.emitRegPush(pair.keyReg)
		c.emitRegPush(pair.valReg)
	}

	// Free the original pair registers now that values are on stack
	for _, pair := range pairs {
		c.freeTempReg(pair.keyReg)
		c.freeTempReg(pair.valReg)
	}

	// Use fixed registers for overflow case to avoid conflicts
	// R252 = map result
	// R253 = key temp
	// R254 = value temp
	// R255 = return register (reserved)
	const overflowMapReg = 252
	const overflowKeyReg = 253
	const overflowValReg = 254

	dst := overflowMapReg
	c.emitRegMapEmpty(dst)

	// Pop pairs from stack and add to map
	for i := len(pairs) - 1; i >= 0; i-- {
		// Pop value first (it was pushed last)
		c.emitRegPop(overflowValReg)
		// Pop key
		c.emitRegPop(overflowKeyReg)
		// Add to map
		c.emitRegMapSet(dst, dst, overflowKeyReg, overflowValReg)
	}

	return dst, nil
}

// compileIndexExpression compiles an index expression
func (c *RegCompiler) compileIndexExpression(n *parser.IndexExpression) (int, error) {
	leftReg, err := c.Compile(n.Left)
	if err != nil {
		return 0, err
	}

	indexReg, err := c.Compile(n.Index)
	if err != nil {
		return 0, err
	}

	dst := c.allocTempReg()
	c.emitRegIndex(dst, leftReg, indexReg)

	c.freeTempReg(leftReg)
	c.freeTempReg(indexReg)

	return dst, nil
}

// compileSliceExpression compiles a slice expression (a[start:end])
func (c *RegCompiler) compileSliceExpression(n *parser.SliceExpression) (int, error) {
	leftReg, err := c.Compile(n.Left)
	if err != nil {
		return 0, err
	}

	var startReg, endReg int

	// Compile start index (use 0 if nil)
	if n.Start != nil {
		startReg, err = c.Compile(n.Start)
		if err != nil {
			return 0, err
		}
	} else {
		startReg = c.allocTempReg()
		c.emitRegNull(startReg)
	}

	// Compile end index (use -1 if nil, will be treated as "to end" in VM)
	if n.End != nil {
		endReg, err = c.Compile(n.End)
		if err != nil {
			return 0, err
		}
	} else {
		endReg = c.allocTempReg()
		c.emitRegNull(endReg)
	}

	dst := c.allocTempReg()
	c.emitRegSlice(dst, leftReg, startReg, endReg)

	c.freeTempReg(leftReg)
	c.freeTempReg(startReg)
	c.freeTempReg(endReg)

	return dst, nil
}

// compileDotExpression compiles a dot expression (obj.field or obj.method)
func (c *RegCompiler) compileDotExpression(n *parser.DotExpression) (int, error) {
	// Compile the object
	objReg, err := c.Compile(n.Object)
	if err != nil {
		return 0, err
	}

	// Get the property name
	nameIdx := c.addConstant(objects.InternString(n.Property.Value))

	// Allocate result register
	dst := c.allocTempReg()

	// Emit get field/method instruction
	c.emitRegGetField(dst, objReg, nameIdx)

	c.freeTempReg(objReg)

	return dst, nil
}

// compileAssignmentExpression compiles an assignment expression
func (c *RegCompiler) compileAssignmentExpression(n *parser.AssignmentExpression) (int, error) {
	// Compile right side
	valReg, err := c.Compile(n.Value)
	if err != nil {
		return 0, err
	}

	// Determine where to store
	switch left := n.Left.(type) {
	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(left.Value)
		if !ok {
			return 0, fmt.Errorf("undefined variable: %s", left.Value)
		}
		switch symbol.Scope {
		case GlobalScope:
			c.emitRegStoreGlobal(valReg, symbol.Index)
		case LocalScope:
			c.emitRegStoreLocal(valReg, symbol.Index)
		case FreeScope:
			c.emitRegStoreFree(valReg, symbol.Index)
		}

	case *parser.IndexExpression:
		objReg, err := c.Compile(left.Left)
		if err != nil {
			return 0, err
		}
		indexReg, err := c.Compile(left.Index)
		if err != nil {
			return 0, err
		}
		c.emitRegSetIndex(objReg, indexReg, valReg)
		c.freeTempReg(objReg)
		c.freeTempReg(indexReg)

	case *parser.DotExpression:
		objReg, err := c.Compile(left.Object)
		if err != nil {
			return 0, err
		}
		nameIdx := c.addConstant(objects.InternString(left.Property.Value))
		c.emitRegSetField(objReg, valReg, nameIdx)
		c.freeTempReg(objReg)
	}

	return valReg, nil
}

// compileBreakStatement compiles a break statement
func (c *RegCompiler) compileBreakStatement(n *parser.BreakStatement) (int, error) {
	if len(c.loopContexts) == 0 {
		return 0, fmt.Errorf("break statement outside of loop")
	}

	// Emit jump to end of loop (will be patched when loop ends)
	breakPos := c.emitRegJump(0)

	// Track the break position for patching
	ctx := &c.loopContexts[len(c.loopContexts)-1]
	ctx.breakPos = append(ctx.breakPos, breakPos)

	return 0, nil
}

// compileContinueStatement compiles a continue statement
func (c *RegCompiler) compileContinueStatement(n *parser.ContinueStatement) (int, error) {
	if len(c.loopContexts) == 0 {
		return 0, fmt.Errorf("continue statement outside of loop")
	}

	// Emit jump to update section or start of loop
	// Will be patched to jump to update position for for-loops
	continuePos := c.emitRegJump(0)

	// Track the continue position for patching
	ctx := &c.loopContexts[len(c.loopContexts)-1]
	ctx.continuePos = append(ctx.continuePos, continuePos)

	return 0, nil
}

// compileTernaryExpression compiles a ternary expression (condition ? consequent : alternative)
func (c *RegCompiler) compileTernaryExpression(n *parser.TernaryExpression) (int, error) {
	// Compile condition
	condReg, err := c.Compile(n.Condition)
	if err != nil {
		return 0, err
	}

	// Jump to alternative if condition is false
	jumpIfFalsePos := c.emitRegJumpIfFalse(condReg, 0)
	c.freeTempReg(condReg)

	// Compile consequent
	consequentReg, err := c.Compile(n.Consequent)
	if err != nil {
		return 0, err
	}

	// Allocate result register and move consequent there
	resultReg := c.allocTempReg()
	if consequentReg != resultReg {
		c.emitRegMove(resultReg, consequentReg)
		c.freeTempReg(consequentReg)
	}

	// Jump over alternative
	jumpPos := c.emitRegJump(0)

	// Patch jump to here (start of alternative)
	c.patchJump(jumpIfFalsePos)

	// Compile alternative
	alternativeReg, err := c.Compile(n.Alternative)
	if err != nil {
		return 0, err
	}

	// Move alternative to result register
	if alternativeReg != resultReg {
		c.emitRegMove(resultReg, alternativeReg)
		c.freeTempReg(alternativeReg)
	}

	// Patch jump to here (end)
	c.patchJump(jumpPos)

	return resultReg, nil
}

// compilePostfixExpression compiles a postfix expression (i++, i--)
func (c *RegCompiler) compilePostfixExpression(n *parser.PostfixExpression) (int, error) {
	// Allocate a register for the result (old value for postfix)
	resultReg := c.allocTempReg()

	// Handle different left expression types
	switch left := n.Left.(type) {
	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(left.Value)
		if !ok {
			return 0, fmt.Errorf("undefined variable: %s", left.Value)
		}

		// Allocate a register for the current value
		valReg := c.allocTempReg()

		// Load current value
		switch symbol.Scope {
		case GlobalScope:
			c.emitRegLoadGlobal(valReg, symbol.Index)
		case LocalScope:
			c.emitRegLoadLocal(valReg, symbol.Index)
		case FreeScope:
			c.emitRegLoadFree(valReg, symbol.Index)
		}

		// Copy old value to result register (postfix returns old value)
		c.emitRegMove(resultReg, valReg)

		// Load constant 1
		oneIdx := c.addConstant(objects.NewInt(1))
		oneReg := c.allocTempReg()
		c.emitRegLoadConst(oneReg, oneIdx)

		// Perform increment or decrement
		switch n.Operator {
		case "++":
			c.emitRegAdd(valReg, valReg, oneReg)
		case "--":
			c.emitRegSub(valReg, valReg, oneReg)
		default:
			return 0, fmt.Errorf("unknown postfix operator: %s", n.Operator)
		}

		// Store back to variable
		switch symbol.Scope {
		case GlobalScope:
			c.emitRegStoreGlobal(valReg, symbol.Index)
		case LocalScope:
			c.emitRegStoreLocal(valReg, symbol.Index)
		case FreeScope:
			c.emitRegStoreFree(valReg, symbol.Index)
		}

		// Free temporary registers
		c.freeTempReg(valReg)
		c.freeTempReg(oneReg)

	case *parser.IndexExpression:
		// Handle a[i]++ or a[i]--
		// Compile the object (array/map)
		objReg, err := c.Compile(left.Left)
		if err != nil {
			return 0, err
		}

		// Compile the index
		indexReg, err := c.Compile(left.Index)
		if err != nil {
			return 0, err
		}

		// Get current value from index
		valReg := c.allocTempReg()
		c.emitRegIndex(valReg, objReg, indexReg)

		// Copy old value to result register (postfix returns old value)
		c.emitRegMove(resultReg, valReg)

		// Load constant 1
		oneIdx := c.addConstant(objects.NewInt(1))
		oneReg := c.allocTempReg()
		c.emitRegLoadConst(oneReg, oneIdx)

		// Perform increment or decrement
		switch n.Operator {
		case "++":
			c.emitRegAdd(valReg, valReg, oneReg)
		case "--":
			c.emitRegSub(valReg, valReg, oneReg)
		default:
			return 0, fmt.Errorf("unknown postfix operator: %s", n.Operator)
		}

		// Store back to index
		c.emitRegSetIndex(objReg, indexReg, valReg)

		// Free temporary registers
		c.freeTempReg(objReg)
		c.freeTempReg(indexReg)
		c.freeTempReg(valReg)
		c.freeTempReg(oneReg)

	case *parser.DotExpression:
		// Handle obj.field++ or obj.field--
		// Compile the object
		objReg, err := c.Compile(left.Object)
		if err != nil {
			return 0, err
		}

		// Get field name
		nameIdx := c.addConstant(objects.InternString(left.Property.Value))

		// Get current value from field
		valReg := c.allocTempReg()
		c.emitRegGetField(valReg, objReg, nameIdx)

		// Copy old value to result register (postfix returns old value)
		c.emitRegMove(resultReg, valReg)

		// Load constant 1
		oneIdx := c.addConstant(objects.NewInt(1))
		oneReg := c.allocTempReg()
		c.emitRegLoadConst(oneReg, oneIdx)

		// Perform increment or decrement
		switch n.Operator {
		case "++":
			c.emitRegAdd(valReg, valReg, oneReg)
		case "--":
			c.emitRegSub(valReg, valReg, oneReg)
		default:
			return 0, fmt.Errorf("unknown postfix operator: %s", n.Operator)
		}

		// Store back to field
		c.emitRegSetField(objReg, valReg, nameIdx)

		// Free temporary registers
		c.freeTempReg(objReg)
		c.freeTempReg(valReg)
		c.freeTempReg(oneReg)

	default:
		return 0, fmt.Errorf("postfix expression not supported for type: %T", left)
	}

	return resultReg, nil
}

// compileCompoundAssignmentExpression compiles compound assignment expressions (+=, -=, *=, /=)
func (c *RegCompiler) compileCompoundAssignmentExpression(n *parser.CompoundAssignmentExpression) (int, error) {
	// Compile right side first (common for all cases)
	rightReg, err := c.Compile(n.Right)
	if err != nil {
		return 0, err
	}

	// Result register for the final value
	resultReg := c.allocTempReg()

	// Handle different left expression types
	switch left := n.Left.(type) {
	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(left.Value)
		if !ok {
			return 0, fmt.Errorf("undefined variable: %s", left.Value)
		}

		// Load current value
		valReg := c.allocTempReg()
		switch symbol.Scope {
		case GlobalScope:
			c.emitRegLoadGlobal(valReg, symbol.Index)
		case LocalScope:
			c.emitRegLoadLocal(valReg, symbol.Index)
		case FreeScope:
			c.emitRegLoadFree(valReg, symbol.Index)
		}

		// Perform operation
		switch n.Operator {
		case "+=":
			c.emitRegAdd(valReg, valReg, rightReg)
		case "-=":
			c.emitRegSub(valReg, valReg, rightReg)
		case "*=":
			c.emitRegMul(valReg, valReg, rightReg)
		case "/=":
			c.emitRegDiv(valReg, valReg, rightReg)
		case "%=":
			c.emitRegMod(valReg, valReg, rightReg)
		default:
			return 0, fmt.Errorf("unknown compound assignment operator: %s", n.Operator)
		}

		// Copy to result register
		c.emitRegMove(resultReg, valReg)

		// Store back to variable
		switch symbol.Scope {
		case GlobalScope:
			c.emitRegStoreGlobal(valReg, symbol.Index)
		case LocalScope:
			c.emitRegStoreLocal(valReg, symbol.Index)
		case FreeScope:
			c.emitRegStoreFree(valReg, symbol.Index)
		}

		c.freeTempReg(valReg)

	case *parser.IndexExpression:
		// Handle a[i] += n, a[i] -= n, etc.
		objReg, err := c.Compile(left.Left)
		if err != nil {
			return 0, err
		}
		indexReg, err := c.Compile(left.Index)
		if err != nil {
			return 0, err
		}

		// Get current value
		valReg := c.allocTempReg()
		c.emitRegIndex(valReg, objReg, indexReg)

		// Perform operation
		switch n.Operator {
		case "+=":
			c.emitRegAdd(valReg, valReg, rightReg)
		case "-=":
			c.emitRegSub(valReg, valReg, rightReg)
		case "*=":
			c.emitRegMul(valReg, valReg, rightReg)
		case "/=":
			c.emitRegDiv(valReg, valReg, rightReg)
		case "%=":
			c.emitRegMod(valReg, valReg, rightReg)
		default:
			return 0, fmt.Errorf("unknown compound assignment operator: %s", n.Operator)
		}

		// Copy to result register
		c.emitRegMove(resultReg, valReg)

		// Store back
		c.emitRegSetIndex(objReg, indexReg, valReg)

		c.freeTempReg(objReg)
		c.freeTempReg(indexReg)
		c.freeTempReg(valReg)

	case *parser.DotExpression:
		// Handle obj.field += n, obj.field -= n, etc.
		objReg, err := c.Compile(left.Object)
		if err != nil {
			return 0, err
		}
		nameIdx := c.addConstant(objects.InternString(left.Property.Value))

		// Get current value
		valReg := c.allocTempReg()
		c.emitRegGetField(valReg, objReg, nameIdx)

		// Perform operation
		switch n.Operator {
		case "+=":
			c.emitRegAdd(valReg, valReg, rightReg)
		case "-=":
			c.emitRegSub(valReg, valReg, rightReg)
		case "*=":
			c.emitRegMul(valReg, valReg, rightReg)
		case "/=":
			c.emitRegDiv(valReg, valReg, rightReg)
		case "%=":
			c.emitRegMod(valReg, valReg, rightReg)
		default:
			return 0, fmt.Errorf("unknown compound assignment operator: %s", n.Operator)
		}

		// Copy to result register
		c.emitRegMove(resultReg, valReg)

		// Store back
		c.emitRegSetField(objReg, valReg, nameIdx)

		c.freeTempReg(objReg)
		c.freeTempReg(valReg)

	default:
		return 0, fmt.Errorf("compound assignment not supported for type: %T", left)
	}

	c.freeTempReg(rightReg)
	return resultReg, nil
}

// Register allocation helpers

func (c *RegCompiler) allocTempReg() int {
	// First, check if there's a freed register we can reuse
	if len(c.freeRegs) > 0 {
		// Pop the last freed register
		reg := c.freeRegs[len(c.freeRegs)-1]
		c.freeRegs = c.freeRegs[:len(c.freeRegs)-1]
		return reg
	}

	// No freed registers available, allocate a new one
	if c.nextTempReg >= NumRegisters-1 {
		// We've run out of registers - this is a critical error
		// Fall back to reusing a high register (will likely cause issues)
		// In a production compiler, we'd spill to stack/local slots
		return NumRegisters - 2
	}
	reg := c.nextTempReg
	c.nextTempReg++
	if c.nextTempReg > c.maxReg {
		c.maxReg = c.nextTempReg
	}
	return reg
}

// allocContiguousRegisters allocates count contiguous registers starting from the returned index.
// Returns the start register index. If not enough space, uses locals for overflow.
func (c *RegCompiler) allocContiguousRegisters(count int) int {
	if count <= 0 {
		return c.nextTempReg
	}

	// Check if we have enough space
	available := NumRegisters - 1 - c.nextTempReg // -1 for ReturnRegister
	if available >= count {
		// Enough space - allocate normally
		start := c.nextTempReg
		for i := 0; i < count; i++ {
			c.allocTempReg()
		}
		return start
	}

	// Not enough contiguous space - allocate as many as we can
	// The VM will handle reading from locals for overflow
	start := c.nextTempReg
	for c.nextTempReg < NumRegisters-1 {
		c.allocTempReg()
	}
	return start
}

// ensureRegisterSpace ensures there are at least 'count' registers available.
// If not, it returns an error (caller should handle by using alternative approach).
func (c *RegCompiler) ensureRegisterSpace(count int) bool {
	return c.nextTempReg+count < NumRegisters-1
}

func (c *RegCompiler) freeTempReg(reg int) {
	// Validate register is in valid range and not a reserved register
	if reg < FirstLocalRegister || reg >= NumRegisters-1 {
		return // Don't free reserved registers
	}

	// If this is the last allocated register, decrement the counter
	if reg == c.nextTempReg-1 {
		c.nextTempReg--
		// Also pop any freed registers that are now contiguous with the end
		for len(c.freeRegs) > 0 && c.freeRegs[len(c.freeRegs)-1] == c.nextTempReg-1 {
			c.nextTempReg--
			c.freeRegs = c.freeRegs[:len(c.freeRegs)-1]
		}
	} else {
		// Add to free list for reuse
		c.freeRegs = append(c.freeRegs, reg)
	}
}

// Emit helpers

func (c *RegCompiler) emitRegAdd(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegAdd, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegSub(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegSub, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegMul(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegMul, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegDiv(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegDiv, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegMod(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegMod, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegNeg(dst, src int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegNeg, dst, src)...)
}

func (c *RegCompiler) emitRegNot(dst, src int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegNot, dst, src)...)
}

func (c *RegCompiler) emitRegAnd(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegAnd, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegOr(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegOr, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegLess(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegLess, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegGreater(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegGreater, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegLessEqual(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegLessEqual, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegGreaterEqual(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegGreaterEqual, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegEqual(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegEqual, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegNotEqual(dst, src1, src2 int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegNotEqual, dst, src1, src2)...)
}

func (c *RegCompiler) emitRegMove(dst, src int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegMove, dst, src)...)
}

func (c *RegCompiler) emitRegLoadConst(dst, constIdx int) {
	c.instructions = append(c.instructions, MakeRegInstructionConst(OpRegLoadConst, dst, constIdx)...)
}

func (c *RegCompiler) emitRegLoadGlobal(dst, globalIdx int) {
	c.instructions = append(c.instructions, MakeRegInstructionConst(OpRegLoadGlobal, dst, globalIdx)...)
}

func (c *RegCompiler) emitRegStoreGlobal(src, globalIdx int) {
	c.instructions = append(c.instructions, MakeRegInstructionConst(OpRegStoreGlobal, src, globalIdx)...)
}

func (c *RegCompiler) emitRegLoadLocal(dst, localIdx int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegLoadLocal, dst, localIdx)...)
}

func (c *RegCompiler) emitRegStoreLocal(src, localIdx int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegStoreLocal, src, localIdx)...)
}

func (c *RegCompiler) emitRegLoadFree(dst, freeIdx int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegLoadFree, dst, freeIdx)...)
}

func (c *RegCompiler) emitRegStoreFree(src, freeIdx int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegStoreFree, src, freeIdx)...)
}

func (c *RegCompiler) emitRegJump(offset int) int {
	pos := len(c.instructions)
	c.instructions = append(c.instructions, MakeRegJump(OpRegJump, offset)...)
	return pos
}

func (c *RegCompiler) emitRegJumpIfFalse(condReg, offset int) int {
	pos := len(c.instructions)
	c.instructions = append(c.instructions, MakeRegJumpCond(OpRegJumpIfFalse, condReg, offset)...)
	return pos
}

func (c *RegCompiler) emitRegNull(dst int) {
	c.instructions = append(c.instructions, MakeRegInstruction1(OpRegNull, dst)...)
}

func (c *RegCompiler) emitRegTrue(dst int) {
	c.instructions = append(c.instructions, MakeRegInstruction1(OpRegTrue, dst)...)
}

func (c *RegCompiler) emitRegAddConst(dst, src, constIdx int) {
	c.instructions = append(c.instructions,
		byte(OpRegAddConst),
		byte(dst),
		byte(src),
		byte(constIdx>>8),
		byte(constIdx),
	)
}

func (c *RegCompiler) emitRegFalse(dst int) {
	c.instructions = append(c.instructions, MakeRegInstruction1(OpRegFalse, dst)...)
}

func (c *RegCompiler) emitRegCall(funcReg, numArgs int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegCall, funcReg, numArgs)...)
}

func (c *RegCompiler) emitRegTailCall(funcReg, numArgs int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegTailCall, funcReg, numArgs)...)
}

func (c *RegCompiler) emitRegBuiltin(builtinIdx, numArgs int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegBuiltin, builtinIdx, numArgs)...)
}

func (c *RegCompiler) emitRegLoadBuiltin(dst, builtinIdx int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegLoadBuiltin, dst, builtinIdx)...)
}

func (c *RegCompiler) emitRegReturn(reg int) {
	c.instructions = append(c.instructions, MakeRegInstruction1(OpRegReturn, reg)...)
}

func (c *RegCompiler) emitRegArray(dst, startReg, count int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegArray, dst, startReg, count)...)
}

func (c *RegCompiler) emitRegArrayEmpty(dst int) {
	c.instructions = append(c.instructions, MakeRegInstruction1(OpRegArrayEmpty, dst)...)
}

func (c *RegCompiler) emitRegArrayAppend(dst, arrReg, elemReg int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegArrayAppend, dst, arrReg, elemReg)...)
}

func (c *RegCompiler) emitRegMapEmpty(dst int) {
	c.instructions = append(c.instructions, MakeRegInstruction1(OpRegMapEmpty, dst)...)
}

func (c *RegCompiler) emitRegMapSet(dst, mapReg, keyReg, valReg int) {
	c.instructions = append(c.instructions, []byte{byte(OpRegMapSet), byte(dst), byte(mapReg), byte(keyReg), byte(valReg)}...)
}

func (c *RegCompiler) emitRegMap(dst, startReg, count int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegMap, dst, startReg, count)...)
}

func (c *RegCompiler) emitRegIndex(dst, objReg, indexReg int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegIndex, dst, objReg, indexReg)...)
}

func (c *RegCompiler) emitRegSetIndex(objReg, indexReg, valReg int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegSetIndex, objReg, indexReg, valReg)...)
}

func (c *RegCompiler) emitRegSlice(dst, objReg, startReg, endReg int) {
	c.instructions = append(c.instructions, MakeRegInstruction4(OpRegSlice, dst, objReg, startReg, endReg)...)
}

func (c *RegCompiler) emitRegIterKey(dst, iterReg, indexReg int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegIterKey, dst, iterReg, indexReg)...)
}

func (c *RegCompiler) emitRegIterValue(dst, iterReg, indexReg int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegIterValue, dst, iterReg, indexReg)...)
}

func (c *RegCompiler) emitRegLoadModule(dst, constIdx int) {
	c.instructions = append(c.instructions, MakeRegInstructionConst(OpRegLoadModule, dst, constIdx)...)
}

func (c *RegCompiler) emitRegGetExport(dst, moduleReg, nameIdx int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegGetExport, dst, moduleReg, nameIdx)...)
}

func (c *RegCompiler) emitRegSetExport(srcReg, nameIdx int) {
	c.instructions = append(c.instructions, MakeRegInstructionConst(OpRegSetExport, srcReg, nameIdx)...)
}

func (c *RegCompiler) emitRegGetField(dst, objReg, nameIdx int) {
	// Format: OpRegGetField dst obj name_idx_hi name_idx_lo
	c.instructions = append(c.instructions,
		byte(OpRegGetField),
		byte(dst),
		byte(objReg),
		byte(nameIdx>>8),
		byte(nameIdx),
	)
}

func (c *RegCompiler) emitRegSetField(objReg, valReg, nameIdx int) {
	// Format: OpRegSetField obj val name_idx_hi name_idx_lo
	c.instructions = append(c.instructions,
		byte(OpRegSetField),
		byte(objReg),
		byte(valReg),
		byte(nameIdx>>8),
		byte(nameIdx),
	)
}

func (c *RegCompiler) emitRegCallMethod(objReg, nameIdx, numArgs int) {
	// Format: OpRegCallMethod obj name_idx_hi name_idx_lo num_args
	c.instructions = append(c.instructions,
		byte(OpRegCallMethod),
		byte(objReg),
		byte(nameIdx>>8),
		byte(nameIdx),
		byte(numArgs),
	)
}

func (c *RegCompiler) emitRegTailCallMethod(objReg, nameIdx, numArgs int) {
	// Format: OpRegTailCallMethod obj name_idx_hi name_idx_lo num_args
	c.instructions = append(c.instructions,
		byte(OpRegTailCallMethod),
		byte(objReg),
		byte(nameIdx>>8),
		byte(nameIdx),
		byte(numArgs),
	)
}

// patchJump patches a jump instruction with the correct offset
func (c *RegCompiler) patchJump(pos int) {
	offset := len(c.instructions) - pos
	// Jump offset is stored at pos+2, pos+3 (16-bit)
	c.instructions[pos+2] = byte(offset >> 8)
	c.instructions[pos+3] = byte(offset)
}

// patchJumpTo patches a jump to a specific position
func (c *RegCompiler) patchJumpTo(pos int, target int) {
	offset := target - pos
	c.instructions[pos+2] = byte(offset >> 8)
	c.instructions[pos+3] = byte(offset)
}

// addConstant adds a constant to the constant pool
func (c *RegCompiler) addConstant(obj objects.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *RegCompiler) emitRegPush(srcReg int) {
	c.instructions = append(c.instructions, byte(OpRegPush), byte(srcReg))
}

func (c *RegCompiler) emitRegPop(dstReg int) {
	c.instructions = append(c.instructions, byte(OpRegPop), byte(dstReg))
}

// emitRegPrimeInnerLoop emits OpRegPrimeInnerLoop
func (c *RegCompiler) emitRegPrimeInnerLoop(nReg, iReg, resultReg int, jumpDoneOffset int) {
	c.instructions = append(c.instructions,
		byte(OpRegPrimeInnerLoop),
		byte(nReg),
		byte(iReg),
		byte(resultReg),
		byte(jumpDoneOffset>>8),
		byte(jumpDoneOffset),
	)
}

// emitRegModCheckZero emits OpRegModCheckZero
func (c *RegCompiler) emitRegModCheckZero(resultReg, nReg, iReg int) {
	c.instructions = append(c.instructions,
		byte(OpRegModCheckZero),
		byte(resultReg),
		byte(nReg),
		byte(iReg),
	)
}

// emitRegInnerLoopPrime emits OpRegInnerLoopPrime
func (c *RegCompiler) emitRegInnerLoopPrime(nReg, iReg, resultReg, jumpIsPrime, jumpDone int) {
	c.instructions = append(c.instructions,
		byte(OpRegInnerLoopPrime),
		byte(nReg),
		byte(iReg),
		byte(resultReg),
		byte(jumpIsPrime>>8),
		byte(jumpIsPrime),
		byte(jumpDone>>8),
		byte(jumpDone),
	)
}

// emitRegLoopMulCheck emits OpRegLoopMulCheck
func (c *RegCompiler) emitRegLoopMulCheck(iReg, nReg int, jumpOutOffset int) {
	c.instructions = append(c.instructions,
		byte(OpRegLoopMulCheck),
		byte(iReg),
		byte(nReg),
		byte(jumpOutOffset>>8),
		byte(jumpOutOffset),
	)
}

// Bytecode returns the compiled bytecode
func (c *RegCompiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
		SourceMap:    c.sourceMap,
	}
}

// DefineGlobal defines a global variable
func (c *RegCompiler) DefineGlobal(name string) {
	c.symbolTable.Define(name)
}

// ResolveSymbol resolves a symbol
func (c *RegCompiler) ResolveSymbol(name string) (Symbol, bool) {
	return c.symbolTable.Resolve(name)
}

// SetSymbolTable sets the symbol table for persistent compilation
func (c *RegCompiler) SetSymbolTable(st *SymbolTable) {
	c.symbolTable = st
}

// SetConstants sets the constants pool for persistent compilation
func (c *RegCompiler) SetConstants(constants []objects.Object) {
	c.constants = constants
}

// SymbolTable returns the current symbol table
func (c *RegCompiler) SymbolTable() *SymbolTable {
	return c.symbolTable
}

// Constants returns the current constants pool
func (c *RegCompiler) Constants() []objects.Object {
	return c.constants
}

// SetSourceFile sets the source file for error reporting
func (c *RegCompiler) SetSourceFile(filename string) {
	c.sourceFile = filename
}

// compileImportStatement compiles an import statement for register VM
func (c *RegCompiler) compileImportStatement(node *parser.ImportStatement) (int, error) {
	// Allocate a register for the module result
	moduleReg := c.allocTempReg()

	// Load the module path constant
	pathIdx := c.addConstant(objects.InternString(node.Path.Value))

	// Emit OpRegLoadModule to load the module into a register
	c.emitRegLoadModule(moduleReg, pathIdx)

	// Handle different import styles
	if node.Name != nil {
		// Default import: import math from "./math"
		symbol := c.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(moduleReg, symbol.Index)
		} else {
			c.emitRegStoreLocal(moduleReg, symbol.Index)
		}
	} else if node.Alias != nil {
		// Namespace import: import * as math from "./math"
		symbol := c.symbolTable.Define(node.Alias.Value)
		if symbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(moduleReg, symbol.Index)
		} else {
			c.emitRegStoreLocal(moduleReg, symbol.Index)
		}
	} else if len(node.Names) > 0 {
		// Destructuring import: import { add, sub } from "./math"
		for _, name := range node.Names {
			// Get the export by name
			nameIdx := c.addConstant(objects.InternString(name.Value))
			exportReg := c.allocTempReg()
			c.emitRegGetExport(exportReg, moduleReg, nameIdx)
			// Store in global
			symbol := c.symbolTable.Define(name.Value)
			if symbol.Scope == GlobalScope {
				c.emitRegStoreGlobal(exportReg, symbol.Index)
			} else {
				c.emitRegStoreLocal(exportReg, symbol.Index)
			}
			c.freeTempReg(exportReg)
		}
	} else {
		// Simple import: import "time" or import "./math"
		path := node.Path.Value
		moduleName := extractModuleName(path)

		if moduleName != "" {
			// Store the module as a global variable
			symbol := c.symbolTable.Define(moduleName)
			if symbol.Scope == GlobalScope {
				c.emitRegStoreGlobal(moduleReg, symbol.Index)
			} else {
				c.emitRegStoreLocal(moduleReg, symbol.Index)
			}
		}
	}

	c.freeTempReg(moduleReg)
	return 0, nil
}

// compileExportStatement compiles an export statement for register VM
func (c *RegCompiler) compileExportStatement(node *parser.ExportStatement) (int, error) {
	// Handle different export types
	switch stmt := node.Exportable.(type) {
	case *parser.VarStatement:
		// Compile the value expression
		valReg, err := c.Compile(stmt.Value)
		if err != nil {
			return 0, err
		}
		// Define the variable in the symbol table
		symbol := c.symbolTable.Define(stmt.Name.Value)
		// Store in global
		if symbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(valReg, symbol.Index)
		} else {
			c.emitRegStoreLocal(valReg, symbol.Index)
		}
		// Export the variable
		nameIdx := c.addConstant(objects.InternString(stmt.Name.Value))
		c.emitRegSetExport(valReg, nameIdx)
		c.freeTempReg(valReg)

	case *parser.ConstStatement:
		// Compile the value expression
		valReg, err := c.Compile(stmt.Value)
		if err != nil {
			return 0, err
		}
		// Define the constant in the symbol table
		symbol := c.symbolTable.Define(stmt.Name.Value)
		// Store in global
		if symbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(valReg, symbol.Index)
		} else {
			c.emitRegStoreLocal(valReg, symbol.Index)
		}
		// Export the constant
		nameIdx := c.addConstant(objects.InternString(stmt.Name.Value))
		c.emitRegSetExport(valReg, nameIdx)
		c.freeTempReg(valReg)

	case *parser.ExpressionStatement:
		// Handle function exports: export func add(a, b) { ... }
		if fn, ok := stmt.Expression.(*parser.FunctionLiteral); ok {
			if fn.Name == "" {
				return 0, fmt.Errorf("exported function must have a name")
			}
			// Compile the function
			valReg, err := c.Compile(fn)
			if err != nil {
				return 0, err
			}
			// The function is now stored in the global
			symbol, ok := c.symbolTable.Resolve(fn.Name)
			if !ok {
				return 0, fmt.Errorf("symbol %s not found", fn.Name)
			}
			if symbol.Scope == GlobalScope {
				c.emitRegStoreGlobal(valReg, symbol.Index)
			} else {
				c.emitRegStoreLocal(valReg, symbol.Index)
			}
			// Export the function
			nameIdx := c.addConstant(objects.InternString(fn.Name))
			c.emitRegSetExport(valReg, nameIdx)
			c.freeTempReg(valReg)
			return 0, nil
		}
		return 0, fmt.Errorf("unsupported export expression type: %T", stmt.Expression)

	default:
		return 0, fmt.Errorf("unsupported export type: %T", stmt)
	}

	return 0, nil
}

// CompileReg compiles a program using the register-based compiler
func CompileReg(program *parser.Program) (*Bytecode, error) {
	c := NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		return nil, err
	}
	return c.Bytecode(), nil
}

// ============================================================================
// PRIME CHECK OPTIMIZATION
// ============================================================================

// tryOptimizePrimeCheckFunc attempts to detect a complete prime check function
// Pattern: func isPrime(n) { var i = 2; while (i * i <= n) { if (n % i == 0) { return false } i++ } return true }
// Returns true if optimization was applied
func (c *RegCompiler) tryOptimizePrimeCheckFunc(fn *parser.FunctionLiteral) bool {
	// Check if function has exactly one parameter
	if len(fn.Parameters) != 1 {
		return false
	}

	// Check body structure: should have variable declaration, while loop, and return
	body := fn.Body
	if len(body.Statements) < 2 {
		return false
	}

	// Try to match the pattern
	// We're looking for a simple prime check implementation
	// For now, let's check for a simpler pattern: just a return with a single expression

	// Actually, let's handle this at the function call level instead
	// by detecting when isPrime(n) is called with a constant or variable

	return false
}

// emitRegPrimeCheck emits OpRegPrimeCheck
func (c *RegCompiler) emitRegPrimeCheck(nReg, resultReg int) {
	c.instructions = append(c.instructions,
		byte(OpRegPrimeCheck),
		byte(nReg),
		byte(resultReg),
	)
}

// emitRegPrimeCheckRange emits OpRegPrimeCheckRange
func (c *RegCompiler) emitRegPrimeCheckRange(startReg, endReg, countReg int) {
	c.instructions = append(c.instructions,
		byte(OpRegPrimeCheckRange),
		byte(startReg),
		byte(endReg),
		byte(countReg),
	)
}

// emitRegNestedLoopMul emits OpRegNestedLoopMul
func (c *RegCompiler) emitRegNestedLoopMul(arrAReg, arrBReg, nConst, mConst, resultReg int) {
	c.instructions = append(c.instructions,
		byte(OpRegNestedLoopMul),
		byte(arrAReg),
		byte(arrBReg),
		byte(nConst>>8), byte(nConst),
		byte(mConst>>8), byte(mConst),
		byte(resultReg),
	)
}

// emitRegMatrixMulElement emits OpRegMatrixMulElement
func (c *RegCompiler) emitRegMatrixMulElement(aReg, bReg, iReg, jReg, kLimit int, resultReg int) {
	c.instructions = append(c.instructions,
		byte(OpRegMatrixMulElement),
		byte(aReg),
		byte(bReg),
		byte(iReg),
		byte(jReg),
		byte(kLimit>>8), byte(kLimit),
		byte(resultReg),
	)
}

// ============================================================================
// LOOP UNROLLING OPTIMIZATION
// ============================================================================

// tryUnrollLoop attempts to unroll small fixed-iteration loops
// Pattern: for (var i = 0; i < N; i++) { body } where N is a small constant
// Returns true if optimization was applied
func (c *RegCompiler) tryUnrollLoop(n *parser.ForStatement) bool {
	// Check for pattern: for (var i = 0; i < limit; i++) { body }
	// where limit is a small integer constant (<= 8)

	if n.Init == nil || n.Condition == nil || n.Update == nil || n.Body == nil {
		return false
	}

	// Check init: var i = 0
	varInit, ok := n.Init.(*parser.VarStatement)
	if !ok {
		return false
	}
	initInt, ok := varInit.Value.(*parser.IntegerLiteral)
	if !ok || initInt.Value != 0 {
		return false
	}

	// Check condition: i < limit
	condInfix, ok := n.Condition.(*parser.InfixExpression)
	if !ok || condInfix.Operator != "<" {
		return false
	}
	condLeft, ok := condInfix.Left.(*parser.Identifier)
	if !ok || condLeft.Value != varInit.Name.Value {
		return false
	}
	limitInt, ok := condInfix.Right.(*parser.IntegerLiteral)
	if !ok {
		return false
	}

	limit := int(limitInt.Value)

	// Only unroll small loops (limit <= 8)
	if limit > 8 || limit < 1 {
		return false
	}

	// Check update: i++ or i += 1
	isIncrement := false
	switch update := n.Update.(type) {
	case *parser.ExpressionStatement:
		if postExpr, ok := update.Expression.(*parser.PostfixExpression); ok {
			if postExpr.Operator == "++" {
				ident, ok := postExpr.Left.(*parser.Identifier)
				if ok && ident.Value == varInit.Name.Value {
					isIncrement = true
				}
			}
		}
		if compoundExpr, ok := update.Expression.(*parser.CompoundAssignmentExpression); ok {
			if compoundExpr.Operator == "+=" {
				ident, ok := compoundExpr.Left.(*parser.Identifier)
				if ok && ident.Value == varInit.Name.Value {
					if rightInt, ok := compoundExpr.Right.(*parser.IntegerLiteral); ok && rightInt.Value == 1 {
						isIncrement = true
					}
				}
			}
		}
	}

	if !isIncrement {
		return false
	}

	// Pattern matched! Unroll the loop by duplicating the body
	// Define the loop variable
	loopVarSymbol := c.symbolTable.Define(varInit.Name.Value)

	// Allocate a register for the loop variable
	loopVarReg := c.allocTempReg()

	// Unroll the loop - emit the body 'limit' times
	// For each iteration, we set the loop variable to the current value
	// and then compile the body
	for i := 0; i < limit; i++ {
		// Set loop variable to current iteration value
		iterValueIdx := c.addConstant(objects.NewInt(int64(i)))
		c.emitRegLoadConst(loopVarReg, iterValueIdx)

		// Store to local slot
		if loopVarSymbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(loopVarReg, loopVarSymbol.Index)
		} else {
			c.emitRegStoreLocal(loopVarReg, loopVarSymbol.Index)
		}

		// Compile the body
		// We need to compile the body for each iteration
		// This is done by recursively compiling each statement
		for _, stmt := range n.Body.Statements {
			_, err := c.Compile(stmt)
			if err != nil {
				c.freeTempReg(loopVarReg)
				return false
			}
		}
	}

	// Set the final value of the loop variable to 'limit'
	finalValueIdx := c.addConstant(objects.NewInt(int64(limit)))
	c.emitRegLoadConst(loopVarReg, finalValueIdx)
	if loopVarSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(loopVarReg, loopVarSymbol.Index)
	} else {
		c.emitRegStoreLocal(loopVarReg, loopVarSymbol.Index)
	}

	c.freeTempReg(loopVarReg)

	return true
}

// compileClassStatement compiles a class declaration
func (c *RegCompiler) compileClassStatement(node *parser.ClassStatement) (int, error) {
	// Define the class name in the symbol table BEFORE compiling methods
	// This allows methods to reference the class name (e.g., for constructor calls)
	classSymbol := c.symbolTable.Define(node.Name.Value)

	// Remember the superclass symbol (if any) to load later
	var superSymbol *Symbol
	if node.SuperClass != nil {
		symbol, ok := c.symbolTable.Resolve(node.SuperClass.Value)
		if !ok {
			return 0, fmt.Errorf("undefined superclass: %s", node.SuperClass.Value)
		}
		superSymbol = &symbol
	}

	// Compile default fields as a map
	// Format: key1, val1, key2, val2, ... -> map
	// We need contiguous registers for OpRegMap, so allocate directly without using free list
	numFields := len(node.Fields)
	fieldsStartReg := c.nextTempReg

	// Reserve registers for all field key-value pairs
	numFieldRegs := numFields * 2
	for i := 0; i < numFieldRegs; i++ {
		c.nextTempReg++
		if c.nextTempReg > c.maxReg {
			c.maxReg = c.nextTempReg
		}
	}

	// Now compile each field's key and value into the reserved registers
	for i, field := range node.Fields {
		keyReg := fieldsStartReg + i*2
		valReg := fieldsStartReg + i*2 + 1

		// Key (field name)
		nameIdx := c.addConstant(objects.NewString(field.Name.Value))
		c.emitRegLoadConst(keyReg, nameIdx)

		// Value (field default value)
		// Compile the value and move it to the target register
		compiledValReg, err := c.Compile(field.Value)
		if err != nil {
			return 0, err
		}
		if compiledValReg != valReg {
			c.emitRegMove(valReg, compiledValReg)
		}
	}

	// Create fields map
	fieldsReg := c.allocTempReg()
	if numFields == 0 {
		c.emitRegMapEmpty(fieldsReg)
	} else {
		c.instructions = append(c.instructions,
			byte(OpRegMap),
			byte(fieldsReg),
			byte(fieldsStartReg),
			byte(numFields*2),
		)
		// Free field key/value registers
		for i := 0; i < numFields*2; i++ {
			c.freeTempReg(fieldsStartReg + i)
		}
	}

	// Compile methods as a map
	// Each method needs 'this' at local slot 0
	// First, compile all methods and store them (to avoid scope issues)
	type methodInfo struct {
		name     string
		fn       *CompiledFunction
		fnIndex  int
		freeVars []Symbol
	}
	methodInfos := make([]methodInfo, 0, len(node.Methods))

	for _, method := range node.Methods {
		// Compile method as function with 'this' at local 0
		methodFn, err := c.compileMethod(method)
		if err != nil {
			return 0, err
		}
		fnIndex := c.addConstant(methodFn)
		methodInfos = append(methodInfos, methodInfo{
			name:     method.Name,
			fn:       methodFn,
			fnIndex:  fnIndex,
			freeVars: methodFn.FreeVariables,
		})
	}

	// Now build the methods map in contiguous registers
	// We need to ensure contiguous allocation, so clear the freeRegs first
	// and allocate all registers sequentially
	numMethodRegs := len(methodInfos) * 2 // Each method needs key and value registers
	methodsStartReg := c.nextTempReg

	// Ensure we allocate contiguous registers by not using the free list
	for i := 0; i < numMethodRegs; i++ {
		c.nextTempReg++
		if c.nextTempReg > c.maxReg {
			c.maxReg = c.nextTempReg
		}
	}

	// Now emit the load instructions for each method
	for i, mi := range methodInfos {
		keyReg := methodsStartReg + i*2
		valReg := methodsStartReg + i*2 + 1

		// Key (method name)
		nameIdx := c.addConstant(objects.NewString(mi.name))
		c.emitRegLoadConst(keyReg, nameIdx)

		// Load the method function
		if len(mi.freeVars) == 0 {
			c.emitRegLoadConst(valReg, mi.fnIndex)
		} else {
			// Handle closures - load free variables
			// Allocate registers for free vars (these don't need to be contiguous with the map)
			freeStartReg := c.nextTempReg
			for j, freeVar := range mi.freeVars {
				c.nextTempReg++
				if c.nextTempReg > c.maxReg {
					c.maxReg = c.nextTempReg
				}
				freeReg := freeStartReg + j
				switch freeVar.Scope {
				case GlobalScope:
					c.emitRegLoadGlobal(freeReg, freeVar.Index)
				case LocalScope:
					c.emitRegLoadLocal(freeReg, freeVar.Index)
				case FreeScope:
					c.emitRegLoadFree(freeReg, freeVar.Index)
				}
			}
			c.instructions = append(c.instructions,
				byte(OpRegClosure),
				byte(valReg),
				byte(mi.fnIndex>>8),
				byte(mi.fnIndex),
				byte(len(mi.freeVars)),
				byte(freeStartReg),
			)
		}
	}
	numMethods := len(node.Methods)

	// Create methods map
	methodsReg := c.allocTempReg()
	if numMethods == 0 {
		c.emitRegMapEmpty(methodsReg)
	} else {
		c.instructions = append(c.instructions,
			byte(OpRegMap),
			byte(methodsReg),
			byte(methodsStartReg),
			byte(numMethods*2),
		)
		// Free method key/value registers
		for i := 0; i < numMethods*2; i++ {
			c.freeTempReg(methodsStartReg + i)
		}
	}

	// Load superclass NOW (just before creating the class) to avoid register conflicts
	superclassReg := 255 // 255 means no superclass
	if superSymbol != nil {
		superclassReg = c.allocTempReg()
		c.emitRegLoadGlobal(superclassReg, superSymbol.Index)
	}

	// Create class
	nameIdx := c.addConstant(objects.NewString(node.Name.Value))
	dst := c.allocTempReg()
	c.instructions = append(c.instructions,
		byte(OpRegClass),
		byte(dst),
		byte(nameIdx>>8),
		byte(nameIdx),
		byte(superclassReg),
		byte(fieldsReg),
		byte(methodsReg),
	)

	// Free temporary registers
	if superclassReg != 255 {
		c.freeTempReg(superclassReg)
	}
	c.freeTempReg(fieldsReg)
	c.freeTempReg(methodsReg)

	// Store class in global (use the symbol we defined at the start)
	c.emitRegStoreGlobal(dst, classSymbol.Index)

	return dst, nil
}

// compileMethod compiles a method (function with 'this' at local 0)
func (c *RegCompiler) compileMethod(n *parser.FunctionLiteral) (*CompiledFunction, error) {
	// Enter function scope
	c.enterScope()

	// Reserve local slot 0 for 'this' (set by VM when calling method)
	thisSymbol := c.symbolTable.Define("this")

	// Verify that 'this' is at index 0 (it should be the first defined variable)
	if thisSymbol.Index != 0 {
		return nil, fmt.Errorf("internal error: 'this' symbol not at index 0")
	}

	// Copy 'this' from R0 to local slot 0
	// This is critical - R0 contains 'this' when the method is called
	c.emitRegStoreLocal(0, thisSymbol.Index)

	// Define parameters as local variables (starting at index 1)
	// Parameters are passed in R1-R7 (R0 is 'this'), copy them to local slots
	for i, p := range n.Parameters {
		symbol := c.symbolTable.Define(p.Value)
		// Copy from argument register to local slot
		// R1-R7 -> Locals[1..n] (slot 0 is 'this')
		c.emitRegStoreLocal(i+1, symbol.Index)
	}

	// Compile body
	lastReg, err := c.Compile(n.Body)
	if err != nil {
		return nil, err
	}

	// Ensure function ends with return
	if len(c.instructions) == 0 || Opcode(c.instructions[len(c.instructions)-1]) != OpRegReturn {
		// Emit implicit return with the last expression's result
		// For methods, the implicit return should return the last expression's value
		if lastReg == 0 {
			// Empty block or statement with no value - return null
			c.emitRegReturn(0)
		} else {
			// Return the last expression's result
			c.emitRegReturn(lastReg)
		}
	}

	// Leave scope and get compiled function
	fn := c.leaveScope()
	fn.NumParameters = len(n.Parameters) + 1 // +1 for 'this'

	return fn, nil
}

// compileNewExpression compiles a new expression
func (c *RegCompiler) compileNewExpression(node *parser.NewExpression) (int, error) {
	// Get class
	symbol, ok := c.symbolTable.Resolve(node.Class.String())
	if !ok {
		return 0, fmt.Errorf("undefined class: %s", node.Class.String())
	}

	classReg := c.allocTempReg()
	c.emitRegLoadGlobal(classReg, symbol.Index)

	// For simplicity, we limit to 8 arguments in registers
	if len(node.Arguments) > 8 {
		c.freeTempReg(classReg)
		return 0, fmt.Errorf("too many arguments for new expression (max 8)")
	}

	// Compile arguments and move them to R0, R1, R2, ...
	// This is required because OpRegNew expects args in R0-R7
	// and will shift them to R1-R8 when adding 'this' to R0
	argRegs := make([]int, len(node.Arguments))
	for i, arg := range node.Arguments {
		reg, err := c.Compile(arg)
		if err != nil {
			c.freeTempReg(classReg)
			return 0, err
		}
		argRegs[i] = reg

		// Move argument to R0, R1, R2, ...
		// Use argument registers (0-7) for passing args to constructor
		if reg != i {
			c.emitRegMove(i, reg)
			c.freeTempReg(reg)
		}
	}

	// Create instance - OpRegNew format: dst, class_reg, num_args
	// Args are now in R0, R1, R2, ...
	dst := c.allocTempReg()
	c.instructions = append(c.instructions, byte(OpRegNew), byte(dst), byte(classReg), byte(len(node.Arguments)))

	// Free temporary registers
	c.freeTempReg(classReg)

	return dst, nil
}

// compileThisExpression compiles this expression
func (c *RegCompiler) compileThisExpression(node *parser.ThisExpression) (int, error) {
	// Load 'this' from local 0 (this is always first local in method context)
	dst := c.allocTempReg()
	c.emitRegLoadLocal(dst, 0)
	return dst, nil
}

// compileSuperCallExpression compiles a super.method() call
func (c *RegCompiler) compileSuperCallExpression(node *parser.SuperCallExpression) (int, error) {
	// Get method name
	methodIdx := c.addConstant(objects.NewString(node.Method))

	// Compile arguments
	// The VM will:
	// 1. Get 'this' from local 0
	// 2. Shift args: R0->R1, R1->R2, etc.
	// 3. Put 'this' in R0
	// So we need to put args in R0, R1, R2, ... (they will become R1, R2, R3 after shift)
	if len(node.Args) > 7 {
		return 0, fmt.Errorf("too many arguments for super call (max 7)")
	}

	// Compile arguments to temp registers first
	argRegs := make([]int, len(node.Args))
	for i, arg := range node.Args {
		argReg, err := c.Compile(arg)
		if err != nil {
			return 0, err
		}
		argRegs[i] = argReg
	}

	// Move arguments to R0, R1, R2, ...
	for i, argReg := range argRegs {
		if argReg != i {
			c.emitRegMove(i, argReg)
		}
	}

	// Free temp registers
	for _, reg := range argRegs {
		c.freeTempReg(reg)
	}

	// Emit super call instruction
	// Format: OpRegSuper, method_idx(2), num_args
	c.instructions = append(c.instructions,
		byte(OpRegSuper),
		byte(methodIdx>>8),
		byte(methodIdx),
		byte(len(node.Args)),
	)

	// Result is in ReturnRegister
	return ReturnRegister, nil
}

// compileTryStatement compiles a try-catch-finally statement
func (c *RegCompiler) compileTryStatement(node *parser.TryStatement) (int, error) {
	// Increment try block depth to disable tail call optimization
	c.tryBlockDepth++

	// Push exception handler with placeholder addresses (2 bytes each)
	pushHandlerPos := len(c.instructions)
	c.instructions = append(c.instructions, byte(OpRegPushHandler), 0, 0, 0, 0) // catchAddr (2 bytes), finallyAddr (2 bytes)

	// Compile try block
	_, err := c.Compile(node.Block)
	if err != nil {
		c.tryBlockDepth--
		return 0, err
	}

	// Pop handler after successful try block
	c.instructions = append(c.instructions, byte(OpRegPopHandler))

	// After try block completes normally, jump to finally (if exists) or past catch
	var jumpPastCatchPos int = -1
	if node.Catch != nil || node.Finally != nil {
		jumpPastCatchPos = c.emitRegJump(0) // Will be patched
	}

	// Record catch address
	catchAddr := 0
	if node.Catch != nil {
		catchAddr = len(c.instructions)

		// The exception value is in R0 - bind it to the variable
		// VM puts the thrown value in R0 before jumping to catch
		symbol := c.symbolTable.Define(node.Catch.Exception.Value)
		// Exception value is in R0, store it to the variable
		if symbol.Scope == GlobalScope {
			c.emitRegStoreGlobal(0, symbol.Index) // R0 -> global
		} else {
			c.emitRegStoreLocal(0, symbol.Index) // R0 -> local
		}

		// Compile catch body
		_, err = c.Compile(node.Catch.Block)
		if err != nil {
			c.tryBlockDepth--
			return 0, err
		}
	}

	// Record finally address
	finallyAddr := 0
	if node.Finally != nil {
		finallyAddr = len(c.instructions)

		// Compile finally block
		_, err = c.Compile(node.Finally.Block)
		if err != nil {
			c.tryBlockDepth--
			return 0, err
		}

		// Emit OpRegEndFinally to check for pending exceptions
		c.instructions = append(c.instructions, byte(OpRegEndFinally))
	}

	// Patch push handler with catch and finally addresses
	// Addresses are 2 bytes each (big-endian)
	c.instructions[pushHandlerPos+1] = byte(catchAddr >> 8)
	c.instructions[pushHandlerPos+2] = byte(catchAddr)
	c.instructions[pushHandlerPos+3] = byte(finallyAddr >> 8)
	c.instructions[pushHandlerPos+4] = byte(finallyAddr)

	// Patch jump after try to land at finally (if exists) or after catch
	if jumpPastCatchPos >= 0 {
		if node.Finally != nil && finallyAddr > 0 {
			// Jump to finally block to execute it
			c.patchJumpTo(jumpPastCatchPos, finallyAddr)
		} else {
			// Jump past catch
			c.patchJump(jumpPastCatchPos)
		}
	}

	// Decrement try block depth
	c.tryBlockDepth--

	// Return null for try statement
	dst := c.allocTempReg()
	c.emitRegNull(dst)
	return dst, nil
}

// compileThrowStatement compiles a throw statement
func (c *RegCompiler) compileThrowStatement(node *parser.ThrowStatement) (int, error) {
	// Compile the expression to throw (if present)
	var errReg int
	if node.ErrExpr != nil {
		var err error
		errReg, err = c.Compile(node.ErrExpr)
		if err != nil {
			return 0, err
		}
	} else {
		// Throw null if no expression
		errReg = c.allocTempReg()
		c.emitRegNull(errReg)
	}

	c.instructions = append(c.instructions, byte(OpRegThrow), byte(errReg))
	c.freeTempReg(errReg)

	// Return null (unreachable, but needed for type consistency)
	dst := c.allocTempReg()
	c.emitRegNull(dst)
	return dst, nil
}

// compileSwitchStatement compiles a switch statement
func (c *RegCompiler) compileSwitchStatement(node *parser.SwitchStatement) (int, error) {
	// Compile switch expression
	switchReg, err := c.Compile(node.Expression)
	if err != nil {
		return 0, err
	}

	// Track positions for patching jumps to end
	var endJumpPositions []int

	// Compile each case
	for _, caseStmt := range node.Cases {
		// Compile case expression
		caseReg, err := c.Compile(caseStmt.Expression)
		if err != nil {
			return 0, err
		}

		// Compare: switchValue == caseValue
		cmpReg := c.allocTempReg()
		c.emitRegEqual(cmpReg, switchReg, caseReg)
		c.freeTempReg(caseReg)

		// Jump to next case/default if not matched
		notMatchedJumpPos := c.emitRegJumpIfFalse(cmpReg, 0)
		c.freeTempReg(cmpReg)

		// Match found - compile case body
		_, err = c.Compile(caseStmt.Consequence)
		if err != nil {
			return 0, err
		}

		// Jump to end of switch (if we fell through without break)
		endJumpPos := c.emitRegJump(0)
		endJumpPositions = append(endJumpPositions, endJumpPos)

		// Patch "not matched" jump to next case
		c.patchJump(notMatchedJumpPos)
	}

	// Compile default case if exists
	if node.Default != nil {
		_, err = c.Compile(node.Default.Consequence)
		if err != nil {
			return 0, err
		}
	}

	// Patch all "jump to end" positions
	endPos := len(c.instructions)
	for _, pos := range endJumpPositions {
		c.patchJumpTo(pos, endPos)
	}

	// Free switch register
	c.freeTempReg(switchReg)

	// Return null for switch statement
	resultReg := c.allocTempReg()
	c.emitRegNull(resultReg)
	return resultReg, nil
}
