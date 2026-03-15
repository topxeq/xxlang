// pkg/module/resolver.go
// Module path resolution functionality.
package module

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	// ErrModuleNotFound is returned when a module cannot be found.
	ErrModuleNotFound = errors.New("module not found")
	// ErrInvalidImportPath is returned when an import path is invalid.
	ErrInvalidImportPath = errors.New("invalid import path")
	// ErrBareImportNotSupported is returned when a bare import path is used (not yet supported).
	ErrBareImportNotSupported = errors.New("bare imports not supported yet")
)

// Resolve resolves an import path relative to the importer.
// It handles:
//   - Standard library paths starting with "std/" -> returns as-is
//   - Plugin paths starting with "plugin/" -> returns as-is
//   - Absolute paths and relative paths -> resolves to file path
//   - .wasm extension -> returns as-is (WASM plugin)
//   - Other paths -> adds .xxl extension if not present (Xxlang module)
func Resolve(importerPath, importPath string) (string, error) {
	// Check if it's a standard library module
	if strings.HasPrefix(importPath, "std/") {
		return importPath, nil
	}

	// Check if it's a plugin module (by name)
	if strings.HasPrefix(importPath, "plugin/") {
		return importPath, nil
	}

	// Check if it's an absolute path
	if filepath.IsAbs(importPath) {
		return resolveFilePath(importPath), nil
	}

	// Check if it's a relative path
	if !strings.HasPrefix(importPath, "./") && !strings.HasPrefix(importPath, "../") {
		return "", ErrBareImportNotSupported
	}

	// Get directory of importer
	importerDir := filepath.Dir(importerPath)

	// Resolve the path
	resolved := filepath.Join(importerDir, importPath)

	return resolveFilePath(resolved), nil
}

// resolveFilePath handles file path resolution.
// .wasm files are returned as-is, others get .xxl extension added if not present.
func resolveFilePath(path string) string {
	// Clean the path
	path = filepath.Clean(path)

	// .wasm files are WASM plugins, return as-is
	if filepath.Ext(path) == ".wasm" {
		return path
	}

	// Add .xxl extension if not present
	if filepath.Ext(path) != ".xxl" {
		path += ".xxl"
	}

	return path
}
