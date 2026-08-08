// pkg/vm/pool.go
// VM instance pool for script execution.
//
// The VMPool reuses RegVM instances to reduce GC pressure and memory churn
// compared to creating a new VM per request. Unlike a fixed-size pool, this
// implementation is unbounded: if no VM is available, a new one is created
// on demand. This provides reuse benefits without imposing any concurrency
// limit or backpressure.

package vm

import (
	"github.com/topxeq/xxlang/pkg/compiler"
)

// VMPool manages a pool of RegVM instances.
// The pool is unbounded: Acquire reuses an available VM if one exists,
// otherwise creates a new one. This avoids blocking callers while still
// reducing allocations under typical load.
type VMPool struct {
	pool     chan *RegVM
	newFunc  func(*compiler.Bytecode) *RegVM
	bytecode *compiler.Bytecode
	closed   bool
}

// NewVMPool creates a VM pool. VMs are created on demand using the provided
// factory function when none are available in the pool.
func NewVMPool(factory func(*compiler.Bytecode) *RegVM) *VMPool {
	// Use a small buffer to hold a few reusable VMs without blocking.
	// The pool is unbounded: if the buffer is empty, a new VM is created.
	return &VMPool{
		pool:    make(chan *RegVM, 4),
		newFunc: factory,
	}
}

// NewVMPoolWithBytecode creates a VM pool where all VMs share the given
// compiled bytecode. VMs are created on demand; the pool simply caches
// returned VMs for reuse.
func NewVMPoolWithBytecode(bytecode *compiler.Bytecode) *VMPool {
	if bytecode == nil {
		panic("VMPool: bytecode must not be nil")
	}

	p := &VMPool{
		pool:     make(chan *RegVM, 4),
		bytecode: bytecode,
	}

	return p
}

// Acquire retrieves a VM from the pool. If the pool has an available VM,
// it is returned immediately. If the pool is empty, a new VM is created
// using the factory function or shared bytecode.
//
// Acquire never blocks. It returns an error only if the pool is closed.
func (p *VMPool) Acquire() (*RegVM, error) {
	if p.closed {
		return nil, ErrPoolClosed
	}

	// Try to reuse an existing VM from the pool
	select {
	case vm := <-p.pool:
		return vm, nil
	default:
		// No VM available; create a new one
		if p.bytecode != nil {
			return NewRegVM(p.bytecode), nil
		}
		if p.newFunc != nil {
			return p.newFunc(nil), nil
		}
		return nil, ErrNoFactory
	}
}

// Release returns a VM to the pool after resetting it to a clean state.
// The VM's globals, frame stack, and pending exceptions are cleared so that
// the next request starts fresh.
//
// If the pool buffer is full, the VM is discarded (will be garbage collected).
// Releasing a nil VM is a no-op.
func (p *VMPool) Release(vm *RegVM) {
	if vm == nil {
		return
	}

	p.resetVM(vm)

	// Non-blocking send: if the pool buffer is full, discard the VM.
	select {
	case p.pool <- vm:
		// Successfully returned to pool for reuse
	default:
		// Pool buffer full; let GC collect the VM
	}
}

// resetVM clears VM state between requests to prevent state leakage.
func (p *VMPool) resetVM(vm *RegVM) {
	// Reset globals to nil/zero values
	for i := range vm.globals {
		vm.globals[i] = ValueNull
	}

	// Reset frame index (the main frame stays, but we reset to just 1 frame)
	vm.frameIndex = 1

	// Reset the main frame's IP to 0 so execution starts from the beginning
	if len(vm.frames) > 0 && vm.frames[0] != nil {
		vm.frames[0].IP = 0
		// Reset registers used by the main frame
		numRegs := vm.frames[0].Fn.NumRegs
		if numRegs <= 0 || numRegs > compiler.NumRegisters {
			numRegs = compiler.NumRegisters
		}
		for i := 0; i < numRegs; i++ {
			vm.frames[0].Registers[i] = ValueNull
		}
		vm.frames[0].Locals = nil
		vm.frames[0].FreeVars = nil
		vm.frames[0].This = ValueNull
		vm.frames[0].CurrentClass = nil
		vm.frames[0].SelectCases = nil
		vm.frames[0].SelectNumCases = 0
	}

	// Clear pending exception
	vm.pendingException = ValueNull

	// Clear temp stack
	if vm.tempStack != nil {
		vm.tempStack.Reset()
	}

	// Reset next global index
	vm.nextGlobalIndex = 0
}

// Close drains and closes the pool. After Close, Acquire will return an error.
// Close is safe to call multiple times.
func (p *VMPool) Close() {
	if p.closed {
		return
	}
	p.closed = true
	close(p.pool)
}

// Available returns the number of VMs currently available in the pool
// (not in use). This is a point-in-time snapshot.
func (p *VMPool) Available() int {
	return len(p.pool)
}

// ErrPoolClosed is returned by Acquire when the pool has been closed.
var ErrPoolClosed = &poolError{"vm pool is closed"}

// ErrNoFactory is returned by Acquire when no factory or bytecode is configured.
var ErrNoFactory = &poolError{"no VM factory configured"}

type poolError struct {
	msg string
}

func (e *poolError) Error() string {
	return e.msg
}
