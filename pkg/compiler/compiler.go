// pkg/compiler/compiler.go
package compiler

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// SymbolScope represents the scope of a symbol
type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
)

// Symbol represents a named variable in a scope
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable manages symbol definitions and resolution
type SymbolTable struct {
	Outer         *SymbolTable
	Store         map[string]Symbol
	NumDefinitions int
	FreeSymbols   []Symbol
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	s := &SymbolTable{
		Store:       make(map[string]Symbol),
		FreeSymbols: []Symbol{},
	}
	// Define built-in functions
	s.DefineBuiltin(0, "len")
	s.DefineBuiltin(1, "print")
	s.DefineBuiltin(2, "println")
	s.DefineBuiltin(3, "typeOf")
	s.DefineBuiltin(4, "substr")
	s.DefineBuiltin(5, "split")
	s.DefineBuiltin(6, "join")
	s.DefineBuiltin(7, "trim")
	s.DefineBuiltin(8, "upper")
	s.DefineBuiltin(9, "lower")
	s.DefineBuiltin(10, "containsStr")
	s.DefineBuiltin(11, "replace")
	s.DefineBuiltin(12, "startsWith")
	s.DefineBuiltin(13, "endsWith")
	s.DefineBuiltin(14, "abs")
	s.DefineBuiltin(15, "floor")
	s.DefineBuiltin(16, "ceil")
	s.DefineBuiltin(17, "sqrt")
	s.DefineBuiltin(18, "pow")
	s.DefineBuiltin(19, "min")
	s.DefineBuiltin(20, "max")
	s.DefineBuiltin(21, "int")
	s.DefineBuiltin(22, "float")
	s.DefineBuiltin(23, "string")
	s.DefineBuiltin(24, "push")
	s.DefineBuiltin(25, "pop")
	s.DefineBuiltin(26, "first")
	s.DefineBuiltin(27, "last")
	s.DefineBuiltin(28, "rest")
	s.DefineBuiltin(29, "concat")
	s.DefineBuiltin(30, "indexOf")
	s.DefineBuiltin(31, "containsArr")
	s.DefineBuiltin(32, "keys")
	s.DefineBuiltin(33, "values")
	s.DefineBuiltin(34, "hasKey")
	s.DefineBuiltin(35, "delete")
	s.DefineBuiltin(36, "range")
	s.DefineBuiltin(37, "sort")
	s.DefineBuiltin(38, "sum")
	s.DefineBuiltin(39, "avg")
	s.DefineBuiltin(40, "reverse")
	return s
}

// NewEnclosedSymbolTable creates a new symbol table with an outer scope
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// Define adds a new symbol to the symbol table
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.NumDefinitions}

	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.Store[name] = symbol
	s.NumDefinitions++
	return symbol
}

// Resolve finds a symbol in the symbol table or outer scopes
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := s.Store[name]
	if !ok && s.Outer != nil {
		symbol, ok = s.Outer.Resolve(name)
		if !ok {
			return symbol, false
		}

		// If we find a symbol in outer scope that's not global or builtin,
		// it becomes a free variable
		if symbol.Scope == GlobalScope || symbol.Scope == BuiltinScope {
			return symbol, true
		}

		// Add to free symbols if not already there
		free := Symbol{Name: name, Scope: FreeScope, Index: len(s.FreeSymbols)}
		s.FreeSymbols = append(s.FreeSymbols, symbol)
		s.Store[name] = free
		return free, true
	}
	return symbol, ok
}

// DefineBuiltin adds a built-in function symbol
func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Scope: BuiltinScope, Index: index}
	s.Store[name] = symbol
	return symbol
}

// CompilationScope represents a compilation scope
type CompilationScope struct {
	instructions []byte
}

// Bytecode represents compiled bytecode
type Bytecode struct {
	Instructions []byte
	Constants    []objects.Object
}

// CompiledFunction represents a compiled function
type CompiledFunction struct {
	Instructions   []byte
	NumLocals      int
	NumParameters  int
	FreeVariables  []Symbol // Free variables captured from outer scope
}

// Type returns the object type
func (cf *CompiledFunction) Type() objects.ObjectType { return objects.FunctionType }

// Inspect returns the string representation
func (cf *CompiledFunction) Inspect() string { return fmt.Sprintf("CompiledFunction[%d]", len(cf.Instructions)) }

// ToBool converts to boolean
func (cf *CompiledFunction) ToBool() *objects.Bool { return objects.TRUE }

// HashKey returns the hash key
func (cf *CompiledFunction) HashKey() objects.HashKey {
	return objects.HashKey{Type: objects.FunctionType, Value: 0}
}

// EmittedInstruction represents the last emitted instruction
type EmittedInstruction struct {
	Opcode   Opcode
	Position int
}

// Compiler transforms AST into bytecode
type Compiler struct {
	constants []objects.Object

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

// New creates a new compiler
func New() *Compiler {
	return &Compiler{
		constants:   []objects.Object{},
		symbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{{instructions: []byte{}}},
		scopeIndex:  0,
	}
}

// NewWithState creates a new compiler with existing state
func NewWithState(s *SymbolTable, constants []objects.Object) *Compiler {
	return &Compiler{
		constants:   constants,
		symbolTable: s,
		scopes:      []CompilationScope{{instructions: []byte{}}},
		scopeIndex:  0,
	}
}

// Compile compiles an AST node into bytecode
func (c *Compiler) Compile(node parser.Node) error {
	switch node := node.(type) {
	case *parser.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

	case *parser.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err
		}
		c.emit(OpPop)

	case *parser.VarStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		switch symbol.Scope {
		case GlobalScope:
			c.emit(OpSetGlobal, symbol.Index)
		case LocalScope:
			c.emit(OpSetLocal, symbol.Index)
		}

	case *parser.ConstStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		switch symbol.Scope {
		case GlobalScope:
			c.emit(OpSetGlobal, symbol.Index)
		case LocalScope:
			c.emit(OpSetLocal, symbol.Index)
		}

	case *parser.ReturnStatement:
		if node.ReturnValue != nil {
			if err := c.Compile(node.ReturnValue); err != nil {
				return err
			}
		} else {
			c.emit(OpNull)
		}
		c.emit(OpReturn)

	case *parser.BlockStatement:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

	case *parser.IfStatement:
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		// Jump to else/end if false
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Compile consequence
		if err := c.Compile(node.Consequence); err != nil {
			return err
		}

		// Remove the last OpPop if the block has a return value
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Jump over alternative
		jumpPos := c.emit(OpJump, 9999)

		// Fix jump to else/end position
		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative != nil {
			if err := c.Compile(node.Alternative); err != nil {
				return err
			}
			// Remove the last OpPop if the block has a return value
			if c.lastInstruction.Opcode == OpPop {
				c.removeLastInstruction()
			}
		} else {
			c.emit(OpNull)
		}

		// Fix jump over alternative
		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

		// Add pop for if expression
		c.emit(OpPop)

	case *parser.WhileStatement:
		// Save position for loop start
		loopStart := len(c.currentInstructions())

		// Compile condition
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		// Jump if false
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Remove the last OpPop from body
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Jump back to loop start
		c.emit(OpJump, loopStart)

		// Fix jump position
		afterBodyPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterBodyPos)

	case *parser.ForStatement:
		// for (init; condition; update) { body }
		// Compile init
		if node.Init != nil {
			if err := c.Compile(node.Init); err != nil {
				return err
			}
		}

		// Save position for loop start
		loopStart := len(c.currentInstructions())

		// Compile condition (if none, use true)
		if node.Condition != nil {
			if err := c.Compile(node.Condition); err != nil {
				return err
			}
		} else {
			c.emit(OpTrue)
		}

		// Jump if false
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Remove the last OpPop from body
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Compile update
		if node.Update != nil {
			if err := c.Compile(node.Update); err != nil {
				return err
			}
			// Remove the result of update statement
			if c.lastInstruction.Opcode == OpPop {
				c.removeLastInstruction()
			}
		}

		// Jump back to loop start
		c.emit(OpJump, loopStart)

		// Fix jump position
		afterBodyPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterBodyPos)

	case *parser.ForInStatement:
		// for (key, value in iterable) { body }
		// for (value in iterable) { body }

		// Compile iterable
		if err := c.Compile(node.Iterable); err != nil {
			return err
		}

		// Initialize index to 0
		indexConst := c.addConstant(&objects.Int{Value: 0})
		c.emit(OpConstant, indexConst)

		// Initialize iterator to null
		c.emit(OpNull)

		// Loop start
		loopStart := len(c.currentInstructions())

		// Jump if finished (when iterator is null after iteration)
		jumpNotTruthyPos := c.emit(OpJumpIfFalse, 9999)

		// Duplicate current iterator state
		c.emit(OpDup)

		// Set value variable (or key if only one variable)
		if node.Value != nil {
			symbol := c.symbolTable.Define(node.Value.Value)
			c.emit(OpSetGlobal, symbol.Index)
		}

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Remove the last OpPop from body
		if c.lastInstruction.Opcode == OpPop {
			c.removeLastInstruction()
		}

		// Jump back to loop start
		c.emit(OpJump, loopStart)

		// Fix jump position
		afterBodyPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterBodyPos)

	case *parser.BreakStatement:
		c.emit(OpBreak)

	case *parser.ContinueStatement:
		c.emit(OpContinue)

	case *parser.IntegerLiteral:
		integer := &objects.Int{Value: node.Value}
		c.emit(OpConstant, c.addConstant(integer))

	case *parser.FloatLiteral:
		float := &objects.Float{Value: node.Value}
		c.emit(OpConstant, c.addConstant(float))

	case *parser.StringLiteral:
		str := &objects.String{Value: node.Value}
		c.emit(OpConstant, c.addConstant(str))

	case *parser.BooleanLiteral:
		if node.Value {
			c.emit(OpTrue)
		} else {
			c.emit(OpFalse)
		}

	case *parser.NullLiteral:
		c.emit(OpNull)

	case *parser.ArrayLiteral:
		for _, el := range node.Elements {
			if err := c.Compile(el); err != nil {
				return err
			}
		}
		c.emit(OpArray, len(node.Elements))

	case *parser.MapLiteral:
		// Sort keys for deterministic order
		keys := make([]parser.Expression, 0, len(node.Pairs))
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		for _, k := range keys {
			if err := c.Compile(k); err != nil {
				return err
			}
			if err := c.Compile(node.Pairs[k]); err != nil {
				return err
			}
		}
		c.emit(OpMap, len(node.Pairs))

	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		switch symbol.Scope {
		case GlobalScope:
			c.emit(OpGetGlobal, symbol.Index)
		case LocalScope:
			c.emit(OpGetLocal, symbol.Index)
		case BuiltinScope:
			c.emit(OpBuiltin, symbol.Index)
		case FreeScope:
			c.emit(OpGetFree, symbol.Index)
		}

	case *parser.PrefixExpression:
		if err := c.Compile(node.Right); err != nil {
			return err
		}

		switch node.Operator {
		case "-":
			c.emit(OpNeg)
		case "!":
			c.emit(OpNot)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *parser.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Right); err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(OpAdd)
		case "-":
			c.emit(OpSub)
		case "*":
			c.emit(OpMul)
		case "/":
			c.emit(OpDiv)
		case "%":
			c.emit(OpMod)
		case "==":
			c.emit(OpEqual)
		case "!=":
			c.emit(OpNotEqual)
		case "<":
			c.emit(OpLess)
		case ">":
			c.emit(OpGreater)
		case "<=":
			c.emit(OpLessEqual)
		case ">=":
			c.emit(OpGreaterEqual)
		case "&&":
			c.emit(OpAnd)
		case "||":
			c.emit(OpOr)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *parser.IndexExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}
		c.emit(OpIndex)

	case *parser.AssignmentExpression:
		// Compile the value first
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		// Handle different left-hand side types
		switch left := node.Left.(type) {
		case *parser.Identifier:
			symbol, ok := c.symbolTable.Resolve(left.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", left.Value)
			}
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpSetFree, symbol.Index)
			}
		case *parser.IndexExpression:
			// For arr[i] = value, we need stack to be: [arr, index, value]
			// But we already compiled value, which is on stack
			// We need to compile arr and index, then swap to get correct order
			if err := c.Compile(left.Left); err != nil {
				return err
			}
			if err := c.Compile(left.Index); err != nil {
				return err
			}
			// Stack is now: [value, arr, index]
			// We need: [arr, index, value]
			// OpSetIndex pops: value, index, left
			// So we need to reorder: pop value, push arr, push index, push value
			// Actually, let's change the VM to handle the order correctly
			c.emit(OpSetIndex)
		default:
			return fmt.Errorf("cannot assign to %T", left)
		}

	case *parser.FunctionLiteral:
		// If named function, define the name first for recursion support
		var funcSymbol Symbol
		if node.Name != "" {
			funcSymbol = c.symbolTable.Define(node.Name)
		}

		// Enter function scope
		c.enterScope()

		// Define parameters as local variables
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		// Compile body
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// Ensure function ends with return
		if c.lastInstruction.Opcode != OpReturn {
			c.emit(OpNull)
			c.emit(OpReturn)
		}

		// Leave function scope and get compiled function
		compiledFn := c.leaveScope()
		compiledFn.NumParameters = len(node.Parameters)

		// Add function to constants
		fnIndex := c.addConstant(compiledFn)

		// Emit OpClosure with function index and number of free variables
		// For each free variable, emit code to push its value onto the stack
		for _, freeVar := range compiledFn.FreeVariables {
			// Push the value of each free variable onto the stack
			// This will be captured in the closure
			switch freeVar.Scope {
			case GlobalScope:
				c.emit(OpGetGlobal, freeVar.Index)
			case LocalScope:
				c.emit(OpGetLocal, freeVar.Index)
			case FreeScope:
				// Free variable in outer function - get it from outer's free vars
				c.emit(OpGetFree, freeVar.Index)
			}
		}

		// Emit closure instruction
		c.emit(OpClosure, fnIndex, len(compiledFn.FreeVariables))

		// If named function, bind it to its name
		if node.Name != "" {
			switch funcSymbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, funcSymbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, funcSymbol.Index)
			}
		}

	case *parser.CallExpression:
		// Compile function
		if err := c.Compile(node.Function); err != nil {
			return err
		}

		// Compile arguments
		for _, arg := range node.Arguments {
			if err := c.Compile(arg); err != nil {
				return err
			}
		}

		c.emit(OpCall, len(node.Arguments))

	case *parser.DotExpression:
		// Compile object
		if err := c.Compile(node.Object); err != nil {
			return err
		}
		// Add property name to constants
		nameConst := c.addConstant(&objects.String{Value: node.Property.Value})
		c.emit(OpGetMethod, nameConst)

	case *parser.TernaryExpression:
		// condition ? consequent : alternative
		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		jumpFalsePos := c.emit(OpJumpIfFalse, 9999)

		if err := c.Compile(node.Consequent); err != nil {
			return err
		}

		jumpEndPos := c.emit(OpJump, 9999)

		afterConsequentPos := len(c.currentInstructions())
		c.changeOperand(jumpFalsePos, afterConsequentPos)

		if err := c.Compile(node.Alternative); err != nil {
			return err
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpEndPos, afterAlternativePos)

	case *parser.CompoundAssignmentExpression:
		// x += 1 is equivalent to x = x + 1
		// Get the current value
		switch left := node.Left.(type) {
		case *parser.Identifier:
			symbol, ok := c.symbolTable.Resolve(left.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", left.Value)
			}

			// Get current value
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpGetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpGetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpGetFree, symbol.Index)
			}

			// Compile right side
			if err := c.Compile(node.Right); err != nil {
				return err
			}

			// Apply operation
			switch node.Operator {
			case "+=":
				c.emit(OpAdd)
			case "-=":
				c.emit(OpSub)
			case "*=":
				c.emit(OpMul)
			case "/=":
				c.emit(OpDiv)
			}

			// Store result
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpSetFree, symbol.Index)
			}
		}

	case *parser.PostfixExpression:
		// x++ or x--
		switch left := node.Left.(type) {
		case *parser.Identifier:
			symbol, ok := c.symbolTable.Resolve(left.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", left.Value)
			}

			// Get current value
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpGetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpGetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpGetFree, symbol.Index)
			}

			// Duplicate for result (postfix returns old value)
			c.emit(OpDup)

			// Add 1 or subtract 1
			one := c.addConstant(&objects.Int{Value: 1})
			c.emit(OpConstant, one)

			switch node.Operator {
			case "++":
				c.emit(OpAdd)
			case "--":
				c.emit(OpSub)
			}

			// Store result
			switch symbol.Scope {
			case GlobalScope:
				c.emit(OpSetGlobal, symbol.Index)
			case LocalScope:
				c.emit(OpSetLocal, symbol.Index)
			case FreeScope:
				c.emit(OpSetFree, symbol.Index)
			}
		}

	default:
		return fmt.Errorf("unknown node type: %T", node)
	}

	return nil
}

// Bytecode returns the compiled bytecode
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

// emit adds an instruction to the bytecode
func (c *Compiler) emit(op Opcode, operands ...int) int {
	ins := Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

// addInstruction adds an instruction to the current scope
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), ins...)
	return posNewInstruction
}

// setLastInstruction updates the last instruction tracking
func (c *Compiler) setLastInstruction(op Opcode, pos int) {
	previous := c.lastInstruction
	c.lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
	c.previousInstruction = previous
}

// lastInstructionIs returns true if the last instruction matches
func (c *Compiler) lastInstructionIs(op Opcode) bool {
	return c.lastInstruction.Opcode == op
}

// removeLastInstruction removes the last instruction
func (c *Compiler) removeLastInstruction() {
	ins := c.currentInstructions()
	c.scopes[c.scopeIndex].instructions = ins[:len(ins)-1]
	c.lastInstruction = c.previousInstruction
}

// replaceInstruction replaces an instruction at a position
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

// changeOperand changes the operand of an instruction
func (c *Compiler) changeOperand(pos int, operand int) {
	op := Opcode(c.currentInstructions()[pos])
	newInstruction := Make(op, operand)
	c.replaceInstruction(pos, newInstruction)
}

// currentInstructions returns the current scope's instructions
func (c *Compiler) currentInstructions() []byte {
	return c.scopes[c.scopeIndex].instructions
}

// addConstant adds a constant to the constant pool
func (c *Compiler) addConstant(obj objects.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// enterScope enters a new compilation scope
func (c *Compiler) enterScope() {
	scope := CompilationScope{instructions: []byte{}}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// leaveScope leaves the current scope and returns the compiled function
func (c *Compiler) leaveScope() *CompiledFunction {
	instructions := c.currentInstructions()

	// Get the number of local variables in this scope
	numLocals := c.symbolTable.NumDefinitions

	// Capture free variables before leaving scope
	freeVars := make([]Symbol, len(c.symbolTable.FreeSymbols))
	copy(freeVars, c.symbolTable.FreeSymbols)

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer

	return &CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		FreeVariables: freeVars,
	}
}
