package dom

import (
	"regexp"
	"strconv"
	"strings"
)

func QuerySelector(root *Node, selector string) *Node {
	results := QuerySelectorAll(root, selector)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

func QuerySelectorAll(root *Node, selector string) []*Node {
	selector = strings.TrimSpace(selector)

	// Handle comma-separated selector groups: "input,textarea,select"
	if strings.Contains(selector, ",") {
		var results []*Node
		seen := make(map[*Node]bool)
		for _, part := range strings.Split(selector, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			for _, n := range QuerySelectorAll(root, part) {
				if !seen[n] {
					seen[n] = true
					results = append(results, n)
				}
			}
		}
		return results
	}

	// Tokenize the selector into simple selector segments connected by
	// combinators (space = descendant, > = child, + = adjacent sibling).
	segments := tokenizeSelector(selector)
	if len(segments) == 1 {
		return selectSimple(root, segments[0].selector)
	}

	return selectByCombinators(root, segments)
}

// selectorSegment represents one segment of a complex CSS selector.
type selectorSegment struct {
	combinator string // "" (first), " " (descendant), ">" (child), "+" (adjacent sibling)
	selector   string // simple selector like "div", ".foo", "#id", "input[type='text']"
}

// tokenizeSelector splits a complex CSS selector into segments.
// E.g. "div > .foo input[type='text']" → [{"" "div"}, {">" ".foo"}, {" " "input[type='text']"}]
func tokenizeSelector(selector string) []selectorSegment {
	var segments []selectorSegment
	i := 0
	n := len(selector)

	for i < n {
		// Skip leading whitespace
		for i < n && selector[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}

		// Check for combinator at current position
		combinator := " " // default is descendant
		if len(segments) > 0 {
			// Look back: if the previous character before whitespace was > or +, it's that combinator
			// We already skipped whitespace, so check if the previous segment ended with a combinator marker
		}

		// Now parse the simple selector until we hit a combinator
		start := i
		inBrackets := 0
		for i < n {
			ch := selector[i]
			if ch == '[' {
				inBrackets++
			} else if ch == ']' {
				inBrackets--
			} else if inBrackets == 0 {
				if ch == ' ' || ch == '>' || ch == '+' {
					// Potential combinator
					sel := strings.TrimSpace(selector[start:i])
					if sel != "" {
						segments = append(segments, selectorSegment{combinator: combinator, selector: sel})
					}
					// Determine next combinator
					if ch == '>' || ch == '+' {
						segments = append(segments, selectorSegment{combinator: string(ch), selector: ""})
						i++
						// Skip whitespace after combinator
						for i < n && selector[i] == ' ' {
							i++
						}
						combinator = " " // reset for next segment
						start = i
						continue
					}
					// It's a space — descendant combinator
					combinator = " "
					i++
					// Skip remaining whitespace
					for i < n && selector[i] == ' ' {
						i++
					}
					start = i
					continue
				}
			}
			i++
		}
		// Final segment
		sel := strings.TrimSpace(selector[start:])
		if sel != "" {
			segments = append(segments, selectorSegment{combinator: combinator, selector: sel})
		}
		break
	}

	// Clean up: merge combinator-only segments into the next segment
	var cleaned []selectorSegment
	for i := 0; i < len(segments); i++ {
		s := segments[i]
		if s.selector == "" && s.combinator != "" {
			// This is a combinator-only segment; apply it to the next segment
			if i+1 < len(segments) {
				segments[i+1].combinator = s.combinator
			}
			continue
		}
		cleaned = append(cleaned, s)
	}

	// First segment has no combinator
	if len(cleaned) > 0 {
		cleaned[0].combinator = ""
	}

	return cleaned
}

// selectSimple handles a single simple selector (no combinators).
func selectSimple(root *Node, selector string) []*Node {
	selector = strings.TrimSpace(selector)

	// Check for pseudo-classes like :first-child, :last-child, :nth-child(n)
	if colonIdx := strings.Index(selector, ":"); colonIdx >= 0 {
		return selectByPseudoClass(root, selector, colonIdx)
	}

	if strings.HasPrefix(selector, "#") && !strings.Contains(selector[1:], "#") && !strings.Contains(selector, ".") && !strings.Contains(selector, "[") {
		return selectByID(root, selector[1:])
	}
	if strings.HasPrefix(selector, ".") && !strings.Contains(selector[1:], ".") && !strings.Contains(selector, "#") && !strings.Contains(selector, "[") {
		return selectByClass(root, selector[1:])
	}

	// Check for attribute selectors like "script[src]" or "link[rel='stylesheet']"
	if strings.Contains(selector, "[") && strings.HasSuffix(selector, "]") {
		return selectByAttributeSelector(root, selector)
	}

	// Check for combined selectors like "p#intro" or "div.className"
	if strings.Contains(selector, "#") || strings.Contains(selector, ".") {
		return selectByCombinedSelector(root, selector)
	}

	return selectByTag(root, selector)
}

// selectByPseudoClass handles selectors with pseudo-classes.
// e.g. "li:first-child", "tr:nth-child(2)", "div:last-child"
func selectByPseudoClass(root *Node, selector string, colonIdx int) []*Node {
	baseSelector := selector[:colonIdx]
	pseudoPart := selector[colonIdx+1:]

	// Get base candidates
	var candidates []*Node
	if baseSelector == "" {
		walk(root, func(n *Node) {
			if n.Type == ElementNode {
				candidates = append(candidates, n)
			}
		})
	} else {
		candidates = selectSimple(root, baseSelector)
	}

	// Filter by pseudo-class
	var results []*Node
	for _, n := range candidates {
		if matchPseudoClass(n, pseudoPart) {
			results = append(results, n)
		}
	}
	return results
}

// matchPseudoClass checks if a node matches a pseudo-class expression.
func matchPseudoClass(n *Node, pseudo string) bool {
	// Handle :first-child
	if pseudo == "first-child" {
		if n.Parent == nil {
			return true
		}
		for _, sib := range n.Parent.Children {
			if sib.Type == ElementNode {
				return sib == n
			}
		}
		return false
	}

	// Handle :last-child
	if pseudo == "last-child" {
		if n.Parent == nil {
			return true
		}
		for i := len(n.Parent.Children) - 1; i >= 0; i-- {
			if n.Parent.Children[i].Type == ElementNode {
				return n.Parent.Children[i] == n
			}
		}
		return false
	}

	// Handle :nth-child(n)
	if strings.HasPrefix(pseudo, "nth-child(") && strings.HasSuffix(pseudo, ")") {
		arg := pseudo[10 : len(pseudo)-1]
		return matchNthChild(n, arg)
	}

	// Handle :not(selector)
	if strings.HasPrefix(pseudo, "not(") && strings.HasSuffix(pseudo, ")") {
		arg := pseudo[4 : len(pseudo)-1]
		matches := selectSimple(n, arg)
		for _, m := range matches {
			if m == n {
				return false
			}
		}
		return true
	}

	return false
}

// matchNthChild checks if a node is the nth child (1-indexed).
func matchNthChild(n *Node, arg string) bool {
	if n.Parent == nil {
		return false
	}

	// Find the 1-based index of this element among its element siblings
	index := 0
	for _, sib := range n.Parent.Children {
		if sib.Type == ElementNode {
			index++
			if sib == n {
				break
			}
		}
	}

	// Handle specific number: "3"
	if num, err := strconv.Atoi(arg); err == nil {
		return index == num
	}

	// Handle "odd" and "even"
	if arg == "odd" {
		return index%2 == 1
	}
	if arg == "even" {
		return index%2 == 0
	}

	// Handle "an+b" pattern (simplified: only handles "2n+1", "3n", "n+2", etc.)
	anbPattern := regexp.MustCompile(`^(\d*)n\s*\+\s*(\d+)$|^(\d*)n$|^n\s*\+\s*(\d+)$`)
	if m := anbPattern.FindStringSubmatch(arg); m != nil {
		aStr := ""
		bStr := ""
		if m[2] != "" { // "an+b"
			aStr = m[1]
			bStr = m[2]
		} else if m[3] != "" { // "an"
			aStr = m[3]
			bStr = "0"
		} else if m[4] != "" { // "n+b"
			aStr = "1"
			bStr = m[4]
		}
		a := 1
		if aStr != "" {
			if v, err := strconv.Atoi(aStr); err == nil {
				a = v
			}
		}
		b, _ := strconv.Atoi(bStr)
		if a == 0 {
			return index == b
		}
		// Check: index = a*n + b for some non-negative integer n
		diff := index - b
		if diff < 0 {
			return false
		}
		return diff%a == 0
	}

	return false
}

// selectByCombinators handles complex selectors with combinators.
func selectByCombinators(root *Node, segments []selectorSegment) []*Node {
	if len(segments) == 0 {
		return nil
	}

	// Start with the first segment
	results := selectSimple(root, segments[0].selector)

	// Apply each subsequent segment with its combinator
	for i := 1; i < len(segments); i++ {
		seg := segments[i]
		var next []*Node

		for _, candidate := range results {
			switch seg.combinator {
			case " ":
				// Descendant: find all descendants of candidate matching the selector
				next = append(next, selectSimple(candidate, seg.selector)...)
			case ">":
				// Child: find direct children of candidate matching the selector
				for _, child := range candidate.Children {
					if matchComplexSelector(child, seg.selector) {
						next = append(next, child)
					}
				}
			case "+":
				// Adjacent sibling: find the next sibling matching the selector
				if candidate.Parent != nil {
					found := false
					for _, sib := range candidate.Parent.Children {
						if found && sib.Type == ElementNode {
							if matchComplexSelector(sib, seg.selector) {
								next = append(next, sib)
							}
							break
						}
						if sib == candidate {
							found = true
						}
					}
				}
			}
		}

		results = next
	}

	// Deduplicate
	seen := make(map[*Node]bool)
	var deduped []*Node
	for _, n := range results {
		if !seen[n] {
			seen[n] = true
			deduped = append(deduped, n)
		}
	}

	return deduped
}

// MatchesSelector checks if a node matches a CSS selector string.
// It handles comma-separated selectors (returns true if any match).
func MatchesSelector(n *Node, selector string) (bool, error) {
	if n == nil {
		return false, nil
	}
	parts := strings.Split(selector, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if matchComplexSelector(n, part) {
			return true, nil
		}
	}
	return false, nil
}

// matchComplexSelector checks if a node matches a simple selector.
func matchComplexSelector(n *Node, selector string) bool {
	if n == nil || n.Type != ElementNode {
		return false
	}
	selector = strings.TrimSpace(selector)
	return matchSimpleSelector(n, selector)
}

// matchSimpleSelector checks if a single node matches a simple CSS selector.
func matchSimpleSelector(n *Node, selector string) bool {
	selector = strings.TrimSpace(selector)
	if selector == "" || selector == "*" {
		return true
	}

	// Check for pseudo-classes like :first-child
	if colonIdx := strings.Index(selector, ":"); colonIdx >= 0 {
		baseSelector := selector[:colonIdx]
		pseudoPart := selector[colonIdx+1:]
		if baseSelector != "" && !matchSimpleSelector(n, baseSelector) {
			return false
		}
		return matchPseudoClass(n, pseudoPart)
	}

	// Check for attribute selectors like "script[src]" or "input[type='text']"
	if strings.Contains(selector, "[") && strings.HasSuffix(selector, "]") {
		return matchAttributeSelector(n, selector)
	}

	// Check for combined selectors like "p#intro" or "div.className" or "input.el-input__inner"
	if strings.Contains(selector, "#") || strings.Contains(selector, ".") {
		return matchCombinedSelector(n, selector)
	}

	// Simple tag selector
	return n.Data == strings.ToLower(selector)
}

// matchAttributeSelector checks if a node matches a selector with attribute brackets.
func matchAttributeSelector(n *Node, selector string) bool {
	bracketIdx := strings.Index(selector, "[")
	if bracketIdx < 0 || !strings.HasSuffix(selector, "]") {
		return false
	}
	tagPart := selector[:bracketIdx]
	attrPart := selector[bracketIdx+1 : len(selector)-1]

	// Check tag part
	if tagPart != "" && !matchSimpleSelector(n, tagPart) {
		return false
	}

	// Parse attribute part
	eqIdx := strings.Index(attrPart, "=")
	if eqIdx < 0 {
		// Just attribute presence: [disabled]
		return n.GetAttribute(attrPart) != ""
	}
	attrName := strings.TrimSpace(attrPart[:eqIdx])
	attrVal := strings.TrimSpace(attrPart[eqIdx+1:])
	// Remove quotes
	if len(attrVal) >= 2 && ((attrVal[0] == '"' && attrVal[len(attrVal)-1] == '"') || (attrVal[0] == '\'' && attrVal[len(attrVal)-1] == '\'')) {
		attrVal = attrVal[1 : len(attrVal)-1]
	}
	return n.GetAttribute(attrName) == attrVal
}

// matchCombinedSelector checks if a node matches a combined selector like "div.cls#id".
func matchCombinedSelector(n *Node, selector string) bool {
	// Parse tag, class, and id parts
	tag := ""
	remaining := selector

	// Extract tag if it starts with a letter (not # or .)
	if len(remaining) > 0 && remaining[0] != '#' && remaining[0] != '.' {
		i := 0
		for i < len(remaining) && remaining[i] != '#' && remaining[i] != '.' && remaining[i] != '[' && remaining[i] != ':' {
			i++
		}
		tag = remaining[:i]
		remaining = remaining[i:]
	}

	// Check tag
	if tag != "" && n.Data != strings.ToLower(tag) {
		return false
	}

	// Parse id and classes from remaining
	for len(remaining) > 0 {
		if remaining[0] == '#' {
			remaining = remaining[1:]
			i := 0
			for i < len(remaining) && remaining[i] != '.' && remaining[i] != '#' && remaining[i] != '[' && remaining[i] != ':' {
				i++
			}
			id := remaining[:i]
			remaining = remaining[i:]
			if n.GetAttribute("id") != id {
				return false
			}
		} else if remaining[0] == '.' {
			remaining = remaining[1:]
			i := 0
			for i < len(remaining) && remaining[i] != '.' && remaining[i] != '#' && remaining[i] != '[' && remaining[i] != ':' {
				i++
			}
			cls := remaining[:i]
			remaining = remaining[i:]
			if !hasClass(n, cls) {
				return false
			}
		} else if remaining[0] == '[' {
			closeIdx := strings.Index(remaining, "]")
			if closeIdx < 0 {
				break
			}
			attrSel := remaining[:closeIdx+1]
			remaining = remaining[closeIdx+1:]
			if !matchAttributeSelector(n, tag+attrSel) {
				return false
			}
		} else {
			break
		}
	}
	return true
}

// hasClass checks if a node has a specific CSS class.
func hasClass(n *Node, cls string) bool {
	classAttr := n.GetAttribute("class")
	for _, c := range strings.Fields(classAttr) {
		if c == cls {
			return true
		}
	}
	return false
}

// selectByAttributeSelector handles selectors like "tag[attr]", "tag[attr='value']", "tag[attr=\"value\"]", "tag[attr=value]"
func selectByAttributeSelector(root *Node, selector string) []*Node {
	var results []*Node

	// Parse: extract the part before [ and the content inside []
	bracketIdx := strings.Index(selector, "[")
	if bracketIdx < 0 || !strings.HasSuffix(selector, "]") {
		return results
	}

	tag := strings.ToLower(selector[:bracketIdx])
	attrExpr := selector[bracketIdx+1 : len(selector)-1]

	// Split attrExpr into attr and optional value
	var attr, value string
	if eqIdx := strings.Index(attrExpr, "="); eqIdx >= 0 {
		attr = attrExpr[:eqIdx]
		valPart := attrExpr[eqIdx+1:]
		// Strip surrounding quotes if present
		if len(valPart) >= 2 && ((valPart[0] == '\'' && valPart[len(valPart)-1] == '\'') || (valPart[0] == '"' && valPart[len(valPart)-1] == '"')) {
			value = valPart[1 : len(valPart)-1]
		} else {
			value = valPart
		}
	} else {
		attr = attrExpr
	}

	walk(root, func(n *Node) {
		if n.Type != ElementNode {
			return
		}

		// Check tag
		if tag != "" && strings.ToLower(n.Data) != tag {
			return
		}

		// Check attribute
		attrVal := n.GetAttribute(attr)
		if attrVal == "" {
			return
		}

		// If value specified, check exact match
		if value != "" && attrVal != value {
			return
		}

		results = append(results, n)
	})

	return results
}

// selectByCombinedSelector handles selectors like "tag#id" or "tag.className"
func selectByCombinedSelector(root *Node, selector string) []*Node {
	var results []*Node

	// Parse selector: extract tag, id, class
	tag := ""
	id := ""
	class := ""

	// Pattern: tag#id or tag.class or tag#id.class etc
	// First, try to extract tag (letters at the start)
	tagPattern := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9]*)`)
	if match := tagPattern.FindString(selector); match != "" {
		tag = strings.ToLower(match)
		selector = selector[len(match):]
	}

	// Now parse remaining #id and .class parts
	for len(selector) > 0 {
		if strings.HasPrefix(selector, "#") {
			// Extract ID
			idPattern := regexp.MustCompile(`^#([a-zA-Z][a-zA-Z0-9_-]*)`)
			if match := idPattern.FindStringSubmatch(selector); match != nil {
				id = match[1]
				selector = selector[len(match[0]):]
			} else {
				break
			}
		} else if strings.HasPrefix(selector, ".") {
			// Extract class
			classPattern := regexp.MustCompile(`^\.([a-zA-Z][a-zA-Z0-9_-]*)`)
			if match := classPattern.FindStringSubmatch(selector); match != nil {
				class = match[1]
				selector = selector[len(match[0]):]
			} else {
				break
			}
		} else {
			break
		}
	}

	walk(root, func(n *Node) {
		if n.Type != ElementNode {
			return
		}

		// Check tag
		if tag != "" && strings.ToLower(n.Data) != tag {
			return
		}

		// Check ID
		if id != "" && n.GetAttribute("id") != id {
			return
		}

		// Check class
		if class != "" {
			classes := strings.Fields(n.GetAttribute("class"))
			found := false
			for _, c := range classes {
				if c == class {
					found = true
					break
				}
			}
			if !found {
				return
			}
		}

		results = append(results, n)
	})

	return results
}

func selectByID(root *Node, id string) []*Node {
	var results []*Node
	walk(root, func(n *Node) {
		if n.GetAttribute("id") == id {
			results = append(results, n)
		}
	})
	return results
}

func selectByClass(root *Node, class string) []*Node {
	var results []*Node
	walk(root, func(n *Node) {
		classes := strings.Fields(n.GetAttribute("class"))
		for _, c := range classes {
			if c == class {
				results = append(results, n)
				return
			}
		}
	})
	return results
}

func selectByTag(root *Node, tag string) []*Node {
	var results []*Node
	tag = strings.ToLower(tag)
	walk(root, func(n *Node) {
		if n.Type == ElementNode && strings.ToLower(n.Data) == tag {
			results = append(results, n)
		}
	})
	return results
}

func walk(n *Node, fn func(*Node)) {
	if n == nil {
		return
	}
	fn(n)
	for _, child := range n.Children {
		walk(child, fn)
	}
}

func GetElementByID(root *Node, id string) *Node {
	return QuerySelector(root, "#"+id)
}

func GetElementsByTagName(root *Node, tag string) []*Node {
	return selectByTag(root, tag)
}

func GetElementsByClassName(root *Node, class string) []*Node {
	return selectByClass(root, class)
}
