package httpclient

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client    *http.Client
	userAgent string
	cookies   []*http.Cookie
	headers   map[string]string
	proxyURL  string
	lastURL   string
	lastResp  *Response
	debug     bool
}

type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       string
	URL        string
}

type Options struct {
	UserAgent string
	Proxy     string
	Timeout   int
	Debug     bool
}

func NewClient(opts *Options) *Client {
	if opts == nil {
		opts = &Options{}
	}

	ua := opts.UserAgent
	if ua == "" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) xxhlbr/0.1.0"
	}

	c := &Client{
		userAgent: ua,
		proxyURL:  opts.Proxy,
		headers:   make(map[string]string),
		debug:     opts.Debug,
	}

	transport := &http.Transport{}
	if opts.Proxy != "" {
		proxyURL, err := url.Parse(opts.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	// Set timeout (default 30 seconds if not specified)
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	c.client = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return c
}

// SetDebug enables or disables debug mode.
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) Get(url string) (*Response, error) {
	return c.Request("GET", url, "", nil)
}

func (c *Client) Post(url string, contentType string, body string) (*Response, error) {
	return c.Request("POST", url, body, map[string]string{
		"Content-Type": contentType,
	})
}

func (c *Client) PostForm(targetURL string, data map[string]string) (*Response, error) {
	form := make(url.Values)
	for k, v := range data {
		form.Set(k, v)
	}
	return c.Request("POST", targetURL, form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}

func (c *Client) Request(method, url string, body string, extraHeaders map[string]string) (*Response, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.lastURL = resp.Request.URL.String()
	c.lastResp = &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       string(bodyBytes),
		URL:        c.lastURL,
	}

	for _, cookie := range resp.Cookies() {
		c.setCookie(cookie)
	}

	return c.lastResp, nil
}

func (c *Client) setCookie(cookie *http.Cookie) {
	for i, existing := range c.cookies {
		if existing.Name == cookie.Name && existing.Domain == cookie.Domain {
			c.cookies[i] = cookie
			return
		}
	}
	c.cookies = append(c.cookies, cookie)
}

func (c *Client) Cookies() []*http.Cookie {
	return c.cookies
}

func (c *Client) SetCookie(name, value, domain string) {
	c.cookies = append(c.cookies, &http.Cookie{
		Name:   name,
		Value:  value,
		Domain: domain,
	})
}

func (c *Client) ClearCookies() {
	c.cookies = nil
}

func (c *Client) SetUserAgent(ua string) {
	c.userAgent = ua
}

func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

func (c *Client) LastURL() string {
	return c.lastURL
}

func (c *Client) LastResponse() *Response {
	return c.lastResp
}

func (c *Client) FollowRedirects(max int) error {
	c.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= max {
			return http.ErrUseLastResponse
		}
		return nil
	}
	return nil
}
