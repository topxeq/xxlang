package hlbr

import (
	"net/http"
	"time"

	"github.com/topxeq/xxlang/pkg/hlbr/browser"
	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/httpclient"
)

type Browser struct {
	browser *browser.Browser
}

type Options struct {
	UserAgent           string
	Proxy               string
	Timeout             time.Duration
	Debug               bool
	JsDebug             bool // Separate JS debug output
	SkipExternalScripts bool // Skip loading external scripts
	SkipScripts         bool // Skip all script execution
}

func New(opts *Options) (*Browser, error) {
	if opts == nil {
		opts = &Options{}
	}

	browserOpts := &browser.Options{
		UserAgent:           opts.UserAgent,
		Proxy:               opts.Proxy,
		Debug:               opts.Debug,
		NoScripts:           opts.SkipScripts,
		SkipExternalScripts: opts.SkipExternalScripts,
	}
	if opts.Timeout > 0 {
		browserOpts.Timeout = int(opts.Timeout.Seconds())
	}

	b := &Browser{
		browser: browser.New(browserOpts),
	}

	// Set JS debug mode if specified
	if opts.JsDebug {
		b.browser.SetJsDebug(true)
	}

	return b, nil
}

// SetDebug enables or disables debug mode.
func (b *Browser) SetDebug(debug bool) {
	b.browser.SetDebug(debug)
}

// SetJsDebug enables or disables JS debug mode (shows JS code being executed).
func (b *Browser) SetJsDebug(jsDebug bool) {
	b.browser.SetJsDebug(jsDebug)
}

func (b *Browser) Navigate(url string) error {
	return b.browser.Navigate(url)
}

func (b *Browser) GetTitle() string {
	return b.browser.GetTitle()
}

func (b *Browser) GetHTML() string {
	return b.browser.GetHTML()
}

func (b *Browser) GetText() string {
	return b.browser.GetText()
}

func (b *Browser) ScreenshotText(width int) string {
	return b.browser.ScreenshotText(width)
}

func (b *Browser) ScreenshotTextToFile(path string, width int) error {
	return b.browser.ScreenshotTextToFile(path, width)
}

func (b *Browser) GetURL() string {
	return b.browser.GetURL()
}

func (b *Browser) QuerySelector(selector string) *dom.Node {
	return b.browser.QuerySelector(selector)
}

func (b *Browser) QuerySelectorAll(selector string) []*dom.Node {
	return b.browser.QuerySelectorAll(selector)
}

func (b *Browser) Evaluate(code string) (any, error) {
	return b.browser.Evaluate(code)
}

func (b *Browser) Document() *dom.Document {
	return b.browser.Document()
}

func (b *Browser) Client() *httpclient.Client {
	return b.browser.Client()
}

func (b *Browser) SetUserAgent(ua string) {
	b.browser.Client().SetUserAgent(ua)
}

func (b *Browser) SetHeader(key, value string) {
	b.browser.Client().SetHeader(key, value)
}

func (b *Browser) GetCookies() []*http.Cookie {
	return b.browser.Client().Cookies()
}

func (b *Browser) History() []string {
	return b.browser.History()
}

func (b *Browser) Back() error {
	return b.browser.Back()
}

// GetLocalStorage returns the localStorage data
func (b *Browser) GetLocalStorage() map[string]string {
	return b.browser.GetLocalStorage()
}

// GetSessionStorage returns the sessionStorage data
func (b *Browser) GetSessionStorage() map[string]string {
	return b.browser.GetSessionStorage()
}

// SetLocalStorageItem sets a localStorage item
func (b *Browser) SetLocalStorageItem(key, value string) {
	b.browser.SetLocalStorageItem(key, value)
}

// SetSessionStorageItem sets a sessionStorage item
func (b *Browser) SetSessionStorageItem(key, value string) {
	b.browser.SetSessionStorageItem(key, value)
}

// GetConsoleOutput returns the console.log output
func (b *Browser) GetConsoleOutput() []string {
	return b.browser.GetConsoleOutput()
}

// WaitStable waits for the page to become stable (no pending timers or JavaScript execution).
func (b *Browser) WaitStable(timeoutMs, stableForMs int) error {
	return b.browser.WaitStable(timeoutMs, stableForMs)
}

// WaitStableDefault waits for the page to become stable with default timeouts.
func (b *Browser) WaitStableDefault() error {
	return b.browser.WaitStableDefault()
}

// Abort cancels any running JavaScript execution.
// This can be called from another goroutine to stop a long-running script.
func (b *Browser) Abort() {
	b.browser.Abort()
}

// IsAborted returns true if the abort flag has been set.
func (b *Browser) IsAborted() bool {
	return b.browser.IsAborted()
}

// AnalyzeVueTemplates analyzes JavaScript source code to find Vue.js templates
// and extract form fields. This is useful for SPA pages that render forms dynamically.
func (b *Browser) AnalyzeVueTemplates() []browser.FormField {
	return b.browser.AnalyzeVueTemplates()
}

// VM returns the JavaScript VM
func (b *Browser) VM() interface{} {
	return b.browser.VM()
}

// Fill sets the value of the first element matching the CSS selector and
// dispatches an "input" event so that Vue v-model bindings are updated.
func (b *Browser) Fill(selector, value string) error {
	return b.browser.Fill(selector, value)
}

// Click clicks the first element matching the CSS selector.
func (b *Browser) Click(selector string) error {
	return b.browser.Click(selector)
}

func Launch(opts *Options) (*Browser, error) {
	return New(opts)
}
