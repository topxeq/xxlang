// +build !amd64

// pkg/jit/jit_vm_stub.go
// Stub implementation of JITVM for non-amd64 platforms
// JIT compilation is not supported on non-amd64 platforms
// This provides a compatible API that just uses the regular VM

package jit

import (
	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/module"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/vm"
)

// JITVM wraps a register VM (stub for non-amd64)
type JITVM struct {
	*vm.RegVM
	config JITConfig
}

// NewJITVM creates a new JIT-enabled VM (stub - just uses regular VM)
func NewJITVM(bytecode *compiler.Bytecode, config JITConfig) *JITVM {
	return &JITVM{
		RegVM:  vm.NewRegVM(bytecode),
		config: config,
	}
}

// NewJITVMWithGlobals creates a JIT VM with custom globals (stub)
func NewJITVMWithGlobals(bytecode *compiler.Bytecode, globals []vm.Value, config JITConfig) *JITVM {
	return &JITVM{
		RegVM:  vm.NewRegVMWithGlobals(bytecode, globals),
		config: config,
	}
}

// NewJITVMWithObjectGlobals creates a JIT VM with globals as objects.Object (stub)
func NewJITVMWithObjectGlobals(bytecode *compiler.Bytecode, globals []objects.Object, config JITConfig) *JITVM {
	valueGlobals := make([]vm.Value, len(globals))
	for i, obj := range globals {
		valueGlobals[i] = vm.NewObject(obj)
	}
	return NewJITVMWithGlobals(bytecode, valueGlobals, config)
}

// Run executes the bytecode (stub - just uses interpreter)
func (j *JITVM) Run() error {
	return j.RegVM.Run()
}

// SetJITEnabled enables or disables JIT compilation (stub - no-op)
func (j *JITVM) SetJITEnabled(enabled bool) {
	// No-op on non-amd64
}

// GetJITStats returns JIT compilation statistics (stub)
func (j *JITVM) GetJITStats() JITStats {
	return JITStats{}
}

// GetNativeStats returns native execution statistics (stub)
func (j *JITVM) GetNativeStats() (nativeExecs, interpExecs int64) {
	return 0, 0
}

// Cleanup releases JIT resources (stub - no-op)
func (j *JITVM) Cleanup() {
	// No-op on non-amd64
}

// SetSourcePath sets the source path for error messages
func (j *JITVM) SetSourcePath(path string) {
	j.RegVM.SetSourcePath(path)
}

// SetCurrentModule sets the current module context
func (j *JITVM) SetCurrentModule(module *objects.Module) {
	j.RegVM.SetCurrentModule(module)
}

// LastPoppedObject returns the last popped object
func (j *JITVM) LastPoppedObject() objects.Object {
	return j.RegVM.LastPoppedObject()
}

// SetLoader sets the module loader
func (j *JITVM) SetLoader(loader *module.Loader) {
	j.RegVM.SetLoader(loader)
}

// GlobalsAsObjects returns globals as objects
func (j *JITVM) GlobalsAsObjects() []objects.Object {
	return j.RegVM.GlobalsAsObjects()
}
