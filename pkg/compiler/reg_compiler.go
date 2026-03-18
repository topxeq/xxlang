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
	nextTempReg int
	maxReg      int

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
	})

	// Start with fresh instructions for the new scope
	c.instructions = []byte{}

	// Reset register allocation for new scope (parameters will use R0-R7)
	c.nextTempReg = FirstLocalRegister
	c.maxReg = FirstLocalRegister

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// leaveScope leaves the current scope and returns the compiled function
func (c *RegCompiler) leaveScope() *CompiledFunction {
	// Capture the function's instructions
	fnInstructions := c.instructions
	numLocals := c.symbolTable.NumDefinitions
	freeVars := make([]Symbol, len(c.symbolTable.FreeSymbols))
	copy(freeVars, c.symbolTable.FreeSymbols)

	// Restore outer scope's state
	if len(c.scopeStack) > 0 {
		outer := c.scopeStack[len(c.scopeStack)-1]
		c.scopeStack = c.scopeStack[:len(c.scopeStack)-1]
		c.instructions = outer.instructions
		c.nextTempReg = outer.nextTempReg
		c.maxReg = outer.maxReg
	} else {
		c.instructions = []byte{}
		c.nextTempReg = FirstLocalRegister
		c.maxReg = FirstLocalRegister
	}

	c.symbolTable = c.symbolTable.Outer

	return &CompiledFunction{
		Instructions:  fnInstructions,
		NumLocals:     numLocals,
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
	case *parser.BlockStatement:
		return c.compileBlockStatement(n)
	case *parser.IfStatement:
		return c.compileIfStatement(n)
	case *parser.WhileStatement:
		return c.compileWhileStatement(n)
	case *parser.ForStatement:
		return c.compileForStatement(n)
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
	case *parser.AssignmentExpression:
		return c.compileAssignmentExpression(n)
	case *parser.BreakStatement:
		return c.compileBreakStatement(n)
	case *parser.ContinueStatement:
		return c.compileContinueStatement(n)
	case *parser.TernaryExpression:
		return c.compileTernaryExpression(n)
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
	constIdx := c.addConstant(&objects.String{Value: n.Value})
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

// compileReturnStatement compiles a return statement
func (c *RegCompiler) compileReturnStatement(n *parser.ReturnStatement) (int, error) {
	if n.ReturnValue == nil {
		c.emitRegReturn(0)
		return 0, nil
	}

	valReg, err := c.Compile(n.ReturnValue)
	if err != nil {
		return 0, err
	}

	c.emitRegReturn(valReg)
	c.freeTempReg(valReg)
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
			// Compile arguments into R0-R7
			for i, arg := range n.Arguments {
				argReg, err := c.Compile(arg)
				if err != nil {
					return 0, err
				}
				// Move to argument register position (R0-R7)
				if argReg != i {
					c.emitRegMove(i, argReg)
					c.freeTempReg(argReg)
				}
				// Don't free if argReg == i, the value is already in the right place
			}

			// Emit builtin call
			c.emitRegBuiltin(symbol.Index, len(n.Arguments))
			return ReturnRegister, nil
		}
	}

	// Regular function call
	// Compile function
	funcReg, err := c.Compile(n.Function)
	if err != nil {
		return 0, err
	}

	// Compile arguments into R0-R7
	for i, arg := range n.Arguments {
		argReg, err := c.Compile(arg)
		if err != nil {
			return 0, err
		}
		// Move to argument register position (R0-R7)
		if argReg != i {
			c.emitRegMove(i, argReg)
		}
		c.freeTempReg(argReg)
	}

	// Emit call
	c.emitRegCall(funcReg, len(n.Arguments))
	c.freeTempReg(funcReg)

	// Result is in ReturnRegister
	return ReturnRegister, nil
}

// compileArrayLiteral compiles an array literal
func (c *RegCompiler) compileArrayLiteral(n *parser.ArrayLiteral) (int, error) {
	// Compile elements
	startReg := c.nextTempReg
	for _, elem := range n.Elements {
		_, err := c.Compile(elem)
		if err != nil {
			return 0, err
		}
	}

	dst := c.allocTempReg()
	c.emitRegArray(dst, startReg, len(n.Elements))

	// Free element registers
	for i := 0; i < len(n.Elements); i++ {
		c.freeTempReg(startReg + i)
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

	// Now allocate contiguous registers and move all pairs
	dst := c.allocTempReg()
	startReg := c.nextTempReg
	count := len(pairs)

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

// Register allocation helpers

func (c *RegCompiler) allocTempReg() int {
	if c.nextTempReg >= NumRegisters-1 {
		// Reserve ReturnRegister
		return NumRegisters - 2
	}
	reg := c.nextTempReg
	c.nextTempReg++
	if c.nextTempReg > c.maxReg {
		c.maxReg = c.nextTempReg
	}
	return reg
}

func (c *RegCompiler) freeTempReg(reg int) {
	if reg == c.nextTempReg-1 {
		c.nextTempReg--
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

func (c *RegCompiler) emitRegBuiltin(builtinIdx, numArgs int) {
	c.instructions = append(c.instructions, MakeRegInstruction2(OpRegBuiltin, builtinIdx, numArgs)...)
}

func (c *RegCompiler) emitRegReturn(reg int) {
	c.instructions = append(c.instructions, MakeRegInstruction1(OpRegReturn, reg)...)
}

func (c *RegCompiler) emitRegArray(dst, startReg, count int) {
	c.instructions = append(c.instructions, MakeRegInstruction(OpRegArray, dst, startReg, count)...)
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

// CompileReg compiles a program using the register-based compiler
func CompileReg(program *parser.Program) (*Bytecode, error) {
	c := NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		return nil, err
	}
	return c.Bytecode(), nil
}
