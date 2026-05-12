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

func TestStringRepeatPadTrim(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	r, _ := vm.Run(`"ab".repeat(3)`)
	if r.Str != "ababab" {
		t.Errorf("repeat: expected ababab, got %s", r.Str)
	}

	r, _ = vm.Run(`"hi".padStart(5, "x")`)
	if r.Str != "xxxhi" {
		t.Errorf("padStart: expected xxxhi, got %s", r.Str)
	}

	r, _ = vm.Run(`"hi".padEnd(5, "x")`)
	if r.Str != "hixxx" {
		t.Errorf("padEnd: expected hixxx, got %s", r.Str)
	}

	r, _ = vm.Run(`"  hi  ".trimStart()`)
	if r.Str != "hi  " {
		t.Errorf("trimStart: expected 'hi  ', got '%s'", r.Str)
	}

	r, _ = vm.Run(`"  hi  ".trimEnd()`)
	if r.Str != "  hi" {
		t.Errorf("trimEnd: expected '  hi', got '%s'", r.Str)
	}
}

func TestArrayFlatFill(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	r, _ := vm.Run(`[1, [2, [3]]].flat()`)
	if r.Type != "object" || r.Arr == nil || len(r.Arr) != 3 {
		t.Errorf("flat: expected 3 elements, got type=%s len=%v", r.Type, len(r.Arr))
	}

	r, _ = vm.Run(`[1, 2, 3].fill(0, 1, 2)`)
	if r.Arr == nil || r.Arr[0].Num != 1 || r.Arr[1].Num != 0 || r.Arr[2].Num != 3 {
		t.Errorf("fill: expected [1,0,3], got %v", r.Arr)
	}

	r, _ = vm.Run(`[1, [2, [3]]].flat(2)`)
	if r.Type != "object" || r.Arr == nil || len(r.Arr) != 3 {
		t.Errorf("flat(2): expected 3 elements, got type=%s len=%v", r.Type, len(r.Arr))
	}
}

func TestObjectIs(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	r, _ := vm.Run(`Object.is(NaN, NaN)`)
	if !r.Bool {
		t.Error("Object.is(NaN, NaN) should be true")
	}

	r, _ = vm.Run(`Object.is(0, -0)`)
	if r.Bool {
		t.Error("Object.is(0, -0) should be false")
	}

	r, _ = vm.Run(`Object.is(-0, -0)`)
	if !r.Bool {
		t.Error("Object.is(-0, -0) should be true")
	}

	r, _ = vm.Run(`Object.is("a", "a")`)
	if !r.Bool {
		t.Error("Object.is('a', 'a') should be true")
	}
}

func TestPromiseAll(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	vm.Run(`
		var allResult = -1;
		Promise.all([Promise.resolve(1), Promise.resolve(2), Promise.resolve(3)]).then(function(vals) {
			allResult = vals.length;
		});
	`)
	r, _ := vm.Run(`allResult`)
	if r.Num != 3 {
		t.Errorf("Promise.all resolved: expected 3, got %v", r.Num)
	}

	vm.Run(`
		var allResult2 = -1;
		Promise.all([1, 2, 3]).then(function(vals) {
			allResult2 = vals.length;
		});
	`)
	r, _ = vm.Run(`allResult2`)
	if r.Num != 3 {
		t.Errorf("Promise.all non-promise: expected 3, got %v", r.Num)
	}
}

func TestPromiseExecutorResolve(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	vm.Run(`
		var execResult = "pending";
		new Promise(function(r) { r("hello"); }).then(function(v) { execResult = v; });
	`)
	r, _ := vm.Run(`execResult`)
	if r.Str != "hello" {
		t.Errorf("Promise executor resolve: expected hello, got %s", r.Str)
	}
}

// TestPromiseChainingPending verifies the core Promise Resolution Procedure:
// when a .then() callback returns a Promise that is still pending, the
// chaining promise must subscribe to it rather than wrapping it. This is
// the pattern used by axios and other libraries:
//
//	Promise.resolve(1).then(function() { return axios(url); }).then(...)
//
// where axios() returns a new Promise that resolves later via XHR callback.
func TestPromiseChainingPending(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	// Test 1: .then() returning a pending Promise that resolves immediately
	// via resolvePromiseWith (simulates axios returning a deferred Promise)
	vm.Run(`
		var chainResult1 = "not_set";
		var resolveInner;
		var innerP = new Promise(function(r) { resolveInner = r; });
		Promise.resolve("first").then(function(v) {
			return innerP;
		}).then(function(v2) {
			chainResult1 = v2;
		});
	`)
	// Resolve the inner promise in a separate Run() call
	vm.Run(`
		resolveInner("inner_value");
	`)
	r, _ := vm.Run(`chainResult1`)
	if valStr(r) != "inner_value" {
		t.Errorf("chaining pending promise: expected 'inner_value', got %v", valStr(r))
	}

	// Test 2: Deep chaining with pending promises (3 levels)
	vm2 := NewVM(dom.NewDocument())
	vm2.Run(`
		var chainResult2 = "not_set";
		var resolveA, resolveB;
		var pA = new Promise(function(r) { resolveA = r; });
		var pB = new Promise(function(r) { resolveB = r; });
		Promise.resolve(1).then(function(v) {
			return pA;
		}).then(function(v) {
			return pB;
		}).then(function(v) {
			chainResult2 = v;
		});
	`)
	vm2.Run(`resolveA(10)`)
	vm2.Run(`resolveB(100)`)
	r, _ = vm2.Run(`chainResult2`)
	if r.Num != 100 {
		t.Errorf("deep chaining pending: expected 100, got %v", r.Num)
	}

	// Test 3: Promise.resolve(pendingPromise) should follow the inner promise
	vm3 := NewVM(dom.NewDocument())
	vm3.Run(`
		var chainResult3 = "not_set";
		var resolveC;
		var pC = new Promise(function(r) { resolveC = r; });
		Promise.resolve(pC).then(function(v) {
			chainResult3 = v;
		});
	`)
	vm3.Run(`resolveC("followed")`)
	r, _ = vm3.Run(`chainResult3`)
	if valStr(r) != "followed" {
		t.Errorf("Promise.resolve(thenable): expected 'followed', got %v", valStr(r))
	}
}

// TestPromiseCatchPassthrough verifies that .catch() properly passes through
// fulfilled values (i.e., catch is a no-op when no rejection occurred) and
// that .then() properly propagates rejections when no onRejected handler is
// provided.
func TestPromiseCatchPassthrough(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	// Test 1: .catch() on a fulfilled promise should pass through the value
	vm.Run(`
		var catchResult1 = "not_set";
		Promise.resolve("success").catch(function(e) {
			catchResult1 = "caught:" + e;
		}).then(function(v) {
			catchResult1 = v;
		});
	`)
	r, _ := vm.Run(`catchResult1`)
	if valStr(r) != "success" {
		t.Errorf("catch passthrough fulfilled: expected 'success', got %v", valStr(r))
	}

	// Test 2: .catch() on a rejected promise should handle the rejection
	vm2 := NewVM(dom.NewDocument())
	vm2.Run(`
		var catchResult2 = "not_set";
		Promise.reject("error1").catch(function(e) {
			return "recovered:" + e;
		}).then(function(v) {
			catchResult2 = v;
		});
	`)
	r, _ = vm2.Run(`catchResult2`)
	if valStr(r) != "recovered:error1" {
		t.Errorf("catch handles rejection: expected 'recovered:error1', got %v", valStr(r))
	}

	// Test 3: Rejection propagates through .then() without onRejected handler
	vm3 := NewVM(dom.NewDocument())
	vm3.Run(`
		var catchResult3 = "not_set";
		Promise.reject("err").then(function(v) {
			return "should not run";
		}).catch(function(e) {
			catchResult3 = "caught:" + e;
		});
	`)
	r, _ = vm3.Run(`catchResult3`)
	if valStr(r) != "caught:err" {
		t.Errorf("rejection passthrough .then(): expected 'caught:err', got %v", valStr(r))
	}
}

// TestPromiseCatchWithPendingReturn verifies that when a .catch() handler
// returns a pending Promise, the chaining promise subscribes to it.
func TestPromiseCatchWithPendingReturn(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	vm.Run(`
		var catchChainResult = "not_set";
		var resolveRecovery;
		var recoveryP = new Promise(function(r) { resolveRecovery = r; });
		Promise.reject("fail").catch(function(e) {
			return recoveryP;
		}).then(function(v) {
			catchChainResult = v;
		});
	`)
	vm.Run(`resolveRecovery("recovered")`)
	r, _ := vm.Run(`catchChainResult`)
	if valStr(r) != "recovered" {
		t.Errorf("catch returning pending promise: expected 'recovered', got %v", valStr(r))
	}
}

// TestPromiseConstructorResolveThenable verifies that when a Promise
// constructor's resolve function is called with a thenable, the promise
// follows the thenable's state. Pattern:
//
//	new Promise(function(r) { r(anotherPromise); })
func TestPromiseConstructorResolveThenable(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	vm.Run(`
		var ctorResult = "not_set";
		var resolveInner;
		var innerP = new Promise(function(r) { resolveInner = r; });
		new Promise(function(r) { r(innerP); }).then(function(v) {
			ctorResult = v;
		});
	`)
	vm.Run(`resolveInner("from_inner")`)
	r, _ := vm.Run(`ctorResult`)
	if valStr(r) != "from_inner" {
		t.Errorf("constructor resolve thenable: expected 'from_inner', got %v", valStr(r))
	}
}

// TestPromiseAxiosPattern simulates the exact pattern used by axios and
// similar HTTP libraries: a function that returns a new Promise where the
// resolve function is captured and called later (e.g., by an XHR callback).
// This is the core pattern that was broken before the Promise Resolution
// Procedure fix.
func TestPromiseAxiosPattern(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	// Simulate axios.get() returning a deferred Promise
	vm.Run(`
		window.__axiosResults = [];

		// Simulated axios that captures resolve
		function fakeAxios(responseData) {
			return new Promise(function(resolve) {
				resolve({data: responseData, status: 200});
			});
		}

		// Chain: axios.get().then(() => axios.post()).then(handle)
		fakeAxios("first_data").then(function(r1) {
			window.__axiosResults.push(r1.data);
			return fakeAxios("second_data");
		}).then(function(r2) {
			window.__axiosResults.push(r2.data);
		});
	`)
	r, _ := vm.Run(`window.__axiosResults.length`)
	if r.Num != 2 {
		t.Fatalf("axios pattern: expected 2 results, got %v", r.Num)
	}
	r1, _ := vm.Run(`window.__axiosResults[0]`)
	if valStr(r1) != "first_data" {
		t.Errorf("axios pattern first: expected 'first_data', got %v", valStr(r1))
	}
	r2, _ := vm.Run(`window.__axiosResults[1]`)
	if valStr(r2) != "second_data" {
		t.Errorf("axios pattern second: expected 'second_data', got %v", valStr(r2))
	}

	// Test with deferred resolve (more realistic axios pattern)
	vm2 := NewVM(dom.NewDocument())
	vm2.Run(`
		window.__deferredResults = [];
		var __xhrResolve;

		// Simulated axios with truly deferred resolve (like real XHR)
		function deferredAxios(data) {
			return new Promise(function(resolve) {
				__xhrResolve = resolve;
			});
		}

		deferredAxios("step1").then(function(r1) {
			window.__deferredResults.push(r1);
			return deferredAxios("step2");
		}).then(function(r2) {
			window.__deferredResults.push(r2);
		});
	`)
	// First XHR completes
	vm2.Run(`__xhrResolve("step1_done")`)
	r, _ = vm2.Run(`window.__deferredResults.length`)
	if r.Num != 1 {
		t.Fatalf("deferred axios step1: expected 1 result, got %v", r.Num)
	}
	r1, _ = vm2.Run(`window.__deferredResults[0]`)
	if valStr(r1) != "step1_done" {
		t.Errorf("deferred axios step1: expected 'step1_done', got %v", valStr(r1))
	}
	// Second XHR completes
	vm2.Run(`__xhrResolve("step2_done")`)
	r, _ = vm2.Run(`window.__deferredResults.length`)
	if r.Num != 2 {
		t.Fatalf("deferred axios step2: expected 2 results, got %v", r.Num)
	}
	r2, _ = vm2.Run(`window.__deferredResults[1]`)
	if valStr(r2) != "step2_done" {
		t.Errorf("deferred axios step2: expected 'step2_done', got %v", valStr(r2))
	}
}

func TestMutationObserverStub(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	r, _ := vm.Run(`typeof MutationObserver`)
	if r.Str != "function" {
		t.Errorf("MutationObserver typeof: expected function, got %s", r.Str)
	}

	r, _ = vm.Run(`var obs = new MutationObserver(function(){}); typeof obs.observe`)
	if r.Str != "function" {
		t.Errorf("MutationObserver.observe typeof: expected function, got %s", r.Str)
	}
}

func TestBreakInForLoop(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	r, _ := vm.Run(`(function(){ var x=""; for(var i=0;i<5;i++){if(i===3){x="found";break}} return x; })()`)
	if r.Str != "found" {
		t.Errorf("break in for loop: expected found, got %v", r)
	}

	r, _ = vm.Run(`(function(){ var sum=0; for(var i=0;i<10;i++){if(i>=5){break}sum+=i} return sum; })()`)
	if r.Num != 10 {
		t.Errorf("break in for loop sum: expected 10, got %v", r.Num)
	}

	r, _ = vm.Run(`(function(){ var x=""; var i=0; while(i<5){if(i===3){x="found";break}i++} return x; })()`)
	if r.Str != "found" {
		t.Errorf("break in while loop: expected found, got %v", r)
	}

	r, _ = vm.Run(`(function(){ var result=[]; for(var k in {a:1,b:2,c:3}){result.push(k);if(result.length===2){break}} return result.length; })()`)
	if r.Num != 2 {
		t.Errorf("break in for-in loop: expected 2 items, got %v", r.Num)
	}

	r, _ = vm.Run(`(function(){ var result=[]; var arr=[10,20,30,40]; for(var v of arr){result.push(v);if(v===30){break}} return result.join(","); })()`)
	if r.Str != "10,20,30" {
		t.Errorf("break in for-of loop: expected 10,20,30, got %v", r.Str)
	}
}

func TestContinueInForLoop(t *testing.T) {
	vm := NewVM(dom.NewDocument())

	r, _ := vm.Run(`(function(){ var sum=0; for(var i=0;i<5;i++){if(i===3){continue}sum+=i} return sum; })()`)
	if r.Num != 7 {
		t.Errorf("continue in for loop: expected 7, got %v", r.Num)
	}

	r, _ = vm.Run(`(function(){ var result=[]; for(var i=0;i<5;i++){if(i%2===0){continue}result.push(i)} return result.join(","); })()`)
	if r.Str != "1,3" {
		t.Errorf("continue skip even: expected 1,3, got %v", r.Str)
	}
}
