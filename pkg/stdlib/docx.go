// pkg/stdlib/docx.go
// DOCX module for Xxlang - Microsoft Word document processing.
package stdlib

import (
	"fmt"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "docx",
		Exports: map[string]objects.Object{
			// ============================================
			// File Operations
			// ============================================

			// open(path) - Opens a docx file
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("open() takes exactly 1 argument")
				}
				path, ok := args[0].(*objects.String)
				if !ok {
					return Error("open() requires a string path argument")
				}

				doc, err := objects.OpenDocx(path.Value)
				if err != nil {
					return Error(err.Error())
				}
				return doc
			}),

			// create() - Creates a new empty document
			"create": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 0 {
					return Error("create() takes no arguments")
				}
				return objects.NewDocxDocument()
			}),

			// save(doc, path) - Saves document to file
			"save": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("save() takes exactly 2 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("save() requires a DocxDocument as first argument")
				}
				path, ok := args[1].(*objects.String)
				if !ok {
					return Error("save() requires a string path as second argument")
				}

				if err := doc.Save(path.Value); err != nil {
					return Error(err.Error())
				}
				return objects.NULL
			}),

			// toBytes(doc) - Converts document to byte array
			"toBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("toBytes() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("toBytes() requires a DocxDocument argument")
				}

				data, err := doc.ToBytes()
				if err != nil {
					return Error(err.Error())
				}
				return &objects.Bytes{Value: data}
			}),

			// fromBytes(data) - Loads document from byte array
			"fromBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("fromBytes() takes exactly 1 argument")
				}
				data, ok := args[0].(*objects.Bytes)
				if !ok {
					return Error("fromBytes() requires a Bytes argument")
				}

				doc, err := objects.OpenDocxFromBytes(data.Value)
				if err != nil {
					return Error(err.Error())
				}
				return doc
			}),

			// close(doc) - Closes document and releases resources
			"close": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("close() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("close() requires a DocxDocument argument")
				}

				if err := doc.Close(); err != nil {
					return Error(err.Error())
				}
				return objects.NULL
			}),

			// ============================================
			// Paragraph Operations
			// ============================================

			// addParagraph(doc, textOrConfig) - Adds a paragraph
			"addParagraph": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return Error("addParagraph() takes 1 or 2 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("addParagraph() requires a DocxDocument as first argument")
				}

				body := doc.GetBody()
				if body == nil {
					return Error("document has no body element")
				}

				// Create paragraph node
				para := objects.NewXMLNode("w:p")

				// Handle text argument
				if len(args) == 2 {
					switch arg := args[1].(type) {
					case *objects.String:
						// Add text run
						run := objects.NewXMLNode("w:r")
						text := objects.NewXMLNode("w:t")
						text.SetText(arg.Value)
						// Preserve spaces
						text.SetAttr("xml:space", "preserve")
						run.AddChild(text)
						para.AddChild(run)

					case *objects.Map:
						// Handle config map
						text := getMapString(arg, "text", "")
						if text != "" {
							run := objects.NewXMLNode("w:r")
							textNode := objects.NewXMLNode("w:t")
							textNode.SetText(text)
							textNode.SetAttr("xml:space", "preserve")
							run.AddChild(textNode)
							para.AddChild(run)
						}

						// Handle alignment
						align := getMapString(arg, "align", "")
						if align != "" {
							pPr := objects.NewXMLNode("w:pPr")
							jc := objects.NewXMLNode("w:jc")
							jc.SetAttr("w:val", align)
							pPr.AddChild(jc)
							para.AddChild(pPr)
						}

						// Handle style
						style := getMapString(arg, "style", "")
						if style != "" {
							pPr := para.FindFirst("w:pPr")
							if pPr == nil {
								pPr = objects.NewXMLNode("w:pPr")
								para.AddChild(pPr)
							}
							styleNode := objects.NewXMLNode("w:pStyle")
							styleNode.SetAttr("w:val", style)
							pPr.AddChild(styleNode)
						}
					}
				}

				body.AddChild(para)
				doc.SetModified(true)

				return &objects.DocxParagraph{
					Document: doc,
					XmlNode:  para,
				}
			}),

			// getParagraphs(doc) - Gets all paragraphs
			"getParagraphs": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getParagraphs() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("getParagraphs() requires a DocxDocument argument")
				}

				body := doc.GetBody()
				if body == nil {
					return Array()
				}

				// Find all paragraph nodes
				paraNodes := body.Find("//w:p")
				elements := paraNodes.Elements

				result := make([]objects.Object, len(elements))
				for i, elem := range elements {
					if node, ok := elem.(*objects.XMLNode); ok {
						result[i] = &objects.DocxParagraph{
							Document: doc,
							XmlNode:  node,
						}
					}
				}

				return Array(result...)
			}),

			// setAlignment(para, align) - Sets paragraph alignment
			"setAlignment": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setAlignment() takes exactly 2 arguments")
				}
				para, ok := args[0].(*objects.DocxParagraph)
				if !ok {
					return Error("setAlignment() requires a DocxParagraph as first argument")
				}
				align, ok := args[1].(*objects.String)
				if !ok {
					return Error("setAlignment() requires a string alignment as second argument")
				}

				if err := setParagraphProperty(para, "w:jc", align.Value); err != nil {
					return Error(err.Error())
				}
				return objects.NULL
			}),

			// setSpacing(para, config) - Sets paragraph spacing
			"setSpacing": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setSpacing() takes exactly 2 arguments")
				}
				para, ok := args[0].(*objects.DocxParagraph)
				if !ok {
					return Error("setSpacing() requires a DocxParagraph as first argument")
				}
				config, ok := args[1].(*objects.Map)
				if !ok {
					return Error("setSpacing() requires a Map config as second argument")
				}

				pPr := getOrCreatePPr(para)
				spacing := objects.NewXMLNode("w:spacing")

				if before := getMapInt(config, "before", -1); before >= 0 {
					spacing.SetAttr("w:before", intToStr(before))
				}
				if after := getMapInt(config, "after", -1); after >= 0 {
					spacing.SetAttr("w:after", intToStr(after))
				}
				if line := getMapInt(config, "line", -1); line >= 0 {
					spacing.SetAttr("w:line", intToStr(line))
					lineRule := getMapString(config, "lineRule", "auto")
					spacing.SetAttr("w:lineRule", lineRule)
				}

				pPr.AddChild(spacing)
				return objects.NULL
			}),

			// ============================================
			// Run Operations
			// ============================================

			// addRun(para, textOrConfig) - Adds a text run to paragraph
			"addRun": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return Error("addRun() takes 1 or 2 arguments")
				}
				para, ok := args[0].(*objects.DocxParagraph)
				if !ok {
					return Error("addRun() requires a DocxParagraph as first argument")
				}

				run := objects.NewXMLNode("w:r")

				if len(args) == 2 {
					switch arg := args[1].(type) {
					case *objects.String:
						text := objects.NewXMLNode("w:t")
						text.SetText(arg.Value)
						text.SetAttr("xml:space", "preserve")
						run.AddChild(text)

					case *objects.Map:
						text := getMapString(arg, "text", "")
						if text != "" {
							textNode := objects.NewXMLNode("w:t")
							textNode.SetText(text)
							textNode.SetAttr("xml:space", "preserve")
							run.AddChild(textNode)
						}

						// Formatting properties
						bold := getMapBool(arg, "bold", false)
						italic := getMapBool(arg, "italic", false)
						underline := getMapString(arg, "underline", "")
						fontSize := getMapInt(arg, "fontSize", 0)
						fontName := getMapString(arg, "fontName", "")
						color := getMapString(arg, "color", "")

						if bold || italic || underline != "" || fontSize > 0 || fontName != "" || color != "" {
							rPr := objects.NewXMLNode("w:rPr")

							if bold {
								b := objects.NewXMLNode("w:b")
								rPr.AddChild(b)
							}
							if italic {
								i := objects.NewXMLNode("w:i")
								rPr.AddChild(i)
							}
							if underline != "" {
								u := objects.NewXMLNode("w:u")
								u.SetAttr("w:val", underline)
								rPr.AddChild(u)
							}
							if fontSize > 0 {
								sz := objects.NewXMLNode("w:sz")
								sz.SetAttr("w:val", intToStr(fontSize*2)) // Convert to half-points
								rPr.AddChild(sz)
							}
							if fontName != "" {
								rFonts := objects.NewXMLNode("w:rFonts")
								rFonts.SetAttr("w:ascii", fontName)
								rFonts.SetAttr("w:hAnsi", fontName)
								rPr.AddChild(rFonts)
							}
							if color != "" {
								c := objects.NewXMLNode("w:color")
								c.SetAttr("w:val", color)
								rPr.AddChild(c)
							}

							run.AddChild(rPr)
						}
					}
				}

				para.XmlNode.AddChild(run)

				return &objects.DocxRun{
					Paragraph: para,
					XmlNode:   run,
				}
			}),

			// getRuns(para) - Gets all runs in paragraph
			"getRuns": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getRuns() takes exactly 1 argument")
				}
				para, ok := args[0].(*objects.DocxParagraph)
				if !ok {
					return Error("getRuns() requires a DocxParagraph argument")
				}

				runNodes := para.XmlNode.Find("//w:r")
				elements := runNodes.Elements

				result := make([]objects.Object, len(elements))
				for i, elem := range elements {
					if node, ok := elem.(*objects.XMLNode); ok {
						result[i] = &objects.DocxRun{
							Paragraph: para,
							XmlNode:   node,
						}
					}
				}

				return Array(result...)
			}),

			// setBold(run, bool) - Sets bold formatting
			"setBold": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setBold() takes exactly 2 arguments")
				}
				run, ok := args[0].(*objects.DocxRun)
				if !ok {
					return Error("setBold() requires a DocxRun as first argument")
				}
				bold, ok := args[1].(*objects.Bool)
				if !ok {
					return Error("setBold() requires a Bool as second argument")
				}

				setRunProperty(run, "w:b", bold.Value)
				return objects.NULL
			}),

			// setItalic(run, bool) - Sets italic formatting
			"setItalic": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setItalic() takes exactly 2 arguments")
				}
				run, ok := args[0].(*objects.DocxRun)
				if !ok {
					return Error("setItalic() requires a DocxRun as first argument")
				}
				italic, ok := args[1].(*objects.Bool)
				if !ok {
					return Error("setItalic() requires a Bool as second argument")
				}

				setRunProperty(run, "w:i", italic.Value)
				return objects.NULL
			}),

			// setUnderline(run, style) - Sets underline style
			"setUnderline": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setUnderline() takes exactly 2 arguments")
				}
				run, ok := args[0].(*objects.DocxRun)
				if !ok {
					return Error("setUnderline() requires a DocxRun as first argument")
				}
				style, ok := args[1].(*objects.String)
				if !ok {
					return Error("setUnderline() requires a String as second argument")
				}

				rPr := getOrCreateRPr(run)
				u := objects.NewXMLNode("w:u")
				u.SetAttr("w:val", style.Value)
				rPr.AddChild(u)
				return objects.NULL
			}),

			// setFontSize(run, size) - Sets font size (in points)
			"setFontSize": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setFontSize() takes exactly 2 arguments")
				}
				run, ok := args[0].(*objects.DocxRun)
				if !ok {
					return Error("setFontSize() requires a DocxRun as first argument")
				}
				size, ok := args[1].(*objects.Int)
				if !ok {
					return Error("setFontSize() requires an Int as second argument")
				}

				rPr := getOrCreateRPr(run)
				sz := objects.NewXMLNode("w:sz")
				sz.SetAttr("w:val", intToStr(int(size.Value*2))) // Convert to half-points
				rPr.AddChild(sz)
				return objects.NULL
			}),

			// setFontName(run, name) - Sets font name
			"setFontName": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setFontName() takes exactly 2 arguments")
				}
				run, ok := args[0].(*objects.DocxRun)
				if !ok {
					return Error("setFontName() requires a DocxRun as first argument")
				}
				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("setFontName() requires a String as second argument")
				}

				rPr := getOrCreateRPr(run)
				rFonts := objects.NewXMLNode("w:rFonts")
				rFonts.SetAttr("w:ascii", name.Value)
				rFonts.SetAttr("w:hAnsi", name.Value)
				rPr.AddChild(rFonts)
				return objects.NULL
			}),

			// setColor(run, color) - Sets text color
			"setColor": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setColor() takes exactly 2 arguments")
				}
				run, ok := args[0].(*objects.DocxRun)
				if !ok {
					return Error("setColor() requires a DocxRun as first argument")
				}
				color, ok := args[1].(*objects.String)
				if !ok {
					return Error("setColor() requires a String as second argument")
				}

				rPr := getOrCreateRPr(run)
				c := objects.NewXMLNode("w:color")
				c.SetAttr("w:val", color.Value)
				rPr.AddChild(c)
				return objects.NULL
			}),

			// ============================================
			// Table Operations
			// ============================================

			// addTable(doc, rows, cols) - Adds a table
			"addTable": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("addTable() takes exactly 3 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("addTable() requires a DocxDocument as first argument")
				}
				rows, ok := args[1].(*objects.Int)
				if !ok {
					return Error("addTable() requires an Int rows as second argument")
				}
				cols, ok := args[2].(*objects.Int)
				if !ok {
					return Error("addTable() requires an Int cols as third argument")
				}

				body := doc.GetBody()
				if body == nil {
					return Error("document has no body element")
				}

				// Create table
				tbl := objects.NewXMLNode("w:tbl")

				// Add table properties
				tblPr := objects.NewXMLNode("w:tblPr")
				tblW := objects.NewXMLNode("w:tblW")
				tblW.SetAttr("w:w", "0")
				tblW.SetAttr("w:type", "auto")
				tblPr.AddChild(tblW)
				tbl.AddChild(tblPr)

				// Add rows and cells
				for r := 0; r < int(rows.Value); r++ {
					tr := objects.NewXMLNode("w:tr")
					for c := 0; c < int(cols.Value); c++ {
						tc := objects.NewXMLNode("w:tc")
						// Add empty paragraph in cell
						p := objects.NewXMLNode("w:p")
						tc.AddChild(p)
						tr.AddChild(tc)
					}
					tbl.AddChild(tr)
				}

				body.AddChild(tbl)
				doc.SetModified(true)

				return &objects.DocxTable{
					Document: doc,
					XmlNode:  tbl,
					Rows:     int(rows.Value),
					Cols:     int(cols.Value),
				}
			}),

			// getTables(doc) - Gets all tables
			"getTables": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getTables() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("getTables() requires a DocxDocument argument")
				}

				body := doc.GetBody()
				if body == nil {
					return Array()
				}

				tblNodes := body.Find("//w:tbl")
				elements := tblNodes.Elements

				result := make([]objects.Object, len(elements))
				for i, elem := range elements {
					if node, ok := elem.(*objects.XMLNode); ok {
						// Count rows and cols
						rows := node.Find("w:tr")
						rowCount := len(rows.Elements)
						colCount := 0
						if rowCount > 0 {
							if firstRow, ok := rows.Elements[0].(*objects.XMLNode); ok {
								cells := firstRow.Find("w:tc")
								colCount = len(cells.Elements)
							}
						}

						result[i] = &objects.DocxTable{
							Document: doc,
							XmlNode:  node,
							Rows:     rowCount,
							Cols:     colCount,
						}
					}
				}

				return Array(result...)
			}),

			// setCellText(table, row, col, text) - Sets cell text
			"setCellText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 4 {
					return Error("setCellText() takes exactly 4 arguments")
				}
				tbl, ok := args[0].(*objects.DocxTable)
				if !ok {
					return Error("setCellText() requires a DocxTable as first argument")
				}
				rowIdx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("setCellText() requires an Int row as second argument")
				}
				colIdx, ok := args[2].(*objects.Int)
				if !ok {
					return Error("setCellText() requires an Int col as third argument")
				}
				text, ok := args[3].(*objects.String)
				if !ok {
					return Error("setCellText() requires a String text as fourth argument")
				}

				// Find the cell
				cell, err := getTableCell(tbl, int(rowIdx.Value), int(colIdx.Value))
				if err != nil {
					return Error(err.Error())
				}

				// Find or create paragraph in cell
				para := cell.FindFirst("w:p")
				if para == nil {
					para = objects.NewXMLNode("w:p")
					cell.AddChild(para)
				} else {
					// Clear existing content
					para.Clear()
				}

				// Add text run
				run := objects.NewXMLNode("w:r")
				textNode := objects.NewXMLNode("w:t")
				textNode.SetText(text.Value)
				textNode.SetAttr("xml:space", "preserve")
				run.AddChild(textNode)
				para.AddChild(run)

				return objects.NULL
			}),

			// getCell(table, row, col) - Gets cell object
			"getCell": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("getCell() takes exactly 3 arguments")
				}
				tbl, ok := args[0].(*objects.DocxTable)
				if !ok {
					return Error("getCell() requires a DocxTable as first argument")
				}
				rowIdx, ok := args[1].(*objects.Int)
				if !ok {
					return Error("getCell() requires an Int row as second argument")
				}
				colIdx, ok := args[2].(*objects.Int)
				if !ok {
					return Error("getCell() requires an Int col as third argument")
				}

				cell, err := getTableCell(tbl, int(rowIdx.Value), int(colIdx.Value))
				if err != nil {
					return Error(err.Error())
				}

				return &objects.DocxTableCell{
					XmlNode: cell,
					ColIdx:  int(colIdx.Value),
				}
			}),

			// setTableBorders(table, config) - Sets table borders
			"setTableBorders": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setTableBorders() takes exactly 2 arguments")
				}
				tbl, ok := args[0].(*objects.DocxTable)
				if !ok {
					return Error("setTableBorders() requires a DocxTable as first argument")
				}
				config, ok := args[1].(*objects.Map)
				if !ok {
					return Error("setTableBorders() requires a Map config as second argument")
				}

				// Find or create tblPr
				tblPr := tbl.XmlNode.FindFirst("w:tblPr")
				if tblPr == nil {
					tblPr = objects.NewXMLNode("w:tblPr")
					tbl.XmlNode.AddChild(tblPr)
				}

				// Create borders
				tblBorders := objects.NewXMLNode("w:tblBorders")

				borderNames := []string{"top", "left", "bottom", "right", "insideH", "insideV"}
				for _, name := range borderNames {
					border := objects.NewXMLNode("w:" + name)
					style := getMapString(config, name+"Style", "single")
					size := getMapInt(config, name+"Size", 4)
					color := getMapString(config, name+"Color", "auto")
					border.SetAttr("w:val", style)
					border.SetAttr("w:sz", intToStr(size))
					border.SetAttr("w:color", color)
					tblBorders.AddChild(border)
				}

				tblPr.AddChild(tblBorders)
				return objects.NULL
			}),

			// setCellShading(cell, config) - Sets cell background
			"setCellShading": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setCellShading() takes exactly 2 arguments")
				}
				cell, ok := args[0].(*objects.DocxTableCell)
				if !ok {
					return Error("setCellShading() requires a DocxTableCell as first argument")
				}
				config, ok := args[1].(*objects.Map)
				if !ok {
					return Error("setCellShading() requires a Map config as second argument")
				}

				// Find or create tcPr
				tcPr := cell.XmlNode.FindFirst("w:tcPr")
				if tcPr == nil {
					tcPr = objects.NewXMLNode("w:tcPr")
					cell.XmlNode.AddChild(tcPr)
				}

				// Create shading
				shd := objects.NewXMLNode("w:shd")
				fill := getMapString(config, "fill", "auto")
				pattern := getMapString(config, "pattern", "clear")
				shd.SetAttr("w:fill", fill)
				shd.SetAttr("w:val", pattern)
				tcPr.AddChild(shd)

				return objects.NULL
			}),

			// ============================================
			// Query Operations
			// ============================================

			// getText(doc) - Gets plain text content
			"getText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getText() takes exactly 1 argument")
				}

				var text strings.Builder

				switch doc := args[0].(type) {
				case *objects.DocxDocument:
					textNodes := doc.FindElements("//w:t")
					for _, elem := range textNodes.Elements {
						if node, ok := elem.(*objects.XMLNode); ok {
							text.WriteString(node.Text())
						}
					}

				case *objects.DocxParagraph:
					textNodes := doc.XmlNode.Find("w:t")
					for _, elem := range textNodes.Elements {
						if node, ok := elem.(*objects.XMLNode); ok {
							text.WriteString(node.Text())
						}
					}
					// Also try without prefix
					if text.Len() == 0 {
						textNodes = doc.XmlNode.Find("t")
						for _, elem := range textNodes.Elements {
							if node, ok := elem.(*objects.XMLNode); ok {
								text.WriteString(node.Text())
							}
						}
					}

				case *objects.DocxTableCell:
					textNodes := doc.XmlNode.Find("w:t")
					for _, elem := range textNodes.Elements {
						if node, ok := elem.(*objects.XMLNode); ok {
							text.WriteString(node.Text())
						}
					}
					// Also try without prefix
					if text.Len() == 0 {
						textNodes = doc.XmlNode.Find("t")
						for _, elem := range textNodes.Elements {
							if node, ok := elem.(*objects.XMLNode); ok {
								text.WriteString(node.Text())
							}
						}
					}

				default:
					return Error("getText() requires a DocxDocument, DocxParagraph, or DocxTableCell argument")
				}

				return String(text.String())
			}),

			// findAll(doc, text) - Finds all occurrences
			"findAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("findAll() takes exactly 2 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("findAll() requires a DocxDocument as first argument")
				}
				searchText, ok := args[1].(*objects.String)
				if !ok {
					return Error("findAll() requires a String as second argument")
				}

				body := doc.GetBody()
				if body == nil {
					return Array()
				}

				// Find all text nodes containing the search text
				textNodes := body.Find("//w:t")
				var positions []objects.Object
				pos := 0

				for _, elem := range textNodes.Elements {
					if node, ok := elem.(*objects.XMLNode); ok {
						text := node.Text()
						idx := strings.Index(text, searchText.Value)
						if idx >= 0 {
							positions = append(positions, Int(int64(pos+idx)))
						}
						pos += len(text)
					}
				}

				return Array(positions...)
			}),

			// replaceAll(doc, old, new) - Replaces all occurrences
			"replaceAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return Error("replaceAll() takes exactly 3 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("replaceAll() requires a DocxDocument as first argument")
				}
				oldText, ok := args[1].(*objects.String)
				if !ok {
					return Error("replaceAll() requires a String old as second argument")
				}
				newText, ok := args[2].(*objects.String)
				if !ok {
					return Error("replaceAll() requires a String new as third argument")
				}

				body := doc.GetBody()
				if body == nil {
					return Int(0)
				}

				// Find and replace in all text nodes
				textNodes := body.Find("//w:t")
				count := 0

				for _, elem := range textNodes.Elements {
					if node, ok := elem.(*objects.XMLNode); ok {
						text := node.Text()
						newTextStr := strings.ReplaceAll(text, oldText.Value, newText.Value)
						if text != newTextStr {
							node.SetText(newTextStr)
							count += strings.Count(text, oldText.Value)
						}
					}
				}

				if count > 0 {
					doc.SetModified(true)
				}

				return Int(int64(count))
			}),

			// ============================================
			// Document Property Operations
			// ============================================

			// getTitle(doc) - Gets document title
			"getTitle": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getTitle() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("getTitle() requires a DocxDocument argument")
				}
				props := doc.GetProperties()
				if props == nil {
					return String("")
				}
				return String(props.Title)
			}),

			// setTitle(doc, title) - Sets document title
			"setTitle": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setTitle() takes exactly 2 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("setTitle() requires a DocxDocument as first argument")
				}
				title, ok := args[1].(*objects.String)
				if !ok {
					return Error("setTitle() requires a String as second argument")
				}
				props := doc.GetProperties()
				if props != nil {
					props.Title = title.Value
				}
				return objects.NULL
			}),

			// getAuthor(doc) - Gets document author
			"getAuthor": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getAuthor() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("getAuthor() requires a DocxDocument argument")
				}
				props := doc.GetProperties()
				if props == nil {
					return String("")
				}
				return String(props.Author)
			}),

			// setAuthor(doc, author) - Sets document author
			"setAuthor": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("setAuthor() takes exactly 2 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("setAuthor() requires a DocxDocument as first argument")
				}
				author, ok := args[1].(*objects.String)
				if !ok {
					return Error("setAuthor() requires a String as second argument")
				}
				props := doc.GetProperties()
				if props != nil {
					props.Author = author.Value
				}
				return objects.NULL
			}),

			// getProperties(doc) - Gets document properties as map
			"getProperties": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getProperties() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("getProperties() requires a DocxDocument argument")
				}
				props := doc.GetProperties()
				if props == nil {
					pairs := make(map[objects.HashKey]objects.MapPair)
					return &objects.Map{Pairs: pairs}
				}

				pairs := make(map[objects.HashKey]objects.MapPair)
				pairs[objects.NewString("title").HashKey()] = objects.MapPair{Key: objects.NewString("title"), Value: String(props.Title)}
				pairs[objects.NewString("subject").HashKey()] = objects.MapPair{Key: objects.NewString("subject"), Value: String(props.Subject)}
				pairs[objects.NewString("author").HashKey()] = objects.MapPair{Key: objects.NewString("author"), Value: String(props.Author)}
				pairs[objects.NewString("keywords").HashKey()] = objects.MapPair{Key: objects.NewString("keywords"), Value: String(props.Keywords)}
				pairs[objects.NewString("description").HashKey()] = objects.MapPair{Key: objects.NewString("description"), Value: String(props.Description)}
				pairs[objects.NewString("created").HashKey()] = objects.MapPair{Key: objects.NewString("created"), Value: String(props.Created)}
				pairs[objects.NewString("modified").HashKey()] = objects.MapPair{Key: objects.NewString("modified"), Value: String(props.Modified)}

				return &objects.Map{Pairs: pairs}
			}),

			// ============================================
			// Page Operations
			// ============================================

			// addPageBreak(doc) - Adds a page break
			"addPageBreak": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("addPageBreak() takes exactly 1 argument")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("addPageBreak() requires a DocxDocument argument")
				}

				body := doc.GetBody()
				if body == nil {
					return Error("document has no body element")
				}

				// Create paragraph with page break
				para := objects.NewXMLNode("w:p")
				run := objects.NewXMLNode("w:r")
				br := objects.NewXMLNode("w:br")
				br.SetAttr("w:type", "page")
				run.AddChild(br)
				para.AddChild(run)
				body.AddChild(para)

				doc.SetModified(true)
				return objects.NULL
			}),

			// ============================================
			// Low-level XML Operations
			// ============================================

			// findXML(doc, path) - Find nodes by path expression
			"findXML": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return Error("findXML() takes exactly 2 arguments")
				}
				doc, ok := args[0].(*objects.DocxDocument)
				if !ok {
					return Error("findXML() requires a DocxDocument as first argument")
				}
				path, ok := args[1].(*objects.String)
				if !ok {
					return Error("findXML() requires a String path as second argument")
				}

				xmlDoc := doc.GetXMLDoc()
				if xmlDoc == nil {
					return Array()
				}

				return xmlDoc.Find(path.Value)
			}),

			// getXMLNode(element) - Gets underlying XML node
			"getXMLNode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("getXMLNode() takes exactly 1 argument")
				}

				switch elem := args[0].(type) {
				case *objects.DocxParagraph:
					return elem.XmlNode
				case *objects.DocxRun:
					return elem.XmlNode
				case *objects.DocxTable:
					return elem.XmlNode
				case *objects.DocxTableCell:
					return elem.XmlNode
				default:
					return Error("getXMLNode() requires a Docx element argument")
				}
			}),

			// ============================================
			// Type Check Functions
			// ============================================

			"isDocxDocument": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDocxDocument() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.DocxDocument)
				return Bool(ok)
			}),

			"isDocxParagraph": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDocxParagraph() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.DocxParagraph)
				return Bool(ok)
			}),

			"isDocxRun": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDocxRun() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.DocxRun)
				return Bool(ok)
			}),

			"isDocxTable": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDocxTable() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.DocxTable)
				return Bool(ok)
			}),

			"isDocxTableCell": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isDocxTableCell() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.DocxTableCell)
				return Bool(ok)
			}),

			// ============================================
			// Constants
			// ============================================

			// Unit conversion constants
			"TwipsPerInch":  Int(1440),
			"TwipsPerPoint": Int(20),
			"EMUsPerInch":   Int(914400),
			"EMUsPerPoint":  Int(12700),

			// Standard page sizes (in twips)
			"PageSizeA4Width":      Int(11906),
			"PageSizeA4Height":     Int(16838),
			"PageSizeLetterWidth":  Int(12240),
			"PageSizeLetterHeight": Int(15840),

			// Standard margins (in twips)
			"MarginNormal":   Int(1440),
			"MarginNarrow":   Int(720),
			"MarginModerate": Int(1080),
			"MarginWide":     Int(2160),
		},
	})
}

// Helper functions

// getMapString gets a string value from a map with a default.
func getMapString(m *objects.Map, key, defVal string) string {
	if m == nil {
		return defVal
	}
	for _, pair := range m.Pairs {
		if k, ok := pair.Key.(*objects.String); ok && k.Value == key {
			if v, ok := pair.Value.(*objects.String); ok {
				return v.Value
			}
		}
	}
	return defVal
}

// getMapInt gets an int value from a map with a default.
func getMapInt(m *objects.Map, key string, defVal int) int {
	if m == nil {
		return defVal
	}
	for _, pair := range m.Pairs {
		if k, ok := pair.Key.(*objects.String); ok && k.Value == key {
			if v, ok := pair.Value.(*objects.Int); ok {
				return int(v.Value)
			}
		}
	}
	return defVal
}

// getMapBool gets a bool value from a map with a default.
func getMapBool(m *objects.Map, key string, defVal bool) bool {
	if m == nil {
		return defVal
	}
	for _, pair := range m.Pairs {
		if k, ok := pair.Key.(*objects.String); ok && k.Value == key {
			if v, ok := pair.Value.(*objects.Bool); ok {
				return v.Value
			}
		}
	}
	return defVal
}

// intToStr converts int to string.
func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}

// setParagraphProperty sets a paragraph property.
func setParagraphProperty(para *objects.DocxParagraph, propName, value string) error {
	pPr := para.XmlNode.FindFirst("w:pPr")
	if pPr == nil {
		pPr = objects.NewXMLNode("w:pPr")
		para.XmlNode.AddChild(pPr)
	}

	// Find or create the property node
	prop := pPr.FindFirst(propName)
	if prop == nil {
		prop = objects.NewXMLNode(propName)
		pPr.AddChild(prop)
	}

	prop.SetAttr("w:val", value)
	return nil
}

// getOrCreatePPr gets or creates paragraph properties.
func getOrCreatePPr(para *objects.DocxParagraph) *objects.XMLNode {
	pPr := para.XmlNode.FindFirst("w:pPr")
	if pPr == nil {
		pPr = objects.NewXMLNode("w:pPr")
		para.XmlNode.AddChild(pPr)
	}
	return pPr
}

// getOrCreateRPr gets or creates run properties.
func getOrCreateRPr(run *objects.DocxRun) *objects.XMLNode {
	rPr := run.XmlNode.FindFirst("w:rPr")
	if rPr == nil {
		rPr = objects.NewXMLNode("w:rPr")
		run.XmlNode.AddChild(rPr)
	}
	return rPr
}

// setRunProperty sets a run property (boolean).
func setRunProperty(run *objects.DocxRun, propName string, value bool) {
	rPr := getOrCreateRPr(run)

	// Find existing property
	prop := rPr.FindFirst(propName)
	if prop != nil {
		if !value {
			// Remove property
			// Note: XMLNode doesn't have a remove method, so we just don't add it
			return
		}
		return
	}

	if value {
		prop = objects.NewXMLNode(propName)
		rPr.AddChild(prop)
	}
}

// getTableCell gets a table cell by row and column index.
func getTableCell(tbl *objects.DocxTable, rowIdx, colIdx int) (*objects.XMLNode, error) {
	rows := tbl.XmlNode.Find("w:tr")
	if rowIdx < 0 || rowIdx >= len(rows.Elements) {
		return nil, fmt.Errorf("row index %d out of range", rowIdx)
	}

	row, ok := rows.Elements[rowIdx].(*objects.XMLNode)
	if !ok {
		return nil, fmt.Errorf("invalid row node")
	}

	cells := row.Find("w:tc")
	if colIdx < 0 || colIdx >= len(cells.Elements) {
		return nil, fmt.Errorf("column index %d out of range", colIdx)
	}

	cell, ok := cells.Elements[colIdx].(*objects.XMLNode)
	if !ok {
		return nil, fmt.Errorf("invalid cell node")
	}

	return cell, nil
}