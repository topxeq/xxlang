// pkg/objects/map_test.go
package objects

import "testing"

func TestMapInspect(t *testing.T) {
	pairs := make(map[HashKey]MapPair)
	key := &String{Value: "a"}
	pairs[key.HashKey()] = MapPair{
		Key:   key,
		Value: &Int{Value: 1},
	}
	m := &Map{Pairs: pairs}

	// Map order is not guaranteed, just check it contains the right parts
	got := m.Inspect()
	if got[0] != '{' || got[len(got)-1] != '}' {
		t.Errorf("Map.Inspect() = %s, expected to start with { and end with }", got)
	}
}

func TestMapType(t *testing.T) {
	m := &Map{Pairs: map[HashKey]MapPair{}}
	if got := m.Type(); got != MapType {
		t.Errorf("Map.Type() = %s, want MAP", got)
	}
}

func TestMapToBool(t *testing.T) {
	empty := &Map{Pairs: map[HashKey]MapPair{}}
	if empty.ToBool() != FALSE {
		t.Error("Map({}).ToBool() should be FALSE")
	}

	pairs := make(map[HashKey]MapPair)
	key := &String{Value: "a"}
	pairs[key.HashKey()] = MapPair{
		Key:   key,
		Value: &Int{Value: 1},
	}
	nonempty := &Map{Pairs: pairs}
	if nonempty.ToBool() != TRUE {
		t.Error("Map({a: 1}).ToBool() should be TRUE")
	}
}
