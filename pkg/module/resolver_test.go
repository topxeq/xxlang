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
	_, err := Resolve("/project/main.xxl", "std/math")
	if err == nil {
		t.Error("expected error for bare import path")
	}
}
