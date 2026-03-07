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
// It handles relative paths starting with "./" or "../".
// The returned path will have the .xxl extension added if not present.
func Resolve(importerPath, importPath string) (string, error) {
	// Check if it's a relative path
	if !strings.HasPrefix(importPath, "./") && !strings.HasPrefix(importPath, "../") {
		return "", ErrBareImportNotSupported
	}

	// Get directory of importer
	importerDir := filepath.Dir(importerPath)

	// Resolve the path
	resolved := filepath.Join(importerDir, importPath)

	// Add .xxl extension if not present
	if filepath.Ext(resolved) != ".xxl" {
		resolved += ".xxl"
	}

	// Clean the path
	resolved = filepath.Clean(resolved)

	return resolved, nil
}
