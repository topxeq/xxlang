// pkg/stdlib/os_test.go
package stdlib

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callOsFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("os")
	if mod == nil {
		panic("os module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

func TestOsModuleExists(t *testing.T) {
	mod := Get("os")
	if mod == nil {
		t.Fatal("os module not found")
	}
}

func TestOsIsDir(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	result := callOsFunc("isDir", String(tmpDir))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isDir() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isDir() should return true for directory")
	}

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	result = callOsFunc("isDir", String(tmpFile.Name()))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isDir() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isDir() should return false for file")
	}
}

func TestOsIsFile(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	result := callOsFunc("isFile", String(tmpFile.Name()))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isFile() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Error("isFile() should return true for file")
	}

	// Test with directory
	tmpDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	result = callOsFunc("isFile", String(tmpDir))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isFile() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isFile() should return false for directory")
	}
}

func TestOsHostname(t *testing.T) {
	result := callOsFunc("hostname")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("hostname() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("hostname() should return non-empty string")
	}
}

func TestOsTempDir(t *testing.T) {
	result := callOsFunc("temp")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("temp() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("temp() should return non-empty string")
	}
}

func TestOsHome(t *testing.T) {
	result := callOsFunc("home")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("home() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("home() should return non-empty string")
	}
}

func TestOsPlatform(t *testing.T) {
	result := callOsFunc("platform")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("platform() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("platform() should return non-empty string")
	}
}

func TestOsArch(t *testing.T) {
	result := callOsFunc("arch")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("arch() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("arch() should return non-empty string")
	}
}

func TestOsCpus(t *testing.T) {
	result := callOsFunc("cpus")
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("cpus() should return Int, got %T", result)
	}
	if i.Value <= 0 {
		t.Errorf("cpus() = %d, should be positive", i.Value)
	}
}

func TestOsAbs(t *testing.T) {
	result := callOsFunc("abs", String("."))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("abs() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("abs() should return non-empty string")
	}
}

func TestOsBase(t *testing.T) {
	result := callOsFunc("base", String("/path/to/file.txt"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("base() should return String, got %T", result)
	}
	if s.Value != "file.txt" {
		t.Errorf("base() = %s, want 'file.txt'", s.Value)
	}
}

func TestOsDir(t *testing.T) {
	result := callOsFunc("dir", String("/path/to/file.txt"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("dir() should return String, got %T", result)
	}
	// Normalize path separators for cross-platform testing
	expected := filepath.ToSlash("/path/to")
	got := filepath.ToSlash(s.Value)
	if got != expected {
		t.Errorf("dir() = %s, want '%s'", s.Value, expected)
	}
}

func TestOsExt(t *testing.T) {
	result := callOsFunc("ext", String("/path/to/file.txt"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("ext() should return String, got %T", result)
	}
	if s.Value != ".txt" {
		t.Errorf("ext() = %s, want '.txt'", s.Value)
	}
}

func TestOsClean(t *testing.T) {
	result := callOsFunc("clean", String("/path/../to/./file.txt"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("clean() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("clean() should return non-empty string")
	}
}

func TestOsIsAbs(t *testing.T) {
	// Use platform-specific absolute path
	absPath := "/absolute/path"
	if runtime.GOOS == "windows" {
		absPath = "C:\\absolute\\path"
	}

	result := callOsFunc("isAbs", String(absPath))
	b, ok := result.(*objects.Bool)
	if !ok {
		t.Fatalf("isAbs() should return Bool, got %T", result)
	}
	if !b.Value {
		t.Errorf("isAbs('%s') should return true", absPath)
	}

	result = callOsFunc("isAbs", String("relative/path"))
	b, ok = result.(*objects.Bool)
	if !ok {
		t.Fatalf("isAbs() should return Bool, got %T", result)
	}
	if b.Value {
		t.Error("isAbs('relative/path') should return false")
	}
}

func TestOsJoin(t *testing.T) {
	result := callOsFunc("join", String("path"), String("to"), String("file"))
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("join() should return String, got %T", result)
	}
	expected := filepath.Join("path", "to", "file")
	if s.Value != expected {
		t.Errorf("join() = %s, want '%s'", s.Value, expected)
	}
}

func TestOsSize(t *testing.T) {
	// Create a temp file with known content
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	content := "hello world"
	tmpFile.WriteString(content)
	tmpFile.Close()

	result := callOsFunc("size", String(tmpFile.Name()))
	i, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("size() should return Int, got %T", result)
	}
	if i.Value != int64(len(content)) {
		t.Errorf("size() = %d, want %d", i.Value, len(content))
	}
}

func TestOsListDir(t *testing.T) {
	// Create temp dir with files
	tmpDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Create(filepath.Join(tmpDir, "file1.txt"))
	os.Create(filepath.Join(tmpDir, "file2.txt"))

	result := callOsFunc("listDir", String(tmpDir))
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("listDir() should return Array, got %T", result)
	}
	if len(arr.Elements) < 2 {
		t.Errorf("listDir() should return at least 2 elements, got %d", len(arr.Elements))
	}
}

func TestOsStat(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	result := callOsFunc("stat", String(tmpFile.Name()))
	// stat returns an Array with file info
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("stat() should return Array, got %T", result)
	}
	if len(arr.Elements) == 0 {
		t.Error("stat() should return array with info")
	}
}

func TestOsRename(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	newPath := tmpFile.Name() + ".renamed"
	defer os.Remove(newPath)

	result := callOsFunc("rename", String(tmpFile.Name()), String(newPath))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("rename() should return Null, got %T", result)
	}

	// Verify file was renamed
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("rename() should have renamed the file")
	}
}

func TestOsTempFile(t *testing.T) {
	result := callOsFunc("tempFile")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("tempFile() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("tempFile() should return non-empty string")
	}
	// Clean up
	os.Remove(s.Value)
}

func TestOsTempDirFunc(t *testing.T) {
	result := callOsFunc("tempDir")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("tempDir() should return String, got %T", result)
	}
	if s.Value == "" {
		t.Error("tempDir() should return non-empty string")
	}
	// Clean up
	os.RemoveAll(s.Value)
}

// Tests for config functions

func TestGetConfigMap(t *testing.T) {
	// GetConfigMap should return a map (empty if no config file exists)
	cfg := GetConfigMap()
	if cfg == nil {
		t.Error("GetConfigMap() should return non-nil map")
	}
}

func TestGetConfigObjImpl(t *testing.T) {
	// getConfigObjImpl should return a map object
	result := getConfigObjImpl()
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("getConfigObjImpl() should return Map, got %T", result)
	}
	if m.Pairs == nil {
		t.Error("getConfigObjImpl() should return map with non-nil Pairs")
	}
}

func TestGetConfigStrImpl(t *testing.T) {
	// Test with non-existent config name
	result := getConfigStrImpl("nonexistent_config_12345")
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("getConfigStrImpl() for non-existent config should return Null, got %T", result)
	}
}

func TestSetConfigStrImpl(t *testing.T) {
	// Test setting a config value
	result := setConfigStrImpl("test_config_name", "test value")
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("setConfigStrImpl() should return Null, got %T", result)
	}

	// Verify we can read it back
	result = getConfigStrImpl("test_config_name")
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("getConfigStrImpl() should return String, got %T", result)
	}
	if s.Value != "test value" {
		t.Errorf("getConfigStrImpl() = %s, want 'test value'", s.Value)
	}

	// Clean up
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".xxl", "test_config_name.cfg")
	os.Remove(configPath)
}

func TestTryReadConfigFile(t *testing.T) {
	// Test with non-existent path
	result := tryReadConfigFile("/nonexistent/path/settings.json")
	if result != nil {
		t.Error("tryReadConfigFile() for non-existent path should return nil")
	}
}

func TestTryReadConfigMap(t *testing.T) {
	// Test with non-existent path
	result := tryReadConfigMap("/nonexistent/path/settings.json")
	if result != nil {
		t.Error("tryReadConfigMap() for non-existent path should return nil")
	}
}

func TestOsGetConfigObj(t *testing.T) {
	result := callOsFunc("getConfigObj")
	m, ok := result.(*objects.Map)
	if !ok {
		t.Fatalf("getConfigObj() should return Map, got %T", result)
	}
	// Should return a map (possibly empty)
	_ = m
}

func TestOsGetConfigStr(t *testing.T) {
	result := callOsFunc("getConfigStr", String("nonexistent_config_12345"))
	// Should return Null for non-existent config
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("getConfigStr() for non-existent should return Null, got %T", result)
	}
}

func TestOsSetConfigStr(t *testing.T) {
	result := callOsFunc("setConfigStr", String("test_os_config"), String("test value"))
	if _, ok := result.(*objects.Null); !ok {
		t.Errorf("setConfigStr() should return Null, got %T", result)
	}

	// Clean up
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".xxl", "test_os_config.cfg")
	os.Remove(configPath)
}

func TestOsUserInfo(t *testing.T) {
	result := callOsFunc("userInfo")
	// userInfo returns an Array with user info
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("userInfo() should return Array, got %T", result)
	}
	if len(arr.Elements) == 0 {
		t.Error("userInfo() should return array with user info")
	}
}

func TestGetConfigStrImpl_ReadsFromHomeCfg(t *testing.T) {
	// Prepare a config file under the user's home (~/.xxl/<name>.cfg)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %v", err)
	}

	cfgDir := filepath.Join(homeDir, ".xxl")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	cfgName := "test_home_cfg"
	cfgPath := filepath.Join(cfgDir, cfgName+".cfg")
	defer os.Remove(cfgPath)

	const cfgContent = "home-cfg-value"
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed to write temp cfg: %v", err)
	}

	// When getConfigStrImpl is called with the name, it should read the file
	result := getConfigStrImpl(cfgName)
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("getConfigStrImpl() should return String, got %T", result)
	}
	if s.Value != cfgContent {
		t.Fatalf("getConfigStrImpl() = %q, want %q", s.Value, cfgContent)
	}
}

func TestSetConfigStrImpl_WritesAndReads(t *testing.T) {
	// Use a unique name to avoid collisions
	name := "test_os_config_set"
	value := "hello-config"
	// Write
	res := setConfigStrImpl(name, value)
	if _, ok := res.(*objects.Null); !ok {
		t.Fatalf("setConfigStrImpl() should return Null, got %T", res)
	}

	// Read back via getConfigStrImpl
	got := getConfigStrImpl(name)
	s, ok := got.(*objects.String)
	if !ok {
		t.Fatalf("getConfigStrImpl() should return String, got %T", got)
	}
	if s.Value != value {
		t.Fatalf("getConfigStrImpl() = %q, want %q", s.Value, value)
	}

	// Cleanup
	homeDir, _ := os.UserHomeDir()
	cfgPath := filepath.Join(homeDir, ".xxl", name+".cfg")
	os.Remove(cfgPath)
}

func TestTryReadConfigFile_ParsesJsonToMap(t *testing.T) {
	// Create a temporary JSON config file with a couple of keys
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")
	content := []byte(`{"foo": "bar", "num": 123}`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write temp json: %v", err)
	}

	obj := tryReadConfigFile(path)
	if obj == nil {
		t.Fatalf("tryReadConfigFile() returned nil, expected Map object")
	}
	m, ok := obj.(*objects.Map)
	if !ok {
		t.Fatalf("tryReadConfigFile() should return Map, got %T", obj)
	}
	// Verify entries exist
	foundFoo := false
	for _, pair := range m.Pairs {
		if pair.Key != nil {
			if ks, ok := pair.Key.(*objects.String); ok && ks.Value == "foo" {
				if s, ok := pair.Value.(*objects.String); ok {
					if s.Value != "bar" {
						t.Fatalf("expected foo.bar value 'bar', got '%s'", s.Value)
					}
					foundFoo = true
				}
			}
		}
	}
	if !foundFoo {
		t.Fatalf("expected to find key 'foo' in config map")
	}
}
