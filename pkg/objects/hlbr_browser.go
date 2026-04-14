// pkg/objects/hlbr_browser.go
// HlbrBrowser object for Xxlang - lightweight headless browser based on hlbr.
// Unlike the Rod-based browser module, this does not require Chrome/Chromium
// and uses a built-in HTTP client, HTML parser, DOM, CSS selector engine,
// and JavaScript interpreter.
package objects

import (
	"fmt"
	"net/http"
	"time"
	"unsafe"

	"github.com/topxeq/xxlang/pkg/hlbr"
	"github.com/topxeq/xxlang/pkg/hlbr/dom"
)

// HlbrBrowser wraps hlbr.Browser as an Xxlang object.
type HlbrBrowser struct {
	browser *hlbr.Browser
}

// Type returns the object type.
func (b *HlbrBrowser) Type() ObjectType { return HlbrBrowserType }

// TypeTag returns the fast type tag.
func (b *HlbrBrowser) TypeTag() TypeTag { return TagHlbrBrowser }

// Inspect returns a string representation.
func (b *HlbrBrowser) Inspect() string {
	url := ""
	if b.browser != nil {
		url = b.browser.GetURL()
	}
	if url == "" {
		url = "about:blank"
	}
	return fmt.Sprintf("<HLBR_BROWSER url=%s>", url)
}

// ToBool converts the object to a boolean.
func (b *HlbrBrowser) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the object.
func (b *HlbrBrowser) HashKey() HashKey {
	return HashKey{
		Type:  HlbrBrowserType,
		Value: uint64(uintptr(unsafe.Pointer(b))),
	}
}

// GetBrowser returns the underlying hlbr.Browser instance.
func (b *HlbrBrowser) GetBrowser() *hlbr.Browser {
	return b.browser
}

// NewHlbrBrowser creates a new HlbrBrowser instance with the given options.
func NewHlbrBrowser(args ...Object) Object {
	var opts *hlbr.Options

	if len(args) > 0 {
		if m, ok := args[0].(*Map); ok {
			opts = &hlbr.Options{}
			if v := hlbrGetStringFromMap(m, "userAgent"); v != "" {
				opts.UserAgent = v
			}
			if v := hlbrGetStringFromMap(m, "proxy"); v != "" {
				opts.Proxy = v
			}
			if v, ok := hlbrGetIntFromMap(m, "timeout"); ok && v > 0 {
				opts.Timeout = time.Duration(v) * time.Second
			}
		}
	}

	b, err := hlbr.New(opts)
	if err != nil {
		return newError("hlbr new failed: %s", err.Error())
	}

	return &HlbrBrowser{browser: b}
}

// HlbrNode wraps dom.Node as an Xxlang object for DOM element access.
type HlbrNode struct {
	node *dom.Node
}

// Type returns the object type.
func (n *HlbrNode) Type() ObjectType { return HlbrNodeType }

// TypeTag returns the fast type tag.
func (n *HlbrNode) TypeTag() TypeTag { return TagHlbrNode }

// Inspect returns a string representation.
func (n *HlbrNode) Inspect() string {
	if n.node == nil {
		return "<HLBR_NODE nil>"
	}
	tag := n.node.Data
	id := n.node.GetAttribute("id")
	if id != "" {
		return fmt.Sprintf("<HLBR_NODE <%s#%s>>", tag, id)
	}
	cls := n.node.GetAttribute("class")
	if cls != "" {
		return fmt.Sprintf("<HLBR_NODE <%s.%s>>", tag, cls)
	}
	return fmt.Sprintf("<HLBR_NODE <%s>>", tag)
}

// ToBool converts the object to a boolean.
func (n *HlbrNode) ToBool() *Bool {
	return &Bool{Value: n.node != nil}
}

// HashKey returns a hash key for the object.
func (n *HlbrNode) HashKey() HashKey {
	return HashKey{
		Type:  HlbrNodeType,
		Value: uint64(uintptr(unsafe.Pointer(n))),
	}
}

// GetNode returns the underlying dom.Node.
func (n *HlbrNode) GetNode() *dom.Node {
	return n.node
}

// NewHlbrNode wraps a dom.Node into an HlbrNode object.
func NewHlbrNode(node *dom.Node) *HlbrNode {
	if node == nil {
		return nil
	}
	return &HlbrNode{node: node}
}

// hlbrGoValueToObject converts a Go value from hlbr.Evaluate to an Xxlang Object.
func hlbrGoValueToObject(v interface{}) Object {
	if v == nil {
		return NULL
	}
	switch val := v.(type) {
	case bool:
		return &Bool{Value: val}
	case float64:
		if val == float64(int64(val)) {
			return NewInt(int64(val))
		}
		return &Float{Value: val}
	case string:
		return NewString(val)
	case []interface{}:
		elems := make([]Object, len(val))
		for i, item := range val {
			elems[i] = hlbrGoValueToObject(item)
		}
		return &Array{Elements: elems}
	case map[string]interface{}:
		pairs := make(map[HashKey]MapPair)
		for k, v := range val {
			key := NewString(k)
			pairs[key.HashKey()] = MapPair{Key: key, Value: hlbrGoValueToObject(v)}
		}
		return &Map{Pairs: pairs}
	default:
		return NewString(fmt.Sprintf("%v", v))
	}
}

// hlbrCookiesToXxArray converts http.Cookie slice to Xxlang Array.
func hlbrCookiesToXxArray(cookies []*http.Cookie) Object {
	elems := make([]Object, len(cookies))
	for i, c := range cookies {
		pairs := make(map[HashKey]MapPair)
		nameKey := NewString("name")
		nameVal := NewString(c.Name)
		pairs[nameKey.HashKey()] = MapPair{Key: nameKey, Value: nameVal}

		valKey := NewString("value")
		valVal := NewString(c.Value)
		pairs[valKey.HashKey()] = MapPair{Key: valKey, Value: valVal}

		domainKey := NewString("domain")
		domainVal := NewString(c.Domain)
		pairs[domainKey.HashKey()] = MapPair{Key: domainKey, Value: domainVal}

		pathKey := NewString("path")
		pathVal := NewString(c.Path)
		pairs[pathKey.HashKey()] = MapPair{Key: pathKey, Value: pathVal}

		elems[i] = &Map{Pairs: pairs}
	}
	return &Array{Elements: elems}
}

// hlbrNodesToXxArray converts dom.Node slice to Xxlang Array of HlbrNode objects.
func hlbrNodesToXxArray(nodes []*dom.Node) Object {
	elems := make([]Object, len(nodes))
	for i, n := range nodes {
		elems[i] = NewHlbrNode(n)
	}
	return &Array{Elements: elems}
}

// hlbrGetStringFromMap extracts a string value from an Xxlang Map by key.
func hlbrGetStringFromMap(m *Map, key string) string {
	keyObj := NewString(key)
	if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
		if s, ok := pair.Value.(*String); ok {
			return s.Value
		}
	}
	return ""
}

// hlbrGetIntFromMap extracts an int64 value from an Xxlang Map by key.
func hlbrGetIntFromMap(m *Map, key string) (int64, bool) {
	keyObj := NewString(key)
	if pair, exists := m.Pairs[keyObj.HashKey()]; exists {
		if i, ok := pair.Value.(*Int); ok {
			return i.Value, true
		}
		if f, ok := pair.Value.(*Float); ok {
			return int64(f.Value), true
		}
	}
	return 0, false
}
