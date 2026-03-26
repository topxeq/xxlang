// pkg/objects/methods_pptx.go
// PPTX methods for Xxlang.
package objects

import "fmt"

// ============================================================
// PPTX Document Methods
// ============================================================

var pptxDocumentMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"close": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for close. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for close must be PPTXDocument, got %s", args[0].Type())
		}
		self.Close()
		return NULL
	}},
	"getSlideCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSlideCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for getSlideCount must be PPTXDocument, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetSlideCount()))
	}},
	"getSlide": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for getSlide. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for getSlide must be PPTXDocument, got %s", args[0].Type())
		}
		index, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT")
		}
		slide := self.GetSlide(int(index.Value))
		if slide == nil {
			return NULL
		}
		return slide
	}},
	"addSlide": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for addSlide. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for addSlide must be PPTXDocument, got %s", args[0].Type())
		}
		return self.AddSlide()
	}},
	"deleteSlide": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for deleteSlide. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for deleteSlide must be PPTXDocument, got %s", args[0].Type())
		}
		index, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT")
		}
		return &Bool{Value: self.DeleteSlide(int(index.Value))}
	}},
	"moveSlide": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for moveSlide. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for moveSlide must be PPTXDocument, got %s", args[0].Type())
		}
		from, ok := args[1].(*Int)
		if !ok {
			return newError("from must be INT")
		}
		to, ok := args[2].(*Int)
		if !ok {
			return newError("to must be INT")
		}
		return &Bool{Value: self.MoveSlide(int(from.Value), int(to.Value))}
	}},
	"duplicateSlide": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for duplicateSlide. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for duplicateSlide must be PPTXDocument, got %s", args[0].Type())
		}
		index, ok := args[1].(*Int)
		if !ok {
			return newError("index must be INT")
		}
		slide := self.DuplicateSlide(int(index.Value))
		if slide == nil {
			return NULL
		}
		return slide
	}},
	"getProperties": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getProperties. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for getProperties must be PPTXDocument, got %s", args[0].Type())
		}
		props := self.GetProperties()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "title", NewString(props.Title))
		setMapKey(m, "subject", NewString(props.Subject))
		setMapKey(m, "author", NewString(props.Author))
		setMapKey(m, "keywords", NewString(props.Keywords))
		setMapKey(m, "description", NewString(props.Description))
		return m
	}},
	"setProperties": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setProperties. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for setProperties must be PPTXDocument, got %s", args[0].Type())
		}
		propsMap, ok := args[1].(*Map)
		if !ok {
			return newError("properties must be MAP")
		}
		props := &PPTXProperties{}
		if v := getMapValue(propsMap, "title"); v != NULL {
			props.Title = v.(*String).Value
		}
		if v := getMapValue(propsMap, "subject"); v != NULL {
			props.Subject = v.(*String).Value
		}
		if v := getMapValue(propsMap, "author"); v != NULL {
			props.Author = v.(*String).Value
		}
		if v := getMapValue(propsMap, "keywords"); v != NULL {
			props.Keywords = v.(*String).Value
		}
		if v := getMapValue(propsMap, "description"); v != NULL {
			props.Description = v.(*String).Value
		}
		self.SetProperties(props)
		return NULL
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return newError("wrong number of arguments for save. got=%d, want>=1", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for save must be PPTXDocument, got %s", args[0].Type())
		}
		if len(args) >= 2 {
			path, ok := args[1].(*String)
			if !ok {
				return newError("path must be STRING")
			}
			if err := self.Save(path.Value); err != nil {
				return newError("failed to save: %s", err.Error())
			}
		} else {
			if err := self.Save(""); err != nil {
				return newError("failed to save: %s", err.Error())
			}
		}
		return NULL
	}},
	"toBytes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for toBytes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXDocument)
		if !ok {
			return newError("receiver for toBytes must be PPTXDocument, got %s", args[0].Type())
		}
		data, err := self.ToBytes()
		if err != nil {
			return newError("failed to convert to bytes: %s", err.Error())
		}
		return NewBytes(data)
	}},
}

// ============================================================
// PPTX Slide Methods
// ============================================================

var pptxSlideMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"getIndex": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getIndex. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getIndex must be PPTXSlide, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetIndex()))
	}},
	"getTexts": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getTexts. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getTexts must be PPTXSlide, got %s", args[0].Type())
		}
		texts := self.GetTexts()
		elements := make([]Object, len(texts))
		for i, t := range texts {
			elements[i] = t
		}
		return &Array{Elements: elements}
	}},
	"getShapes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getShapes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getShapes must be PPTXSlide, got %s", args[0].Type())
		}
		shapes := self.GetShapes()
		elements := make([]Object, len(shapes))
		for i, s := range shapes {
			elements[i] = s
		}
		return &Array{Elements: elements}
	}},
	"getImages": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getImages. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getImages must be PPTXSlide, got %s", args[0].Type())
		}
		images := self.GetImages()
		elements := make([]Object, len(images))
		for i, img := range images {
			elements[i] = img
		}
		return &Array{Elements: elements}
	}},
	"getTables": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getTables. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getTables must be PPTXSlide, got %s", args[0].Type())
		}
		tables := self.GetTables()
		elements := make([]Object, len(tables))
		for i, t := range tables {
			elements[i] = t
		}
		return &Array{Elements: elements}
	}},
	"getCharts": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getCharts. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getCharts must be PPTXSlide, got %s", args[0].Type())
		}
		charts := self.GetCharts()
		elements := make([]Object, len(charts))
		for i, c := range charts {
			elements[i] = c
		}
		return &Array{Elements: elements}
	}},
	"getAllText": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getAllText. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getAllText must be PPTXSlide, got %s", args[0].Type())
		}
		return NewString(self.GetAllText())
	}},
	"addText": {Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for addText. got=%d, want>=2", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for addText must be PPTXSlide, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("text must be STRING")
		}
		options := make(map[string]interface{})
		if len(args) >= 3 {
			if opts, ok := args[2].(*Map); ok {
				for _, pair := range opts.Pairs {
					key := pair.Key.(*String).Value
					switch val := pair.Value.(type) {
					case *Int:
						options[key] = val.Value
					case *Float:
						options[key] = val.Value
					case *String:
						options[key] = val.Value
					case *Bool:
						options[key] = val.Value
					}
				}
			}
		}
		return self.AddText(text.Value, options)
	}},
	"addShape": {Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for addShape. got=%d, want>=2", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for addShape must be PPTXSlide, got %s", args[0].Type())
		}
		shapeType, ok := args[1].(*String)
		if !ok {
			return newError("shapeType must be STRING")
		}
		options := make(map[string]interface{})
		if len(args) >= 3 {
			if opts, ok := args[2].(*Map); ok {
				for _, pair := range opts.Pairs {
					key := pair.Key.(*String).Value
					switch val := pair.Value.(type) {
					case *Int:
						options[key] = val.Value
					case *Float:
						options[key] = val.Value
					case *String:
						options[key] = val.Value
					case *Bool:
						options[key] = val.Value
					}
				}
			}
		}
		return self.AddShape(PPTXShapeKind(shapeType.Value), options)
	}},
	"addImage": {Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return newError("wrong number of arguments for addImage. got=%d, want>=2", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for addImage must be PPTXSlide, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		options := make(map[string]interface{})
		if len(args) >= 3 {
			if opts, ok := args[2].(*Map); ok {
				for _, pair := range opts.Pairs {
					key := pair.Key.(*String).Value
					switch val := pair.Value.(type) {
					case *Int:
						options[key] = val.Value
					case *Float:
						options[key] = val.Value
					case *String:
						options[key] = val.Value
					case *Bool:
						options[key] = val.Value
					}
				}
			}
		}
		img := self.AddImage(path.Value, options)
		if img == nil {
			return NULL
		}
		return img
	}},
	"addTable": {Fn: func(args ...Object) Object {
		if len(args) < 3 {
			return newError("wrong number of arguments for addTable. got=%d, want>=3", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for addTable must be PPTXSlide, got %s", args[0].Type())
		}
		rows, ok := args[1].(*Int)
		if !ok {
			return newError("rows must be INT")
		}
		cols, ok := args[2].(*Int)
		if !ok {
			return newError("cols must be INT")
		}
		options := make(map[string]interface{})
		if len(args) >= 4 {
			if opts, ok := args[3].(*Map); ok {
				for _, pair := range opts.Pairs {
					key := pair.Key.(*String).Value
					switch val := pair.Value.(type) {
					case *Int:
						options[key] = val.Value
					case *Float:
						options[key] = val.Value
					case *String:
						options[key] = val.Value
					case *Bool:
						options[key] = val.Value
					}
				}
			}
		}
		return self.AddTable(int(rows.Value), int(cols.Value), options)
	}},
	"getNotes": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getNotes. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for getNotes must be PPTXSlide, got %s", args[0].Type())
		}
		return NewString(self.GetNotes())
	}},
	"setNotes": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setNotes. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXSlide)
		if !ok {
			return newError("receiver for setNotes must be PPTXSlide, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("text must be STRING")
		}
		self.SetNotes(text.Value)
		return NULL
	}},
}

// ============================================================
// PPTX TextFrame Methods
// ============================================================

var pptxTextFrameMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"getText": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getText. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextFrame)
		if !ok {
			return newError("receiver for getText must be PPTXTextFrame, got %s", args[0].Type())
		}
		return NewString(self.GetText())
	}},
	"setText": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setText. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXTextFrame)
		if !ok {
			return newError("receiver for setText must be PPTXTextFrame, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("text must be STRING")
		}
		self.SetText(text.Value)
		return NULL
	}},
	"getParagraphs": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getParagraphs. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextFrame)
		if !ok {
			return newError("receiver for getParagraphs must be PPTXTextFrame, got %s", args[0].Type())
		}
		paragraphs := self.GetParagraphs()
		elements := make([]Object, len(paragraphs))
		for i, p := range paragraphs {
			elements[i] = p
		}
		return &Array{Elements: elements}
	}},
	"getPosition": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPosition. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextFrame)
		if !ok {
			return newError("receiver for getPosition must be PPTXTextFrame, got %s", args[0].Type())
		}
		pos := self.GetPosition()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "x", NewInt(pos.X))
		setMapKey(m, "y", NewInt(pos.Y))
		return m
	}},
	"setPosition": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setPosition. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXTextFrame)
		if !ok {
			return newError("receiver for setPosition must be PPTXTextFrame, got %s", args[0].Type())
		}
		x, ok := args[1].(*Int)
		if !ok {
			return newError("x must be INT")
		}
		y, ok := args[2].(*Int)
		if !ok {
			return newError("y must be INT")
		}
		self.SetPosition(x.Value, y.Value)
		return NULL
	}},
	"getSize": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSize. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextFrame)
		if !ok {
			return newError("receiver for getSize must be PPTXTextFrame, got %s", args[0].Type())
		}
		size := self.GetSize()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "width", NewInt(size.Width))
		setMapKey(m, "height", NewInt(size.Height))
		return m
	}},
	"setSize": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setSize. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXTextFrame)
		if !ok {
			return newError("receiver for setSize must be PPTXTextFrame, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("width must be INT")
		}
		height, ok := args[2].(*Int)
		if !ok {
			return newError("height must be INT")
		}
		self.SetSize(width.Value, height.Value)
		return NULL
	}},
}

// ============================================================
// PPTX TextRun Methods
// ============================================================

var pptxTextRunMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"getText": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getText. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for getText must be PPTXTextRun, got %s", args[0].Type())
		}
		return NewString(self.GetText())
	}},
	"setText": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setText. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for setText must be PPTXTextRun, got %s", args[0].Type())
		}
		text, ok := args[1].(*String)
		if !ok {
			return newError("text must be STRING")
		}
		self.SetText(text.Value)
		return NULL
	}},
	"getFontName": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getFontName. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for getFontName must be PPTXTextRun, got %s", args[0].Type())
		}
		return NewString(self.GetFontName())
	}},
	"setFontName": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setFontName. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for setFontName must be PPTXTextRun, got %s", args[0].Type())
		}
		name, ok := args[1].(*String)
		if !ok {
			return newError("name must be STRING")
		}
		self.SetFontName(name.Value)
		return NULL
	}},
	"getFontSize": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getFontSize. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for getFontSize must be PPTXTextRun, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetFontSize()))
	}},
	"setFontSize": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setFontSize. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for setFontSize must be PPTXTextRun, got %s", args[0].Type())
		}
		size, ok := args[1].(*Int)
		if !ok {
			return newError("size must be INT")
		}
		self.SetFontSize(int(size.Value))
		return NULL
	}},
	"isBold": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isBold. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for isBold must be PPTXTextRun, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsBold()}
	}},
	"setBold": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setBold. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for setBold must be PPTXTextRun, got %s", args[0].Type())
		}
		bold, ok := args[1].(*Bool)
		if !ok {
			return newError("bold must be BOOL")
		}
		self.SetBold(bold.Value)
		return NULL
	}},
	"isItalic": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for isItalic. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for isItalic must be PPTXTextRun, got %s", args[0].Type())
		}
		return &Bool{Value: self.IsItalic()}
	}},
	"setItalic": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setItalic. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for setItalic must be PPTXTextRun, got %s", args[0].Type())
		}
		italic, ok := args[1].(*Bool)
		if !ok {
			return newError("italic must be BOOL")
		}
		self.SetItalic(italic.Value)
		return NULL
	}},
	"getColor": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getColor. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for getColor must be PPTXTextRun, got %s", args[0].Type())
		}
		return NewString(self.GetColor())
	}},
	"setColor": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setColor. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXTextRun)
		if !ok {
			return newError("receiver for setColor must be PPTXTextRun, got %s", args[0].Type())
		}
		color, ok := args[1].(*String)
		if !ok {
			return newError("color must be STRING")
		}
		self.SetColor(color.Value)
		return NULL
	}},
}

// ============================================================
// PPTX Shape Methods
// ============================================================

var pptxShapeMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"getType": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getType. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for getType must be PPTXShape, got %s", args[0].Type())
		}
		return NewString(string(self.GetKind()))
	}},
	"getPosition": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPosition. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for getPosition must be PPTXShape, got %s", args[0].Type())
		}
		pos := self.GetPosition()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "x", NewInt(pos.X))
		setMapKey(m, "y", NewInt(pos.Y))
		return m
	}},
	"setPosition": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setPosition. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for setPosition must be PPTXShape, got %s", args[0].Type())
		}
		x, ok := args[1].(*Int)
		if !ok {
			return newError("x must be INT")
		}
		y, ok := args[2].(*Int)
		if !ok {
			return newError("y must be INT")
		}
		self.SetPosition(x.Value, y.Value)
		return NULL
	}},
	"getSize": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSize. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for getSize must be PPTXShape, got %s", args[0].Type())
		}
		size := self.GetSize()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "width", NewInt(size.Width))
		setMapKey(m, "height", NewInt(size.Height))
		return m
	}},
	"setSize": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setSize. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for setSize must be PPTXShape, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("width must be INT")
		}
		height, ok := args[2].(*Int)
		if !ok {
			return newError("height must be INT")
		}
		self.SetSize(width.Value, height.Value)
		return NULL
	}},
	"getFill": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getFill. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for getFill must be PPTXShape, got %s", args[0].Type())
		}
		fill := self.GetFill()
		if fill == nil {
			return NULL
		}
		return NewString(fmt.Sprintf("%02X%02X%02X", fill.R, fill.G, fill.B))
	}},
	"setFill": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setFill. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for setFill must be PPTXShape, got %s", args[0].Type())
		}
		color, ok := args[1].(*String)
		if !ok {
			return newError("color must be STRING")
		}
		self.SetFill(color.Value)
		return NULL
	}},
	"getTextFrame": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getTextFrame. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for getTextFrame must be PPTXShape, got %s", args[0].Type())
		}
		tf := self.GetTextFrame()
		if tf == nil {
			return NULL
		}
		return tf
	}},
	"addTextFrame": {Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return newError("wrong number of arguments for addTextFrame. got=%d, want>=1", len(args))
		}
		self, ok := args[0].(*PPTXShape)
		if !ok {
			return newError("receiver for addTextFrame must be PPTXShape, got %s", args[0].Type())
		}
		text := ""
		if len(args) >= 2 {
			if t, ok := args[1].(*String); ok {
				text = t.Value
			}
		}
		return self.AddTextFrame(text)
	}},
}

// ============================================================
// PPTX Table Methods
// ============================================================

var pptxTableMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"getRowCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getRowCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for getRowCount must be PPTXTable, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetRowCount()))
	}},
	"getColCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getColCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for getColCount must be PPTXTable, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetColCount()))
	}},
	"getValue": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for getValue. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for getValue must be PPTXTable, got %s", args[0].Type())
		}
		row, ok := args[1].(*Int)
		if !ok {
			return newError("row must be INT")
		}
		col, ok := args[2].(*Int)
		if !ok {
			return newError("col must be INT")
		}
		return NewString(self.GetValue(int(row.Value), int(col.Value)))
	}},
	"setValue": {Fn: func(args ...Object) Object {
		if len(args) != 4 {
			return newError("wrong number of arguments for setValue. got=%d, want=4", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for setValue must be PPTXTable, got %s", args[0].Type())
		}
		row, ok := args[1].(*Int)
		if !ok {
			return newError("row must be INT")
		}
		col, ok := args[2].(*Int)
		if !ok {
			return newError("col must be INT")
		}
		value, ok := args[3].(*String)
		if !ok {
			return newError("value must be STRING")
		}
		self.SetValue(int(row.Value), int(col.Value), value.Value)
		return NULL
	}},
	"getPosition": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPosition. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for getPosition must be PPTXTable, got %s", args[0].Type())
		}
		pos := self.GetPosition()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "x", NewInt(pos.X))
		setMapKey(m, "y", NewInt(pos.Y))
		return m
	}},
	"setPosition": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setPosition. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for setPosition must be PPTXTable, got %s", args[0].Type())
		}
		x, ok := args[1].(*Int)
		if !ok {
			return newError("x must be INT")
		}
		y, ok := args[2].(*Int)
		if !ok {
			return newError("y must be INT")
		}
		self.SetPosition(x.Value, y.Value)
		return NULL
	}},
	"getSize": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSize. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for getSize must be PPTXTable, got %s", args[0].Type())
		}
		size := self.GetSize()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "width", NewInt(size.Width))
		setMapKey(m, "height", NewInt(size.Height))
		return m
	}},
	"setSize": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setSize. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXTable)
		if !ok {
			return newError("receiver for setSize must be PPTXTable, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("width must be INT")
		}
		height, ok := args[2].(*Int)
		if !ok {
			return newError("height must be INT")
		}
		self.SetSize(width.Value, height.Value)
		return NULL
	}},
}

// ============================================================
// PPTX Chart Methods
// ============================================================

var pptxChartMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"getType": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getType. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for getType must be PPTXChart, got %s", args[0].Type())
		}
		return NewString(string(self.GetKind()))
	}},
	"getTitle": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getTitle. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for getTitle must be PPTXChart, got %s", args[0].Type())
		}
		return NewString(self.GetTitle())
	}},
	"setTitle": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for setTitle. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for setTitle must be PPTXChart, got %s", args[0].Type())
		}
		title, ok := args[1].(*String)
		if !ok {
			return newError("title must be STRING")
		}
		self.SetTitle(title.Value)
		return NULL
	}},
	"getSeriesCount": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSeriesCount. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for getSeriesCount must be PPTXChart, got %s", args[0].Type())
		}
		return NewInt(int64(self.GetSeriesCount()))
	}},
	"getPosition": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPosition. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for getPosition must be PPTXChart, got %s", args[0].Type())
		}
		pos := self.GetPosition()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "x", NewInt(pos.X))
		setMapKey(m, "y", NewInt(pos.Y))
		return m
	}},
	"setPosition": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setPosition. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for setPosition must be PPTXChart, got %s", args[0].Type())
		}
		x, ok := args[1].(*Int)
		if !ok {
			return newError("x must be INT")
		}
		y, ok := args[2].(*Int)
		if !ok {
			return newError("y must be INT")
		}
		self.SetPosition(x.Value, y.Value)
		return NULL
	}},
	"getSize": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSize. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for getSize must be PPTXChart, got %s", args[0].Type())
		}
		size := self.GetSize()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "width", NewInt(size.Width))
		setMapKey(m, "height", NewInt(size.Height))
		return m
	}},
	"setSize": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setSize. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXChart)
		if !ok {
			return newError("receiver for setSize must be PPTXChart, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("width must be INT")
		}
		height, ok := args[2].(*Int)
		if !ok {
			return newError("height must be INT")
		}
		self.SetSize(width.Value, height.Value)
		return NULL
	}},
}

// ============================================================
// PPTX Image Methods
// ============================================================

var pptxImageMethods = map[string]*Builtin{
	"typeOf": {Fn: universalTypeOf},
	"toStr":  {Fn: universalToStr},
	"getData": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getData. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for getData must be PPTXImage, got %s", args[0].Type())
		}
		return NewBytes(self.GetData())
	}},
	"getDataBase64": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getDataBase64. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for getDataBase64 must be PPTXImage, got %s", args[0].Type())
		}
		return NewString(self.GetDataBase64())
	}},
	"save": {Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("wrong number of arguments for save. got=%d, want=2", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for save must be PPTXImage, got %s", args[0].Type())
		}
		path, ok := args[1].(*String)
		if !ok {
			return newError("path must be STRING")
		}
		if err := self.Save(path.Value); err != nil {
			return newError("failed to save image: %s", err.Error())
		}
		return NULL
	}},
	"getFormat": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getFormat. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for getFormat must be PPTXImage, got %s", args[0].Type())
		}
		return NewString(self.GetFormat())
	}},
	"getPosition": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getPosition. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for getPosition must be PPTXImage, got %s", args[0].Type())
		}
		pos := self.GetPosition()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "x", NewInt(pos.X))
		setMapKey(m, "y", NewInt(pos.Y))
		return m
	}},
	"setPosition": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setPosition. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for setPosition must be PPTXImage, got %s", args[0].Type())
		}
		x, ok := args[1].(*Int)
		if !ok {
			return newError("x must be INT")
		}
		y, ok := args[2].(*Int)
		if !ok {
			return newError("y must be INT")
		}
		self.SetPosition(x.Value, y.Value)
		return NULL
	}},
	"getSize": {Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("wrong number of arguments for getSize. got=%d, want=1", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for getSize must be PPTXImage, got %s", args[0].Type())
		}
		size := self.GetSize()
		m := &Map{Pairs: make(map[HashKey]MapPair)}
		setMapKey(m, "width", NewInt(size.Width))
		setMapKey(m, "height", NewInt(size.Height))
		return m
	}},
	"setSize": {Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("wrong number of arguments for setSize. got=%d, want=3", len(args))
		}
		self, ok := args[0].(*PPTXImage)
		if !ok {
			return newError("receiver for setSize must be PPTXImage, got %s", args[0].Type())
		}
		width, ok := args[1].(*Int)
		if !ok {
			return newError("width must be INT")
		}
		height, ok := args[2].(*Int)
		if !ok {
			return newError("height must be INT")
		}
		self.SetSize(width.Value, height.Value)
		return NULL
	}},
}

// Helper functions for PPTX methods

func setMapKey(m *Map, key string, value Object) {
	k := NewString(key)
	m.Pairs[k.HashKey()] = MapPair{Key: k, Value: value}
}

func getMapValue(m *Map, key string) Object {
	k := NewString(key)
	if pair, ok := m.Pairs[k.HashKey()]; ok {
		return pair.Value
	}
	return NULL
}
