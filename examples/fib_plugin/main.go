// examples/fib_plugin/main.go
// 演示如何使用 Xxlang 作为第三方库编写插件
package main

import (
	"fmt"
	"time"

	// 导入 Xxlang 解释器
	"github.com/topxeq/xxlang/pkg/interpreter"

	// 导入我们的插件 - 这会触发 init() 自动注册
	_ "github.com/topxeq/xxlang/examples/fib_plugin/plugin"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Xxlang 插件系统演示")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("本示例展示如何编写第三方插件供 Xxlang 调用")
	fmt.Println()

	// 创建解释器
	interp := interpreter.New(interpreter.WithStdlib())

	fmt.Println("==============================================")
	fmt.Println("1. 测试插件功能")
	fmt.Println("==============================================")
	fmt.Println()

	// 测试 Xxlang 调用 Go 插件
	testCode := `
// 导入我们编写的 fib 插件
import "plugin/fib"

println("插件版本: " + fib.version)

// 测试高性能斐波那契计算
println("")
println("=== fib.fast (O(n) 算法) ===")
println("fib.fast(10) = " + fib.fast(10).toStr())
println("fib.fast(50) = " + fib.fast(50).toStr())

// 测试矩阵快速幂
println("")
println("=== fib.matrix (O(log n) 算法) ===")
println("fib.matrix(10) = " + fib.matrix(10).toStr())
println("fib.matrix(50) = " + fib.matrix(50).toStr())

// 测试批量计算
println("")
println("=== fib.range_ (批量计算) ===")
var fibs = fib.range_(10)
println("fib.range_(10) 返回 " + len(fibs).toStr() + " 个结果")
println("fib(0) 到 fib(10): " + fibs.toStr())

// 测试斐波那契数检测
println("")
println("=== fib.isFib (检测斐波那契数) ===")
println("isFib(13) = " + fib.isFib(13).toStr())
println("isFib(14) = " + fib.isFib(14).toStr())
println("isFib(21) = " + fib.isFib(21).toStr())
println("isFib(22) = " + fib.isFib(22).toStr())
`

	result, err := interp.Eval(testCode)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("2. 性能对比测试")
	fmt.Println("==============================================")
	fmt.Println()

	// 性能对比
	perfTest := `
import "plugin/fib"

// Xxlang 朴素递归
func fibNaive(n) {
    if (n <= 1) { return n }
    return fibNaive(n - 1) + fibNaive(n - 2)
}

// Xxlang 尾递归
func fibTail(n, a, b) {
    if (n == 0) { return a }
    if (n == 1) { return b }
    return fibTail(n - 1, b, a + b)
}

// 对比测试
println("fib(35) 性能对比:")
`

	interp.Eval(perfTest)

	// 测试 Xxlang 朴素递归
	start := time.Now()
	interp.Eval("fibNaive(35)")
	naiveTime := time.Since(start)

	// 测试 Xxlang 尾递归
	start = time.Now()
	interp.Eval("fibTail(35, 0, 1)")
	tailTime := time.Since(start)

	// 测试插件 O(n) 算法
	start = time.Now()
	interp.Eval("fib.fast(35)")
	fastTime := time.Since(start)

	// 测试插件 O(log n) 算法
	start = time.Now()
	interp.Eval("fib.matrix(35)")
	matrixTime := time.Since(start)

	fmt.Printf("  Xxlang 朴素递归:    %v\n", naiveTime)
	fmt.Printf("  Xxlang 尾递归:      %v\n", tailTime)
	fmt.Printf("  插件 fib.fast:      %v (O(n))\n", fastTime)
	fmt.Printf("  插件 fib.matrix:    %v (O(log n))\n", matrixTime)

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("3. 边界值测试 (int64 最大斐波那契数)")
	fmt.Println("==============================================")
	fmt.Println()

	// int64 最大斐波那契数是 fib(92) = 7540113804746346429
	// fib(93) = 12200160415121876738 超出 int64 正数范围
	fmt.Println("int64 最大值: 9223372036854775807")
	fmt.Println("fib(92) = 7540113804746346429 (int64 范围内最大)")
	fmt.Println("fib(93) = 12200160415121876738 (超出 int64)")
	fmt.Println()

	start = time.Now()
	result, _ = interp.Eval("fib.matrix(92)")
	fmt.Printf("fib.matrix(92) 计算完成，耗时: %v\n", time.Since(start))
	fmt.Printf("结果: %s\n", result.Inspect())

	// 测试更大的数来展示性能（注意会溢出）
	fmt.Println()
	fmt.Println("注意: fib(93) 会溢出 int64，结果不正确")
	start = time.Now()
	result, _ = interp.Eval("fib.matrix(93)")
	fmt.Printf("fib.matrix(93) 计算完成，耗时: %v\n", time.Since(start))
	fmt.Printf("结果: %s (溢出后不正确)\n", result.Inspect())

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("总结")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("通过编写 Go 插件，Xxlang 获得：")
	fmt.Println("1. 原生 Go 性能")
	fmt.Println("2. 复杂算法支持 (如矩阵快速幂)")
	fmt.Println("3. 扩展功能 (如斐波那契数检测)")
	fmt.Println("4. 批量处理能力")
	fmt.Println()
	fmt.Println("插件开发步骤：")
	fmt.Println("1. 实现 plugin.Plugin 接口")
	fmt.Println("2. 在 init() 中调用 plugin.Register()")
	fmt.Println("3. 在主程序中 import 插件包")
	fmt.Println("4. 在 Xxlang 中 import \"plugin/插件名\"")
}
