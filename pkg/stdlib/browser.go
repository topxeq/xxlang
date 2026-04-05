// pkg/stdlib/browser.go
// Browser module for Xxlang - Web scraping with Rod
package stdlib

import (
	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	// Register browser module
	Register(&Module{
		Name: "browser",
		Exports: map[string]objects.Object{
			"open": &objects.Builtin{Fn: objects.NewBrowser},
		},
	})
}
