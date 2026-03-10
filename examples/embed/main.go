// examples/embed/main.go
// Embedding example - demonstrates how to use xxlang as a Go library
package main

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
	fmt.Println("=== Xxlang Embedding Example ===")
	fmt.Println()

	// Create an interpreter with stdlib enabled
	interp := interpreter.New(interpreter.WithStdlib())

	// Simple evaluation
	fmt.Println("1. Simple evaluation:")
	result, err := interp.Eval("2 + 2 * 3")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("   2 + 2 * 3 = %s\n", result.Inspect())
	fmt.Println()

	// Pass values from Go
	fmt.Println("2. Pass values from Go:")
	err = interp.SetGlobal("x", 42)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	result, err = interp.Eval("println(x)")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println()

	// Get values back
	fmt.Println("3. Get values back:")
	_, err = interp.Eval("var result = x * 2")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	val, ok := interp.GetGlobal("result")
	if ok {
		fmt.Printf("   result = %s\n", val.Inspect())
	}
	fmt.Println()

	// Convert to Go types
	fmt.Println("4. Type conversion:")
	_, err = interp.Eval("var arr = [1, 2, 3, 4, 5]")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if arrVal, ok := interp.GetGlobalAs("arr"); ok {
		fmt.Printf("   arr (Go) = %v\n", arrVal)
	}
	fmt.Println()

	// Using functions
	fmt.Println("5. Define and call functions:")
	_, err = interp.Eval("func greet(name) { return \"Hello, \" + name + \"!\" }")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	result, err = interp.Eval("greet(\"World\")")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("   greet(\"World\") = %s\n", result.Inspect())
	fmt.Println()

	fmt.Println("=== Embedding Example Complete ===")
}
