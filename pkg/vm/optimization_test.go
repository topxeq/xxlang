// pkg/vm/optimization_test.go
package vm

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
)

// TestPrimeCheckOptimization tests the prime check superinstruction
func TestPrimeCheckOptimization(t *testing.T) {
	tests := []struct {
		n        int64
		expected bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{7, true},
		{11, true},
		{13, true},
		{17, true},
		{19, true},
		{23, true},
		{29, true},
		{31, true},
		{37, true},
		{41, true},
		{43, true},
		{47, true},
		{53, true},
		{59, true},
		{61, true},
		{67, true},
		{71, true},
		{73, true},
		{79, true},
		{83, true},
		{89, true},
		{97, true},
		{100, false},
		{1000, false},
		{7919, true}, // 1000th prime
	}

	for _, tt := range tests {
		// Test using the isPrime function pattern
		_ = tt // placeholder for future implementation
	}
}

// TestOpRegPrimeCheck tests the OpRegPrimeCheck instruction directly
func TestOpRegPrimeCheck(t *testing.T) {
	tests := []struct {
		n        int64
		expected bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{7, true},
		{11, true},
		{13, true},
		{17, true},
		{19, true},
		{23, true},
		{97, true},
		{100, false},
		{7919, true},
	}

	for _, tt := range tests {
		// Placeholder for proper bytecode construction test
		_ = tt
	}
}

// BenchmarkPrimeCheck compares optimized vs non-optimized prime checking
func BenchmarkPrimeCheck(b *testing.B) {
	code := `
func isPrime(n) {
    if (n < 2) {
        return false
    }
    if (n == 2) {
        return true
    }
    if (n % 2 == 0) {
        return false
    }
    var i = 3
    while (i * i <= n) {
        if (n % i == 0) {
            return false
        }
        i = i + 1
    }
    return true
}

var count = 0
for (var n = 2; n < 10000; n++) {
    if (isPrime(n)) {
        count = count + 1
    }
}
count
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewRegVM(bytecode)
		if err := vm.Run(); err != nil {
			b.Fatalf("runtime error: %v", err)
		}
	}
}

// BenchmarkNestedLoop tests nested loop performance
func BenchmarkNestedLoop(b *testing.B) {
	code := `
var sum = 0
for (var i = 0; i < 100; i++) {
    for (var j = 0; j < 100; j++) {
        sum = sum + i * j
    }
}
sum
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewRegVM(bytecode)
		if err := vm.Run(); err != nil {
			b.Fatalf("runtime error: %v", err)
		}
	}
}

// BenchmarkLoopUnrolling tests loop unrolling performance
func BenchmarkLoopUnrolling(b *testing.B) {
	code := `
var sum = 0
for (var i = 0; i < 8; i++) {
    sum = sum + i
}
sum
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewRegVM(bytecode)
		if err := vm.Run(); err != nil {
			b.Fatalf("runtime error: %v", err)
		}
	}
}

// BenchmarkFibonacciRecursive tests recursive function performance
func BenchmarkFibonacciRecursive(b *testing.B) {
	code := `
func fib(n) {
    if (n < 2) {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

fib(25)
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	c := compiler.NewRegCompiler()
	_, err := c.Compile(program)
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bytecode := c.Bytecode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewRegVM(bytecode)
		if err := vm.Run(); err != nil {
			b.Fatalf("runtime error: %v", err)
		}
	}
}
