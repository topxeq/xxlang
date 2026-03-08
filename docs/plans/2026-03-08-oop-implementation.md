# OOP System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement single-inheritance OOP system with class, extends, new, this, and super keywords.

**Architecture:** Prototype-chain based inheritance with Class and Instance objects. Methods stored on Class, fields on Instance. VM tracks current instance for `this` binding.

**Tech Stack:** Go standard library, existing lexer/parser/compiler/VM infrastructure.

---

## Task 1: Add Class and Instance Object Types

**Files:**
- Create: `pkg/objects/class.go`
- Test: `pkg/objects/class_test.go`

**Step 1: Write the failing test**

```go
// pkg/objects/class_test.go
package objects

import "testing"

func TestClassObject(t *testing.T) {
	class := &Class{
		Name:       "Person",
		SuperClass: nil,
		Methods:    make(map[string]*CompiledFunction),
		Fields:     make(map[string]Object),
	}

	if class.Type() != ClassType {
		t.Errorf("expected ClassType, got %s", class.Type())
	}
	if class.Inspect() != "class Person" {
		t.Errorf("expected 'class Person', got %s", class.Inspect())
	}
}

func TestInstanceObject(t *testing.T) {
	class := &Class{Name: "Person"}
	instance := &Instance{
		Class:  class,
		Fields: map[string]Object{"name": NewString("Alice")},
	}

	if instance.Type() != InstanceType {
		t.Errorf("expected InstanceType, got %s", instance.Type())
	}
	if instance.Inspect() != "Person instance" {
		t.Errorf("expected 'Person instance', got %s", instance.Inspect())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/objects/... -run TestClassObject`
Expected: FAIL - undefined: Class

**Step 3: Implement Class and Instance types**

```go
// pkg/objects/class.go
package objects

import "bytes"

// Class represents a class definition
type Class struct {
	Name       string
	SuperClass *Class
	Methods    map[string]*CompiledFunction
	InitMethod *CompiledFunction // constructor
	Fields     map[string]Object // default field values
}

func (c *Class) Type() ObjectType { return ClassType }
func (c *Class) Inspect() string  { return "class " + c.Name }
func (c *Class) ToBool() *Bool    { return TRUE }
func (c *Class) HashKey() HashKey {
	return HashKey{Type: ClassType, Value: uint64(uintptr(unsafe.Pointer(c)))}
}

// Instance represents an instance of a class
type Instance struct {
	Class  *Class
	Fields map[string]Object
}

func (i *Instance) Type() ObjectType { return InstanceType }
func (i *Instance) Inspect() string  { return i.Class.Name + " instance" }
func (i *Instance) ToBool() *Bool    { return TRUE }
func (i *Instance) HashKey() HashKey {
	return HashKey{Type: InstanceType, Value: uint64(uintptr(unsafe.Pointer(i)))}
}
```

**Step 4: Add unsafe import to class.go**

```go
import (
	"bytes"
	"unsafe"
)
```

**Step 5: Run test to verify it passes**

Run: `go test ./pkg/objects/... -run "TestClassObject|TestInstanceObject"`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/objects/class.go pkg/objects/class_test.go
git commit -m "feat(objects): add Class and Instance object types"
```

---

## Task 2: Add OOP Opcodes

**Files:**
- Modify: `pkg/compiler/opcode.go`
- Test: `pkg/compiler/opcode_test.go`

**Step 1: Add new opcode constants**

In `pkg/compiler/opcode.go`, add after `OpSetExport`:

```go
	// Class operations
	OpClass    // Create class object
	OpNew      // Create instance
	OpGetField // Get instance field
	OpSetField // Set instance field
	OpSuper    // Get superclass method
```

**Step 2: Add opcode definitions**

In the `definitions` map, add:

```go
	// Class operations
	OpClass:    {"OpClass", []int{2}},    // 2-byte: class name constant index
	OpNew:      {"OpNew", []int{1}},      // 1-byte: argument count
	OpGetField: {"OpGetField", []int{2}}, // 2-byte: field name constant index
	OpSetField: {"OpSetField", []int{2}}, // 2-byte: field name constant index
	OpSuper:    {"OpSuper", []int{2}},    // 2-byte: method name constant index
```

**Step 3: Run tests to verify no breakage**

Run: `go test ./pkg/compiler/...`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/compiler/opcode.go
git commit -m "feat(compiler): add OOP opcodes (OpClass, OpNew, OpGetField, OpSetField, OpSuper)"
```

---

## Task 3: Add SuperExpression AST Node

**Files:**
- Modify: `pkg/parser/ast.go`
- Test: `pkg/parser/ast_test.go`

**Step 1: Add SuperExpression struct**

In `pkg/parser/ast.go`, add after `ThisExpression`:

```go
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
```

**Step 2: Run tests to verify no breakage**

Run: `go test ./pkg/parser/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/parser/ast.go
git commit -m "feat(parser): add SuperExpression and SuperCallExpression AST nodes"
```

---

## Task 4: Add Class Statement Parsing

**Files:**
- Modify: `pkg/parser/parser.go`
- Test: `pkg/parser/parser_test.go`

**Step 1: Write the failing test**

Add to `pkg/parser/parser_test.go`:

```go
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
			`class Person { func init(name) { this.name = name; }  var name = ""; }`,
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/parser/... -run TestClassStatements`
Expected: FAIL

**Step 3: Add class parsing to parseStatement**

In `pkg/parser/parser.go`, in the `parseStatement` switch, add:

```go
	case lexer.TokenClass:
		return p.parseClassStatement()
```

**Step 4: Implement parseClassStatement**

Add to `pkg/parser/parser.go`:

```go
// parseClassStatement parses a class declaration
func (p *Parser) parseClassStatement() *ClassStatement {
	stmt := &ClassStatement{Token: p.curToken}

	// Expect class name
	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for extends
	if p.peekTokenIs(lexer.TokenExtends) {
		p.nextToken()
		if !p.expectPeek(lexer.TokenIdent) {
			return nil
		}
		stmt.SuperClass = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// Expect {
	if !p.expectPeek(lexer.TokenLBrace) {
		return nil
	}

	// Parse class body
	stmt.Methods = []*FunctionLiteral{}
	stmt.Fields = []*VarStatement{}

	for !p.curTokenIs(lexer.TokenRBrace) && !p.curTokenIs(lexer.TokenEOF) {
		p.nextToken()

		if p.curTokenIs(lexer.TokenFunc) {
			method := p.parseFunctionLiteral()
			if method != nil {
				stmt.Methods = append(stmt.Methods, method)
			}
		} else if p.curTokenIs(lexer.TokenVar) {
			field := p.parseVarStatement()
			if field != nil {
				stmt.Fields = append(stmt.Fields, field)
			}
		} else {
			p.addError(fmt.Sprintf("unexpected token in class body: %s", p.curToken.Type))
			return nil
		}

		// Skip semicolons
		if p.curTokenIs(lexer.TokenSemicolon) {
			p.nextToken()
		}
	}

	return stmt
}
```

**Step 5: Add new and this prefix parsing**

In `pkg/parser/parser.go`, in `New()` function, register prefix parsers:

```go
	p.registerPrefix(lexer.TokenNew, p.parseNewExpression)
	p.registerPrefix(lexer.TokenThis, p.parseThisExpression)
	p.registerPrefix(lexer.TokenSuper, p.parseSuperExpression)
```

**Step 6: Implement expression parsers**

Add to `pkg/parser/parser.go`:

```go
// parseNewExpression parses a new expression
func (p *Parser) parseNewExpression() Expression {
	expr := &NewExpression{Token: p.curToken}

	// Expect class name
	if !p.expectPeek(lexer.TokenIdent) {
		return nil
	}
	expr.Class = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect (
	if !p.expectPeek(lexer.TokenLParen) {
		return nil
	}

	// Parse arguments
	expr.Arguments = p.parseExpressionList(lexer.TokenRParen)

	return expr
}

// parseThisExpression parses this expression
func (p *Parser) parseThisExpression() Expression {
	return &ThisExpression{Token: p.curToken}
}

// parseSuperExpression parses super expression
func (p *Parser) parseSuperExpression() Expression {
	expr := &SuperExpression{Token: p.curToken}

	// Check for super.method() pattern
	if p.peekTokenIs(lexer.TokenDot) {
		p.nextToken() // consume dot
		if !p.expectPeek(lexer.TokenIdent) {
			return nil
		}
		methodName := p.curToken.Literal

		// Check for method call
		if p.peekTokenIs(lexer.TokenLParen) {
			p.nextToken()
			callExpr := &SuperCallExpression{
				Token:  expr.Token,
				Method: methodName,
				Args:   p.parseExpressionList(lexer.TokenRParen),
			}
			return callExpr
		}
	}

	return expr
}
```

**Step 7: Add parseExpressionList helper**

Add to `pkg/parser/parser.go`:

```go
// parseExpressionList parses a comma-separated list of expressions
func (p *Parser) parseExpressionList(end lexer.TokenType) []Expression {
	args := []Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.TokenComma) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}
```

**Step 8: Run tests to verify they pass**

Run: `go test ./pkg/parser/... -run "TestClassStatements|TestNewExpression|TestThisExpression"`
Expected: PASS

**Step 9: Commit**

```bash
git add pkg/parser/parser.go pkg/parser/parser_test.go
git commit -m "feat(parser): add class, new, this, and super parsing"
```

---

## Task 5: Add Class Compilation

**Files:**
- Modify: `pkg/compiler/compiler.go`
- Test: `pkg/compiler/compiler_test.go`

**Step 1: Write the failing test**

Add to `pkg/compiler/compiler_test.go`:

```go
func TestClassStatementCompilation(t *testing.T) {
	input := `
		class Person {
			var name = ""
			func init(name) { this.name = name }
		}
	`
	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	bytecode := compiler.Bytecode()
	if len(bytecode.Instructions) == 0 {
		t.Error("expected instructions to be generated")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/compiler/... -run TestClassStatementCompilation`
Expected: FAIL

**Step 3: Add class compilation to Compile switch**

In `pkg/compiler/compiler.go`, in the `Compile` switch, add:

```go
	case *parser.ClassStatement:
		return c.compileClassStatement(node)

	case *parser.NewExpression:
		return c.compileNewExpression(node)

	case *parser.ThisExpression:
		return c.compileThisExpression(node)

	case *parser.SuperCallExpression:
		return c.compileSuperCallExpression(node)
```

**Step 4: Implement compileClassStatement**

Add to `pkg/compiler/compiler.go`:

```go
// compileClassStatement compiles a class declaration
func (c *Compiler) compileClassStatement(node *parser.ClassStatement) error {
	// Compile superclass (push null if none)
	if node.SuperClass != nil {
		symbol := c.symbolTable.Resolve(node.SuperClass.Value)
		if symbol == nil {
			return fmt.Errorf("undefined superclass: %s", node.SuperClass.Value)
		}
		c.emit(OpGetGlobal, symbol.Index)
	} else {
		c.emit(OpNull)
	}

	// Compile default fields as a map
	c.emit(OpMap, 0)
	for _, field := range node.Fields {
		// Duplicate map
		c.emit(OpDup)
		// Key
		nameIdx := c.addConstant(objects.NewString(field.Name.Value))
		c.emit(OpConstant, nameIdx)
		// Value
		if err := c.Compile(field.Value); err != nil {
			return err
		}
		c.emit(OpSetIndex)
	}

	// Compile methods as a map
	c.emit(OpMap, 0)
	for _, method := range node.Methods {
		// Duplicate map
		c.emit(OpDup)
		// Key (method name)
		nameIdx := c.addConstant(objects.NewString(method.Name))
		c.emit(OpConstant, nameIdx)
		// Compile method as function
		if err := c.Compile(method); err != nil {
			return err
		}
		c.emit(OpSetIndex)
	}

	// Create class
	nameIdx := c.addConstant(objects.NewString(node.Name.Value))
	c.emit(OpClass, nameIdx)

	// Store class in global
	symbol := c.symbolTable.Define(node.Name.Value)
	c.emit(OpSetGlobal, symbol.Index)

	return nil
}
```

**Step 5: Implement compileNewExpression**

Add to `pkg/compiler/compiler.go`:

```go
// compileNewExpression compiles a new expression
func (c *Compiler) compileNewExpression(node *parser.NewExpression) error {
	// Get class
	symbol := c.symbolTable.Resolve(node.Class.String())
	if symbol == nil {
		return fmt.Errorf("undefined class: %s", node.Class.String())
	}
	c.emit(OpGetGlobal, symbol.Index)

	// Compile arguments
	for _, arg := range node.Arguments {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}

	// Create instance
	c.emit(OpNew, len(node.Arguments))

	return nil
}
```

**Step 6: Implement compileThisExpression**

Add to `pkg/compiler/compiler.go`:

```go
// compileThisExpression compiles this expression
func (c *Compiler) compileThisExpression(node *parser.ThisExpression) error {
	// Push current instance reference onto stack
	// This is handled at VM level - we emit a special marker
	c.emit(OpGetLocal, 0) // this is always first local in method context
	return nil
}
```

**Step 7: Implement compileSuperCallExpression**

Add to `pkg/compiler/compiler.go`:

```go
// compileSuperCallExpression compiles a super.method() call
func (c *Compiler) compileSuperCallExpression(node *parser.SuperCallExpression) error {
	// Get super method
	nameIdx := c.addConstant(objects.NewString(node.Method))
	c.emit(OpSuper, nameIdx)

	// Push this as first argument
	c.emit(OpGetLocal, 0)

	// Compile remaining arguments
	for _, arg := range node.Args {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}

	// Call method
	c.emit(OpCall, len(node.Args)+1) // +1 for this

	return nil
}
```

**Step 8: Add imports if needed**

Ensure `fmt` and `objects` are imported in compiler.go.

**Step 9: Run tests**

Run: `go test ./pkg/compiler/... -run TestClassStatementCompilation`
Expected: PASS

**Step 10: Commit**

```bash
git add pkg/compiler/compiler.go pkg/compiler/compiler_test.go
git commit -m "feat(compiler): add class, new, this, and super compilation"
```

---

## Task 6: Add VM Execution for Classes

**Files:**
- Modify: `pkg/vm/vm.go`
- Test: `pkg/vm/vm_test.go`

**Step 1: Write the failing test**

Add to `pkg/vm/vm_test.go`:

```go
func TestClassCreation(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Person {
				var name = ""
			}
			var p = new Person()
			p.name
			`,
			"",
		},
		{
			`
			class Counter {
				var count = 0
				func inc() {
					this.count = this.count + 1
				}
			}
			var c = new Counter()
			c.inc()
			c.count
			`,
			1,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObject(t, evaluated, tt.expected)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/vm/... -run TestClassCreation`
Expected: FAIL

**Step 3: Add currentInstance to VM struct**

In `pkg/vm/vm.go`, update the VM struct:

```go
type VM struct {
	constants     []objects.Object
	stack         *Stack
	frames        []*Frame
	frameIndex    int
	globals       []objects.Object
	loader        *module.Loader
	currentModule *objects.Module
	sourcePath    string
	currentInstance *objects.Instance  // Current instance for this binding
}
```

**Step 4: Add opcode handlers in Run switch**

In `pkg/vm/vm.go`, in the `Run` switch, add:

```go
		case compiler.OpClass:
			if err := vm.executeOpClass(); err != nil {
				return err
			}

		case compiler.OpNew:
			if err := vm.executeOpNew(); err != nil {
				return err
			}

		case compiler.OpGetField:
			if err := vm.executeOpGetField(); err != nil {
				return err
			}

		case compiler.OpSetField:
			if err := vm.executeOpSetField(); err != nil {
				return err
			}

		case compiler.OpSuper:
			if err := vm.executeOpSuper(); err != nil {
				return err
			}
```

**Step 5: Implement OpClass execution**

Add to `pkg/vm/vm.go`:

```go
// executeOpClass creates a class object
func (vm *VM) executeOpClass() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	name := vm.constants[nameIdx].(*objects.String).Value

	// Pop methods map
	methodsObj := vm.stack.Pop()
	methods, ok := methodsObj.(*objects.Map)
	if !ok {
		return fmt.Errorf("methods must be a map")
	}

	// Pop fields map
	fieldsObj := vm.stack.Pop()
	fields, ok := fieldsObj.(*objects.Map)
	if !ok {
		return fmt.Errorf("fields must be a map")
	}

	// Pop superclass (or null)
	superObj := vm.stack.Pop()
	var superClass *objects.Class
	if superObj != objects.NULL {
		superClass, ok = superObj.(*objects.Class)
		if !ok {
			return fmt.Errorf("superclass must be a class")
		}
	}

	// Build methods map
	classMethods := make(map[string]*objects.CompiledFunction)
	var initMethod *objects.CompiledFunction
	for _, pair := range methods.Pairs {
		name := pair.Key.(*objects.String).Value
		if fn, ok := pair.Value.(*objects.CompiledFunction); ok {
			classMethods[name] = fn
			if name == "init" {
				initMethod = fn
			}
		}
	}

	// Build fields map
	classFields := make(map[string]objects.Object)
	for _, pair := range fields.Pairs {
		name := pair.Key.(*objects.String).Value
		classFields[name] = pair.Value
	}

	// Create class
	class := &objects.Class{
		Name:       name,
		SuperClass: superClass,
		Methods:    classMethods,
		InitMethod: initMethod,
		Fields:     classFields,
	}

	vm.stack.Push(class)
	return nil
}
```

**Step 6: Implement OpNew execution**

Add to `pkg/vm/vm.go`:

```go
// executeOpNew creates a new instance
func (vm *VM) executeOpNew() error {
	argCount := int(vm.readUint16())
	vm.currentFrame().IP += 2

	classObj := vm.stack.Pop()
	class, ok := classObj.(*objects.Class)
	if !ok {
		return fmt.Errorf("cannot use 'new' on non-class type")
	}

	// Collect fields from class hierarchy
	fields := make(map[string]objects.Object)
	vm.collectFields(class, fields)

	// Create instance
	instance := &objects.Instance{
		Class:  class,
		Fields: fields,
	}

	// Call init if exists
	if class.InitMethod != nil {
		// Set current instance for this binding
		vm.currentInstance = instance

		// Push instance as first argument (this)
		vm.stack.Push(instance)

		// Pop arguments and push them back in order
		args := make([]objects.Object, argCount)
		for i := argCount - 1; i >= 0; i-- {
			args[i] = vm.stack.Pop()
		}
		for _, arg := range args {
			vm.stack.Push(arg)
		}

		// Call init
		if err := vm.callFunction(class.InitMethod, argCount+1); err != nil {
			return err
		}

		// Pop return value
		vm.stack.Pop()
	}

	vm.stack.Push(instance)
	return nil
}

// collectFields collects fields from class hierarchy
func (vm *VM) collectFields(class *objects.Class, fields map[string]objects.Object) {
	// Collect from parent first
	if class.SuperClass != nil {
		vm.collectFields(class.SuperClass, fields)
	}
	// Collect from this class (overrides parent)
	for name, value := range class.Fields {
		fields[name] = value
	}
}
```

**Step 7: Implement OpGetField execution**

Add to `pkg/vm/vm.go`:

```go
// executeOpGetField gets a field from an instance
func (vm *VM) executeOpGetField() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	name := vm.constants[nameIdx].(*objects.String).Value

	obj := vm.stack.Pop()
	instance, ok := obj.(*objects.Instance)
	if !ok {
		return fmt.Errorf("cannot access field '%s' on %s", name, obj.Type())
	}

	value, ok := instance.Fields[name]
	if !ok {
		vm.stack.Push(objects.NULL)
	} else {
		vm.stack.Push(value)
	}
	return nil
}
```

**Step 8: Implement OpSetField execution**

Add to `pkg/vm/vm.go`:

```go
// executeOpSetField sets a field on an instance
func (vm *VM) executeOpSetField() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	name := vm.constants[nameIdx].(*objects.String).Value

	value := vm.stack.Pop()
	obj := vm.stack.Pop()

	instance, ok := obj.(*objects.Instance)
	if !ok {
		return fmt.Errorf("cannot set field '%s' on %s", name, obj.Type())
	}

	instance.Fields[name] = value
	vm.stack.Push(value)
	return nil
}
```

**Step 9: Implement OpSuper execution**

Add to `pkg/vm/vm.go`:

```go
// executeOpSuper gets a method from the superclass
func (vm *VM) executeOpSuper() error {
	nameIdx := int(vm.readUint16())
	vm.currentFrame().IP += 2

	name := vm.constants[nameIdx].(*objects.String).Value

	if vm.currentInstance == nil {
		return fmt.Errorf("cannot use 'super' outside of method context")
	}

	class := vm.currentInstance.Class
	if class.SuperClass == nil {
		return fmt.Errorf("class '%s' has no superclass", class.Name)
	}

	// Find method in superclass chain
	method := vm.findMethod(class.SuperClass, name)
	if method == nil {
		return fmt.Errorf("method '%s' not found in superclass", name)
	}

	vm.stack.Push(method)
	return nil
}

// findMethod finds a method in class hierarchy
func (vm *VM) findMethod(class *objects.Class, name string) *objects.CompiledFunction {
	for c := class; c != nil; c = c.SuperClass {
		if method, ok := c.Methods[name]; ok {
			return method
		}
	}
	return nil
}
```

**Step 10: Update GetMethod to handle instances**

Modify `executeGetMethod` in `pkg/vm/vm.go` to also handle Instance:

```go
// Add this case in executeGetMethod:
	// Handle Instance objects (for method access)
	if inst, ok := obj.(*objects.Instance); ok {
		method := vm.findMethod(inst.Class, name)
		if method != nil {
			// Set current instance for this binding
			vm.currentInstance = inst
			vm.stack.Push(method)
			return nil
		}
		// Check if it's a field access
		if value, ok := inst.Fields[name]; ok {
			vm.stack.Push(value)
			return nil
		}
		return fmt.Errorf("method or field '%s' not found on class '%s'", name, inst.Class.Name)
	}
```

**Step 11: Run tests**

Run: `go test ./pkg/vm/... -run TestClassCreation`
Expected: PASS

**Step 12: Commit**

```bash
git add pkg/vm/vm.go pkg/vm/vm_test.go
git commit -m "feat(vm): add class/instance execution support"
```

---

## Task 7: Integration Tests

**Files:**
- Create: `tests/oop_test.go`

**Step 1: Write comprehensive integration tests**

```go
// tests/oop_test.go
package tests

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

func TestOOPBasicClass(t *testing.T) {
	input := `
		class Point {
			var x = 0
			var y = 0

			func init(x, y) {
				this.x = x
				this.y = y
			}

			func add() {
				return this.x + this.y
			}
		}

		var p = new Point(3, 4)
		p.add()
	`

	result := runXxlang(t, input)
	if result != "7" {
		t.Errorf("expected 7, got %s", result)
	}
}

func TestOOPInheritance(t *testing.T) {
	input := `
		class Animal {
			var name = ""
			func init(name) { this.name = name }
			func speak() { return this.name }
		}

		class Dog extends Animal {
			func speak() { return this.name + " barks" }
		}

		var d = new Dog("Buddy")
		d.speak()
	`

	result := runXxlang(t, input)
	if result != "Buddy barks" {
		t.Errorf("expected 'Buddy barks', got %s", result)
	}
}

func TestOOPSuperCall(t *testing.T) {
	input := `
		class Animal {
			var name = ""
			func init(name) { this.name = name }
		}

		class Dog extends Animal {
			var breed = ""
			func init(name, breed) {
				super.init(name)
				this.breed = breed
			}
		}

		var d = new Dog("Buddy", "Golden")
		d.name + " " + d.breed
	`

	result := runXxlang(t, input)
	if result != "Buddy Golden" {
		t.Errorf("expected 'Buddy Golden', got %s", result)
	}
}

func TestOOPMethodLookup(t *testing.T) {
	input := `
		class A { func foo() { return "A" } }
		class B extends A { }
		class C extends B { }

		var c = new C()
		c.foo()
	`

	result := runXxlang(t, input)
	if result != "A" {
		t.Errorf("expected 'A', got %s", result)
	}
}

func TestOOPFieldInheritance(t *testing.T) {
	input := `
		class Animal {
			var name = "unknown"
			var age = 0
		}

		class Dog extends Animal {
			var breed = "mixed"
		}

		var d = new Dog()
		d.name + " " + d.age + " " + d.breed
	`

	result := runXxlang(t, input)
	if result != "unknown 0 mixed" {
		t.Errorf("expected 'unknown 0 mixed', got %s", result)
	}
}

// Helper function
func runXxlang(t *testing.T, input string) string {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		t.Fatalf("vm error: %v", err)
	}

	result := v.LastPopped()
	if result != nil {
		return result.Inspect()
	}
	return ""
}
```

**Step 2: Run all tests**

Run: `go test ./tests/... -run TestOOP`
Expected: PASS

**Step 3: Run full test suite**

Run: `go test ./...`
Expected: PASS

**Step 4: Commit**

```bash
git add tests/oop_test.go
git commit -m "test: add comprehensive OOP integration tests"
```

---

## Task 8: Update Documentation

**Files:**
- Modify: `README.md`

**Step 1: Add OOP section to README**

Add after the module system section:

```markdown
## Object-Oriented Programming

Xxlang supports lightweight OOP with single inheritance.

### Class Definition

```xxl
class Person {
    var name = ""
    var age = 0

    func init(name, age) {
        this.name = name
        this.age = age
    }

    func greet() {
        return "Hello, " + this.name
    }
}
```

### Inheritance

```xxl
class Animal {
    var name = ""
    func init(name) { this.name = name }
    func speak() { return this.name + " makes a sound" }
}

class Dog extends Animal {
    func speak() { return this.name + " barks" }
}

var dog = new Dog("Buddy")
println(dog.speak())  // "Buddy barks"
```

### Super Calls

```xxl
class Dog extends Animal {
    var breed = ""
    func init(name, breed) {
        super.init(name)  // Call parent constructor
        this.breed = breed
    }
}
```

### Features

- Single inheritance with `extends`
- Automatic `this` binding in methods
- `init` constructor called by `new`
- `super.method()` for parent method calls
- All members are public
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add OOP documentation to README"
```

---

## Summary

After completing all tasks, the OOP system will support:

- ✅ `class` declarations with fields and methods
- ✅ `extends` for single inheritance
- ✅ `new` for instance creation
- ✅ `this` for instance reference
- ✅ `super.method()` for parent calls
- ✅ Method lookup through inheritance chain
- ✅ Field inheritance from parent classes

All tests should pass: `go test ./...`
