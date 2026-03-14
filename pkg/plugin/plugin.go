// pkg/plugin/plugin.go
// Plugin system for loading WebAssembly plugins at runtime.
//
// WASM plugins work on all platforms (Windows, Linux, macOS) without CGO.
// Plugins can be written in TinyGo, Rust, C/C++, Zig, AssemblyScript, etc.
//
// Usage from xxlang:
//
//	import "plugin/myplugin"
//	myplugin.hello()
//
// Building a WASM plugin:
//
//	tinygo build -o myplugin.wasm -target=wasi myplugin.go
//	# or with Rust:
//	cargo build --target wasm32-wasi --release
package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/topxeq/xxlang/pkg/objects"
)

// Plugin is the interface that WASM plugins must implement.
type Plugin interface {
	// Name returns the plugin name (used as plugin/name in imports).
	Name() string

	// Exports returns the module's exported symbols.
	// The map keys are the names accessible from xxlang code.
	Exports() map[string]objects.Object
}

// Registry tracks loaded plugins.
// It is safe for concurrent use.
var Registry = struct {
	sync.RWMutex
	plugins map[string]Plugin
}{
	plugins: make(map[string]Plugin),
}

// Register registers a plugin with the registry.
// This is typically called from a plugin's init() function (for static plugins)
// or by the loader after loading a WASM plugin.
func Register(p Plugin) {
	if p == nil {
		return
	}
	Registry.Lock()
	defer Registry.Unlock()
	Registry.plugins[p.Name()] = p
}

// Get retrieves a plugin from the registry.
// Returns the plugin and true if found, nil and false otherwise.
func Get(name string) (Plugin, bool) {
	Registry.RLock()
	defer Registry.RUnlock()
	p, ok := Registry.plugins[name]
	return p, ok
}

// Has checks if a plugin is registered.
func Has(name string) bool {
	Registry.RLock()
	defer Registry.RUnlock()
	_, ok := Registry.plugins[name]
	return ok
}

// List returns all registered plugin names.
func List() []string {
	Registry.RLock()
	defer Registry.RUnlock()
	names := make([]string, 0, len(Registry.plugins))
	for name := range Registry.plugins {
		names = append(names, name)
	}
	return names
}

// Loader handles loading WASM plugin files.
type Loader struct {
	mu      sync.RWMutex
	paths   []string        // search paths for plugins
	loading map[string]bool // cycle detection
}

// NewLoader creates a new plugin loader with default search paths.
func NewLoader() *Loader {
	return &Loader{
		paths:   []string{"./plugins", "./plugin"},
		loading: make(map[string]bool),
	}
}

// AddPath adds a search path for plugins.
func (l *Loader) AddPath(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paths = append(l.paths, path)
}

// SetPaths sets the search paths for plugins.
func (l *Loader) SetPaths(paths []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paths = paths
}

// Paths returns the current search paths.
func (l *Loader) Paths() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	paths := make([]string, len(l.paths))
	copy(paths, l.paths)
	return paths
}

// Load loads a plugin by name.
// It first checks the registry, then searches for a .wasm file.
// Returns the plugin and any error encountered.
func (l *Loader) Load(name string) (Plugin, error) {
	// Check if already registered
	if p, ok := Get(name); ok {
		return p, nil
	}

	// Check for cycle
	l.mu.Lock()
	if l.loading[name] {
		l.mu.Unlock()
		return nil, fmt.Errorf("circular plugin load: %s", name)
	}
	l.loading[name] = true
	l.mu.Unlock()

	defer func() {
		l.mu.Lock()
		delete(l.loading, name)
		l.mu.Unlock()
	}()

	// Try to load from file
	plugin, err := l.loadFromFile(name)
	if err != nil {
		return nil, err
	}

	// Register the plugin
	Register(plugin)
	return plugin, nil
}

// loadFromFile attempts to load a plugin from a .wasm file.
func (l *Loader) loadFromFile(name string) (Plugin, error) {
	// Get search paths
	l.mu.RLock()
	paths := make([]string, len(l.paths))
	copy(paths, l.paths)
	l.mu.RUnlock()

	// Search for plugin file
	for _, searchPath := range paths {
		pluginPath := filepath.Join(searchPath, name+".wasm")

		// Check if file exists
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			continue
		}

		// Try to load the WASM plugin
		plugin, err := loadPluginWASM(pluginPath)
		if err == nil {
			return plugin, nil
		}
		return nil, fmt.Errorf("error loading plugin %s from %s: %v", name, pluginPath, err)
	}

	return nil, fmt.Errorf("plugin not found: %s (searched: %s)", name, strings.Join(paths, ", "))
}

// ToModule converts a Plugin to an objects.Module.
func ToModule(p Plugin) *objects.Module {
	exports := p.Exports()
	return &objects.Module{
		Name:    p.Name(),
		Exports: exports,
		Globals: nil, // Plugins don't have isolated globals
	}
}

// isNotExist checks if an error indicates that a plugin was not found.
func isNotExist(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "no such file") ||
		strings.Contains(msg, "cannot find")
}
