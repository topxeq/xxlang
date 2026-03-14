# Xxlang 标准库文档

Xxlang 提供了一套标准库模块，可以通过 `import` 语句导入使用。

## 导入语法

```xxl
import "std/modulename"

// 使用模块中的函数
modulename.functionName()
```

## 可用模块

### std/io - 输入输出

文件读写和控制台输入输出。

```xxl
import "std/io"

// 控制台输出
io.print("Hello")
io.println("World")
io.printf("Value: %d\n", 42)

// 控制台输入
var line = io.readLine()

// 文件操作
var content = io.readFile("file.txt")
io.writeFile("output.txt", "content")
io.appendFile("log.txt", "new line\n")

// 文件系统
var exists = io.exists("file.txt")
io.remove("file.txt")
io.mkdir("newdir")
var dir = io.cwd()  // 当前工作目录

// 环境变量
var path = io.env("PATH")
io.setEnv("MY_VAR", "value")

// 命令行参数
var args = io.args()

// 退出程序
io.exit(0)
```

### std/time - 时间函数

时间获取、格式化和休眠。

```xxl
import "std/time"

// 获取时间戳
var ts = time.unix()        // 秒
var ms = time.unixMs()      // 毫秒
var ns = time.unixNano()    // 纳秒

// 获取当前时间各部分
var now = time.now()        // 返回 {year, month, day, hour, minute, second, nanosecond}
var y = time.year()
var m = time.month()
var d = time.day()
var h = time.hour()
var min = time.minute()
var sec = time.second()
var wd = time.weekday()     // 0=Sunday, 6=Saturday

// 休眠
time.sleep(1000)            // 毫秒
time.sleepSec(1)            // 秒

// 格式化时间
// Go 时间格式: 2006-01-02 15:04:05
var formatted = time.format("2006-01-02 15:04:05")
var custom = time.formatUnix(1704067200, "2006-01-02")

// 解析时间
var timestamp = time.parse("2006-01-02", "2024-01-01")

// 时间计算
var elapsed = time.since(startMs)  // 毫秒
var tomorrow = time.addDays(1)
var nextMonth = time.addMonths(1)
var nextYear = time.addYears(1)

// 日期工具
var isLeap = time.isLeapYear(2024)
var days = time.daysInMonth(2024, 2)  // 29
```

### std/math - 数学函数

```xxl
import "std/math"

// 基本运算
var a = math.abs(-5)        // 5
var m = math.max(1, 2)      // 2
var n = math.min(1, 2)      // 1

// 幂运算
var p = math.pow(2, 10)     // 1024
var s = math.sqrt(16)       // 4

// 取整
var f = math.floor(3.7)     // 3
var c = math.ceil(3.2)      // 4
var r = math.round(3.5)     // 4

// 三角函数
var sin = math.sin(3.14159/2)
var cos = math.cos(0)
var tan = math.tan(1)

// 对数
var ln = math.log(2.71828)
var lg = math.log10(100)

// 随机数
var rand = math.random()    // 0.0 - 1.0
var randInt = math.randomInt(1, 100)  // 1 - 100

// 常量
var pi = math.PI
var e = math.E
```

### std/string - 字符串工具

```xxl
import "std/string"

// 构建
var s = string.build("Hello", " ", "World")

// 重复
var rep = string.repeat("ab", 3)  // "ababab"

// 分割和连接
var parts = string.split("a,b,c", ",")
var joined = string.join(["a", "b", "c"], "-")

// 查找和替换
var has = string.contains("hello", "ell")  // true
var idx = string.indexOf("hello", "l")     // 2
var replaced = string.replace("hello", "l", "L")

// 大小写
var up = string.toUpper("hello")  // "HELLO"
var lo = string.toLower("HELLO")  // "hello"

// 修剪
var t = string.trim("  hello  ")   // "hello"
var tl = string.trimLeft("  hello")
var tr = string.trimRight("hello  ")

// 前缀后缀
var hasPre = string.hasPrefix("hello", "he")
var hasSuf = string.hasSuffix("hello", "lo")
```

### std/array - 数组工具

```xxl
import "std/array"

// 映射
var doubled = array.map([1, 2, 3], func(x) { return x * 2 })

// 过滤
var evens = array.filter([1, 2, 3, 4], func(x) { return x % 2 == 0 })

// 归约
var sum = array.reduce([1, 2, 3], 0, func(acc, x) { return acc + x })

// 查找
var found = array.find([1, 2, 3], func(x) { return x > 1 })
var idx = array.findIndex([1, 2, 3], func(x) { return x > 1 })
var has = array.some([1, 2, 3], func(x) { return x > 2 })
var all = array.every([1, 2, 3], func(x) { return x > 0 })

// 排序
var sorted = array.sort([3, 1, 2])
var sortedDesc = array.sortBy([{"name": "b"}, {"name": "a"}], func(a, b) { return a.name > b.name })

// 切片
var flat = array.flatten([[1, 2], [3, 4]])
var unique = array.unique([1, 2, 2, 3])
```

### std/json - JSON 处理

```xxl
import "std/json"

// 编码
var s = json.encode({"name": "test", "value": 42})

// 解码
var obj = json.decode("{\"name\": \"test\"}")
var name = obj.name
```

### std/regex - 正则表达式

```xxl
import "std/regex"

// 匹配
var matched = regex.match("hello world", "hello")

// 查找
var found = regex.find("hello 123 world", "\\d+")

// 替换
var replaced = regex.replace("hello 123", "\\d+", "456")

// 分割
var parts = regex.split("a1b2c3", "\\d+")
```

### std/crypto - 加密

```xxl
import "std/crypto"

// 哈希
var md5 = crypto.md5("hello")
var sha256 = crypto.sha256("hello")

// Base64
var encoded = crypto.base64Encode("hello")
var decoded = crypto.base64Decode(encoded)

// Hex
var hex = crypto.hexEncode("hello")
var raw = crypto.hexDecode(hex)
```

## 使用示例

### 计时程序

```xxl
import "std/time"
import "std/io"

var start = time.unixMs()

// 执行一些操作
var sum = 0
for (var i = 0; i < 1000000; i++) {
    sum = sum + i
}

var elapsed = time.since(start)
io.println("Elapsed: " + string(elapsed) + "ms")
```

### 读取配置文件

```xxl
import "std/io"
import "std/json"

var configStr = io.readFile("config.json")
var config = json.decode(configStr)

io.println("Database: " + config.database)
io.println("Port: " + string(config.port))
```

### 日志记录

```xxl
import "std/io"
import "std/time"

func log(message) {
    var timestamp = time.format("2006-01-02 15:04:05")
    io.appendFile("app.log", "[" + timestamp + "] " + message + "\n")
}

log("Application started")
log("Processing data...")
log("Done")
```

## 注意事项

1. 所有标准库模块都以 `std/` 开头
2. 导入后使用模块名（不含 `std/`）作为前缀调用函数
3. 某些函数可能返回 Error 对象，请检查返回值
