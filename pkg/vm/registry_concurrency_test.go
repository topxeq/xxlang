// pkg/vm/registry_concurrency_test.go
// Concurrency tests for the global object registry:
// verifies that concurrent VM executions + registry clearing do not
// corrupt object references (no race, no use-after-clear).
package vm

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/parser"
)

// compileSimple compiles a tiny script that creates objects.
func compileSimple(t *testing.T, code string) *compiler.Bytecode {
	t.Helper()
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return c.Bytecode()
}

// TestConcurrentExecutionsWithClear hammers the registry with many concurrent
// VM executions. The last one to finish clears the registry; others must not
// be affected. Run with -race to detect data races.
func TestConcurrentExecutionsWithClear(t *testing.T) {
	bytecode := compileSimple(t, `
var s = ""
for (var i = 0; i < 100; i = i + 1) {
	s = s + "x" + i
}
var m = {}
m["k1"] = s
m["k2"] = [1, 2, 3]
return m
`)

	const workers = 32
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for iter := 0; iter < 50; iter++ {
				BeginExecution()
				func() {
					defer EndExecution()
					v := NewRegVMWithGlobals(bytecode, make([]Value, GlobalsSize))
					if err := v.Run(); err != nil {
						errCh <- fmt.Errorf("worker %d iter %d: %v", id, iter, err)
						return
					}
					res := v.LastResult()
					if res == ValueNull {
						errCh <- fmt.Errorf("worker %d iter %d: null result", id, iter)
						return
					}
					// Force a read of the object through the registry
					obj := res.ToObject()
					if obj == nil {
						errCh <- fmt.Errorf("worker %d iter %d: nil object", id, iter)
						return
					}
				}()
			}
		}(w)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
}

// TestBeginEndBalanced verifies the counter returns to zero after balanced calls.
func TestBeginEndBalanced(t *testing.T) {
	before := atomic.LoadInt32(&activeVMCount)
	BeginExecution()
	BeginExecution()
	EndExecution()
	EndExecution()
	after := atomic.LoadInt32(&activeVMCount)
	if before != after {
		t.Fatalf("counter mismatch: before=%d after=%d", before, after)
	}
}

// TestConcurrentHigherOrderFuncs verifies that closures created by different
// VMs call back into the CORRECT VM when higher-order builtins (map/filter)
// are used concurrently. Before the fix, closures shared a global callback
// that raced between VMs.
func TestConcurrentHigherOrderFuncs(t *testing.T) {
	bytecode := compileSimple(t, `
var arr = [1, 2, 3, 4, 5]
var sum = 0
for (var i = 0; i < len(arr); i = i + 1) {
	sum = sum + arr[i]
}
var doubled = []
for (var i = 0; i < len(arr); i = i + 1) {
	doubled = doubled.push(arr[i] * 2)
}
return sum
`)

	const workers = 16
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for iter := 0; iter < 30; iter++ {
				BeginExecution()
				func() {
					defer EndExecution()
					v := NewRegVMWithGlobals(bytecode, make([]Value, GlobalsSize))
					if err := v.Run(); err != nil {
						errCh <- fmt.Errorf("worker %d iter %d: %v", id, iter, err)
						return
					}
					res := v.LastResult()
					obj := res.ToObject()
					if obj == nil {
						errCh <- fmt.Errorf("worker %d iter %d: nil result", id, iter)
						return
					}
					if obj.Type() != objects.IntType {
						errCh <- fmt.Errorf("worker %d iter %d: expected int result, got %s", id, iter, obj.Type())
						return
					}
					if obj.Inspect() != "15" {
						errCh <- fmt.Errorf("worker %d iter %d: expected 15, got %s", id, iter, obj.Inspect())
					}
				}()
			}
		}(w)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
}

// TestConcurrentClosureCallbacks exercises higher-order builtins (map/filter)
// from concurrent VMs. Each VM creates its own closure; the closure must call
// back into ITS OWN VM. With the old global callback, closures raced and
// executed in the wrong VM.
func TestConcurrentClosureCallbacks(t *testing.T) {
	bytecode := compileSimple(t, `
var arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
var mapped = mapArray(arr, func(x) { return x * 2 })
var filtered = filterArray(mapped, func(x) { return x % 4 == 0 })
var sum = 0
for (var i = 0; i < len(filtered); i = i + 1) {
	sum = sum + filtered[i]
}
return sum
`)

	const workers = 24
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for iter := 0; iter < 40; iter++ {
				BeginExecution()
				func() {
					defer EndExecution()
					v := NewRegVMWithGlobals(bytecode, make([]Value, GlobalsSize))
					if err := v.Run(); err != nil {
						errCh <- fmt.Errorf("worker %d iter %d: %v", id, iter, err)
						return
					}
					res := v.LastResult()
					obj := res.ToObject()
					if obj == nil {
						errCh <- fmt.Errorf("worker %d iter %d: nil result", id, iter)
						return
					}
					// 2,4,6,...,20 mapped; filtered (x%4==0): 4,8,12,16,20 -> sum=60
					if obj.Inspect() != "60" {
						errCh <- fmt.Errorf("worker %d iter %d: expected 60, got %s", id, iter, obj.Inspect())
					}
				}()
			}
		}(w)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
}
