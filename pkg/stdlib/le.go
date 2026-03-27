// pkg/stdlib/le.go
// Line Editor (le) module for Xxlang - line-based text editing functionality.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

// emptyArray is a reusable empty array for le module.
var emptyArray = &objects.Array{Elements: []objects.Object{}}

func init() {
	Register(&Module{
		Name: "le",
		Exports: map[string]objects.Object{
			// ============================================================
			// Creation Functions
			// ============================================================

			// open opens a file and returns a LineEditor object
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("open takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				return le
			}),

			// fromText creates a LineEditor from a string
			"fromText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromText takes exactly 1 argument")
				}
				text, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				return objects.NewLineEditorFromText(text.Value)
			}),

			// fromLines creates a LineEditor from a string array
			"fromLines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromLines takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("argument must be an array")
				}
				lines := make([]string, len(arr.Elements))
				for i, elem := range arr.Elements {
					if s, ok := elem.(*objects.String); ok {
						lines[i] = s.Value
					} else {
						lines[i] = elem.Inspect()
					}
				}
				return objects.NewLineEditorFromLines(lines)
			}),

			// create creates an empty LineEditor
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewLineEditor()
			}),

			// isLineEditor checks if an object is a LineEditor
			"isLineEditor": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isLineEditor takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.LineEditor)
				return Bool(ok)
			}),

			// ============================================================
			// Quick Functions - One-time file operations
			// ============================================================

			// replaceInFile replaces text in a file
			"replaceInFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("replaceInFile takes exactly 3 arguments (path, old, new)")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string")
				}
				old, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string")
				}
				newStr, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				le.Replace(old.Value, newStr.Value)
				if err := le.Save(); err != nil {
					return Error("failed to save file: " + err.Error())
				}
				return objects.NULL
			}),

			// sortFile sorts a file (overwrites)
			"sortFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sortFile takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				le.Sort()
				if err := le.Save(); err != nil {
					return Error("failed to save file: " + err.Error())
				}
				return objects.NULL
			}),

			// sortFileTo sorts a file and saves to a new path
			"sortFileTo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("sortFileTo takes exactly 2 arguments (src, dst)")
				}
				src, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string")
				}
				dst, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(src.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				le.Sort()
				if err := le.SaveAs(dst.Value); err != nil {
					return Error("failed to save file: " + err.Error())
				}
				return objects.NULL
			}),

			// uniqueFile removes duplicate lines from a file (overwrites)
			"uniqueFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("uniqueFile takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				le.Unique()
				if err := le.Save(); err != nil {
					return Error("failed to save file: " + err.Error())
				}
				return objects.NULL
			}),

			// uniqueFileTo removes duplicates and saves to a new path
			"uniqueFileTo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("uniqueFileTo takes exactly 2 arguments (src, dst)")
				}
				src, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string")
				}
				dst, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(src.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				le.Unique()
				if err := le.SaveAs(dst.Value); err != nil {
					return Error("failed to save file: " + err.Error())
				}
				return objects.NULL
			}),

			// grepFile filters a file by pattern and returns matching lines
			"grepFile": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("grepFile takes exactly 2 arguments (path, pattern)")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string")
				}
				pattern, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				lines := le.FindAll(pattern.Value)
				elements := make([]objects.Object, len(lines))
				for i, line := range lines {
					elements[i] = objects.NewString(line)
				}
				return &objects.Array{Elements: elements}
			}),

			// grepFileTo filters a file and saves to a new path
			"grepFileTo": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("grepFileTo takes exactly 3 arguments (src, pattern, dst)")
				}
				src, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string")
				}
				pattern, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string")
				}
				dst, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(src.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				filtered := le.Grep(pattern.Value)
				if err := filtered.SaveAs(dst.Value); err != nil {
					return Error("failed to save file: " + err.Error())
				}
				return objects.NULL
			}),

			// removeEmptyLines removes empty lines from a file
			"removeEmptyLines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("removeEmptyLines takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				le.RemoveEmpty()
				if err := le.Save(); err != nil {
					return Error("failed to save file: " + err.Error())
				}
				return objects.NULL
			}),

			// countLines counts lines in a file
			"countLines": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("countLines takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				return objects.NewInt(int64(le.LineCount()))
			}),

			// head gets first n lines from a file
			"head": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("head takes exactly 2 arguments (path, n)")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("second argument must be an integer")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}

				count := int(n.Value)
				if count <= 0 {
					return emptyArray
				}
				if count > le.LineCount() {
					count = le.LineCount()
				}
				lines := le.GetLines(1, count)
				elements := make([]objects.Object, len(lines))
				for i, line := range lines {
					elements[i] = objects.NewString(line)
				}
				return &objects.Array{Elements: elements}
			}),

			// tail gets last n lines from a file
			"tail": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("tail takes exactly 2 arguments (path, n)")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("second argument must be an integer")
				}

				le, err := objects.NewLineEditorFromFile(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}

				count := int(n.Value)
				if count <= 0 {
					return emptyArray
				}
				lineCount := le.LineCount()
				if count > lineCount {
					count = lineCount
				}
				start := lineCount - count + 1
				lines := le.GetLines(start, lineCount)
				elements := make([]objects.Object, len(lines))
				for i, line := range lines {
					elements[i] = objects.NewString(line)
				}
				return &objects.Array{Elements: elements}
			}),

			// ============================================================
			// SSH Operations - Load/Save files via SSH
			// ============================================================

			// loadFromSsh loads a remote file via SSH and returns a LineEditor.
			// Parameters: host, port, user, password, remotePath
			// Returns: LineEditor object or Error
			"loadFromSsh": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 5 {
					return Error("loadFromSsh takes exactly 5 arguments (host, port, user, password, remotePath)")
				}
				host, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (host)")
				}
				port, ok := args[1].(*objects.Int)
				if !ok {
					return Error("second argument must be an integer (port)")
				}
				user, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string (user)")
				}
				password, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (password)")
				}
				remotePath, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				content, err := client.ReadFile(remotePath.Value)
				if err != nil {
					return Error("failed to read remote file: " + err.Error())
				}

				return objects.NewLineEditorFromText(content)
			}),

			// loadFromSshWithKey loads a remote file via SSH with key authentication.
			// Parameters: host, port, user, keyPath, remotePath
			// Returns: LineEditor object or Error
			"loadFromSshWithKey": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 5 {
					return Error("loadFromSshWithKey takes exactly 5 arguments (host, port, user, keyPath, remotePath)")
				}
				host, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (host)")
				}
				port, ok := args[1].(*objects.Int)
				if !ok {
					return Error("second argument must be an integer (port)")
				}
				user, ok := args[2].(*objects.String)
				if !ok {
					return Error("third argument must be a string (user)")
				}
				keyPath, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (keyPath)")
				}
				remotePath, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.ConnectWithKey(host.Value, int(port.Value), user.Value, keyPath.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				content, err := client.ReadFile(remotePath.Value)
				if err != nil {
					return Error("failed to read remote file: " + err.Error())
				}

				return objects.NewLineEditorFromText(content)
			}),

			// saveToSsh saves a LineEditor content to a remote file via SSH.
			// Parameters: lineEditor, host, port, user, password, remotePath
			// Returns: null on success, Error on failure
			"saveToSsh": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 6 {
					return Error("saveToSsh takes exactly 6 arguments (lineEditor, host, port, user, password, remotePath)")
				}
				le, ok := args[0].(*objects.LineEditor)
				if !ok {
					return Error("first argument must be a LineEditor")
				}
				host, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (host)")
				}
				port, ok := args[2].(*objects.Int)
				if !ok {
					return Error("third argument must be an integer (port)")
				}
				user, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (user)")
				}
				password, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (password)")
				}
				remotePath, ok := args[5].(*objects.String)
				if !ok {
					return Error("sixth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				content := le.ToText()
				if err := client.WriteFile(remotePath.Value, content); err != nil {
					return Error("failed to write remote file: " + err.Error())
				}

				return objects.NULL
			}),

			// saveToSshWithKey saves a LineEditor content to a remote file via SSH with key authentication.
			// Parameters: lineEditor, host, port, user, keyPath, remotePath
			// Returns: null on success, Error on failure
			"saveToSshWithKey": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 6 {
					return Error("saveToSshWithKey takes exactly 6 arguments (lineEditor, host, port, user, keyPath, remotePath)")
				}
				le, ok := args[0].(*objects.LineEditor)
				if !ok {
					return Error("first argument must be a LineEditor")
				}
				host, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (host)")
				}
				port, ok := args[2].(*objects.Int)
				if !ok {
					return Error("third argument must be an integer (port)")
				}
				user, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (user)")
				}
				keyPath, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (keyPath)")
				}
				remotePath, ok := args[5].(*objects.String)
				if !ok {
					return Error("sixth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.ConnectWithKey(host.Value, int(port.Value), user.Value, keyPath.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				content := le.ToText()
				if err := client.WriteFile(remotePath.Value, content); err != nil {
					return Error("failed to write remote file: " + err.Error())
				}

				return objects.NULL
			}),

			// appendToSsh appends a LineEditor content to a remote file via SSH.
			// Parameters: lineEditor, host, port, user, password, remotePath
			// Returns: null on success, Error on failure
			"appendToSsh": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 6 {
					return Error("appendToSsh takes exactly 6 arguments (lineEditor, host, port, user, password, remotePath)")
				}
				le, ok := args[0].(*objects.LineEditor)
				if !ok {
					return Error("first argument must be a LineEditor")
				}
				host, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (host)")
				}
				port, ok := args[2].(*objects.Int)
				if !ok {
					return Error("third argument must be an integer (port)")
				}
				user, ok := args[3].(*objects.String)
				if !ok {
					return Error("fourth argument must be a string (user)")
				}
				password, ok := args[4].(*objects.String)
				if !ok {
					return Error("fifth argument must be a string (password)")
				}
				remotePath, ok := args[5].(*objects.String)
				if !ok {
					return Error("sixth argument must be a string (remotePath)")
				}

				client := objects.NewSSHClient()
				if err := client.Connect(host.Value, int(port.Value), user.Value, password.Value); err != nil {
					return Error("SSH connection failed: " + err.Error())
				}
				defer client.Close()

				// Read existing content first
				existingContent, _ := client.ReadFile(remotePath.Value)
				newContent := existingContent + "\n" + le.ToText()
				if err := client.WriteFile(remotePath.Value, newContent); err != nil {
					return Error("failed to append to remote file: " + err.Error())
				}

				return objects.NULL
			}),
		},
	})
}
