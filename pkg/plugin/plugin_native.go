//go:build linux || darwin || freebsd

// pkg/plugin/plugin_native.go
// Native plugin loading for supported platforms (Linux, macOS, FreeBSD).
package plugin

import (
	"fmt"
	pluginpkg "plugin"

	"github.com/topxeq/xxlang/pkg/objects"
)

// loadPluginSO loads a .so file using Go's plugin package.
func loadPluginSO(path string) (Plugin, error) {
	// Open the plugin
	p, err := pluginpkg.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %v", err)
	}

	// Try to look up different symbol names in order
	symbolNames := []string{"Plugin", "New", "Instance"}

	var pluginInstance Plugin

	for _, symName := range symbolNames {
		sym, err := p.Lookup(symName)
		if err != nil {
			continue // Try next symbol
		}

		// Try different type assertions
		switch v := sym.(type) {
		case Plugin:
			pluginInstance = v
		case *Plugin:
			if v != nil {
				pluginInstance = *v
			}
		case func() Plugin:
			pluginInstance = v()
		case func() map[string]interface{}:
			// Support the common pattern where plugins return a map
			exports := v()
			return newMapPlugin(exports)
		case map[string]interface{}:
			return newMapPlugin(v)
		}

		if pluginInstance != nil {
			break
		}
	}

	if pluginInstance == nil {
		return nil, fmt.Errorf("plugin does not export a valid Plugin symbol (tried: %v)", symbolNames)
	}

	return pluginInstance, nil
}

// mapPlugin is a simple Plugin implementation from a map of exports.
type mapPlugin struct {
	name    string
	exports map[string]interface{}
}

func newMapPlugin(exports map[string]interface{}) (Plugin, error) {
	// Get the plugin name from exports or use "unknown"
	name := "unknown"
	if n, ok := exports["name"].(string); ok {
		name = n
	} else if n, ok := exports["Name"].(string); ok {
		name = n
	}

	return &mapPlugin{
		name:    name,
		exports: exports,
	}, nil
}

func (m *mapPlugin) Name() string {
	return m.name
}

func (m *mapPlugin) Exports() map[string]objects.Object {
	result := make(map[string]objects.Object)

	for key, value := range m.exports {
		// Skip the name field
		if key == "name" || key == "Name" {
			continue
		}

		// Convert to xxlang object
		if obj, ok := value.(objects.Object); ok {
			result[key] = obj
		} else if fn, ok := value.(func(...objects.Object) objects.Object); ok {
			result[key] = &objects.Builtin{Fn: fn}
		}
		// Skip values that can't be converted
	}

	return result
}
