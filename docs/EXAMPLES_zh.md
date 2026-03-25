# Xxlang 常用场景代码示例

本文档提供由浅入深的实用代码示例，涵盖常见编程场景。

## 目录

1. [基础语法](#基础语法)
2. [控制流程](#控制流程)
3. [函数](#函数)
4. [字符串操作](#字符串操作)
5. [数组操作](#数组操作)
6. [Map操作](#map操作)
7. [文件操作](#文件操作)
8. [JSON处理](#json处理)
9. [错误处理](#错误处理)
10. [类与面向对象](#类与面向对象)
11. [正则表达式](#正则表达式)
12. [HTTP请求](#http请求)
13. [StringBuilder](#stringbuilder)
14. [实用案例](#实用案例)
15. [并发编程](#并发编程)
16. [大整数与大浮点数](#大整数与大浮点数)
17. [Chars类型（Unicode处理）](#chars类型unicode处理)
18. [技巧与最佳实践](#技巧与最佳实践)
19. [常见问题解决](#常见问题解决)

---

## 基础语法

### 变量与类型

```xxl
// 变量声明
var name = "Alice"
var age = 30
var score = 95.5
var active = true
var nothing = null

// 数组
var numbers = [1, 2, 3, 4, 5]
var names = ["Alice", "Bob", "Charlie"]

// Map（字典）
var person = {
    "name": "Alice",
    "age": 30,
    "city": "New York"
}

// 访问Map值
pln(person["name"])  // "Alice"
```

#### 大整数与大浮点数

当需要处理超出普通整数范围的数值或需要高精度计算时，可使用BigInt和BigFloat：

```xxl
// BigInt - 任意精度整数（后缀 n）
var big = 123456789012345678901234567890n
pln(big)                    // 123456789012345678901234567890
pln(typeOf(big))            // "BIGINT"

// BigFloat - 任意精度浮点数（后缀 m）
var pi = 3.141592653589793238462643383279m
pln(pi)                     // 3.141592653589793238462643383279
pln(typeOf(pi))             // "BIGFLOAT"

// 精确计算（避免浮点数误差）
var a = 0.1m
var b = 0.2m
pln(a + b)                  // 0.3（精确值，而非0.30000000000000004）

// 大数运算
var factorial = 1n
for (var i = 1; i <= 50; i++) {
    factorial = factorial * toBigInt(i)
}
pln(factorial)              // 50的阶乘
```

**适用场景：**
- **BigInt**: 加密算法、阶乘计算、超过int64范围的ID
- **BigFloat**: 金融计算、科学计算、避免浮点误差

### 类型检查

```xxl
// 检查类型
var x = 42
pln(isInt(x))      // true
pln(isFloat(3.14)) // true
pln(isString("hi")) // true
pln(isArray([1,2,3])) // true
pln(isMap({"a": 1})) // true
pln(isBool(true))  // true
pln(isNull(null))  // true
```

### 算术运算

```xxl
var a = 10
var b = 3

pln(a + b)   // 13
pln(a - b)   // 7
pln(a * b)   // 30
pln(a / b)   // 3
pln(a % b)   // 1

// 自增
var count = 0
count = count + 1
pln(count)  // 1
```

### 字符串基础

```xxl
var s = "Hello, World!"

// 字符串拼接
var greeting = "Hello" + " " + "World"
pln(greeting)  // "Hello World"

// 字符串方法
pln(upper(s))      // "HELLO, WORLD!"
pln(lower(s))      // "hello, world!"
pln(len(s))        // 13
pln(trim("  hi  ")) // "hi"
```

---

## 控制流程

### 条件判断

```xxl
var score = 85

if (score >= 90) {
    pln("等级: A")
} else if (score >= 80) {
    pln("等级: B")
} else if (score >= 70) {
    pln("等级: C")
} else {
    pln("等级: F")
}
```

### For循环

```xxl
// 传统for循环
for (var i = 0; i < 5; i = i + 1) {
    pln("迭代:", i)
}

// 遍历数组
var fruits = ["apple", "banana", "cherry"]
for (fruit in fruits) {
    pln("水果:", fruit)
}

// 遍历Map
var scores = {"Alice": 95, "Bob": 87, "Charlie": 92}
for (name in scores) {
    pln(name, "得分", scores[name])
}
```

### While循环

```xxl
var count = 5
while (count > 0) {
    pln("倒计时:", count)
    count = count - 1
}
pln("发射！")
```

### Switch语句

Switch语句提供了一种清晰的方式来根据值进行分支。**注意：Xxlang的switch不会穿透(fall-through)，匹配到case后会自动退出。**

```xxl
var day = 3

switch (day) {
case 1:
    pln("星期一")
case 2:
    pln("星期二")
case 3:
    pln("星期三")    // 输出这个，然后退出switch
case 4:
    pln("星期四")    // 不会执行
case 5:
    pln("星期五")    // 不会执行
default:
    pln("周末")      // 不会执行
}
```

#### 字符串匹配

```xxl
var fruit = "apple"

switch (fruit) {
case "apple":
    pln("苹果")
case "banana":
    pln("香蕉")
case "orange":
    pln("橙子")
default:
    pln("未知水果")
}
// 输出: 苹果
```

#### 使用表达式

```xxl
var x = 10
var y = 20

switch (x + y) {
case 10:
    pln("和为10")
case 20:
    pln("和为20")
case 30:
    pln("和为30")
default:
    pln("其他值")
}
// 输出: 和为30
```

#### 条件Switch模式

使用`switch (true)`可以优雅地处理范围判断和复杂条件：

```xxl
// 范围判断
var score = 85

switch (true) {
case score >= 90:
    pln("等级: A")
case score >= 80:
    pln("等级: B")
case score >= 70:
    pln("等级: C")
case score >= 60:
    pln("等级: D")
default:
    pln("等级: F")
}
// 输出: 等级: B

// 复杂条件判断
var age = 25

switch (true) {
case age < 13:
    pln("儿童")
case age < 20:
    pln("青少年")
case age < 60:
    pln("成年人")
default:
    pln("老年人")
}
// 输出: 成年人

// 多条件组合
var x = 10
var y = 20

switch (true) {
case x < 0 || y < 0:
    pln("存在负数")
case x > 100 && y > 100:
    pln("都是大数")
case x == y:
    pln("相等")
default:
    pln("其他情况")
}
```

> **注意**: case按顺序评估，将更具体的条件放在前面。

#### 与if-else链的比较

```xxl
// 使用switch更清晰
switch (status) {
case 200:
    pln("成功")
case 404:
    pln("未找到")
case 500:
    pln("服务器错误")
default:
    pln("未知状态")
}

// 等价的if-else链更冗长
if (status == 200) {
    pln("成功")
} else if (status == 404) {
    pln("未找到")
} else if (status == 500) {
    pln("服务器错误")
} else {
    pln("未知状态")
}
```

> 详细说明请参阅 [LANGUAGE.md](LANGUAGE.md#switch-statement) 中的 Switch Statement 章节。

### 三元运算符

三元运算符`?:`提供了一种简洁的条件表达式写法：

```xxl
// 基本语法: 条件 ? 真值 : 假值
var age = 20
var status = age >= 18 ? "成年" : "未成年"
pln(status)  // "成年"

// 等价于:
var status2
if (age >= 18) {
    status2 = "成年"
} else {
    status2 = "未成年"
}
```

#### 常用示例

```xxl
// 简单值选择
var a = 10
var b = 20
var max = a > b ? a : b
pln(max)  // 20

// 字符串格式化
var count = 1
var msg = count == 1 ? "1个项目" : count + "个项目"
pln(msg)  // "1个项目"

// 默认值
var name = ""
var displayName = name != "" ? name : "匿名用户"
pln(displayName)  // "匿名用户"

// 在表达式中使用
var found = true
pln("结果: " + (found ? "找到" : "未找到"))

// 与函数调用结合
var isValid = true
var result = isValid ? processData() : getDefault()
```

#### 最佳实践

```xxl
// ✅ 好：简单清晰的选择
var discount = isMember ? 0.1 : 0

// ❌ 不好：过于复杂
var x = a > b ? c > d ? e : f : g > h ? i : j  // 难以阅读

// ✅ 更好：复杂逻辑使用switch或if-else
var x
switch (true) {
case a > b && c > d:
    x = e
case a > b:
    x = f
case g > h:
    x = i
default:
    x = j
}
```

---

## 函数

### 基础函数

```xxl
func greet(name) {
    return "你好, " + name + "!"
}

pln(greet("Alice"))  // "你好, Alice!"
```

### 多参数函数

```xxl
func add(a, b) {
    return a + b
}

func calculate(a, b, op) {
    if (op == "+") {
        return a + b
    } else if (op == "-") {
        return a - b
    } else if (op == "*") {
        return a * b
    } else if (op == "/") {
        return a / b
    }
    return 0
}

pln(calculate(10, 5, "+"))  // 15
pln(calculate(10, 5, "*"))  // 50
```

### 可变参数函数

使用`...`语法定义可变参数函数，可以接受任意数量的参数：

```xxl
// 基本可变参数函数
func sum(...args) {
    var total = 0
    for (x in args) {
        total = total + x
    }
    return total
}

pln(sum())            // 0
pln(sum(1))           // 1
pln(sum(1, 2, 3))     // 6
pln(sum(1, 2, 3, 4, 5))  // 15
```

#### 固定参数与可变参数混合

固定参数必须放在可变参数之前：

```xxl
// 前缀 + 可变参数
func formatList(prefix, ...items) {
    var result = prefix + ": "
    for (i, item in items) {
        if (i > 0) {
            result = result + ", "
        }
        result = result + toStr(item)
    }
    return result
}

pln(formatList("水果", "苹果", "香蕉", "橙子"))
// 输出: 水果: 苹果, 香蕉, 橙子
```

#### 可变参数是数组

在函数内部，可变参数作为数组处理：

```xxl
func logAll(...messages) {
    // messages 是数组
    pln("参数数量:", len(messages))
    for (msg in messages) {
        pln("-", msg)
    }
}

logAll("你好", "世界", 42, true)
// 参数数量: 4
// - 你好
// - 世界
// - 42
// - true
```

#### 实用示例

```xxl
// 最大值函数
func max(...nums) {
    if (len(nums) == 0) {
        return null
    }
    var result = nums[0]
    for (num in nums) {
        if (num > result) {
            result = num
        }
    }
    return result
}

pln(max(3, 1, 4, 1, 5, 9, 2, 6))  // 9

// 字符串拼接
func concat(...parts) {
    var result = ""
    for (part in parts) {
        result = result + toStr(part)
    }
    return result
}

pln(concat("Hello", " ", "World", "!"))  // "Hello World!"

// 创建数组
func list(...items) {
    return items
}

var arr = list(1, 2, 3, 4, 5)
pln(arr)  // [1, 2, 3, 4, 5]
```

### 闭包

```xxl
// 返回函数的函数
func makeCounter() {
    var count = 0
    func() {
        count = count + 1
        return count
    }
}

var counter = makeCounter()
pln(counter())  // 1
pln(counter())  // 2
pln(counter())  // 3

// 另一个计数器是独立的
var counter2 = makeCounter()
pln(counter2())  // 1
```

### 高阶函数

```xxl
// 接受函数作为参数的函数
func apply(arr, fn) {
    var result = []
    for (item in arr) {
        result = push(result, fn(item))
    }
    return result
}

var numbers = [1, 2, 3, 4, 5]

var doubled = apply(numbers, func(x) { return x * 2 })
pln(doubled)  // [2, 4, 6, 8, 10]

var squared = apply(numbers, func(x) { return x * x })
pln(squared)  // [1, 4, 9, 16, 25]
```

---

## 字符串操作

### 字符串处理

```xxl
import "string"

var text = "  Hello, World!  "

// 去除空白
pln(trim(text))                    // "Hello, World!"
pln(trimLeft(text))                // "Hello, World!  "
pln(trimRight(text))               // "  Hello, World!"

// 分割与连接
var parts = split("a,b,c", ",")
pln(parts)                         // ["a", "b", "c"]

pln(join(["x", "y", "z"], "-"))    // "x-y-z"

// 包含与查找
pln(contains("hello", "ell"))      // true
pln(indexOf("hello", "l"))         // 2
pln(lastIndexOf("hello", "l"))     // 3

// 前缀与后缀
pln(startsWith("hello", "he"))     // true
pln(endsWith("hello", "lo"))       // true

// 大小写转换
pln(upper("hello"))                // "HELLO"
pln(lower("HELLO"))                // "hello"

// 重复
pln(repeat("ab", 3))               // "ababab"
```

### 字符串格式化

```xxl
import "fmt"

// Printf风格格式化
pln(fmt.sprintf("姓名: %s, 年龄: %d", "Alice", 30))
pln(fmt.sprintf("价格: $%.2f", 19.99))
pln(fmt.sprintf("十六进制: 0x%X", 255))

// 填充对齐
pln(lpad("42", 5, "0"))   // "00042"
pln(rpad("42", 5, "-"))   // "42---"
```

---

## 数组操作

### 数组基础

```xxl
var arr = [3, 1, 4, 1, 5, 9, 2, 6]

// 长度
pln(len(arr))  // 8

// 访问元素
pln(arr[0])    // 3
pln(arr[-1])   // 6 (最后一个元素)

// 添加元素
arr = push(arr, 5)
pln(arr)       // [..., 5]

// 切片
var sub = arr[2:5]
pln(sub)       // [4, 1, 5]

// 反转
pln(reverse(arr))

// 排序
pln(sort(arr))  // [1, 1, 2, 3, 4, 5, 6, 9]
```

### 数组工具函数

```xxl
import "array"

var numbers = [1, 2, 3, 4, 5]

// 函数式操作
var doubled = array.map(numbers, func(x) { return x * 2 })
pln(doubled)  // [2, 4, 6, 8, 10]

var evens = array.filter(numbers, func(x) { return x % 2 == 0 })
pln(evens)    // [2, 4]

var sum = array.reduce(numbers, 0, func(acc, x) { return acc + x })
pln(sum)      // 15

// 查找
pln(array.includes(numbers, 3))  // true
pln(array.indexOf(numbers, 4))   // 3

// 工具函数
pln(array.first(numbers))   // 1
pln(array.last(numbers))    // 5
pln(array.min(numbers))     // 1
pln(array.max(numbers))     // 5
```

---

## Map操作

### Map基础

```xxl
var person = {
    "name": "Alice",
    "age": 30,
    "city": "New York"
}

// 访问
pln(person["name"])  // "Alice"

// 修改
person["age"] = 31
person["email"] = "alice@example.com"

// 删除
person = delete(person, "city")

// 检查键是否存在
pln(hasKey(person, "name"))   // true
pln(hasKey(person, "phone"))  // false

// 获取所有键和值
pln(keys(person))    // ["name", "age", "email"]
pln(values(person))  // ["Alice", 31, "alice@example.com"]

// 遍历
for (key in person) {
    pln(key, ":", person[key])
}
```

### 嵌套Map

```xxl
var company = {
    "name": "Tech Corp",
    "employees": [
        {"name": "Alice", "role": "Engineer"},
        {"name": "Bob", "role": "Designer"},
        {"name": "Charlie", "role": "Manager"}
    ],
    "address": {
        "city": "San Francisco",
        "country": "USA"
    }
}

pln(company["name"])
pln(company["address"]["city"])
pln(company["employees"][0]["name"])

// 遍历员工
for (emp in company["employees"]) {
    pln(emp["name"], "-", emp["role"])
}
```

---

## 文件操作

### 读写文件

```xxl
import "io"

// 写入文件
io.writeFile("test.txt", "Hello, File!")

// 追加内容
io.appendFile("test.txt", "\n新的一行")

// 读取整个文件
var content = io.readFile("test.txt")
pln(content)

// 检查文件是否存在
if (io.exists("test.txt")) {
    pln("文件存在！")
}

// 删除文件
io.remove("test.txt")

// 创建目录
io.mkdir("mydir")

// 列出目录内容
var files = io.readDir(".")
for (f in files) {
    pln(f)
}
```

### 路径操作

```xxl
import "os"

// 获取当前目录
pln(os.cwd())

// 环境变量
var home = os.getenv("HOME")
pln("主目录:", home)

// 命令行参数
var args = os.args()
for (arg in args) {
    pln("参数:", arg)
}
```

---

## JSON处理

### 解析与生成JSON

```xxl
import "json"

// 解析JSON字符串
var jsonString = '{"name": "Alice", "age": 30, "skills": ["Go", "Python"]}'
var data = json.parse(jsonString)

pln(data["name"])              // "Alice"
pln(data["skills"][0])         // "Go"

// 生成JSON字符串
var person = {
    "name": "Bob",
    "age": 25,
    "active": true,
    "tags": ["developer", "golang"]
}

var jsonOutput = json.stringify(person)
pln(jsonOutput)

// 美化输出
var pretty = json.stringifyIndent(person, "  ")
pln(pretty)
```

### JSON数组处理

```xxl
import "json"

// 解析JSON数组
var jsonArray = '[1, 2, 3, 4, 5]'
var numbers = json.parse(jsonArray)
pln(numbers[2])  // 3

// 复杂JSON
var complex = '''
{
    "users": [
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"}
    ],
    "total": 2
}
'''

var result = json.parse(complex)
for (user in result["users"]) {
    pln("用户", user["id"], ":", user["name"])
}
```

---

## 错误处理

### 错误检查模式

```xxl
import "io"

func safeReadFile(path) {
    if (!io.exists(path)) {
        return {"ok": false, "error": "文件不存在"}
    }

    var content = io.readFile(path)
    if (isError(content)) {
        return {"ok": false, "error": content}
    }

    return {"ok": true, "data": content}
}

var result = safeReadFile("config.json")
if (result["ok"]) {
    pln("内容:", result["data"])
} else {
    pln("错误:", result["error"])
}
```

### 错误处理模式

```xxl
// 使用isError检查错误
func divide(a, b) {
    if (b == 0) {
        return error("除零错误")
    }
    return a / b
}

var result = divide(10, 0)
if (isError(result)) {
    pln("错误:", result)
} else {
    pln("结果:", result)
}
```

---

## 类与面向对象

### 基础类

```xxl
class Point {
    func init(x, y) {
        this.x = x
        this.y = y
    }

    func toString() {
        return "(" + str(this.x) + ", " + str(this.y) + ")"
    }

    func add(other) {
        return new Point(this.x + other.x, this.y + other.y)
    }

    func distance() {
        import "math"
        return math.sqrt(this.x * this.x + this.y * this.y)
    }
}

var p1 = new Point(3, 4)
var p2 = new Point(1, 2)

pln(p1.toString())      // "(3, 4)"
pln(p1.distance())      // 5.0

var p3 = p1.add(p2)
pln(p3.toString())      // "(4, 6)"
```

### 继承

```xxl
class Animal {
    func init(name) {
        this.name = name
    }

    func speak() {
        return "..."
    }
}

class Dog extends Animal {
    func init(name, breed) {
        this.name = name
        this.breed = breed
    }

    func speak() {
        return this.name + "说汪汪！"
    }
}

class Cat extends Animal {
    func speak() {
        return this.name + "说喵喵！"
    }
}

var dog = new Dog("Buddy", "Golden Retriever")
var cat = new Cat("Whiskers")

pln(dog.speak())  // "Buddy说汪汪！"
pln(cat.speak())  // "Whiskers说喵喵！"
```

### 银行账户示例

```xxl
class BankAccount {
    func init(owner, initialBalance) {
        this.owner = owner
        this.balance = initialBalance
        this.transactions = []
    }

    func deposit(amount) {
        if (amount <= 0) {
            return error("无效的存款金额")
        }
        this.balance = this.balance + amount
        this.transactions = push(this.transactions, "存款: +" + str(amount))
        return this.balance
    }

    func withdraw(amount) {
        if (amount <= 0) {
            return error("无效的取款金额")
        }
        if (amount > this.balance) {
            return error("余额不足")
        }
        this.balance = this.balance - amount
        this.transactions = push(this.transactions, "取款: -" + str(amount))
        return this.balance
    }

    func getBalance() {
        return this.balance
    }

    func getStatement() {
        var result = "账户: " + this.owner + "\n"
        result = result + "余额: $" + str(this.balance) + "\n"
        result = result + "交易记录:\n"
        for (t in this.transactions) {
            result = result + "  " + t + "\n"
        }
        return result
    }
}

var account = new BankAccount("Alice", 1000)
account.deposit(500)
account.withdraw(200)
pln(account.getStatement())
```

---

## 正则表达式

### 模式匹配

```xxl
import "regex"

// 检查是否匹配
pln(regex.matches("hello123", "\\d+"))    // true
pln(regex.matches("hello", "\\d+"))       // false

// 查找匹配
var text = "The quick brown fox jumps over 123 lazy dogs."
var numbers = regex.findall(text, "\\d+")
pln(numbers)  // ["123"]

// 查找所有单词
var words = regex.findall(text, "[a-z]+")
pln(words)    // ["The", "quick", "brown", ...]

// 替换
var replaced = regex.replace("hello world", "world", "Xxlang")
pln(replaced)  // "hello Xxlang"

// 替换所有
var cleaned = regex.replaceAll("123-456-7890", "\\D", "")
pln(cleaned)   // "1234567890"
```

### 邮箱验证

```xxl
import "regex"

func isValidEmail(email) {
    var pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    return regex.matches(email, pattern)
}

pln(isValidEmail("user@example.com"))     // true
pln(isValidEmail("invalid-email"))        // false
```

### 电话号码解析

```xxl
import "regex"

func parsePhone(phone) {
    // 移除非数字字符
    var digits = regex.replaceAll(phone, "\\D", "")

    if (len(digits) != 11) {
        return error("无效的电话号码")
    }

    return {
        "areaCode": digits[:3],
        "prefix": digits[3:7],
        "number": digits[7:]
    }
}

var parsed = parsePhone("138-1234-5678")
pln(parsed)
// {"areaCode": "138", "prefix": "1234", "number": "5678"}
```

---

## HTTP请求

### GET请求

```xxl
import "net"

// 简单GET请求
var response = net.get("https://httpbin.org/get")
pln("状态:", response["status"])
pln("内容:", response["body"])

// 带请求头的GET
var headers = {
    "User-Agent": "Xxlang/1.0",
    "Accept": "application/json"
}
response = net.get("https://httpbin.org/headers", headers)
pln(response["body"])
```

### POST请求

```xxl
import "net"
import "json"

// POST JSON数据
var data = {
    "name": "Alice",
    "email": "alice@example.com"
}

var response = net.post(
    "https://httpbin.org/post",
    json.stringify(data),
    {"Content-Type": "application/json"}
)

pln("状态:", response["status"])
```

### API客户端示例

```xxl
import "net"
import "json"

class APIClient {
    func init(baseUrl) {
        this.baseUrl = baseUrl
        this.headers = {"Content-Type": "application/json"}
    }

    func setHeader(key, value) {
        this.headers[key] = value
    }

    func get(endpoint) {
        var url = this.baseUrl + endpoint
        var response = net.get(url, this.headers)
        if (response["status"] != 200) {
            return error("HTTP " + str(response["status"]))
        }
        return json.parse(response["body"])
    }

    func post(endpoint, data) {
        var url = this.baseUrl + endpoint
        var response = net.post(url, json.stringify(data), this.headers)
        if (response["status"] != 200 && response["status"] != 201) {
            return error("HTTP " + str(response["status"]))
        }
        return json.parse(response["body"])
    }
}

// 使用示例
var api = new APIClient("https://jsonplaceholder.typicode.com")
var posts = api.get("/posts")
pln("第一篇文章标题:", posts[0]["title"])
```

---

## StringBuilder

### 基础用法

```xxl
import "stringbuilder"

// 创建StringBuilder
var sb = stringbuilder.create()

// 高效构建字符串
sb.write("你好")
sb.write(" ")
sb.writeLine("世界！")
sb.write("答案是: ")
sb.write(str(42))

pln(sb.toString())
// 你好 世界！
// 答案是: 42

pln("长度:", sb.len())
pln("是否为空:", sb.isEmpty())

// 清空并复用
sb.clear()
pln("清空后是否为空:", sb.isEmpty())
```

### 构建报告

```xxl
import "stringbuilder"

func buildReport(title, items) {
    var sb = stringbuilder.create()

    // 标题
    sb.writeLine("=" + repeat("=", len(title)))
    sb.writeLine(title)
    sb.writeLine("=" + repeat("=", len(title)))
    sb.writeLine("")

    // 列表项
    var i = 1
    for (item in items) {
        sb.write(str(i) + ". ")
        sb.writeLine(item)
        i = i + 1
    }

    // 页脚
    sb.writeLine("")
    sb.writeLine("总计: " + str(len(items)) + " 项")

    return sb.toString()
}

var report = buildReport("购物清单", [
    "牛奶",
    "鸡蛋",
    "面包",
    "黄油"
])
pln(report)
```

### CSV构建器

```xxl
import "stringbuilder"

func buildCSV(headers, rows) {
    var sb = stringbuilder.create()

    // 表头
    var first = true
    for (h in headers) {
        if (!first) {
            sb.write(",")
        }
        sb.write(h)
        first = false
    }
    sb.writeLine("")

    // 数据行
    for (row in rows) {
        first = true
        for (val in row) {
            if (!first) {
                sb.write(",")
            }
            // 如果包含逗号则加引号
            if (contains(val, ",")) {
                sb.write("\"" + val + "\"")
            } else {
                sb.write(val)
            }
            first = false
        }
        sb.writeLine("")
    }

    return sb.toString()
}

var csv = buildCSV(
    ["姓名", "邮箱", "城市"],
    [
        ["Alice", "alice@example.com", "北京"],
        ["Bob", "bob@example.com", "上海"],
        ["Charlie", "charlie@example.com", "广州"]
    ]
)
pln(csv)
```

---

## 实用案例

### 配置文件解析器

```xxl
import "io"
import "string"

func parseConfig(path) {
    if (!io.exists(path)) {
        return error("配置文件不存在: " + path)
    }

    var content = io.readFile(path)
    var lines = split(content, "\n")

    var config = {}

    for (line in lines) {
        line = trim(line)

        // 跳过空行和注释
        if (len(line) == 0 || startsWith(line, "#")) {
            continue
        }

        // 解析 key=value
        var parts = split(line, "=")
        if (len(parts) == 2) {
            var key = trim(parts[0])
            var value = trim(parts[1])

            // 移除引号
            if (startsWith(value, "\"") && endsWith(value, "\"")) {
                value = value[1:len(value)-1]
            }

            config[key] = value
        }
    }

    return config
}

// config.txt:
// # 数据库设置
// host = localhost
// port = 5432
// name = mydb

var config = parseConfig("config.txt")
pln("主机:", config["host"])
pln("端口:", config["port"])
pln("数据库:", config["name"])
```

### 词频统计

```xxl
import "string"
import "regex"

func wordFrequency(text) {
    // 转换为小写并提取单词
    text = lower(text)
    var words = regex.findall(text, "[a-z]+")

    // 统计频率
    var freq = {}
    for (word in words) {
        if (hasKey(freq, word)) {
            freq[word] = freq[word] + 1
        } else {
            freq[word] = 1
        }
    }

    return freq
}

var text = "The quick brown fox jumps over the lazy dog. The dog was not impressed."
var freq = wordFrequency(text)

for (word in freq) {
    pln(word, ":", freq[word])
}
```

### 待办事项应用

```xxl
import "io"
import "json"

class TodoList {
    func init(filename) {
        this.filename = filename
        this.todos = []
        this.load()
    }

    func load() {
        if (io.exists(this.filename)) {
            var content = io.readFile(this.filename)
            var data = json.parse(content)
            if (!isError(data)) {
                this.todos = data
            }
        }
    }

    func save() {
        io.writeFile(this.filename, json.stringify(this.todos))
    }

    func add(task) {
        var todo = {
            "id": len(this.todos) + 1,
            "task": task,
            "done": false,
            "created": now()
        }
        this.todos = push(this.todos, todo)
        this.save()
        return todo
    }

    func complete(id) {
        var i = 0
        for (todo in this.todos) {
            if (todo["id"] == id) {
                this.todos[i]["done"] = true
                this.save()
                return true
            }
            i = i + 1
        }
        return false
    }

    func list() {
        pln("=== 待办事项 ===")
        for (todo in this.todos) {
            var status = "[ ]"
            if (todo["done"]) {
                status = "[x]"
            }
            pln(status, todo["id"], "-", todo["task"])
        }
        pln("================")
    }
}

// 使用示例
var todos = new TodoList("todos.json")
todos.add("学习 Xxlang")
todos.add("构建项目")
todos.complete(1)
todos.list()
```

### 数据转换管道

```xxl
import "array"

// 通过管道操作转换数据
func pipeline(data, operations) {
    var result = data
    for (op in operations) {
        result = op(result)
    }
    return result
}

// 示例数据
var sales = [
    {"product": "Widget", "price": 10.0, "quantity": 5},
    {"product": "Gadget", "price": 25.0, "quantity": 3},
    {"product": "Widget", "price": 10.0, "quantity": 2},
    {"product": "Gizmo", "price": 15.0, "quantity": 10},
    {"product": "Gadget", "price": 25.0, "quantity": 1}
]

// 计算每笔销售的总金额
var withTotal = array.map(sales, func(sale) {
    sale["total"] = sale["price"] * sale["quantity"]
    return sale
})

// 筛选大额销售
var bigSales = array.filter(withTotal, func(sale) {
    return sale["total"] > 50
})

// 按总金额降序排列
var sorted = array.sortBy(bigSales, func(a, b) {
    return b["total"] - a["total"]
})

pln("大额销售 (>$50):")
for (sale in sorted) {
    pln("  " + sale["product"] + ": $" + str(sale["total"]))
}
```

---

## 并发编程

### Tube（通道）

Tube 是 Xxlang 中的并发通信机制，类似于 Go 语言的 channel：

```xxl
// 创建一个带缓冲的 tube
var tube = makeTube(2)

// 发送值
tube <- 10
tube <- 20

// 接收值
pln("接收:", <- tube)
pln("接收:", <- tube)

// 检查容量和长度
pln("Tube容量:", tubeCap(tube))
pln("Tube长度:", tubeLen(tube))
```

### Select 语句

Select 用于处理多个 tube 的并发操作：

```xxl
var ch1 = makeTube(1)
var ch2 = makeTube(1)

ch1 <- "来自ch1"
ch2 <- "来自ch2"

// Select 从最先就绪的 tube 接收
for (var i = 0; i < 2; i++) {
    select {
        case var v = <- ch1:
            pln("从ch1收到:", v)
        case var v = <- ch2:
            pln("从ch2收到:", v)
    }
}
```

### Context 超时控制

Context 用于超时和取消控制：

```xxl
// 创建带超时的 context
var ctx = contextWithTimeout(100)  // 100毫秒超时

pln("初始状态:", contextDone(ctx))

sleep(50)
pln("50ms后状态:", contextDone(ctx))

sleep(100)
pln("150ms后状态:", contextDone(ctx))  // true，已超时
```

### 并发执行

使用 `run` 关键字启动并发任务：

```xxl
var results = makeTube(3)

// 并发运行多个任务
run func() {
    sleep(30)
    results <- "任务1完成"
}

run func() {
    sleep(20)
    results <- "任务2完成"
}

run func() {
    sleep(10)
    results <- "任务3完成"
}

// 收集结果（可能以任意顺序到达）
for (var i = 0; i < 3; i++) {
    pln(<- results)
}
```

### 并发安全计数器

```xxl
func createSafeCounter() {
    var count = 0
    var mutex = makeMutex()

    return {
        "increment": func() {
            mutexLock(mutex)
            count = count + 1
            mutexUnlock(mutex)
            return count
        },
        "decrement": func() {
            mutexLock(mutex)
            count = count - 1
            mutexUnlock(mutex)
            return count
        },
        "get": func() {
            mutexLock(mutex)
            var val = count
            mutexUnlock(mutex)
            return val
        }
    }
}

var counter = createSafeCounter()
pln(counter["increment"]())  // 1
pln(counter["increment"]())  // 2
pln(counter["get"]())        // 2
```

---

## 大整数与大浮点数

### BigInt 任意精度整数

当需要处理超出普通整数范围的数值时，使用 BigInt（后缀 `n`）：

```xxl
// 普通整数有限制（64位）
var maxInt64 = 9223372036854775807

// BigInt 没有实际限制
var huge = 1234567890123456789012345678901234567890n
pln("巨大数字:", huge)

// BigInt 运算
var a = 1000000000000000000n
var b = 2000000000000000000n
pln("a + b =", a + b)
pln("a * b =", a * b)

// 大数阶乘
func factorialBig(n) {
    var result = 1n
    for (var i = 2; i <= n; i++) {
        result = result * toBigInt(i)
    }
    return result
}

pln("100! =", factorialBig(100))
```

### BigFloat 任意精度浮点数

BigFloat 解决浮点数精度问题（后缀 `m`）：

```xxl
// 普通浮点数有精度问题
var regular = 0.1 + 0.2
pln("普通 0.1 + 0.2:", regular)  // 可能有精度误差

// BigFloat 保持精确
var precise = 0.1m + 0.2m
pln("BigFloat 0.1m + 0.2m:", precise)  // 精确的 0.3

// 高精度计算
var pi = 3.14159265358979323846264338327950288419716939937510m
pln("高精度圆周率:", pi)

// 金融计算示例
var price = 19.99m
var tax = price * 0.08m
var total = price + tax
pln("价格:", price)
pln("税费 (8%):", tax)
pln("总计:", total)
```

**适用场景：**
- **BigInt**: 加密算法、阶乘计算、超过 int64 范围的 ID
- **BigFloat**: 金融计算、科学计算、避免浮点误差

---

## Chars 类型（Unicode 处理）

### 字符级操作

String 是字节导向的，而 Chars 是字符导向的：

```xxl
// String 按字节计算
var s = "中文测试"
pln("字符串长度（字节）:", len(s))  // 12 字节

// Chars 按字符计算
var c = toChars(s)
pln("Chars长度（字符）:", len(c))  // 4 个字符

// 字符访问
pln("c[0]:", c[0])  // 中
pln("c[1]:", c[1])  // 文

// 混合内容和 emoji
var mixed = toChars("Hello世界🎉")
pln("长度:", len(mixed))
for (ch in mixed) {
    pln(" -", ch)
}
```

### Chars 方法

```xxl
var text = toChars("Hello World 世界")

// 按字符位置截取
pln("subStr(0, 5):", text.subStr(0, 5).toStr())  // Hello

// 大小写转换（支持 Unicode）
pln("upper():", text.upper().toStr())
pln("lower():", text.lower().toStr())

// 反转
pln("reverse():", text.reverse().toStr())

// 包含检测
pln("contains('世界'):", text.contains("世界"))
```

---

## 技巧与最佳实践

### 代码组织

```xxl
// 使用模块组织代码
// utils.xxl
func helper1() { ... }
func helper2() { ... }

// main.xxl
import "utils"
utils.helper1()
```

### 错误处理模式

```xxl
// 始终检查错误
func safeOperation() {
    var result = riskyOperation()
    if (isError(result)) {
        pln("错误:", result)
        return null
    }
    return result
}
```

### 性能优化

```xxl
// 使用StringBuilder进行多次字符串拼接
import "stringbuilder"

var sb = stringbuilder.create()
for (i in [1, 2, 3, 4, 5]) {
    sb.write(str(i))  // 高效
}

// 避免这样写:
var s = ""
for (i in [1, 2, 3, 4, 5]) {
    s = s + str(i)  // 每次都创建新字符串
}
```

### 内存管理

```xxl
// 尽可能复用对象
var sb = stringbuilder.create(1000)  // 预分配容量

// 清空并复用
sb.clear()
// 再次使用...
```

---

## 常见问题解决

### 如何实现枚举

```xxl
// 使用常量模拟枚举
var Status = {
    "PENDING": 0,
    "PROCESSING": 1,
    "COMPLETED": 2,
    "FAILED": 3
}

var status = Status["PENDING"]
```

### 如何实现单例

```xxl
// 使用模块级变量实现单例
var _instance = null

func getInstance() {
    if (_instance == null) {
        _instance = new SomeClass()
    }
    return _instance
}
```

### 如何延迟执行

```xxl
import "time"

func delay(ms, fn) {
    time.sleep(ms)
    return fn()
}

delay(1000, func() {
    pln("延迟1秒后执行")
})
```

### 如何实现简单的状态机

```xxl
class StateMachine {
    func init(initialState) {
        this.state = initialState
        this.transitions = {}
    }

    func addTransition(from, event, to, action) {
        var key = from + ":" + event
        this.transitions[key] = {"to": to, "action": action}
    }

    func trigger(event) {
        var key = this.state + ":" + event
        if (hasKey(this.transitions, key)) {
            var transition = this.transitions[key]
            if (transition["action"] != null) {
                transition["action"]()
            }
            this.state = transition["to"]
            return true
        }
        return false
    }

    func getState() {
        return this.state
    }
}

// 使用示例
var sm = new StateMachine("idle")
sm.addTransition("idle", "start", "running", func() { pln("启动!") })
sm.addTransition("running", "stop", "stopped", func() { pln("停止!") })

sm.trigger("start")  // 输出: 启动!
pln(sm.getState())   // "running"
sm.trigger("stop")   // 输出: 停止!
pln(sm.getState())   // "stopped"
```
