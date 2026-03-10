// tests/oop_test.go
package tests

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// Helper function
func runOOPCode(t *testing.T, input string) string {
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

// ============================================
// Basic Class Tests
// ============================================

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

	result := runOOPCode(t, input)
	if result != "7" {
		t.Errorf("expected 7, got %s", result)
	}
}

func TestOOPClassWithNoConstructor(t *testing.T) {
	input := `
		class Point {
			var x = 10
			var y = 20
		}

		var p = new Point()
		p.x + p.y
	`

	result := runOOPCode(t, input)
	if result != "30" {
		t.Errorf("expected 30, got %s", result)
	}
}

func TestOOPClassMultipleInstances(t *testing.T) {
	input := `
		class Box {
			var value = 0

			func init(v) {
				this.value = v
			}
		}

		var b1 = new Box(10)
		var b2 = new Box(20)
		b1.value + b2.value
	`

	result := runOOPCode(t, input)
	if result != "30" {
		t.Errorf("expected 30, got %s", result)
	}
}

// ============================================
// Inheritance Tests
// ============================================

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

	result := runOOPCode(t, input)
	if result != "Buddy barks" {
		t.Errorf("expected 'Buddy barks', got %s", result)
	}
}

func TestOOPSuperCall(t *testing.T) {
	input := `
		class Animal {
			var name = ""
			func init(n) { this.name = n }
		}

		class Dog extends Animal {
			var breed = ""
			func init(n, b) {
				super.init(n)
				this.breed = b
			}
		}

		var d = new Dog("Buddy", "Golden")
		d.name
	`

	result := runOOPCode(t, input)
	if result != "Buddy" {
		t.Errorf("expected 'Buddy', got %s", result)
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

	result := runOOPCode(t, input)
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
		d.name
	`

	result := runOOPCode(t, input)
	if result != "unknown" {
		t.Errorf("expected 'unknown', got %s", result)
	}
}
