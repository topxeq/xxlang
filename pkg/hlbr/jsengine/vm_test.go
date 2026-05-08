package jsengine

import (
	"testing"

	"github.com/topxeq/xxlang/pkg/hlbr/dom"
)

func newTestVM() *VM {
	doc := &dom.Document{Root: &dom.Node{Type: dom.ElementNode, Data: "html"}}
	return NewVM(doc)
}

func valNum(v *Value) float64 {
	if v == nil {
		return 0
	}
	return v.Num
}

func valStr(v *Value) string {
	if v == nil {
		return ""
	}
	return v.Str
}

func valBool(v *Value) bool {
	if v == nil {
		return false
	}
	return v.Bool
}

func TestBasicArithmetic(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run("2 + 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 5 {
		t.Errorf("expected 5, got %v", valNum(result))
	}
}

func TestVariableAssignment(t *testing.T) {
	vm := newTestVM()
	_, err := vm.Run("var x = 42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := vm.Run("x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 42 {
		t.Errorf("expected 42, got %v", valNum(result))
	}
}

func TestStringConcat(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`"hello" + " " + "world"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valStr(result) != "hello world" {
		t.Errorf("expected 'hello world', got %v", valStr(result))
	}
}

func TestArrayPush(t *testing.T) {
	vm := newTestVM()
	_, err := vm.Run("var arr = [1, 2, 3]; arr.push(4)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := vm.Run("arr.length")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 4 {
		t.Errorf("expected 4, got %v", valNum(result))
	}
}

func TestObjectKeys(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`Object.keys({a: 1, b: 2}).length`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 2 {
		t.Errorf("expected 2, got %v", valNum(result))
	}
}

func TestObjectAssign(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`Object.assign({a: 1}, {b: 2}).b`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 2 {
		t.Errorf("expected 2, got %v", valNum(result))
	}
}

func TestFunctionCall(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`(function() { return 42 })()`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 42 {
		t.Errorf("expected 42, got %v", valNum(result))
	}
}

func TestArrowFunction(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`(x => x * 2)(21)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 42 {
		t.Errorf("expected 42, got %v", valNum(result))
	}
}

func TestTemplateLiteral(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run("`hello ${'world'}`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valStr(result) != "hello world" {
		t.Errorf("expected 'hello world', got %v", valStr(result))
	}
}

func TestJSONParse(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`JSON.parse('{"a":1}').a`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 1 {
		t.Errorf("expected 1, got %v", valNum(result))
	}
}

func TestJSONStringify(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`JSON.stringify({a: 1})`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valStr(result) != `{"a":1}` {
		t.Errorf("expected '{\"a\":1}', got %v", valStr(result))
	}
}

func TestNumberIsInteger(t *testing.T) {
	vm := newTestVM()
	// Number.isInteger is installed as a polyfill at the browser level,
	// not as a VM built-in. Test it only if available.
	result, err := vm.Run(`typeof Number.isInteger`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valStr(result) != "function" {
		t.Skip("Number.isInteger not available in bare VM")
	}
	result, err = vm.Run(`Number.isInteger(42)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("expected true, got %v", valBool(result))
	}
}

func TestStringIncludes(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`"hello world".includes("world")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("expected true, got %v", valBool(result))
	}
}

func TestArrayFind(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`[1, 2, 3, 4].find(function(x) { return x > 2 })`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 3 {
		t.Errorf("expected 3, got %v", valNum(result))
	}
}

func TestDestructuring(t *testing.T) {
	vm := newTestVM()
	_, err := vm.Run(`var [a, b] = [1, 2]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := vm.Run("a + b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 3 {
		t.Errorf("expected 3, got %v", valNum(result))
	}
}

func TestSpreadOperator(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`var a = [1, 2]; var b = [...a, 3]; b.length`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 3 {
		t.Errorf("expected 3, got %v", valNum(result))
	}
}

func TestMathFloor(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`Math.floor(3.7)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valNum(result) != 3 {
		t.Errorf("expected 3, got %v", valNum(result))
	}
}

func TestTypeof(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`typeof 42`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valStr(result) != "number" {
		t.Errorf("expected 'number', got %v", valStr(result))
	}
}

func TestTernaryOperator(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`true ? "yes" : "no"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valStr(result) != "yes" {
		t.Errorf("expected 'yes', got %v", valStr(result))
	}
}
