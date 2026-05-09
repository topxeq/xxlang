package jsengine

import (
	"fmt"
	"math"
	"reflect"
	"sort"
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
	store     map[string]*Value
	outer     *Environment
	globalObj *Value // global object (window) for property-as-variable lookup
	withObj   *Value // with-statement object: property lookups check this first
	vm        *VM    // reference to VM for calling getters/setters in with-objects
}

func NewEnvironment(outer *Environment) *Environment {
	return &Environment{
		store: make(map[string]*Value),
		outer: outer,
	}
}

func (e *Environment) Get(name string) *Value {
	if e.withObj != nil && e.vm != nil {
		v := e.vm.lookupWithObject(e.withObj, name)
		if v != nil && v.Type != "undefined" {
			return v
		}
	}
	if v, ok := e.store[name]; ok {
		return v
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	// At global scope: also check the global object (window) properties
	if e.globalObj != nil && e.globalObj.Obj != nil {
		if v, ok := e.globalObj.Obj[name]; ok {
			return v
		}
	}
	return &Value{Type: "undefined"}
}

// lookupWithObject looks up a property on a with-object, walking the prototype chain.
func (vm *VM) lookupWithObject(obj *Value, name string) *Value {
	visited := 0
	current := obj
	for current != nil && visited < 20 {
		if current.Descriptors != nil {
			if desc, ok := current.Descriptors[name]; ok {
				if desc.Get != nil {
					result := vm.callGetter(desc.Get, current)
					if result != nil && (result.Type == "function" || result.Type == "native") && result.ThisBinding == nil {
						result.ThisBinding = obj
					}
					return result
				}
				if desc.Value != nil {
					result := desc.Value
					if result != nil && (result.Type == "function" || result.Type == "native") && result.ThisBinding == nil {
						result.ThisBinding = obj
					}
					return result
				}
			}
		}
		if current.Obj != nil {
			if v, ok := current.Obj[name]; ok {
				if v != nil && (v.Type == "function" || v.Type == "native") && v.ThisBinding == nil {
					copied := *v
					copied.ThisBinding = obj
					return &copied
				}
				return v
			}
		}
		if current.Proto != nil {
			current = current.Proto
			visited++
			continue
		}
		break
	}
	return nil
}

// setOnWithObj tries to set a property on the with-object or its prototype chain.
// Returns true if the property was found and set.
func (vm *VM) setOnWithObj(obj *Value, name string, val *Value) bool {
	visited := 0
	current := obj
	for current != nil && visited < 20 {
		if current.Descriptors != nil {
			if desc, ok := current.Descriptors[name]; ok {
				if desc.Set != nil {
					vm.callSetter(desc.Set, current, val)
					return true
				}
				if desc.Get != nil {
					return true
				}
			}
		}
		if current.Obj != nil {
			if _, ok := current.Obj[name]; ok {
				current.Obj[name] = val
				return true
			}
		}
		if current.Proto != nil {
			current = current.Proto
			visited++
			continue
		}
		break
	}
		return false
}

func (e *Environment) Set(name string, val *Value) {
	if e.withObj != nil && e.vm != nil {
		if e.vm.setOnWithObj(e.withObj, name, val) {
			return
		}
	}
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

// Delete removes a variable from the environment (used by the delete operator).
func (e *Environment) Delete(name string) {
	delete(e.store, name)
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
	callStack    []string      // function call stack for debugging
	// Memory allocation tracking to prevent memory exhaustion
	allocCount   int64         // number of objects/arrays allocated
	maxAllocs    int64         // maximum allocations before error (0 = unlimited)
	// Abort flag for external cancellation
	abortFlag    bool          // set to true to abort execution
	// running flag prevents step counter resets during nested Run() calls
	running bool
	// Step profiling callback: called at step milestones for debugging infinite loops
	stepProfileFn func(step int64, stack []string)
	_stepProfileLast int64 // last step count at which profile was logged
	// Step profile AST capture: when set, captures AST of current function at milestones
	stepProfileAST string // last captured AST snippet for the looping function
	// accessorDepth tracks nested getter/setter calls to prevent infinite re-entrancy.
	// Vue 2's reactivity can create getter→setter→getter→... chains when Dep.target
	// tracking goes wrong. Real browsers handle this via the JS engine's call stack,
	// but our VM needs an explicit guard.
	accessorDepth int
	// accessorCalls counts total getter/setter invocations to cap flat (non-nested) loops.
	// Vue 2's Dep/Watcher system can create shallow but very long reactive chains.
	// After a threshold, we suppress further accessor calls to let the program proceed.
	accessorCalls int
	accessorMax   int
	// forIterCount tracks total for/while loop iterations to break infinite reactive cycles.
	// Vue's scheduler flush loops can restart infinitely via nextTick. A cumulative
	// iteration cap ensures the program eventually progresses past the reactivity phase.
	forIterCount int
	forIterMax   int
}

// SetAccessorMax sets the maximum number of accessor (getter/setter) invocations
// allowed before suppressing further calls. Set to 0 for unlimited.
func (vm *VM) SetAccessorMax(max int) {
	vm.accessorMax = max
}

// ResetAccessorCalls resets the accessor invocation counter.
func (vm *VM) ResetAccessorCalls() {
	vm.accessorCalls = 0
}

// SetForIterMax sets the maximum cumulative for/while loop iterations.
// Set to 0 for unlimited.
func (vm *VM) SetForIterMax(max int) {
	vm.forIterMax = max
}

// ResetForIterCount resets the for/while loop iteration counter.
func (vm *VM) ResetForIterCount() {
	vm.forIterCount = 0
}

// GetForIterCount returns the current for/while loop iteration count.
func (vm *VM) GetForIterCount() int {
	return vm.forIterCount
}

// GetAccessorCalls returns the current accessor invocation count.
func (vm *VM) GetAccessorCalls() int {
	return vm.accessorCalls
}

// SetStepProfileFn sets a callback for step profiling during execution.
func (vm *VM) SetStepProfileFn(fn func(step int64, stack []string)) {
	vm.stepProfileFn = fn
}

// GetStepProfileAST returns the last captured AST snippet for the looping function.
func (vm *VM) GetStepProfileAST() string {
	return vm.stepProfileAST
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

// SetStepProfile sets a callback for step profiling. Called every 1M steps with
// the current step count and call stack. Pass nil to disable.
func (vm *VM) SetStepProfile(fn func(step int64, stack []string)) {
	vm.stepProfileFn = fn
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

// DefineGlobal defines a global variable in the top-level environment.
func (vm *VM) DefineGlobal(name string, val *Value) {
	vm.env.Define(name, val)
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

// SetStepCount sets the current step count (for resetting between evaluations).
func (vm *VM) SetStepCount(count int64) {
	vm.stepCount = count
}

// GetStepCount returns the current step count.
func (vm *VM) GetStepCount() int64 {
	return vm.stepCount
}

// Env returns the current global environment for external access.
func (vm *VM) Env() *Environment {
	return vm.env
}

// GetMaxSteps returns the maximum step limit.
func (vm *VM) GetMaxSteps() int64 {
	return vm.maxSteps
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
func (vm *VM) enterCall() {
	vm.callDepth++
	if vm.maxCallDepth > 0 && vm.callDepth > vm.maxCallDepth {
		// Build stack trace from callStack
		msg := fmt.Sprintf("Recursion limit: exceeded %d call depth", vm.maxCallDepth)
		if len(vm.callStack) > 0 {
			last := vm.callStack
			if len(last) > 10 {
				last = last[len(last)-10:]
			}
			msg += "\nCall stack (last 10):\n"
			for _, s := range last {
				msg += "  " + s + "\n"
			}
		}
		ThrowJS(NewError("RangeError", msg))
	}
}

// exitCall decrements the call depth.
func (vm *VM) exitCall() {
	if vm.callDepth > 0 {
		vm.callDepth--
	}
	if len(vm.callStack) > 0 {
		vm.callStack = vm.callStack[:len(vm.callStack)-1]
	}
}

// checkSteps increments the step counter and checks for timeout conditions.
// It panics with a JS exception if the step limit or wall-clock timeout is exceeded.
// This is called from execStmt, execBlock, and evalExpr to prevent infinite loops.
func (vm *VM) checkSteps() {
	vm.stepCount++
	// Step profiling: log call stack at milestones (every 1M steps)
	if vm.stepProfileFn != nil && vm.stepCount%1_000_000 == 0 && vm.stepCount != vm._stepProfileLast {
		vm._stepProfileLast = vm.stepCount
		vm.stepProfileFn(vm.stepCount, vm.callStack)
	}
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
		"warn": {Type: "native", Native: func(args []*Value) *Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = valueToString(a)
			}
			vm.output = append(vm.output, "[warn] "+strings.Join(parts, " "))
			return &Value{Type: "undefined"}
		}},
		"error": {Type: "native", Native: func(args []*Value) *Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = valueToString(a)
			}
			vm.output = append(vm.output, "[error] "+strings.Join(parts, " "))
			return &Value{Type: "undefined"}
		}},
		"info": {Type: "native", Native: func(args []*Value) *Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = valueToString(a)
			}
			vm.output = append(vm.output, "[info] "+strings.Join(parts, " "))
			return &Value{Type: "undefined"}
		}},
		"debug": {Type: "native", Native: func(args []*Value) *Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = valueToString(a)
			}
			vm.output = append(vm.output, "[debug] "+strings.Join(parts, " "))
			return &Value{Type: "undefined"}
		}},
		"trace": {Type: "native", Native: func(args []*Value) *Value {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = valueToString(a)
			}
			vm.output = append(vm.output, "[trace] "+strings.Join(parts, " "))
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
	// Document native functions use nativeThisOffset to skip the prepended 'this'
	// argument that evalCall adds when calling methods via member expression (e.g. document.getElementById).
	vm.env.Define("document", &Value{Type: "object", Obj: map[string]*Value{
		"title": {Type: "string", Str: title},
		"querySelector": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset || vm.root == nil {
				return &Value{Type: "null"}
			}
			sel := args[offset].Str
			node := dom.QuerySelector(vm.root, sel)
			if node == nil {
				return &Value{Type: "null"}
			}
			return vm.wrapNode(node)
		}},
		"querySelectorAll": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset || vm.root == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			sel := args[offset].Str
			nodes := dom.QuerySelectorAll(vm.root, sel)
			arr := make([]*Value, len(nodes))
			for i, n := range nodes {
				arr[i] = vm.wrapNode(n)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"getElementById": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset || vm.root == nil {
				return &Value{Type: "null"}
			}
			id := args[offset].Str
			node := dom.GetElementByID(vm.root, id)
			if node == nil {
				return &Value{Type: "null"}
			}
			return vm.wrapNode(node)
		}},
		"getElementsByClassName": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset || vm.root == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			className := args[offset].Str
			nodes := dom.GetElementsByClassName(vm.root, className)
			arr := make([]*Value, len(nodes))
			for i, n := range nodes {
				arr[i] = vm.wrapNode(n)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"getElementsByTagName": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset || vm.root == nil {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			tag := args[offset].Str
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
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "null"}
			}
			tagArg := args[offset]
			tag := ""
			if tagArg.Type == "string" {
				tag = tagArg.Str
			} else if tagArg.Type == "object" && tagArg.Obj != nil {
				// Component object: try to extract name from name or template
				if name, ok := tagArg.Obj["name"]; ok && name.Type == "string" {
					tag = name.Str
				} else {
					tag = "div"
				}
			} else if tagArg.Type == "function" {
				if tagArg.Func != nil {
					tag = "div"
				}
			}
			if tag == "" {
				tag = "div"
			}
			newNode := &dom.Node{Type: dom.ElementNode, Data: strings.ToLower(tag)}
			return vm.wrapNode(newNode)
		}},
		"createElementNS": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return &Value{Type: "null"}
			}
			tag := args[offset+1].Str
			newNode := &dom.Node{Type: dom.ElementNode, Data: strings.ToLower(tag)}
			return vm.wrapNode(newNode)
		}},
		"createTextNode": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			text := ""
			if len(args) > offset {
				text = args[offset].Str
			}
			newNode := &dom.Node{Type: dom.TextNode, Data: text}
			return vm.wrapNode(newNode)
		}},
		"createDocumentFragment": {Type: "native", Native: func(args []*Value) *Value {
			newNode := &dom.Node{Type: dom.DocumentNode, Data: "#document-fragment"}
			return vm.wrapNode(newNode)
		}},
		"createComment": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			text := ""
			if len(args) > offset {
				text = args[offset].Str
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
			"reload": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
			"replace": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
			"assign": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
		}},
		"navigator": {Type: "object", Obj: map[string]*Value{
			"userAgent": {Type: "string", Str: "Xxlang-HLBR/1.0"},
			"language": {Type: "string", Str: "en-US"},
			"platform": {Type: "string", Str: "Xxlang"},
			"cookieEnabled": {Type: "bool", Bool: true},
		}},
		"history": {Type: "object", Obj: map[string]*Value{
			"length": {Type: "number", Num: 1},
			"scrollRestoration": {Type: "string", Str: "auto"},
			"state": {Type: "null"},
			"pushState": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
			"replaceState": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
			"go": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
			"back": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
			"forward": {Type: "native", Native: func(args []*Value) *Value {
				return &Value{Type: "undefined"}
			}},
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
		"getComputedStyle": {Type: "native", Native: func(args []*Value) *Value {
			// Stub: returns a CSSStyleDeclaration-like object
			return &Value{Type: "object", Obj: map[string]*Value{
				"getPropertyValue": {Type: "native", Native: func(innerArgs []*Value) *Value {
					return &Value{Type: "string", Str: ""}
				}},
				"display":    {Type: "string", Str: ""},
				"visibility": {Type: "string", Str: ""},
				"position":   {Type: "string", Str: ""},
				"width":      {Type: "string", Str: ""},
				"height":     {Type: "string", Str: ""},
			}}
		}},
	}}
	vm.env.Define("window", windowObj)
	// Set the global object so that window properties are accessible as global variables
	vm.env.globalObj = windowObj
	vm.env.Define("this", windowObj)
	vm.env.Define("self", windowObj)

	vm.env.Define("Math", &Value{Type: "object", Obj: map[string]*Value{
		"PI": {Type: "number", Num: math.Pi},
		"E":  {Type: "number", Num: math.E},
		"sqrt": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Sqrt(args[offset].Num)}
		}},
		"abs": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Abs(args[offset].Num)}
		}},
		"floor": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Floor(args[offset].Num)}
		}},
		"ceil": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Ceil(args[offset].Num)}
		}},
		"round": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "number"}
			}
			return &Value{Type: "number", Num: math.Round(args[offset].Num)}
		}},
		"max": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "number", Num: math.Inf(-1)}
			}
			m := args[offset].Num
			for _, a := range args[offset+1:] {
				if a.Num > m {
					m = a.Num
				}
			}
			return &Value{Type: "number", Num: m}
		}},
		"min": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "number", Num: math.Inf(1)}
			}
			m := args[offset].Num
			for _, a := range args[offset+1:] {
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
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "string"}
			}
			return &Value{Type: "string", Str: jsonStringify(args[offset])}
		}},
		"parse": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "undefined"}
			}
			return jsonParse(args[offset].Str)
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

	vm.env.Define("String", &Value{Type: "native", BuiltInConstructor: "String", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "string"}
		}
		return &Value{Type: "string", Str: valueToString(args[0])}
	}})

	vm.env.Define("Number", &Value{Type: "native", BuiltInConstructor: "Number", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "number"}
		}
		n, _ := strconv.ParseFloat(args[0].Str, 64)
		return &Value{Type: "number", Num: n}
	}})

	vm.env.Define("Boolean", &Value{Type: "native", BuiltInConstructor: "Boolean", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: isTruthy(args[0])}
	}})

	vm.env.Define("Date", &Value{Type: "native", BuiltInConstructor: "Date", Native: func(args []*Value) *Value {
		ts := float64(time.Now().UnixNano() / 1e6)
		offset := NativeThisOffset(args)
		if len(args) > offset {
			if args[offset].Type == "number" {
				ts = args[offset].Num
			} else if args[offset].Type == "string" {
				if t, err := time.Parse(time.RFC3339, args[offset].Str); err == nil {
					ts = float64(t.UnixNano() / 1e6)
				}
			}
		}
		obj := make(map[string]*Value)
		obj["getTime"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			return &Value{Type: "number", Num: ts}
		}}
		obj["valueOf"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			return &Value{Type: "number", Num: ts}
		}}
		obj["toString"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "string", Str: t.UTC().Format("Mon Jan 02 2006 15:04:05 GMT+0000 (Coordinated Universal Time)")}
		}}
		obj["toISOString"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "string", Str: t.UTC().Format("2006-01-02T15:04:05.000Z")}
		}}
		obj["toJSON"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "string", Str: t.UTC().Format("2006-01-02T15:04:05.000Z")}
		}}
		obj["getFullYear"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "number", Num: float64(t.Year())}
		}}
		obj["getMonth"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "number", Num: float64(t.Month() - 1)}
		}}
		obj["getDate"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "number", Num: float64(t.Day())}
		}}
		obj["getHours"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "number", Num: float64(t.Hour())}
		}}
		obj["getMinutes"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "number", Num: float64(t.Minute())}
		}}
		obj["getSeconds"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "number", Num: float64(t.Second())}
		}}
		obj["getMilliseconds"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			t := time.Unix(0, int64(ts)*1e6)
			return &Value{Type: "number", Num: float64(t.Nanosecond() / 1e6)}
		}}
		obj["getTimezoneOffset"] = &Value{Type: "native", Native: func(a2 []*Value) *Value {
			_, offset := time.Unix(0, int64(ts)*1e6).Zone()
			return &Value{Type: "number", Num: float64(-offset / 60)}
		}}
		return &Value{Type: "object", Obj: obj, BuiltInConstructor: "Date"}
	}})
	dateObj := vm.env.Get("Date")
	if dateObj.Type == "native" {
		if dateObj.Obj == nil {
			dateObj.Obj = make(map[string]*Value)
		}
		dateObj.Obj["now"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: float64(time.Now().UnixNano() / 1e6)}
		}}
		dateObj.Obj["parse"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 || args[0].Type != "string" {
				return &Value{Type: "number", Num: math.NaN()}
			}
			if t, err := time.Parse(time.RFC3339, args[0].Str); err == nil {
				return &Value{Type: "number", Num: float64(t.UnixNano() / 1e6)}
			}
			return &Value{Type: "number", Num: math.NaN()}
		}}
		dateObj.Obj["UTC"] = &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: math.NaN()}
		}}
	}
	vm.env.Define("DatePrototype", &Value{Type: "object", Obj: make(map[string]*Value)})

	vm.env.Define("Event", &Value{Type: "native", BuiltInConstructor: "Event", Native: func(args []*Value) *Value {
		offset := NativeThisOffset(args)
		eventType := ""
		if len(args) > offset && args[offset] != nil {
			eventType = args[offset].Str
		}
		bubbles := false
		cancelable := false
		if len(args) > offset+1 && args[offset+1] != nil && args[offset+1].Type == "object" && args[offset+1].Obj != nil {
			if b, ok := args[offset+1].Obj["bubbles"]; ok && b != nil {
				bubbles = b.Bool
			}
			if c, ok := args[offset+1].Obj["cancelable"]; ok && c != nil {
				cancelable = c.Bool
			}
		}
		return &Value{Type: "object", Obj: map[string]*Value{
			"type":             {Type: "string", Str: eventType},
			"bubbles":          {Type: "bool", Bool: bubbles},
			"cancelable":       {Type: "bool", Bool: cancelable},
			"defaultPrevented": {Type: "bool", Bool: false},
			"target":           {Type: "null"},
			"currentTarget":    {Type: "null"},
			"timeStamp":        {Type: "number", Num: float64(0)},
		}}
	}})

	vm.env.Define("CustomEvent", &Value{Type: "native", BuiltInConstructor: "CustomEvent", Native: func(args []*Value) *Value {
		offset := NativeThisOffset(args)
		eventType := ""
		if len(args) > offset && args[offset] != nil {
			eventType = args[offset].Str
		}
		bubbles := false
		cancelable := false
		var detail *Value
		if len(args) > offset+1 && args[offset+1] != nil && args[offset+1].Type == "object" && args[offset+1].Obj != nil {
			if b, ok := args[offset+1].Obj["bubbles"]; ok && b != nil {
				bubbles = b.Bool
			}
			if c, ok := args[offset+1].Obj["cancelable"]; ok && c != nil {
				cancelable = c.Bool
			}
			if d, ok := args[offset+1].Obj["detail"]; ok {
				detail = d
			}
		}
		obj := map[string]*Value{
			"type":             {Type: "string", Str: eventType},
			"bubbles":          {Type: "bool", Bool: bubbles},
			"cancelable":       {Type: "bool", Bool: cancelable},
			"defaultPrevented": {Type: "bool", Bool: false},
			"target":           {Type: "null"},
			"currentTarget":    {Type: "null"},
			"timeStamp":        {Type: "number", Num: float64(0)},
		}
		if detail != nil {
			obj["detail"] = detail
		} else {
			obj["detail"] = &Value{Type: "null"}
		}
		return &Value{Type: "object", Obj: obj}
	}})

	vm.env.Define("MouseEvent", &Value{Type: "native", BuiltInConstructor: "MouseEvent", Native: func(args []*Value) *Value {
		offset := NativeThisOffset(args)
		eventType := "click"
		if len(args) > offset && args[offset] != nil {
			eventType = args[offset].Str
		}
		return &Value{Type: "object", Obj: map[string]*Value{
			"type":             {Type: "string", Str: eventType},
			"bubbles":          {Type: "bool", Bool: true},
			"cancelable":       {Type: "bool", Bool: true},
			"defaultPrevented": {Type: "bool", Bool: false},
			"target":           {Type: "null"},
			"currentTarget":    {Type: "null"},
			"timeStamp":        {Type: "number", Num: float64(0)},
			"clientX":          {Type: "number", Num: 0},
			"clientY":          {Type: "number", Num: 0},
			"button":           {Type: "number", Num: 0},
		}}
	}})

	arrayConstructor := &Value{Type: "native", BuiltInConstructor: "Array", Native: func(args []*Value) *Value {
		arr := vm.newArray(args)
		arr.BuiltInConstructor = "Array"
		return arr
	}}
	// Add static methods to the Array constructor
	arrayConstructor.Obj = map[string]*Value{
		"isArray": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "bool", Bool: false}
			}
			return &Value{Type: "bool", Bool: args[offset].Type == "object" && args[offset].Arr != nil}
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

	// Make Array static methods non-writable and non-configurable to prevent
	// polyfills from replacing them with broken JS implementations.
	if arrayConstructor.Obj != nil {
		if arrayConstructor.Descriptors == nil {
			arrayConstructor.Descriptors = make(map[string]*PropertyDescriptor)
		}
		for k := range arrayConstructor.Obj {
			if _, exists := arrayConstructor.Descriptors[k]; !exists {
				arrayConstructor.Descriptors[k] = &PropertyDescriptor{
					Writable:     false,
					Configurable: false,
					Enumerable:   false,
				}
			}
		}
	}

	objectMethods := GetObjectMethods(vm)
	objectVal := &Value{Type: "native", BuiltInConstructor: "Object", Obj: objectMethods}
	// Make Object static methods non-writable and non-configurable to prevent
	// polyfills (e.g. core-js) from replacing them with broken JS implementations.
	// Our native Go implementations work correctly; polyfill JS ones do not.
	objectDescs := make(map[string]*PropertyDescriptor)
	for k := range objectMethods {
		objectDescs[k] = &PropertyDescriptor{
			Writable:     false,
			Configurable: false,
			Enumerable:   false,
		}
	}
	objectVal.Descriptors = objectDescs
	vm.env.Define("Object", objectVal)

	// Set Object.prototype to the ObjectPrototype
	objectProto := vm.env.Get("ObjectPrototype")
	if objectProto.Type == "object" {
		objectVal := vm.env.Get("Object")
		if objectVal.Obj != nil {
			objectVal.Obj["prototype"] = objectProto
		}
		// Set constructor on ObjectPrototype to point back to Object
		if objectProto.Obj != nil {
			objectProto.Obj["constructor"] = objectVal
		}
	}

	// Define Function constructor with prototype
	functionProto := vm.env.Get("FunctionPrototype")
	functionConstructor := &Value{Type: "native", BuiltInConstructor: "Function", Native: func(args []*Value) *Value {
		// Function constructor: new Function(p1, p2, ..., body)
		// Last argument is the function body, preceding arguments are parameter names.
		if len(args) == 0 {
			return vm.newFunction(&Function{Params: []string{}, Body: []Statement{}, Env: vm.env})
		}
		offset := nativeThisOffset(args)
		if offset >= len(args) {
			return vm.newFunction(&Function{Params: []string{}, Body: []Statement{}, Env: vm.env})
		}
		bodyStr := ""
		paramNames := []string{}
		for i := offset; i < len(args); i++ {
			if i == len(args)-1 {
				bodyStr = valueToString(args[i])
			} else {
				paramNames = append(paramNames, valueToString(args[i]))
			}
		}
		p := NewParser(bodyStr)
		prog := p.Parse()
		fn := &Function{Params: paramNames, Body: prog.Statements, Env: vm.env}
		return vm.newFunction(fn)
	}}
	if functionProto.Type == "object" {
		functionConstructor.Obj = map[string]*Value{
			"prototype": functionProto,
		}
		// Set constructor on FunctionPrototype to point back to Function
		if functionProto.Obj != nil {
			functionProto.Obj["constructor"] = functionConstructor
		}
	}
	vm.env.Define("Function", functionConstructor)

	// Set Array.prototype to the ArrayPrototype
	arrayProto := vm.env.Get("ArrayPrototype")
	if arrayProto.Type == "object" {
		arrayObj := vm.env.Get("Array")
		arrayObjType := "object"
		if arrayObj.Type == "native" || arrayObj.Type == "function" {
			arrayObjType = arrayObj.Type
		}
		if (arrayObjType == "object" || arrayObjType == "native") && arrayObj.Obj != nil {
			arrayObj.Obj["prototype"] = arrayProto
		}
		// Set constructor on ArrayPrototype to point back to Array
		if arrayProto.Obj != nil {
			arrayProto.Obj["constructor"] = arrayObj
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
	vm.env.Define("RegExp", &Value{Type: "native", BuiltInConstructor: "RegExp", Native: func(args []*Value) *Value {
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
	vm.env.Define("Error", &Value{Type: "native", BuiltInConstructor: "Error", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewError("Error", message)
	}})

	vm.env.Define("TypeError", &Value{Type: "native", BuiltInConstructor: "TypeError", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewTypeError(message)
	}})

	vm.env.Define("ReferenceError", &Value{Type: "native", BuiltInConstructor: "ReferenceError", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewReferenceError(message)
	}})

	vm.env.Define("SyntaxError", &Value{Type: "native", BuiltInConstructor: "SyntaxError", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewSyntaxError(message)
	}})

	vm.env.Define("RangeError", &Value{Type: "native", BuiltInConstructor: "RangeError", Native: func(args []*Value) *Value {
		message := ""
		if len(args) > 0 {
			message = args[0].Str
		}
		return NewRangeError(message)
	}})


	vm.env.Define("undefined", &Value{Type: "undefined"})
	vm.env.Define("NaN", &Value{Type: "number", Num: math.NaN()})
	vm.env.Define("Infinity", &Value{Type: "number", Num: math.Inf(1)})

	// performance.now() - returns high-resolution timestamp
	// Used by Vue.js for timing measurements
	vm.env.Define("performance", &Value{Type: "object", Obj: map[string]*Value{
		"now": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: float64(time.Now().UnixNano()) / 1e6}
		}},
		"timing": {Type: "object", Obj: map[string]*Value{
			"navigationStart": {Type: "number", Num: float64(time.Now().UnixNano()) / 1e6},
		}},
	}})

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
		windowObj.Obj["performance"] = vm.env.Get("performance")
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
		// When called via callSetter: args = [obj, val]
		// The value is always the last argument.
		valIdx := len(args) - 1
		if valIdx >= 0 {
			newChildren := htmlparser.ParseFragment(args[valIdx].Str)
			n.Children = newChildren
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
		// When called via callSetter: args = [obj, val]
		// The value is always the last argument.
		valIdx := len(args) - 1
		if valIdx >= 0 {
			n.Children = []*dom.Node{
				{Type: dom.TextNode, Data: args[valIdx].Str, Parent: n},
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
		"nodeType":    {Type: "number", Num: float64(n.Type)},
		"getAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "null"}
			}
			return &Value{Type: "string", Str: n.GetAttribute(args[offset].Str)}
		}},
		"setAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) >= offset+2 {
				n.SetAttribute(args[offset].Str, args[offset+1].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"removeAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) > offset {
				n.RemoveAttribute(args[offset].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"hasAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "bool", Bool: false}
			}
			return &Value{Type: "bool", Bool: n.HasAttribute(args[offset].Str)}
		}},
		"querySelector": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "null"}
			}
			child := dom.QuerySelector(n, args[offset].Str)
			return vm.wrapNode(child)
		}},
		"querySelectorAll": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			children := dom.QuerySelectorAll(n, args[offset].Str)
			arr := make([]*Value, len(children))
			for i, c := range children {
				arr[i] = vm.wrapNode(c)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"firstChild": {Type: "null"},
		"lastChild": {Type: "null"},
		"parentNode": {Type: "null"},
		"parentElement": {Type: "null"},
		"nextSibling": {Type: "null"},
		"previousSibling": {Type: "null"},
		"nextElementSibling": {Type: "null"},
		"previousElementSibling": {Type: "null"},
		// DOM mutation methods
		"appendChild": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "undefined"}
			}
			child := args[offset]
			if childNode := vm.unwrapNode(child); childNode != nil {
				n.AppendChild(childNode)
				return child
			}
			return &Value{Type: "undefined"}
		}},
		"removeChild": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "undefined"}
			}
			child := args[offset]
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
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "undefined"}
			}
			newNode := args[offset]
			newDomNode := vm.unwrapNode(newNode)
			if newDomNode == nil {
				return &Value{Type: "undefined"}
			}
			// If no refNode or refNode is null, append at end (like appendChild)
			if len(args) <= offset+1 || args[offset+1].Type == "null" || args[offset+1].Type == "undefined" {
				newDomNode.Parent = n
				n.Children = append(n.Children, newDomNode)
				return newNode
			}
			refNode := args[offset+1]
			refDomNode := vm.unwrapNode(refNode)
			if refDomNode != nil {
				for i, c := range n.Children {
					if c == refDomNode {
						newDomNode.Parent = n
						n.Children = append(n.Children[:i], append([]*dom.Node{newDomNode}, n.Children[i:]...)...)
						break
					}
				}
			}
			return newNode
		}},
		"replaceChild": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return &Value{Type: "undefined"}
			}
			newNode := args[offset]
			oldNode := args[offset+1]
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
			offset := nativeThisOffset(args)
			deep := false
			if len(args) > offset {
				deep = args[offset].Bool
			}
			clone := vm.cloneNode(n, deep)
			return vm.wrapNode(clone)
		}},
		"contains": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "bool", Bool: false}
			}
			other := vm.unwrapNode(args[offset])
			return &Value{Type: "bool", Bool: vm.nodeContains(n, other)}
		}},
		"addEventListener": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return &Value{Type: "undefined"}
			}
			eventType := args[offset].Str
			handler := args[offset+1]
			if n.EventListeners == nil {
				n.EventListeners = make(map[string][]interface{})
			}
			n.EventListeners[eventType] = append(n.EventListeners[eventType], handler)
			return &Value{Type: "undefined"}
		}},
		"removeEventListener": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return &Value{Type: "undefined"}
			}
			eventType := args[offset].Str
			handler := args[offset+1]
			if n.EventListeners != nil {
				handlers := n.EventListeners[eventType]
				for i, h := range handlers {
					if h == handler {
						n.EventListeners[eventType] = append(handlers[:i], handlers[i+1:]...)
						break
					}
				}
			}
			return &Value{Type: "undefined"}
		}},
		"dispatchEvent": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "bool", Bool: true}
			}
			event := args[offset]
			var eventType string
			if event != nil && event.Type == "object" && event.Obj != nil {
				if t, ok := event.Obj["type"]; ok && t != nil {
					eventType = t.Str
				}
			}
			if eventType == "" {
				return &Value{Type: "bool", Bool: true}
			}
			if n.EventListeners != nil {
				handlers := n.EventListeners[eventType]
				for _, h := range handlers {
					if fn, ok := h.(*Value); ok && fn != nil {
						vm.callFunction(fn, []*Value{{Type: "object", Obj: map[string]*Value{
							"type":       {Type: "string", Str: eventType},
							"target":     args[0],
							"currentTarget": args[0],
							"bubbles":    {Type: "bool", Bool: true},
							"cancelable": {Type: "bool", Bool: true},
							"defaultPrevented": {Type: "bool", Bool: false},
						}}})
					}
				}
			}
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
			thisObj := getNativeThis(args)
			event := &Value{Type: "object", Obj: map[string]*Value{
				"type":           {Type: "string", Str: "click"},
				"target":         thisObj,
				"currentTarget":  thisObj,
				"bubbles":        {Type: "bool", Bool: true},
				"cancelable":     {Type: "bool", Bool: true},
				"defaultPrevented": {Type: "bool", Bool: false},
			}}
			if n.EventListeners != nil {
				handlers := n.EventListeners["click"]
				for _, h := range handlers {
					if fn, ok := h.(*Value); ok && fn != nil {
						vm.callFunction(fn, []*Value{event})
					}
				}
			}
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
			offset := nativeThisOffset(args)
			if len(args) <= offset || n == nil {
				return &Value{Type: "null"}
			}
			selector := args[offset].Str
			current := n
			for current != nil {
				matched, _ := dom.MatchesSelector(current, selector)
				if matched {
					return vm.wrapNode(current)
				}
				current = current.Parent
			}
			return &Value{Type: "null"}
		}},
		"matches": {Type: "native", Native: func(args []*Value) *Value {
			if len(args) == 0 || n == nil {
				return &Value{Type: "bool", Bool: false}
			}
			offset := nativeThisOffset(args)
			selector := ""
			if len(args) > offset {
				selector = args[offset].Str
			}
			matched, _ := dom.MatchesSelector(n, selector)
			return &Value{Type: "bool", Bool: matched}
		}},
		"classList": {Type: "object", Obj: map[string]*Value{
			"add": {Type: "native", Native: func(args []*Value) *Value {
				if len(args) > 0 {
					classes := strings.Fields(n.GetAttribute("class"))
					classSet := make(map[string]bool)
					for _, c := range classes {
						classSet[c] = true
					}
					for _, a := range args {
						cls := a.Str
						if cls != "" && !classSet[cls] {
							classes = append(classes, cls)
							classSet[cls] = true
						}
					}
					n.SetAttribute("class", strings.Join(classes, " "))
				}
				return &Value{Type: "undefined"}
			}},
			"remove": {Type: "native", Native: func(args []*Value) *Value {
				if len(args) > 0 {
					classes := strings.Fields(n.GetAttribute("class"))
					removeSet := make(map[string]bool)
					for _, a := range args {
						removeSet[a.Str] = true
					}
					var kept []string
					for _, c := range classes {
						if !removeSet[c] {
							kept = append(kept, c)
						}
					}
					n.SetAttribute("class", strings.Join(kept, " "))
				}
				return &Value{Type: "undefined"}
			}},
			"toggle": {Type: "native", Native: func(args []*Value) *Value {
				if len(args) > 0 {
					cls := args[0].Str
					classes := strings.Fields(n.GetAttribute("class"))
					found := false
					var kept []string
					for _, c := range classes {
						if c == cls {
							found = true
						} else {
							kept = append(kept, c)
						}
					}
					if found {
						n.SetAttribute("class", strings.Join(kept, " "))
						return &Value{Type: "bool", Bool: false}
					}
					kept = append(kept, cls)
					n.SetAttribute("class", strings.Join(kept, " "))
					return &Value{Type: "bool", Bool: true}
				}
				return &Value{Type: "bool", Bool: false}
			}},
			"contains": {Type: "native", Native: func(args []*Value) *Value {
				if len(args) > 0 {
					classes := strings.Fields(n.GetAttribute("class"))
					for _, c := range classes {
						if c == args[0].Str {
							return &Value{Type: "bool", Bool: true}
						}
					}
				}
				return &Value{Type: "bool", Bool: false}
			}},
		}},
		"style": {Type: "object", Obj: func() map[string]*Value {
			styleObj := make(map[string]*Value)
			styleAttr := n.GetAttribute("style")
			if styleAttr != "" {
				for _, part := range strings.Split(styleAttr, ";") {
					kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
					if len(kv) == 2 {
						camelKey := cssToCamel(strings.TrimSpace(kv[0]))
						styleObj[camelKey] = &Value{Type: "string", Str: strings.TrimSpace(kv[1])}
					}
				}
			}
			styleObj["setProperty"] = &Value{Type: "native", Native: func(args []*Value) *Value {
				offset := nativeThisOffset(args)
				if len(args) >= offset+2 {
					prop := args[offset].Str
					val := args[offset+1].Str
					camelKey := cssToCamel(prop)
					currentStyle := n.GetAttribute("style")
					styleMap := parseStyleAttr(currentStyle)
					styleMap[camelKey] = val
					n.SetAttribute("style", buildStyleAttr(styleMap))
				}
				return &Value{Type: "undefined"}
			}}
			styleObj["removeProperty"] = &Value{Type: "native", Native: func(args []*Value) *Value {
				offset := nativeThisOffset(args)
				if len(args) >= offset+1 {
					prop := args[offset].Str
					camelKey := cssToCamel(prop)
					currentStyle := n.GetAttribute("style")
					styleMap := parseStyleAttr(currentStyle)
					oldVal := ""
					if v, ok := styleMap[camelKey]; ok {
						oldVal = v
						delete(styleMap, camelKey)
					}
					n.SetAttribute("style", buildStyleAttr(styleMap))
					return &Value{Type: "string", Str: oldVal}
				}
				return &Value{Type: "string", Str: ""}
			}}
			styleObj["cssText"] = &Value{Type: "string", Str: styleAttr}
			return styleObj
		}()},
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

	// Add descriptors for innerHTML, textContent, and other getter/setter properties
	obj.Descriptors = map[string]*PropertyDescriptor{
		"innerHTML":   {Get: innerHTMLGetter, Set: innerHTMLSetter, Enumerable: true, Configurable: true},
		"textContent": {Get: textContentGetter, Set: textContentSetter, Enumerable: true, Configurable: true},
		"outerHTML": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.OuterHTML()}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 && n.Parent != nil {
				newChildren := htmlparser.ParseFragment(args[valIdx].Str)
				parent := n.Parent
				insertIdx := -1
				for i, c := range parent.Children {
					if c == n {
						insertIdx = i
						break
					}
				}
				if insertIdx >= 0 {
					newSlice := make([]*dom.Node, 0, len(parent.Children)-1+len(newChildren))
					newSlice = append(newSlice, parent.Children[:insertIdx]...)
					for _, c := range newChildren {
						c.Parent = parent
						newSlice = append(newSlice, c)
					}
					newSlice = append(newSlice, parent.Children[insertIdx+1:]...)
					parent.Children = newSlice
				}
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"id": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.GetAttribute("id")}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.SetAttribute("id", args[valIdx].Str)
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"className": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.GetAttribute("class")}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.SetAttribute("class", args[valIdx].Str)
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"value": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.GetAttribute("value")}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.SetAttribute("value", args[valIdx].Str)
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"innerText": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.TextContent()}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.Children = []*dom.Node{
					{Type: dom.TextNode, Data: args[valIdx].Str, Parent: n},
				}
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"style": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			styleStr := n.GetAttribute("style")
			styleObj := make(map[string]*Value)
			styleObj["cssText"] = &Value{Type: "string", Str: styleStr}
			styleObj["display"] = &Value{Type: "string", Str: ""}
			if styleStr != "" {
				for _, part := range strings.Split(styleStr, ";") {
					kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
					if len(kv) == 2 {
						camelKey := cssToCamel(strings.TrimSpace(kv[0]))
						styleObj[camelKey] = &Value{Type: "string", Str: strings.TrimSpace(kv[1])}
					}
				}
			}
			styleObj["setProperty"] = &Value{Type: "native", Native: func(innerArgs []*Value) *Value {
				offset := nativeThisOffset(innerArgs)
				if len(innerArgs) >= offset+2 {
					prop := innerArgs[offset].Str
					val := innerArgs[offset+1].Str
					camelKey := cssToCamel(prop)
					currentStyle := n.GetAttribute("style")
					styleMap := parseStyleAttr(currentStyle)
					styleMap[camelKey] = val
					n.SetAttribute("style", buildStyleAttr(styleMap))
				}
				return &Value{Type: "undefined"}
			}}
			styleObj["removeProperty"] = &Value{Type: "native", Native: func(innerArgs []*Value) *Value {
				offset := nativeThisOffset(innerArgs)
				if len(innerArgs) >= offset+1 {
					prop := innerArgs[offset].Str
					camelKey := cssToCamel(prop)
					currentStyle := n.GetAttribute("style")
					styleMap := parseStyleAttr(currentStyle)
					oldVal := ""
					if v, ok := styleMap[camelKey]; ok {
						oldVal = v
						delete(styleMap, camelKey)
					}
					n.SetAttribute("style", buildStyleAttr(styleMap))
					return &Value{Type: "string", Str: oldVal}
				}
				return &Value{Type: "string", Str: ""}
			}}
			return &Value{Type: "object", Obj: styleObj}
		}}, Enumerable: true, Configurable: true},
		"children": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeList(n.Children)
		}}, Enumerable: true, Configurable: true},
		"childNodes": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeList(n.Children)
		}}, Enumerable: true, Configurable: true},
		"firstChild": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(firstChild(n))
		}}, Enumerable: true, Configurable: true},
		"lastChild": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(lastChild(n))
		}}, Enumerable: true, Configurable: true},
		"parentNode": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(n.Parent)
		}}, Enumerable: true, Configurable: true},
		"parentElement": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			if n.Parent != nil && n.Parent.Type == dom.ElementNode {
				return vm.wrapNodeShallow(n.Parent)
			}
			return &Value{Type: "null"}
		}}, Enumerable: true, Configurable: true},
		"nextSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(n.NextSibling)
		}}, Enumerable: true, Configurable: true},
		"previousSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(n.PrevSibling)
		}}, Enumerable: true, Configurable: true},
		"nextElementSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			sibling := n.NextSibling
			for sibling != nil && sibling.Type != dom.ElementNode {
				sibling = sibling.NextSibling
			}
			return vm.wrapNode(sibling)
		}}, Enumerable: true, Configurable: true},
		"previousElementSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			sibling := n.PrevSibling
			for sibling != nil && sibling.Type != dom.ElementNode {
				sibling = sibling.PrevSibling
			}
			return vm.wrapNode(sibling)
		}}, Enumerable: true, Configurable: true},
		"childElementCount": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			count := 0
			for _, c := range n.Children {
				if c.Type == dom.ElementNode {
					count++
				}
			}
			return &Value{Type: "number", Num: float64(count)}
		}}, Enumerable: true, Configurable: true},
		"offsetWidth": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 100}
		}}, Enumerable: true, Configurable: true},
		"offsetHeight": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 30}
		}}, Enumerable: true, Configurable: true},
		"offsetLeft": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Enumerable: true, Configurable: true},
		"offsetTop": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Enumerable: true, Configurable: true},
		"clientWidth": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 100}
		}}, Enumerable: true, Configurable: true},
		"clientHeight": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 30}
		}}, Enumerable: true, Configurable: true},
		"scrollWidth": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 100}
		}}, Enumerable: true, Configurable: true},
		"scrollHeight": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 30}
		}}, Enumerable: true, Configurable: true},
		"scrollTop": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"scrollLeft": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"offsetParent": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			if n.Parent != nil && n.Parent.Type == dom.ElementNode {
				return vm.wrapNodeShallow(n.Parent)
			}
			return &Value{Type: "null"}
		}}, Enumerable: true, Configurable: true},
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
		// When called via callSetter: args = [obj, val]
		valIdx := len(args) - 1
		if valIdx >= 0 {
			newChildren := htmlparser.ParseFragment(args[valIdx].Str)
			n.Children = newChildren
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
		// When called via callSetter: args = [obj, val]
		valIdx := len(args) - 1
		if valIdx >= 0 {
			n.Children = []*dom.Node{
				{Type: dom.TextNode, Data: args[valIdx].Str, Parent: n},
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
		"getAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "null"}
			}
			return &Value{Type: "string", Str: n.GetAttribute(args[offset].Str)}
		}},
		"setAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) >= offset+2 {
				n.SetAttribute(args[offset].Str, args[offset+1].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"querySelector": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "null"}
			}
			child := dom.QuerySelector(n, args[offset].Str)
			return vm.wrapNodeShallow(child)
		}},
		"querySelectorAll": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "object", Arr: []*Value{}}
			}
			children := dom.QuerySelectorAll(n, args[offset].Str)
			arr := make([]*Value, len(children))
			for i, c := range children {
				arr[i] = vm.wrapNodeShallow(c)
			}
			return &Value{Type: "object", Arr: arr}
		}},
		"hasAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "bool", Bool: false}
			}
			return &Value{Type: "bool", Bool: n.HasAttribute(args[offset].Str)}
		}},
		"removeAttribute": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) > offset {
				n.RemoveAttribute(args[offset].Str)
			}
			return &Value{Type: "undefined"}
		}},
		"appendChild": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "undefined"}
			}
			child := args[offset]
			if childNode := vm.unwrapNode(child); childNode != nil {
				n.AppendChild(childNode)
				return child
			}
			return &Value{Type: "undefined"}
		}},
		"insertBefore": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return &Value{Type: "undefined"}
			}
			newNode := args[offset]
			refNode := args[offset+1]
			newDomNode := vm.unwrapNode(newNode)
			refDomNode := vm.unwrapNode(refNode)
			if newDomNode != nil {
				if newDomNode.Parent != nil {
					for i, c := range newDomNode.Parent.Children {
						if c == newDomNode {
							newDomNode.Parent.Children = append(newDomNode.Parent.Children[:i], newDomNode.Parent.Children[i+1:]...)
							break
						}
					}
				}
				if refDomNode != nil {
					for i, c := range n.Children {
						if c == refDomNode {
							newDomNode.Parent = n
							n.Children = append(n.Children[:i], append([]*dom.Node{newDomNode}, n.Children[i:]...)...)
							return newNode
						}
					}
				}
				n.AppendChild(newDomNode)
			}
			return newNode
		}},
		"removeChild": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "undefined"}
			}
			child := args[offset]
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
		"replaceChild": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset+1 {
				return &Value{Type: "undefined"}
			}
			newNode := args[offset]
			oldNode := args[offset+1]
			newDomNode := vm.unwrapNode(newNode)
			oldDomNode := vm.unwrapNode(oldNode)
			if newDomNode != nil && oldDomNode != nil {
				if newDomNode.Parent != nil {
					for i, c := range newDomNode.Parent.Children {
						if c == newDomNode {
							newDomNode.Parent.Children = append(newDomNode.Parent.Children[:i], newDomNode.Parent.Children[i+1:]...)
							break
						}
					}
				}
				for i, c := range n.Children {
					if c == oldDomNode {
						newDomNode.Parent = n
						n.Children[i] = newDomNode
						oldDomNode.Parent = nil
						break
					}
				}
			}
			return oldNode
		}},
		"cloneNode": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			deep := false
			if len(args) > offset {
				deep = args[offset].Bool
			}
			clone := vm.cloneNode(n, deep)
			return vm.wrapNodeShallow(clone)
		}},
		"contains": {Type: "native", Native: func(args []*Value) *Value {
			offset := nativeThisOffset(args)
			if len(args) <= offset {
				return &Value{Type: "bool", Bool: false}
			}
			other := vm.unwrapNode(args[offset])
			return &Value{Type: "bool", Bool: vm.nodeContains(n, other)}
		}},
		"addEventListener": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"removeEventListener": {Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}},
		"firstChild": {Type: "null"},
		"lastChild": {Type: "null"},
		"parentNode": {Type: "null"},
		"parentElement": {Type: "null"},
		"nextSibling": {Type: "null"},
		"previousSibling": {Type: "null"},
		"childElementCount": {Type: "number", Num: 0},
	}}

	obj.Descriptors = map[string]*PropertyDescriptor{
		"innerHTML":   {Get: innerHTMLGetter, Set: innerHTMLSetter, Enumerable: true, Configurable: true},
		"textContent": {Get: textContentGetter, Set: textContentSetter, Enumerable: true, Configurable: true},
		"outerHTML": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.OuterHTML()}
		}}, Enumerable: true, Configurable: true},
		"id": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.GetAttribute("id")}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.SetAttribute("id", args[valIdx].Str)
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"className": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.GetAttribute("class")}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.SetAttribute("class", args[valIdx].Str)
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"value": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.GetAttribute("value")}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.SetAttribute("value", args[valIdx].Str)
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"innerText": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "string", Str: n.TextContent()}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			valIdx := len(args) - 1
			if valIdx >= 0 {
				n.Children = []*dom.Node{
					{Type: dom.TextNode, Data: args[valIdx].Str, Parent: n},
				}
			}
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"style": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			styleStr := n.GetAttribute("style")
			styleObj := make(map[string]*Value)
			styleObj["cssText"] = &Value{Type: "string", Str: styleStr}
			styleObj["display"] = &Value{Type: "string", Str: ""}
			if styleStr != "" {
				for _, part := range strings.Split(styleStr, ";") {
					kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
					if len(kv) == 2 {
						camelKey := cssToCamel(strings.TrimSpace(kv[0]))
						styleObj[camelKey] = &Value{Type: "string", Str: strings.TrimSpace(kv[1])}
					}
				}
			}
			styleObj["setProperty"] = &Value{Type: "native", Native: func(innerArgs []*Value) *Value {
				offset := nativeThisOffset(innerArgs)
				if len(innerArgs) >= offset+2 {
					prop := innerArgs[offset].Str
					val := innerArgs[offset+1].Str
					camelKey := cssToCamel(prop)
					currentStyle := n.GetAttribute("style")
					styleMap := parseStyleAttr(currentStyle)
					styleMap[camelKey] = val
					n.SetAttribute("style", buildStyleAttr(styleMap))
				}
				return &Value{Type: "undefined"}
			}}
			styleObj["removeProperty"] = &Value{Type: "native", Native: func(innerArgs []*Value) *Value {
				offset := nativeThisOffset(innerArgs)
				if len(innerArgs) >= offset+1 {
					prop := innerArgs[offset].Str
					camelKey := cssToCamel(prop)
					currentStyle := n.GetAttribute("style")
					styleMap := parseStyleAttr(currentStyle)
					oldVal := ""
					if v, ok := styleMap[camelKey]; ok {
						oldVal = v
						delete(styleMap, camelKey)
					}
					n.SetAttribute("style", buildStyleAttr(styleMap))
					return &Value{Type: "string", Str: oldVal}
				}
				return &Value{Type: "string", Str: ""}
			}}
			return &Value{Type: "object", Obj: styleObj}
		}}, Enumerable: true, Configurable: true},
		"children": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeListShallow(n.Children)
		}}, Enumerable: true, Configurable: true},
		"childNodes": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeListShallow(n.Children)
		}}, Enumerable: true, Configurable: true},
		"firstChild": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(firstChild(n))
		}}, Enumerable: true, Configurable: true},
		"lastChild": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(lastChild(n))
		}}, Enumerable: true, Configurable: true},
		"parentNode": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(n.Parent)
		}}, Enumerable: true, Configurable: true},
		"parentElement": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			if n.Parent != nil && n.Parent.Type == dom.ElementNode {
				return vm.wrapNodeShallow(n.Parent)
			}
			return &Value{Type: "null"}
		}}, Enumerable: true, Configurable: true},
		"nextSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(n.NextSibling)
		}}, Enumerable: true, Configurable: true},
		"previousSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNodeShallow(n.PrevSibling)
		}}, Enumerable: true, Configurable: true},
		"nextElementSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			sibling := n.NextSibling
			for sibling != nil && sibling.Type != dom.ElementNode {
				sibling = sibling.NextSibling
			}
			if sibling != nil {
				return vm.wrapNodeShallow(sibling)
			}
			return &Value{Type: "null"}
		}}, Enumerable: true, Configurable: true},
		"previousElementSibling": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			sibling := n.PrevSibling
			for sibling != nil && sibling.Type != dom.ElementNode {
				sibling = sibling.PrevSibling
			}
			if sibling != nil {
				return vm.wrapNodeShallow(sibling)
			}
			return &Value{Type: "null"}
		}}, Enumerable: true, Configurable: true},
		"childElementCount": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			count := 0
			for _, c := range n.Children {
				if c.Type == dom.ElementNode {
					count++
				}
			}
			return &Value{Type: "number", Num: float64(count)}
		}}, Enumerable: true, Configurable: true},
		"offsetWidth": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 100}
		}}, Enumerable: true, Configurable: true},
		"offsetHeight": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 30}
		}}, Enumerable: true, Configurable: true},
		"offsetLeft": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Enumerable: true, Configurable: true},
		"offsetTop": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Enumerable: true, Configurable: true},
		"clientWidth": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 100}
		}}, Enumerable: true, Configurable: true},
		"clientHeight": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 30}
		}}, Enumerable: true, Configurable: true},
		"scrollWidth": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 100}
		}}, Enumerable: true, Configurable: true},
		"scrollHeight": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 30}
		}}, Enumerable: true, Configurable: true},
		"scrollTop": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"scrollLeft": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "number", Num: 0}
		}}, Set: &Value{Type: "native", Native: func(args []*Value) *Value {
			return &Value{Type: "undefined"}
		}}, Enumerable: true, Configurable: true},
		"offsetParent": {Get: &Value{Type: "native", Native: func(args []*Value) *Value {
			if n.Parent != nil && n.Parent.Type == dom.ElementNode {
				return vm.wrapNodeShallow(n.Parent)
			}
			return &Value{Type: "null"}
		}}, Enumerable: true, Configurable: true},
	}
	// Store the dom.Node reference for unwrapNode (used by DOM mutation methods)
	obj.NodeRef = n

	return obj
}

// newArray creates a new array Value with the ArrayPrototype link.
func (vm *VM) newArray(elements []*Value) *Value {
	if elements == nil {
		elements = []*Value{}
	}
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

	// Only reset step counter when starting a fresh execution, not during nested calls
	if !vm.running {
		vm.ResetSteps()
	}
	vm.running = true
	defer func() { vm.running = false }()

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
				vm.debugLog("RECOVERED panic: %T: %v", r, r)
				switch v := r.(type) {
				case *JSException:
					execErr = fmt.Errorf("JS error: %s", valueToString(v.Value))
				default:
					execErr = fmt.Errorf("execution error: %v", r)
				}
			}
		}()

	// Hoist function declarations in the global scope
	vm.hoistFunctionDecls(prog.Statements)

	for i, stmt := range prog.Statements {
			if returning {
				vm.debugLog("Breaking due to returning flag")
				break
			}
			vm.debugLog("Executing statement %d: %T", i+1, stmt)
			result, returning, returnVal = vm.execStmt(stmt)
			vm.debugLog("Statement %d done: returning=%v, resultType=%s", i+1, returning, result.Type)
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
		if len(s.Decls) > 0 {
			var lastVal *Value = &Value{Type: "undefined"}
			for _, d := range s.Decls {
				if d.IsDestructuring && d.DestructPattern != nil {
					vm.execDestructuring(d.DestructPattern, d.Value)
					continue
				}
				var val *Value = &Value{Type: "undefined"}
				if d.Value != nil {
					val = vm.evalExpr(d.Value)
				}
				vm.env.Define(d.Name, val)
				lastVal = val
			}
			return lastVal, false, nil
		}
		// Legacy path for VarDecls without Decls
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
			vm.checkSteps()
			if s.Cond != nil {
				cond := vm.evalExpr(s.Cond)
				if !isTruthy(cond) {
					break
				}
			}
			vm.forIterCount++
			if vm.forIterMax > 0 && vm.forIterCount > vm.forIterMax {
				break
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
			// Collect all enumerable keys from Obj, Descriptors, and prototype chain
			keys := make(map[string]bool)
			current := obj
			depth := 0
			for current != nil && current.Type == "object" {
				if current.Obj != nil {
					for key := range current.Obj {
						if !keys[key] {
							// Check if this property has a non-enumerable descriptor
							if current.Descriptors != nil {
								if desc, ok := current.Descriptors[key]; ok && !desc.Enumerable {
									continue
								}
							}
							// Skip built-in prototype methods on non-own objects (depth > 0)
							if depth > 0 && isBuiltinPrototypeProperty(key) {
								continue
							}
							keys[key] = true
						}
					}
				}
				if current.Descriptors != nil {
					for key, desc := range current.Descriptors {
						if desc.Enumerable && !keys[key] {
							keys[key] = true
						}
					}
				}
				if current.Arr != nil {
					for i := range current.Arr {
						key := strconv.Itoa(i)
						if !keys[key] {
							keys[key] = true
						}
					}
				}
				current = current.Proto
				depth++
				if depth > 100 {
					break
				}
			}
			// Debug: log large for-in iterations
			if len(keys) > 100 && vm.debug {
				arrLen := 0
				if obj.Arr != nil {
					arrLen = len(obj.Arr)
				}
				fmt.Printf("[HLBR JS DEBUG] for-in: %d keys, obj type properties: %d, arr len: %d, depth: %d\n",
					len(keys), len(obj.Obj), arrLen, depth)
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
			vm.checkSteps()
			cond := vm.evalExpr(s.Cond)
			if !isTruthy(cond) {
				break
			}
			vm.forIterCount++
			if vm.forIterMax > 0 && vm.forIterCount > vm.forIterMax {
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
	case *SwitchStmt:
		disc := vm.evalExpr(s.Discriminant)
		matched := false
		var defaultIdx int = -1
		for i, c := range s.Cases {
			if c.Test == nil {
				defaultIdx = i
				continue
			}
			testVal := vm.evalExpr(c.Test)
			if !matched && valuesEqual(disc, testVal) {
				matched = true
			}
			if matched {
				for _, stmt := range c.Consequent {
					val, brk, retVal := vm.execStmt(stmt)
					if brk {
						if retVal != nil {
							return val, true, retVal
						}
						return &Value{Type: "undefined"}, false, nil
					}
				}
			}
		}
		if !matched && defaultIdx >= 0 {
			for i := defaultIdx; i < len(s.Cases); i++ {
				for _, stmt := range s.Cases[i].Consequent {
					val, brk, retVal := vm.execStmt(stmt)
					if brk {
						if retVal != nil {
							return val, true, retVal
						}
						return &Value{Type: "undefined"}, false, nil
					}
				}
			}
		}
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
	case *WithStmt:
		// With statement: with (obj) { body }
		// Creates a scope where obj's properties are checked first during variable lookup
		obj := vm.evalExpr(s.Object)
		var withResult *Value = &Value{Type: "undefined"}
		var withReturning bool
		var withRetVal *Value
		if obj == nil || (obj.Type != "object" && obj.Type != "function") {
			for _, stmt := range s.Body {
				withResult, withReturning, withRetVal = vm.execStmt(stmt)
				if withReturning {
					return withResult, withReturning, withRetVal
				}
			}
			return withResult, false, nil
		}
		withEnv := NewEnvironment(vm.env)
		withEnv.withObj = obj
		withEnv.vm = vm
		oldEnv := vm.env
		vm.env = withEnv
		for _, stmt := range s.Body {
			withResult, withReturning, withRetVal = vm.execStmt(stmt)
			if withReturning {
				vm.env = oldEnv
				return withResult, withReturning, withRetVal
			}
		}
		vm.env = oldEnv
		return withResult, false, nil
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

	// Hoist function declarations to the top of the scope
	vm.hoistFunctionDecls(stmts)

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

// hoistFunctionDecls scans statements for function declarations and defines
// them in the current scope before any code executes, implementing JavaScript's
// function hoisting behavior.
func (vm *VM) hoistFunctionDecls(stmts []Statement) {
	hoisted := 0
	for _, stmt := range stmts {
		if fd, ok := stmt.(*FunctionDecl); ok {
			hoisted++
			fn := &Function{Params: fd.Params, DefaultVals: fd.DefaultVals, RestParam: fd.RestParam, Body: fd.Body, Env: vm.env}
			val := &Value{Type: "function", Func: fn}
			if fd.IsAsync {
				val.IsAsync = true
			}
			protoObj := &Value{Type: "object", Obj: make(map[string]*Value)}
			if objProto := vm.env.Get("ObjectPrototype"); objProto.Type == "object" {
				protoObj.Proto = objProto
			}
			protoObj.Obj["constructor"] = val
			val.PrototypeObj = protoObj
			vm.env.Define(fd.Name, val)
		}
	}
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
	case *RegexLit:
		return newRegExp(e.Pattern, e.Flags)
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
		result := vm.newArray(arr)
		result.BuiltInConstructor = "Array"
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
		result := &Value{Type: "object", Obj: obj, BuiltInConstructor: "Object"}
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
		obj := vm.evalExpr(e.Object)
		if obj.Type == "undefined" || obj.Type == "null" {
			return &Value{Type: "undefined"}
		}
		if e.Computed {
			prop := vm.evalExpr(e.Property)
			key := valueToString(prop)
			if (obj.Type == "object" || obj.Type == "function" || obj.Type == "native") && obj.Obj != nil {
				if val, ok := obj.Obj[key]; ok {
					return val
				}
			}
		} else {
			if ident, ok := e.Property.(*Ident); ok {
				if (obj.Type == "object" || obj.Type == "function" || obj.Type == "native") && obj.Obj != nil {
					if val, ok := obj.Obj[ident.Name]; ok {
						return val
					}
				}
			}
		}
		return &Value{Type: "undefined"}
	case *NullishCoalescingExpr:
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
		key := valueToString(prop)
		switch obj.Type {
		case "object", "function", "native":
			if obj.Obj != nil {
				if _, ok := obj.Obj[key]; ok {
					return &Value{Type: "bool", Bool: true}
				}
			}
			if obj.Descriptors != nil {
				if _, ok := obj.Descriptors[key]; ok {
					return &Value{Type: "bool", Bool: true}
				}
			}
			// Walk the prototype chain for the 'in' operator
			current := obj.Proto
			for current != nil {
				if current.Obj != nil {
					if _, ok := current.Obj[key]; ok {
						return &Value{Type: "bool", Bool: true}
					}
				}
				if current.Descriptors != nil {
					if _, ok := current.Descriptors[key]; ok {
						return &Value{Type: "bool", Bool: true}
					}
				}
				current = current.Proto
			}
		case "string":
			return &Value{Type: "bool", Bool: key == "length" || isStringMethod(key)}
		case "arrayMethod", "stringMethod":
			return &Value{Type: "bool", Bool: false}
		}
		return &Value{Type: "bool", Bool: false}
	case *SequenceExpr:
		var result *Value = &Value{Type: "undefined"}
		for _, expr := range e.Expressions {
			result = vm.evalExpr(expr)
		}
		return result
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
	// Short-circuit operators must evaluate lazily
	switch e.Op {
	case "&&":
		left := vm.evalExpr(e.Left)
		if !isTruthy(left) {
			return left
		}
		return vm.evalExpr(e.Right)
	case "||":
		left := vm.evalExpr(e.Left)
		if isTruthy(left) {
			return left
		}
		return vm.evalExpr(e.Right)
	}

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
	case "&":
		l, r := toInt32(left), toInt32(right)
		return &Value{Type: "number", Num: float64(l & r)}
	case "|":
		l, r := toInt32(left), toInt32(right)
		return &Value{Type: "number", Num: float64(l | r)}
	case "^":
		l, r := toInt32(left), toInt32(right)
		return &Value{Type: "number", Num: float64(l ^ r)}
	case "<<":
		l, r := toInt32(left), toUint32(right)
		return &Value{Type: "number", Num: float64(l << (r & 31))}
	case ">>":
		l, r := toInt32(left), toUint32(right)
		return &Value{Type: "number", Num: float64(l >> (r & 31))}
	case ">>>":
		l, r := toUint32(left), toUint32(right)
		return &Value{Type: "number", Num: float64(l >> (r & 31))}
	}
	return &Value{Type: "undefined"}
}

func (vm *VM) evalUnary(e *UnaryExpr) *Value {
	switch e.Op {
	case "delete":
		return vm.evalDelete(e.Expr)
	case "!":
		val := vm.evalExpr(e.Expr)
		return &Value{Type: "bool", Bool: !isTruthy(val)}
	case "-":
		val := vm.evalExpr(e.Expr)
		return &Value{Type: "number", Num: -val.Num}
	case "+":
		val := vm.evalExpr(e.Expr)
		return &Value{Type: "number", Num: val.Num}
	case "typeof":
		// typeof doesn't evaluate the expression for undefined variables
		if ident, ok := e.Expr.(*Ident); ok {
			val := vm.env.Get(ident.Name)
			if val == nil || val.Type == "undefined" {
				return &Value{Type: "string", Str: "undefined"}
			}
			return &Value{Type: "string", Str: jsTypeOf(val)}
		}
		val := vm.evalExpr(e.Expr)
		return &Value{Type: "string", Str: jsTypeOf(val)}
	case "void":
		vm.evalExpr(e.Expr)
		return &Value{Type: "undefined"}
	case "~":
		val := vm.evalExpr(e.Expr)
		return &Value{Type: "number", Num: float64(^toInt32(val))}
	}
	val := vm.evalExpr(e.Expr)
	return val
}

// evalDelete handles the delete operator which removes a property from an object.
func (vm *VM) evalDelete(expr Expression) *Value {
	switch e := expr.(type) {
	case *MemberExpr:
		obj := vm.evalExpr(e.Object)
		if obj == nil || (obj.Type != "object" && obj.Type != "function" && obj.Type != "native") {
			return &Value{Type: "bool", Bool: true}
		}
		var prop string
		if e.Computed {
			prop = valueToString(vm.evalExpr(e.Property))
		} else if ident, ok := e.Property.(*Ident); ok {
			prop = ident.Name
		}
		if obj.Obj != nil {
			// Check if property is non-configurable before deleting
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[prop]; ok && !desc.Configurable {
					return &Value{Type: "bool", Bool: false}
				}
			}
			delete(obj.Obj, prop)
		}
		if obj.Descriptors != nil {
			delete(obj.Descriptors, prop)
		}
		return &Value{Type: "bool", Bool: true}
	case *Ident:
		// delete on a global variable: remove from environment
		vm.env.Delete(e.Name)
		return &Value{Type: "bool", Bool: true}
	default:
		// delete on non-member expressions returns true
		return &Value{Type: "bool", Bool: true}
	}
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
		// For assignment, we need the actual object reference (not a bindThis copy).
		// evalMember may return a copy for function/native types (for method binding),
		// but assignment must modify the original.
		obj := vm.evalMemberRaw(left.Object)

		// Handle Proxy set trap
		if obj.Type == "proxy" && obj.Proxy != nil {
			return vm.evalProxySet(obj.Proxy, left, val)
		}

		if obj.Type == "object" || obj.Type == "function" || obj.Type == "native" {
			// Check if object is frozen
			if obj.Frozen {
				return val
			}
			// Initialize Obj map if needed (functions can have properties in JS)
			if obj.Obj == nil {
				obj.Obj = make(map[string]*Value)
			}
			prop := ""
			if left.Computed {
				prop = valueToString(vm.evalExpr(left.Property))
			} else if ident, ok := left.Property.(*Ident); ok {
				prop = ident.Name
			}
			// Check for setter via descriptor
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[prop]; ok {
					// If there's a setter, invoke it
					if desc.Set != nil {
						vm.callSetter(desc.Set, obj, val)
						return val
					}
					// If not writable, silently ignore the assignment (strict mode would throw)
					if !desc.Writable {
						return val
					}
				}
			}
		obj.Obj[prop] = val
		// When assigning to a property that has a data descriptor,
		// also update the descriptor's Value to keep it in sync.
		// Without this, the descriptor's stale Value takes priority
		// over Obj during property reads, causing reads to return
		// the old value instead of the newly assigned one.
		if obj.Descriptors != nil {
			if desc, ok := obj.Descriptors[prop]; ok && desc.Get == nil && desc.Set == nil {
				desc.Value = val
			}
		}
		// When assigning to a function's .prototype property, also update PrototypeObj
		// so that evalNew uses the updated prototype chain (e.g., after Sub.prototype = Object.create(Super.prototype))
		if prop == "prototype" && (obj.Type == "function" || obj.Type == "native") {
			obj.PrototypeObj = val
		}
	}

	return val
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

	// Handle Object(value) call - per ECMAScript spec, Object(value) returns
	// value itself if it's already an object. For primitives, it wraps them.
	// Critical for webpack code like Object(v.c)() where Object(fn) must
	// return fn itself.
	if callee.Type == "native" && callee.BuiltInConstructor == "Object" && callee.Native == nil {
		if len(args) > 0 {
			val := args[0]
			switch val.Type {
			case "undefined", "null":
				obj := make(map[string]*Value)
				return &Value{Type: "object", Obj: obj, BuiltInConstructor: "Object"}
			case "string":
				obj := make(map[string]*Value)
				obj["valueOf"] = &Value{Type: "native", Native: func(a2 []*Value) *Value { return val }}
				obj["toString"] = &Value{Type: "native", Native: func(a2 []*Value) *Value { return val }}
				obj["length"] = &Value{Type: "number", Num: float64(len(val.Str))}
				for i, ch := range val.Str {
					idx := strconv.Itoa(i)
					char := string(ch)
					obj[idx] = &Value{Type: "string", Str: char}
				}
				return &Value{Type: "object", Obj: obj, BuiltInConstructor: "String"}
			case "number", "bool", "bigint":
				obj := make(map[string]*Value)
				obj["valueOf"] = &Value{Type: "native", Native: func(a2 []*Value) *Value { return val }}
				obj["toString"] = &Value{Type: "native", Native: func(a2 []*Value) *Value { return &Value{Type: "string", Str: valueToString(val)} }}
				return &Value{Type: "object", Obj: obj, BuiltInConstructor: "Object"}
			default:
				// Per ECMAScript spec, Object(value) returns value for objects.
				// However, returning the same pointer for constructor objects
				// (Number, RegExp, Array, etc.) causes core-js polyfill to loop
				// endlessly (it iterates over the object's properties and re-processes).
				// Return val directly for: functions, native functions without
				// BuiltInConstructor (these are regular closures, not constructors),
				// and primitive wrappers (Object/Number/String/Boolean).
				// Do NOT return constructors like Number, RegExp, Array directly.
				if val.Type == "function" {
					return val
				}
				if val.Type == "native" && val.BuiltInConstructor == "" {
					return val
				}
				if val.BuiltInConstructor == "Object" || val.BuiltInConstructor == "String" || val.BuiltInConstructor == "Number" || val.BuiltInConstructor == "Boolean" {
					return val
				}
				obj := make(map[string]*Value)
				return &Value{Type: "object", Obj: obj, BuiltInConstructor: "Object"}
			}
		}
		obj := make(map[string]*Value)
		return &Value{Type: "object", Obj: obj, BuiltInConstructor: "Object"}
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

		// Create the 'arguments' object (array-like with numeric indices, length, and callee)
		argsObj := &Value{Type: "object", Arr: args, Obj: make(map[string]*Value)}
		argsObj.Obj["length"] = &Value{Type: "number", Num: float64(len(args))}
		argsObj.Obj["callee"] = &Value{Type: "function", Func: fn}
		childEnv.Define("arguments", argsObj)

		oldEnv := vm.env
		vm.env = childEnv

		// Track recursion depth — include first statement type for identification
		var firstStmt string
		if len(fn.Body) > 0 {
			firstStmt = fmt.Sprintf(" %T", fn.Body[0])
		}
		vm.callStack = append(vm.callStack, fmt.Sprintf("fn(%v)%s", fn.Params, firstStmt))
		vm.enterCall()
		defer vm.exitCall()

		// Hoist function declarations in the function body
		vm.hoistFunctionDecls(fn.Body)

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

// evalMemberRaw evaluates a MemberExpr and returns the raw object value
// without bindThis copying. Used by evalAssign to ensure property writes
// go to the actual object, not a temporary copy.
func (vm *VM) evalMemberRaw(expr Expression) *Value {
	if me, ok := expr.(*MemberExpr); ok {
		obj := vm.evalMemberRaw(me.Object)
		if obj == nil {
			return &Value{Type: "undefined"}
		}
		var key string
		if me.Computed {
			key = valueToString(vm.evalExpr(me.Property))
		} else if ident, ok := me.Property.(*Ident); ok {
			key = ident.Name
		}
		// Handle function.prototype — return the constructor's prototype object
		if (obj.Type == "function" || obj.Type == "native") && key == "prototype" {
			if obj.PrototypeObj != nil {
				return obj.PrototypeObj
			}
			protoObj := &Value{Type: "object", Obj: make(map[string]*Value)}
			if objProto := vm.env.Get("ObjectPrototype"); objProto != nil && objProto.Type == "object" {
				protoObj.Proto = objProto
			}
			protoObj.Obj["constructor"] = obj
			obj.PrototypeObj = protoObj
			return protoObj
		}
		if obj.Obj != nil {
			if v, ok := obj.Obj[key]; ok {
				return v
			}
		}
		// Handle descriptors (getters should be called for raw access too)
		if obj.Descriptors != nil {
			if desc, ok := obj.Descriptors[key]; ok {
				if desc.Get != nil {
					return vm.callGetter(desc.Get, obj)
				}
				if desc.Value != nil {
					return desc.Value
				}
			}
		}
		if obj.Type == "object" && obj.Arr != nil {
			if key == "length" {
				return &Value{Type: "number", Num: float64(len(obj.Arr))}
			}
		}
		return &Value{Type: "undefined"}
	}
	return vm.evalExpr(expr)
}

func (vm *VM) evalMember(e *MemberExpr) *Value {
	obj := vm.evalExpr(e.Object)

	// Precompute the property key for computed member expressions.
	// This is crucial for prototype chain lookups: re-evaluating e.Property
	// inside lookupInPrototypeWithKey would use the wrong environment scope,
	// causing variable-based computed keys to resolve incorrectly.
	computedKey := ""
	var computedProp *Value
	if e.Computed {
		computedProp = vm.evalExpr(e.Property)
		computedKey = valueToString(computedProp)
	} else if ident, ok := e.Property.(*Ident); ok {
		computedKey = ident.Name
	}

	// Property access on null/undefined: in real browsers, this throws TypeError.
	// However, we currently can't throw here because too much existing code and
	// library code relies on safe property access patterns like `obj && obj.foo`.
	// The vhPU patch in browser.go handles the core-js case instead.
	// TODO: Re-enable TypeError once all downstream code is fixed.
	_ = obj

	// Handle Proxy get trap
	if obj.Type == "proxy" && obj.Proxy != nil {
		return vm.evalProxyGet(obj.Proxy, e)
	}

	// Handle native functions with Obj field (e.g., Array.isArray, Promise.resolve)
	// These are constructors/static methods attached to native functions
	if obj.Type == "native" {
		if ident, ok := e.Property.(*Ident); ok {
			// Check for getter via descriptor first
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[ident.Name]; ok && desc.Get != nil {
					return vm.callGetter(desc.Get, obj)
				}
			}
			if obj.Obj != nil {
				if v, ok := obj.Obj[ident.Name]; ok {
					if (v.Type == "function" || v.Type == "native") && v.ThisBinding == nil {
						return vm.bindThis(v, obj)
					}
					return v
				}
			}
		}
	}

	// Handle function types with Obj field (functions are objects in JS)
	if obj.Type == "function" {
		if ident, ok := e.Property.(*Ident); ok {
			// Check for getter via descriptor first
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[ident.Name]; ok && desc.Get != nil {
					return vm.callGetter(desc.Get, obj)
				}
			}
			if obj.Obj != nil {
				if v, ok := obj.Obj[ident.Name]; ok {
					if (v.Type == "function" || v.Type == "native") && v.ThisBinding == nil {
						return vm.bindThis(v, obj)
					}
					return v
				}
			}
		}
	}

	// Handle Map and Set types - they store methods in Obj field
	if obj.Type == "map" || obj.Type == "set" || obj.Type == "weakmap" || obj.Type == "weakset" || obj.Type == "iterator" {
		if obj.Obj != nil {
			if ident, ok := e.Property.(*Ident); ok {
				if v, ok := obj.Obj[ident.Name]; ok {
					if v.Type == "native" || v.Type == "function" {
						return vm.bindThis(v, obj)
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
				if v.Type == "native" || v.Type == "function" {
					return vm.bindThis(v, obj)
				}
				return v
			}
		}
	}

	if obj.Type == "object" {
		if e.Computed {
			key := computedKey
			prop := computedProp
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[key]; ok && desc.Get != nil {
					return vm.callGetter(desc.Get, obj)
				}
			}
			if obj.Obj != nil {
				if v, ok := obj.Obj[key]; ok {
					if v.Type == "function" || v.Type == "native" {
						return vm.bindThis(v, obj)
					}
					return v
				}
			}
			// Check for Symbol.iterator on arrays - look up in ArrayPrototype
			if obj.Arr != nil && prop.Type == "symbol" {
				arrayProto := vm.env.Get("ArrayPrototype")
				if arrayProto != nil && arrayProto.Obj != nil {
					if v, ok := arrayProto.Obj[key]; ok {
						if v.Type == "native" || v.Type == "function" {
							return vm.bindThis(v, obj)
						}
						return v
					}
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
			if obj.Descriptors != nil {
				if desc, ok := obj.Descriptors[ident.Name]; ok && desc.Get != nil {
					return vm.callGetter(desc.Get, obj)
				}
			}
				if obj.Obj != nil {
					if v, ok := obj.Obj[ident.Name]; ok {
						if v.Type == "function" || v.Type == "native" {
							return vm.bindThis(v, obj)
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
	if (obj.Type == "function" && obj.Func != nil) || (obj.Type == "native" && obj.Native != nil) || (obj.Type == "arrayMethod" && obj.Native != nil) {
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
				fn := obj.Func
				nativeFn := obj.Native
				fnType := obj.Type
				methodName := obj.Str
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
						// For arrayMethod, use callArrayMethod directly with the new this
						if fnType == "arrayMethod" && methodName != "" && boundThis != nil {
							return callArrayMethod(methodName, boundThis, fnArgs, vm)
						}
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
				fn := obj.Func
				nativeFn := obj.Native
				fnType := obj.Type
				methodName := obj.Str
				originalThis := obj.ThisBinding
				return &Value{
					Type: "native",
					Native: func(args []*Value) *Value {
						var boundThis *Value = originalThis
						var boundArgs []*Value
						if len(args) > 0 {
							boundThis = args[0]
							boundArgs = args[1:]
						}
						// For native/arrayMethod functions, return a native bound wrapper
						if fnType == "arrayMethod" && methodName != "" {
							return &Value{
								Type: "native",
								Native: func(innerArgs []*Value) *Value {
									callArgs := innerArgs
									if len(innerArgs) > 0 && innerArgs[0]._isThisArg {
										callArgs = innerArgs[1:]
									}
									allArgs := make([]*Value, 0, len(boundArgs)+len(callArgs))
									allArgs = append(allArgs, boundArgs...)
									allArgs = append(allArgs, callArgs...)
									target := boundThis
									if target == nil {
										target = originalThis
									}
									if target != nil {
										return callArrayMethod(methodName, target, allArgs, vm)
									}
									return &Value{Type: "undefined"}
								},
							}
						}
						if (fnType == "native") && nativeFn != nil {
							return &Value{
								Type: "native",
								Native: func(innerArgs []*Value) *Value {
									// Strip any this-binding prepended by evalCall
									callArgs := innerArgs
									if len(innerArgs) > 0 && innerArgs[0]._isThisArg {
										callArgs = innerArgs[1:]
									}
									allArgs := make([]*Value, 0, len(boundArgs)+len(callArgs)+1)
									if boundThis != nil {
										boundThisCopy := &Value{}
										*boundThisCopy = *boundThis
										boundThisCopy._isThisArg = true
										allArgs = append(allArgs, boundThisCopy)
									}
									allArgs = append(allArgs, boundArgs...)
									allArgs = append(allArgs, callArgs...)
									return nativeFn(allArgs)
								},
							}
						}
						// For user-defined functions
						return &Value{
							Type:        "function",
							Func:        fn,
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
		return vm.lookupInPrototypeWithKey(obj.Proto, e, obj, computedKey)
	}

	// Function types: look up in FunctionPrototype
	if obj.Type == "function" || obj.Type == "native" || obj.Type == "arrayMethod" {
		if fnProto := vm.env.Get("FunctionPrototype"); fnProto.Type == "object" {
			result := vm.lookupInPrototypeWithKey(fnProto, e, obj, computedKey)
			if result.Type != "undefined" {
				return result
			}
		}
	}

	return &Value{Type: "undefined"}
}

// lookupInPrototype walks the prototype chain to find a property.
// Returns undefined if not found in any prototype.
// The originalObj parameter is used to set ThisBinding on found methods,
// so that when obj.hasOwnProperty() is called, 'this' refers to obj, not the prototype.
func (vm *VM) lookupInPrototype(proto *Value, e *MemberExpr) *Value {
	return vm.lookupInPrototypeWithKey(proto, e, nil, "")
}

// lookupInPrototypeWithKey walks the prototype chain with an explicit original object
// and an optional precomputed key string. When precomputedKey is non-empty, it is used
// directly instead of re-evaluating e.Property. This is critical for computed member
// expressions like obj[varKey] where re-evaluating the property expression would use
// the wrong environment scope.
func (vm *VM) lookupInPrototypeWithKey(proto *Value, e *MemberExpr, originalObj *Value, precomputedKey string) *Value {
	if proto == nil || proto.Type != "object" {
		return &Value{Type: "undefined"}
	}

	var key string
	if precomputedKey != "" {
		key = precomputedKey
	} else if ident, ok := e.Property.(*Ident); ok {
		key = ident.Name
	} else if e.Computed {
		key = valueToString(vm.evalExpr(e.Property))
	}

		if key != "" && proto.Obj != nil {
		if v, ok := proto.Obj[key]; ok {
			thisVal := originalObj
			if thisVal == nil {
				thisVal = proto
			}
			if v.Type == "function" || v.Type == "native" {
				return &Value{
					Type:        v.Type,
					Native:      v.Native,
					Func:        v.Func,
					Obj:         v.Obj,
					ThisBinding: thisVal,
				}
			}
			return v
		}
	}

	// Check descriptors in the prototype chain (for Object.defineProperty getters/setters)
	if key != "" && proto.Descriptors != nil {
		if desc, ok := proto.Descriptors[key]; ok {
			thisVal := originalObj
			if thisVal == nil {
				thisVal = proto
			}
			if desc.Get != nil {
				return vm.callGetter(desc.Get, thisVal)
			}
			if desc.Value != nil {
				return desc.Value
			}
		}
	}

	// Walk up the prototype chain
	if proto.Proto != nil {
		return vm.lookupInPrototypeWithKey(proto.Proto, e, originalObj, precomputedKey)
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
		argsObj := &Value{Type: "object", Arr: args, Obj: make(map[string]*Value)}
		argsObj.Obj["length"] = &Value{Type: "number", Num: float64(len(args))}
		argsObj.Obj["callee"] = &Value{Type: "function", Func: fn}
		childEnv.Define("arguments", argsObj)

		obj := &Value{Type: "object", Obj: make(map[string]*Value)}
		// Link the object to the constructor's prototype chain
		// In JS: obj.__proto__ = Constructor.prototype
		protoObj := callee.PrototypeObj
		if protoObj == nil && callee.Obj != nil {
			if p, ok := callee.Obj["prototype"]; ok {
				protoObj = p
			}
		}
		if protoObj != nil {
			obj.Proto = protoObj
		}
		childEnv.Define("this", obj)

		oldEnv := vm.env
		vm.env = childEnv
		var result *Value
		for _, stmt := range fn.Body {
			ret, returning, retVal := vm.execStmt(stmt)
			if returning {
				result = ret
				// In JS, if a constructor returns an object, new returns that object.
				// If it returns a primitive (or undefined), new returns this.
				if retVal != nil && (retVal.Type == "object" || retVal.Type == "function" || retVal.Type == "native") {
					obj = retVal
				}
				break
			}
		}
		_ = result
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

func toInt32(v *Value) int32 {
	n := v.Num
	if v.Type == "bool" {
		if v.Bool {
			n = 1
		} else {
			n = 0
		}
	}
	return int32(n)
}

func toUint32(v *Value) uint32 {
	n := v.Num
	if v.Type == "bool" {
		if v.Bool {
			n = 1
		} else {
			n = 0
		}
	}
	return uint32(n)
}

// isBuiltinPrototypeProperty returns true for built-in Object.prototype and
// Array.prototype property names that should be non-enumerable and thus
// excluded from for-in loops when found on the prototype chain (depth > 0).
var builtinNonEnumerableProps = map[string]bool{
	"hasOwnProperty":          true,
	"isPrototypeOf":           true,
	"propertyIsEnumerable":    true,
	"toString":                true,
	"valueOf":                 true,
	"toLocaleString":          true,
	"constructor":             true,
	"__defineGetter__":        true,
	"__defineSetter__":        true,
	"__lookupGetter__":        true,
	"__lookupSetter__":        true,
	"__proto__":               true,
	// Array.prototype methods (also non-enumerable in real JS)
	"push":    true,
	"pop":     true,
	"shift":   true,
	"unshift": true,
	"slice":   true,
	"splice":  true,
	"concat":  true,
	"join":    true,
	"indexOf": true,
	"lastIndexOf": true,
	"forEach": true,
	"map":     true,
	"filter":  true,
	"reduce":  true,
	"reduceRight": true,
	"some":    true,
	"every":   true,
	"find":    true,
	"findIndex": true,
	"includes": true,
	"fill":    true,
	"reverse": true,
	"sort":    true,
	"flat":    true,
	"flatMap": true,
	"keys":    true,
	"values":  true,
	"entries": true,
	// Function.prototype methods
	"call":    true,
	"apply":   true,
	"bind":    true,
	// String.prototype methods
	"charAt":     true,
	"charCodeAt": true,
	"substring":  true,
	"substr":     true,
	"split":      true,
	"replace":    true,
	"toUpperCase": true,
	"toLowerCase": true,
	"trim":       true,
	"trimStart":  true,
	"trimEnd":    true,
	"startsWith": true,
	"endsWith":   true,
	"repeat":     true,
	"padStart":   true,
	"padEnd":     true,
	"match":      true,
	"search":     true,
}

func isBuiltinPrototypeProperty(key string) bool {
	return builtinNonEnumerableProps[key]
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
	case "object", "regexp", "promise":
		return true
	case "function", "native", "class":
		return true
	case "map", "set", "weakmap", "weakset", "iterator":
		return true
	case "arrayMethod", "stringMethod":
		return true
	}
	return false
}

func valuesEqual(a, b *Value) bool {
	// ECMAScript Abstract Equality Comparison (==)
	// Step 1: null == undefined is true
	if (a.Type == "null" || a.Type == "undefined") && (b.Type == "null" || b.Type == "undefined") {
		return true
	}
	// Step 2: same type — use strict equality
	if a.Type == b.Type {
		return valuesStrictEqualSameType(a, b)
	}
	// Step 3: Number == String → ToNumber(string) == number
	if a.Type == "number" && b.Type == "string" {
		nb := toNumberValue(b)
		return numbersEqual(a.Num, nb.Num)
	}
	if a.Type == "string" && b.Type == "number" {
		na := toNumberValue(a)
		return numbersEqual(na.Num, b.Num)
	}
	// Step 4: Boolean == X → ToNumber(boolean) == X
	if a.Type == "bool" {
		return valuesEqual(toNumberValue(a), b)
	}
	if b.Type == "bool" {
		return valuesEqual(a, toNumberValue(b))
	}
	// Step 5: String/Number == Object → ToPrimitive(object) == string/number
	if (a.Type == "string" || a.Type == "number") && (b.Type == "object" || b.Type == "regexp") {
		pb := toPrimitiveValue(b)
		return valuesEqual(a, pb)
	}
	if (b.Type == "string" || b.Type == "number") && (a.Type == "object" || a.Type == "regexp") {
		pa := toPrimitiveValue(a)
		return valuesEqual(pa, b)
	}
	return false
}

// valuesStrictEqualSameType compares two values of the same type for strict equality.
func valuesStrictEqualSameType(a, b *Value) bool {
	switch a.Type {
	case "undefined", "null":
		return true
	case "number":
		return numbersEqual(a.Num, b.Num)
	case "bigint":
		return a.Str == b.Str
	case "string":
		return a.Str == b.Str
	case "bool":
		return a.Bool == b.Bool
	case "native":
		if a.Native != nil && b.Native != nil {
			av := reflect.ValueOf(a.Native)
			bv := reflect.ValueOf(b.Native)
			if av.Pointer() == bv.Pointer() {
				return true
			}
		}
		return a == b
	case "function":
		if a.Func != nil && b.Func != nil {
			if a.Func == b.Func {
				return true
			}
		}
		if a.Native != nil && b.Native != nil {
			av := reflect.ValueOf(a.Native)
			bv := reflect.ValueOf(b.Native)
			if av.Pointer() == bv.Pointer() {
				return true
			}
		}
		return a == b
	case "object", "regexp":
		if a.NodeRef != nil && b.NodeRef != nil {
			return a.NodeRef == b.NodeRef
		}
		return a == b
	case "symbol":
		return a == b
	}
	return false
}

// numbersEqual compares two float64 values following JS semantics (NaN != NaN).
func numbersEqual(a, b float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return a == b
}

// toNumberValue converts a Value to a number Value following JS ToNumber rules.
func toNumberValue(v *Value) *Value {
	switch v.Type {
	case "number":
		return v
	case "bool":
		if v.Bool {
			return &Value{Type: "number", Num: 1}
		}
		return &Value{Type: "number", Num: 0}
	case "string":
		n, ok := jsParseNumber(v.Str)
		if ok {
			return &Value{Type: "number", Num: n}
		}
		return &Value{Type: "number", Num: math.NaN()}
	case "null":
		return &Value{Type: "number", Num: 0}
	case "undefined":
		return &Value{Type: "number", Num: math.NaN()}
	default:
		return &Value{Type: "number", Num: math.NaN()}
	}
}

// toPrimitiveValue converts an object Value to a primitive Value (string or number).
func toPrimitiveValue(v *Value) *Value {
	if v.Type == "object" && v.Obj != nil {
		if v, ok := v.Obj["valueOf"]; ok && (v.Type == "function" || v.Type == "native") {
			return &Value{Type: "string", Str: valueToString(v)}
		}
		if ts, ok := v.Obj["toString"]; ok && (ts.Type == "function" || ts.Type == "native") {
			return &Value{Type: "string", Str: valueToString(v)}
		}
	}
	return &Value{Type: "string", Str: valueToString(v)}
}

// jsParseNumber parses a string to a float64 following JS ToNumber semantics.
// Returns the number and true if successful, or 0 and false if the string is not a valid number.
func jsParseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, true
	}
	if s == "Infinity" || s == "+Infinity" {
		return math.Inf(1), true
	}
	if s == "-Infinity" {
		return math.Inf(-1), true
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN(), false
	}
	return n, true
}

func valuesStrictEqual(a, b *Value) bool {
	if a.Type != b.Type {
		return false
	}
	return valuesStrictEqualSameType(a, b)
}

// jsTypeOf returns the JavaScript typeof string for a value.
func jsTypeOf(v *Value) string {
	switch v.Type {
	case "undefined":
		return "undefined"
	case "null":
		return "object"
	case "bool":
		return "boolean"
	case "number", "bigint":
		return v.Type
	case "string":
		return "string"
	case "function", "native":
		return "function"
	case "object":
		if v.Arr != nil {
			return "object"
		}
		return "object"
	case "regexp":
		return "object"
	case "symbol":
		return "symbol"
	default:
		return "object"
	}
}

// ValueToString converts a Value to its string representation.
// This is the exported version of valueToString for use by other packages.
func ValueToString(v *Value) string {
	return valueToString(v)
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
			// Check for Error objects with name/message
			if nameVal, ok := v.Obj["name"]; ok && nameVal.Type == "string" {
				if msgVal, ok := v.Obj["message"]; ok && msgVal.Type == "string" {
					return nameVal.Str + ": " + msgVal.Str
				}
				return nameVal.Str
			}
			return "[object Object]"
		}
		return "[object]"
	case "function":
		return "[function]"
	case "native":
		return "[native function]"
	case "symbol":
		if v.Str != "" {
			return "Symbol(" + v.Str + ")"
		}
		return "Symbol(" + strconv.FormatFloat(v.Num, 'f', 0, 64) + ")"
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
// callGetter invokes a getter function with the object as 'this'.
func (vm *VM) callGetter(fn *Value, obj *Value) *Value {
	vm.accessorDepth++
	vm.accessorCalls++
	if vm.accessorDepth > 100 || (vm.accessorMax > 0 && vm.accessorCalls > vm.accessorMax) {
		vm.accessorDepth--
		return &Value{Type: "undefined"}
	}
	result := &Value{Type: "undefined"}
	if fn.Type == "function" && fn.Func != nil {
		result = vm.callFunctionWithThis(fn.Func, obj, []*Value{})
	} else if fn.Type == "native" && fn.Native != nil {
		result = fn.Native([]*Value{obj})
	}
	vm.accessorDepth--
	return result
}

// callSetter invokes a setter function with the object as 'this' and the value.
func (vm *VM) callSetter(fn *Value, obj *Value, val *Value) {
	vm.accessorDepth++
	vm.accessorCalls++
	if vm.accessorDepth > 100 || (vm.accessorMax > 0 && vm.accessorCalls > vm.accessorMax) {
		vm.accessorDepth--
		return
	}
	if fn.Type == "function" && fn.Func != nil {
		vm.callFunctionWithThis(fn.Func, obj, []*Value{val})
	} else if fn.Type == "native" && fn.Native != nil {
		fn.Native([]*Value{obj, val})
	}
	vm.accessorDepth--
}

func (vm *VM) callFunction(fn *Value, args []*Value) *Value {
	if fn.Type == "native" && fn.Native != nil {
		return fn.Native(args)
	}

	if fn.Type == "function" && fn.Func != nil {
		return vm.callFunctionWithThis(fn.Func, fn.ThisBinding, args)
	}

	return &Value{Type: "undefined"}
}

// bindThis returns a copy of v with ThisBinding set to obj.
// It preserves all fields so that property access on the returned value
// still works (e.g., Fn.version where Fn was read from window.Fn).
func (vm *VM) bindThis(v *Value, obj *Value) *Value {
	return &Value{
		Type:               v.Type,
		Num:                v.Num,
		Str:                v.Str,
		Bool:               v.Bool,
		Obj:                v.Obj,
		Arr:                v.Arr,
		Func:               v.Func,
		Class:              v.Class,
		Promise:            v.Promise,
		Proxy:              v.Proxy,
		MapData:            v.MapData,
		SetData:            v.SetData,
		Native:             v.Native,
		ThisBinding:        obj,
		Descriptors:        v.Descriptors,
		IsAsync:            v.IsAsync,
		BuiltInConstructor: v.BuiltInConstructor,
		NodeRef:            v.NodeRef,
		Proto:              v.Proto,
		Frozen:             v.Frozen,
		PrototypeObj:       v.PrototypeObj,
	}
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

	// Create the 'arguments' object (array-like with numeric indices, length, and callee)
	argsObj := &Value{Type: "object", Arr: args, Obj: make(map[string]*Value)}
	argsObj.Obj["length"] = &Value{Type: "number", Num: float64(len(args))}
	argsObj.Obj["callee"] = &Value{Type: "function", Func: f}
	childEnv.Define("arguments", argsObj)

	oldEnv := vm.env
	vm.env = childEnv

	var firstStmt string
	if len(f.Body) > 0 {
		firstStmt = fmt.Sprintf(" %T", f.Body[0])
	}
	vm.callStack = append(vm.callStack, fmt.Sprintf("callFnWithThis(%v) body=%d%s", f.Params, len(f.Body), firstStmt))
	vm.enterCall()
	defer vm.exitCall()

	// Hoist function declarations in the function body
	vm.hoistFunctionDecls(f.Body)

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
		offset := nativeThisOffset(args)
		if len(args) > offset {
			p.Value = args[offset]
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
		offset := nativeThisOffset(args)
		if len(args) > offset {
			p.Rejection = args[offset]
		}
		return &Value{Type: "promise", Promise: p}
	}}

	// Promise.all static method
	promiseAll := &Value{Type: "native", Native: func(args []*Value) *Value {
		offset := nativeThisOffset(args)
		if len(args) <= offset {
			return &Value{Type: "promise", Promise: &Promise{
				State:     "fulfilled",
				Value:     &Value{Type: "object", Arr: []*Value{}},
				OnFulfill: make([]*Function, 0),
				OnReject:  make([]*Function, 0),
				Env:       vm.env,
			}}
		}
		promises := args[offset]
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

	// Promise.race static method
	promiseRace := &Value{Type: "native", Native: func(args []*Value) *Value {
		offset := nativeThisOffset(args)
		if len(args) <= offset {
			return &Value{Type: "promise", Promise: &Promise{
				State:     "pending",
				OnFulfill: make([]*Function, 0),
				OnReject:  make([]*Function, 0),
				Env:       vm.env,
			}}
		}
		promises := args[offset]
		if promises.Type != "object" || promises.Arr == nil {
			return &Value{Type: "promise", Promise: &Promise{
				State:     "rejected",
				Rejection: &Value{Type: "string", Str: "Promise.race requires an array"},
				OnFulfill: make([]*Function, 0),
				OnReject:  make([]*Function, 0),
				Env:       vm.env,
			}}
		}
		for _, p := range promises.Arr {
			if p.Type == "promise" && p.Promise != nil && (p.Promise.State == "fulfilled" || p.Promise.State == "rejected") {
				return p
			} else if p.Type != "promise" {
				return &Value{Type: "promise", Promise: &Promise{
					State:     "fulfilled",
					Value:     p,
					OnFulfill: make([]*Function, 0),
					OnReject:  make([]*Function, 0),
					Env:       vm.env,
				}}
			}
		}
		return &Value{Type: "promise", Promise: &Promise{
			State:     "pending",
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
			"race":        promiseRace,
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
	vm.env.Define("Map", &Value{Type: "native", BuiltInConstructor: "Map", Native: func(args []*Value) *Value {
		m := &Value{
			Type:    "map",
			BuiltInConstructor: "Map",
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
	vm.env.Define("Set", &Value{Type: "native", BuiltInConstructor: "Set", Native: func(args []*Value) *Value {
		s := &Value{
			Type:    "set",
			BuiltInConstructor: "Set",
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
	vm.env.Define("WeakMap", &Value{Type: "native", BuiltInConstructor: "WeakMap", Native: func(args []*Value) *Value {
		wm := &Value{
			Type:    "weakmap",
			BuiltInConstructor: "WeakMap",
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
	vm.env.Define("WeakSet", &Value{Type: "native", BuiltInConstructor: "WeakSet", Native: func(args []*Value) *Value {
		ws := &Value{
			Type:    "weakset",
			BuiltInConstructor: "WeakSet",
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

	symbolConstructor := &Value{Type: "native", BuiltInConstructor: "Symbol", Native: func(args []*Value) *Value {
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

	// Add Symbol.iterator to Array.prototype so that for...of works on arrays.
	// Returns an iterator object with a next() method that yields array values.
	arrayProto := vm.env.Get("ArrayPrototype")
	if arrayProto != nil && arrayProto.Obj != nil {
		symKey := valueToString(iteratorSymbol)
		arrayProto.Obj[symKey] = &Value{Type: "native", Native: func(args []*Value) *Value {
			var thisObj *Value
			if len(args) > 0 && args[0]._isThisArg {
				thisObj = args[0]
			}
			if thisObj == nil || thisObj.Arr == nil {
				return &Value{Type: "object", Obj: map[string]*Value{
					"next": {Type: "native", Native: func(nArgs []*Value) *Value {
						return &Value{Type: "object", Obj: map[string]*Value{
							"value": {Type: "undefined"},
							"done":  {Type: "bool", Bool: true},
						}}
					}},
				}}
			}
			arr := thisObj.Arr
			idx := 0
			iterObj := &Value{Type: "object", Obj: map[string]*Value{
				"next": {Type: "native", Native: func(nArgs []*Value) *Value {
					if idx < len(arr) {
						val := arr[idx]
						idx++
						return &Value{Type: "object", Obj: map[string]*Value{
							"value": val,
							"done":  {Type: "bool", Bool: false},
						}}
					}
					return &Value{Type: "object", Obj: map[string]*Value{
						"value": {Type: "undefined"},
						"done":  {Type: "bool", Bool: true},
					}}
				}},
			}}
			return iterObj
		}}
	}
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
	vm.env.Define("Proxy", &Value{Type: "native", BuiltInConstructor: "Proxy", Native: func(args []*Value) *Value {
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
		offset := nativeThisOffset(args)
		if len(args) < offset+2 {
			return &Value{Type: "number", Num: 0}
		}
		callback := args[offset]
		delay := int64(0)
		if args[offset+1].Type == "number" {
			delay = int64(args[offset+1].Num)
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
		offset := nativeThisOffset(args)
		if len(args) < offset+2 {
			return &Value{Type: "number", Num: 0}
		}
		callback := args[offset]
		delay := int64(0)
		if args[offset+1].Type == "number" {
			delay = int64(args[offset+1].Num)
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
		offset := nativeThisOffset(args)
		if len(args) > offset && args[offset].Type == "number" {
			id := int(args[offset].Num)
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
		offset := nativeThisOffset(args)
		if len(args) > offset && args[offset].Type == "number" {
			id := int(args[offset].Num)
			for i, t := range vm.pendingTimers {
				if t.id == id && t.interval {
					vm.pendingTimers = append(vm.pendingTimers[:i], vm.pendingTimers[i+1:]...)
					break
				}
			}
		}
		return &Value{Type: "undefined"}
	}}

	// requestAnimationFrame(callback) - schedules a callback for next animation frame
	// Used by Vue.js for batching DOM updates
	requestAnimationFrame := &Value{Type: "native", Native: func(args []*Value) *Value {
		offset := nativeThisOffset(args)
		if len(args) > offset {
			callback := args[offset]
			if callback.Type == "function" || callback.Type == "native" {
				timerID++
				task := timerTask{
					id:        timerID,
					callback:  callback,
					delay:     16 * 1e6, // 16ms in ns (approx 60fps)
					interval:  false,
				}
				task.executeAt = time.Now().UnixNano() + task.delay
				vm.pendingTimers = append(vm.pendingTimers, task)
				return &Value{Type: "number", Num: float64(timerID)}
			}
		}
		return &Value{Type: "number", Num: 0}
	}}

	// cancelAnimationFrame(id)
	cancelAnimationFrame := &Value{Type: "native", Native: func(args []*Value) *Value {
		offset := nativeThisOffset(args)
		if len(args) > offset && args[offset].Type == "number" {
			id := int(args[offset].Num)
			for i, t := range vm.pendingTimers {
				if t.id == id {
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
	vm.env.Define("requestAnimationFrame", requestAnimationFrame)
	vm.env.Define("cancelAnimationFrame", cancelAnimationFrame)

	// Also add to window object if it exists
	if window := vm.env.Get("window"); window.Type == "object" && window.Obj != nil {
		window.Obj["setTimeout"] = setTimeout
		window.Obj["setInterval"] = setInterval
		window.Obj["clearTimeout"] = clearTimeout
		window.Obj["clearInterval"] = clearInterval
		window.Obj["requestAnimationFrame"] = requestAnimationFrame
		window.Obj["cancelAnimationFrame"] = cancelAnimationFrame
	}
}

// RunTimers executes all pending timers that are due.
// Returns the number of timers still pending.
func (vm *VM) RunTimers() int {
	now := time.Now().UnixNano()
	var remaining []timerTask

	// Iterate over a copy so callbacks can safely add new timers
	tasks := make([]timerTask, len(vm.pendingTimers))
	copy(tasks, vm.pendingTimers)
	// Track original timer IDs to detect new timers added during execution
	originalIDs := make(map[int]bool)
	for _, t := range tasks {
		originalIDs[t.id] = true
	}

	for _, task := range tasks {
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
	// Append any new timers that were added during callback execution
	for _, t := range vm.pendingTimers {
		if !originalIDs[t.id] {
			remaining = append(remaining, t)
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

// cssToCamel converts a CSS property name to camelCase.
// e.g. "background-color" → "backgroundColor", "font-size" → "fontSize"
func cssToCamel(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// parseStyleAttr parses a CSS style attribute string into a map.
func parseStyleAttr(styleAttr string) map[string]string {
	result := make(map[string]string)
	if styleAttr == "" {
		return result
	}
	for _, part := range strings.Split(styleAttr, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 {
			camelKey := cssToCamel(strings.TrimSpace(kv[0]))
			result[camelKey] = strings.TrimSpace(kv[1])
		}
	}
	return result
}

// buildStyleAttr builds a CSS style attribute string from a map.
func buildStyleAttr(styleMap map[string]string) string {
	keys := make([]string, 0, len(styleMap))
	for k := range styleMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		cssKey := camelToCSS(k)
		parts = append(parts, cssKey+": "+styleMap[k])
	}
	return strings.Join(parts, "; ")
}

// camelToCSS converts a camelCase property name to CSS kebab-case.
// e.g. "backgroundColor" → "background-color"
func camelToCSS(s string) string {
	var buf strings.Builder
	for i, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			if i > 0 {
				buf.WriteByte('-')
			}
			buf.WriteByte(byte(ch + 32))
		} else {
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}
