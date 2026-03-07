// pkg/module/loader.go
// Module loading functionality with caching and cycle detection.
package module

import (
	"sync"
)

// Loader handles module loading with caching and cycle detection.
// It maintains a cache of loaded modules and tracks which modules
// are currently being loaded to detect circular dependencies.
type Loader struct {
	mu      sync.RWMutex
	modules map[string]*Module // cached modules
	loading map[string]bool    // modules currently being loaded (cycle detection)
}

// NewLoader creates a new module loader with empty caches.
func NewLoader() *Loader {
	return &Loader{
		modules: make(map[string]*Module),
		loading: make(map[string]bool),
	}
}

// Get retrieves a cached module or returns error if not cached.
func (l *Loader) Get(path string) (*Module, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	m, ok := l.modules[path]
	if !ok {
		return nil, ErrModuleNotFound
	}
	return m, nil
}

// Set caches a module with the given path as key.
func (l *Loader) Set(path string, m *Module) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.modules[path] = m
}

// IsLoading checks if a module is currently being loaded.
// This is used for cycle detection during module resolution.
func (l *Loader) IsLoading(path string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loading[path]
}

// MarkLoading marks a module as being loaded.
// Call this before starting to load a module to enable cycle detection.
func (l *Loader) MarkLoading(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loading[path] = true
}

// MarkDone marks a module as done loading.
// Call this after a module has been successfully loaded and cached.
func (l *Loader) MarkDone(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.loading, path)
}

// HasModule checks if a module is cached.
func (l *Loader) HasModule(path string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.modules[path]
	return ok
}
