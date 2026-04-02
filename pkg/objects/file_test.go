package objects

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Helper to create a File object from a fresh temp file.
func newTestFile(t *testing.T, dir, name string, mode FileMode) *File {
	t.Helper()
	path := filepath.Join(dir, name)
	fh, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() { _ = fh.Close() })
	return NewFile(fh, path, mode)
}

func TestFileBasicIOAndState(t *testing.T) {
	dir := t.TempDir()
	f := newTestFile(t, dir, "sample.txt", ModeRW)
	defer func() { _ = f.Close() }()

	// WriteString and Tell
	n, err := f.WriteString("Hello")
	if err != nil || n != 5 {
		t.Fatalf("WriteString failed: n=%d err=%v", n, err)
	}
	if got := f.Tell(); got != 5 {
		t.Fatalf("Tell() after write should be 5, got %d", got)
	}

	// Read back from start
	if _, err := f.Seek(0, SeekStart); err != nil {
		t.Fatalf("Seek start failed: %v", err)
	}
	data, err := f.Read(5)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(data) != "Hello" {
		t.Fatalf("expected 'Hello', got %q", string(data))
	}

	// Stat and Truncate
	if fi, err := f.Stat(); err != nil {
		t.Fatalf("Stat error: %v", err)
	} else if fi.Size() != 5 {
		t.Fatalf("expected size 5, got %d", fi.Size())
	}
	if err := f.Truncate(3); err != nil {
		t.Fatalf("Truncate error: %v", err)
	}
	if fi, err := f.Stat(); err != nil {
		t.Fatalf("Stat after truncate error: %v", err)
	} else if fi.Size() != 3 {
		t.Fatalf("expected size 3 after truncate, got %d", fi.Size())
	}

	// ReadAll after seek to start
	if _, err := f.Seek(0, SeekStart); err != nil {
		t.Fatalf("Seek start failed: %v", err)
	}
	dataAll, err := f.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(dataAll) != "Hel" {
		t.Fatalf("ReadAll expected 'Hel', got %q", string(dataAll))
	}

	// Write more and verify final size
	if _, err := f.Seek(0, SeekEnd); err != nil {
		t.Fatalf("SeekEnd failed: %v", err)
	}
	if n, err := f.WriteString("OK"); err != nil || n != 2 {
		t.Fatalf("WriteString after end failed: n=%d err=%v", n, err)
	}
	if fi, err := f.Stat(); err != nil {
		t.Fatalf("Stat after final write: %v", err)
	} else if fi.Size() != 5 {
		t.Fatalf("expected final size 5, got %d", fi.Size())
	}

	// Flush and Close
	if err := f.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if f.IsOpen() {
		t.Fatalf("expected file to be closed after Close()")
	}
}

func TestFileReadLineAndReadVariants(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	content := "one\ntwo\nthree"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fh, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	t.Cleanup(func() { _ = fh.Close() })
	f := NewFile(fh, path, ModeRW)
	t.Cleanup(func() { _ = f.Close() })

	// Read first line
	line, err := f.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine1 error: %v", err)
	}
	if line != "one" {
		t.Fatalf("ReadLine1 expected 'one', got %q", line)
	}
	// Read second line
	line, err = f.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine2 error: %v", err)
	}
	if line != "two" {
		t.Fatalf("ReadLine2 expected 'two', got %q", line)
	}
	// Third line without trailing newline
	line, err = f.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine3 expected data, got error: %v", err)
	}
	if line != "three" {
		t.Fatalf("ReadLine3 expected 'three', got %q", line)
	}
	// EOF afterwards
	_, err = f.ReadLine()
	if err != io.EOF {
		t.Fatalf("expected EOF after reading all lines, got: %v", err)
	}
}

func TestFileInfoHelpersAndIsRegularIsSymlink(t *testing.T) {
	// Create a temporary file and get FileInfo
	dir := t.TempDir()
	path := filepath.Join(dir, "info.txt")
	if err := os.WriteFile(path, []byte("hi"), 0644); err != nil {
		t.Fatalf("write info file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	fi := NewFileInfo(info, path)
	// Type and TypeTag for FileInfo
	if fi.Type() != FileInfoType {
		t.Fatalf("expected FileInfoType, got %v", fi.Type())
	}
	if fi.TypeTag() != TagFileInfo {
		t.Fatalf("expected TagFileInfo, got %v", fi.TypeTag())
	}
	// IsRegular should be true for regular file
	if !fi.IsRegular() {
		t.Fatalf("expected IsRegular to be true for a regular file")
	}
	if fi.IsSymlink() {
		t.Fatalf("expected IsSymlink to be false for a regular file")
	}
	// Get mod time string and unix
	_ = fi.GetModTimeString()
	_ = fi.GetModTimeUnix()
	_ = fi.GetModeString()

	// NewFileInfo with a directory should report IsDir and IsRegular false
	dirInfo, _ := os.Stat(dir)
	dirFi := NewFileInfo(dirInfo, dir)
	if dirFi.IsRegular() {
		t.Fatalf("expected IsRegular to be false for a directory")
	}
	// Create a symlink test if supported
	if runtime.GOOS != "windows" {
		// Create a temporary target file
		target := filepath.Join(dir, "target.txt")
		if err := os.WriteFile(target, []byte("x"), 0644); err == nil {
			link := filepath.Join(dir, "link.txt")
			if err := os.Symlink(target, link); err == nil {
				info2, _ := os.Lstat(link)
				fi2 := NewFileInfo(info2, link)
				if !fi2.IsSymlink() {
					t.Fatalf("expected symlink flag on Lstat info")
				}
			}
		}
	}
}

func TestGetModTimeAndModeStringsForFileInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "m.txt")
	if err := os.WriteFile(path, []byte("ok"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	fi := NewFileInfo(info, path)
	if got := fi.GetModTimeString(); got != fi.ModTime.Format("2006-01-02 15:04:05") {
		t.Fatalf("GetModTimeString mismatch: %q != %q", got, fi.ModTime.Format("2006-01-02 15:04:05"))
	}
	if fi.GetModTimeUnix() != fi.ModTime.UnixMilli() {
		t.Fatalf("GetModTimeUnix mismatch: %d != %d", fi.GetModTimeUnix(), fi.ModTime.UnixMilli())
	}
	if fi.GetModeString() != fmt.Sprintf("%03o", fi.Mode) {
		t.Fatalf("GetModeString mismatch: %q != %q", fi.GetModeString(), fmt.Sprintf("%03o", fi.Mode))
	}
	// IsRegular/IsSymlink for the file should be true/false as appropriate
	if !fi.IsRegular() {
		t.Fatalf("expected IsRegular to be true for a regular file")
	}
	if fi.IsSymlink() {
		t.Fatalf("unexpected symlink flag for regular file")
	}
}

func TestHashKeyNonZeroAndTypes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "h.bin")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	fh, _ := os.OpenFile(path, os.O_RDWR, 0644)
	t.Cleanup(func() { _ = fh.Close() })
	f := NewFile(fh, path, ModeRW)
	t.Cleanup(func() { _ = f.Close() })
	kh := f.HashKey()
	if kh.Type != FileType {
		t.Fatalf("HashKey type mismatch: %v", kh.Type)
	}
	if kh.Value == 0 {
		t.Fatalf("HashKey should have a non-zero value")
	}
}
