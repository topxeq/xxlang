// tests/cli_test.go
// Integration tests for the xxlang command-line interface
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================
// CLI Test Helpers
// ============================================

// buildXxlang builds the xxlang binary for testing
func buildXxlang(t *testing.T) string {
	t.Helper()

	// Create temp directory for binary
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "xxlang")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/xxlang")
	cmd.Dir = getProjectRoot()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build xxlang: %v\n%s", err, output)
	}

	return binPath
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Walk up from current directory to find go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

// createTestFile creates a temporary .xxl file with the given content
func createTestFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.xxl")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	return path
}

// runXxlang runs the xxlang binary with the given file
func runXxlang(t *testing.T, binPath, filePath string) (string, error) {
	t.Helper()

	cmd := exec.Command(binPath, filePath)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ============================================
// CLI Tests
// ============================================

func TestCLIBasicExecution(t *testing.T) {
	binPath := buildXxlang(t)

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "integer output",
			code:     "print(42)",
			expected: "42",
		},
		{
			name:     "string output",
			code:     `print("hello world")`,
			expected: "hello world",
		},
		{
			name:     "arithmetic",
			code:     "print(2 + 3 * 4)",
			expected: "14",
		},
		{
			name:     "function call",
			code:     "func add(a, b) { return a + b } print(add(5, 7))",
			expected: "12",
		},
		{
			name:     "boolean",
			code:     "print(true && false)",
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTestFile(t, tt.code)
			output, err := runXxlang(t, binPath, filePath)
			if err != nil {
				t.Fatalf("execution failed: %v\n%s", err, output)
			}
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestCLIFileNotFound(t *testing.T) {
	binPath := buildXxlang(t)

	_, err := runXxlang(t, binPath, "/nonexistent/path/file.xxl")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCLISyntaxError(t *testing.T) {
	binPath := buildXxlang(t)

	// Invalid syntax
	code := "var x = "
	filePath := createTestFile(t, code)

	_, err := runXxlang(t, binPath, filePath)
	if err == nil {
		t.Error("expected error for syntax error")
	}
}

func TestCLIRuntimeError(t *testing.T) {
	binPath := buildXxlang(t)

	// Division by zero
	code := "var x = 1 / 0"
	filePath := createTestFile(t, code)

	output, err := runXxlang(t, binPath, filePath)
	if err == nil {
		t.Error("expected error for runtime error")
	}
	if !strings.Contains(output, "error") && !strings.Contains(output, "Error") {
		t.Errorf("expected error message in output, got %q", output)
	}
}

func TestCLIMultiStatement(t *testing.T) {
	binPath := buildXxlang(t)

	code := `
var a = 10
var b = 20
var c = a + b
print(c)
`
	filePath := createTestFile(t, code)
	output, err := runXxlang(t, binPath, filePath)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !strings.Contains(output, "30") {
		t.Errorf("expected output to contain '30', got %q", output)
	}
}

func TestCLILoops(t *testing.T) {
	binPath := buildXxlang(t)

	code := `
var sum = 0
for (var i = 1; i <= 5; i = i + 1) {
    sum = sum + i
}
print(sum)
`
	filePath := createTestFile(t, code)
	output, err := runXxlang(t, binPath, filePath)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !strings.Contains(output, "15") {
		t.Errorf("expected output to contain '15', got %q", output)
	}
}

func TestCLIFibonacci(t *testing.T) {
	binPath := buildXxlang(t)

	code := `
func fib(n) {
    if (n <= 1) {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}
print(fib(15))
`
	filePath := createTestFile(t, code)
	output, err := runXxlang(t, binPath, filePath)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !strings.Contains(output, "610") {
		t.Errorf("expected output to contain '610', got %q", output)
	}
}

func TestCLIClosures(t *testing.T) {
	binPath := buildXxlang(t)

	code := `
func makeCounter() {
    var count = 0
    func counter() {
        count = count + 1
        return count
    }
    return counter
}

var c = makeCounter()
print(c())
print(c())
print(c())
`
	filePath := createTestFile(t, code)
	output, err := runXxlang(t, binPath, filePath)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Should output 1, 2, 3
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") || !strings.Contains(output, "3") {
		t.Errorf("expected output to contain '1', '2', '3', got %q", output)
	}
}

func TestCLIArrays(t *testing.T) {
	binPath := buildXxlang(t)

	code := `
var arr = [1, 2, 3, 4, 5]
var sum = 0
for (var i = 0; i < len(arr); i = i + 1) {
    sum = sum + arr[i]
}
print(sum)
`
	filePath := createTestFile(t, code)
	output, err := runXxlang(t, binPath, filePath)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !strings.Contains(output, "15") {
		t.Errorf("expected output to contain '15', got %q", output)
	}
}

func TestCLIMaps(t *testing.T) {
	binPath := buildXxlang(t)

	code := `
var person = {
    "name": "Alice",
    "age": 30
}
print(person["name"])
`
	filePath := createTestFile(t, code)
	output, err := runXxlang(t, binPath, filePath)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !strings.Contains(output, "Alice") {
		t.Errorf("expected output to contain 'Alice', got %q", output)
	}
}

func TestCLIBuiltinFunctions(t *testing.T) {
	binPath := buildXxlang(t)

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"len string", `print(len("hello"))`, "5"},
		{"len array", `print(len([1, 2, 3]))`, "3"},
		{"typeOf int", `print(typeOf(42))`, "INT"},
		{"typeOf string", `print(typeOf("hello"))`, "STRING"},
		{"upper", `print(upper("hello"))`, "HELLO"},
		{"lower", `print(lower("HELLO"))`, "hello"},
		{"abs", `print(abs(-42))`, "42"},
		{"sqrt", `print(sqrt(16))`, "4"},
		{"pow", `print(pow(2, 8))`, "256"},
		{"min", `print(min(3, 7))`, "3"},
		{"max", `print(max(3, 7))`, "7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTestFile(t, tt.code)
			output, err := runXxlang(t, binPath, filePath)
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestCLIComplexProgram(t *testing.T) {
	binPath := buildXxlang(t)

	// Bubble sort implementation
	code := `
func bubbleSort(arr) {
    var n = len(arr)
    for (var i = 0; i < n - 1; i = i + 1) {
        for (var j = 0; j < n - i - 1; j = j + 1) {
            if (arr[j] > arr[j + 1]) {
                var temp = arr[j]
                arr[j] = arr[j + 1]
                arr[j + 1] = temp
            }
        }
    }
    return arr
}

var numbers = [64, 34, 25, 12, 22, 11, 90]
var sorted = bubbleSort(numbers)
print(sorted[0])
print(sorted[len(sorted) - 1])
`
	filePath := createTestFile(t, code)
	output, err := runXxlang(t, binPath, filePath)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// First element should be 11, last should be 90
	if !strings.Contains(output, "11") || !strings.Contains(output, "90") {
		t.Errorf("expected sorted array to start with 11 and end with 90, got %q", output)
	}
}
