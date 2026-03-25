// pkg/stdlib/xlsx.go
// XLSX module for Xxlang - Excel file handling.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "xlsx",
		Exports: map[string]objects.Object{
			// open opens an existing xlsx file
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("open takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("path must be a string")
				}
				wb, err := objects.OpenXLSX(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				return wb
			}),

			// create creates a new empty workbook
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewXLSX()
			}),

			// isXLSX checks if an object is an XLSX workbook
			"isXLSX": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isXLSX takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.XLSX)
				return Bool(ok)
			}),

			// colToIndex converts column letter to index (A=1, B=2, ...)
			"colToIndex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("colToIndex takes exactly 1 argument")
				}
				col, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				return Int(int64(objects.ColToIndex(col.Value)))
			}),

			// indexToCol converts index to column letter (1=A, 2=B, ...)
			"indexToCol": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("indexToCol takes exactly 1 argument")
				}
				idx, ok := args[0].(*objects.Int)
				if !ok {
					return Error("argument must be an integer")
				}
				return String(objects.IndexToCol(int(idx.Value)))
			}),

			// parseCellRef parses a cell reference (e.g., "A1" -> ["A", 1])
			"parseCellRef": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("parseCellRef takes exactly 1 argument")
				}
				ref, ok := args[0].(*objects.String)
				if !ok {
					return Error("argument must be a string")
				}
				col, row := objects.ParseCellRef(ref.Value)
				return Array(String(col), Int(int64(row)))
			}),
		},
	})
}