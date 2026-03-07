// pkg/objects/string_test.go
package objects

import "testing"

func TestStringInspect(t *testing.T) {
	s := &String{Value: "hello"}
	if got := s.Inspect(); got != "hello" {
		t.Errorf("String.Inspect() = %s, want hello", got)
	}
}

func TestStringType(t *testing.T) {
	s := &String{Value: "hello"}
	if got := s.Type(); got != StringType {
		t.Errorf("String.Type() = %s, want STRING", got)
	}
}

func TestStringToBool(t *testing.T) {
	empty := &String{Value: ""}
	if empty.ToBool() != FALSE {
		t.Error("String(\"\").ToBool() should be FALSE")
	}

	nonempty := &String{Value: "hello"}
	if nonempty.ToBool() != TRUE {
		t.Error("String(\"hello\").ToBool() should be TRUE")
	}
}

func TestStringHashKey(t *testing.T) {
	a := &String{Value: "hello"}
	b := &String{Value: "hello"}
	c := &String{Value: "world"}

	if a.HashKey() != b.HashKey() {
		t.Error("same string values should have same hash keys")
	}
	if a.HashKey() == c.HashKey() {
		t.Error("different string values should have different hash keys")
	}
}
