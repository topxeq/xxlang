# Built-in Functions Reference

This document provides a comprehensive reference for all built-in functions in Xxlang.

## Table of Contents

- [Preset Global Variables](#preset-global-variables)
- [Basic Functions](#basic-functions)
- [String Functions](#string-functions)
- [Math Functions](#math-functions)
- [Type Conversion Functions](#type-conversion-functions)
- [Array Functions](#array-functions)
- [Map Functions](#map-functions)
- [Command Line Argument Functions](#command-line-argument-functions)
- [Utility Functions](#utility-functions)
- [Dynamic Code Execution](#dynamic-code-execution)
- [Type Methods](#type-methods)
- [Standard Library Modules](#standard-library-modules)

---

## Preset Global Variables

Xxlang provides preset global variables that are automatically available in all scripts:

### argsG

A string array containing all command line arguments (including the program name and script path).

```xxl
// Example: running `xxl script.xxl -- -port=8080 -verbose`
// argsG would be: ["xxl", "script.xxl", "--", "-port=8080", "-verbose"]

println("Arguments: ", argsG)
println("First arg: ", argsG[0])
```

### scriptPathG

The path of the currently executing script. This can be:
- A file path when running a local script
- A URL when running a script from a URL
- An empty string when in REPL mode or embedded execution

```xxl
println("Script path: ", scriptPathG)

if (scriptPathG == "") {
    println("Running in REPL or embedded mode")
}
```

### Example: Command Line Argument Parsing

```xxl
// Parse command line arguments
var port = getSwitch(argsG, "-port=", "8080")
var host = getSwitch(argsG, "-host=", "localhost")
var verbose = includes(argsG, "-verbose")

println("Port: ", port)
println("Host: ", host)
println("Verbose: ", verbose)
```

---

## Basic Functions

### len(obj)

Returns the length of a string, array, or map.

```xxl
len("hello")      // 5
len([1, 2, 3])    // 3
len({"a": 1})     // 1
```

### pr(args...)

Prints arguments to stdout without a newline.

```xxl
pr("Hello", " ", "World")  // Hello World
```

### pln(args...)

Prints arguments to stdout with a trailing newline.

```xxl
pln("Hello", "World")  // Hello World
```

### pl(format, args...)

Formatted print with a trailing newline. Similar to Go's `fmt.Printf` but automatically adds `\n` at the end.

```xxl
pl("Hello, World!")                    // Hello, World!
pl("Name: %s, Age: %d", "Alice", 30)   // Name: Alice, Age: 30
pl("Value: %.2f", 3.14159)             // Value: 3.14
pl("Bool: %v, Int: %d", true, 42)      // Bool: true, Int: 42
```

**Format verbs:**
- `%s` - String
- `%d` - Integer (decimal)
- `%f` - Float
- `%.Nf` - Float with N decimal places
- `%v` - Default format for any value
- `%t` - Boolean (true/false)
- `%x` - Integer (hexadecimal)
- `%o` - Integer (octal)
- `%b` - Integer (binary)
- `%%` - Literal percent sign

### prf(format, args...)

Formatted print without a trailing newline. Equivalent to Go's `fmt.Printf`.

```xxl
prf("Name: %s, Age: %d", "Alice", 30)  // Name: Alice, Age: 30 (no newline)
prf("Value: %.2f", 3.14159)             // Value: 3.14 (no newline)
pln()                                   // Add newline separately
```

**Format verbs:** Same as `pl`.

### checkErr(obj)

Checks if the argument is an error object. If it is, prints the error message to stderr and exits with code 1. Otherwise, does nothing.

```xxl
import "io"

var content = io.readFile("config.json")
checkErr(content)  // Exit if file read failed
// Continue processing if no error
pln("File loaded successfully")
```

### checkEmpty(str, message?)

Checks if a string is empty. If it is, exits with code 1. Optionally accepts a second argument as an error message to print to stderr before exiting.

```xxl
var name = ""
checkEmpty(name)                      // Exit silently if empty
checkEmpty(name, "name cannot be empty")  // Exit with message if empty

var value = "hello"
checkEmpty(value)  // Does nothing, continues execution
pln("Value: ", value)
```

### pl(format, args...)

```xxl
pl("Hello, World!")                    // Hello, World!
pl("Name: %s, Age: %d", "Alice", 30)   // Name: Alice, Age: 30
pl("Value: %.2f", 3.14159)             // Value: 3.14
pl("Bool: %v, Int: %d", true, 42)      // Bool: true, Int: 42
```

**Format verbs:**
- `%s` - String
- `%d` - Integer (decimal)
- `%f` - Float
- `%.Nf` - Float with N decimal places
- `%v` - Default format for any value
- `%t` - Boolean (true/false)
- `%x` - Integer (hexadecimal)
- `%o` - Integer (octal)
- `%b` - Integer (binary)
- `%%` - Literal percent sign

### typeOf(obj)

Returns the type of an object as a string.

```xxl
typeOf(42)        // "INT"
typeOf("hello")   // "STRING"
typeOf([1, 2])    // "ARRAY"
typeOf({"a": 1})  // "MAP"
typeOf(true)      // "BOOL"
typeOf(null)      // "NULL"
```

---

## String Functions

### substr(str, start, end?)

Extracts a substring from `start` to `end` (exclusive). If `end` is omitted, extracts to the end of the string.

```xxl
substr("hello", 1, 4)   // "ell"
substr("hello", 2)      // "llo"
```

### split(str, separator)

Splits a string by a separator and returns an array.

```xxl
split("a,b,c", ",")     // ["a", "b", "c"]
split("hello world", " ")  // ["hello", "world"]
```

### join(array, separator)

Joins array elements with a separator into a string.

```xxl
join(["a", "b", "c"], "-")  // "a-b-c"
```

### trim(str)

Removes leading and trailing whitespace.

```xxl
trim("  hello  ")  // "hello"
```

### upper(str)

Converts a string to uppercase.

```xxl
upper("hello")  // "HELLO"
```

### lower(str)

Converts a string to lowercase.

```xxl
lower("HELLO")  // "hello"
```

### containsStr(str, substr)

Checks if a string contains a substring.

```xxl
containsStr("hello world", "world")  // true
containsStr("hello", "xyz")          // false
```

### replace(str, old, new)

Replaces all occurrences of `old` with `new`.

```xxl
replace("hello world", "world", "Xxlang")  // "hello Xxlang"
```

### startsWith(str, prefix)

Checks if a string starts with a prefix.

```xxl
startsWith("hello", "he")  // true
```

### endsWith(str, suffix)

Checks if a string ends with a suffix.

```xxl
endsWith("hello", "lo")  // true
```

---

## Math Functions

### abs(num)

Returns the absolute value of a number.

```xxl
abs(-42)   // 42
abs(-3.14) // 3.14
```

### floor(num)

Returns the largest integer less than or equal to a number.

```xxl
floor(3.7)  // 3
floor(-3.7) // -4
```

### ceil(num)

Returns the smallest integer greater than or equal to a number.

```xxl
ceil(3.2)   // 4
ceil(-3.2)  // -3
```

### sqrt(num)

Returns the square root of a number.

```xxl
sqrt(16)   // 4
sqrt(2)    // 1.4142135623730951
```

### pow(base, exp)

Returns `base` raised to the power of `exp`.

```xxl
pow(2, 10)  // 1024
pow(3, 2)   // 9
```

### min(a, b)

Returns the smaller of two values.

```xxl
min(5, 3)    // 3
min(1.5, 2)  // 1.5
```

### max(a, b)

Returns the larger of two values.

```xxl
max(5, 3)    // 5
max(1.5, 2)  // 2
```

---

## Type Conversion Functions

### int(value)

Converts a value to an integer.

```xxl
int(3.7)       // 3
int("42")      // 42
int(true)      // 1
int(false)     // 0
```

### float(value)

Converts a value to a float.

```xxl
float(42)      // 42.0
float("3.14")  // 3.14
float(true)    // 1.0
```

### string(value)

Converts a value to its string representation.

```xxl
string(42)     // "42"
string(3.14)   // "3.14"
string(true)   // "true"
string([1, 2]) // "[1, 2]"
```

---

## Array Functions

### push(array, element)

Returns a new array with the element appended.

```xxl
push([1, 2, 3], 4)  // [1, 2, 3, 4]
```

### pop(array)

Returns a new array with the last element removed.

```xxl
pop([1, 2, 3])  // [1, 2]
```

### first(array)

Returns the first element of an array, or null if empty.

```xxl
first([1, 2, 3])  // 1
first([])         // null
```

### last(array)

Returns the last element of an array, or null if empty.

```xxl
last([1, 2, 3])  // 3
last([])         // null
```

### rest(array, start, end?)

Returns a slice of the array from `start` to `end` (exclusive).

```xxl
rest([1, 2, 3, 4], 1, 3)  // [2, 3]
rest([1, 2, 3, 4], 2)     // [3, 4]
```

### concat(array1, array2)

Concatenates two arrays.

```xxl
concat([1, 2], [3, 4])  // [1, 2, 3, 4]
```

### indexOf(array, element)

Returns the index of an element, or -1 if not found.

```xxl
indexOf([1, 2, 3], 2)   // 1
indexOf([1, 2, 3], 5)   // -1
```

### containsArr(array, element)

Checks if an array contains an element.

```xxl
containsArr([1, 2, 3], 2)  // true
containsArr([1, 2, 3], 5)  // false
```

### sort(array)

Returns a sorted copy of the array.

```xxl
sort([3, 1, 2])  // [1, 2, 3]
```

### sum(array)

Returns the sum of numeric elements in an array.

```xxl
sum([1, 2, 3, 4, 5])  // 15
```

### avg(array)

Returns the average of numeric elements in an array.

```xxl
avg([1, 2, 3, 4, 5])  // 3.0
```

### reverse(array)

Returns a reversed copy of the array.

```xxl
reverse([1, 2, 3])  // [3, 2, 1]
```

---

## Map Functions

### keys(map)

Returns an array of all keys in a map.

```xxl
keys({"a": 1, "b": 2})  // ["a", "b"]
```

### values(map)

Returns an array of all values in a map.

```xxl
values({"a": 1, "b": 2})  // [1, 2]
```

### hasKey(map, key)

Checks if a map contains a key.

```xxl
hasKey({"a": 1}, "a")  // true
hasKey({"a": 1}, "b")  // false
```

### delete(map, key)

Returns a new map with the specified key removed.

```xxl
delete({"a": 1, "b": 2}, "a")  // {"b": 2}
```

---

## Command Line Argument Functions

### getSwitch(array, prefix, default)

Searches an array for an element that starts with the given prefix and returns the value after the prefix. If not found, returns the default value.

This is particularly useful for parsing command line arguments.

```xxl
// With argsG = ["script.xxl", "-port=8080", "-host=localhost", "-verbose"]

var port = getSwitch(argsG, "-port=", "3000")      // "8080"
var host = getSwitch(argsG, "-host=", "127.0.0.1") // "localhost"
var debug = getSwitch(argsG, "-debug=", "false")   // "false" (not found)

// Check for flag-style arguments (no value after prefix)
var hasVerbose = getSwitch(argsG, "-verbose", "")  // "" (found, but no value)
if (hasVerbose == "" && includes(argsG, "-verbose")) {
    println("Verbose mode enabled")
}
```

**Parameters:**
- `array` - An array of strings to search (typically `argsG`)
- `prefix` - The prefix to search for (e.g., "-port=")
- `default` - The default value to return if the prefix is not found

**Returns:**
- The value after the prefix if found
- The default value if the prefix is not found

### switchExists(array, switchName)

Checks if a switch argument exists in the array with exact match. Returns `true` only if the element exactly matches the switch name.

```xxl
// With argsG = ["script.xxl", "-port=8080", "-verbose", "-debug=true"]

var hasPort = switchExists(argsG, "-port")        // false (no exact match)
var hasPortEq = switchExists(argsG, "-port=8080") // true (exact match)
var hasVerbose = switchExists(argsG, "-verbose")  // true (exact match)
var hasDebug = switchExists(argsG, "-debug")      // false (no exact match)
var hasDebugEq = switchExists(argsG, "-debug=true") // true (exact match)
```

**Parameters:**
- `array` - An array of strings to search (typically `argsG`)
- `switchName` - The switch name to look for (e.g., "-verbose")

**Returns:**
- `true` if the switch exists with exact match
- `false` if the switch is not found

---

## Utility Functions

### range(end) or range(start, end)

Generates an array of integers from start to end (inclusive).

```xxl
range(5)       // [0, 1, 2, 3, 4, 5]
range(2, 5)    // [2, 3, 4, 5]
range(5, 2)    // [5, 4, 3, 2]
```

### runCode(code, args?)

Executes Xxlang code dynamically. Optional `args` map provides variables.

```xxl
runCode("1 + 2")                    // 3
runCode("a + b", {"a": 10, "b": 20}) // 30
```

### loadPlugin(path)

Loads a native Go plugin from the specified path.

```xxl
loadPlugin("./myplugin.so")
```

### copy(obj)

Creates a shallow copy of an array or map.

```xxl
var arr = [1, 2, 3]
var arrCopy = copy(arr)
arrCopy[0] = 99
println(arr[0])     // 1 (original unchanged)

var map = {"a": 1}
var mapCopy = copy(map)
```

### clone(obj)

Creates a deep copy of an array or map.

```xxl
var nested = {"a": [1, 2, 3]}
var cloned = clone(nested)
cloned["a"][0] = 99
println(nested["a"][0])  // 1 (original unchanged)
```

### equals(a, b)

Performs deep equality comparison.

```xxl
equals([1, 2], [1, 2])           // true
equals({"a": 1}, {"a": 1})       // true
equals([1, [2, 3]], [1, [2, 3]]) // true
```

### defaults(obj, defaultObj)

Fills in missing keys from default object.

```xxl
var config = {"host": "localhost"}
var result = defaults(config, {"host": "127.0.0.1", "port": 8080})
// {"host": "localhost", "port": 8080}
```

### base64Encode(s)

Encodes string to base64.

```xxl
base64Encode("hello")  // "aGVsbG8="
```

### base64Decode(s)

Decodes base64 string.

```xxl
base64Decode("aGVsbG8=")  // "hello"
```

### hexEncode(s)

Encodes string to hexadecimal.

```xxl
hexEncode("hello")  // "68656c6c6f"
```

### hexDecode(s)

Decodes hexadecimal string.

```xxl
hexDecode("68656c6c6f")  // "hello"
```

### md5(s)

Returns MD5 hash as hex string.

```xxl
md5("hello")  // "5d41402abc4b2a76b9719d911017c592"
```

### sha256(s)

Returns SHA256 hash as hex string.

```xxl
sha256("hello")  // "2cf24dba5fb0a30e26e83b2ac5b9e29e..."
```

### sleep(ms)

Pauses execution for specified milliseconds.

```xxl
println("Starting...")
sleep(1000)  // Sleep 1 second
println("Done!")
```

### now()

Returns current Unix timestamp in seconds.

```xxl
var ts = now()  // e.g., 1710422400
```

### nowMs()

Returns current Unix timestamp in milliseconds.

```xxl
var ms = nowMs()  // e.g., 1710422400000
```

### uuid()

Generates a random UUID string.

```xxl
var id = uuid()  // "550e8400-e29b-41d4-a716-446655440000"
```

### trimPrefix(s, prefix)

Removes prefix from string if present.

```xxl
trimPrefix("hello_world", "hello_")  // "world"
trimPrefix("hello", "x")             // "hello"
```

### trimSuffix(s, suffix)

Removes suffix from string if present.

```xxl
trimSuffix("hello.txt", ".txt")  // "hello"
trimSuffix("hello", ".txt")      // "hello"
```

### count(arr)

Returns the length of an array.

```xxl
count([1, 2, 3, 4, 5])  // 5
```

### isDigit(s)

Returns true if string contains only digits.

```xxl
isDigit("12345")   // true
isDigit("12a45")   // false
```

### isAlpha(s)

Returns true if string contains only letters.

```xxl
isAlpha("hello")   // true
isAlpha("hello1")  // false
```

### isAlphaNum(s)

Returns true if string contains only letters and digits.

```xxl
isAlphaNum("hello123")  // true
isAlphaNum("hello!")    // false
```

### find(arr, predicate)

Returns first element that matches predicate, or null.

```xxl
find([1, 2, 3, 4, 5], func(x) { return x > 3 })  // 4
```

### findIndex(arr, predicate)

Returns index of first element that matches predicate, or -1.

```xxl
findIndex([1, 2, 3, 4, 5], func(x) { return x > 3 })  // 3
```

### includes(arr, value)

Returns true if array contains value.

```xxl
includes([1, 2, 3], 2)   // true
includes([1, 2, 3], 5)   // false
```

### shuffle(arr)

Returns a shuffled copy of the array.

```xxl
shuffle([1, 2, 3, 4, 5])  // [3, 1, 5, 2, 4] (random order)
```

### sample(arr, n)

Returns n random elements from array.

```xxl
sample([1, 2, 3, 4, 5], 2)  // [3, 5] (random selection)
```

### chunk(arr, size)

Splits array into chunks of specified size.

```xxl
chunk([1, 2, 3, 4, 5, 6], 2)  // [[1, 2], [3, 4], [5, 6]]
chunk([1, 2, 3, 4, 5], 2)     // [[1, 2], [3, 4], [5]]
```

---

## Type Methods

All types have the following universal methods:

- `typeOf()` - Returns the type as a string
- `toStr()` - Returns the string representation

### Int Methods

```xxl
(-5).abs()      // 5
42.toFloat()    // 42.0
42.typeOf()     // "INT"
```

### Float Methods

```xxl
3.7.floor()     // 3
3.2.ceil()      // 4
3.5.round()     // 4
(-3.14).abs()   // 3.14
3.14.toInt()    // 3
```

### String Methods

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

### Array Methods

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

### Map Methods

```xxl
{"a": 1, "b": 2}.len()     // 2
{"a": 1, "b": 2}.keys()    // ["a", "b"]
{"a": 1, "b": 2}.values()  // [1, 2]
{"a": 1}.hasKey("a")       // true
{"a": 1, "b": 2}.delete("a") // {"b": 2}
```

---

## Standard Library Modules

Import standard library modules with the `import` statement:

```xxl
import "math"
import "io" { readFile, writeFile }
```

### math

Mathematical functions and constants.

| Function | Description | Example |
|----------|-------------|---------|
| `PI` | Pi constant | `math.PI` |
| `E` | Euler's number | `math.E` |
| `abs(x)` | Absolute value | `math.abs(-5)` |
| `ceil(x)` | Round up | `math.ceil(3.2)` |
| `floor(x)` | Round down | `math.floor(3.7)` |
| `round(x)` | Round to nearest | `math.round(3.5)` |
| `sqrt(x)` | Square root | `math.sqrt(16)` |
| `pow(x, y)` | Power | `math.pow(2, 8)` |
| `sin(x)` | Sine (radians) | `math.sin(1.57)` |
| `cos(x)` | Cosine (radians) | `math.cos(0)` |
| `tan(x)` | Tangent (radians) | `math.tan(1)` |
| `asin(x)` | Arc sine | `math.asin(1)` |
| `acos(x)` | Arc cosine | `math.acos(0)` |
| `atan(x)` | Arc tangent | `math.atan(1)` |
| `log(x)` | Natural logarithm | `math.log(2.7)` |
| `log10(x)` | Base-10 logarithm | `math.log10(100)` |
| `exp(x)` | Exponential | `math.exp(1)` |
| `min(args...)` | Minimum of values | `math.min(3, 1, 2)` |
| `max(args...)` | Maximum of values | `math.max(3, 1, 2)` |
| `random()` | Random number [0, 1) | `math.random()` |

### io

Input/output operations.

| Function | Description | Example |
|----------|-------------|---------|
| `print(args...)` | Print without newline | `io.print("Hello")` |
| `println(args...)` | Print with newline | `io.println("Hello")` |
| `printf(fmt, args...)` | Formatted print | `io.printf("Value: %d", 42)` |
| `readLine()` | Read a line from stdin | `io.readLine()` |
| `readFile(path)` | Read file contents | `io.readFile("data.txt")` |
| `readBytes(path)` | Read file as byte array | `io.readBytes("data.bin")` |
| `writeFile(path, content)` | Write string to file | `io.writeFile("out.txt", "data")` |
| `writeBytes(path, bytes)` | Write bytes to file | `io.writeBytes("out.bin", bytes)` |
| `appendFile(path, content)` | Append to file | `io.appendFile("log.txt", "msg")` |
| `exists(path)` | Check if file exists | `io.exists("data.txt")` |
| `remove(path)` | Delete file | `io.remove("temp.txt")` |
| `mkdir(path)` | Create directory | `io.mkdir("mydir")` |
| `cwd()` | Get current directory | `io.cwd()` |
| `exit(code)` | Exit program | `io.exit(0)` |
| `env(key)` | Get environment variable | `io.env("HOME")` |
| `setEnv(key, value)` | Set environment variable | `io.setEnv("DEBUG", "1")` |
| `args()` | Get command line arguments | `io.args()` |

### os

Operating system utilities.

| Function | Description | Example |
|----------|-------------|---------|
| `join(paths...)` | Join path components | `os.join("dir", "file.txt")` |
| `base(path)` | Get file basename | `os.base("/a/b/c.txt")` |
| `dir(path)` | Get directory path | `os.dir("/a/b/c.txt")` |
| `ext(path)` | Get file extension | `os.ext("file.txt")` |
| `abs(path)` | Get absolute path | `os.abs("./file")` |
| `clean(path)` | Clean path | `os.clean("./a/../b")` |
| `isAbs(path)` | Check if absolute path | `os.isAbs("/home")` |
| `stat(path)` | Get file info | `os.stat("file.txt")` |
| `size(path)` | Get file size | `os.size("file.txt")` |
| `isDir(path)` | Check if directory | `os.isDir("mydir")` |
| `isFile(path)` | Check if file | `os.isFile("file.txt")` |
| `listDir(path)` | List directory contents | `os.listDir(".")` |
| `walk(path)` | Walk directory tree | `os.walk("/home")` |
| `exec(cmd)` | Execute command | `os.exec("ls -la")` |
| `shell(cmd)` | Execute shell command | `os.shell("echo hello")` |
| `hostname()` | Get hostname | `os.hostname()` |
| `platform()` | Get OS name | `os.platform()` |
| `arch()` | Get CPU architecture | `os.arch()` |
| `home()` | Get home directory | `os.home()` |
| `temp()` | Get temp directory | `os.temp()` |
| `rename(old, new)` | Rename file | `os.rename("a.txt", "b.txt")` |
| `copy(src, dst)` | Copy file | `os.copy("a.txt", "b.txt")` |
| `chmod(path, mode)` | Change permissions | `os.chmod("file", 0755)` |
| `tempFile(pattern)` | Create temp file | `os.tempFile("app-*")` |
| `tempDir(pattern)` | Create temp directory | `os.tempDir("app-*")` |

### json

JSON encoding and decoding.

| Function | Description | Example |
|----------|-------------|---------|
| `parse(str)` | Parse JSON string | `json.parse('{"a": 1}')` |
| `stringify(obj, indent?)` | Convert to JSON string | `json.stringify(obj, 2)` |
| `encode(obj)` | Encode to JSON | `json.encode(obj)` |
| `decode(str)` | Decode JSON string | `json.decode('{"a": 1}')` |

### regex

Regular expression operations (PCRE compatible).

| Function | Description | Example |
|----------|-------------|---------|
| `compile(pattern)` | Compile regex | `regex.compile("\\d+")` |
| `match(pattern, str)` | Check if matches | `regex.match("\\d+", "abc123")` |
| `find(pattern, str)` | Find first match | `regex.find("\\d+", "abc123")` |
| `findAll(pattern, str, limit?)` | Find all matches | `regex.findAll("\\d+", "a1b2c3")` |
| `findGroups(pattern, str)` | Get captured groups | `regex.findGroups("(\\d+)-(\\d+)", "1-2")` |
| `replace(pattern, str, repl)` | Replace matches | `regex.replace("\\d+", "a1b2", "X")` |
| `split(pattern, str, limit?)` | Split by regex | `regex.split("\\s+", "a b c")` |
| `quote(str)` | Escape regex chars | `regex.quote("a.b")` |
| `count(pattern, str)` | Count matches | `regex.count("\\d+", "a1b2c3")` |
| `test(pattern)` | Validate pattern | `regex.test("\\d+")` |

### time

Time and date operations.

| Function | Description | Example |
|----------|-------------|---------|
| `unix()` | Current Unix timestamp (seconds) | `time.unix()` |
| `unixMs()` | Current Unix timestamp (ms) | `time.unixMs()` |
| `unixNano()` | Current Unix timestamp (ns) | `time.unixNano()` |
| `now()` | Current time as map | `time.now()` |
| `year()` | Current year | `time.year()` |
| `month()` | Current month (1-12) | `time.month()` |
| `day()` | Current day | `time.day()` |
| `hour()` | Current hour (0-23) | `time.hour()` |
| `minute()` | Current minute | `time.minute()` |
| `second()` | Current second | `time.second()` |
| `weekday()` | Day of week (0=Sunday) | `time.weekday()` |
| `sleep(ms)` | Sleep milliseconds | `time.sleep(1000)` |
| `sleepSec(sec)` | Sleep seconds | `time.sleepSec(1)` |
| `format(layout)` | Format current time | `time.format("2006-01-02")` |
| `formatUnix(ts, layout)` | Format timestamp | `time.formatUnix(ts, "2006-01-02")` |
| `parse(layout, value)` | Parse time string | `time.parse("2006-01-02", "2024-01-15")` |
| `since(ms)` | Milliseconds since timestamp | `time.since(start)` |
| `addDays(days)` | Add days to now | `time.addDays(7)` |
| `addMonths(months)` | Add months to now | `time.addMonths(1)` |
| `addYears(years)` | Add years to now | `time.addYears(1)` |
| `isLeapYear(year)` | Check leap year | `time.isLeapYear(2024)` |
| `daysInMonth(year, month)` | Days in month | `time.daysInMonth(2024, 2)` |

### string

String utilities module.

| Function | Description | Example |
|----------|-------------|---------|
| `len(s)` | String length | `string.len("hello")` |
| `substr(s, start, end?)` | Substring | `string.substr("hello", 1, 4)` |
| `indexOf(s, substr)` | Find substring | `string.indexOf("hello", "ll")` |
| `contains(s, substr)` | Contains check | `string.contains("hello", "ell")` |
| `hasPrefix(s, prefix)` | Has prefix | `string.hasPrefix("hello", "he")` |
| `hasSuffix(s, suffix)` | Has suffix | `string.hasSuffix("hello", "lo")` |
| `toUpper(s)` | To uppercase | `string.toUpper("hello")` |
| `toLower(s)` | To lowercase | `string.toLower("HELLO")` |
| `trim(s)` | Trim whitespace | `string.trim("  hello  ")` |
| `trimSpace(s)` | Trim whitespace | `string.trimSpace("  hello  ")` |
| `split(s, sep)` | Split string | `string.split("a,b,c", ",")` |
| `join(arr, sep)` | Join strings | `string.join(["a", "b"], "-")` |
| `repeat(s, n)` | Repeat string | `string.repeat("ab", 3)` |
| `replace(s, old, new)` | Replace all | `string.replace("hello", "l", "L")` |
| `parseInt(s)` | Parse integer | `string.parseInt("42")` |
| `parseFloat(s)` | Parse float | `string.parseFloat("3.14")` |
| `toString(x)` | Convert to string | `string.toString(42)` |
| `reverse(s)` | Reverse string | `string.reverse("hello")` |

### crypto

Cryptographic functions.

| Function | Description | Example |
|----------|-------------|---------|
| `md5(s)` | MD5 hash | `crypto.md5("hello")` |
| `sha1(s)` | SHA1 hash | `crypto.sha1("hello")` |
| `sha256(s)` | SHA256 hash | `crypto.sha256("hello")` |
| `sha512(s)` | SHA512 hash | `crypto.sha512("hello")` |
| `base64Encode(s)` | Base64 encode | `crypto.base64Encode("hello")` |
| `base64Decode(s)` | Base64 decode | `crypto.base64Decode("aGVsbG8=")` |
| `hexEncode(s)` | Hex encode | `crypto.hexEncode("hello")` |
| `hexDecode(s)` | Hex decode | `crypto.hexDecode("68656c6c6f")` |

### fmt

Formatting utilities.

| Function | Description | Example |
|----------|-------------|---------|
| `sprintf(format, args...)` | Format string | `fmt.sprintf("Name: %s, Age: %d", "John", 25)` |
| `printf(format, args...)` | Print formatted | `fmt.printf("Value: %d\n", 42)` |

### array

Extended array utilities.

| Function | Description | Example |
|----------|-------------|---------|
| `map(arr, fn)` | Map elements | `array.map([1, 2, 3], fn(x) { x * 2 })` |
| `filter(arr, fn)` | Filter elements | `array.filter([1, 2, 3], fn(x) { x > 1 })` |
| `reduce(arr, fn, init)` | Reduce elements | `array.reduce([1, 2, 3], fn(a, b) { a + b }, 0)` |
| `forEach(arr, fn)` | Iterate elements | `array.forEach([1, 2, 3], fn(x) { println(x) })` |

### collections

Collection utilities (sets, stacks, queues).

### bytes

Byte array operations.

### csv

CSV file parsing and writing.

### debug

Debugging utilities.

### encoding

Encoding/decoding utilities (Base64, Hex).

### env

Environment variable utilities.

### log

Logging utilities.

### net

Network utilities.

### sort

Advanced sorting utilities.

### strconv

String conversion utilities.

### text

Text processing utilities.

### uuid

UUID generation.

---

## See Also

- [Language Reference](LANGUAGE.md) - Complete language syntax
- [Standard Library](STDLIB.md) - Standard library overview
- [Embedding Guide](EMBEDDING.md) - Using Xxlang in Go applications
