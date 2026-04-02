// pkg/stdlib/le_file_ops_test.go
// Tests for le module file operation functions (replaceInFile, sortFile, etc.)
package stdlib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func callLeOpFunc(name string, args ...objects.Object) objects.Object {
	mod := Get("le")
	if mod == nil {
		panic("le module not found")
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		panic("function not found: " + name)
	}
	return fn.Fn(args...)
}

// TestLe_ReplaceInFile tests replaceInFile function.
func TestLe_ReplaceInFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	// Create file with content
	os.WriteFile(filePath, []byte("hello world, hello again"), 0644)

	result := callLeOpFunc("replaceInFile", String(filePath), String("hello"), String("hi"))
	if result != objects.NULL {
		t.Fatalf("replaceInFile should return NULL, got %v", result)
	}

	// Verify replacement
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	// Note: SaveAs adds newline after each line, so content ends with \n
	expected := "hi world, hi again\n"
	if string(content) != expected {
		t.Errorf("after replaceInFile, content = %q, want %q", string(content), expected)
	}
}

// TestLe_ReplaceInFile_Errors tests replaceInFile error cases.
func TestLe_ReplaceInFile_Errors(t *testing.T) {
	// Missing file
	result := callLeOpFunc("replaceInFile", String("/nonexistent/file.txt"), String("a"), String("b"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("replaceInFile with nonexistent file should error")
	}

	// Wrong argument count
	result = callLeOpFunc("replaceInFile", String("path"))
	if _, ok := result.(*objects.Error); !ok {
		t.Error("replaceInFile with 1 arg should error")
	}
}

// TestLe_SortFile tests sortFile function.
func TestLe_SortFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sort.txt")
	// Write unsorted lines
	os.WriteFile(filePath, []byte("z\na\nm\nb\n"), 0644)

	result := callLeOpFunc("sortFile", String(filePath))
	if result != objects.NULL {
		t.Fatalf("sortFile should return NULL, got %v", result)
	}

	content, _ := os.ReadFile(filePath)
	expected := "a\nb\nm\nz\n"
	if string(content) != expected {
		t.Errorf("after sortFile, content = %q, want %q", string(content), expected)
	}
}

// TestLe_SortFileTo tests sortFileTo function.
func TestLe_SortFileTo(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")
	os.WriteFile(src, []byte("z\na\nm\n"), 0644)

	result := callLeOpFunc("sortFileTo", String(src), String(dst))
	if result != objects.NULL {
		t.Fatalf("sortFileTo should return NULL, got %v", result)
	}

	// Source should remain unchanged
	orig, _ := os.ReadFile(src)
	if string(orig) != "z\na\nm\n" {
		t.Error("sortFileTo should not modify source")
	}
	// Destination should be sorted
	sorted, _ := os.ReadFile(dst)
	if string(sorted) != "a\nm\nz\n" {
		t.Errorf("sorted destination = %q, want sorted", string(sorted))
	}
}

// TestLe_UniqueFile tests uniqueFile function.
func TestLe_UniqueFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "uniq.txt")
	os.WriteFile(filePath, []byte("a\nb\na\nc\nb\n"), 0644)

	result := callLeOpFunc("uniqueFile", String(filePath))
	if result != objects.NULL {
		t.Fatalf("uniqueFile should return NULL, got %v", result)
	}

	content, _ := os.ReadFile(filePath)
	expected := "a\nb\nc\n"
	if string(content) != expected {
		t.Errorf("after uniqueFile, content = %q, want %q", string(content), expected)
	}
}

// TestLe_UniqueFileTo tests uniqueFileTo function.
func TestLe_UniqueFileTo(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")
	os.WriteFile(src, []byte("a\nb\na\nc\n"), 0644)

	result := callLeOpFunc("uniqueFileTo", String(src), String(dst))
	if result != objects.NULL {
		t.Fatalf("uniqueFileTo should return NULL, got %v", result)
	}

	// Source unchanged
	orig, _ := os.ReadFile(src)
	if string(orig) != "a\nb\na\nc\n" {
		t.Error("uniqueFileTo modified source")
	}
	// Destination unique
	dest, _ := os.ReadFile(dst)
	if string(dest) != "a\nb\nc\n" {
		t.Errorf("unique destination = %q, want a\\nb\\nc\\n", string(dest))
	}
}

// TestLe_RemoveEmptyLines tests removeEmptyLines function.
func TestLe_RemoveEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.txt")
	os.WriteFile(filePath, []byte("a\n\nb\n\n\nc\n"), 0644)

	result := callLeOpFunc("removeEmptyLines", String(filePath))
	if result != objects.NULL {
		t.Fatalf("removeEmptyLines should return NULL, got %v", result)
	}

	content, _ := os.ReadFile(filePath)
	expected := "a\nb\nc\n"
	if string(content) != expected {
		t.Errorf("after removeEmptyLines, content = %q, want %q", string(content), expected)
	}
}

// TestLe_CountLines tests countLines function.
func TestLe_CountLines(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "count.txt")
	os.WriteFile(filePath, []byte("a\nb\nc\nd\n"), 0644)

	result := callLeOpFunc("countLines", String(filePath))
	if result == nil {
		t.Fatal("countLines returned nil")
	}
	count, ok := result.(*objects.Int)
	if !ok {
		t.Fatalf("expected Int, got %T", result)
	}
	if count.Value != 4 {
		t.Errorf("countLines = %d, want 4", count.Value)
	}
}

// TestLe_Head tests head function.
func TestLe_Head(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "head.txt")
	os.WriteFile(filePath, []byte("1\n2\n3\n4\n5\n"), 0644)

	result := callLeOpFunc("head", String(filePath), Int(2))
	if result == nil {
		t.Fatal("head returned nil")
	}
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("head should return 2 lines, got %d", len(arr.Elements))
	}
	// Check values
	if s, ok := arr.Elements[0].(*objects.String); !ok || s.Value != "1" {
		t.Errorf("first line = %q, want \"1\"", arr.Elements[0])
	}
	if s, ok := arr.Elements[1].(*objects.String); !ok || s.Value != "2" {
		t.Errorf("second line = %q, want \"2\"", arr.Elements[1])
	}
}

// TestLe_Tail tests tail function.
func TestLe_Tail(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "tail.txt")
	os.WriteFile(filePath, []byte("1\n2\n3\n4\n5\n"), 0644)

	result := callLeOpFunc("tail", String(filePath), Int(2))
	if result == nil {
		t.Fatal("tail returned nil")
	}
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("tail should return 2 lines, got %d", len(arr.Elements))
	}
	if s, ok := arr.Elements[0].(*objects.String); !ok || s.Value != "4" {
		t.Errorf("first tail line = %q, want \"4\"", arr.Elements[0])
	}
	if s, ok := arr.Elements[1].(*objects.String); !ok || s.Value != "5" {
		t.Errorf("second tail line = %q, want \"5\"", arr.Elements[1])
	}
}

// TestLe_GrepFile tests grepFile function.
func TestLe_GrepFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "grep.txt")
	os.WriteFile(filePath, []byte("apple\nbanana\napricot\ncherry\n"), 0644)

	result := callLeOpFunc("grepFile", String(filePath), String("ap"))
	if result == nil {
		t.Fatal("grepFile returned nil")
	}
	arr, ok := result.(*objects.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}
	if len(arr.Elements) != 2 {
		t.Errorf("grepFile should return 2 lines, got %d", len(arr.Elements))
	}
	if s, ok := arr.Elements[0].(*objects.String); !ok || s.Value != "apple" {
		t.Errorf("first match = %q, want \"apple\"", arr.Elements[0])
	}
	if s, ok := arr.Elements[1].(*objects.String); !ok || s.Value != "apricot" {
		t.Errorf("second match = %q, want \"apricot\"", arr.Elements[1])
	}
}

// TestLe_GrepFileTo tests grepFileTo function.
func TestLe_GrepFileTo(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")
	os.WriteFile(src, []byte("apple\nbanana\napricot\n"), 0644)

	result := callLeOpFunc("grepFileTo", String(src), String("ap"), String(dst))
	if result != objects.NULL {
		t.Fatalf("grepFileTo should return NULL, got %v", result)
	}
	// Verify destination
	content, _ := os.ReadFile(dst)
	lines := splitLines(string(content))
	if len(lines) != 2 {
		t.Errorf("grepFileTo produced %d lines, want 2", len(lines))
	}
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
