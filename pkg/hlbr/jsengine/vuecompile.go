package jsengine

import (
	"regexp"
	"strings"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
	"github.com/topxeq/xxlang/pkg/hlbr/htmlparser"
)

// CompileVueTemplate compiles a Vue template string and returns a Value
// containing render and staticRenderFns. It works by:
// 1. Parsing the template HTML in Go (fast, no VM step limit)
// 2. Generating render function code (_c/_v/_s calls)
// 3. Using new Function('with(this){return ...}') to create the render
//
// This approach is generic — it handles any Vue component with a template,
// not specific libraries. The with(this) scope ensures that template
// expressions like `authorized`, `ruleForm.username`, `$t('key')` resolve
// against the Vue component instance at render time.
func (vm *VM) CompileVueTemplate(template string) *Value {
	template = strings.TrimSpace(template)
	if template == "" {
		return vm.compileResult(`with(this){return _c("div")}`)
	}

	nodes := htmlparser.ParseFragment(template)
	if len(nodes) == 0 {
		return vm.compileResult(`with(this){return _c("div")}`)
	}

	var code string
	if len(nodes) == 1 {
		code = genCode(nodes[0])
	} else {
		var childCodes []string
		for _, n := range nodes {
			childCodes = append(childCodes, genCode(n))
		}
		code = `_c("div",[` + strings.Join(childCodes, ",") + `])`
	}

	renderCode := "with(this){return " + code + "}"
	return vm.compileResult(renderCode)
}

// compileResult creates a Value with render and staticRenderFns from code string.
func (vm *VM) compileResult(renderCode string) *Value {
	// Create the render function using new Function(code)
	fnVal := vm.newFunctionFromCode(renderCode)
	return &Value{
		Type: "object",
		Obj: map[string]*Value{
			"render":          fnVal,
			"staticRenderFns": {Type: "object", Arr: []*Value{}},
			"_compiled":       {Type: "bool", Bool: true},
			"_renderCode":     {Type: "string", Str: renderCode},
		},
	}
}

// newFunctionFromCode creates a function value using new Function(renderCode).
func (vm *VM) newFunctionFromCode(code string) *Value {
	functionConstructor := vm.env.Get("Function")
	if functionConstructor == nil || (functionConstructor.Type != "native" && functionConstructor.Type != "function") {
		// Fallback: can't create function
		return &Value{Type: "null"}
	}

	// Call new Function(code) which parses the code and creates a function
	result := vm.callFunction(functionConstructor, []*Value{
		{Type: "string", Str: code},
	})
	return result
}

// ---- Code Generation ----

// genCode generates render function code for a DOM node.
func genCode(node *dom.Node) string {
	if node.Type == dom.TextNode {
		return genTextCode(node.Data)
	}
	if node.Type != dom.ElementNode {
		return `_v("")`
	}
	return genElementCode(node)
}

// interpRe matches {{ expr }} interpolation in text.
var interpRe = regexp.MustCompile(`\{\{([\s\S]*?)\}\}`)

// genTextCode generates code for a text node, handling {{ expr }}.
func genTextCode(text string) string {
	if text == "" {
		return `_v("")`
	}
	if !strings.Contains(text, "{{") {
		return `_v("` + escapeJS(text) + `")`
	}

	parts := interpRe.Split(text, -1)
	matches := interpRe.FindAllStringSubmatch(text, -1)

	var items []string
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			items = append(items, `"`+escapeJS(trimmed)+`"`)
		}
		if i < len(matches) {
			expr := strings.TrimSpace(matches[i][1])
			items = append(items, `_s(`+expr+`)`)
		}
	}
	if len(items) == 0 {
		return `_v("")`
	}
	if len(items) == 1 {
		return `_v(` + items[0] + `)`
	}
	return `_v(` + strings.Join(items, "+") + `)`
}

// dirInfo holds parsed Vue directive information for code generation.
type dirInfo struct {
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

// genElementCode generates render function code for an element node.
func genElementCode(node *dom.Node) string {
	tag := node.Data
	dir := parseDirectives(node.Attr)

	if dir.ifExp != "" {
		return genIfCode(dir.ifExp, node)
	}
	if dir.elseIfExp != "" || dir.else_ {
		return `_c("div")`
	}

	code := genElementInnerCode(tag, dir, node)

	if dir.forExp != "" {
		code = wrapForCode(dir.forExp, code)
	}

	return code
}

// genIfCode generates conditional render code for v-if/v-else-if/v-else chains.
func genIfCode(ifExp string, ifNode *dom.Node) string {
	type branch struct {
		exp  string
		node *dom.Node
	}
	branches := []branch{{exp: ifExp, node: ifNode}}

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
			cd := parseDirectives(child.Attr)
			if cd.elseIfExp != "" {
				branches = append(branches, branch{exp: cd.elseIfExp, node: child})
			} else if cd.else_ {
				branches = append(branches, branch{exp: "", node: child})
				break
			} else {
				break
			}
		}
	}

	var parts []string
	for i, br := range branches {
		tag := br.node.Data
		d := parseDirectives(br.node.Attr)
		d.ifExp, d.elseIfExp = "", ""
		d.else_ = false
		inner := genElementInnerCode(tag, d, br.node)
		if d.forExp != "" {
			inner = wrapForCode(d.forExp, inner)
		}
		if i == 0 {
			parts = append(parts, `(`+br.exp+`)?`+inner+`:`)
		} else if br.exp != "" {
			parts = append(parts, `(`+br.exp+`)?`+inner+`:`)
		} else {
			parts = append(parts, inner)
		}
	}
	if branches[len(branches)-1].exp != "" {
		parts = append(parts, `_c("div")`)
	}
	return `(` + strings.Join(parts, "") + `)`
}

// wrapForCode generates v-for render code using _l().
func wrapForCode(forExp, innerCode string) string {
	forExp = strings.TrimSpace(forExp)
	inIdx := strings.LastIndex(forExp, " in ")
	if inIdx < 0 {
		inIdx = strings.LastIndex(forExp, " of ")
	}
	if inIdx < 0 {
		return innerCode
	}
	left := strings.TrimSpace(forExp[:inIdx])
	listExpr := strings.TrimSpace(forExp[inIdx+4:])
	left = strings.Trim(left, "()")
	parts := strings.SplitN(left, ",", 2)
	itemVar := strings.TrimSpace(parts[0])
	indexVar := ""
	if len(parts) > 1 {
		indexVar = strings.TrimSpace(parts[1])
	}
	iter := itemVar
	if indexVar != "" {
		iter = itemVar + "," + indexVar
	}
	return `_l((` + listExpr + `),function(` + iter + `){return ` + innerCode + `})`
}

// genElementInnerCode generates the _c() call for an element.
func genElementInnerCode(tag string, dir *dirInfo, node *dom.Node) string {
	var dataEntries []string

	if dir.staticCls != "" {
		dataEntries = append(dataEntries, `staticClass:`+quoteJS(dir.staticCls))
	}
	if dir.dynaCls != "" {
		dataEntries = append(dataEntries, `class:(`+dir.dynaCls+`)`)
	}
	if dir.staticSty != "" {
		dataEntries = append(dataEntries, `staticStyle:`+quoteJS(dir.staticSty))
	}
	if dir.dynaSty != "" {
		dataEntries = append(dataEntries, `style:(`+dir.dynaSty+`)`)
	}
	if dir.keyVal != "" {
		dataEntries = append(dataEntries, `key:(`+dir.keyVal+`)`)
	}
	if dir.refVal != "" {
		dataEntries = append(dataEntries, `ref:(`+dir.refVal+`)`)
	}
	if dir.slotVal != "" {
		dataEntries = append(dataEntries, `slot:(`+dir.slotVal+`)`)
	}
	if dir.showExp != "" {
		dataEntries = append(dataEntries, `directives:[{name:"show",rawName:"v-show",value:(`+dir.showExp+`)}]`)
	}
	if dir.htmlExp != "" {
		dataEntries = append(dataEntries, `domProps:{innerHTML:(`+dir.htmlExp+`)}`)
	}
	if dir.modelExp != "" {
		dir.plainAttr["data-vmodel"] = dir.modelExp
		dataEntries = append(dataEntries, `domProps:{value:(`+dir.modelExp+`)}`)
		dataEntries = append(dataEntries, `on:{input:function($event){`+dir.modelExp+`=$event.target.value}}`)
	}
	var allAttrItems []string
	if len(dir.bindAttrs) > 0 {
		for k, v := range dir.bindAttrs {
			allAttrItems = append(allAttrItems, quoteJS(k)+`:(`+v+`)`)
		}
	}
	if len(dir.plainAttr) > 0 {
		for k, v := range dir.plainAttr {
			allAttrItems = append(allAttrItems, quoteJS(k)+":"+quoteJS(v))
		}
	}
	if len(allAttrItems) > 0 {
		dataEntries = append(dataEntries, `attrs:{`+strings.Join(allAttrItems, ",")+`}`)
	}
	if len(dir.onEvents) > 0 {
		var items []string
		for k, v := range dir.onEvents {
			items = append(items, quoteJS(k)+`:function($event){`+v+`}`)
		}
		dataEntries = append(dataEntries, `on:{`+strings.Join(items, ",")+`}`)
	}

	var childCodes []string
	if dir.textExp != "" {
		childCodes = []string{`_v(_s(` + dir.textExp + `))`}
	} else {
		for _, child := range node.Children {
			childCodes = append(childCodes, genCode(child))
		}
	}

	dataStr := ""
	if len(dataEntries) > 0 {
		dataStr = `{` + strings.Join(dataEntries, ",") + `}`
	}
	childStr := ""
	if len(childCodes) > 0 {
		childStr = `[` + strings.Join(childCodes, ",") + `]`
	}

	parts := []string{quoteJS(tag)}
	if dataStr != "" {
		parts = append(parts, dataStr)
		if childStr != "" {
			parts = append(parts, childStr)
		}
	} else if childStr != "" {
		parts = append(parts, childStr)
	}
	return `_c(` + strings.Join(parts, ",") + `)`
}

// parseDirectives extracts Vue directives from element attributes.
func parseDirectives(attrs []dom.Attribute) *dirInfo {
	dir := &dirInfo{
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
			dir.keyVal = quoteJS(value)
		case name == ":ref":
			dir.refVal = value
		case name == "ref":
			dir.refVal = quoteJS(value)
		case name == ":slot":
			dir.slotVal = value
		case name == "slot":
			dir.slotVal = quoteJS(value)
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
			dir.onEvents[eventName] = normalizeHandler(value)
		case strings.HasPrefix(name, "v-on:"):
			eventName := name[5:]
			dir.onEvents[eventName] = normalizeHandler(value)
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

// simpleNameRe matches simple JavaScript identifiers.
var simpleNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// normalizeHandler converts event handler expressions to function bodies.
func normalizeHandler(handler string) string {
	handler = strings.TrimSpace(handler)
	if strings.HasPrefix(handler, "function(") || strings.HasPrefix(handler, "(") {
		return handler
	}
	if strings.Contains(handler, "(") || strings.Contains(handler, "++") || strings.Contains(handler, "--") {
		return handler
	}
	if simpleNameRe.MatchString(handler) {
		return handler + "($event)"
	}
	return handler
}

// escapeJS escapes a string for use inside JavaScript string literals.
func escapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// quoteJS wraps a string in double quotes with escaping.
func quoteJS(s string) string {
	return `"` + escapeJS(s) + `"`
}
