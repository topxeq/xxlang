// pkg/plugin/plugin_test.go
package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// mockPlugin is a test implementation of the Plugin interface
type mockPlugin struct {
	name    string
	exports map[string]objects.Object
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Exports() map[string]objects.Object {
	return m.exports
}

func TestPlugin_Interface(t *testing.T) {
	t.Run("plugin interface implementation", func(t *testing.T) {
		p := &mockPlugin{
			name: "test",
			exports: map[string]objects.Object{
				"hello": &objects.String{Value: "world"},
			},
		}

		if p.Name() != "test" {
			t.Errorf("expected name 'test', got %q", p.Name())
		}
		if len(p.Exports()) != 1 {
			t.Errorf("expected 1 export, got %d", len(p.Exports()))
		}
		exports := p.Exports()
		if s, ok := exports["hello"].(*objects.String); !ok || s.Value != "world" {
			t.Errorf("expected export 'hello' to be string 'world'")
		}
	})
}

func TestPlugin_Registry(t *testing.T) {
	t.Run("register and get", func(t *testing.T) {
		// Clear registry for test
		Registry.Lock()
		Registry.plugins = make(map[string]Plugin)
		Registry.Unlock()

		p := &mockPlugin{name: "testplugin", exports: nil}
		Register(p)

		got, ok := Get("testplugin")
		if !ok {
			t.Fatal("expected plugin to be found")
		}
		if got == nil {
			t.Fatal("expected plugin, got nil")
		}
		if got.Name() != "testplugin" {
			t.Errorf("expected name 'testplugin', got %q", got.Name())
		}
	})

	t.Run("get non-existent", func(t *testing.T) {
		got, ok := Get("nonexistent")
		if ok {
			t.Errorf("expected ok to be false for non-existent plugin")
		}
		if got != nil {
			t.Errorf("expected nil for non-existent plugin, got %v", got)
		}
	})

	t.Run("register nil", func(t *testing.T) {
		Registry.Lock()
		Registry.plugins = make(map[string]Plugin)
		Registry.Unlock()

		// Register nil should not panic and should not add anything
		Register(nil)

		if len(Registry.plugins) != 0 {
			t.Errorf("expected empty registry after registering nil, got %d plugins", len(Registry.plugins))
		}
	})
}

func TestPlugin_Has(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.plugins["existing"] = &mockPlugin{name: "existing"}
	Registry.Unlock()

	t.Run("existing plugin", func(t *testing.T) {
		if !Has("existing") {
			t.Error("expected Has('existing') to return true")
		}
	})

	t.Run("non-existent plugin", func(t *testing.T) {
		if Has("nonexistent") {
			t.Error("expected Has('nonexistent') to return false")
		}
	})
}

func TestPlugin_List(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.plugins["a"] = &mockPlugin{name: "a"}
	Registry.plugins["b"] = &mockPlugin{name: "b"}
	Registry.Unlock()

	list := List()
	if len(list) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(list))
	}

	// Check that both plugins are in the list
	found := make(map[string]bool)
	for _, name := range list {
		found[name] = true
	}
	if !found["a"] || !found["b"] {
		t.Errorf("expected list to contain 'a' and 'b', got %v", list)
	}
}

func TestLoader_New(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("expected loader, got nil")
	}

	paths := loader.Paths()
	if len(paths) != 2 {
		t.Errorf("expected 2 default paths, got %d", len(paths))
	}
}

func TestLoader_Paths(t *testing.T) {
	loader := NewLoader()

	t.Run("add path", func(t *testing.T) {
		loader.AddPath("/custom/path")
		paths := loader.Paths()
		if len(paths) != 3 {
			t.Errorf("expected 3 paths after AddPath, got %d", len(paths))
		}
		// Check the last path is the one we added
		if paths[len(paths)-1] != "/custom/path" {
			t.Errorf("expected last path to be '/custom/path', got %q", paths[len(paths)-1])
		}
	})

	t.Run("set paths", func(t *testing.T) {
		loader.SetPaths([]string{"/path1", "/path2"})
		paths := loader.Paths()
		if len(paths) != 2 {
			t.Errorf("expected 2 paths after SetPaths, got %d", len(paths))
		}
		if paths[0] != "/path1" || paths[1] != "/path2" {
			t.Errorf("expected paths ['/path1', '/path2'], got %v", paths)
		}
	})

	t.Run("paths copy", func(t *testing.T) {
		loader.SetPaths([]string{"/original"})
		paths := loader.Paths()
		paths[0] = "/modified"

		// Original should not be modified
		originalPaths := loader.Paths()
		if originalPaths[0] != "/original" {
			t.Errorf("expected Paths() to return a copy, but original was modified")
		}
	})
}

func TestLoader_Load_AlreadyRegistered(t *testing.T) {
	// Clear and register a plugin
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.plugins["registered"] = &mockPlugin{name: "registered", exports: map[string]objects.Object{"test": &objects.Int{Value: 1}}}
	Registry.Unlock()

	loader := NewLoader()
	p, err := loader.Load("registered")
	if err != nil {
		t.Errorf("expected no error for already registered plugin, got %v", err)
	}
	if p == nil {
		t.Error("expected plugin, got nil")
	}
	if p.Name() != "registered" {
		t.Errorf("expected name 'registered', got %q", p.Name())
	}
}

func TestLoader_Load_NotFound(t *testing.T) {
	// Clear registry
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()
	loader.SetPaths([]string{"/nonexistent/path"})

	p, err := loader.Load("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestToModule(t *testing.T) {
	p := &mockPlugin{
		name: "testmodule",
		exports: map[string]objects.Object{
			"func1": &objects.Builtin{Fn: func(args ...objects.Object) objects.Object { return objects.NULL }},
			"var1":  &objects.String{Value: "value1"},
		},
	}

	m := ToModule(p)
	if m == nil {
		t.Fatal("expected module, got nil")
	}
	if m.Name != "testmodule" {
		t.Errorf("expected name 'testmodule', got %q", m.Name)
	}
	if len(m.Exports) != 2 {
		t.Errorf("expected 2 exports, got %d", len(m.Exports))
	}
	if _, ok := m.Exports["func1"]; !ok {
		t.Error("expected 'func1' in exports")
	}
	if _, ok := m.Exports["var1"]; !ok {
		t.Error("expected 'var1' in exports")
	}
}

func TestLoader_CycleDetection(t *testing.T) {
	// Clear registry
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()
	loader.SetPaths([]string{"/nonexistent"})

	// Manually mark a plugin as loading to simulate a cycle
	loader.mu.Lock()
	loader.loading["cyclic"] = true
	loader.mu.Unlock()

	p, err := loader.Load("cyclic")
	if err == nil {
		t.Error("expected error for cyclic load")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
	if err != nil && !containsStr(err.Error(), "circular") {
		t.Errorf("expected circular error message, got: %v", err)
	}

	// Clean up
	loader.mu.Lock()
	delete(loader.loading, "cyclic")
	loader.mu.Unlock()
}

func TestIsNotExist(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{fmt.Errorf("plugin not found"), true},
		{fmt.Errorf("no such file or directory"), true},
		{fmt.Errorf("cannot find module"), true},
		{fmt.Errorf("some other error"), false},
	}

	for _, tt := range tests {
		result := isNotExist(tt.err)
		if result != tt.expected {
			t.Errorf("isNotExist(%v) = %v, expected %v", tt.err, result, tt.expected)
		}
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s[1:], substr) || s[:len(substr)] == substr)
}

func TestPlugin_NilExports(t *testing.T) {
	p := &mockPlugin{
		name:    "nilexports",
		exports: nil,
	}

	// Should not panic when Exports() returns nil
	exports := p.Exports()
	if exports != nil {
		t.Errorf("expected nil exports, got %v", exports)
	}

	// ToModule should handle nil exports
	m := ToModule(p)
	if m == nil {
		t.Fatal("expected module, got nil")
	}
	if m.Name != "nilexports" {
		t.Errorf("expected name 'nilexports', got %q", m.Name)
	}
}

func TestPlugin_Overwrite(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	// Register first plugin
	p1 := &mockPlugin{name: "overwrite", exports: map[string]objects.Object{"v": &objects.Int{Value: 1}}}
	Register(p1)

	// Register second plugin with same name (should overwrite)
	p2 := &mockPlugin{name: "overwrite", exports: map[string]objects.Object{"v": &objects.Int{Value: 2}}}
	Register(p2)

	// Should get the second plugin
	got, ok := Get("overwrite")
	if !ok {
		t.Fatal("expected plugin to be found")
	}
	if got != p2 {
		t.Error("expected second plugin to overwrite first")
	}
}

func TestLoader_LoadPath(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()

	t.Run("file not found", func(t *testing.T) {
		p, err := loader.LoadPath("/nonexistent/path/plugin.wasm")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
		if p != nil {
			t.Errorf("expected nil plugin, got %v", p)
		}
	})
}

func TestLoader_loadFromFile_ExistingFile(t *testing.T) {
	// Create a temp file to test file existence check
	tmpFile, err := os.CreateTemp("", "test*.wasm")
	if err != nil {
		t.Skip("could not create temp file")
	}
	defer os.Remove(tmpFile.Name())

	loader := NewLoader()
	loader.SetPaths([]string{filepath.Dir(tmpFile.Name())})

	// Load a file that exists but isn't valid WASM
	p, err := loader.loadFromFile(filepath.Base(tmpFile.Name())[:len(filepath.Base(tmpFile.Name()))-5])
	// Should fail because the file isn't a valid WASM file
	if err == nil {
		t.Error("expected error for invalid WASM file")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestLoader_loadFromFile_AllPaths(t *testing.T) {
	loader := NewLoader()
	loader.SetPaths([]string{"/nonexistent1", "/nonexistent2", "/nonexistent3"})

	// Try to load from multiple paths - all should fail
	p, err := loader.loadFromFile("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestLoader_Concurrent(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()

	// Test concurrent access to Paths
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			loader.AddPath("/test/path")
			_ = loader.Paths()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================
// Additional LoadPath tests
// ============================================

func TestLoader_LoadPath_NilPaths(t *testing.T) {
	loader := NewLoader()
	loader.SetPaths(nil)

	p, err := loader.LoadPath("/nonexistent/plugin.wasm")
	if err == nil {
		t.Error("expected error for nil paths")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestLoader_LoadPath_EmptyPath(t *testing.T) {
	loader := NewLoader()

	p, err := loader.LoadPath("")
	if err == nil {
		t.Error("expected error for empty path")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestLoader_Load_WithPluginExtension(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()
	loader.SetPaths([]string{"/nonexistent"})

	// Test loading plugin with explicit .plugin extension
	p, err := loader.Load("myplugin.plugin")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestLoader_Load_WithWasmExtension(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()
	loader.SetPaths([]string{"/nonexistent"})

	// Test loading plugin with explicit .wasm extension
	p, err := loader.Load("myplugin.wasm")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestLoader_Load_WithSoExtension(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()
	loader.SetPaths([]string{"/nonexistent"})

	// Test loading plugin with .so extension
	p, err := loader.Load("myplugin.so")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

func TestLoader_LoadPath_RegisteredPlugin(t *testing.T) {
	// Register a plugin
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.plugins["test_registered"] = &mockPlugin{
		name: "test_registered",
		exports: map[string]objects.Object{"x": &objects.Int{Value: 1}},
	}
	Registry.Unlock()

	loader := NewLoader()
	// LoadPath should work even for registered plugins
	p, err := loader.Load("test_registered")
	if err != nil {
		t.Errorf("expected no error for registered plugin, got %v", err)
	}
	if p == nil {
		t.Error("expected plugin, got nil")
	}
	if p.Name() != "test_registered" {
		t.Errorf("expected name 'test_registered', got %q", p.Name())
	}
}

func TestLoader_LoadPath_StatError(t *testing.T) {
	loader := NewLoader()

	// Try to load from a path that can't be accessed
	p, err := loader.LoadPath("/root/some_nonexistent_plugin.wasm")
	if err == nil {
		t.Error("expected error for inaccessible path")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

// TestLoader_Load_MultiplePaths tests searching through multiple paths
func TestLoader_Load_MultiplePaths(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	loader := NewLoader()
	loader.SetPaths([]string{
		"/nonexistent/path1",
		"/nonexistent/path2",
		"/nonexistent/path3",
	})

	p, err := loader.Load("test")
	if err == nil {
		t.Error("expected error for plugin not in any path")
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %v", p)
	}
}

// TestToModule_NilExports tests ToModule with nil exports
func TestToModule_NilExports(t *testing.T) {
	p := &mockPlugin{
		name:    "nil_exports",
		exports: nil,
	}

	m := ToModule(p)
	if m == nil {
		t.Fatal("expected module, got nil")
	}
	if m.Name != "nil_exports" {
		t.Errorf("expected name 'nil_exports', got %q", m.Name)
	}
	// Exports can be nil if the plugin returns nil
}

// TestPlugin_RegistryConcurrent tests concurrent access to the registry
func TestPlugin_RegistryConcurrent(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func(id int) {
			name := fmt.Sprintf("plugin_%d", id)
			Register(&mockPlugin{name: name})
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func(id int) {
			name := fmt.Sprintf("plugin_%d", id)
			_ = Has(name)
			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestList_EmptyRegistry tests List with empty registry
func TestList_EmptyRegistry(t *testing.T) {
	Registry.Lock()
	Registry.plugins = make(map[string]Plugin)
	Registry.Unlock()

	list := List()
	if list == nil {
		t.Error("expected non-nil list")
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %v", list)
	}
}
