package jsengine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
)

type Value struct {
	Type        string
	Num         float64
	Str         string
	Bool        bool
	Obj         map[string]*Value
	Arr         []*Value
	Func        *Function
	Class       *Class
	Promise     *Promise
	Proxy       *Proxy
	MapData     map[string]*Value // for Map type
	SetData     map[string]bool   // for Set type
	Native      func(args []*Value) *Value
	ThisBinding *Value // Bound this context for methods
	// Property descriptors for object properties (used by Object.defineProperty)
	Descriptors map[string]*PropertyDescriptor
	IsAsync     bool // true for async functions
	// BuiltInConstructor is the name of the built-in constructor (e.g., "Array", "Object")
	BuiltInConstructor string
}

// Proxy represents a JavaScript Proxy object
type Proxy struct {
	Target    *Value
	Handlers  *Value // handler object with traps
	Env       *Environment
}

// Promise represents a JavaScript Promise
type Promise struct {
	State     string        // "pending", "fulfilled", "rejected"
	Value     *Value        // resolved value
	Rejection *Value        // rejection reason
	OnFulfill []*Function   // then callbacks
	OnReject  []*Function   // catch callbacks
	Env       *Environment
}

type Function struct {
	Params     []string
	DefaultVals map[string]Expression // default values for parameters
	RestParam  string                 // rest parameter name
	Body       []Statement
	Env        *Environment
	Bytecode   *Bytecode // compiled bytecode for faster execution
}

// Class represents a JavaScript class
type Class struct {
	Name       string
	SuperClass string
	Methods    map[string]*Function      // instance methods
	Static     map[string]*Function      // static methods
	Getters    map[string]*Function      // getters
	Setters    map[string]*Function      // setters
	Env        *Environment
}

type Environment struct {
	store map[string]*Value
	outer *Environment
}

func NewEnvironment(outer *Environment) *Environment {
	return &Environment{
		store: make(map[string]*Value),
		outer: outer,
	}
}

func (e *Environment) Get(name string) *Value {
	if v, ok := e.store[name]; ok {
		return v
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return &Value{Type: "undefined"}
}

func (e *Environment) Set(name string, val *Value) {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return
	}
	if e.outer != nil {
		e.outer.Set(name, val)
		return
	}
	e.store[name] = val
}

func (e *Environment) Define(name string, val *Value) {
	e.store[name] = val
}

type VM struct {
	env            *Environment
	doc            *dom.Document
	root           *dom.Node
	output         []string
	LocalStorage   map[string]string
	SessionStorage map[string]string
	debug          bool
	// For super() support
	currentClass *Class
	// Timer support for setTimeout/setInterval
	pendingTimers []timerTask
}

// timerTask represents a scheduled timer callback
type timerTask struct {
	id       int
	callback *Value
	delay    int64
	executeAt int64 // Unix nano timestamp
	interval bool   // true for setInterval, false for setTimeout
}

func NewVM(doc *dom.Document) *VM {
	vm := &VM{
		doc:            doc,
		LocalStorage:   make(map[string]string),
		SessionStorage: make(map[string]string),
		pendingTimers:  make([]timerTask, 0),
	}
	if doc != nil {
		vm.root = doc.Root
	}
	vm.env = NewEnvironment(nil)
	vm.setupBuiltins()
	vm.setupPromise()
	vm.setupTimers()
	vm.setupProxy()
		vm.setupMapSet()
		vm.setupSymbol()
		vm.setupReflect()
		vm.setupPrototypes()
	return vm
}

// SetDebug enables or disables debug mode.
func (vm *VM) SetDebug(debug bool) {
	vm.debug = debug
}

func (vm *VM) debugLog(format string, args ...interface{}) {
	if vm.debug {
		fmt.Printf("[HLBR JS DEBUG] "+format+"\n", args...)
	}
}

func (vm *VM) setupBuiltins() {
	vm.env.Define("console", &Value{Type: "object", Obj: map[string]*Value{
		"log": {Type: "native", Native: func(args []*Value) *Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = valueToString(a)
			}
			vm.output = append(vm.output, strings.Join(parts, " "))
			return &Value{Type: "undefined"}
		}},
	}})

	var title string
	if vm.doc != nil {
		title = vm.doc.Title()
	}
	var bodyNode, headNode *dom.Node
	if vm.doc != nil {
		bodyNode = vm.doc.Body()
		headNode = vm.doc.Head()
	}

	// Create document object
	vm.env.Define("document", &Value{Type: "object", Obj: map[string]*Value{
		"title": {Type: "string", Str: title},
		"querySelector": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 || vm.root == nil {
				return &Value{Type: "null"}
			}
			sel := args[0].Str
			node := dom.QuerySelector(vm.root, sel)
			if node == nil {
				return &Value{Type: "null"}
			}
			return vm.wrapNode(node)
		}},
		"querySelectorAll": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 || vm.root == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			sel := args[0].Str
			nodes := dom.QuerySelectorAll(vm.root, sel)
			arr := make([]*Value, len(nodes))
			for i, n := range nodes {
				arr[i] = vm.wrapNode(n)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"getElementById": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 || vm.root == nil {
				return &Value{Type: "null"}
			}
			id := args[0].Str
			node := dom.GetElementByID(vm.root, id)
			if node == nil {
				return &Value{Type: "null"}
			}
			return vm.wrapNode(node)
		}},
		"getElementsByTagName": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 || vm.root == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			tag := args[0].Str
			nodes := dom.GetElementsByTagName(vm.root, tag)
			arr := make([]*Value, len(nodes))
			for i, n := range nodes {
				arr[i] = vm.wrapNode(n)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"body": vm.wrapNode(bodyNode),
		"head": vm.wrapNode(headNode),
	}})

	// Create window object - the global object in browsers
	vm.env.Define("window", &Value{Type: "object", Obj: map[string]*Value{
		"document": {Type: "native", Native: func(args []*Value) *Value {
			return vm.env.Get("document")
		}},
		"console": {Type: "native", Native: func(args []*Value) *Value {
			return vm.env.Get("console")
		}},
		"Math": {Type: "native", Native: func(args []*Value) *Value {
			return vm.env.Get("Math")
		}},
		"localStorage": {Type: "native", Native: func(args []*Value) *Value {
			return vm.env.Get("localStorage")
		}},
		"sessionStorage": {Type: "native", Native: func(args []*Value) *Value {
			return vm.env.Get("sessionStorage")
		}},
		"location": {Type: "object", Obj: map[string]*Value{
			"href": {Type: "string", Str: ""},
		}},
		"navigator": {Type: "object", Obj: map[string]*Value{
			"userAgent": {Type: "string", Str: "Xxlang-HLBR/1.0"},
		}},
	}})

	vm.env.Define("Math", &Value{Type: "object", Obj: map[string]*Value{
		"PI": {Type: "number", Num: math.Pi},
		"E":  {Type: "number", Num: math.E},
		"sqrt": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Sqrt(args[0].Num)}
		}},
		"abs": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Abs(args[0].Num)}
		}},
		"floor": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Floor(args[0].Num)}
		}},
		"ceil": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Ceil(args[0].Num)}
		}},
		"round": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Round(args[0].Num)}
		}},
		"max": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "number", Num: math.Inf(-1)}
			}
			m := args[0].Num
			for _, a := range args[1:] {
				if a.Num > m {
					m = a.Num
				}
			}
			return &Value{Type: "number", Num: m}
		}},
		"min": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "number", Num: math.Inf(1)}
			}
			m := args[0].Num
			for _, a := range args[1:] {
				if a.Num < m {
					m = a.Num
				}
			}
			return &Value{Type: "number", Num: m}
		}},
		"random": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: float64(len(vm.output)) / 1000}
		}},
	}})

	vm.env.Define("JSON", &Value{Type: "object", Obj: map[string]*Value{
		"stringify": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "string"}
			}
			return &Value{Type: "string", Str: valueToString(args[0])}
		}},
		"parse": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "undefined"}
			}
			return &Value{Type: "object", Str: args[0].Str}
		}},
	}})

	vm.env.Define("parseInt", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "number"}
		}
		n, _ := strconv.Atoi(args[0].Str)
		return &Value{Type: "number", Num: float64(n)}
	}})

	vm.env.Define("parseFloat", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "number"}
		}
		n, _ := strconv.ParseFloat(args[0].Str, 64)
		return &Value{Type: "number", Num: n}
	}})

	vm.env.Define("isNaN", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: args[0].Type == "number" && math.IsNaN(args[0].Num)}
	}})

	vm.env.Define("isFinite", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: args[0].Type == "number" && !math.IsInf(args[0].Num, 0) && !math.IsNaN(args[0].Num)}
	}})

	vm.env.Define("String", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "string"}
		}
		return &Value{Type: "string", Str: valueToString(args[0])}
	}})

	vm.env.Define("Number", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "number"}
		}
		n, _ := strconv.ParseFloat(args[0].Str, 64)
		return &Value{Type: "number", Num: n}
	}})

	vm.env.Define("Boolean", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: isTruthy(args[0])}
	}})

	arrayConstructor := &Value{Type: "native", BuiltInConstructor: "Array", Native: func(args []*Value) *Value {
		return &Value{Type: "object", Arr: args}
	}}
	vm.env.Define("Array", arrayConstructor)

	vm.env.Define("Object", &Value{Type: "object", Obj: GetObjectMethods(vm)})

	// RegExp constructor
	vm.env.Define("RegExp", &Value{Type: "native", Native: func(args []*Value) *Value {
		pattern := ""
		flags := ""
		if len(args) > 0 {
			// Check if first arg is already a RegExp
			if args[0].Type == "regexp" {
				// Return the same RegExp or create a new one
				return args[0]
			}
			pattern = args[0].Str
		}
		if len(args) > 1 {
			flags = args[1].Str
		}
		return newRegExp(pattern, flags)
	}})

	// Error constructors
	vm.env.Define("Error", &Value{Type: "native", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewError("Error", message)
	}})

	vm.env.Define("TypeError", &Value{Type: "native", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewTypeError(message)
	}})

	vm.env.Define("ReferenceError", &Value{Type: "native", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewReferenceError(message)
	}})

	vm.env.Define("SyntaxError", &Value{Type: "native", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewSyntaxError(message)
	}})

	vm.env.Define("RangeError", &Value{Type: "native", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewRangeError(message)
	}})


	vm.env.Define("undefined", &Value{Type: "undefined"})
	vm.env.Define("NaN", &Value{Type: "number", Num: math.NaN()})
	vm.env.Define("Infinity", &Value{Type: "number", Num: math.Inf(1)})

	// localStorage - simple key-value storage
	vm.env.Define("localStorage", &Value{Type: "object", Obj: map[string]*Value{
		"setItem": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 2 {
				vm.LocalStorage[args[0].Str] = args[1].Str
			}
			return &Value{Type: "undefined"}
		}},
		"getItem": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "null"}
			}
			if v, ok := vm.LocalStorage[args[0].Str]; ok {
				return &Value{Type: "string", Str: v}
			}
			return &Value{Type: "null"}
		}},
		"removeItem": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) > 0 {
				delete(vm.LocalStorage, args[0].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"clear": {Type: "native", Native: func(args []*Value) *Value {
			vm.LocalStorage = make(map[string]string)
			return &Value{Type: "undefined"}
		}},
		"length": {Type: "number", Num: 0}, // dynamic length not supported in simple impl
	}})

	// sessionStorage - simple key-value storage (per-session)
	vm.env.Define("sessionStorage", &Value{Type: "object", Obj: map[string]*Value{
		"setItem": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 2 {
				vm.SessionStorage[args[0].Str] = args[1].Str
			}
			return &Value{Type: "undefined"}
		}},
		"getItem": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "null"}
			}
			if v, ok := vm.SessionStorage[args[0].Str]; ok {
				return &Value{Type: "string", Str: v}
			}
			return &Value{Type: "null"}
		}},
		"removeItem": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) > 0 {
				delete(vm.SessionStorage, args[0].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"clear": {Type: "native", Native: func(args []*Value) *Value {
			vm.SessionStorage = make(map[string]string)
			return &Value{Type: "undefined"}
		}},
		"length": {Type: "number", Num: 0},
	}})
}

func (vm *VM) wrapNode(n *dom.Node) *Value {
	if n == nil {
		return &Value{Type: "null"}
	}

	// Create innerHTML getter/setter descriptor
	innerHTMLGetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		return &Value{Type: "string", Str: n.InnerHTML()}
	}}
	innerHTMLSetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) > 0 {
			// Parse the HTML fragment and replace children
			newChildren := htmlparser.ParseFragment(args[0].Str)
			n.Children = newChildren
			// Update parent references
			for _, child := range n.Children {
				child.Parent = n
			}
		}
		return &Value{Type: "undefined"}
	}}

	// Create textContent getter/setter descriptor
	textContentGetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		return &Value{Type: "string", Str: n.TextContent()}
	}}
	textContentSetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) > 0 {
			// Set text content by replacing all children with a single text node
			n.Children = []*dom.Node{
				{Type: dom.TextNode, Data: args[0].Str, Parent: n},
			}
		}
		return &Value{Type: "undefined"}
	}}

	obj := &Value{Type: "object", Obj: map[string]*Value{
		"tagName":     {Type: "string", Str: n.TagName()},
		"id":          {Type: "string", Str: n.ID()},
		"className":   {Type: "string", Str: n.ClassName()},
		"nodeName":    {Type: "string", Str: n.Data},
		"nodeValue":   {Type: "string", Str: n.GetAttribute("value")},
		"outerHTML":   {Type: "string", Str: n.OuterHTML()},
		"innerText":   {Type: "string", Str: n.TextContent()},
		"getAttribute": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "null"}
			}
			return &Value{Type: "string", Str: n.GetAttribute(args[0].Str)}
		}},
		"setAttribute": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 2 {
				n.SetAttribute(args[0].Str, args[1].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"querySelector": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "null"}
			}
			child := dom.QuerySelector(n, args[0].Str)
			return vm.wrapNode(child)
		}},
		"querySelectorAll": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			children := dom.QuerySelectorAll(n, args[0].Str)
			arr := make([]*Value, len(children))
			for i, c := range children {
				arr[i] = vm.wrapNode(c)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"children": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeList(n.Children)
		}},
		"firstChild": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(firstChild(n))
		}},
		"lastChild": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(lastChild(n))
		}},
		"parentNode": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(n.Parent)
		}},
	}}

	// Add descriptors for innerHTML and textContent with getters/setters
	obj.Descriptors = map[string]*PropertyDescriptor{
		"innerHTML":   {Get: innerHTMLGetter, Set: innerHTMLSetter, Enumerable: true, Configurable: true},
		"textContent": {Get: textContentGetter, Set: textContentSetter, Enumerable: true, Configurable: true},
	}

	return obj
}

// wrapNodeShallow wraps a node without eagerly wrapping children or parent,
// to avoid infinite recursion on circular DOM references (parent -> child -> parent).
// Access to children, firstChild, lastChild, parentNode is via lazy native functions.
func (vm *VM) wrapNodeShallow(n *dom.Node) *Value {
	if n == nil {
		return &Value{Type: "null"}
	}

	// Create innerHTML getter/setter descriptor
	innerHTMLGetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		return &Value{Type: "string", Str: n.InnerHTML()}
	}}
	innerHTMLSetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) > 0 {
			// Parse the HTML fragment and replace children
			newChildren := htmlparser.ParseFragment(args[0].Str)
			n.Children = newChildren
			// Update parent references
			for _, child := range n.Children {
				child.Parent = n
			}
		}
		return &Value{Type: "undefined"}
	}}

	// Create textContent getter/setter descriptor
	textContentGetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		return &Value{Type: "string", Str: n.TextContent()}
	}}
	textContentSetter := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) > 0 {
			// Set text content by replacing all children with a single text node
			n.Children = []*dom.Node{
				{Type: dom.TextNode, Data: args[0].Str, Parent: n},
			}
		}
		return &Value{Type: "undefined"}
	}}

	obj := &Value{Type: "object", Obj: map[string]*Value{
		"tagName":     {Type: "string", Str: n.TagName()},
		"id":          {Type: "string", Str: n.ID()},
		"className":   {Type: "string", Str: n.ClassName()},
		"nodeName":    {Type: "string", Str: n.Data},
		"nodeValue":   {Type: "string", Str: n.GetAttribute("value")},
		"outerHTML":   {Type: "string", Str: n.OuterHTML()},
		"innerText":   {Type: "string", Str: n.TextContent()},
		"getAttribute": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "null"}
			}
			return &Value{Type: "string", Str: n.GetAttribute(args[0].Str)}
		}},
		"setAttribute": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 2 {
				n.SetAttribute(args[0].Str, args[1].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"querySelector": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "null"}
			}
			child := dom.QuerySelector(n, args[0].Str)
			return vm.wrapNodeShallow(child)
		}},
		"querySelectorAll": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			children := dom.QuerySelectorAll(n, args[0].Str)
			arr := make([]*Value, len(children))
			for i, c := range children {
				arr[i] = vm.wrapNodeShallow(c)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"children": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeListShallow(n.Children)
		}},
		"firstChild": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(firstChild(n))
		}},
		"lastChild": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(lastChild(n))
		}},
		"parentNode": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(n.Parent)
		}},
		"childElementCount": {Type: "number", Num: float64(len(n.Children))},
	}}

	// Add descriptors for innerHTML and textContent with getters/setters
	obj.Descriptors = map[string]*PropertyDescriptor{
		"innerHTML":   {Get: innerHTMLGetter, Set: innerHTMLSetter, Enumerable: true, Configurable: true},
		"textContent": {Get: textContentGetter, Set: textContentSetter, Enumerable: true, Configurable: true},
	}

	return obj
}

// wrapNodeListShallow wraps a list of nodes without deep recursion.
func (vm *VM) wrapNodeListShallow(nodes []*dom.Node) *Value {
	arr := make([]*Value, len(nodes))
	for i, n := range nodes {
		arr[i] = vm.wrapNodeShallow(n)
	}
	return &Value{Type: "object", Arr: arr}
}

func (vm *VM) wrapNodeList(nodes []*dom.Node) *Value {
	arr := make([]*Value, len(nodes))
	for i, n := range nodes {
		arr[i] = vm.wrapNode(n)
	}
	return &Value{Type: "object", Arr: arr}
}

func firstChild(n *dom.Node) *dom.Node {
	if len(n.Children) > 0 {
		return n.Children[0]
	}
	return nil
}

func lastChild(n *dom.Node) *dom.Node {
	if len(n.Children) > 0 {
		return n.Children[len(n.Children)-1]
	}
	return nil
}

func (vm *VM) Run(code string) (*Value, error) {
	vm.debugLog("=== Running JS code ===")
	if vm.debug {
		// Show first 200 chars of code, or full code if shorter
		preview := code
		if len(code) > 200 {
			preview = code[:200] + "..."
		}
		vm.debugLog("Code: %s", preview)
	}

	parser := NewParser(code)
	prog := parser.Parse()

	vm.debugLog("Parsed %d statements", len(prog.Statements))

	var result *Value = &Value{Type: "undefined"}
	var returnVal *Value
	var returning bool

	for i, stmt := range prog.Statements {
		if returning {
			break
		}
		vm.debugLog("Executing statement %d: %T", i+1, stmt)
		result, returning, returnVal = vm.execStmt(stmt)
	}

	if returning && returnVal != nil {
		vm.debugLog("Returning: %s", valueToString(returnVal))
		return returnVal, nil
	}
	vm.debugLog("=== JS execution complete ===")
	return result, nil
}

func (vm *VM) execStmt(stmt Statement) (*Value, bool, *Value) {
	switch s := stmt.(type) {
	case *ExpressionStmt:
		v := vm.evalExpr(s.Expr)
		return v, false, nil
	case *VarDecl:
		if s.IsDestructuring && s.DestructPattern != nil {
			vm.execDestructuring(s.DestructPattern, s.Value)
			return &Value{Type: "undefined"}, false, nil
		}
		var val *Value = &Value{Type: "undefined"}
		if s.Value != nil {
			val = vm.evalExpr(s.Value)
		}
		vm.env.Define(s.Name, val)
		return val, false, nil
	case *IfStmt:
		cond := vm.evalExpr(s.Cond)
		if isTruthy(cond) {
			return vm.execBlock(s.Body)
		} else if len(s.Else) > 0 {
			return vm.execBlock(s.Else)
		}
		return &Value{Type: "undefined"}, false, nil
	case *ForStmt:
		if s.Init != nil {
			vm.execStmt(s.Init)
		}
		for {
			if s.Cond != nil {
				cond := vm.evalExpr(s.Cond)
				if !isTruthy(cond) {
					break
				}
			}
			ret, brk, retVal := vm.execBlock(s.Body)
			if brk {
				return ret, true, retVal
			}
			if s.Post != nil {
				vm.execStmt(s.Post)
			}
		}
		return &Value{Type: "undefined"}, false, nil
	case *ForInStmt:
		obj := vm.evalExpr(s.Object)
		if obj.Type == "object" && obj.Obj != nil {
			for key := range obj.Obj {
				childEnv := NewEnvironment(vm.env)
				childEnv.Define(s.VarName, &Value{Type: "string", Str: key})
				oldEnv := vm.env
				vm.env = childEnv
				ret, brk, retVal := vm.execBlock(s.Body)
				vm.env = oldEnv
				if brk {
					return ret, true, retVal
				}
			}
		}
		return &Value{Type: "undefined"}, false, nil
	case *ForOfStmt:
		obj := vm.evalExpr(s.Object)
		if obj.Type == "object" && obj.Arr != nil {
			for _, val := range obj.Arr {
				childEnv := NewEnvironment(vm.env)
				childEnv.Define(s.VarName, val)
				oldEnv := vm.env
				vm.env = childEnv
				ret, brk, retVal := vm.execBlock(s.Body)
				vm.env = oldEnv
				if brk {
					return ret, true, retVal
				}
			}
		} else if obj.Type == "object" && obj.Obj != nil {
			for _, val := range obj.Obj {
				childEnv := NewEnvironment(vm.env)
				childEnv.Define(s.VarName, val)
				oldEnv := vm.env
				vm.env = childEnv
				ret, brk, retVal := vm.execBlock(s.Body)
				vm.env = oldEnv
				if brk {
					return ret, true, retVal
				}
			}
		}
		return &Value{Type: "undefined"}, false, nil
	case *WhileStmt:
		for {
			cond := vm.evalExpr(s.Cond)
			if !isTruthy(cond) {
				break
			}
			ret, brk, retVal := vm.execBlock(s.Body)
			if brk {
				return ret, true, retVal
			}
		}
		return &Value{Type: "undefined"}, false, nil
	case *ReturnStmt:
		var val *Value = &Value{Type: "undefined"}
		if s.Value != nil {
			val = vm.evalExpr(s.Value)
		}
		return val, true, val
	case *BreakStmt:
		return &Value{Type: "undefined"}, true, nil
	case *ContinueStmt:
		return &Value{Type: "undefined"}, false, nil
	case *BlockStmt:
		return vm.execBlock(s.Statements)
	case *TryStmt:
		// Execute try block with panic recovery
		var result *Value = &Value{Type: "undefined"}
		var returning bool
		var retVal *Value
		var caughtException *JSException

		// Use defer/recover to catch panics
		func() {
			defer func() {
				if r := recover(); r != nil {
					switch v := r.(type) {
					case *JSException:
						caughtException = v
					default:
						caughtException = &JSException{
							Value: ToJSValue(v),
						}
					}
				}
			}()

			for _, stmt := range s.Body {
				result, returning, retVal = vm.execStmt(stmt)
				if returning {
					break
				}
			}
		}()

		// If we caught an exception or had a return, handle appropriately
		if returning {
			// Execute finally before returning
			if len(s.Finally) > 0 {
				for _, fs := range s.Finally {
					vm.execStmt(fs)
				}
			}
			if retVal != nil {
				return retVal, true, retVal
			}
			return result, true, nil
		}

		// If we caught an exception, execute catch block
		if caughtException != nil && len(s.Catch) > 0 {
			// Create a new scope with the catch variable
			childEnv := NewEnvironment(vm.env)
			if s.CatchVar != "" && caughtException.Value != nil {
				childEnv.Define(s.CatchVar, caughtException.Value)
			}
			oldEnv := vm.env
			vm.env = childEnv

			for _, stmt := range s.Catch {
				result, returning, retVal = vm.execStmt(stmt)
				if returning {
					vm.env = oldEnv
					// Execute finally before returning
					if len(s.Finally) > 0 {
						for _, fs := range s.Finally {
							vm.execStmt(fs)
						}
					}
					if retVal != nil {
						return retVal, true, retVal
					}
					return result, true, nil
				}
			}
			vm.env = oldEnv
		} else if caughtException != nil {
			// No catch block, but have an exception - execute finally then re-panic
			if len(s.Finally) > 0 {
				for _, fs := range s.Finally {
					vm.execStmt(fs)
				}
			}
			panic(caughtException)
		}

		// Execute finally block
		for _, stmt := range s.Finally {
			result, returning, retVal = vm.execStmt(stmt)
			if returning {
				if retVal != nil {
					return retVal, true, retVal
				}
				return result, true, nil
			}
		}

		return result, false, nil
	case *ThrowStmt:
		// Throw statement - evaluate the value and panic
		throwVal := vm.evalExpr(s.Value)
		ThrowJS(throwVal)
		return &Value{Type: "undefined"}, false, nil
	case *FunctionDecl:
		fn := &Function{Params: s.Params, DefaultVals: s.DefaultVals, RestParam: s.RestParam, Body: s.Body, Env: vm.env}
		val := &Value{Type: "function", Func: fn}
		if s.IsAsync {
			val.IsAsync = true
		}
		vm.env.Define(s.Name, val)
		return &Value{Type: "undefined"}, false, nil
		case *GeneratorDecl:
			// Generator function - create a generator object
			gen := &Value{
				Type: "generator",
				Func: &Function{Params: s.Params, RestParam: s.RestParam, Body: s.Body, Env: vm.env},
				Obj:  make(map[string]*Value),
			}
			// Add next() method
			gen.Obj["next"] = &Value{Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "object", Obj: map[string]*Value{
					"value":  &Value{Type: "undefined"},
					"done":   &Value{Type: "bool", Bool: true},
				}}
			}}
			vm.env.Define(s.Name, gen)
			return &Value{Type: "undefined"}, false, nil
		case *YieldStmt:
			// Yield statement - return the value with returning=true
			var val *Value = &Value{Type: "undefined"}
			if s.Value != nil {
				val = vm.evalExpr(s.Value)
			}
			return val, true, val
	case *ClassDecl:
		cls := vm.createClass(s)
		vm.env.Define(s.Name, &Value{Type: "class", Class: cls})
		return &Value{Type: "undefined"}, false, nil
	}
	return &Value{Type: "undefined"}, false, nil
}

// execDestructuring executes a destructuring assignment
func (vm *VM) execDestructuring(pattern *DestructPattern, source Expression) {
	if source == nil {
		return
	}

	val := vm.evalExpr(source)
	if val == nil {
		return
	}

	if pattern.IsArray {
		// Array destructuring
		for i, elem := range pattern.Elements {
			var elemVal *Value = &Value{Type: "undefined"}
			if val.Arr != nil && i < len(val.Arr) {
				elemVal = val.Arr[i]
			}
			if elemVal.Type == "undefined" && elem.Default != nil {
				elemVal = vm.evalExpr(elem.Default)
			}
			vm.env.Define(elem.Name, elemVal)
		}
	} else {
		// Object destructuring
		for _, elem := range pattern.Elements {
			var elemVal *Value = &Value{Type: "undefined"}
			if val.Obj != nil {
				prop := elem.Property
				if prop == "" {
					prop = elem.Name
				}
				if v, ok := val.Obj[prop]; ok {
					elemVal = v
				}
			}
			if elemVal.Type == "undefined" && elem.Default != nil {
				elemVal = vm.evalExpr(elem.Default)
			}
			vm.env.Define(elem.Name, elemVal)
		}
	}
}

func (vm *VM) execBlock(stmts []Statement) (*Value, bool, *Value) {
	childEnv := NewEnvironment(vm.env)
	oldEnv := vm.env
	vm.env = childEnv

	var result *Value = &Value{Type: "undefined"}
	for _, stmt := range stmts {
		ret, brk, retVal := vm.execStmt(stmt)
		result = ret
		if brk {
			vm.env = oldEnv
			return ret, true, retVal
		}
	}

	vm.env = oldEnv
	return result, false, nil
}

func (vm *VM) evalExpr(expr Expression) *Value {
	switch e := expr.(type) {
	case *NumberLit:
		return &Value{Type: "number", Num: e.Value}
	case *BigIntLit:
		return &Value{Type: "bigint", Str: e.Value}
	case *StringLit:
		return &Value{Type: "string", Str: e.Value}
	case *TemplateLit:
		return vm.evalTemplateLit(e)
	case *BoolLit:
		return &Value{Type: "bool", Bool: e.Value}
	case *NullLit:
		return &Value{Type: "null"}
	case *UndefinedLit:
		return &Value{Type: "undefined"}
	case *Ident:
		return vm.env.Get(e.Name)
	case *BinaryExpr:
		return vm.evalBinary(e)
	case *UnaryExpr:
		return vm.evalUnary(e)
	case *UpdateExpr:
		return vm.evalUpdate(e)
	case *AssignExpr:
		return vm.evalAssign(e)
	case *CallExpr:
		return vm.evalCall(e)
	case *MemberExpr:
		return vm.evalMember(e)
	case *ArrayLit:
		var arr []*Value
		for _, el := range e.Elements {
			if spread, ok := el.(*SpreadExpr); ok {
				// Spread the array
				val := vm.evalExpr(spread.Argument)
				if val.Type == "object" && val.Arr != nil {
					arr = append(arr, val.Arr...)
				}
			} else {
				arr = append(arr, vm.evalExpr(el))
			}
		}
		return &Value{Type: "object", Arr: arr}
	case *ObjectLit:
		obj := make(map[string]*Value)
		for _, prop := range e.Properties {
			if prop.Spread {
				// Spread operator: copy properties from source object
				src := vm.evalExpr(prop.Value)
				if src.Type == "object" && src.Obj != nil {
					for k, v := range src.Obj {
						obj[k] = v
					}
				}
			} else if prop.Computed {
				// Computed property name: [expr]
				keyVal := vm.evalExpr(prop.KeyExpr)
				key := valueToString(keyVal)
				obj[key] = vm.evalExpr(prop.Value)
			} else {
				obj[prop.Key] = vm.evalExpr(prop.Value)
			}
		}
		return &Value{Type: "object", Obj: obj}
	case *FunctionExpr:
		fn := &Function{Params: e.Params, DefaultVals: e.DefaultVals, RestParam: e.RestParam, Body: e.Body, Env: vm.env}
		return &Value{Type: "function", Func: fn}
	case *ArrowFunctionExpr:
		fn := &Function{Params: e.Params, RestParam: e.RestParam, Body: e.Body, Env: vm.env}
		return &Value{Type: "function", Func: fn}
	case *TernaryExpr:
		cond := vm.evalExpr(e.Cond)
		if isTruthy(cond) {
			return vm.evalExpr(e.True)
		}
		return vm.evalExpr(e.False)
	case *TypeOfExpr:
		v := vm.evalExpr(e.Expr)
		return &Value{Type: "string", Str: v.Type}
	case *AwaitExpr:
		// await extracts the value from a Promise
		promise := vm.evalExpr(e.Expr)
		if promise.Type == "promise" && promise.Promise != nil {
			// If promise is fulfilled, return the value
			if promise.Promise.State == "fulfilled" {
				return promise.Promise.Value
			}
			// If promise is pending, we need to wait for it
			// For now, we'll process any pending promises synchronously
			vm.processPendingPromises()
			if promise.Promise.State == "fulfilled" {
				return promise.Promise.Value
			}
		}
		return promise
		case *OptionalChainExpr:
			// Optional chaining: obj?.prop - returns undefined if obj is null/undefined
			obj := vm.evalExpr(e.Object)
			if obj.Type == "undefined" || obj.Type == "null" {
				return &Value{Type: "undefined"}
			}
			// Evaluate property access
			if e.Computed {
				prop := vm.evalExpr(e.Property)
				key := valueToString(prop)
				if obj.Type == "object" && obj.Obj != nil {
					if val, ok := obj.Obj[key]; ok {
						return val
					}
				}
			} else {
				if ident, ok := e.Property.(*Ident); ok {
					if obj.Type == "object" && obj.Obj != nil {
						if val, ok := obj.Obj[ident.Name]; ok {
							return val
						}
					}
				}
			}
			return &Value{Type: "undefined"}
		case *NullishCoalescingExpr:
			// Nullish coalescing: a ?? b - returns b if a is null/undefined
			left := vm.evalExpr(e.Left)
			if left.Type == "undefined" || left.Type == "null" {
				return vm.evalExpr(e.Right)
			}
			return left
	case *ClassExpr:
		cls := vm.createClassFromExpr(e)
		return &Value{Type: "class", Class: cls}
	case *ThisExpr:
		return vm.env.Get("this")
	case *SuperExpr:
		// super() calls parent constructor
		if vm.currentClass != nil && vm.currentClass.SuperClass != "" {
			// Get parent class from environment
			parentClassVal := vm.env.Get(vm.currentClass.SuperClass)
			if parentClassVal.Type == "class" && parentClassVal.Class != nil {
				parentClass := parentClassVal.Class
				thisVal := vm.env.Get("this")
				if constructor, ok := parentClass.Methods["constructor"]; ok {
					childEnv := NewEnvironment(constructor.Env)
					childEnv.Define("this", thisVal)
					for i, param := range constructor.Params {
						if i < len(e.Args) {
							childEnv.Define(param, vm.evalExpr(e.Args[i]))
						}
					}
					oldEnv := vm.env
					oldClass := vm.currentClass
					vm.currentClass = parentClass
					vm.env = childEnv
					for _, stmt := range constructor.Body {
						vm.execStmt(stmt)
					}
					vm.env = oldEnv
					vm.currentClass = oldClass
				}
			}
		}
		return &Value{Type: "undefined"}
	case *InstanceofExpr:
		obj := vm.evalExpr(e.Object)
		constructor := vm.evalExpr(e.Constructor)
		return &Value{Type: "bool", Bool: vm.instanceof(obj, constructor)}
	case *InExpr:
		prop := vm.evalExpr(e.Property)
		obj := vm.evalExpr(e.Object)
		if obj.Type == "object" && obj.Obj != nil {
			key := valueToString(prop)
			_, ok := obj.Obj[key]
			return &Value{Type: "bool", Bool: ok}
		}
		return &Value{Type: "bool", Bool: false}
	case *NewExpr:
		return vm.evalNew(e)
	}
	return &Value{Type: "undefined"}
}


// evalTemplateLit evaluates a template literal
func (vm *VM) evalTemplateLit(e *TemplateLit) *Value {
	var result string
	for _, part := range e.Parts {
		if part.IsExpr {
			val := vm.evalExpr(part.Expr)
			result += valueToString(val)
		} else {
			result += part.Text
		}
	}
	return &Value{Type: "string", Str: result}
}

func (vm *VM) evalBinary(e *BinaryExpr) *Value {
	left := vm.evalExpr(e.Left)
	right := vm.evalExpr(e.Right)

	switch e.Op {
	case "+":
		if left.Type == "string" || right.Type == "string" {
			return &Value{Type: "string", Str: valueToString(left) + valueToString(right)}
		}
		return &Value{Type: "number", Num: left.Num + right.Num}
	case "-":
		return &Value{Type: "number", Num: left.Num - right.Num}
	case "*":
		return &Value{Type: "number", Num: left.Num * right.Num}
	case "/":
		return &Value{Type: "number", Num: left.Num / right.Num}
	case "%":
		return &Value{Type: "number", Num: math.Mod(left.Num, right.Num)}
	case "==":
		return &Value{Type: "bool", Bool: valuesEqual(left, right)}
	case "===":
		return &Value{Type: "bool", Bool: valuesStrictEqual(left, right)}
	case "!=":
		return &Value{Type: "bool", Bool: !valuesEqual(left, right)}
	case "!==":
		return &Value{Type: "bool", Bool: !valuesStrictEqual(left, right)}
	case "<":
		return &Value{Type: "bool", Bool: left.Num < right.Num}
	case ">":
		return &Value{Type: "bool", Bool: left.Num > right.Num}
	case "<=":
		return &Value{Type: "bool", Bool: left.Num <= right.Num}
	case ">=":
		return &Value{Type: "bool", Bool: left.Num >= right.Num}
	case "&&":
		if isTruthy(left) {
			return right
		}
		return left
	case "||":
		if isTruthy(left) {
			return left
		}
		return right
	}
	return &Value{Type: "undefined"}
}

func (vm *VM) evalUnary(e *UnaryExpr) *Value {
	val := vm.evalExpr(e.Expr)
	switch e.Op {
	case "!":
		return &Value{Type: "bool", Bool: !isTruthy(val)}
	case "-":
		return &Value{Type: "number", Num: -val.Num}
	case "+":
		return &Value{Type: "number", Num: val.Num}
	case "typeof":
		return &Value{Type: "string", Str: val.Type}
	}
	return val
}

func (vm *VM) evalAssign(e *AssignExpr) *Value {
	val := vm.evalExpr(e.Right)

	switch left := e.Left.(type) {
	case *Ident:
		switch e.Op {
		case "=":
			vm.env.Set(left.Name, val)
		case "+=":
			cur := vm.env.Get(left.Name)
			vm.env.Set(left.Name, &Value{Type: "number", Num: cur.Num + val.Num})
		case "-=":
			cur := vm.env.Get(left.Name)
			vm.env.Set(left.Name, &Value{Type: "number", Num: cur.Num - val.Num})
		case "*=":
			cur := vm.env.Get(left.Name)
			vm.env.Set(left.Name, &Value{Type: "number", Num: cur.Num * val.Num})
		case "/=":
			cur := vm.env.Get(left.Name)
			vm.env.Set(left.Name, &Value{Type: "number", Num: cur.Num / val.Num})
		}
	case *MemberExpr:
		obj := vm.evalExpr(left.Object)

		// Handle Proxy set trap
		if obj.Type == "proxy" && obj.Proxy != nil {
			return vm.evalProxySet(obj.Proxy, left, val)
		}

		if obj.Type == "object" && obj.Obj != nil {
			prop := ""
			if left.Computed {
				prop = valueToString(vm.evalExpr(left.Property))
			} else if ident, ok := left.Property.(*Ident); ok {
				prop = ident.Name
			}
				// Check for setter via descriptor
				if obj.Descriptors != nil {
					if desc, ok := obj.Descriptors[prop]; ok && desc.Set != nil {
						vm.callFunction(desc.Set, []*Value{val})
						return val
					}
				}
			obj.Obj[prop] = val
		}
	}

	return val
}

func (vm *VM) evalCall(e *CallExpr) *Value {
	var thisBinding *Value = nil

	// Check if this is a method call (obj.method())
	if member, ok := e.Callee.(*MemberExpr); ok {
		thisBinding = vm.evalExpr(member.Object)
	}

	callee := vm.evalExpr(e.Callee)

	var args []*Value
	for _, arg := range e.Args {
		args = append(args, vm.evalExpr(arg))
	}

	// Handle native functions (including array methods)
	if callee.Type == "native" && callee.Native != nil {
		return callee.Native(args)
	}

	// Handle array methods (type set in evalMember)
	if callee.Type == "arrayMethod" && callee.Native != nil {
		return callee.Native(args)
	}

	// Handle string methods (type set in evalMember)
	if callee.Type == "stringMethod" && callee.Native != nil {
		return callee.Native(args)
	}

	// Handle generator function call - return a new iterator object
	if callee.Type == "generator" && callee.Func != nil {
		// Create an iterator object that captures the function and its environment
		iteratorObj := &Value{
			Type: "iterator",
			Obj: map[string]*Value{
				"_genFunc":  callee,
				"_args":     &Value{Type: "object", Arr: args},
				"_state":    &Value{Type: "number", Num: 0}, // 0 = initial, 1 = running, 2 = done
				"_yielded":  &Value{Type: "undefined"},
				"_done":     &Value{Type: "bool", Bool: false},
			},
		}
		// Add next() method
		iteratorObj.Obj["next"] = &Value{Type: "native", Native: func(callArgs []*Value) *Value {
			genFunc := iteratorObj.Obj["_genFunc"]
			argsVal := iteratorObj.Obj["_args"]
			state := iteratorObj.Obj["_state"]

			// If already done, return done result
			if state.Num >= 2 {
				return &Value{Type: "object", Obj: map[string]*Value{
					"value": &Value{Type: "undefined"},
					"done":  &Value{Type: "bool", Bool: true},
				}}
			}

			// Mark as running
			iteratorObj.Obj["_state"] = &Value{Type: "number", Num: 1}

			// Execute the generator function
			fn := genFunc.Func
			childEnv := NewEnvironment(fn.Env)
			if argsVal != nil && argsVal.Arr != nil {
				for i, param := range fn.Params {
					if i < len(argsVal.Arr) {
						childEnv.Define(param, argsVal.Arr[i])
					} else {
						childEnv.Define(param, &Value{Type: "undefined"})
					}
				}
				if fn.RestParam != "" {
					restArgs := make([]*Value, 0)
					if len(argsVal.Arr) > len(fn.Params) {
						restArgs = argsVal.Arr[len(fn.Params):]
					}
					childEnv.Define(fn.RestParam, &Value{Type: "object", Arr: restArgs})
				}
			}

			oldEnv := vm.env
			vm.env = childEnv

			var yieldedValue *Value = &Value{Type: "undefined"}
			var isDone bool = true

			for _, stmt := range fn.Body {
				ret, returning, retVal := vm.execStmt(stmt)
				if returning {
					// Check if this is a yield (not return)
					if _, isYield := stmt.(*YieldStmt); isYield {
						yieldedValue = ret
						isDone = false
						// Store state for potential resume
						iteratorObj.Obj["_yielded"] = yieldedValue
						break
					} else {
						// Actual return - generator is done
						if retVal != nil {
							yieldedValue = retVal
						} else {
							yieldedValue = ret
						}
						isDone = true
						break
					}
				}
			}

			vm.env = oldEnv

			if isDone {
				iteratorObj.Obj["_state"] = &Value{Type: "number", Num: 2}
				iteratorObj.Obj["_done"] = &Value{Type: "bool", Bool: true}
			}

			return &Value{Type: "object", Obj: map[string]*Value{
				"value": yieldedValue,
				"done":  &Value{Type: "bool", Bool: isDone},
			}}
		}}
		return iteratorObj
	}

	if callee.Type == "function" && callee.Func != nil {
		fn := callee.Func
		childEnv := NewEnvironment(fn.Env)

		// Bind 'this' - prefer ThisBinding from the function value, otherwise use context
		if callee.ThisBinding != nil {
			childEnv.Define("this", callee.ThisBinding)
		} else if thisBinding != nil {
			childEnv.Define("this", thisBinding)
		}

		for i, param := range fn.Params {
			if i < len(args) {
				childEnv.Define(param, args[i])
			} else if fn.DefaultVals != nil && fn.DefaultVals[param] != nil {
				// Use default value
				childEnv.Define(param, vm.evalExpr(fn.DefaultVals[param]))
			} else {
				childEnv.Define(param, &Value{Type: "undefined"})
			}
		}

		// Handle rest parameter
		if fn.RestParam != "" {
			restArgs := make([]*Value, 0)
			if len(args) > len(fn.Params) {
				restArgs = args[len(fn.Params):]
			}
			childEnv.Define(fn.RestParam, &Value{Type: "object", Arr: restArgs})
		}

		oldEnv := vm.env
		vm.env = childEnv

		// Handle async function - return a Promise
		if callee.IsAsync {
			p := &Promise{
				State:     "pending",
				OnFulfill: make([]*Function, 0),
				OnReject:  make([]*Function, 0),
				Env:       vm.env,
			}

			var result *Value = &Value{Type: "undefined"}
			for _, stmt := range fn.Body {
				ret, returning, retVal := vm.execStmt(stmt)
				result = ret
				if returning {
					// Resolve the promise with the return value
					p.State = "fulfilled"
					if retVal != nil {
						p.Value = retVal
					} else {
						p.Value = result
					}
					break
				}
			}

			// If no return, resolve with undefined
			if p.State == "pending" {
				p.State = "fulfilled"
				p.Value = result
			}

			vm.env = oldEnv
			return &Value{Type: "promise", Promise: p}
		}

		var result *Value = &Value{Type: "undefined"}
		for _, stmt := range fn.Body {
			ret, returning, retVal := vm.execStmt(stmt)
			result = ret
			if returning {
				vm.env = oldEnv
				if retVal != nil {
					return retVal
				}
				return result
			}
		}

		vm.env = oldEnv
		return result
	}

	return &Value{Type: "undefined"}
}

func (vm *VM) evalMember(e *MemberExpr) *Value {
	obj := vm.evalExpr(e.Object)

	// Handle Proxy get trap
	if obj.Type == "proxy" && obj.Proxy != nil {
		return vm.evalProxyGet(obj.Proxy, e)
	}

	// Handle Map and Set types - they store methods in Obj field
	if obj.Type == "map" || obj.Type == "set" || obj.Type == "weakmap" || obj.Type == "weakset" || obj.Type == "iterator" {
		if obj.Obj != nil {
			if ident, ok := e.Property.(*Ident); ok {
				if v, ok := obj.Obj[ident.Name]; ok {
					// If the value is a function, bind this
					if v.Type == "native" || v.Type == "function" {
						return &Value{
							Type:        v.Type,
							Native:      v.Native,
							Func:        v.Func,
							ThisBinding: obj,
						}
					}
					return v
				}
			}
		}
	}

	// Handle RegExp type
	if obj.Type == "regexp" && obj.Obj != nil {
		if ident, ok := e.Property.(*Ident); ok {
			if v, ok := obj.Obj[ident.Name]; ok {
				// If the value is a function (exec, test), bind this
				if v.Type == "native" || v.Type == "function" {
					return &Value{
						Type:        v.Type,
						Native:      v.Native,
						Func:        v.Func,
						ThisBinding: obj,
					}
				}
				return v
			}
		}
	}

	if obj.Type == "object" {
		if e.Computed {
			prop := vm.evalExpr(e.Property)
			key := valueToString(prop)
			// Check for getter via descriptor
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[key]; ok && desc.Get != nil {
					return vm.callFunction(desc.Get, []*Value{})
				}
			}
			if obj.Obj != nil {
				if v, ok := obj.Obj[key]; ok {
					// If the value is a function, bind this
					if v.Type == "function" {
						return &Value{
							Type:        "function",
							Func:        v.Func,
							ThisBinding: obj,
						}
					}
					return v
				}
			}
			if obj.Arr != nil {
				// Check for array methods
				if key == "length" {
					return &Value{Type: "number", Num: float64(len(obj.Arr))}
				}
				// Check for array index
				idx := int(prop.Num)
				if idx >= 0 && idx < len(obj.Arr) {
					return obj.Arr[idx]
				}
				// Check for array method
				if isArrayMethod(key) {
					return &Value{Type: "arrayMethod", Arr: obj.Arr, Str: key, Native: func(args []*Value) *Value {
						return callArrayMethod(key, obj, args, vm)
					}}
				}
			}
		} else if ident, ok := e.Property.(*Ident); ok {
			// Check for getter via descriptor
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[ident.Name]; ok && desc.Get != nil {
					return vm.callFunction(desc.Get, []*Value{})
				}
			}
				if obj.Obj != nil {
					if v, ok := obj.Obj[ident.Name]; ok {
						// If the value is a function, bind this
						if v.Type == "function" {
							return &Value{
								Type:        "function",
								Func:        v.Func,
								ThisBinding: obj,
							}
						}
						return v
					}
				}
			// Check for array methods (support empty arrays too)
			// An array is identified by having Arr field (even if nil/empty) and no Obj field
			if obj.Type == "object" {
				if ident.Name == "length" {
					return &Value{Type: "number", Num: float64(len(obj.Arr))}
				}
				// Check for array method
				if isArrayMethod(ident.Name) {
					return &Value{Type: "arrayMethod", Arr: obj.Arr, Str: ident.Name, Native: func(args []*Value) *Value {
						return callArrayMethod(ident.Name, obj, args, vm)
					}}
				}
			}
		}
	}

	if obj.Type == "string" {
		if ident, ok := e.Property.(*Ident); ok {
			switch ident.Name {
			case "length":
				return &Value{Type: "number", Num: float64(utf8.RuneCountInString(obj.Str))}
			}
			// Check for string methods
			if isStringMethod(ident.Name) {
				str := obj.Str
				return &Value{Type: "stringMethod", Str: str, Native: func(args []*Value) *Value {
					return callStringMethod(ident.Name, str, args, vm)
				}}
			}
		}
		// Computed property access for strings
		if e.Computed {
			prop := vm.evalExpr(e.Property)
			key := valueToString(prop)
			// Try numeric index
			if idx, err := strconv.Atoi(key); err == nil {
				runes := []rune(obj.Str)
				if idx >= 0 && idx < len(runes) {
					return &Value{Type: "string", Str: string(runes[idx])}
				}
			}
		}
	}

	// Handle Promise methods
	if obj.Type == "promise" && obj.Promise != nil {
		if ident, ok := e.Property.(*Ident); ok {
			switch ident.Name {
			case "then":
				return &Value{Type: "native", Native: func(args []*Value) *Value {
					if len(args) > 0 && (args[0].Type == "function" || args[0].Type == "native") {
						if obj.Promise.State == "fulfilled" {
							vm.callFunction(args[0], []*Value{obj.Promise.Value})
						} else {
							if args[0].Func != nil {
								obj.Promise.OnFulfill = append(obj.Promise.OnFulfill, args[0].Func)
							}
						}
					}
					return obj
				}}
			case "catch":
				return &Value{Type: "native", Native: func(args []*Value) *Value {
					if len(args) > 0 && (args[0].Type == "function" || args[0].Type == "native") {
						if obj.Promise.State == "rejected" {
							vm.callFunction(args[0], []*Value{obj.Promise.Rejection})
						} else {
							if args[0].Func != nil {
								obj.Promise.OnReject = append(obj.Promise.OnReject, args[0].Func)
							}
						}
					}
					return obj
				}}
			}
		}
	}

	// Handle class static methods
	if obj.Type == "class" && obj.Class != nil {
		if ident, ok := e.Property.(*Ident); ok {
			if method, ok := obj.Class.Static[ident.Name]; ok {
				return &Value{Type: "function", Func: method}
			}
		}
	}

	// Handle Function.prototype.call
	if obj.Type == "function" && obj.Func != nil {
		if ident, ok := e.Property.(*Ident); ok {
			if ident.Name == "call" {
				// Return a native function that calls the original function with given this
				fn := obj.Func
				originalThis := obj.ThisBinding
				return &Value{
					Type: "native",
					Native: func(args []*Value) *Value {
						childEnv := NewEnvironment(fn.Env)
						// First arg is this, rest are function args
						var boundThis *Value = originalThis
						var fnArgs []*Value
						if len(args) > 0 {
							boundThis = args[0]
							fnArgs = args[1:]
						}
						if boundThis != nil {
							childEnv.Define("this", boundThis)
						}
						for i, param := range fn.Params {
							if i < len(fnArgs) {
								childEnv.Define(param, fnArgs[i])
							} else {
								childEnv.Define(param, &Value{Type: "undefined"})
							}
						}
						oldEnv := vm.env
						vm.env = childEnv
						var result *Value = &Value{Type: "undefined"}
						for _, stmt := range fn.Body {
							ret, returning, retVal := vm.execStmt(stmt)
							result = ret
							if returning {
								vm.env = oldEnv
								if retVal != nil {
									return retVal
								}
								return result
							}
						}
						vm.env = oldEnv
						return result
					},
				}
			}
		}
	}

	return &Value{Type: "undefined"}
}

func (vm *VM) evalNew(e *NewExpr) *Value {
	callee := vm.evalExpr(e.Callee)

	// Handle class instantiation
	if callee.Type == "class" && callee.Class != nil {
		return vm.instantiateClass(callee.Class, e.Args)
	}

	// Handle object with constructor (like Promise)
	if callee.Type == "object" && callee.Obj != nil {
		if constructor, ok := callee.Obj["constructor"]; ok {
			if constructor.Type == "native" && constructor.Native != nil {
				var args []*Value
				for _, arg := range e.Args {
					args = append(args, vm.evalExpr(arg))
				}
				return constructor.Native(args)
			}
			if constructor.Type == "function" && constructor.Func != nil {
				var args []*Value
				for _, arg := range e.Args {
					args = append(args, vm.evalExpr(arg))
				}
				return vm.callFunction(constructor, args)
			}
		}
	}

	// Handle native function as constructor (like Proxy)
	if callee.Type == "native" && callee.Native != nil {
		var args []*Value
		for _, arg := range e.Args {
			args = append(args, vm.evalExpr(arg))
		}
		return callee.Native(args)
	}

	// Handle function as constructor
	if callee.Type == "function" && callee.Func != nil {
		fn := callee.Func
		childEnv := NewEnvironment(fn.Env)

		var args []*Value
		for _, arg := range e.Args {
			args = append(args, vm.evalExpr(arg))
		}
		for i, param := range fn.Params {
			if i < len(args) {
				childEnv.Define(param, args[i])
			}
		}

		obj := &Value{Type: "object", Obj: make(map[string]*Value)}
		childEnv.Define("this", obj)

		oldEnv := vm.env
		vm.env = childEnv
		for _, stmt := range fn.Body {
			vm.execStmt(stmt)
		}
		vm.env = oldEnv

		return obj
	}
	return &Value{Type: "object", Obj: make(map[string]*Value)}
}

// createClass creates a Class struct from a ClassDecl
func (vm *VM) createClass(decl *ClassDecl) *Class {
	cls := &Class{
		Name:       decl.Name,
		SuperClass: decl.SuperClass,
		Methods:    make(map[string]*Function),
		Static:     make(map[string]*Function),
		Getters:    make(map[string]*Function),
		Setters:    make(map[string]*Function),
		Env:        vm.env,
	}

	for _, elem := range decl.Body {
		fn := &Function{Params: elem.Params, Body: elem.Body, Env: vm.env}
		switch {
		case elem.IsStatic:
			cls.Static[elem.Name] = fn
		case elem.IsGetter:
			cls.Getters[elem.Name] = fn
		case elem.IsSetter:
			cls.Setters[elem.Name] = fn
		default:
			cls.Methods[elem.Name] = fn
		}
	}

	return cls
}

// createClassFromExpr creates a Class struct from a ClassExpr
func (vm *VM) createClassFromExpr(expr *ClassExpr) *Class {
	cls := &Class{
		Name:       expr.Name,
		SuperClass: expr.SuperClass,
		Methods:    make(map[string]*Function),
		Static:     make(map[string]*Function),
		Getters:    make(map[string]*Function),
		Setters:    make(map[string]*Function),
		Env:        vm.env,
	}

	for _, elem := range expr.Body {
		fn := &Function{Params: elem.Params, Body: elem.Body, Env: vm.env}
		switch {
		case elem.IsStatic:
			cls.Static[elem.Name] = fn
		case elem.IsGetter:
			cls.Getters[elem.Name] = fn
		case elem.IsSetter:
			cls.Setters[elem.Name] = fn
		default:
			cls.Methods[elem.Name] = fn
		}
	}

	return cls
}

// instantiateClass creates a new instance from a class
func (vm *VM) instantiateClass(cls *Class, args []Expression) *Value {
	// Create the instance object
	obj := &Value{
		Type: "object",
		Obj:  make(map[string]*Value),
	}

	// Set up prototype chain (methods)
	for name, method := range cls.Methods {
		obj.Obj[name] = &Value{
			Type: "function",
			Func: method,
		}
	}

	// Call constructor if exists
	if constructor, ok := cls.Methods["constructor"]; ok {
		childEnv := NewEnvironment(constructor.Env)
		childEnv.Define("this", obj)

		for i, param := range constructor.Params {
			if i < len(args) {
				childEnv.Define(param, vm.evalExpr(args[i]))
			}
		}

		oldEnv := vm.env
		oldClass := vm.currentClass
		vm.currentClass = cls
		vm.env = childEnv
		for _, stmt := range constructor.Body {
			vm.execStmt(stmt)
		}
		vm.env = oldEnv
		vm.currentClass = oldClass
	}

	return obj
}

func isTruthy(v *Value) bool {
	switch v.Type {
	case "undefined", "null":
		return false
	case "bool":
		return v.Bool
	case "number":
		return v.Num != 0 && !math.IsNaN(v.Num)
	case "string":
		return v.Str != ""
	case "object":
		return true
	}
	return false
}

func valuesEqual(a, b *Value) bool {
	if a.Type == "undefined" && b.Type == "undefined" {
		return true
	}
	if a.Type == "null" && b.Type == "null" {
		return true
	}
	if a.Type == "number" && b.Type == "number" {
		return a.Num == b.Num
	}
	if a.Type == "string" && b.Type == "string" {
		return a.Str == b.Str
	}
	if a.Type == "bool" && b.Type == "bool" {
		return a.Bool == b.Bool
	}
	return false
}

func valuesStrictEqual(a, b *Value) bool {
	if a.Type != b.Type {
		return false
	}
	return valuesEqual(a, b)
}

func valueToString(v *Value) string {
	switch v.Type {
	case "undefined":
		return "undefined"
	case "null":
		return "null"
	case "bool":
		if v.Bool {
			return "true"
		}
		return "false"
	case "number":
		if math.IsNaN(v.Num) {
			return "NaN"
		}
		if math.IsInf(v.Num, 1) {
			return "Infinity"
		}
		if math.IsInf(v.Num, -1) {
			return "-Infinity"
		}
		return strconv.FormatFloat(v.Num, 'f', -1, 64)
	case "string":
		return v.Str
	case "object":
		if v.Arr != nil {
			parts := make([]string, len(v.Arr))
			for i, a := range v.Arr {
				parts[i] = valueToString(a)
			}
			return "[" + strings.Join(parts, ",") + "]"
		}
		if v.Obj != nil {
			return "[object Object]"
		}
		return "[object]"
	case "function":
		return "[function]"
	case "native":
		return "[native function]"
	}
	return ""
}

func (vm *VM) Output() []string {
	return vm.output
}

// isArrayMethod checks if a name is a built-in array method.
func isArrayMethod(name string) bool {
	switch name {
	case "forEach", "map", "filter", "find", "findIndex", "some", "every",
		"reduce", "reduceRight", "indexOf", "lastIndexOf", "includes",
		"join", "reverse", "sort", "slice", "splice", "push", "pop",
		"shift", "unshift", "concat", "flat", "flatMap", "fill",
		"copyWithin", "at", "toReversed", "toSorted", "toSpliced", "with":
		return true
	}
	return false
}

// callFunction calls a function value with the given arguments.
// This is used by array methods to invoke callback functions.
func (vm *VM) callFunction(fn *Value, args []*Value) *Value {
	if fn.Type == "native" && fn.Native != nil {
		return fn.Native(args)
	}

	if fn.Type == "function" && fn.Func != nil {
		f := fn.Func
		childEnv := NewEnvironment(f.Env)

		for i, param := range f.Params {
			if i < len(args) {
				childEnv.Define(param, args[i])
			} else {
				childEnv.Define(param, &Value{Type: "undefined"})
			}
		}

		oldEnv := vm.env
		vm.env = childEnv

		var result *Value = &Value{Type: "undefined"}
		for _, stmt := range f.Body {
			ret, returning, retVal := vm.execStmt(stmt)
			result = ret
			if returning {
				vm.env = oldEnv
				if retVal != nil {
					return retVal
				}
				return result
			}
		}

		vm.env = oldEnv
		return result
	}

	return &Value{Type: "undefined"}
}

// setupPromise adds Promise constructor and methods to the VM
func (vm *VM) setupPromise() {
	// Promise constructor function
	promiseConstructor := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "undefined"}
		}
		executor := args[0]
		if executor.Type != "function" && executor.Type != "native" {
			return &Value{Type: "undefined"}
		}

		p := &Promise{
			State:     "pending",
			OnFulfill: make([]*Function, 0),
			OnReject:  make([]*Function, 0),
			Env:       vm.env,
		}

		// Create resolve and reject functions
		resolve := &Value{Type: "native", Native: func(args []*Value) *Value {
			if p.State == "pending" {
				p.State = "fulfilled"
				if len(args) > 0 {
					p.Value = args[0]
				}
				// Call all onFulfill callbacks
				for _, fn := range p.OnFulfill {
					vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{p.Value})
				}
			}
			return &Value{Type: "undefined"}
		}}

		reject := &Value{Type: "native", Native: func(args []*Value) *Value {
			if p.State == "pending" {
				p.State = "rejected"
				if len(args) > 0 {
					p.Rejection = args[0]
				}
				// Call all onReject callbacks
				for _, fn := range p.OnReject {
					vm.callFunction(&Value{Type: "function", Func: fn}, []*Value{p.Rejection})
				}
			}
			return &Value{Type: "undefined"}
		}}

		// Execute the executor function
		if executor.Type == "function" && executor.Func != nil {
			childEnv := NewEnvironment(executor.Func.Env)
			childEnv.Define("resolve", resolve)
			childEnv.Define("reject", reject)
			oldEnv := vm.env
			vm.env = childEnv
			for _, stmt := range executor.Func.Body {
				vm.execStmt(stmt)
			}
			vm.env = oldEnv
		} else if executor.Type == "native" && executor.Native != nil {
			executor.Native([]*Value{resolve, reject})
		}

		return &Value{Type: "promise", Promise: p}
	}}

	// Promise.resolve static method
	promiseResolve := &Value{Type: "native", Native: func(args []*Value) *Value {
		p := &Promise{
			State:     "fulfilled",
			OnFulfill: make([]*Function, 0),
			OnReject:  make([]*Function, 0),
			Env:       vm.env,
		}
		if len(args) > 0 {
			p.Value = args[0]
		}
		return &Value{Type: "promise", Promise: p}
	}}

	// Promise.reject static method
	promiseReject := &Value{Type: "native", Native: func(args []*Value) *Value {
		p := &Promise{
			State:     "rejected",
			OnReject:  make([]*Function, 0),
			OnFulfill: make([]*Function, 0),
			Env:       vm.env,
		}
		if len(args) > 0 {
			p.Rejection = args[0]
		}
		return &Value{Type: "promise", Promise: p}
	}}

	// Promise.all static method
	promiseAll := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "promise", Promise: &Promise{
				State:     "fulfilled",
				Value:     &Value{Type: "object", Arr: []*Value{}},
				OnFulfill: make([]*Function, 0),
				OnReject:  make([]*Function, 0),
				Env:       vm.env,
			}}
		}
		promises := args[0]
		if promises.Type != "object" || promises.Arr == nil {
			return &Value{Type: "promise", Promise: &Promise{
				State:     "rejected",
				Rejection: &Value{Type: "string", Str: "Promise.all requires an array"},
				OnFulfill: make([]*Function, 0),
				OnReject:  make([]*Function, 0),
				Env:       vm.env,
			}}
		}
		results := make([]*Value, len(promises.Arr))
		for i, p := range promises.Arr {
			if p.Type == "promise" && p.Promise != nil && p.Promise.State == "fulfilled" {
				results[i] = p.Promise.Value
			} else if p.Type != "promise" {
				results[i] = p
			} else {
				results[i] = &Value{Type: "undefined"}
			}
		}
		return &Value{Type: "promise", Promise: &Promise{
			State:     "fulfilled",
			Value:     &Value{Type: "object", Arr: results},
			OnFulfill: make([]*Function, 0),
			OnReject:  make([]*Function, 0),
			Env:       vm.env,
		}}
	}}

	// Define Promise as an object with constructor and static methods
	vm.env.Define("Promise", &Value{
		Type: "object",
		Obj: map[string]*Value{
			"constructor": promiseConstructor,
			"resolve":     promiseResolve,
			"reject":      promiseReject,
			"all":         promiseAll,
		},
	})

	// Also define as callable for new Promise()
	// Store the constructor for NewExpr to use
	vm.env.Define("_PromiseConstructor", promiseConstructor)
}

// processPendingPromises processes any pending promises synchronously
func (vm *VM) processPendingPromises() {
	// This is a simplified synchronous implementation
}

// setupMapSet adds Map and Set constructors to the VM
func (vm *VM) setupMapSet() {
	// Map constructor
	vm.env.Define("Map", &Value{Type: "native", Native: func(args []*Value) *Value {
		m := &Value{
			Type:    "map",
			MapData: make(map[string]*Value),
			Obj:     make(map[string]*Value),
		}
		// Add methods
		m.Obj["set"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 2 && m.MapData != nil {
				key := valueToString(args[0])
				m.MapData[key] = args[1]
			}
			return m
		}}
		m.Obj["get"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && m.MapData != nil {
				key := valueToString(args[0])
				if val, ok := m.MapData[key]; ok {
					return val
				}
			}
			return &Value{Type: "undefined"}
		}}
		m.Obj["has"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && m.MapData != nil {
				key := valueToString(args[0])
				_, ok := m.MapData[key]
				return &Value{Type: "bool", Bool: ok}
			}
			return &Value{Type: "bool", Bool: false}
		}}
		m.Obj["delete"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && m.MapData != nil {
				key := valueToString(args[0])
				delete(m.MapData, key)
			}
			return &Value{Type: "undefined"}
		}}
		m.Obj["clear"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			m.MapData = make(map[string]*Value)
			return &Value{Type: "undefined"}
		}}
		m.Obj["size"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: float64(len(m.MapData))}
		}}
		return m
	}})

	// Set constructor
	vm.env.Define("Set", &Value{Type: "native", Native: func(args []*Value) *Value {
		s := &Value{
			Type:    "set",
			SetData: make(map[string]bool),
			Obj:     make(map[string]*Value),
		}
		// Add methods
		s.Obj["add"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && s.SetData != nil {
				key := valueToString(args[0])
				s.SetData[key] = true
			}
			return s
		}}
		s.Obj["has"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && s.SetData != nil {
				key := valueToString(args[0])
				_, ok := s.SetData[key]
				return &Value{Type: "bool", Bool: ok}
			}
			return &Value{Type: "bool", Bool: false}
		}}
		s.Obj["delete"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && s.SetData != nil {
				key := valueToString(args[0])
				delete(s.SetData, key)
			}
			return &Value{Type: "undefined"}
		}}
		s.Obj["clear"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			s.SetData = make(map[string]bool)
			return &Value{Type: "undefined"}
		}}
		s.Obj["size"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: float64(len(s.SetData))}
		}}
		return s
	}})

	// WeakMap constructor
	vm.env.Define("WeakMap", &Value{Type: "native", Native: func(args []*Value) *Value {
		wm := &Value{
			Type:    "weakmap",
			MapData: make(map[string]*Value),
			Obj:     make(map[string]*Value),
		}
		wm.Obj["set"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 2 && wm.MapData != nil {
				key := valueToString(args[0])
				wm.MapData[key] = args[1]
			}
			return wm
		}}
		wm.Obj["get"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && wm.MapData != nil {
				key := valueToString(args[0])
				if val, ok := wm.MapData[key]; ok {
					return val
				}
			}
			return &Value{Type: "undefined"}
		}}
		wm.Obj["has"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && wm.MapData != nil {
				key := valueToString(args[0])
				_, ok := wm.MapData[key]
				return &Value{Type: "bool", Bool: ok}
			}
			return &Value{Type: "bool", Bool: false}
		}}
		wm.Obj["delete"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && wm.MapData != nil {
				key := valueToString(args[0])
				delete(wm.MapData, key)
			}
			return &Value{Type: "undefined"}
		}}
		return wm
	}})

	// WeakSet constructor
	vm.env.Define("WeakSet", &Value{Type: "native", Native: func(args []*Value) *Value {
		ws := &Value{
			Type:    "weakset",
			SetData: make(map[string]bool),
			Obj:     make(map[string]*Value),
		}
		ws.Obj["add"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && ws.SetData != nil {
				key := valueToString(args[0])
				ws.SetData[key] = true
			}
			return ws
		}}
		ws.Obj["has"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && ws.SetData != nil {
				key := valueToString(args[0])
				_, ok := ws.SetData[key]
				return &Value{Type: "bool", Bool: ok}
			}
			return &Value{Type: "bool", Bool: false}
		}}
		ws.Obj["delete"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) >= 1 && ws.SetData != nil {
				key := valueToString(args[0])
				delete(ws.SetData, key)
			}
			return &Value{Type: "undefined"}
		}}
		return ws
	}})
// setupSymbol adds Symbol constructor to the VM
}
func (vm *VM) setupSymbol() {
	symbolCounter := 0
	vm.env.Define("Symbol", &Value{Type: "native", Native: func(args []*Value) *Value {
		symbolCounter++
		desc := ""
		if len(args) > 0 {
			desc = valueToString(args[0])
		}
		return &Value{
			Type: "symbol",
			Str:  desc,
			Num:  float64(symbolCounter),
		}
	}})
}

// setupReflect adds Reflect API to the VM
func (vm *VM) setupReflect() {
	reflectObj := make(map[string]*Value)

	reflectObj["get"] = &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) >= 2 {
			target := args[0]
			key := valueToString(args[1])
			if target.Type == "object" && target.Obj != nil {
				if val, ok := target.Obj[key]; ok {
					return val
				}
			}
		}
		return &Value{Type: "undefined"}
	}}

	reflectObj["set"] = &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) >= 3 {
			target := args[0]
			key := valueToString(args[1])
			value := args[2]
			if target.Type == "object" && target.Obj != nil {
				target.Obj[key] = value
				return &Value{Type: "bool", Bool: true}
			}
		}
		return &Value{Type: "bool", Bool: false}
	}}

	reflectObj["has"] = &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) >= 2 {
			target := args[0]
			key := valueToString(args[1])
			if target.Type == "object" && target.Obj != nil {
				_, ok := target.Obj[key]
				return &Value{Type: "bool", Bool: ok}
			}
		}
		return &Value{Type: "bool", Bool: false}
	}}

	reflectObj["deleteProperty"] = &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) >= 2 {
			target := args[0]
			key := valueToString(args[1])
			if target.Type == "object" && target.Obj != nil {
				delete(target.Obj, key)
				return &Value{Type: "bool", Bool: true}
			}
		}
		return &Value{Type: "bool", Bool: false}
	}}

	reflectObj["ownKeys"] = &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) >= 1 {
			target := args[0]
			if target.Type == "object" && target.Obj != nil {
				keys := make([]*Value, 0, len(target.Obj))
				for k := range target.Obj {
					keys = append(keys, &Value{Type: "string", Str: k})
				}
				return &Value{Type: "object", Arr: keys}
			}
		}
		return &Value{Type: "object", Arr: []*Value{}}
	}}

	vm.env.Define("Reflect", &Value{Type: "object", Obj: reflectObj})
}

// setupProxy adds Proxy constructor to the VM
func (vm *VM) setupProxy() {
	// Proxy constructor: new Proxy(target, handler)
	vm.env.Define("Proxy", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) < 2 {
			return &Value{Type: "undefined"}
		}
		target := args[0]
		handlers := args[1]

		if target.Type != "object" || handlers.Type != "object" {
			return &Value{Type: "undefined"}
		}

		proxy := &Proxy{
			Target:   target,
			Handlers: handlers,
			Env:      vm.env,
		}

		return &Value{Type: "proxy", Proxy: proxy}
	}})
}


// evalProxyGet handles property access through a Proxy with get trap
func (vm *VM) evalProxyGet(proxy *Proxy, e *MemberExpr) *Value {
	// Get the property name
	var propName string
	if e.Computed {
		// Computed property: obj[expr]
		prop := vm.evalExpr(e.Property)
		propName = valueToString(prop)
	} else {
		// Dot notation: obj.prop
		if ident, ok := e.Property.(*Ident); ok {
			propName = ident.Name
		} else {
			propName = ""
		}
	}

	// Check if handler has a "get" trap
	if proxy.Handlers.Type == "object" {
		if getTrap, exists := proxy.Handlers.Obj["get"]; exists && getTrap != nil {
			// Call the get trap: handler.get(target, prop, receiver)
			propValue := &Value{Type: "string", Str: propName}
			receiver := &Value{Type: "proxy", Proxy: proxy}

			result := vm.callFunction(getTrap, []*Value{proxy.Target, propValue, receiver})
			return result
		}
	}

	// No get trap - forward to target directly
	target := proxy.Target
	if target.Type == "object" && target.Obj != nil {
		if val, exists := target.Obj[propName]; exists {
			return val
		}
	}

	// Check array access
	if target.Arr != nil {
		if propName == "length" {
			return &Value{Type: "number", Num: float64(len(target.Arr))}
		}
		// Try numeric index
		if idx, err := strconv.Atoi(propName); err == nil && idx >= 0 && idx < len(target.Arr) {
			return target.Arr[idx]
		}
	}

	return &Value{Type: "undefined"}
}

// evalProxySet handles property assignment through a Proxy with set trap
func (vm *VM) evalProxySet(proxy *Proxy, e *MemberExpr, val *Value) *Value {
	// Get the property name
	var propName string
	if e.Computed {
		prop := vm.evalExpr(e.Property)
		propName = valueToString(prop)
	} else {
		if ident, ok := e.Property.(*Ident); ok {
			propName = ident.Name
		} else {
			propName = ""
		}
	}

	// Check if handler has a "set" trap
	if proxy.Handlers.Type == "object" {
		if setTrap, exists := proxy.Handlers.Obj["set"]; exists && setTrap != nil {
			// Call the set trap: handler.set(target, prop, value, receiver)
			propValue := &Value{Type: "string", Str: propName}
			receiver := &Value{Type: "proxy", Proxy: proxy}

			vm.callFunction(setTrap, []*Value{proxy.Target, propValue, val, receiver})
			return val
		}
	}

	// No set trap - forward to target
	if proxy.Target.Type == "object" && proxy.Target.Obj != nil {
		proxy.Target.Obj[propName] = val
	}

	return val
}


// setupTimers adds setTimeout, setInterval, clearTimeout, clearInterval to the VM
func (vm *VM) setupTimers() {
	timerID := 0

	// setTimeout(callback, delay, ...args)
	setTimeout := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) < 2 {
			return &Value{Type: "number", Num: 0}
		}
		callback := args[0]
		delay := int64(0)
		if args[1].Type == "number" {
			delay = int64(args[1].Num)
		}
		timerID++
		task := timerTask{
			id:        timerID,
			callback:  callback,
			delay:     delay * 1e6, // ms to ns
			interval:  false,
		}
		task.executeAt = time.Now().UnixNano() + task.delay
		vm.pendingTimers = append(vm.pendingTimers, task)
		return &Value{Type: "number", Num: float64(timerID)}
	}}

	// setInterval(callback, delay, ...args)
	setInterval := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) < 2 {
			return &Value{Type: "number", Num: 0}
		}
		callback := args[0]
		delay := int64(0)
		if args[1].Type == "number" {
			delay = int64(args[1].Num)
		}
		timerID++
		task := timerTask{
			id:        timerID,
			callback:  callback,
			delay:     delay * 1e6, // ms to ns
			interval:  true,
		}
		task.executeAt = time.Now().UnixNano() + task.delay
		vm.pendingTimers = append(vm.pendingTimers, task)
		return &Value{Type: "number", Num: float64(timerID)}
	}}

	// clearTimeout(id)
	clearTimeout := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) > 0 && args[0].Type == "number" {
			id := int(args[0].Num)
			for i, t := range vm.pendingTimers {
				if t.id == id && !t.interval {
					vm.pendingTimers = append(vm.pendingTimers[:i], vm.pendingTimers[i+1:]...)
					break
				}
			}
		}
		return &Value{Type: "undefined"}
	}}

	// clearInterval(id)
	clearInterval := &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) > 0 && args[0].Type == "number" {
			id := int(args[0].Num)
			for i, t := range vm.pendingTimers {
				if t.id == id && t.interval {
					vm.pendingTimers = append(vm.pendingTimers[:i], vm.pendingTimers[i+1:]...)
					break
				}
			}
		}
		return &Value{Type: "undefined"}
	}}

	// Define on global and window
	vm.env.Define("setTimeout", setTimeout)
	vm.env.Define("setInterval", setInterval)
	vm.env.Define("clearTimeout", clearTimeout)
	vm.env.Define("clearInterval", clearInterval)

	// Also add to window object if it exists
	if window := vm.env.Get("window"); window.Type == "object" && window.Obj != nil {
		window.Obj["setTimeout"] = setTimeout
		window.Obj["setInterval"] = setInterval
		window.Obj["clearTimeout"] = clearTimeout
		window.Obj["clearInterval"] = clearInterval
	}
}

// RunTimers executes all pending timers that are due.
// Returns the number of timers still pending.
func (vm *VM) RunTimers() int {
	now := time.Now().UnixNano()
	var remaining []timerTask

	for _, task := range vm.pendingTimers {
		if task.executeAt <= now {
			// Execute the callback
			if task.callback != nil {
				if task.callback.Type == "function" && task.callback.Func != nil {
					vm.callFunction(task.callback, []*Value{})
				} else if task.callback.Type == "native" && task.callback.Native != nil {
					task.callback.Native([]*Value{})
				}
			}
			// Check if this timer was cleared during callback execution
			// by checking if it still exists in pendingTimers
			stillExists := false
			for _, t := range vm.pendingTimers {
				if t.id == task.id {
					stillExists = true
					break
				}
			}
			// If interval and still exists (not cleared), reschedule
			if task.interval && stillExists {
				task.executeAt = now + task.delay
				remaining = append(remaining, task)
			}
		} else {
			remaining = append(remaining, task)
		}
	}
	vm.pendingTimers = remaining
	return len(vm.pendingTimers)
}

// WaitForTimers blocks until all timers complete or timeout is reached.
// This is useful for implementing waitStable.
func (vm *VM) WaitForTimers(timeoutMs int) {
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	for time.Now().Before(deadline) {
		pending := vm.RunTimers()
		if pending == 0 {
			break
		}
		// Sleep a short time before checking again
		time.Sleep(10 * time.Millisecond)
	}
}

// HasPendingTimers returns true if there are pending timers.
func (vm *VM) HasPendingTimers() bool {
	return len(vm.pendingTimers) > 0
}
