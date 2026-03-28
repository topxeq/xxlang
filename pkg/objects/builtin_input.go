//go:build !js

// pkg/objects/builtin_input.go
// Clipboard and input built-in functions for Xxlang
package objects

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func init() {
	// Input functions
	Builtins["getInput"] = &Builtin{Fn: builtinGetInput}
	Builtins["getInputf"] = &Builtin{Fn: builtinGetInputf}
	Builtins["getChar"] = &Builtin{Fn: builtinGetChar}
	Builtins["getMultiLineInput"] = &Builtin{Fn: builtinGetMultiLineInput}
	Builtins["getPassword"] = &Builtin{Fn: builtinGetPassword}
	Builtins["confirm"] = &Builtin{Fn: builtinConfirm}
	Builtins["readLine"] = &Builtin{Fn: builtinReadLine}

	// Clipboard functions (platform dependent - basic implementation)
	Builtins["getClipText"] = &Builtin{Fn: builtinGetClipText}
	Builtins["setClipText"] = &Builtin{Fn: builtinSetClipText}
}

var stdinReader = bufio.NewReader(os.Stdin)

// builtinGetInput - get user input
// Usage: getInput() -> string
func builtinGetInput(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getInput. got=%d, want=0", len(args))
	}

	input, err := stdinReader.ReadString('\n')
	if err != nil {
		return NewString("")
	}

	return NewString(strings.TrimRight(input, "\r\n"))
}

// builtinGetInputf - get user input with prompt
// Usage: getInputf(format, args...) -> string
func builtinGetInputf(args ...Object) Object {
	if len(args) < 1 {
		return newError("wrong number of arguments for getInputf. got=%d, want>=1", len(args))
	}

	format, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'getInputf' must be STRING, got %s", args[0].Type())
	}

	// Build prompt
	prompt := format.Value
	if len(args) > 1 {
		promptArgs := make([]interface{}, len(args)-1)
		for i, arg := range args[1:] {
			switch v := arg.(type) {
			case *Int:
				promptArgs[i] = v.Value
			case *Float:
				promptArgs[i] = v.Value
			case *String:
				promptArgs[i] = v.Value
			case *Bool:
				promptArgs[i] = v.Value
			default:
				promptArgs[i] = v.Inspect()
			}
		}
		prompt = fmt.Sprintf(format.Value, promptArgs...)
	}

	fmt.Print(prompt)

	input, err := stdinReader.ReadString('\n')
	if err != nil {
		return NewString("")
	}

	return NewString(strings.TrimRight(input, "\r\n"))
}

// builtinGetChar - get single character input
// Usage: getChar() -> string
func builtinGetChar(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getChar. got=%d, want=0", len(args))
	}

	b, err := stdinReader.ReadByte()
	if err != nil {
		return NewString("")
	}

	return NewString(string(b))
}

// builtinGetMultiLineInput - get multi-line input until empty line
// Usage: getMultiLineInput() -> string
//
//	getMultiLineInput(endMarker) -> string
func builtinGetMultiLineInput(args ...Object) Object {
	if len(args) > 1 {
		return newError("wrong number of arguments for getMultiLineInput. got=%d, want=0 or 1", len(args))
	}

	endMarker := ""
	if len(args) == 1 {
		m, ok := args[0].(*String)
		if !ok {
			return newError("argument to 'getMultiLineInput' must be STRING, got %s", args[0].Type())
		}
		endMarker = m.Value
	}

	var lines []string
	for {
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimRight(line, "\r\n")

		if endMarker == "" && line == "" {
			break
		}
		if endMarker != "" && line == endMarker {
			break
		}

		lines = append(lines, line)
	}

	return NewString(strings.Join(lines, "\n"))
}

// builtinGetPassword - get password input (no echo)
// Note: This is a simplified version that doesn't hide input
// Usage: getPassword(prompt) -> string
func builtinGetPassword(args ...Object) Object {
	if len(args) > 1 {
		return newError("wrong number of arguments for getPassword. got=%d, want=0 or 1", len(args))
	}

	if len(args) == 1 {
		prompt, ok := args[0].(*String)
		if !ok {
			return newError("argument to 'getPassword' must be STRING, got %s", args[0].Type())
		}
		fmt.Print(prompt.Value)
	}

	input, err := stdinReader.ReadString('\n')
	if err != nil {
		return NewString("")
	}

	return NewString(strings.TrimRight(input, "\r\n"))
}

// builtinConfirm - get yes/no confirmation
// Usage: confirm(prompt) -> bool
//
//	confirm(prompt, defaultValue) -> bool
func builtinConfirm(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments for confirm. got=%d, want=1 or 2", len(args))
	}

	prompt, ok := args[0].(*String)
	if !ok {
		return newError("first argument to 'confirm' must be STRING, got %s", args[0].Type())
	}

	defaultValue := false
	defaultStr := "[y/N]"
	if len(args) == 2 {
		if d, ok := args[1].(*Bool); ok {
			defaultValue = d.Value
			if defaultValue {
				defaultStr = "[Y/n]"
			}
		}
	}

	fmt.Printf("%s %s: ", prompt.Value, defaultStr)

	input, err := stdinReader.ReadString('\n')
	if err != nil {
		return &Bool{Value: defaultValue}
	}

	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return &Bool{Value: defaultValue}
	}

	return &Bool{Value: input == "y" || input == "yes"}
}

// builtinReadLine - read a line from stdin
// Usage: readLine() -> string or null
func builtinReadLine(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for readLine. got=%d, want=0", len(args))
	}

	input, err := stdinReader.ReadString('\n')
	if err != nil {
		return NULL
	}

	return NewString(strings.TrimRight(input, "\r\n"))
}

// builtinGetClipText - get clipboard text
// Note: This is a placeholder implementation
// Real implementation requires platform-specific code
// Usage: getClipText() -> string
func builtinGetClipText(args ...Object) Object {
	if len(args) != 0 {
		return newError("wrong number of arguments for getClipText. got=%d, want=0", len(args))
	}

	// Try to get clipboard content using system command
	// This is a simplified implementation
	return newError("clipboard not supported in this environment")
}

// builtinSetClipText - set clipboard text
// Note: This is a placeholder implementation
// Real implementation requires platform-specific code
// Usage: setClipText(text) -> null
func builtinSetClipText(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments for setClipText. got=%d, want=1", len(args))
	}

	text, ok := args[0].(*String)
	if !ok {
		return newError("argument to 'setClipText' must be STRING, got %s", args[0].Type())
	}

	// Placeholder - would need platform-specific implementation
	_ = text.Value

	return newError("clipboard not supported in this environment")
}
