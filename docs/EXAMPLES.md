# Xxlang Common Scenarios Code Examples

This document provides practical code examples organized from simple to complex, covering common programming scenarios.

## Table of Contents

1. [Basics](#basics)
2. [Control Flow](#control-flow)
3. [Functions](#functions)
4. [Strings](#strings)
5. [Arrays](#arrays)
6. [Maps](#maps)
7. [File Operations](#file-operations)
8. [JSON Processing](#json-processing)
9. [Error Handling](#error-handling)
10. [Classes and OOP](#classes-and-oop)
11. [Regular Expressions](#regular-expressions)
12. [HTTP Requests](#http-requests)
13. [StringBuilder](#stringbuilder)
14. [Practical Examples](#practical-examples)

---

## Basics

### Variables and Types

```xxl
// Variable declarations
var name = "Alice"
var age = 30
var score = 95.5
var active = true
var nothing = null

// Arrays
var numbers = [1, 2, 3, 4, 5]
var names = ["Alice", "Bob", "Charlie"]

// Maps
var person = {
    "name": "Alice",
    "age": 30,
    "city": "New York"
}

// Access map values
pln(person["name"])  // "Alice"
```

### Type Checking

```xxl
// Check types
var x = 42
pln(isInt(x))      // true
pln(isFloat(3.14)) // true
pln(isString("hi")) // true
pln(isArray([1,2,3])) // true
pln(isMap({"a": 1})) // true
pln(isBool(true))  // true
pln(isNull(null))  // true
```

### Arithmetic Operations

```xxl
var a = 10
var b = 3

pln(a + b)   // 13
pln(a - b)   // 7
pln(a * b)   // 30
pln(a / b)   // 3
pln(a % b)   // 1

// Increment
var count = 0
count = count + 1
pln(count)  // 1
```

### String Operations

```xxl
var s = "Hello, World!"

// String concatenation
var greeting = "Hello" + " " + "World"
pln(greeting)  // "Hello World"

// String methods
pln(upper(s))      // "HELLO, WORLD!"
pln(lower(s))      // "hello, world!"
pln(len(s))        // 13
pln(trim("  hi  ")) // "hi"
```

---

## Control Flow

### If-Else

```xxl
var score = 85

if (score >= 90) {
    pln("Grade: A")
} else if (score >= 80) {
    pln("Grade: B")
} else if (score >= 70) {
    pln("Grade: C")
} else {
    pln("Grade: F")
}
```

### For Loop

```xxl
// Traditional for loop
for (var i = 0; i < 5; i = i + 1) {
    pln("Iteration:", i)
}

// For-in loop with array
var fruits = ["apple", "banana", "cherry"]
for (fruit in fruits) {
    pln("Fruit:", fruit)
}

// For-in loop with map
var scores = {"Alice": 95, "Bob": 87, "Charlie": 92}
for (name in scores) {
    pln(name, "scored", scores[name])
}
```

### While Loop

```xxl
var count = 5
while (count > 0) {
    pln("Countdown:", count)
    count = count - 1
}
pln("Liftoff!")
```

### Switch Statement

The switch statement provides clean branching based on a value. **Note: Xxlang's switch does not fall through - it automatically exits after matching a case.**

```xxl
var day = 3

switch (day) {
case 1:
    pln("Monday")
case 2:
    pln("Tuesday")
case 3:
    pln("Wednesday")  // Outputs this, then exits
case 4:
    pln("Thursday")   // Not executed
case 5:
    pln("Friday")     // Not executed
default:
    pln("Weekend")    // Not executed
}
```

#### String Matching

```xxl
var fruit = "apple"

switch (fruit) {
case "apple":
    pln("Red or Green")
case "banana":
    pln("Yellow")
default:
    pln("Unknown")
}
// Output: Red or Green
```

#### With Expressions

```xxl
var x = 10
var y = 20

switch (x + y) {
case 10:
    pln("Sum is 10")
case 30:
    pln("Sum is 30")
default:
    pln("Other sum")
}
// Output: Sum is 30
```

> See [LANGUAGE.md](LANGUAGE.md#switch-statement) for detailed documentation.

---

## Functions

### Basic Function

```xxl
func greet(name) {
    return "Hello, " + name + "!"
}

pln(greet("Alice"))  // "Hello, Alice!"
```

### Multiple Parameters

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

### Closures

```xxl
// Function that returns a function
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

// Another counter is independent
var counter2 = makeCounter()
pln(counter2())  // 1
```

### Higher-Order Functions

```xxl
// Function taking another function as parameter
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

## Strings

### String Manipulation

```xxl
import "string"

var text = "  Hello, World!  "

// Trim whitespace
pln(trim(text))                    // "Hello, World!"
pln(trimLeft(text))                // "Hello, World!  "
pln(trimRight(text))               // "  Hello, World!"

// Split and join
var parts = split("a,b,c", ",")
pln(parts)                         // ["a", "b", "c"]

pln(join(["x", "y", "z"], "-"))    // "x-y-z"

// Contains and index
pln(contains("hello", "ell"))      // true
pln(indexOf("hello", "l"))         // 2
pln(lastIndexOf("hello", "l"))     // 3

// Prefix and suffix
pln(startsWith("hello", "he"))     // true
pln(endsWith("hello", "lo"))       // true

// Case conversion
pln(upper("hello"))                // "HELLO"
pln(lower("HELLO"))                // "hello"

// Repeat
pln(repeat("ab", 3))               // "ababab"
```

### String Formatting

```xxl
import "fmt"

// Printf-style formatting
pln(fmt.sprintf("Name: %s, Age: %d", "Alice", 30))
pln(fmt.sprintf("Price: $%.2f", 19.99))
pln(fmt.sprintf("Hex: 0x%X", 255))

// Format with padding
pln(lpad("42", 5, "0"))   // "00042"
pln(rpad("42", 5, "-"))   // "42---"
```

---

## Arrays

### Array Operations

```xxl
var arr = [3, 1, 4, 1, 5, 9, 2, 6]

// Length
pln(len(arr))  // 8

// Access elements
pln(arr[0])    // 3
pln(arr[-1])   // 6 (last element)

// Push and pop
arr = push(arr, 5)
pln(arr)       // [..., 5]

var last = arr[-1]
arr = arr[:len(arr)-1]  // Remove last

// Slice
var sub = arr[2:5]
pln(sub)       // [4, 1, 5]

// Reverse
pln(reverse(arr))

// Sort
pln(sort(arr))  // [1, 1, 2, 3, 4, 5, 6, 9]
```

### Array Utilities

```xxl
import "array"

var numbers = [1, 2, 3, 4, 5]

// Functional operations
var doubled = array.map(numbers, func(x) { return x * 2 })
pln(doubled)  // [2, 4, 6, 8, 10]

var evens = array.filter(numbers, func(x) { return x % 2 == 0 })
pln(evens)    // [2, 4]

var sum = array.reduce(numbers, 0, func(acc, x) { return acc + x })
pln(sum)      // 15

// Search
pln(array.includes(numbers, 3))  // true
pln(array.indexOf(numbers, 4))   // 3

// Utilities
pln(array.first(numbers))   // 1
pln(array.last(numbers))    // 5
pln(array.min(numbers))     // 1
pln(array.max(numbers))     // 5
```

---

## Maps

### Map Operations

```xxl
var person = {
    "name": "Alice",
    "age": 30,
    "city": "New York"
}

// Access
pln(person["name"])  // "Alice"

// Modify
person["age"] = 31
person["email"] = "alice@example.com"

// Delete
person = delete(person, "city")

// Check key exists
pln(hasKey(person, "name"))   // true
pln(hasKey(person, "phone"))  // false

// Keys and values
pln(keys(person))    // ["name", "age", "email"]
pln(values(person))  // ["Alice", 31, "alice@example.com"]

// Iterate
for (key in person) {
    pln(key, ":", person[key])
}
```

### Nested Maps

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

// Iterate employees
for (emp in company["employees"]) {
    pln(emp["name"], "-", emp["role"])
}
```

---

## File Operations

### Read and Write Files

```xxl
import "io"

// Write to file
io.writeFile("test.txt", "Hello, File!")

// Append to file
io.appendFile("test.txt", "\nNew line added.")

// Read entire file
var content = io.readFile("test.txt")
pln(content)

// Check if file exists
if (io.exists("test.txt")) {
    pln("File exists!")
}

// Delete file
io.remove("test.txt")

// Create directory
io.mkdir("mydir")

// List directory
var files = io.readDir(".")
for (f in files) {
    pln(f)
}
```

### Working with Paths

```xxl
import "os"

// Get current directory
pln(os.cwd())

// Environment variables
var home = os.getenv("HOME")
pln("Home:", home)

// Command line arguments
var args = os.args()
for (arg in args) {
    pln("Arg:", arg)
}
```

---

## JSON Processing

### Parse and Generate JSON

```xxl
import "json"

// Parse JSON string
var jsonString = '{"name": "Alice", "age": 30, "skills": ["Go", "Python"]}'
var data = json.parse(jsonString)

pln(data["name"])              // "Alice"
pln(data["skills"][0])         // "Go"

// Generate JSON string
var person = {
    "name": "Bob",
    "age": 25,
    "active": true,
    "tags": ["developer", "golang"]
}

var jsonOutput = json.stringify(person)
pln(jsonOutput)

// Pretty print
var pretty = json.stringifyIndent(person, "  ")
pln(pretty)
```

### JSON with Arrays

```xxl
import "json"

// Parse JSON array
var jsonArray = '[1, 2, 3, 4, 5]'
var numbers = json.parse(jsonArray)
pln(numbers[2])  // 3

// Complex JSON
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
    pln("User", user["id"], ":", user["name"])
}
```

---

## Error Handling

### Error Checking Pattern

```xxl
import "io"

func safeReadFile(path) {
    if (!io.exists(path)) {
        return {"ok": false, "error": "File not found"}
    }

    var content = io.readFile(path)
    if (isError(content)) {
        return {"ok": false, "error": content}
    }

    return {"ok": true, "data": content}
}

var result = safeReadFile("config.json")
if (result["ok"]) {
    pln("Content:", result["data"])
} else {
    pln("Error:", result["error"])
}
```

### Try-Catch Pattern

```xxl
// Using isError to check for errors
func divide(a, b) {
    if (b == 0) {
        return error("division by zero")
    }
    return a / b
}

var result = divide(10, 0)
if (isError(result)) {
    pln("Error:", result)
} else {
    pln("Result:", result)
}
```

---

## Classes and OOP

### Basic Class

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

### Inheritance

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
        return this.name + " says Woof!"
    }
}

class Cat extends Animal {
    func speak() {
        return this.name + " says Meow!"
    }
}

var dog = new Dog("Buddy", "Golden Retriever")
var cat = new Cat("Whiskers")

pln(dog.speak())  // "Buddy says Woof!"
pln(cat.speak())  // "Whiskers says Meow!"
```

### Bank Account Example

```xxl
class BankAccount {
    func init(owner, initialBalance) {
        this.owner = owner
        this.balance = initialBalance
        this.transactions = []
    }

    func deposit(amount) {
        if (amount <= 0) {
            return error("Invalid deposit amount")
        }
        this.balance = this.balance + amount
        this.transactions = push(this.transactions, "Deposit: +" + str(amount))
        return this.balance
    }

    func withdraw(amount) {
        if (amount <= 0) {
            return error("Invalid withdrawal amount")
        }
        if (amount > this.balance) {
            return error("Insufficient funds")
        }
        this.balance = this.balance - amount
        this.transactions = push(this.transactions, "Withdraw: -" + str(amount))
        return this.balance
    }

    func getBalance() {
        return this.balance
    }

    func getStatement() {
        var result = "Account: " + this.owner + "\n"
        result = result + "Balance: $" + str(this.balance) + "\n"
        result = result + "Transactions:\n"
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

## Regular Expressions

### Pattern Matching

```xxl
import "regex"

// Check if matches
pln(regex.matches("hello123", "\\d+"))    // true
pln(regex.matches("hello", "\\d+"))       // false

// Find matches
var text = "The quick brown fox jumps over 123 lazy dogs."
var numbers = regex.findall(text, "\\d+")
pln(numbers)  // ["123"]

// Find all words
var words = regex.findall(text, "[a-z]+")
pln(words)    // ["The", "quick", "brown", ...]

// Replace
var replaced = regex.replace("hello world", "world", "Xxlang")
pln(replaced)  // "hello Xxlang"

// Replace all
var cleaned = regex.replaceAll("123-456-7890", "\\D", "")
pln(cleaned)   // "1234567890"
```

### Email Validation

```xxl
import "regex"

func isValidEmail(email) {
    var pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    return regex.matches(email, pattern)
}

pln(isValidEmail("user@example.com"))     // true
pln(isValidEmail("invalid-email"))        // false
```

### Phone Number Parsing

```xxl
import "regex"

func parsePhone(phone) {
    // Remove non-digits
    var digits = regex.replaceAll(phone, "\\D", "")

    if (len(digits) != 10) {
        return error("Invalid phone number")
    }

    return {
        "areaCode": digits[:3],
        "exchange": digits[3:6],
        "number": digits[6:]
    }
}

var parsed = parsePhone("(555) 123-4567")
pln(parsed)
// {"areaCode": "555", "exchange": "123", "number": "4567"}
```

---

## HTTP Requests

### GET Request

```xxl
import "net"

// Simple GET request
var response = net.get("https://httpbin.org/get")
pln("Status:", response["status"])
pln("Body:", response["body"])

// GET with headers
var headers = {
    "User-Agent": "Xxlang/1.0",
    "Accept": "application/json"
}
response = net.get("https://httpbin.org/headers", headers)
pln(response["body"])
```

### POST Request

```xxl
import "net"
import "json"

// POST with JSON body
var data = {
    "name": "Alice",
    "email": "alice@example.com"
}

var response = net.post(
    "https://httpbin.org/post",
    json.stringify(data),
    {"Content-Type": "application/json"}
)

pln("Status:", response["status"])
```

### API Client Example

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

// Usage
var api = new APIClient("https://jsonplaceholder.typicode.com")
var posts = api.get("/posts")
pln("First post title:", posts[0]["title"])
```

---

## StringBuilder

### Basic Usage

```xxl
import "stringbuilder"

// Create a StringBuilder
var sb = stringbuilder.create()

// Build string efficiently
sb.write("Hello")
sb.write(" ")
sb.writeLine("World!")
sb.write("The answer is: ")
sb.write(str(42))

pln(sb.toString())
// Hello World!
// The answer is: 42

pln("Length:", sb.len())
pln("IsEmpty:", sb.isEmpty())

// Clear and reuse
sb.clear()
pln("After clear, IsEmpty:", sb.isEmpty())
```

### Building Reports

```xxl
import "stringbuilder"

func buildReport(title, items) {
    var sb = stringbuilder.create()

    // Header
    sb.writeLine("=" + repeat("=", len(title)))
    sb.writeLine(title)
    sb.writeLine("=" + repeat("=", len(title)))
    sb.writeLine("")

    // Items
    var i = 1
    for (item in items) {
        sb.write(str(i) + ". ")
        sb.writeLine(item)
        i = i + 1
    }

    // Footer
    sb.writeLine("")
    sb.writeLine("Total items: " + str(len(items)))

    return sb.toString()
}

var report = buildReport("Shopping List", [
    "Milk",
    "Eggs",
    "Bread",
    "Butter"
])
pln(report)
```

### CSV Builder

```xxl
import "stringbuilder"

func buildCSV(headers, rows) {
    var sb = stringbuilder.create()

    // Header row
    var first = true
    for (h in headers) {
        if (!first) {
            sb.write(",")
        }
        sb.write(h)
        first = false
    }
    sb.writeLine("")

    // Data rows
    for (row in rows) {
        first = true
        for (val in row) {
            if (!first) {
                sb.write(",")
            }
            // Quote if contains comma
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
    ["Name", "Email", "City"],
    [
        ["Alice", "alice@example.com", "New York"],
        ["Bob", "bob@example.com", "Los Angeles"],
        ["Charlie", "charlie@example.com", "Chicago"]
    ]
)
pln(csv)
```

---

## Practical Examples

### Configuration File Parser

```xxl
import "io"
import "string"

func parseConfig(path) {
    if (!io.exists(path)) {
        return error("Config file not found: " + path)
    }

    var content = io.readFile(path)
    var lines = split(content, "\n")

    var config = {}

    for (line in lines) {
        line = trim(line)

        // Skip empty lines and comments
        if (len(line) == 0 || startsWith(line, "#")) {
            continue
        }

        // Parse key=value
        var parts = split(line, "=")
        if (len(parts) == 2) {
            var key = trim(parts[0])
            var value = trim(parts[1])

            // Remove quotes if present
            if (startsWith(value, "\"") && endsWith(value, "\"")) {
                value = value[1:len(value)-1]
            }

            config[key] = value
        }
    }

    return config
}

// config.txt:
// # Database settings
// host = localhost
// port = 5432
// name = mydb

var config = parseConfig("config.txt")
pln("Host:", config["host"])
pln("Port:", config["port"])
pln("Database:", config["name"])
```

### Word Frequency Counter

```xxl
import "string"
import "regex"

func wordFrequency(text) {
    // Convert to lowercase and extract words
    text = lower(text)
    var words = regex.findall(text, "[a-z]+")

    // Count frequencies
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

### Simple Web Scraper

```xxl
import "net"
import "regex"

func extractLinks(html) {
    var pattern = '<a[^>]+href="([^"]+)"'
    return regex.findall(html, pattern)
}

func scrapePage(url) {
    var response = net.get(url)
    if (response["status"] != 200) {
        return error("Failed to fetch: " + url)
    }

    var html = response["body"]
    var links = extractLinks(html)

    return {
        "url": url,
        "links": links,
        "linkCount": len(links)
    }
}

// Example usage
var result = scrapePage("https://example.com")
if (!isError(result)) {
    pln("Found", result["linkCount"], "links on", result["url"])
    for (link in result["links"]) {
        pln("  -", link)
    }
}
```

### Todo List Application

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
        pln("=== Todo List ===")
        for (todo in this.todos) {
            var status = "[ ]"
            if (todo["done"]) {
                status = "[x]"
            }
            pln(status, todo["id"], "-", todo["task"])
        }
        pln("=================")
    }
}

// Usage
var todos = new TodoList("todos.json")
todos.add("Learn Xxlang")
todos.add("Build a project")
todos.complete(1)
todos.list()
```

### Simple HTTP Server

```xxl
// Note: This is a conceptual example showing API design patterns
// Actual HTTP server functionality depends on net module capabilities

import "net"
import "json"

func handleRequest(request) {
    var path = request["path"]
    var method = request["method"]

    // Route: GET /api/hello
    if (path == "/api/hello" && method == "GET") {
        return {
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "body": json.stringify({"message": "Hello, World!"})
        }
    }

    // Route: POST /api/echo
    if (path == "/api/echo" && method == "POST") {
        return {
            "status": 200,
            "headers": {"Content-Type": "application/json"},
            "body": request["body"]
        }
    }

    // 404 Not Found
    return {
        "status": 404,
        "body": json.stringify({"error": "Not Found"})
    }
}

// Server would run and call handleRequest for each incoming request
```

### Data Transformation Pipeline

```xxl
import "array"

// Transform data through a pipeline of operations
func pipeline(data, operations) {
    var result = data
    for (op in operations) {
        result = op(result)
    }
    return result
}

// Sample data
var sales = [
    {"product": "Widget", "price": 10.0, "quantity": 5},
    {"product": "Gadget", "price": 25.0, "quantity": 3},
    {"product": "Widget", "price": 10.0, "quantity": 2},
    {"product": "Gizmo", "price": 15.0, "quantity": 10},
    {"product": "Gadget", "price": 25.0, "quantity": 1}
]

// Calculate total for each sale
var withTotal = array.map(sales, func(sale) {
    sale["total"] = sale["price"] * sale["quantity"]
    return sale
})

// Filter sales over $50
var bigSales = array.filter(withTotal, func(sale) {
    return sale["total"] > 50
})

// Sort by total descending
var sorted = array.sortBy(bigSales, func(a, b) {
    return b["total"] - a["total"]
})

pln("Big Sales (>$50):")
for (sale in sorted) {
    pln("  " + sale["product"] + ": $" + str(sale["total"]))
}
```

---

## Tips and Best Practices

### Code Organization

```xxl
// Use modules to organize code
// utils.xxl
func helper1() { ... }
func helper2() { ... }

// main.xxl
import "utils"
utils.helper1()
```

### Error Handling Pattern

```xxl
// Always check for errors
func safeOperation() {
    var result = riskyOperation()
    if (isError(result)) {
        pln("Error:", result)
        return null
    }
    return result
}
```

### Performance Tips

```xxl
// Use StringBuilder for multiple concatenations
import "stringbuilder"

var sb = stringbuilder.create()
for (i in [1, 2, 3, 4, 5]) {
    sb.write(str(i))  // Efficient
}

// Instead of:
var s = ""
for (i in [1, 2, 3, 4, 5]) {
    s = s + str(i)  // Creates new string each time
}
```

### Memory Management

```xxl
// Reuse objects when possible
var sb = stringbuilder.create(1000)  // Pre-allocate capacity

// Clear and reuse
sb.clear()
// Use again...
```
