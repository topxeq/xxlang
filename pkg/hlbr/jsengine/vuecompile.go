package jsengine

import (
	"strings"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
)

// CompileVueTemplate compiles a Vue template string and returns a Value
// containing render and staticRenderFns. The render function is a native
// Go function that walks the parsed template AST and generates VNodes
// by calling the createElement (h) function passed by Vue.
func (vm *VM) CompileVueTemplate(template string) *Value {
	template = strings.TrimSpace(template)
	if template == "" {
		return vm.makeEmptyCompileResult()
	}

	nodes := htmlparser.ParseFragment(template)
	if len(nodes) == 0 {
		return vm.makeEmptyCompileResult()
	}

	astNodes := nodes

	renderFn := &Value{
		Type: "native",
		Native: func(args []*Value) *Value {
			var h *Value
			var vueInst *Value
			offset := nativeThisOffset(args)
			// When Vue calls render.call(vm, h), the first arg is `this` (Vue instance)
			// and the second arg is `h` (createElement).
			if offset > 0 && len(args) > 0 {
				vueInst = args[0]
			}
			if len(args) > offset {
				h = args[offset]
			}
			if h == nil {
				h = vm.env.Get("_c")
			}
			if h == nil {
				return &Value{Type: "null"}
			}

			// Set the Vue instance as `this` in the VM environment so that
			// expression evaluation (e.g., `authorized`, `ruleForm.username`)
			// resolves against the component's data properties.
			savedThis := vm.env.Get("this")
			if vueInst != nil {
				vm.env.Set("this", vueInst)
			}

			var result *Value
			if len(astNodes) == 1 {
				result = vm.vueRenderNode(astNodes[0], h)
			} else {
				var childVNodes []*Value
				for _, n := range astNodes {
					vn := vm.vueRenderNode(n, h)
					if vn != nil {
						childVNodes = append(childVNodes, vn)
					}
				}
				result = vm.vueCallH(h, "div", nil, childVNodes)
			}

			// Restore `this`
			if savedThis != nil {
				vm.env.Set("this", savedThis)
			}

			return result
		},
	}

	return &Value{
		Type: "object",
		Obj: map[string]*Value{
			"render":          renderFn,
			"staticRenderFns": {Type: "object", Arr: []*Value{}},
			"_compiled":       {Type: "bool", Bool: true},
		},
	}
}

func (vm *VM) makeEmptyCompileResult() *Value {
	emptyRender := &Value{
		Type: "native",
		Native: func(args []*Value) *Value {
			var h *Value
			offset := nativeThisOffset(args)
			if len(args) > offset {
				h = args[offset]
			}
			if h != nil {
				return vm.vueCallH(h, "div", nil, nil)
			}
			return &Value{Type: "null"}
		},
	}
	return &Value{
		Type: "object",
		Obj: map[string]*Value{
			"render":          emptyRender,
			"staticRenderFns": {Type: "object", Arr: []*Value{}},
		},
	}
}

// vueRenderNode generates a VNode for a DOM node by calling h().
func (vm *VM) vueRenderNode(node *dom.Node, h *Value) *Value {
	if node.Type == dom.TextNode {
		return vm.vueRenderTextNode(node.Data, h)
	}
	if node.Type != dom.ElementNode {
		return nil
	}
	return vm.vueRenderElementNode(node, h)
}

// vueRenderTextNode generates a VNode for a text node, handling {{ expr }}.
func (vm *VM) vueRenderTextNode(text string, h *Value) *Value {
	if text == "" {
		return &Value{Type: "string", Str: ""}
	}

	if !strings.Contains(text, "{{") {
		return &Value{Type: "string", Str: text}
	}

	// Text with interpolation - evaluate and concatenate
	parts := vueSplitInterp(text)
	var result string
	for _, p := range parts {
		if p.isExpr {
			val := vm.vueEvalStringExpr(p.content)
			result += valueToString(val)
		} else {
			result += p.content
		}
	}
	return &Value{Type: "string", Str: result}
}

// vueInterpPart represents a part of text (plain or interpolated).
type vueInterpPart struct {
	isExpr  bool
	content string
}

// vueSplitInterp splits text into plain and {{ expr }} parts.
func vueSplitInterp(text string) []vueInterpPart {
	var parts []vueInterpPart
	for {
		idx := strings.Index(text, "{{")
		if idx < 0 {
			if text != "" {
				parts = append(parts, vueInterpPart{isExpr: false, content: text})
			}
			break
		}
		if idx > 0 {
			parts = append(parts, vueInterpPart{isExpr: false, content: text[:idx]})
		}
		endIdx := strings.Index(text[idx:], "}}")
		if endIdx < 0 {
			parts = append(parts, vueInterpPart{isExpr: false, content: text[idx:]})
			break
		}
		expr := strings.TrimSpace(text[idx+2 : idx+endIdx])
		parts = append(parts, vueInterpPart{isExpr: true, content: expr})
		text = text[idx+endIdx+2:]
	}
	return parts
}

// vueEvalStringExpr evaluates a JavaScript expression string in the current VM context.
// For Vue render functions, expressions like `authorized` or `ruleForm.username`
// need to resolve against the Vue instance (this).
func (vm *VM) vueEvalStringExpr(expr string) *Value {
	// Try evaluating with `this.` prefix for simple identifiers and member expressions
	thisPrefixed := "this." + expr
	p := NewParser(thisPrefixed)
	node := p.parseExpression()
	if node != nil {
		val := vm.evalExpr(node)
		if val != nil && val.Type != "undefined" {
			return val
		}
	}
	// Fallback: evaluate without this prefix (for globals like $t, $router, etc.)
	p2 := NewParser(expr)
	node2 := p2.parseExpression()
	if node2 == nil {
		return &Value{Type: "undefined"}
	}
	return vm.evalExpr(node2)
}

// vueRenderDir holds parsed Vue directives for runtime rendering.
type vueRenderDir struct {
	ifExp     string
	elseIfExp string
	else_     bool
	forExp    string
	showExp   string
	htmlExp   string
	textExp   string
	modelExp  string
	bindAttrs map[string]string
	onEvents  map[string]string
	plainAttr map[string]string
	staticCls string
	dynaCls   string
	staticSty string
	dynaSty   string
	refVal    string
	keyVal    string
	slotVal   string
}

// vueRenderElementNode generates a VNode for an element.
func (vm *VM) vueRenderElementNode(node *dom.Node, h *Value) *Value {
	dir := vueParseDirectives(node.Attr)

	if dir.ifExp != "" {
		return vm.vueRenderIf(dir.ifExp, node, h)
	}
	if dir.elseIfExp != "" || dir.else_ {
		return nil
	}

	if dir.forExp != "" {
		return vm.vueRenderFor(dir.forExp, node, h)
	}

	return vm.vueRenderElementInner(node.Data, dir, node, h)
}

// vueRenderIf handles v-if/v-else-if/v-else conditional rendering.
func (vm *VM) vueRenderIf(ifExp string, ifNode *dom.Node, h *Value) *Value {
	type condBranch struct {
		exp  string
		node *dom.Node
	}

	branches := []condBranch{{exp: ifExp, node: ifNode}}

	parent := ifNode.Parent
	if parent != nil {
		foundIf := false
		for _, child := range parent.Children {
			if child == ifNode {
				foundIf = true
				continue
			}
			if !foundIf {
				continue
			}
			childDir := vueParseDirectives(child.Attr)
			if childDir.elseIfExp != "" {
				branches = append(branches, condBranch{exp: childDir.elseIfExp, node: child})
			} else if childDir.else_ {
				branches = append(branches, condBranch{exp: "", node: child})
				break
			} else {
				break
			}
		}
	}

	for _, br := range branches {
		if br.exp == "" {
			dir := vueParseDirectives(br.node.Attr)
			dir.ifExp, dir.elseIfExp = "", ""
			dir.else_ = false
			if dir.forExp != "" {
				return vm.vueRenderFor(dir.forExp, br.node, h)
			}
			return vm.vueRenderElementInner(br.node.Data, dir, br.node, h)
		}
		condVal := vm.vueEvalStringExpr(br.exp)
		if isTruthy(condVal) {
			dir := vueParseDirectives(br.node.Attr)
			dir.ifExp, dir.elseIfExp = "", ""
			dir.else_ = false
			if dir.forExp != "" {
				return vm.vueRenderFor(dir.forExp, br.node, h)
			}
			return vm.vueRenderElementInner(br.node.Data, dir, br.node, h)
		}
	}

	return nil
}

// vueRenderFor handles v-for iteration.
func (vm *VM) vueRenderFor(forExp string, node *dom.Node, h *Value) *Value {
	forExp = strings.TrimSpace(forExp)
	inIdx := strings.LastIndex(forExp, " in ")
	if inIdx < 0 {
		inIdx = strings.LastIndex(forExp, " of ")
	}
	if inIdx < 0 {
		return vm.vueRenderElementInner(node.Data, vueParseDirectives(node.Attr), node, h)
	}

	leftPart := strings.TrimSpace(forExp[:inIdx])
	listExpr := strings.TrimSpace(forExp[inIdx+4:])

	leftPart = strings.Trim(leftPart, "()")
	parts := strings.SplitN(leftPart, ",", 2)
	itemVar := strings.TrimSpace(parts[0])
	indexVar := ""
	if len(parts) > 1 {
		indexVar = strings.TrimSpace(parts[1])
	}

	listVal := vm.vueEvalStringExpr(listExpr)
	if listVal == nil || listVal.Arr == nil {
		return nil
	}

	var vnodes []*Value
	for i, item := range listVal.Arr {
		vm.env.Set(itemVar, item)
		if indexVar != "" {
			vm.env.Set(indexVar, &Value{Type: "number", Num: float64(i)})
		}

		dir := vueParseDirectives(node.Attr)
		dir.ifExp, dir.elseIfExp = "", ""
		dir.else_ = false
		dir.forExp = ""
		vn := vm.vueRenderElementInner(node.Data, dir, node, h)
		if vn != nil {
			vnodes = append(vnodes, vn)
		}
	}

	return &Value{Type: "object", Arr: vnodes}
}

// vueRenderElementInner generates a VNode for an element without conditional/loop handling.
func (vm *VM) vueRenderElementInner(tag string, dir *vueRenderDir, node *dom.Node, h *Value) *Value {
	dataObj := make(map[string]*Value)

	if dir.staticCls != "" {
		dataObj["staticClass"] = &Value{Type: "string", Str: dir.staticCls}
	}
	if dir.dynaCls != "" {
		dataObj["class"] = vm.vueEvalStringExpr(dir.dynaCls)
	}
	if dir.staticSty != "" {
		dataObj["staticStyle"] = &Value{Type: "string", Str: dir.staticSty}
	}
	if dir.dynaSty != "" {
		dataObj["style"] = vm.vueEvalStringExpr(dir.dynaSty)
	}
	if dir.keyVal != "" {
		dataObj["key"] = vm.vueEvalStringExpr(dir.keyVal)
	}
	if dir.refVal != "" {
		dataObj["ref"] = &Value{Type: "string", Str: dir.refVal}
	}
	if dir.slotVal != "" {
		dataObj["slot"] = &Value{Type: "string", Str: dir.slotVal}
	}

	if dir.showExp != "" {
		dataObj["directives"] = &Value{Type: "object", Arr: []*Value{
			{Type: "object", Obj: map[string]*Value{
				"name":    {Type: "string", Str: "show"},
				"rawName": {Type: "string", Str: "v-show"},
				"value":   vm.vueEvalStringExpr(dir.showExp),
			}},
		}}
	}

	if dir.htmlExp != "" {
		dataObj["domProps"] = &Value{Type: "object", Obj: map[string]*Value{
			"innerHTML": vm.vueEvalStringExpr(dir.htmlExp),
		}}
	}

	if dir.modelExp != "" {
		modelVal := vm.vueEvalStringExpr(dir.modelExp)
		if dp, ok := dataObj["domProps"]; ok && dp.Obj != nil {
			dp.Obj["value"] = modelVal
		} else {
			dataObj["domProps"] = &Value{Type: "object", Obj: map[string]*Value{
				"value": modelVal,
			}}
		}
		modelExpr := dir.modelExp
		if on, ok := dataObj["on"]; ok && on.Obj != nil {
			on.Obj["input"] = &Value{Type: "native", Native: func(args []*Value) *Value {
				vm.vueEvalStringExpr(modelExpr + "=$event.target.value")
				return &Value{Type: "undefined"}
			}}
		} else {
			dataObj["on"] = &Value{Type: "object", Obj: map[string]*Value{
				"input": {Type: "native", Native: func(args []*Value) *Value {
					vm.vueEvalStringExpr(modelExpr + "=$event.target.value")
					return &Value{Type: "undefined"}
				}},
			}}
		}
	}

	if len(dir.bindAttrs) > 0 {
		attrsMap := make(map[string]*Value)
		for k, v := range dir.bindAttrs {
			attrsMap[k] = vm.vueEvalStringExpr(v)
		}
		if existing, ok := dataObj["attrs"]; ok && existing.Obj != nil {
			for k, v := range attrsMap {
				existing.Obj[k] = v
			}
		} else {
			dataObj["attrs"] = &Value{Type: "object", Obj: attrsMap}
		}
	}

	if len(dir.onEvents) > 0 {
		eventsMap := make(map[string]*Value)
		for k, v := range dir.onEvents {
			handlerExpr := v
			eventsMap[k] = &Value{Type: "native", Native: func(args []*Value) *Value {
				return vm.vueEvalStringExpr(handlerExpr)
			}}
		}
		if existing, ok := dataObj["on"]; ok && existing.Obj != nil {
			for k, v := range eventsMap {
				existing.Obj[k] = v
			}
		} else {
			dataObj["on"] = &Value{Type: "object", Obj: eventsMap}
		}
	}

	if len(dir.plainAttr) > 0 {
		attrsMap := make(map[string]*Value)
		for k, v := range dir.plainAttr {
			attrsMap[k] = &Value{Type: "string", Str: v}
		}
		if existing, ok := dataObj["attrs"]; ok && existing.Obj != nil {
			for k, v := range attrsMap {
				existing.Obj[k] = v
			}
		} else {
			dataObj["attrs"] = &Value{Type: "object", Obj: attrsMap}
		}
	}

	var childVNodes []*Value
	if dir.textExp != "" {
		val := vm.vueEvalStringExpr(dir.textExp)
		childVNodes = append(childVNodes, &Value{Type: "string", Str: valueToString(val)})
	} else {
		for _, child := range node.Children {
			vn := vm.vueRenderNode(child, h)
			if vn != nil {
				childVNodes = append(childVNodes, vn)
			}
		}
	}

	var dataVal *Value
	if len(dataObj) > 0 {
		dataVal = &Value{Type: "object", Obj: dataObj}
	}

	return vm.vueCallH(h, tag, dataVal, childVNodes)
}

// vueCallH calls the createElement (h) function with the given arguments.
func (vm *VM) vueCallH(h *Value, tag string, data *Value, children []*Value) *Value {
	var args []*Value
	args = append(args, &Value{Type: "string", Str: tag})

	if data != nil {
		args = append(args, data)
	}

	if len(children) > 0 {
		if data == nil {
			args = append(args, &Value{Type: "object", Obj: map[string]*Value{}})
		}
		args = append(args, &Value{Type: "object", Arr: children})
	}

	return vm.callFunction(h, args)
}

// vueParseDirectives extracts Vue directives from element attributes.
func vueParseDirectives(attrs []dom.Attribute) *vueRenderDir {
	dir := &vueRenderDir{
		bindAttrs: make(map[string]string),
		onEvents:  make(map[string]string),
		plainAttr: make(map[string]string),
	}

	for _, attr := range attrs {
		name := attr.Key
		value := attr.Value

		switch {
		case name == "v-if":
			dir.ifExp = value
		case name == "v-else-if":
			dir.elseIfExp = value
		case name == "v-else":
			dir.else_ = true
		case name == "v-for":
			dir.forExp = value
		case name == "v-show":
			dir.showExp = value
		case name == "v-html":
			dir.htmlExp = value
		case name == "v-text":
			dir.textExp = value
		case name == "v-model":
			dir.modelExp = value
		case name == ":key":
			dir.keyVal = value
		case name == "key":
			dir.keyVal = value
		case name == ":ref":
			dir.refVal = value
		case name == "ref":
			dir.refVal = value
		case name == ":slot":
			dir.slotVal = value
		case name == "slot":
			dir.slotVal = value
		case strings.HasPrefix(name, ":"):
			bindAttr := name[1:]
			switch bindAttr {
			case "class":
				dir.dynaCls = value
			case "style":
				dir.dynaSty = value
			default:
				dir.bindAttrs[bindAttr] = value
			}
		case strings.HasPrefix(name, "v-bind:"):
			bindAttr := name[7:]
			switch bindAttr {
			case "class":
				dir.dynaCls = value
			case "style":
				dir.dynaSty = value
			default:
				dir.bindAttrs[bindAttr] = value
			}
		case strings.HasPrefix(name, "@"):
			eventName := name[1:]
			dir.onEvents[eventName] = value
		case strings.HasPrefix(name, "v-on:"):
			eventName := name[5:]
			dir.onEvents[eventName] = value
		case name == "class":
			dir.staticCls = value
		case name == "style":
			dir.staticSty = value
		default:
			if !strings.HasPrefix(name, "v-") {
				dir.plainAttr[name] = value
			}
		}
	}

	return dir
}
