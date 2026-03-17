// examples/fib_plugin/fibplugin.go
// 这是一个示例插件，展示如何将 Xxlang 作为第三方库来编写插件
//
// 插件提供高性能的斐波那契计算功能，供 Xxlang 代码调用
//
// 使用方式：
//
//	作为静态插件：在 Go 程序中 import 此包，插件自动注册
package fibplugin

import (
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/plugin"
)

// FibPlugin 实现 plugin.Plugin 接口
type FibPlugin struct{}

// Name 返回插件名称，用于 import "plugin/fib"
func (p *FibPlugin) Name() string {
	return "fib"
}

// Exports 返回插件导出的函数和变量
func (p *FibPlugin) Exports() map[string]objects.Object {
	return map[string]objects.Object{
		// 高性能斐波那契计算 (O(n) 时间复杂度)
		"fast": &objects.Builtin{
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return &objects.Error{Message: "fib.fast requires 1 argument"}
				}

				n, ok := args[0].(*objects.Int)
				if !ok {
					return &objects.Error{Message: "argument must be integer"}
				}

				result := fibFast(n.Value)
				return &objects.Int{Value: result}
			},
		},

		// 矩阵快速幂斐波那契 (O(log n) 时间复杂度)
		"matrix": &objects.Builtin{
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return &objects.Error{Message: "fib.matrix requires 1 argument"}
				}

				n, ok := args[0].(*objects.Int)
				if !ok {
					return &objects.Error{Message: "argument must be integer"}
				}

				result := fibMatrix(n.Value)
				return &objects.Int{Value: result}
			},
		},

		// 批量计算斐波那契数列
		"range_": &objects.Builtin{
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return &objects.Error{Message: "fib.range_ requires 1 argument"}
				}

				n, ok := args[0].(*objects.Int)
				if !ok {
					return &objects.Error{Message: "argument must be integer"}
				}

				results := make([]objects.Object, n.Value+1)
				a, b := int64(0), int64(1)
				for i := int64(0); i <= n.Value; i++ {
					results[i] = &objects.Int{Value: a}
					a, b = b, a+b
				}
				return &objects.Array{Elements: results}
			},
		},

		// 检查一个数是否为斐波那契数
		"isFib": &objects.Builtin{
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return &objects.Error{Message: "fib.isFib requires 1 argument"}
				}

				n, ok := args[0].(*objects.Int)
				if !ok {
					return &objects.Error{Message: "argument must be integer"}
				}

				return boolToObject(isFibonacci(n.Value))
			},
		},

		// 插件版本
		"version": &objects.String{Value: "1.0.0"},
	}
}

// Go 原生实现 - O(n) 时间复杂度
func fibFast(n int64) int64 {
	if n <= 1 {
		return n
	}
	a, b := int64(0), int64(1)
	for i := int64(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// 矩阵快速幂 - O(log n) 时间复杂度
func fibMatrix(n int64) int64 {
	if n <= 1 {
		return n
	}

	// 矩阵乘法
	mul := func(a, b [2][2]int64) [2][2]int64 {
		return [2][2]int64{
			{a[0][0]*b[0][0] + a[0][1]*b[1][0], a[0][0]*b[0][1] + a[0][1]*b[1][1]},
			{a[1][0]*b[0][0] + a[1][1]*b[1][0], a[1][0]*b[0][1] + a[1][1]*b[1][1]},
		}
	}

	// 矩阵快速幂
	result := [2][2]int64{{1, 0}, {0, 1}}
	base := [2][2]int64{{1, 1}, {1, 0}}

	for n > 0 {
		if n&1 == 1 {
			result = mul(result, base)
		}
		base = mul(base, base)
		n >>= 1
	}

	return result[0][1]
}

// 检查是否为斐波那契数
// 一个数是斐波那契数当且仅当 5*n^2+4 或 5*n^2-4 是完全平方数
func isFibonacci(n int64) bool {
	if n < 0 {
		return false
	}

	// 检查 5*n^2+4 或 5*n^2-4 是否为完全平方数
	check := func(x int64) bool {
		sqrt := int64(0)
		for sqrt*sqrt < x {
			sqrt++
		}
		return sqrt*sqrt == x
	}

	n2 := n * n
	return check(5*n2+4) || check(5*n2-4)
}

func boolToObject(b bool) *objects.Bool {
	if b {
		return objects.TRUE
	}
	return objects.FALSE
}

// 在 init 中自动注册插件
func init() {
	plugin.Register(&FibPlugin{})
}
