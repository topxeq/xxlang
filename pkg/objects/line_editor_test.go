// pkg/objects/line_editor_test.go
// Tests for LineEditor object.
package objects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLineEditor(t *testing.T) {
	le := NewLineEditor()
	if le == nil {
		t.Fatal("NewLineEditor returned nil")
	}
	if le.LineCount() != 0 {
		t.Errorf("Expected 0 lines, got %d", le.LineCount())
	}
	if !le.IsEmpty() {
		t.Error("Expected IsEmpty to be true")
	}
}

func TestNewLineEditorFromText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantLines int
	}{
		{"empty", "", 0},
		{"single line", "hello", 1},
		{"multiple lines", "one\ntwo\nthree", 3},
		{"trailing newline", "one\ntwo\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			le := NewLineEditorFromText(tt.text)
			if le.LineCount() != tt.wantLines {
				t.Errorf("Expected %d lines, got %d", tt.wantLines, le.LineCount())
			}
		})
	}
}

func TestNewLineEditorFromLines(t *testing.T) {
	lines := []string{"a", "b", "c"}
	le := NewLineEditorFromLines(lines)
	if le.LineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", le.LineCount())
	}
}

func TestLineEditorGetSetLine(t *testing.T) {
	le := NewLineEditorFromText("one\ntwo\nthree")

	// Test GetLine
	line, ok := le.GetLine(1)
	if !ok || line != "one" {
		t.Errorf("GetLine(1) = %q, want 'one'", line)
	}

	// Test negative index
	line, ok = le.GetLine(-1)
	if !ok || line != "three" {
		t.Errorf("GetLine(-1) = %q, want 'three'", line)
	}

	// Test SetLine
	le.SetLine(2, "TWO")
	line, _ = le.GetLine(2)
	if line != "TWO" {
		t.Errorf("After SetLine, GetLine(2) = %q, want 'TWO'", line)
	}

	// Test out of range
	_, ok = le.GetLine(10)
	if ok {
		t.Error("GetLine(10) should return false for out of range")
	}
}

func TestLineEditorAddInsertDelete(t *testing.T) {
	le := NewLineEditor()

	// AddLine
	le.AddLine("first")
	le.AddLine("second")
	if le.LineCount() != 2 {
		t.Errorf("After AddLine, LineCount = %d, want 2", le.LineCount())
	}

	// InsertLine
	le.InsertLine(2, "middle")
	line, _ := le.GetLine(2)
	if line != "middle" {
		t.Errorf("After InsertLine, GetLine(2) = %q, want 'middle'", line)
	}

	// DeleteLine
	le.DeleteLine(2)
	if le.LineCount() != 2 {
		t.Errorf("After DeleteLine, LineCount = %d, want 2", le.LineCount())
	}
}

func TestLineEditorFind(t *testing.T) {
	le := NewLineEditorFromText("apple\nbanana\ncherry\napple\ndate")

	// Find
	found := le.Find("apple")
	if len(found) != 2 {
		t.Errorf("Find('apple') found %d lines, want 2", len(found))
	}
	if found[0] != 1 || found[1] != 4 {
		t.Errorf("Find('apple') = %v, want [1, 4]", found)
	}

	// FindFirst
	first := le.FindFirst("apple")
	if first != 1 {
		t.Errorf("FindFirst('apple') = %d, want 1", first)
	}

	// FindLast
	last := le.FindLast("apple")
	if last != 4 {
		t.Errorf("FindLast('apple') = %d, want 4", last)
	}

	// FindAll
	all := le.FindAll("a")
	if len(all) != 4 {
		t.Errorf("FindAll('a') found %d lines, want 4", len(all))
	}
}

func TestLineEditorReplace(t *testing.T) {
	le := NewLineEditorFromText("hello world\nhello universe")

	// Replace
	count := le.Replace("hello", "hi")
	if count != 2 {
		t.Errorf("Replace count = %d, want 2", count)
	}

	line, _ := le.GetLine(1)
	if line != "hi world" {
		t.Errorf("After Replace, line 1 = %q, want 'hi world'", line)
	}

	// ReplaceFirst
	le2 := NewLineEditorFromText("foo bar foo")
	le2.ReplaceFirst("foo", "baz")
	line, _ = le2.GetLine(1)
	if line != "baz bar foo" {
		t.Errorf("After ReplaceFirst, line = %q, want 'baz bar foo'", line)
	}

	// ReplaceLast
	le3 := NewLineEditorFromText("foo bar foo")
	le3.ReplaceLast("foo", "baz")
	line, _ = le3.GetLine(1)
	if line != "foo bar baz" {
		t.Errorf("After ReplaceLast, line = %q, want 'foo bar baz'", line)
	}
}

func TestLineEditorSort(t *testing.T) {
	le := NewLineEditorFromText("cherry\napple\nbanana")

	le.Sort()
	line, _ := le.GetLine(1)
	if line != "apple" {
		t.Errorf("After Sort, first line = %q, want 'apple'", line)
	}

	le.SortDesc()
	line, _ = le.GetLine(1)
	if line != "cherry" {
		t.Errorf("After SortDesc, first line = %q, want 'cherry'", line)
	}
}

func TestLineEditorUnique(t *testing.T) {
	le := NewLineEditorFromText("a\nb\na\nc\nb")

	le.Unique()
	if le.LineCount() != 3 {
		t.Errorf("After Unique, LineCount = %d, want 3", le.LineCount())
	}
}

func TestLineEditorTrim(t *testing.T) {
	le := NewLineEditorFromText("  hello  \n  world  ")

	le.Trim()
	line, _ := le.GetLine(1)
	if line != "hello" {
		t.Errorf("After Trim, line 1 = %q, want 'hello'", line)
	}
}

func TestLineEditorGrep(t *testing.T) {
	le := NewLineEditorFromText("apple\nbanana\ncherry\napricot")

	filtered := le.Grep("a")
	if filtered.LineCount() != 3 {
		t.Errorf("Grep('a') returned %d lines, want 3", filtered.LineCount())
	}

	filtered = le.GrepNot("a")
	if filtered.LineCount() != 1 {
		t.Errorf("GrepNot('a') returned %d lines, want 1", filtered.LineCount())
	}
}

func TestLineEditorToFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create editor and save
	le := NewLineEditorFromText("line1\nline2\nline3")
	err := le.SaveAs(testFile)
	if err != nil {
		t.Fatalf("SaveAs failed: %v", err)
	}

	// Read back and verify
	le2, err := NewLineEditorFromFile(testFile)
	if err != nil {
		t.Fatalf("NewLineEditorFromFile failed: %v", err)
	}

	if le2.LineCount() != 3 {
		t.Errorf("Read back LineCount = %d, want 3", le2.LineCount())
	}

	line, _ := le2.GetLine(1)
	if line != "line1" {
		t.Errorf("Read back line 1 = %q, want 'line1'", line)
	}
}

func TestLineEditorType(t *testing.T) {
	le := NewLineEditor()
	if le.Type() != LineEditorType {
		t.Errorf("Type() = %s, want %s", le.Type(), LineEditorType)
	}
	if le.TypeTag() != TagLineEditor {
		t.Errorf("TypeTag() = %d, want %d", le.TypeTag(), TagLineEditor)
	}
}

func TestLineEditorToText(t *testing.T) {
	le := NewLineEditorFromText("a\nb\nc")
	text := le.ToText()
	if text != "a\nb\nc" {
		t.Errorf("ToText() = %q, want 'a\\nb\\nc'", text)
	}
}

func TestLineEditorReverse(t *testing.T) {
	le := NewLineEditorFromText("a\nb\nc")
	le.Reverse()

	line, _ := le.GetLine(1)
	if line != "c" {
		t.Errorf("After Reverse, first line = %q, want 'c'", line)
	}
}

func TestLineEditorNumberLines(t *testing.T) {
	le := NewLineEditorFromText("a\nb\nc")
	le.NumberLines(1)

	line, _ := le.GetLine(1)
	if line != "1: a" {
		t.Errorf("After NumberLines(1), first line = %q, want '1: a'", line)
	}
}

func TestLineEditorJoin(t *testing.T) {
	le := NewLineEditorFromText("a\nb\nc")
	result := le.Join(",")

	if result != "a,b,c" {
		t.Errorf("Join(',') = %q, want 'a,b,c'", result)
	}
}

func TestLineEditorSplitLines(t *testing.T) {
	le := NewLineEditorFromText("a,b,c\nd,e,f")
	le.SplitLines(",")

	if le.LineCount() != 6 {
		t.Errorf("After SplitLines, LineCount = %d, want 6", le.LineCount())
	}
}

func TestLineEditorPrefixSuffix(t *testing.T) {
	le := NewLineEditorFromText("a\nb")
	le.Prefix(">")
	le.Suffix("<")

	line, _ := le.GetLine(1)
	if line != ">a<" {
		t.Errorf("After Prefix and Suffix, line = %q, want '>a<'", line)
	}
}

func TestLineEditorRemoveEmpty(t *testing.T) {
	le := NewLineEditorFromText("a\n\nb\n\n\nc")
	le.RemoveEmpty()

	if le.LineCount() != 3 {
		t.Errorf("After RemoveEmpty, LineCount = %d, want 3", le.LineCount())
	}
}

func TestLineEditorFindDupes(t *testing.T) {
	le := NewLineEditorFromText("a\nb\na\nc\nb\nb")
	dupes := le.FindDupes()

	if len(dupes) != 2 {
		t.Errorf("FindDupes found %d duplicates, want 2", len(dupes))
	}
	if dupes["a"] != 2 || dupes["b"] != 3 {
		t.Errorf("FindDupes = %v, want {a: 2, b: 3}", dupes)
	}
}

func TestLineEditorKeepRemoveDupes(t *testing.T) {
	le := NewLineEditorFromText("a\nb\na\nc")
	le.RemoveDupes()

	if le.LineCount() != 2 {
		t.Errorf("After RemoveDupes, LineCount = %d, want 2", le.LineCount())
	}

	le2 := NewLineEditorFromText("a\nb\na\nc")
	le2.KeepDupes()

	// KeepDupes keeps only lines that appear more than once ('a' appears twice)
	// Both occurrences of 'a' are kept, so we expect 2 lines
	if le2.LineCount() != 2 {
		t.Errorf("After KeepDupes, LineCount = %d, want 2", le2.LineCount())
	}
}

func TestLineEditorSortByCol(t *testing.T) {
	le := NewLineEditorFromText("a,3\nb,1\nc,2")

	le.SortByCol(2, ",")
	line, _ := le.GetLine(1)
	if line != "b,1" {
		t.Errorf("After SortByCol, first line = %q, want 'b,1'", line)
	}

	le.SortByColNum(2, ",")
	line, _ = le.GetLine(1)
	if line != "b,1" {
		t.Errorf("After SortByColNum, first line = %q, want 'b,1'", line)
	}
}

func TestLineEditorCaseConversion(t *testing.T) {
	le := NewLineEditorFromText("Hello\nWorld")

	le.ToUpperCase()
	line, _ := le.GetLine(1)
	if line != "HELLO" {
		t.Errorf("After ToUpperCase, line = %q, want 'HELLO'", line)
	}

	le.ToLowerCase()
	line, _ = le.GetLine(1)
	if line != "hello" {
		t.Errorf("After ToLowerCase, line = %q, want 'hello'", line)
	}
}

func TestLineEditorDedent(t *testing.T) {
	le := NewLineEditorFromText("  a\n  b\n    c")
	le.Dedent()

	line, _ := le.GetLine(1)
	if line != "a" {
		t.Errorf("After Dedent, first line = %q, want 'a'", line)
	}
}

func TestLineEditorFromFile(t *testing.T) {
	// Test error case - non-existent file
	_, err := NewLineEditorFromFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLineEditorIsModified(t *testing.T) {
	le := NewLineEditorFromText("test")
	if le.IsModified() {
		t.Error("New editor should not be modified")
	}

	le.AddLine("new")
	if !le.IsModified() {
		t.Error("Editor should be modified after AddLine")
	}

	// Save to temp file
	tmpFile := filepath.Join(os.TempDir(), "line_editor_test.txt")
	defer os.Remove(tmpFile)

	le.SetFilePath(tmpFile)
	le.SaveAs(tmpFile)
	if le.IsModified() {
		t.Error("Editor should not be modified after save")
	}
}

func TestLineEditorGetLines(t *testing.T) {
	le := NewLineEditorFromText("a\nb\nc\nd\ne")

	lines := le.GetLines(2, 4)
	if len(lines) != 3 {
		t.Errorf("GetLines(2,4) returned %d lines, want 3", len(lines))
	}
	if lines[0] != "b" || lines[2] != "d" {
		t.Errorf("GetLines(2,4) = %v, want [b, c, d]", lines)
	}
}

func TestLineEditorDeleteLines(t *testing.T) {
	le := NewLineEditorFromText("a\nb\nc\nd\ne")

	le.DeleteLines(2, 4)
	if le.LineCount() != 2 {
		t.Errorf("After DeleteLines(2,4), LineCount = %d, want 2", le.LineCount())
	}
}