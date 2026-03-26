// pkg/stdlib/csv.go
// CSV parsing and generation utilities for the Xxlang standard library.
package stdlib

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

// findColumnIndex finds the column index by name in a header row.
// Returns -1 if not found.
func findColumnIndex(header *objects.Array, name string) int {
	for i, elem := range header.Elements {
		if str, ok := elem.(*objects.String); ok && str.Value == name {
			return i
		}
	}
	return -1
}

func init() {
	Register(&Module{
		Name: "csv",
		Exports: map[string]objects.Object{
			// Parse CSV string to array of arrays
			"parse": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("parse() takes at least 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parse() requires a string argument")
				}
				comma := ','
				if len(args) > 1 {
					c, ok := args[1].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}
				reader := csv.NewReader(strings.NewReader(s.Value))
				reader.Comma = comma
				result := []objects.Object{}
				for {
					record, err := reader.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						return Error(err.Error())
					}
					row := make([]objects.Object, len(record))
					for i, field := range record {
						row[i] = String(field)
					}
					result = append(result, Array(row...))
				}
				return Array(result...)
			}),

			// Parse CSV with header to array of maps
			"parseWithHeader": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("parseWithHeader() takes at least 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseWithHeader() requires a string argument")
				}
				comma := ','
				if len(args) > 1 {
					c, ok := args[1].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}
				reader := csv.NewReader(strings.NewReader(s.Value))
				reader.Comma = comma
				// Read header
				header, err := reader.Read()
				if err != nil {
					return Error(err.Error())
				}
				result := []objects.Object{}
				for {
					record, err := reader.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						return Error(err.Error())
					}
					pairs := make(map[objects.HashKey]objects.MapPair)
					for i, field := range record {
						if i < len(header) {
							key := String(header[i])
							pairs[key.HashKey()] = objects.MapPair{
								Key:   key,
								Value: String(field),
							}
						}
					}
					result = append(result, &objects.Map{Pairs: pairs})
				}
				return Array(result...)
			}),

			// Generate CSV from array of arrays
			"stringify": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("stringify() takes at least 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("stringify() requires an array argument")
				}
				comma := ','
				if len(args) > 1 {
					c, ok := args[1].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}
				var result strings.Builder
				writer := csv.NewWriter(&result)
				writer.Comma = comma
				for _, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok {
						return Error("stringify() requires array of arrays")
					}
					record := make([]string, len(rowArr.Elements))
					for i, field := range rowArr.Elements {
						switch v := field.(type) {
						case *objects.String:
							record[i] = v.Value
						default:
							record[i] = field.Inspect()
						}
					}
					if err := writer.Write(record); err != nil {
						return Error(err.Error())
					}
				}
				writer.Flush()
				if err := writer.Error(); err != nil {
					return Error(err.Error())
				}
				return String(result.String())
			}),

			// Generate CSV from array of maps
			"stringifyMaps": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("stringifyMaps() takes at least 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("stringifyMaps() requires an array as first argument")
				}
				headers, ok := args[1].(*objects.Array)
				if !ok {
					return Error("stringifyMaps() requires an array of headers")
				}
				comma := ','
				if len(args) > 2 {
					c, ok := args[2].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}
				// Extract header strings
				headerStrs := make([]string, len(headers.Elements))
				for i, h := range headers.Elements {
					s, ok := h.(*objects.String)
					if !ok {
						return Error("stringifyMaps() requires string headers")
					}
					headerStrs[i] = s.Value
				}
				var result strings.Builder
				writer := csv.NewWriter(&result)
				writer.Comma = comma
				// Write header
				if err := writer.Write(headerStrs); err != nil {
					return Error(err.Error())
				}
				// Write rows
				for _, row := range arr.Elements {
					m, ok := row.(*objects.Map)
					if !ok {
						return Error("stringifyMaps() requires array of maps")
					}
					record := make([]string, len(headerStrs))
					for i, key := range headerStrs {
						val := ""
						for _, pair := range m.Pairs {
							if k, ok := pair.Key.(*objects.String); ok && k.Value == key {
								switch v := pair.Value.(type) {
								case *objects.String:
									val = v.Value
								default:
									val = pair.Value.Inspect()
								}
								break
							}
						}
						record[i] = val
					}
					if err := writer.Write(record); err != nil {
						return Error(err.Error())
					}
				}
				writer.Flush()
				if err := writer.Error(); err != nil {
					return Error(err.Error())
				}
				return String(result.String())
			}),

			// Get column from parsed CSV
			"column": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("column() takes 2 or 3 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("column() requires an array as first argument")
				}

				// Check if column identifier is int or string
				var colIdx int = -1
				switch v := args[1].(type) {
				case *objects.Int:
					colIdx = int(v.Value)
				case *objects.String:
					// Column by name - need header
					if len(args) < 3 {
						return Error("column() by name requires header array as third argument")
					}
					header, ok := args[2].(*objects.Array)
					if !ok {
						return Error("column() header must be an array")
					}
					colIdx = findColumnIndex(header, v.Value)
					if colIdx < 0 {
						return Error("column() column name not found: " + v.Value)
					}
				default:
					return Error("column() requires integer index or string name")
				}

				result := []objects.Object{}
				for _, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok {
						continue
					}
					if colIdx >= 0 && colIdx < len(rowArr.Elements) {
						result = append(result, rowArr.Elements[colIdx])
					}
				}
				return Array(result...)
			}),

			// Get header (first row) from CSV data
			"getHeader": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getHeader() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("getHeader() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Array()
				}
				rowArr, ok := arr.Elements[0].(*objects.Array)
				if !ok {
					return Array()
				}
				return rowArr
			}),

			// Get column index by name from header
			"colIndex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("colIndex() takes exactly 2 arguments")
				}
				header, ok := args[0].(*objects.Array)
				if !ok {
					return Error("colIndex() requires header array as first argument")
				}
				colName, ok := args[1].(*objects.String)
				if !ok {
					return Error("colIndex() requires string column name as second argument")
				}
				idx := findColumnIndex(header, colName.Value)
				if idx < 0 {
					return Error("colIndex() column not found: " + colName.Value)
				}
				return Int(int64(idx))
			}),

			// Get column name by index from header
			"colName": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("colName() takes exactly 2 arguments")
				}
				header, ok := args[0].(*objects.Array)
				if !ok {
					return Error("colName() requires header array as first argument")
				}
				colIdx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("colName() requires integer column index as second argument")
				}
				idx := int(colIdx.Value)
				if idx < 0 || idx >= len(header.Elements) {
					return Error("colName() index out of range")
				}
				str, ok := header.Elements[idx].(*objects.String)
				if !ok {
					return String(header.Elements[idx].Inspect())
				}
				return str
			}),

			// Get column count
			"colCount": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("colCount() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("colCount() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Int(0)
				}
				rowArr, ok := arr.Elements[0].(*objects.Array)
				if !ok {
					return Int(0)
				}
				return Int(int64(len(rowArr.Elements)))
			}),

			// Set column value for all rows
			"setColumn": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 || len(args) > 4 {
					return Error("setColumn() takes 3 or 4 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("setColumn() requires an array as first argument")
				}

				var colIdx int = -1
				switch v := args[1].(type) {
				case *objects.Int:
					colIdx = int(v.Value)
				case *objects.String:
					if len(args) < 4 {
						return Error("setColumn() by name requires header array as fourth argument")
					}
					header, ok := args[3].(*objects.Array)
					if !ok {
						return Error("setColumn() header must be an array")
					}
					colIdx = findColumnIndex(header, v.Value)
					if colIdx < 0 {
						return Error("setColumn() column name not found: " + v.Value)
					}
				default:
					return Error("setColumn() requires integer index or string name")
				}

				value := args[2]

				result := make([]objects.Object, len(arr.Elements))
				for i, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok {
						result[i] = row
						continue
					}
					newRow := make([]objects.Object, len(rowArr.Elements))
					copy(newRow, rowArr.Elements)
					if colIdx >= 0 && colIdx < len(newRow) {
						newRow[colIdx] = value
					} else if colIdx >= len(newRow) {
						// Extend row
						extended := make([]objects.Object, colIdx+1)
						copy(extended, newRow)
						for j := len(newRow); j < colIdx; j++ {
							extended[j] = String("")
						}
						extended[colIdx] = value
						newRow = extended
					}
					result[i] = Array(newRow...)
				}
				return Array(result...)
			}),

			// Insert column at position
			"insertColumn": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("insertColumn() takes exactly 3 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("insertColumn() requires an array as first argument")
				}
				colIdx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("insertColumn() requires integer column index")
				}
				value := args[2]

				idx := int(colIdx.Value)
				result := make([]objects.Object, len(arr.Elements))
				for i, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok {
						result[i] = row
						continue
					}
					newRow := make([]objects.Object, len(rowArr.Elements)+1)
					if idx > len(rowArr.Elements) {
						idx = len(rowArr.Elements)
					}
					copy(newRow, rowArr.Elements[:idx])
					newRow[idx] = value
					copy(newRow[idx+1:], rowArr.Elements[idx:])
					result[i] = Array(newRow...)
				}
				return Array(result...)
			}),

			// Remove column at position
			"removeColumn": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("removeColumn() takes 2 or 3 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("removeColumn() requires an array as first argument")
				}

				var colIdx int = -1
				switch v := args[1].(type) {
				case *objects.Int:
					colIdx = int(v.Value)
				case *objects.String:
					if len(args) < 3 {
						return Error("removeColumn() by name requires header array as third argument")
					}
					header, ok := args[2].(*objects.Array)
					if !ok {
						return Error("removeColumn() header must be an array")
					}
					colIdx = findColumnIndex(header, v.Value)
					if colIdx < 0 {
						return Error("removeColumn() column name not found: " + v.Value)
					}
				default:
					return Error("removeColumn() requires integer index or string name")
				}

				result := make([]objects.Object, len(arr.Elements))
				for i, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok || colIdx < 0 || colIdx >= len(rowArr.Elements) {
						result[i] = row
						continue
					}
					newRow := make([]objects.Object, len(rowArr.Elements)-1)
					copy(newRow, rowArr.Elements[:colIdx])
					copy(newRow[colIdx:], rowArr.Elements[colIdx+1:])
					result[i] = Array(newRow...)
				}
				return Array(result...)
			}),

			// Rename column in header
			"renameColumn": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("renameColumn() takes exactly 3 arguments")
				}
				header, ok := args[0].(*objects.Array)
				if !ok {
					return Error("renameColumn() requires header array as first argument")
				}
				oldName, ok := args[1].(*objects.String)
				if !ok {
					return Error("renameColumn() requires string old name")
				}
				newName, ok := args[2].(*objects.String)
				if !ok {
					return Error("renameColumn() requires string new name")
				}

				idx := findColumnIndex(header, oldName.Value)
				if idx < 0 {
					return Error("renameColumn() column not found: " + oldName.Value)
				}

				newHeader := make([]objects.Object, len(header.Elements))
				copy(newHeader, header.Elements)
				newHeader[idx] = newName
				return Array(newHeader...)
			}),

			// ============================================================
			// Row operations
			// ============================================================

			// Get a specific row from CSV data
			"row": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("row() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("row() requires an array as first argument")
				}
				rowIdx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("row() requires integer row index")
				}
				idx := int(rowIdx.Value)
				if idx < 0 || idx >= len(arr.Elements) {
					return Error("row() index out of range")
				}
				return arr.Elements[idx]
			}),

			// Get row count
			"rowCount": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("rowCount() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("rowCount() requires an array argument")
				}
				return Int(int64(len(arr.Elements)))
			}),

			// Skip first n rows
			"skip": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("skip() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("skip() requires an array as first argument")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("skip() requires integer count")
				}
				count := int(n.Value)
				if count < 0 {
					count = 0
				}
				if count >= len(arr.Elements) {
					return Array()
				}
				return Array(arr.Elements[count:]...)
			}),

			// Take first n rows
			"take": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("take() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("take() requires an array as first argument")
				}
				n, ok := args[1].(*objects.Int)
				if !ok {
					return Error("take() requires integer count")
				}
				count := int(n.Value)
				if count < 0 {
					count = 0
				}
				if count >= len(arr.Elements) {
					return arr
				}
				return Array(arr.Elements[:count]...)
			}),

			// Append a row to CSV data
			"appendRow": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("appendRow() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("appendRow() requires an array as first argument")
				}
				newRow := args[1]
				result := make([]objects.Object, len(arr.Elements)+1)
				copy(result, arr.Elements)
				result[len(arr.Elements)] = newRow
				return Array(result...)
			}),

			// Prepend a row to CSV data
			"prependRow": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("prependRow() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("prependRow() requires an array as first argument")
				}
				newRow := args[1]
				result := make([]objects.Object, len(arr.Elements)+1)
				result[0] = newRow
				copy(result[1:], arr.Elements)
				return Array(result...)
			}),

			// Transpose CSV data (rows become columns)
			"transpose": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("transpose() takes exactly 1 argument")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("transpose() requires an array argument")
				}
				if len(arr.Elements) == 0 {
					return Array()
				}

				// Find max columns
				maxCols := 0
				for _, row := range arr.Elements {
					if rowArr, ok := row.(*objects.Array); ok {
						if len(rowArr.Elements) > maxCols {
							maxCols = len(rowArr.Elements)
						}
					}
				}

				// Create transposed result
				result := make([]objects.Object, maxCols)
				for col := 0; col < maxCols; col++ {
					newRow := make([]objects.Object, len(arr.Elements))
					for row := 0; row < len(arr.Elements); row++ {
						if rowArr, ok := arr.Elements[row].(*objects.Array); ok && col < len(rowArr.Elements) {
							newRow[row] = rowArr.Elements[col]
						} else {
							newRow[row] = String("")
						}
					}
					result[col] = Array(newRow...)
				}
				return Array(result...)
			}),

			// Filter rows by predicate function
			"filterRows": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("filterRows() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("filterRows() requires an array as first argument")
				}
				pred, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("filterRows() requires a function as second argument")
				}

				result := []objects.Object{}
				for _, row := range arr.Elements {
					res := pred.Fn(row)
					if b, ok := res.(*objects.Bool); ok && b.Value {
						result = append(result, row)
					}
				}
				return Array(result...)
			}),

			// Map rows by function
			"mapRows": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("mapRows() takes exactly 2 arguments")
				}
				arr, ok := args[0].(*objects.Array)
				if !ok {
					return Error("mapRows() requires an array as first argument")
				}
				fn, ok := args[1].(*objects.Builtin)
				if !ok {
					return Error("mapRows() requires a function as second argument")
				}

				result := make([]objects.Object, len(arr.Elements))
				for i, row := range arr.Elements {
					result[i] = fn.Fn(row)
				}
				return Array(result...)
			}),

			// ============================================================
			// File-based CSV operations
			// ============================================================

			// read reads a CSV file and returns an array of arrays.
			// Usage: data = csv.read(path, [delimiter])
			"read": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return Error("read() takes 1 or 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("read() requires a string path")
				}

				comma := ','
				if len(args) == 2 {
					c, ok := args[1].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}

				file, err := os.Open(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				reader := csv.NewReader(file)
				reader.Comma = comma
				reader.LazyQuotes = true

				var result []objects.Object
				for {
					record, err := reader.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						return Error(err.Error())
					}

					row := make([]objects.Object, len(record))
					for i, field := range record {
						row[i] = String(field)
					}
					result = append(result, Array(row...))
				}

				return Array(result...)
			}),

			// readWithHeader reads a CSV file with header and returns array of maps.
			// Usage: data = csv.readWithHeader(path, [delimiter])
			"readWithHeader": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return Error("readWithHeader() takes 1 or 2 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("readWithHeader() requires a string path")
				}

				comma := ','
				if len(args) == 2 {
					c, ok := args[1].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}

				file, err := os.Open(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				reader := csv.NewReader(file)
				reader.Comma = comma
				reader.LazyQuotes = true

				// Read header
				header, err := reader.Read()
				if err != nil {
					return Error(err.Error())
				}

				var result []objects.Object
				for {
					record, err := reader.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						return Error(err.Error())
					}

					pairs := make(map[objects.HashKey]objects.MapPair)
					for i, field := range record {
						if i < len(header) {
							key := String(header[i])
							pairs[key.HashKey()] = objects.MapPair{
								Key:   key,
								Value: String(field),
							}
						}
					}
					result = append(result, &objects.Map{Pairs: pairs})
				}

				return Array(result...)
			}),

			// write writes array of arrays to a CSV file.
			// Usage: csv.write(path, data, [delimiter])
			"write": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("write() takes 2 or 3 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("write() requires a string path")
				}
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("write() requires an array argument")
				}

				comma := ','
				if len(args) == 3 {
					c, ok := args[2].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}

				file, err := os.Create(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				writer.Comma = comma

				for _, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok {
						return Error("write() requires array of arrays")
					}
					record := make([]string, len(rowArr.Elements))
					for i, field := range rowArr.Elements {
						switch v := field.(type) {
						case *objects.String:
							record[i] = v.Value
						default:
							record[i] = field.Inspect()
						}
					}
					if err := writer.Write(record); err != nil {
						return Error(err.Error())
					}
				}

				writer.Flush()
				if err := writer.Error(); err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// writeWithHeader writes array of maps to CSV with header.
			// Usage: csv.writeWithHeader(path, data, headers, [delimiter])
			"writeWithHeader": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 || len(args) > 4 {
					return Error("writeWithHeader() takes 3 or 4 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("writeWithHeader() requires a string path")
				}
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("writeWithHeader() requires an array argument")
				}
				headers, ok := args[2].(*objects.Array)
				if !ok {
					return Error("writeWithHeader() requires headers array")
				}

				comma := ','
				if len(args) == 4 {
					c, ok := args[3].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}

				file, err := os.Create(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				writer.Comma = comma

				// Write header
				headerRecord := make([]string, len(headers.Elements))
				for i, h := range headers.Elements {
					s, ok := h.(*objects.String)
					if !ok {
						return Error("headers must be strings")
					}
					headerRecord[i] = s.Value
				}
				if err := writer.Write(headerRecord); err != nil {
					return Error(err.Error())
				}

				// Write data rows
				for _, row := range arr.Elements {
					m, ok := row.(*objects.Map)
					if !ok {
						return Error("writeWithHeader() requires array of maps")
					}

					record := make([]string, len(headers.Elements))
					for i, h := range headers.Elements {
						key, ok := h.(*objects.String)
						if !ok {
							continue
						}
						if pair, exists := m.Pairs[key.HashKey()]; exists {
							switch v := pair.Value.(type) {
							case *objects.String:
								record[i] = v.Value
							default:
								record[i] = pair.Value.Inspect()
							}
						}
					}
					if err := writer.Write(record); err != nil {
						return Error(err.Error())
					}
				}

				writer.Flush()
				if err := writer.Error(); err != nil {
					return Error(err.Error())
				}

				return Null()
			}),

			// append appends rows to an existing CSV file.
			// Usage: csv.append(path, data, [delimiter])
			"append": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 || len(args) > 3 {
					return Error("append() takes 2 or 3 arguments")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("append() requires a string path")
				}
				arr, ok := args[1].(*objects.Array)
				if !ok {
					return Error("append() requires an array argument")
				}

				comma := ','
				if len(args) == 3 {
					c, ok := args[2].(*objects.String)
					if ok && len(c.Value) > 0 {
						comma = rune(c.Value[0])
					}
				}

				file, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return Error(err.Error())
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				writer.Comma = comma

				for _, row := range arr.Elements {
					rowArr, ok := row.(*objects.Array)
					if !ok {
						return Error("append() requires array of arrays")
					}
					record := make([]string, len(rowArr.Elements))
					for i, field := range rowArr.Elements {
						switch v := field.(type) {
						case *objects.String:
							record[i] = v.Value
						default:
							record[i] = field.Inspect()
						}
					}
					if err := writer.Write(record); err != nil {
						return Error(err.Error())
					}
				}

				writer.Flush()
				if err := writer.Error(); err != nil {
					return Error(err.Error())
				}

				return Null()
			}),
		},
	})
}
