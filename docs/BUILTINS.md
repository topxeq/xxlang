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
- [Encryption Functions](#encryption-functions)
- [HTTP Client Functions](#http-client-functions)
- [HTTP Server Functions](#http-server-functions)
- [WebSocket Functions](#websocket-functions)
- [Concurrency Functions](#concurrency-functions)
- [Context Functions](#context-functions)
- [Database Functions](#database-functions)
- [Type Checking Functions](#type-checking-functions)
- [Formatting Functions](#formatting-functions)
- [Dynamic Code Execution](#dynamic-code-execution)
- [BigInt/BigFloat Functions](#bigintbigfloat-functions)
- [Reader/Writer Functions](#readerwriter-functions)
- [System Command Functions](#system-command-functions)
- [Collection Functions](#collection-functions)
- [String Processing Functions](#string-processing-functions)
- [Bitwise Functions](#bitwise-functions)
- [Check/Validation Functions](#checkvalidation-functions)
- [Bytes Functions](#bytes-functions)
- [Miscellaneous Functions](#miscellaneous-functions)
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

pln("Arguments: ", argsG)
pln("First arg: ", argsG[0])
```

### scriptPathG

The path of the currently executing script. This can be:
- A file path when running a local script
- A URL when running a script from a URL
- An empty string when in REPL mode or embedded execution

```xxl
pln("Script path: ", scriptPathG)

if (scriptPathG == "") {
    pln("Running in REPL or embedded mode")
}
```

### Example: Command Line Argument Parsing

```xxl
// Parse command line arguments
var port = getSwitch(argsG, "-port=", "8080")
var host = getSwitch(argsG, "-host=", "localhost")
var verbose = includes(argsG, "-verbose")

pln("Port: ", port)
pln("Host: ", host)
pln("Verbose: ", verbose)
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

### checkErr(obj, message?)

Checks if the argument is an error object. If it is, prints the error message to stderr and exits with code 1. Otherwise, does nothing. Optionally accepts a custom message. If the message contains format verbs like `%v`, the error message is used as the argument.

```xxl
import "io"

var content = io.readFile("config.json")
checkErr(content)                        // Exit with default error message
checkErr(content, "Failed to read file") // Exit with custom message
checkErr(content, "Error: %v")           // Exit with formatted message
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

### genOtpCode(secret)

Generates a TOTP (Time-based One-Time Password) code from a base32-encoded secret. Returns the 6-digit OTP code as a string, or an error object if the secret is invalid.

```xxl
// Generate OTP code from a TOTP secret
var code = genOtpCode("JBSWY3DPEHPK3PXP")
pln("OTP Code: ", code)  // e.g., "398139"

// Check for errors
checkErr(code, "Failed to generate OTP: %v")

// Use with authenticator apps (Google Authenticator, Authy, etc.)
var secret = "HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ"
var otp = genOtpCode(secret)
pln("Your OTP: ", otp)
```

**Note:** The secret must be a valid base32-encoded string as used by TOTP authenticator apps.

### typeOf(obj, detailed?)

Returns the type of an object as a string. With `detailed=true`, returns the class name for instances instead of "INSTANCE".

```xxl
typeOf(42)        // "INT"
typeOf("hello")   // "STRING"
typeOf([1, 2])    // "ARRAY"
typeOf({"a": 1})  // "MAP"
typeOf(true)      // "BOOL"
typeOf(null)      // "NULL"

// For class instances
class Person { var name = "" }
var p = new Person()
typeOf(p)           // "INSTANCE"
typeOf(p, true)     // "Person" (returns class name)
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

### padLeft(str, width, padChar?)

Pads a string on the left to the specified width.

```xxl
padLeft("5", 4)         // "   5" (padded with spaces)
padLeft("5", 4, "0")    // "0005"
padLeft("hello", 3)     // "hello" (already wider than width)
```

### padRight(str, width, padChar?)

Pads a string on the right to the specified width.

```xxl
padRight("5", 4)        // "5   " (padded with spaces)
padRight("5", 4, "0")   // "5000"
padRight("hello", 3)    // "hello" (already wider than width)
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

### round(value, precision?)

Rounds a number to the nearest integer or to specified decimal places.

```xxl
round(3.7)           // 4
round(3.14159, 2)    // 3.14
round(3.14159, 4)    // 3.1416
```

### clamp(value, min, max)

Clamps a value to the range [min, max].

```xxl
clamp(15, 0, 10)     // 10
clamp(-5, 0, 10)     // 0
clamp(5, 0, 10)      // 5
```

### sign(value)

Returns the sign of a number (-1, 0, or 1).

```xxl
sign(-42)    // -1
sign(0)      // 0
sign(42)     // 1
```

### random()

Returns a random float between 0 and 1.

```xxl
var r = random()  // 0.0 <= r < 1.0
```

### randomInt(min, max)

Returns a random integer in the range [min, max] (inclusive).

```xxl
var die = randomInt(1, 6)   // 1, 2, 3, 4, 5, or 6
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

### sortByFunc(array, comparator)

Sorts an array in-place using a custom comparator function. The comparator receives two indices (idx1, idx2) and returns true if the element at idx1 should come before idx2.

```xxl
// Sort by custom criteria
var users = [
    {"name": "Bob", "age": 30},
    {"name": "Alice", "age": 25},
    {"name": "Charlie", "age": 35}
]

// Sort by age (ascending)
users.sortByFunc(func(i, j) {
    return users[i]["age"] < users[j]["age"]
})
// Result: [{"name": "Alice", "age": 25}, {"name": "Bob", "age": 30}, {"name": "Charlie", "age": 35}]

// Sort by name (alphabetically)
users.sortByFunc(func(i, j) {
    return users[i]["name"] < users[j]["name"]
})
// Result: [{"name": "Alice", "age": 25}, {"name": "Bob", "age": 30}, {"name": "Charlie", "age": 35}]

// Sort numbers by absolute value
var nums = [-5, 2, -1, 8, -3]
nums.sortByFunc(func(i, j) {
    return abs(nums[i]) < abs(nums[j])
})
// Result: [-1, 2, -3, -5, 8]
```

**Note:** Unlike `sort()`, `sortByFunc` sorts the array in-place and returns the same array.

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

### unique(array)

Returns an array with duplicate values removed.

```xxl
unique([1, 2, 2, 3, 3, 3])  // [1, 2, 3]
```

### flatten(array, depth?)

Flattens nested arrays to the specified depth.

```xxl
flatten([[1, 2], [3, 4]])           // [1, 2, 3, 4]
flatten([[[1]], [2]], 1)            // [[1], 2] (depth 1)
flatten([[[1]], [2]])               // [1, 2] (full flatten)
```

### without(array, values...)

Returns an array with the specified values removed.

```xxl
without([1, 2, 3, 4], 2, 4)  // [1, 3]
```

### take(array, n)

Returns the first n elements of an array.

```xxl
take([1, 2, 3, 4, 5], 3)  // [1, 2, 3]
```

### drop(array, n)

Returns an array without the first n elements.

```xxl
drop([1, 2, 3, 4, 5], 2)  // [3, 4, 5]
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

### merge(map1, map2, ...)

Merges multiple maps. Later maps override earlier ones for duplicate keys.

```xxl
merge({"a": 1}, {"b": 2})           // {"a": 1, "b": 2}
merge({"a": 1}, {"a": 2, "b": 3})    // {"a": 2, "b": 3}
```

### entries(map)

Returns an array of [key, value] pairs.

```xxl
entries({"x": 10, "y": 20})  // [["x", 10], ["y", 20]]
```

---

## OrderedMap Functions

OrderedMap is a map that preserves insertion order of key-value pairs. Unlike regular Maps, OrderedMaps maintain the order in which keys were added, which is useful for:

- Database query results where column order matters
- Configuration files where key order should be preserved
- JSON serialization with predictable key order
- Building ordered data structures

### newOrderedMap([source])

Creates a new empty OrderedMap, or creates one from an existing Map or Array.

```xxl
// Create empty OrderedMap
var om = newOrderedMap()

// Create from Map
var om2 = newOrderedMap({"a": 1, "b": 2})

// Create from Array of [key, value] pairs
var om3 = newOrderedMap([["x", 10], ["y", 20]])

// Add entries
om["name"] = "Alice"
om["age"] = 30
om["city"] = "Beijing"

pln(om.keys())    // ["name", "age", "city"] - insertion order preserved
pln(om.values())  // ["Alice", 30, "Beijing"]
```

### make("orderedMap", [capacity])

Creates a new OrderedMap with optional pre-allocated capacity.

```xxl
var om = make("orderedMap", 100)  // Pre-allocate space for 100 entries
```

### isOrderedMap(value)

Returns true if value is an OrderedMap.

```xxl
isOrderedMap(newOrderedMap())  // true
isOrderedMap({"a": 1})         // false (regular Map)
```

---

## OrderedMap Methods

OrderedMap supports the following methods:

### Basic Operations

```xxl
var om = newOrderedMap()
om["a"] = 1           // Set value
var v = om["a"]       // Get value: 1
om.hasKey("a")        // Check key: true
om.len()              // Length: 1
var newOm = om.delete("a")  // Delete (returns new OrderedMap)
```

### Ordered Access

```xxl
var om = newOrderedMap()
om["x"] = 1
om["y"] = 2
om["z"] = 3

om.keys()             // ["x", "y", "z"] - keys in insertion order
om.values()           // [1, 2, 3] - values in insertion order
om.entries()          // [["x", 1], ["y", 2], ["z", 3]]
om.indexOf("y")       // 1 - index of key
om.getAt(0)           // ["x", 1] - [key, value] at index
```

### Reordering Operations

```xxl
var om = newOrderedMap()
om["a"] = 1
om["b"] = 2
om["c"] = 3

om.moveToFront("c")   // Move key to front: ["c", "a", "b"]
om.moveToBack("a")    // Move key to back: ["c", "b", "a"]
om.swap("c", "a")     // Swap positions: ["a", "b", "c"]
om.reverse()          // Reverse order: ["c", "b", "a"]
om.sortByKey()        // Sort by key alphabetically
```

### Conversion

```xxl
var om = newOrderedMap()
om["a"] = 1
om["b"] = 2

om.toMap()            // Convert to regular Map
om.clone()            // Create a copy
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
    pln("Verbose mode enabled")
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

### toStr(obj)

Converts any value to its string representation.

```xxl
toStr(42)       // "42"
toStr(3.14)     // "3.14"
toStr(true)     // "true"
toStr([1, 2])   // "[1, 2]"
```

### toChars(s)

Converts a string to a chars array for character-based operations. Essential for proper Unicode handling where operations work on characters (code points) rather than bytes.

```xxl
// Byte vs Character counting
var s = "Hello世界🎉"
pln(len(s))         // 15 (bytes)
pln(len(toChars(s))) // 8 (characters)

// Character indexing
var c = toChars("中文测试")
pln(c[0])           // "中"
pln(c[1])           // "文"
pln(c[-1])          // "试" (negative index)

// Character slicing
var c2 = toChars("Hello World 你好")
pln(c2.subStr(0, 5).toStr())   // "Hello"
pln(c2.subStr(5, 7).toStr())   // "世界" error - correct is subStr(6, 11)
```

**Chars Methods:**
- `toStr()` - Convert back to string
- `upper()` - Uppercase (character-aware)
- `lower()` - Lowercase (character-aware)
- `contains(sub)` - Check if contains substring
- `indexOf(sub)` - Find character index of substring
- `startsWith(prefix)` - Check prefix
- `endsWith(suffix)` - Check suffix
- `reverse()` - Reverse characters
- `repeat(n)` - Repeat n times
- `subStr(start, end)` - Slice by character indices

### charLen(s)

Returns the number of Unicode characters in a string without creating a chars object.

```xxl
charLen("Hello世界🎉")   // 8
charLen("中文测试")      // 4
charLen("hello")         // 5

// Compare with len() which returns bytes
pln(len("中文"))    // 6 (bytes)
pln(charLen("中文")) // 2 (characters)
```

---

## Utility Functions

### range(end) or range(start, end) or range(start, end, step)

Generates an array of integers from start to end (inclusive).

```xxl
range(5)           // [0, 1, 2, 3, 4, 5]
range(2, 5)        // [2, 3, 4, 5]
range(5, 2)        // [5, 4, 3, 2]
range(0, 10, 2)    // [0, 2, 4, 6, 8] (with step)
range(10, 0, -2)   // [10, 8, 6, 4, 2] (negative step)
```

**Parameters:**
- `end` - End value (inclusive), start defaults to 0
- `start` - Start value (optional)
- `step` - Step value (optional, default 1). Cannot be zero.

**Returns:** Array of integers

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

### delegate(code)

Delegates code execution to the VM context. This is primarily used internally for embedded execution scenarios where the host application needs to execute Xxlang code within a controlled context.

```xxl
// Execute code in the delegated context
var result = delegate("2 + 2")
pln(result)  // 4

// Can be used for sandboxed execution
delegate("import 'math'; math.sqrt(16)")  // 4
```

**Note:** The `delegate` function is primarily intended for advanced use cases and embedding scenarios. For most dynamic code execution needs, use `runCode()` instead.

### copy(obj)

Creates a shallow copy of an array or map.

```xxl
var arr = [1, 2, 3]
var arrCopy = copy(arr)
arrCopy[0] = 99
pln(arr[0])     // 1 (original unchanged)

var map = {"a": 1}
var mapCopy = copy(map)
```

### clone(obj)

Creates a deep copy of an array or map.

```xxl
var nested = {"a": [1, 2, 3]}
var cloned = clone(nested)
cloned["a"][0] = 99
pln(nested["a"][0])  // 1 (original unchanged)
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
pln("Starting...")
sleep(1000)  // Sleep 1 second
pln("Done!")
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

## Encryption Functions

Xxlang provides Charlang-compatible encryption/decryption functions. These are implemented without third-party dependencies and maintain full cross-compatibility with Charlang.

### encryptTextByTXTE(text, code)

Encrypts text using the TXTE (simple text encryption) algorithm. Returns a hex string. This is a deterministic encryption - same inputs always produce same output.

```xxl
var encrypted = encryptTextByTXTE("Hello", "mykey")
// "8F9FA29CA39E" (deterministic output)
```

### decryptTextByTXTE(hexStr, code)

Decrypts a hex string encrypted by TXTE.

```xxl
var decrypted = decryptTextByTXTE("8F9FA29CA39E", "mykey")
// "Hello"
```

### encryptDataByTXDEE(data, code)

Encrypts byte data using TXDEE (enhanced data encryption) with random prefix/suffix bytes. Returns a byte array.

```xxl
var data = [72, 101, 108, 108, 111]  // "Hello" bytes
var encrypted = encryptDataByTXDEE(data, "mykey")
// Returns byte array with random padding
```

### decryptDataByTXDEE(data, code)

Decrypts byte data encrypted by TXDEE.

```xxl
var decrypted = decryptDataByTXDEE(encrypted, "mykey")
// Returns original byte array
```

### encryptTextByTXDEE(text, code)

Encrypts text using TXDEE and returns a hex string.

```xxl
var encrypted = encryptTextByTXDEE("Hello", "mykey")
// Different output each time due to random bytes
```

### decryptTextByTXDEE(hexStr, code)

Decrypts a hex string encrypted by TXDEE.

```xxl
var decrypted = decryptTextByTXDEE(encrypted, "mykey")
// "Hello"
```

### encryptDataByTXDEF(data, code)

Encrypts byte data using TXDEF (flexible data encryption) with dynamic padding based on code. Returns a byte array.

```xxl
var data = [72, 101, 108, 108, 111]
var encrypted = encryptDataByTXDEF(data, "mykey")
```

### decryptDataByTXDEF(data, code)

Decrypts byte data encrypted by TXDEF.

```xxl
var decrypted = decryptDataByTXDEF(encrypted, "mykey")
```

### encryptTextByTXDEF(text, code)

Encrypts text using TXDEF and returns a hex string.

```xxl
var encrypted = encryptTextByTXDEF("Hello", "mykey")
```

### decryptTextByTXDEF(hexStr, code)

Decrypts a hex string encrypted by TXDEF.

```xxl
var decrypted = decryptTextByTXDEF(encrypted, "mykey")
// "Hello"
```

### encryptData(data, code) / decryptData(data, code)

Default data encryption using TXDEF algorithm.

```xxl
var encrypted = encryptData([1, 2, 3], "secret")
var decrypted = decryptData(encrypted, "secret")
```

### encryptBytes(data, code) / decryptBytes(data, code)

Byte array encryption aliases.

```xxl
var encrypted = encryptBytes([72, 101, 108, 108, 111], "key")
var decrypted = decryptBytes(encrypted, "key")
```

### encryptText(text, code) / decryptText(hexStr, code)

Default text encryption using TXDEF.

```xxl
var encrypted = encryptText("Hello World", "mykey")
var decrypted = decryptText(encrypted, "mykey")
// "Hello World"
```

### encryptStr(text, code) / decryptStr(hexStr, code)

String encryption aliases (same as encryptText/decryptText).

```xxl
var encrypted = encryptStr("secret message", "password")
var decrypted = decryptStr(encrypted, "password")
```

### encryptStream(reader, code, writer) / decryptStream(reader, code, writer)

Stream-based encryption/decryption for large data.

```xxl
import "io"

var reader = io.newReader("large content to encrypt")
var writer = io.newBytesWriter()
encryptStream(reader, "mykey", writer)
var encrypted = writer.bytes()

// Decrypt
var reader2 = io.newReader(encrypted)
var writer2 = io.newBytesWriter()
decryptStream(reader2, "mykey", writer2)
```

### aesEncrypt(data, key, mode?) / aesDecrypt(data, key, mode?)

AES encryption/decryption. Supports ECB-like mode (CBC with zero IV) and CBC mode.

```xxl
// ECB-like mode (default)
var encrypted = aesEncrypt([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16], "16bytekey1234567")
var decrypted = aesDecrypt(encrypted, "16bytekey1234567")

// CBC mode
var encryptedCBC = aesEncrypt(data, "16bytekey1234567", "cbc")
var decryptedCBC = aesDecrypt(encryptedCBC, "16bytekey1234567", "cbc")
```

**Note:** Key is truncated to 16 bytes if longer. For CBC mode, the key prefix is used as IV.

---

## HTTP Client Functions

### getWeb(url, options?)

Fetches URL content and returns as string. With `-object` or `-json` flag, parses JSON response.

```xxl
var content = getWeb("https://example.com/api/data")
var data = getWeb("https://api.example.com/json", "-object")
```

### getWebBytes(url, timeout?)

Fetches URL content and returns as byte array.

```xxl
var bytes = getWebBytes("https://example.com/image.png")
```

### getWebObject(url, timeout?)

Fetches URL content and parses as JSON object.

```xxl
var obj = getWebObject("https://api.example.com/data")
pln(obj["name"])
```

### postWeb(url, body, options?)

Posts data to URL and returns response as string.

```xxl
var response = postWeb("https://api.example.com/submit", '{"key": "value"}')
```

### postWebObject(url, body, options?)

Posts data to URL and parses JSON response.

```xxl
var result = postWebObject("https://api.example.com/submit", '{"name": "test"}')
```

### urlExists(url)

Returns true if URL exists (HTTP 200).

```xxl
if (urlExists("https://example.com/file.txt")) {
    pln("File exists")
}
```

### httpStatus(url)

Returns HTTP status information as a map with `statusCode`, `status`, and `headers`.

```xxl
var info = httpStatus("https://example.com")
pln("Status:", info["statusCode"])
pln("Headers:", info["headers"])
```

### downloadFile(url, localPath, timeout?)

Downloads a file from URL to local path. Returns null on success, error on failure.

```xxl
// Download a file
var result = downloadFile("https://example.com/file.zip", "/path/to/local/file.zip")
checkErr(result, "Download failed: %v")

// With custom timeout (seconds)
downloadFile("https://example.com/large.zip", "/path/to/large.zip", 120)
```

**Parameters:**
- `url` - The URL to download from
- `localPath` - The local file path to save to
- `timeout` - Optional timeout in seconds (default: 60)

**Returns:** null on success, ERROR on failure

---

## HTTP Server Functions

These functions are available in server mode (when running with `-server` flag or as a microservice).

### writeResp(body)

Writes response body to the HTTP response.

```xxl
writeResp("Hello, World!")
writeResp('{"status": "ok"}')
```

### setRespHeader(key, value)

Sets a response header.

```xxl
setRespHeader("Content-Type", "application/json")
setRespHeader("X-Custom-Header", "value")
```

### addRespHeader(key, value)

Adds a response header (appends to existing).

```xxl
addRespHeader("Set-Cookie", "session=abc123")
```

### getReqHeader(key)

Gets a request header value.

```xxl
var contentType = getReqHeader("Content-Type")
var auth = getReqHeader("Authorization")
```

### getReqHeaders()

Gets all request headers as a map.

```xxl
var headers = getReqHeaders()
pln(headers["Content-Type"])
```

### setCookie(name, value, options?)

Sets a cookie in the response.

```xxl
setCookie("session", "abc123")
setCookie("token", "xyz", {"HttpOnly": true, "Secure": true, "MaxAge": 3600})
```

### getCookie(name)

Gets a cookie value from the request.

```xxl
var session = getCookie("session")
```

### getCookies()

Gets all cookies as a map.

```xxl
var cookies = getCookies()
pln(cookies["session"])
```

### parseForm()

Parses URL-encoded form data from the request body.

```xxl
var form = parseForm()
pln(form["username"])
```

### parseJSON()

Parses JSON body from the request.

```xxl
var data = parseJSON()
pln(data["name"])
```

### getReqBody()

Gets the raw request body as a string.

```xxl
var body = getReqBody()
pln(body)
```

### getReqBodyBytes()

Gets the raw request body as a byte array.

```xxl
var bytes = getReqBodyBytes()
```

### status(code)

Sets the HTTP response status code.

```xxl
status(404)
status(500)
writeResp("Not Found")
```

### redirect(url)

Redirects to another URL.

```xxl
redirect("/login")
redirect("https://example.com")
```

### serveFile(path)

Serves a static file.

```xxl
serveFile("./static/index.html")
serveFile("/var/www/files/document.pdf")
```

### getMimeType(filename)

Gets the MIME type for a file.

```xxl
var mime = getMimeType("document.pdf")  // "application/pdf"
var mime2 = getMimeType("image.png")     // "image/png"
```

### setContentType(contentType)

Sets the Content-Type header.

```xxl
setContentType("application/json")
setContentType("text/html; charset=utf-8")
```

### queryParam(key)

Gets a query parameter value.

```xxl
// URL: /search?q=hello&page=1
var q = queryParam("q")     // "hello"
var page = queryParam("page") // "1"
```

### queryParams()

Gets all query parameters as a map.

```xxl
var params = queryParams()
pln(params["q"])
```

### formValue(key)

Gets a form value (from URL-encoded POST body).

```xxl
var username = formValue("username")
```

### httpStatusName(code)

Gets the HTTP status name for a code.

```xxl
httpStatusName(200)  // "OK"
httpStatusName(404)  // "Not Found"
httpStatusName(500)  // "Internal Server Error"
```

### isHttpReq(value)

Returns true if value is an HTTP request object.

```xxl
isHttpReq(req)  // true
isHttpReq({})   // false
```

### isHttpResp(value)

Returns true if value is an HTTP response object.

```xxl
isHttpResp(resp)  // true
isHttpResp({})    // false
```

### urlEncode(str)

URL-encodes a string.

```xxl
urlEncode("hello world")  // "hello+world"
urlEncode("a=b&c=d")      // "a%3Db%26c%3Dd"
```

### urlDecode(str)

URL-decodes a string.

```xxl
urlDecode("hello+world")  // "hello world"
urlDecode("a%3Db")        // "a=b"
```

---

## WebSocket Functions

### webSocket(upgrade?)

Upgrades an HTTP connection to WebSocket. Returns a WebSocket connection object.

```xxl
var ws = webSocket()
if (isWebSocket(ws)) {
    // Handle WebSocket connection
    var msg = wsReadMsg(ws)
    wsSendText(ws, "Echo: " + msg)
    wsClose(ws)
}
```

### wsReadMsg(ws)

Reads a message from a WebSocket connection. Returns the message as a string.

```xxl
var msg = wsReadMsg(ws)
pln("Received:", msg)
```

### wsSendText(ws, message)

Sends a text message through a WebSocket connection.

```xxl
wsSendText(ws, "Hello, client!")
```

### wsSendBinary(ws, data)

Sends binary data through a WebSocket connection.

```xxl
wsSendBinary(ws, [1, 2, 3, 4, 5])
```

### wsSendClose(ws)

Sends a close frame to the WebSocket client.

```xxl
wsSendClose(ws)
```

### wsClose(ws)

Closes the WebSocket connection.

```xxl
wsClose(ws)
```

### isWebSocket(value)

Returns true if value is a WebSocket connection object.

```xxl
isWebSocket(ws)  // true
isWebSocket({})  // false
```

---

## Concurrency Functions

Xxlang provides Go-style concurrency primitives. See [CONCURRENCY.md](CONCURRENCY.md) for detailed documentation.

### makeTube(type?, buffer?)

Creates a new tube (channel). Optional type string and buffer size.

```xxl
var tube = makeTube(10)          // Buffered tube (capacity 10)
var intTube = makeTube("INT", 5) // Typed tube for integers
var syncTube = makeTube(0)       // Unbuffered (synchronous)
```

### closeTube(tube)

Closes a tube. After closing, sends will panic and receives will return null.

```xxl
closeTube(tube)
```

### tubeLen(tube)

Returns the number of elements in the tube.

```xxl
tubeLen(tube)  // Current number of buffered elements
```

### tubeCap(tube)

Returns the capacity of the tube.

```xxl
tubeCap(tube)  // Buffer size
```

### tubeClosed(tube)

Returns true if the tube is closed.

```xxl
if (tubeClosed(tube)) {
    pln("Tube is closed")
}
```

### tubeSend(tube, value)

Sends a value to a tube. Blocking if tube is full.

```xxl
tubeSend(tube, 42)
```

### tubeRecv(tube)

Receives a value from a tube. Blocking if tube is empty. Returns null if closed.

```xxl
var val = tubeRecv(tube)
```

### tubeTrySend(tube, value)

Attempts to send without blocking. Returns true if sent, false if tube is full.

```xxl
if (!tubeTrySend(tube, data)) {
    pln("Tube is full, couldn't send")
}
```

### tubeTryRecv(tube)

Attempts to receive without blocking. Returns `[value, ok]` where ok is true if received.

```xxl
var val, ok = tubeTryRecv(tube)
if (!ok) {
    pln("No data available")
}
```

### newMutex()

Creates a new mutex for mutual exclusion.

```xxl
var mutex = newMutex()
mutex.lock()
// Critical section
mutex.unlock()
```

### newRWMutex()

Creates a new read-write mutex.

```xxl
var rwmutex = newRWMutex()
rwmutex.rLock()   // Read lock
// Read operations
rwmutex.rUnlock()
rwmutex.lock()    // Write lock
// Write operations
rwmutex.unlock()
```

### newWaitGroup()

Creates a new wait group for coordinating goroutines.

```xxl
var wg = newWaitGroup()

wg.add(1)
run {
    // Do work
    wg.done()
}

wg.wait()  // Wait for all goroutines
```

### newOnce()

Creates a one-time execution guard.

```xxl
var once = newOnce()
once.do(func() {
    pln("This will only execute once")
})
```

### newCond()

Creates a condition variable for synchronization.

```xxl
var cond = newCond()
cond.wait()     // Wait for signal
cond.signal()   // Signal one waiter
cond.broadcast() // Signal all waiters
```

### newAtomic(value?)

Creates an atomic integer for lock-free operations.

```xxl
var counter = newAtomic(0)

counter.add(1)    // Atomic add
counter.load()    // Atomic load
counter.store(10) // Atomic store
counter.swap(5)   // Atomic swap (returns old value)
```

---

## Context Functions

Context provides timeout and cancellation for operations.

### newContext()

Creates a new context.

```xxl
var ctx = newContext()
```

### contextWithTimeout(timeoutMs)

Creates a context with a timeout.

```xxl
var ctx = contextWithTimeout(5000)  // 5 second timeout
```

### contextWithCancel(parent?)

Creates a context that can be cancelled.

```xxl
var ctx = contextWithCancel()
contextCancel(ctx)  // Cancel the context
```

### contextWithDeadline(deadlineMs)

Creates a context with a deadline (absolute Unix timestamp in milliseconds).

```xxl
var deadline = nowMs() + 10000  // 10 seconds from now
var ctx = contextWithDeadline(deadline)
```

### contextCancel(ctx)

Cancels a context.

```xxl
contextCancel(ctx)
```

### contextDone(ctx)

Returns true if the context is done (cancelled or timed out).

```xxl
if (contextDone(ctx)) {
    pln("Operation cancelled or timed out")
}
```

### contextErr(ctx)

Returns the context error (null if not done, error string if cancelled/timed out).

```xxl
var err = contextErr(ctx)
if (err != null) {
    pln("Context error:", err)
}
```

### contextIsDone(ctx)

Returns true if the context is done (alias for contextDone).

```xxl
contextIsDone(ctx)
```

### contextDeadline(ctx)

Returns the deadline of the context, or 0 if no deadline.

```xxl
var deadline = contextDeadline(ctx)
```

---

## Database Functions

Xxlang provides comprehensive database builtin functions with two versions: **string-based** (Charlang compatible) and **typed** (preserve native types).

### String-Based Functions (Charlang Compatible)

These functions convert all database values to strings, providing compatibility with Charlang. NULL values are returned as empty strings.

#### formatSQLValue(str)

Escapes a string for safe SQL usage by escaping single quotes and backslashes.

```xxl
formatSQLValue("O'Brien's test")  // "O''Brien''s test"
formatSQLValue("Line1\nLine2")    // "Line1\\nLine2"
```

#### dbConnect(driver, dataSource)

Connects to a database and returns a database connection object. Returns ERROR on failure.

**Supported drivers:** `sqlite`, `sqlite3`, `mysql`, `postgres`

```xxl
// SQLite in-memory database
var db = dbConnect("sqlite", ":memory:")

// SQLite file database
var db = dbConnect("sqlite", "/path/to/database.sqlite")

// MySQL
var db = dbConnect("mysql", "user:password@tcp(localhost:3306)/dbname")

// PostgreSQL
var db = dbConnect("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")

if (typeOf(db) == "ERROR") {
    pln("Connection failed:", db)
}
```

#### dbClose(db)

Closes a database connection.

```xxl
dbClose(db)
```

#### dbQuery(db, query, params...)

Executes a query and returns an array of maps where all values are strings. Supports parameterized queries.

```xxl
// Query all rows
var rows = dbQuery(db, "SELECT * FROM users ORDER BY id")

// Query with parameters
var rows = dbQuery(db, "SELECT * FROM users WHERE age > ?", 25)

// Access values (all as strings)
for (row in rows) {
    pln("Name:", row["name"], "Age:", row["age"])  // age is string "30"
}
```

#### dbQueryOrdered(db, query, params...)

Executes a query and returns an array of OrderedMaps, where each OrderedMap preserves column order from the query. All values are converted to strings.

```xxl
var ordered = dbQueryOrdered(db, "SELECT name, salary FROM employees ORDER BY salary DESC LIMIT 3")
for (row in ordered) {
    // row is an OrderedMap with keys in query column order
    pln("columns:", row.keys())        // ["name", "salary"]
    pln("values:", row.values())       // ["Alice", "50000"]
    pln(row["name"], "=", row["salary"])  // Direct access by column name
}
```

#### dbQueryRecs(db, query, params...)

Executes a query and returns a 2D array with the first row as column headers.

```xxl
var recs = dbQueryRecs(db, "SELECT name, age FROM users LIMIT 3")
// recs[0] = ["name", "age"]      (header row)
// recs[1] = ["Alice", "30"]      (data row)
// recs[2] = ["Bob", "25"]
```

#### dbQueryMap(db, query, keyColumn, params...)

Executes a query and returns a map where each key maps to a single row. Useful for GROUP BY queries with aggregation.

```xxl
// Group by department, get count
var byDept = dbQueryMap(db, "SELECT department, COUNT(*) as cnt FROM employees GROUP BY department", "department")
var keys = keys(byDept)
for (k in keys) {
    pln(k, ":", byDept[k]["cnt"])
}
```

#### dbQueryMapArray(db, query, keyColumn, params...)

Executes a query and returns a map where each key maps to an array of rows. Useful for grouping multiple rows by a column.

```xxl
// Group employees by department
var byDept = dbQueryMapArray(db, "SELECT * FROM employees ORDER BY salary", "department")
var keys = keys(byDept)
for (k in keys) {
    var arr = byDept[k]
    pln(k, ":", len(arr), "employees")
}
```

#### dbQueryCount(db, query, params...)

Executes a query and returns the first column of the first row as an integer. Useful for COUNT queries.

```xxl
var total = dbQueryCount(db, "SELECT COUNT(*) FROM employees")
var active = dbQueryCount(db, "SELECT COUNT(*) FROM employees WHERE active = ?", 1)
```

#### dbQueryFloat(db, query, params...)

Executes a query and returns the first column of the first row as a float. Useful for SUM, AVG queries.

```xxl
var totalSalary = dbQueryFloat(db, "SELECT SUM(salary) FROM employees")
var avgAge = dbQueryFloat(db, "SELECT AVG(age) FROM employees WHERE department = ?", "Engineering")
```

#### dbQueryString(db, query, params...)

Executes a query and returns the first column of the first row as a string.

```xxl
var name = dbQueryString(db, "SELECT name FROM users WHERE id = ?", 1)
var topDept = dbQueryString(db, "SELECT department FROM employees GROUP BY department ORDER BY COUNT(*) DESC LIMIT 1")
```

#### dbExec(db, query, params...)

Executes a SQL statement (INSERT, UPDATE, DELETE, CREATE, etc.) and returns `[lastInsertId, rowsAffected]`.

```xxl
// CREATE TABLE
var result = dbExec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")

// INSERT
var result = dbExec(db, "INSERT INTO users (name, age) VALUES (?, ?)", "Alice", 30)
pln("Inserted ID:", result[0], "Affected rows:", result[1])

// UPDATE
var result = dbExec(db, "UPDATE users SET age = ? WHERE name = ?", 31, "Alice")
pln("Affected rows:", result[1])

// DELETE
var result = dbExec(db, "DELETE FROM users WHERE id = ?", 1)
```

### Typed Functions (Preserve Native Types)

These functions preserve native data types (int, float, bool, string, time). NULL values are returned as `null`.

#### dbQueryTyped(db, query, params...)

Executes a query and returns an array of maps with native types preserved.

```xxl
var rows = dbQueryTyped(db, "SELECT * FROM users ORDER BY id")
for (row in rows) {
    pln("Name:", row["name"], "type:", typeOf(row["name"]))  // STRING
    pln("Age:", row["age"], "type:", typeOf(row["age"]))      // INT
    pln("Salary:", row["salary"], "type:", typeOf(row["salary"]))  // FLOAT
}
```

#### dbQueryRowTyped(db, query, params...)

Executes a query and returns a single row as a map with native types preserved. Returns `null` if no rows found.

```xxl
var user = dbQueryRowTyped(db, "SELECT * FROM users WHERE id = ?", 1)
if (user != null) {
    pln("Name:", user["name"])
}

var notFound = dbQueryRowTyped(db, "SELECT * FROM users WHERE id = ?", 999)
// notFound is null
```

#### dbQueryArrayTyped(db, query, params...)

Executes a query and returns an array of arrays (rows as arrays, not maps) with native types preserved.

```xxl
var rows = dbQueryArrayTyped(db, "SELECT name, age FROM employees ORDER BY age")
for (row in rows) {
    pln("Name:", row[0], "Age:", row[1], "Age type:", typeOf(row[1]))
}
```

#### dbQueryValueTyped(db, query, params...)

Executes a query and returns a single value with native type preserved. Returns `null` if no rows found.

```xxl
var maxAge = dbQueryValueTyped(db, "SELECT MAX(age) FROM employees")  // INT
var avgSalary = dbQueryValueTyped(db, "SELECT AVG(salary) FROM employees")  // FLOAT
var name = dbQueryValueTyped(db, "SELECT name FROM users WHERE id = ?", 1)  // STRING
```

### NULL Handling

**String-based functions:** NULL values become empty strings `""`.

**Typed functions:** NULL values become `null`.

```xxl
// Insert with NULL
dbExec(db, "INSERT INTO users (name, age) VALUES (?, ?)", "Unknown", null)

// String version - NULL becomes ""
var rows = dbQuery(db, "SELECT * FROM users WHERE name = ?", "Unknown")
pln(rows[0]["age"])  // "" (empty string)

// Typed version - NULL becomes null
var rowsTyped = dbQueryTyped(db, "SELECT * FROM users WHERE name = ?", "Unknown")
pln(rowsTyped[0]["age"])  // null
pln(typeOf(rowsTyped[0]["age"]))  // "NULL"
```

### Complete Example

```xxl
// Connect to SQLite in-memory database
var db = dbConnect("sqlite", ":memory:")

// Create table
dbExec(db, "CREATE TABLE employees (id INTEGER PRIMARY KEY, name TEXT, age INTEGER, salary REAL, department TEXT)")

// Insert data
dbExec(db, "INSERT INTO employees (name, age, salary, department) VALUES (?, ?, ?, ?)", "Alice", 30, 75000.50, "Engineering")
dbExec(db, "INSERT INTO employees (name, age, salary, department) VALUES (?, ?, ?, ?)", "Bob", 25, 65000.00, "Marketing")

// Query with string-based function (all values are strings)
var rows = dbQuery(db, "SELECT * FROM employees")
for (row in rows) {
    pln(row["name"], "age:", row["age"], "(type:", typeOf(row["age"]), ")")
}
// Output: Alice age: 30 (type: STRING)

// Query with typed function (native types preserved)
var rowsTyped = dbQueryTyped(db, "SELECT * FROM employees")
for (row in rowsTyped) {
    pln(row["name"], "age:", row["age"], "(type:", typeOf(row["age"]), ")")
}
// Output: Alice age: 30 (type: INT)

// Aggregate queries
var total = dbQueryCount(db, "SELECT COUNT(*) FROM employees")
var avgSalary = dbQueryFloat(db, "SELECT AVG(salary) FROM employees")

// Close connection
dbClose(db)
```

---

## Type Checking Functions

### isEmpty(value)

Returns true if value is empty string, empty array, empty map, or null.

```xxl
isEmpty("")          // true
isEmpty([])          // true
isEmpty({})          // true
isEmpty(null)        // true
isEmpty([1, 2])      // false
```

### isString(value)

Returns true if value is a string.

```xxl
isString("hello")    // true
isString(42)         // false
```

### isNumber(value)

Returns true if value is an integer or float.

```xxl
isNumber(42)         // true
isNumber(3.14)       // true
isNumber("42")       // false
```

### isInt(value)

Returns true if value is an integer.

```xxl
isInt(42)            // true
isInt(3.14)          // false
```

### isFloat(value)

Returns true if value is a float.

```xxl
isFloat(3.14)        // true
isFloat(42)          // false
```

### isBigInt(value)

Returns true if value is a BigInt.

```xxl
isBigInt(12345678901234567890n)  // true
isBigInt(42)                      // false
isBigInt(3.14)                    // false
```

### isBigFloat(value)

Returns true if value is a BigFloat.

```xxl
isBigFloat(3.14159265358979323846m)  // true
isBigFloat(3.14)                      // false
isBigFloat(42)                        // false
```

### isArray(value)

Returns true if value is an array.

```xxl
isArray([1, 2, 3])   // true
isArray("hello")     // false
```

### isMap(value)

Returns true if value is a map.

```xxl
isMap({"a": 1})      // true
isMap([1, 2])        // false
```

### isOrderedMap(value)

Returns true if value is an OrderedMap.

```xxl
isOrderedMap(newOrderedMap())  // true
isOrderedMap({"a": 1})         // false (regular Map)
isOrderedMap([1, 2])           // false
```

### isBool(value)

Returns true if value is a boolean.

```xxl
isBool(true)         // true
isBool(1)            // false
```

### isFunction(value)

Returns true if value is a function.

```xxl
isFunction(len)      // true
isFunction(42)       // false
```

### isNull(value)

Returns true if value is null.

```xxl
isNull(null)         // true
isNull(0)            // false
```

---

## Formatting Functions

### format(template, args...)

Formats a string using Go-style format specifiers.

```xxl
format("Hello %s, you are %d", "Alice", 30)  // "Hello Alice, you are 30"
format("Pi is %.2f", 3.14159)                 // "Pi is 3.14"
```

---

## BigInt/BigFloat Functions

Xxlang supports arbitrary precision numbers for high-precision calculations.

### bigInt(value)

Creates a BigInt from a string or integer.

```xxl
var big = bigInt("123456789012345678901234567890")
var fromInt = bigInt(42)
```

### bigFloat(value)

Creates a BigFloat from a string or float.

```xxl
var precise = bigFloat("3.141592653589793238462643383279")
var fromFloat = bigFloat(3.14)
```

### isBigInt(value)

Returns true if value is a BigInt.

```xxl
isBigInt(12345678901234567890n)  // true
isBigInt(42)                      // false
```

### isBigFloat(value)

Returns true if value is a BigFloat.

```xxl
isBigFloat(3.14159265358979323846m)  // true
isBigFloat(3.14)                      // false
```

**BigInt Literals:** Use the `n` suffix: `12345678901234567890n`

**BigFloat Literals:** Use the `m` suffix: `3.14159265358979323846m`

```xxl
// BigInt operations
var a = 1000000000000000000n
var b = 2000000000000000000n
pln(a + b)  // 3000000000000000000

// BigFloat operations (exact precision)
var pi = 3.14159265358979323846m
var e = 2.71828182845904523536m
pln(pi + e)  // Exact result, no floating-point errors
```

---

## Reader/Writer Functions

Functions for working with I/O readers and writers.

### getWebReader(url)

Returns a reader for streaming HTTP response body.

```xxl
var reader = getWebReader("https://example.com/large-file.zip")
```

### ioCopy(dst, src)

Copies data from a reader to a writer. Returns bytes copied and error.

```xxl
var bytesCopied, err = ioCopy(dstWriter, srcReader)
```

### isReader(value)

Returns true if value is a reader.

```xxl
isReader(reader)  // true
isReader("text")  // false
```

### isWriter(value)

Returns true if value is a writer.

```xxl
isWriter(writer)  // true
isWriter({})      // false
```

### newBytesReader(bytes)

Creates a reader from a byte array.

```xxl
var data = [72, 101, 108, 108, 111]  // "Hello"
var reader = newBytesReader(data)
```

### newStringReader(str)

Creates a reader from a string.

```xxl
var reader = newStringReader("Hello, World!")
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
| `scan(prompt?)` | Read a line from stdin | `io.scan("Enter name: ")` |
| `scanInt(prompt?)` | Read an integer from stdin | `io.scanInt("Enter age: ")` |
| `scanFloat(prompt?)` | Read a float from stdin | `io.scanFloat("Enter price: ")` |
| `scanBool(prompt?)` | Read a boolean from stdin | `io.scanBool("Continue? ")` |
| `scanN(n)` | Read n tokens from stdin | `io.scanN(3)` |
| `scanSplit(sep)` | Read line and split | `io.scanSplit(",")` |
| `scan2()` | Read two tokens | `a, b = io.scan2()` |
| `scan3()` | Read three tokens | `a, b, c = io.scan3()` |
| `scanf(format)` | Read with format | `io.scanf("{} {}")` |
| `newScanner(reader?)` | Create Scanner object | `io.newScanner()` |

#### Input/Scan Functions

Xxlang provides convenient functions for reading user input from stdin:

```xxl
import "io"

// Basic usage - read a line
var name = io.scan("Enter your name: ")
pln("Hello, " + name + "!")

// Read specific types
var age = io.scanInt("Enter your age: ")
var price = io.scanFloat("Enter price: ")
var confirmed = io.scanBool("Continue? (true/false): ")

// Read multiple values
var a, b = io.scan2()  // Read two whitespace-delimited tokens
var x, y, z = io.scan3()  // Read three tokens

// Read n values into array
var tokens = io.scanN(3)  // Returns ["token1", "token2", "token3"]

// Read and split by delimiter
var parts = io.scanSplit(",")  // Read line, split by comma

// Format-based reading
var values = io.scanf("{} {} {}")  // Read three space-separated values
```

#### Scanner Object

For more control over input reading, use the Scanner object:

```xxl
import "io"

// Create scanner from stdin
var scanner = io.newScanner()

// Or create from a reader
var reader = io.newReader("hello world\n42\n3.14")
var scanner2 = io.newScanner(reader)

// Read tokens
var token = scanner.next()       // Read whitespace-delimited token
var line = scanner.nextLine()    // Read entire line
var num = scanner.nextInt()      // Read integer
var f = scanner.nextFloat()      // Read float
var b = scanner.nextBool()       // Read boolean

// Check for more input
if (scanner.hasNext()) {
    pln("More input available")
}

// Skip current line
scanner.skipLine()

// Close when done
scanner.close()
```

**Scanner Methods:**

| Method | Description | Return Type |
|--------|-------------|-------------|
| `next()` | Read next whitespace-delimited token | STRING or NULL |
| `nextLine()` | Read next line | STRING or NULL |
| `nextInt()` | Read next token as integer | INT or ERROR |
| `nextFloat()` | Read next token as float | FLOAT or ERROR |
| `nextBool()` | Read next token as boolean | BOOL or ERROR |
| `hasNext()` | Check if more input available | BOOL |
| `skipLine()` | Skip current line | NULL |
| `close()` | Close the scanner | NULL |

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

Cryptographic functions including hashing, encoding, and encryption.

**Hash Functions:**

| Function | Description | Example |
|----------|-------------|---------|
| `md5(s)` | MD5 hash | `crypto.md5("hello")` |
| `sha1(s)` | SHA1 hash | `crypto.sha1("hello")` |
| `sha256(s)` | SHA256 hash | `crypto.sha256("hello")` |
| `sha512(s)` | SHA512 hash | `crypto.sha512("hello")` |

**Encoding Functions:**

| Function | Description | Example |
|----------|-------------|---------|
| `base64Encode(s)` | Base64 encode | `crypto.base64Encode("hello")` |
| `base64Decode(s)` | Base64 decode | `crypto.base64Decode("aGVsbG8=")` |
| `hexEncode(s)` | Hex encode | `crypto.hexEncode("hello")` |
| `hexDecode(s)` | Hex decode | `crypto.hexDecode("68656c6c6f")` |

**Encryption Functions (Charlang compatible):**

| Function | Description | Example |
|----------|-------------|---------|
| `encryptTextByTXTE(text, code)` | TXTE text encryption | `crypto.encryptTextByTXTE("hello", "key")` |
| `decryptTextByTXTE(hexStr, code)` | TXTE text decryption | `crypto.decryptTextByTXTE("...", "key")` |
| `encryptTextByTXDEE(text, code)` | TXDEE text encryption | `crypto.encryptTextByTXDEE("hello", "key")` |
| `decryptTextByTXDEE(hexStr, code)` | TXDEE text decryption | `crypto.decryptTextByTXDEE("...", "key")` |
| `encryptTextByTXDEF(text, code)` | TXDEF text encryption | `crypto.encryptTextByTXDEF("hello", "key")` |
| `decryptTextByTXDEF(hexStr, code)` | TXDEF text decryption | `crypto.decryptTextByTXDEF("...", "key")` |
| `encryptText(text, code)` | Default text encryption | `crypto.encryptText("hello", "key")` |
| `decryptText(hexStr, code)` | Default text decryption | `crypto.decryptText("...", "key")` |
| `encryptData(data, code)` | Default data encryption | `crypto.encryptData([1,2,3], "key")` |
| `decryptData(data, code)` | Default data decryption | `crypto.decryptData(encData, "key")` |
| `aesEncrypt(data, key, mode?)` | AES encryption | `crypto.aesEncrypt(data, "16bytekey1234567")` |
| `aesDecrypt(data, key, mode?)` | AES decryption | `crypto.aesDecrypt(encData, "16bytekey1234567")` |

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
| `forEach(arr, fn)` | Iterate elements | `array.forEach([1, 2, 3], fn(x) { pln(x) })` |

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

## System Command Functions

Xxlang provides built-in functions for executing system commands and starting applications.

### systemCmd(cmd, args...)

Executes a system command synchronously and returns the result as an OrderedMap.

**Parameters:**
- `cmd` (string): The command to execute
- `args...` (strings, optional): Additional arguments for the command

**Returns:** An OrderedMap with:
- `success` (bool): Whether the command succeeded (exit code 0)
- `exitCode` (int): The exit code of the command
- `output` (string): Combined stdout and stderr output
- `error` (string): Error message if any

```xxl
// Simple command
var r = systemCmd("echo hello")
pln(r["output"])     // "hello"
pln(r["success"])    // true
pln(r["exitCode"])   // 0

// Command with arguments
var r = systemCmd("dir", ".")
if (r["success"]) {
    pln("Output: ", r["output"])
}

// Handle errors
var r = systemCmd("nonexistent_command")
if (!r["success"]) {
    pln("Error: ", r["error"])
    pln("Exit code: ", r["exitCode"])
}
```

### systemCmdDetached(cmd, args...)

Executes a system command in detached (background) mode. The command runs asynchronously without waiting for completion.

**Parameters:**
- `cmd` (string): The command to execute
- `args...` (strings, optional): Additional arguments for the command

**Returns:** An OrderedMap with:
- `success` (bool): Whether the command was started successfully
- `pid` (int): Process ID (0 if not available)
- `error` (string): Error message if any

```xxl
// Start a background process
var r = systemCmdDetached("notepad")
pln(r["success"])    // true
pln(r["pid"])        // Process ID (e.g., 12345)

// Start a background script
var r = systemCmdDetached("python", "server.py")
if (r["success"]) {
    pln("Server started with PID: ", r["pid"])
}

// Handle errors
var r = systemCmdDetached("nonexistent_command")
if (!r["success"]) {
    pln("Failed to start: ", r["error"])
}
```

### systemStart(path, workingDir)

Opens a file or URL with the default application, or starts a program. This is equivalent to:
- Windows: `start "" path`
- macOS: `open path`
- Linux: `xdg-open path`

**Parameters:**
- `path` (string): The file path, URL, or program to open
- `workingDir` (string, optional): Working directory for the process

**Returns:** An OrderedMap with:
- `success` (bool): Whether the operation succeeded
- `error` (string): Error message if any

```xxl
// Open a URL in the default browser
var r = systemStart("https://www.example.com")
pln(r["success"])    // true

// Open a file with its default application
var r = systemStart("C:\\Documents\\report.pdf")
pln(r["success"])    // true

// Open a file in a specific working directory
var r = systemStart("readme.txt", "C:\\MyProject")

// Handle errors
if (!r["success"]) {
    pln("Failed to open: ", r["error"])
}
```

---

## Test Assertion Functions

These functions help with testing and debugging Xxlang code.

### testByText(actual, expected, testName?, testGroup?)

Tests that two strings are exactly equal.

**Parameters:**
- `actual` (string): The actual value
- `expected` (string): The expected value
- `testName` (string, optional): A name for the test
- `testGroup` (string, optional): A group identifier for the test

**Returns:** `null` on success, or an `ERROR` object with details on failure.

```xxl
// Basic test
testByText("hello", "hello")  // Prints: test 1 passed

// Test with custom name
testByText("hello", "hello", "greeting test")  // Prints: test greeting test passed

// Test with group
testByText("hello", "hello", "greeting", "group1")  // Prints: test greeting(group1) passed

// Failed test shows diff position
testByText("hello world", "hello")
// ERROR: test 2 failed at position 5:
// -----
// hello world
// -----
// hello
```

### testByStartsWith(str, prefix, testName?, testGroup?)

Tests that a string starts with a given prefix.

**Parameters:**
- `str` (string): The string to check
- `prefix` (string): The expected prefix
- `testName` (string, optional): A name for the test
- `testGroup` (string, optional): A group identifier for the test

**Returns:** `null` on success, or an `ERROR` object with details on failure.

```xxl
testByStartsWith("hello world", "hello")  // Prints: test 1 passed
testByStartsWith("hello world", "world")  // ERROR: string does not start with prefix
```

### testByEndsWith(str, suffix, testName?, testGroup?)

Tests that a string ends with a given suffix.

**Parameters:**
- `str` (string): The string to check
- `suffix` (string): The expected suffix
- `testName` (string, optional): A name for the test
- `testGroup` (string, optional): A group identifier for the test

**Returns:** `null` on success, or an `ERROR` object with details on failure.

```xxl
testByEndsWith("hello world", "world")  // Prints: test 1 passed
testByEndsWith("hello world", "hello")  // ERROR: string does not end with suffix
```

### testByContains(str, substr, testName?, testGroup?)

Tests that a string contains a given substring.

**Parameters:**
- `str` (string): The string to check
- `substr` (string): The substring to find
- `testName` (string, optional): A name for the test
- `testGroup` (string, optional): A group identifier for the test

**Returns:** `null` on success, or an `ERROR` object with details on failure.

```xxl
testByContains("hello world", "lo wo")  // Prints: test 1 passed
testByContains("hello world", "xyz")    // ERROR: string does not contain substring
```

### testByReg(str, pattern, testName?, testGroup?)

Tests that a string matches a regex pattern (full match).

**Parameters:**
- `str` (string): The string to check
- `pattern` (string): The regex pattern (must match the entire string)
- `testName` (string, optional): A name for the test
- `testGroup` (string, optional): A group identifier for the test

**Returns:** `null` on success, or an `ERROR` object with details on failure.

```xxl
testByReg("hello123", "hello[0-9]+")  // Prints: test 1 passed
testByReg("hello", "[0-9]+")          // ERROR: string does not match regex pattern
```

### testByRegContains(str, pattern, testName?, testGroup?)

Tests that a string contains a match for a regex pattern (partial match).

**Parameters:**
- `str` (string): The string to check
- `pattern` (string): The regex pattern to find in the string
- `testName` (string, optional): A name for the test
- `testGroup` (string, optional): A group identifier for the test

**Returns:** `null` on success, or an `ERROR` object with details on failure.

```xxl
testByRegContains("The price is $100", "\\$[0-9]+")  // Prints: test 1 passed
testByRegContains("hello world", "[0-9]+")           // ERROR: string does not contain regex match
```

### dumpVar(value)

Dumps a variable for debugging, showing its type and contents.

**Parameters:**
- `value` (any): The value to dump

**Returns:** `null` (output is printed to stdout)

```xxl
var testMap = {"name": "Alice", "age": 30}
dumpVar(testMap)
// Output:
// Dump: {age: 30, name: Alice}
// Type: MAP
// Contents:
//   name: Alice
//   age: 30

var testArr = [1, 2, 3, "four"]
dumpVar(testArr)
// Output:
// Dump: [1, 2, 3, four]
// Type: ARRAY
// Elements:
//   [0]: 1
//   [1]: 2
//   [2]: 3
//   [3]: four
```

### debugInfo(args...)

Returns debug information about the arguments passed.

**Parameters:**
- `args...` (any, optional): Values to include in debug info

**Returns:** A string with debug information.

```xxl
var info = debugInfo("test", 42)
pln(info)
// Output:
// === Debug Info ===
// Note: Full debug info requires VM context.
// Arguments:
//   [0]: test (type: STRING)
//   [1]: 42 (type: INT)
```

---

## Collection Functions (Batch 11)

High-order functions for array manipulation.

### mapArray(array, fn)

Applies a function to each element and returns a new array.

```xxl
func double(x) { return x * 2 }
var arr = [1, 2, 3, 4, 5]
var doubled = mapArray(arr, double)
// [2, 4, 6, 8, 10]
```

### filterArray(array, fn)

Filters array elements that satisfy the predicate function.

```xxl
func isEven(x) { return x % 2 == 0 }
var arr = [1, 2, 3, 4, 5]
var evens = filterArray(arr, isEven)
// [2, 4]
```

### reduceArray(array, fn, initial?)

Reduces array to a single value using an accumulator function.

```xxl
func add(a, b) { return a + b }
var arr = [1, 2, 3, 4, 5]
var sum = reduceArray(arr, add, 0)
// 15
```

### forEach(array, fn)

Iterates over array elements, calling the function for each.

```xxl
forEach([1, 2, 3], func(x, i) {
    pln("Index:", i, "Value:", x)
})
```

### flatMap(array, fn)

Maps each element and flattens the result.

```xxl
func duplicate(x) { return [x, x] }
var result = flatMap([1, 2, 3], duplicate)
// [1, 1, 2, 2, 3, 3]
```

### every(array, fn)

Returns true if all elements satisfy the predicate.

```xxl
every([2, 4, 6], func(x) { return x % 2 == 0 })  // true
every([2, 3, 4], func(x) { return x % 2 == 0 })  // false
```

### some(array, fn)

Returns true if any element satisfies the predicate.

```xxl
some([1, 3, 5], func(x) { return x > 4 })  // true
some([1, 3, 5], func(x) { return x > 6 })  // false
```

### groupBy(array, fn)

Groups array elements by the key returned from the function.

```xxl
var people = [
    {"name": "Alice", "age": 30},
    {"name": "Bob", "age": 25},
    {"name": "Charlie", "age": 30}
]
var grouped = groupBy(people, func(p) { return p["age"] })
// {"25": [...], "30": [...]}
```

### partition(array, fn)

Splits array into two groups based on predicate.

```xxl
var nums = [1, 2, 3, 4, 5, 6]
var result = partition(nums, func(x) { return x % 2 == 0 })
// [[2, 4, 6], [1, 3, 5]]
```

### zip(array1, array2, ...)

Combines multiple arrays into an array of tuples.

```xxl
var names = ["Alice", "Bob"]
var ages = [30, 25]
zip(names, ages)
// [["Alice", 30], ["Bob", 25]]
```

### unzip(array)

Splits an array of tuples into separate arrays.

```xxl
unzip([["Alice", 30], ["Bob", 25]])
// [["Alice", "Bob"], [30, 25]]
```

### fill(array, value, start?, end?)

Fills array elements with a value in the specified range.

```xxl
fill([1, 2, 3, 4, 5], 0)        // [0, 0, 0, 0, 0]
fill([1, 2, 3, 4, 5], 0, 1, 3)   // [1, 0, 0, 4, 5]
```

### rangeNum(end) / rangeNum(start, end, step?)

Generates an array of numbers in a range.

```xxl
rangeNum(5)           // [0, 1, 2, 3, 4]
rangeNum(1, 5)        // [1, 2, 3, 4]
rangeNum(0, 10, 2)    // [0, 2, 4, 6, 8]
```

### intersection(arr1, arr2)

Returns elements common to both arrays.

```xxl
intersection([1, 2, 3], [2, 3, 4])
// [2, 3]
```

### difference(arr1, arr2)

Returns elements in arr1 but not in arr2.

```xxl
difference([1, 2, 3], [2, 3, 4])
// [1]
```

### union(arr1, arr2)

Returns unique elements from both arrays.

```xxl
union([1, 2, 3], [2, 3, 4])
// [1, 2, 3, 4]
```

### countBy(array, fn)

Counts elements grouped by the key function.

```xxl
countBy([1, 2, 3, 4, 5], func(x) {
    return x % 2 == 0 ? "even" : "odd"
})
// {"even": 2, "odd": 3}
```

### sortBy(array, fn)

Sorts array by the value returned from the function.

```xxl
var users = [{"name": "Bob"}, {"name": "Alice"}]
sortBy(users, func(u) { return u["name"] })
// [{"name": "Alice"}, {"name": "Bob"}]
```

---

## Utility Functions (Batch 12)

General-purpose utility functions.

### sprintf(format, args...)

Formats a string and returns the result.

```xxl
sprintf("Hello %s, you are %d years old", "Alice", 25)
// "Hello Alice, you are 25 years old"
```

### toBool(value)

Converts a value to boolean.

```xxl
toBool(0)      // false
toBool(1)      // true
toBool("")     // false
toBool("hello") // true
```

### toInt(value, base?)

Converts a value to integer.

```xxl
toInt("123")       // 123
toInt("ff", 16)     // 255
toInt(3.14)         // 3
```

### toFloat(value)

Converts a value to float.

```xxl
toFloat("3.14")   // 3.14
toFloat(42)        // 42.0
```

### isUndefined(value) / isNull(value)

Returns true if value is null/undefined.

```xxl
isNull(null)     // true
isNull(0)        // false
```

### isCallable(value)

Returns true if value is a callable function.

```xxl
isCallable(len)         // true (builtin)
isCallable(func(){})    // true (user function)
isCallable([1,2,3])     // false
```

### isIterable(value)

Returns true if value can be iterated.

```xxl
isIterable([1,2,3])   // true
isIterable("hello")   // true
isIterable(123)       // false
```

### isError(value)

Returns true if value is an error object.

```xxl
isError(error("test"))  // true
isError("hello")        // false
```

### error(message)

Creates an error object.

```xxl
var e = error("something went wrong")
pln(e)  // ERROR: something went wrong
```

### getErrStr(value)

Extracts error message from error object.

```xxl
getErrStr(error("failed"))  // "failed"
```

### isErrStr(value)

Returns true if string is an error message.

```xxl
isErrStr("ERROR: failed")   // true
isErrStr("hello")           // false
```

### typeCode(value)

Returns the type code of an object.

```xxl
typeCode(123)       // 1 (INT)
typeCode("hello")   // 3 (STRING)
typeCode([1,2,3])   // 6 (ARRAY)
```

### clone(value)

Creates a deep copy of an object.

```xxl
var arr = [1, 2, 3]
var arr2 = clone(arr)
arr2[0] = 99
pln(arr[0])   // 1 (original unchanged)
```

### swap(array, i, j)

Returns a new array with elements at indices i and j swapped.

```xxl
swap([1, 2, 3, 4, 5], 0, 4)
// [5, 2, 3, 4, 1]
```

### coalesce(values...)

Returns the first non-null/non-error value.

```xxl
coalesce(null, null, "first", "second")
// "first"
```

### defaultVal(value, default)

Returns value if not null/error, otherwise returns default.

```xxl
defaultVal(null, "default")     // "default"
defaultVal("actual", "default") // "actual"
```

---

## String Processing Functions (Batch 13)

Enhanced string manipulation functions.

### strSplitLines(str)

Splits a string into lines.

```xxl
strSplitLines("a\nb\nc")
// ["a", "b", "c"]
```

### strContainsAny(str, chars)

Returns true if string contains any of the specified characters.

```xxl
strContainsAny("hello", "aeiou")  // true
strContainsAny("xyz", "aeiou")     // false
```

### strIndex(str, substr)

Returns the index of substring, or -1 if not found.

```xxl
strIndex("hello world", "world")  // 6
strIndex("hello", "xyz")           // -1
```

### strLastIndex(str, substr)

Returns the last index of substring.

```xxl
strLastIndex("hello hello", "hello")  // 6
```

### strSplitN(str, sep, n)

Splits string with a limit on number of parts.

```xxl
strSplitN("a,b,c,d", ",", 2)
// ["a", "b,c,d"]
```

### strPad(str, length, padStr?, padRight?)

Pads string to specified length.

```xxl
strPad("abc", 10)              // "       abc"
strPad("abc", 10, "-")         // "-------abc"
strPad("abc", 10, "-", true)   // "abc-------"
```

### strSub(str, start, end?)

Extracts substring with Unicode support.

```xxl
strSub("hello world", 0, 5)  // "hello"
strSub("hello world", -5)     // "world"
```

### intToStr(n, base?)

Converts integer to string.

```xxl
intToStr(123)       // "123"
intToStr(255, 16)   // "ff"
```

### floatToStr(f, precision?)

Converts float to string.

```xxl
floatToStr(3.14159)     // "3.14159"
floatToStr(3.14159, 2)  // "3.14"
```

### charCode(str, index?)

Returns the Unicode code point of a character.

```xxl
charCode("ABC")      // 65
charCode("ABC", 1)   // 66
```

### charFromCode(code)

Creates a character from a Unicode code point.

```xxl
charFromCode(65)   // "A"
charFromCode(20013) // "中"
```

### reverseMap(map)

Returns a new map with keys and values swapped.

```xxl
reverseMap({"a": "1", "b": "2"})
// {"1": "a", "2": "b"}
```

### simpleStrToMap(str, sep1?, sep2?)

Parses a simple string to a map.

```xxl
simpleStrToMap("a=1,b=2,c=3")
// {"a": "1", "b": "2", "c": "3"}
```

### mapToStr(map, sep1?, sep2?)

Converts a map to a simple string.

```xxl
mapToStr({"a": "1", "b": "2"})
// "a=1,b=2"
```

### Bitwise Functions

```xxl
bitNot(5)           // -6
bitAnd(15, 7)       // 7
bitOr(8, 3)         // 11
bitXor(12, 10)      // 6
bitShiftLeft(4, 2)   // 16
bitShiftRight(16, 2) // 4
```

---

## Check/Validation Functions (Batch 14)

Functions for checking and validating values.

### isNil(value) / isNull(value)

Returns true if value is null.

```xxl
isNull(null)  // true
isNull(0)     // false
```

### isNilOrEmpty(value)

Returns true if value is null or empty.

```xxl
isNilOrEmpty(null)    // true
isNilOrEmpty("")      // true
isNilOrEmpty([])      // true
isNilOrEmpty("hello") // false
```

### isNilOrErr(value)

Returns true if value is null or an error.

```xxl
isNilOrErr(null)            // true
isNilOrErr(error("test"))   // true
isNilOrErr(123)             // false
```

### isBytes(value)

Returns true if value is a bytes object.

```xxl
isBytes(bytes(1,2,3))  // false (returns array)
isBytes([1,2,3])       // false
```

### isChars(value)

Returns true if value is a chars object.

```xxl
isChars(toChars("hello"))  // true
isChars("hello")           // false
```

### pass()

Does nothing and returns null. Useful as a placeholder.

```xxl
pass()  // null
```

### errStrf(format, args...)

Creates a formatted error string.

```xxl
errStrf("failed: %s", "timeout")
// "ERROR: failed: timeout"
```

### errf(format, args...)

Creates a formatted error object.

```xxl
var e = errf("error: %d", 123)
isError(e)  // true
```

### errToEmpty(value)

Converts error to empty string, passes through other values.

```xxl
errToEmpty(error("test"))   // ""
errToEmpty("hello")         // "hello"
errToEmpty("ERROR: bad")    // ""
```

### sscanf(str, format)

Parses string according to format.

```xxl
sscanf("name:Alice age:25", "name:%s age:%d")
// ["Alice", 25]
```

### bytesStartsWith(data, prefix)

Checks if bytes data starts with prefix.

```xxl
bytesStartsWith(bytes(72,101,108,108,111), bytes(72,101))
// true
```

### bytesEndsWith(data, suffix)

Checks if bytes data ends with suffix.

```xxl
bytesEndsWith(bytes(72,101,108,108,111), bytes(108,111))
// true
```

### bytesContains(data, sub)

Checks if bytes data contains sub.

```xxl
bytesContains(bytes(1,2,3,4,5), bytes(3,4))
// true
```

### bytesIndex(data, sub)

Returns index of sub in bytes data.

```xxl
bytesIndex(bytes(1,2,3,4,5), bytes(3,4))
// 2
```

### compareBytes(a, b)

Compares two byte arrays. Returns -1, 0, or 1.

```xxl
compareBytes("abc", "abd")   // -1
compareBytes("abc", "abc")   // 0
compareBytes("abd", "abc")   // 1
```

### compareText(a, b)

Compares two text values. Returns -1, 0, or 1.

```xxl
compareText(123, 124)   // -1
compareText("a", "a")   // 0
```

---

## Random/Temp/URL Functions (Batch 15)

Random number generation, temporary files, and URL utilities.

### getRandomInt(max) / getRandomInt(min, max)

Returns a random integer.

```xxl
getRandomInt(100)      // 0-99
getRandomInt(50, 100)  // 50-100
```

### getRandomFloat()

Returns a random float between 0 and 1.

```xxl
getRandomFloat()  // 0.0 <= x < 1.0
```

### getRandomStr(length, charset?)

Generates a random string.

```xxl
getRandomStr(16)                     // "aB3dE5fG7hI9jK1m"
getRandomStr(8, "0123456789")         // "12345678"
```

### createTempDir(dir?, pattern?)

Creates a temporary directory.

```xxl
var tmpDir = createTempDir()
// "/tmp/xxlang_123456"
```

### createTempFile(dir?, pattern?)

Creates a temporary file.

```xxl
var tmpFile = createTempFile()
// "/tmp/xxlang_789012"
```

### changeDir(path)

Changes the current working directory.

```xxl
changeDir("/home/user")
```

### lookPath(name)

Finds an executable in PATH.

```xxl
lookPath("go")   // "/usr/local/go/bin/go"
lookPath("xyz")  // null
```

### joinUrlPath(base, elements...)

Joins URL path components.

```xxl
joinUrlPath("https://example.com", "api", "v1", "users")
// "https://example.com/api/v1/users"
```

### parseUrl(urlStr)

Parses a URL into its components.

```xxl
var u = parseUrl("https://example.com/path?q=hello#frag")
u["scheme"]    // "https"
u["host"]      // "example.com"
u["path"]      // "/path"
u["rawQuery"]  // "q=hello"
u["fragment"]  // "frag"
```

### parseQuery(queryStr)

Parses a URL query string.

```xxl
var q = parseQuery("a=1&b=2&c=3")
q["a"]  // "1"
q["b"]  // "2"
```

### isHttps(urlStr)

Returns true if URL uses HTTPS.

```xxl
isHttps("https://example.com")  // true
isHttps("http://example.com")   // false
```

### genToken(length?)

Generates a random token.

```xxl
genToken()      // 32-character base64 token
genToken(16)    // 16-byte token
```

### genOtpCode(secret, digits?)

Generates a simple OTP code.

```xxl
var code = genOtpCode("mysecret")
// "502718"
```

### checkOtpCode(secret, code, digits?)

Validates an OTP code.

```xxl
checkOtpCode("mysecret", "502718")
// true
```

---

## Collection Functions

### mapArray(arr, fn)

Applies a function to each element and returns a new array.

```xxl
mapArray([1, 2, 3], func(x) { return x * 2 })  // [2, 4, 6]
```

### filterArray(arr, fn)

Filters array elements that satisfy the condition.

```xxl
filterArray([1, 2, 3, 4, 5], func(x) { return x > 2 })  // [3, 4, 5]
```

### reduceArray(arr, fn, initial?)

Reduces array to a single value.

```xxl
reduceArray([1, 2, 3, 4], func(acc, x) { return acc + x })      // 10
reduceArray([1, 2, 3, 4], func(acc, x) { return acc + x }, 0)   // 10
```

### forEach(arr, fn)

Iterates over array elements.

```xxl
forEach([1, 2, 3], func(x, i) { pln(i, ":", x) })
```

### flatMap(arr, fn)

Maps then flattens the result.

```xxl
flatMap([1, 2, 3], func(x) { return [x, x * 2] })  // [1, 2, 2, 4, 3, 6]
```

### every(arr, fn)

Checks if all elements satisfy the condition.

```xxl
every([2, 4, 6], func(x) { return x % 2 == 0 })  // true
every([2, 3, 4], func(x) { return x % 2 == 0 })  // false
```

### some(arr, fn)

Checks if any element satisfies the condition.

```xxl
some([1, 3, 5], func(x) { return x % 2 == 0 })  // false
some([1, 2, 3], func(x) { return x % 2 == 0 })  // true
```

### groupBy(arr, fn)

Groups array elements by key function.

```xxl
var users = [{"name": "Alice", "dept": "Eng"}, {"name": "Bob", "dept": "Sales"}]
groupBy(users, func(u) { return u["dept"] })
// {"Eng": [...], "Sales": [...]}
```

### partition(arr, fn)

Splits array into two groups by condition.

```xxl
partition([1, 2, 3, 4, 5], func(x) { return x > 2 })
// [[3, 4, 5], [1, 2]]
```

### zip(arr1, arr2, ...)

Combines multiple arrays into array of pairs.

```xxl
zip([1, 2, 3], ["a", "b", "c"])  // [[1, "a"], [2, "b"], [3, "c"]]
```

### unzip(arr)

Splits array of pairs into separate arrays.

```xxl
unzip([[1, "a"], [2, "b"]])  // [[1, 2], ["a", "b"]]
```

### fill(arr, value, start?, end?)

Fills array range with value.

```xxl
fill([1, 2, 3, 4], 0)         // [0, 0, 0, 0]
fill([1, 2, 3, 4], 0, 1, 3)   // [1, 0, 0, 4]
```

### rangeNum(end) or rangeNum(start, end, step?)

Generates a range of numbers (exclusive end).

```xxl
rangeNum(5)           // [0, 1, 2, 3, 4]
rangeNum(1, 5)        // [1, 2, 3, 4]
rangeNum(0, 10, 2)    // [0, 2, 4, 6, 8]
```

### intersection(arr1, arr2)

Finds intersection of two arrays.

```xxl
intersection([1, 2, 3], [2, 3, 4])  // [2, 3]
```

### difference(arr1, arr2)

Finds elements in arr1 but not in arr2.

```xxl
difference([1, 2, 3], [2, 3, 4])  // [1]
```

### union(arr1, arr2)

Returns union of two arrays (unique elements).

```xxl
union([1, 2], [2, 3])  // [1, 2, 3]
```

### countBy(arr, fn)

Counts elements by key function.

```xxl
countBy([1, 2, 3, 4, 5], func(x) { return x % 2 == 0 ? "even" : "odd" })
// {"even": 2, "odd": 3}
```

### sortBy(arr, fn)

Sorts array by key function.

```xxl
sortBy([3, 1, 2], func(x) { return x })  // [1, 2, 3]
sortBy(["banana", "apple"], func(x) { return len(x) })  // ["apple", "banana"]
```

---

## String Processing Functions

### strSplitLines(str)

Splits string by lines.

```xxl
strSplitLines("line1\nline2\nline3")  // ["line1", "line2", "line3"]
```

### strContainsAny(str, chars)

Checks if string contains any of the characters.

```xxl
strContainsAny("hello", "aeiou")  // true (contains 'e' and 'o')
```

### strIndex(str, substr)

Finds index of substring.

```xxl
strIndex("hello world", "world")  // 6
```

### strLastIndex(str, substr)

Finds last index of substring.

```xxl
strLastIndex("hello hello", "hello")  // 6
```

### strSplitN(str, sep, n)

Splits string with limit.

```xxl
strSplitN("a,b,c,d", ",", 2)  // ["a", "b,c,d"]
```

### strPad(str, length, padStr?, padRight?)

Pads string to specified length.

```xxl
strPad("5", 5)              // "    5"
strPad("5", 5, "0")         // "00005"
strPad("5", 5, "0", true)   // "50000"
```

### strSub(str, start, end?)

Gets substring (character-aware).

```xxl
strSub("hello world", 0, 5)  // "hello"
strSub("你好世界", 1, 3)       // "好世"
```

### intToStr(n, base?)

Converts integer to string.

```xxl
intToStr(42)      // "42"
intToStr(255, 16) // "ff"
```

### floatToStr(f, prec?)

Converts float to string.

```xxl
floatToStr(3.14159)     // "3.14159"
floatToStr(3.14159, 2)  // "3.14"
```

### charCode(str, index?)

Gets character code (Unicode code point).

```xxl
charCode("A")      // 65
charCode("你好", 0)  // 20320
```

### charFromCode(code)

Creates character from code point.

```xxl
charFromCode(65)     // "A"
charFromCode(20320)  // "你"
```

### reverseMap(map)

Reverses map keys and values.

```xxl
reverseMap({"a": 1, "b": 2})  // {1: "a", 2: "b"}
```

### simpleStrToMap(str, sep1?, sep2?)

Parses simple string to map.

```xxl
simpleStrToMap("a=1,b=2")            // {"a": "1", "b": "2"}
simpleStrToMap("a:1;b:2", ";", ":")  // {"a": "1", "b": "2"}
```

### mapToStr(map, sep1?, sep2?)

Converts map to simple string.

```xxl
mapToStr({"a": 1, "b": 2})  // "a=1,b=2"
```

---

## Bitwise Functions

### bitNot(n)

Bitwise NOT.

```xxl
bitNot(5)  // -6
```

### bitAnd(a, b)

Bitwise AND.

```xxl
bitAnd(5, 3)  // 1
```

### bitOr(a, b)

Bitwise OR.

```xxl
bitOr(5, 3)  // 7
```

### bitXor(a, b)

Bitwise XOR.

```xxl
bitXor(5, 3)  // 6
```

### bitShiftLeft(n, shift)

Bitwise left shift.

```xxl
bitShiftLeft(1, 4)  // 16
```

### bitShiftRight(n, shift)

Bitwise right shift.

```xxl
bitShiftRight(16, 2)  // 4
```

---

## Check/Validation Functions

### isNil(value) / isNull(value)

Checks if value is null.

```xxl
isNil(null)   // true
isNil(0)      // false
```

### isNilOrEmpty(value)

Checks if value is null or empty.

```xxl
isNilOrEmpty(null)   // true
isNilOrEmpty("")     // true
isNilOrEmpty([])     // true
isNilOrEmpty([1])    // false
```

### isNilOrErr(value)

Checks if value is null or error.

```xxl
isNilOrErr(null)         // true
isNilOrErr(error("x"))   // true
```

### isBytes(value)

Checks if value is bytes.

```xxl
isBytes(bytes("hello"))  // true
```

### isChars(value)

Checks if value is chars.

```xxl
isChars(toChars("hello"))  // true
```

### isUndefined(value)

Checks if value is undefined/null.

```xxl
isUndefined(null)  // true
```

### isCallable(value)

Checks if value is callable.

```xxl
isCallable(func() {})  // true
isCallable(len)        // true
```

### isIterable(value)

Checks if value is iterable.

```xxl
isIterable([1, 2, 3])  // true
isIterable("hello")    // true
```

### isError(value)

Checks if value is an error.

```xxl
isError(error("failed"))  // true
```

### error(message)

Creates an error object.

```xxl
var e = error("something went wrong")
```

### getErrStr(value)

Gets error string from error object.

```xxl
getErrStr(error("failed"))  // "failed"
```

### isErrStr(value)

Checks if string is an error message.

```xxl
isErrStr("ERROR: failed")  // true
```

### typeCode(value)

Gets type code of an object.

```xxl
typeCode(42)     // 2 (INT)
typeCode("hi")   // 4 (STRING)
```

### pass()

Does nothing and returns null.

```xxl
pass()  // null
```

### errStrf(format, args...)

Formats error string.

```xxl
errStrf("file not found: %s", "test.txt")  // "ERROR: file not found: test.txt"
```

### errf(format, args...)

Creates formatted error.

```xxl
errf("invalid value: %d", 42)  // Error{Message: "invalid value: 42"}
```

### errToEmpty(value)

Converts error to empty string.

```xxl
errToEmpty(error("x"))  // ""
errToEmpty("hello")     // "hello"
```

### sscanf(str, format)

Parses string according to format.

```xxl
sscanf("hello 42", "hello %d")  // [42]
```

### coalesce(val1, val2, ...)

Returns first non-null/non-error value.

```xxl
coalesce(null, null, "found")  // "found"
```

### defaultVal(value, default)

Returns default value if null or error.

```xxl
defaultVal(null, "default")     // "default"
defaultVal("value", "default")  // "value"
```

---

## Bytes Functions

### bytesStartsWith(data, prefix)

Checks if bytes starts with prefix.

```xxl
bytesStartsWith(bytes("hello"), "he")  // true
```

### bytesEndsWith(data, suffix)

Checks if bytes ends with suffix.

```xxl
bytesEndsWith(bytes("hello"), "lo")  // true
```

### bytesContains(data, sub)

Checks if bytes contains substring.

```xxl
bytesContains(bytes("hello world"), "world")  // true
```

### bytesIndex(data, sub)

Finds index of bytes in bytes.

```xxl
bytesIndex(bytes("hello"), "ll")  // 2
```

### compareBytes(a, b)

Compares two byte arrays. Returns -1, 0, or 1.

```xxl
compareBytes("abc", "abd")  // -1
compareBytes("abc", "abc")  // 0
```

### compareText(a, b)

Compares two text values.

```xxl
compareText("abc", "abd")  // -1
```

---

## Miscellaneous Functions

### getRandomInt(min, max) or getRandomInt(max)

Returns random integer in range [min, max] (inclusive).

```xxl
getRandomInt(10)      // 0-10
getRandomInt(1, 6)    // 1-6 (dice roll)
```

### getRandomFloat()

Returns random float in [0, 1).

```xxl
getRandomFloat()  // e.g., 0.723456
```

### getRandomStr(length, charset?)

Generates random string.

```xxl
getRandomStr(8)                    // "aB3dEfGh"
getRandomStr(4, "0123456789")      // "5273"
```

### createTempDir(dir?, pattern?)

Creates temporary directory.

```xxl
createTempDir()                    // "/tmp/xxlang_123456"
createTempDir("/mydir", "test_*")  // "/mydir/test_789"
```

### createTempFile(dir?, pattern?)

Creates temporary file.

```xxl
createTempFile()  // "/tmp/xxlang_123456"
```

### changeDir(path)

Changes current working directory.

```xxl
changeDir("/home/user")
```

### lookPath(name)

Finds executable file in PATH.

```xxl
lookPath("python")  // "/usr/bin/python"
```

### joinUrlPath(base, elem...)

Joins URL path components.

```xxl
joinUrlPath("https://example.com", "api", "users")
// "https://example.com/api/users"
```

### parseUrl(urlStr)

Parses URL and returns map.

```xxl
parseUrl("https://user:pass@example.com:8080/path?q=1")
// {"scheme": "https", "host": "example.com:8080", ...}
```

### parseQuery(queryStr)

Parses URL query string.

```xxl
parseQuery("a=1&b=2")  // {"a": "1", "b": "2"}
```

### isHttps(urlStr)

Checks if URL is HTTPS.

```xxl
isHttps("https://example.com")  // true
```

### genToken(length?)

Generates random token.

```xxl
genToken()      // 32-byte base64 token
genToken(16)    // 16-byte token
```

### genOtpCode(secret, digits?)

Generates simple OTP code.

```xxl
genOtpCode("mysecret")  // "502718"
```

### checkOtpCode(secret, code, digits?)

Validates OTP code.

```xxl
checkOtpCode("mysecret", "502718")  // true
```

---

## See Also

- [Language Reference](LANGUAGE.md) - Complete language syntax
- [Standard Library](STDLIB.md) - Standard library overview
- [Embedding Guide](EMBEDDING.md) - Using Xxlang in Go applications
