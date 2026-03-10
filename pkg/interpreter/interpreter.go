// pkg/interpreter/interpreter.go
// Clean embedding API for using xxlang as a Go library.
package interpreter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/stdlib"
	"github.com/topxeq/xxlang/pkg/vm"
)

// Interpreter wraps VM and Compiler for easy embedding.
// It provides a high-level API for evaluating xxlang code
// and passing values between Go and xxlang.
type Interpreter struct {
	symbolTable *compiler.SymbolTable
	constants   []objects.Object
	globals     []objects.Object
	loader      *module.Loader
	stdlib      bool
}

// New creates a new interpreter with the given options.
// Default interpreter has no stdlib and empty globals.
func New(opts ...Option) *Interpreter {
	i := &Interpreter{
		symbolTable: compiler.NewSymbolTable(),
		constants:   make([]objects.Object, 0),
		globals:     make([]objects.Object, compiler.GlobalsSize),
		loader:      module.NewLoader(),
		stdlib:      false,
	}

	for _, opt := range opts {
		opt(i)
	}

	return i
}

// Eval compiles and executes xxlang code, returning the result.
// This is the primary method for running xxlang code from Go.
func (i *Interpreter) Eval(code string) (objects.Object, error) {
	// Lexical analysis
	l := lexer.New(code)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		return nil, formatParserErrors(p.Errors())
	}

	// Compilation with persistent state
	c := compiler.NewWithState(i.symbolTable, i.constants)
	if err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compiler error: %v", err)
	}

	// Update constants for next execution
	i.constants = c.Bytecode().Constants

	// Execution with persistent globals
	bytecode := c.Bytecode()
	v := vm.NewWithGlobalsStore(bytecode, i.globals)
	v.SetLoader(i.loader)

	if err := v.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}

	// Update globals for next execution
	i.globals = v.Globals()

	return v.LastPopped(), nil
}

// EvalFile loads and executes an xxlang source file.
// The file path is used for module resolution and error messages.
func (i *Interpreter) EvalFile(path string) (objects.Object, error) {
	// Get absolute path for module resolution
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("error resolving path '%s': %v", path, err)
	}

	// Read the file
	code, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file '%s': %v", path, err)
	}

	// Lexical analysis
	l := lexer.New(string(code))

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		return nil, formatParserErrors(p.Errors())
	}

	// Compilation
	c := compiler.NewWithState(i.symbolTable, i.constants)
	if err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compiler error: %v", err)
	}

	// Update constants
	i.constants = c.Bytecode().Constants

	// Create main module for exports
	mainModule := &objects.Module{
		Name:    absPath,
		Exports: make(map[string]objects.Object),
	}

	// Execution
	bytecode := c.Bytecode()
	v := vm.NewWithGlobalsStore(bytecode, i.globals)
	v.SetLoader(i.loader)
	v.SetSourcePath(absPath)
	v.SetCurrentModule(mainModule)

	if err := v.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}

	// Update globals for next execution
	i.globals = v.Globals()

	return v.LastPopped(), nil
}

// SetGlobal sets a global variable in the interpreter.
// The value is converted from Go to xxlang using FromGo.
func (i *Interpreter) SetGlobal(name string, value interface{}) error {
	obj, err := FromGo(value)
	if err != nil {
		return fmt.Errorf("cannot convert value: %v", err)
	}

	// Define the symbol
	symbol := i.symbolTable.Define(name)

	// Ensure globals slice is large enough
	if symbol.Index >= len(i.globals) {
		return fmt.Errorf("global index %d out of bounds", symbol.Index)
	}

	i.globals[symbol.Index] = obj
	return nil
}

// GetGlobal retrieves a global variable from the interpreter.
// Returns the xxlang object directly (not converted to Go).
func (i *Interpreter) GetGlobal(name string) (objects.Object, bool) {
	symbol, ok := i.symbolTable.Resolve(name)
	if !ok {
		return nil, false
	}

	if symbol.Index >= len(i.globals) {
		return nil, false
	}

	obj := i.globals[symbol.Index]
	if obj == nil {
		return nil, false
	}

	return obj, true
}

// GetGlobalAs retrieves a global variable and converts it to a Go value.
// This is a convenience method that combines GetGlobal and ToGo.
func (i *Interpreter) GetGlobalAs(name string) (interface{}, bool) {
	obj, ok := i.GetGlobal(name)
	if !ok {
		return nil, false
	}
	return ToGo(obj), true
}

// Globals returns the interpreter's global variables slice.
// This is useful for advanced use cases where direct access is needed.
func (i *Interpreter) Globals() []objects.Object {
	return i.globals
}

// Loader returns the interpreter's module loader.
// This can be used to pre-load modules or check loaded modules.
func (i *Interpreter) Loader() *module.Loader {
	return i.loader
}

// Reset clears all state and creates a fresh interpreter.
// This is useful when you want to start over without creating a new instance.
func (i *Interpreter) Reset() {
	i.symbolTable = compiler.NewSymbolTable()
	i.constants = make([]objects.Object, 0)
	i.globals = make([]objects.Object, compiler.GlobalsSize)
	i.loader = module.NewLoader()
}

// formatParserErrors formats parser errors into a single error.
func formatParserErrors(errors []string) error {
	return fmt.Errorf("parser errors: %v", errors)
}

// Check stdlib imports at compile time
var _ = stdlib.Registry
