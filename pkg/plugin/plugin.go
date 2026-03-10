// pkg/plugin/plugin.go
// Plugin system for loading native Go plugins at runtime.
//
// Plugins are Go shared libraries (.so on Linux/macOS) that implement
// the Plugin interface. They can export functions, variables, and other
// values to xxlang code.
//
// Usage from xxlang:
//
//	import "plugin/myplugin"
//	myplugin.hello()
//
// Building a plugin:
//
//	go build -buildmode=plugin -o myplugin.so myplugin.go
package plugin

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/topxeq/xxlang/pkg/objects"
)

// Plugin is the interface that native plugins must implement.
// Plugins are loaded from .so files and register their exports
// through this interface.
type Plugin interface {
	// Name returns the plugin name (used as plugin/name in imports).
	// This should match the filename (without .so extension).
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
// This is typically called from a plugin's init() function.
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

// Loader handles loading .so plugin files.
type Loader struct {
	mu       sync.RWMutex
	paths    []string          // search paths for plugins
	loading  map[string]bool   // cycle detection
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
// It first checks the registry, then searches for a .so file.
// Returns the plugin and any error encountered.
//
// Note: Go's plugin package only works on Linux, macOS, and FreeBSD.
// On Windows, this will return an error.
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

// loadFromFile attempts to load a plugin from a .so file.
func (l *Loader) loadFromFile(name string) (Plugin, error) {
	// Get search paths
	l.mu.RLock()
	paths := make([]string, len(l.paths))
	copy(paths, l.paths)
	l.mu.RUnlock()

	// Search for plugin file
	for _, searchPath := range paths {
		soPath := filepath.Join(searchPath, name+".so")

		// Try to load the plugin
		plugin, err := loadPluginSO(soPath)
		if err == nil {
			return plugin, nil
		}
		// Continue to next path if file not found
		if !isNotExist(err) {
			return nil, fmt.Errorf("error loading plugin %s from %s: %v", name, soPath, err)
		}
	}

	return nil, fmt.Errorf("plugin not found: %s (searched: %s)", name, strings.Join(paths, ", "))
}

// loadPluginSO is implemented in plugin_native.go for supported platforms
// and plugin_stub.go for unsupported platforms (Windows).
// The signature is:
//   func loadPluginSO(path string) (Plugin, error)

// isNotExist checks if an error indicates file not found.
func isNotExist(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "no such file") ||
		strings.Contains(err.Error(), "cannot find"))
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
