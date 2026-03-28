// examples/test_stdlib_main.go - Test the new stdlib modules
package main

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/interpreter"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("Xxlang Standard Library Test")
	fmt.Println("==============================================")
	fmt.Println()

	interp := interpreter.New(interpreter.WithStdlib())

	// Test os
	fmt.Println("1. os module")
	fmt.Println("-------------------------------------------")
	testOS := `
import "os"

println("Platform: " + os.platform())
println("Arch: " + os.arch())
println("CPUs: " + os.cpus().toStr())
println("Home: " + os.home())
println("Temp: " + os.temp())
println("Hostname: " + os.hostname())
println("")

// Path operations
println("Path join: " + os.join("a", "b", "c.txt"))
println("Path base: " + os.base("/path/to/file.txt"))
println("Path dir: " + os.dir("/path/to/file.txt"))
println("Path ext: " + os.ext("/path/to/file.txt"))
println("Is absolute: " + os.isAbs("/path/to/file"))
`
	_, err := interp.Eval(testOS)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test encoding
	fmt.Println()
	fmt.Println("2. encoding module")
	fmt.Println("-------------------------------------------")
	testEncoding := `
import "encoding"

var original = "Hello, World!"
var b64 = encoding.base64Encode(original)
println("Base64 encode: " + b64)
println("Base64 decode: " + encoding.base64Decode(b64))

var hexEnc = encoding.hexEncode(original)
println("Hex encode: " + hexEnc)
println("Hex decode: " + encoding.hexDecode(hexEnc))

var urlEnc = encoding.urlEncode("hello world&key=value")
println("URL encode: " + urlEnc)
println("URL decode: " + encoding.urlDecode(urlEnc))

var urlParts = encoding.parseURL("https://example.com/path?q=test#section")
println("Parse URL: scheme=" + urlParts[0] + " host=" + urlParts[1] + " path=" + urlParts[2])
`
	_, err = interp.Eval(testEncoding)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test uuid
	fmt.Println()
	fmt.Println("3. uuid module")
	fmt.Println("-------------------------------------------")
	testUUID := `
import "uuid"

println("UUID v4: " + uuid.v4())
println("UUID v4 short: " + uuid.v4Short())
println("Simple ID: " + uuid.simple())
println("Time ID: " + uuid.timeID())
println("Random string: " + uuid.random(16))
println("Hex string: " + uuid.hex(8))
println("Is valid UUID: " + uuid.isValid("550e8400-e29b-41d4-a716-446655440000"))
`
	_, err = interp.Eval(testUUID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test strconv
	fmt.Println()
	fmt.Println("4. strconv module")
	fmt.Println("-------------------------------------------")
	testStrconv := `
import "strconv"

// Type conversions
println("parseInt('42'): " + strconv.parseInt("42").toStr())
println("parseFloat('3.14'): " + strconv.parseFloat("3.14").toStr())
println("parseBool('true'): " + strconv.parseBool("true").toStr())

// Format
println("formatInt(255, 16): " + strconv.formatInt(255, 16))
println("formatFloat(3.14159, 2): " + strconv.formatFloat(3.14159, 2))

// Type helpers
println("toInt('42'): " + strconv.toInt("42").toStr())
println("toFloat(42): " + strconv.toFloat(42).toStr())
println("toBool(1): " + strconv.toBool(1).toStr())
println("toString(123): " + strconv.toString(123))

// Format helpers
println("formatBytes(1536): " + strconv.formatBytes(1536))
println("formatBytes(1048576): " + strconv.formatBytes(1048576))
println("formatDuration(3661500): " + strconv.formatDuration(3661500))
`
	_, err = interp.Eval(testStrconv)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test math
	fmt.Println()
	fmt.Println("5. math module")
	fmt.Println("-------------------------------------------")
	testMath := `
import "math"

println("PI: " + math.PI)
println("E: " + math.E)
println("sin(0): " + math.sin(0))
println("cos(0): " + math.cos(0))
println("sqrt(16): " + math.sqrt(16))
println("pow(2, 10): " + math.pow(2, 10))
println("abs(-42): " + math.abs(-42))
println("floor(3.7): " + math.floor(3.7))
println("ceil(3.2): " + math.ceil(3.2))
println("round(3.5): " + math.round(3.5))
println("min(3, 5): " + math.min(3, 5))
println("max(3, 5): " + math.max(3, 5))
`
	_, err = interp.Eval(testMath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test net (basic test without actual network calls)
	fmt.Println()
	fmt.Println("6. net module")
	fmt.Println("-------------------------------------------")
	testNet := `
import "net"

println("isOK(200): " + net.isOK(200))
println("isOK(404): " + net.isOK(404))
println("isRedirect(301): " + net.isRedirect(301))
println("isClientError(400): " + net.isClientError(400))
println("isServerError(500): " + net.isServerError(500))
`
	_, err = interp.Eval(testNet)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test io
	fmt.Println()
	fmt.Println("7. io module")
	fmt.Println("-------------------------------------------")
	testIO := `
import "io"

println("Current dir: " + io.cwd())
var args = io.args()
println("Args count: " + args.len().toStr())

var home = io.env("HOME")
println("HOME env: " + home)

var path = io.env("PATH")
println("PATH exists: " + (path != "" && path != null).toStr())
`
	_, err = interp.Eval(testIO)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test string
	fmt.Println()
	fmt.Println("8. strings module")
	fmt.Println("-------------------------------------------")
	testString := `
import "strings"

var s = "Hello, World!"
println("len: " + strings.len(s))
println("toUpper: " + strings.toUpper(s))
println("toLower: " + strings.toLower(s))
println("contains: " + strings.contains(s, "World"))
println("indexOf: " + strings.indexOf(s, "World"))
println("hasPrefix: " + strings.hasPrefix(s, "Hello"))
println("hasSuffix: " + strings.hasSuffix(s, "!"))
println("split: " + strings.split("a,b,c", ","))
println("repeat: " + strings.repeat("ab", 3))
println("reverse: " + strings.reverse("hello"))
`
	_, err = interp.Eval(testString)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Test array
	fmt.Println()
	fmt.Println("9. array module")
	fmt.Println("-------------------------------------------")
	testArray := `
import "array"

var arr = [3, 1, 4, 1, 5, 9, 2, 6]
println("len: " + array.len(arr))
println("first: " + array.first(arr))
println("last: " + array.last(arr))
println("contains: " + array.contains(arr, 5))
println("indexOf: " + array.indexOf(arr, 5))
println("slice: " + array.slice(arr, 2, 5))
println("unique: " + array.unique([1, 2, 2, 3, 3, 3]))
println("isEmpty: " + array.isEmpty([]))
`
	_, err = interp.Eval(testArray)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("All tests completed!")
	fmt.Println("==============================================")
}
