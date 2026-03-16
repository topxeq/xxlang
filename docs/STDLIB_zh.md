# Xxlang 标准库参考手册

## 概述

Xxlang 包含一个按模块组织的完整标准库。所有模块使用 `import` 语句导入。

```xxl
import "io"
io.println("你好，世界！")
```

## 目录

- [io](#io) - 输入输出操作
- [os](#os) - 操作系统工具和配置
- [string](#string) - 字符串工具
- [math](#math) - 数学函数
- [array](#array) - 数组工具
- [json](#json) - JSON 编解码
- [regex](#regex) - 正则表达式
- [crypto](#crypto) - 加密函数
- [time](#time) - 时间日期函数

---

## io

输入输出操作，包括读写和控制台交互。

### 控制台函数

#### print(args...)
打印参数，不换行。

```xxl
print("你好")
print(" ")
print("世界")
// 输出: 你好 世界
```

#### println(args...)
打印参数，用空格分隔，末尾换行。

```xxl
println("你好", "世界")  // "你好 世界\n"
println(42, "是答案")     // "42 是答案\n"
```

#### printf(format, args...)
使用 Go 风格格式说明符格式化打印。

```xxl
printf("姓名: %s, 年龄: %d\n", "张三", 30)
printf("圆周率: %.2f\n", 3.14159)  // "圆周率: 3.14"
```

#### readLine()
从标准输入读取一行。

```xxl
print("请输入姓名: ")
var name = readLine()
println("你好, " + name)
```

### 文件函数

#### readFile(path)
读取整个文件内容为字符串。

```xxl
var content = readFile("data.txt")
println(content)
```

#### writeFile(path, content)
将字符串内容写入文件。

```xxl
writeFile("output.txt", "你好，文件！")
```

#### appendFile(path, content)
追加字符串内容到文件。

```xxl
appendFile("log.txt", "新日志条目\n")
```

#### exists(path)
如果文件或目录存在则返回 true。

```xxl
if (exists("config.json")) {
    println("找到配置文件")
}
```

#### remove(path)
删除文件。

```xxl
remove("temp.txt")
```

#### mkdir(path)
创建目录（包括父目录）。

```xxl
mkdir("path/to/directory")
```

### 系统函数

#### cwd()
返回当前工作目录。

```xxl
println(cwd())  // "/home/user/project"
```

#### exit(code)
以状态码退出。

```xxl
exit(0)   // 成功
exit(1)   // 错误
```

#### env(key)
获取环境变量值。

```xxl
var home = env("HOME")
var path = env("PATH")
```

#### setEnv(key, value)
设置环境变量。

```xxl
setEnv("DEBUG", "true")
```

#### args()
返回命令行参数数组。

```xxl
var args = args()
println(args[0])  // 程序名
```

---

## os

操作系统工具和配置管理。

### 配置

#### getConfigObj()
返回 Xxlang 配置对象（映射类型）。配置从 JSON 文件读取，搜索优先级如下：

1. `~/.xxl/settings.json`（用户主目录）
2. `/.xxl/settings.json`（Linux/Unix 系统）
3. `C:\.xxl\settings.json`（Windows 系统）

如果未找到配置文件，返回空映射。

```xxl
import "os"
var cfg = os.getConfigObj()
println(cfg["cloudUrlBase"])

// 示例配置文件 (~/.xxl/settings.json):
// {
//   "cloudUrlBase": "https://script.topget.org/",
//   "timeout": 30,
//   "debug": true
// }
```

#### getConfigStr(name)
从 `.cfg` 文件读取配置字符串。文件搜索顺序如下：

1. `~/.xxl/<name>.cfg`（用户主目录）
2. `/.xxl/<name>.cfg`（Linux/Unix 系统）
3. `C:\.xxl\<name>.cfg`（Windows 系统）

如果文件不存在，返回 `null`。

```xxl
import "os"
var token = os.getConfigStr("api_token")
if (token != null) {
    println("找到令牌: " + token)
}
```

#### setConfigStr(name, value)
将配置字符串写入用户主目录下的 `.cfg` 文件（`~/.xxl/<name>.cfg`）。如果 `.xxl` 目录不存在，会自动创建。

```xxl
import "os"
os.setConfigStr("api_token", "my-secret-token")

// 之后读取
var token = os.getConfigStr("api_token")
println(token)  // "my-secret-token"
```

### 系统信息

#### platform()
返回当前操作系统名称。

```xxl
println(os.platform())  // "linux"、"windows"、"darwin"
```

#### arch()
返回 CPU 架构。

```xxl
println(os.arch())  // "amd64"、"arm64"
```

#### hostname()
返回系统主机名。

```xxl
println(os.hostname())
```

#### home()
返回用户主目录。

```xxl
println(os.home())  // "/home/user" 或 "C:\Users\user"
```

#### temp()
返回系统临时目录。

```xxl
println(os.temp())  // Unix 上为 "/tmp"
```

#### cpus()
返回 CPU 核心数。

```xxl
println(os.cpus())  // 8
```

### 文件系统操作

#### join(paths...)
连接路径组件。

```xxl
os.join("a", "b", "c.txt")  // Unix: "a/b/c.txt", Windows: "a\b\c.txt"
```

#### base(path)
返回路径的最后一个元素。

```xxl
os.base("/path/to/file.txt")  // "file.txt"
```

#### dir(path)
返回路径的目录部分。

```xxl
os.dir("/path/to/file.txt")  // "/path/to"
```

#### ext(path)
返回文件扩展名。

```xxl
os.ext("/path/to/file.txt")  // ".txt"
```

#### abs(path)
返回绝对路径。

```xxl
println(os.abs("./file.txt"))  // "/current/dir/file.txt"
```

#### isAbs(path)
判断是否为绝对路径。

```xxl
os.isAbs("/path/to/file")  // true
os.isAbs("./file")          // false
```

#### stat(path)
返回文件信息数组 [名称, 大小, 是否目录, 修改时间]。

```xxl
var info = os.stat("file.txt")
println(info[0])  // 名称
println(info[1])  // 大小（字节）
println(info[2])  // 是否目录
println(info[3])  // 修改时间
```

#### isDir(path)
判断是否为目录。

```xxl
os.isDir("/path/to/dir")  // true
```

#### isFile(path)
判断是否为文件。

```xxl
os.isFile("/path/to/file.txt")  // true
```

#### listDir(path)
返回目录条目名称数组。

```xxl
var files = os.listDir(".")
for (f in files) {
    println(f)
}
```

#### mkdir(path)
创建目录。

```xxl
os.mkdir("path/to/new/dir")
```

#### rename(oldPath, newPath)
重命名文件或目录。

```xxl
os.rename("old.txt", "new.txt")
```

#### copy(srcPath, dstPath)
复制文件。

```xxl
os.copy("source.txt", "dest.txt")
```

#### chmod(path, mode)
修改文件权限（仅 Unix）。

```xxl
os.chmod("script.sh", 0o755)
```

### 进程执行

#### exec(command)
执行 shell 命令，返回 [输出, 退出码, 错误]。

```xxl
var result = os.exec("ls -la")
println(result[0])  // 输出
println(result[1])  // 退出码
```

#### shell(command)
通过系统 shell 执行命令。

```xxl
var result = os.shell("echo hello")
println(result[0])  // "hello\n"
```

### 临时文件

#### tempFile(pattern)
创建临时文件并返回路径。

```xxl
var tmp = os.tempFile("myapp-*.txt")
// 使用文件...
```

#### tempDir(pattern)
创建临时目录并返回路径。

```xxl
var tmpDir = os.tempDir("myapp-*")
// 使用目录...
```

---

## string

字符串操作工具。

#### len(s)
返回字符串长度。

```xxl
len("hello")  // 5
```

#### upper(s)
转换为大写。

```xxl
upper("hello")  // "HELLO"
```

#### lower(s)
转换为小写。

```xxl
lower("HELLO")  // "hello"
```

#### trim(s)
移除首尾空白字符。

```xxl
trim("  hello  ")  // "hello"
```

#### substr(s, start, end)
提取子字符串。

```xxl
substr("hello", 1, 4)  // "ell"
```

#### split(s, sep)
用分隔符分割字符串。

```xxl
split("a,b,c", ",")  // ["a", "b", "c"]
```

#### join(arr, sep)
用分隔符连接数组元素。

```xxl
join(["a", "b", "c"], "-")  // "a-b-c"
```

#### containsStr(s, substr)
如果字符串包含子串则返回 true。

```xxl
containsStr("hello", "ell")  // true
containsStr("hello", "xyz")  // false
```

#### replace(s, old, new)
替换所有匹配项。

```xxl
replace("hello world", "l", "L")  // "heLLo worLd"
```

#### startsWith(s, prefix)
如果字符串以前缀开头则返回 true。

```xxl
startsWith("hello", "he")  // true
```

#### endsWith(s, suffix)
如果字符串以后缀结尾则返回 true。

```xxl
endsWith("hello", "lo")  // true
```

#### indexOf(s, substr)
返回第一次出现的索引，未找到返回 -1。

```xxl
indexOf("hello", "l")   // 2
indexOf("hello", "x")   // -1
```

---

## math

数学函数。

#### abs(x)
返回绝对值。

```xxl
abs(-42)    // 42
abs(-3.14)  // 3.14
```

#### floor(x)
返回向下取整结果。

```xxl
floor(3.7)   // 3
floor(-3.7)  // -4
```

#### ceil(x)
返回向上取整结果。

```xxl
ceil(3.2)    // 4
ceil(-3.2)   // -3
```

#### round(x)
返回最接近的整数。

```xxl
round(3.4)   // 3
round(3.6)   // 4
```

#### sqrt(x)
返回平方根。

```xxl
sqrt(16)   // 4
sqrt(2)    // 1.414...
```

#### pow(base, exp)
返回 base 的 exp 次方。

```xxl
pow(2, 8)   // 256
pow(10, 3)  // 1000
```

#### min(a, b)
返回较小值。

```xxl
min(3, 7)   // 3
```

#### max(a, b)
返回较大值。

```xxl
max(3, 7)   // 7
```

#### sin(x), cos(x), tan(x)
三角函数（弧度）。

```xxl
sin(0)              // 0
cos(0)              // 1
tan(0)              // 0
```

#### log(x), log10(x)
自然对数和以10为底的对数。

```xxl
log(2.718281828)   // ~1
log10(100)         // 2
```

#### random()
返回 0 到 1 之间的随机浮点数。

```xxl
var r = random()  // 0.0 <= r < 1.0
```

#### randomInt(min, max)
返回范围内的随机整数。

```xxl
var die = randomInt(1, 6)
```

---

## array

数组操作工具。

#### len(arr)
返回数组长度。

```xxl
len([1, 2, 3])  // 3
```

#### first(arr)
返回第一个元素，空数组返回 null。

```xxl
first([1, 2, 3])  // 1
first([])         // null
```

#### last(arr)
返回最后一个元素，空数组返回 null。

```xxl
last([1, 2, 3])  // 3
last([])         // null
```

#### push(arr, value)
返回追加了元素的新数组。

```xxl
push([1, 2], 3)  // [1, 2, 3]
```

#### pop(arr)
返回移除最后一个元素后的新数组。

```xxl
pop([1, 2, 3])  // [1, 2]
```

#### sort(arr)
返回升序排序后的数组。

```xxl
sort([3, 1, 4, 1, 5])  // [1, 1, 3, 4, 5]
```

#### reverse(arr)
返回反转后的数组。

```xxl
reverse([1, 2, 3])  // [3, 2, 1]
```

#### sum(arr)
返回数值元素的和。

```xxl
sum([1, 2, 3, 4, 5])  // 15
```

#### avg(arr)
返回数值元素的平均值。

```xxl
avg([1, 2, 3, 4, 5])  // 3
```

#### indexOf(arr, value)
返回值的索引，未找到返回 -1。

```xxl
indexOf([1, 2, 3], 2)  // 1
indexOf([1, 2, 3], 5)  // -1
```

#### containsArr(arr, value)
如果数组包含值则返回 true。

```xxl
containsArr([1, 2, 3], 2)  // true
```

#### concat(arr1, arr2)
连接两个数组。

```xxl
concat([1, 2], [3, 4])  // [1, 2, 3, 4]
```

#### slice(arr, start, end)
返回子数组。

```xxl
slice([1, 2, 3, 4, 5], 1, 4)  // [2, 3, 4]
```

#### isEmpty(arr)
如果数组为空则返回 true。

```xxl
isEmpty([])      // true
isEmpty([1])     // false
```

---

## json

JSON 编解码。

#### parse(jsonString)
将 JSON 字符串解析为 Xxlang 值。

```xxl
var data = parse('{"name": "张三", "age": 30}')
println(data["name"])  // "张三"

var arr = parse('[1, 2, 3]')
println(arr[0])  // 1
```

#### stringify(value, indent)
将 Xxlang 值转换为 JSON 字符串。

```xxl
var obj = {"name": "张三", "age": 30}
println(stringify(obj))        // {"name":"张三","age":30}
println(stringify(obj, "  "))  // 带2空格缩进的格式化输出
println(stringify(obj, 4))     // 带4空格缩进的格式化输出
```

#### encode(value)
不带缩进的 stringify 别名。

```xxl
encode({"a": 1})  // '{"a":1}'
```

#### decode(jsonString)
parse 的别名。

```xxl
decode('{"x": 10}')["x"]  // 10
```

---

## regex

正则表达式操作。

#### match(pattern, str)
如果字符串匹配模式则返回 true。

```xxl
match("\\d+", "123")       // true
match("[a-z]+", "Hello")   // false
```

#### find(pattern, str)
返回第一个匹配或 null。

```xxl
find("\\d+", "abc123def")  // "123"
```

#### findAll(pattern, str)
返回所有匹配的数组。

```xxl
findAll("\\d+", "a1b22c333")  // ["1", "22", "333"]
```

#### replace(pattern, str, replacement)
替换所有匹配。

```xxl
replace("\\d+", "a1b2c3", "X")  // "aXbXcX"
```

#### split(pattern, str)
按模式分割字符串。

```xxl
split("\\s+", "a  b   c")  // ["a", "b", "c"]
```

---

## crypto

加密函数。

#### md5(s)
返回 MD5 哈希十六进制字符串。

```xxl
md5("hello")  // "5d41402abc4b2a76b9719d911017c592"
```

#### sha1(s)
返回 SHA1 哈希十六进制字符串。

```xxl
sha1("hello")  // "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
```

#### sha256(s)
返回 SHA256 哈希十六进制字符串。

```xxl
sha256("hello")  // "2cf24dba5fb0a30e26e83b2ac5b9e29e..."
```

#### sha512(s)
返回 SHA512 哈希十六进制字符串。

```xxl
sha512("hello")  // "9b71d224bd62f3785d96d46ad3ea3d..."
```

#### hmac(algorithm, key, message)
返回 HMAC 哈希。

```xxl
hmac("sha256", "secret", "message")  // 十六进制字符串
```

#### base64Encode(s)
将字符串编码为 base64。

```xxl
base64Encode("hello")  // "aGVsbG8="
```

#### base64Decode(s)
解码 base64 字符串。

```xxl
base64Decode("aGVsbG8=")  // "hello"
```

#### base64UrlEncode(s)
URL 安全的 base64 编码。

```xxl
base64UrlEncode("hello?world")
```

#### base64UrlDecode(s)
URL 安全的 base64 解码。

```xxl
base64UrlDecode(base64UrlEncode("hello?world"))  // "hello?world"
```

#### hexEncode(bytes)
编码为十六进制字符串。

```xxl
hexEncode("hello")  // "68656c6c6f"
```

#### hexDecode(hex)
解码十六进制字符串。

```xxl
hexDecode("68656c6c6f")  // "hello"
```

#### randomBytes(n)
返回 n 个随机字节作为字符串。

```xxl
var bytes = randomBytes(16)
```

#### randomHex(n)
返回 n 个随机字节的十六进制字符串。

```xxl
randomHex(16)  // "a1b2c3d4..." (32个十六进制字符)
```

---

## time

处理时间戳和持续时间的时间日期函数。

### 时间戳

#### unix()
返回当前 Unix 时间戳（秒）。

```xxl
import "time"
var ts = time.unix()  // 例如: 1710422400
```

#### unixMs()
返回当前 Unix 时间戳（毫秒）。

```xxl
var ms = time.unixMs()  // 例如: 1710422400000
```

#### unixNano()
返回当前 Unix 时间戳（纳秒）。

```xxl
var ns = time.unixNano()  // 例如: 1710422400000000000
```

### 日期/时间组件

#### now()
返回当前时间作为包含组件的映射。

```xxl
var t = time.now()
println(t["year"])     // 2026
println(t["month"])    // 3 (三月)
println(t["day"])      // 14
println(t["hour"])     // 20
println(t["minute"])   // 30
println(t["second"])   // 45
println(t["nanosecond"])  // 123456789
```

#### year(), month(), day(), hour(), minute(), second()
返回当前日期/时间组件。

```xxl
println(time.year())    // 2026
println(time.month())   // 3
println(time.day())     // 14
println(time.hour())    // 20
println(time.minute())  // 30
println(time.second())  // 45
```

#### weekday()
返回当前星期几（0=周日，1=周一，...，6=周六）。

```xxl
var wd = time.weekday()  // 例如: 4 (周四)
```

### 格式化

#### format(layout)
使用 Go 风格布局格式化当前时间。

```xxl
println(time.format("2006-01-02"))           // "2026-03-14"
println(time.format("15:04:05"))             // "20:30:45"
println(time.format("2006-01-02 15:04:05"))  // "2026-03-14 20:30:45"
```

#### formatUnix(timestamp, layout)
格式化 Unix 时间戳。

```xxl
var ts = time.unix()
println(time.formatUnix(ts, "2006-01-02"))
```

#### parse(layout, value)
解析时间字符串并返回 Unix 时间戳。

```xxl
var ts = time.parse("2006-01-02", "2026-03-14")
```

### 休眠

#### sleep(ms)
暂停执行指定的毫秒数。

```xxl
println("开始...")
time.sleep(1000)  // 休眠1秒
println("完成！")
```

#### sleepSec(seconds)
暂停执行指定的秒数。

```xxl
time.sleepSec(2)  // 休眠2秒
```

### 日期运算

#### addDays(days)
向当前时间添加天数，返回 Unix 时间戳。

```xxl
var tomorrow = time.addDays(1)
var nextWeek = time.addDays(7)
```

#### addMonths(months)
向当前时间添加月数。

```xxl
var nextMonth = time.addMonths(1)
```

#### addYears(years)
向当前时间添加年数。

```xxl
var nextYear = time.addYears(1)
```

#### since(startMs)
返回自给定时间戳以来经过的毫秒数。

```xxl
var start = time.unixMs()
// ... 执行一些操作 ...
var elapsed = time.since(start)
println("耗时 " + elapsed.toStr() + " 毫秒")
```

### 日历

#### isLeapYear(year)
如果是闰年则返回 true。

```xxl
time.isLeapYear(2024)  // true
time.isLeapYear(2023)  // false
```

#### daysInMonth(year, month)
返回月份的天数。

```xxl
time.daysInMonth(2024, 2)  // 29 (闰年)
time.daysInMonth(2023, 2)  // 28
time.daysInMonth(2024, 1)  // 31
```

---

## 内置函数（全局）

这些函数无需导入任何模块即可使用。

### 类型函数

#### typeOf(value)
返回类型名称字符串。

```xxl
typeOf(42)        // "INT"
typeOf("hello")   // "STRING"
typeOf([1, 2])    // "ARRAY"
```

#### len(value)
返回字符串、数组或映射的长度。

```xxl
len("hello")      // 5
len([1, 2, 3])    // 3
len({"a": 1})     // 1
```

### 类型转换

#### int(value)
转换为整数。

```xxl
int(3.14)      // 3
int("42")      // 42
int(true)      // 1
```

#### float(value)
转换为浮点数。

```xxl
float(42)      // 42.0
float("3.14")  // 3.14
```

#### string(value)
转换为字符串。

```xxl
string(42)     // "42"
string(3.14)   // "3.14"
string(true)   // "true"
```

### 映射操作

#### keys(map)
返回映射键的数组。

```xxl
keys({"a": 1, "b": 2})  // ["a", "b"]
```

#### values(map)
返回映射值的数组。

```xxl
values({"a": 1, "b": 2})  // [1, 2]
```

#### hasKey(map, key)
如果映射包含键则返回 true。

```xxl
hasKey({"a": 1}, "a")   // true
hasKey({"a": 1}, "b")   // false
```

#### delete(map, key)
返回移除键后的映射。

```xxl
delete({"a": 1, "b": 2}, "a")  // {"b": 2}
```

### 范围

#### range(start, stop)
返回整数数组。

```xxl
range(0, 5)    // [0, 1, 2, 3, 4]
range(1, 4)    // [1, 2, 3]
```

### 断言

#### assert(condition)
如果条件为 false 则抛出错误。

```xxl
assert(1 + 1 == 2)
assert(x > 0, "x 必须为正数")
```

### 字符串工具

#### repeat(s, n)
返回重复 n 次的字符串。

```xxl
repeat("ab", 3)  // "ababab"
repeat("x", 5)   // "xxxxx"
```

#### lpad(s, length, padChar)
左侧填充字符串到指定长度。

```xxl
lpad("5", 4, "0")     // "0005"
lpad("hello", 10)      // "     hello"
```

#### rpad(s, length, padChar)
右侧填充字符串到指定长度。

```xxl
rpad("5", 4, "0")      // "5000"
rpad("hello", 10)      // "hello     "
```

#### charAt(s, index)
返回指定索引的字符，越界返回 null。

```xxl
charAt("hello", 1)  // "e"
charAt("hello", 10) // null
```

#### trimLeft(s, cutset)
移除前导字符。

```xxl
trimLeft("  hello")          // "hello"
trimLeft("xxhelloxx", "x")   // "helloxx"
```

#### trimRight(s, cutset)
移除尾部字符。

```xxl
trimRight("hello  ")         // "hello"
trimRight("xxhelloxx", "x")  // "xxhello"
```

### 类型检查

#### isEmpty(value)
如果值是空字符串、空数组、空映射或 null 则返回 true。

```xxl
isEmpty("")          // true
isEmpty([])          // true
isEmpty({})          // true
isEmpty(null)        // true
isEmpty([1, 2])      // false
```

#### isString(value), isNumber(value), isInt(value), isFloat(value)
如果值是指定类型则返回 true。

```xxl
isString("hello")    // true
isNumber(42)         // true
isNumber(3.14)       // true
isInt(42)            // true
isFloat(3.14)        // true
```

#### isArray(value), isMap(value), isBool(value), isNull(value), isFunction(value)
如果值是指定类型则返回 true。

```xxl
isArray([1, 2, 3])     // true
isMap({"a": 1})         // true
isBool(true)            // true
isNull(null)            // true
isFunction(len)         // true
```

### 数学工具

#### round(value, precision)
四舍五入到最接近的整数或指定小数位。

```xxl
round(3.7)           // 4
round(3.14159, 2)    // 3.14
round(3.14159, 4)    // 3.1416
```

#### clamp(value, min, max)
将值限制在 [min, max] 范围内。

```xxl
clamp(15, 0, 10)     // 10
clamp(-5, 0, 10)     // 0
clamp(5, 0, 10)      // 5
```

#### sign(value)
返回数字的符号（-1、0 或 1）。

```xxl
sign(-42)    // -1
sign(0)      // 0
sign(42)     // 1
```

#### random()
返回 0 到 1 之间的随机浮点数。

```xxl
var r = random()  // 0.0 <= r < 1.0
```

#### randomInt(min, max)
返回 [min, max] 范围内的随机整数。

```xxl
var die = randomInt(1, 6)   // 1, 2, 3, 4, 5 或 6
```

### 数组工具

#### unique(arr)
返回移除重复值后的数组。

```xxl
unique([1, 2, 2, 3, 3, 3])  // [1, 2, 3]
```

#### flatten(arr, depth)
展平嵌套数组。

```xxl
flatten([[1, 2], [3, 4]])           // [1, 2, 3, 4]
flatten([[[1]], [2]], 1)             // [[1], 2]
```

#### without(arr, values...)
返回移除指定值后的数组。

```xxl
without([1, 2, 3, 4], 2, 4)  // [1, 3]
```

#### take(arr, n)
返回前 n 个元素。

```xxl
take([1, 2, 3, 4, 5], 3)  // [1, 2, 3]
```

#### drop(arr, n)
返回移除前 n 个元素后的数组。

```xxl
drop([1, 2, 3, 4, 5], 2)  // [3, 4, 5]
```

### 映射工具

#### merge(map1, map2, ...)
合并多个映射（后面的映射覆盖前面的）。

```xxl
merge({"a": 1}, {"b": 2})           // {"a": 1, "b": 2}
merge({"a": 1}, {"a": 2, "b": 3})    // {"a": 2, "b": 3}
```

#### entries(map)
返回 [键, 值] 对数组。

```xxl
entries({"x": 10, "y": 20})  // [["x", 10], ["y", 20]]
```

### 格式化

#### format(template, args...)
使用 Go 风格格式说明符格式化字符串。

```xxl
format("你好 %s，你 %d 岁", "张三", 30)  // "你好 张三，你 30 岁"
format("圆周率是 %.2f", 3.14159)                 // "圆周率是 3.14"
```
