// pkg/objects/browser_element.go
// RodHTMLElement object for web scraping with Rod
package objects

import (
	"fmt"
	"time"
	"unsafe"
	"github.com/go-rod/rod"
)

// RodHTMLElement - HTML Element object for Rod-based browser
type RodHTMLElement struct {
	Element *rod.Element
	Page    *rod.Page
}

// GetText - get element text content
func (el *RodHTMLElement) GetText(args ...Object) Object {
	text, err := el.Element.Text()
	if err != nil {
		return &String{Value: ""}
	}
	return &String{Value: text}
}

// GetAttr - get element attribute
func (el *RodHTMLElement) GetAttr(args ...Object) Object {
	if len(args) < 1 {
		return NULL
	}

	name, ok := args[0].(*String)
	if !ok {
		return NULL
	}

	value, _ := el.Element.Attribute(name.Value)
	if value == nil {
		return NULL
	}
	return &String{Value: *value}
}

// GetProperty - get element property
func (el *RodHTMLElement) GetProperty(args ...Object) Object {
	if len(args) < 1 {
		return NULL
	}

	name, ok := args[0].(*String)
	if !ok {
		return NULL
	}

	result, err := el.Element.Eval(`function() { return this[name] }`, name.Value)
	if err != nil {
		return NULL
	}

	return parseJSResult(result.Value)
}

// GetInnerHTML - get inner HTML
func (el *RodHTMLElement) GetInnerHTML(args ...Object) Object {
	html, err := el.Element.HTML()
	if err != nil {
		return &String{Value: ""}
	}
	return &String{Value: html}
}

// GetOuterHTML - get outer HTML
func (el *RodHTMLElement) GetOuterHTML(args ...Object) Object {
	result, err := el.Element.Eval(`function() { return this.outerHTML; }`)
	if err != nil {
		return &String{Value: ""}
	}
	return parseJSResult(result.Value)
}

// GetTagName - get tag name
func (el *RodHTMLElement) GetTagName(args ...Object) Object {
	result, err := el.Element.Eval(`function() { return this.tagName; }`)
	if err != nil {
		return &String{Value: ""}
	}
	return parseJSResult(result.Value)
}

// Click - click element
func (el *RodHTMLElement) Click(args ...Object) Object {
	el.Element.MustClick()
	return el
}

// Input - input text
func (el *RodHTMLElement) Input(args ...Object) Object {
	if len(args) < 1 {
		return el
	}

	value, ok := args[0].(*String)
	if !ok {
		return el
	}

	el.Element.MustInput(value.Value)
	return el
}

// Select - select option (for select elements)
func (el *RodHTMLElement) Select(args ...Object) Object {
	if len(args) < 1 {
		return el
	}

	value, ok := args[0].(*String)
	if !ok {
		return el
	}

	// Use JavaScript to select the option
	el.Element.MustEval(`function() { this.value = '` + value.Value + `'; this.dispatchEvent(new Event('change')); }`)
	return el
}

// Check - check checkbox/radio
func (el *RodHTMLElement) Check(args ...Object) Object {
	el.Element.MustEval(`function() { this.checked = true; }`)
	return el
}

// Uncheck - uncheck checkbox/radio
func (el *RodHTMLElement) Uncheck(args ...Object) Object {
	el.Element.MustEval(`function() { this.checked = false; }`)
	return el
}

// Focus - focus element
func (el *RodHTMLElement) Focus(args ...Object) Object {
	el.Element.MustFocus()
	return el
}

// Blur - blur element
func (el *RodHTMLElement) Blur(args ...Object) Object {
	el.Element.MustBlur()
	return el
}

// Find - find child element
func (el *RodHTMLElement) Find(args ...Object) Object {
	if len(args) < 1 {
		return NULL
	}

	selector, ok := args[0].(*String)
	if !ok {
		return NULL
	}

	child, err := el.Element.Element(selector.Value)
	if err != nil {
		return NULL
	}

	return &RodHTMLElement{Element: child, Page: el.Page}
}

// FindAll - find all child elements
func (el *RodHTMLElement) FindAll(args ...Object) Object {
	if len(args) < 1 {
		return &Array{Elements: []Object{}}
	}

	selector, ok := args[0].(*String)
	if !ok {
		return &Array{Elements: []Object{}}
	}

	children, err := el.Element.Elements(selector.Value)
	if err != nil {
		return &Array{Elements: []Object{}}
	}

	values := make([]Object, len(children))
	for i, child := range children {
		values[i] = &RodHTMLElement{Element: child, Page: el.Page}
	}

	return &Array{Elements: values}
}

// Screenshot - screenshot of element
func (el *RodHTMLElement) Screenshot(args ...Object) Object {
	if len(args) < 1 {
		return NewString("[error] screenshot requires path")
	}

	path, ok := args[0].(*String)
	if !ok {
		return NewString("[error] screenshot requires string path")
	}

	el.Element.MustScreenshot(path.Value)
	return NULL
}

// GetBoundingClientRect - get element position and size
func (el *RodHTMLElement) GetBoundingClientRect(args ...Object) Object {
	result, err := el.Element.Eval(`function() {
		var rect = this.getBoundingClientRect();
		return {
			x: rect.x,
			y: rect.y,
			width: rect.width,
			height: rect.height,
			top: rect.top,
			right: rect.right,
			bottom: rect.bottom,
			left: rect.left
		};
	}`)
	if err != nil {
		return NULL
	}

	return parseJSResult(result.Value)
}

// IsVisible - check if element is visible
func (el *RodHTMLElement) IsVisible(args ...Object) Object {
	result, err := el.Element.Eval(`function() {
		var style = window.getComputedStyle(this);
		return style.display !== 'none' &&
		       style.visibility !== 'hidden' &&
		       style.opacity !== '0' &&
		       this.offsetWidth > 0 &&
		       this.offsetHeight > 0;
	}`)
	if err != nil {
		return &Bool{Value: false}
	}

	if b, ok := parseJSResult(result.Value).(*Bool); ok {
		return b
	}
	return &Bool{Value: false}
}

// IsEnabled - check if element is enabled
func (el *RodHTMLElement) IsEnabled(args ...Object) Object {
	result, err := el.Element.Eval(`function() { return !this.disabled; }`)
	if err != nil {
		return &Bool{Value: false}
	}

	if b, ok := parseJSResult(result.Value).(*Bool); ok {
		return b
	}
	return &Bool{Value: false}
}

// GetValue - get form element value
func (el *RodHTMLElement) GetValue(args ...Object) Object {
	result, err := el.Element.Eval(`function() { return this.value; }`)
	if err != nil {
		return &String{Value: ""}
	}

	return parseJSResult(result.Value)
}

// SetValue - set form element value
func (el *RodHTMLElement) SetValue(args ...Object) Object {
	if len(args) < 1 {
		return el
	}

	value, ok := args[0].(*String)
	if !ok {
		return el
	}

	el.Element.MustInput(value.Value)
	return el
}

// WaitFor - wait for element state
func (el *RodHTMLElement) WaitFor(args ...Object) Object {
	if len(args) < 1 {
		return el
	}

	state, ok := args[0].(*String)
	if !ok {
		return el
	}

	switch state.Value {
	case "visible":
		el.Element.MustWaitVisible()
	case "hidden":
		el.Element.MustWaitInvisible()
	case "enabled":
		el.Element.MustWaitEnabled()
	case "disabled":
		// Wait until disabled
		for i := 0; i < 50; i++ {
			if el.Element.MustDisabled() {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	return el
}

// Type - type text with delay (simulate real typing)
func (el *RodHTMLElement) TypeText(args ...Object) Object {
	if len(args) < 1 {
		return el
	}

	text, ok := args[0].(*String)
	if !ok {
		return el
	}

	// Use Input with type simulation
	el.Element.MustInput(text.Value)
	return el
}

// Hover - hover over element
func (el *RodHTMLElement) Hover(args ...Object) Object {
	el.Element.MustHover()
	return el
}

// Drag - drag element to target
func (el *RodHTMLElement) Drag(args ...Object) Object {
	if len(args) < 2 {
		return el
	}

	x := getInt(args[0])
	y := getInt(args[1])

	// Use JavaScript to trigger drag and drop
	el.Element.MustEval(`function(x, y) {
		var event = new MouseEvent('dragstart', {
			bubbles: true,
			cancelable: true,
			view: window,
			screenX: x,
			screenY: y
		});
		this.dispatchEvent(event);
	}`, x, y)
	return el
}

// Press - press key
func (el *RodHTMLElement) Press(args ...Object) Object {
	if len(args) < 1 {
		return el
	}

	key, ok := args[0].(*String)
	if !ok {
		return el
	}

	// Focus the element first
	el.Element.MustFocus()
	// Use JavaScript to simulate key press
	el.Element.MustEval(`function(key) {
		var event = new KeyboardEvent('keydown', {key: key, bubbles: true});
		this.dispatchEvent(event);
		var event2 = new KeyboardEvent('keyup', {key: key, bubbles: true});
		this.dispatchEvent(event2);
	}`, key.Value)
	return el
}

// SelectAll - select all text in input
func (el *RodHTMLElement) SelectAll(args ...Object) Object {
	el.Element.MustEval(`function() { this.select(); }`)
	return el
}

// Type returns the object type
func (el *RodHTMLElement) Type() ObjectType { return RodHTMLElementType }

// TypeTag returns the type tag for fast type checking
func (el *RodHTMLElement) TypeTag() TypeTag { return TagRodHTMLElement }

// Inspect returns the string representation
func (el *RodHTMLElement) Inspect() string {
	result, err := el.Element.Eval(`function() { return this.tagName; }`)
	if err != nil {
		return "<HTMLElement error>"
	}
	// Use Description field which contains string representation
	if result.Description != "" {
		return fmt.Sprintf("<HTMLElement %s>", result.Description)
	}
	return "<HTMLElement>"
}

// ToBool returns the boolean value
func (el *RodHTMLElement) ToBool() *Bool { return TRUE }

// HashKey returns the hash key
func (el *RodHTMLElement) HashKey() HashKey {
	return HashKey{
		Type:  RodHTMLElementType,
		Value: uint64(uintptr(unsafe.Pointer(el.Element))),
	}
}
