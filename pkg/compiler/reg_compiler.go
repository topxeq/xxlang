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
		// Builtins are handled at call time
		c.emitRegLoadConst(dst, symbol.Index) // Builtin index
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

	// Patch breaks
	ctx := c.loopContexts[len(c.loopContexts)-1]
	for _, pos := range ctx.breakPos {
		c.patchJump(pos)
	}
	c.loopContexts = c.loopContexts[:len(c.loopContexts)-1]

	return 0, nil
}

// compileForStatement compiles a for statement
func (c *RegCompiler) compileForStatement(n *parser.ForStatement) (int, error) {
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

// compileForInStatement compiles a for-in statement
// for (value in iterable) { body }
// for (key, value in iterable) { body }
func (c *RegCompiler) compileForInStatement(n *parser.ForInStatement) (int, error) {
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
	iterSymbol := c.symbolTable.Define("__for_in_iter__")
	if iterSymbol.Scope == GlobalScope {
		c.emitRegStoreGlobal(iterReg, iterSymbol.Index)
	} else {
		c.emitRegStoreLocal(iterReg, iterSymbol.Index)
	}
	c.freeTempReg(iterReg)

	// Define index variable (hidden, used for iteration)
	indexSymbol := c.symbolTable.Define("__for_in_index__")

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
	condReg := c.allocTempReg()
	c.emitRegLess(condReg, indexReg, lenReg)

	// Jump to end if condition is false
	jumpIfFalsePos := c.emitRegJumpIfFalse(condReg, 0)
	c.freeTempReg(condReg)
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
	if _, ok := n.Function.(*parser.DotExpression); ok {
		// Method calls need special handling for TCO
		// For now, fall back to normal call + return
		// TODO: Implement method TCO
		valReg, err := c.compileCallExpression(n)
		if err != nil {
			return 0, err
		}
		c.emitRegReturn(valReg)
		c.freeTempReg(valReg)
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

	// Compile body
	_, err := c.Compile(n.Body)
	if err != nil {
		return 0, err
	}

	// Ensure function ends with return
	if len(c.instructions) == 0 || Opcode(c.instructions[len(c.instructions)-1]) != OpRegReturn {
		// Emit implicit return null
		c.emitRegReturn(0)
	}

	// Leave scope and get compiled function
	fn := c.leaveScope()
	fn.NumParameters = len(n.Parameters)

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
		dst := c.allocTempReg()
		startReg := c.nextTempReg

		for i, elemReg := range elementRegs {
			targetReg := startReg + i
			c.allocTempReg() // allocate the register slot
			if targetReg != elemReg {
				c.emitRegMove(targetReg, elemReg)
				c.freeTempReg(elemReg)
			}
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
	// Collect all key-value pairs first to ensure deterministic order
	type kvPair struct {
		keyReg, valReg int
	}
	pairs := make([]kvPair, 0, len(n.Pairs))

	// Compile all key-value pairs
	for key, val := range n.Pairs {
		keyReg, err := c.Compile(key)
		if err != nil {
			return 0, err
		}
		// Save keyReg by moving to a safe position if needed
		savedKeyReg := c.allocTempReg()
		if savedKeyReg != keyReg {
			c.emitRegMove(savedKeyReg, keyReg)
			c.freeTempReg(keyReg)
		}

		valReg, err := c.Compile(val)
		if err != nil {
			return 0, err
		}

		pairs = append(pairs, kvPair{keyReg: savedKeyReg, valReg: valReg})
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
	// Handle identifier postfix expressions: x++ or x--
	switch left := n.Left.(type) {
	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(left.Value)
		if !ok {
			return 0, fmt.Errorf("undefined variable: %s", left.Value)
		}

		// Allocate a register for the result (old value)
		resultReg := c.allocTempReg()

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

		return resultReg, nil

	default:
		return 0, fmt.Errorf("postfix expression not supported for type: %T", left)
	}
}

// compileCompoundAssignmentExpression compiles compound assignment expressions (+=, -=, *=, /=)
func (c *RegCompiler) compileCompoundAssignmentExpression(n *parser.CompoundAssignmentExpression) (int, error) {
	// Handle identifier compound assignments: x += 1, x -= 2, etc.
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

		// Compile right side
		rightReg, err := c.Compile(n.Right)
		if err != nil {
			return 0, err
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
		c.freeTempReg(rightReg)

		return valReg, nil

	default:
		return 0, fmt.Errorf("compound assignment not supported for type: %T", left)
	}
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
