# 内置函数参考手册

本文档提供 Xxlang 所有内置函数的完整参考。

## 目录

- [预置全局变量](#预置全局变量)
- [基础函数](#基础函数)
- [字符串函数](#字符串函数)
- [数学函数](#数学函数)
- [类型转换函数](#类型转换函数)
- [类型检查函数](#类型检查函数)
- [数组函数](#数组函数)
- [映射函数](#映射函数)
- [命令行参数函数](#命令行参数函数)
- [工具函数](#工具函数)
- [加密函数](#加密函数)
- [动态代码执行](#动态代码执行)
- [类型方法](#类型方法)
- [标准库模块](#标准库模块)

---

## 预置全局变量

Xxlang 提供预置的全局变量，在所有脚本中自动可用：

### argsG

包含所有命令行参数的字符串数组（包括程序名和脚本路径）。

```xxl
// 示例：运行 `xxl script.xxl -- -port=8080 -verbose`
// argsG 为：["xxl", "script.xxl", "--", "-port=8080", "-verbose"]

pln("参数：", argsG)
pln("第一个参数：", argsG[0])
```

### scriptPathG

当前执行脚本的路径。可能是：
- 文件路径（执行本地脚本时）
- URL（执行远程脚本时）
- 空字符串（REPL 模式或嵌入式运行时）

```xxl
pln("脚本路径：", scriptPathG)

if (scriptPathG == "") {
    pln("运行在 REPL 或嵌入式模式")
}
```

### 示例：命令行参数解析

```xxl
// 解析命令行参数
var port = getSwitch(argsG, "-port=", "8080")
var host = getSwitch(argsG, "-host=", "localhost")
var verbose = includes(argsG, "-verbose")

pln("端口：", port)
pln("主机：", host)
pln("详细模式：", verbose)
```

---

## 基础函数

### len(obj)

返回字符串、数组或映射的长度。

```xxl
len("hello")      // 5
len([1, 2, 3])    // 3
len({"a": 1})     // 1
```

### pr(args...)

打印参数到标准输出，不换行。

```xxl
pr("Hello", " ", "World")  // Hello World
```

### pln(args...)

打印参数到标准输出，末尾换行。

```xxl
pln("Hello", "World")  // Hello World
```

### pl(format, args...)

格式化打印，末尾自动换行。类似 Go 语言的 `fmt.Printf`，但自动在末尾添加 `\n`。

```xxl
pl("Hello, World!")                    // Hello, World!
pl("姓名: %s, 年龄: %d", "张三", 30)      // 姓名: 张三, 年龄: 30
pl("数值: %.2f", 3.14159)               // 数值: 3.14
pl("布尔: %v, 整数: %d", true, 42)       // 布尔: true, 整数: 42
```

**格式化动词：**
- `%s` - 字符串
- `%d` - 整数（十进制）
- `%f` - 浮点数
- `%.Nf` - 浮点数保留 N 位小数
- `%v` - 任意值的默认格式
- `%t` - 布尔值 (true/false)
- `%x` - 整数（十六进制）
- `%o` - 整数（八进制）
- `%b` - 整数（二进制）
- `%%` - 百分号字面量

### prf(format, args...)

格式化打印，不自动换行。等效于 Go 语言的 `fmt.Printf`。

```xxl
prf("姓名: %s, 年龄: %d", "张三", 30)  // 姓名: 张三, 年龄: 30（无换行）
prf("数值: %.2f", 3.14159)             // 数值: 3.14（无换行）
pln()                                  // 单独添加换行
```

**格式化动词：** 与 `pl` 相同。

### checkErr(obj, message?)

检查参数是否为错误对象。如果是，则将错误信息打印到标准错误输出并以代码 1 退出。否则不做任何操作。可选地接受自定义消息参数。如果消息中包含格式化占位符（如 `%v`），则将错误信息作为参数进行格式化输出。

```xxl
import "io"

var content = io.readFile("config.json")
checkErr(content)                        // 使用默认错误消息退出
checkErr(content, "读取文件失败")          // 使用自定义消息退出
checkErr(content, "错误: %v")             // 使用格式化消息退出
// 没有错误则继续执行
pln("文件加载成功")
```

### checkEmpty(str, message?)

检查字符串是否为空。如果为空，则以代码 1 退出。可选地接受第二个参数作为错误信息，在退出前打印到标准错误输出。

```xxl
var name = ""
checkEmpty(name)                      // 如果为空则静默退出
checkEmpty(name, "name 不能为空")       // 如果为空则打印消息并退出

var value = "hello"
checkEmpty(value)  // 不为空，不做任何操作，继续执行
pln("值: ", value)
```

### genOtpCode(secret)

从 base32 编码的密钥生成 TOTP（基于时间的一次性密码）代码。返回 6 位数字的 OTP 代码字符串，如果密钥无效则返回错误对象。

```xxl
// 从 TOTP 密钥生成 OTP 代码
var code = genOtpCode("JBSWY3DPEHPK3PXP")
pln("OTP 代码: ", code)  // 例如："398139"

// 检查错误
checkErr(code, "生成 OTP 失败: %v")

// 配合验证器应用使用（Google Authenticator、Authy 等）
var secret = "HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ"
var otp = genOtpCode(secret)
pln("您的 OTP: ", otp)
```

**注意：** 密钥必须是 TOTP 验证器应用使用的有效 base32 编码字符串。

### typeOf(obj, detailed?)

返回对象的类型字符串。当 `detailed=true` 时，对实例对象返回类名而非 "INSTANCE"。

```xxl
typeOf(42)        // "INT"
typeOf("hello")   // "STRING"
typeOf([1, 2])    // "ARRAY"
typeOf({"a": 1})  // "MAP"
typeOf(true)      // "BOOL"
typeOf(null)      // "NULL"

// 对于类实例
class Person { var name = "" }
var p = new Person()
typeOf(p)           // "INSTANCE"
typeOf(p, true)     // "Person"（返回类名）
```

---

## 字符串函数

### substr(str, start, end?)

从 `start` 到 `end`（不包含）提取子字符串。如果省略 `end`，则提取到字符串末尾。

```xxl
substr("hello", 1, 4)   // "ell"
substr("hello", 2)      // "llo"
```

### split(str, separator)

用分隔符分割字符串，返回数组。

```xxl
split("a,b,c", ",")     // ["a", "b", "c"]
split("hello world", " ")  // ["hello", "world"]
```

### join(array, separator)

用分隔符连接数组元素为字符串。

```xxl
join(["a", "b", "c"], "-")  // "a-b-c"
```

### trim(str)

移除首尾空白字符。

```xxl
trim("  hello  ")  // "hello"
```

### upper(str)

转换为大写。

```xxl
upper("hello")  // "HELLO"
```

### lower(str)

转换为小写。

```xxl
lower("HELLO")  // "hello"
```

### containsStr(str, substr)

检查字符串是否包含子串。

```xxl
containsStr("hello world", "world")  // true
containsStr("hello", "xyz")          // false
```

### replace(str, old, new)

替换所有匹配项。

```xxl
replace("hello world", "world", "Xxlang")  // "hello Xxlang"
```

### startsWith(str, prefix)

检查是否以指定前缀开头。

```xxl
startsWith("hello", "he")  // true
```

### endsWith(str, suffix)

检查是否以指定后缀结尾。

```xxl
endsWith("hello", "lo")  // true
```

### padLeft(str, width, padChar?)

在字符串左侧填充到指定宽度。

```xxl
padLeft("5", 4)         // "   5"（用空格填充）
padLeft("5", 4, "0")    // "0005"
padLeft("hello", 3)     // "hello"（已超过宽度）
```

### padRight(str, width, padChar?)

在字符串右侧填充到指定宽度。

```xxl
padRight("5", 4)        // "5   "（用空格填充）
padRight("5", 4, "0")   // "5000"
padRight("hello", 3)    // "hello"（已超过宽度）
```

### toChars(s)

将字符串转换为 chars 数组，用于基于字符的操作。对于正确处理 Unicode 字符至关重要，操作基于字符（码点）而非字节。

```xxl
// 字节与字符计数
var s = "Hello世界🎉"
pln(len(s))          // 15（字节）
pln(len(toChars(s))) // 8（字符）

// 字符索引
var c = toChars("中文测试")
pln(c[0])            // "中"
pln(c[1])            // "文"
pln(c[-1])           // "试"（负索引）

// 字符切片
var c2 = toChars("Hello World 你好")
pln(c2.subStr(0, 5).toStr())   // "Hello"
pln(c2.subStr(6, 11).toStr())  // "World"
```

**chars 方法：**
- `toStr()` - 转换回字符串
- `upper()` - 转大写（字符感知）
- `lower()` - 转小写（字符感知）
- `contains(sub)` - 检查是否包含子字符串
- `indexOf(sub)` - 查找子字符串的字符索引
- `startsWith(prefix)` - 检查前缀
- `endsWith(suffix)` - 检查后缀
- `reverse()` - 反转字符
- `repeat(n)` - 重复 n 次
- `subStr(start, end)` - 按字符索引切片

### charLen(s)

返回字符串中 Unicode 字符的数量，无需创建 chars 对象。

```xxl
charLen("Hello世界🎉")   // 8
charLen("中文测试")      // 4
charLen("hello")         // 5

// 对比 len() 返回字节数
pln(len("中文"))    // 6（字节）
pln(charLen("中文")) // 2（字符）
```

---

## 数学函数

### abs(num)

返回绝对值。

```xxl
abs(-42)   // 42
abs(-3.14) // 3.14
```

### floor(num)

返回小于或等于数值的最大整数。

```xxl
floor(3.7)  // 3
floor(-3.7) // -4
```

### ceil(num)

返回大于或等于数值的最小整数。

```xxl
ceil(3.2)   // 4
ceil(-3.2)  // -3
```

### sqrt(num)

返回平方根。

```xxl
sqrt(16)   // 4
sqrt(2)    // 1.4142135623730951
```

### pow(base, exp)

返回 `base` 的 `exp` 次方。

```xxl
pow(2, 10)  // 1024
pow(3, 2)   // 9
```

### min(a, b)

返回较小值。

```xxl
min(5, 3)    // 3
min(1.5, 2)  // 1.5
```

### max(a, b)

返回较大值。

```xxl
max(5, 3)    // 5
max(1.5, 2)  // 2
```

---

## 类型转换函数

### int(value)

转换为整数。

```xxl
int(3.7)       // 3
int("42")      // 42
int(true)      // 1
int(false)     // 0
```

### float(value)

转换为浮点数。

```xxl
float(42)      // 42.0
float("3.14")  // 3.14
float(true)    // 1.0
```

### string(value)

转换为字符串表示。

```xxl
string(42)     // "42"
string(3.14)   // "3.14"
string(true)   // "true"
string([1, 2]) // "[1, 2]"
```

---

## 类型检查函数

### isString(value)

判断值是否为字符串。

```xxl
isString("hello")    // true
isString(42)         // false
```

### isNumber(value)

判断值是否为整数或浮点数。

```xxl
isNumber(42)         // true
isNumber(3.14)       // true
isNumber("42")       // false
```

### isInt(value)

判断值是否为整数。

```xxl
isInt(42)            // true
isInt(3.14)          // false
```

### isFloat(value)

判断值是否为浮点数。

```xxl
isFloat(3.14)        // true
isFloat(42)          // false
```

### isBigInt(value)

判断值是否为大整数（BigInt）。

```xxl
isBigInt(12345678901234567890n)  // true
isBigInt(42)                      // false
isBigInt(3.14)                    // false
```

### isBigFloat(value)

判断值是否为大浮点数（BigFloat）。

```xxl
isBigFloat(3.14159265358979323846m)  // true
isBigFloat(3.14)                      // false
isBigFloat(42)                        // false
```

### isArray(value)

判断值是否为数组。

```xxl
isArray([1, 2, 3])   // true
isArray("hello")     // false
```

### isMap(value)

判断值是否为映射（Map）。

```xxl
isMap({"a": 1})      // true
isMap([1, 2])        // false
```

### isBool(value)

判断值是否为布尔值。

```xxl
isBool(true)         // true
isBool(1)            // false
```

### isFunction(value)

判断值是否为函数。

```xxl
isFunction(len)      // true
isFunction(42)       // false
```

### isNull(value)

判断值是否为 null。

```xxl
isNull(null)         // true
isNull(0)            // false
```

---

## 数组函数

### push(array, element)

返回追加元素后的新数组。

```xxl
push([1, 2, 3], 4)  // [1, 2, 3, 4]
```

### pop(array)

返回移除最后一个元素后的新数组。

```xxl
pop([1, 2, 3])  // [1, 2]
```

### first(array)

返回数组第一个元素，空数组返回 null。

```xxl
first([1, 2, 3])  // 1
first([])         // null
```

### last(array)

返回数组最后一个元素，空数组返回 null。

```xxl
last([1, 2, 3])  // 3
last([])         // null
```

### rest(array, start, end?)

返回从 `start` 到 `end`（不包含）的切片。

```xxl
rest([1, 2, 3, 4], 1, 3)  // [2, 3]
rest([1, 2, 3, 4], 2)     // [3, 4]
```

### concat(array1, array2)

连接两个数组。

```xxl
concat([1, 2], [3, 4])  // [1, 2, 3, 4]
```

### indexOf(array, element)

返回元素索引，未找到返回 -1。

```xxl
indexOf([1, 2, 3], 2)   // 1
indexOf([1, 2, 3], 5)   // -1
```

### containsArr(array, element)

检查数组是否包含元素。

```xxl
containsArr([1, 2, 3], 2)  // true
containsArr([1, 2, 3], 5)  // false
```

### sort(array)

返回排序后的数组副本。

```xxl
sort([3, 1, 2])  // [1, 2, 3]
```

### sum(array)

返回数组元素的和。

```xxl
sum([1, 2, 3, 4, 5])  // 15
```

### avg(array)

返回数组元素的平均值。

```xxl
avg([1, 2, 3, 4, 5])  // 3.0
```

### reverse(array)

返回反转后的数组副本。

```xxl
reverse([1, 2, 3])  // [3, 2, 1]
```

---

## 映射函数

### keys(map)

返回所有键组成的数组。

```xxl
keys({"a": 1, "b": 2})  // ["a", "b"]
```

### values(map)

返回所有值组成的数组。

```xxl
values({"a": 1, "b": 2})  // [1, 2]
```

### hasKey(map, key)

检查是否包含指定键。

```xxl
hasKey({"a": 1}, "a")  // true
hasKey({"a": 1}, "b")  // false
```

### delete(map, key)

返回移除指定键后的新映射。

```xxl
delete({"a": 1, "b": 2}, "a")  // {"b": 2}
```

---

## 命令行参数函数

### getSwitch(array, prefix, default)

在数组中搜索以指定前缀开头的元素，返回前缀后的值。如果未找到，返回默认值。

这对于解析命令行参数特别有用。

```xxl
// 假设 argsG = ["script.xxl", "-port=8080", "-host=localhost", "-verbose"]

var port = getSwitch(argsG, "-port=", "3000")       // "8080"
var host = getSwitch(argsG, "-host=", "127.0.0.1")  // "localhost"
var debug = getSwitch(argsG, "-debug=", "false")    // "false"（未找到）

// 检查标志型参数（前缀后无值）
var hasVerbose = getSwitch(argsG, "-verbose", "")   // ""（找到了，但没有值）
if (hasVerbose == "" && includes(argsG, "-verbose")) {
    pln("详细模式已启用")
}
```

**参数：**
- `array` - 要搜索的字符串数组（通常是 `argsG`）
- `prefix` - 要搜索的前缀（如 "-port="）
- `default` - 未找到时返回的默认值

**返回值：**
- 找到时返回前缀后的值
- 未找到时返回默认值

### switchExists(array, switchName)

检查开关参数是否存在于数组中，要求精确匹配。只有当元素完全匹配开关名称时才返回 `true`。

```xxl
// 假设 argsG = ["script.xxl", "-port=8080", "-verbose", "-debug=true"]

var hasPort = switchExists(argsG, "-port")         // false（无精确匹配）
var hasPortEq = switchExists(argsG, "-port=8080")  // true（精确匹配）
var hasVerbose = switchExists(argsG, "-verbose")   // true（精确匹配）
var hasDebug = switchExists(argsG, "-debug")       // false（无精确匹配）
var hasDebugEq = switchExists(argsG, "-debug=true") // true（精确匹配）
```

**参数：**
- `array` - 要搜索的字符串数组（通常是 `argsG`）
- `switchName` - 要查找的开关名称（如 "-verbose"）

**返回值：**
- `true` 如果开关精确匹配存在
- `false` 如果未找到

---

## 工具函数

### range(end) 或 range(start, end) 或 range(start, end, step)

生成从 start 到 end（包含）的整数数组。

```xxl
range(5)           // [0, 1, 2, 3, 4, 5]
range(2, 5)        // [2, 3, 4, 5]
range(5, 2)        // [5, 4, 3, 2]
range(0, 10, 2)    // [0, 2, 4, 6, 8]（指定步长）
range(10, 0, -2)   // [10, 8, 6, 4, 2]（负步长）
```

**参数：**
- `end` - 结束值（包含），起始值默认为 0
- `start` - 起始值（可选）
- `step` - 步长值（可选，默认为 1）。不能为零。

**返回：** 整数数组

### runCode(code, args?)

动态执行 Xxlang 代码。可选的 `args` 映射提供变量。

```xxl
runCode("1 + 2")                    // 3
runCode("a + b", {"a": 10, "b": 20}) // 30
```

### loadPlugin(path)

从指定路径加载原生 Go 插件。

```xxl
loadPlugin("./myplugin.so")
```

### copy(obj)

创建数组或映射的浅拷贝。

```xxl
var arr = [1, 2, 3]
var arrCopy = copy(arr)
arrCopy[0] = 99
pln(arr[0])     // 1 (原数组不变)

var map = {"a": 1}
var mapCopy = copy(map)
```

### clone(obj)

创建数组或映射的深拷贝。

```xxl
var nested = {"a": [1, 2, 3]}
var cloned = clone(nested)
cloned["a"][0] = 99
pln(nested["a"][0])  // 1 (原对象不变)
```

### equals(a, b)

执行深度相等比较。

```xxl
equals([1, 2], [1, 2])           // true
equals({"a": 1}, {"a": 1})       // true
equals([1, [2, 3]], [1, [2, 3]]) // true
```

### defaults(obj, defaultObj)

用默认对象填充缺失的键。

```xxl
var config = {"host": "localhost"}
var result = defaults(config, {"host": "127.0.0.1", "port": 8080})
// {"host": "localhost", "port": 8080}
```

### base64Encode(s)

将字符串编码为 base64。

```xxl
base64Encode("hello")  // "aGVsbG8="
```

### base64Decode(s)

解码 base64 字符串。

```xxl
base64Decode("aGVsbG8=")  // "hello"
```

### hexEncode(s)

将字符串编码为十六进制。

```xxl
hexEncode("hello")  // "68656c6c6f"
```

### hexDecode(s)

解码十六进制字符串。

```xxl
hexDecode("68656c6c6f")  // "hello"
```

### md5(s)

返回 MD5 哈希十六进制字符串。

```xxl
md5("hello")  // "5d41402abc4b2a76b9719d911017c592"
```

### sha256(s)

返回 SHA256 哈希十六进制字符串。

```xxl
sha256("hello")  // "2cf24dba5fb0a30e26e83b2ac5b9e29e..."
```

### sleep(ms)

暂停执行指定的毫秒数。

```xxl
pln("开始...")
sleep(1000)  // 休眠1秒
pln("完成！")
```

### now()

返回当前 Unix 时间戳（秒）。

```xxl
var ts = now()  // 例如: 1710422400
```

### nowMs()

返回当前 Unix 时间戳（毫秒）。

```xxl
var ms = nowMs()  // 例如: 1710422400000
```

### uuid()

生成随机 UUID 字符串。

```xxl
var id = uuid()  // "550e8400-e29b-41d4-a716-446655440000"
```

### trimPrefix(s, prefix)

移除字符串的前缀（如果存在）。

```xxl
trimPrefix("hello_world", "hello_")  // "world"
trimPrefix("hello", "x")             // "hello"
```

### trimSuffix(s, suffix)

移除字符串的后缀（如果存在）。

```xxl
trimSuffix("hello.txt", ".txt")  // "hello"
trimSuffix("hello", ".txt")      // "hello"
```

### count(arr)

返回数组的长度。

```xxl
count([1, 2, 3, 4, 5])  // 5
```

### isDigit(s)

如果字符串只包含数字则返回 true。

```xxl
isDigit("12345")   // true
isDigit("12a45")   // false
```

### isAlpha(s)

如果字符串只包含字母则返回 true。

```xxl
isAlpha("hello")   // true
isAlpha("hello1")  // false
```

### isAlphaNum(s)

如果字符串只包含字母和数字则返回 true。

```xxl
isAlphaNum("hello123")  // true
isAlphaNum("hello!")    // false
```

### find(arr, predicate)

返回第一个匹配谓词的元素，或 null。

```xxl
find([1, 2, 3, 4, 5], func(x) { return x > 3 })  // 4
```

### findIndex(arr, predicate)

返回第一个匹配谓词的元素的索引，或 -1。

```xxl
findIndex([1, 2, 3, 4, 5], func(x) { return x > 3 })  // 3
```

### includes(arr, value)

如果数组包含值则返回 true。

```xxl
includes([1, 2, 3], 2)   // true
includes([1, 2, 3], 5)   // false
```

### shuffle(arr)

返回随机打乱顺序的数组副本。

```xxl
shuffle([1, 2, 3, 4, 5])  // [3, 1, 5, 2, 4] (随机顺序)
```

### sample(arr, n)

从数组中随机选取 n 个元素。

```xxl
sample([1, 2, 3, 4, 5], 2)  // [3, 5] (随机选择)
```

### chunk(arr, size)

将数组分割为指定大小的块。

```xxl
chunk([1, 2, 3, 4, 5, 6], 2)  // [[1, 2], [3, 4], [5, 6]]
chunk([1, 2, 3, 4, 5], 2)     // [[1, 2], [3, 4], [5]]
```

---

## 加密函数

Xxlang 提供与 Charlang 兼容的加密/解密函数。这些函数不依赖第三方库实现，并保持与 Charlang 的完全交叉兼容。

### encryptTextByTXTE(text, code)

使用 TXTE（简单文本加密）算法加密文本。返回十六进制字符串。这是确定性加密 - 相同的输入总是产生相同的输出。

```xxl
var encrypted = encryptTextByTXTE("Hello", "mykey")
// "8F9FA29CA39E"（确定性输出）
```

### decryptTextByTXTE(hexStr, code)

解密由 TXTE 加密的十六进制字符串。

```xxl
var decrypted = decryptTextByTXTE("8F9FA29CA39E", "mykey")
// "Hello"
```

### encryptDataByTXDEE(data, code)

使用 TXDEE（增强数据加密）算法加密字节数据，带有随机前缀/后缀字节。返回字节数组。

```xxl
var data = [72, 101, 108, 108, 111]  // "Hello" 字节
var encrypted = encryptDataByTXDEE(data, "mykey")
// 返回带随机填充的字节数组
```

### decryptDataByTXDEE(data, code)

解密由 TXDEE 加密的字节数据。

```xxl
var decrypted = decryptDataByTXDEE(encrypted, "mykey")
// 返回原始字节数组
```

### encryptTextByTXDEE(text, code)

使用 TXDEE 加密文本并返回十六进制字符串。

```xxl
var encrypted = encryptTextByTXDEE("Hello", "mykey")
// 由于随机字节，每次输出不同
```

### decryptTextByTXDEE(hexStr, code)

解密由 TXDEE 加密的十六进制字符串。

```xxl
var decrypted = decryptTextByTXDEE(encrypted, "mykey")
// "Hello"
```

### encryptDataByTXDEF(data, code)

使用 TXDEF（灵活数据加密）算法加密字节数据，根据密钥动态填充。返回字节数组。

```xxl
var data = [72, 101, 108, 108, 111]
var encrypted = encryptDataByTXDEF(data, "mykey")
```

### decryptDataByTXDEF(data, code)

解密由 TXDEF 加密的字节数据。

```xxl
var decrypted = decryptDataByTXDEF(encrypted, "mykey")
```

### encryptTextByTXDEF(text, code)

使用 TXDEF 加密文本并返回十六进制字符串。

```xxl
var encrypted = encryptTextByTXDEF("Hello", "mykey")
```

### decryptTextByTXDEF(hexStr, code)

解密由 TXDEF 加密的十六进制字符串。

```xxl
var decrypted = decryptTextByTXDEF(encrypted, "mykey")
// "Hello"
```

### encryptData(data, code) / decryptData(data, code)

使用 TXDEF 算法的默认数据加密。

```xxl
var encrypted = encryptData([1, 2, 3], "secret")
var decrypted = decryptData(encrypted, "secret")
```

### encryptBytes(data, code) / decryptBytes(data, code)

字节数组加密别名。

```xxl
var encrypted = encryptBytes([72, 101, 108, 108, 111], "key")
var decrypted = decryptBytes(encrypted, "key")
```

### encryptText(text, code) / decryptText(hexStr, code)

使用 TXDEF 的默认文本加密。

```xxl
var encrypted = encryptText("Hello World", "mykey")
var decrypted = decryptText(encrypted, "mykey")
// "Hello World"
```

### encryptStr(text, code) / decryptStr(hexStr, code)

字符串加密别名（与 encryptText/decryptText 相同）。

```xxl
var encrypted = encryptStr("secret message", "password")
var decrypted = decryptStr(encrypted, "password")
```

### encryptStream(reader, code, writer) / decryptStream(reader, code, writer)

用于大数据的流式加密/解密。

```xxl
import "io"

var reader = io.newReader("large content to encrypt")
var writer = io.newBytesWriter()
encryptStream(reader, "mykey", writer)
var encrypted = writer.bytes()

// 解密
var reader2 = io.newReader(encrypted)
var writer2 = io.newBytesWriter()
decryptStream(reader2, "mykey", writer2)
```

### aesEncrypt(data, key, mode?) / aesDecrypt(data, key, mode?)

AES 加密/解密。支持 ECB-like 模式（CBC 配合零 IV）和 CBC 模式。

```xxl
// ECB-like 模式（默认）
var encrypted = aesEncrypt([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], "16bytekey1234567")
var decrypted = aesDecrypt(encrypted, "16bytekey1234567")

// CBC 模式
var encryptedCBC = aesEncrypt(data, "16bytekey1234567", "cbc")
var decryptedCBC = aesDecrypt(encryptedCBC, "16bytekey1234567", "cbc")
```

**注意：** 密钥如果超过 16 字节会被截断。CBC 模式下，密钥前缀用作 IV。

---

## 类型方法

所有类型都有以下通用方法：

- `typeOf()` - 返回类型字符串
- `toStr()` - 返回字符串表示

### 整数方法

```xxl
(-5).abs()      // 5
42.toFloat()    // 42.0
42.typeOf()     // "INT"
```

### 浮点数方法

```xxl
3.7.floor()     // 3
3.2.ceil()      // 4
3.5.round()     // 4
(-3.14).abs()   // 3.14
3.14.toInt()    // 3
```

### 字符串方法

```xxl
"hello".len()           // 5
"hello".upper()         // "HELLO"
"HELLO".lower()         // "hello"
"  hello  ".trim()      // "hello"
"a,b,c".split(",")      // ["a", "b", "c"]
"hello".contains("ell") // true
"hello".indexOf("l")    // 2
"hello".startsWith("he") // true
"hello".endsWith("lo")  // true
"42".toInt()            // 42
"3.14".toFloat()        // 3.14
```

### chars 方法

```xxl
var c = toChars("Hello世界🎉")

c.len()                    // 8（字符）
c[0]                       // "H"
c[5]                       // "世"
c.subStr(0, 5).toStr()     // "Hello"
c.subStr(5, 7).toStr()     // "世界"
c.upper().toStr()          // "HELLO世界🎉"
c.lower().toStr()          // "hello世界🎉"
c.contains("世界")          // true
c.indexOf("世界")           // 5
c.startsWith("Hello")      // true
c.endsWith("🎉")            // true
c.reverse().toStr()        // "🎉界世olleH"
c.repeat(2).toStr()        // "Hello世界🎉Hello世界🎉"
```

### 数组方法

```xxl
[1, 2, 3].len()        // 3
[1, 2].push(3)         // [1, 2, 3]
[1, 2, 3].pop()        // [1, 2]
[1, 2, 3].first()      // 1
[1, 2, 3].last()       // 3
[1, 2, 3].indexOf(2)   // 1
[1, 2, 3].contains(2)  // true
[1, 2, 3].reverse()    // [3, 2, 1]
["a", "b"].join("-")   // "a-b"
```

### 映射方法

```xxl
{"a": 1, "b": 2}.len()     // 2
{"a": 1, "b": 2}.keys()    // ["a", "b"]
{"a": 1, "b": 2}.values()  // [1, 2]
{"a": 1}.hasKey("a")       // true
{"a": 1, "b": 2}.delete("a") // {"b": 2}
```

---

## 标准库模块

使用 `import` 语句导入标准库模块：

```xxl
import "math"
import "io" { readFile, writeFile }
```

### math

数学函数和常量。

| 函数 | 说明 | 示例 |
|------|------|------|
| `PI` | 圆周率常量 | `math.PI` |
| `E` | 自然常数 | `math.E` |
| `abs(x)` | 绝对值 | `math.abs(-5)` |
| `ceil(x)` | 向上取整 | `math.ceil(3.2)` |
| `floor(x)` | 向下取整 | `math.floor(3.7)` |
| `round(x)` | 四舍五入 | `math.round(3.5)` |
| `sqrt(x)` | 平方根 | `math.sqrt(16)` |
| `pow(x, y)` | 幂运算 | `math.pow(2, 8)` |
| `sin(x)` | 正弦（弧度） | `math.sin(1.57)` |
| `cos(x)` | 余弦（弧度） | `math.cos(0)` |
| `tan(x)` | 正切（弧度） | `math.tan(1)` |
| `asin(x)` | 反正弦 | `math.asin(1)` |
| `acos(x)` | 反余弦 | `math.acos(0)` |
| `atan(x)` | 反正切 | `math.atan(1)` |
| `log(x)` | 自然对数 | `math.log(2.7)` |
| `log10(x)` | 以10为底对数 | `math.log10(100)` |
| `exp(x)` | 指数函数 | `math.exp(1)` |
| `min(args...)` | 最小值 | `math.min(3, 1, 2)` |
| `max(args...)` | 最大值 | `math.max(3, 1, 2)` |
| `random()` | 随机数 [0, 1) | `math.random()` |

### io

输入输出操作。

| 函数 | 说明 | 示例 |
|------|------|------|
| `print(args...)` | 打印（不换行） | `io.print("Hello")` |
| `println(args...)` | 打印（换行） | `io.println("Hello")` |
| `printf(fmt, args...)` | 格式化打印 | `io.printf("值: %d", 42)` |
| `readLine()` | 读取一行输入 | `io.readLine()` |
| `readFile(path)` | 读取文件内容 | `io.readFile("data.txt")` |
| `readBytes(path)` | 读取文件为字节数组 | `io.readBytes("data.bin")` |
| `writeFile(path, content)` | 写入字符串到文件 | `io.writeFile("out.txt", "data")` |
| `writeBytes(path, bytes)` | 写入字节到文件 | `io.writeBytes("out.bin", bytes)` |
| `appendFile(path, content)` | 追加内容到文件 | `io.appendFile("log.txt", "msg")` |
| `exists(path)` | 检查文件是否存在 | `io.exists("data.txt")` |
| `remove(path)` | 删除文件 | `io.remove("temp.txt")` |
| `mkdir(path)` | 创建目录 | `io.mkdir("mydir")` |
| `cwd()` | 获取当前目录 | `io.cwd()` |
| `exit(code)` | 退出程序 | `io.exit(0)` |
| `env(key)` | 获取环境变量 | `io.env("HOME")` |
| `setEnv(key, value)` | 设置环境变量 | `io.setEnv("DEBUG", "1")` |
| `args()` | 获取命令行参数 | `io.args()` |
| `scan(prompt?)` | 读取一行输入 | `io.scan("请输入姓名: ")` |
| `scanInt(prompt?)` | 读取整数 | `io.scanInt("请输入年龄: ")` |
| `scanFloat(prompt?)` | 读取浮点数 | `io.scanFloat("请输入价格: ")` |
| `scanBool(prompt?)` | 读取布尔值 | `io.scanBool("继续吗? ")` |
| `scanN(n)` | 读取n个token | `io.scanN(3)` |
| `scanSplit(sep)` | 读取并分割 | `io.scanSplit(",")` |
| `scan2()` | 读取两个值 | `a, b = io.scan2()` |
| `scan3()` | 读取三个值 | `a, b, c = io.scan3()` |
| `scanf(format)` | 格式化读取 | `io.scanf("{} {}")` |
| `newScanner(reader?)` | 创建Scanner对象 | `io.newScanner()` |

#### 输入/扫描函数

Xxlang提供了便捷的函数用于从标准输入读取用户输入：

```xxl
import "io"

// 基本用法 - 读取一行
var name = io.scan("请输入姓名: ")
pln("你好, " + name + "!")

// 读取特定类型
var age = io.scanInt("请输入年龄: ")
var price = io.scanFloat("请输入价格: ")
var confirmed = io.scanBool("继续吗? (true/false): ")

// 读取多个值
var a, b = io.scan2()  // 读取两个空白分隔的token
var x, y, z = io.scan3()  // 读取三个token

// 读取n个值到数组
var tokens = io.scanN(3)  // 返回 ["token1", "token2", "token3"]

// 读取并按分隔符分割
var parts = io.scanSplit(",")  // 读取一行，按逗号分割

// 格式化读取
var values = io.scanf("{} {} {}")  // 读取三个空格分隔的值
```

#### Scanner对象

如果需要更多控制，可以使用Scanner对象：

```xxl
import "io"

// 从标准输入创建扫描器
var scanner = io.newScanner()

// 或从reader创建
var reader = io.newReader("hello world\n42\n3.14")
var scanner2 = io.newScanner(reader)

// 读取token
var token = scanner.next()       // 读取空白分隔的token
var line = scanner.nextLine()    // 读取整行
var num = scanner.nextInt()      // 读取整数
var f = scanner.nextFloat()      // 读取浮点数
var b = scanner.nextBool()       // 读取布尔值

// 检查是否还有输入
if (scanner.hasNext()) {
    pln("还有更多输入")
}

// 跳过当前行
scanner.skipLine()

// 完成后关闭
scanner.close()
```

**Scanner对象方法：**

| 方法 | 说明 | 返回类型 |
|------|------|----------|
| `next()` | 读取下一个空白分隔的token | STRING 或 NULL |
| `nextLine()` | 读取下一行 | STRING 或 NULL |
| `nextInt()` | 读取下一个整数 | INT 或 ERROR |
| `nextFloat()` | 读取下一个浮点数 | FLOAT 或 ERROR |
| `nextBool()` | 读取下一个布尔值 | BOOL 或 ERROR |
| `hasNext()` | 检查是否还有输入 | BOOL |
| `skipLine()` | 跳过当前行 | NULL |
| `close()` | 关闭扫描器 | NULL |

### os

操作系统工具。

| 函数 | 说明 | 示例 |
|------|------|------|
| `join(paths...)` | 连接路径 | `os.join("dir", "file.txt")` |
| `base(path)` | 获取文件名 | `os.base("/a/b/c.txt")` |
| `dir(path)` | 获取目录路径 | `os.dir("/a/b/c.txt")` |
| `ext(path)` | 获取扩展名 | `os.ext("file.txt")` |
| `abs(path)` | 获取绝对路径 | `os.abs("./file")` |
| `clean(path)` | 清理路径 | `os.clean("./a/../b")` |
| `isAbs(path)` | 检查是否绝对路径 | `os.isAbs("/home")` |
| `stat(path)` | 获取文件信息 | `os.stat("file.txt")` |
| `size(path)` | 获取文件大小 | `os.size("file.txt")` |
| `isDir(path)` | 检查是否目录 | `os.isDir("mydir")` |
| `isFile(path)` | 检查是否文件 | `os.isFile("file.txt")` |
| `listDir(path)` | 列出目录内容 | `os.listDir(".")` |
| `walk(path)` | 遍历目录树 | `os.walk("/home")` |
| `exec(cmd)` | 执行命令 | `os.exec("ls -la")` |
| `shell(cmd)` | 执行 shell 命令 | `os.shell("echo hello")` |
| `hostname()` | 获取主机名 | `os.hostname()` |
| `platform()` | 获取操作系统 | `os.platform()` |
| `arch()` | 获取 CPU 架构 | `os.arch()` |
| `home()` | 获取主目录 | `os.home()` |
| `temp()` | 获取临时目录 | `os.temp()` |
| `rename(old, new)` | 重命名文件 | `os.rename("a.txt", "b.txt")` |
| `copy(src, dst)` | 复制文件 | `os.copy("a.txt", "b.txt")` |
| `chmod(path, mode)` | 修改权限 | `os.chmod("file", 0755)` |
| `tempFile(pattern)` | 创建临时文件 | `os.tempFile("app-*")` |
| `tempDir(pattern)` | 创建临时目录 | `os.tempDir("app-*")` |

### json

JSON 编解码。

| 函数 | 说明 | 示例 |
|------|------|------|
| `parse(str)` | 解析 JSON 字符串 | `json.parse('{"a": 1}')` |
| `stringify(obj, indent?)` | 转换为 JSON 字符串 | `json.stringify(obj, 2)` |
| `encode(obj)` | 编码为 JSON | `json.encode(obj)` |
| `decode(str)` | 解码 JSON 字符串 | `json.decode('{"a": 1}')` |

### regex

正则表达式操作（兼容 PCRE）。

| 函数 | 说明 | 示例 |
|------|------|------|
| `compile(pattern)` | 编译正则表达式 | `regex.compile("\\d+")` |
| `match(pattern, str)` | 检查是否匹配 | `regex.match("\\d+", "abc123")` |
| `find(pattern, str)` | 查找第一个匹配 | `regex.find("\\d+", "abc123")` |
| `findAll(pattern, str, limit?)` | 查找所有匹配 | `regex.findAll("\\d+", "a1b2c3")` |
| `findGroups(pattern, str)` | 获取捕获组 | `regex.findGroups("(\\d+)-(\\d+)", "1-2")` |
| `replace(pattern, str, repl)` | 替换匹配 | `regex.replace("\\d+", "a1b2", "X")` |
| `split(pattern, str, limit?)` | 按正则分割 | `regex.split("\\s+", "a b c")` |
| `quote(str)` | 转义正则字符 | `regex.quote("a.b")` |
| `count(pattern, str)` | 计数匹配 | `regex.count("\\d+", "a1b2c3")` |
| `test(pattern)` | 验证模式 | `regex.test("\\d+")` |

### time

时间和日期操作。

| 函数 | 说明 | 示例 |
|------|------|------|
| `unix()` | 当前 Unix 时间戳（秒） | `time.unix()` |
| `unixMs()` | 当前 Unix 时间戳（毫秒） | `time.unixMs()` |
| `unixNano()` | 当前 Unix 时间戳（纳秒） | `time.unixNano()` |
| `now()` | 当前时间（映射） | `time.now()` |
| `year()` | 当前年份 | `time.year()` |
| `month()` | 当前月份（1-12） | `time.month()` |
| `day()` | 当前日期 | `time.day()` |
| `hour()` | 当前小时（0-23） | `time.hour()` |
| `minute()` | 当前分钟 | `time.minute()` |
| `second()` | 当前秒数 | `time.second()` |
| `weekday()` | 星期几（0=周日） | `time.weekday()` |
| `sleep(ms)` | 休眠毫秒 | `time.sleep(1000)` |
| `sleepSec(sec)` | 休眠秒 | `time.sleepSec(1)` |
| `format(layout)` | 格式化当前时间 | `time.format("2006-01-02")` |
| `formatUnix(ts, layout)` | 格式化时间戳 | `time.formatUnix(ts, "2006-01-02")` |
| `parse(layout, value)` | 解析时间字符串 | `time.parse("2006-01-02", "2024-01-15")` |
| `since(ms)` | 距时间戳的毫秒数 | `time.since(start)` |
| `addDays(days)` | 加天数 | `time.addDays(7)` |
| `addMonths(months)` | 加月数 | `time.addMonths(1)` |
| `addYears(years)` | 加年数 | `time.addYears(1)` |
| `isLeapYear(year)` | 是否闰年 | `time.isLeapYear(2024)` |
| `daysInMonth(year, month)` | 月天数 | `time.daysInMonth(2024, 2)` |

### string

字符串工具模块。

| 函数 | 说明 | 示例 |
|------|------|------|
| `len(s)` | 字符串长度 | `string.len("hello")` |
| `substr(s, start, end?)` | 子串 | `string.substr("hello", 1, 4)` |
| `indexOf(s, substr)` | 查找子串 | `string.indexOf("hello", "ll")` |
| `contains(s, substr)` | 包含检查 | `string.contains("hello", "ell")` |
| `hasPrefix(s, prefix)` | 前缀检查 | `string.hasPrefix("hello", "he")` |
| `hasSuffix(s, suffix)` | 后缀检查 | `string.hasSuffix("hello", "lo")` |
| `toUpper(s)` | 转大写 | `string.toUpper("hello")` |
| `toLower(s)` | 转小写 | `string.toLower("HELLO")` |
| `trim(s)` | 去空白 | `string.trim("  hello  ")` |
| `trimSpace(s)` | 去空白 | `string.trimSpace("  hello  ")` |
| `split(s, sep)` | 分割字符串 | `string.split("a,b,c", ",")` |
| `join(arr, sep)` | 连接字符串 | `string.join(["a", "b"], "-")` |
| `repeat(s, n)` | 重复字符串 | `string.repeat("ab", 3)` |
| `replace(s, old, new)` | 替换全部 | `string.replace("hello", "l", "L")` |
| `parseInt(s)` | 解析整数 | `string.parseInt("42")` |
| `parseFloat(s)` | 解析浮点数 | `string.parseFloat("3.14")` |
| `toString(x)` | 转字符串 | `string.toString(42)` |
| `reverse(s)` | 反转字符串 | `string.reverse("hello")` |

### crypto

加密函数，包括哈希、编码和加密。

**哈希函数：**

| 函数 | 说明 | 示例 |
|------|------|------|
| `md5(s)` | MD5 哈希 | `crypto.md5("hello")` |
| `sha1(s)` | SHA1 哈希 | `crypto.sha1("hello")` |
| `sha256(s)` | SHA256 哈希 | `crypto.sha256("hello")` |
| `sha512(s)` | SHA512 哈希 | `crypto.sha512("hello")` |

**编码函数：**

| 函数 | 说明 | 示例 |
|------|------|------|
| `base64Encode(s)` | Base64 编码 | `crypto.base64Encode("hello")` |
| `base64Decode(s)` | Base64 解码 | `crypto.base64Decode("aGVsbG8=")` |
| `hexEncode(s)` | 十六进制编码 | `crypto.hexEncode("hello")` |
| `hexDecode(s)` | 十六进制解码 | `crypto.hexDecode("68656c6c6f")` |

**加密函数（Charlang 兼容）：**

| 函数 | 说明 | 示例 |
|------|------|------|
| `encryptTextByTXTE(text, code)` | TXTE 文本加密 | `crypto.encryptTextByTXTE("hello", "key")` |
| `decryptTextByTXTE(hexStr, code)` | TXTE 文本解密 | `crypto.decryptTextByTXTE("...", "key")` |
| `encryptTextByTXDEE(text, code)` | TXDEE 文本加密 | `crypto.encryptTextByTXDEE("hello", "key")` |
| `decryptTextByTXDEE(hexStr, code)` | TXDEE 文本解密 | `crypto.decryptTextByTXDEE("...", "key")` |
| `encryptTextByTXDEF(text, code)` | TXDEF 文本加密 | `crypto.encryptTextByTXDEF("hello", "key")` |
| `decryptTextByTXDEF(hexStr, code)` | TXDEF 文本解密 | `crypto.decryptTextByTXDEF("...", "key")` |
| `encryptText(text, code)` | 默认文本加密 | `crypto.encryptText("hello", "key")` |
| `decryptText(hexStr, code)` | 默认文本解密 | `crypto.decryptText("...", "key")` |
| `encryptData(data, code)` | 默认数据加密 | `crypto.encryptData([1,2,3], "key")` |
| `decryptData(data, code)` | 默认数据解密 | `crypto.decryptData(encData, "key")` |
| `aesEncrypt(data, key, mode?)` | AES 加密 | `crypto.aesEncrypt(data, "16bytekey1234567")` |
| `aesDecrypt(data, key, mode?)` | AES 解密 | `crypto.aesDecrypt(encData, "16bytekey1234567")` |

### fmt

格式化工具。

| 函数 | 说明 | 示例 |
|------|------|------|
| `sprintf(format, args...)` | 格式化字符串 | `fmt.sprintf("名字: %s, 年龄: %d", "张三", 25)` |
| `printf(format, args...)` | 格式化打印 | `fmt.printf("值: %d\n", 42)` |

### array

扩展数组工具。

| 函数 | 说明 | 示例 |
|------|------|------|
| `map(arr, fn)` | 映射元素 | `array.map([1, 2, 3], fn(x) { x * 2 })` |
| `filter(arr, fn)` | 过滤元素 | `array.filter([1, 2, 3], fn(x) { x > 1 })` |
| `reduce(arr, fn, init)` | 归约元素 | `array.reduce([1, 2, 3], fn(a, b) { a + b }, 0)` |
| `forEach(arr, fn)` | 遍历元素 | `array.forEach([1, 2, 3], fn(x) { pln(x) })` |

### collections

集合工具（集合、栈、队列）。

### bytes

字节数组操作。

### csv

CSV 文件解析和写入。

### debug

调试工具。

### encoding

编解码工具（Base64、Hex）。

### env

环境变量工具。

### log

日志工具。

### net

网络工具。

### sort

高级排序工具。

### strconv

字符串转换工具。

### text

文本处理工具。

### uuid

UUID 生成。

---

## 另见

- [语言参考](LANGUAGE.md) - 完整语言语法
- [标准库](STDLIB_zh.md) - 标准库概述
- [嵌入指南](EMBEDDING.md) - 在 Go 应用中使用 Xxlang
