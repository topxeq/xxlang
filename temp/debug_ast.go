package main

import (
	"fmt"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

func main() {
	// Single-line closure
	input1 := `for (var i = 0; i < 3; i = i + 1) { funcs = push(funcs, func() { return i }) }`
	l1 := lexer.New(input1)
	p1 := parser.New(l1)
	prog1 := p1.ParseProgram()
	fmt.Println("=== Single-line closure ===")
	fmt.Printf("Errors: %v\n", p1.Errors())
	for i, stmt := range prog1.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
	}

	// Multi-line closure
	input2 := `for (var i = 0; i < 3; i = i + 1) { funcs = push(funcs, func() {
        return i
    }) }`
	l2 := lexer.New(input2)
	p2 := parser.New(l2)
	prog2 := p2.ParseProgram()
	fmt.Println("\n=== Multi-line closure ===")
	fmt.Printf("Errors: %v\n", p2.Errors())
	for i, stmt := range prog2.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
	}
}
