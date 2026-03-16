// pkg/stdlib/io.go
// I/O utilities for the Xxlang standard library.
package stdlib

import (
	"bufio"
	"fmt"
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
		},
	})
}
