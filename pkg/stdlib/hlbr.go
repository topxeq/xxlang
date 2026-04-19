// pkg/stdlib/hlbr.go
// HLBR module for Xxlang - lightweight headless browser.
// This module provides a zero-dependency headless browser that does not
// require Chrome/Chromium. It uses a built-in HTTP client, HTML parser,
// DOM tree, CSS selector engine, and JavaScript interpreter.
package stdlib

import (
	"fmt"

	"github.com/topxeq/xxlang/pkg/hlbr"
	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
	"github.com/topxeq/xxlang/pkg/hlbr/renderer"

	"github.com/topxeq/xxlang/pkg/objects"
)

// hlbrGetStringFromMap extracts a string value from an Xxlang Map by key.
func hlbrGetStringFromMap(m *objects.Map, key string) string {
	keyObj := objects.NewString(key)
	if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
		if s, ok := pair.Value.(*objects.String); ok {
			return s.Value
		}
	}
	return ""
}

func init() {
	Register(&Module{
		Name: "hlbr",
		Exports: map[string]objects.Object{

			// open creates a new HlbrBrowser instance.
			// Options map keys: userAgent (string), proxy (string), timeout (int, seconds).
			"open": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewHlbrBrowser(args...)
			}),

			// parseHTML parses an HTML string and returns an HlbrNode (document root).
			"parseHTML": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("parseHTML() requires 1 argument (html string)")
				}
				htmlStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("parseHTML() argument must be STRING")
				}
				doc := htmlparser.Parse(htmlStr.Value)
				if doc == nil || doc.Root == nil {
					return objects.NULL
				}
				return objects.NewHlbrNode(doc.Root)
			}),

			// screenshotText renders an HTML string to text at the given width.
			"screenshotText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("screenshotText() requires at least 1 argument (html string)")
				}
				htmlStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("screenshotText() first argument must be STRING")
				}
				width := 80
				if len(args) >= 2 {
					if w, ok := args[1].(*objects.Int); ok {
						width = int(w.Value)
					} else if w, ok := args[1].(*objects.Float); ok {
						width = int(w.Value)
					}
				}
				doc := htmlparser.Parse(htmlStr.Value)
				if doc == nil {
					return Error("screenshotText() failed to parse HTML")
				}
				return String(renderer.ScreenshotText(doc, width))
			}),

			// querySelector performs a CSS selector query on an HlbrNode.
			"querySelector": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("querySelector() requires 2 arguments (node, selector)")
				}
				node, ok := args[0].(*objects.HlbrNode)
				if !ok {
					return Error("querySelector() first argument must be HLBR_NODE")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("querySelector() second argument must be STRING")
				}
				result := dom.QuerySelector(node.GetNode(), sel.Value)
				if result == nil {
					return objects.NULL
				}
				return objects.NewHlbrNode(result)
			}),

			// querySelectorAll performs a CSS selector query on an HlbrNode, returning all matches.
			"querySelectorAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("querySelectorAll() requires 2 arguments (node, selector)")
				}
				node, ok := args[0].(*objects.HlbrNode)
				if !ok {
					return Error("querySelectorAll() first argument must be HLBR_NODE")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("querySelectorAll() second argument must be STRING")
				}
				nodes := dom.QuerySelectorAll(node.GetNode(), sel.Value)
				elems := make([]objects.Object, len(nodes))
				for i, n := range nodes {
					elems[i] = objects.NewHlbrNode(n)
				}
				return Array(elems...)
			}),

			// launch is an alias for open (creates a new HlbrBrowser).
			"launch": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return objects.NewHlbrBrowser(args...)
			}),

			// quickGet is a convenience function that creates a browser,
			// navigates to a URL, and returns the page text.
			"quickGet": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("quickGet() requires 1 argument (url)")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("quickGet() argument must be STRING")
				}

				var opts *hlbr.Options
				if len(args) >= 2 {
					if m, ok := args[1].(*objects.Map); ok {
						opts = &hlbr.Options{}
						if v := hlbrGetStringFromMap(m, "userAgent"); v != "" {
							opts.UserAgent = v
						}
						if v := hlbrGetStringFromMap(m, "proxy"); v != "" {
							opts.Proxy = v
						}
					}
				}

				b, err := hlbr.New(opts)
				if err != nil {
					return Error("quickGet() failed to create browser: " + err.Error())
				}
				if err := b.Navigate(url.Value); err != nil {
					return Error("quickGet() navigate failed: " + err.Error())
				}
				return String(b.GetText())
			}),

			// quickGetHTML navigates to a URL and returns the page HTML.
			"quickGetHTML": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("quickGetHTML() requires 1 argument (url)")
				}
				url, ok := args[0].(*objects.String)
				if !ok {
					return Error("quickGetHTML() argument must be STRING")
				}

				b, err := hlbr.New(nil)
				if err != nil {
					return Error("quickGetHTML() failed to create browser: " + err.Error())
				}
				if err := b.Navigate(url.Value); err != nil {
					return Error("quickGetHTML() navigate failed: " + err.Error())
				}
				return String(b.GetHTML())
			}),

			// ========== Storage Functions ==========

			// getLocalStorage returns the localStorage data from a browser.
			"getLocalStorage": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getLocalStorage() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getLocalStorage() argument must be HLBR_BROWSER")
				}
				storage := br.GetBrowser().GetLocalStorage()
				if storage == nil {
					return &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
				}
				pairs := make(map[objects.HashKey]objects.MapPair)
				for k, v := range storage {
					keyObj := objects.NewString(k)
					pairs[keyObj.HashKey()] = objects.MapPair{Key: keyObj, Value: objects.NewString(v)}
				}
				return &objects.Map{Pairs: pairs}
			}),

			// getSessionStorage returns the sessionStorage data from a browser.
			"getSessionStorage": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getSessionStorage() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getSessionStorage() argument must be HLBR_BROWSER")
				}
				storage := br.GetBrowser().GetSessionStorage()
				if storage == nil {
					return &objects.Map{Pairs: make(map[objects.HashKey]objects.MapPair)}
				}
				pairs := make(map[objects.HashKey]objects.MapPair)
				for k, v := range storage {
					keyObj := objects.NewString(k)
					pairs[keyObj.HashKey()] = objects.MapPair{Key: keyObj, Value: objects.NewString(v)}
				}
				return &objects.Map{Pairs: pairs}
			}),

			// setLocalStorage sets a localStorage item.
			"setLocalStorage": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("setLocalStorage() requires 3 arguments (browser, key, value)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setLocalStorage() first argument must be HLBR_BROWSER")
				}
				key, ok := args[1].(*objects.String)
				if !ok {
					return Error("setLocalStorage() key must be STRING")
				}
				val, ok := args[2].(*objects.String)
				if !ok {
					return Error("setLocalStorage() value must be STRING")
				}
				br.GetBrowser().SetLocalStorageItem(key.Value, val.Value)
				return objects.NULL
			}),

			// setSessionStorage sets a sessionStorage item.
			"setSessionStorage": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("setSessionStorage() requires 3 arguments (browser, key, value)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setSessionStorage() first argument must be HLBR_BROWSER")
				}
				key, ok := args[1].(*objects.String)
				if !ok {
					return Error("setSessionStorage() key must be STRING")
				}
				val, ok := args[2].(*objects.String)
				if !ok {
					return Error("setSessionStorage() value must be STRING")
				}
				br.GetBrowser().SetSessionStorageItem(key.Value, val.Value)
				return objects.NULL
			}),

			// getConsoleOutput returns the console.log output from a browser.
			"getConsoleOutput": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getConsoleOutput() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getConsoleOutput() argument must be HLBR_BROWSER")
				}
				output := br.GetBrowser().GetConsoleOutput()
				if output == nil {
					return Array()
				}
				elems := make([]objects.Object, len(output))
				for i, s := range output {
					elems[i] = objects.NewString(s)
				}
				return Array(elems...)
			}),

			// ========== Form Interaction Functions ==========

			// setValue sets a form value via DOM attribute.
			"setValue": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("setValue() requires 3 arguments (browser, selector, value)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setValue() first argument must be HLBR_BROWSER")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("setValue() selector must be STRING")
				}
				val, ok := args[2].(*objects.String)
				if !ok {
					return Error("setValue() value must be STRING")
				}
				node := br.GetBrowser().QuerySelector(sel.Value)
				if node == nil {
					return Error("setValue: element not found for selector '" + sel.Value + "'")
				}
				node.SetAttribute("value", val.Value)
				return br
			}),

			// setValueByJS sets a form value via JavaScript.
			"setValueByJS": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("setValueByJS() requires 3 arguments (browser, selector, value)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setValueByJS() first argument must be HLBR_BROWSER")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("setValueByJS() selector must be STRING")
				}
				val, ok := args[2].(*objects.String)
				if !ok {
					return Error("setValueByJS() value must be STRING")
				}
				jsCode := fmt.Sprintf("var el = document.querySelector('%s'); if (el) { el.value = '%s'; }", sel.Value, val.Value)
				_, err := br.GetBrowser().Evaluate(jsCode)
				if err != nil {
					return Error("setValueByJS failed: " + err.Error())
				}
				return br
			}),

			// clickByJS clicks an element via JavaScript.
			"clickByJS": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("clickByJS() requires 2 arguments (browser, selector)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("clickByJS() first argument must be HLBR_BROWSER")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("clickByJS() selector must be STRING")
				}
				jsCode := fmt.Sprintf("var el = document.querySelector('%s'); if (el) { el.click(); }", sel.Value)
				_, err := br.GetBrowser().Evaluate(jsCode)
				if err != nil {
					return Error("clickByJS failed: " + err.Error())
				}
				return br
			}),

			// submitForm submits a form via JavaScript.
			"submitForm": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("submitForm() requires 2 arguments (browser, selector)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("submitForm() first argument must be HLBR_BROWSER")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("submitForm() selector must be STRING")
				}
				jsCode := fmt.Sprintf("var form = document.querySelector('%s'); if (form && form.submit) { form.submit(); }", sel.Value)
				_, err := br.GetBrowser().Evaluate(jsCode)
				if err != nil {
					return Error("submitForm failed: " + err.Error())
				}
				return br
			}),
		},
	})
}