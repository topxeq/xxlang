package module

import (
	"path/filepath"
	"testing"
)

func TestResolveRelativePath(t *testing.T) {
	tests := []struct {
		importer   string
		importPath string
		want       string
	}{
		{"/project/main.xxl", "./math", "/project/math.xxl"},
		{"/project/main.xxl", "./utils/helper", "/project/utils/helper.xxl"},
		{"/project/main.xxl", "../lib", "/lib.xxl"},
		{"/project/src/main.xxl", "./math", "/project/src/math.xxl"},
	}

	for _, tt := range tests {
		got, err := Resolve(tt.importer, tt.importPath)
		if err != nil {
			t.Errorf("Resolve(%s, %s) error: %v", tt.importer, tt.importPath, err)
			continue
		}
		wantNorm := filepath.Clean(tt.want)
		gotNorm := filepath.Clean(got)
		if gotNorm != wantNorm {
			t.Errorf("Resolve(%s, %s) = %s, want %s", tt.importer, tt.importPath, gotNorm, wantNorm)
		}
	}
}

func TestResolveBarePath(t *testing.T) {
	_, err := Resolve("/project/main.xxl", "some/bare/path")
	if err == nil {
		t.Error("expected error for bare import path without std/ prefix")
	}
}

func TestResolveStdlibPath(t *testing.T) {
	got, err := Resolve("/project/main.xxl", "std/math")
	if err != nil {
		t.Errorf("Resolve std/math error: %v", err)
	}
	if got != "std/math" {
		t.Errorf("Resolve std/math = %s, want std/math", got)
	}
}

func TestResolveAbsolutePath(t *testing.T) {
	tests := []struct {
		importer   string
		importPath string
		want       string
	}{
		// Absolute paths on Unix
		{"/project/main.xxl", "/abs/path/module.xxl", "/abs/path/module.xxl"},
		{"/project/main.xxl", "/home/user/project/utils.xxl", "/home/user/project/utils.xxl"},
		// Absolute path without extension - should add .xxl
		{"/project/main.xxl", "/abs/path/module", "/abs/path/module.xxl"},
		// Absolute path with subdirectories
		{"/project/main.xxl", "/usr/local/lib/xxlang/math", "/usr/local/lib/xxlang/math.xxl"},
	}

	for _, tt := range tests {
		got, err := Resolve(tt.importer, tt.importPath)
		if err != nil {
			t.Errorf("Resolve(%s, %s) error: %v", tt.importer, tt.importPath, err)
			continue
		}
		wantNorm := filepath.Clean(tt.want)
		gotNorm := filepath.Clean(got)
		if gotNorm != wantNorm {
			t.Errorf("Resolve(%s, %s) = %s, want %s", tt.importer, tt.importPath, gotNorm, wantNorm)
		}
	}
}

func TestResolveWasmPath(t *testing.T) {
	tests := []struct {
		importer   string
		importPath string
		want       string
	}{
		// Absolute path to .wasm file - should NOT add extension
		{"/project/main.xxl", "/plugins/fib.wasm", "/plugins/fib.wasm"},
		// Relative path to .wasm file
		{"/project/main.xxl", "./plugins/fib.wasm", "/project/plugins/fib.wasm"},
		{"/project/src/main.xxl", "../plugins/math.wasm", "/project/plugins/math.wasm"},
	}

	for _, tt := range tests {
		got, err := Resolve(tt.importer, tt.importPath)
		if err != nil {
			t.Errorf("Resolve(%s, %s) error: %v", tt.importer, tt.importPath, err)
			continue
		}
		wantNorm := filepath.Clean(tt.want)
		gotNorm := filepath.Clean(got)
		if gotNorm != wantNorm {
			t.Errorf("Resolve(%s, %s) = %s, want %s", tt.importer, tt.importPath, gotNorm, wantNorm)
		}
	}
}
