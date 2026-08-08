// pkg/server/concurrent.go
// Concurrent script execution helper.
//
// RunConcurrent executes multiple scripts concurrently using errgroup.
// There are no artificial limits on concurrency or execution time — scripts
// run to completion (or until they error/panic). The VM pool is used purely
// for instance reuse, not as a semaphore.

package server

import (
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/topxeq/xxlang/pkg/vm"
)

// ConcurrentConfig holds configuration for concurrent execution.
// Currently empty — reserved for future options.
type ConcurrentConfig struct{}

// RunConcurrent executes the given list of script paths concurrently.
// Each script is loaded from the cache, a VM is acquired from the pool,
// the script is run, and the VM is released back to the pool.
//
// There is no timeout and no concurrency limit. Scripts run to completion.
// If any script fails, the error is returned after all goroutines finish.
func RunConcurrent(
	pool *vm.VMPool,
	cache *ScriptCache,
	scripts []string,
	_ ConcurrentConfig,
) error {
	if len(scripts) == 0 {
		return nil
	}

	g := errgroup.Group{}

	for _, path := range scripts {
		scriptPath := path
		g.Go(func() error {
			entry, err := cache.GetOrLoad(scriptPath)
			if err != nil {
				return fmt.Errorf("cache miss for %s: %w", scriptPath, err)
			}

			v, err := pool.Acquire()
			if err != nil {
				return fmt.Errorf("acquire VM for %s: %w", scriptPath, err)
			}

			// Track registry lifecycle (same protocol as RunScriptOnHttp)
			vm.BeginExecution()
			defer vm.EndExecution()

			var execErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						execErr = fmt.Errorf("panic in %s: %v", scriptPath, r)
					}
					pool.Release(v)
				}()

				if err := v.Run(); err != nil {
					execErr = fmt.Errorf("execute %s: %w", scriptPath, err)
				}
			}()

			_ = entry // bytecode is referenced by the VM already
			return execErr
		})
	}

	return g.Wait()
}

// RunConcurrentWithResults executes scripts concurrently and returns individual
// results indexed in the same order as the input scripts slice.
func RunConcurrentWithResults(
	pool *vm.VMPool,
	cache *ScriptCache,
	scripts []string,
	_ ConcurrentConfig,
) ([]Result, error) {
	if len(scripts) == 0 {
		return nil, nil
	}

	results := make([]Result, len(scripts))
	g := errgroup.Group{}

	for i, path := range scripts {
		idx := i
		scriptPath := path
		g.Go(func() error {
			entry, err := cache.GetOrLoad(scriptPath)
			if err != nil {
				results[idx] = Result{Path: scriptPath, Err: err}
				return nil
			}

			v, err := pool.Acquire()
			if err != nil {
				results[idx] = Result{Path: scriptPath, Err: err}
				return nil
			}

			// Track registry lifecycle (same protocol as RunScriptOnHttp)
			vm.BeginExecution()
			defer vm.EndExecution()

			var execErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						execErr = fmt.Errorf("panic in %s: %v", scriptPath, r)
					}
					pool.Release(v)
				}()

				if err := v.Run(); err != nil {
					execErr = err
				}
			}()

			results[idx] = Result{Path: scriptPath, Err: execErr}
			_ = entry
			return nil
		})
	}

	_ = g.Wait()

	for _, r := range results {
		if r.Err != nil {
			return results, r.Err
		}
	}

	return results, nil
}

// Result holds the outcome of a single script execution.
type Result struct {
	Path string
	Err  error
}
