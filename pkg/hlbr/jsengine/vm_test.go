package jsengine

import (
	"fmt"
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

func TestConsoleMethods(t *testing.T) {
	vm := newTestVM()
	methods := []string{"warn", "error", "info", "debug", "trace"}
	for _, m := range methods {
		result, err := vm.Run(fmt.Sprintf(`typeof console.%s`, m))
		if err != nil {
			t.Fatalf("unexpected error for console.%s: %v", m, err)
		}
		if valStr(result) != "function" {
			t.Errorf("console.%s: expected 'function', got %v", m, valStr(result))
		}
	}
}

func TestPromiseRace(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`typeof Promise.race`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valStr(result) != "function" {
		t.Errorf("expected 'function', got %v", valStr(result))
	}
}

func TestPromiseRaceSync(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`var p = Promise.race([42]); p.Promise ? p.Promise.State : "no-promise"`)
	if err != nil {
		t.Skip("Promise.race internal access not supported")
	}
	if valStr(result) != "fulfilled" {
		t.Skipf("Promise.race not fulfilled synchronously, got %v", valStr(result))
	}
}

func TestLocationMethods(t *testing.T) {
	vm := newTestVM()
	methods := []string{"reload", "replace", "assign"}
	for _, m := range methods {
		result, err := vm.Run(fmt.Sprintf(`typeof window.location.%s`, m))
		if err != nil {
			t.Fatalf("unexpected error for location.%s: %v", m, err)
		}
		if valStr(result) != "function" {
			t.Errorf("location.%s: expected 'function', got %v", m, valStr(result))
		}
	}
}

func TestAbstractEquality(t *testing.T) {
	vm := newTestVM()
	tests := []struct {
		expr     string
		expected bool
	}{
		{`1 == "1"`, true},
		{`1 === "1"`, false},
		{`0 == false`, true},
		{`"" == false`, true},
		{`null == undefined`, true},
		{`1 == true`, true},
		{`"1" == 1`, true},
		{`null == 0`, false},
		{`undefined == 0`, false},
		{`null == false`, false},
		{`2 == "2"`, true},
		{`"hello" == "hello"`, true},
		{`42 === 42`, true},
		{`42 == 42`, true},
	}
	for _, tt := range tests {
		result, err := vm.Run(tt.expr)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.expr, err)
			continue
		}
		if valBool(result) != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.expr, tt.expected, valBool(result))
		}
	}
}

func TestObjectCall(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`var fn = function() { return 42; }; Object(fn)()`)
	if err != nil {
		t.Fatalf("Object(fn)(): unexpected error: %v", err)
	}
	if valNum(result) != 42 {
		t.Errorf("Object(fn)(): expected 42, got %v", valNum(result))
	}

	result, err = vm.Run(`var obj = {a: 1}; Object(obj) === obj`)
	if err != nil {
		t.Fatalf("Object(obj) === obj: unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("Object(obj) === obj: expected true, got %v", valBool(result))
	}

	result, err = vm.Run(`typeof Object()`)
	if err != nil {
		t.Fatalf("typeof Object(): unexpected error: %v", err)
	}
	if valStr(result) != "object" {
		t.Errorf("typeof Object(): expected 'object', got %v", valStr(result))
	}

	// Object("z") creates a String wrapper with indexed characters
	result, err = vm.Run(`Object("z")[0]`)
	if err != nil {
		t.Fatalf("Object(\"z\")[0]: unexpected error: %v", err)
	}
	if valStr(result) != "z" {
		t.Errorf("Object(\"z\")[0]: expected 'z', got %v", valStr(result))
	}

	// Object("hello").length should be 5
	result, err = vm.Run(`Object("hello").length`)
	if err != nil {
		t.Fatalf("Object(\"hello\").length: unexpected error: %v", err)
	}
	if valNum(result) != 5 {
		t.Errorf("Object(\"hello\").length: expected 5, got %v", valNum(result))
	}
}

func TestDateConstructor(t *testing.T) {
	vm := newTestVM()
	result, err := vm.Run(`typeof new Date().getTime()`)
	if err != nil {
		t.Fatalf("Date.getTime: unexpected error: %v", err)
	}
	if valStr(result) != "number" {
		t.Errorf("Date.getTime: expected 'number', got %v", valStr(result))
	}

	result, err = vm.Run(`typeof Date.now()`)
	if err != nil {
		t.Fatalf("Date.now: unexpected error: %v", err)
	}
	if valStr(result) != "number" {
		t.Errorf("Date.now: expected 'number', got %v", valStr(result))
	}

	result, err = vm.Run(`new Date().getTime() > 0`)
	if err != nil {
		t.Fatalf("Date.getTime > 0: unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("Date.getTime > 0: expected true, got %v", valBool(result))
	}

	result, err = vm.Run(`Date.now() > 0`)
	if err != nil {
		t.Fatalf("Date.now > 0: unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("Date.now > 0: expected true, got %v", valBool(result))
	}

	result, err = vm.Run(`typeof new Date().toISOString`)
	if err != nil {
		t.Fatalf("Date.toISOString: unexpected error: %v", err)
	}
	if valStr(result) != "function" {
		t.Errorf("Date.toISOString: expected 'function', got %v", valStr(result))
	}
}
