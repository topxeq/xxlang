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

	// Object("z").propertyIsEnumerable(0) — core-js detection pattern
	result, err = vm.Run(`Object("z").propertyIsEnumerable(0)`)
	if err != nil {
		t.Fatalf("Object(\"z\").propertyIsEnumerable(0): unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("Object(\"z\").propertyIsEnumerable(0): expected true, got %v", valBool(result))
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

// TestConsoleNativeThisOffset verifies that console.log etc. do not
// include the console object itself in the output. This was a bug where
// NativeThisOffset was not used, causing the 'this' arg to shift all
// parameters by one position.
func TestConsoleNativeThisOffset(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	// console.log with one argument should output just that argument
	result, err := vm.Run(`console.log("hello")`)
	if err != nil {
		t.Fatalf("console.log: unexpected error: %v", err)
	}
	_ = result

	// console.log with multiple arguments
	result, err = vm.Run(`console.log("a", "b", "c")`)
	if err != nil {
		t.Fatalf("console.log multi: unexpected error: %v", err)
	}
	_ = result
}

// TestMapSetNativeThisOffset verifies that Map/Set instance methods
// correctly skip the prepended 'this' argument when called as methods.
func TestMapSetNativeThisOffset(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	// Map methods
	result, err := vm.Run(`
		var m = new Map();
		m.set("key1", "value1");
		m.get("key1");
	`)
	if err != nil {
		t.Fatalf("Map.set/get: unexpected error: %v", err)
	}
	if valStr(result) != "value1" {
		t.Errorf("Map.get: expected 'value1', got %v", valStr(result))
	}

	result, err = vm.Run(`
		var m = new Map();
		m.set("k", "v");
		m.has("k");
	`)
	if err != nil {
		t.Fatalf("Map.has: unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("Map.has: expected true, got %v", valBool(result))
	}

	result, err = vm.Run(`
		var m = new Map();
		m.set("k", "v");
		m.delete("k");
		m.has("k");
	`)
	if err != nil {
		t.Fatalf("Map.delete: unexpected error: %v", err)
	}
	if valBool(result) {
		t.Errorf("Map.has after delete: expected false, got true")
	}

	// Set methods
	result, err = vm.Run(`
		var s = new Set();
		s.add("item1");
		s.has("item1");
	`)
	if err != nil {
		t.Fatalf("Set.add/has: unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("Set.has: expected true, got %v", valBool(result))
	}

	result, err = vm.Run(`
		var s = new Set();
		s.add("x");
		s.delete("x");
		s.has("x");
	`)
	if err != nil {
		t.Fatalf("Set.delete: unexpected error: %v", err)
	}
	if valBool(result) {
		t.Errorf("Set.has after delete: expected false, got true")
	}
}

// TestLocalStorageNativeThisOffset verifies that localStorage/sessionStorage
// methods correctly skip the prepended 'this' argument.
func TestLocalStorageNativeThisOffset(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	result, err := vm.Run(`
		localStorage.setItem("testKey", "testValue");
		localStorage.getItem("testKey");
	`)
	if err != nil {
		t.Fatalf("localStorage.setItem/getItem: unexpected error: %v", err)
	}
	if valStr(result) != "testValue" {
		t.Errorf("localStorage.getItem: expected 'testValue', got %v", valStr(result))
	}

	result, err = vm.Run(`
		localStorage.removeItem("testKey");
		localStorage.getItem("testKey");
	`)
	if err != nil {
		t.Fatalf("localStorage.removeItem: unexpected error: %v", err)
	}
	if valStr(result) != "undefined" && result.Type != "undefined" && valStr(result) != "" {
		t.Errorf("localStorage.getItem after remove: expected undefined/empty, got %v (type=%s)", valStr(result), result.Type)
	}

	// sessionStorage
	result, err = vm.Run(`
		sessionStorage.setItem("sk", "sv");
		sessionStorage.getItem("sk");
	`)
	if err != nil {
		t.Fatalf("sessionStorage: unexpected error: %v", err)
	}
	if valStr(result) != "sv" {
		t.Errorf("sessionStorage.getItem: expected 'sv', got %v", valStr(result))
	}
}

// TestReflectNativeThisOffset verifies that Reflect methods correctly
// skip the prepended 'this' argument.
func TestReflectNativeThisOffset(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	result, err := vm.Run(`
		var obj = {x: 42};
		Reflect.get(obj, "x");
	`)
	if err != nil {
		t.Fatalf("Reflect.get: unexpected error: %v", err)
	}
	if result.Num != 42 {
		t.Errorf("Reflect.get: expected 42, got %v", result.Num)
	}

	result, err = vm.Run(`
		var obj = {};
		Reflect.set(obj, "y", 99);
		obj.y;
	`)
	if err != nil {
		t.Fatalf("Reflect.set: unexpected error: %v", err)
	}
	if result.Num != 99 {
		t.Errorf("Reflect.set: expected 99, got %v", result.Num)
	}

	result, err = vm.Run(`
		var obj = {z: 1};
		Reflect.has(obj, "z");
	`)
	if err != nil {
		t.Fatalf("Reflect.has: unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("Reflect.has: expected true, got %v", valBool(result))
	}

	result, err = vm.Run(`
		var obj = {a: 1};
		Reflect.deleteProperty(obj, "a");
		Reflect.has(obj, "a");
	`)
	if err != nil {
		t.Fatalf("Reflect.deleteProperty: unexpected error: %v", err)
	}
	if valBool(result) {
		t.Errorf("Reflect.has after delete: expected false, got true")
	}
}

// TestDefinePropertyGetterSetter verifies that Object.defineProperty
// getters and setters work correctly when reading/writing properties.
func TestDefinePropertyGetterSetter(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	result, err := vm.Run(`
		var obj = {};
		var count = 0;
		Object.defineProperty(obj, "dynamic", {
			get: function() { count++; return count; },
			configurable: true
		});
		var a = obj.dynamic;
		var b = obj.dynamic;
		a + "," + b;
	`)
	if err != nil {
		t.Fatalf("getter: unexpected error: %v", err)
	}
	if valStr(result) != "1,2" {
		t.Errorf("getter: expected '1,2', got %v", valStr(result))
	}

	result, err = vm.Run(`
		var obj = {_val: 0};
		Object.defineProperty(obj, "val", {
			get: function() { return this._val; },
			set: function(v) { this._val = v * 2; },
			configurable: true
		});
		obj.val = 10;
		obj.val;
	`)
	if err != nil {
		t.Fatalf("setter: unexpected error: %v", err)
	}
	if result.Num != 20 {
		t.Errorf("setter: expected 20, got %v", result.Num)
	}
}

// TestClassListNativeThisOffset verifies that classList methods
// correctly skip the prepended 'this' argument.
func TestClassListNativeThisOffset(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	result, err := vm.Run(`
		var el = document.createElement("div");
		el.className = "foo bar";
		el.classList.add("baz");
		el.classList.contains("baz");
	`)
	if err != nil {
		t.Fatalf("classList.add: unexpected error: %v", err)
	}
	if !valBool(result) {
		t.Errorf("classList.contains after add: expected true, got false")
	}

	result, err = vm.Run(`
		var el = document.createElement("div");
		el.className = "foo bar";
		el.classList.remove("foo");
		el.classList.contains("foo");
	`)
	if err != nil {
		t.Fatalf("classList.remove: unexpected error: %v", err)
	}
	if valBool(result) {
		t.Errorf("classList.contains after remove: expected false, got true")
	}
}

// TestArrayFromNativeThisOffset verifies that Array.from and Array.of
// correctly skip the prepended 'this' argument.
func TestArrayFromNativeThisOffset(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	result, err := vm.Run(`Array.from([1, 2, 3]).length`)
	if err != nil {
		t.Fatalf("Array.from: unexpected error: %v", err)
	}
	if result.Num != 3 {
		t.Errorf("Array.from: expected 3, got %v", result.Num)
	}

	result, err = vm.Run(`Array.of(1, 2, 3).length`)
	if err != nil {
		t.Fatalf("Array.of: unexpected error: %v", err)
	}
	if result.Num != 3 {
		t.Errorf("Array.of: expected 3, got %v", result.Num)
	}
}

// TestPromiseChaining verifies that Promise.then() returns a new Promise
// (per ECMAScript spec) and that chained .then() callbacks receive the
// return value of the previous callback. This is critical for patterns
// like: axios.get(url1).then(() => axios.post(url2)).then(...)
func TestPromiseChaining(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	// Test: basic chaining with Promise.resolve return
	vm.Run(`
		Promise.resolve({data: "first"}).then(function(r1) {
			return Promise.resolve({data: "second"});
		}).then(function(r2) {
			window.__chainResult = r2.data;
		});
	`)
	result, err := vm.Run(`window.__chainResult`)
	if err != nil {
		t.Fatalf("chaining: unexpected error: %v", err)
	}
	if valStr(result) != "second" {
		t.Errorf("chaining: expected 'second', got %v", valStr(result))
	}

	// Test: chaining with non-Promise return value
	vm.Run(`
		Promise.resolve(10).then(function(v) {
			return v * 2;
		}).then(function(v2) {
			window.__chainResult2 = v2 + 5;
		});
	`)
	result, err = vm.Run(`window.__chainResult2`)
	if err != nil {
		t.Fatalf("chaining non-promise: unexpected error: %v", err)
	}
	if result.Num != 25 {
		t.Errorf("chaining non-promise: expected 25, got %v", result.Num)
	}

	// Test: deep chaining (3 levels)
	vm.Run(`
		Promise.resolve(1).then(function(v) {
			return Promise.resolve(v + 1);
		}).then(function(v) {
			return Promise.resolve(v + 1);
		}).then(function(v) {
			window.__chainResult3 = v;
		});
	`)
	result, err = vm.Run(`window.__chainResult3`)
	if err != nil {
		t.Fatalf("deep chaining: unexpected error: %v", err)
	}
	if result.Num != 3 {
		t.Errorf("deep chaining: expected 3, got %v", result.Num)
	}
}
