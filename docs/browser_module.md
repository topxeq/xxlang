# Browser Module for Xxlang

The browser module provides web scraping and browser automation capabilities using the Rod library (a Chrome DevTools Protocol driver).

## Installation

The browser module is included in the standard library. No additional installation is required.

## Usage

### Import the module

```xxl
import "browser"
```

### Create a browser instance

```xxl
// Create browser with options
var opts = {"headless": true}
var br = browser.open(opts)
```

**Options:**
- `headless` (boolean): Run browser in headless mode (default: true)
- `chromePath` (string): Path to Chrome/Chromium executable
- `autoDownload` (boolean): Auto-download Chromium if not found (default: true)
- `proxy` (string): Proxy server URL
- `userAgent` (string): Custom user agent string

### Navigation

```xxl
// Navigate to URL
br.get("https://example.com")

// Wait for page load
br.waitLoad()

// Wait for page to become stable (no network activity)
br.waitStable()

// Set browser window to fullscreen
br.fullscreen()

// Refresh page
br.refresh()

// Navigate back/forward
br.back()
br.forward()
```

### Finding Elements

```xxl
// Find single element
var el = br.find("#myId")
var el = br.find(".myClass")
var el = br.find("input[name='username']")

// Find multiple elements
var elements = br.findAll("a")
var elements = br.findAll(".item")

// Check if element exists
var exists = br.exists(".myClass")
```

### Element Operations

```xxl
// Get element text
var text = el.getText()

// Get element attribute
var href = el.getAttr("href")
var src = el.getAttr("src")

// Get inner/outer HTML
var inner = el.getInnerHTML()
var outer = el.getOuterHTML()

// Get tag name
var tag = el.getTagName()

// Check visibility/enabled state
var visible = el.isVisible()
var enabled = el.isEnabled()

// Get form element value
var value = el.getValue()
```

### Interaction

```xxl
// Click element
br.click("#submitBtn")
el.click()

// Input text
br.fill("#username", "john")
el.input("hello")

// Type text (simulated)
el.typeText("hello world")

// Select option
el.select("optionValue")

// Check/uncheck checkbox
el.check()
el.uncheck()

// Focus/blur
el.focus()
el.blur()

// Hover
el.hover()

// Press key
el.press("Enter")

// Select all text
el.selectAll()
```

### JavaScript Execution

```xxl
// Execute JavaScript
var result = br.eval("document.title")
var result = br.eval("window.innerWidth")
var result = br.eval("2 + 2")
```

### Storage Operations

```xxl
// localStorage
br.setLocalStorage("key", "value")
var data = br.getLocalStorage()

// sessionStorage
br.setSessionStorage("key", "value")
var data = br.getSessionStorage()
```

### Cookies

```xxl
// Get all cookies
var cookies = br.getCookies()

// Set cookies
var cookieList = [
    {"name": "session", "value": "abc123", "domain": "example.com"},
    {"name": "user", "value": "john", "domain": "example.com"}
]
br.setCookies(cookieList)

// Clear all cookies
br.clearCookies()
```

### Screenshots

```xxl
// Take screenshot of page
br.screenshot("page.png")

// Take screenshot of element
el.screenshot("element.png")
```

### Viewport

```xxl
// Set viewport size
br.setViewport(1920, 1080)
```

### Storage Persistence

```xxl
// Save storage to file
br.saveStorage("storage.json")

// Load storage from file
br.loadStorage("storage.json")
```

### Get Page Content

```xxl
// Get page HTML
var html = br.html()

// Get page text
var text = br.text()
```

### User Agent

```xxl
// Set custom user agent
br.setUserAgent("Mozilla/5.0 ...")
```

### Close Browser

```xxl
br.close()
```

## Complete Example

```xxl
import "browser"
import "strings"

// Launch browser
var opts = {"headless": true}
var br = browser.open(opts)

// Check for errors
if (typeOf(br) == "STRING" && strings.hasPrefix(br, "[error]")) {
    pln("Error:", br)
    return
}

// Navigate to page
br.get("https://example.com")

// Get page info
pln("Title:", br.eval("document.title"))
pln("HTML length:", br.html().len())

// Find and interact with elements
var h1 = br.find("h1")
if (h1 != null) {
    pln("H1 text:", h1.getText())
}

// Find all links
var links = br.findAll("a")
pln("Links found:", links.len())

// Take screenshot
br.screenshot("example.png")

// Close browser
br.close()
```

## Error Handling

All browser operations return error messages as strings when they fail. Check the return type:

```xxl
var result = br.get("https://example.com")
if (typeOf(result) == "STRING" && strings.hasPrefix(result, "[error]")) {
    pln("Error occurred:", result)
}
```

## Notes

1. **Headless mode**: Set `headless=false` to see the browser window
2. **Chrome path**: If Chrome is not found automatically, set `chromePath` to the Chrome executable location
3. **Auto-download**: By default, Chromium is auto-downloaded if not found. Set `autoDownload=false` to disable
4. **Performance**: Headless mode is faster and recommended for automation/scraping
5. **Storage paths**: Storage files should use absolute or relative paths from the current directory
