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
	UserAgent string
	Proxy     string
	Timeout   time.Duration
}

func New(opts *Options) (*Browser, error) {
	if opts == nil {
		opts = &Options{}
	}

	browserOpts := &browser.Options{
		UserAgent: opts.UserAgent,
		Proxy:     opts.Proxy,
	}
	if opts.Timeout > 0 {
		browserOpts.Timeout = int(opts.Timeout.Seconds())
	}

	return &Browser{
		browser: browser.New(browserOpts),
	}, nil
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

func Launch(opts *Options) (*Browser, error) {
	return New(opts)
}
