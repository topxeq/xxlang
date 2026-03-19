// benchmarks/full_comparison_test.go
// Full comparison benchmarks: Stack VM vs Register VM vs Go vs Python
package benchmarks

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/lexer"
	"github.com/topxeq/xxlang/pkg/parser"
	"github.com/topxeq/xxlang/pkg/vm"
)

// runStackVM runs code using the stack-based VM
func runStackVM(input string) error {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil
	}
	c := compiler.New()
	if err := c.Compile(program); err != nil {
		return err
	}
	v := vm.New(c.Bytecode())
	return v.Run()
}

// runRegisterVM runs code using the register-based VM
func runRegisterVM(input string) error {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil
	}
	c := compiler.NewRegCompiler()
	if _, err := c.Compile(program); err != nil {
		return err
	}
	v := vm.NewRegVM(c.Bytecode())
	return v.Run()
}

// ============================================================
// Fibonacci Benchmarks
// ============================================================

var fibCode15 = `
func fib(n) {
	if (n < 2) { return n }
	return fib(n - 1) + fib(n - 2)
}
fib(15)
`

var fibCode20 = `
func fib(n) {
	if (n < 2) { return n }
	return fib(n - 1) + fib(n - 2)
}
fib(20)
`

var fibIterCode = `
func fib(n) {
	var a = 0
	var b = 1
	for (var i = 0; i < n; i = i + 1) {
		var temp = a + b
		a = b
		b = temp
	}
	return a
}
fib(30)
`

func BenchmarkStackVM_Fibonacci15(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(fibCode15)
	}
}

func BenchmarkRegisterVM_Fibonacci15(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(fibCode15)
	}
}

func BenchmarkGo_Fibonacci15(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goFib(15)
	}
}

func BenchmarkStackVM_Fibonacci20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(fibCode20)
	}
}

func BenchmarkRegisterVM_Fibonacci20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(fibCode20)
	}
}

func BenchmarkGo_Fibonacci20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goFib(20)
	}
}

func BenchmarkStackVM_FibonacciIterative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(fibIterCode)
	}
}

func BenchmarkRegisterVM_FibonacciIterative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(fibIterCode)
	}
}

func BenchmarkGo_FibonacciIterative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goFibIter(30)
	}
}

func goFib(n int) int {
	if n < 2 {
		return n
	}
	return goFib(n-1) + goFib(n-2)
}

func goFibIter(n int) int {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}

// ============================================================
// Factorial Benchmark
// ============================================================

var factorialCode = `
func fact(n) {
	if (n <= 1) { return 1 }
	return n * fact(n - 1)
}
fact(12)
`

func BenchmarkStackVM_Factorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(factorialCode)
	}
}

func BenchmarkRegisterVM_Factorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(factorialCode)
	}
}

func BenchmarkGo_Factorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		goFact(12)
	}
}

func goFact(n int) int {
	if n <= 1 {
		return 1
	}
	return n * goFact(n-1)
}

// ============================================================
// Loop Benchmarks
// ============================================================

var forLoop100Code = `
var sum = 0
for (var i = 0; i < 100; i = i + 1) {
	sum = sum + i
}
`

var forLoop1000Code = `
var sum = 0
for (var i = 0; i < 1000; i = i + 1) {
	sum = sum + i
}
`

var whileLoopCode = `
var i = 0
var sum = 0
while (i < 100) {
	sum = sum + i
	i = i + 1
}
`

func BenchmarkStackVM_ForLoop100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(forLoop100Code)
	}
}

func BenchmarkRegisterVM_ForLoop100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(forLoop100Code)
	}
}

func BenchmarkGo_ForLoop100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 100; j++ {
			sum += j
		}
	}
}

func BenchmarkStackVM_ForLoop1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(forLoop1000Code)
	}
}

func BenchmarkRegisterVM_ForLoop1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(forLoop1000Code)
	}
}

func BenchmarkGo_ForLoop1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 1000; j++ {
			sum += j
		}
	}
}

func BenchmarkStackVM_WhileLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(whileLoopCode)
	}
}

func BenchmarkRegisterVM_WhileLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(whileLoopCode)
	}
}

func BenchmarkGo_WhileLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := 0
		sum := 0
		for j < 100 {
			sum += j
			j++
		}
	}
}

// ============================================================
// Function Call Benchmarks
// ============================================================

var functionCallsCode = `
func add(a, b) { return a + b }
func mul(a, b) { return a * b }
var sum = 0
for (var i = 0; i < 100; i = i + 1) {
	sum = add(sum, i)
	sum = mul(sum, 1)
}
`

func BenchmarkStackVM_FunctionCalls(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(functionCallsCode)
	}
}

func BenchmarkRegisterVM_FunctionCalls(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(functionCallsCode)
	}
}

func BenchmarkGo_FunctionCalls(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 100; j++ {
			sum = goAdd(sum, j)
			sum = goMul(sum, 1)
		}
	}
}

func goAdd(a, b int) int { return a + b }
func goMul(a, b int) int { return a * b }

// ============================================================
// Arithmetic Benchmarks
// ============================================================

var arithmeticCode = `
var a = 10
var b = 20
var c = a + b
var d = c * 2
var e = d - 5
var f = e / 3
`

var intensiveArithmeticCode = `
var x = 0
for (var i = 0; i < 1000; i = i + 1) {
	x = x + i * 2 - i / 2
	x = x % 1000
}
`

var nestedExpressionsCode = `
var a = 1 + 2 * 3 - 4 / 2
var c = 10
var d = (a + 5) * (c - a) / (5 + 1)
`

func BenchmarkStackVM_Arithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(arithmeticCode)
	}
}

func BenchmarkRegisterVM_Arithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(arithmeticCode)
	}
}

func BenchmarkGo_Arithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := 10
		b := 20
		c := a + b
		d := c * 2
		e := d - 5
		_ = e / 3
	}
}

func BenchmarkStackVM_IntensiveArithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(intensiveArithmeticCode)
	}
}

func BenchmarkRegisterVM_IntensiveArithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(intensiveArithmeticCode)
	}
}

func BenchmarkGo_IntensiveArithmetic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x := 0
		for j := 0; j < 1000; j++ {
			x = x + j*2 - j/2
			x = x % 1000
		}
	}
}

func BenchmarkStackVM_NestedExpressions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(nestedExpressionsCode)
	}
}

func BenchmarkRegisterVM_NestedExpressions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(nestedExpressionsCode)
	}
}

func BenchmarkGo_NestedExpressions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := 1 + 2*3 - 4/2
		c := 10
		_ = (a + 5) * (c - a) / (5 + 1)
	}
}

// ============================================================
// Comparison Benchmarks
// ============================================================

var comparisonsCode = `
var a = 10
var b = 20
var c = 0
for (var i = 0; i < 100; i = i + 1) {
	if (a < b) { c = c + 1 }
	if (a > b) { c = c + 1 }
	if (a == b) { c = c + 1 }
	if (a != b) { c = c + 1 }
}
`

func BenchmarkStackVM_Comparisons(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(comparisonsCode)
	}
}

func BenchmarkRegisterVM_Comparisons(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(comparisonsCode)
	}
}

func BenchmarkGo_Comparisons(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := 10
		b := 20
		c := 0
		for j := 0; j < 100; j++ {
			if a < b {
				c++
			}
			if a > b {
				c++
			}
			if a == b {
				c++
			}
			if a != b {
				c++
			}
		}
	}
}

// ============================================================
// Algorithm Benchmarks
// ============================================================

var primeCheckCode = `
func isPrime(n) {
	if (n < 2) { return false }
	var i = 2
	while (i * i <= n) {
		if (n % i == 0) { return false }
		i = i + 1
	}
	return true
}
for (var j = 0; j < 100; j = j + 1) {
	isPrime(j)
}
`

var bubbleSortCode = `
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
}
bubbleSort([5, 3, 8, 4, 2, 7, 1, 10, 6, 9])
`

func BenchmarkStackVM_PrimeCheck100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(primeCheckCode)
	}
}

func BenchmarkRegisterVM_PrimeCheck100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(primeCheckCode)
	}
}

func BenchmarkGo_PrimeCheck100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			goIsPrime(j)
		}
	}
}

func goIsPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func BenchmarkStackVM_BubbleSort10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(bubbleSortCode)
	}
}

func BenchmarkRegisterVM_BubbleSort10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(bubbleSortCode)
	}
}

func BenchmarkGo_BubbleSort10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := []int{5, 3, 8, 4, 2, 7, 1, 10, 6, 9}
		goBubbleSort(arr)
	}
}

func goBubbleSort(arr []int) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

// ============================================================
// Array and String Benchmarks
// ============================================================

var arraySumCode = `
var arr = []
for (var i = 0; i < 1000; i = i + 1) {
	arr[i] = i
}
var sum = 0
for (var i = 0; i < 1000; i = i + 1) {
	sum = sum + arr[i]
}
`

var stringConcatCode = `
var s = ""
for (var i = 0; i < 100; i = i + 1) {
	s = s + "x"
}
`

func BenchmarkStackVM_ArraySum1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(arraySumCode)
	}
}

func BenchmarkRegisterVM_ArraySum1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(arraySumCode)
	}
}

func BenchmarkGo_ArraySum1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := make([]int, 1000)
		for j := 0; j < 1000; j++ {
			arr[j] = j
		}
		sum := 0
		for j := 0; j < 1000; j++ {
			sum += arr[j]
		}
	}
}

func BenchmarkStackVM_StringConcat100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runStackVM(stringConcatCode)
	}
}

func BenchmarkRegisterVM_StringConcat100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRegisterVM(stringConcatCode)
	}
}

func BenchmarkGo_StringConcat100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < 100; j++ {
			s += "x"
		}
	}
}
