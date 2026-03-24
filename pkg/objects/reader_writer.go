// pkg/objects/reader_writer.go
// Reader and Writer objects for streaming I/O operations.
package objects

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// Reader wraps an io.Reader for streaming read operations.
// It implements io.ReadCloser interface.
type Reader struct {
	Value   io.Reader
	bufReader *bufio.Reader // Lazy-initialized for line reading
}

// getBufReader returns a bufio.Reader, creating one if needed.
func (r *Reader) getBufReader() *bufio.Reader {
	if r.bufReader == nil && r.Value != nil {
		r.bufReader = bufio.NewReader(r.Value)
	}
	return r.bufReader
}

// Type returns the object type.
func (r *Reader) Type() ObjectType { return ReaderType }

// TypeTag returns the type tag for fast type checking.
func (r *Reader) TypeTag() TypeTag { return TagReader }

// Inspect returns a string representation of the Reader.
func (r *Reader) Inspect() string {
	if r.Value == nil {
		return "[reader nil]"
	}
	return "[reader]"
}

// ToBool converts the Reader to a boolean (true if not nil).
func (r *Reader) ToBool() *Bool {
	return &Bool{Value: r.Value != nil}
}

// HashKey returns a hash key for the Reader.
func (r *Reader) HashKey() HashKey {
	return HashKey{Type: ReaderType, Value: 0}
}

// Read reads up to n bytes from the reader and returns as byte array.
func (r *Reader) Read(n int) Object {
	if r.Value == nil {
		return newError("reader is nil")
	}

	buf := make([]byte, n)
	numRead, err := r.Value.Read(buf)
	if err != nil && err != io.EOF {
		return newError("read failed: %v", err)
	}

	// Return as array of integers (bytes)
	elements := make([]Object, numRead)
	for i := 0; i < numRead; i++ {
		elements[i] = NewInt(int64(buf[i]))
	}
	return NewArray(elements)
}

// ReadStr reads up to n bytes and returns as string.
func (r *Reader) ReadStr(n int) Object {
	if r.Value == nil {
		return newError("reader is nil")
	}

	buf := make([]byte, n)
	numRead, err := r.Value.Read(buf)
	if err != nil && err != io.EOF {
		return newError("read failed: %v", err)
	}

	return NewString(string(buf[:numRead]))
}

// ReadAll reads all remaining content and returns as string.
func (r *Reader) ReadAllStr() Object {
	if r.Value == nil {
		return newError("reader is nil")
	}

	data, err := io.ReadAll(r.Value)
	if err != nil {
		return newError("readAll failed: %v", err)
	}

	return NewString(string(data))
}

// ReadAllBytes reads all remaining content and returns as byte array.
func (r *Reader) ReadAllBytes() Object {
	if r.Value == nil {
		return newError("reader is nil")
	}

	data, err := io.ReadAll(r.Value)
	if err != nil {
		return newError("readAll failed: %v", err)
	}

	elements := make([]Object, len(data))
	for i, b := range data {
		elements[i] = NewInt(int64(b))
	}
	return NewArray(elements)
}

// ReadLine reads a single line from the reader.
// Returns null on EOF, or error on failure.
func (r *Reader) ReadLine() Object {
	if r.Value == nil {
		return newError("reader is nil")
	}

	br := r.getBufReader()
	if br == nil {
		return newError("reader is nil")
	}

	line, err := br.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			if len(line) == 0 {
				return NULL
			}
			// Return remaining content even on EOF
			return NewString(line)
		}
		return newError("readLine failed: %v", err)
	}

	// Remove trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	// Remove carriage return if present (Windows line endings)
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	return NewString(line)
}

// Close closes the reader if it implements io.Closer.
func (r *Reader) Close() Object {
	if r.Value == nil {
		return NULL
	}

	closer, ok := r.Value.(io.Closer)
	if !ok {
		return newError("reader does not support close")
	}

	if err := closer.Close(); err != nil {
		return newError("close failed: %v", err)
	}

	r.Value = nil
	return NULL
}

// GetMember returns a member by name for script access.
func (r *Reader) GetMember(name string) Object {
	switch name {
	case "read":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("read() takes exactly 1 argument")
			}
			n, ok := args[0].(*Int)
			if !ok {
				return newError("read() requires an integer argument")
			}
			return r.Read(int(n.Value))
		}}
	case "readStr":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("readStr() takes exactly 1 argument")
			}
			n, ok := args[0].(*Int)
			if !ok {
				return newError("readStr() requires an integer argument")
			}
			return r.ReadStr(int(n.Value))
		}}
	case "readAllStr":
		return &Builtin{Fn: func(args ...Object) Object {
			return r.ReadAllStr()
		}}
	case "readAllBytes":
		return &Builtin{Fn: func(args ...Object) Object {
			return r.ReadAllBytes()
		}}
	case "readLine":
		return &Builtin{Fn: func(args ...Object) Object {
			return r.ReadLine()
		}}
	case "close":
		return &Builtin{Fn: func(args ...Object) Object {
			return r.Close()
		}}
	}
	return NULL
}

// Writer wraps an io.Writer for streaming write operations.
type Writer struct {
	Value io.Writer
}

// Type returns the object type.
func (w *Writer) Type() ObjectType { return WriterType }

// TypeTag returns the type tag for fast type checking.
func (w *Writer) TypeTag() TypeTag { return TagWriter }

// Inspect returns a string representation of the Writer.
func (w *Writer) Inspect() string {
	if w.Value == nil {
		return "[writer nil]"
	}
	return "[writer]"
}

// ToBool converts the Writer to a boolean (true if not nil).
func (w *Writer) ToBool() *Bool {
	return &Bool{Value: w.Value != nil}
}

// HashKey returns a hash key for the Writer.
func (w *Writer) HashKey() HashKey {
	return HashKey{Type: WriterType, Value: 0}
}

// Write writes bytes to the writer.
func (w *Writer) Write(data []byte) Object {
	if w.Value == nil {
		return newError("writer is nil")
	}

	n, err := w.Value.Write(data)
	if err != nil {
		return newError("write failed: %v", err)
	}

	return NewInt(int64(n))
}

// WriteStr writes a string to the writer.
func (w *Writer) WriteStr(s string) Object {
	if w.Value == nil {
		return newError("writer is nil")
	}

	n, err := w.Value.Write([]byte(s))
	if err != nil {
		return newError("writeStr failed: %v", err)
	}

	return NewInt(int64(n))
}

// WriteBytes writes a byte array (array of integers) to the writer.
func (w *Writer) WriteBytes(arr *Array) Object {
	if w.Value == nil {
		return newError("writer is nil")
	}

	data := make([]byte, len(arr.Elements))
	for i, elem := range arr.Elements {
		b, ok := elem.(*Int)
		if !ok {
			return newError("writeBytes requires array of integers (0-255)")
		}
		if b.Value < 0 || b.Value > 255 {
			return newError("writeBytes: byte value out of range (0-255)")
		}
		data[i] = byte(b.Value)
	}

	n, err := w.Value.Write(data)
	if err != nil {
		return newError("writeBytes failed: %v", err)
	}

	return NewInt(int64(n))
}

// Close closes the writer if it implements io.Closer.
func (w *Writer) Close() Object {
	if w.Value == nil {
		return NULL
	}

	closer, ok := w.Value.(io.Closer)
	if !ok {
		return newError("writer does not support close")
	}

	if err := closer.Close(); err != nil {
		return newError("close failed: %v", err)
	}

	w.Value = nil
	return NULL
}

// GetMember returns a member by name for script access.
func (w *Writer) GetMember(name string) Object {
	switch name {
	case "write":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("write() takes exactly 1 argument")
			}
			// Accept byte array
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("write() requires an array argument")
			}
			return w.WriteBytes(arr)
		}}
	case "writeStr":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("writeStr() takes exactly 1 argument")
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("writeStr() requires a string argument")
			}
			return w.WriteStr(s.Value)
		}}
	case "writeBytes":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("writeBytes() takes exactly 1 argument")
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("writeBytes() requires an array argument")
			}
			return w.WriteBytes(arr)
		}}
	case "close":
		return &Builtin{Fn: func(args ...Object) Object {
			return w.Close()
		}}
	}
	return NULL
}

// NewReader creates a new Reader object.
func NewReader(r io.Reader) *Reader {
	return &Reader{Value: r}
}

// NewWriter creates a new Writer object.
func NewWriter(w io.Writer) *Writer {
	return &Writer{Value: w}
}

// IoCopy copies data from reader to writer.
// This is exposed as a builtin function.
func IoCopy(dst *Writer, src *Reader) Object {
	if dst == nil || dst.Value == nil {
		return newError("destination writer is nil")
	}
	if src == nil || src.Value == nil {
		return newError("source reader is nil")
	}

	n, err := io.Copy(dst.Value, src.Value)
	if err != nil {
		return newError("ioCopy failed: %v", err)
	}

	return NewInt(n)
}

// NewStringReader creates an io.Reader from a string.
// This is useful for creating a Reader from a string value.
func NewStringReader(s string) io.Reader {
	return strings.NewReader(s)
}

// NewBytesReader creates an io.Reader from a byte slice.
// This is useful for creating a Reader from byte array data.
func NewBytesReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}
