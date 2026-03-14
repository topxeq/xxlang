# Xxlang 示例脚本

本目录包含循序渐进的 Xxlang 语言示例脚本，帮助用户学习语言的各个方面。

## 目录结构

```
scripts/
├── basics/          # 基础语法 (01-05)
├── intermediate/    # 中级特性 (06-10)
├── advanced/        # 高级特性 (11-15)
└── stdlib/          # 标准库使用 (16-20)
```

## 运行脚本

```bash
xx run scripts/basics/01_variables.xxl
```

## 脚本列表

### 基础 (basics/)

1. **01_variables.xxl** - 变量声明和类型
2. **02_operators.xxl** - 算术、比较、逻辑运算符
3. **03_control_flow.xxl** - if-else 和 switch 语句
4. **04_loops.xxl** - while、for、for-in 循环
5. **05_functions.xxl** - 函数定义、参数、返回值、递归

### 中级 (intermediate/)

6. **06_arrays.xxl** - 数组创建、访问、方法
7. **07_maps.xxl** - Map 创建、访问、方法
8. **08_closures.xxl** - 闭包和捕获变量
9. **09_recursion.xxl** - 递归算法：阶乘、斐波那契、二分查找
10. **10_strings.xxl** - 字符串操作和方法

### 高级 (advanced/)

11. **11_classes.xxl** - 类定义、实例、方法
12. **12_inheritance.xxl** - 继承、super 调用、多态
13. **13_modules.xxl** - 模块模式、命名空间
14. **14_error_handling.xxl** - 错误处理模式
15. **15_functional.xxl** - 函数式编程：map、filter、reduce

### 标准库 (stdlib/)

16. **16_math.xxl** - 数学函数
17. **17_string_funcs.xxl** - 字符串函数
18. **18_io.xxl** - 输入输出操作
19. **19_encoding.xxl** - 编码模式和简单密码
20. **20_utilities.xxl** - 工具函数

## 内置函数

Xxlang 提供以下内置函数：

### 类型操作
- `typeOf(value)` - 返回类型名称
- `int(value)` - 转换为整数
- `float(value)` - 转换为浮点数
- `string(value)` - 转换为字符串

### 数学函数
- `abs(n)` - 绝对值
- `min(a, b)` - 最小值
- `max(a, b)` - 最大值
- `floor(n)` - 向下取整
- `ceil(n)` - 向上取整
- `sqrt(n)` - 平方根
- `pow(base, exp)` - 幂运算

### 数组函数
- `len(arr)` - 长度
- `push(arr, item)` - 添加元素
- `pop(arr)` - 移除末尾元素
- `first(arr)` - 第一个元素
- `last(arr)` - 最后一个元素
- `sort(arr)` - 排序
- `reverse(arr)` - 反转
- `sum(arr)` - 求和
- `avg(arr)` - 平均值
- `concat(arr1, arr2)` - 连接数组
- `indexOf(arr, item)` - 查找索引
- `containsArr(arr, item)` - 是否包含

### 字符串函数
- `upper(s)` - 大写
- `lower(s)` - 小写
- `trim(s)` - 去除空白
- `substr(s, start, end)` - 子字符串
- `split(s, delimiter)` - 分割
- `join(arr, delimiter)` - 连接
- `replace(s, old, new)` - 替换
- `indexOf(s, sub)` - 查找位置
- `containsStr(s, sub)` - 是否包含
- `startsWith(s, prefix)` - 是否以前缀开始
- `endsWith(s, suffix)` - 是否以后缀结束

### Map 函数
- `keys(map)` - 所有键
- `values(map)` - 所有值
- `hasKey(map, key)` - 是否包含键
- `delete(map, key)` - 删除键

### 其他
- `range(start, end, step)` - 生成范围

## 注意事项

### 已知限制

当前版本存在以下限制：

1. **字符串拼接**：不支持自动类型转换，拼接字符串和数字时需使用 `string()` 函数
2. **else if**：不支持 `else if` 语法，请使用嵌套的 `if-else`
3. **switch 语句**：当前未实现，请使用 `if-else` 链替代
4. **for-in 循环**：数组遍历可能有问题，建议使用索引遍历
5. **break/continue**：在某些情况下可能不正常工作
6. **尾递归优化**：可能导致运行时错误，建议使用迭代方式

### 字符串拼接示例

```xxl
// 错误
var msg = "Count: " + 5

// 正确
var msg = "Count: " + string(5)

// 或者使用 println 的多参数
println("Count: ", 5)
```

### if-else 链示例

```xxl
// 不支持 else if
if (grade >= 90) {
    println("A")
} else if (grade >= 80) {  // 错误
    println("B")
}

// 使用嵌套 if-else
if (grade >= 90) {
    println("A")
} else {
    if (grade >= 80) {
        println("B")
    } else {
        println("C")
    }
}
```

### 自定义脚本

用户添加的以 `hm_` 开头的脚本不应被删除。

## 学习路径

建议按以下顺序学习：

1. 先完成 basics/ 下的所有脚本
2. 然后学习 intermediate/ 的数据结构
3. 接着理解 advanced/ 的面向对象
4. 最后掌握 stdlib/ 的标准库使用
