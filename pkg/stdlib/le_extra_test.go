// pkg/stdlib/le_extra_test.go
// Additional tests for le (LineEditor) module to increase coverage
package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// TestLe_Line tests GetLine() method on LineEditor
func TestLe_Line(t *testing.T) {
	le := callLeFunc("fromText", String("line1\nline2\nline3")).(*objects.LineEditor)

	// Test valid line numbers (1-indexed)
	line1, ok := le.GetLine(1)
	if !ok || line1 != "line1" {
		t.Fatalf("GetLine(1) = %q, ok=%v, want \"line1\", true", line1, ok)
	}
	line2, ok := le.GetLine(2)
	if !ok || line2 != "line2" {
		t.Fatalf("GetLine(2) = %q, ok=%v, want \"line2\", true", line2, ok)
	}
	// Out of range
	_, ok = le.GetLine(0)
	if ok {
		t.Error("GetLine(0) should return false")
	}
	_, ok = le.GetLine(4)
	if ok {
		t.Error("GetLine(4) should return false")
	}
}

// TestLe_LineCount tests LineCount()
func TestLe_LineCount(t *testing.T) {
	le := callLeFunc("fromText", String("a\nb\nc\nd")).(*objects.LineEditor)
	count := le.LineCount()
	if count != 4 {
		t.Fatalf("LineCount() = %d, want 4", count)
	}
}

// TestLe_SetLine tests SetLine()
func TestLe_SetLine(t *testing.T) {
	le := callLeFunc("fromText", String("old1\nold2")).(*objects.LineEditor)
	ok := le.SetLine(1, "new1")
	if !ok {
		t.Fatalf("SetLine returned false")
	}
	line, _ := le.GetLine(1)
	if line != "new1" {
		t.Fatalf("after SetLine, GetLine(1) = %q, want \"new1\"", line)
	}
	// Out of range error
	ok = le.SetLine(3, "oob")
	if ok {
		t.Error("SetLine out of range should return false")
	}
}

// TestLe_InsertLine tests InsertLine()
func TestLe_InsertLine(t *testing.T) {
	le := callLeFunc("fromText", String("a\nc")).(*objects.LineEditor)
	ok := le.InsertLine(2, "b") // insert before line 2
	if !ok {
		t.Fatalf("InsertLine returned false")
	}
	if le.LineCount() != 3 {
		t.Fatalf("LineCount after insert = %d, want 3", le.LineCount())
	}
	line2, _ := le.GetLine(2)
	if line2 != "b" {
		t.Fatalf("GetLine(2) after insert = %q, want \"b\"", line2)
	}
	// Check that c moved to line 3
	line3, _ := le.GetLine(3)
	if line3 != "c" {
		t.Fatalf("GetLine(3) after insert = %q, want \"c\"", line3)
	}
}

// TestLe_DeleteLine tests DeleteLine()
func TestLe_DeleteLine(t *testing.T) {
	le := callLeFunc("fromText", String("a\nb\nc")).(*objects.LineEditor)
	ok := le.DeleteLine(2)
	if !ok {
		t.Fatalf("DeleteLine returned false")
	}
	if le.LineCount() != 2 {
		t.Fatalf("LineCount after delete = %d, want 2", le.LineCount())
	}
	_, ok = le.GetLine(2)
	if !ok {
		t.Fatalf("GetLine(2) after delete should exist (now 'c'), returned false")
	}
	line2, _ := le.GetLine(2)
	if line2 != "c" {
		t.Fatalf("GetLine(2) = %q, want \"c\"", line2)
	}
	// Delete out of range
	ok = le.DeleteLine(0)
	if ok {
		t.Error("DeleteLine(0) should return false")
	}
}

// TestLe_Replace tests Replace(old, new string)
func TestLe_Replace(t *testing.T) {
	le := callLeFunc("fromText", String("foo bar foo")).(*objects.LineEditor)
	count := le.Replace("foo", "baz")
	if count != 2 {
		t.Fatalf("Replace count = %d, want 2", count)
	}
	text := le.ToText()
	if text != "baz bar baz" {
		t.Fatalf("after Replace, text = %q, want \"baz bar baz\"", text)
	}
	// Replace non-existent returns 0
	le2 := callLeFunc("fromText", String("hello")).(*objects.LineEditor)
	count = le2.Replace("xyz", "abc")
	if count != 0 {
		t.Fatalf("Replace non-existent should return 0, got %d", count)
	}
}

// TestLe_Search tests FindAll() method (Search equivalent)
func TestLe_Search(t *testing.T) {
	le := callLeFunc("fromText", String("hello\nworld\nhello again")).(*objects.LineEditor)
	// Search for "hello"
	results := le.FindAll("hello")
	if len(results) != 2 {
		t.Fatalf("FindAll results length = %d, want 2", len(results))
	}
	// Verify content
	if results[0] != "hello" || results[1] != "hello again" {
		t.Fatalf("FindAll results = %v, want ['hello', 'hello again']", results)
	}
}

// TestLe_WriteToFile tests SaveAs() method (WriteToFile equivalent)
func TestLe_WriteToFile(t *testing.T) {
	le := callLeFunc("fromText", String("content1\ncontent2")).(*objects.LineEditor)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.txt")
	err := le.SaveAs(path)
	if err != nil {
		t.Fatalf("SaveAs error: %v", err)
	}
	// Read back
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	// Note: SaveAs adds newline after each line, so content ends with \n
	if string(data) != "content1\ncontent2\n" {
		t.Fatalf("file content = %q, want \"content1\\ncontent2\\n\"", string(data))
	}
}

// TestLe_ReadFromFile tests opening a file (ReadFromFile equivalent)
func TestLe_ReadFromFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "in.txt")
	os.WriteFile(path, []byte("x\ny\nz"), 0644)
	// Use open function to read from file
	result := callLeFunc("open", String(path))
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("open returned error: %v", result)
	}
	le, ok := result.(*objects.LineEditor)
	if !ok {
		t.Fatalf("expected LineEditor, got %T", result)
	}
	if le.LineCount() != 3 {
		t.Fatalf("LineCount after open = %d, want 3", le.LineCount())
	}
	line2, _ := le.GetLine(2)
	if line2 != "y" {
		t.Fatalf("GetLine(2) = %q, want \"y\"", line2)
	}
}

// TestLe_Sort tests Sort() method
func TestLe_Sort(t *testing.T) {
	le := callLeFunc("fromText", String("c\na\nb")).(*objects.LineEditor)
	le.Sort()
	text := le.ToText()
	// Should be a\nb\nc
	if text != "a\nb\nc" {
		t.Fatalf("after Sort, text = %q, want \"a\\nb\\nc\"", text)
	}
}

// TestLe_Unique tests Unique() method
func TestLe_Unique(t *testing.T) {
	le := callLeFunc("fromText", String("a\nb\na\nc\nb")).(*objects.LineEditor)
	le.Unique()
	// Text should have duplicates removed, order preserved first occurrences: a, b, c
	text := le.ToText()
	expected := "a\nb\nc"
	if text != expected {
		t.Fatalf("after Unique, text = %q, want %q", text, expected)
	}
	// Calling Unique on already unique should not change anything
	le2 := callLeFunc("fromText", String("a\nb\nc")).(*objects.LineEditor)
	le2.Unique()
	text2 := le2.ToText()
	if text2 != "a\nb\nc" {
		t.Fatalf("Unique on already unique changed text to %q", text2)
	}
}
