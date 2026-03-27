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
		},
	})
}