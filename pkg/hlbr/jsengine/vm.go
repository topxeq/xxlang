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
	// NodeRef holds a reference to a dom.Node for wrapped DOM nodes.
	NodeRef *dom.Node
	// Proto is the prototype link for prototype chain lookup.
	// When a property is not found on this value, the VM walks up the Proto chain.
	Proto *Value
	// Frozen indicates if the object is frozen (Object.freeze)
	Frozen bool
	// PrototypeObj is the function's prototype property (for constructor functions).
	// When `new Fn()` is called, the new object's Proto is set to Fn.PrototypeObj.
	PrototypeObj *Value
	// _isThisArg is an internal marker: when true, this Value was prepended
	// as the 'this' argument in a native method call by evalCall.
	_isThisArg bool
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
	// Execution limit to prevent infinite loops
	stepCount    int64         // current step count, incremented on each eval/exec
	maxSteps     int64         // maximum steps before timeout error (0 = unlimited)
	timeoutAt    int64         // Unix nano timestamp for wall-clock timeout (0 = none)
	maxTimeoutMs int64         // default timeout in milliseconds (0 = none)
	// Recursion depth limit to prevent stack overflow
	callDepth    int           // current function call depth
	maxCallDepth int           // maximum function call depth (0 = unlimited)
	// Memory allocation tracking to prevent memory exhaustion
	allocCount   int64         // number of objects/arrays allocated
	maxAllocs    int64         // maximum allocations before error (0 = unlimited)
	// Abort flag for external cancellation
	abortFlag    bool          // set to true to abort execution
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
		maxSteps:       10_000_000, // default: 10M steps
		maxTimeoutMs:   30_000,     // default: 30 seconds wall-clock timeout
		maxCallDepth:   500,        // default: 500 levels of recursion
		maxAllocs:      1_000_000,  // default: 1M object allocations
	}
	if doc != nil {
		vm.root = doc.Root
	}
	vm.env = NewEnvironment(nil)
	vm.setupPrototypes() // Call setupPrototypes first to define prototype objects
	vm.setupBuiltins()
	vm.setupPromise()
	vm.setupTimers()
	vm.setupProxy()
	vm.setupMapSet()
	vm.setupSymbol()
	vm.setupReflect()
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

// SetMaxSteps sets the maximum number of execution steps before timeout.
// Set to 0 for unlimited (not recommended).
func (vm *VM) SetMaxSteps(max int64) {
	vm.maxSteps = max
}

// SetTimeoutMs sets the wall-clock timeout in milliseconds.
// Set to 0 for no timeout.
func (vm *VM) SetTimeoutMs(ms int64) {
	vm.maxTimeoutMs = ms
}

// SetMaxCallDepth sets the maximum recursion depth.
// Set to 0 for unlimited (not recommended, risk of stack overflow).
func (vm *VM) SetMaxCallDepth(depth int) {
	vm.maxCallDepth = depth
}

// SetMaxAllocs sets the maximum number of object allocations.
// Set to 0 for unlimited (not recommended, risk of memory exhaustion).
func (vm *VM) SetMaxAllocs(max int64) {
	vm.maxAllocs = max
}

// Abort sets the abort flag to stop execution from outside.
// This can be called from another goroutine to cancel a running script.
func (vm *VM) Abort() {
	vm.abortFlag = true
}

// IsAborted returns true if the abort flag has been set.
func (vm *VM) IsAborted() bool {
	return vm.abortFlag
}

// trackAlloc increments the allocation counter and checks the limit.
// Call this when creating new objects, arrays, or other heap-allocated values.
func (vm *VM) trackAlloc() {
	vm.allocCount++
	if vm.maxAllocs > 0 && vm.allocCount > vm.maxAllocs {
		ThrowJS(NewError("RangeError", fmt.Sprintf("Memory limit: exceeded %d allocations", vm.maxAllocs)))
	}
}

// enterCall increments the call depth and checks the limit.
// Call this when entering a function call.
func (vm *VM) enterCall() {
	vm.callDepth++
	if vm.maxCallDepth > 0 && vm.callDepth > vm.maxCallDepth {
		ThrowJS(NewError("RangeError", fmt.Sprintf("Recursion limit: exceeded %d call depth", vm.maxCallDepth)))
	}
}

// exitCall decrements the call depth.
// Call this when exiting a function call.
func (vm *VM) exitCall() {
	if vm.callDepth > 0 {
		vm.callDepth--
	}
}

// checkSteps increments the step counter and checks for timeout conditions.
// It panics with a JS exception if the step limit or wall-clock timeout is exceeded.
// This is called from execStmt, execBlock, and evalExpr to prevent infinite loops.
func (vm *VM) checkSteps() {
	vm.stepCount++
	// Check abort flag
	if vm.abortFlag {
		ThrowJS(NewError("AbortError", "Execution aborted by external request"))
	}
	// Check step limit
	if vm.maxSteps > 0 && vm.stepCount > vm.maxSteps {
		ThrowJS(NewError("RangeError", fmt.Sprintf("Execution timeout: exceeded %d steps", vm.maxSteps)))
	}
	// Check wall-clock timeout (only every 1000 steps to reduce overhead)
	if vm.maxTimeoutMs > 0 && vm.stepCount%1000 == 0 {
		if vm.timeoutAt == 0 {
			vm.timeoutAt = time.Now().UnixNano() + vm.maxTimeoutMs*1e6
		}
		if time.Now().UnixNano() > vm.timeoutAt {
			ThrowJS(NewError("RangeError", fmt.Sprintf("Execution timeout: exceeded %dms wall-clock time", vm.maxTimeoutMs)))
		}
	}
}

// ResetSteps resets the step counter and timeout deadline.
// Call this before executing a new script.
func (vm *VM) ResetSteps() {
	vm.stepCount = 0
	vm.timeoutAt = 0
	vm.callDepth = 0
	vm.allocCount = 0
	vm.abortFlag = false
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
		"documentElement": vm.wrapNode(vm.root),
		"createElement": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "null"}
			}
			tag := args[0].Str
			newNode := &dom.Node{Type: dom.ElementNode, Data: strings.ToLower(tag)}
			return vm.wrapNode(newNode)
		}},
		"createTextNode": {Type: "native", Native: func(args []*Value) *Value {
			text := ""
			if len(args) > 0 {
				text = args[0].Str
			}
			newNode := &dom.Node{Type: dom.TextNode, Data: text}
			return vm.wrapNode(newNode)
		}},
		"createDocumentFragment": {Type: "native", Native: func(args []*Value) *Value {
			newNode := &dom.Node{Type: dom.DocumentNode, Data: "#document-fragment"}
			return vm.wrapNode(newNode)
		}},
		"createComment": {Type: "native", Native: func(args []*Value) *Value {
			text := ""
			if len(args) > 0 {
				text = args[0].Str
			}
			newNode := &dom.Node{Type: dom.CommentNode, Data: text}
			return vm.wrapNode(newNode)
		}},
		"addEventListener": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: event listener registration (events not yet dispatched)
			return &Value{Type: "undefined"}
		}},
		"removeEventListener": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: event listener removal
			return &Value{Type: "undefined"}
		}},
		"cookie": {Type: "string", Str: ""},
		"readyState": {Type: "string", Str: "complete"},
		"implementation": {Type: "object", Obj: map[string]*Value{
			"createHTMLDocument": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "object", Obj: map[string]*Value{}}
			}},
		}},
	}})

	// Create window object first (as a shell) - the global object in browsers
	windowObj := &Value{Type: "object", Obj: map[string]*Value{
		"location": {Type: "object", Obj: map[string]*Value{
			"href": {Type: "string", Str: ""},
			"protocol": {Type: "string", Str: "http:"},
			"host": {Type: "string", Str: ""},
			"hostname": {Type: "string", Str: ""},
			"pathname": {Type: "string", Str: "/"},
			"search": {Type: "string", Str: ""},
			"hash": {Type: "string", Str: ""},
		}},
		"navigator": {Type: "object", Obj: map[string]*Value{
			"userAgent": {Type: "string", Str: "Xxlang-HLBR/1.0"},
			"language": {Type: "string", Str: "en-US"},
			"platform": {Type: "string", Str: "Xxlang"},
			"cookieEnabled": {Type: "bool", Bool: true},
		}},
		"innerWidth":  {Type: "number", Num: 1024},
		"innerHeight": {Type: "number", Num: 768},
		"outerWidth":  {Type: "number", Num: 1024},
		"outerHeight": {Type: "number", Num: 768},
		"screenX":     {Type: "number", Num: 0},
		"screenY":     {Type: "number", Num: 0},
		"pageXOffset": {Type: "number", Num: 0},
		"pageYOffset": {Type: "number", Num: 0},
		"scrollX":     {Type: "number", Num: 0},
		"scrollY":     {Type: "number", Num: 0},
		"devicePixelRatio": {Type: "number", Num: 1},
		"name":        {Type: "string", Str: ""},
		"origin":      {Type: "string", Str: ""},
		"parent":      &Value{Type: "undefined"}, // will be set to self below
		"top":         &Value{Type: "undefined"}, // will be set to self below
		"frames":      &Value{Type: "undefined"}, // will be set to self below
		"self":        &Value{Type: "undefined"}, // will be set to self below
		"opener":      &Value{Type: "null"},
	}}
	vm.env.Define("window", windowObj)
	// Also define 'this' and 'self' to point to window (for browser compatibility in global scope)
	vm.env.Define("this", windowObj)
	vm.env.Define("self", windowObj)

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
			return &Value{Type: "string", Str: jsonStringify(args[0])}
		}},
		"parse": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "undefined"}
			}
			return jsonParse(args[0].Str)
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
	// Add static methods to the Array constructor
	arrayConstructor.Obj = map[string]*Value{
		"isArray": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "bool", Bool: false}
			}
			return &Value{Type: "bool", Bool: args[0].Type == "object" && args[0].Arr != nil}
		}},
		"from": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 1 {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			src := args[0]
			// Direct array
			if src.Type == "object" && src.Arr != nil {
				// Apply map function if provided
				if len(args) > 1 && args[1].Type == "native" {
					result := make([]*Value, len(src.Arr))
					for i, v := range src.Arr {
						result[i] = args[1].Native([]*Value{v, {Type: "number", Num: float64(i)}})
					}
					return &Value{Type: "object", Arr: result}
				}
				return src
			}
			// String
			if src.Type == "string" {
				arr := make([]*Value, 0, len(src.Str))
				for _, ch := range src.Str {
					arr = append(arr, &Value{Type: "string", Str: string(ch)})
				}
				return &Value{Type: "object", Arr: arr}
			}
			// Check for iterator protocol
			if src.Type == "object" && src.Obj != nil {
				iteratorSymbol := vm.env.Get("_symbolIterator")
				if iterMethod, ok := src.Obj[valueToString(iteratorSymbol)]; ok && iterMethod.Type == "native" {
					iterator := iterMethod.Native(nil)
					if iterator.Type == "object" && iterator.Obj != nil {
						if nextFn, ok := iterator.Obj["next"]; ok && nextFn.Type == "native" {
							arr := make([]*Value, 0)
							for {
								result := nextFn.Native(nil)
								if result.Type == "object" && result.Obj != nil {
									if done, ok := result.Obj["done"]; ok && done.Bool {
										break
									}
									value := result.Obj["value"]
									if value == nil {
										value = &Value{Type: "undefined"}
									}
									arr = append(arr, value)
								} else {
									break
								}
							}
							// Apply map function if provided
							if len(args) > 1 && args[1].Type == "native" {
								for i, v := range arr {
									arr[i] = args[1].Native([]*Value{v, {Type: "number", Num: float64(i)}})
								}
							}
							return &Value{Type: "object", Arr: arr}
						}
					}
				}
				// Object values
				arr := make([]*Value, 0, len(src.Obj))
				for _, v := range src.Obj {
					arr = append(arr, v)
				}
				return &Value{Type: "object", Arr: arr}
			}
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"of": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: args}
		}},
	}
	vm.env.Define("Array", arrayConstructor)

	vm.env.Define("Object", &Value{Type: "object", Obj: GetObjectMethods(vm)})

	// Set Object.prototype to the ObjectPrototype
	objectProto := vm.env.Get("ObjectPrototype")
	if objectProto.Type == "object" {
		objectVal := vm.env.Get("Object")
		if objectVal.Type == "object" && objectVal.Obj != nil {
			objectVal.Obj["prototype"] = objectProto
		}
	}

	// Define Function constructor with prototype
	functionProto := vm.env.Get("FunctionPrototype")
	functionConstructor := &Value{Type: "native", Native: func(args []*Value) *Value {
		// Function constructor - create a new function
		return &Value{Type: "function", Func: &Function{Params: []string{}, Body: []Statement{}, Env: vm.env}}
	}}
	if functionProto.Type == "object" {
		functionConstructor.Obj = map[string]*Value{
			"prototype": functionProto,
		}
	}
	vm.env.Define("Function", functionConstructor)

	// Set Array.prototype to the ArrayPrototype
	arrayProto := vm.env.Get("ArrayPrototype")
	if arrayProto.Type == "object" {
		arrayObj := vm.env.Get("Array")
		if arrayObj.Type == "object" && arrayObj.Obj != nil {
			arrayObj.Obj["prototype"] = arrayProto
		}
	}

	// Set Function.prototype to the FunctionPrototype
	// Already handled above when defining Function constructor

	// Set String.prototype to the StringPrototype
	stringProto := vm.env.Get("StringPrototype")
	if stringProto.Type == "object" {
		stringConstructor := vm.env.Get("String")
		if stringConstructor.Type == "native" {
			if stringConstructor.Obj == nil {
				stringConstructor.Obj = make(map[string]*Value)
			}
			stringConstructor.Obj["prototype"] = stringProto
		}
	}

	// Set Number.prototype to the NumberPrototype
	numberProto := vm.env.Get("NumberPrototype")
	if numberProto.Type == "object" {
		numberConstructor := vm.env.Get("Number")
		if numberConstructor.Type == "native" {
			if numberConstructor.Obj == nil {
				numberConstructor.Obj = make(map[string]*Value)
			}
			numberConstructor.Obj["prototype"] = numberProto
		}
	}

	// Set Boolean.prototype to the BooleanPrototype
	booleanProto := vm.env.Get("BooleanPrototype")
	if booleanProto.Type == "object" {
		booleanConstructor := vm.env.Get("Boolean")
		if booleanConstructor.Type == "native" {
			if booleanConstructor.Obj == nil {
				booleanConstructor.Obj = make(map[string]*Value)
			}
			booleanConstructor.Obj["prototype"] = booleanProto
		}
	}

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

	// Wire up direct references from window to builtins (not native functions)
	// This must happen after all builtins are defined so we can reference them directly
	if windowObj.Obj != nil {
		windowObj.Obj["document"] = vm.env.Get("document")
		windowObj.Obj["console"] = vm.env.Get("console")
		windowObj.Obj["Math"] = vm.env.Get("Math")
		windowObj.Obj["localStorage"] = vm.env.Get("localStorage")
		windowObj.Obj["sessionStorage"] = vm.env.Get("sessionStorage")
		windowObj.Obj["JSON"] = vm.env.Get("JSON")
		windowObj.Obj["Array"] = vm.env.Get("Array")
		windowObj.Obj["Object"] = vm.env.Get("Object")
		windowObj.Obj["String"] = vm.env.Get("String")
		windowObj.Obj["Number"] = vm.env.Get("Number")
		windowObj.Obj["Boolean"] = vm.env.Get("Boolean")
		windowObj.Obj["RegExp"] = vm.env.Get("RegExp")
		windowObj.Obj["Error"] = vm.env.Get("Error")
		windowObj.Obj["TypeError"] = vm.env.Get("TypeError")
		windowObj.Obj["ReferenceError"] = vm.env.Get("ReferenceError")
		windowObj.Obj["SyntaxError"] = vm.env.Get("SyntaxError")
		windowObj.Obj["RangeError"] = vm.env.Get("RangeError")
		windowObj.Obj["Promise"] = vm.env.Get("Promise")
		windowObj.Obj["Map"] = vm.env.Get("Map")
		windowObj.Obj["Set"] = vm.env.Get("Set")
		windowObj.Obj["WeakMap"] = vm.env.Get("WeakMap")
		windowObj.Obj["WeakSet"] = vm.env.Get("WeakSet")
		windowObj.Obj["Symbol"] = vm.env.Get("Symbol")
		windowObj.Obj["Reflect"] = vm.env.Get("Reflect")
		windowObj.Obj["Proxy"] = vm.env.Get("Proxy")
		windowObj.Obj["parseInt"] = vm.env.Get("parseInt")
		windowObj.Obj["parseFloat"] = vm.env.Get("parseFloat")
		windowObj.Obj["isNaN"] = vm.env.Get("isNaN")
		windowObj.Obj["isFinite"] = vm.env.Get("isFinite")
		windowObj.Obj["undefined"] = vm.env.Get("undefined")
		windowObj.Obj["NaN"] = vm.env.Get("NaN")
		windowObj.Obj["Infinity"] = vm.env.Get("Infinity")
		windowObj.Obj["setTimeout"] = vm.env.Get("setTimeout")
		windowObj.Obj["setInterval"] = vm.env.Get("setInterval")
		windowObj.Obj["clearTimeout"] = vm.env.Get("clearTimeout")
		windowObj.Obj["clearInterval"] = vm.env.Get("clearInterval")
		windowObj.Obj["self"] = windowObj
		windowObj.Obj["parent"] = windowObj
		windowObj.Obj["top"] = windowObj
		windowObj.Obj["frames"] = windowObj
		windowObj.Obj["window"] = windowObj
	}
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
		"nodeType":    {Type: "number", Num: float64(n.Type)},
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
		"removeAttribute": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) > 0 {
				n.RemoveAttribute(args[0].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"hasAttribute": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "bool", Bool: false}
			}
			return &Value{Type: "bool", Bool: n.HasAttribute(args[0].Str)}
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
		"parentElement": {Type: "native", Native: func(args []*Value) *Value {
			if n.Parent != nil && n.Parent.Type == dom.ElementNode {
				return vm.wrapNodeShallow(n.Parent)
			}
			return &Value{Type: "null"}
		}},
		"nextSibling": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(n.NextSibling)
		}},
		"previousSibling": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(n.PrevSibling)
		}},
		"nextElementSibling": {Type: "native", Native: func(args []*Value) *Value {
			sibling := n.NextSibling
			for sibling != nil && sibling.Type != dom.ElementNode {
				sibling = sibling.NextSibling
			}
			return vm.wrapNode(sibling)
		}},
		"previousElementSibling": {Type: "native", Native: func(args []*Value) *Value {
			sibling := n.PrevSibling
			for sibling != nil && sibling.Type != dom.ElementNode {
				sibling = sibling.PrevSibling
			}
			return vm.wrapNode(sibling)
		}},
		// DOM mutation methods
		"appendChild": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "undefined"}
			}
			child := args[0]
			// Extract the dom.Node from the wrapped value
			if childNode := vm.unwrapNode(child); childNode != nil {
				n.AppendChild(childNode)
				return child
			}
			return &Value{Type: "undefined"}
		}},
		"removeChild": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "undefined"}
			}
			child := args[0]
			if childNode := vm.unwrapNode(child); childNode != nil {
				for i, c := range n.Children {
					if c == childNode {
						n.Children = append(n.Children[:i], n.Children[i+1:]...)
						break
					}
				}
				return child
			}
			return &Value{Type: "undefined"}
		}},
		"insertBefore": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 2 {
				return args[0]
			}
			newNode := args[0]
			refNode := args[1]
			newDomNode := vm.unwrapNode(newNode)
			refDomNode := vm.unwrapNode(refNode)
			if newDomNode != nil {
				for i, c := range n.Children {
					if c == refDomNode {
						newDomNode.Parent = n
						n.Children = append(n.Children[:i], append([]*dom.Node{newDomNode}, n.Children[i:]...)...)
						break
					}
				}
				return newNode
			}
			return &Value{Type: "undefined"}
		}},
		"replaceChild": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) < 2 {
				return &Value{Type: "undefined"}
			}
			newNode := args[0]
			oldNode := args[1]
			newDomNode := vm.unwrapNode(newNode)
			oldDomNode := vm.unwrapNode(oldNode)
			if newDomNode != nil && oldDomNode != nil {
				for i, c := range n.Children {
					if c == oldDomNode {
						newDomNode.Parent = n
						n.Children[i] = newDomNode
						break
					}
				}
				return oldNode
			}
			return &Value{Type: "undefined"}
		}},
		"cloneNode": {Type: "native", Native: func(args []*Value) *Value {
			deep := false
			if len(args) > 0 {
				deep = args[0].Bool
			}
			clone := vm.cloneNode(n, deep)
			return vm.wrapNode(clone)
		}},
		"contains": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 {
				return &Value{Type: "bool", Bool: false}
			}
			other := vm.unwrapNode(args[0])
			return &Value{Type: "bool", Bool: vm.nodeContains(n, other)}
		}},
		"addEventListener": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: event listener registration
			return &Value{Type: "undefined"}
		}},
		"removeEventListener": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: event listener removal
			return &Value{Type: "undefined"}
		}},
		"dispatchEvent": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: event dispatch
			return &Value{Type: "bool", Bool: true}
		}},
		"focus": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: focus
			return &Value{Type: "undefined"}
		}},
		"blur": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: blur
			return &Value{Type: "undefined"}
		}},
		"click": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: click
			return &Value{Type: "undefined"}
		}},
		"scrollIntoView": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: scrollIntoView
			return &Value{Type: "undefined"}
		}},
		"getBoundingClientRect": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: returns a DOMRect-like object
			return &Value{Type: "object", Obj: map[string]*Value{
				"x":      {Type: "number", Num: 0},
				"y":      {Type: "number", Num: 0},
				"width":  {Type: "number", Num: 0},
				"height": {Type: "number", Num: 0},
				"top":    {Type: "number", Num: 0},
				"right":  {Type: "number", Num: 0},
				"bottom": {Type: "number", Num: 0},
				"left":   {Type: "number", Num: 0},
			}}
		}},
		"getClientRects": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "object", Arr: []*Value{}}
		}},
		"closest": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: closest selector
			return &Value{Type: "null"}
		}},
		"matches": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: matches selector
			return &Value{Type: "bool", Bool: false}
		}},
		"classList": {Type: "object", Obj: map[string]*Value{
			"add": {Type: "native", Native: func(args []*Value) *Value {
				// Stub: add class
				return &Value{Type: "undefined"}
			}},
			"remove": {Type: "native", Native: func(args []*Value) *Value {
				// Stub: remove class
				return &Value{Type: "undefined"}
			}},
			"toggle": {Type: "native", Native: func(args []*Value) *Value {
				// Stub: toggle class
				return &Value{Type: "bool", Bool: false}
			}},
			"contains": {Type: "native", Native: func(args []*Value) *Value {
				// Stub: contains class
				return &Value{Type: "bool", Bool: false}
			}},
		}},
		"style": {Type: "object", Obj: make(map[string]*Value)},
		"dataset": {Type: "object", Obj: make(map[string]*Value)},
		"checked": {Type: "bool", Bool: false},
		"disabled": {Type: "bool", Bool: false},
		"selected": {Type: "bool", Bool: false},
		"value": {Type: "string", Str: n.GetAttribute("value")},
		"type": {Type: "string", Str: n.GetAttribute("type")},
		"name": {Type: "string", Str: n.GetAttribute("name")},
		"placeholder": {Type: "string", Str: n.GetAttribute("placeholder")},
		"src": {Type: "string", Str: n.GetAttribute("src")},
		"href": {Type: "string", Str: n.GetAttribute("href")},
		"action": {Type: "string", Str: n.GetAttribute("action")},
		"method": {Type: "string", Str: n.GetAttribute("method")},
		"tabIndex": {Type: "number", Num: 0},
	}}

	// Add descriptors for innerHTML and textContent with getters/setters
	obj.Descriptors = map[string]*PropertyDescriptor{
		"innerHTML":   {Get: innerHTMLGetter, Set: innerHTMLSetter, Enumerable: true, Configurable: true},
		"textContent": {Get: textContentGetter, Set: textContentSetter, Enumerable: true, Configurable: true},
	}
	// Store the dom.Node reference for unwrapNode (used by DOM mutation methods)
	obj.NodeRef = n

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
	// Store the dom.Node reference for unwrapNode (used by DOM mutation methods)
	obj.NodeRef = n

	return obj
}

// newArray creates a new array Value with the ArrayPrototype link.
func (vm *VM) newArray(elements []*Value) *Value {
	arr := &Value{Type: "object", Arr: elements}
	if proto := vm.env.Get("ArrayPrototype"); proto.Type == "object" {
		arr.Proto = proto
	}
	return arr
}

// newFunction creates a new function Value with the FunctionPrototype link.
func (vm *VM) newFunction(fn *Function) *Value {
	val := &Value{Type: "function", Func: fn}
	if proto := vm.env.Get("FunctionPrototype"); proto.Type == "object" {
		val.Proto = proto
	}
	return val
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

// unwrapNode extracts a dom.Node from a JS wrapped node value.
// Returns nil if the value is not a wrapped DOM node.
func (vm *VM) unwrapNode(v *Value) *dom.Node {
	if v == nil || v.Type != "object" {
		return nil
	}
	return v.NodeRef
}

// nodePtrValue stores a dom.Node pointer as a JS value for internal use.
type nodePtrValue struct {
	Node *dom.Node
}

// cloneNode creates a deep or shallow copy of a DOM node.
func (vm *VM) cloneNode(n *dom.Node, deep bool) *dom.Node {
	if n == nil {
		return nil
	}
	clone := &dom.Node{
		Type: n.Type,
		Data: n.Data,
		Attr: make([]dom.Attribute, len(n.Attr)),
	}
	copy(clone.Attr, n.Attr)
	if deep {
		for _, child := range n.Children {
			childClone := vm.cloneNode(child, true)
			childClone.Parent = clone
			clone.Children = append(clone.Children, childClone)
		}
	}
	return clone
}

// nodeContains checks if parent contains child in the DOM tree.
func (vm *VM) nodeContains(parent, child *dom.Node) bool {
	if child == nil || parent == nil {
		return false
	}
	if child == parent {
		return true
	}
	for _, c := range parent.Children {
		if vm.nodeContains(c, child) {
			return true
		}
	}
	return false
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

	vm.ResetSteps() // reset step counter for each Run call

	parser := NewParser(code)
	prog := parser.Parse()

	vm.debugLog("Parsed %d statements", len(prog.Statements))

	var result *Value = &Value{Type: "undefined"}
	var returnVal *Value
	var returning bool

	// Wrap execution in a recovery function to catch step timeout panics
	var execErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case *JSException:
					execErr = fmt.Errorf("JS error: %s", valueToString(v.Value))
				default:
					execErr = fmt.Errorf("execution error: %v", r)
				}
			}
		}()

		for i, stmt := range prog.Statements {
			if returning {
				break
			}
			vm.debugLog("Executing statement %d: %T", i+1, stmt)
			result, returning, returnVal = vm.execStmt(stmt)
		}
	}()

	if execErr != nil {
		return &Value{Type: "undefined"}, execErr
	}

	if returning && returnVal != nil {
		vm.debugLog("Returning: %s", valueToString(returnVal))
		return returnVal, nil
	}
	vm.debugLog("=== JS execution complete ===")
	return result, nil
}

func (vm *VM) execStmt(stmt Statement) (*Value, bool, *Value) {
	vm.checkSteps()
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
			vm.checkSteps() // Check for infinite loop in for-statement
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
		if obj.Type == "object" {
			// Collect all enumerable keys from Obj and Descriptors
			keys := make(map[string]bool)
			if obj.Obj != nil {
				for key := range obj.Obj {
					keys[key] = true
				}
			}
			if obj.Descriptors != nil {
				for key, desc := range obj.Descriptors {
					if desc.Enumerable {
						keys[key] = true
					}
				}
			}
			for key := range keys {
				vm.checkSteps()
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
		// Check if object has Symbol.iterator
		if obj.Type == "object" && obj.Obj != nil {
			iteratorSymbol := vm.env.Get("_symbolIterator")
			if iterMethod, ok := obj.Obj[valueToString(iteratorSymbol)]; ok && iterMethod.Type == "native" {
				// Use iterator protocol
				iterator := iterMethod.Native(nil)
				if iterator.Type == "object" && iterator.Obj != nil {
					if nextFn, ok := iterator.Obj["next"]; ok && nextFn.Type == "native" {
						for {
							vm.checkSteps()
							result := nextFn.Native(nil)
							if result.Type == "object" && result.Obj != nil {
								if done, ok := result.Obj["done"]; ok && done.Bool {
									break
								}
								value := result.Obj["value"]
								if value == nil {
									value = &Value{Type: "undefined"}
								}
								childEnv := NewEnvironment(vm.env)
								childEnv.Define(s.VarName, value)
								oldEnv := vm.env
								vm.env = childEnv
								ret, brk, retVal := vm.execBlock(s.Body)
								vm.env = oldEnv
								if brk {
									return ret, true, retVal
								}
							} else {
								break
							}
						}
						return &Value{Type: "undefined"}, false, nil
					}
				}
			}
		}
		// Fallback to array iteration
		if obj.Type == "object" && obj.Arr != nil {
			for _, val := range obj.Arr {
				vm.checkSteps() // Check for infinite loop in for-of-statement
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
				vm.checkSteps() // Check for infinite loop in for-of-statement
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
		} else if obj.Type == "string" {
			// String iteration
			for _, ch := range obj.Str {
				vm.checkSteps()
				childEnv := NewEnvironment(vm.env)
				childEnv.Define(s.VarName, &Value{Type: "string", Str: string(ch)})
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
			vm.checkSteps() // Check for infinite loop in while-statement
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
		// Create prototype object for the constructor function
		protoObj := &Value{Type: "object", Obj: make(map[string]*Value)}
		if objProto := vm.env.Get("ObjectPrototype"); objProto.Type == "object" {
			protoObj.Proto = objProto
		}
		// Store constructor reference on prototype
		protoObj.Obj["constructor"] = val
		val.PrototypeObj = protoObj
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
		vm.checkSteps()
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
	vm.checkSteps()
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
		vm.trackAlloc() // Track array allocation
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
		result := &Value{Type: "object", Arr: arr}
		// Set prototype link for array methods
		if proto := vm.env.Get("ArrayPrototype"); proto.Type == "object" {
			result.Proto = proto
		}
		return result
	case *ObjectLit:
		vm.trackAlloc() // Track object allocation
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
		result := &Value{Type: "object", Obj: obj}
		// Set prototype link to Object.prototype
		if proto := vm.env.Get("ObjectPrototype"); proto.Type == "object" {
			result.Proto = proto
		}
		return result
	case *FunctionExpr:
		fn := &Function{Params: e.Params, DefaultVals: e.DefaultVals, RestParam: e.RestParam, Body: e.Body, Env: vm.env}
		result := &Value{Type: "function", Func: fn}
		if proto := vm.env.Get("FunctionPrototype"); proto.Type == "object" {
			result.Proto = proto
		}
		// Create prototype object for the constructor function
		protoObj := &Value{Type: "object", Obj: make(map[string]*Value)}
		if objProto := vm.env.Get("ObjectPrototype"); objProto.Type == "object" {
			protoObj.Proto = objProto
		}
		protoObj.Obj["constructor"] = result
		result.PrototypeObj = protoObj
		return result
	case *ArrowFunctionExpr:
		fn := &Function{Params: e.Params, RestParam: e.RestParam, Body: e.Body, Env: vm.env}
		result := &Value{Type: "function", Func: fn}
		if proto := vm.env.Get("FunctionPrototype"); proto.Type == "object" {
			result.Proto = proto
		}
		return result
	case *TernaryExpr:
		cond := vm.evalExpr(e.Cond)
		if isTruthy(cond) {
			return vm.evalExpr(e.True)
		}
		return vm.evalExpr(e.False)
	case *TypeOfExpr:
		v := vm.evalExpr(e.Expr)
		// In JavaScript, typeof null should return "object"
		if v.Type == "null" {
			return &Value{Type: "string", Str: "object"}
		}
		if v.Type == "native" || v.Type == "function" {
			return &Value{Type: "string", Str: "function"}
		}
		// Map internal types to JS typeof strings
		switch v.Type {
		case "arrayMethod", "stringMethod":
			return &Value{Type: "string", Str: "function"}
		case "bigint":
			return &Value{Type: "string", Str: "bigint"}
		case "regexp":
			return &Value{Type: "string", Str: "object"}
		case "promise", "proxy", "map", "set", "weakmap", "weakset", "iterator", "generator":
			return &Value{Type: "string", Str: "object"}
		}
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
	// Handle nil right side (e.g., from incomplete expressions)
	if e.Right == nil {
		return &Value{Type: "undefined"}
	}
	val := vm.evalExpr(e.Right)

	switch left := e.Left.(type) {
	case *Ident:
		switch e.Op {
		case "=":
			vm.env.Set(left.Name, val)
		case "+=":
			cur := vm.env.Get(left.Name)
			if cur.Type == "string" || val.Type == "string" {
				vm.env.Set(left.Name, &Value{Type: "string", Str: valueToString(cur) + valueToString(val)})
			} else {
				vm.env.Set(left.Name, &Value{Type: "number", Num: cur.Num + val.Num})
			}
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
			// Check if object is frozen
			if obj.Frozen {
				// In strict mode, this would throw an error
				// For now, just return the value without modifying
				return val
			}
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

// evalMember evaluates a member expression (property access).

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
	// Convention: only prepend 'this' if the callee has a ThisBinding set by
	// evalMember (via prototype chain lookup). Do NOT prepend this for regular
	// native function calls where thisBinding comes from the call expression,
	// as that would break built-in functions like setTimeout.
	if callee.Type == "native" && callee.Native != nil {
		nativeArgs := args
		if callee.ThisBinding != nil {
			thisVal := &Value{
				Type:        callee.ThisBinding.Type,
				Num:         callee.ThisBinding.Num,
				Str:         callee.ThisBinding.Str,
				Bool:        callee.ThisBinding.Bool,
				Obj:         callee.ThisBinding.Obj,
				Arr:         callee.ThisBinding.Arr,
				Func:        callee.ThisBinding.Func,
				Class:       callee.ThisBinding.Class,
				Promise:     callee.ThisBinding.Promise,
				Proxy:       callee.ThisBinding.Proxy,
				MapData:     callee.ThisBinding.MapData,
				SetData:     callee.ThisBinding.SetData,
				Native:      callee.ThisBinding.Native,
				ThisBinding: callee.ThisBinding.ThisBinding,
				Descriptors: callee.ThisBinding.Descriptors,
				IsAsync:     callee.ThisBinding.IsAsync,
				BuiltInConstructor: callee.ThisBinding.BuiltInConstructor,
				NodeRef:     callee.ThisBinding.NodeRef,
				Proto:       callee.ThisBinding.Proto,
				Frozen:      callee.ThisBinding.Frozen,
				_isThisArg:  true,
			}
			nativeArgs = make([]*Value, 0, len(args)+1)
			nativeArgs = append(nativeArgs, thisVal)
			nativeArgs = append(nativeArgs, args...)
		}
		return callee.Native(nativeArgs)
	}

	// Handle array methods (type set in evalMember)
	if callee.Type == "arrayMethod" && callee.Native != nil {
		nativeArgs := args
		if callee.ThisBinding != nil {
			thisVal := &Value{
				Type:        callee.ThisBinding.Type,
				Arr:         callee.ThisBinding.Arr,
				Obj:         callee.ThisBinding.Obj,
				_isThisArg:  true,
			}
			nativeArgs = make([]*Value, 0, len(args)+1)
			nativeArgs = append(nativeArgs, thisVal)
			nativeArgs = append(nativeArgs, args...)
		}
		return callee.Native(nativeArgs)
	}

	// Handle string methods (type set in evalMember)
	if callee.Type == "stringMethod" && callee.Native != nil {
		nativeArgs := args
		if callee.ThisBinding != nil {
			thisVal := &Value{
				Type:       callee.ThisBinding.Type,
				Str:        callee.ThisBinding.Str,
				_isThisArg: true,
			}
			nativeArgs = make([]*Value, 0, len(args)+1)
			nativeArgs = append(nativeArgs, thisVal)
			nativeArgs = append(nativeArgs, args...)
		}
		return callee.Native(nativeArgs)
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

		// Track recursion depth
		vm.enterCall()
		defer vm.exitCall()

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

	// Handle native functions with Obj field (e.g., Array.isArray, Promise.resolve)
	// These are constructors/static methods attached to native functions
	if obj.Type == "native" && obj.Obj != nil {
		vm.debugLog("evalMember: native function with Obj field, Obj keys: %d", len(obj.Obj))
		if ident, ok := e.Property.(*Ident); ok {
			vm.debugLog("evalMember: looking for property '%s' in native Obj", ident.Name)
			if v, ok := obj.Obj[ident.Name]; ok {
				vm.debugLog("evalMember: found property '%s' in native Obj", ident.Name)
				return v
			}
		}
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

	// Handle number methods (e.g., Math.PI.toFixed(2))
	if obj.Type == "number" {
		if ident, ok := e.Property.(*Ident); ok {
			num := obj.Num
			switch ident.Name {
			case "toFixed":
				return &Value{Type: "native", Native: func(args []*Value) *Value {
					precision := 0
					if len(args) > 0 {
						precision = int(args[0].Num)
					}
					return &Value{Type: "string", Str: strconv.FormatFloat(num, 'f', precision, 64)}
				}}
			case "toPrecision":
				return &Value{Type: "native", Native: func(args []*Value) *Value {
					precision := 0
					if len(args) > 0 {
						precision = int(args[0].Num)
					}
					return &Value{Type: "string", Str: strconv.FormatFloat(num, 'g', precision, 64)}
				}}
			case "toExponential":
				return &Value{Type: "native", Native: func(args []*Value) *Value {
					precision := -1
					if len(args) > 0 {
						precision = int(args[0].Num)
					}
					return &Value{Type: "string", Str: strconv.FormatFloat(num, 'e', precision, 64)}
				}}
			case "toString":
				return &Value{Type: "native", Native: func(args []*Value) *Value {
					radix := 10
					if len(args) > 0 {
						radix = int(args[0].Num)
					}
					if radix == 10 {
						return &Value{Type: "string", Str: strconv.FormatFloat(num, 'f', -1, 64)}
					}
					return &Value{Type: "string", Str: strconv.FormatInt(int64(num), radix)}
				}}
			case "valueOf":
				return &Value{Type: "number", Num: num}
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

	// Handle Function.prototype.call, apply, and bind
	if (obj.Type == "function" && obj.Func != nil) || (obj.Type == "native" && obj.Native != nil) {
		if ident, ok := e.Property.(*Ident); ok {
			// Handle fn.prototype - return the constructor's prototype object
			if ident.Name == "prototype" {
				if obj.PrototypeObj != nil {
					return obj.PrototypeObj
				}
				// Create a default prototype object if not set
				protoObj := &Value{Type: "object", Obj: make(map[string]*Value)}
				if objProto := vm.env.Get("ObjectPrototype"); objProto.Type == "object" {
					protoObj.Proto = objProto
				}
				protoObj.Obj["constructor"] = obj
				obj.PrototypeObj = protoObj
				return protoObj
			}
			if ident.Name == "call" {
				// Return a native function that calls the original function with given this
				fn := obj.Func
				nativeFn := obj.Native
				originalThis := obj.ThisBinding
				return &Value{
					Type: "native",
					Native: func(args []*Value) *Value {
						// Skip 'this' arg if prepended by evalCall
						offset := 0
						if len(args) > 0 && args[0]._isThisArg {
							offset = 1
						}
						if len(args) <= offset {
							return &Value{Type: "undefined"}
						}
						var boundThis *Value = originalThis
						var fnArgs []*Value
						boundThis = args[offset]
						fnArgs = args[offset+1:]
						// If it's a native function, prepend the 'this' and call
						if nativeFn != nil {
							nativeCallArgs := make([]*Value, 0, len(fnArgs)+1)
							if boundThis != nil {
								thisVal := &Value{
									Type:        boundThis.Type,
									Num:         boundThis.Num,
									Str:         boundThis.Str,
									Bool:        boundThis.Bool,
									Obj:         boundThis.Obj,
									Arr:         boundThis.Arr,
									Func:        boundThis.Func,
									Class:       boundThis.Class,
									Promise:     boundThis.Promise,
									Proxy:       boundThis.Proxy,
									MapData:     boundThis.MapData,
									SetData:     boundThis.SetData,
									Native:      boundThis.Native,
									ThisBinding: boundThis.ThisBinding,
									Descriptors: boundThis.Descriptors,
									IsAsync:     boundThis.IsAsync,
									BuiltInConstructor: boundThis.BuiltInConstructor,
									NodeRef:     boundThis.NodeRef,
									Proto:       boundThis.Proto,
									Frozen:      boundThis.Frozen,
									_isThisArg:  true,
								}
								nativeCallArgs = append(nativeCallArgs, thisVal)
							}
							nativeCallArgs = append(nativeCallArgs, fnArgs...)
							return nativeFn(nativeCallArgs)
						}
						// If it's a user-defined function, execute it
						if fn != nil {
							return vm.callFunctionWithThis(fn, boundThis, fnArgs)
						}
						return &Value{Type: "undefined"}
					},
				}
			}
			if ident.Name == "apply" {
				fn := obj.Func
				nativeFn := obj.Native
				originalThis := obj.ThisBinding
				return &Value{
					Type: "native",
					Native: func(args []*Value) *Value {
						// Skip 'this' arg if prepended by evalCall
						offset := 0
						if len(args) > 0 && args[0]._isThisArg {
							offset = 1
						}
						var boundThis *Value = originalThis
						var fnArgs []*Value
						if len(args) > offset {
							boundThis = args[offset]
						}
						if len(args) > offset+1 {
							// Second arg is an array of arguments
							arrVal := args[offset+1]
							if arrVal.Type == "object" && arrVal.Arr != nil {
								fnArgs = arrVal.Arr
							}
						}
						if nativeFn != nil {
							nativeCallArgs := make([]*Value, 0, len(fnArgs)+1)
							if boundThis != nil {
								thisVal := &Value{
									Type:        boundThis.Type,
									Obj:         boundThis.Obj,
									Arr:         boundThis.Arr,
									Str:         boundThis.Str,
									Num:         boundThis.Num,
									Bool:        boundThis.Bool,
									Descriptors: boundThis.Descriptors,
									Proto:       boundThis.Proto,
									_isThisArg:  true,
								}
								nativeCallArgs = append(nativeCallArgs, thisVal)
							}
							nativeCallArgs = append(nativeCallArgs, fnArgs...)
							return nativeFn(nativeCallArgs)
						}
						if fn != nil {
							return vm.callFunctionWithThis(fn, boundThis, fnArgs)
						}
						return &Value{Type: "undefined"}
					},
				}
			}
			// Handle Function.prototype.bind
			if ident.Name == "bind" {
				// Return a native function that creates a bound function
				fn := obj.Func
				originalThis := obj.ThisBinding
				return &Value{
					Type: "native",
					Native: func(args []*Value) *Value {
						// First arg is the this value to bind
						var boundThis *Value = originalThis
						var boundArgs []*Value
						if len(args) > 0 {
							boundThis = args[0]
							boundArgs = args[1:]
						}
						// Return a new function with the bound this and partial arguments
						return &Value{
							Type: "function",
							Func: fn,
							ThisBinding: boundThis,
							Native: func(innerArgs []*Value) *Value {
								// Combine bound args with call args
								allArgs := make([]*Value, 0, len(boundArgs)+len(innerArgs))
								allArgs = append(allArgs, boundArgs...)
								allArgs = append(allArgs, innerArgs...)
								// Call the original function
								childEnv := NewEnvironment(fn.Env)
								if boundThis != nil {
									childEnv.Define("this", boundThis)
								}
								for i, param := range fn.Params {
									if i < len(allArgs) {
										childEnv.Define(param, allArgs[i])
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
					},
				}
			}
		}
	}

	// Walk prototype chain if property not found on this object
	if obj.Type == "object" && obj.Proto != nil {
		return vm.lookupInPrototypeWithOriginal(obj.Proto, e, obj)
	}

	return &Value{Type: "undefined"}
}

// lookupInPrototype walks the prototype chain to find a property.
// Returns undefined if not found in any prototype.
// The originalObj parameter is used to set ThisBinding on found methods,
// so that when obj.hasOwnProperty() is called, 'this' refers to obj, not the prototype.
func (vm *VM) lookupInPrototype(proto *Value, e *MemberExpr) *Value {
	return vm.lookupInPrototypeWithOriginal(proto, e, nil)
}

// lookupInPrototypeWithOriginal walks the prototype chain with an explicit original object.
func (vm *VM) lookupInPrototypeWithOriginal(proto *Value, e *MemberExpr, originalObj *Value) *Value {
	if proto == nil || proto.Type != "object" {
		return &Value{Type: "undefined"}
	}

	if ident, ok := e.Property.(*Ident); ok {
		if proto.Obj != nil {
			if v, ok := proto.Obj[ident.Name]; ok {
				// If the value is a function, bind this to the original object
				// (not the prototype) so method calls work correctly
				thisVal := originalObj
				if thisVal == nil {
					thisVal = proto
				}
				if v.Type == "function" || v.Type == "native" {
					return &Value{
						Type:        v.Type,
						Native:      v.Native,
						Func:        v.Func,
						ThisBinding: thisVal,
					}
				}
				return v
			}
		}
	}

	// Walk up the prototype chain
	if proto.Proto != nil {
		return vm.lookupInPrototypeWithOriginal(proto.Proto, e, originalObj)
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
		// Link the object to the constructor's prototype chain
		// In JS: obj.__proto__ = Constructor.prototype
		if callee.PrototypeObj != nil {
			obj.Proto = callee.PrototypeObj
		}
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
	// For objects, compare by reference (pointer equality)
	if a.Type == "object" && b.Type == "object" {
		// Check if they are the same object (same pointer)
		// This is a simple reference comparison
		return a == b
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

// jsonStringify converts a JS Value to a JSON string.
// Unlike valueToString, this produces valid JSON with quoted keys and proper escaping.
func jsonStringify(v *Value) string {
	switch v.Type {
	case "undefined", "null", "function", "native":
		return "null"
	case "bool":
		if v.Bool {
			return "true"
		}
		return "false"
	case "number":
		if math.IsNaN(v.Num) || math.IsInf(v.Num, 0) {
			return "null"
		}
		s := strconv.FormatFloat(v.Num, 'f', -1, 64)
		// Ensure valid JSON number format
		if strings.Contains(s, ".") {
			return strings.TrimRight(strings.TrimRight(s, "0"), ".")
		}
		return s
	case "string":
		return jsonStringEscape(v.Str)
	case "object":
		if v.Arr != nil {
			parts := make([]string, len(v.Arr))
			for i, a := range v.Arr {
				parts[i] = jsonStringify(a)
			}
			return "[" + strings.Join(parts, ",") + "]"
		}
		if v.Obj != nil {
			parts := make([]string, 0, len(v.Obj))
			for k, val := range v.Obj {
				// Skip function values in JSON serialization
				if val.Type == "function" || val.Type == "native" {
					continue
				}
				parts = append(parts, jsonStringEscape(k)+":"+jsonStringify(val))
			}
			return "{" + strings.Join(parts, ",") + "}"
		}
		return "{}"
	}
	return "null"
}

// jsonStringEscape escapes a string for use in JSON.
func jsonStringEscape(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		case '\b':
			sb.WriteString("\\b")
		case '\f':
			sb.WriteString("\\f")
		default:
			if r < 0x20 {
				sb.WriteString(fmt.Sprintf("\\u%04x", r))
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// jsonParse parses a JSON string into a JS Value.
func jsonParse(s string) *Value {
	s = strings.TrimSpace(s)
	if s == "" {
		return &Value{Type: "undefined"}
	}
	return jsonParseValue(s)
}

// jsonParseValue parses a JSON value from a string.
func jsonParseValue(s string) *Value {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return &Value{Type: "undefined"}
	}
	switch s[0] {
	case '"':
		str, _ := jsonUnescapeString(s)
		return &Value{Type: "string", Str: str}
	case '{':
		return jsonParseObject(s)
	case '[':
		return jsonParseArray(s)
	case 't':
		return &Value{Type: "bool", Bool: true}
	case 'f':
		return &Value{Type: "bool", Bool: false}
	case 'n':
		return &Value{Type: "null"}
	default:
		// Try number
		if s[0] == '-' || (s[0] >= '0' && s[0] <= '9') {
			n, err := strconv.ParseFloat(s, 64)
			if err == nil {
				return &Value{Type: "number", Num: n}
			}
		}
		return &Value{Type: "undefined"}
	}
}

// jsonParseObject parses a JSON object string.
func jsonParseObject(s string) *Value {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '{' {
		return &Value{Type: "object", Obj: make(map[string]*Value)}
	}
	obj := make(map[string]*Value)
	s = s[1 : len(s)-1] // Remove { and }
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return &Value{Type: "object", Obj: obj}
	}
	// Simple parsing: split by commas at the top level
	pairs := jsonSplitTopLevel(s, ',')
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		colonIdx := strings.Index(pair, ":")
		if colonIdx < 0 {
			continue
		}
		key := strings.TrimSpace(pair[:colonIdx])
		val := strings.TrimSpace(pair[colonIdx+1:])
		// Remove quotes from key
		if len(key) >= 2 && key[0] == '"' && key[len(key)-1] == '"' {
			key = key[1 : len(key)-1]
		}
		obj[key] = jsonParseValue(val)
	}
	return &Value{Type: "object", Obj: obj}
}

// jsonParseArray parses a JSON array string.
func jsonParseArray(s string) *Value {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' {
		return &Value{Type: "object", Arr: []*Value{}}
	}
	inner := s[1 : len(s)-1] // Remove [ and ]
	inner = strings.TrimSpace(inner)
	if len(inner) == 0 {
		return &Value{Type: "object", Arr: []*Value{}}
	}
	parts := jsonSplitTopLevel(inner, ',')
	arr := make([]*Value, len(parts))
	for i, part := range parts {
		arr[i] = jsonParseValue(strings.TrimSpace(part))
	}
	return &Value{Type: "object", Arr: arr}
}

// jsonSplitTopLevel splits a string by a delimiter, respecting nested braces and brackets.
func jsonSplitTopLevel(s string, delim byte) []string {
	var parts []string
	depth := 0
	inString := false
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' && (i == 0 || s[i-1] != '\\') {
			inString = !inString
		}
		if inString {
			continue
		}
		if c == '{' || c == '[' {
			depth++
		} else if c == '}' || c == ']' {
			depth--
		} else if c == delim && depth == 0 {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// jsonUnescapeString unescapes a JSON string.
func jsonUnescapeString(s string) (string, error) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return s, nil
	}
	s = s[1 : len(s)-1]
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			case '/':
				sb.WriteByte('/')
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case 'u':
				if i+4 < len(s) {
					hex := s[i+1 : i+5]
					n, err := strconv.ParseInt(hex, 16, 32)
					if err == nil {
						sb.WriteRune(rune(n))
					}
					i += 4
				}
			default:
				sb.WriteByte(s[i])
			}
		} else {
			sb.WriteByte(s[i])
		}
	}
	return sb.String(), nil
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
		return vm.callFunctionWithThis(fn.Func, fn.ThisBinding, args)
	}

	return &Value{Type: "undefined"}
}

// callFunctionWithThis calls a user-defined function with an explicit this binding.
func (vm *VM) callFunctionWithThis(f *Function, thisVal *Value, args []*Value) *Value {
	childEnv := NewEnvironment(f.Env)

	if thisVal != nil {
		childEnv.Define("this", thisVal)
	}

	for i, param := range f.Params {
		if i < len(args) {
			childEnv.Define(param, args[i])
		} else if f.DefaultVals != nil && f.DefaultVals[param] != nil {
			childEnv.Define(param, vm.evalExpr(f.DefaultVals[param]))
		} else {
			childEnv.Define(param, &Value{Type: "undefined"})
		}
	}

	if f.RestParam != "" {
		restArgs := make([]*Value, 0)
		if len(args) > len(f.Params) {
			restArgs = args[len(f.Params):]
		}
		childEnv.Define(f.RestParam, &Value{Type: "object", Arr: restArgs})
	}

	oldEnv := vm.env
	vm.env = childEnv

	vm.enterCall()
	defer vm.exitCall()

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
}

// setupSymbol adds Symbol constructor to the VM
func (vm *VM) setupSymbol() {
	symbolCounter := 0
	symbolCache := make(map[string]*Value)

	symbolConstructor := &Value{Type: "native", Native: func(args []*Value) *Value {
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
	}}

	// Create well-known symbols
	symbolCounter++
	iteratorSymbol := &Value{Type: "symbol", Str: "Symbol.iterator", Num: float64(symbolCounter)}
	symbolCache["iterator"] = iteratorSymbol

	symbolCounter++
	toPrimitiveSymbol := &Value{Type: "symbol", Str: "Symbol.toPrimitive", Num: float64(symbolCounter)}
	symbolCache["toPrimitive"] = toPrimitiveSymbol

	symbolCounter++
	toStringTagSymbol := &Value{Type: "symbol", Str: "Symbol.toStringTag", Num: float64(symbolCounter)}
	symbolCache["toStringTag"] = toStringTagSymbol

	symbolCounter++
	speciesSymbol := &Value{Type: "symbol", Str: "Symbol.species", Num: float64(symbolCounter)}
	symbolCache["species"] = speciesSymbol

	symbolCounter++
	hasInstanceSymbol := &Value{Type: "symbol", Str: "Symbol.hasInstance", Num: float64(symbolCounter)}
	symbolCache["hasInstance"] = hasInstanceSymbol

	// Add well-known symbols as properties of Symbol constructor
	symbolConstructor.Obj = map[string]*Value{
		"iterator":     iteratorSymbol,
		"toPrimitive":  toPrimitiveSymbol,
		"toStringTag":  toStringTagSymbol,
		"species":      speciesSymbol,
		"hasInstance":   hasInstanceSymbol,
		"for": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) > 0 {
				key := valueToString(args[0])
				if s, ok := symbolCache[key]; ok {
					return s
				}
				symbolCounter++
				s := &Value{Type: "symbol", Str: key, Num: float64(symbolCounter)}
				symbolCache[key] = s
				return s
			}
			return &Value{Type: "undefined"}
		}},
		"keyFor": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) > 0 && args[0].Type == "symbol" {
				for k, s := range symbolCache {
					if s.Num == args[0].Num {
						return &Value{Type: "string", Str: k}
					}
				}
			}
			return &Value{Type: "undefined"}
		}},
	}

	vm.env.Define("Symbol", symbolConstructor)
	// Store iterator symbol for internal use
	vm.env.Define("_symbolIterator", iteratorSymbol)
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
