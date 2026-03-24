// pkg/stdlib/io.go
// I/O utilities for the Xxlang standard library.
package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "io",
		Exports: map[string]objects.Object{
			"print": BuiltinFunc(func(args ...objects.Object) objects.Object {
				for _, arg := range args {
					fmt.Print(arg.Inspect())
				}
				return Null()
			}),

			"println": BuiltinFunc(func(args ...objects.Object) objects.Object {
				for i, arg := range args {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Print(arg.Inspect())
				}
				fmt.Println()
				return Null()
			}),

			"printf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("printf() takes at least 1 argument")
				}
				format, ok := args[0].(*objects.String)
				if !ok {
					return Error("printf() requires a format string")
				}
				// Convert remaining args to interface{} for fmt.Printf
				ifaces := make([]interface{}, len(args)-1)
				for i, arg := range args[1:] {
					switch a := arg.(type) {
					case *objects.Int:
						ifaces[i] = a.Value
					case *objects.Float:
						ifaces[i] = a.Value
					case *objects.String:
						ifaces[i] = a.Value
					case *objects.Bool:
						ifaces[i] = a.Value
					default:
						ifaces[i] = a.Inspect()
					}
				}
				fmt.Printf(format.Value, ifaces...)
				return Null()
			}),

			"readLine": BuiltinFunc(func(args ...objects.Object) objects.Object {
				reader := bufio.NewReader(os.Stdin)
				line, err := reader.ReadString('\n')
				if err != nil {
					return Error(err.Error())
				}
				// Remove trailing newline
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}
				if len(line) > 0 && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				return String(line)
			}),

			"readFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readFile() takes exactly 1 argument")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("readFile() requires a string filename")
				}
				content, err := os.ReadFile(filename.Value)
				if err != nil {
					return Error(err.Error())
				}
				return String(string(content))
			}),

			"readBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readBytes() takes exactly 1 argument")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("readBytes() requires a string filename")
				}
				content, err := os.ReadFile(filename.Value)
				if err != nil {
					return Error(err.Error())
				}
				result := make([]objects.Object, len(content))
				for i, b := range content {
					result[i] = Int(int64(b))
				}
				return Array(result...)
			}),

			"writeFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeFile() takes exactly 2 arguments")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeFile() requires a string filename")
				}
				content, ok := args[1].(*objects.String)
				if !ok {
					return Error("writeFile() requires a string content")
				}
				err := os.WriteFile(filename.Value, []byte(content.Value), 0644)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			"writeBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeBytes() takes exactly 2 arguments")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeBytes() requires a string filename")
				}
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("writeBytes() requires an array of bytes")
				}
				data := make([]byte, len(arr.Elements))
				for i, elem := range arr.Elements {
					n, ok := elem.(*objects.Int)
					if !ok {
						return Error("writeBytes() requires integer array elements")
					}
					if n.Value < 0 || n.Value > 255 {
						return Error("writeBytes() byte values must be 0-255")
					}
					data[i] = byte(n.Value)
				}
				err := os.WriteFile(filename.Value, data, 0644)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			"appendFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("appendFile() takes exactly 2 arguments")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("appendFile() requires a string filename")
				}
				content, ok := args[1].(*objects.String)
				if !ok {
					return Error("appendFile() requires a string content")
				}
				file, err := os.OpenFile(filename.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()
				_, err = file.WriteString(content.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			"exists": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("exists() takes exactly 1 argument")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("exists() requires a string filename")
				}
				_, err := os.Stat(filename.Value)
				return Bool(err == nil)
			}),

			"remove": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("remove() takes exactly 1 argument")
				}
				filename, ok := args[0].(*objects.String)
				if !ok {
					return Error("remove() requires a string filename")
				}
				err := os.Remove(filename.Value)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			"mkdir": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("mkdir() takes exactly 1 argument")
				}
				dirname, ok := args[0].(*objects.String)
				if !ok {
					return Error("mkdir() requires a string directory name")
				}
				err := os.MkdirAll(dirname.Value, 0755)
				if err != nil {
					return Error(err.Error())
				}
				return Null()
			}),

			"cwd": BuiltinFunc(func(args ...objects.Object) objects.Object {
				dir, err := os.Getwd()
				if err != nil {
					return Error(err.Error())
				}
				return String(dir)
			}),

			"exit": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) > 0 {
					code, ok := args[0].(*objects.Int)
					if ok {
						os.Exit(int(code.Value))
					}
				}
				os.Exit(0)
				return Null()
			}),

			"env": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("env() takes exactly 1 argument")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("env() requires a string key")
				}
				value := os.Getenv(key.Value)
				if value == "" {
					return Null()
				}
				return String(value)
			}),

			"setEnv": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setEnv() takes exactly 2 arguments")
				}
				key, ok := args[0].(*objects.String)
				if !ok {
					return Error("setEnv() requires a string key")
				}
				value, ok := args[1].(*objects.String)
				if !ok {
					return Error("setEnv() requires a string value")
				}
				os.Setenv(key.Value, value.Value)
				return Null()
			}),

			"args": BuiltinFunc(func(args ...objects.Object) objects.Object {
				result := []objects.Object{}
				for _, arg := range os.Args {
					result = append(result, String(arg))
				}
				return Array(result...)
			}),

			// ioCopy copies data from a Reader to a Writer using streaming I/O.
			// Usage: n = io.copy(dstWriter, srcReader)
			// Returns the number of bytes copied.
			"ioCopy": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("ioCopy() takes exactly 2 arguments")
				}
				dst, ok := args[0].(*objects.Writer)
				if !ok {
					return Error("ioCopy() requires a WRITER as first argument")
				}
				src, ok := args[1].(*objects.Reader)
				if !ok {
					return Error("ioCopy() requires a READER as second argument")
				}
				return objects.IoCopy(dst, src)
			}),

			// newReader creates a new Reader from any io.Reader source.
			// Currently supports File, FileUpload, HttpReq (body), String, Bytes,
			// BytesBuffer, Chars (as UTF-8 string), and Array of integers (0-255, as bytes).
			// Usage: r = io.newReader(source)
			"newReader": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newReader() takes exactly 1 argument")
				}
				switch src := args[0].(type) {
				case *objects.File:
					return objects.NewReader(src.Handle)
				case *objects.FileUpload:
					file, err := src.Open()
					if err != nil {
						return Error("newReader() failed to open FileUpload: " + err.Error())
					}
					return objects.NewReader(file)
				case *objects.HttpReq:
					if src.Value == nil || src.Value.Body == nil {
						return Error("newReader() HttpReq has no body")
					}
					return objects.NewReader(src.Value.Body)
				case *objects.String:
					return objects.NewReader(objects.NewStringReader(src.Value))
				case *objects.BytesBuffer:
					return objects.NewReader(src.GetIOReader())
				case *objects.Chars:
					// Convert Chars (Unicode code points) to UTF-8 string reader
					return objects.NewReader(objects.NewStringReader(src.String()))
				case *objects.Array:
					// Treat as byte array - convert integers 0-255 to []byte
					data := make([]byte, len(src.Elements))
					for i, elem := range src.Elements {
						n, ok := elem.(*objects.Int)
						if !ok {
							return Error("newReader() array must contain only integers")
						}
						if n.Value < 0 || n.Value > 255 {
							return Error("newReader() array integers must be 0-255")
						}
						data[i] = byte(n.Value)
					}
					return objects.NewReader(objects.NewBytesReader(data))
				case *objects.Bytes:
					// Immutable Bytes object
					return objects.NewReader(src.GetIOReader())
				default:
					return Error("newReader() requires FILE, FILE_UPLOAD, HTTP_REQ, STRING, BYTES, BYTES_BUFFER, CHARS, or ARRAY of bytes")
				}
			}),

			// newWriter creates a new Writer from any io.Writer destination.
			// Currently supports File, HttpResp, StringBuilder, and BytesBuffer.
			// Usage: w = io.newWriter(destination)
			"newWriter": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("newWriter() takes exactly 1 argument")
				}
				switch dst := args[0].(type) {
				case *objects.File:
					return objects.NewWriter(dst.Handle)
				case *objects.HttpResp:
					if dst.Value == nil {
						return Error("newWriter() HttpResp is nil")
					}
					return objects.NewWriter(dst.Value)
				case *objects.StringBuilder:
					return objects.NewWriter(dst.GetIOWriter())
				case *objects.BytesBuffer:
					return objects.NewWriter(dst.GetIOWriter())
				default:
					return Error("newWriter() requires FILE, HTTP_RESP, STRING_BUILDER, or BYTES_BUFFER")
				}
			}),

			// read reads up to n bytes from a Reader and returns as byte array.
			// Usage: data = io.read(reader, n)
			"read": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("read() takes exactly 2 arguments")
				}
				reader, ok := args[0].(*objects.Reader)
				if !ok {
					return Error("read() requires a READER as first argument")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("read() requires an INT as second argument")
				}
				return reader.Read(int(n.Value))
			}),

			// readStr reads up to n bytes from a Reader and returns as string.
			// Usage: s = io.readStr(reader, n)
			"readStr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("readStr() takes exactly 2 arguments")
				}
				reader, ok := args[0].(*objects.Reader)
				if !ok {
					return Error("readStr() requires a READER as first argument")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("readStr() requires an INT as second argument")
				}
				return reader.ReadStr(int(n.Value))
			}),

			// readAllStr reads all remaining content from a Reader as string.
			// Usage: s = io.readAllStr(reader)
			"readAllStr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readAllStr() takes exactly 1 argument")
				}
				reader, ok := args[0].(*objects.Reader)
				if !ok {
					return Error("readAllStr() requires a READER")
				}
				return reader.ReadAllStr()
			}),

			// readAllBytes reads all remaining content from a Reader as byte array.
			// Usage: data = io.readAllBytes(reader)
			"readAllBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readAllBytes() takes exactly 1 argument")
				}
				reader, ok := args[0].(*objects.Reader)
				if !ok {
					return Error("readAllBytes() requires a READER")
				}
				return reader.ReadAllBytes()
			}),

			// readLineFrom reads a single line from a Reader.
			// Usage: line = io.readLineFrom(reader)
			"readLineFrom": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("readLineFrom() takes exactly 1 argument")
				}
				reader, ok := args[0].(*objects.Reader)
				if !ok {
					return Error("readLineFrom() requires a READER")
				}
				return reader.ReadLine()
			}),

			// writeTo writes a byte array to a Writer.
			// Usage: n = io.writeTo(writer, data)
			"writeTo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeTo() takes exactly 2 arguments")
				}
				writer, ok := args[0].(*objects.Writer)
				if !ok {
					return Error("writeTo() requires a WRITER as first argument")
				}
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("writeTo() requires an ARRAY as second argument")
				}
				return writer.WriteBytes(arr)
			}),

			// writeStrTo writes a string to a Writer.
			// Usage: n = io.writeStrTo(writer, str)
			"writeStrTo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeStrTo() takes exactly 2 arguments")
				}
				writer, ok := args[0].(*objects.Writer)
				if !ok {
					return Error("writeStrTo() requires a WRITER as first argument")
				}
				s, ok := args[1].(*objects.String)
				if !ok {
					return Error("writeStrTo() requires a STRING as second argument")
				}
				return writer.WriteStr(s.Value)
			}),

			// writeBytesTo writes a byte array to a Writer.
			// Usage: n = io.writeBytesTo(writer, data)
			"writeBytesTo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("writeBytesTo() takes exactly 2 arguments")
				}
				writer, ok := args[0].(*objects.Writer)
				if !ok {
					return Error("writeBytesTo() requires a WRITER as first argument")
				}
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("writeBytesTo() requires an ARRAY as second argument")
				}
				return writer.WriteBytes(arr)
			}),

			// copyStream copies data from a Reader to a Writer (alias for ioCopy).
			// Usage: n = io.copyStream(dstWriter, srcReader)
			"copyStream": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("copyStream() takes exactly 2 arguments")
				}
				dst, ok := args[0].(*objects.Writer)
				if !ok {
					return Error("copyStream() requires a WRITER as first argument")
				}
				src, ok := args[1].(*objects.Reader)
				if !ok {
					return Error("copyStream() requires a READER as second argument")
				}
				return objects.IoCopy(dst, src)
			}),

			// pipe creates a pipe with a Reader and Writer.
			// Usage: r, w = io.pipe()
			"pipe": BuiltinFunc(func(args ...objects.Object) objects.Object {
				r, w := io.Pipe()
				return Array(objects.NewReader(r), objects.NewWriter(w))
			}),

			// ============================================
			// Scan functions for reading input
			// ============================================

			// scan reads a line from stdin and returns it as a string.
			// Optional argument: prompt string to display before reading.
			// Usage: s = io.scan() or s = io.scan("Enter name: ")
			"scan": BuiltinFunc(func(args ...objects.Object) objects.Object {
				prompt := ""
				if len(args) > 0 {
					if p, ok := args[0].(*objects.String); ok {
						prompt = p.Value
					}
				}
				return objects.Scan(prompt)
			}),

			// scanInt reads an integer from stdin.
			// Optional argument: prompt string.
			// Usage: n = io.scanInt() or n = io.scanInt("Enter number: ")
			"scanInt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				prompt := ""
				if len(args) > 0 {
					if p, ok := args[0].(*objects.String); ok {
						prompt = p.Value
					}
				}
				return objects.ScanInt(prompt)
			}),

			// scanFloat reads a float from stdin.
			// Optional argument: prompt string.
			// Usage: f = io.scanFloat() or f = io.scanFloat("Enter decimal: ")
			"scanFloat": BuiltinFunc(func(args ...objects.Object) objects.Object {
				prompt := ""
				if len(args) > 0 {
					if p, ok := args[0].(*objects.String); ok {
						prompt = p.Value
					}
				}
				return objects.ScanFloat(prompt)
			}),

			// scanBool reads a boolean from stdin.
			// Accepts "true", "false", "1", "0" (case-insensitive).
			// Optional argument: prompt string.
			// Usage: b = io.scanBool() or b = io.scanBool("Continue? ")
			"scanBool": BuiltinFunc(func(args ...objects.Object) objects.Object {
				prompt := ""
				if len(args) > 0 {
					if p, ok := args[0].(*objects.String); ok {
						prompt = p.Value
					}
				}
				return objects.ScanBool(prompt)
			}),

			// scanN reads n whitespace-delimited tokens from stdin.
			// Returns an array of strings.
			// Usage: tokens = io.scanN(3)
			"scanN": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("scanN() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("scanN() requires an integer argument")
				}
				return objects.ScanN(int(n.Value))
			}),

			// scanSplit reads a line and splits it by the given separator.
			// If separator is empty, splits by whitespace.
			// Returns an array of strings.
			// Usage: parts = io.scanSplit(",") or parts = io.scanSplit("")
			"scanSplit": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("scanSplit() takes exactly 1 argument")
				}
				sep, ok := args[0].(*objects.String)
				if !ok {
					return Error("scanSplit() requires a string argument")
				}
				return objects.ScanSplit(sep.Value)
			}),

			// scan2 reads two whitespace-delimited tokens from stdin.
			// Returns two values for multiple assignment.
			// Usage: a, b = io.scan2()
			"scan2": BuiltinFunc(func(args ...objects.Object) objects.Object {
				v1, v2 := objects.Scan2()
				return Array(v1, v2)
			}),

			// scan3 reads three whitespace-delimited tokens from stdin.
			// Returns three values for multiple assignment.
			// Usage: a, b, c = io.scan3()
			"scan3": BuiltinFunc(func(args ...objects.Object) objects.Object {
				v1, v2, v3 := objects.Scan3()
				return Array(v1, v2, v3)
			}),

			// scanf reads input according to a format string with {} placeholders.
			// Returns an array of parsed string values.
			// Usage: values = io.scanf("{} {} {}")
			"scanf": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("scanf() takes exactly 1 argument")
				}
				format, ok := args[0].(*objects.String)
				if !ok {
					return Error("scanf() requires a string format argument")
				}
				return objects.Scanf(format.Value)
			}),

			// newScanner creates a new Scanner object for reading input.
			// Without arguments, creates a scanner for stdin.
			// With a Reader argument, creates a scanner for that reader.
			// Usage: s = io.newScanner() or s = io.newScanner(reader)
			"newScanner": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) == 0 {
					return objects.NewScanner(nil)
				}
				if len(args) != 1 {
					return Error("newScanner() takes at most 1 argument")
				}
				reader, ok := args[0].(*objects.Reader)
				if !ok {
					return Error("newScanner() requires a READER argument")
				}
				return objects.NewScanner(reader.Value)
			}),
		},
	})
}
