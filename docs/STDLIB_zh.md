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
- [chars](#chars) - Unicode 字符处理
- [stringbuilder](#stringbuilder) - 高效字符串拼接
- [bytesbuffer](#bytesbuffer) - 高效字节缓冲区操作
- [math](#math) - 数学函数
- [array](#array) - 数组工具
- [json](#json) - JSON 编解码
- [regex](#regex) - 正则表达式
- [crypto](#crypto) - 加密函数
- [time](#time) - 时间日期函数
- [fmt](#fmt) - 格式化工具
- [encoding](#encoding) - Base64 和十六进制编解码
- [uuid](#uuid) - UUID 生成
- [debug](#debug) - 调试工具

---

## io

输入输出操作，包括读写和控制台交互。

### 控制台函数

#### print(args...)
打印参数，不换行。

```xxl
io.print("你好")
io.print(" ")
io.print("世界")
// 输出: 你好 世界
```

#### println(args...)
打印参数，用空格分隔，末尾换行。

```xxl
io.println("你好", "世界")  // "你好 世界\n"
io.println(42, "是答案")     // "42 是答案\n"
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
io.print("请输入姓名: ")
var name = io.readLine()
io.println("你好, " + name)
```

### 文件函数

#### readFile(path)
读取整个文件内容为字符串。

```xxl
var content = io.readFile("data.txt")
io.println(content)
```

#### writeFile(path, content)
将字符串内容写入文件。

```xxl
io.writeFile("output.txt", "你好，文件！")
```

#### appendFile(path, content)
追加字符串内容到文件。

```xxl
io.appendFile("log.txt", "新日志条目\n")
```

#### exists(path)
如果文件或目录存在则返回 true。

```xxl
if (io.exists("config.json")) {
    io.println("找到配置文件")
}
```

#### remove(path)
删除文件。

```xxl
io.remove("temp.txt")
```

#### mkdir(path)
创建目录（包括父目录）。

```xxl
io.mkdir("path/to/directory")
```

### 系统函数

#### cwd()
返回当前工作目录。

```xxl
io.println(io.cwd())  // "/home/user/project"
```

#### exit(code)
以状态码退出。

```xxl
io.exit(0)   // 成功
io.exit(1)   // 错误
```

#### env(key)
获取环境变量值。

```xxl
var home = io.env("HOME")
var path = io.env("PATH")
```

#### setEnv(key, value)
设置环境变量。

```xxl
io.setEnv("DEBUG", "true")
```

#### args()
返回命令行参数数组。

```xxl
var args = io.args()
io.println(args[0])  // 程序名
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
pln(cfg["cloudUrlBase"])

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
    pln("找到令牌: " + token)
}
```

#### setConfigStr(name, value)
将配置字符串写入用户主目录下的 `.cfg` 文件（`~/.xxl/<name>.cfg`）。如果 `.xxl` 目录不存在，会自动创建。

```xxl
import "os"
os.setConfigStr("api_token", "my-secret-token")

// 之后读取
var token = os.getConfigStr("api_token")
pln(token)  // "my-secret-token"
```

### 系统信息

#### platform()
返回当前操作系统名称。

```xxl
pln(os.platform())  // "linux"、"windows"、"darwin"
```

#### arch()
返回 CPU 架构。

```xxl
pln(os.arch())  // "amd64"、"arm64"
```

#### hostname()
返回系统主机名。

```xxl
pln(os.hostname())
```

#### home()
返回用户主目录。

```xxl
pln(os.home())  // "/home/user" 或 "C:\Users\user"
```

#### temp()
返回系统临时目录。

```xxl
pln(os.temp())  // Unix 上为 "/tmp"
```

#### cpus()
返回 CPU 核心数。

```xxl
pln(os.cpus())  // 8
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
pln(os.abs("./file.txt"))  // "/current/dir/file.txt"
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
pln(info[0])  // 名称
pln(info[1])  // 大小（字节）
pln(info[2])  // 是否目录
pln(info[3])  // 修改时间
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
    pln(f)
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
pln(result[0])  // 输出
pln(result[1])  // 退出码
```

#### shell(command)
通过系统 shell 执行命令。

```xxl
var result = os.shell("echo hello")
pln(result[0])  // "hello\n"
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

## chars

`chars` 类型提供正确的 Unicode 字符处理，操作基于字符（码点）而非字节。这对于正确处理包含中文、日文、韩文、emoji 等 Unicode 字符的文本至关重要。

### 为什么需要 chars？

在 Xxlang 中，`string` 类型是面向字节的（类似 Go），这意味着：
- `len("中文")` 返回 6（字节），而不是 2（字符）
- `"中文"[0]` 返回字节值，而不是字符
- `substr` 使用字节索引

`chars` 类型提供面向字符的操作：
- `len(toChars("中文"))` 返回 2（字符）
- `toChars("中文")[0]` 返回 "中"（完整字符）
- `subStr` 使用字符索引

### toChars(s)

将字符串转换为 chars 数组，用于基于字符的操作。

```xxl
var s = "Hello世界🎉"
var c = toChars(s)

pln(len(s))      // 15（字节）
pln(len(c))      // 8（字符）
```

### charLen(s)

返回字符串中 Unicode 字符的数量（无需创建 chars 对象）。

```xxl
charLen("Hello世界🎉")  // 8
charLen("中文测试")     // 4
charLen("hello")        // 5
```

### chars 索引

通过字符索引（而非字节索引）访问字符：

```xxl
var c = toChars("Hello世界🎉")

pln(c[0])   // "H"
pln(c[5])   // "世"
pln(c[7])   // "🎉"
pln(c[-1])  // "🎉"（负索引从末尾开始）
```

### chars 切片

使用字符索引提取子字符串：

```xxl
var c = toChars("Hello世界🎉")

pln(c.subStr(0, 5).toStr())   // "Hello"
pln(c.subStr(5, 7).toStr())   // "世界"
```

### chars 方法

#### toStr()

将 chars 转换回字符串。

```xxl
var c = toChars("你好")
pln(c.toStr())  // "你好"
```

#### upper()

返回大写版本（字符感知）。

```xxl
var c = toChars("Hello World 你好")
pln(c.upper().toStr())  // "HELLO WORLD 你好"
```

#### lower()

返回小写版本（字符感知）。

```xxl
var c = toChars("HELLO WORLD 你好")
pln(c.lower().toStr())  // "hello world 你好"
```

#### contains(substring)

检查 chars 是否包含子字符串（字符感知）。

```xxl
var c = toChars("Hello World 你好")
pln(c.contains("World"))  // true
pln(c.contains("你好"))    // true
pln(c.contains("xyz"))    // false
```

#### indexOf(substring)

返回第一次出现的字符索引，未找到返回 -1。

```xxl
var c = toChars("Hello World 你好")
pln(c.indexOf("World"))  // 6
pln(c.indexOf("你好"))    // 12
pln(c.indexOf("xyz"))    // -1
```

#### startsWith(prefix)

检查 chars 是否以指定前缀开头。

```xxl
var c = toChars("Hello World")
pln(c.startsWith("Hello"))  // true
pln(c.startsWith("World"))  // false
```

#### endsWith(suffix)

检查 chars 是否以指定后缀结尾。

```xxl
var c = toChars("Hello World 你好")
pln(c.endsWith("你好"))    // true
pln(c.endsWith("World"))  // false
```

#### reverse()

返回 chars 的反转副本。

```xxl
var c = toChars("abc世")
pln(c.reverse().toStr())  // "世cba"
```

#### repeat(n)

返回 chars 重复 n 次的结果。

```xxl
var c = toChars("abc世")
pln(c.repeat(3).toStr())  // "abc世abc世abc世"
```

### string 与 chars 对比

| 操作 | `string`（字节） | `chars`（字符） |
|-----|-----------------|----------------|
| `len("中文")` | 6（字节） | N/A |
| `len(toChars("中文"))` | N/A | 2（字符） |
| `"中文"[0]` | 字节值 | N/A |
| `toChars("中文")[0]` | N/A | "中"（字符） |
| `substr(s, 0, 1)` | 字节切片 | N/A |
| `c.subStr(0, 1)` | N/A | 字符切片 |

### 何时使用 chars

使用 `chars` 当你需要：
- 统计 Unicode 文本中的字符数
- 提取或操作单个 Unicode 字符
- 执行基于字符的切片
- 处理包含中文、日文、韩文、emoji 等的文本

使用 `string` 当你需要：
- 使用面向字节的 API
- 仅处理 ASCII 文本以优化性能
- 保持与 Go 字符串模型的兼容性

### 示例：处理多语言文本

```xxl
var text = "日本語English中文한국어"
var c = toChars(text)

pln("文本: ", text)
pln("字节数: ", len(text))       // 20 字节
pln("字符数: ", len(c))           // 13 字符

// 遍历每个字符
for (var i = 0; i < len(c); i = i + 1) {
    pln("  [", i, "] = ", c[i])
}
```

---

## stringbuilder

高效字符串拼接器，使用可变字符串构建器。与每次都创建新字符串的普通字符串拼接不同，StringBuilder 使用内部缓冲区以获得更好的性能。

#### create(capacity?)
创建新的 StringBuilder 实例。可选的 capacity 参数用于预分配缓冲区大小。

```xxl
import "stringbuilder"

// 创建空构建器
var sb = stringbuilder.create()

// 创建指定容量的构建器
var sb2 = stringbuilder.create(1000)
```

#### isStringBuilder(obj)
检查对象是否为 StringBuilder。

```xxl
var sb = stringbuilder.create()
stringbuilder.isStringBuilder(sb)   // true
stringbuilder.isStringBuilder(42)   // false
```

### StringBuilder 方法

#### write(str)
追加字符串到构建器。返回写入的字节数。

```xxl
var sb = stringbuilder.create()
sb.write("你好")
sb.write(" ")
sb.write("世界")
pln(sb.toString())  // "你好 世界"
```

#### writeLine(str)
追加字符串并换行。

```xxl
var sb = stringbuilder.create()
sb.writeLine("第一行")
sb.writeLine("第二行")
pln(sb.toString())
// 第一行
// 第二行
```

#### toString()
返回累积的字符串。

```xxl
var sb = stringbuilder.create()
sb.write("测试")
var result = sb.toString()  // "测试"
```

#### len()
返回当前累积字符串的长度。

```xxl
var sb = stringbuilder.create()
sb.write("你好")
pln(sb.len())  // 6 (UTF-8编码)
```

#### isEmpty()
检查构建器是否为空。

```xxl
var sb = stringbuilder.create()
pln(sb.isEmpty())  // true
sb.write("测试")
pln(sb.isEmpty())  // false
```

#### clear()
清空构建器中的所有内容。

```xxl
var sb = stringbuilder.create()
sb.write("你好")
sb.clear()
pln(sb.len())  // 0
```

#### reset()
`clear()` 的别名。重置构建器为空状态。

```xxl
var sb = stringbuilder.create()
sb.write("你好")
sb.reset()
pln(sb.isEmpty())  // true
```

#### grow(n)
预分配缓冲区容量，当已知最终大小时可提高性能。

```xxl
var sb = stringbuilder.create()
sb.grow(1000)  // 预分配 1000 字节
sb.write("你好")
```

### 性能示例

```xxl
import "stringbuilder"

// 高效字符串构建
var sb = stringbuilder.create()
for (var i = 0; i < 1000; i = i + 1) {
    sb.writeLine("第 " + str(i) + " 行")
}
var result = sb.toString()
pln("总字符数:", result.len())
```

---

## bytesbuffer

高效字节缓冲区操作，用于处理二进制数据。BytesBuffer 提供可变缓冲区用于读写字节，类似于 Go 的 `bytes.Buffer`。

**注意：** BytesBuffer 不是线程安全的。如需并发访问，请使用 `sync` 模块的 `Mutex` 或 `RWMutex` 进行外部同步。

### 模块函数

#### create(capacity?)
创建新的 BytesBuffer 实例。可选的 capacity 参数用于预分配缓冲区大小。

```xxl
import "bytesbuffer"

// 创建空缓冲区
var buf = bytesbuffer.create()

// 创建指定容量的缓冲区
var buf2 = bytesbuffer.create(1024)
```

#### fromBytes(arr)
从整数数组（0-255）创建 BytesBuffer。

```xxl
var buf = bytesbuffer.fromBytes([72, 101, 108, 108, 111])  // "Hello"
```

#### fromString(str)
从字符串创建 BytesBuffer。

```xxl
var buf = bytesbuffer.fromString("Hello World")
```

#### isBytesBuffer(obj)
检查对象是否为 BytesBuffer。

```xxl
var buf = bytesbuffer.create()
bytesbuffer.isBytesBuffer(buf)   // true
bytesbuffer.isBytesBuffer(42)    // false
```

### BytesBuffer 方法

#### write(data)
写入字符串或字节数组到缓冲区。返回写入的字节数。

```xxl
var buf = bytesbuffer.create()
buf.write("Hello")
buf.write([32, 87, 111, 114, 108, 100])  // " World"
pln(buf.toString())  // "Hello World"
```

#### writeByte(b)
写入单个字节（0-255）。

```xxl
var buf = bytesbuffer.create()
buf.writeByte(72)   // 'H'
buf.writeByte(105)  // 'i'
pln(buf.toString())  // "Hi"
```

#### writeInt16(n), writeInt32(n), writeInt64(n)
以小端序写入整数。

```xxl
var buf = bytesbuffer.create()
buf.writeInt32(123456)
buf.writeInt64(9876543210)
```

#### writeFloat32(n), writeFloat64(n)
以小端序写入浮点数。

```xxl
var buf = bytesbuffer.create()
buf.writeFloat32(3.14)
buf.writeFloat64(2.718281828)
```

#### bytes()
返回缓冲区内容为整数数组（0-255）。

```xxl
var buf = bytesbuffer.fromString("Hi")
var arr = buf.bytes()  // [72, 105]
```

#### toString()
返回缓冲区内容为字符串。

```xxl
var buf = bytesbuffer.create()
buf.write("Hello")
pln(buf.toString())  // "Hello"
```

#### len()
返回缓冲区当前长度。

```xxl
var buf = bytesbuffer.fromString("Hello")
pln(buf.len())  // 5
```

#### cap()
返回缓冲区当前容量。

```xxl
var buf = bytesbuffer.create(100)
pln(buf.cap())  // >= 100
```

#### readByte()
读取并返回单个字节，缓冲区为空时返回 null。

```xxl
var buf = bytesbuffer.fromString("Hi")
var b1 = buf.readByte()  // 72 ('H')
var b2 = buf.readByte()  // 105 ('i')
var b3 = buf.readByte()  // null (空)
```

#### readInt16(), readInt32(), readInt64()
以小端序读取整数。出错时返回 null。

```xxl
var buf = bytesbuffer.create()
buf.writeInt32(12345)
var n = buf.readInt32()  // 12345
```

#### readFloat32(), readFloat64()
以小端序读取浮点数。出错时返回 null。

```xxl
var buf = bytesbuffer.create()
buf.writeFloat64(3.14159)
var f = buf.readFloat64()  // 3.14159
```

#### peek(n)
返回接下来的 n 个字节，但不移动读取位置。

```xxl
var buf = bytesbuffer.fromString("Hello")
var preview = buf.peek(3)  // [72, 101, 108] ("Hel")
pln(buf.len())  // 5 (不变)
```

#### clear()
清空缓冲区中的所有内容。

```xxl
var buf = bytesbuffer.fromString("Hello")
buf.clear()
pln(buf.len())  // 0
```

#### reset()
`clear()` 的别名。重置缓冲区为空状态。

#### grow(n)
预分配缓冲区容量以提高性能。

```xxl
var buf = bytesbuffer.create()
buf.grow(1024)  // 预分配 1KB
```

#### truncate(n)
丢弃前 n 个字节之外的所有内容。

```xxl
var buf = bytesbuffer.fromString("Hello World")
buf.truncate(5)
pln(buf.toString())  // "Hello"
```

#### isEmpty()
检查缓冲区是否为空。

```xxl
var buf = bytesbuffer.create()
pln(buf.isEmpty())  // true
buf.write("test")
pln(buf.isEmpty())  // false
```

### 二进制协议示例

```xxl
import "bytesbuffer"

// 构建简单的二进制消息
var buf = bytesbuffer.create()

// 消息头：魔数（4字节）+ 版本（2字节）+ 长度（4字节）
buf.writeInt32(0x4D455353)  // "MESS" 魔数
buf.writeInt16(1)           // 版本 1
buf.writeInt32(12)          // 负载长度

// 负载数据
buf.write("Hello World!")

pln("总字节数:", buf.len())  // 22 字节
```

### 线程安全示例

```xxl
import "bytesbuffer"
load "sync"

var buf = bytesbuffer.create()
var mu = sync.createMutex()

// 安全的并发写入
spawn {
    mu.lock()
    buf.write("thread1")
    mu.unlock()
}

spawn {
    mu.lock()
    buf.write("thread2")
    mu.unlock()
}
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
pln(data["name"])  // "张三"

var arr = parse('[1, 2, 3]')
pln(arr[0])  // 1
```

#### stringify(value, indent)
将 Xxlang 值转换为 JSON 字符串。

```xxl
var obj = {"name": "张三", "age": 30}
pln(stringify(obj))        // {"name":"张三","age":30}
pln(stringify(obj, "  "))  // 带2空格缩进的格式化输出
pln(stringify(obj, 4))     // 带4空格缩进的格式化输出
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

#### hmacMd5(key, data)
返回 HMAC-MD5 哈希。

```xxl
hmacMd5("secret", "message")  // 十六进制字符串
```

#### hmacSha1(key, data)
返回 HMAC-SHA1 哈希。

```xxl
hmacSha1("secret", "message")  // 十六进制字符串
```

#### hmacSha256(key, data)
返回 HMAC-SHA256 哈希。

```xxl
hmacSha256("secret", "message")  // 十六进制字符串
```

#### hmacSha512(key, data)
返回 HMAC-SHA512 哈希。

```xxl
hmacSha512("secret", "message")  // 十六进制字符串
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

#### randomBase64(n)
返回 n 个随机字节的 base64 字符串。

```xxl
randomBase64(16)  // 随机 base64 编码字符串
```

#### uuid()
生成随机 UUID (v4) 字符串。

```xxl
uuid()  // "550e8400-e29b-41d4-a716-446655440000"
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
pln(t["year"])     // 2026
pln(t["month"])    // 3 (三月)
pln(t["day"])      // 14
pln(t["hour"])     // 20
pln(t["minute"])   // 30
pln(t["second"])   // 45
pln(t["nanosecond"])  // 123456789
```

#### year(), month(), day(), hour(), minute(), second()
返回当前日期/时间组件。

```xxl
pln(time.year())    // 2026
pln(time.month())   // 3
pln(time.day())     // 14
pln(time.hour())    // 20
pln(time.minute())  // 30
pln(time.second())  // 45
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
pln(time.format("2006-01-02"))           // "2026-03-14"
pln(time.format("15:04:05"))             // "20:30:45"
pln(time.format("2006-01-02 15:04:05"))  // "2026-03-14 20:30:45"
```

#### formatUnix(timestamp, layout)
格式化 Unix 时间戳。

```xxl
var ts = time.unix()
pln(time.formatUnix(ts, "2006-01-02"))
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
pln("开始...")
time.sleep(1000)  // 休眠1秒
pln("完成！")
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
pln("耗时 " + elapsed.toStr() + " 毫秒")
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
