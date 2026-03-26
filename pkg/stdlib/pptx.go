// pkg/stdlib/pptx.go
// PPTX module for Xxlang - PowerPoint file handling.
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "pptx",
		Exports: map[string]objects.Object{
			// create creates a new empty presentation
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("create takes no arguments")
				}
				return objects.NewPPTX()
			}),

			// open opens an existing pptx file
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("open takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("path must be a string")
				}
				doc, err := objects.OpenPPTX(path.Value)
				if err != nil {
					return Error("failed to open file: " + err.Error())
				}
				return doc
			}),

			// fromBytes opens a pptx from byte data
			"fromBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromBytes takes exactly 1 argument")
				}
				data, ok := args[0].(*objects.Bytes)
				if !ok {
					return Error("data must be bytes")
				}
				doc, err := objects.OpenPPTXFromBytes(data.Value)
				if err != nil {
					return Error("failed to parse pptx data: " + err.Error())
				}
				return doc
			}),

			// isPPTX checks if an object is a PPTX document
			"isPPTX": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isPPTX takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.PPTXDocument)
				return Bool(ok)
			}),

			// isSlide checks if an object is a PPTX slide
			"isSlide": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isSlide takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.PPTXSlide)
				return Bool(ok)
			}),

			// isTextFrame checks if an object is a PPTX text frame
			"isTextFrame": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isTextFrame takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.PPTXTextFrame)
				return Bool(ok)
			}),

			// isShape checks if an object is a PPTX shape
			"isShape": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isShape takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.PPTXShape)
				return Bool(ok)
			}),

			// isTable checks if an object is a PPTX table
			"isTable": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isTable takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.PPTXTable)
				return Bool(ok)
			}),

			// isChart checks if an object is a PPTX chart
			"isChart": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isChart takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.PPTXChart)
				return Bool(ok)
			}),

			// isImage checks if an object is a PPTX image
			"isImage": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isImage takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.PPTXImage)
				return Bool(ok)
			}),

			// EMU conversion helpers
			// inchesToEMU converts inches to EMUs
			"inchesToEMU": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("inchesToEMU takes exactly 1 argument")
				}
				val, ok := args[0].(*objects.Float)
				if !ok {
					if intVal, ok := args[0].(*objects.Int); ok {
						return Int(int64(float64(intVal.Value) * 914400))
					}
					return Error("value must be a number")
				}
				return Int(int64(val.Value * 914400))
			}),

			// emuToInches converts EMUs to inches
			"emuToInches": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("emuToInches takes exactly 1 argument")
				}
				val, ok := args[0].(*objects.Int)
				if !ok {
					return Error("value must be an integer")
				}
				return Float(float64(val.Value) / 914400)
			}),

			// pointsToEMU converts points to EMUs
			"pointsToEMU": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("pointsToEMU takes exactly 1 argument")
				}
				val, ok := args[0].(*objects.Float)
				if !ok {
					if intVal, ok := args[0].(*objects.Int); ok {
						return Int(int64(float64(intVal.Value) * 12700))
					}
					return Error("value must be a number")
				}
				return Int(int64(val.Value * 12700))
			}),

			// emuToPoints converts EMUs to points
			"emuToPoints": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("emuToPoints takes exactly 1 argument")
				}
				val, ok := args[0].(*objects.Int)
				if !ok {
					return Error("value must be an integer")
				}
				return Float(float64(val.Value) / 12700)
			}),

			// pixelsToEMU converts pixels to EMUs (at 96 DPI)
			"pixelsToEMU": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("pixelsToEMU takes exactly 1 argument")
				}
				val, ok := args[0].(*objects.Float)
				if !ok {
					if intVal, ok := args[0].(*objects.Int); ok {
						return Int(int64(float64(intVal.Value) * 9525))
					}
					return Error("value must be a number")
				}
				return Int(int64(val.Value * 9525))
			}),

			// emuToPixels converts EMUs to pixels (at 96 DPI)
			"emuToPixels": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("emuToPixels takes exactly 1 argument")
				}
				val, ok := args[0].(*objects.Int)
				if !ok {
					return Error("value must be an integer")
				}
				return Float(float64(val.Value) / 9525)
			}),
		},
	})
}
