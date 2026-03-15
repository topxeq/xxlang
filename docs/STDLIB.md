# Xxlang Standard Library Reference

## Overview

Xxlang includes a comprehensive standard library organized into modules. All modules are imported using the `import` statement.

```xxl
import "std/io"
io.println("Hello, World!")
```

## Table of Contents

- [std/io](#stdio) - Input/output operations
- [std/string](#stdstring) - String utilities
- [std/math](#stdmath) - Mathematical functions
- [std/array](#stdarray) - Array utilities
- [std/json](#stdjson) - JSON encoding/decoding
- [std/regex](#stdregex) - Regular expressions
- [std/crypto](#stdcrypto) - Cryptographic functions
- [std/time](#stdtime) - Time and date functions

---

## std/io

Input/output operations for reading, writing, and console interaction.

### Console Functions

#### print(args...)
Prints arguments without a trailing newline.

```xxl
print("Hello")
print(" ")
print("World")
// Output: Hello World
```

#### println(args...)
Prints arguments separated by spaces, followed by a newline.

```xxl
println("Hello", "World")  // "Hello World\n"
println(42, "is the answer")  // "42 is the answer\n"
```

#### printf(format, args...)
Formatted print using Go-style format specifiers.

```xxl
printf("Name: %s, Age: %d\n", "Alice", 30)
printf("Pi: %.2f\n", 3.14159)  // "Pi: 3.14"
```

#### readLine()
Reads a line from stdin.

```xxl
print("Enter name: ")
var name = readLine()
println("Hello, " + name)
```

### File Functions

#### readFile(path)
Reads entire file content as string.

```xxl
var content = readFile("data.txt")
println(content)
```

#### writeFile(path, content)
Writes string content to file.

```xxl
writeFile("output.txt", "Hello, File!")
```

#### appendFile(path, content)
Appends string content to file.

```xxl
appendFile("log.txt", "New log entry\n")
```

#### exists(path)
Returns true if file or directory exists.

```xxl
if (exists("config.json")) {
    println("Config found")
}
```

#### remove(path)
Deletes a file.

```xxl
remove("temp.txt")
```

#### mkdir(path)
Creates directory (including parent directories).

```xxl
mkdir("path/to/directory")
```

### System Functions

#### cwd()
Returns current working directory.

```xxl
println(cwd())  // "/home/user/project"
```

#### exit(code)
Exits with status code.

```xxl
exit(0)   // Success
exit(1)   // Error
```

#### env(key)
Gets environment variable value.

```xxl
var home = env("HOME")
var path = env("PATH")
```

#### setEnv(key, value)
Sets environment variable.

```xxl
setEnv("DEBUG", "true")
```

#### args()
Returns command-line arguments as array.

```xxl
var args = args()
println(args[0])  // Program name
```

---

## std/string

String manipulation utilities.

#### len(s)
Returns string length.

```xxl
len("hello")  // 5
```

#### upper(s)
Converts to uppercase.

```xxl
upper("hello")  // "HELLO"
```

#### lower(s)
Converts to lowercase.

```xxl
lower("HELLO")  // "hello"
```

#### trim(s)
Removes leading and trailing whitespace.

```xxl
trim("  hello  ")  // "hello"
```

#### substr(s, start, end)
Extracts substring.

```xxl
substr("hello", 1, 4)  // "ell"
```

#### split(s, sep)
Splits string by separator.

```xxl
split("a,b,c", ",")  // ["a", "b", "c"]
```

#### join(arr, sep)
Joins array elements with separator.

```xxl
join(["a", "b", "c"], "-")  // "a-b-c"
```

#### containsStr(s, substr)
Returns true if string contains substring.

```xxl
containsStr("hello", "ell")  // true
containsStr("hello", "xyz")  // false
```

#### replace(s, old, new)
Replaces all occurrences.

```xxl
replace("hello world", "l", "L")  // "heLLo worLd"
```

#### startsWith(s, prefix)
Returns true if string starts with prefix.

```xxl
startsWith("hello", "he")  // true
```

#### endsWith(s, suffix)
Returns true if string ends with suffix.

```xxl
endsWith("hello", "lo")  // true
```

#### indexOf(s, substr)
Returns index of first occurrence, or -1.

```xxl
indexOf("hello", "l")   // 2
indexOf("hello", "x")   // -1
```

---

## std/math

Mathematical functions.

#### abs(x)
Returns absolute value.

```xxl
abs(-42)    // 42
abs(-3.14)  // 3.14
```

#### floor(x)
Returns floor (round down).

```xxl
floor(3.7)   // 3
floor(-3.7)  // -4
```

#### ceil(x)
Returns ceiling (round up).

```xxl
ceil(3.2)    // 4
ceil(-3.2)   // -3
```

#### round(x)
Returns nearest integer.

```xxl
round(3.4)   // 3
round(3.6)   // 4
```

#### sqrt(x)
Returns square root.

```xxl
sqrt(16)   // 4
sqrt(2)    // 1.414...
```

#### pow(base, exp)
Returns base raised to power.

```xxl
pow(2, 8)   // 256
pow(10, 3)  // 1000
```

#### min(a, b)
Returns smaller value.

```xxl
min(3, 7)   // 3
```

#### max(a, b)
Returns larger value.

```xxl
max(3, 7)   // 7
```

#### sin(x), cos(x), tan(x)
Trigonometric functions (radians).

```xxl
sin(0)              // 0
cos(0)              // 1
tan(0)              // 0
```

#### log(x), log10(x)
Natural and base-10 logarithm.

```xxl
log(2.718281828)   // ~1
log10(100)         // 2
```

#### random()
Returns random float between 0 and 1.

```xxl
var r = random()  // 0.0 <= r < 1.0
```

#### randomInt(min, max)
Returns random integer in range.

```xxl
var die = randomInt(1, 6)
```

---

## std/array

Array manipulation utilities.

#### len(arr)
Returns array length.

```xxl
len([1, 2, 3])  // 3
```

#### first(arr)
Returns first element, or null if empty.

```xxl
first([1, 2, 3])  // 1
first([])         // null
```

#### last(arr)
Returns last element, or null if empty.

```xxl
last([1, 2, 3])  // 3
last([])         // null
```

#### push(arr, value)
Returns new array with value appended.

```xxl
push([1, 2], 3)  // [1, 2, 3]
```

#### pop(arr)
Returns new array without last element.

```xxl
pop([1, 2, 3])  // [1, 2]
```

#### sort(arr)
Returns sorted array (ascending).

```xxl
sort([3, 1, 4, 1, 5])  // [1, 1, 3, 4, 5]
```

#### reverse(arr)
Returns reversed array.

```xxl
reverse([1, 2, 3])  // [3, 2, 1]
```

#### sum(arr)
Returns sum of numeric elements.

```xxl
sum([1, 2, 3, 4, 5])  // 15
```

#### avg(arr)
Returns average of numeric elements.

```xxl
avg([1, 2, 3, 4, 5])  // 3
```

#### indexOf(arr, value)
Returns index of value, or -1.

```xxl
indexOf([1, 2, 3], 2)  // 1
indexOf([1, 2, 3], 5)  // -1
```

#### containsArr(arr, value)
Returns true if value is in array.

```xxl
containsArr([1, 2, 3], 2)  // true
```

#### concat(arr1, arr2)
Concatenates two arrays.

```xxl
concat([1, 2], [3, 4])  // [1, 2, 3, 4]
```

#### slice(arr, start, end)
Returns subarray.

```xxl
slice([1, 2, 3, 4, 5], 1, 4)  // [2, 3, 4]
```

#### isEmpty(arr)
Returns true if array is empty.

```xxl
isEmpty([])      // true
isEmpty([1])     // false
```

---

## std/json

JSON encoding and decoding.

#### parse(jsonString)
Parses JSON string to Xxlang value.

```xxl
var data = parse('{"name": "Alice", "age": 30}')
println(data["name"])  // "Alice"

var arr = parse('[1, 2, 3]')
println(arr[0])  // 1
```

#### stringify(value, indent)
Converts Xxlang value to JSON string.

```xxl
var obj = {"name": "Alice", "age": 30}
println(stringify(obj))        // {"name":"Alice","age":30}
println(stringify(obj, "  "))  // Pretty printed with 2-space indent
println(stringify(obj, 4))     // Pretty printed with 4-space indent
```

#### encode(value)
Alias for stringify without indent.

```xxl
encode({"a": 1})  // '{"a":1}'
```

#### decode(jsonString)
Alias for parse.

```xxl
decode('{"x": 10}')["x"]  // 10
```

---

## std/regex

Regular expression operations.

#### match(pattern, str)
Returns true if string matches pattern.

```xxl
match("\\d+", "123")       // true
match("[a-z]+", "Hello")   // false
```

#### find(pattern, str)
Returns first match or null.

```xxl
find("\\d+", "abc123def")  // "123"
```

#### findAll(pattern, str)
Returns array of all matches.

```xxl
findAll("\\d+", "a1b22c333")  // ["1", "22", "333"]
```

#### replace(pattern, str, replacement)
Replaces all matches.

```xxl
replace("\\d+", "a1b2c3", "X")  // "aXbXcX"
```

#### split(pattern, str)
Splits string by pattern.

```xxl
split("\\s+", "a  b   c")  // ["a", "b", "c"]
```

---

## std/crypto

Cryptographic functions.

#### md5(s)
Returns MD5 hash as hex string.

```xxl
md5("hello")  // "5d41402abc4b2a76b9719d911017c592"
```

#### sha1(s)
Returns SHA1 hash as hex string.

```xxl
sha1("hello")  // "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
```

#### sha256(s)
Returns SHA256 hash as hex string.

```xxl
sha256("hello")  // "2cf24dba5fb0a30e26e83b2ac5b9e29e..."
```

#### sha512(s)
Returns SHA512 hash as hex string.

```xxl
sha512("hello")  // "9b71d224bd62f3785d96d46ad3ea3d..."
```

#### hmac(algorithm, key, message)
Returns HMAC hash.

```xxl
hmac("sha256", "secret", "message")  // Hex string
```

#### base64Encode(s)
Encodes string to base64.

```xxl
base64Encode("hello")  // "aGVsbG8="
```

#### base64Decode(s)
Decodes base64 string.

```xxl
base64Decode("aGVsbG8=")  // "hello"
```

#### base64UrlEncode(s)
URL-safe base64 encoding.

```xxl
base64UrlEncode("hello?world")
```

#### base64UrlDecode(s)
URL-safe base64 decoding.

```xxl
base64UrlDecode(base64UrlEncode("hello?world"))  // "hello?world"
```

#### hexEncode(bytes)
Encodes to hex string.

```xxl
hexEncode("hello")  // "68656c6c6f"
```

#### hexDecode(hex)
Decodes hex string.

```xxl
hexDecode("68656c6c6f")  // "hello"
```

#### randomBytes(n)
Returns n random bytes as string.

```xxl
var bytes = randomBytes(16)
```

#### randomHex(n)
Returns n random bytes as hex string.

```xxl
randomHex(16)  // "a1b2c3d4..." (32 hex chars)
```

---

## std/time

Time and date functions for working with timestamps and durations.

### Timestamps

#### unix()
Returns current Unix timestamp in seconds.

```xxl
import "std/time"
var ts = time.unix()  // e.g., 1710422400
```

#### unixMs()
Returns current Unix timestamp in milliseconds.

```xxl
var ms = time.unixMs()  // e.g., 1710422400000
```

#### unixNano()
Returns current Unix timestamp in nanoseconds.

```xxl
var ns = time.unixNano()  // e.g., 1710422400000000000
```

### Date/Time Components

#### now()
Returns current time as a map with components.

```xxl
var t = time.now()
println(t["year"])     // 2026
println(t["month"])    // 3 (March)
println(t["day"])      // 14
println(t["hour"])     // 20
println(t["minute"])   // 30
println(t["second"])   // 45
println(t["nanosecond"])  // 123456789
```

#### year(), month(), day(), hour(), minute(), second()
Returns current date/time component.

```xxl
println(time.year())    // 2026
println(time.month())   // 3
println(time.day())     // 14
println(time.hour())    // 20
println(time.minute())  // 30
println(time.second())  // 45
```

#### weekday()
Returns current weekday (0=Sunday, 1=Monday, ..., 6=Saturday).

```xxl
var wd = time.weekday()  // e.g., 4 (Thursday)
```

### Formatting

#### format(layout)
Formats current time using Go-style layout.

```xxl
println(time.format("2006-01-02"))           // "2026-03-14"
println(time.format("15:04:05"))             // "20:30:45"
println(time.format("2006-01-02 15:04:05"))  // "2026-03-14 20:30:45"
```

#### formatUnix(timestamp, layout)
Formats a Unix timestamp.

```xxl
var ts = time.unix()
println(time.formatUnix(ts, "2006-01-02"))
```

#### parse(layout, value)
Parses a time string and returns Unix timestamp.

```xxl
var ts = time.parse("2006-01-02", "2026-03-14")
```

### Sleep

#### sleep(ms)
Pauses execution for specified milliseconds.

```xxl
println("Starting...")
time.sleep(1000)  // Sleep 1 second
println("Done!")
```

#### sleepSec(seconds)
Pauses execution for specified seconds.

```xxl
time.sleepSec(2)  // Sleep 2 seconds
```

### Date Arithmetic

#### addDays(days)
Adds days to current time, returns Unix timestamp.

```xxl
var tomorrow = time.addDays(1)
var nextWeek = time.addDays(7)
```

#### addMonths(months)
Adds months to current time.

```xxl
var nextMonth = time.addMonths(1)
```

#### addYears(years)
Adds years to current time.

```xxl
var nextYear = time.addYears(1)
```

#### since(startMs)
Returns milliseconds elapsed since given timestamp.

```xxl
var start = time.unixMs()
// ... do some work ...
var elapsed = time.since(start)
println("Took " + elapsed.toStr() + " ms")
```

### Calendar

#### isLeapYear(year)
Returns true if year is a leap year.

```xxl
time.isLeapYear(2024)  // true
time.isLeapYear(2023)  // false
```

#### daysInMonth(year, month)
Returns number of days in month.

```xxl
time.daysInMonth(2024, 2)  // 29 (leap year)
time.daysInMonth(2023, 2)  // 28
time.daysInMonth(2024, 1)  // 31
```

---

## Built-in Functions (Global)

These functions are available without importing any module.

### Type Functions

#### typeOf(value)
Returns type name as string.

```xxl
typeOf(42)        // "INT"
typeOf("hello")   // "STRING"
typeOf([1, 2])    // "ARRAY"
```

#### len(value)
Returns length of string, array, or map.

```xxl
len("hello")      // 5
len([1, 2, 3])    // 3
len({"a": 1})     // 1
```

### Type Conversion

#### int(value)
Converts to integer.

```xxl
int(3.14)      // 3
int("42")      // 42
int(true)      // 1
```

#### float(value)
Converts to float.

```xxl
float(42)      // 42.0
float("3.14")  // 3.14
```

#### string(value)
Converts to string.

```xxl
string(42)     // "42"
string(3.14)   // "3.14"
string(true)   // "true"
```

### Map Operations

#### keys(map)
Returns array of map keys.

```xxl
keys({"a": 1, "b": 2})  // ["a", "b"]
```

#### values(map)
Returns array of map values.

```xxl
values({"a": 1, "b": 2})  // [1, 2]
```

#### hasKey(map, key)
Returns true if map has key.

```xxl
hasKey({"a": 1}, "a")   // true
hasKey({"a": 1}, "b")   // false
```

#### delete(map, key)
Returns map with key removed.

```xxl
delete({"a": 1, "b": 2}, "a")  // {"b": 2}
```

### Range

#### range(start, stop)
Returns array of integers.

```xxl
range(0, 5)    // [0, 1, 2, 3, 4]
range(1, 4)    // [1, 2, 3]
```

### Assertions

#### assert(condition)
Throws error if condition is false.

```xxl
assert(1 + 1 == 2)
assert(x > 0, "x must be positive")
```

### String Utilities

#### repeat(s, n)
Returns string repeated n times.

```xxl
repeat("ab", 3)  // "ababab"
repeat("x", 5)   // "xxxxx"
```

#### lpad(s, length, padChar)
Pads string on the left to specified length.

```xxl
lpad("5", 4, "0")     // "0005"
lpad("hello", 10)      // "     hello"
```

#### rpad(s, length, padChar)
Pads string on the right to specified length.

```xxl
rpad("5", 4, "0")      // "5000"
rpad("hello", 10)      // "hello     "
```

#### charAt(s, index)
Returns character at index, or null if out of bounds.

```xxl
charAt("hello", 1)  // "e"
charAt("hello", 10) // null
```

#### trimLeft(s, cutset)
Removes leading characters.

```xxl
trimLeft("  hello")          // "hello"
trimLeft("xxhelloxx", "x")   // "helloxx"
```

#### trimRight(s, cutset)
Removes trailing characters.

```xxl
trimRight("hello  ")         // "hello"
trimRight("xxhelloxx", "x")  // "xxhello"
```

### Type Checking

#### isEmpty(value)
Returns true if value is empty string, array, map, or null.

```xxl
isEmpty("")          // true
isEmpty([])          // true
isEmpty({})          // true
isEmpty(null)        // true
isEmpty([1, 2])      // false
```

#### isString(value), isNumber(value), isInt(value), isFloat(value)
Returns true if value is of the specified type.

```xxl
isString("hello")    // true
isNumber(42)         // true
isNumber(3.14)       // true
isInt(42)            // true
isFloat(3.14)        // true
```

#### isArray(value), isMap(value), isBool(value), isNull(value), isFunction(value)
Returns true if value is of the specified type.

```xxl
isArray([1, 2, 3])     // true
isMap({"a": 1})         // true
isBool(true)            // true
isNull(null)            // true
isFunction(len)         // true
```

### Math Utilities

#### round(value, precision)
Rounds to nearest integer or specified decimal places.

```xxl
round(3.7)           // 4
round(3.14159, 2)    // 3.14
round(3.14159, 4)    // 3.1416
```

#### clamp(value, min, max)
Clamps value to range [min, max].

```xxl
clamp(15, 0, 10)     // 10
clamp(-5, 0, 10)     // 0
clamp(5, 0, 10)      // 5
```

#### sign(value)
Returns sign of number (-1, 0, or 1).

```xxl
sign(-42)    // -1
sign(0)      // 0
sign(42)     // 1
```

#### random()
Returns random float between 0 and 1.

```xxl
var r = random()  // 0.0 <= r < 1.0
```

#### randomInt(min, max)
Returns random integer in range [min, max].

```xxl
var die = randomInt(1, 6)   // 1, 2, 3, 4, 5, or 6
```

### Array Utilities

#### unique(arr)
Returns array with duplicate values removed.

```xxl
unique([1, 2, 2, 3, 3, 3])  // [1, 2, 3]
```

#### flatten(arr, depth)
Flattens nested arrays.

```xxl
flatten([[1, 2], [3, 4]])           // [1, 2, 3, 4]
flatten([[[1]], [2]], 1)             // [[1], 2]
```

#### without(arr, values...)
Returns array with specified values removed.

```xxl
without([1, 2, 3, 4], 2, 4)  // [1, 3]
```

#### take(arr, n)
Returns first n elements.

```xxl
take([1, 2, 3, 4, 5], 3)  // [1, 2, 3]
```

#### drop(arr, n)
Returns array without first n elements.

```xxl
drop([1, 2, 3, 4, 5], 2)  // [3, 4, 5]
```

### Map Utilities

#### merge(map1, map2, ...)
Merges multiple maps (later maps override earlier).

```xxl
merge({"a": 1}, {"b": 2})           // {"a": 1, "b": 2}
merge({"a": 1}, {"a": 2, "b": 3})    // {"a": 2, "b": 3}
```

#### entries(map)
Returns array of [key, value] pairs.

```xxl
entries({"x": 10, "y": 20})  // [["x", 10], ["y", 20]]
```

### Formatting

#### format(template, args...)
Formats string with Go-style format specifiers.

```xxl
format("Hello %s, you are %d", "Alice", 30)  // "Hello Alice, you are 30"
format("Pi is %.2f", 3.14159)                 // "Pi is 3.14"
```
