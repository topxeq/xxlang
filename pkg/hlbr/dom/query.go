package dom

import (
	"regexp"
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

	if strings.HasPrefix(selector, "#") {
		return selectByID(root, selector[1:])
	}
	if strings.HasPrefix(selector, ".") {
		return selectByClass(root, selector[1:])
	}

	if strings.Contains(selector, " ") {
		return selectDescendant(root, selector)
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

// selectByAttributeSelector handles selectors like "tag[attr]" or "tag[attr='value']"
func selectByAttributeSelector(root *Node, selector string) []*Node {
	var results []*Node

	// Parse selector: tag[attr] or tag[attr='value']
	attrPattern := regexp.MustCompile(`^([a-zA-Z]*)\[([a-zA-Z-]+)(?:=[\'"]([^\'"]*)[\'"])?\]$`)
	matches := attrPattern.FindStringSubmatch(selector)

	if matches == nil {
		return results
	}

	tag := strings.ToLower(matches[1])
	attr := matches[2]
	value := matches[3] // Empty if no value specified

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

func selectDescendant(root *Node, selector string) []*Node {
	parts := strings.Fields(selector)
	if len(parts) == 0 {
		return nil
	}

	var candidates []*Node
	candidates = append(candidates, root)

	for _, part := range parts {
		var next []*Node
		for _, c := range candidates {
			children := queryNode(c, part)
			next = append(next, children...)
		}
		candidates = next
	}

	return candidates
}

func queryNode(root *Node, selector string) []*Node {
	var results []*Node
	walk(root, func(n *Node) {
		if matchSelector(n, selector) {
			results = append(results, n)
		}
	})
	return results
}

func matchSelector(n *Node, selector string) bool {
	if n.Type != ElementNode {
		return false
	}

	selector = strings.TrimSpace(selector)

	if strings.HasPrefix(selector, "#") {
		return n.GetAttribute("id") == selector[1:]
	}
	if strings.HasPrefix(selector, ".") {
		classes := strings.Fields(n.GetAttribute("class"))
		for _, c := range classes {
			if c == selector[1:] {
				return true
			}
		}
		return false
	}

	return strings.ToLower(n.Data) == strings.ToLower(selector)
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
