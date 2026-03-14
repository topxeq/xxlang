package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

const (
	PROMPT          = ">> "
	CONTINUE_PROMPT = ".. "
)

const (
	Version   = "0.3.0"
)

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
	// No arguments - start REPL
	if len(os.Args) < 2 {
		startREPL()
		return
	}

	// Check for help or version flags
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" || arg == "help" {
			printUsage()
			return
		}
		if arg == "--version" || arg == "-v" || arg == "version" {
			printVersion()
			return
		}
	}

	switch os.Args[1] {
	case "compile":
		if err := compileCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: xx run <file>")
			os.Exit(1)
		}
		runFile(os.Args[2])
	case "version":
		printVersion()
	case "help":
		printUsage()
	default:
		// Unknown subcommand - show error
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
    fmt.Printf("Xxlang v%s\n", Version)
    fmt.Println()
    fmt.Println("Usage:")
    fmt.Println("  xx                    Start interactive REPL")
    fmt.Println("  xx run <file>         Execute a .xxl file")
    fmt.Println("  xx compile <file>     Compile file to bytecode")
    fmt.Println("  xx version            Print version information")
    fmt.Println("  xx help               Print this help message")
    fmt.Println()
    fmt.Println("Options:")
    fmt.Println("  -o, --output path     Output path for compiled file")
    fmt.Println("      --target os/arch  Cross-compile for target OS/architecture")
    fmt.Println("      --bytecode        Output as bytecode (.xxb) instead of executable")
    fmt.Println()
    fmt.Println("Examples:")
    fmt.Println("  xx")
    fmt.Println("  xx run script.xxl")
    fmt.Println("  xx compile -o program script.xxl")
    fmt.Println("  xx compile -o program.exe --target windows/amd64 script.xxl")
    fmt.Println("  xx compile --bytecode script.xxl")
}

func printVersion() {
    fmt.Printf("Xxlang v%s (Go %s)\n", Version, runtime.Version())
}

func startREPL() {
    repl := NewREPL()
    scanner := bufio.NewScanner(os.Stdin)

    fmt.Println("Xxlang REPL v" + Version)
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

// runFile runs an xxlang source file
func runFile(filename string) {
    // Get absolute path for module resolution
    absPath, err := filepath.Abs(filename)
    if err != nil {
        fmt.Printf("Error resolving path '%s': %v\n", filename, err)
        os.Exit(1)
    }

    // Check if it's a bytecode file
    if strings.HasSuffix(absPath, ".xxb") {
        runBytecodeFile(absPath)
        return
    }

    // Read the source file
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
    c.SetSource(absPath, string(code)) // Enable source mapping
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
        // Format runtime error with source location and call stack
        fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
        fmt.Fprintf(os.Stderr, "\n%s", v.GetCallStack())
        os.Exit(1)
    }

    // Print result if it's meaningful
    result := v.LastPopped()
    if result != nil && result != objects.NULL {
        fmt.Println(result.Inspect())
    }
}

// runBytecodeFile runs a compiled bytecode file
func runBytecodeFile(filename string) {
    // Read the bytecode file
    bytecodeData, err := os.ReadFile(filename)
    if err != nil {
        fmt.Printf("Error reading bytecode file '%s': %v\n", filename, err)
        os.Exit(1)
    }

    // Deserialize the bytecode
    bytecode, err := compiler.Deserialize(bytecodeData)
    if err != nil {
        fmt.Printf("Error loading bytecode: %v\n", err)
        os.Exit(1)
    }

    // Create main module for exports
    mainModule := &objects.Module{
        Name:    filename,
        Exports: make(map[string]objects.Object),
    }

    // Execution
    v := vm.NewWithGlobalsStore(bytecode, make([]objects.Object, compiler.GlobalsSize))
    v.SetSourcePath(filename)
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
