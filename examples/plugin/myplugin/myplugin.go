// examples/plugin/myplugin/myplugin.go
// Plugin example - demonstrates how to create a native Go plugin
// and use it from xxlang code.
//
// Build command:
//   go build -buildmode=plugin -o myplugin.so myplugin.go
//
// Usage from xxlang:
//   import "plugin/myplugin"
//   myplugin.hello()
package main

import (
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/plugin"
)

// MyPlugin is a simple example plugin that exports a hello function
type MyPlugin struct{}

// Name returns the plugin name
func (p *MyPlugin) Name() string {
	return "myplugin"
}

// Exports returns the plugin's exports
func (p *MyPlugin) Exports() map[string]objects.Object {
	return map[string]objects.Object{
		"hello": &objects.Builtin{
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.String{Value: "Hello from plugin!"}
			},
		},
		"add": &objects.Builtin{
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return &objects.Error{Message: "add requires 2 arguments"}
				}
				a, ok1 := args[0].(*objects.Int)
				b, ok2 := args[1].(*objects.Int)
				if !ok1 || !ok2 {
					return &objects.Error{Message: "arguments must be integers"}
				}
				return &objects.Int{Value: a.Value + b.Value}
			},
		},
	}
}

// Plugin instance (exported for plugin loading)
var Plugin plugin.Plugin = &MyPlugin{}

func init() {
	// Register the plugin
	plugin.Register(Plugin)
}

func main() {} // Required for plugin build
