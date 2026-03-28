// pkg/objects/scanner_test.go
package objects

import (
	"strings"
	"testing"
)

func TestScannerNext(t *testing.T) {
	input := "hello world\nfoo bar"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Test next()
	result := scanner.next()
	if result == NULL {
		t.Fatal("expected 'hello', got NULL")
	}
	if str, ok := result.(*String); ok {
		if str.Value != "hello" {
			t.Errorf("expected 'hello', got '%s'", str.Value)
		}
	} else {
		t.Errorf("expected String, got %T", result)
	}

	// Test second next()
	result = scanner.next()
	if str, ok := result.(*String); ok {
		if str.Value != "world" {
			t.Errorf("expected 'world', got '%s'", str.Value)
		}
	}

	// Test nextLine()
	result = scanner.nextLine()
	if str, ok := result.(*String); ok {
		if str.Value != "foo bar" {
			t.Errorf("expected 'foo bar', got '%s'", str.Value)
		}
	}

	// Test EOF
	result = scanner.next()
	if result != NULL {
		t.Errorf("expected NULL on EOF, got %v", result)
	}
}

func TestScannerNextInt(t *testing.T) {
	input := "42\n-100\nnotanumber"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Test valid integer
	result := scanner.nextInt()
	if n, ok := result.(*Int); ok {
		if n.Value != 42 {
			t.Errorf("expected 42, got %d", n.Value)
		}
	} else {
		t.Errorf("expected Int, got %T", result)
	}

	// Test negative integer
	result = scanner.nextInt()
	if n, ok := result.(*Int); ok {
		if n.Value != -100 {
			t.Errorf("expected -100, got %d", n.Value)
		}
	}

	// Test invalid integer
	result = scanner.nextInt()
	if _, ok := result.(*Error); !ok {
		t.Errorf("expected Error for invalid integer, got %T", result)
	}
}

func TestScannerNextFloat(t *testing.T) {
	input := "3.14\n-2.5\nnotanumber"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Test valid float
	result := scanner.nextFloat()
	if f, ok := result.(*Float); ok {
		if f.Value != 3.14 {
			t.Errorf("expected 3.14, got %f", f.Value)
		}
	} else {
		t.Errorf("expected Float, got %T", result)
	}

	// Test negative float
	result = scanner.nextFloat()
	if f, ok := result.(*Float); ok {
		if f.Value != -2.5 {
			t.Errorf("expected -2.5, got %f", f.Value)
		}
	}

	// Test invalid float
	result = scanner.nextFloat()
	if _, ok := result.(*Error); !ok {
		t.Errorf("expected Error for invalid float, got %T", result)
	}
}

func TestScannerNextBool(t *testing.T) {
	input := "true\nfalse\n1\n0\nnotabool"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Test true
	result := scanner.nextBool()
	if result != TRUE {
		t.Errorf("expected TRUE, got %v", result)
	}

	// Test false
	result = scanner.nextBool()
	if result != FALSE {
		t.Errorf("expected FALSE, got %v", result)
	}

	// Test 1
	result = scanner.nextBool()
	if result != TRUE {
		t.Errorf("expected TRUE for '1', got %v", result)
	}

	// Test 0
	result = scanner.nextBool()
	if result != FALSE {
		t.Errorf("expected FALSE for '0', got %v", result)
	}

	// Test invalid
	result = scanner.nextBool()
	if _, ok := result.(*Error); !ok {
		t.Errorf("expected Error for invalid boolean, got %T", result)
	}
}

func TestScannerHasNext(t *testing.T) {
	input := "hello world"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Should have next
	result := scanner.hasNext()
	if result != TRUE {
		t.Errorf("expected TRUE, got %v", result)
	}

	// Read all content
	scanner.next()
	scanner.next()

	// Should not have next
	result = scanner.hasNext()
	if result != FALSE {
		t.Errorf("expected FALSE after EOF, got %v", result)
	}
}

func TestScannerSkipLine(t *testing.T) {
	input := "line1\nline2\nline3"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Skip first line
	scanner.skipLine()

	// Read second line
	result := scanner.nextLine()
	if str, ok := result.(*String); ok {
		if str.Value != "line2" {
			t.Errorf("expected 'line2', got '%s'", str.Value)
		}
	}
}

func TestScannerType(t *testing.T) {
	input := "test"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	if scanner.Type() != ScannerType {
		t.Errorf("expected ScannerType, got %s", scanner.Type())
	}

	if scanner.TypeTag() != TagScanner {
		t.Errorf("expected TagScanner, got %d", scanner.TypeTag())
	}

	if scanner.Inspect() != "[scanner]" {
		t.Errorf("expected '[scanner]', got '%s'", scanner.Inspect())
	}
}

func TestStandaloneFunctions(t *testing.T) {
	// Test reading multiple tokens
	inputReader := strings.NewReader("apple banana cherry")
	scanner := NewScanner(inputReader)

	tokens := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		result := scanner.next()
		if result == NULL {
			break
		}
		if str, ok := result.(*String); ok {
			tokens = append(tokens, str.Value)
		}
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}

	// Test line reading
	inputReader2 := strings.NewReader("a,b,c")
	scanner2 := NewScanner(inputReader2)
	line := scanner2.nextLine()
	if str, ok := line.(*String); ok {
		if str.Value != "a,b,c" {
			t.Errorf("expected 'a,b,c', got '%s'", str.Value)
		}
	}

	// Test reading two values
	inputReader3 := strings.NewReader("first second")
	scanner3 := NewScanner(inputReader3)
	v1 := scanner3.next()
	v2 := scanner3.next()
	if v1 == NULL || v2 == NULL {
		t.Error("expected two values from scan")
	}
}

func TestScannerClose(t *testing.T) {
	// Test close on non-closable reader
	input := "test"
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Close should return NULL (no error) for non-closable readers
	result := scanner.close()
	if result != NULL {
		t.Errorf("expected NULL for close on non-closable reader, got %v", result)
	}
}

func TestScannerEOF(t *testing.T) {
	input := ""
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Test EOF on empty input
	result := scanner.next()
	if result != NULL {
		t.Errorf("expected NULL on empty input, got %v", result)
	}

	result = scanner.nextLine()
	if result != NULL {
		t.Errorf("expected NULL on empty input for nextLine, got %v", result)
	}

	result = scanner.nextInt()
	if result != NULL {
		t.Errorf("expected NULL on empty input for nextInt, got %v", result)
	}
}

func TestScannerWithWhitespace(t *testing.T) {
	// Test next() skips whitespace properly
	input := "  hello   world  "
	reader := strings.NewReader(input)
	scanner := NewScanner(reader)

	// Test that leading/trailing whitespace is handled
	result := scanner.next()
	if str, ok := result.(*String); ok {
		if str.Value != "hello" {
			t.Errorf("expected 'hello', got '%s'", str.Value)
		}
	}

	result = scanner.next()
	if str, ok := result.(*String); ok {
		if str.Value != "world" {
			t.Errorf("expected 'world', got '%s'", str.Value)
		}
	}

	// Test nextLine preserves whitespace
	input2 := "  foo  bar  \n  baz  "
	reader2 := strings.NewReader(input2)
	scanner2 := NewScanner(reader2)

	// nextLine should preserve leading/trailing whitespace
	line := scanner2.nextLine()
	if str, ok := line.(*String); ok {
		expected := "  foo  bar  "
		if str.Value != expected {
			t.Errorf("expected '%s', got '%s'", expected, str.Value)
		}
	}
}
