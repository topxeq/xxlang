// pkg/objects/browser.go
// Browser object for web scraping with Rod
package objects

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"os"
	"os/exec"
	"runtime"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"strings"
	"unsafe"
)

// RodBrowser - Browser object for Xxlang using Rod
type RodBrowser struct {
	browser    *rod.Browser
	page       *rod.Page
	chromePath string
	headless   bool
}

// Browser - alias for RodBrowser for backward compatibility
type Browser = RodBrowser

// NewBrowser - create new browser instance
func NewBrowser(args ...Object) Object {
	headless := true
	chromePath := ""
	autoDownload := true
	proxy := ""
	userAgent := ""

	// Parse options
	if len(args) > 0 {
		if opts, ok := args[0].(*Map); ok {
			if hPair, ok := opts.Pairs[NewString("headless").HashKey()]; ok {
				if hb, ok := hPair.Value.(*Bool); ok {
					headless = hb.Value
				}
			}
			if pPair, ok := opts.Pairs[NewString("chromePath").HashKey()]; ok {
				if ps, ok := pPair.Value.(*String); ok {
					chromePath = ps.Value
				}
			}
			if aPair, ok := opts.Pairs[NewString("autoDownload").HashKey()]; ok {
				if ab, ok := aPair.Value.(*Bool); ok {
					autoDownload = ab.Value
				}
			}
			if prPair, ok := opts.Pairs[NewString("proxy").HashKey()]; ok {
				if prs, ok := prPair.Value.(*String); ok {
					proxy = prs.Value
				}
			}
			if uaPair, ok := opts.Pairs[NewString("userAgent").HashKey()]; ok {
				if uas, ok := uaPair.Value.(*String); ok {
					userAgent = uas.Value
				}
			}
		}
	}

	// Create launcher
	l := launcher.New()

	// Set headless mode
	if headless {
		// Use the newer headless mode (--headless=new) which is more stable
		l = l.HeadlessNew(true)
		// Additional flags to prevent window flash on Windows
		l = l.Set("no-startup-window")
		l = l.Set("disable-backgrounding-occluded-windows")
		l = l.Set("disable-gpu")
	} else {
		l = l.Headless(false)
	}

	// Set Chrome path or find system Chrome
	if chromePath != "" {
		l = l.Bin(chromePath)
	} else if !autoDownload {
		// Try to find system Chrome
		found := findSystemChrome(l)
		if !found {
			return NewString("[error] Chrome not found. Set chromePath or enable autoDownload")
		}
	}
	// else: use auto-download (default)

	// Set proxy if provided
	if proxy != "" {
		l = l.Set(flags.ProxyServer, proxy)
	}

	// Set user agent if provided
	if userAgent != "" {
		l = l.Set("user-agent", userAgent)
	}

	// Launch browser
	u, err := l.Launch()
	if err != nil {
		return NewString("[error] Failed to launch browser: " + err.Error())
	}

	// Connect to browser
	browser := rod.New().ControlURL(u).MustConnect()

	// Create default page
	page := browser.MustPage("")

	return &RodBrowser{
		browser:    browser,
		page:       page,
		chromePath: chromePath,
		headless:   headless,
	}
}

// findSystemChrome - try to find system Chrome installation
func findSystemChrome(l *launcher.Launcher) bool {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			os.ExpandEnv(`%LOCALAPPDATA%\Google\Chrome\Application\chrome.exe`),
			os.ExpandEnv(`%PROGRAMFILES%\Microsoft\Edge\Application\msedge.exe`),
		}
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	case "linux":
		paths = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/usr/bin/microsoft-edge",
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			l.Bin(p)
			return true
		}
	}

	return false
}

// RodBrowser methods

// Get - navigate to URL
func (b *RodBrowser) Get(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] get requires url")
	}

	url, ok := args[0].(*String)
	if !ok {
		return NewString("[error] get requires string url")
	}

	b.page = b.browser.MustPage(url.Value)
	b.page.MustWaitLoad()
	return b
}

// Close - close browser
func (b *RodBrowser) Close(args ...Object) Object {
	b.browser.Close()
	return NULL
}

// Fill - fill input field
func (b *RodBrowser) Fill(args ...Object) Object {
	if len(args) < 2 {
		return NewString("[error] fill requires selector and value")
	}

	selector, ok1 := args[0].(*String)
	value, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewString("[error] fill requires string arguments")
	}

	b.page.MustElement(selector.Value).MustInput(value.Value)
	return b
}

// Click - click element
func (b *RodBrowser) Click(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] click requires selector")
	}

	selector, ok := args[0].(*String)
	if !ok {
		return NewString("[error] click requires string selector")
	}

	b.page.MustElement(selector.Value).MustClick()
	return b
}

// Wait - wait for element
func (b *RodBrowser) Wait(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] wait requires selector")
	}

	selector, ok := args[0].(*String)
	if !ok {
		return NewString("[error] wait requires string selector")
	}

	b.page.MustElement(selector.Value)
	return b
}

// WaitLoad - wait for page load
func (b *RodBrowser) WaitLoad(args ...Object) Object {
	b.page.MustWaitLoad()
	return b
}

// WaitStable - wait for page to become stable (no network activity)
func (b *RodBrowser) WaitStable(args ...Object) Object {
	b.page.MustWaitStable()
	return b
}

// Fullscreen - set browser window to fullscreen
func (b *RodBrowser) Fullscreen(args ...Object) Object {
	b.page.MustWindowFullscreen()
	return b
}

// Exists - check if element exists
func (b *RodBrowser) Exists(args ...Object) Object {
	if len(args) < 1 {
		return &Bool{Value: false}
	}

	selector, ok := args[0].(*String)
	if !ok {
		return &Bool{Value: false}
	}

	_, err := b.page.Element(selector.Value)
	return &Bool{Value: err == nil}
}

// Find - find single element
func (b *RodBrowser) Find(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] find requires selector")
	}

	selector, ok := args[0].(*String)
	if !ok {
		return NewString("[error] find requires string selector")
	}

	el, err := b.page.Element(selector.Value)
	if err != nil {
		return NULL
	}

	return &RodHTMLElement{Element: el, Page: b.page}
}

// FindAll - find all elements
func (b *RodBrowser) FindAll(args ...Object) Object {
	if len(args) < 1 {
		return &Array{Elements: []Object{}}
	}

	selector, ok := args[0].(*String)
	if !ok {
		return &Array{Elements: []Object{}}
	}

	els, err := b.page.Elements(selector.Value)
	if err != nil {
		return &Array{Elements: []Object{}}
	}

	values := make([]Object, len(els))
	for i, el := range els {
		values[i] = &RodHTMLElement{Element: el, Page: b.page}
	}

	return &Array{Elements: values}
}

// Eval - execute JavaScript
func (b *RodBrowser) Eval(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] eval requires js string")
	}

	js, ok := args[0].(*String)
	if !ok {
		return NewString("[error] eval requires string js")
	}

	// Wrap simple expressions in an arrow function with no args
	jsCode := js.Value
	trimmed := strings.TrimSpace(jsCode)
	// Check if it's already a function or IIFE
	isIIFE := strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")()")
	isFunction := strings.HasPrefix(trimmed, "function")
	// Check for arrow functions: () => or (args) => or async () =>
	isArrowFunc := strings.Contains(trimmed, "=>")

	if !isIIFE && !isFunction && !isArrowFunc {
		// Wrap as arrow function with no arguments: () => expression
		jsCode = "() => " + jsCode
	}

	result, err := b.page.Eval(jsCode)
	if err != nil {
		return NewString("[error] " + err.Error())
	}

	return parseJSResult(result.Value)
}

// GetLocalStorage - get localStorage
func (b *RodBrowser) GetLocalStorage(args ...Object) Object {
	// Use arrow function syntax that Rod expects
	js := `() => {
		var data = {};
		for (var i = 0; i < localStorage.length; i++) {
			var key = localStorage.key(i);
			data[key] = localStorage.getItem(key);
		}
		return data;
	}`

	result, err := b.page.Eval(js)
	if err != nil {
		return NewString("[error] " + err.Error())
	}

	return parseJSResult(result.Value)
}

// GetSessionStorage - get sessionStorage
func (b *RodBrowser) GetSessionStorage(args ...Object) Object {
	// Use arrow function syntax that Rod expects
	js := `() => {
		var data = {};
		for (var i = 0; i < sessionStorage.length; i++) {
			var key = sessionStorage.key(i);
			data[key] = sessionStorage.getItem(key);
		}
		return data;
	}`

	result, err := b.page.Eval(js)
	if err != nil {
		return NewString("[error] " + err.Error())
	}

	return parseJSResult(result.Value)
}

// SetLocalStorage - set localStorage item
func (b *RodBrowser) SetLocalStorage(args ...Object) Object {
	if len(args) < 2 {
		return NewString("[error] setLocalStorage requires key and value")
	}

	key, ok1 := args[0].(*String)
	value, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewString("[error] setLocalStorage requires string arguments")
	}

	// Use arrow function IIFE syntax
	b.page.MustEval(`((k, v) => localStorage.setItem(k, v))`, key.Value, value.Value)
	return NULL
}

// SetSessionStorage - set sessionStorage item
func (b *RodBrowser) SetSessionStorage(args ...Object) Object {
	if len(args) < 2 {
		return NewString("[error] setSessionStorage requires key and value")
	}

	key, ok1 := args[0].(*String)
	value, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewString("[error] setSessionStorage requires string arguments")
	}

	// Use arrow function IIFE syntax
	b.page.MustEval(`((k, v) => sessionStorage.setItem(k, v))`, key.Value, value.Value)
	return NULL
}

// GetCookies - get cookies
func (b *RodBrowser) GetCookies(args ...Object) Object {
	cookies, err := b.page.Cookies([]string{})
	if err != nil {
		return &Array{Elements: []Object{}}
	}

	values := make([]Object, len(cookies))
	for i, c := range cookies {
		pairs := make(map[HashKey]MapPair)
		namePair := MapPair{Key: NewString("name"), Value: NewString(c.Name)}
		valuePair := MapPair{Key: NewString("value"), Value: NewString(c.Value)}
		domainPair := MapPair{Key: NewString("domain"), Value: NewString(c.Domain)}
		pathPair := MapPair{Key: NewString("path"), Value: NewString(c.Path)}
		pairs[namePair.Key.HashKey()] = namePair
		pairs[valuePair.Key.HashKey()] = valuePair
		pairs[domainPair.Key.HashKey()] = domainPair
		pairs[pathPair.Key.HashKey()] = pathPair

		values[i] = &Map{Pairs: pairs}
	}

	return &Array{Elements: values}
}

// SetCookies - set cookies
func (b *RodBrowser) SetCookies(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] setCookies requires cookies array")
	}

	cookies, ok := args[0].(*Array)
	if !ok {
		return NewString("[error] setCookies requires array")
	}

	var protoCookies []*proto.NetworkCookieParam
	for _, c := range cookies.Elements {
		if m, ok := c.(*Map); ok {
			name := getStringFromMap(m, "name")
			value := getStringFromMap(m, "value")
			domain := getStringFromMap(m, "domain")
			path := getStringFromMap(m, "path")
			if name != "" && value != "" {
				cookie := &proto.NetworkCookieParam{
					Name:  name,
					Value: value,
				}
				if domain != "" {
					cookie.Domain = domain
				}
				if path != "" {
					cookie.Path = path
				}
				protoCookies = append(protoCookies, cookie)
			}
		}
	}

	if len(protoCookies) > 0 {
		b.page.MustSetCookies(protoCookies...)
	}

	return NULL
}

// ClearCookies - clear all cookies
func (b *RodBrowser) ClearCookies(args ...Object) Object {
	// Clear browser cookies using proto call
	proto.NetworkClearBrowserCookies{}.Call(b.browser.Client(nil))
	return NULL
}

// SaveStorage - save storage to file
func (b *RodBrowser) SaveStorage(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] saveStorage requires path")
	}

	path, ok := args[0].(*String)
	if !ok {
		return NewString("[error] saveStorage requires string path")
	}

	local := b.GetLocalStorage()
	session := b.GetSessionStorage()
	cookies := b.GetCookies()

	// Convert to map for JSON
	data := map[string]interface{}{
		"localStorage":   toGoValue(local),
		"sessionStorage": toGoValue(session),
		"cookies":        toGoValue(cookies),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return NewString("[error] " + err.Error())
	}

	// Ensure directory exists
	dir := filepath.Dir(path.Value)
	os.MkdirAll(dir, 0755)

	err = ioutil.WriteFile(path.Value, jsonData, 0644)
	if err != nil {
		return NewString("[error] " + err.Error())
	}

	return NULL
}

// LoadStorage - load storage from file
func (b *RodBrowser) LoadStorage(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] loadStorage requires path")
	}

	path, ok := args[0].(*String)
	if !ok {
		return NewString("[error] loadStorage requires string path")
	}

	data, err := ioutil.ReadFile(path.Value)
	if err != nil {
		return NewString("[error] " + err.Error())
	}

	var storage map[string]interface{}
	err = json.Unmarshal(data, &storage)
	if err != nil {
		return NewString("[error] " + err.Error())
	}

	// Restore localStorage
	if local, ok := storage["localStorage"].(map[string]interface{}); ok {
		for k, v := range local {
			b.page.MustEval(`(function(k, v) { localStorage.setItem(k, v); })()`, k, toStringInterface(v))
		}
	}

	// Restore cookies
	if cookies, ok := storage["cookies"].([]interface{}); ok {
		var protoCookies []*proto.NetworkCookieParam
		for _, c := range cookies {
			if cm, ok := c.(map[string]interface{}); ok {
				if name, ok := cm["name"].(string); ok {
					if value, ok := cm["value"].(string); ok {
						cookie := &proto.NetworkCookieParam{
							Name:  name,
							Value: value,
						}
						if domain, ok := cm["domain"].(string); ok && domain != "" {
							cookie.Domain = domain
						}
						if path, ok := cm["path"].(string); ok && path != "" {
							cookie.Path = path
						}
						protoCookies = append(protoCookies, cookie)
					}
				}
			}
		}
		if len(protoCookies) > 0 {
			b.page.MustSetCookies(protoCookies...)
		}
	}

	return NULL
}

// Refresh - refresh page
func (b *RodBrowser) Refresh(args ...Object) Object {
	b.page.MustReload()
	return b
}

// Back - go back
func (b *RodBrowser) Back(args ...Object) Object {
	b.page.MustNavigateBack()
	return b
}

// Forward - go forward
func (b *RodBrowser) Forward(args ...Object) Object {
	b.page.MustNavigateForward()
	return b
}

// Screenshot - take screenshot
func (b *RodBrowser) Screenshot(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] screenshot requires path")
	}

	path, ok := args[0].(*String)
	if !ok {
		return NewString("[error] screenshot requires string path")
	}

	b.page.MustScreenshot(path.Value)
	return NULL
}

// HTML - get page HTML
func (b *RodBrowser) HTML(args ...Object) Object {
	return &String{Value: b.page.MustHTML()}
}

// Text - get page text
func (b *RodBrowser) Text(args ...Object) Object {
	return &String{Value: b.page.MustElement("body").MustText()}
}

// SetViewport - set viewport size
func (b *RodBrowser) SetViewport(args ...Object) Object {
	if len(args) < 2 {
		return NewString("[error] setViewport requires width and height")
	}

	width := getInt(args[0])
	height := getInt(args[1])

	b.page.MustSetViewport(int(width), int(height), 1, false)
	return b
}

// SetUserAgent - set user agent (using JavaScript injection)
func (b *RodBrowser) SetUserAgent(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] setUserAgent requires ua string")
	}

	ua, ok := args[0].(*String)
	if !ok {
		return NewString("[error] setUserAgent requires string")
	}

	// Override navigator.userAgent using JavaScript
	b.page.MustEval(`Object.defineProperty(navigator, 'userAgent', {value: '` + ua.Value + `', configurable: true});`)
	return b
}

// Inject - inject JavaScript
func (b *RodBrowser) Inject(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] inject requires js string")
	}

	js, ok := args[0].(*String)
	if !ok {
		return NewString("[error] inject requires string js")
	}

	b.page.MustEval(js.Value)
	return b
}

// parseJSResult - parse JavaScript result to Xxlang Object
func parseJSResult(jsonData interface{}) Object {
	var value interface{}
	switch v := jsonData.(type) {
	case []byte:
		json.Unmarshal(v, &value)
	case string:
		json.Unmarshal([]byte(v), &value)
	default:
		value = jsonData
	}
	return toXxValue(value)
}

// toXxValue - convert Go value to Xxlang Object
func toXxValue(v interface{}) Object {
	switch val := v.(type) {
	case string:
		return &String{Value: val}
	case float64:
		return &Float{Value: val}
	case bool:
		return &Bool{Value: val}
	case nil:
		return NULL
	case map[string]interface{}:
		pairs := make(map[HashKey]MapPair)
		for k, v := range val {
			key := NewString(k)
			valObj := toXxValue(v)
			pairs[key.HashKey()] = MapPair{Key: key, Value: valObj}
		}
		return &Map{Pairs: pairs}
	case []interface{}:
		arr := make([]Object, len(val))
		for i, v := range val {
			arr[i] = toXxValue(v)
		}
		return &Array{Elements: arr}
	default:
		return &String{Value: toStringInterface(v)}
	}
}

// toGoValue - convert Xxlang Object to Go value
func toGoValue(obj Object) interface{} {
	switch v := obj.(type) {
	case *String:
		return v.Value
	case *Float:
		return v.Value
	case *Bool:
		return v.Value
	case *Null:
		return nil
	case *Map:
		result := make(map[string]interface{})
		for _, val := range v.Pairs {
			result[val.Key.Inspect()] = toGoValue(val.Value)
		}
		return result
	case *Array:
		result := make([]interface{}, len(v.Elements))
		for i, val := range v.Elements {
			result[i] = toGoValue(val)
		}
		return result
	default:
		return v.Inspect()
	}
}

// Helpers
func getString(obj Object) string {
	if s, ok := obj.(*String); ok {
		return s.Value
	}
	return ""
}

// getStringFromMap gets a string value from a Map by key
func getStringFromMap(m *Map, key string) string {
	if pair, ok := m.Pairs[NewString(key).HashKey()]; ok {
		if s, ok := pair.Value.(*String); ok {
			return s.Value
		}
	}
	return ""
}

func getInt(obj Object) int64 {
	switch v := obj.(type) {
	case *Float:
		return int64(v.Value)
	case *Int:
		return v.Value
	default:
		return 0
	}
}

// toStringInterface converts an interface{} to string
func toStringInterface(v interface{}) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// Check if binary is Chrome/Chromium
func isChromePath(path string) bool {
	_, err := exec.LookPath(path)
	return err == nil
}

// Type returns the object type
func (b *RodBrowser) Type() ObjectType { return RodBrowserType }

// TypeTag returns the type tag for fast type checking
func (b *RodBrowser) TypeTag() TypeTag { return TagRodBrowser }

// Inspect returns the string representation
func (b *RodBrowser) Inspect() string {
	return fmt.Sprintf("<Browser headless=%v>", b.headless)
}

// ToBool returns the boolean value
func (b *RodBrowser) ToBool() *Bool { return TRUE }

// HashKey returns the hash key
func (b *RodBrowser) HashKey() HashKey {
	return HashKey{
		Type:  RodBrowserType,
		Value: uint64(uintptr(unsafe.Pointer(b.browser))),
	}
}
