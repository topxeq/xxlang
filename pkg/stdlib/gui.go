// Package stdlib provides the standard library modules for Xxlang.
// This file contains the GUI module for WebView2 support.
package stdlib

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/webview2"
)

func init() {
	Register(&Module{
		Name: "gui",
		Exports: map[string]objects.Object{
			// ============================================================
			// Window Creation
			// ============================================================

			// createWindow creates a new WebView2 window.
			// Usage: handle = gui.createWindow(title, width, height)
			//        handle = gui.createWindow(title, width, height, userDataFolder)
			"createWindow": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("createWindow() requires at least 3 arguments: title, width, height")
				}

				title, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument must be a string (title)")
				}

				width, ok := args[1].(*objects.Int)
				if !ok {
					return Error("second argument must be an integer (width)")
				}

				height, ok := args[2].(*objects.Int)
				if !ok {
					return Error("third argument must be an integer (height)")
				}

				config := webview2.WebView2Config{
					Title:  title.Value,
					Width:  int(width.Value),
					Height: int(height.Value),
					Debug:  false,
				}

				// Set default user data folder if not provided
				if len(args) > 3 {
					if folder, ok := args[3].(*objects.String); ok {
						config.UserDataFolder = folder.Value
					}
				}

				// If no user data folder specified, use temp directory
				if config.UserDataFolder == "" {
					tempDir := os.TempDir()
					config.UserDataFolder = filepath.Join(tempDir, "xxlang_webview2")
				}

				// Optional: debug mode
				if len(args) > 4 {
					if debug, ok := args[4].(*objects.Bool); ok {
						config.Debug = debug.Value
					} else {
						return Error("fourth argument (debug) must be a boolean")
					}
				}

				wv, err := webview2.NewWebView2(config)
				if err != nil {
					return Error("failed to create WebView: " + err.Error())
				}

				return objects.NewWebView(wv)
			}),

			// ============================================================
			// Content Setting
			// ============================================================

			// setHTML sets the HTML content of the WebView.
			// Usage: gui.setHTML(handle, html)
			"setHTML": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("setHTML() requires 2 arguments: handle, html")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("first argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return Error("WebView is closed")
				}

				html, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (html)")
				}

				handle := wv.Handle.(interface {
					NavigateToString(string) error
				})
				if err := handle.NavigateToString(html.Value); err != nil {
					return Error("failed to set HTML: " + err.Error())
				}

				return objects.NULL
			}),

			// loadURL navigates to a URL.
			// Usage: gui.loadURL(handle, url)
			"loadURL": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("loadURL() requires 2 arguments: handle, url")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("first argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return Error("WebView is closed")
				}

				url, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (url)")
				}

				handle := wv.Handle.(interface {
					Navigate(string) error
				})
				if err := handle.Navigate(url.Value); err != nil {
					return Error("failed to navigate: " + err.Error())
				}

				return objects.NULL
			}),

			// ============================================================
			// JavaScript Interaction
			// ============================================================

			// evalJS executes JavaScript and returns the result.
			// Usage: result = gui.evalJS(handle, script)
			"evalJS": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("evalJS() requires 2 arguments: handle, script")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("first argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return Error("WebView is closed")
				}

				script, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (script)")
				}

				// Use synchronous wrapper
				type syncResult struct {
					result string
					err    error
				}
				resultChan := make(chan syncResult, 1)

				handle := wv.Handle.(interface {
					ExecuteScript(string, func(string, error)) error
				})
				err := handle.ExecuteScript(script.Value, func(result string, err error) {
					resultChan <- syncResult{result: result, err: err}
				})

				if err != nil {
					return Error("script execution failed: " + err.Error())
				}

				// Wait for result (this is a simplified sync version)
				// In production, this should use proper async handling
				res := <-resultChan
				if res.err != nil {
					return Error("script execution error: " + res.err.Error())
				}

				return String(res.result)
			}),

			// evalJSAsync executes JavaScript without waiting for result.
			// Non-blocking - use this with poll() mode.
			// Usage: gui.evalJSAsync(handle, script)
			"evalJSAsync": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("evalJSAsync() requires 2 arguments: handle, script")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("first argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return Error("WebView is closed")
				}

				script, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (script)")
				}

				handle := wv.Handle.(interface {
					ExecuteScript(string, func(string, error)) error
				})
				// Pass nil callback for fire-and-forget
				err := handle.ExecuteScript(script.Value, nil)
				if err != nil {
					return Error("script execution failed: " + err.Error())
				}

				return objects.NULL
			}),

			// bind binds a Xxlang function to be callable from JavaScript.
			// Usage: gui.bind(handle, name, function)
			"bind": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("bind() requires 3 arguments: handle, name, function")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("first argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return Error("WebView is closed")
				}

				name, ok := args[1].(*objects.String)
				if !ok {
					return Error("second argument must be a string (name)")
				}

				fn, ok := args[2].(*objects.Function)
				if !ok {
					return Error("third argument must be a function")
				}

				wv.AddCallback(name.Value, fn)

				// Register with WebView2
				handle := wv.Handle.(interface {
					BindFunction(string, func(args map[string]interface{}))
				})
				handle.BindFunction(name.Value, func(args map[string]interface{}) {
					// This would need proper VM integration to call the Xxlang function
					// For now, we just store the callback
				})

				return objects.NULL
			}),

			// ============================================================
			// Window Control
			// ============================================================

			// loop runs the WebView message loop (blocking).
			// Usage: gui.loop(handle)
			"loop": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("run() requires 1 argument: handle")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return Error("WebView is closed")
				}

				handle := wv.Handle.(interface {
					Run()
				})
				handle.Run()

				return objects.NULL
			}),

			// poll processes a single message without blocking.
			// Returns true if a message was processed, false if no message.
			// Usage: processed = gui.poll(handle)
			"poll": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("poll() requires 1 argument: handle")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return objects.FALSE
				}

				handle := wv.Handle.(interface {
					Poll() bool
				})
				if handle.Poll() {
					return objects.TRUE
				}
				return objects.FALSE
			}),

			// popMessage gets the next message from the queue.
			// Returns the message string or empty string if none.
			// Usage: msg = gui.popMessage(handle)
			"popMessage": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("popMessage() requires 1 argument: handle")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return String("")
				}

				handle := wv.Handle.(interface {
					PopMessage() string
				})
				return String(handle.PopMessage())
			}),

			// hasMessages returns true if there are pending messages.
			// Usage: has = gui.hasMessages(handle)
			"hasMessages": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("hasMessages() requires 1 argument: handle")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return objects.FALSE
				}

				handle := wv.Handle.(interface {
					HasMessages() bool
				})
				if handle.HasMessages() {
					return objects.TRUE
				}
				return objects.FALSE
			}),

			// close closes the WebView.
			// Usage: gui.close(handle)
			"close": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("close() requires 1 argument: handle")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("argument must be a WebView handle")
				}

				if !wv.IsClosed() {
					handle := wv.Handle.(interface {
						Close()
					})
					handle.Close()
					wv.SetClosed(true)
				}

				return objects.NULL
			}),

			// ============================================================
			// Utility Functions
			// ============================================================

			// getVersion returns the installed WebView2 runtime version.
			// Usage: version = gui.getVersion()
			"getVersion": BuiltinFunc(func(args ...objects.Object) objects.Object {
				version, err := webview2.GetInstalledVersion()
				if err != nil {
					return Error("failed to get WebView2 version: " + err.Error())
				}
				return String(version)
			}),

			// isClosed checks if the WebView is closed.
			// Usage: closed = gui.isClosed(handle)
			"isClosed": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("isClosed() requires 1 argument: handle")
				}

				wv, ok := args[0].(*objects.WebView)
				if !ok {
					return Error("argument must be a WebView handle")
				}

				if wv.IsClosed() {
					return objects.TRUE
				}
				return objects.FALSE
			}),
		},
	})
}

// Helper for JSON marshaling (used in bind)
func _() {
	_ = json.Marshal
}
