// benchmarks/benchmark_test.go
// Benchmark tests comparing xxlang with Go baseline
package benchmarks

import (
	"fmt"
	"testing"
	"time"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// runXxlangCodeWithFlags compiles and runs xxlang code with custom optimization flags
func runXxlangCodeWithFlags(code string, flags compiler.OptimizationFlags) (time.Duration, error) {
	start := time.Now()

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return 0, fmt.Errorf("parse error: %s", p.Errors()[0])
	}

	c := compiler.NewWithOptions(flags)
	if err := c.Compile(program); err != nil {
		return 0, err
	}

	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// runXxlangCode compiles and runs xxlang code, Returns elapsed time.
func runXxlangCode(code string) (time.Duration, error) {
	start := time.Now()

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return 0, fmt.Errorf("parse error: %s", p.Errors()[0])
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		return 0, err
	}

	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// ============================================================
// Go Baseline Benchmarks
// ============================================================

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func BenchmarkGoFib10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(10)
	}
}

func BenchmarkGoFib20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(20)
	}
}

func BenchmarkGoFib30(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(30)
	}
}

func BenchmarkGoFib35(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(35)
	}
}

// ============================================================
// Xxlang Benchmarks
// ============================================================

func BenchmarkXxlangFib10(b *testing.B) {
	code := `func fib(n) { if (n <= 1) { return n } return fib(n - 1) + fib(n - 2) } fib(10)`
	for i := 0; i < b.N; i++ {
		runXxlangCode(code)
	}
}

func BenchmarkXxlangFib15(b *testing.B) {
	code := `func fib(n) { if (n <= 1) { return n } return fib(n - 1) + fib(n - 2) } fib(15)`
	for i := 0; i < b.N; i++ {
		runXxlangCode(code)
	}
}

func BenchmarkXxlangFib20(b *testing.B) {
	code := `func fib(n) { if (n <= 1) { return n } return fib(n - 1) + fib(n - 2) } fib(20)`
	for i := 0; i < b.N; i++ {
		runXxlangCode(code)
	}
}

// Loop benchmark
func BenchmarkGoLoopSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 10000; j++ {
			sum += j
		}
	}
}

func BenchmarkXxlangLoopSum(b *testing.B) {
	code := `
var sum = 0
for (var i = 0; i < 10000; i = i + 1) {
    sum = sum + i
}
`
	for i := 0; i < b.N; i++ {
		runXxlangCode(code)
	}
}

// Array operations benchmark
func BenchmarkGoArraySum(b *testing.B) {
	arr := make([]int, 1000)
	for i := range arr {
		arr[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, v := range arr {
			sum += v
		}
	}
}

func BenchmarkXxlangArraySum(b *testing.B) {
	code := `
var arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
var sum = 0
for (var i = 0; i < len(arr); i = i + 1) {
    sum = sum + arr[i]
}
`
	for i := 0; i < b.N; i++ {
		runXxlangCode(code)
	}
}

// Function call overhead benchmark
func add(a, b int) int {
	return a + b
}

func BenchmarkGoFunctionCalls(b *testing.B) {
	sum := 0
	for i := 0; i < b.N; i++ {
		sum = add(sum, i)
	}
}

func BenchmarkXxlangFunctionCalls(b *testing.B) {
	code := `
func add(a, b) {
    return a + b
}
var sum = 0
for (var i = 0; i < 1000; i = i + 1) {
    sum = add(sum, i)
}
`
	for i := 0; i < b.N; i++ {
		runXxlangCode(code)
	}
}

// ============================================================
// Compilation-Only Benchmarks (to measure compile time)
// ============================================================

func BenchmarkXxlangCompileFib(b *testing.B) {
	code := `func fib(n) { if (n <= 1) { return n } return fib(n - 1) + fib(n - 2) }`
	for i := 0; i < b.N; i++ {
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()
		c := compiler.New()
		c.Compile(program)
	}
}

func BenchmarkXxlangCompileComplex(b *testing.B) {
	code := `
class Counter {
    func init(start) {
        this.count = start
    }
    func increment() {
        this.count = this.count + 1
        return this.count
    }
    func get() {
        return this.count
    }
}

var c = new Counter(0)
for (var i = 0; i < 100; i = i + 1) {
    c.increment()
}
`
	for i := 0; i < b.N; i++ {
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()
		c := compiler.New()
		c.Compile(program)
	}
}

// ============================================================
// Function Inlining Benchmarks
// ============================================================

// BenchmarkXxlangInlineSimpleFunction tests a simple add function that should be inlined
// A function is inlineable if:
// - It has a single return expression
// - No free variables (not a closure)
// - No side effects (no assignments, function calls, etc.)
func BenchmarkXxlangInlineSimpleFunction_WithInlining(b *testing.B) {
	// Simple function: return a + b (inlineable)
	code := `
func add(a, b) { return a + b }
var sum = 0
for (var i = 0; i < 10000; i = i + 1) {
    sum = add(sum, i)
}
`
	flags := compiler.DefaultOptimizations()
	for i := 0; i < b.N; i++ {
		runXxlangCodeWithFlags(code, flags)
	}
}

func BenchmarkXxlangInlineSimpleFunction_NoInlining(b *testing.B) {
	// Same code, but inlining disabled
	code := `
func add(a, b) { return a + b }
var sum = 0
for (var i = 0; i < 10000; i = i + 1) {
    sum = add(sum, i)
}
`
	flags := compiler.NoInliningOptimizations()
	for i := 0; i < b.N; i++ {
		runXxlangCodeWithFlags(code, flags)
	}
}

// BenchmarkXxlangInlineMultiFunc tests multiple small inlineable functions
func BenchmarkXxlangInlineMultiFunc_WithInlining(b *testing.B) {
	code := `
func double(x) { return x + x }
func square(x) { return x * x }
func add(a, b) { return a + b }
var result = 0
for (var i = 0; i < 5000; i = i + 1) {
    result = add(double(i), square(i % 10))
}
`
	flags := compiler.DefaultOptimizations()
	for i := 0; i < b.N; i++ {
		runXxlangCodeWithFlags(code, flags)
	}
}

func BenchmarkXxlangInlineMultiFunc_NoInlining(b *testing.B) {
	code := `
func double(x) { return x + x }
func square(x) { return x * x }
func add(a, b) { return a + b }
var result = 0
for (var i = 0; i < 5000; i = i + 1) {
    result = add(double(i), square(i % 10))
}
`
	flags := compiler.NoInliningOptimizations()
	for i := 0; i < b.N; i++ {
		runXxlangCodeWithFlags(code, flags)
	}
}

// BenchmarkXxlangInlineNestedCalls tests nested inlineable function calls
func BenchmarkXxlangInlineNestedCalls_WithInlining(b *testing.B) {
	code := `
func inc(x) { return x + 1 }
func add(a, b) { return a + b }
var sum = 0
for (var i = 0; i < 10000; i = i + 1) {
    sum = add(inc(sum), i)
}
`
	flags := compiler.DefaultOptimizations()
	for i := 0; i < b.N; i++ {
		runXxlangCodeWithFlags(code, flags)
	}
}

func BenchmarkXxlangInlineNestedCalls_NoInlining(b *testing.B) {
	code := `
func inc(x) { return x + 1 }
func add(a, b) { return a + b }
var sum = 0
for (var i = 0; i < 10000; i = i + 1) {
    sum = add(inc(sum), i)
}
`
	flags := compiler.NoInliningOptimizations()
	for i := 0; i < b.N; i++ {
		runXxlangCodeWithFlags(code, flags)
	}
}
