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

func TestMap_Methods(t *testing.T) {
	// Create a map with pairs
	pairs := make(map[HashKey]MapPair)
	keyA := &String{Value: "a"}
	keyB := &String{Value: "b"}
	pairs[keyA.HashKey()] = MapPair{Key: keyA, Value: &Int{Value: 1}}
	pairs[keyB.HashKey()] = MapPair{Key: keyB, Value: &Int{Value: 2}}
	m := &Map{Pairs: pairs}

	tests := []struct {
		name     string
		method   string
		args     []Object
		expected Object
	}{
		{"len", "len", nil, &Int{Value: 2}},
		{"hasKey true", "hasKey", []Object{&String{Value: "a"}}, TRUE},
		{"hasKey false", "hasKey", []Object{&String{Value: "c"}}, FALSE},
		{"typeOf", "typeOf", nil, &String{Value: "MAP"}},
		{"toStr", "toStr", nil, &String{Value: m.Inspect()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, ok := GetMethod(MapType, tt.method)
			if !ok {
				t.Fatalf("method %s not found for MapType", tt.method)
			}

			// Build args with receiver as first argument
			args := []Object{m}
			args = append(args, tt.args...)

			result := method.Fn(args...)
			if isError(result) {
				t.Fatalf("unexpected error: %v", result.Inspect())
			}
			compareObjectsForTest(t, result, tt.expected)
		})
    }
}
