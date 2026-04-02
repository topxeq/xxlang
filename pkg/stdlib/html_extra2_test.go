// pkg/stdlib/html_extra2_test.go
// Additional tests for html module to increase coverage.
package stdlib

import (
	"os"
	"strings"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

// htmlCall calls an html module builtin.
func htmlCall(name string, args ...objects.Object) objects.Object {
	mod := Get("html")
	if mod == nil {
		return &objects.Error{Message: "html module not found"}
	}
	fn, ok := mod.Exports[name].(*objects.Builtin)
	if !ok {
		return &objects.Error{Message: "function not found: " + name}
	}
	return fn.Fn(args...)
}

// TestHTMLExtra2_Basic tests basic html functions.
func TestHTMLExtra2_Basic(t *testing.T) {
	// Test newDocument
	doc := htmlCall("newDocument")
	if _, ok := doc.(*objects.Error); ok {
		t.Fatalf("newDocument error: %s", doc.Inspect())
	}
	if _, ok := doc.(*objects.HTMLDocument); !ok {
		t.Fatalf("newDocument did not return HTMLDocument, got %T", doc)
	}

	// Test newDocumentWithTitle
	docWithTitle := htmlCall("newDocumentWithTitle", String("Test Title"))
	if _, ok := docWithTitle.(*objects.Error); ok {
		t.Fatalf("newDocumentWithTitle error: %s", docWithTitle.Inspect())
	}
	if _, ok := docWithTitle.(*objects.HTMLDocument); !ok {
		t.Fatalf("newDocumentWithTitle did not return HTMLDocument, got %T", docWithTitle)
	}

	// Test newElement
	elem := htmlCall("newElement", String("div"))
	if _, ok := elem.(*objects.Error); ok {
		t.Fatalf("newElement error: %s", elem.Inspect())
	}
	if _, ok := elem.(*objects.HTMLElement); !ok {
		t.Fatalf("newElement did not return HTMLElement, got %T", elem)
	}

	// Test newTextNode
	textNode := htmlCall("newTextNode", String("Hello"))
	if _, ok := textNode.(*objects.Error); ok {
		t.Fatalf("newTextNode error: %s", textNode.Inspect())
	}
	// Text node is an HTMLElement as well? Actually it's a different type but can be considered element.

	// Test newComment
	comment := htmlCall("newComment", String("comment"))
	if _, ok := comment.(*objects.Error); ok {
		t.Fatalf("newComment error: %s", comment.Inspect())
	}
}

// TestHTMLExtra2_Escape tests html.escape, escapeAttr, unescape, stripTags, sanitize.
func TestHTMLExtra2_Escape(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		input    string
		wantCont string
	}{
		{"escape", "escape", "<script>", "&lt;"},
		{"escape", "escape", "a & b", "&amp;"},
		{"escapeAttr", "escapeAttr", `"quotes"`, "&quot;"},
		{"unescape", "unescape", "&lt;", "<"},
		{"stripTags", "stripTags", "<p>Hello</p>", "Hello"},
		// sanitize test moved to separate test due to empty expectation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := htmlCall(tt.funcName, String(tt.input))
			if _, ok := result.(*objects.Error); ok {
				t.Fatalf("%s error: %s", tt.funcName, result.Inspect())
			}
			str, ok := result.(*objects.String)
			if !ok {
				t.Fatalf("Expected String, got %T", result)
			}
			if !strings.Contains(str.Value, tt.wantCont) {
				t.Errorf("%s(%q) = %q; expected to contain %q", tt.funcName, tt.input, str.Value, tt.wantCont)
			}
		})
	}
}

// TestHTMLExtra2_Sanitize tests html.sanitize specifically.
func TestHTMLExtra2_Sanitize(t *testing.T) {
	// Sanitize should remove script tags and their content
	result := htmlCall("sanitize", String("<script>alert('xss')</script>"))
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("sanitize error: %s", result.Inspect())
	}
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	// After sanitization, the script tag and its content should be removed, leaving empty or only safe text
	if strings.Contains(str.Value, "<script>") {
		t.Errorf("sanitize did not remove script tag: %q", str.Value)
	}
	if strings.Contains(str.Value, "alert") {
		t.Errorf("sanitize did not remove script content: %q", str.Value)
	}
}

// TestHTMLExtra2_Parse tests html.parse and parseFragment.
func TestHTMLExtra2_Parse(t *testing.T) {
	// Test parse with valid HTML
	htmlStr := "<!DOCTYPE html><html><head><title>Test</title></head><body><h1>Hello</h1></body></html>"
	doc := htmlCall("parse", String(htmlStr))
	if _, ok := doc.(*objects.Error); ok {
		t.Fatalf("parse error: %s", doc.Inspect())
	}
	if _, ok := doc.(*objects.HTMLDocument); !ok {
		t.Fatalf("parse did not return HTMLDocument, got %T", doc)
	}

	// Test parseFragment
	frag := htmlCall("parseFragment", String("<p>Para</p><div>Div</div>"))
	if _, ok := frag.(*objects.Error); ok {
		t.Fatalf("parseFragment error: %s", frag.Inspect())
	}
	arr, ok := frag.(*objects.Array)
	if !ok {
		t.Fatalf("parseFragment did not return Array, got %T", frag)
	}
	if len(arr.Elements) == 0 {
		t.Error("parseFragment returned empty array")
	}

	// Test parse with invalid HTML (should error)
	badDoc := htmlCall("parse", String("<<<invalid>>>"))
	if _, ok := badDoc.(*objects.Error); !ok {
		t.Errorf("parse with invalid HTML should error, got %v", badDoc)
	}

	// Test parseFragment with invalid (may succeed or error)
	badFrag := htmlCall("parseFragment", String("<<<invalid>>>"))
	if _, ok := badFrag.(*objects.Error); ok {
		t.Logf("parseFragment with invalid HTML error: %s", badFrag.Inspect())
	}
}

// TestHTMLExtra2_FileOperations tests parseFile.
func TestHTMLExtra2_FileOperations(t *testing.T) {
	// Create a temporary HTML file
	tmpDir := t.TempDir()
	filePath := tmpDir + "/test.html"
	content := "<html><body><h1>Test</h1></body></html>"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Test parseFile
	doc := htmlCall("parseFile", String(filePath))
	if _, ok := doc.(*objects.Error); ok {
		t.Fatalf("parseFile error: %s", doc.Inspect())
	}
	if _, ok := doc.(*objects.HTMLDocument); !ok {
		t.Fatalf("parseFile did not return HTMLDocument, got %T", doc)
	}

	// Test parseFile with non-existent file
	badDoc := htmlCall("parseFile", String("/nonexistent/file.html"))
	if _, ok := badDoc.(*objects.Error); !ok {
		t.Errorf("parseFile with non-existent file should error, got %v", badDoc)
	}
}

// TestHTMLExtra2_IsChecks tests isHTMLDocument and isHTMLElement.
func TestHTMLExtra2_IsChecks(t *testing.T) {
	// Create a document and element
	doc := htmlCall("newDocument")
	elem := htmlCall("newElement", String("span"))
	nonHTML := String("not html")

	// isHTMLDocument
	res := htmlCall("isHTMLDocument", doc)
	if boolObj, ok := res.(*objects.Bool); ok {
		if !boolObj.Value {
			t.Errorf("isHTMLDocument(doc) = false; want true")
		}
	} else {
		t.Errorf("isHTMLDocument(doc) returned non-Bool: %T", res)
	}

	res = htmlCall("isHTMLDocument", elem)
	if boolObj, ok := res.(*objects.Bool); ok {
		if boolObj.Value {
			t.Errorf("isHTMLDocument(elem) = true; want false")
		}
	} else {
		t.Errorf("isHTMLDocument(elem) returned non-Bool: %T", res)
	}

	res = htmlCall("isHTMLDocument", nonHTML)
	if boolObj, ok := res.(*objects.Bool); ok {
		if boolObj.Value {
			t.Errorf("isHTMLDocument(nonHTML) = true; want false")
		}
	} else {
		t.Errorf("isHTMLDocument(nonHTML) returned non-Bool: %T", res)
	}

	// isHTMLElement
	res = htmlCall("isHTMLElement", elem)
	if boolObj, ok := res.(*objects.Bool); ok {
		if !boolObj.Value {
			t.Errorf("isHTMLElement(elem) = false; want true")
		}
	} else {
		t.Errorf("isHTMLElement(elem) returned non-Bool: %T", res)
	}

	res = htmlCall("isHTMLElement", doc)
	if boolObj, ok := res.(*objects.Bool); ok {
		if boolObj.Value {
			t.Errorf("isHTMLElement(doc) = true; want false")
		}
	} else {
		t.Errorf("isHTMLElement(doc) returned non-Bool: %T", res)
	}

	res = htmlCall("isHTMLElement", nonHTML)
	if boolObj, ok := res.(*objects.Bool); ok {
		if boolObj.Value {
			t.Errorf("isHTMLElement(nonHTML) = true; want false")
		}
	} else {
		t.Errorf("isHTMLElement(nonHTML) returned non-Bool: %T", res)
	}
}

// TestHTMLExtra2_Encode tests html.encode.
func TestHTMLExtra2_Encode(t *testing.T) {
	// Encode a simple map
	mapObj := &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
	mapObj.Pairs[String("name").HashKey()] = objects.MapPair{Key: String("name"), Value: String("Test")}
	mapObj.Pairs[String("value").HashKey()] = objects.MapPair{Key: String("value"), Value: Int(123)}

	result := htmlCall("encode", mapObj)
	if _, ok := result.(*objects.Error); ok {
		t.Fatalf("encode error: %s", result.Inspect())
	}
	str, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if !strings.Contains(str.Value, "<div") {
		t.Error("encode output should contain a div wrapper")
	}

	// With custom root name
	result2 := htmlCall("encode", mapObj, String("section"))
	if _, ok := result2.(*objects.Error); ok {
		t.Fatalf("encode with root error: %s", result2.Inspect())
	}
	str2, ok := result2.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result2)
	}
	if !strings.Contains(str2.Value, "<section") {
		t.Error("encode with root should contain section wrapper")
	}

	// Encode with wrong second arg type (should ignore and use default)
	badRoot := htmlCall("encode", mapObj, Int(456))
	if _, ok := badRoot.(*objects.Error); ok {
		t.Logf("encode with int root gave error: %s", badRoot.Inspect())
	} else {
		strBad, ok := badRoot.(*objects.String)
		if ok && !strings.Contains(strBad.Value, "<div") {
			t.Error("encode with int root should still produce div wrapper")
		}
	}
}

// TestHTMLExtra2_ErrorHandling tests error handling for all html functions.
func TestHTMLExtra2_ErrorHandling(t *testing.T) {
	// parse
	result := htmlCall("parse")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parse with no args should error, got %v", result)
	}
	result = htmlCall("parse", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parse with int should error, got %v", result)
	}

	// parseFile
	result = htmlCall("parseFile")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseFile with no args should error, got %v", result)
	}
	result = htmlCall("parseFile", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseFile with int should error, got %v", result)
	}

	// parseFragment
	result = htmlCall("parseFragment")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseFragment with no args should error, got %v", result)
	}
	result = htmlCall("parseFragment", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("parseFragment with int should error, got %v", result)
	}

	// newDocumentWithTitle
	result = htmlCall("newDocumentWithTitle")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newDocumentWithTitle with no args should error, got %v", result)
	}
	result = htmlCall("newDocumentWithTitle", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newDocumentWithTitle with int should error, got %v", result)
	}

	// newElement
	result = htmlCall("newElement")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newElement with no args should error, got %v", result)
	}
	result = htmlCall("newElement", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newElement with int should error, got %v", result)
	}

	// newTextNode
	result = htmlCall("newTextNode")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newTextNode with no args should error, got %v", result)
	}
	result = htmlCall("newTextNode", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newTextNode with int should error, got %v", result)
	}

	// newComment
	result = htmlCall("newComment")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newComment with no args should error, got %v", result)
	}
	result = htmlCall("newComment", objects.NewInt(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("newComment with int should error, got %v", result)
	}

	// escape
	result = htmlCall("escape")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("escape with no args should error, got %v", result)
	}
	result = htmlCall("escape", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("escape with int should error, got %v", result)
	}

	// escapeAttr
	result = htmlCall("escapeAttr")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("escapeAttr with no args should error, got %v", result)
	}
	result = htmlCall("escapeAttr", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("escapeAttr with int should error, got %v", result)
	}

	// unescape
	result = htmlCall("unescape")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("unescape with no args should error, got %v", result)
	}
	result = htmlCall("unescape", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("unescape with int should error, got %v", result)
	}

	// stripTags
	result = htmlCall("stripTags")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("stripTags with no args should error, got %v", result)
	}
	result = htmlCall("stripTags", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("stripTags with int should error, got %v", result)
	}

	// sanitize
	result = htmlCall("sanitize")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("sanitize with no args should error, got %v", result)
	}
	result = htmlCall("sanitize", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("sanitize with int should error, got %v", result)
	}

	// isHTMLDocument and isHTMLElement already tested with types; also test no args
	result = htmlCall("isHTMLDocument")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("isHTMLDocument with no args should error, got %v", result)
	}
	result = htmlCall("isHTMLElement")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("isHTMLElement with no args should error, got %v", result)
	}

	// encode
	result = htmlCall("encode")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("encode with no args should error, got %v", result)
	}

	// createElement alias
	result = htmlCall("createElement")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("createElement with no args should error, got %v", result)
	}
	result = htmlCall("createElement", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("createElement with int should error, got %v", result)
	}

	// createTextNode alias
	result = htmlCall("createTextNode")
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("createTextNode with no args should error, got %v", result)
	}
	result = htmlCall("createTextNode", Int(123))
	if _, ok := result.(*objects.Error); !ok {
		t.Errorf("createTextNode with int should error, got %v", result)
	}
}
