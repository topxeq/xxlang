// pkg/stdlib/hlbr.go
// HLBR module for Xxlang - lightweight headless browser.
// This module provides a zero-dependency headless browser that does not
// require Chrome/Chromium. It uses a built-in HTTP client, HTML parser,
// DOM tree, CSS selector engine, and JavaScript interpreter.
package stdlib

import (
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
		},
	})
}
