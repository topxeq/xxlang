// pkg/server/cache.go
// Thread-safe script cache with hot-reload capabilities.
//
// The ScriptCache stores pre-compiled bytecode keyed by file path. This avoids
// the expensive read+lex+parse+compile cycle on every request. When a script
// file changes on disk, UpdateScript atomically swaps the cached bytecode so
// subsequent requests use the new version without downtime.
//
// Concurrency: sync.RWMutex allows multiple concurrent readers (execution)
// while writes (hot-reload) are exclusive. The mutex is held only for the
// duration of the map lookup, not during script execution.

package server

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// CacheEntry holds a compiled script and its metadata.
type CacheEntry struct {
	// Bytecode is the pre-compiled script, safe to share across VMs.
	Bytecode *compiler.Bytecode
	// ModTime is the last modification time of the source file.
	ModTime time.Time
	// SourcePath is the original file path (for error reporting).
	SourcePath string
}

// ScriptCache is a thread-safe cache of compiled scripts.
type ScriptCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
}

// NewScriptCache creates an empty script cache.
func NewScriptCache() *ScriptCache {
	return &ScriptCache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get returns the cached bytecode for the given path, or nil if not cached.
// Get is safe for concurrent use and does not block other readers.
func (c *ScriptCache) Get(path string) *CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries[path]
}

// GetOrLoad returns the cached bytecode for path. If not cached, it reads the
// file from disk, compiles it, and stores the result in the cache.
//
// GetOrLoad uses a write lock for the compile path (to prevent thundering herd
// compilation), but a read lock for the cache hit path.
func (c *ScriptCache) GetOrLoad(path string) (*CacheEntry, error) {
	// Fast path: check cache with read lock
	c.mu.RLock()
	entry, ok := c.entries[path]
	c.mu.RUnlock()
	if ok {
		return entry, nil
	}

	// Slow path: compile and store with write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have
	// loaded it while we were waiting)
	if entry, ok = c.entries[path]; ok {
		return entry, nil
	}

	entry, err := c.loadFromFile(path)
	if err != nil {
		return nil, err
	}

	c.entries[path] = entry
	return entry, nil
}

// UpdateScript checks if the source file at path has been modified since it
// was last cached. If so, it recompiles and atomically replaces the cached
// entry. If the file hasn't changed, UpdateScript returns nil without doing
// any work (returns false as the bool to indicate no reload).
//
// Returns (true, nil) if the script was reloaded, (false, nil) if unchanged,
// or (_, err) if compilation failed.
func (c *ScriptCache) UpdateScript(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", path, err)
	}
	newModTime := info.ModTime()

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[path]
	if ok && !newModTime.After(entry.ModTime) {
		// File hasn't changed since last load
		return false, nil
	}

	newEntry, err := c.loadFromFile(path)
	if err != nil {
		return false, err
	}

	c.entries[path] = newEntry
	return true, nil
}

// Invalidate removes the entry for path from the cache. The next request for
// this path will trigger a fresh compile.
func (c *ScriptCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, path)
}

// Clear removes all entries from the cache.
func (c *ScriptCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Len returns the number of cached scripts.
func (c *ScriptCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// loadFromFile reads a script from disk, lexes, parses, and compiles it.
// Must be called with c.mu held (write lock).
func (c *ScriptCache) loadFromFile(path string) (*CacheEntry, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	// Lexical analysis
	l := lexer.New(string(source))

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parser errors for %s: %v", path, p.Errors())
	}

	// Compilation
	comp := compiler.NewRegCompiler()
	comp.SetSourceFile(path)

	if _, err := comp.Compile(program); err != nil {
		return nil, fmt.Errorf("compiler error for %s: %v", path, err)
	}

	bytecode := comp.Bytecode()

	return &CacheEntry{
		Bytecode:   bytecode,
		ModTime:    info.ModTime(),
		SourcePath: path,
	}, nil
}
