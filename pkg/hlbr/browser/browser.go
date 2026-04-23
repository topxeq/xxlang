package browser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
	"github.com/topxeq/xxlang/pkg/hlbr/httpclient"
	"github.com/topxeq/xxlang/pkg/hlbr/jsengine"
	"github.com/topxeq/xxlang/pkg/hlbr/renderer"
)

type Browser struct {
	client     *httpclient.Client
	doc        *dom.Document
	vm         *jsengine.VM
	currentURL string
	history    []string
	debug      bool
}

type Options struct {
	UserAgent string
	Proxy     string
	Timeout   int
	Debug     bool
}

func New(opts *Options) *Browser {
	if opts == nil {
		opts = &Options{}
	}
	httpOpts := &httpclient.Options{
		UserAgent: opts.UserAgent,
		Proxy:     opts.Proxy,
		Timeout:   opts.Timeout,
		Debug:     opts.Debug,
	}
	return &Browser{
		client:  httpclient.NewClient(httpOpts),
		history: make([]string, 0),
		debug:   opts.Debug,
	}
}

// SetDebug enables or disables debug mode.
func (b *Browser) SetDebug(debug bool) {
	b.debug = debug
	b.client.SetDebug(debug)
}

func (b *Browser) debugLog(format string, args ...interface{}) {
	if b.debug {
		fmt.Printf("[HLBR DEBUG] "+format+"\n", args...)
	}
}

func (b *Browser) Navigate(rawURL string) error {
	b.debugLog("Navigate started: %s", rawURL)

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		b.debugLog("URL parse error: %v", err)
		return err
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}
	fullURL := parsedURL.String()
	b.debugLog("Full URL: %s", fullURL)

	b.debugLog("Sending HTTP GET request...")
	resp, err := b.client.Get(fullURL)
	if err != nil {
		b.debugLog("HTTP request error: %v", err)
		return err
	}
	b.debugLog("Response received: status=%d, bodyLen=%d", resp.StatusCode, len(resp.Body))

	b.currentURL = resp.URL
	b.history = append(b.history, resp.URL)

	b.debugLog("Parsing HTML...")
	b.doc = htmlparser.Parse(resp.Body)
	b.debugLog("HTML parsed, creating JS VM...")
	b.vm = jsengine.NewVM(b.doc)

	b.debugLog("Executing scripts...")
	b.executeScripts()
	b.debugLog("Navigate completed")

	return nil
}

func (b *Browser) executeScripts() {
	if b.doc == nil || b.vm == nil {
		return
	}

	scripts := dom.GetElementsByTagName(b.doc.Root, "script")
	for _, script := range scripts {
		src := script.GetAttribute("src")
		if src != "" {
			b.loadExternalScript(src)
			continue
		}

		code := script.TextContent()
		if code != "" {
			b.vm.Run(code)
		}
	}
}

func (b *Browser) loadExternalScript(src string) {
	if !strings.HasPrefix(src, "http") {
		baseURL, _ := url.Parse(b.currentURL)
		if strings.HasPrefix(src, "/") {
			src = baseURL.Scheme + "://" + baseURL.Host + src
		} else {
			src = baseURL.Scheme + "://" + baseURL.Host + "/" + src
		}
	}

	resp, err := b.client.Get(src)
	if err != nil {
		return
	}

	if b.vm != nil {
		b.vm.Run(resp.Body)
	}
}

func (b *Browser) Evaluate(code string) (any, error) {
	if b.vm == nil {
		return nil, nil
	}
	val, err := b.vm.Run(code)
	if err != nil {
		return nil, err
	}
	return jsValueToGo(val), nil
}

func (b *Browser) QuerySelector(selector string) *dom.Node {
	if b.doc == nil {
		return nil
	}
	return dom.QuerySelector(b.doc.Root, selector)
}

func (b *Browser) QuerySelectorAll(selector string) []*dom.Node {
	if b.doc == nil {
		return nil
	}
	return dom.QuerySelectorAll(b.doc.Root, selector)
}

func (b *Browser) GetTitle() string {
	if b.doc == nil {
		return ""
	}
	return b.doc.Title()
}

func (b *Browser) GetHTML() string {
	if b.doc == nil || b.doc.Root == nil {
		return ""
	}
	return b.doc.Root.InnerHTML()
}

func (b *Browser) GetText() string {
	if b.doc == nil || b.doc.Root == nil {
		return ""
	}
	return b.doc.Root.TextContent()
}

func (b *Browser) ScreenshotText(width int) string {
	return renderer.ScreenshotText(b.doc, width)
}

func (b *Browser) ScreenshotTextToFile(path string, width int) error {
	return renderer.ScreenshotTextToFile(b.doc, path, width)
}

func (b *Browser) GetURL() string {
	return b.currentURL
}

func (b *Browser) Document() *dom.Document {
	return b.doc
}

func (b *Browser) VM() *jsengine.VM {
	return b.vm
}

func (b *Browser) Client() *httpclient.Client {
	return b.client
}

func (b *Browser) History() []string {
	return b.history
}

func (b *Browser) Back() error {
	if len(b.history) < 2 {
		return nil
	}
	b.history = b.history[:len(b.history)-1]
	prevURL := b.history[len(b.history)-1]
	return b.Navigate(prevURL)
}

// GetLocalStorage returns the localStorage data as a map
func (b *Browser) GetLocalStorage() map[string]string {
	if b.vm == nil {
		return nil
	}
	return b.vm.LocalStorage
}

// GetSessionStorage returns the sessionStorage data as a map
func (b *Browser) GetSessionStorage() map[string]string {
	if b.vm == nil {
		return nil
	}
	return b.vm.SessionStorage
}

// SetLocalStorageItem sets a key-value pair in localStorage
func (b *Browser) SetLocalStorageItem(key, value string) {
	if b.vm != nil {
		b.vm.LocalStorage[key] = value
	}
}

// SetSessionStorageItem sets a key-value pair in sessionStorage
func (b *Browser) SetSessionStorageItem(key, value string) {
	if b.vm != nil {
		b.vm.SessionStorage[key] = value
	}
}

// GetConsoleOutput returns the console.log output
func (b *Browser) GetConsoleOutput() []string {
	if b.vm == nil {
		return nil
	}
	return b.vm.Output()
}

// SetJsDebug enables or disables JS debug mode (compatibility stub).
func (b *Browser) SetJsDebug(jsDebug bool) {
	// No dedicated JS debug toggle in this minimal browser; stub for compatibility
	b.debugLog("SetJsDebug called (stub): %v", jsDebug)
}

// WaitStable waits for the page to become stable (compatibility stub).
func (b *Browser) WaitStable(timeoutMs, stableForMs int) error {
	b.debugLog("WaitStable called (stub): timeoutMs=%d, stableForMs=%d", timeoutMs, stableForMs)
	return nil
}

// WaitStableDefault waits for the page to become stable with default timeouts (compatibility stub).
func (b *Browser) WaitStableDefault() error {
	b.debugLog("WaitStableDefault called (stub)")
	return nil
}

func jsValueToGo(v *jsengine.Value) any {
	if v == nil {
		return nil
	}
	switch v.Type {
	case "undefined", "null":
		return nil
	case "bool":
		return v.Bool
	case "number":
		return v.Num
	case "string":
		return v.Str
	case "object":
		if v.Arr != nil {
			arr := make([]any, len(v.Arr))
			for i, a := range v.Arr {
				arr[i] = jsValueToGo(a)
			}
			return arr
		}
		if v.Obj != nil {
			obj := make(map[string]any)
			for k, val := range v.Obj {
				obj[k] = jsValueToGo(val)
			}
			return obj
		}
		return nil
	case "function", "native":
		return "[function]"
	}
	return nil
}
