// pkg/compiler/sourcemap_test.go
package compiler

import (
	"bytes"
	"testing"
)

func TestSourceMap_New(t *testing.T) {
	sm := NewSourceMap()
	if sm == nil {
		t.Fatal("expected SourceMap, got nil")
	}
	if sm.Locations == nil {
		t.Fatal("expected Locations map to be initialized")
	}
}

func TestSourceMap_AddAndGet(t *testing.T) {
	sm := NewSourceMap()

	sm.Add(10, SourceLocation{Line: 5, Column: 10})

	loc, ok := sm.Get(10)
	if !ok {
		t.Fatal("expected to find location at offset 10")
	}
	if loc.Line != 5 || loc.Column != 10 {
		t.Fatalf("expected (5, 10), got (%d, %d)", loc.Line, loc.Column)
	}

	loc, ok = sm.Get(15)
	if !ok {
		t.Fatal("expected to find closest location")
	}
	if loc.Line != 5 {
		t.Fatalf("expected line 5, got %d", loc.Line)
	}

	_, ok = sm.Get(0)
	if ok {
		t.Fatal("expected not to find location at offset 0")
	}
}

func TestSourceMap_SetSourceFileAndGetLine(t *testing.T) {
	sm := NewSourceMap()
	source := "line1\nline2\nline3"

	sm.SetSourceFile("test.xxl", source)

	if sm.SourceFile != "test.xxl" {
		t.Fatalf("expected 'test.xxl', got %q", sm.SourceFile)
	}
	if len(sm.SourceLines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(sm.SourceLines))
	}
	if sm.GetLine(1) != "line1" {
		t.Errorf("expected 'line1', got %q", sm.GetLine(1))
	}
	if sm.GetLine(3) != "line3" {
		t.Errorf("expected 'line3', got %q", sm.GetLine(3))
	}
	if sm.GetLine(0) != "" {
		t.Errorf("expected empty string for line 0")
	}
	if sm.GetLine(4) != "" {
		t.Errorf("expected empty string for line 4")
	}
}

func TestSourceMap_FormatErrorBasic(t *testing.T) {
	sm := NewSourceMap()
	sm.SetSourceFile("test.xxl", "var x = 42")
	sm.Add(0, SourceLocation{Line: 1, Column: 5})

	err := sm.FormatError(0, "unexpected token")
	if err == "" {
		t.Fatal("expected error message")
	}

	sm2 := NewSourceMap()
	err = sm2.FormatError(100, "some error")
	if err != "some error" {
		t.Errorf("expected 'some error', got %q", err)
	}
}

func TestSourceMap_FormatErrorEdgeCases(t *testing.T) {
	sm := NewSourceMap()
	sm.SetSourceFile("", "var x = 42")
	sm.Add(0, SourceLocation{Line: 1, Column: 5})

	err := sm.FormatError(0, "unexpected token")
	if err == "" {
		t.Fatal("expected error message")
	}

	sm2 := NewSourceMap()
	sm2.SetSourceFile("test.xxl", "var x = 42")
	sm2.Add(100, SourceLocation{Line: 1, Column: 5})

	err = sm2.FormatError(50, "error")
	if err != "error" {
		t.Errorf("expected plain error message, got %q", err)
	}
}

func TestSourceMap_Serialize(t *testing.T) {
	sm := NewSourceMap()
	sm.SetSourceFile("test.xxl", "line1\nline2")
	sm.Add(0, SourceLocation{Line: 1, Column: 1})
	sm.Add(10, SourceLocation{Line: 2, Column: 5})

	var buf bytes.Buffer
	err := sm.Serialize(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sm2, err := DeserializeSourceMap(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm2.SourceFile != "test.xxl" {
		t.Errorf("expected 'test.xxl', got %q", sm2.SourceFile)
	}
	if len(sm2.SourceLines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(sm2.SourceLines))
	}

	loc, ok := sm2.Get(0)
	if !ok {
		t.Fatal("expected to find location at offset 0")
	}
	if loc.Line != 1 || loc.Column != 1 {
		t.Errorf("expected (1, 1), got (%d, %d)", loc.Line, loc.Column)
	}
}

func TestSourceMap_DeserializeEmpty(t *testing.T) {
	sm := NewSourceMap()
	sm.SetSourceFile("", "")

	var buf bytes.Buffer
	err := sm.Serialize(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sm2, err := DeserializeSourceMap(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm2.SourceFile != "" {
		t.Errorf("expected empty source file")
	}
	if len(sm2.SourceLines) != 0 {
		t.Errorf("expected 0 lines")
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"multiple lines", "line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"single line", "line1", []string{"line1"}},
		{"empty", "", []string{}},
		{"trailing newline", "line1\nline2\n", []string{"line1", "line2"}},
		{"empty lines", "line1\n\nline3", []string{"line1", "", "line3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d", len(tt.expected), len(result))
			}
			for i := range tt.expected {
				if result[i] != tt.expected[i] {
					t.Errorf("line %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}
