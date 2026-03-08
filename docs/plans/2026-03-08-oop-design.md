# OOP System Design Document

**Date**: 2026-03-08
**Version**: 1.0
**Status**: Approved

## Overview

This document defines the design for Xxlang's Object-Oriented Programming (OOP) system, featuring single inheritance, automatic `this` binding, and prototype-chain-based method lookup.

## Design Decisions Summary

| Feature | Decision |
|---------|----------|
| Constructor | `init` method, automatically called by `new` |
| `this` Binding | Automatic, no explicit declaration needed |
| Inheritance | `super` keyword to call parent methods |
| Access Control | All public, no private/protected |
| Member Definition | `var`/`func` declarations |
| Static Members | Not supported, use modules instead |
| Implementation | Prototype chain approach |

## Object Model

### Class Object

```go
// pkg/objects/class.go

// Class represents a class definition
type Class struct {
    Name       string
    SuperClass *Class                    // nil if no parent
    Methods    map[string]*CompiledFunction
    InitMethod *CompiledFunction         // constructor, may be nil
    Fields     map[string]Object         // default field values
}
```

### Instance Object

```go
// Instance represents an instance of a class
type Instance struct {
    Class  *Class
    Fields map[string]Object            // instance fields
}
```

### Inheritance Rules

- **Field Inheritance**: Copy default field values from class and all parent classes when creating instance
- **Method Lookup**: Traverse prototype chain `instance.Class -> SuperClass -> ...`

## Syntax

### Class Definition

```xxl
class Animal {
    var name = ""

    func init(name) {
        this.name = name
    }

    func speak() {
        return this.name + " makes a sound"
    }
}
```

### Inheritance

```xxl
class Dog extends Animal {
    var breed = ""

    func init(name, breed) {
        super.init(name)
        this.breed = breed
    }

    func speak() {
        return this.name + " barks"
    }
}
```

### Instance Creation

```xxl
var dog = new Dog("Buddy", "Golden")
println(dog.speak())  // "Buddy barks"
```

## AST Nodes

```go
// ClassStatement - class declaration (already defined)
type ClassStatement struct {
    Token      lexer.Token
    Name       *Identifier
    SuperClass *Identifier      // nil if no extends
    Body       *BlockStatement
}

// NewExpression - create instance (already defined)
type NewExpression struct {
    Token  lexer.Token
    Class  Expression         // class name
    Args   []Expression       // constructor arguments
}

// ThisExpression - this keyword (already defined)
type ThisExpression struct {
    Token lexer.Token
}

// SuperExpression - super keyword (NEW)
type SuperExpression struct {
    Token lexer.Token
}

// SuperCallExpression - super.method() (NEW)
type SuperCallExpression struct {
    Token    lexer.Token
    Method   string
    Args     []Expression
}
```

## Opcodes

### New Opcodes

```go
const (
    // Class operations
    OpClass       // Create class object
    OpNew         // Create instance
    OpGetField    // Get instance field
    OpSetField    // Set instance field
    OpSuper       // Get superclass method
)
```

### Opcode Definitions

```go
OpClass:    {"OpClass", []int{2}},      // 2-byte: class name constant index
OpNew:      {"OpNew", []int{1}},        // 1-byte: argument count
OpGetField: {"OpGetField", []int{2}},   // 2-byte: field name constant index
OpSetField: {"OpSetField", []int{2}},   // 2-byte: field name constant index
OpSuper:    {"OpSuper", []int{2}},      // 2-byte: method name constant index
```

### Compilation Example

Class definition:
```xxl
class Person {
    var name = ""

    func init(name) {
        this.name = name
    }
}
```

Compiles to:
```
OpConstant 0      ; "Person" class name
OpConstant 1      ; {"name": ""} default fields
OpConstant 2      ; init method
OpClass 0         ; create class, store in global
```

Instance creation:
```xxl
var p = new Person("Alice")
```

Compiles to:
```
OpGetGlobal 0     ; get Person class
OpConstant 3      ; "Alice"
OpNew 1           ; create instance, call init
OpSetGlobal 1     ; store in p
```

## VM Execution

### VM Structure Update

```go
type VM struct {
    // ...existing fields...
    currentInstance *Instance  // current instance for method execution (for this)
}
```

### Opcode Execution

#### OpClass
```
1. Pop fields map from stack
2. Pop methods map from stack
3. Pop superclass (or null) from stack
4. Create Class object
5. Push class onto stack
```

#### OpNew
```
1. Pop class from stack
2. Pop arguments from stack
3. Copy default fields from class hierarchy
4. Create Instance object
5. Call init method if exists
6. Push instance onto stack
```

#### OpGetField
```
1. Pop instance from stack
2. Get field value from instance.Fields
3. Push value onto stack (null if not found)
```

#### OpSetField
```
1. Pop value from stack
2. Pop instance from stack
3. Set instance.Fields[name] = value
4. Push value onto stack
```

#### OpSuper
```
1. Get current instance from VM context
2. Find method in superclass chain
3. Push method onto stack
```

### `this` Handling

- When calling a method, VM records `currentInstance`
- `this.name` compiles to `OpGetField`
- `this.name = x` compiles to `OpSetField`

## Error Handling

### Error Types

```go
// Class definition errors
"class 'Person' already defined"
"'extends' requires a valid class name"
"class cannot extend itself"

// Instance creation errors
"class 'Person' has no 'init' method"
"wrong number of arguments for init, expected 2, got 1"
"cannot use 'new' on non-class type"

// Field access errors
"cannot access field 'name' on null"
"field 'age' not found on instance of 'Person'"

// Method call errors
"method 'speak' not found on class 'Person'"
"cannot call 'super.init' outside of class method"
"no superclass for 'Animal'"
```

### Edge Cases

| Case | Behavior |
|------|----------|
| `new` on class without init | Create instance, no constructor call |
| Access non-existent field | Return `null` |
| `this` outside class method | Compile error |
| `super` in class without parent | Runtime error |
| Circular inheritance | Compile-time detection, error |
| `new` on non-class type | Runtime error |

### Field vs Method Same Name

```xxl
class Person {
    var name = ""
    func name() { return this.name }  // Allowed, method takes priority
}
var p = new Person()
p.name()        // Call method
p.name          // Access field (if method name is shadowed)
```

## Test Cases

### Basic Class

```xxl
class Point {
    var x = 0
    var y = 0

    func init(x, y) {
        this.x = x
        this.y = y
    }
}
var p = new Point(3, 4)
p.x + p.y  // => 7
```

### Method Call and `this`

```xxl
class Counter {
    var count = 0

    func inc() {
        this.count = this.count + 1
    }
}
var c = new Counter()
c.inc()
c.inc()
c.count  // => 2
```

### Inheritance

```xxl
class Animal {
    var name = ""
    func init(name) { this.name = name }
    func speak() { return this.name }
}
class Dog extends Animal {
    func speak() { return this.name + " barks" }
}
var d = new Dog("Buddy")
d.speak()  // => "Buddy barks"
```

### Super Call

```xxl
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
d.name + " " + d.breed  // => "Buddy Golden"
```

### Method Lookup Chain

```xxl
class A { func foo() { return "A" } }
class B extends A { }
class C extends B { }
var c = new C()
c.foo()  // => "A"
```

## Implementation Priority

1. **Lexer**: Already complete (class, extends, new, this keywords defined)
2. **Parser AST**: Already complete (ClassStatement, NewExpression, ThisExpression defined)
3. **Parser**: Add class parsing logic
4. **Objects**: Add Class and Instance types
5. **Opcodes**: Add OpClass, OpNew, OpGetField, OpSetField, OpSuper
6. **Compiler**: Add class statement compilation
7. **VM**: Add class/instance execution logic
8. **Tests**: Add comprehensive unit tests
