// pkg/objects/scanner.go
// Scanner object for reading input from stdin or any io.Reader.
package objects

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Scanner provides convenient methods for reading input.
// It wraps a bufio.Reader for efficient reading.
type Scanner struct {
	reader    *bufio.Reader
	rawReader io.Reader
}

// Type returns the object type.
func (s *Scanner) Type() ObjectType { return ScannerType }

// TypeTag returns the type tag for fast type checking.
func (s *Scanner) TypeTag() TypeTag { return TagScanner }

// Inspect returns a string representation of the Scanner.
func (s *Scanner) Inspect() string {
	return "[scanner]"
}

// ToBool converts the Scanner to a boolean (always true).
func (s *Scanner) ToBool() *Bool {
	return TRUE
}

// HashKey returns a hash key for the Scanner.
func (s *Scanner) HashKey() HashKey {
	return HashKey{Type: ScannerType, Value: 0}
}

// NewScanner creates a new Scanner from an io.Reader.
// If reader is nil, it defaults to os.Stdin.
func NewScanner(reader io.Reader) *Scanner {
	if reader == nil {
		reader = os.Stdin
	}
	return &Scanner{
		reader:    bufio.NewReader(reader),
		rawReader: reader,
	}
}

// getReader returns the underlying bufio.Reader.
func (s *Scanner) getReader() *bufio.Reader {
	if s.reader == nil && s.rawReader != nil {
		s.reader = bufio.NewReader(s.rawReader)
	}
	return s.reader
}

// next reads and returns the next token (whitespace-delimited).
// Returns null on EOF or error.
func (s *Scanner) next() Object {
	r := s.getReader()
	if r == nil {
		return newError("scanner is nil")
	}

	// Skip leading whitespace
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return NULL
			}
			return newError("scan failed: %v", err)
		}
		if !isWhitespace(ch) {
			r.UnreadRune()
			break
		}
	}

	// Read until whitespace
	var sb strings.Builder
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return newError("scan failed: %v", err)
		}
		if isWhitespace(ch) {
			break
		}
		sb.WriteRune(ch)
	}

	if sb.Len() == 0 {
		return NULL
	}

	return NewString(sb.String())
}

// isWhitespace checks if a rune is whitespace.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// nextLine reads and returns the next line.
// Returns null on EOF or error.
func (s *Scanner) nextLine() Object {
	r := s.getReader()
	if r == nil {
		return newError("scanner is nil")
	}

	line, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			if line == "" {
				return NULL
			}
			return NewString(strings.TrimSuffix(line, "\r"))
		}
		return newError("scanLine failed: %v", err)
	}

	// Remove trailing newline and carriage return
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return NewString(line)
}

// nextInt reads the next token and parses it as an integer.
// Returns error if the token is not a valid integer.
func (s *Scanner) nextInt() Object {
	token := s.next()
	if token == NULL {
		return token
	}

	str, ok := token.(*String)
	if !ok {
		return newError("scan failed to get string")
	}

	n, err := strconv.ParseInt(str.Value, 10, 64)
	if err != nil {
		return newError("scanInt: invalid integer '%s'", str.Value)
	}

	return NewInt(n)
}

// nextFloat reads the next token and parses it as a float.
// Returns error if the token is not a valid float.
func (s *Scanner) nextFloat() Object {
	token := s.next()
	if token == NULL {
		return token
	}

	str, ok := token.(*String)
	if !ok {
		return newError("scan failed to get string")
	}

	f, err := strconv.ParseFloat(str.Value, 64)
	if err != nil {
		return newError("scanFloat: invalid float '%s'", str.Value)
	}

	return NewFloat(f)
}

// nextBool reads the next token and parses it as a boolean.
// Accepts "true", "false", "1", "0" (case-insensitive).
func (s *Scanner) nextBool() Object {
	token := s.next()
	if token == NULL {
		return token
	}

	str, ok := token.(*String)
	if !ok {
		return newError("scan failed to get string")
	}

	switch strings.ToLower(str.Value) {
	case "true", "1":
		return TRUE
	case "false", "0":
		return FALSE
	default:
		return newError("scanBool: invalid boolean '%s'", str.Value)
	}
}

// hasNext checks if there is more input to read.
func (s *Scanner) hasNext() Object {
	r := s.getReader()
	if r == nil {
		return FALSE
	}

	// Peek at the next byte
	_, err := r.Peek(1)
	if err != nil {
		if err == io.EOF {
			return FALSE
		}
		return FALSE
	}
	return TRUE
}

// skipLine skips the current line.
func (s *Scanner) skipLine() Object {
	r := s.getReader()
	if r == nil {
		return NULL
	}

	_, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return newError("skipLine failed: %v", err)
	}
	return NULL
}

// close closes the underlying reader if it implements io.Closer.
func (s *Scanner) close() Object {
	if s.rawReader == nil {
		return NULL
	}

	closer, ok := s.rawReader.(io.Closer)
	if !ok {
		return NULL // Not a closer, nothing to do
	}

	if err := closer.Close(); err != nil {
		return newError("close failed: %v", err)
	}

	s.rawReader = nil
	s.reader = nil
	return NULL
}

// GetMember returns a member by name for script access.
func (s *Scanner) GetMember(name string) Object {
	switch name {
	case "next":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("next() takes no arguments")
			}
			return s.next()
		}}
	case "nextLine":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("nextLine() takes no arguments")
			}
			return s.nextLine()
		}}
	case "nextInt":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("nextInt() takes no arguments")
			}
			return s.nextInt()
		}}
	case "nextFloat":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("nextFloat() takes no arguments")
			}
			return s.nextFloat()
		}}
	case "nextBool":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("nextBool() takes no arguments")
			}
			return s.nextBool()
		}}
	case "hasNext":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("hasNext() takes no arguments")
			}
			return s.hasNext()
		}}
	case "skipLine":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("skipLine() takes no arguments")
			}
			return s.skipLine()
		}}
	case "close":
		return &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("close() takes no arguments")
			}
			return s.close()
		}}
	}
	return NULL
}

// ============================================
// Standalone scan functions (not methods)
// ============================================

// Global reader for standalone scan functions
var globalReader = bufio.NewReader(os.Stdin)

// scanToken reads a single whitespace-delimited token from stdin.
func scanToken() (string, error) {
	r := globalReader

	// Skip leading whitespace
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return "", err
		}
		if !isWhitespace(ch) {
			r.UnreadRune()
			break
		}
	}

	// Read until whitespace
	var sb strings.Builder
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if isWhitespace(ch) {
			break
		}
		sb.WriteRune(ch)
	}

	return sb.String(), nil
}

// scanLine reads a line from stdin.
func scanLine() (string, error) {
	line, err := globalReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	// Remove trailing newline and carriage return
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line, nil
}

// Scan reads a line from stdin and returns it as a string.
// If a prompt string is provided, it is printed first.
func Scan(prompt string) Object {
	if prompt != "" {
		fmt.Print(prompt)
	}

	line, err := scanLine()
	if err != nil {
		if err == io.EOF {
			return NULL
		}
		return newError("scan failed: %v", err)
	}

	return NewString(line)
}

// ScanInt reads an integer from stdin.
func ScanInt(prompt string) Object {
	if prompt != "" {
		fmt.Print(prompt)
	}

	token, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NULL
		}
		return newError("scanInt failed: %v", err)
	}

	n, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		return newError("scanInt: invalid integer '%s'", token)
	}

	return NewInt(n)
}

// ScanFloat reads a float from stdin.
func ScanFloat(prompt string) Object {
	if prompt != "" {
		fmt.Print(prompt)
	}

	token, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NULL
		}
		return newError("scanFloat failed: %v", err)
	}

	f, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return newError("scanFloat: invalid float '%s'", token)
	}

	return NewFloat(f)
}

// ScanBool reads a boolean from stdin.
func ScanBool(prompt string) Object {
	if prompt != "" {
		fmt.Print(prompt)
	}

	token, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NULL
		}
		return newError("scanBool failed: %v", err)
	}

	switch strings.ToLower(token) {
	case "true", "1":
		return TRUE
	case "false", "0":
		return FALSE
	default:
		return newError("scanBool: invalid boolean '%s'", token)
	}
}

// ScanN reads n whitespace-delimited tokens from stdin.
func ScanN(n int) Object {
	if n <= 0 {
		return EMPTY_ARRAY
	}

	elements := make([]Object, 0, n)
	for i := 0; i < n; i++ {
		token, err := scanToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return newError("scanN failed: %v", err)
		}
		elements = append(elements, NewString(token))
	}

	return NewArray(elements)
}

// ScanSplit reads a line and splits it by the given separator.
func ScanSplit(sep string) Object {
	line, err := scanLine()
	if err != nil {
		if err == io.EOF {
			return EMPTY_ARRAY
		}
		return newError("scanSplit failed: %v", err)
	}

	if sep == "" {
		// Split by whitespace
		fields := strings.Fields(line)
		elements := make([]Object, len(fields))
		for i, f := range fields {
			elements[i] = NewString(f)
		}
		return NewArray(elements)
	}

	parts := strings.Split(line, sep)
	elements := make([]Object, len(parts))
	for i, p := range parts {
		elements[i] = NewString(strings.TrimSpace(p))
	}
	return NewArray(elements)
}

// Scan2 reads two whitespace-delimited tokens from stdin.
func Scan2() (Object, Object) {
	token1, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NULL, NULL
		}
		return newError("scan2 failed: %v", err), NULL
	}

	token2, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NewString(token1), NULL
		}
		return newError("scan2 failed: %v", err), NULL
	}

	return NewString(token1), NewString(token2)
}

// Scan3 reads three whitespace-delimited tokens from stdin.
func Scan3() (Object, Object, Object) {
	token1, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NULL, NULL, NULL
		}
		return newError("scan3 failed: %v", err), NULL, NULL
	}

	token2, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NewString(token1), NULL, NULL
		}
		return newError("scan3 failed: %v", err), NULL, NULL
	}

	token3, err := scanToken()
	if err != nil {
		if err == io.EOF {
			return NewString(token1), NewString(token2), NULL
		}
		return newError("scan3 failed: %v", err), NULL, NULL
	}

	return NewString(token1), NewString(token2), NewString(token3)
}

// Scanf reads input according to a format string.
// The format string can contain {} placeholders for values.
// Returns an array of parsed values.
func Scanf(format string) Object {
	// Read a line
	line, err := scanLine()
	if err != nil {
		if err == io.EOF {
			return EMPTY_ARRAY
		}
		return newError("scanf failed: %v", err)
	}

	// Count placeholders
	placeholderCount := strings.Count(format, "{}")
	if placeholderCount == 0 {
		return EMPTY_ARRAY
	}

	// Simple implementation: split line by whitespace and return tokens
	// corresponding to the number of placeholders
	fields := strings.Fields(line)
	elements := make([]Object, 0, placeholderCount)
	for i := 0; i < placeholderCount && i < len(fields); i++ {
		elements = append(elements, NewString(fields[i]))
	}

	return NewArray(elements)
}

// ResetGlobalReader resets the global reader for stdin.
// This is useful for testing or when stdin needs to be reinitialized.
func ResetGlobalReader() {
	globalReader = bufio.NewReader(os.Stdin)
}
