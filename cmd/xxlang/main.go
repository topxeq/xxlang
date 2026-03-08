// cmd/xxlang/main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

const PROMPT = ">> "
const CONTINUE_PROMPT = ".. "

// REPL represents an interactive REPL session
type REPL struct {
	symbolTable *compiler.SymbolTable
	constants   []objects.Object
	globals     []objects.Object
	history     []string
}

// NewREPL creates a new REPL session
func NewREPL() *REPL {
	return &REPL{
		symbolTable: compiler.NewSymbolTable(),
		constants:   make([]objects.Object, 0),
		globals:     make([]objects.Object, compiler.GlobalsSize),
		history:     make([]string, 0),
	}
}

func main() {
	// Command-line flags
	flag.Parse()

	// Check if running a file
	args := flag.Args()
	if len(args) > 0 {
		runFile(args[0])
		return
	}

	// Start interactive REPL
	startREPL()
}

func startREPL() {
	repl := NewREPL()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Xxlang REPL v0.2")
	fmt.Println("Type 'exit' or 'quit' to exit, 'help' for help, 'history' for command history")
	fmt.Println("Multi-line: end line with '{' to continue")
	fmt.Print(PROMPT)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			fmt.Print(PROMPT)
			continue
		}

		// Handle special commands
		switch line {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return
		case "help":
			printHelp()
			fmt.Print(PROMPT)
			continue
		case "history":
			printHistory(repl.history)
			fmt.Print(PROMPT)
			continue
		case "clear":
			repl = NewREPL()
			fmt.Println("Cleared all variables and functions")
			fmt.Print(PROMPT)
			continue
		}

		// Check for multi-line input
		input := line
		for countOpenBraces(input) > 0 {
			fmt.Print(CONTINUE_PROMPT)
			if !scanner.Scan() {
				break
			}
			nextLine := scanner.Text()
			input += "\n" + nextLine
		}

		// Add to history
		repl.history = append(repl.history, input)

		// Parse and execute the input
		result, err := repl.Execute(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if result != nil && result != objects.NULL {
			fmt.Println(result.Inspect())
		}

		fmt.Print(PROMPT)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

// countOpenBraces counts unclosed braces in the input
func countOpenBraces(input string) int {
	count := 0
	inString := false
	escape := false

	for _, ch := range input {
		if escape {
			escape = false
			continue
		}
		switch ch {
		case '\\':
			escape = true
		case '"':
			inString = !inString
		case '{':
			if !inString {
				count++
			}
		case '}':
			if !inString {
				count--
			}
		}
	}

	return count
}

func printHelp() {
	fmt.Print(`Xxlang REPL Commands:
  exit, quit  - Exit the REPL
  help        - Show this help message
  history     - Show command history
  clear       - Clear all variables and functions

Language Features:
  Variables:    var x = 10
  Constants:   const PI = 3.14
  Functions:   func add(a, b) { return a + b }
  Closures:    func makeCounter() { var c = 0; func() { c++ }; }
  Arrays:      [1, 2, 3]
  Maps:        {"key": "value"}
  Loops:       for (var i = 0; i < 10; i++) { }
  Conditionals: if (x > 0) { } else { }

Examples:
  >> var x = 10
  10
  >> x + 5
  15
  >> func fib(n) { if (n <= 1) { return n; } return fib(n-1) + fib(n-2); }
  >> fib(10)
  55

`)
}

func printHistory(history []string) {
	if len(history) == 0 {
		fmt.Println("No commands in history")
		return
	}
	for i, cmd := range history {
		// Truncate long commands for display
		display := cmd
		if len(display) > 60 {
			display = display[:57] + "..."
		}
		// Replace newlines with escaped version
		display = strings.ReplaceAll(display, "\n", "\\n")
		fmt.Printf("%3d: %s\n", i+1, display)
	}
}

func runFile(filename string) {
	// Get absolute path for module resolution
	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Printf("Error resolving path '%s': %v\n", filename, err)
		os.Exit(1)
	}

	// Read the file
	code, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Printf("Error reading file '%s': %v\n", filename, err)
		os.Exit(1)
	}

	// Lexical analysis
	l := lexer.New(string(code))

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		fmt.Printf("Error: %v\n", formatParserErrors(p.Errors()))
		os.Exit(1)
	}

	// Compilation
	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create main module for exports
	mainModule := &objects.Module{
		Name:    absPath,
		Exports: make(map[string]objects.Object),
	}

	// Execution
	bytecode := c.Bytecode()
	v := vm.NewWithGlobalsStore(bytecode, make([]objects.Object, compiler.GlobalsSize))
	v.SetSourcePath(absPath)
	v.SetCurrentModule(mainModule)

	if err := v.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print result if it's meaningful
	result := v.LastPopped()
	if result != nil && result != objects.NULL {
		fmt.Println(result.Inspect())
	}
}

// Execute compiles and runs code in the REPL context
func (r *REPL) Execute(input string) (objects.Object, error) {
	// Lexical analysis
	l := lexer.New(input)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		return nil, formatParserErrors(p.Errors())
	}

	// Compilation with persistent state
	c := compiler.NewWithState(r.symbolTable, r.constants)
	if err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compiler error: %v", err)
	}

	// Update constants
	r.constants = c.Bytecode().Constants

	// Execution with persistent globals
	bytecode := c.Bytecode()
	v := vm.NewWithGlobalsStore(bytecode, r.globals)

	if err := v.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}

	// Update globals for next execution
	r.globals = v.Globals()

	return v.LastPopped(), nil
}

func formatParserErrors(errors []string) error {
	var sb strings.Builder
	sb.WriteString("parser errors:\n")
	for _, err := range errors {
		sb.WriteString("  ")
		sb.WriteString(err)
		sb.WriteString("\n")
	}
	return fmt.Errorf("%s", sb.String())
}
