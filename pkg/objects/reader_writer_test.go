// pkg/objects/reader_writer_test.go
// Tests for Reader and Writer objects
package objects

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewReader(t *testing.T) {
	r := strings.NewReader("test content")
	reader := NewReader(r)

	if reader == nil {
		t.Fatal("expected non-nil Reader")
	}
	if reader.Type() != ReaderType {
		t.Errorf("expected type READER, got %s", reader.Type())
	}
	if reader.TypeTag() != TagReader {
		t.Errorf("expected TypeTag TagReader, got %d", reader.TypeTag())
	}
}

func TestReader_Inspect(t *testing.T) {
	r := strings.NewReader("test")
	reader := NewReader(r)

	inspect := reader.Inspect()
	if inspect == "" {
		t.Errorf("expected non-empty Inspect")
	}
	if !strings.Contains(inspect, "reader") {
		t.Errorf("expected Inspect to contain 'reader', got %s", inspect)
	}

	// Test nil reader
	nilReader := &Reader{Value: nil}
	if nilReader.Inspect() != "[reader nil]" {
		t.Errorf("expected '[reader nil]', got %s", nilReader.Inspect())
	}
}

func TestReader_ToBool(t *testing.T) {
	r := strings.NewReader("test")
	reader := NewReader(r)

	if !reader.ToBool().Value {
		t.Errorf("Reader with value should return true")
	}

	nilReader := &Reader{Value: nil}
	if nilReader.ToBool().Value {
		t.Errorf("Reader with nil value should return false")
	}
}

func TestReader_HashKey(t *testing.T) {
	r := strings.NewReader("test")
	reader := NewReader(r)

	key := reader.HashKey()
	if key.Type != ReaderType {
		t.Errorf("expected HashKey.Type READER, got %s", key.Type)
	}
}

func TestReader_Read(t *testing.T) {
	r := strings.NewReader("Hello, World!")
	reader := NewReader(r)

	result := reader.Read(5)
	if result.Type() != ArrayType {
		t.Fatalf("expected Array, got %s", result.Type())
	}

	arr := result.(*Array)
	if len(arr.Elements) != 5 {
		t.Errorf("expected 5 elements, got %d", len(arr.Elements))
	}

	// Check first byte is 'H' (72)
	if arr.Elements[0].(*Int).Value != 72 {
		t.Errorf("expected first byte to be 72 ('H'), got %d", arr.Elements[0].(*Int).Value)
	}
}

func TestReader_ReadStr(t *testing.T) {
	r := strings.NewReader("Hello, World!")
	reader := NewReader(r)

	result := reader.ReadStr(5)
	if result.Type() != StringType {
		t.Fatalf("expected String, got %s", result.Type())
	}

	str := result.(*String)
	if str.Value != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", str.Value)
	}
}

func TestReader_ReadAllStr(t *testing.T) {
	r := strings.NewReader("Hello, World!")
	reader := NewReader(r)

	result := reader.ReadAllStr()
	if result.Type() != StringType {
		t.Fatalf("expected String, got %s", result.Type())
	}

	str := result.(*String)
	if str.Value != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", str.Value)
	}
}

func TestReader_ReadAllBytes(t *testing.T) {
	r := strings.NewReader("Hello")
	reader := NewReader(r)

	result := reader.ReadAllBytes()
	if result.Type() != ArrayType {
		t.Fatalf("expected Array, got %s", result.Type())
	}

	arr := result.(*Array)
	if len(arr.Elements) != 5 {
		t.Errorf("expected 5 elements, got %d", len(arr.Elements))
	}
}

func TestReader_ReadLine(t *testing.T) {
	r := strings.NewReader("Line 1\nLine 2\nLine 3")
	reader := NewReader(r)

	line := reader.ReadLine()
	if line.Type() != StringType {
		t.Fatalf("expected String, got %s", line.Type())
	}
	if line.(*String).Value != "Line 1" {
		t.Errorf("expected 'Line 1', got '%s'", line.(*String).Value)
	}

	line = reader.ReadLine()
	if line.(*String).Value != "Line 2" {
		t.Errorf("expected 'Line 2', got '%s'", line.(*String).Value)
	}
}

func TestReader_NilReader(t *testing.T) {
	reader := &Reader{Value: nil}

	result := reader.Read(5)
	if result.Type() != ErrorType {
		t.Errorf("expected Error for nil reader, got %s", result.Type())
	}

	result = reader.ReadStr(5)
	if result.Type() != ErrorType {
		t.Errorf("expected Error for nil reader, got %s", result.Type())
	}

	result = reader.ReadAllStr()
	if result.Type() != ErrorType {
		t.Errorf("expected Error for nil reader, got %s", result.Type())
	}
}

// Writer tests
func TestNewWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	if writer == nil {
		t.Fatal("expected non-nil Writer")
	}
	if writer.Type() != WriterType {
		t.Errorf("expected type WRITER, got %s", writer.Type())
	}
	if writer.TypeTag() != TagWriter {
		t.Errorf("expected TypeTag TagWriter, got %d", writer.TypeTag())
	}
}

func TestWriter_Inspect(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	inspect := writer.Inspect()
	if inspect == "" {
		t.Errorf("expected non-empty Inspect")
	}
	if !strings.Contains(inspect, "writer") {
		t.Errorf("expected Inspect to contain 'writer', got %s", inspect)
	}

	// Test nil writer
	nilWriter := &Writer{Value: nil}
	if nilWriter.Inspect() != "[writer nil]" {
		t.Errorf("expected '[writer nil]', got %s", nilWriter.Inspect())
	}
}

func TestWriter_ToBool(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	if !writer.ToBool().Value {
		t.Errorf("Writer with value should return true")
	}

	nilWriter := &Writer{Value: nil}
	if nilWriter.ToBool().Value {
		t.Errorf("Writer with nil value should return false")
	}
}

func TestWriter_HashKey(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	key := writer.HashKey()
	if key.Type != WriterType {
		t.Errorf("expected HashKey.Type WRITER, got %s", key.Type)
	}
}

func TestWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	result := writer.Write([]byte("Hello"))
	if result.Type() != IntType {
		t.Fatalf("expected Int, got %s", result.Type())
	}

	n := result.(*Int).Value
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}

	if buf.String() != "Hello" {
		t.Errorf("expected buffer to contain 'Hello', got '%s'", buf.String())
	}
}

func TestWriter_WriteBytes(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	// Create array of byte values
	elements := []Object{
		NewInt(72),  // H
		NewInt(101), // e
		NewInt(108), // l
		NewInt(108), // l
		NewInt(111), // o
	}
	arr := NewArray(elements)

	result := writer.WriteBytes(arr)
	if result.Type() != IntType {
		t.Fatalf("expected Int, got %s", result.Type())
	}

	n := result.(*Int).Value
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}

	if buf.String() != "Hello" {
		t.Errorf("expected buffer to contain 'Hello', got '%s'", buf.String())
	}
}

func TestWriter_WriteStr(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	result := writer.WriteStr("Hello")
	if result.Type() != IntType {
		t.Fatalf("expected Int, got %s", result.Type())
	}

	n := result.(*Int).Value
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}

	if buf.String() != "Hello" {
		t.Errorf("expected buffer to contain 'Hello', got '%s'", buf.String())
	}
}

func TestWriter_NilWriter(t *testing.T) {
	writer := &Writer{Value: nil}

	result := writer.Write([]byte("test"))
	if result.Type() != ErrorType {
		t.Errorf("expected Error for nil writer, got %s", result.Type())
	}

	result = writer.WriteBytes(NewArray([]Object{NewInt(1)}))
	if result.Type() != ErrorType {
		t.Errorf("expected Error for nil writer, got %s", result.Type())
	}
}
