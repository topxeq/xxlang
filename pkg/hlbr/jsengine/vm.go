package jsengine

import (
	"math"
	"strconv"
	"strings"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
)

type Value struct {
	Type   string
	Num    float64
	Str    string
	Bool   bool
	Obj    map[string]*Value
	Arr    []*Value
	Func   *Function
	Native func(args []*Value) *Value
}

type Function struct {
	Params []string
	Body   []Statement
	Env    *Environment
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
	env           *Environment
	doc           *dom.Document
	root          *dom.Node
	output        []string
	LocalStorage  map[string]string
	SessionStorage map[string]string
}

func NewVM(doc *dom.Document) *VM {
	vm := &VM{doc: doc, LocalStorage: make(map[string]string), SessionStorage: make(map[string]string)}
	if doc != nil {
		vm.root = doc.Root
	}
	vm.env = NewEnvironment(nil)
	vm.setupBuiltins()
	return vm
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
		"body": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(bodyNode)
		}},
		"head": {Type: "native", Native: func(args []*Value) *Value {
			return vm.wrapNode(headNode)
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

	vm.env.Define("Array", &Value{Type: "native", Native: func(args []*Value) *Value {
		return &Value{Type: "object", Arr: args}
	}})

	vm.env.Define("Object", &Value{Type: "native", Native: func(args []*Value) *Value {
		if len(args) == 0 {
			return &Value{Type: "object", Obj: make(map[string]*Value)}
		}
		return args[0]
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
	return &Value{Type: "object", Obj: map[string]*Value{
		"tagName":     {Type: "string", Str: n.TagName()},
		"id":          {Type: "string", Str: n.ID()},
		"className":   {Type: "string", Str: n.ClassName()},
		"nodeName":    {Type: "string", Str: n.Data},
		"nodeValue":   {Type: "string", Str: n.GetAttribute("value")},
		"textContent": {Type: "string", Str: n.TextContent()},
		"innerHTML":   {Type: "string", Str: n.InnerHTML()},
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
}

// wrapNodeShallow wraps a node without eagerly wrapping children or parent,
// to avoid infinite recursion on circular DOM references (parent -> child -> parent).
// Access to children, firstChild, lastChild, parentNode is via lazy native functions.
func (vm *VM) wrapNodeShallow(n *dom.Node) *Value {
	if n == nil {
		return &Value{Type: "null"}
	}
	return &Value{Type: "object", Obj: map[string]*Value{
		"tagName":     {Type: "string", Str: n.TagName()},
		"id":          {Type: "string", Str: n.ID()},
		"className":   {Type: "string", Str: n.ClassName()},
		"nodeName":    {Type: "string", Str: n.Data},
		"nodeValue":   {Type: "string", Str: n.GetAttribute("value")},
		"textContent": {Type: "string", Str: n.TextContent()},
		"innerHTML":   {Type: "string", Str: n.InnerHTML()},
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
	parser := NewParser(code)
	prog := parser.Parse()

	var result *Value = &Value{Type: "undefined"}
	var returnVal *Value
	var returning bool

	for _, stmt := range prog.Statements {
		if returning {
			break
		}
		result, returning, returnVal = vm.execStmt(stmt)
	}

	if returning && returnVal != nil {
		return returnVal, nil
	}
	return result, nil
}

func (vm *VM) execStmt(stmt Statement) (*Value, bool, *Value) {
	switch s := stmt.(type) {
	case *ExpressionStmt:
		v := vm.evalExpr(s.Expr)
		return v, false, nil
	case *VarDecl:
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
	case *FunctionDecl:
		fn := &Function{Params: s.Params, Body: s.Body, Env: vm.env}
		vm.env.Define(s.Name, &Value{Type: "function", Func: fn})
		return &Value{Type: "undefined"}, false, nil
	}
	return &Value{Type: "undefined"}, false, nil
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
	case *StringLit:
		return &Value{Type: "string", Str: e.Value}
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
	case *AssignExpr:
		return vm.evalAssign(e)
	case *CallExpr:
		return vm.evalCall(e)
	case *MemberExpr:
		return vm.evalMember(e)
	case *ArrayLit:
		arr := make([]*Value, len(e.Elements))
		for i, el := range e.Elements {
			arr[i] = vm.evalExpr(el)
		}
		return &Value{Type: "object", Arr: arr}
	case *ObjectLit:
		obj := make(map[string]*Value)
		for _, prop := range e.Properties {
			obj[prop.Key] = vm.evalExpr(prop.Value)
		}
		return &Value{Type: "object", Obj: obj}
	case *FunctionExpr:
		fn := &Function{Params: e.Params, Body: e.Body, Env: vm.env}
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
	case *NewExpr:
		return vm.evalNew(e)
	}
	return &Value{Type: "undefined"}
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
		if obj.Type == "object" && obj.Obj != nil {
			prop := ""
			if left.Computed {
				prop = valueToString(vm.evalExpr(left.Property))
			} else if ident, ok := left.Property.(*Ident); ok {
				prop = ident.Name
			}
			obj.Obj[prop] = val
		}
	}

	return val
}

func (vm *VM) evalCall(e *CallExpr) *Value {
	callee := vm.evalExpr(e.Callee)

	var args []*Value
	for _, arg := range e.Args {
		args = append(args, vm.evalExpr(arg))
	}

	if callee.Type == "native" && callee.Native != nil {
		return callee.Native(args)
	}

	if callee.Type == "function" && callee.Func != nil {
		fn := callee.Func
		childEnv := NewEnvironment(fn.Env)

		for i, param := range fn.Params {
			if i < len(args) {
				childEnv.Define(param, args[i])
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
	}

	return &Value{Type: "undefined"}
}

func (vm *VM) evalMember(e *MemberExpr) *Value {
	obj := vm.evalExpr(e.Object)

	if obj.Type == "object" {
		if e.Computed {
			prop := vm.evalExpr(e.Property)
			key := valueToString(prop)
			if obj.Obj != nil {
				if v, ok := obj.Obj[key]; ok {
					return v
				}
			}
			if obj.Arr != nil {
				idx := int(prop.Num)
				if idx >= 0 && idx < len(obj.Arr) {
					return obj.Arr[idx]
				}
			}
		} else if ident, ok := e.Property.(*Ident); ok {
			if obj.Obj != nil {
				if v, ok := obj.Obj[ident.Name]; ok {
					return v
				}
			}
		}
	}

	if obj.Type == "string" {
		if ident, ok := e.Property.(*Ident); ok {
			switch ident.Name {
			case "length":
				return &Value{Type: "number", Num: float64(len(obj.Str))}
			}
		}
	}

	return &Value{Type: "undefined"}
}

func (vm *VM) evalNew(e *NewExpr) *Value {
	callee := vm.evalExpr(e.Callee)
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
