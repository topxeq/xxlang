package main

import (
	"fmt"
	"os"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: disasm <file.xxl>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Printf("Parser error: %s\n", e)
		}
		os.Exit(1)
	}

	c := compiler.NewRegCompiler()
	_, err = c.Compile(program)
	if err != nil {
		fmt.Printf("Compile error: %v\n", err)
		os.Exit(1)
	}

	bytecode := c.Bytecode()
	defs := compiler.GetDefinitions()
	code := bytecode.Instructions
	fmt.Printf("Total instructions: %d bytes\n", len(code))
	fmt.Printf("Constants: %d\n", len(bytecode.Constants))

	ip := 0
	for ip < len(code) {
		op := compiler.Opcode(code[ip])
		def, ok := defs[op]
		if !ok {
			fmt.Printf("%04d: UNKNOWN opcode %d\n", ip, op)
			ip++
			continue
		}

		operands, read := compiler.ReadOperands(def, code[ip+1:])
		fmt.Printf("%04d: %s", ip, def.Name)
		for _, operand := range operands {
			fmt.Printf(" %d", operand)
		}

		if op == compiler.OpConstant || op == compiler.OpRegLoadConst {
			if len(operands) > 0 && int(operands[0]) < len(bytecode.Constants) {
				fmt.Printf("  ; const[%d] = %s", operands[0], bytecode.Constants[operands[0]].Inspect())
			}
		}

		fmt.Println()
		ip += 1 + read
	}
}
