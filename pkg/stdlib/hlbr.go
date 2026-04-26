// pkg/stdlib/hlbr.go
// HLBR module for Xxlang - lightweight headless browser.
// This module provides a zero-dependency headless browser that does not
// require Chrome/Chromium. It uses a built-in HTTP client, HTML parser,
// DOM tree, CSS selector engine, and JavaScript interpreter.
package stdlib

import (
	"fmt"
	"net/http"
	"time"

	"github.com/topxeq/xxlang/pkg/hlbr"
	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
	"github.com/topxeq/xxlang/pkg/hlbr/renderer"

	"github.com/topxeq/xxlang/pkg/objects"
)

// hlbrGoValueToObject converts a Go value from hlbr.Evaluate to an Xxlang Object.
func hlbrGoValueToObject(v interface{}) objects.Object {
	if v == nil {
		return objects.NULL
	}
	switch val := v.(type) {
	case bool:
		return Bool(val)
	case float64:
		if val == float64(int64(val)) {
			return Int(int64(val))
		}
		return Float(val)
	case string:
		return String(val)
	case []interface{}:
		elems := make([]objects.Object, len(val))
		for i, item := range val {
			elems[i] = hlbrGoValueToObject(item)
		}
		return Array(elems...)
	case map[string]interface{}:
		pairs := make(map[objects.HashKey]objects.MapPair)
		for k, v := range val {
			key := String(k)
			pairs[key.HashKey()] = objects.MapPair{Key: key, Value: hlbrGoValueToObject(v)}
		}
		return &objects.Map{Pairs: pairs}
	default:
		return String(fmt.Sprintf("%v", v))
	}
}

// hlbrCookiesToXxArray converts http.Cookie slice to Xxlang Array.
func hlbrCookiesToXxArray(cookies []*http.Cookie) objects.Object {
	elems := make([]objects.Object, len(cookies))
	for i, c := range cookies {
		pairs := make(map[objects.HashKey]objects.MapPair)
		addToCookieMap(pairs, "name", c.Name)
		addToCookieMap(pairs, "value", c.Value)
		addToCookieMap(pairs, "domain", c.Domain)
		addToCookieMap(pairs, "path", c.Path)
		elems[i] = &objects.Map{Pairs: pairs}
	}
	return Array(elems...)
}

// addToCookieMap adds a key-value pair to a cookie map.
func addToCookieMap(pairs map[objects.HashKey]objects.MapPair, key, value string) {
	keyObj := String(key)
	pairs[keyObj.HashKey()] = objects.MapPair{Key: keyObj, Value: String(value)}
}

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

			// navigate navigates browser to a URL.
			// Usage: navigate(browser, url)
			"navigate": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("navigate() requires 2 arguments (browser, url)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("navigate() first argument must be HLBR_BROWSER")
				}
				url, ok := args[1].(*objects.String)
				if !ok {
					return Error("navigate() second argument must be STRING")
				}
				if err := br.GetBrowser().Navigate(url.Value); err != nil {
					return Error("navigate() failed: " + err.Error())
				}
				return br
			}),

			// evaluate executes JavaScript code in browser.
			// Usage: evaluate(browser, jsCode) -> result
			"evaluate": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("evaluate() requires 2 arguments (browser, jsCode)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("evaluate() first argument must be HLBR_BROWSER")
				}
				jsCode, ok := args[1].(*objects.String)
				if !ok {
					return Error("evaluate() second argument must be STRING")
				}
				result, err := br.GetBrowser().Evaluate(jsCode.Value)
				if err != nil {
					return Error("evaluate() failed: " + err.Error())
				}
				return hlbrGoValueToObject(result)
			}),

			// getHTML returns the current page HTML.
			// Usage: getHTML(browser) -> string
			"getHTML": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getHTML() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getHTML() argument must be HLBR_BROWSER")
				}
				return String(br.GetBrowser().GetHTML())
			}),

			// getText returns the current page text content.
			// Usage: getText(browser) -> string
			"getText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getText() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getText() argument must be HLBR_BROWSER")
				}
				return String(br.GetBrowser().GetText())
			}),

			// getURL returns the current page URL.
			// Usage: getURL(browser) -> string
			"getURL": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getURL() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getURL() argument must be HLBR_BROWSER")
				}
				return String(br.GetBrowser().GetURL())
			}),

			// getCookies returns browser cookies.
			// Usage: getCookies(browser) -> array
			"getCookies": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getCookies() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getCookies() argument must be HLBR_BROWSER")
				}
				return hlbrCookiesToXxArray(br.GetBrowser().GetCookies())
			}),

			// wait waits for specified seconds (simple implementation).
			// Usage: wait(browser, seconds)
			"wait": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("wait() requires 2 arguments (browser, seconds)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("wait() first argument must be HLBR_BROWSER")
				}
				secs, ok := args[1].(*objects.Int)
				if !ok {
					if f, ok := args[1].(*objects.Float); ok {
						time.Sleep(time.Duration(f.Value * float64(time.Second)))
						return br
					}
					return Error("wait() second argument must be INT or FLOAT")
				}
				time.Sleep(time.Duration(secs.Value) * time.Second)
				return br
			}),

			// getSessionStorageItem gets a single sessionStorage item.
			// Usage: getSessionStorageItem(browser, key) -> string
			"getSessionStorageItem": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("getSessionStorageItem() requires 2 arguments (browser, key)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getSessionStorageItem() first argument must be HLBR_BROWSER")
				}
				key, ok := args[1].(*objects.String)
				if !ok {
					return Error("getSessionStorageItem() second argument must be STRING")
				}
				storage := br.GetBrowser().GetSessionStorage()
				if storage == nil {
					return objects.NULL
				}
				if val, exists := storage[key.Value]; exists {
					return String(val)
				}
				return objects.NULL
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
			// setDebug enables or disables debug mode for a browser.
			// Usage: setDebug(browser, enabled)
			"setDebug": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("setDebug() requires 2 arguments (browser, enabled)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setDebug() first argument must be HLBR_BROWSER")
				}
				enabled, ok := args[1].(*objects.Bool)
				if !ok {
					return Error("setDebug() second argument must be BOOL")
				}
				br.GetBrowser().SetDebug(enabled.Value)
				return br
			}),

			// setUserAgent sets the User-Agent header for the browser's HTTP client.
			// Usage: setUserAgent(browser, userAgent)
			"setUserAgent": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("setUserAgent() requires 2 arguments (browser, userAgent)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setUserAgent() first argument must be HLBR_BROWSER")
				}
				ua, ok := args[1].(*objects.String)
				if !ok {
					return Error("setUserAgent() second argument must be STRING")
				}
				br.GetBrowser().Client().SetUserAgent(ua.Value)
				return br
			}),

			// setHeader sets a custom HTTP header for the browser's HTTP client.
			// Usage: setHeader(browser, key, value)
			"setHeader": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("setHeader() requires 3 arguments (browser, key, value)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setHeader() first argument must be HLBR_BROWSER")
				}
				key, ok := args[1].(*objects.String)
				if !ok {
					return Error("setHeader() second argument must be STRING")
				}
				val, ok := args[2].(*objects.String)
				if !ok {
					return Error("setHeader() third argument must be STRING")
				}
				br.GetBrowser().Client().SetHeader(key.Value, val.Value)
				return br
			}),

			// setHeaders sets multiple HTTP headers from a map.
			// Usage: setHeaders(browser, headersMap)
			"setHeaders": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("setHeaders() requires 2 arguments (browser, headers)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("setHeaders() first argument must be HLBR_BROWSER")
				}
				headers, ok := args[1].(*objects.Map)
				if !ok {
					return Error("setHeaders() second argument must be MAP")
				}
				for _, pair := range headers.Pairs {
					key, ok1 := pair.Key.(*objects.String)
					val, ok2 := pair.Value.(*objects.String)
					if ok1 && ok2 {
						br.GetBrowser().Client().SetHeader(key.Value, val.Value)
					}
				}
				return br
			}),

			// getTitle returns the page title.
			// Usage: getTitle(browser) -> string
			"getTitle": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("getTitle() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("getTitle() argument must be HLBR_BROWSER")
				}
				return String(br.GetBrowser().GetTitle())
			}),

			// find queries the browser's document with a CSS selector.
			// Usage: find(browser, selector) -> HlbrNode or NULL
			"find": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("find() requires 2 arguments (browser, selector)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("find() first argument must be HLBR_BROWSER")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("find() second argument must be STRING")
				}
				result := br.GetBrowser().QuerySelector(sel.Value)
				if result == nil {
					return objects.NULL
				}
				return objects.NewHlbrNode(result)
			}),

			// findAll queries all matching elements in the browser's document.
			// Usage: findAll(browser, selector) -> array of HlbrNode
			"findAll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("findAll() requires 2 arguments (browser, selector)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("findAll() first argument must be HLBR_BROWSER")
				}
				sel, ok := args[1].(*objects.String)
				if !ok {
					return Error("findAll() second argument must be STRING")
				}
				nodes := br.GetBrowser().QuerySelectorAll(sel.Value)
				elems := make([]objects.Object, len(nodes))
				for i, n := range nodes {
					elems[i] = objects.NewHlbrNode(n)
				}
				return Array(elems...)
			}),

			// analyzeVueTemplates analyzes JavaScript source code to find Vue.js templates
			// and extract form fields. This is useful for SPA pages that render forms dynamically.
			// Usage: analyzeVueTemplates(browser) -> array of maps with name, type, placeholder, etc.
			"analyzeVueTemplates": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("analyzeVueTemplates() requires 1 argument (browser)")
				}
				br, ok := args[0].(*objects.HlbrBrowser)
				if !ok {
					return Error("analyzeVueTemplates() argument must be HLBR_BROWSER")
				}
				fields := br.GetBrowser().AnalyzeVueTemplates()
				elems := make([]objects.Object, len(fields))
				for i, f := range fields {
					pairs := make(map[objects.HashKey]objects.MapPair)
					addToMap(pairs, "name", f.Name)
					addToMap(pairs, "type", f.Type)
					addToMap(pairs, "placeholder", f.Placeholder)
					addToMap(pairs, "label", f.Label)
					addToMap(pairs, "id", f.ID)
					addToMapBool(pairs, "required", f.Required)
					elems[i] = &objects.Map{Pairs: pairs}
				}
				return Array(elems...)
			}),

			// rsaEncrypt encrypts plaintext using RSA with hex-encoded modulus and exponent.
			// This is useful for web applications that use RSA encryption for login passwords.
			// Usage: rsaEncrypt(plaintext, hexModulus, hexExponent) -> hex-encoded ciphertext
			// rsaEncrypt encrypts plaintext using RSA PKCS1v15 with hex-encoded modulus and exponent.
			// Usage: rsaEncrypt(plaintext, hexModulus, hexExponent) -> hex-encoded ciphertext
			"rsaEncrypt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("rsaEncrypt() requires 3 arguments (plaintext, hexModulus, hexExponent)")
				}
				plaintext, ok := args[0].(*objects.String)
				if !ok {
					return Error("rsaEncrypt() first argument must be STRING (plaintext)")
				}
				hexModulus, ok := args[1].(*objects.String)
				if !ok {
					return Error("rsaEncrypt() second argument must be STRING (hex modulus)")
				}
				hexExponent, ok := args[2].(*objects.String)
				if !ok {
					return Error("rsaEncrypt() third argument must be STRING (hex exponent)")
				}
				ciphertext, err := hlbr.RSAEncryptHex(plaintext.Value, hexModulus.Value, hexExponent.Value)
				if err != nil {
					return Error("rsaEncrypt() failed: " + err.Error())
				}
				return String(ciphertext)
			}),

			// rsaEncryptRaw encrypts plaintext using raw RSA (no padding, reversed byte order)
			// with hex-encoded modulus and exponent. This matches the behavior of common
			// JavaScript RSA libraries (e.g. JSEncrypt) that use NoPadding with reversed bytes.
			// Usage: rsaEncryptRaw(plaintext, hexModulus, hexExponent) -> hex-encoded ciphertext
			"rsaEncryptRaw": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("rsaEncryptRaw() requires 3 arguments (plaintext, hexModulus, hexExponent)")
				}
				plaintext, ok := args[0].(*objects.String)
				if !ok {
					return Error("rsaEncryptRaw() first argument must be STRING (plaintext)")
				}
				hexModulus, ok := args[1].(*objects.String)
				if !ok {
					return Error("rsaEncryptRaw() second argument must be STRING (hex modulus)")
				}
				hexExponent, ok := args[2].(*objects.String)
				if !ok {
					return Error("rsaEncryptRaw() third argument must be STRING (hex exponent)")
				}
				ciphertext, err := hlbr.RSAEncryptHexRaw(plaintext.Value, hexModulus.Value, hexExponent.Value)
				if err != nil {
					return Error("rsaEncryptRaw() failed: " + err.Error())
				}
				return String(ciphertext)
			}),
		},
	})
}

// addToMap adds a string key-value pair to a map.
func addToMap(pairs map[objects.HashKey]objects.MapPair, key, value string) {
	keyObj := objects.NewString(key)
	pairs[keyObj.HashKey()] = objects.MapPair{Key: keyObj, Value: objects.NewString(value)}
}

// addToMapBool adds a bool key-value pair to a map.
func addToMapBool(pairs map[objects.HashKey]objects.MapPair, key string, value bool) {
	keyObj := objects.NewString(key)
	pairs[keyObj.HashKey()] = objects.MapPair{Key: keyObj, Value: &objects.Bool{Value: value}}
}
