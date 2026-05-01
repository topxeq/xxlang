package browser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
	"github.com/topxeq/xxlang/pkg/hlbr/httpclient"
	"github.com/topxeq/xxlang/pkg/hlbr/jsengine"
	"github.com/topxeq/xxlang/pkg/hlbr/renderer"
)

type Browser struct {
	client              *httpclient.Client
	doc                 *dom.Document
	vm                  *jsengine.VM
	currentURL          string
	history             []string
	debug               bool
	noScripts           bool // if true, skip script execution during navigate
	skipExternalScripts bool // if true, skip loading external scripts
}

type Options struct {
	UserAgent           string
	Proxy               string
	Timeout             int
	Debug               bool
	NoScripts           bool // if true, skip script execution during navigate
	SkipExternalScripts bool // if true, skip loading external scripts
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
		client:              httpclient.NewClient(httpOpts),
		history:             make([]string, 0),
		debug:               opts.Debug,
		noScripts:           opts.NoScripts,
		skipExternalScripts: opts.SkipExternalScripts,
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

	// Handle about:blank - create empty document without HTTP request
	if rawURL == "about:blank" || rawURL == "" {
		b.currentURL = "about:blank"
		b.doc = htmlparser.Parse("<html><head></head><body></body></html>")
		b.vm = jsengine.NewVM(b.doc)
		b.vm.SetTimeoutMs(30_000) // 30 seconds default (reduced from 120s)
		b.vm.SetMaxCallDepth(1000)
		b.vm.SetMaxAllocs(100_000) // 100K object allocations max
		b.debugLog("Navigate completed (about:blank)")
		return nil
	}

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
	b.vm.SetTimeoutMs(60_000)  // 60 seconds for heavy pages (reduced from 180s)
	b.vm.SetMaxSteps(100_000_000) // 100M steps for heavy pages (reduced from 500M)
	b.vm.SetMaxCallDepth(1000)
	b.vm.SetMaxAllocs(500_000) // 500K object allocations max

	if !b.noScripts {
		b.debugLog("Executing scripts...")
		b.executeScripts()
	} else {
		b.debugLog("Skipping scripts (noScripts=true)")
	}
	b.debugLog("Navigate completed")

	return nil
}

func (b *Browser) executeScripts() {
	if b.doc == nil || b.vm == nil {
		return
	}

	scripts := dom.GetElementsByTagName(b.doc.Root, "script")
	vueMountPatched := false
	for _, script := range scripts {
		src := script.GetAttribute("src")
		if src != "" {
			if !b.skipExternalScripts {
				b.loadExternalScript(src)
			} else {
				b.debugLog("Skipping external script: %s", src)
			}
		} else {
			code := script.TextContent()
			if code != "" {
				b.vm.Run(code)
			}
		}

		// After each script, check if Vue is now defined and patch $mount
		if !vueMountPatched {
			result, _ := b.vm.Run("typeof Vue !== 'undefined' && typeof Vue.compile === 'function'")
			if result != nil && result.Bool {
				b.vm.Run(`(function(){function simpleCompile(tpl){if(!tpl)return '_c("div")';var m;m=tpl.match(/^<([A-Z][a-zA-Z0-9]*)\s*\/?\s*>$/);if(m)return '_c("'+m[1]+'")';m=tpl.match(/^<([a-z][a-z0-9]*)\s*\/\s*>$/);if(m)return '_c("'+m[1]+'")';var closeTag=tpl.match(/<\/([a-z][a-z0-9]*)>$/);if(closeTag){var openTag=tpl.match(/^<([a-z][a-z0-9]*)>/);if(openTag&&openTag[1]===closeTag[1]){var tag=openTag[1];var inner=tpl.substring(openTag[0].length,tpl.length-closeTag[0].length);var children=[];if(inner){var textParts=inner.split(/({{[^}]+}})/);for(var i=0;i<textParts.length;i++){var part=textParts[i];if(!part)continue;var exprMatch=part.match(/^{{(.+)}}$/);if(exprMatch){children.push('_v(_s('+exprMatch[1].trim()+'))')}else{children.push('_v("'+part.replace(/"/g,'\\"')+'")')}}}return '_c("'+tag+'",'+(children.length>0?'['+children.join(',')+']':'')+')'}}return '_c("div",_v('+JSON.stringify(tpl)+'))'}Vue.compile=function(template){template=template.trim();var staticRenderFns=[];var code=simpleCompile(template);var render=new Function('with(this){return '+code+'}');return{render:render,staticRenderFns:staticRenderFns}};function fixDataProxy(vm){if(!vm._data||!vm.$options)return;var keys=Object.keys(vm._data);for(var i=0;i<keys.length;i++){var key=keys[i];if(key.charCodeAt(0)!==36&&key.charCodeAt(0)!==95&&!vm.hasOwnProperty(key)){(function(k){Object.defineProperty(vm,k,{get:function(){return this._data[k]},set:function(v){this._data[k]=v},enumerable:true,configurable:true})})(key)}}}var origInit=Vue.prototype._init;Vue.prototype._init=function(options){origInit.call(this,options);fixDataProxy(this)};var cs=Vue.prototype.$mount;Vue.prototype.$mount=function(el,hydrating){var vm=this,n=vm.$options;fixDataProxy(vm);if(!n.render){var r=n.template;if(r){if(typeof r==="string"){if(r.charAt(0)==="#"){var t=document.querySelector(r);if(t){r=t.innerHTML}}if(r){try{var compiled=Vue.compile(r);n.render=compiled.render;n.staticRenderFns=compiled.staticRenderFns}catch(e){}}}}else if(r&&r.nodeType){r=r.innerHTML}}return cs.call(this,el,hydrating)}})()`)
				vueMountPatched = true
				b.debugLog("Vue $mount polyfill applied")
			}
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
		b.debugLog("Failed to load external script: %v", err)
		return
	}

	// Limit external script size to 1MB to prevent memory exhaustion
	if len(resp.Body) > 1_000_000 {
		b.debugLog("External script too large (%d bytes), skipping", len(resp.Body))
		return
	}

	if b.vm != nil {
		// Reset steps before executing external script to prevent cumulative timeout
		b.vm.ResetSteps()
		b.vm.Run(resp.Body)
	}
}

func (b *Browser) Evaluate(code string) (any, error) {
	if b.vm == nil {
		return nil, nil
	}
	// Save and reset step counter for per-evaluate limit
	prevSteps := b.vm.GetStepCount()
	prevMax := b.vm.GetMaxSteps()
	b.vm.SetMaxSteps(10_000_000) // 10M steps per evaluate call
	b.vm.SetStepCount(0)
	val, err := b.vm.Run(code)
	b.vm.SetStepCount(prevSteps)
	b.vm.SetMaxSteps(prevMax)
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

// SetJsDebug enables or disables JS debug mode.
func (b *Browser) SetJsDebug(jsDebug bool) {
	b.debugLog("SetJsDebug called: %v", jsDebug)
	if b.vm != nil {
		b.vm.SetDebug(jsDebug)
	}
}

// WaitStable waits for the page to become stable by draining pending JavaScript
// timers (setTimeout/setInterval) and allowing Vue/SPA rendering to complete.
// timeoutMs is the maximum time to wait, stableForMs is how long the page must
// remain stable (no new timers) before returning.
func (b *Browser) WaitStable(timeoutMs, stableForMs int) error {
	b.debugLog("WaitStable called: timeoutMs=%d, stableForMs=%d", timeoutMs, stableForMs)
	if b.vm == nil {
		return nil
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	stableDeadline := time.Time{}
	maxRounds := 50 // Safety limit to prevent infinite loops

	for round := 0; round < maxRounds; round++ {
		if time.Now().After(deadline) {
			b.debugLog("WaitStable: timeout reached after %d rounds", round)
			break
		}

		// Brief sleep to allow timers to become due
		time.Sleep(10 * time.Millisecond)

		// Process all pending timers synchronously
		hadTimers := b.vm.HasPendingTimers()
		if hadTimers {
			b.debugLog("WaitStable: processing pending timers (round %d)", round)
			b.vm.RunTimers()
		}

		// Check if new timers were created (e.g., by Vue's nextTick)
		if b.vm.HasPendingTimers() {
			// More timers were scheduled, keep processing
			stableDeadline = time.Time{}
			continue
		}

		// No pending timers - check if we've been stable for long enough
		if stableDeadline.IsZero() {
			stableDeadline = time.Now().Add(time.Duration(stableForMs) * time.Millisecond)
		}

		if time.Now().After(stableDeadline) {
			b.debugLog("WaitStable: page stable after %d rounds", round)
			break
		}
	}

	return nil
}

// WaitStableDefault waits for the page to become stable with default timeouts.
func (b *Browser) WaitStableDefault() error {
	return b.WaitStable(5000, 100)
}

// Abort cancels any running JavaScript execution.
// This can be called from another goroutine to stop a long-running script.
func (b *Browser) Abort() {
	if b.vm != nil {
		b.vm.Abort()
	}
}

// IsAborted returns true if the abort flag has been set.
func (b *Browser) IsAborted() bool {
	if b.vm != nil {
		return b.vm.IsAborted()
	}
	return false
}

// FormField represents a form field found in the page.
type FormField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Placeholder string `json:"placeholder"`
	Label       string `json:"label"`
	ID          string `json:"id"`
	Required    bool   `json:"required"`
}

// AnalyzeVueTemplates analyzes JavaScript source code to find Vue.js templates
// and extract form fields. This is useful for SPA pages that render forms dynamically.
func (b *Browser) AnalyzeVueTemplates() []FormField {
	var fields []FormField

	// Collect all JavaScript sources
	var jsSources []string

	// Add inline scripts
	if b.doc != nil {
		scripts := dom.GetElementsByTagName(b.doc.Root, "script")
		for _, script := range scripts {
			src := script.GetAttribute("src")
			if src == "" {
				code := script.TextContent()
				if code != "" {
					jsSources = append(jsSources, code)
				}
			} else {
				// Download external script
				if !strings.HasPrefix(src, "http") {
					baseURL, _ := url.Parse(b.currentURL)
					if strings.HasPrefix(src, "/") {
						src = baseURL.Scheme + "://" + baseURL.Host + src
					} else {
						src = baseURL.Scheme + "://" + baseURL.Host + "/" + src
					}
				}
				resp, err := b.client.Get(src)
				if err == nil && len(resp.Body) <= 2_000_000 { // Limit to 2MB
					jsSources = append(jsSources, resp.Body)
				}
			}
		}
	}

	// Analyze each JavaScript source for Vue templates
	for _, js := range jsSources {
		fields = append(fields, b.extractFormFieldsFromJS(js)...)
	}

	return fields
}

// extractFormFieldsFromJS extracts form fields from JavaScript source code.
func (b *Browser) extractFormFieldsFromJS(js string) []FormField {
	var fields []FormField

	// Look for el-form-item patterns
	// Pattern: <el-form-item prop="fieldName">
	formItemPattern := `el-form-item[^>]*prop=["']([^"']+)["']`
	formItemRegex := regexp.MustCompile(formItemPattern)
	matches := formItemRegex.FindAllStringSubmatch(js, -1)

	for _, match := range matches {
		if len(match) > 1 {
			field := FormField{
				Name: match[1],
				Type: "text",
			}

			// Try to find the associated el-input
			inputPattern := `v-model=["']ruleForm\.` + regexp.QuoteMeta(match[1]) + `["'][^>]*>`
			inputRegex := regexp.MustCompile(inputPattern)
			inputMatch := inputRegex.FindString(js)

			if inputMatch != "" {
				// Extract type
				typePattern := `type=["']([^"']+)["']`
				typeRegex := regexp.MustCompile(typePattern)
				typeMatch := typeRegex.FindStringSubmatch(inputMatch)
				if len(typeMatch) > 1 {
					field.Type = typeMatch[1]
				}

				// Extract placeholder
				placeholderPattern := `:placeholder=["']\$t\(['"]([^'"]+)['"]\)["']|placeholder=["']([^"']+)["']`
				placeholderRegex := regexp.MustCompile(placeholderPattern)
				placeholderMatch := placeholderRegex.FindStringSubmatch(inputMatch)
				if len(placeholderMatch) > 1 {
					if placeholderMatch[1] != "" {
						field.Placeholder = placeholderMatch[1]
					} else if placeholderMatch[2] != "" {
						field.Placeholder = placeholderMatch[2]
					}
				}
			}

			fields = append(fields, field)
		}
	}

	// Also look for v-model patterns directly
	vModelPattern := `v-model=["']([^"']+)["']`
	vModelRegex := regexp.MustCompile(vModelPattern)
	vModelMatches := vModelRegex.FindAllStringSubmatch(js, -1)

	existingFields := make(map[string]bool)
	for _, f := range fields {
		existingFields[f.Name] = true
	}

	for _, match := range vModelMatches {
		if len(match) > 1 {
			fieldName := match[1]
			// Skip if already found or if it's a complex expression
			if existingFields[fieldName] || strings.Contains(fieldName, ".") {
				continue
			}
			// This is a simple v-model, might be a form field
			fields = append(fields, FormField{
				Name: fieldName,
				Type: "text",
			})
			existingFields[fieldName] = true
		}
	}

	return fields
}

func jsValueToGo(v *jsengine.Value) any {
	return jsValueToGoWithVisited(v, make(map[*jsengine.Value]bool))
}

func jsValueToGoWithVisited(v *jsengine.Value, visited map[*jsengine.Value]bool) any {
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
		if visited[v] {
			return "[circular]"
		}
		visited[v] = true
		if v.Arr != nil {
			arr := make([]any, len(v.Arr))
			for i, a := range v.Arr {
				arr[i] = jsValueToGoWithVisited(a, visited)
			}
			return arr
		}
		if v.Obj != nil {
			obj := make(map[string]any)
			for k, val := range v.Obj {
				obj[k] = jsValueToGoWithVisited(val, visited)
			}
			return obj
		}
		return nil
	case "function", "native":
		return "[function]"
	}
	return nil
}
