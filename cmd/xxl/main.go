package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/jit"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/stdlib"
	"github.com/topxeq/xxlang/pkg/vm"
)

// JITEnabled controls whether JIT compilation is enabled
// Default is false (interpreter mode)
var useJIT = false

// JITConfig holds JIT compiler settings
var jitHotThreshold = 100
var jitMaxCodeSize = 4096
var jitDebug = false

// DebugMode controls whether comprehensive debug output is enabled
var debugMode = false

const (
	PROMPT          = ">> "
	CONTINUE_PROMPT = ".. "
)

// Version is set via -ldflags at build time. Default is "dev" for local builds.
var Version = "dev"

// REPL represents an interactive REPL session
type REPL struct {
	symbolTable *compiler.SymbolTable
	constants   []objects.Object
	globals     []objects.Object
	history     []string
}

// NewREPL creates a new REPL session
func NewREPL() *REPL {
	st := compiler.NewSymbolTable()

	// Define preset global variables
	argsGSymbol := st.Define("argsG")
	scriptPathGSymbol := st.Define("scriptPathG")

	// Initialize globals array
	globals := make([]objects.Object, compiler.GlobalsSize)

	// Set argsG - command line arguments as string array
	argsElements := make([]objects.Object, len(os.Args))
	for i, arg := range os.Args {
		argsElements[i] = &objects.String{Value: arg}
	}
	globals[argsGSymbol.Index] = &objects.Array{Elements: argsElements}

	// Set scriptPathG - empty for REPL mode
	globals[scriptPathGSymbol.Index] = &objects.String{Value: ""}

	return &REPL{
		symbolTable: st,
		constants:   make([]objects.Object, 0),
		globals:     globals,
		history:     make([]string, 0),
	}
}

// splitArgs splits command line arguments at the -- separator.
// Returns (interpreterArgs, scriptArgs).
// Everything after -- is passed to the script.
func splitArgs(args []string) (interpreterArgs []string, scriptArgs []string) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

func main() {
	// No arguments - start REPL
	if len(os.Args) < 2 {
		startREPL()
		return
	}

	// Split arguments at -- separator
	interpreterArgs, scriptArgs := splitArgs(os.Args[1:])

	// Parse JIT flags
	interpreterArgs = parseFlags(interpreterArgs)

	// Set script args for scripts to access via env::scriptArgs()
	stdlib.SetScriptArgs(scriptArgs)

	// Check for help or version flags (only in interpreter args)
	for _, arg := range interpreterArgs {
		if arg == "--help" || arg == "-h" || arg == "help" {
			printUsage()
			return
		}
		if arg == "--version" || arg == "-v" || arg == "version" {
			printVersion()
			return
		}
	}

	// Check for -cloud flag
	if len(interpreterArgs) >= 2 && interpreterArgs[0] == "-cloud" {
		runFromCloud(interpreterArgs[1])
		return
	}

	switch interpreterArgs[0] {
	case "compile":
		if err := compileCmd(interpreterArgs[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "run":
		if len(interpreterArgs) < 2 {
			fmt.Println("Usage: xxl run <file|url> [-- script args...]")
			os.Exit(1)
		}
		runFileOrURL(interpreterArgs[1])
	case "update":
		if err := updateCmd(interpreterArgs[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		printVersion()
	case "help":
		printUsage()
	default:
		// Check if it's a file or URL that we can run directly
		arg := interpreterArgs[0]
		if isURL(arg) || strings.HasSuffix(arg, ".xxl") || strings.HasSuffix(arg, ".xxb") {
			runFileOrURL(arg)
		} else {
			// Unknown subcommand - show error
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", arg)
			printUsage()
			os.Exit(1)
		}
	}
}

// parseFlags extracts and processes JIT-related and debug flags
// Returns the remaining arguments
func parseFlags(args []string) []string {
	var result []string
	for _, arg := range args {
		if arg == "--jit" {
			useJIT = true
		} else if arg == "--jit-debug" {
			useJIT = true
			jitDebug = true
		} else if strings.HasPrefix(arg, "--jit-threshold=") {
			threshold := strings.TrimPrefix(arg, "--jit-threshold=")
			fmt.Sscanf(threshold, "%d", &jitHotThreshold)
			useJIT = true
		} else if arg == "--no-jit" {
			useJIT = false
		} else if arg == "--debug" {
			debugMode = true
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func printUsage() {
	fmt.Printf("Xxlang v%s\n", Version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  xxl                         Start interactive REPL")
	fmt.Println("  xxl <file|url> [-- args...] Execute a .xxl file or script from URL")
	fmt.Println("  xxl run <file|url> [-- args...]")
	fmt.Println("                              Execute a .xxl file or script from URL")
	fmt.Println("  xxl -cloud <script>         Execute script from cloud URL base")
	fmt.Println("  xxl compile <file>          Compile file to bytecode")
	fmt.Println("  xxl update                  Self-update to latest version from GitHub")
	fmt.Println("  xxl version                 Print version information")
	fmt.Println("  xxl help                    Print this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o, --output path     Output path for compiled file")
	fmt.Println("      --target os/arch  Cross-compile for target OS/architecture")
	fmt.Println("      --bytecode        Output as bytecode (.xxb) instead of executable")
	fmt.Println("  -cloud <script>       Execute script from configured cloudUrlBase")
	fmt.Println("      --jit             Enable JIT compilation for hot paths (experimental)")
	fmt.Println("      --jit-threshold=N Set JIT hot path threshold (default: 100)")
	fmt.Println("      --no-jit          Disable JIT compilation (default)")
	fmt.Println("      --debug           Show debug info (bytecode count, runtime, JIT usage)")
	fmt.Println()
	fmt.Println("Script Arguments:")
	fmt.Println("  Use '--' to separate interpreter arguments from script arguments.")
	fmt.Println("  Script arguments are accessible via env::scriptArgs().")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  xxl")
	fmt.Println("  xxl script.xxl")
	fmt.Println("  xxl script.xxl -- arg1 arg2 --help")
	fmt.Println("  xxl run script.xxl -- --verbose -f file.txt")
	fmt.Println("  xxl https://raw.githubusercontent.com/user/repo/main/script.xxl")
	fmt.Println("  xxl compile -o program script.xxl")
	fmt.Println("  xxl compile -o program.exe --target windows/amd64 script.xxl")
	fmt.Println("  xxl compile --bytecode script.xxl")
	fmt.Println("  xxl -cloud basic.xxl")
	fmt.Println("  xxl update")
}

func printVersion() {
	fmt.Printf("Xxlang v%s\n", Version)
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

	return r.executeRegister(program)
}

// executeRegister uses the register-based VM for REPL execution
func (r *REPL) executeRegister(program *parser.Program) (objects.Object, error) {
	// Compilation with register compiler
	c := compiler.NewRegCompiler()
	c.SetSymbolTable(r.symbolTable)
	c.SetConstants(r.constants)
	if _, err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compiler error: %v", err)
	}

	// Update constants
	r.constants = c.Bytecode().Constants

	// Execution with persistent globals
	bytecode := c.Bytecode()
	v := vm.NewRegVMWithObjectGlobals(bytecode, r.globals)

	if err := v.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}

	// Update globals for next execution
	r.globals = v.GlobalsAsObjects()

	return v.LastPoppedObject(), nil
}

// isURL checks if the given string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// runFromCloud fetches and executes a script from the cloud URL base configured in settings
func runFromCloud(scriptName string) {
	// Get config map
	cfg := stdlib.GetConfigMap()

	// Get cloudUrlBase from config
	cloudUrlBase, ok := cfg["cloudUrlBase"].(string)
	if !ok || cloudUrlBase == "" {
		fmt.Fprintln(os.Stderr, "Error: cloudUrlBase not configured in settings.json")
		fmt.Fprintln(os.Stderr, "Please add a 'cloudUrlBase' field to your ~/.xxl/settings.json file")
		os.Exit(1)
	}

	// Construct full URL
	// Ensure the base URL ends with / and the script name is appended
	baseURL := strings.TrimSuffix(cloudUrlBase, "/")
	fullURL := baseURL + "/" + scriptName

	// Run from the constructed URL
	runFromURL(fullURL)
}

// runFileOrURL runs an xxlang source file or script from URL
func runFileOrURL(source string) {
	if isURL(source) {
		runFromURL(source)
	} else {
		runFile(source)
	}
}

// runFromURL fetches and executes an xxlang script from a URL
func runFromURL(url string) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the script
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching URL '%s': %v\n", url, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: HTTP status %d (%s)\n", resp.StatusCode, resp.Status)
		os.Exit(1)
	}

	// Read the response body
	code, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Execute the code
	executeCode(string(code), url)
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

	// Execute the code
	executeCode(string(code), absPath)
}

// executeCode compiles and executes xxlang source code
func executeCode(code, sourcePath string) {
	// Lexical analysis
	l := lexer.New(code)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		fmt.Printf("Error: %v\n", formatParserErrors(p.Errors()))
		os.Exit(1)
	}

	executeCodeRegister(program, sourcePath, code)
}

// executeCodeRegister uses the register-based VM
func executeCodeRegister(program *parser.Program, sourcePath, code string) {
	// Compilation
	c := compiler.NewRegCompiler()
	// Define preset global variables before compilation
	argsGSymbol := c.SymbolTable().Define("argsG")
	scriptPathGSymbol := c.SymbolTable().Define("scriptPathG")
	c.SetSourceFile(sourcePath)

	compileStart := time.Now()
	if _, err := c.Compile(program); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	compileTime := time.Since(compileStart)

	// Create main module for exports
	mainModule := &objects.Module{
		Name:    sourcePath,
		Exports: make(map[string]objects.Object),
	}

	// Prepare globals array with preset values
	globals := make([]vm.Value, compiler.GlobalsSize)

	// Set argsG - command line arguments as string array
	argsElements := make([]objects.Object, len(os.Args))
	for i, arg := range os.Args {
		argsElements[i] = &objects.String{Value: arg}
	}
	globals[argsGSymbol.Index] = vm.NewObject(&objects.Array{Elements: argsElements})

	// Set scriptPathG - script path
	globals[scriptPathGSymbol.Index] = vm.NewObject(&objects.String{Value: sourcePath})

	// Execution
	bytecode := c.Bytecode()

	// Print debug info before execution
	if debugMode {
		fmt.Fprintf(os.Stderr, "[Debug] Source: %s\n", sourcePath)
		fmt.Fprintf(os.Stderr, "[Debug] Source size: %d bytes\n", len(code))
		fmt.Fprintf(os.Stderr, "[Debug] Bytecode instructions: %d\n", len(bytecode.Instructions))
		fmt.Fprintf(os.Stderr, "[Debug] Constants: %d\n", len(bytecode.Constants))
		fmt.Fprintf(os.Stderr, "[Debug] Compile time: %v\n", compileTime)
		fmt.Fprintf(os.Stderr, "[Debug] JIT enabled: %v\n", useJIT)
	}

	execStart := time.Now()

	// Use JIT VM if enabled
	if useJIT {
		jitConfig := jit.JITConfig{
			HotThreshold: jitHotThreshold,
			MaxCodeSize:  jitMaxCodeSize,
			Debug:        jitDebug,
		}
		jitVM := jit.NewJITVMWithGlobals(bytecode, globals, jitConfig)
		jitVM.SetSourcePath(sourcePath)
		jitVM.SetCurrentModule(mainModule)

		if err := jitVM.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n%s", jitVM.GetCallStack())
			os.Exit(1)
		}
		execTime := time.Since(execStart)

		// Print JIT stats
		stats := jitVM.GetJITStats()
		nativeExecs, interpExecs := jitVM.GetNativeStats()

		if debugMode {
			fmt.Fprintf(os.Stderr, "[Debug] VM mode: JIT (hybrid)\n")
			fmt.Fprintf(os.Stderr, "[Debug] Execution time: %v\n", execTime)
			fmt.Fprintf(os.Stderr, "[Debug] JIT compiled functions: %d\n", stats.CompiledFunctions)
			fmt.Fprintf(os.Stderr, "[Debug] JIT total code size: %d bytes\n", stats.TotalCodeSize)
			fmt.Fprintf(os.Stderr, "[Debug] Native executions: %d\n", nativeExecs)
			fmt.Fprintf(os.Stderr, "[Debug] Interpreter executions: %d\n", interpExecs)
			fmt.Fprintf(os.Stderr, "[Debug] Total time: %v\n", compileTime+execTime)
		} else if stats.CompiledFunctions > 0 {
			fmt.Fprintf(os.Stderr, "[JIT] Compiled %d functions, %d bytes\n", stats.CompiledFunctions, stats.TotalCodeSize)
		}

		// Print result if it's meaningful
		result := jitVM.LastPoppedObject()
		if result != nil && result != objects.NULL {
			fmt.Println(result.Inspect())
		}

		jitVM.Cleanup()
		return
	}

	// Standard register VM execution
	v := vm.NewRegVMWithGlobals(bytecode, globals)
	v.SetSourcePath(sourcePath)
	v.SetCurrentModule(mainModule)

	if err := v.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\n%s", v.GetCallStack())
		os.Exit(1)
	}
	execTime := time.Since(execStart)

	if debugMode {
		fmt.Fprintf(os.Stderr, "[Debug] VM mode: Interpreter\n")
		fmt.Fprintf(os.Stderr, "[Debug] Execution time: %v\n", execTime)
		fmt.Fprintf(os.Stderr, "[Debug] Total time: %v\n", compileTime+execTime)
	}

	// Print result if it's meaningful
	result := v.LastPoppedObject()
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

	// Prepare globals array with preset values
	globals := make([]vm.Value, compiler.GlobalsSize)

	// Set argsG - command line arguments as string array (index 0)
	argsElements := make([]objects.Object, len(os.Args))
	for i, arg := range os.Args {
		argsElements[i] = &objects.String{Value: arg}
	}
	globals[0] = vm.NewObject(&objects.Array{Elements: argsElements})

	// Set scriptPathG - script path (index 1)
	globals[1] = vm.NewObject(&objects.String{Value: filename})

	// Execution
	v := vm.NewRegVMWithGlobals(bytecode, globals)
	v.SetSourcePath(filename)
	v.SetCurrentModule(mainModule)

	if err := v.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print result if it's meaningful
	result := v.LastPoppedObject()
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
