package dom

import (
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

	return selectByTag(root, selector)
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
