package dom

import "testing"

func buildTestTree() *Node {
	root := &Node{Type: ElementNode, Data: "div"}
	root.SetAttribute("id", "root")
	root.SetAttribute("class", "container")

	child1 := &Node{Type: ElementNode, Data: "form", Parent: root}
	child1.SetAttribute("id", "myForm")
	child1.SetAttribute("class", "form-class")
	root.Children = append(root.Children, child1)

	input1 := &Node{Type: ElementNode, Data: "input", Parent: child1}
	input1.SetAttribute("type", "text")
	input1.SetAttribute("placeholder", "username")
	input1.SetAttribute("class", "el-input__inner")
	child1.Children = append(child1.Children, input1)

	input2 := &Node{Type: ElementNode, Data: "input", Parent: child1}
	input2.SetAttribute("type", "password")
	input2.SetAttribute("placeholder", "password")
	input2.SetAttribute("class", "el-input__inner")
	child1.Children = append(child1.Children, input2)

	formItem := &Node{Type: ElementNode, Data: "div", Parent: child1}
	formItem.SetAttribute("class", "el-form-item")
	child1.Children = append(child1.Children, formItem)

	btn := &Node{Type: ElementNode, Data: "button", Parent: child1}
	btn.SetAttribute("type", "submit")
	btn.SetAttribute("class", "btn-primary")
	child1.Children = append(child1.Children, btn)

	child2 := &Node{Type: ElementNode, Data: "span", Parent: root}
	child2.SetAttribute("class", "info-text")
	root.Children = append(root.Children, child2)

	return root
}

func TestSelectByTag(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "input")
	if len(results) != 2 {
		t.Errorf("expected 2 inputs, got %d", len(results))
	}
}

func TestSelectByID(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "#myForm")
	if len(results) != 1 {
		t.Errorf("expected 1 element with id=myForm, got %d", len(results))
	}
	if results[0].Data != "form" {
		t.Errorf("expected form, got %s", results[0].Data)
	}
}

func TestSelectByClass(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, ".el-input__inner")
	if len(results) != 2 {
		t.Errorf("expected 2 elements with class el-input__inner, got %d", len(results))
	}
}

func TestCommaSelector(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "input,button")
	if len(results) != 3 {
		t.Errorf("expected 3 (2 inputs + 1 button), got %d", len(results))
	}
}

func TestAttributeSelector(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "input[type='text']")
	if len(results) != 1 {
		t.Errorf("expected 1 input with type=text, got %d", len(results))
	}
}

func TestDescendantSelector(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "form input")
	if len(results) != 2 {
		t.Errorf("expected 2 inputs inside form, got %d", len(results))
	}
}

func TestChildCombinator(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "form > div")
	if len(results) != 1 {
		t.Errorf("expected 1 div child of form, got %d", len(results))
	}
}

func TestCombinedSelector(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "button.btn-primary")
	if len(results) != 1 {
		t.Errorf("expected 1 button.btn-primary, got %d", len(results))
	}
}

func TestQuerySelector(t *testing.T) {
	root := buildTestTree()
	result := QuerySelector(root, "input")
	if result == nil {
		t.Error("expected non-nil result")
	}
	if result.Data != "input" {
		t.Errorf("expected input, got %s", result.Data)
	}
}

func TestNonExistentSelector(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "select")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAttributeSelectorNoValue(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "input[type]")
	if len(results) != 2 {
		t.Errorf("expected 2 inputs with type attr, got %d", len(results))
	}
}

func TestFirstChildPseudoClass(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "form > :first-child")
	if len(results) != 1 {
		t.Errorf("expected 1 first-child, got %d", len(results))
	}
	if results[0].Data != "input" {
		t.Errorf("expected input as first-child, got %s", results[0].Data)
	}
}

func TestLastChildPseudoClass(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "form > :last-child")
	if len(results) != 1 {
		t.Errorf("expected 1 last-child, got %d", len(results))
	}
	if results[0].Data != "button" {
		t.Errorf("expected button as last-child, got %s", results[0].Data)
	}
}

func TestNthChildPseudoClass(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "form > :nth-child(2)")
	if len(results) != 1 {
		t.Errorf("expected 1 nth-child(2), got %d", len(results))
	}
	if results[0].Data != "input" {
		t.Errorf("expected 2nd input as nth-child(2), got %s", results[0].Data)
	}
}

func TestNthChildOddPseudoClass(t *testing.T) {
	root := buildTestTree()
	results := QuerySelectorAll(root, "form > :nth-child(odd)")
	// 1st=input, 3rd=div -> 2 odd elements
	if len(results) != 2 {
		t.Errorf("expected 2 odd children, got %d", len(results))
	}
}
