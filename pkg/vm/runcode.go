// pkg/vm/runcode.go
// Support for dynamic code execution via runCode builtin
package vm

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/objects"
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
