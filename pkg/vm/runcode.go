// pkg/vm/runcode.go
// Support for dynamic code execution via runCode builtin
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// RunCodeFunc is the signature for the runCode callback
type RunCodeFunc func(code string, args *objects.Map) (objects.Object, error)

// runCodeCallback is set by the VM when executing
var runCodeCallback RunCodeFunc

// SetRunCodeCallback registers the callback for runCode builtin
func SetRunCodeCallback(fn RunCodeFunc) {
	runCodeCallback = fn
}

// GetRunCodeCallback returns the current callback
func GetRunCodeCallback() RunCodeFunc {
	return runCodeCallback
}

// ExecuteRunCode is called by the runCode builtin
func ExecuteRunCode(code string, args *objects.Map) (objects.Object, error) {
	if runCodeCallback == nil {
		return nil, fmt.Errorf("runCode not available in this context")
	}
	return runCodeCallback(code, args)
}

// RunCodeInVM executes code in the context of the given VM
func RunCodeInVM(code string, args *objects.Map, vm *VM) (objects.Object, error) {
	// Lexical analysis
	l := lexer.New(code)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse error: %s", p.Errors()[0])
	}

	// Create a new compiler
	c := compiler.New()

	// If args are provided, define them as globals before compilation
	// This allows the code to reference these variables
	if args != nil && len(args.Pairs) > 0 {
		for _, pair := range args.Pairs {
			key, ok := pair.Key.(*objects.String)
			if !ok {
				return nil, fmt.Errorf("argument keys must be strings")
			}
			c.DefineGlobal(key.Value)
		}
	}

	if err := c.Compile(program); err != nil {
		return nil, fmt.Errorf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	// Create a new globals array for this execution
	newGlobals := make([]objects.Object, compiler.GlobalsSize)

	// Set argument values in globals
	if args != nil {
		for _, pair := range args.Pairs {
			key := pair.Key.(*objects.String)
			// Find the symbol to get its index
			symbol, ok := c.ResolveSymbol(key.Value)
			if ok && symbol.Scope == compiler.GlobalScope {
				newGlobals[symbol.Index] = pair.Value
			}
		}
	}

	// Create a new VM with the prepared globals
	newVM := NewWithGlobalsStore(bytecode, newGlobals)

	// Share the loader if available
	if vm.loader != nil {
		newVM.SetLoader(vm.loader)
	}

	if err := newVM.Run(); err != nil {
		return nil, fmt.Errorf("runtime error: %v", err)
	}

	return newVM.LastPopped(), nil
}
