//go:build !linux && !darwin && !freebsd

// pkg/plugin/plugin_stub.go
// Stub implementation for platforms that don't support Go plugins (Windows, etc).
package plugin

import (
	"fmt"
	"runtime"
)

// loadPluginSO returns an error on unsupported platforms.
func loadPluginSO(path string) (Plugin, error) {
	return nil, fmt.Errorf("native plugins are not supported on %s (only Linux, macOS, and FreeBSD)", runtime.GOOS)
}
